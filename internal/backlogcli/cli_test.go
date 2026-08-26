package backlogcli_test

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/FMakareev/muster-backlog/internal/backlog"
	"github.com/FMakareev/muster-backlog/internal/backlogcli"
	"github.com/FMakareev/muster-backlog/internal/project"
)

// These tests run the real Backlog.md CLI against scratch projects rather than
// a mock. The entire premise of this package is that the CLI owns the format,
// so a mock would only test my own assumptions about it.
func runner(t *testing.T) *backlogcli.Runner {
	t.Helper()
	if _, err := exec.LookPath(backlogcli.BinaryName); err != nil {
		t.Skipf("the %s CLI is not installed", backlogcli.BinaryName)
	}
	r, err := backlogcli.New()
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	return r
}

// newProject initialises a real project through the CLI itself.
func newProject(t *testing.T, r *backlogcli.Runner) string {
	t.Helper()
	dir := t.TempDir()
	err := r.Init(context.Background(), dir, backlogcli.InitOptions{
		Name:            "Scratch",
		NoGit:           true,
		IntegrationMode: "none",
	})
	if err != nil {
		t.Fatalf("Init: %v", err)
	}
	return dir
}

// scan reads the project back from disk, which is the only way to know what
// the CLI actually did.
func scan(t *testing.T, dir string) *backlog.Project {
	t.Helper()
	loc, err := project.Discover(dir)
	if err != nil {
		t.Fatalf("Discover: %v", err)
	}
	p, err := backlog.Scan(loc)
	if err != nil {
		t.Fatalf("Scan: %v", err)
	}
	return p
}

func createTask(t *testing.T, r *backlogcli.Runner, dir, title string) string {
	t.Helper()
	out, err := r.Exec(context.Background(), dir, "task", "create", title, "--plain")
	if err != nil {
		t.Fatalf("task create: %v", err)
	}
	for _, field := range strings.Fields(out) {
		if strings.HasPrefix(strings.ToUpper(field), "TASK-") {
			return strings.TrimSuffix(strings.ToUpper(field), ":")
		}
	}
	t.Fatalf("no task id in CLI output: %q", out)
	return ""
}

func TestVersionIsCheckedAtStartup(t *testing.T) {
	r := runner(t)
	if r.Version() == "" {
		t.Error("no version recorded")
	}
	if r.Binary() == "" {
		t.Error("no binary path recorded")
	}
}

// A missing CLI must be one clear report, not the same failure arriving again
// on every action the user tries.
func TestMissingBinaryIsReportedOnce(t *testing.T) {
	_, err := backlogcli.NewAt(filepath.Join(t.TempDir(), "definitely-not-here"))
	if err == nil {
		t.Fatal("want an error for a missing binary")
	}
	if !errors.Is(err, backlogcli.ErrNotInstalled) {
		t.Errorf("err = %v, want ErrNotInstalled", err)
	}
}

// A binary that runs but is not the CLI must not be mistaken for one.
func TestNonBacklogBinaryIsRejected(t *testing.T) {
	fake := filepath.Join(t.TempDir(), "fake")
	script := "#!/bin/sh\necho 0.1.0\n"
	if err := os.WriteFile(fake, []byte(script), 0o755); err != nil {
		t.Fatalf("write fake: %v", err)
	}
	_, err := backlogcli.NewAt(fake)
	if !errors.Is(err, backlogcli.ErrUnsupportedVersion) {
		t.Fatalf("err = %v, want ErrUnsupportedVersion", err)
	}
	if !strings.Contains(err.Error(), backlogcli.MinimumVersion) {
		t.Errorf("err = %v, want it to name the version needed", err)
	}
}

func TestCommandsRunInTheTargetProject(t *testing.T) {
	r := runner(t)
	alpha := newProject(t, r)
	beta := newProject(t, r)

	createTask(t, r, alpha, "Only in alpha")

	if got := len(scan(t, alpha).Tasks); got != 1 {
		t.Errorf("alpha has %d tasks, want 1", got)
	}
	if got := len(scan(t, beta).Tasks); got != 0 {
		t.Errorf("beta has %d tasks, want none — the command ran in the wrong project", got)
	}
}

func TestSetStatusAndPriority(t *testing.T) {
	r := runner(t)
	dir := newProject(t, r)
	id := createTask(t, r, dir, "A task to move")

	if err := r.SetStatus(context.Background(), dir, id, "In Progress"); err != nil {
		t.Fatalf("SetStatus: %v", err)
	}
	if err := r.SetPriority(context.Background(), dir, id, "High"); err != nil {
		t.Fatalf("SetPriority: %v", err)
	}

	tasks := scan(t, dir).Tasks
	if len(tasks) != 1 {
		t.Fatalf("got %d tasks", len(tasks))
	}
	if tasks[0].Status != "In Progress" {
		t.Errorf("status = %q", tasks[0].Status)
	}
	// The CLI writes priority lowercase against a capitalised config, which is
	// exactly why nothing compares these case-sensitively.
	if !strings.EqualFold(tasks[0].Priority, "High") {
		t.Errorf("priority = %q", tasks[0].Priority)
	}
}

func TestLabelsAndMilestone(t *testing.T) {
	r := runner(t)
	dir := newProject(t, r)
	id := createTask(t, r, dir, "A task to label")
	ctx := context.Background()

	if err := r.AddLabel(ctx, dir, id, "backend"); err != nil {
		t.Fatalf("AddLabel: %v", err)
	}
	if err := r.AddLabel(ctx, dir, id, "urgent"); err != nil {
		t.Fatalf("AddLabel: %v", err)
	}
	if got := scan(t, dir).Tasks[0].Labels; len(got) != 2 {
		t.Errorf("labels = %v, want both kept", got)
	}

	if err := r.RemoveLabel(ctx, dir, id, "urgent"); err != nil {
		t.Fatalf("RemoveLabel: %v", err)
	}
	got := scan(t, dir).Tasks[0].Labels
	if len(got) != 1 || got[0] != "backend" {
		t.Errorf("labels = %v, want only backend — removing one must not touch the other", got)
	}
}

// Titles in real projects contain quotes, backticks, semicolons and dollar
// signs. Arguments go to the process directly, never through a shell, so a
// title like this is a title rather than a command.
func TestHostileTitlesAreNotInterpreted(t *testing.T) {
	r := runner(t)
	dir := newProject(t, r)

	canary := filepath.Join(t.TempDir(), "canary")
	titles := []string{
		"Fix $(touch " + canary + ") in the parser",
		"Handle `rm -rf /` safely",
		"Quote \"handling\" and 'apostrophes'; then continue",
		"Риг байка на 2D-физике: рама и колёса",
		"Semicolons; ampersands & pipes | all of it",
	}

	for _, title := range titles {
		createTask(t, r, dir, title)
	}

	if _, err := os.Stat(canary); err == nil {
		t.Fatal("a title was executed by a shell")
	}

	tasks := scan(t, dir).Tasks
	if len(tasks) != len(titles) {
		t.Fatalf("got %d tasks, want %d", len(tasks), len(titles))
	}
	stored := map[string]bool{}
	for _, task := range tasks {
		stored[task.Title] = true
	}
	for _, title := range titles {
		if !stored[title] {
			t.Errorf("title %q did not survive the round trip", title)
		}
	}
}

func TestFailureCarriesWhatTheCLISaid(t *testing.T) {
	r := runner(t)
	dir := newProject(t, r)

	err := r.SetStatus(context.Background(), dir, "TASK-404", "Done")
	if err == nil {
		t.Fatal("want an error for a task that does not exist")
	}

	var cmdErr *backlogcli.CommandError
	if !errors.As(err, &cmdErr) {
		t.Fatalf("err = %T, want a *CommandError", err)
	}
	if cmdErr.ExitCode == 0 {
		t.Error("exit code not captured")
	}
	if len(cmdErr.Args) == 0 || cmdErr.Args[0] != "task" {
		t.Errorf("Args = %v, want the command that was run", cmdErr.Args)
	}
	if cmdErr.Dir != dir {
		t.Errorf("Dir = %q, want the project", cmdErr.Dir)
	}
	if strings.TrimSpace(cmdErr.Stderr)+strings.TrimSpace(cmdErr.Stdout) == "" {
		t.Error("nothing captured from the CLI to show the user")
	}
	if !strings.Contains(err.Error(), "task edit") {
		t.Errorf("Error() = %q, want it to name the command", err.Error())
	}
}

// Two edits in one project can collide over ids and ordinals, so writes to a
// project are serialised. Different projects have no reason to wait.
func TestConcurrentWritesToOneProjectAreSerialised(t *testing.T) {
	r := runner(t)
	dir := newProject(t, r)

	const count = 8
	var wg sync.WaitGroup
	errs := make([]error, count)
	for i := range count {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := r.Exec(context.Background(), dir,
				"task", "create", "Concurrent task", "--plain")
			errs[i] = err
		}()
	}
	wg.Wait()

	for i, err := range errs {
		if err != nil {
			t.Errorf("write %d failed: %v", i, err)
		}
	}

	tasks := scan(t, dir).Tasks
	if len(tasks) != count {
		t.Fatalf("got %d tasks, want %d", len(tasks), count)
	}
	// Every id must be distinct: a lost update would show up here as two tasks
	// claiming the same id.
	ids := map[string]bool{}
	for _, task := range tasks {
		if ids[task.ID] {
			t.Errorf("duplicate id %s — writes were not serialised", task.ID)
		}
		ids[task.ID] = true
	}
}

func TestDraftsAndPromotion(t *testing.T) {
	r := runner(t)
	dir := newProject(t, r)
	ctx := context.Background()

	if err := r.CreateDraft(ctx, dir, "An idea worth keeping", "Some context."); err != nil {
		t.Fatalf("CreateDraft: %v", err)
	}

	p := scan(t, dir)
	if len(p.Drafts) != 1 {
		t.Fatalf("got %d drafts, want 1", len(p.Drafts))
	}
	// Drafts stay off the board, which is what makes capture cheap.
	if len(p.Tasks) != 0 {
		t.Errorf("a draft reached the task list")
	}

	if err := r.PromoteDraft(ctx, dir, p.Drafts[0].ID); err != nil {
		t.Fatalf("PromoteDraft: %v", err)
	}
	p = scan(t, dir)
	if len(p.Tasks) != 1 || len(p.Drafts) != 0 {
		t.Errorf("after promotion: %d tasks, %d drafts", len(p.Tasks), len(p.Drafts))
	}
}

// Every prompt `backlog init` asks has a flag, which is what lets the UI put a
// form over it instead of emulating a dialogue.
func TestInitIsNonInteractive(t *testing.T) {
	r := runner(t)
	dir := t.TempDir()

	err := r.Init(context.Background(), dir, backlogcli.InitOptions{
		Name:            "Custom",
		BacklogDir:      ".backlog",
		TaskPrefix:      "story",
		ZeroPaddedIDs:   3,
		NoGit:           true,
		IntegrationMode: "none",
	})
	if err != nil {
		t.Fatalf("Init: %v", err)
	}

	loc, err := project.Discover(dir)
	if err != nil {
		t.Fatalf("Discover: %v", err)
	}
	if loc.Layout != project.LayoutHidden {
		t.Errorf("Layout = %q, want the hidden layout that was asked for", loc.Layout)
	}
	cfg, err := backlog.LoadConfig(loc.ConfigPath)
	if err != nil {
		t.Fatalf("LoadConfig: %v", err)
	}
	if cfg.TaskPrefix != "story" {
		t.Errorf("TaskPrefix = %q, want story", cfg.TaskPrefix)
	}
	if cfg.ZeroPaddedIDs != 3 {
		t.Errorf("ZeroPaddedIDs = %d, want 3", cfg.ZeroPaddedIDs)
	}
}

// The CLI refuses agent instructions when AI integration is off. Catching it
// before running means the form in the Projects screen can prevent the
// combination rather than show a failure afterwards.
func TestConflictingInitOptionsAreRefusedBeforeRunning(t *testing.T) {
	r := runner(t)
	err := r.Init(context.Background(), t.TempDir(), backlogcli.InitOptions{
		IntegrationMode:   "none",
		AgentInstructions: "claude",
	})
	if !errors.Is(err, backlogcli.ErrConflictingInitOptions) {
		t.Fatalf("err = %v, want ErrConflictingInitOptions", err)
	}
}

func TestAcceptanceCriteriaCanBeTicked(t *testing.T) {
	r := runner(t)
	dir := newProject(t, r)
	ctx := context.Background()

	out, err := r.Exec(ctx, dir, "task", "create", "A task with criteria",
		"--ac", "First", "--ac", "Second", "--plain")
	if err != nil {
		t.Fatalf("task create: %v", err)
	}
	var id string
	for _, field := range strings.Fields(out) {
		if strings.HasPrefix(strings.ToUpper(field), "TASK-") {
			id = strings.TrimSuffix(strings.ToUpper(field), ":")
			break
		}
	}

	if err := r.CheckAcceptanceCriterion(ctx, dir, id, 2); err != nil {
		t.Fatalf("CheckAcceptanceCriterion: %v", err)
	}
	criteria := scan(t, dir).Tasks[0].AcceptanceCriteria
	if len(criteria) != 2 {
		t.Fatalf("got %d criteria", len(criteria))
	}
	if criteria[0].Checked || !criteria[1].Checked {
		t.Errorf("checked states = %v, %v; want only the second ticked",
			criteria[0].Checked, criteria[1].Checked)
	}
}

func TestCreateTaskWithEverySupportedField(t *testing.T) {
	r := runner(t)
	dir := newProject(t, r)

	id, err := r.CreateTask(context.Background(), dir, backlogcli.NewTask{
		Title:              "A fully specified task",
		Description:        "Why it matters.",
		Status:             "In Progress",
		Priority:           "High",
		Type:               "feature",
		Assignee:           "@someone",
		Labels:             []string{"backend", "urgent"},
		AcceptanceCriteria: []string{"First criterion", "Second criterion"},
	})
	if err != nil {
		t.Fatalf("CreateTask: %v", err)
	}
	// The id comes from the CLI rather than being predicted, because generation
	// handles collisions and zero-padding per project.
	if id == "" {
		t.Fatal("no id returned")
	}

	tasks := scan(t, dir).Tasks
	if len(tasks) != 1 {
		t.Fatalf("got %d tasks", len(tasks))
	}
	task := tasks[0]
	if task.ID != id {
		t.Errorf("id = %q, want the reported %q", task.ID, id)
	}
	if task.Title != "A fully specified task" || task.Status != "In Progress" {
		t.Errorf("title/status = %q/%q", task.Title, task.Status)
	}
	if !strings.EqualFold(task.Priority, "High") || task.Type != "feature" {
		t.Errorf("priority/type = %q/%q", task.Priority, task.Type)
	}
	if len(task.Labels) != 2 || len(task.Assignee) != 1 {
		t.Errorf("labels = %v, assignee = %v", task.Labels, task.Assignee)
	}
	if len(task.AcceptanceCriteria) != 2 {
		t.Fatalf("criteria = %+v, want two", task.AcceptanceCriteria)
	}
	if body, ok := task.Section("description"); !ok || !strings.Contains(body, "Why it matters") {
		t.Errorf("description = %q", body)
	}
}

func TestCreateTaskNeedsATitle(t *testing.T) {
	r := runner(t)
	if _, err := r.CreateTask(context.Background(), newProject(t, r),
		backlogcli.NewTask{Title: "   "}); err == nil {
		t.Fatal("want an error for a blank title")
	}
}

func TestEditTaskBody(t *testing.T) {
	r := runner(t)
	dir := newProject(t, r)
	ctx := context.Background()
	id := createTask(t, r, dir, "A task to rewrite")

	if err := r.SetTitle(ctx, dir, id, "A renamed task"); err != nil {
		t.Fatalf("SetTitle: %v", err)
	}
	if err := r.SetDescription(ctx, dir, id, "New description.\n\nWith a second paragraph."); err != nil {
		t.Fatalf("SetDescription: %v", err)
	}
	if err := r.SetPlan(ctx, dir, id, "1. First\n2. Second"); err != nil {
		t.Fatalf("SetPlan: %v", err)
	}
	if err := r.SetNotes(ctx, dir, id, "What happened."); err != nil {
		t.Fatalf("SetNotes: %v", err)
	}

	task := scan(t, dir).Tasks[0]
	if task.Title != "A renamed task" {
		t.Errorf("title = %q", task.Title)
	}
	for _, want := range []struct {
		section backlog.Section
		text    string
	}{
		{"description", "second paragraph"},
		{"plan", "1. First"},
		{"notes", "What happened."},
	} {
		body, ok := task.Section(want.section)
		if !ok || !strings.Contains(body, want.text) {
			t.Errorf("%s = %q, want it to contain %q", want.section, body, want.text)
		}
	}
}

// Replacing the whole list is what makes adding, removing and reordering one
// operation, and keeps the per-item indexes the CLI renumbers out of the UI.
func TestAcceptanceCriteriaAreReplacedWholesale(t *testing.T) {
	r := runner(t)
	dir := newProject(t, r)
	ctx := context.Background()
	id := createTask(t, r, dir, "A task with criteria")

	if err := r.SetAcceptanceCriteria(ctx, dir, id, []string{"One", "Two", "Three"}); err != nil {
		t.Fatalf("SetAcceptanceCriteria: %v", err)
	}
	if got := scan(t, dir).Tasks[0].AcceptanceCriteria; len(got) != 3 {
		t.Fatalf("criteria = %+v, want three", got)
	}

	// Reorder and drop one in a single call.
	if err := r.SetAcceptanceCriteria(ctx, dir, id, []string{"Three", "One"}); err != nil {
		t.Fatalf("SetAcceptanceCriteria: %v", err)
	}
	got := scan(t, dir).Tasks[0].AcceptanceCriteria
	if len(got) != 2 || got[0].Text != "Three" || got[1].Text != "One" {
		t.Fatalf("criteria = %+v, want Three then One", got)
	}
	for i, c := range got {
		if c.Index != i+1 {
			t.Errorf("criterion %d has index %d, want them renumbered", i, c.Index)
		}
	}

	if err := r.SetAcceptanceCriteria(ctx, dir, id, nil); err != nil {
		t.Fatalf("SetAcceptanceCriteria(nil): %v", err)
	}
	if got := scan(t, dir).Tasks[0].AcceptanceCriteria; len(got) != 0 {
		t.Errorf("criteria = %+v, want none left", got)
	}
}
