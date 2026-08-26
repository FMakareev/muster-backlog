package watcher_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/FMakareev/muster-backlog/internal/registry"
	"github.com/FMakareev/muster-backlog/internal/store"
	"github.com/FMakareev/muster-backlog/internal/watcher"
)

// A file changing on disk must end with the store holding the new task, with
// no polling and nothing else rescanned. This is the whole promise of the
// watcher, so it is worth proving end to end rather than at the seam.
func TestChangeOnDiskReachesTheStore(t *testing.T) {
	root := t.TempDir()
	proj := filepath.Join(root, "alpha")
	tasks := filepath.Join(proj, "backlog", "tasks")
	if err := os.MkdirAll(tasks, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(proj, "backlog", "config.yml"),
		[]byte("project_name: Alpha\n"), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	regPath := filepath.Join(root, registry.FileName)
	if err := os.WriteFile(regPath,
		[]byte("projects:\n  - path: "+proj+"\n"), 0o644); err != nil {
		t.Fatalf("write registry: %v", err)
	}
	reg, err := registry.LoadFrom(regPath)
	if err != nil {
		t.Fatalf("LoadFrom: %v", err)
	}

	s := store.New()
	s.Load(reg)
	if n := s.Count(store.Query{}); n != 0 {
		t.Fatalf("store starts with %d tasks, want none", n)
	}

	reloaded := make(chan string, 8)
	var projects []watcher.Project
	for _, p := range s.Projects() {
		projects = append(projects, watcher.Project{
			Path: p.Registry.Path, DataDir: p.Registry.Location.DataDir,
		})
	}
	w, err := watcher.New(projects, watcher.Options{
		Debounce: debounce,
		OnChange: func(path string) {
			s.Reload(path)
			reloaded <- path
		},
		OnError: func(err error) { t.Logf("watcher error: %v", err) },
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	defer func() {
		if err := w.Close(); err != nil {
			t.Errorf("Close: %v", err)
		}
	}()

	// An agent writes a task, exactly as the CLI would.
	if err := os.WriteFile(filepath.Join(tasks, "task-1 - new.md"),
		[]byte("---\nid: TASK-1\ntitle: Written by an agent\nstatus: To Do\n---\n"),
		0o644); err != nil {
		t.Fatalf("write task: %v", err)
	}

	select {
	case <-reloaded:
	case <-time.After(settle):
		t.Fatal("the store was never reloaded")
	}

	items := s.Query(store.Query{})
	if len(items) != 1 {
		t.Fatalf("store holds %d tasks, want the new one", len(items))
	}
	if got := items[0].Entity.Title; got != "Written by an agent" {
		t.Errorf("title = %q", got)
	}
}
