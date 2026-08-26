package backlogcli

import (
	"context"
	"errors"
	"fmt"
	"strings"
)

// The write surface Muster uses.
//
// Each of these maps onto exactly one CLI command. Nothing here builds a task
// file, renumbers an ordinal or invents an id: those belong to Backlog.md, and
// the whole point of this package is that they stay there.

// SetStatus moves a task to another status.
func (r *Runner) SetStatus(ctx context.Context, dir, taskID, status string) error {
	return r.edit(ctx, dir, taskID, "-s", status)
}

// SetPriority sets a task's priority.
func (r *Runner) SetPriority(ctx context.Context, dir, taskID, priority string) error {
	return r.edit(ctx, dir, taskID, "--priority", priority)
}

// SetAssignee replaces a task's assignee.
func (r *Runner) SetAssignee(ctx context.Context, dir, taskID, assignee string) error {
	return r.edit(ctx, dir, taskID, "-a", assignee)
}

// SetMilestone assigns a task to a milestone.
func (r *Runner) SetMilestone(ctx context.Context, dir, taskID, milestone string) error {
	if strings.TrimSpace(milestone) == "" {
		return r.edit(ctx, dir, taskID, "--clear-milestone")
	}
	return r.edit(ctx, dir, taskID, "-m", milestone)
}

// AddLabel adds a label without touching the others.
func (r *Runner) AddLabel(ctx context.Context, dir, taskID, label string) error {
	return r.edit(ctx, dir, taskID, "--add-label", label)
}

// RemoveLabel removes one label without touching the others.
func (r *Runner) RemoveLabel(ctx context.Context, dir, taskID, label string) error {
	return r.edit(ctx, dir, taskID, "--remove-label", label)
}

// CheckAcceptanceCriterion ticks a criterion by its 1-based index.
//
// That index is the CLI's own handle and is renumbered whenever criteria are
// inserted or removed, so it is only valid for the revision of the file the
// caller just read.
func (r *Runner) CheckAcceptanceCriterion(ctx context.Context, dir, taskID string, index int) error {
	return r.edit(ctx, dir, taskID, "--check-ac", fmt.Sprint(index))
}

// UncheckAcceptanceCriterion unticks a criterion by its 1-based index.
func (r *Runner) UncheckAcceptanceCriterion(ctx context.Context, dir, taskID string, index int) error {
	return r.edit(ctx, dir, taskID, "--uncheck-ac", fmt.Sprint(index))
}

// CreateDraft captures text into a project's drafts.
//
// Drafts stay off the board by design, which is what makes capture cheap.
func (r *Runner) CreateDraft(ctx context.Context, dir, title, description string) error {
	args := []string{"draft", "create", title}
	if strings.TrimSpace(description) != "" {
		args = append(args, "-d", description)
	}
	_, err := r.Exec(ctx, dir, args...)
	return err
}

// PromoteDraft turns a draft into a task.
func (r *Runner) PromoteDraft(ctx context.Context, dir, draftID string) error {
	_, err := r.Exec(ctx, dir, "draft", "promote", draftID)
	return err
}

// InitOptions configures `backlog init`.
//
// Every prompt that command asks has a corresponding flag, which is what lets
// the interface put a form over it rather than emulate a dialogue.
type InitOptions struct {
	// Name defaults to the folder name when empty.
	Name string
	// BacklogDir is "backlog", ".backlog" or a project-relative path.
	BacklogDir string
	// ConfigLocation is "folder" or "root".
	ConfigLocation string
	// TaskPrefix defaults to "task".
	TaskPrefix string
	// ZeroPaddedIDs is the number of digits, 0 to disable.
	ZeroPaddedIDs int
	// NoGit initialises without git integration, for folders that are not
	// repositories.
	NoGit bool
	// AgentInstructions is the comma-separated list, or "none".
	AgentInstructions string
	// IntegrationMode is "cli", "mcp" or "none".
	IntegrationMode string
}

// ErrConflictingInitOptions reports a combination the CLI refuses.
var ErrConflictingInitOptions = errors.New("conflicting init options")

// Init runs `backlog init` in a folder.
//
// The combination checked below is a real constraint of the CLI rather than a
// preference: 1.48.0 rejects agent instructions when AI integration is turned
// off. Catching it here means the UI can prevent the combination instead of
// showing the user a failure after the fact.
func (r *Runner) Init(ctx context.Context, dir string, opts InitOptions) error {
	if opts.IntegrationMode == "none" && opts.AgentInstructions != "" {
		return fmt.Errorf(
			"%w: agent instructions cannot be chosen when AI integration is none",
			ErrConflictingInitOptions)
	}

	args := []string{"init"}
	if strings.TrimSpace(opts.Name) != "" {
		args = append(args, opts.Name)
	}
	// --defaults answers every prompt this call does not set explicitly, which
	// is what keeps init non-interactive.
	args = append(args, "--defaults")

	if opts.BacklogDir != "" {
		args = append(args, "--backlog-dir", opts.BacklogDir)
	}
	if opts.ConfigLocation != "" {
		args = append(args, "--config-location", opts.ConfigLocation)
	}
	if opts.TaskPrefix != "" {
		args = append(args, "--task-prefix", opts.TaskPrefix)
	}
	if opts.ZeroPaddedIDs > 0 {
		args = append(args, "--zero-padded-ids", fmt.Sprint(opts.ZeroPaddedIDs))
	}
	if opts.NoGit {
		args = append(args, "--no-git")
	}
	if opts.AgentInstructions != "" {
		args = append(args, "--agent-instructions", opts.AgentInstructions)
	}
	if opts.IntegrationMode != "" {
		args = append(args, "--integration-mode", opts.IntegrationMode)
	}

	_, err := r.Exec(ctx, dir, args...)
	return err
}

// edit is the shared shape of every task mutation.
func (r *Runner) edit(ctx context.Context, dir, taskID string, args ...string) error {
	if strings.TrimSpace(taskID) == "" {
		return fmt.Errorf("no task id given")
	}
	full := append([]string{"task", "edit", taskID}, args...)
	_, err := r.Exec(ctx, dir, full...)
	return err
}
