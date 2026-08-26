package app_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/FMakareev/muster-backlog/internal/app"
)

func task(id, title, status string) string {
	return "---\nid: " + id + "\ntitle: " + title + "\nstatus: " + status +
		"\n---\n\n## Description\n\n<!-- SECTION:DESCRIPTION:BEGIN -->\nx\n" +
		"<!-- SECTION:DESCRIPTION:END -->\n"
}

func names(views []app.ProjectView) []string {
	var out []string
	for _, v := range views {
		out = append(out, v.Name)
	}
	return out
}

func TestInspectFolderTellsRegistrationFromInitialisation(t *testing.T) {
	backlogged := newProject(t, "alpha", map[string]string{
		"task-1 - a.md": task("TASK-1", "A", "To Do"),
	})
	plain := t.TempDir()
	file := filepath.Join(t.TempDir(), "notes.md")
	if err := os.WriteFile(file, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}

	s := startService(t, withRegistry(t, backlogged))

	registered := s.InspectFolder(backlogged)
	if !registered.HasBacklog || !registered.Registered {
		t.Errorf("a registered project reads as %+v", registered)
	}

	fresh := s.InspectFolder(plain)
	switch {
	case fresh.HasBacklog:
		t.Error("an empty folder was reported as holding a backlog")
	case fresh.Registered:
		t.Error("an empty folder was reported as registered")
	case !fresh.IsDir || fresh.Problem != "":
		t.Errorf("an ordinary folder reads as %+v", fresh)
	}
	if fresh.Name != filepath.Base(plain) {
		t.Errorf("name is %q, want the folder name", fresh.Name)
	}

	if got := s.InspectFolder(file); got.Problem == "" || got.IsDir {
		t.Errorf("a file reads as %+v", got)
	}
	if got := s.InspectFolder(filepath.Join(plain, "nowhere")); got.Problem == "" {
		t.Error("a path that is not there was accepted")
	}
	if got := s.InspectFolder("  "); got.Problem == "" {
		t.Error("an empty path was accepted")
	}
}

func TestAddRemoveAndReorderProjects(t *testing.T) {
	one := newProject(t, "one", map[string]string{"task-1 - a.md": task("TASK-1", "A", "To Do")})
	two := newProject(t, "two", map[string]string{"task-1 - b.md": task("TASK-1", "B", "To Do")})
	path := withRegistry(t, one)
	s := startService(t, path)

	if result := s.AddProject(two, app.ProjectEdit{Name: "Second", Colour: "#123456"}); !result.OK {
		t.Fatalf("AddProject: %+v", result.Problem)
	}
	if got := names(s.Projects()); strings.Join(got, ",") != "one,Second" {
		t.Fatalf("projects are %v, want one then Second", got)
	}

	// The board sees it without a restart: the write reloads.
	found := false
	for _, view := range s.Projects() {
		if view.Name == "Second" && view.TaskCount == 1 {
			found = true
		}
	}
	if !found {
		t.Error("the new project's tasks were not loaded")
	}

	if result := s.MoveProject(two, 0); !result.OK {
		t.Fatalf("MoveProject: %+v", result.Problem)
	}
	if got := names(s.Projects()); strings.Join(got, ",") != "Second,one" {
		t.Errorf("after moving, projects are %v", got)
	}

	if result := s.RemoveProject(one); !result.OK {
		t.Fatalf("RemoveProject: %+v", result.Problem)
	}
	if got := names(s.Projects()); strings.Join(got, ",") != "Second" {
		t.Errorf("after removing, projects are %v", got)
	}
	// Unregistering is not deleting.
	if _, err := os.Stat(filepath.Join(one, "backlog", "config.yml")); err != nil {
		t.Errorf("removing the entry touched the folder: %v", err)
	}
}

func TestAddRefusesWhatCannotBeAdded(t *testing.T) {
	one := newProject(t, "one", map[string]string{"task-1 - a.md": task("TASK-1", "A", "To Do")})
	plain := t.TempDir()
	s := startService(t, withRegistry(t, one))

	cases := map[string]string{
		"a folder already registered": one,
		"a folder with no backlog":    plain,
		"a path that is not there":    filepath.Join(plain, "nowhere"),
	}
	for name, path := range cases {
		result := s.AddProject(path, app.ProjectEdit{})
		if result.OK {
			t.Errorf("%s was accepted", name)
			continue
		}
		if result.Problem == nil || result.Problem.Detail == "" {
			t.Errorf("%s was refused without an explanation", name)
		}
	}
	if len(s.Projects()) != 1 {
		t.Errorf("a refused add still changed the registry: %v", names(s.Projects()))
	}
}

func TestSaveProjectRenamesAndRecolours(t *testing.T) {
	one := newProject(t, "one", map[string]string{"task-1 - a.md": task("TASK-1", "A", "To Do")})
	s := startService(t, withRegistry(t, one))

	if result := s.SaveProject(one, app.ProjectEdit{
		Name: "Renamed", Colour: "#abcdef",
	}); !result.OK {
		t.Fatalf("SaveProject: %+v", result.Problem)
	}
	view := s.Projects()[0]
	if view.Name != "Renamed" || view.Colour != "#abcdef" {
		t.Errorf("got %q / %q, want Renamed / #abcdef", view.Name, view.Colour)
	}
}

// Hiding is a display choice: the project stays registered and loaded, and
// everything that answers a question about the work leaves it out.
func TestHidingAProjectKeepsItRegisteredAndOutOfTheWay(t *testing.T) {
	one := newProject(t, "one", map[string]string{"task-1 - a.md": task("TASK-1", "A", "To Do")})
	two := newProject(t, "two", map[string]string{"task-9 - b.md": task("TASK-9", "Bee", "To Do")})
	s := startService(t, withRegistry(t, one, two))

	if got := len(s.Tasks(app.QueryInput{})); got != 2 {
		t.Fatalf("started with %d tasks, want 2", got)
	}

	if result := s.SaveProject(two, app.ProjectEdit{Hidden: true}); !result.OK {
		t.Fatalf("SaveProject: %+v", result.Problem)
	}

	// Still registered, still counted on its own row.
	views := s.Projects()
	if len(views) != 2 {
		t.Fatalf("hiding removed a project: %v", names(views))
	}
	var hidden app.ProjectView
	for _, v := range views {
		if v.Path == two {
			hidden = v
		}
	}
	if !hidden.Hidden {
		t.Error("the project is not marked hidden")
	}
	if hidden.TaskCount != 1 {
		t.Errorf("a hidden project shows %d tasks, want its own count", hidden.TaskCount)
	}

	// And out of everything that answers a question about the work.
	if got := s.Tasks(app.QueryInput{}); len(got) != 1 {
		t.Errorf("a hidden project's tasks are still on the board: %d", len(got))
	}
	if got := s.Search("Bee", 10); len(got) != 0 {
		t.Errorf("a hidden project's tasks are still searchable: %+v", got)
	}
	for _, a := range s.Analytics() {
		if a.Project == two {
			t.Error("a hidden project is still in the figures")
		}
	}

	// Unhiding brings it straight back.
	if result := s.SaveProject(two, app.ProjectEdit{Hidden: false}); !result.OK {
		t.Fatalf("unhide: %+v", result.Problem)
	}
	if got := len(s.Tasks(app.QueryInput{})); got != 2 {
		t.Errorf("unhiding left %d tasks, want 2", got)
	}
}

// The registry is a file people write by hand, and editing it from the
// application must not cost them their notes.
func TestRegistryEditsKeepTheFileAsWritten(t *testing.T) {
	one := newProject(t, "one", map[string]string{"task-1 - a.md": task("TASK-1", "A", "To Do")})
	two := newProject(t, "two", map[string]string{"task-1 - b.md": task("TASK-1", "B", "To Do")})
	path := withRegistry(t, one)

	annotated := "# the ones I care about\n" +
		"projects:\n  - path: " + one + "   # first\n\n" +
		"# advisory only\nwip_limits:\n  In Progress: 2\n"
	if err := os.WriteFile(path, []byte(annotated), 0o644); err != nil {
		t.Fatal(err)
	}

	s := startService(t, path)
	if result := s.AddProject(two, app.ProjectEdit{Name: "Two"}); !result.OK {
		t.Fatalf("AddProject: %+v", result.Problem)
	}

	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	got := string(raw)
	for _, want := range []string{
		"# the ones I care about", "# advisory only", "# first",
		"wip_limits:", "In Progress: 2",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("lost %q from the registry:\n%s", want, got)
		}
	}
}

// The whole path, through the real CLI: an ordinary folder that is not a git
// repository becomes a registered project on the board.
func TestInitialiseAFolderAndRegisterIt(t *testing.T) {
	if _, err := exec.LookPath("backlog"); err != nil {
		t.Skip("the backlog CLI is not installed")
	}
	s := startService(t, withRegistry(t))
	folder := t.TempDir()

	before := s.InspectFolder(folder)
	if before.HasBacklog || before.IsGit {
		t.Fatalf("the scratch folder is not what the test assumes: %+v", before)
	}

	result := s.InitProject(app.InitFolder{
		Path:            folder,
		Name:            "Fresh",
		TaskPrefix:      "job",
		ZeroPaddedIDs:   3,
		IntegrationMode: "none",
		Colour:          "#445566",
	})
	if !result.OK {
		t.Fatalf("InitProject: %+v", result.Problem)
	}

	after := s.InspectFolder(folder)
	if !after.HasBacklog {
		t.Error("no backlog was created")
	}

	views := s.Projects()
	if len(views) != 1 {
		t.Fatalf("got %d projects, want the one just created", len(views))
	}
	if views[0].Name != "Fresh" || views[0].Colour != "#445566" {
		t.Errorf("registered as %q / %q", views[0].Name, views[0].Colour)
	}
	if !views[0].OK {
		t.Errorf("the new project did not load: %s", views[0].Problem)
	}
	// The choices the form made reached the CLI.
	config, err := os.ReadFile(filepath.Join(folder, "backlog", "config.yml"))
	if err != nil {
		t.Fatalf("reading the new config: %v", err)
	}
	if !strings.Contains(string(config), "job") {
		t.Errorf("the task prefix did not reach the CLI:\n%s", config)
	}
	if !strings.Contains(string(config), "zero_padded_ids: 3") {
		t.Errorf("zero padding did not reach the CLI:\n%s", config)
	}
}

// A folder that already holds a backlog is offered for registration, not
// initialised over the top of what is there.
func TestInitialisingOverAnExistingBacklogIsRefused(t *testing.T) {
	one := newProject(t, "one", map[string]string{"task-1 - a.md": task("TASK-1", "A", "To Do")})
	s := startService(t, withRegistry(t))

	result := s.InitProject(app.InitFolder{Path: one})
	if result.OK {
		t.Fatal("initialising over an existing backlog was allowed")
	}
	if result.Problem == nil || !strings.Contains(result.Problem.Detail, "Add the folder") {
		t.Errorf("refused without pointing at the alternative: %+v", result.Problem)
	}
}

// A person who wrote ~/Dev/thing must still have ~/Dev/thing after the
// application rewrites the entry. Resolution expands it everywhere else, and
// writing the expanded form back would quietly rewrite their file.
func TestEditingKeepsThePathAsWritten(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Skip("no home directory")
	}
	folder := filepath.Join(home, ".muster-test-project")
	if err := os.MkdirAll(filepath.Join(folder, "backlog", "tasks"), 0o755); err != nil {
		t.Skipf("cannot write under the home directory: %v", err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(folder) })
	if err := os.WriteFile(filepath.Join(folder, "backlog", "config.yml"),
		[]byte("project_name: \"Tilde\"\nstatuses: [\"To Do\", \"Done\"]\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	path := withRegistry(t)
	if err := os.WriteFile(path,
		[]byte("projects:\n  - path: ~/.muster-test-project\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	s := startService(t, path)

	if result := s.SaveProject(folder, app.ProjectEdit{Name: "Tilde"}); !result.OK {
		t.Fatalf("SaveProject: %+v", result.Problem)
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(raw), "path: ~/.muster-test-project") {
		t.Errorf("the tilde was expanded in the person's own file:\n%s", raw)
	}
}

// When init fails, what the CLI printed is what the person needs: it writes
// files, and a half-written folder needs the real reason rather than a summary.
func TestInitFailureCarriesTheCLIsOwnWords(t *testing.T) {
	if _, err := exec.LookPath("backlog"); err != nil {
		t.Skip("the backlog CLI is not installed")
	}
	s := startService(t, withRegistry(t))

	// The CLI requires a prefix of letters only.
	result := s.InitProject(app.InitFolder{
		Path: t.TempDir(), TaskPrefix: "12345", IntegrationMode: "none",
	})
	if result.OK {
		t.Fatal("a prefix the CLI rejects was accepted")
	}
	if result.Problem == nil {
		t.Fatal("failed without a problem to show")
	}
	detail := strings.ToLower(result.Problem.Detail)
	if !strings.Contains(detail, "prefix") {
		t.Errorf("the CLI's own explanation did not reach the interface: %q",
			result.Problem.Detail)
	}
	if len(s.Projects()) != 0 {
		t.Error("a folder that failed to initialise was registered anyway")
	}
}

// Asking for agent instruction files with AI integration off is a combination
// the CLI refuses, and it is caught before anything is written.
func TestInitRefusesTheImpossibleCombinationBeforeWriting(t *testing.T) {
	if _, err := exec.LookPath("backlog"); err != nil {
		t.Skip("the backlog CLI is not installed")
	}
	s := startService(t, withRegistry(t))
	folder := t.TempDir()

	result := s.InitProject(app.InitFolder{
		Path: folder, IntegrationMode: "none", AgentInstructions: "claude",
	})
	if result.OK {
		t.Fatal("the combination was accepted")
	}
	if entries, err := os.ReadDir(folder); err != nil || len(entries) != 0 {
		t.Errorf("the folder was written to before the refusal: %v", entries)
	}
}

// --defaults does not answer the project name: it is a positional argument,
// and without one the CLI prompts for it and then exits 0 having created
// nothing. Leaving the name empty in the form must still produce a project.
func TestInitWithNoNameStillCreatesAProject(t *testing.T) {
	if _, err := exec.LookPath("backlog"); err != nil {
		t.Skip("the backlog CLI is not installed")
	}
	s := startService(t, withRegistry(t))
	folder := t.TempDir()

	result := s.InitProject(app.InitFolder{Path: folder, IntegrationMode: "none"})
	if !result.OK {
		t.Fatalf("InitProject: %+v", result.Problem)
	}
	if !s.InspectFolder(folder).HasBacklog {
		t.Fatal("the CLI exited without creating anything")
	}
	if got := s.Projects(); len(got) != 1 || got[0].Name != filepath.Base(folder) {
		t.Errorf("registered as %+v, want the folder name", names(got))
	}
}

// There is no application during tests and none in the server build, so the
// interface must be told there is no dialog rather than offered a button that
// can only fail.
func TestNoFolderDialogWithoutADesktop(t *testing.T) {
	if app.FolderDialogAvailable() {
		t.Fatal("a dialog was reported available with no application running")
	}
	s := startService(t, withRegistry(t))
	if got := s.ChooseFolder(); got != "" {
		t.Errorf("ChooseFolder returned %q with no dialog to open", got)
	}
}
