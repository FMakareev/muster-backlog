package store_test

import (
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/FMakareev/muster-backlog/internal/backlog"
	"github.com/FMakareev/muster-backlog/internal/registry"
	"github.com/FMakareev/muster-backlog/internal/store"
)

// fixture builds a registry over temporary projects and returns a loaded store.
type fixture struct {
	root  string
	paths map[string]string
	reg   registry.Registry
	store *store.Store
}

func write(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

// newProject creates a project with the given task files, each written as
// "id|title|status|priority|type|milestone|labels".
func newProject(t *testing.T, root, name, statuses string, tasks ...string) string {
	t.Helper()
	dir := filepath.Join(root, name)
	write(t, filepath.Join(dir, "backlog", "config.yml"),
		"project_name: \""+name+"\"\nstatuses: "+statuses+"\n")
	if err := os.MkdirAll(filepath.Join(dir, "backlog", "tasks"), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	for _, spec := range tasks {
		f := strings.Split(spec, "|")
		content := "---\nid: " + f[0] + "\ntitle: " + f[1] + "\nstatus: " + f[2] + "\n"
		if len(f) > 3 && f[3] != "" {
			content += "priority: " + f[3] + "\n"
		}
		if len(f) > 4 && f[4] != "" {
			content += "type: " + f[4] + "\n"
		}
		if len(f) > 5 && f[5] != "" {
			content += "milestone: " + f[5] + "\n"
		}
		if len(f) > 6 && f[6] != "" {
			content += "labels: [" + f[6] + "]\n"
		}
		content += "---\n\n## Description\n\n<!-- SECTION:DESCRIPTION:BEGIN -->\n" +
			f[1] + " description.\n<!-- SECTION:DESCRIPTION:END -->\n"
		write(t, filepath.Join(dir, "backlog", "tasks",
			strings.ToLower(f[0])+" - t.md"), content)
	}
	return dir
}

func newFixture(t *testing.T) *fixture {
	t.Helper()
	root := t.TempDir()

	alpha := newProject(t, root, "alpha", `["To Do", "In Progress", "Done"]`,
		"TASK-1|Alpha one|To Do|high|feature||backend",
		"TASK-2|Alpha two|In Progress|medium|bug|m-1|frontend",
	)
	beta := newProject(t, root, "beta", `["To Do", "In Review", "Done"]`,
		"TASK-1|Beta one|To Do|low|chore||backend",
		"TASK-3|Beta three|Done|high|feature|m-2|",
	)

	regFile := filepath.Join(root, "projects.yml")
	write(t, regFile, "projects:\n  - path: "+alpha+"\n    name: Alpha\n"+
		"    color: \"#111111\"\n  - path: "+beta+"\n    name: Beta\n")

	reg, err := registry.LoadFrom(regFile)
	if err != nil {
		t.Fatalf("LoadFrom: %v", err)
	}
	s := store.New()
	s.Load(reg)

	return &fixture{
		root:  root,
		paths: map[string]string{"alpha": alpha, "beta": beta},
		reg:   reg,
		store: s,
	}
}

// The same id in two projects must yield two items, not one shadowing the other.
func TestIDsCollideAcrossProjectsWithoutShadowing(t *testing.T) {
	f := newFixture(t)

	items := f.store.Query(store.Query{})
	if len(items) != 4 {
		t.Fatalf("got %d items, want 4", len(items))
	}

	var task1 []store.Item
	for _, it := range items {
		if it.Entity.ID == "TASK-1" {
			task1 = append(task1, it)
		}
	}
	if len(task1) != 2 {
		t.Fatalf("got %d items with id TASK-1, want one per project", len(task1))
	}
	if task1[0].Ref == task1[1].Ref {
		t.Fatal("refs collide; identity is not project-qualified")
	}
	if task1[0].Entity.Title == task1[1].Entity.Title {
		t.Error("the two TASK-1s should be different tasks")
	}

	// And each is retrievable by its own ref.
	for _, want := range task1 {
		got, ok := f.store.Get(want.Ref)
		if !ok {
			t.Fatalf("Get(%s) missed", want.Ref)
		}
		if got.Entity.Title != want.Entity.Title {
			t.Errorf("Get(%s) returned %q, want %q",
				want.Ref, got.Entity.Title, want.Entity.Title)
		}
	}
}

func TestQueryFilters(t *testing.T) {
	f := newFixture(t)

	cases := []struct {
		name  string
		query store.Query
		want  []string
	}{
		{"status", store.Query{Statuses: []string{"To Do"}},
			[]string{"Alpha one", "Beta one"}},
		{"project by display name", store.Query{Projects: []string{"Beta"}},
			[]string{"Beta one", "Beta three"}},
		{"milestone", store.Query{Milestones: []string{"m-2"}},
			[]string{"Beta three"}},
		{"priority", store.Query{Priorities: []string{"High"}},
			[]string{"Alpha one", "Beta three"}},
		{"type", store.Query{Types: []string{"feature"}},
			[]string{"Alpha one", "Beta three"}},
		{"label", store.Query{Labels: []string{"backend"}},
			[]string{"Alpha one", "Beta one"}},
		{"text matches description", store.Query{Text: "beta three description"},
			[]string{"Beta three"}},
		{"filters combine", store.Query{
			Statuses: []string{"To Do"}, Labels: []string{"backend"},
			Projects: []string{"Alpha"}},
			[]string{"Alpha one"}},
		{"no match", store.Query{Statuses: []string{"Nonexistent"}}, nil},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var got []string
			for _, it := range f.store.Query(tc.query) {
				got = append(got, it.Entity.Title)
			}
			if strings.Join(got, "|") != strings.Join(tc.want, "|") {
				t.Errorf("got %v, want %v", got, tc.want)
			}
		})
	}
}

// Priority is written lowercase on disk against a capitalised config list, so
// a case-sensitive filter would silently match nothing.
func TestFiltersAreCaseInsensitive(t *testing.T) {
	f := newFixture(t)
	for _, spelling := range []string{"high", "High", "HIGH"} {
		if n := f.store.Count(store.Query{Priorities: []string{spelling}}); n != 2 {
			t.Errorf("Count(priority=%q) = %d, want 2", spelling, n)
		}
	}
}

// Registry order is the order the user sees.
func TestQueryPreservesRegistryOrder(t *testing.T) {
	f := newFixture(t)
	items := f.store.Query(store.Query{})
	if items[0].ProjectName != "Alpha" || items[len(items)-1].ProjectName != "Beta" {
		t.Errorf("order = %s..%s, want Alpha first and Beta last",
			items[0].ProjectName, items[len(items)-1].ProjectName)
	}
}

// A change on disk must cost one project rescan, not a corpus reload.
func TestReloadTouchesOneProjectOnly(t *testing.T) {
	f := newFixture(t)
	alpha, beta := f.paths["alpha"], f.paths["beta"]

	before := map[string]int{}
	for _, p := range f.store.Projects() {
		before[p.Registry.Path] = len(p.Scanned.Tasks)
	}

	write(t, filepath.Join(alpha, "backlog", "tasks", "task-9 - new.md"),
		"---\nid: TASK-9\ntitle: Alpha nine\nstatus: Done\n---\n")

	betaLoadedAt := loadedAt(t, f.store, beta)

	if !f.store.Reload(alpha) {
		t.Fatal("Reload reported the project as unknown")
	}

	after := map[string]int{}
	for _, p := range f.store.Projects() {
		after[p.Registry.Path] = len(p.Scanned.Tasks)
	}
	if after[alpha] != before[alpha]+1 {
		t.Errorf("alpha has %d tasks, want %d", after[alpha], before[alpha]+1)
	}
	if after[beta] != before[beta] {
		t.Errorf("beta changed from %d to %d tasks", before[beta], after[beta])
	}
	if got := loadedAt(t, f.store, beta); !got.Equal(betaLoadedAt) {
		t.Error("beta was rescanned; only the changed project should be")
	}
}

func loadedAt(t *testing.T, s *store.Store, path string) time.Time {
	t.Helper()
	for _, p := range s.Projects() {
		if p.Registry.Path == path {
			return p.LoadedAt
		}
	}
	t.Fatalf("project %s not in the store", path)
	return time.Time{}
}

func TestReloadUnknownProject(t *testing.T) {
	f := newFixture(t)
	if f.store.Reload(filepath.Join(f.root, "nope")) {
		t.Error("Reload of an unregistered project should report false")
	}
}

// The union algorithm belongs to the board; the store hands over the lists.
func TestStatusListsAreExposedPerProject(t *testing.T) {
	f := newFixture(t)
	lists := f.store.StatusLists()
	if len(lists) != 2 {
		t.Fatalf("got %d lists, want one per project", len(lists))
	}
	if strings.Join(lists[0], "|") == strings.Join(lists[1], "|") {
		t.Error("the fixture projects declare different statuses; the lists should differ")
	}
}

func TestCountByStatus(t *testing.T) {
	f := newFixture(t)
	counts := f.store.CountByStatus(f.paths["alpha"])
	if counts["To Do"] != 1 || counts["In Progress"] != 1 {
		t.Errorf("counts = %v, want one To Do and one In Progress", counts)
	}
}

func TestValuesForFilterMenus(t *testing.T) {
	f := newFixture(t)
	if got := strings.Join(f.store.Values("label"), "|"); got != "backend|frontend" {
		t.Errorf("labels = %q", got)
	}
	if got := strings.Join(f.store.Values("milestone"), "|"); got != "m-1|m-2" {
		t.Errorf("milestones = %q", got)
	}
}

// A broken project is reported, never dropped, and never stops the others.
func TestBrokenProjectIsReportedNotDropped(t *testing.T) {
	root := t.TempDir()
	good := newProject(t, root, "good", `["To Do"]`, "TASK-1|Fine|To Do")
	missing := filepath.Join(root, "gone")

	regFile := filepath.Join(root, "projects.yml")
	write(t, regFile, "projects:\n  - path: "+missing+"\n  - path: "+good+"\n")
	reg, err := registry.LoadFrom(regFile)
	if err != nil {
		t.Fatalf("LoadFrom: %v", err)
	}
	s := store.New()
	s.Load(reg)

	states := s.Projects()
	if len(states) != 2 {
		t.Fatalf("got %d projects, want both reported", len(states))
	}
	if states[0].OK() {
		t.Error("the missing project should not be OK")
	}
	if !states[1].OK() {
		t.Errorf("the good project failed to load: %v", states[1].Err)
	}
	if n := len(s.Query(store.Query{})); n != 1 {
		t.Errorf("got %d items, want the good project's task", n)
	}
	// A broken project is reported on its ProjectState, not as a skipped file.
	// Reporting it in both places would make one broken folder look like two
	// problems, the second of them mislabelled.
	for _, d := range s.Diagnostics() {
		if d.Path == missing {
			t.Errorf("the broken project also appeared as a file diagnostic: %+v", d)
		}
	}
}

// The watcher reloads while the UI reads, so this must be race-free.
func TestConcurrentReadsAndReloads(t *testing.T) {
	f := newFixture(t)
	var wg sync.WaitGroup

	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 50; j++ {
				f.store.Query(store.Query{Statuses: []string{"To Do"}})
				f.store.Projects()
				f.store.Diagnostics()
			}
		}()
	}
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 25; j++ {
				f.store.Reload(f.paths["alpha"])
				f.store.Reload(f.paths["beta"])
			}
		}()
	}
	wg.Wait()

	if n := f.store.Count(store.Query{}); n != 4 {
		t.Errorf("got %d items after concurrent access, want 4", n)
	}
}

func TestQueryKinds(t *testing.T) {
	f := newFixture(t)
	if n := len(f.store.Query(store.Query{Kinds: []backlog.Kind{backlog.KindDraft}})); n != 0 {
		t.Errorf("got %d drafts, want none", n)
	}
	if n := len(f.store.Query(store.Query{Kinds: []backlog.Kind{backlog.KindTask}})); n != 4 {
		t.Errorf("got %d tasks, want 4", n)
	}
}

// Archiving is a soft delete and completed/ is where finished work is filed.
// A query that names no class means the live ones; the board carried archived
// tasks as though they were live until this was noticed.
func TestAnUnqualifiedQueryReturnsOnlyLiveTasks(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Join(root, "solo")
	write(t, filepath.Join(dir, "backlog", "config.yml"),
		"project_name: \"Solo\"\nstatuses: [\"To Do\", \"Done\"]\n")

	body := func(id, title string) string {
		return "---\nid: " + id + "\ntitle: " + title + "\nstatus: To Do\n---\n\n" +
			"## Description\n\n<!-- SECTION:DESCRIPTION:BEGIN -->\nx\n" +
			"<!-- SECTION:DESCRIPTION:END -->\n"
	}
	write(t, filepath.Join(dir, "backlog", "tasks", "task-1 - a.md"), body("TASK-1", "Live"))
	write(t, filepath.Join(dir, "backlog", "completed", "task-2 - b.md"), body("TASK-2", "Filed"))
	write(t, filepath.Join(dir, "backlog", "archive", "tasks", "task-3 - c.md"), body("TASK-3", "Gone"))

	regFile := filepath.Join(root, "projects.yml")
	write(t, regFile, "projects:\n  - path: "+dir+"\n")
	reg, err := registry.LoadFrom(regFile)
	if err != nil {
		t.Fatalf("LoadFrom: %v", err)
	}
	s := store.New()
	s.Load(reg)

	got := s.Query(store.Query{})
	if len(got) != 1 || got[0].Entity.Title != "Live" {
		var titles []string
		for _, item := range got {
			titles = append(titles, item.Entity.Title)
		}
		t.Fatalf("an unqualified query returned %v, want only the live task", titles)
	}
	if n := s.Count(store.Query{}); n != 1 {
		t.Errorf("Count returned %d, want 1", n)
	}

	// Asking for them explicitly still works.
	all := s.Query(store.Query{Classes: []backlog.Class{
		backlog.ClassActive, backlog.ClassCompleted, backlog.ClassArchived,
	}})
	if len(all) != 3 {
		t.Errorf("asking for every class returned %d, want 3", len(all))
	}
}
