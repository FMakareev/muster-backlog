package app_test

import (
	"context"
	"os/exec"
	"strings"
	"testing"

	"github.com/FMakareev/muster-backlog/internal/app"
	"github.com/FMakareev/muster-backlog/internal/backlogcli"
)

// realProject initialises a project through the CLI and returns its path.
func realProject(t *testing.T) string {
	t.Helper()
	if _, err := exec.LookPath(backlogcli.BinaryName); err != nil {
		t.Skipf("the %s CLI is not installed", backlogcli.BinaryName)
	}
	r, err := backlogcli.New()
	if err != nil {
		t.Skipf("the CLI is not usable: %v", err)
	}
	dir := t.TempDir()
	if err := r.Init(context.Background(), dir, backlogcli.InitOptions{
		Name: "Scratch", NoGit: true, IntegrationMode: "none",
	}); err != nil {
		t.Fatalf("Init: %v", err)
	}
	if _, err := r.Exec(context.Background(), dir,
		"task", "create", "A task to move", "--plain"); err != nil {
		t.Fatalf("task create: %v", err)
	}
	return dir
}

func firstTaskID(t *testing.T, s *app.BoardService) string {
	t.Helper()
	tasks := s.Tasks(app.QueryInput{})
	if len(tasks) == 0 {
		t.Fatal("no tasks loaded")
	}
	return tasks[0].ID
}

// A write must end with the store reflecting what the files say, not what the
// caller asked for.
func TestWriteIsFollowedByARescan(t *testing.T) {
	dir := realProject(t)
	s := startService(t, withRegistry(t, dir))

	id := firstTaskID(t, s)
	if got := s.Tasks(app.QueryInput{})[0].Entity.Status; got != "To Do" {
		t.Fatalf("status = %q at the start", got)
	}

	result := s.SetStatus(dir, id, "In Progress")
	if !result.OK {
		t.Fatalf("SetStatus: %+v", result.Problem)
	}

	// No sleeping and no waiting on the watcher: the write itself re-read the
	// project, so the new state is already visible.
	if got := s.Tasks(app.QueryInput{})[0].Entity.Status; got != "In Progress" {
		t.Errorf("status = %q, want the rescan to have picked up the write", got)
	}
}

func TestPriorityAndCriteriaWrites(t *testing.T) {
	dir := realProject(t)
	s := startService(t, withRegistry(t, dir))
	id := firstTaskID(t, s)

	if r := s.SetPriority(dir, id, "High"); !r.OK {
		t.Fatalf("SetPriority: %+v", r.Problem)
	}
	if got := s.Tasks(app.QueryInput{})[0].Entity.Priority; !strings.EqualFold(got, "High") {
		t.Errorf("priority = %q", got)
	}

	if r := s.SetAssignee(dir, id, "@someone"); !r.OK {
		t.Fatalf("SetAssignee: %+v", r.Problem)
	}
	if got := s.Tasks(app.QueryInput{})[0].Entity.Assignee; len(got) != 1 {
		t.Errorf("assignee = %v", got)
	}
}

// A failed write must leave the store showing what the files actually say, so
// the interface can return the card to where it was.
func TestFailedWriteLeavesTheStoreTruthful(t *testing.T) {
	dir := realProject(t)
	s := startService(t, withRegistry(t, dir))

	before := s.Tasks(app.QueryInput{})[0].Entity.Status

	result := s.SetStatus(dir, "TASK-404", "Done")
	if result.OK {
		t.Fatal("want a failure for a task that does not exist")
	}
	if result.Problem == nil {
		t.Fatal("a failed write must carry a problem the interface can show")
	}
	if result.Problem.Title == "" || result.Problem.Detail == "" {
		t.Errorf("problem = %+v, want something renderable", result.Problem)
	}
	if !strings.Contains(result.Problem.Detail, "TASK-404") {
		t.Errorf("detail = %q, want it to name what failed", result.Problem.Detail)
	}

	if got := s.Tasks(app.QueryInput{})[0].Entity.Status; got != before {
		t.Errorf("status = %q, want the real task untouched at %q", got, before)
	}
}

// Titles containing shell metacharacters must survive a round trip through the
// write path unchanged.
func TestHostileTextSurvivesTheWritePath(t *testing.T) {
	dir := realProject(t)
	s := startService(t, withRegistry(t, dir))

	title := "Fix $(whoami) and `ls`; then \"quote\" it — Риг байка"
	if r := s.CaptureDraft(dir, title, "Body with 'quotes' & pipes | here"); !r.OK {
		t.Fatalf("CaptureDraft: %+v", r.Problem)
	}

	drafts := s.Tasks(app.QueryInput{Kinds: []string{"draft"}})
	if len(drafts) != 1 {
		t.Fatalf("got %d drafts, want 1", len(drafts))
	}
	if drafts[0].Entity.Title != title {
		t.Errorf("title = %q, want it unchanged", drafts[0].Entity.Title)
	}
}

func TestCLIVersionIsReported(t *testing.T) {
	dir := realProject(t)
	s := startService(t, withRegistry(t, dir))
	if s.CLIVersion() == "" {
		t.Error("no CLI version recorded")
	}
}
