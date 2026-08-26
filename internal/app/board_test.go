package app_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/FMakareev/muster-backlog/internal/app"
)

// withRegistry points the XDG config home at a temporary directory holding a
// registry over the given projects, so the service loads exactly what the test
// wrote and never touches the developer's own configuration.
func withRegistry(t *testing.T, projects ...string) string {
	t.Helper()

	configHome := t.TempDir()
	dir := filepath.Join(configHome, "muster")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	var lines []string
	for _, p := range projects {
		lines = append(lines, "  - path: "+p)
	}
	content := "projects:\n" + strings.Join(lines, "\n") + "\n"
	if len(projects) == 0 {
		content = "projects: []\n"
	}
	path := filepath.Join(dir, "projects.yml")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write registry: %v", err)
	}
	return path
}

func newProject(t *testing.T, name string, tasks map[string]string) string {
	t.Helper()
	root := filepath.Join(t.TempDir(), name)
	tasksDir := filepath.Join(root, "backlog", "tasks")
	if err := os.MkdirAll(tasksDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "backlog", "config.yml"),
		[]byte("project_name: \""+name+"\"\nstatuses: [\"To Do\", \"Done\"]\n"),
		0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}
	for file, content := range tasks {
		if err := os.WriteFile(filepath.Join(tasksDir, file), []byte(content), 0o644); err != nil {
			t.Fatalf("write task: %v", err)
		}
	}
	return root
}

func startService(t *testing.T, registryPath string) *app.BoardService {
	t.Helper()
	s := app.NewBoardServiceAt(registryPath)
	if err := s.ServiceStartup(t.Context(), serviceOptions()); err != nil {
		t.Fatalf("ServiceStartup: %v", err)
	}
	t.Cleanup(func() {
		if err := s.ServiceShutdown(); err != nil {
			t.Errorf("ServiceShutdown: %v", err)
		}
	})
	return s
}

// The frontend must be able to ask for the registry and the whole task set.
func TestServiceExposesProjectsAndTasks(t *testing.T) {
	alpha := newProject(t, "alpha", map[string]string{
		"task-1 - a.md": "---\nid: TASK-1\ntitle: Alpha one\nstatus: To Do\npriority: high\n---\n",
		"task-2 - b.md": "---\nid: TASK-2\ntitle: Alpha two\nstatus: Done\n---\n",
	})
	beta := newProject(t, "beta", map[string]string{
		"task-1 - c.md": "---\nid: TASK-1\ntitle: Beta one\nstatus: To Do\n---\n",
	})
	s := startService(t, withRegistry(t, alpha, beta))

	projects := s.Projects()
	if len(projects) != 2 {
		t.Fatalf("got %d projects, want 2", len(projects))
	}
	if projects[0].Name != "alpha" || !projects[0].OK {
		t.Errorf("first project = %+v", projects[0])
	}
	if projects[0].TaskCount != 2 {
		t.Errorf("alpha has %d tasks, want 2", projects[0].TaskCount)
	}
	if len(projects[0].Statuses) != 2 {
		t.Errorf("statuses = %v, want the project's own list", projects[0].Statuses)
	}

	tasks := s.Tasks(app.QueryInput{})
	if len(tasks) != 3 {
		t.Fatalf("got %d tasks, want 3", len(tasks))
	}

	// The same id in two projects must arrive as two tasks the frontend can
	// tell apart.
	var ones []app.TaskView
	for _, task := range tasks {
		if task.ID == "TASK-1" {
			ones = append(ones, task)
		}
	}
	if len(ones) != 2 {
		t.Fatalf("got %d tasks with id TASK-1, want one per project", len(ones))
	}
	if ones[0].Project == ones[1].Project {
		t.Error("both TASK-1s claim the same project")
	}

	filtered := s.Tasks(app.QueryInput{Statuses: []string{"To Do"}})
	if len(filtered) != 2 {
		t.Errorf("got %d To Do tasks, want 2", len(filtered))
	}
}

func TestServiceReturnsOneTaskByRef(t *testing.T) {
	alpha := newProject(t, "alpha", map[string]string{
		"task-1 - a.md": "---\nid: TASK-1\ntitle: Alpha one\nstatus: To Do\n---\n\n" +
			"## Description\n\n<!-- SECTION:DESCRIPTION:BEGIN -->\nThe body.\n" +
			"<!-- SECTION:DESCRIPTION:END -->\n",
	})
	s := startService(t, withRegistry(t, alpha))

	got, ok := s.Task(alpha, "task", "active", "TASK-1")
	if !ok {
		t.Fatal("Task missed a task that exists")
	}
	if got.Entity.Title != "Alpha one" {
		t.Errorf("title = %q", got.Entity.Title)
	}
	// The full body travels, which is the whole point of the task panel.
	if body, present := got.Entity.Section("description"); !present ||
		!strings.Contains(body, "The body") {
		t.Errorf("description = %q, present = %v", body, present)
	}

	if _, ok := s.Task(alpha, "task", "active", "TASK-404"); ok {
		t.Error("Task reported a task that does not exist")
	}
}

// First run is not an error state, and it must not look like one.
func TestNoRegistryIsReportedAsFirstRun(t *testing.T) {
	s := startService(t, filepath.Join(t.TempDir(), "muster", "projects.yml"))

	problems := s.Problems()
	if len(problems) != 1 {
		t.Fatalf("got %d problems, want one: %+v", len(problems), problems)
	}
	if problems[0].Kind != app.ProblemNoRegistry {
		t.Errorf("kind = %q, want %q", problems[0].Kind, app.ProblemNoRegistry)
	}
	if problems[0].Title == "" || problems[0].Detail == "" {
		t.Error("a problem with no title or detail cannot be rendered")
	}
	if !strings.Contains(s.RegistryPath(), "muster") {
		t.Errorf("RegistryPath = %q, want it to name where the file belongs",
			s.RegistryPath())
	}
	if len(s.Tasks(app.QueryInput{})) != 0 {
		t.Error("want no tasks on first run")
	}
}

// A broken project is a structured problem the UI can render, not a crash and
// not a silent omission.
func TestBrokenProjectBecomesAStructuredProblem(t *testing.T) {
	good := newProject(t, "good", map[string]string{
		"task-1 - a.md": "---\nid: TASK-1\ntitle: Fine\nstatus: To Do\n---\n",
	})
	missing := filepath.Join(t.TempDir(), "gone")
	s := startService(t, withRegistry(t, missing, good))

	var found *app.Problem
	for i, p := range s.Problems() {
		if p.Kind == app.ProblemProject {
			found = &s.Problems()[i]
			break
		}
	}
	if found == nil {
		t.Fatalf("no project problem reported: %+v", s.Problems())
	}
	if found.Path != missing {
		t.Errorf("problem path = %q, want %q", found.Path, missing)
	}
	if found.Title == "" || found.Detail == "" {
		t.Error("a problem with no title or detail cannot be rendered")
	}
	// The healthy project still works.
	if n := len(s.Tasks(app.QueryInput{})); n != 1 {
		t.Errorf("got %d tasks, want the good project's task", n)
	}
}

// A file skipped during a scan reaches the frontend as a file-level problem
// rather than disappearing into a log.
func TestSkippedFileBecomesAProblem(t *testing.T) {
	proj := newProject(t, "alpha", map[string]string{
		"task-1 - a.md": "---\nid: TASK-1\ntitle: Fine\nstatus: To Do\n---\n",
		"README.md":     "# Notes\n",
	})
	s := startService(t, withRegistry(t, proj))

	var fileProblems int
	for _, p := range s.Problems() {
		if p.Kind == app.ProblemFile {
			fileProblems++
			if !strings.HasSuffix(p.Path, "README.md") {
				t.Errorf("problem names %q", p.Path)
			}
		}
	}
	if fileProblems != 1 {
		t.Errorf("got %d file problems, want 1: %+v", fileProblems, s.Problems())
	}
}

// The whole promise: a file changes on disk and the service reports the new
// state without anyone asking it to rescan.
func TestChangeOnDiskReachesTheService(t *testing.T) {
	proj := newProject(t, "alpha", map[string]string{
		"task-1 - a.md": "---\nid: TASK-1\ntitle: Alpha one\nstatus: To Do\n---\n",
	})
	s := startService(t, withRegistry(t, proj))

	if n := len(s.Tasks(app.QueryInput{})); n != 1 {
		t.Fatalf("got %d tasks at startup, want 1", n)
	}

	// An agent writes a second task, exactly as the CLI would.
	if err := os.WriteFile(
		filepath.Join(proj, "backlog", "tasks", "task-2 - b.md"),
		[]byte("---\nid: TASK-2\ntitle: Written by an agent\nstatus: To Do\n---\n"),
		0o644); err != nil {
		t.Fatalf("write: %v", err)
	}

	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		if len(s.Tasks(app.QueryInput{})) == 2 {
			return
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Fatalf("the new task never appeared; service still reports %d",
		len(s.Tasks(app.QueryInput{})))
}

func TestFilterValuesAndStatusLists(t *testing.T) {
	alpha := newProject(t, "alpha", map[string]string{
		"task-1 - a.md": "---\nid: TASK-1\ntitle: One\nstatus: To Do\n" +
			"labels: [backend]\nmilestone: m-1\n---\n",
	})
	s := startService(t, withRegistry(t, alpha))

	if got := s.FilterValues("label"); len(got) != 1 || got[0] != "backend" {
		t.Errorf("labels = %v", got)
	}
	if got := s.FilterValues("milestone"); len(got) != 1 || got[0] != "m-1" {
		t.Errorf("milestones = %v", got)
	}
	lists := s.StatusLists()
	if len(lists) != 1 || len(lists[0]) != 2 {
		t.Errorf("status lists = %v, want the project's own list", lists)
	}
}

func TestReloadPicksUpRegistryEdits(t *testing.T) {
	alpha := newProject(t, "alpha", map[string]string{
		"task-1 - a.md": "---\nid: TASK-1\ntitle: One\nstatus: To Do\n---\n",
	})
	beta := newProject(t, "beta", map[string]string{
		"task-1 - b.md": "---\nid: TASK-1\ntitle: Two\nstatus: To Do\n---\n",
	})
	regPath := withRegistry(t, alpha)
	s := startService(t, regPath)

	if n := len(s.Projects()); n != 1 {
		t.Fatalf("got %d projects, want 1", n)
	}

	if err := os.WriteFile(regPath,
		[]byte("projects:\n  - path: "+alpha+"\n  - path: "+beta+"\n"), 0o644); err != nil {
		t.Fatalf("rewrite registry: %v", err)
	}
	s.Reload()

	if n := len(s.Projects()); n != 2 {
		t.Errorf("got %d projects after reload, want 2", n)
	}
	if n := len(s.Tasks(app.QueryInput{})); n != 2 {
		t.Errorf("got %d tasks after reload, want 2", n)
	}
}
