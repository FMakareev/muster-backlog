package app_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/FMakareev/muster-backlog/internal/app"
)

func ptr(s string) *string { return &s }

// What a task's frontmatter says now, read from the file rather than the store.
func frontmatter(t *testing.T, project, file string) string {
	t.Helper()
	content, err := os.ReadFile(filepath.Join(project, "backlog", "tasks", file))
	if err != nil {
		t.Fatalf("read task: %v", err)
	}
	_, rest, ok := strings.Cut(string(content), "---\n")
	if !ok {
		t.Fatalf("no frontmatter in %s", file)
	}
	head, _, _ := strings.Cut(rest, "\n---")
	return head
}

// One change, many tasks, one project.
func TestManyTasksTakeOneChange(t *testing.T) {
	if _, err := exec.LookPath("backlog"); err != nil {
		t.Skip("the backlog CLI is not installed")
	}
	one := newProject(t, "one", map[string]string{
		"task-1 - a.md": task("TASK-1", "First", "To Do"),
		"task-2 - b.md": task("TASK-2", "Second", "To Do"),
		"task-3 - c.md": task("TASK-3", "Third", "To Do"),
	})
	s := startService(t, withRegistry(t, one))

	result := s.ChangeMany(app.BulkChange{
		Tasks: []app.BulkTarget{
			{Project: one, TaskID: "TASK-1"},
			{Project: one, TaskID: "TASK-2"},
		},
		Status:    ptr("Done"),
		AddLabels: []string{"swept"},
	})

	if result.Problem != nil {
		t.Fatalf("refused: %+v", result.Problem)
	}
	if result.Changed != 2 || len(result.Failures) != 0 {
		t.Fatalf("changed %d, failures %+v", result.Changed, result.Failures)
	}

	for _, file := range []string{"task-1 - a.md", "task-2 - b.md"} {
		head := frontmatter(t, one, file)
		if !strings.Contains(head, "status: Done") {
			t.Errorf("%s did not move: %s", file, head)
		}
		if !strings.Contains(head, "swept") {
			t.Errorf("%s was not labelled: %s", file, head)
		}
	}
	// The one nobody chose is untouched, which is the whole basis of trusting
	// a bulk change at all.
	if head := frontmatter(t, one, "task-3 - c.md"); !strings.Contains(head, "status: To Do") {
		t.Errorf("an unselected task was changed: %s", head)
	}
}

// A run of twenty writes can partly fail, and saying "done" would be a lie.
func TestAFailedTaskIsNamedAndTheRestStillChange(t *testing.T) {
	if _, err := exec.LookPath("backlog"); err != nil {
		t.Skip("the backlog CLI is not installed")
	}
	one := newProject(t, "one", map[string]string{
		"task-1 - a.md": task("TASK-1", "First", "To Do"),
		"task-2 - b.md": task("TASK-2", "Second", "To Do"),
	})
	s := startService(t, withRegistry(t, one))

	result := s.ChangeMany(app.BulkChange{
		Tasks: []app.BulkTarget{
			{Project: one, TaskID: "TASK-1"},
			{Project: one, TaskID: "TASK-404"}, // there is no such task
			{Project: one, TaskID: "TASK-2"},
		},
		Status: ptr("Done"),
	})

	if result.Changed != 2 {
		t.Errorf("changed %d, want the two that exist", result.Changed)
	}
	if len(result.Failures) != 1 {
		t.Fatalf("failures %+v, want the missing one named", result.Failures)
	}
	if result.Failures[0].TaskID != "TASK-404" {
		t.Errorf("named the wrong task: %+v", result.Failures[0])
	}
	// The CLI's own words, not a sentence invented here: it is the only thing
	// that knows why.
	if !strings.Contains(strings.ToLower(result.Failures[0].Detail), "not found") {
		t.Errorf("the reason is missing: %q", result.Failures[0].Detail)
	}
	if result.Failures[0].ProjectName != "one" {
		t.Errorf("failure does not say where: %+v", result.Failures[0])
	}
	// A failure in the middle does not stop the queue.
	if head := frontmatter(t, one, "task-2 - b.md"); !strings.Contains(head, "status: Done") {
		t.Errorf("the task after the failure was skipped: %s", head)
	}
}

// Status and labels cross projects perfectly well.
func TestAChangeSpanningProjectsIsApplied(t *testing.T) {
	if _, err := exec.LookPath("backlog"); err != nil {
		t.Skip("the backlog CLI is not installed")
	}
	one := newProject(t, "one", map[string]string{
		"task-1 - a.md": task("TASK-1", "First", "To Do"),
	})
	two := newProject(t, "two", map[string]string{
		"task-1 - a.md": task("TASK-1", "Also first", "To Do"),
	})
	s := startService(t, withRegistry(t, one, two))

	result := s.ChangeMany(app.BulkChange{
		Tasks: []app.BulkTarget{
			{Project: one, TaskID: "TASK-1"},
			{Project: two, TaskID: "TASK-1"},
		},
		AddLabels: []string{"release"},
	})

	if result.Problem != nil || result.Changed != 2 {
		t.Fatalf("changed %d, problem %+v", result.Changed, result.Problem)
	}
	// Ids collide across projects; both were reached, not one of them twice.
	for _, project := range []string{one, two} {
		if head := frontmatter(t, project, "task-1 - a.md"); !strings.Contains(head, "release") {
			t.Errorf("%s was not labelled: %s", project, head)
		}
	}
}

// The measured reason: `task edit -m` accepts an id the project does not have,
// writes it, and reports success. Across projects that plants a dangling
// reference everywhere but one, so the run is refused before any of it starts.
func TestAMilestoneIsRefusedAcrossProjects(t *testing.T) {
	one := newProject(t, "one", map[string]string{
		"task-1 - a.md": task("TASK-1", "First", "To Do"),
	})
	two := newProject(t, "two", map[string]string{
		"task-1 - a.md": task("TASK-1", "Also first", "To Do"),
	})
	s := startService(t, withRegistry(t, one, two))

	result := s.ChangeMany(app.BulkChange{
		Tasks: []app.BulkTarget{
			{Project: one, TaskID: "TASK-1"},
			{Project: two, TaskID: "TASK-1"},
		},
		Milestone: ptr("m-1"),
		Status:    ptr("Done"),
	})

	if result.Problem == nil {
		t.Fatalf("a milestone was accepted across projects: changed %d", result.Changed)
	}
	if result.Changed != 0 {
		t.Errorf("%d tasks were changed by a refused run", result.Changed)
	}
	// It names both projects, so the refusal is actionable rather than a rule.
	for _, name := range []string{"one", "two"} {
		if !strings.Contains(result.Problem.Detail, name) {
			t.Errorf("the refusal does not name %s: %s", name, result.Problem.Detail)
		}
	}
	// Nothing else in the change was applied either: refusing half of it would
	// be worse than refusing all of it.
	if head := frontmatter(t, one, "task-1 - a.md"); strings.Contains(head, "status: Done") {
		t.Errorf("the rest of a refused change was applied anyway: %s", head)
	}

	// Inside one project the same change is fine.
	within := s.ChangeMany(app.BulkChange{
		Tasks:     []app.BulkTarget{{Project: one, TaskID: "TASK-1"}},
		Milestone: ptr("m-1"),
	})
	if within.Problem != nil {
		t.Errorf("refused inside a single project: %+v", within.Problem)
	}
}

// An edit that asks for nothing still rewrites updated_date, which would make
// a mistaken run look like work.
func TestAnEmptyChangeIsRefused(t *testing.T) {
	one := newProject(t, "one", map[string]string{
		"task-1 - a.md": task("TASK-1", "First", "To Do"),
	})
	s := startService(t, withRegistry(t, one))

	result := s.ChangeMany(app.BulkChange{
		Tasks:     []app.BulkTarget{{Project: one, TaskID: "TASK-1"}},
		AddLabels: []string{"  ", ""},
	})
	if result.Problem == nil {
		t.Fatalf("an empty change was run: changed %d", result.Changed)
	}

	if empty := s.ChangeMany(app.BulkChange{Status: ptr("Done")}); empty.Problem == nil {
		t.Error("a change with no tasks was run")
	}
}
