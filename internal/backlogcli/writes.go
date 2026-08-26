package backlogcli

import (
	"context"
	"errors"
	"fmt"
	"regexp"
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

// NewDraft is everything `backlog draft create` accepts.
//
// Deliberately short, because the command is: a draft holds a title, a body,
// labels and an assignee, and nothing else. Priority, type and milestone are
// not fields a draft has - they arrive when it becomes a task.
type NewDraft struct {
	Title       string
	Description string
	Labels      []string
	Assignee    string
}

// CreateDraft captures a note into a project's drafts.
//
// Drafts stay off the board by design, which is what makes capture cheap.
func (r *Runner) CreateDraft(ctx context.Context, dir string, draft NewDraft) error {
	title := strings.TrimSpace(draft.Title)
	if title == "" {
		return fmt.Errorf("a draft needs a title")
	}

	args := []string{"draft", "create", title}
	if body := strings.TrimSpace(draft.Description); body != "" {
		args = append(args, "-d", body)
	}
	if labels := joinLabels(draft.Labels); labels != "" {
		args = append(args, "-l", labels)
	}
	if who := strings.TrimSpace(draft.Assignee); who != "" {
		args = append(args, "-a", who)
	}
	_, err := r.Exec(ctx, dir, args...)
	return err
}

// joinLabels renders a label list the way the CLI takes one.
func joinLabels(labels []string) string {
	var kept []string
	for _, label := range labels {
		if trimmed := strings.TrimSpace(label); trimmed != "" {
			kept = append(kept, trimmed)
		}
	}
	return strings.Join(kept, ",")
}

// PromoteDraft turns a draft into a task: the file moves into tasks/, the id
// changes from DRAFT-n to the next task id, and the status becomes the
// project's first. The capture date is kept.
func (r *Runner) PromoteDraft(ctx context.Context, dir, draftID string) error {
	_, err := r.Exec(ctx, dir, "draft", "promote", draftID)
	return err
}

// ArchiveDraft moves a draft into archive/drafts.
//
// This is what discarding is: Backlog.md has no delete, and inventing one by
// removing the file would be Muster writing to a project behind the CLI's
// back. The note stays recoverable, which is the right default for something
// captured in a hurry.
func (r *Runner) ArchiveDraft(ctx context.Context, dir, draftID string) error {
	_, err := r.Exec(ctx, dir, "draft", "archive", draftID)
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

// NewTask is everything the create form can set.
//
// Only Title is required; the CLI applies each project's own defaults for the
// rest, which is why nothing here carries a default of its own.
type NewTask struct {
	Title       string
	Description string
	Status      string
	Priority    string
	Type        string
	Milestone   string
	Assignee    string
	Labels      []string
	// AcceptanceCriteria are added as individual items, in order.
	AcceptanceCriteria []string
	// Parent makes this a subtask of an existing task.
	Parent string
	// DependsOn are task ids in the same project.
	DependsOn []string
}

// CreateTask creates a task and returns the id the CLI assigned.
//
// The id comes back from the CLI rather than being predicted: generation
// handles collisions and zero-padding per project, and guessing it would be
// the first step towards owning the format.
func (r *Runner) CreateTask(ctx context.Context, dir string, task NewTask) (string, error) {
	if strings.TrimSpace(task.Title) == "" {
		return "", fmt.Errorf("a task needs a title")
	}

	args := []string{"task", "create", task.Title, "--plain"}
	args = appendIf(args, "-d", task.Description)
	args = appendIf(args, "-s", task.Status)
	args = appendIf(args, "--priority", task.Priority)
	args = appendIf(args, "--type", task.Type)
	args = appendIf(args, "-m", task.Milestone)
	args = appendIf(args, "-a", task.Assignee)
	args = appendIf(args, "-p", task.Parent)
	if len(task.Labels) > 0 {
		args = append(args, "-l", strings.Join(task.Labels, ","))
	}
	if len(task.DependsOn) > 0 {
		args = append(args, "--dep", strings.Join(task.DependsOn, ","))
	}
	for _, criterion := range task.AcceptanceCriteria {
		if strings.TrimSpace(criterion) != "" {
			args = append(args, "--ac", criterion)
		}
	}

	out, err := r.Exec(ctx, dir, args...)
	if err != nil {
		return "", err
	}
	return taskIDFrom(out), nil
}

// SetTitle renames a task.
func (r *Runner) SetTitle(ctx context.Context, dir, taskID, title string) error {
	if strings.TrimSpace(title) == "" {
		return fmt.Errorf("a task needs a title")
	}
	return r.edit(ctx, dir, taskID, "-t", title)
}

// SetDescription replaces a task's description.
func (r *Runner) SetDescription(ctx context.Context, dir, taskID, text string) error {
	return r.edit(ctx, dir, taskID, "-d", text)
}

// SetPlan replaces a task's implementation plan.
func (r *Runner) SetPlan(ctx context.Context, dir, taskID, text string) error {
	return r.edit(ctx, dir, taskID, "--plan", text)
}

// SetNotes replaces a task's implementation notes.
func (r *Runner) SetNotes(ctx context.Context, dir, taskID, text string) error {
	return r.edit(ctx, dir, taskID, "--notes", text)
}

// SetAcceptanceCriteria replaces the whole acceptance-criteria list.
//
// Replacing wholesale rather than patching item by item is what makes adding,
// removing and reordering one operation, and it keeps the per-item indexes -
// which the CLI renumbers on every insertion - out of the interface entirely.
func (r *Runner) SetAcceptanceCriteria(ctx context.Context, dir, taskID string, items []string) error {
	args := []string{"task", "edit", taskID}
	kept := 0
	for _, item := range items {
		if strings.TrimSpace(item) == "" {
			continue
		}
		args = append(args, "--acceptance-criteria", item)
		kept++
	}
	if kept == 0 {
		args = append(args, "--clear-ac")
	}
	_, err := r.Exec(ctx, dir, args...)
	return err
}

// appendIf adds a flag only when its value is set, so an untouched field in a
// form does not become an instruction to clear anything.
func appendIf(args []string, flag, value string) []string {
	if strings.TrimSpace(value) == "" {
		return args
	}
	return append(args, flag, value)
}

// idPattern matches a Backlog.md id. The prefix is per-project configuration,
// so it is matched as any word rather than assumed to be TASK.
var idPattern = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9]*-\d+(?:\.\d+)*$`)

// taskIDFrom picks the created id out of the CLI's plain output.
func taskIDFrom(out string) string {
	for _, field := range strings.Fields(out) {
		candidate := strings.Trim(field, ":,.")
		if idPattern.MatchString(candidate) {
			return strings.ToUpper(candidate)
		}
	}
	return ""
}
