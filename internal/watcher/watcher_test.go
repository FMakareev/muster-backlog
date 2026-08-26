package watcher_test

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/FMakareev/muster-backlog/internal/watcher"
)

const (
	debounce = 40 * time.Millisecond
	// settle is how long a test waits for a debounced callback: comfortably
	// more than the debounce, but short enough to keep the suite quick.
	settle = 400 * time.Millisecond
)

// recorder collects change callbacks.
type recorder struct {
	mu      sync.Mutex
	calls   []string
	changed chan string
}

func newRecorder() *recorder {
	return &recorder{changed: make(chan string, 64)}
}

func (r *recorder) onChange(project string) {
	r.mu.Lock()
	r.calls = append(r.calls, project)
	r.mu.Unlock()
	select {
	case r.changed <- project:
	default:
	}
}

func (r *recorder) count() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return len(r.calls)
}

// await waits for one change callback naming project.
func (r *recorder) await(t *testing.T, project string) {
	t.Helper()
	deadline := time.After(settle)
	for {
		select {
		case got := <-r.changed:
			if got == project {
				return
			}
		case <-deadline:
			t.Fatalf("no change reported for %s within %s", project, settle)
		}
	}
}

// newProject creates a project data directory with the usual layout.
func newProject(t *testing.T, name string) watcher.Project {
	t.Helper()
	root := t.TempDir()
	data := filepath.Join(root, "backlog")
	for _, sub := range []string{
		"tasks", "drafts", "milestones", "docs", "decisions", "completed",
		filepath.Join("archive", "tasks"),
	} {
		if err := os.MkdirAll(filepath.Join(data, sub), 0o755); err != nil {
			t.Fatalf("mkdir: %v", err)
		}
	}
	return watcher.Project{Path: root, DataDir: data}
}

func start(t *testing.T, projects []watcher.Project) (*watcher.Watcher, *recorder) {
	t.Helper()
	rec := newRecorder()
	w, err := watcher.New(projects, watcher.Options{
		Debounce: debounce,
		OnChange: rec.onChange,
		OnError:  func(err error) { t.Logf("watcher error: %v", err) },
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	t.Cleanup(func() {
		if err := w.Close(); err != nil {
			t.Errorf("Close: %v", err)
		}
	})
	return w, rec
}

func writeTask(t *testing.T, dir, name string) {
	t.Helper()
	path := filepath.Join(dir, name)
	content := "---\nid: TASK-1\ntitle: Something\n---\n"
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
}

// Every entity directory is watched, not only tasks: drafts feed the inbox and
// documents feed the viewer, and both must stay live.
func TestEveryEntityDirectoryIsWatched(t *testing.T) {
	p := newProject(t, "alpha")
	_, rec := start(t, []watcher.Project{p})

	for _, sub := range []string{
		"tasks", "drafts", "milestones", "docs", "decisions", "completed",
		filepath.Join("archive", "tasks"),
	} {
		t.Run(sub, func(t *testing.T) {
			writeTask(t, filepath.Join(p.DataDir, sub), "task-1 - x.md")
			rec.await(t, p.Path)
		})
	}
}

func TestCreateWriteRenameDeleteAllReport(t *testing.T) {
	p := newProject(t, "alpha")
	_, rec := start(t, []watcher.Project{p})
	tasks := filepath.Join(p.DataDir, "tasks")

	path := filepath.Join(tasks, "task-1 - x.md")
	writeTask(t, tasks, "task-1 - x.md")
	rec.await(t, p.Path)

	if err := os.WriteFile(path, []byte("---\nid: TASK-1\ntitle: Edited\n---\n"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	rec.await(t, p.Path)

	renamed := filepath.Join(tasks, "task-2 - y.md")
	if err := os.Rename(path, renamed); err != nil {
		t.Fatalf("rename: %v", err)
	}
	rec.await(t, p.Path)

	if err := os.Remove(renamed); err != nil {
		t.Fatalf("remove: %v", err)
	}
	rec.await(t, p.Path)
}

// A CLI write touches a file several times in milliseconds; a branch switch
// rewrites hundreds. Both must collapse into one reload.
func TestBurstsAreDebouncedIntoOneReload(t *testing.T) {
	p := newProject(t, "alpha")
	_, rec := start(t, []watcher.Project{p})
	tasks := filepath.Join(p.DataDir, "tasks")

	for i := 0; i < 50; i++ {
		writeTask(t, tasks, "task-"+string(rune('a'+i%26))+" - x.md")
	}
	rec.await(t, p.Path)
	time.Sleep(settle)

	if got := rec.count(); got > 3 {
		t.Errorf("got %d reloads for one burst of 50 writes, want it collapsed", got)
	}
}

// A burst in one project must not delay another, so debouncing is per project.
func TestDebounceIsPerProject(t *testing.T) {
	alpha := newProject(t, "alpha")
	beta := newProject(t, "beta")
	_, rec := start(t, []watcher.Project{alpha, beta})

	writeTask(t, filepath.Join(alpha.DataDir, "tasks"), "task-1 - a.md")
	writeTask(t, filepath.Join(beta.DataDir, "tasks"), "task-1 - b.md")

	seen := map[string]bool{}
	deadline := time.After(settle)
	for len(seen) < 2 {
		select {
		case got := <-rec.changed:
			seen[got] = true
		case <-deadline:
			t.Fatalf("only %v reported", seen)
		}
	}
}

// Only the project that changed is reported; the other stays silent.
func TestOnlyTheAffectedProjectIsReported(t *testing.T) {
	alpha := newProject(t, "alpha")
	beta := newProject(t, "beta")
	_, rec := start(t, []watcher.Project{alpha, beta})

	writeTask(t, filepath.Join(alpha.DataDir, "tasks"), "task-1 - a.md")
	rec.await(t, alpha.Path)
	time.Sleep(settle)

	rec.mu.Lock()
	calls := append([]string(nil), rec.calls...)
	rec.mu.Unlock()
	for _, got := range calls {
		if got != alpha.Path {
			t.Fatalf("project %s was reported, want only %s", got, alpha.Path)
		}
	}
}

// Editors save by writing a temporary file and renaming it over the target.
func TestAtomicSavePatternIsSeen(t *testing.T) {
	p := newProject(t, "alpha")
	_, rec := start(t, []watcher.Project{p})
	tasks := filepath.Join(p.DataDir, "tasks")

	target := filepath.Join(tasks, "task-1 - x.md")
	writeTask(t, tasks, "task-1 - x.md")
	rec.await(t, p.Path)
	time.Sleep(settle)

	tmp := filepath.Join(tasks, ".task-1.md.swp")
	if err := os.WriteFile(tmp, []byte("---\nid: TASK-1\ntitle: Saved\n---\n"), 0o644); err != nil {
		t.Fatalf("write temp: %v", err)
	}
	if err := os.Rename(tmp, target); err != nil {
		t.Fatalf("rename over target: %v", err)
	}
	rec.await(t, p.Path)
}

// fsnotify watches inodes, so a directory replaced wholesale - by git, or by a
// tool that rebuilds it - silently stops delivering events unless the watch is
// re-established.
func TestWatchSurvivesDirectoryRecreation(t *testing.T) {
	p := newProject(t, "alpha")
	_, rec := start(t, []watcher.Project{p})
	tasks := filepath.Join(p.DataDir, "tasks")

	writeTask(t, tasks, "task-1 - x.md")
	rec.await(t, p.Path)

	if err := os.RemoveAll(tasks); err != nil {
		t.Fatalf("remove directory: %v", err)
	}
	rec.await(t, p.Path)

	if err := os.MkdirAll(tasks, 0o755); err != nil {
		t.Fatalf("recreate: %v", err)
	}
	// Give the watcher a moment to re-establish the watch.
	time.Sleep(settle)

	writeTask(t, tasks, "task-2 - y.md")
	rec.await(t, p.Path)
}

// A docs subdirectory created after startup must become watched in its own
// right, because fsnotify is not recursive.
func TestNewSubdirectoryBecomesWatched(t *testing.T) {
	p := newProject(t, "alpha")
	_, rec := start(t, []watcher.Project{p})

	nested := filepath.Join(p.DataDir, "docs", "guides")
	if err := os.MkdirAll(nested, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	rec.await(t, p.Path)
	time.Sleep(settle)

	writeTask(t, nested, "doc-1 - x.md")
	rec.await(t, p.Path)
}

// One project's directory vanishing must not take the others down.
func TestVanishedProjectDoesNotStopOthers(t *testing.T) {
	alpha := newProject(t, "alpha")
	beta := newProject(t, "beta")
	_, rec := start(t, []watcher.Project{alpha, beta})

	if err := os.RemoveAll(alpha.DataDir); err != nil {
		t.Fatalf("remove: %v", err)
	}
	time.Sleep(settle)

	writeTask(t, filepath.Join(beta.DataDir, "tasks"), "task-1 - b.md")
	rec.await(t, beta.Path)
}

// A project that does not exist at all is skipped, not fatal.
func TestMissingProjectIsSkipped(t *testing.T) {
	good := newProject(t, "good")
	missing := watcher.Project{
		Path:    filepath.Join(t.TempDir(), "gone"),
		DataDir: filepath.Join(t.TempDir(), "gone", "backlog"),
	}
	w, rec := start(t, []watcher.Project{missing, good})

	for _, dir := range w.WatchedDirs() {
		if strings.Contains(dir, "gone") {
			t.Errorf("watching a directory that does not exist: %s", dir)
		}
	}
	writeTask(t, filepath.Join(good.DataDir, "tasks"), "task-1 - g.md")
	rec.await(t, good.Path)
}

// This runs for a whole working day; a leak here is not academic.
func TestCloseIsCleanAndIdempotent(t *testing.T) {
	before := runtime.NumGoroutine()

	p := newProject(t, "alpha")
	rec := newRecorder()
	w, err := watcher.New([]watcher.Project{p}, watcher.Options{
		Debounce: debounce,
		OnChange: rec.onChange,
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	writeTask(t, filepath.Join(p.DataDir, "tasks"), "task-1 - x.md")
	rec.await(t, p.Path)

	// Close while a debounce timer is in flight.
	writeTask(t, filepath.Join(p.DataDir, "tasks"), "task-2 - y.md")
	if err := w.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}
	if err := w.Close(); err != nil {
		t.Fatalf("second Close: %v", err)
	}

	countBefore := rec.count()
	time.Sleep(settle)
	if got := rec.count(); got != countBefore {
		t.Errorf("a callback fired after Close: %d then %d", countBefore, got)
	}

	// Goroutines wind down asynchronously in the runtime; give them a moment.
	for i := 0; i < 20; i++ {
		if runtime.NumGoroutine() <= before+1 {
			return
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Errorf("goroutines leaked: %d before, %d after", before, runtime.NumGoroutine())
}
