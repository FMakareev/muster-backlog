package app

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/FMakareev/muster-backlog/internal/backlog"
	"github.com/FMakareev/muster-backlog/internal/backlogcli"
)

// DraftView is one captured note waiting to be dealt with.
//
// Carries the whole entity like every other view, plus the one number the
// inbox exists to make visible: how long it has been sitting there.
type DraftView struct {
	TaskView
	// WaitingDays is whole days since capture, or -1 when the file carries no
	// created_date. Counted rather than formatted here, so the interface can
	// say "today" without parsing anything back.
	WaitingDays int `json:"waitingDays"`
}

// Drafts returns every waiting draft, oldest first.
//
// Oldest first because that is the order an inbox has to be emptied in: the
// note that has been ignored longest is the one that most needs a decision.
// Archived drafts are not here - discarding one is how it leaves.
func (s *BoardService) Drafts() []DraftView {
	now := time.Now()

	var out []DraftView
	for _, item := range s.store.Entities(backlog.KindDraft) {
		if item.Ref.Class != backlog.ClassActive {
			continue
		}
		view := DraftView{TaskView: taskView(item, nil), WaitingDays: -1}
		if !item.Entity.Created.IsZero() {
			view.WaitingDays = int(now.Sub(item.Entity.Created).Hours() / 24)
			if view.WaitingDays < 0 {
				// A file dated in the future is someone's clock, not a
				// negative wait.
				view.WaitingDays = 0
			}
		}
		out = append(out, view)
	}

	// Oldest first; undated notes last, since nothing is known about them.
	sortDrafts(out)
	return out
}

// sortDrafts orders by wait, longest first, and by project and id within a
// day so the list does not shuffle between refreshes.
func sortDrafts(drafts []DraftView) {
	sort.SliceStable(drafts, func(i, j int) bool {
		a, b := drafts[i], drafts[j]
		if a.WaitingDays != b.WaitingDays {
			// -1 means undated, which sorts last rather than first.
			switch {
			case a.WaitingDays < 0:
				return false
			case b.WaitingDays < 0:
				return true
			default:
				return a.WaitingDays > b.WaitingDays
			}
		}
		if a.ProjectName != b.ProjectName {
			return a.ProjectName < b.ProjectName
		}
		return a.ID < b.ID
	})
}

// PromoteDraft turns a draft into a task in its own project, and says which
// task it became.
//
// The CLI does not print the new id - only "Promoted draft DRAFT-1" - so it is
// found by seeing which task id was not there a moment ago. That is exact:
// promotion is the only thing happening between the two readings, and both are
// taken under the same store.
func (s *BoardService) PromoteDraft(projectPath, draftID string) CreateResult {
	before := s.taskIDs(projectPath)

	result := s.write(projectPath, fmt.Sprintf("%s could not be promoted", draftID),
		func(cli *backlogcli.Runner) error {
			return cli.PromoteDraft(context.Background(), s.dataDirFor(projectPath), draftID)
		})
	result = s.confirmDraftLeft(result, projectPath, draftID, "promoted")
	if !result.OK {
		return CreateResult{Problem: result.Problem}
	}

	out := CreateResult{OK: true}
	for id := range s.taskIDs(projectPath) {
		if !before[id] {
			out.TaskID = id
			break
		}
	}
	return out
}

// taskIDs is the set of live task ids in one project.
func (s *BoardService) taskIDs(projectPath string) map[string]bool {
	ids := map[string]bool{}
	for _, item := range s.store.Entities(backlog.KindTask) {
		if item.Ref.Project == projectPath && item.Ref.Class == backlog.ClassActive {
			ids[item.Ref.ID] = true
		}
	}
	return ids
}

// DiscardDraft archives a draft.
//
// Archiving rather than deleting: Backlog.md has no delete, and a note
// captured in a hurry is exactly the kind of thing someone wants back.
func (s *BoardService) DiscardDraft(projectPath, draftID string) WriteResult {
	result := s.write(projectPath, fmt.Sprintf("%s could not be discarded", draftID),
		func(cli *backlogcli.Runner) error {
			return cli.ArchiveDraft(context.Background(), s.dataDirFor(projectPath), draftID)
		})
	return s.confirmDraftLeft(result, projectPath, draftID, "discarded")
}

// confirmDraftLeft checks that a draft actually left the inbox.
//
// Not defensiveness for its own sake: on 1.48.0 both `draft promote` and
// `draft archive` exit 0 when the id does not resolve, so a write that did
// nothing at all arrives here as a success. 1.50.1 exits 1, but the check is
// cheap and the store has already been re-read by this point, so it costs a
// map lookup and holds on every version.
func (s *BoardService) confirmDraftLeft(
	result WriteResult, projectPath, draftID, what string,
) WriteResult {
	if !result.OK || !s.draftPresent(projectPath, draftID) {
		return result
	}
	return WriteResult{Problem: &Problem{
		Kind:  ProblemCLI,
		Title: fmt.Sprintf("%s was not %s", draftID, what),
		Detail: "Backlog.md reported no error and the note is still in the " +
			"inbox. On CLI versions before " + backlogcli.RecommendedVersion +
			" that is what an id it could not find looks like.",
		Path: projectPath,
	}}
}

// draftPresent reports whether a draft is still waiting in a project.
func (s *BoardService) draftPresent(projectPath, draftID string) bool {
	for _, item := range s.store.Entities(backlog.KindDraft) {
		if item.Ref.Class != backlog.ClassActive {
			continue
		}
		if item.Ref.Project == projectPath &&
			strings.EqualFold(item.Ref.ID, draftID) {
			return true
		}
	}
	return false
}

// DraftEdit is everything a draft can carry.
//
// The whole field surface `task create --draft` accepts, not the four that
// `draft create` does: a note that cannot be given a priority or a milestone
// until it stops being a note is a note nobody triages.
type DraftEdit struct {
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Labels      []string `json:"labels"`
	Assignee    string   `json:"assignee"`
	Priority    string   `json:"priority"`
	Type        string   `json:"type"`
	Milestone   string   `json:"milestone"`
	// AcceptanceCriteria are added in order, as items.
	AcceptanceCriteria []string `json:"acceptanceCriteria"`
	// Project is where the draft should end up. Empty means where it is.
	Project string `json:"project"`
}

// asNewTask maps an edit onto what the CLI takes.
func (e DraftEdit) asNewTask() backlogcli.NewTask {
	return backlogcli.NewTask{
		Title:              strings.TrimSpace(e.Title),
		Description:        e.Description,
		Priority:           strings.TrimSpace(e.Priority),
		Type:               strings.TrimSpace(e.Type),
		Milestone:          strings.TrimSpace(e.Milestone),
		Assignee:           strings.TrimSpace(e.Assignee),
		Labels:             e.Labels,
		AcceptanceCriteria: e.AcceptanceCriteria,
	}
}

// CaptureNote writes a new draft into a project.
//
// The other half of a usable inbox: until now a note could only be triaged,
// never made, so the inbox could only ever be as full as something else had
// filled it.
func (s *BoardService) CaptureNote(projectPath string, edit DraftEdit) WriteResult {
	if strings.TrimSpace(edit.Title) == "" {
		return WriteResult{Problem: &Problem{
			Kind: ProblemCLI, Title: "A note needs a title",
			Detail: "Backlog.md captures a note by its title; it cannot be empty.",
			Path:   projectPath,
		}}
	}
	return s.write(projectPath, "The note could not be captured",
		func(cli *backlogcli.Runner) error {
			_, err := cli.CreateDraft(context.Background(),
				s.dataDirFor(projectPath), edit.asNewTask())
			return err
		})
}

// ReviseDraft rewrites a draft, in its own project or another one.
//
// Backlog.md has no `draft edit`: `task edit` refuses a DRAFT id outright.
// So revising is capture-and-discard - the new note is created through the
// CLI and the old one archived - which is also the only way to move a draft
// between projects, since ids and files belong to one project.
//
// The cost is that the new note is captured now, so its wait starts again.
// That is stated in the interface rather than hidden, because the wait is the
// number the inbox is for.
//
// Order matters: create first, archive second. If the create fails there is
// still a draft; if the archive fails there are two, which is visible and
// fixable. The reverse could lose the note entirely.
func (s *BoardService) ReviseDraft(projectPath, draftID string, edit DraftEdit) WriteResult {
	target := strings.TrimSpace(edit.Project)
	if target == "" {
		target = projectPath
	}
	if strings.TrimSpace(edit.Title) == "" {
		return WriteResult{Problem: &Problem{
			Kind: ProblemCLI, Title: "A draft needs a title",
			Detail: "Backlog.md captures a note by its title; it cannot be empty.",
			Path:   projectPath,
		}}
	}
	if !s.known(target) {
		return WriteResult{Problem: &Problem{
			Kind: ProblemRegistry, Title: "No such project",
			Detail: "That project is not registered.", Path: target,
		}}
	}

	created := s.write(target, fmt.Sprintf("%s could not be rewritten", draftID),
		func(cli *backlogcli.Runner) error {
			_, err := cli.CreateDraft(context.Background(),
				s.dataDirFor(target), edit.asNewTask())
			return err
		})
	if !created.OK {
		return created
	}

	discarded := s.DiscardDraft(projectPath, draftID)
	if !discarded.OK {
		// The new note exists; say plainly that the old one is still there
		// rather than reporting a clean failure that would send someone
		// looking for a note that was in fact written.
		detail := "The rewritten note was captured, but the original could not " +
			"be discarded, so both are in the inbox."
		if discarded.Problem != nil {
			detail += " " + discarded.Problem.Detail
		}
		return WriteResult{Problem: &Problem{
			Kind: ProblemCLI, Title: "Both copies are still there",
			Detail: detail, Path: projectPath,
		}}
	}
	return WriteResult{OK: true}
}

// known reports whether a path is a registered, loadable project.
func (s *BoardService) known(path string) bool {
	for _, p := range s.store.Projects() {
		if p.Registry.Path == path && p.OK() {
			return true
		}
	}
	return false
}
