package store

import (
	"strings"

	"github.com/FMakareev/muster-backlog/internal/backlog"
)

// Kin is a task's place among a parent and its subtasks.
//
// The format offers two signals for this one relationship and they are not
// equally useful. Every subtask carries parent_task_id, and its own id is the
// parent's with a segment appended - task-59.4 under TASK-59. The subtasks
// field exists in the 1.48.0 serialiser and is populated in none of the 92
// subtasks across the author's nine projects, so the forward direction is
// derived from the back-links rather than read from the file.
type Kin struct {
	// Parent is where the parent actually lives, which is not always where the
	// child does: 6 of those 92 have a parent in another directory, because a
	// parent can be completed or archived while its subtasks are not.
	Parent *Ref
	// ParentTitle saves the interface a second lookup for a link's text.
	ParentTitle string
	// Children are the subtasks, in the order the project scan produced.
	Children []Ref
	// Done counts children that have reached their project's last declared
	// status - the same definition analytics uses, since it is the only one
	// the format offers. Archived children are counted in neither Done nor
	// Children: archiving is a soft delete, and a board that does not show
	// them should not count them either.
	Done int
}

// KinIndex resolves every parent and subtask link the store holds.
//
// One pass over every project rather than a lookup per task: the board asks
// for hundreds of cards at once, and a query per card would be hundreds of
// scans of the same slice.
//
// Ids are matched case-insensitively and nothing else is normalised. Files
// write the id as TASK-7 and the reference as TASK-7, so all 92 links resolve
// on plain equality; stripping zero padding to make TASK-7 also match TASK-007
// would invent a match in a project that has both.
func (s *Store) KinIndex() map[Ref]Kin {
	s.mu.RLock()
	defer s.mu.RUnlock()

	out := map[Ref]Kin{}

	for _, path := range s.order {
		state, ok := s.projects[path]
		if !ok || !state.Visible() {
			continue
		}

		// Every class is indexed, so a parent that has since been completed is
		// still found. Archived tasks resolve as parents but never appear as
		// children, for the same reason they are absent from the board.
		byID := make(map[string]Ref, len(state.Scanned.Tasks))
		titles := make(map[string]string, len(state.Scanned.Tasks))
		for _, task := range state.Scanned.Tasks {
			ref := Ref{Project: path, Key: task.Key}
			id := strings.ToUpper(task.ID)
			// tasks/ wins over completed/ and archive/ for an id that appears
			// twice, which archiving makes possible inside one project.
			if _, taken := byID[id]; taken && task.Class != backlog.ClassActive {
				continue
			}
			byID[id] = ref
			titles[id] = task.Title
		}

		terminal := terminalStatus(state)

		for _, task := range state.Scanned.Tasks {
			declared := strings.TrimSpace(task.ParentTaskID)
			if declared == "" {
				continue
			}
			child := Ref{Project: path, Key: task.Key}
			parent, found := byID[strings.ToUpper(declared)]

			// A declared parent that no file answers to leaves Parent nil.
			// It happens in none of the 92 links today, but a hand-deleted
			// task would produce it, and the caller still has the declared id
			// to say so with.
			kin := out[child]
			if found {
				kin.Parent = &parent
				kin.ParentTitle = titles[strings.ToUpper(declared)]
			}
			out[child] = kin

			if !found || task.Class == backlog.ClassArchived {
				continue
			}
			up := out[parent]
			up.Children = append(up.Children, child)
			if isDone(task, terminal) {
				up.Done++
			}
			out[parent] = up
		}
	}

	return out
}

// terminalStatus is the last status a project declares, which is the only
// definition of finished the format offers.
func terminalStatus(state *ProjectState) string {
	list := state.Scanned.Config.Statuses
	if len(list) == 0 {
		return ""
	}
	return list[len(list)-1]
}

// isDone reports whether a task has finished, by status or by having been
// moved into completed/, which the CLI does on the way out.
func isDone(task backlog.Entity, terminal string) bool {
	if task.Class == backlog.ClassCompleted {
		return true
	}
	return terminal != "" && strings.EqualFold(task.Status, terminal)
}
