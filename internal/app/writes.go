package app

import (
	"context"
	"errors"
	"fmt"

	"github.com/FMakareev/muster-backlog/internal/backlogcli"
)

// ProblemCLI reports that the backlog CLI is missing or unusable. It is
// recorded once at startup rather than raised again on every action.
const ProblemCLI ProblemKind = "cli"

// WriteResult is what the frontend gets back from a write.
//
// A write is not "did the call return": it is what the files say afterwards.
// So this reports whether the CLI succeeded and carries the problem when it
// did not, and the caller re-reads rather than assuming.
type WriteResult struct {
	OK      bool     `json:"ok"`
	Problem *Problem `json:"problem"`
}

func writeOK() WriteResult { return WriteResult{OK: true} }

func writeFailed(title string, err error, path string) WriteResult {
	return WriteResult{OK: false, Problem: &Problem{
		Kind:   ProblemCLI,
		Title:  title,
		Detail: err.Error(),
		Path:   path,
	}}
}

// cliProblem returns the standing reason writes are unavailable, or nil.
func (s *BoardService) cliProblem() *Problem {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.cli != nil {
		return nil
	}
	detail := "Install it and make sure it is on PATH."
	if s.cliErr != nil {
		detail = s.cliErr.Error()
	}
	return &Problem{
		Kind:   ProblemCLI,
		Title:  "Changes cannot be saved",
		Detail: detail,
	}
}

// CLIVersion reports the CLI version in use, empty when there is none.
func (s *BoardService) CLIVersion() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.cli == nil {
		return ""
	}
	return s.cli.Version()
}

// SetStatus moves a task to another status.
func (s *BoardService) SetStatus(projectPath, taskID, status string) WriteResult {
	return s.editTask(projectPath, taskID, fmt.Sprintf("%s could not be moved", taskID),
		func(cli *backlogcli.Runner) error {
			return cli.SetStatus(context.Background(), s.dataDirFor(projectPath), taskID, status)
		})
}

// SetPriority sets a task's priority.
func (s *BoardService) SetPriority(projectPath, taskID, priority string) WriteResult {
	return s.editTask(projectPath, taskID, fmt.Sprintf("%s could not be reprioritised", taskID),
		func(cli *backlogcli.Runner) error {
			return cli.SetPriority(context.Background(), s.dataDirFor(projectPath), taskID, priority)
		})
}

// SetAssignee replaces a task's assignee.
func (s *BoardService) SetAssignee(projectPath, taskID, assignee string) WriteResult {
	return s.editTask(projectPath, taskID, fmt.Sprintf("%s could not be reassigned", taskID),
		func(cli *backlogcli.Runner) error {
			return cli.SetAssignee(context.Background(), s.dataDirFor(projectPath), taskID, assignee)
		})
}

// CheckCriterion ticks or unticks an acceptance criterion by its 1-based index.
func (s *BoardService) CheckCriterion(projectPath, taskID string, index int, checked bool) WriteResult {
	return s.editTask(projectPath, taskID, fmt.Sprintf("%s could not be updated", taskID),
		func(cli *backlogcli.Runner) error {
			dir := s.dataDirFor(projectPath)
			if checked {
				return cli.CheckAcceptanceCriterion(context.Background(), dir, taskID, index)
			}
			return cli.UncheckAcceptanceCriterion(context.Background(), dir, taskID, index)
		})
}

// CaptureDraft writes text into a project's drafts.
func (s *BoardService) CaptureDraft(projectPath, title, description string) WriteResult {
	return s.write(projectPath, "The note could not be captured",
		func(cli *backlogcli.Runner) error {
			_, err := cli.CreateDraft(context.Background(), s.dataDirFor(projectPath),
				backlogcli.NewTask{Title: title, Description: description})
			return err
		})
}

// editTask is the shared shape of every mutation that runs `task edit`.
//
// It refuses a draft before running anything. Backlog.md has no way to edit a
// draft - `task edit` answers "Task DRAFT-4 not found" - so every one of these
// would fail against one, and the CLI's message names a task that plainly
// exists, which reads like a bug in Muster rather than a limit of the format.
// The interface does not offer these on a note either; this is the floor
// under that, because a note opened in the panel used to offer all of them.
func (s *BoardService) editTask(
	projectPath, taskID, failureTitle string, do func(*backlogcli.Runner) error,
) WriteResult {
	if s.draftPresent(projectPath, taskID) {
		return WriteResult{Problem: &Problem{
			Kind:  ProblemCLI,
			Title: fmt.Sprintf("%s is a note, not a task", taskID),
			Detail: "Backlog.md cannot edit a draft. Promote it first, or " +
				"rewrite it in the inbox, which captures a new note in its place.",
			Path: projectPath,
		}}
	}
	return s.write(projectPath, failureTitle, do)
}

// write runs one mutation and then re-reads the project.
//
// The rescan is the point. Nothing here assumes the CLI did what was asked:
// the store is refreshed from the files, and the interface settles on that.
// It is also what makes a failed write visibly harmless — the board returns to
// what the files actually say.
func (s *BoardService) write(projectPath, failureTitle string, do func(*backlogcli.Runner) error) WriteResult {
	s.mu.Lock()
	cli := s.cli
	s.mu.Unlock()

	if cli == nil {
		if p := s.cliProblem(); p != nil {
			return WriteResult{OK: false, Problem: p}
		}
	}

	err := do(cli)

	// Re-read whether the write succeeded or not: a partial write is exactly
	// the case where an assumption would be wrong.
	if s.store.Reload(projectPath) {
		emit(EventProjectChanged, ProjectChanged{Project: projectPath})
	}

	if err != nil {
		return writeFailed(failureTitle, err, projectPath)
	}
	return writeOK()
}

// dataDirFor returns the working directory a command should run in.
//
// Commands run at the project root rather than in the data directory, which is
// how the CLI discovers its own configuration.
func (s *BoardService) dataDirFor(projectPath string) string {
	for _, p := range s.store.Projects() {
		if p.Registry.Path == projectPath && p.OK() {
			return p.Registry.Location.Root
		}
	}
	return projectPath
}

// resolveCLI locates the CLI once, recording why if it cannot.
func (s *BoardService) resolveCLI() {
	cli, err := backlogcli.New()

	s.mu.Lock()
	s.cli, s.cliErr = cli, err
	s.mu.Unlock()

	if err != nil {
		title := "Changes cannot be saved"
		if errors.Is(err, backlogcli.ErrUnsupportedVersion) {
			title = "The Backlog.md CLI is too old"
		}
		s.mu.Lock()
		s.standingProblems = append(s.standingProblems, Problem{
			Kind:   ProblemCLI,
			Title:  title,
			Detail: err.Error(),
		})
		s.mu.Unlock()
	}
}

// AddLabel adds a label to a task without touching the others.
func (s *BoardService) AddLabel(projectPath, taskID, label string) WriteResult {
	return s.editTask(projectPath, taskID, fmt.Sprintf("%s could not be labelled", taskID),
		func(cli *backlogcli.Runner) error {
			return cli.AddLabel(context.Background(), s.dataDirFor(projectPath), taskID, label)
		})
}

// RemoveLabel removes one label without touching the others.
func (s *BoardService) RemoveLabel(projectPath, taskID, label string) WriteResult {
	return s.editTask(projectPath, taskID, fmt.Sprintf("%s could not be relabelled", taskID),
		func(cli *backlogcli.Runner) error {
			return cli.RemoveLabel(context.Background(), s.dataDirFor(projectPath), taskID, label)
		})
}

// NewTaskInput is the create form as the frontend sends it.
type NewTaskInput struct {
	Project            string   `json:"project"`
	Title              string   `json:"title"`
	Description        string   `json:"description"`
	Status             string   `json:"status"`
	Priority           string   `json:"priority"`
	Type               string   `json:"type"`
	Milestone          string   `json:"milestone"`
	Assignee           string   `json:"assignee"`
	Labels             []string `json:"labels"`
	AcceptanceCriteria []string `json:"acceptanceCriteria"`
}

// CreateResult reports what happened, including the id the CLI assigned so the
// interface can open the task it just made.
type CreateResult struct {
	OK      bool     `json:"ok"`
	TaskID  string   `json:"taskId"`
	Problem *Problem `json:"problem"`
}

// CreateTask writes a new task into a project.
func (s *BoardService) CreateTask(input NewTaskInput) CreateResult {
	s.mu.Lock()
	cli := s.cli
	s.mu.Unlock()

	if cli == nil {
		if p := s.cliProblem(); p != nil {
			return CreateResult{Problem: p}
		}
	}

	id, err := cli.CreateTask(context.Background(), s.dataDirFor(input.Project),
		backlogcli.NewTask{
			Title:              input.Title,
			Description:        input.Description,
			Status:             input.Status,
			Priority:           input.Priority,
			Type:               input.Type,
			Milestone:          input.Milestone,
			Assignee:           input.Assignee,
			Labels:             input.Labels,
			AcceptanceCriteria: input.AcceptanceCriteria,
		})

	if s.store.Reload(input.Project) {
		emit(EventProjectChanged, ProjectChanged{Project: input.Project})
	}
	if err != nil {
		return CreateResult{Problem: &Problem{
			Kind:   ProblemCLI,
			Title:  "The task could not be created",
			Detail: err.Error(),
			Path:   input.Project,
		}}
	}
	return CreateResult{OK: true, TaskID: id}
}

// SetMilestone moves a task to a milestone, or clears it when empty.
//
// A milestone is the axis a backlog is planned on, so moving a task between
// them is ordinary work rather than something to be decided once at creation.
func (s *BoardService) SetMilestone(projectPath, taskID, milestone string) WriteResult {
	return s.editTask(projectPath, taskID, fmt.Sprintf("%s could not be moved", taskID),
		func(cli *backlogcli.Runner) error {
			return cli.SetMilestone(context.Background(),
				s.dataDirFor(projectPath), taskID, milestone)
		})
}

// SetTitle renames a task.
func (s *BoardService) SetTitle(projectPath, taskID, title string) WriteResult {
	return s.editTask(projectPath, taskID, fmt.Sprintf("%s could not be renamed", taskID),
		func(cli *backlogcli.Runner) error {
			return cli.SetTitle(context.Background(), s.dataDirFor(projectPath), taskID, title)
		})
}

// SetSection replaces one of a task's body sections.
//
// The section names match what the parser exposes, so the frontend never has
// to know which CLI flag writes which part of the file.
func (s *BoardService) SetSection(projectPath, taskID, section, text string) WriteResult {
	return s.editTask(projectPath, taskID, fmt.Sprintf("%s could not be saved", taskID),
		func(cli *backlogcli.Runner) error {
			dir := s.dataDirFor(projectPath)
			switch section {
			case "description":
				return cli.SetDescription(context.Background(), dir, taskID, text)
			case "plan":
				return cli.SetPlan(context.Background(), dir, taskID, text)
			case "notes":
				return cli.SetNotes(context.Background(), dir, taskID, text)
			default:
				return fmt.Errorf("%q is not a section this application writes", section)
			}
		})
}

// SetAcceptanceCriteria replaces the whole list, which is what makes adding,
// removing and reordering a single operation.
func (s *BoardService) SetAcceptanceCriteria(projectPath, taskID string, items []string) WriteResult {
	return s.editTask(projectPath, taskID, fmt.Sprintf("%s could not be saved", taskID),
		func(cli *backlogcli.Runner) error {
			return cli.SetAcceptanceCriteria(
				context.Background(), s.dataDirFor(projectPath), taskID, items)
		})
}
