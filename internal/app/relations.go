package app

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/FMakareev/muster-backlog/internal/backlog"
	"github.com/FMakareev/muster-backlog/internal/backlogcli"
)

// SetDependencies replaces what a task waits on.
//
// Ids are checked against the project before the CLI is asked, because the
// commonest mistake here is a reference to a task in another project - ids
// collide freely across projects - and saying so without running a command is
// both faster and clearer than the CLI's own message.
func (s *BoardService) SetDependencies(projectPath, taskID string, ids []string) WriteResult {
	var missing []string
	for _, id := range ids {
		id = strings.TrimSpace(id)
		if id == "" {
			continue
		}
		if strings.EqualFold(id, taskID) {
			return WriteResult{Problem: &Problem{
				Kind: ProblemCLI, Title: "A task cannot wait on itself",
				Detail: fmt.Sprintf("%s depends on %s.", taskID, id), Path: projectPath,
			}}
		}
		if !s.taskLive(projectPath, id) {
			missing = append(missing, id)
		}
	}
	if len(missing) > 0 {
		return WriteResult{Problem: &Problem{
			Kind:  ProblemCLI,
			Title: "No such task in this project",
			Detail: fmt.Sprintf(
				"%s: ids resolve inside their own project only, and this one has no %s.",
				strings.Join(missing, ", "), plural(missing)),
			Path: projectPath,
		}}
	}

	return s.write(projectPath, fmt.Sprintf("%s's dependencies could not be changed", taskID),
		func(cli *backlogcli.Runner) error {
			return cli.SetDependencies(context.Background(),
				s.dataDirFor(projectPath), taskID, ids)
		})
}

// plural reads correctly for one id and for several.
func plural(ids []string) string {
	if len(ids) == 1 {
		return "such task"
	}
	return "such tasks"
}

// SetReferences replaces a task's references.
func (s *BoardService) SetReferences(projectPath, taskID string, refs []string) WriteResult {
	return s.write(projectPath, fmt.Sprintf("%s's references could not be changed", taskID),
		func(cli *backlogcli.Runner) error {
			return cli.SetReferences(context.Background(),
				s.dataDirFor(projectPath), taskID, refs)
		})
}

// SetDocumentation replaces a task's documentation links.
func (s *BoardService) SetDocumentation(projectPath, taskID string, docs []string) WriteResult {
	return s.write(projectPath, fmt.Sprintf("%s's documentation could not be changed", taskID),
		func(cli *backlogcli.Runner) error {
			return cli.SetDocumentation(context.Background(),
				s.dataDirFor(projectPath), taskID, docs)
		})
}

// SetModifiedFiles replaces a task's modified-file list.
//
// The list can be changed but not emptied, because Backlog.md has no way to
// empty it: --modified-file "" exits 0 and does nothing. Saying so is better
// than a control that appears to work.
func (s *BoardService) SetModifiedFiles(projectPath, taskID string, files []string) WriteResult {
	if len(backlogcli.NonEmpty(files)) == 0 {
		return WriteResult{Problem: &Problem{
			Kind:  ProblemCLI,
			Title: "This list cannot be emptied",
			Detail: "Backlog.md can set the files a task touches but has no " +
				"command to clear them. Edit the file by hand if the list is wrong.",
			Path: projectPath,
		}}
	}
	return s.write(projectPath, fmt.Sprintf("%s's file list could not be changed", taskID),
		func(cli *backlogcli.Runner) error {
			return cli.SetModifiedFiles(context.Background(),
				s.dataDirFor(projectPath), taskID, files)
		})
}

// ordinalStep is what the CLI allocates between tasks: (count+1) * 1000.
const ordinalStep = 1000

// Reorder puts a task in front of another one in its own column.
//
// beforeID is the task it was dropped in front of, or "" for the end of the
// column - which is exactly what the board's drag reports. The new ordinal is
// the midpoint between the neighbours, which is the shape the format already
// carries: two ordinals in the author's corpus are hand-set midpoints written
// by Backlog.md's own web interface.
//
// Ordinals are neither unique nor mandatory, so where there is no room between
// the neighbours the column is restacked at multiples of 1000 first. That is a
// write per task in one column of one project, and only when the gaps have
// been used up.
func (s *BoardService) Reorder(projectPath, taskID, beforeID string) WriteResult {
	column := s.columnOf(projectPath, taskID)
	if len(column) == 0 {
		return WriteResult{Problem: &Problem{
			Kind: ProblemCLI, Title: "No such task",
			Detail: fmt.Sprintf("%s is not on the board in this project.", taskID),
			Path:   projectPath,
		}}
	}

	ordinal, restack := placeAmong(column, taskID, beforeID)
	if restack {
		if result := s.restack(projectPath, column); !result.OK {
			return result
		}
		column = s.columnOf(projectPath, taskID)
		ordinal, _ = placeAmong(column, taskID, beforeID)
	}

	return s.write(projectPath, fmt.Sprintf("%s could not be reordered", taskID),
		func(cli *backlogcli.Runner) error {
			return cli.SetOrdinal(context.Background(),
				s.dataDirFor(projectPath), taskID, ordinal)
		})
}

// columnOf returns the tasks sharing a task's status, in the order the board
// shows them.
func (s *BoardService) columnOf(projectPath, taskID string) []backlog.Entity {
	var status string
	var all []backlog.Entity
	for _, item := range s.store.Entities(backlog.KindTask) {
		if item.Ref.Project != projectPath || item.Ref.Class != backlog.ClassActive {
			continue
		}
		all = append(all, item.Entity)
		if strings.EqualFold(item.Ref.ID, taskID) {
			status = item.Entity.Status
		}
	}
	if status == "" {
		return nil
	}

	var column []backlog.Entity
	for _, e := range all {
		if strings.EqualFold(e.Status, status) {
			column = append(column, e)
		}
	}
	// The store hands them over already sorted the way Backlog.md sorts them;
	// this keeps that order stable if the source ever stops guaranteeing it.
	sort.SliceStable(column, func(i, j int) bool {
		a, b := column[i], column[j]
		if a.HasOrdinal != b.HasOrdinal {
			return a.HasOrdinal
		}
		if a.HasOrdinal && b.HasOrdinal && *a.Ordinal != *b.Ordinal {
			return *a.Ordinal < *b.Ordinal
		}
		return false
	})
	return column
}

// placeAmong works out the ordinal a task needs to sit in front of another,
// and says whether the column has to be restacked first for want of a gap.
func placeAmong(column []backlog.Entity, taskID, beforeID string) (int, bool) {
	var above, below *backlog.Entity
	for i := range column {
		if strings.EqualFold(column[i].ID, taskID) {
			continue
		}
		if beforeID != "" && strings.EqualFold(column[i].ID, beforeID) {
			below = &column[i]
			break
		}
		above = &column[i]
	}

	switch {
	case above == nil && below == nil:
		return ordinalStep, false
	case above == nil:
		// Dropped at the top: half of whatever is now second.
		if !below.HasOrdinal {
			return ordinalStep, false
		}
		if *below.Ordinal <= 1 {
			return 0, true
		}
		return *below.Ordinal / 2, false
	case below == nil:
		// Dropped at the end.
		if !above.HasOrdinal {
			return ordinalStep, false
		}
		return *above.Ordinal + ordinalStep, false
	default:
		if !above.HasOrdinal || !below.HasOrdinal {
			return 0, true
		}
		gap := *below.Ordinal - *above.Ordinal
		if gap < 2 {
			return 0, true
		}
		return *above.Ordinal + gap/2, false
	}
}

// restack renumbers a column at multiples of 1000, keeping its current order,
// so that a gap exists wherever one is needed.
func (s *BoardService) restack(projectPath string, column []backlog.Entity) WriteResult {
	for i, e := range column {
		id := e.ID
		want := (i + 1) * ordinalStep
		if e.HasOrdinal && *e.Ordinal == want {
			continue
		}
		result := s.write(projectPath, fmt.Sprintf("%s could not be reordered", id),
			func(cli *backlogcli.Runner) error {
				return cli.SetOrdinal(context.Background(),
					s.dataDirFor(projectPath), id, want)
			})
		if !result.OK {
			return result
		}
	}
	return WriteResult{OK: true}
}
