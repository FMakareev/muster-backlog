package store_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/FMakareev/muster-backlog/internal/registry"
	"github.com/FMakareev/muster-backlog/internal/store"
)

// smokeEnv names a colon-separated list of real project roots.
//
// Opt-in, because those projects live on one machine and must never be a
// prerequisite for a green build anywhere else.
const smokeEnv = "MUSTER_SMOKE_PROJECTS"

// registryOverSmokeProjects writes a registry covering the smoke projects.
func registryOverSmokeProjects(t testing.TB) (registry.Registry, int) {
	t.Helper()
	raw := os.Getenv(smokeEnv)
	if raw == "" {
		t.Skipf("set %s to a colon-separated list of project roots to run this", smokeEnv)
	}

	var lines []string
	var count int
	for _, root := range strings.Split(raw, ":") {
		if root = strings.TrimSpace(root); root != "" {
			lines = append(lines, "  - path: "+root)
			count++
		}
	}
	path := filepath.Join(t.TempDir(), registry.FileName)
	content := "projects:\n" + strings.Join(lines, "\n") + "\n"
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write registry: %v", err)
	}

	reg, err := registry.LoadFrom(path)
	if err != nil {
		t.Fatalf("LoadFrom: %v", err)
	}
	return reg, count
}

// TestLoadRealCorpusWithinBudget is the documented startup budget.
//
// The whole point of holding everything in memory is that a full load is cheap
// enough to do at startup and again on demand. If this ever fails, the answer
// is not to raise the number.
func TestLoadRealCorpusWithinBudget(t *testing.T) {
	const budget = 2 * time.Second

	reg, projects := registryOverSmokeProjects(t)
	s := store.New()

	start := time.Now()
	s.Load(reg)
	elapsed := time.Since(start)

	loaded := 0
	for _, p := range s.Projects() {
		if p.OK() {
			loaded++
		} else {
			t.Errorf("%s failed to load: %v", p.Registry.DisplayName, p.Err)
		}
	}
	tasks := s.Count(store.Query{})

	t.Logf("loaded %d projects and %d tasks in %s", loaded, tasks, elapsed)
	if loaded != projects {
		t.Errorf("loaded %d of %d projects", loaded, projects)
	}
	if elapsed > budget {
		t.Errorf("load took %s, over the %s budget", elapsed, budget)
	}

	start = time.Now()
	for _, p := range s.Projects() {
		s.Reload(p.Registry.Path)
	}
	t.Logf("reloading every project one by one took %s", time.Since(start))

	start = time.Now()
	got := s.Query(store.Query{Statuses: []string{"To Do"}})
	t.Logf("querying %d of %d tasks took %s", len(got), tasks, time.Since(start))

	// The kin index is rebuilt on every board refresh, so its cost is paid
	// alongside the query rather than once at startup.
	start = time.Now()
	kin := s.KinIndex()
	elapsed = time.Since(start)
	var resolved, counted int
	for _, k := range kin {
		if k.Parent != nil {
			resolved++
		}
		counted += len(k.Children)
	}
	// The two differ by the archived subtasks, which keep their link upwards
	// and are counted under no parent.
	t.Logf("resolved %d parent links, %d of them counted as subtasks, in %s",
		resolved, counted, elapsed)
	if elapsed > 50*time.Millisecond {
		t.Errorf("kin index took %s over the whole corpus", elapsed)
	}
}
