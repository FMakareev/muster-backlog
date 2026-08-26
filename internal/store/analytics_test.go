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

// analyticsFixture builds a project whose numbers are known exactly.
func analyticsFixture(t *testing.T) *store.Store {
	t.Helper()
	root := t.TempDir()
	dir := filepath.Join(root, "alpha")
	tasks := filepath.Join(dir, "backlog", "tasks")
	if err := os.MkdirAll(tasks, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	write(t, filepath.Join(dir, "backlog", "config.yml"),
		"project_name: Alpha\nstatuses: [\"To Do\", \"In Progress\", \"Done\"]\n")

	// Two open, one finished. One open task is old and untouched; one waits on
	// a task that is not finished, and one waits on a task that is.
	write(t, filepath.Join(tasks, "task-1 - a.md"),
		"---\nid: TASK-1\ntitle: Fresh\nstatus: To Do\npriority: high\ntype: feature\n"+
			"created_date: '2026-08-20 10:00'\nupdated_date: '2026-08-25 10:00'\n---\n")
	write(t, filepath.Join(tasks, "task-2 - b.md"),
		"---\nid: TASK-2\ntitle: Forgotten\nstatus: In Progress\n"+
			"created_date: '2026-01-01 10:00'\nupdated_date: '2026-01-02 10:00'\n"+
			"dependencies:\n  - TASK-3\n  - TASK-4\n---\n")
	write(t, filepath.Join(tasks, "task-3 - c.md"),
		"---\nid: TASK-3\ntitle: Finished\nstatus: Done\npriority: low\n"+
			"created_date: '2026-02-01 10:00'\n---\n")
	write(t, filepath.Join(tasks, "task-4 - d.md"),
		"---\nid: TASK-4\ntitle: Also open\nstatus: To Do\n"+
			"created_date: '2026-08-24 10:00'\nupdated_date: '2026-08-26 10:00'\n---\n")

	regPath := filepath.Join(root, registry.FileName)
	write(t, regPath, "projects:\n  - path: "+dir+"\n    name: Alpha\n")
	reg, err := registry.LoadFrom(regPath)
	if err != nil {
		t.Fatalf("LoadFrom: %v", err)
	}
	s := store.New()
	s.Load(reg)
	return s
}

func TestAnalyticsCounts(t *testing.T) {
	s := analyticsFixture(t)
	now := time.Date(2026, 8, 26, 12, 0, 0, 0, time.UTC)

	all := s.Analytics(store.AnalyticsOptions{Now: now})
	if len(all) != 2 {
		t.Fatalf("got %d entries, want one project and the total", len(all))
	}
	a := all[0]

	if a.Tasks != 4 {
		t.Errorf("tasks = %d, want 4", a.Tasks)
	}
	// Three of the four carry no priority - the symptom that started this
	// project, and the reason the count exists at all.
	if a.Unprioritised != 2 {
		t.Errorf("unprioritised = %d, want 2", a.Unprioritised)
	}

	byStatus := map[string]int{}
	for _, c := range a.Statuses {
		byStatus[c.Label] = c.Total
	}
	if byStatus["To Do"] != 2 || byStatus["In Progress"] != 1 || byStatus["Done"] != 1 {
		t.Errorf("statuses = %+v", a.Statuses)
	}

	// The total is the last entry and matches.
	if total := all[len(all)-1]; total.Tasks != a.Tasks {
		t.Errorf("total tasks = %d, want %d", total.Tasks, a.Tasks)
	}
}

// Stale is measured on open tasks only: a finished task nobody has touched for
// a year is not a problem.
func TestAnalyticsStale(t *testing.T) {
	s := analyticsFixture(t)
	now := time.Date(2026, 8, 26, 12, 0, 0, 0, time.UTC)

	a := s.Analytics(store.AnalyticsOptions{Now: now, StaleAfter: 30 * 24 * time.Hour})[0]
	if len(a.Stale) != 1 {
		t.Fatalf("stale = %+v, want the one forgotten open task", titles(a.Stale))
	}
	if a.Stale[0].Entity.Title != "Forgotten" {
		t.Errorf("stale = %q", a.Stale[0].Entity.Title)
	}

	// A shorter threshold catches more; a longer one catches none.
	if n := len(s.Analytics(store.AnalyticsOptions{Now: now, StaleAfter: time.Hour})[0].Stale); n != 3 {
		t.Errorf("stale with a one-hour threshold = %d, want every open task", n)
	}
	if n := len(s.Analytics(store.AnalyticsOptions{
		Now: now, StaleAfter: 365 * 24 * time.Hour,
	})[0].Stale); n != 0 {
		t.Errorf("stale with a one-year threshold = %d, want none", n)
	}
}

// A dependency that is finished does not block; one that is not does.
func TestAnalyticsBlocked(t *testing.T) {
	s := analyticsFixture(t)
	a := s.Analytics(store.AnalyticsOptions{
		Now: time.Date(2026, 8, 26, 12, 0, 0, 0, time.UTC),
	})[0]

	if len(a.Blocked) != 1 {
		t.Fatalf("blocked = %d, want the one waiting task", len(a.Blocked))
	}
	b := a.Blocked[0]
	if b.Item.Entity.ID != "TASK-2" {
		t.Errorf("blocked task = %q", b.Item.Entity.ID)
	}
	// TASK-3 is Done and must not count; TASK-4 is open and must.
	if len(b.On) != 1 || b.On[0] != "TASK-4" {
		t.Errorf("waiting on %v, want only the unfinished dependency", b.On)
	}
}

func TestAnalyticsAverageAgeIgnoresFinishedWork(t *testing.T) {
	s := analyticsFixture(t)
	now := time.Date(2026, 8, 26, 12, 0, 0, 0, time.UTC)
	a := s.Analytics(store.AnalyticsOptions{Now: now})[0]

	// Open tasks are 6, 237 and 2 days old; the finished one is excluded.
	if a.AverageAgeDays < 80 || a.AverageAgeDays > 84 {
		t.Errorf("average age = %.1f days, want about 82", a.AverageAgeDays)
	}
}

func titles(items []store.Item) []string {
	out := make([]string, 0, len(items))
	for _, i := range items {
		out = append(out, i.Entity.Title)
	}
	return out
}

// Search ranks titles above bodies, because someone typing a few words is
// usually reaching for something they remember by name.
func TestSearchRanksTitlesFirst(t *testing.T) {
	f := newFixture(t)
	hits := f.store.Search("alpha", 20)
	if len(hits) == 0 {
		t.Fatal("no hits")
	}
	seenBody := false
	for _, hit := range hits {
		if hit.Field == "body" {
			seenBody = true
		} else if seenBody {
			t.Fatalf("a %s hit came after a body hit", hit.Field)
		}
	}
}

func TestSearchFindsBodiesWithAnExcerpt(t *testing.T) {
	f := newFixture(t)
	hits := f.store.Search("description", 20)
	if len(hits) == 0 {
		t.Fatal("nothing matched a word that appears only in bodies")
	}
	for _, hit := range hits {
		if hit.Field == "body" && !strings.Contains(strings.ToLower(hit.Excerpt), "description") {
			t.Errorf("excerpt %q does not contain the match", hit.Excerpt)
		}
	}
}

func TestSearchEmptyAndUnmatched(t *testing.T) {
	f := newFixture(t)
	if hits := f.store.Search("   ", 10); hits != nil {
		t.Errorf("blank search returned %d hits", len(hits))
	}
	if hits := f.store.Search("nothingmatchesthis", 10); len(hits) != 0 {
		t.Errorf("got %d hits for a word nothing contains", len(hits))
	}
}

func TestSearchRespectsTheLimit(t *testing.T) {
	f := newFixture(t)
	if hits := f.store.Search("a", 1); len(hits) > 1 {
		t.Errorf("got %d hits, want at most 1", len(hits))
	}
}
