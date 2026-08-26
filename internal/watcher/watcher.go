// Package watcher turns filesystem changes into per-project reload signals.
//
// Liveness is the reason Muster is an application rather than a report: an
// agent changes a task's status in a repository and the board follows without
// anyone pressing anything.
package watcher

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
)

// DefaultDebounce is how long a project stays quiet before it is reloaded.
//
// A single CLI write touches a file several times within milliseconds, and a
// git branch switch rewrites hundreds of files at once. Both have to collapse
// into one reload, while still feeling immediate to someone watching the board.
const DefaultDebounce = 150 * time.Millisecond

// watchedDirs are the entity directories of a project.
//
// All of them are watched, not only tasks: drafts feed the inbox, documents and
// decisions feed the viewer, and the archive subtree changes whenever a task is
// archived. A watcher that covered only tasks/ would make every other screen
// quietly stale.
var watchedDirs = []string{
	"", // the data directory itself, so config.yml changes are noticed
	"tasks",
	"drafts",
	"milestones",
	"docs",
	"decisions",
	"completed",
	"archive",
	filepath.Join("archive", "tasks"),
	filepath.Join("archive", "drafts"),
	filepath.Join("archive", "milestones"),
}

// Project is one project to watch.
type Project struct {
	// Path is the registered project path, used as the identity handed back to
	// the callback.
	Path string
	// DataDir is the resolved Backlog.md data directory.
	DataDir string
}

// Watcher reports which project changed, debounced per project.
//
// It holds no store and applies no reload policy of its own: it says what
// changed, and the caller decides what that means.
type Watcher struct {
	debounce time.Duration
	onChange func(projectPath string)
	onError  func(error)

	fsw *fsnotify.Watcher

	mu sync.Mutex
	// dirs maps a watched directory to the project it belongs to.
	dirs map[string]string
	// projects maps a project path to its data directory.
	projects map[string]string
	// timers holds the in-flight debounce per project.
	timers map[string]*time.Timer
	closed bool

	done chan struct{}
	wg   sync.WaitGroup
}

// Options configure a Watcher.
type Options struct {
	// Debounce defaults to DefaultDebounce.
	Debounce time.Duration
	// OnChange is called with a project path once that project has been quiet
	// for the debounce interval. It runs on the watcher's own goroutine.
	OnChange func(projectPath string)
	// OnError receives watcher-level problems: a directory that vanished, an
	// fsnotify failure. It must never be used to report normal churn.
	OnError func(error)
}

// New starts a watcher over the given projects.
//
// A project whose directories cannot be watched is reported through OnError and
// skipped; it never prevents the others from being watched.
func New(projects []Project, opts Options) (*Watcher, error) {
	fsw, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("creating watcher: %w", err)
	}

	w := &Watcher{
		debounce: opts.Debounce,
		onChange: opts.OnChange,
		onError:  opts.OnError,
		fsw:      fsw,
		dirs:     map[string]string{},
		projects: map[string]string{},
		timers:   map[string]*time.Timer{},
		done:     make(chan struct{}),
	}
	if w.debounce <= 0 {
		w.debounce = DefaultDebounce
	}
	if w.onChange == nil {
		w.onChange = func(string) {}
	}
	if w.onError == nil {
		w.onError = func(error) {}
	}

	for _, p := range projects {
		w.addProject(p)
	}

	w.wg.Add(1)
	go w.run()
	return w, nil
}

// addProject registers a project's directories.
func (w *Watcher) addProject(p Project) {
	w.mu.Lock()
	w.projects[p.Path] = p.DataDir
	w.mu.Unlock()

	for _, sub := range watchedDirs {
		w.addDir(filepath.Join(p.DataDir, sub), p.Path)
	}
}

// addDir starts watching one directory, if it exists.
//
// A missing directory is ordinary - not every project has every one - and is
// silently skipped rather than reported.
func (w *Watcher) addDir(dir, projectPath string) {
	info, err := os.Stat(dir)
	if err != nil || !info.IsDir() {
		return
	}

	w.mu.Lock()
	if w.closed {
		w.mu.Unlock()
		return
	}
	_, already := w.dirs[dir]
	w.mu.Unlock()
	if already {
		return
	}

	if err := w.fsw.Add(dir); err != nil {
		w.onError(fmt.Errorf("watching %s: %w", dir, err))
		return
	}
	w.mu.Lock()
	w.dirs[dir] = projectPath
	w.mu.Unlock()
}

// run is the event loop.
func (w *Watcher) run() {
	defer w.wg.Done()
	for {
		select {
		case <-w.done:
			return
		case event, ok := <-w.fsw.Events:
			if !ok {
				return
			}
			w.handle(event)
		case err, ok := <-w.fsw.Errors:
			if !ok {
				return
			}
			w.onError(err)
		}
	}
}

func (w *Watcher) handle(event fsnotify.Event) {
	dir := filepath.Dir(event.Name)

	w.mu.Lock()
	projectPath, known := w.dirs[dir]
	if !known {
		// The event may be about a watched directory itself rather than a file
		// inside it - a rename or removal of the directory.
		projectPath, known = w.dirs[event.Name]
	}
	w.mu.Unlock()
	if !known {
		return
	}

	// A newly created directory has to be watched in its own right: fsnotify is
	// not recursive, so a docs subdirectory or a recreated archive/ would
	// otherwise be invisible.
	if event.Has(fsnotify.Create) {
		if info, err := os.Stat(event.Name); err == nil && info.IsDir() {
			w.addDir(event.Name, projectPath)
		}
	}

	// fsnotify watches inodes. When git or an editor replaces a directory
	// rather than writing into it, the old watch stops delivering events
	// silently. Re-resolving the project's directories on every change means a
	// dead watch is restored by the next event anywhere in the project, rather
	// than being lost until restart.
	if event.Has(fsnotify.Remove) || event.Has(fsnotify.Rename) {
		w.rewatch(projectPath)
	}

	w.schedule(projectPath)
}

// rewatch re-adds any of a project's directories that are no longer watched.
func (w *Watcher) rewatch(projectPath string) {
	w.mu.Lock()
	dataDir, known := w.projects[projectPath]
	// Forget directories that have gone, so a later recreation is treated as
	// new rather than as already-watched.
	for dir := range w.dirs {
		if w.dirs[dir] != projectPath {
			continue
		}
		if info, err := os.Stat(dir); err != nil || !info.IsDir() {
			delete(w.dirs, dir)
			_ = w.fsw.Remove(dir)
		}
	}
	w.mu.Unlock()
	if !known {
		return
	}

	for _, sub := range watchedDirs {
		w.addDir(filepath.Join(dataDir, sub), projectPath)
	}
}

// schedule collapses a burst into one reload per project.
func (w *Watcher) schedule(projectPath string) {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.closed {
		return
	}

	if timer, ok := w.timers[projectPath]; ok {
		timer.Reset(w.debounce)
		return
	}
	w.timers[projectPath] = time.AfterFunc(w.debounce, func() {
		w.mu.Lock()
		delete(w.timers, projectPath)
		closed := w.closed
		w.mu.Unlock()
		if closed {
			return
		}
		w.onChange(projectPath)
	})
}

// Close stops the watcher and releases every descriptor.
//
// It is idempotent and synchronous: once it returns, no further callback will
// fire and no goroutine of the watcher's remains. This runs for a whole working
// day, so leaks here are not academic.
func (w *Watcher) Close() error {
	w.mu.Lock()
	if w.closed {
		w.mu.Unlock()
		return nil
	}
	w.closed = true
	for path, timer := range w.timers {
		timer.Stop()
		delete(w.timers, path)
	}
	w.mu.Unlock()

	close(w.done)
	err := w.fsw.Close()
	w.wg.Wait()

	if err != nil && !errors.Is(err, os.ErrClosed) {
		return fmt.Errorf("closing watcher: %w", err)
	}
	return nil
}

// WatchedDirs returns the directories currently being watched, for diagnostics.
func (w *Watcher) WatchedDirs() []string {
	w.mu.Lock()
	defer w.mu.Unlock()

	out := make([]string, 0, len(w.dirs))
	for dir := range w.dirs {
		out = append(out, dir)
	}
	return out
}
