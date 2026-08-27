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

// CreateDraft captures a note into a project's drafts.
//
// Drafts stay off the board by design, which is what makes capture cheap.
//
// It goes through `task create --draft` rather than `draft create`, because
// the two commands do not take the same fields: `draft create` knows title,
// body, labels and assignee, while `task create --draft` also takes priority,
// type, milestone, acceptance criteria, a parent and dependencies, and writes
// the same file. Using one path means a draft can hold everything the format
// allows it to.
func (r *Runner) CreateDraft(ctx context.Context, dir string, draft NewTask) (string, error) {
	draft.Draft = true
	return r.CreateTask(ctx, dir, draft)
}

// PromoteDraft turns a draft into a task: the file moves into tasks/, the id
// changes from DRAFT-n to the next task id, and the status becomes the
// project's first. The capture date is kept.
func (r *Runner) PromoteDraft(ctx context.Context, dir, draftID string) error {
	_, err := r.Exec(ctx, dir, "draft", "promote", draftID)
	return err
}

// AddMilestone creates a milestone and returns its id.
func (r *Runner) AddMilestone(ctx context.Context, dir, name, description string) (string, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return "", fmt.Errorf("a milestone needs a name")
	}
	args := []string{"milestone", "add", name}
	args = appendIf(args, "-d", description)

	out, err := r.Exec(ctx, dir, args...)
	if err != nil {
		return "", err
	}
	return milestoneID.FindString(out), nil
}

var milestoneID = regexp.MustCompile(`(?i)\bm-[0-9]+`)

// RenameMilestone renames a milestone, keeping its id.
//
// The tasks that point at it are updated by default, and that is left on: a
// milestone renamed while its tasks still name the old one is a rename that
// only half happened.
func (r *Runner) RenameMilestone(ctx context.Context, dir, from, to string) error {
	_, err := r.Exec(ctx, dir, "milestone", "rename", from, to)
	return err
}

// ArchiveMilestone moves a milestone into archive/milestones and leaves the
// tasks pointing at it exactly as they are.
func (r *Runner) ArchiveMilestone(ctx context.Context, dir, name string) error {
	_, err := r.Exec(ctx, dir, "milestone", "archive", name)
	return err
}

// TaskHandling is what becomes of the tasks when a milestone is removed.
type TaskHandling string

// The three the CLI offers.
const (
	// HandlingClear empties the milestone field on every task that named it.
	HandlingClear TaskHandling = "clear"
	// HandlingKeep leaves the tasks pointing at a milestone that is no longer
	// in the active list.
	HandlingKeep TaskHandling = "keep"
	// HandlingReassign moves them to another milestone.
	HandlingReassign TaskHandling = "reassign"
)

// RemoveMilestone removes a milestone and decides what happens to its tasks.
//
// It archives the file exactly as ArchiveMilestone does - measured, not
// assumed: both end up in archive/milestones. The difference between them is
// entirely in what they do to the tasks, which is why the choice has to be
// made explicitly rather than hidden behind a word.
func (r *Runner) RemoveMilestone(
	ctx context.Context, dir, name string, handling TaskHandling, reassignTo string,
) error {
	if handling == HandlingReassign && strings.TrimSpace(reassignTo) == "" {
		return fmt.Errorf("reassigning needs a milestone to reassign to")
	}
	args := []string{"milestone", "remove", name, "--task-handling", string(handling)}
	if handling == HandlingReassign {
		args = append(args, "--reassign-to", reassignTo)
	}
	_, err := r.Exec(ctx, dir, args...)
	return err
}

// NewDocument is what `backlog doc create` accepts.
//
// No content: the command does not take any. A document is created empty and
// filled by a second call, which is why CreateDocument returns the id.
type NewDocument struct {
	Title string
	// Type is readme, guide, specification or other. Empty lets the CLI
	// choose its own default rather than Muster inventing one.
	Type string
	// Path is a docs-relative sub-path. Absolute paths and .. are rejected by
	// the CLI itself.
	Path string
}

// CreateDocument creates an empty document and returns its id.
func (r *Runner) CreateDocument(ctx context.Context, dir string, doc NewDocument) (string, error) {
	title := strings.TrimSpace(doc.Title)
	if title == "" {
		return "", fmt.Errorf("a document needs a title")
	}
	args := []string{"doc", "create", title, "--plain"}
	args = appendIf(args, "-t", doc.Type)
	args = appendIf(args, "-p", doc.Path)

	out, err := r.Exec(ctx, dir, args...)
	if err != nil {
		return "", err
	}
	return documentIDFrom(out), nil
}

// documentID finds a doc id in the CLI's own output.
var documentID = regexp.MustCompile(`(?i)\bdoc-[0-9]+`)

func documentIDFrom(out string) string {
	return documentID.FindString(out)
}

// DocumentEdit is what `backlog doc update` changes. Empty fields are left
// alone; Content replaces the whole body, which is the only form the command
// takes.
type DocumentEdit struct {
	Title   string
	Type    string
	Tags    []string
	Content string
	// SetContent distinguishes "leave the body alone" from "make it empty".
	SetContent bool
}

// UpdateDocument changes a document.
func (r *Runner) UpdateDocument(ctx context.Context, dir, docID string, edit DocumentEdit) error {
	if strings.TrimSpace(docID) == "" {
		return fmt.Errorf("no document id given")
	}
	args := []string{"doc", "update", docID}
	args = appendIf(args, "--title", edit.Title)
	args = appendIf(args, "-t", edit.Type)
	if tags := joinValues(edit.Tags); tags != "" {
		args = append(args, "--tags", tags)
	}
	if edit.SetContent {
		args = append(args, "--content", edit.Content)
	}
	if len(args) == 3 {
		return fmt.Errorf("nothing to change")
	}
	_, err := r.Exec(ctx, dir, args...)
	return err
}

// CreateDecision creates a decision and returns its id.
//
// The status is free-form and defaults to proposed. The body cannot be set:
// the CLI writes a skeleton with Context, Decision and Consequences headings
// and offers no command to fill them - there is no `decision update`.
func (r *Runner) CreateDecision(ctx context.Context, dir, title, status string) (string, error) {
	title = strings.TrimSpace(title)
	if title == "" {
		return "", fmt.Errorf("a decision needs a title")
	}
	args := []string{"decision", "create", title, "--plain"}
	args = appendIf(args, "-s", status)

	out, err := r.Exec(ctx, dir, args...)
	if err != nil {
		return "", err
	}
	return decisionID.FindString(out), nil
}

var decisionID = regexp.MustCompile(`(?i)\bdecision-[0-9]+`)

// joinValues renders a list the way the CLI takes one.
func joinValues(values []string) string { return strings.Join(trimmed(values), ",") }

// AddComment appends a comment to a task.
//
// The author is optional and omitted when empty, which is what the CLI does
// with it: the file gets no author line at all rather than a placeholder. An
// unsigned comment is a real state, not a failure to record one.
func (r *Runner) AddComment(ctx context.Context, dir, taskID, text, author string) error {
	if strings.TrimSpace(text) == "" {
		return fmt.Errorf("a comment needs something in it")
	}
	args := []string{"--comment", text}
	if who := strings.TrimSpace(author); who != "" {
		args = append(args, "--comment-author", who)
	}
	return r.edit(ctx, dir, taskID, args...)
}

// SetDependencies replaces a task's dependency list.
//
// The CLI checks that each id exists in the project and refuses the whole
// edit naming the ones it could not find, so nothing partial is written.
func (r *Runner) SetDependencies(ctx context.Context, dir, taskID string, ids []string) error {
	if len(trimmed(ids)) == 0 {
		return r.edit(ctx, dir, taskID, "--clear-deps")
	}
	return r.edit(ctx, dir, taskID, "--dep", strings.Join(trimmed(ids), ","))
}

// SetReferences replaces a task's references. Empty clears them.
func (r *Runner) SetReferences(ctx context.Context, dir, taskID string, refs []string) error {
	if len(trimmed(refs)) == 0 {
		return r.edit(ctx, dir, taskID, "--clear-refs")
	}
	args := []string{}
	for _, ref := range trimmed(refs) {
		args = append(args, "--ref", ref)
	}
	return r.edit(ctx, dir, taskID, args...)
}

// SetDocumentation replaces a task's documentation links. Empty clears them.
func (r *Runner) SetDocumentation(ctx context.Context, dir, taskID string, docs []string) error {
	if len(trimmed(docs)) == 0 {
		return r.edit(ctx, dir, taskID, "--clear-docs")
	}
	args := []string{}
	for _, doc := range trimmed(docs) {
		args = append(args, "--doc", doc)
	}
	return r.edit(ctx, dir, taskID, args...)
}

// ErrCannotClear reports a field the CLI can set but not empty.
var ErrCannotClear = errors.New("this field cannot be cleared")

// SetModifiedFiles replaces a task's modified-file list.
//
// It cannot empty one. Measured on 1.50.1: there is no --clear-modified-files,
// and --modified-file "" exits 0 having changed nothing - a silent no-op, the
// worst of the three possible answers. Refusing here is what keeps the
// interface from offering a control that does nothing.
func (r *Runner) SetModifiedFiles(ctx context.Context, dir, taskID string, files []string) error {
	kept := trimmed(files)
	if len(kept) == 0 {
		return fmt.Errorf("%w: modified_files", ErrCannotClear)
	}
	args := []string{}
	for _, file := range kept {
		args = append(args, "--modified-file", file)
	}
	return r.edit(ctx, dir, taskID, args...)
}

// SetOrdinal sets a task's manual ordering value.
func (r *Runner) SetOrdinal(ctx context.Context, dir, taskID string, ordinal int) error {
	return r.edit(ctx, dir, taskID, "--ordinal", fmt.Sprint(ordinal))
}

// NonEmpty drops blanks, so an interface that renders an empty row does not
// write one. Exported because a caller has to know whether a list is empty
// before deciding whether the CLI can express what it wants.
func NonEmpty(values []string) []string { return trimmed(values) }

// trimmed drops blanks, so an interface that renders an empty row does not
// write one.
func trimmed(values []string) []string {
	var kept []string
	for _, value := range values {
		if v := strings.TrimSpace(value); v != "" {
			kept = append(kept, v)
		}
	}
	return kept
}

// CompleteTask moves a finished task into completed/.
//
// The CLI refuses anything that is not at its project's own last declared
// status - not the literal word "Done", which was worth checking: a project
// whose statuses end in "Shipped" completes a Shipped task and refuses the
// others. So the interface can offer this exactly when it will work.
func (r *Runner) CompleteTask(ctx context.Context, dir, taskID string) error {
	_, err := r.Exec(ctx, dir, "task", "complete", taskID)
	return err
}

// ArchiveTask moves a task into archive/tasks.
//
// Archiving is Backlog.md's soft delete: the file stays, the board stops
// showing it, and the id becomes free for reuse - which is why archived ids
// collide with live ones and identity is never the id alone.
func (r *Runner) ArchiveTask(ctx context.Context, dir, taskID string) error {
	_, err := r.Exec(ctx, dir, "task", "archive", taskID)
	return err
}

// DemoteTask sends a task back to drafts.
//
// The only way back: `task edit -s Draft` is refused, because Draft is not one
// of a project's configured statuses. This is a separate command precisely
// because it moves the file.
func (r *Runner) DemoteTask(ctx context.Context, dir, taskID string) error {
	_, err := r.Exec(ctx, dir, "task", "demote", taskID)
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
	// Draft writes the file into drafts/ instead of tasks/, with a DRAFT id
	// and status Draft.
	//
	// `task create --draft` takes the whole field surface below, which
	// `draft create` does not: that command knows only title, body, labels
	// and assignee. Everything a draft is allowed to carry is therefore
	// created through here.
	Draft bool
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
	if task.Draft {
		args = append(args, "--draft")
	}
	args = appendIf(args, "-d", task.Description)
	// A draft's status is Draft and the CLI rejects any other, so it is not
	// passed for one at all.
	if !task.Draft {
		args = appendIf(args, "-s", task.Status)
	}
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
