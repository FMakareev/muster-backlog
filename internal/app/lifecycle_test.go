package app_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/FMakareev/muster-backlog/internal/app"
	"github.com/FMakareev/muster-backlog/internal/testcli"
)

// project builds a project whose statuses end in something other than "Done",
// so the rules below are tested against a project's own terminal status rather
// than against a word.
func shippedProject(t *testing.T, name string, tasks map[string]string) string {
	t.Helper()
	root := filepath.Join(t.TempDir(), name)
	dir := filepath.Join(root, "backlog", "tasks")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "backlog", "config.yml"),
		[]byte("project_name: \""+name+"\"\nstatuses: [\"To Do\", \"Shipped\"]\n"),
		0o644); err != nil {
		t.Fatalf("config: %v", err)
	}
	for file, content := range tasks {
		if err := os.WriteFile(filepath.Join(dir, file), []byte(content), 0o644); err != nil {
			t.Fatalf("task: %v", err)
		}
	}
	return root
}

func liveTitles(views []app.TaskView) []string {
	var out []string
	for _, v := range views {
		out = append(out, v.Entity.Title)
	}
	return out
}

func TestCompletingMovesAFinishedTaskOutOfTheWay(t *testing.T) {
	testcli.Require(t)
	one := shippedProject(t, "one", map[string]string{
		"task-1 - done.md": task("TASK-1", "Finished", "Shipped"),
		"task-2 - open.md": task("TASK-2", "Open", "To Do"),
	})
	s := startService(t, withRegistry(t, one))

	if result := s.CompleteTask(one, "TASK-1"); !result.OK {
		t.Fatalf("CompleteTask: %+v", result.Problem)
	}
	if got := liveTitles(s.Tasks(app.QueryInput{})); strings.Join(got, ",") != "Open" {
		t.Errorf("the board still holds %v", got)
	}
	// Moved, not deleted: Backlog.md keeps finished work in completed/.
	entries, err := os.ReadDir(filepath.Join(one, "backlog", "completed"))
	if err != nil || len(entries) != 1 {
		t.Fatalf("nothing in completed/: %v %v", entries, err)
	}

	// And a task that is not at its project's last status is refused, with the
	// CLI's own reason rather than a guess.
	result := s.CompleteTask(one, "TASK-2")
	if result.OK {
		t.Fatal("an unfinished task was completed")
	}
	if result.Problem == nil || result.Problem.Detail == "" {
		t.Error("refused without an explanation")
	}
	if got := liveTitles(s.Tasks(app.QueryInput{})); len(got) != 1 {
		t.Errorf("a refused completion changed the board: %v", got)
	}
}

func TestArchivingTakesATaskOffTheBoardWithoutDeletingIt(t *testing.T) {
	testcli.Require(t)
	one := shippedProject(t, "one", map[string]string{
		"task-1 - a.md": task("TASK-1", "Never mind", "To Do"),
	})
	s := startService(t, withRegistry(t, one))

	if result := s.ArchiveTask(one, "TASK-1"); !result.OK {
		t.Fatalf("ArchiveTask: %+v", result.Problem)
	}
	if got := s.Tasks(app.QueryInput{}); len(got) != 0 {
		t.Errorf("the board still holds %v", liveTitles(got))
	}
	entries, err := os.ReadDir(filepath.Join(one, "backlog", "archive", "tasks"))
	if err != nil || len(entries) != 1 {
		t.Errorf("the file was not archived: %v %v", entries, err)
	}
}

// The other half of the inbox: a note promoted a moment too early goes back.
func TestDemotingReturnsATaskToTheInbox(t *testing.T) {
	testcli.Require(t)
	one := shippedProject(t, "one", map[string]string{
		"task-1 - a.md": task("TASK-1", "Not ready after all", "To Do"),
	})
	s := startService(t, withRegistry(t, one))

	if result := s.DemoteTask(one, "TASK-1"); !result.OK {
		t.Fatalf("DemoteTask: %+v", result.Problem)
	}
	if got := s.Tasks(app.QueryInput{}); len(got) != 0 {
		t.Errorf("the board still holds %v", liveTitles(got))
	}
	drafts := s.Drafts()
	if len(drafts) != 1 || drafts[0].Entity.Title != "Not ready after all" {
		t.Fatalf("the note did not arrive in the inbox: %+v", drafts)
	}
	// Measured, not assumed: demote moves the file and renames the id, and
	// leaves the status exactly as it was. A demoted note is a draft by where
	// it lives, not by what its status says - which is why the inbox lists by
	// directory rather than by status.
	if drafts[0].ID != "DRAFT-1" {
		t.Errorf("id is %q, want a DRAFT id", drafts[0].ID)
	}
	if drafts[0].Entity.Status != "To Do" {
		t.Errorf("status is %q; demote is documented as leaving it alone",
			drafts[0].Entity.Status)
	}
}

// None of the three is taken on the strength of an exit code.
func TestLifecycleWritesAreConfirmedByOutcome(t *testing.T) {
	testcli.Require(t)
	one := shippedProject(t, "one", map[string]string{
		"task-1 - a.md": task("TASK-1", "Still here", "To Do"),
	})
	s := startService(t, withRegistry(t, one))

	for name, ok := range map[string]bool{
		"complete": s.CompleteTask(one, "TASK-404").OK,
		"archive":  s.ArchiveTask(one, "TASK-404").OK,
		"demote":   s.DemoteTask(one, "TASK-404").OK,
	} {
		if ok {
			t.Errorf("%s of an id that does not resolve reported success", name)
		}
	}
	if got := s.Tasks(app.QueryInput{}); len(got) != 1 {
		t.Errorf("the board changed: %v", liveTitles(got))
	}
}
