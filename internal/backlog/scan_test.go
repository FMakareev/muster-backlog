package backlog_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/FMakareev/muster-backlog/internal/backlog"
	"github.com/FMakareev/muster-backlog/internal/project"
)

// The hidden-layout fixture is a whole project: a .backlog directory holding
// the same id in tasks/ and archive/tasks/, which is a real occurrence in the
// author's projects and the reason a task key includes its directory class.
func TestScanHiddenLayoutFixture(t *testing.T) {
	loc, err := project.Discover(filepath.Join(corpus, "04-layout-dot-backlog"))
	if err != nil {
		t.Fatalf("Discover: %v", err)
	}

	p, err := backlog.Scan(loc)
	if err != nil {
		t.Fatalf("Scan: %v", err)
	}
	if len(p.Tasks) != 2 {
		t.Fatalf("got %d tasks, want the active one and the archived one", len(p.Tasks))
	}

	byClass := map[backlog.Class]backlog.Entity{}
	for _, task := range p.Tasks {
		byClass[task.Class] = task
	}
	active, hasActive := byClass[backlog.ClassActive]
	archived, hasArchived := byClass[backlog.ClassArchived]
	if !hasActive || !hasArchived {
		t.Fatalf("want one active and one archived task, got %+v", byClass)
	}
	if active.ID != archived.ID {
		t.Fatalf("fixture should hold the same id twice, got %q and %q",
			active.ID, archived.ID)
	}
	if active.Title == archived.Title {
		t.Error("the two tasks sharing an id should be different tasks")
	}
	// Same id, different entities: only the class tells them apart.
	if active.Key == archived.Key {
		t.Error("keys collide; the directory class is not part of identity")
	}
}

func newScannableProject(t *testing.T) (string, project.Location) {
	t.Helper()
	root := t.TempDir()
	for _, dir := range []string{
		"tasks", "drafts", "milestones", "docs", "decisions", "completed",
		filepath.Join("archive", "tasks"),
	} {
		if err := os.MkdirAll(filepath.Join(root, "backlog", dir), 0o755); err != nil {
			t.Fatalf("mkdir: %v", err)
		}
	}
	write(t, filepath.Join(root, "backlog", "config.yml"),
		"project_name: Sample\nstatuses: [\"To Do\", \"Done\"]\n")

	loc, err := project.Discover(root)
	if err != nil {
		t.Fatalf("Discover: %v", err)
	}
	return root, loc
}

func write(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

func TestScanReadsEveryEntityDirectory(t *testing.T) {
	root, loc := newScannableProject(t)
	dir := filepath.Join(root, "backlog")

	write(t, filepath.Join(dir, "tasks", "task-1 - A.md"),
		"---\nid: TASK-1\ntitle: A task\nstatus: To Do\n---\n\n## Description\n\n<!-- SECTION:DESCRIPTION:BEGIN -->\nBody.\n<!-- SECTION:DESCRIPTION:END -->\n")
	write(t, filepath.Join(dir, "drafts", "draft-1 - B.md"),
		"---\nid: DRAFT-1\ntitle: A draft\n---\n\nIdea.\n")
	write(t, filepath.Join(dir, "milestones", "m-1 - C.md"),
		"---\nid: m-1\ntitle: A milestone\n---\n\n## Description\n\nPlain prose.\n")
	write(t, filepath.Join(dir, "docs", "doc-1 - D.md"),
		"---\nid: doc-1\ntitle: A document\ntags: [x]\n---\n\nFree markdown.\n")
	write(t, filepath.Join(dir, "decisions", "decision-1 - E.md"),
		"---\nid: decision-1\ntitle: A decision\nstatus: accepted\n---\n\n## Context\n\nWhy.\n")
	write(t, filepath.Join(dir, "completed", "task-2 - F.md"),
		"---\nid: TASK-2\ntitle: A finished task\nstatus: Done\n---\n\nDone.\n")
	write(t, filepath.Join(dir, "archive", "tasks", "task-3 - G.md"),
		"---\nid: TASK-3\ntitle: An archived task\n---\n\nGone.\n")

	p, err := backlog.Scan(loc)
	if err != nil {
		t.Fatalf("Scan: %v", err)
	}
	if got := len(p.Tasks); got != 3 {
		t.Errorf("tasks = %d, want active, completed and archived", got)
	}
	if len(p.Drafts) != 1 || len(p.Milestones) != 1 ||
		len(p.Documents) != 1 || len(p.Decisions) != 1 {
		t.Errorf("drafts=%d milestones=%d docs=%d decisions=%d, want one of each",
			len(p.Drafts), len(p.Milestones), len(p.Documents), len(p.Decisions))
	}
	if p.Config.ProjectName != "Sample" {
		t.Errorf("ProjectName = %q", p.Config.ProjectName)
	}
	if len(p.Diagnostics) != 0 {
		t.Errorf("unexpected diagnostics: %+v", p.Diagnostics)
	}
	// Milestones, documents and decisions have no marker grammar at all.
	if len(p.Milestones[0].Sections) != 0 || len(p.Documents[0].Sections) != 0 {
		t.Error("marker grammar was applied to a non-task entity")
	}
}

func TestScanWalksNestedDocumentDirectories(t *testing.T) {
	root, loc := newScannableProject(t)
	nested := filepath.Join(root, "backlog", "docs", "guides")
	if err := os.MkdirAll(nested, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	write(t, filepath.Join(nested, "doc-2 - Nested.md"),
		"---\nid: doc-2\ntitle: Nested document\n---\n\nContent.\n")

	p, err := backlog.Scan(loc)
	if err != nil {
		t.Fatalf("Scan: %v", err)
	}
	if len(p.Documents) != 1 {
		t.Fatalf("got %d documents, want the nested one found", len(p.Documents))
	}
}

// A README or an editor leftover in a task directory is normal. It must cost a
// diagnostic, never the rest of the project.
func TestNonEntityFileBecomesADiagnostic(t *testing.T) {
	root, loc := newScannableProject(t)
	dir := filepath.Join(root, "backlog", "tasks")

	write(t, filepath.Join(dir, "README.md"), "# Notes about these tasks\n")
	write(t, filepath.Join(dir, ".gitkeep"), "")
	write(t, filepath.Join(dir, "task-1 - Good.md"),
		"---\nid: TASK-1\ntitle: A good task\n---\n\nBody.\n")

	p, err := backlog.Scan(loc)
	if err != nil {
		t.Fatalf("Scan: %v", err)
	}
	if len(p.Tasks) != 1 {
		t.Fatalf("got %d tasks, want the one valid task", len(p.Tasks))
	}
	if len(p.Diagnostics) != 1 {
		t.Fatalf("diagnostics = %+v, want exactly one for the README", p.Diagnostics)
	}
	if !strings.HasSuffix(p.Diagnostics[0].Path, "README.md") {
		t.Errorf("diagnostic names %q, want the README", p.Diagnostics[0].Path)
	}
	if p.Diagnostics[0].Reason == "" {
		t.Error("a diagnostic with no reason helps nobody")
	}
}

// Ordinals are optional and not unique, so both fallbacks in the comparator
// carry real weight.
func TestOrderingMatchesTheCLIComparator(t *testing.T) {
	root, loc := newScannableProject(t)
	dir := filepath.Join(root, "backlog", "tasks")

	write(t, filepath.Join(dir, "task-10 - J.md"),
		"---\nid: TASK-10\ntitle: Ten\nordinal: 1000\n---\n")
	write(t, filepath.Join(dir, "task-9 - I.md"),
		"---\nid: TASK-9\ntitle: Nine\nordinal: 1000\n---\n")
	write(t, filepath.Join(dir, "task-2 - B.md"),
		"---\nid: TASK-2\ntitle: Two\nordinal: 5\n---\n")
	write(t, filepath.Join(dir, "task-1 - A.md"),
		"---\nid: TASK-1\ntitle: One\n---\n")

	p, err := backlog.Scan(loc)
	if err != nil {
		t.Fatalf("Scan: %v", err)
	}
	var order []string
	for _, task := range p.Tasks {
		order = append(order, task.ID)
	}
	want := "TASK-2|TASK-9|TASK-10|TASK-1"
	if got := strings.Join(order, "|"); got != want {
		t.Errorf("order = %q, want %q (ordinal first, colliding ordinals broken "+
			"by numeric id, no ordinal last)", got, want)
	}
}

func TestScanFailsOnlyWhenTheProjectItselfCannotBeRead(t *testing.T) {
	_, err := backlog.Scan(project.Location{
		Root:       t.TempDir(),
		DataDir:    t.TempDir(),
		ConfigPath: filepath.Join(t.TempDir(), "config.yml"),
	})
	if err == nil {
		t.Fatal("want an error when the config cannot be read")
	}
}
