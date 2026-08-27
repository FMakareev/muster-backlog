package app_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/FMakareev/muster-backlog/internal/app"
	"github.com/FMakareev/muster-backlog/internal/testcli"
)

func milestoneOf(t *testing.T, s *app.BoardService, project, taskID string) string {
	t.Helper()
	view, ok := s.Task(project, "task", "active", taskID)
	if !ok {
		t.Fatalf("%s is gone", taskID)
	}
	return view.Entity.Milestone
}

func milestoneTitles(s *app.BoardService, project string) []string {
	var out []string
	for _, m := range s.Milestones() {
		if m.Project == project {
			out = append(out, m.Title)
		}
	}
	return out
}

func TestAddingAndRenamingAMilestone(t *testing.T) {
	testcli.Require(t)
	one := initialised(t, "one")
	s := startService(t, withRegistry(t, one))

	created := s.AddMilestone(one, "First push", "The opening moves")
	if !created.OK {
		t.Fatalf("AddMilestone: %+v", created.Problem)
	}
	if created.TaskID == "" {
		t.Error("the new milestone's id was not reported")
	}
	if got := milestoneTitles(s, one); strings.Join(got, ",") != "First push" {
		t.Fatalf("milestones are %v", got)
	}

	if result := s.RenameMilestone(one, "First push", "Opening push"); !result.OK {
		t.Fatalf("RenameMilestone: %+v", result.Problem)
	}
	if got := milestoneTitles(s, one); strings.Join(got, ",") != "Opening push" {
		t.Errorf("after renaming, milestones are %v", got)
	}
	// The id survives a rename, which is what keeps the tasks pointing at it.
	if !strings.HasPrefix(created.TaskID, "m-") {
		t.Errorf("id is %q", created.TaskID)
	}

	if result := s.AddMilestone(one, "   ", ""); result.OK {
		t.Error("a milestone with no name was accepted")
	}
}

// Retiring a milestone archives the file either way. What the caller is really
// choosing is what happens to the tasks, so all three answers are tested.
func TestRetiringAMilestoneDecidesWhatBecomesOfItsTasks(t *testing.T) {
	testcli.Require(t)

	cases := []struct {
		handling string
		want     string
	}{
		{"keep", "m-0"},
		{"clear", ""},
	}
	for _, c := range cases {
		t.Run(c.handling, func(t *testing.T) {
			one := initialised(t, "one")
			s := startService(t, withRegistry(t, one))
			if result := s.AddMilestone(one, "Doomed", ""); !result.OK {
				t.Fatalf("AddMilestone: %+v", result.Problem)
			}
			create := exec.Command("backlog", "task", "create", "Inside it", "-m", "m-0")
			create.Dir = one
			if out, err := create.CombinedOutput(); err != nil {
				t.Fatalf("task create: %v\n%s", err, out)
			}
			s.Reload()

			if result := s.RetireMilestone(one, "Doomed", c.handling, ""); !result.OK {
				t.Fatalf("RetireMilestone(%s): %+v", c.handling, result.Problem)
			}
			if got := milestoneTitles(s, one); len(got) != 0 {
				t.Errorf("the milestone is still listed: %v", got)
			}
			if got := milestoneOf(t, s, one, "TASK-1"); got != c.want {
				t.Errorf("the task's milestone is %q, want %q", got, c.want)
			}
			// Archived rather than deleted, both ways.
			entries, err := os.ReadDir(filepath.Join(one, "backlog", "archive", "milestones"))
			if err != nil || len(entries) != 1 {
				t.Errorf("the milestone file was not archived: %v %v", entries, err)
			}
		})
	}
}

func TestReassigningTasksToAnotherMilestone(t *testing.T) {
	testcli.Require(t)
	one := initialised(t, "one")
	s := startService(t, withRegistry(t, one))
	for _, name := range []string{"Leaving", "Staying"} {
		if result := s.AddMilestone(one, name, ""); !result.OK {
			t.Fatalf("AddMilestone(%s): %+v", name, result.Problem)
		}
	}
	create := exec.Command("backlog", "task", "create", "Moves across", "-m", "m-0")
	create.Dir = one
	if out, err := create.CombinedOutput(); err != nil {
		t.Fatalf("task create: %v\n%s", err, out)
	}
	s.Reload()

	// A target that does not exist is refused before anything is written.
	if result := s.RetireMilestone(one, "Leaving", "reassign", "Nowhere"); result.OK {
		t.Fatal("reassigning to a milestone that does not exist was accepted")
	}
	if got := milestoneTitles(s, one); len(got) != 2 {
		t.Fatalf("a refused retirement changed the milestones: %v", got)
	}

	if result := s.RetireMilestone(one, "Leaving", "reassign", "Staying"); !result.OK {
		t.Fatalf("RetireMilestone: %+v", result.Problem)
	}
	if got := milestoneOf(t, s, one, "TASK-1"); got != "m-1" {
		t.Errorf("the task points at %q, want the milestone it was moved to", got)
	}
}

// A task's milestone is ordinary work, not a decision taken once at creation.
func TestATasksMilestoneCanBeChangedAndCleared(t *testing.T) {
	testcli.Require(t)
	one := initialised(t, "one")
	s := startService(t, withRegistry(t, one))
	for _, name := range []string{"First", "Second"} {
		if result := s.AddMilestone(one, name, ""); !result.OK {
			t.Fatalf("AddMilestone(%s): %+v", name, result.Problem)
		}
	}
	create := exec.Command("backlog", "task", "create", "Moves about")
	create.Dir = one
	if out, err := create.CombinedOutput(); err != nil {
		t.Fatalf("task create: %v\n%s", err, out)
	}
	s.Reload()

	if result := s.SetMilestone(one, "TASK-1", "m-0"); !result.OK {
		t.Fatalf("SetMilestone: %+v", result.Problem)
	}
	if got := milestoneOf(t, s, one, "TASK-1"); got != "m-0" {
		t.Errorf("milestone is %q", got)
	}

	if result := s.SetMilestone(one, "TASK-1", "m-1"); !result.OK {
		t.Fatalf("moving it: %+v", result.Problem)
	}
	if got := milestoneOf(t, s, one, "TASK-1"); got != "m-1" {
		t.Errorf("after moving, milestone is %q", got)
	}

	// Empty clears it, which is a different command behind the same control.
	if result := s.SetMilestone(one, "TASK-1", ""); !result.OK {
		t.Fatalf("clearing it: %+v", result.Problem)
	}
	if got := milestoneOf(t, s, one, "TASK-1"); got != "" {
		t.Errorf("after clearing, milestone is %q", got)
	}
}
