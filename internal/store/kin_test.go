package store_test

import (
	"path/filepath"
	"testing"

	"github.com/FMakareev/muster-backlog/internal/backlog"
	"github.com/FMakareev/muster-backlog/internal/registry"
	"github.com/FMakareev/muster-backlog/internal/store"
)

// task writes one task file into a class directory, with an optional parent.
func task(t *testing.T, dir, class, id, title, status, parent string) {
	t.Helper()
	body := "---\nid: " + id + "\ntitle: " + title + "\nstatus: " + status + "\n"
	if parent != "" {
		body += "parent_task_id: " + parent + "\n"
	}
	// subtasks is in the serialiser and written empty in every real file, so
	// the fixtures write it empty too: the forward direction must come from
	// the back-links, not from here.
	body += "subtasks: []\n---\n\n## Description\n\n" +
		"<!-- SECTION:DESCRIPTION:BEGIN -->\nx\n<!-- SECTION:DESCRIPTION:END -->\n"
	write(t, filepath.Join(dir, "backlog", class, id+" - t.md"), body)
}

// kinFixture builds one project whose relationships cover every measured case.
func kinFixture(t *testing.T) (*store.Store, string) {
	t.Helper()
	root := t.TempDir()
	dir := filepath.Join(root, "solo")
	write(t, filepath.Join(dir, "backlog", "config.yml"),
		"project_name: \"Solo\"\nstatuses: [\"To Do\", \"In Review\", \"Shipped\"]\n")

	task(t, dir, "tasks", "TASK-1", "A parent", "In Review", "")
	task(t, dir, "tasks", "TASK-1.1", "Open child", "To Do", "TASK-1")
	task(t, dir, "tasks", "TASK-1.2", "Finished child", "Shipped", "TASK-1")
	// Lowercased reference to an uppercase id: files write both, and neither
	// spelling may decide whether the link exists.
	task(t, dir, "tasks", "TASK-1.3", "Case child", "To Do", "task-1")
	task(t, dir, "completed", "TASK-1.4", "Completed child", "In Review", "TASK-1")
	// Archiving is a soft delete: absent from the board, so absent from both
	// halves of the count.
	task(t, dir, filepath.Join("archive", "tasks"), "TASK-1.5", "Gone", "To Do", "TASK-1")

	// A parent that has itself been completed while its subtask has not.
	task(t, dir, "completed", "TASK-2", "Completed parent", "Shipped", "")
	task(t, dir, "tasks", "TASK-2.1", "Child of a completed parent", "To Do", "TASK-2")

	// A parent nothing answers to.
	task(t, dir, "tasks", "TASK-3.1", "Orphan", "To Do", "TASK-99")

	// No relationship at all.
	task(t, dir, "tasks", "TASK-4", "Alone", "To Do", "")

	regFile := filepath.Join(root, "projects.yml")
	write(t, regFile, "projects:\n  - path: "+dir+"\n    name: Solo\n")
	reg, err := registry.LoadFrom(regFile)
	if err != nil {
		t.Fatalf("LoadFrom: %v", err)
	}
	s := store.New()
	s.Load(reg)
	return s, dir
}

func ref(project, class, id string) store.Ref {
	return store.Ref{
		Project: project,
		Key: backlog.Key{
			Kind: backlog.KindTask, Class: backlog.Class(class), ID: id,
		},
	}
}

func TestKinResolvesChildrenFromBackLinks(t *testing.T) {
	s, dir := kinFixture(t)
	kin := s.KinIndex()

	parent := kin[ref(dir, "active", "TASK-1")]
	if len(parent.Children) != 4 {
		t.Fatalf("got %d children, want 4 (three active, one completed, "+
			"the archived one excluded)", len(parent.Children))
	}
	// Two finished: the one at the project's own last status, and the one the
	// CLI moved into completed/.
	if parent.Done != 2 {
		t.Errorf("got Done %d, want 2", parent.Done)
	}

	for _, id := range []string{"TASK-1.1", "TASK-1.2", "TASK-1.3"} {
		child := kin[ref(dir, "active", id)]
		if child.Parent == nil {
			t.Fatalf("%s has no parent", id)
		}
		if child.Parent.ID != "TASK-1" {
			t.Errorf("%s points at %s, want TASK-1", id, child.Parent.ID)
		}
		if child.ParentTitle != "A parent" {
			t.Errorf("%s carries title %q", id, child.ParentTitle)
		}
	}
}

// The archived subtask keeps its own link upwards - it is soft-deleted, not
// severed - but does not count towards its parent's progress.
func TestKinExcludesArchivedChildrenFromCounts(t *testing.T) {
	s, dir := kinFixture(t)
	kin := s.KinIndex()

	archived := kin[ref(dir, "archived", "TASK-1.5")]
	if archived.Parent == nil || archived.Parent.ID != "TASK-1" {
		t.Fatal("an archived subtask should still know its parent")
	}
	for _, child := range kin[ref(dir, "active", "TASK-1")].Children {
		if child.ID == "TASK-1.5" {
			t.Error("archived subtask counted among a parent's children")
		}
	}
}

// A parent can be completed while its subtasks are not: 6 of the 92 real
// links cross directories this way.
func TestKinResolvesAParentInAnotherClass(t *testing.T) {
	s, dir := kinFixture(t)
	kin := s.KinIndex()

	child := kin[ref(dir, "active", "TASK-2.1")]
	if child.Parent == nil {
		t.Fatal("child of a completed parent found no parent")
	}
	if child.Parent.Class != backlog.ClassCompleted {
		t.Errorf("parent resolved to class %q, want completed", child.Parent.Class)
	}
	if len(kin[ref(dir, "completed", "TASK-2")].Children) != 1 {
		t.Error("a completed parent should still list its subtasks")
	}
}

func TestKinLeavesAMissingParentUnresolved(t *testing.T) {
	s, dir := kinFixture(t)
	kin := s.KinIndex()

	orphan, ok := kin[ref(dir, "active", "TASK-3.1")]
	if !ok {
		t.Fatal("a task declaring a parent should appear in the index")
	}
	if orphan.Parent != nil {
		t.Errorf("resolved a parent that does not exist: %v", orphan.Parent)
	}
}

func TestKinSkipsTasksWithNoRelationship(t *testing.T) {
	s, dir := kinFixture(t)
	kin := s.KinIndex()

	if _, ok := kin[ref(dir, "active", "TASK-4")]; ok {
		t.Error("a task with neither parent nor subtasks should be absent")
	}
}

// Ids collide across projects freely, so a link must never reach out of its
// own project to find a parent.
func TestKinDoesNotLinkAcrossProjects(t *testing.T) {
	f := newFixture(t)
	alpha, beta := f.paths["alpha"], f.paths["beta"]
	// beta's TASK-3 declares alpha's TASK-1 as a parent; alpha has one.
	task(t, beta, "tasks", "TASK-1.1", "Beta child", "To Do", "TASK-1")
	f.store.Reload(beta)

	kin := f.store.KinIndex()
	child := kin[ref(beta, "active", "TASK-1.1")]
	if child.Parent == nil {
		t.Fatal("beta's subtask should find beta's own TASK-1")
	}
	if child.Parent.Project != beta {
		t.Errorf("parent resolved into %s, want %s", child.Parent.Project, beta)
	}
	if len(kin[ref(alpha, "active", "TASK-1")].Children) != 0 {
		t.Error("alpha's TASK-1 adopted a subtask from another project")
	}
}
