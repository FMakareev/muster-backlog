package mcpserver_test

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/FMakareev/muster-backlog/internal/mcpserver"
	"github.com/FMakareev/muster-backlog/internal/testcli"
)

// bench builds two real projects and a registry over them, and returns a
// connected client. Real projects because the point of this server is what the
// files say, and a real client because the schemas are half the contract.
func bench(t *testing.T) (*mcp.ClientSession, string) {
	t.Helper()
	testcli.Require(t)

	root := t.TempDir()
	var paths []string
	for _, name := range []string{"alpha", "beta"} {
		dir := filepath.Join(root, name)
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
		run(t, dir, "init", name, "--defaults", "--no-git", "--integration-mode", "none")
		paths = append(paths, dir)
	}
	run(t, paths[0], "task", "create", "First in alpha", "--priority", "high")
	run(t, paths[0], "task", "create", "Second in alpha", "-s", "Done")
	run(t, paths[0], "milestone", "add", "Opening")
	run(t, paths[1], "task", "create", "Only in beta", "-l", "backend")

	registryPath := filepath.Join(root, "projects.yml")
	content := "projects:\n"
	for i, path := range paths {
		content += "  - path: " + path + "\n    name: " + []string{"Alpha", "Beta"}[i] + "\n"
	}
	if err := os.WriteFile(registryPath, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	server, err := mcpserver.New(registryPath)
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	clientTransport, serverTransport := mcp.NewInMemoryTransports()
	go func() { _ = server.MCP().Run(context.Background(), serverTransport) }()

	client := mcp.NewClient(&mcp.Implementation{Name: "test", Version: "1"}, nil)
	session, err := client.Connect(context.Background(), clientTransport, nil)
	if err != nil {
		t.Fatalf("Connect: %v", err)
	}
	t.Cleanup(func() { _ = session.Close() })
	return session, root
}

func run(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("backlog", args...)
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("backlog %s: %v\n%s", strings.Join(args, " "), err, out)
	}
}

// call runs a tool and decodes its structured answer.
func call[T any](t *testing.T, s *mcp.ClientSession, name string, args any) T {
	t.Helper()
	result, err := s.CallTool(context.Background(), &mcp.CallToolParams{
		Name: name, Arguments: args,
	})
	if err != nil {
		t.Fatalf("%s: %v", name, err)
	}
	if result.IsError {
		t.Fatalf("%s refused: %s", name, textOf(result))
	}
	var out T
	raw, err := json.Marshal(result.StructuredContent)
	if err != nil {
		t.Fatalf("%s: %v", name, err)
	}
	if err := json.Unmarshal(raw, &out); err != nil {
		t.Fatalf("%s: decoding %s: %v", name, raw, err)
	}
	return out
}

func refuse(t *testing.T, s *mcp.ClientSession, name string, args any) string {
	t.Helper()
	result, err := s.CallTool(context.Background(), &mcp.CallToolParams{
		Name: name, Arguments: args,
	})
	if err != nil {
		t.Fatalf("%s: %v", name, err)
	}
	if !result.IsError {
		t.Fatalf("%s was accepted when it should not have been", name)
	}
	return textOf(result)
}

func textOf(result *mcp.CallToolResult) string {
	var parts []string
	for _, content := range result.Content {
		if text, ok := content.(*mcp.TextContent); ok {
			parts = append(parts, text.Text)
		}
	}
	return strings.Join(parts, " ")
}

// The whole point: one answer spanning every registered project.
func TestTheAggregateIsWhatItAnswersWith(t *testing.T) {
	session, _ := bench(t)

	type project struct {
		Name     string         `json:"name"`
		Tasks    int            `json:"tasks"`
		ByStatus map[string]int `json:"byStatus"`
		Statuses []string       `json:"statuses"`
	}
	projects := call[[]project](t, session, "list_projects", struct{}{})
	if len(projects) != 2 {
		t.Fatalf("got %d projects, want both", len(projects))
	}
	if projects[0].Tasks != 2 || projects[1].Tasks != 1 {
		t.Errorf("counts are %d and %d", projects[0].Tasks, projects[1].Tasks)
	}
	// Per-project counts by status, so an agent can see the load.
	if projects[0].ByStatus["Done"] != 1 {
		t.Errorf("alpha's statuses are %v", projects[0].ByStatus)
	}
	if len(projects[0].Statuses) == 0 {
		t.Error("a project did not report the statuses it declares")
	}

	type task struct {
		Project string `json:"project"`
		ID      string `json:"id"`
		Title   string `json:"title"`
		Status  string `json:"status"`
	}
	all := call[[]task](t, session, "list_tasks", map[string]any{})
	if len(all) != 3 {
		t.Fatalf("got %d tasks across both projects, want 3: %+v", len(all), all)
	}
	seen := map[string]bool{}
	for _, item := range all {
		seen[item.Project] = true
	}
	if len(seen) != 2 {
		t.Errorf("the answer covered only %v", seen)
	}

	// The same filters the board has.
	high := call[[]task](t, session, "list_tasks", map[string]any{"priorities": []string{"high"}})
	if len(high) != 1 || high[0].Title != "First in alpha" {
		t.Errorf("filtering by priority gave %+v", high)
	}
	beta := call[[]task](t, session, "list_tasks", map[string]any{"project": "Beta"})
	if len(beta) != 1 || beta[0].Project != "Beta" {
		t.Errorf("filtering by project gave %+v", beta)
	}
}

// Ids are unique only inside a project, so every tool that names one is
// qualified, and an id from the wrong project does not resolve.
func TestOneTaskInFullIsFoundInsideItsOwnProject(t *testing.T) {
	session, _ := bench(t)

	type detail struct {
		Title              string   `json:"title"`
		Description        string   `json:"description"`
		AcceptanceCriteria []string `json:"acceptanceCriteria"`
		Path               string   `json:"path"`
	}
	got := call[detail](t, session, "get_task", map[string]any{
		"project": "Alpha", "id": "TASK-1",
	})
	if got.Title != "First in alpha" {
		t.Errorf("got %+v", got)
	}
	if got.Path == "" {
		t.Error("the answer does not say where the file is")
	}

	said := refuse(t, session, "get_task", map[string]any{
		"project": "Beta", "id": "TASK-2",
	})
	if !strings.Contains(said, "unique only inside") {
		t.Errorf("refused with %q", said)
	}
}

// Access is confined to what the person registered.
func TestAnUnregisteredProjectIsRefused(t *testing.T) {
	session, _ := bench(t)

	for _, name := range []string{"/etc", "nowhere", "../beta"} {
		said := refuse(t, session, "list_tasks", map[string]any{"project": name})
		if !strings.Contains(said, "not a registered project") {
			t.Errorf("%q was refused with %q", name, said)
		}
	}
	// An omitted project is not an unregistered one: it means every project,
	// which is what this server is for.
	type task struct {
		Project string `json:"project"`
	}
	if all := call[[]task](t, session, "list_tasks", map[string]any{}); len(all) != 3 {
		t.Errorf("omitting the project gave %d tasks, want every one", len(all))
	}
	// And the refusal says what is registered, so an agent can correct itself.
	said := refuse(t, session, "get_task", map[string]any{"project": "Gamma", "id": "TASK-1"})
	if !strings.Contains(said, "Alpha") || !strings.Contains(said, "Beta") {
		t.Errorf("the refusal does not name what is registered: %q", said)
	}
}

// Writes go through the same CLI the interface uses, and the files afterwards
// are what the answer claims.
func TestWritesGoThroughTheCLIAndLandOnDisk(t *testing.T) {
	session, root := bench(t)

	type written struct {
		Project string `json:"project"`
		ID      string `json:"id"`
		Did     string `json:"did"`
	}
	made := call[written](t, session, "create_task", map[string]any{
		"project":     "Beta",
		"title":       "Written by an agent",
		"description": "With a reason.",
		"priority":    "high",
		"labels":      []string{"agent"},
	})
	if made.ID == "" || made.Project != "Beta" {
		t.Fatalf("create_task answered %+v", made)
	}

	body := fileWith(t, filepath.Join(root, "beta", "backlog", "tasks"), "Written-by-an-agent")
	if !strings.Contains(body, "priority: high") || !strings.Contains(body, "With a reason.") {
		t.Errorf("the file is not what was asked for:\n%s", body)
	}

	// One field at a time, the way the board does it.
	call[written](t, session, "set_field", map[string]any{
		"project": "Beta", "id": made.ID, "field": "status", "value": "Done",
	})
	body = fileWith(t, filepath.Join(root, "beta", "backlog", "tasks"), "Written-by-an-agent")
	if !strings.Contains(body, "status: Done") {
		t.Errorf("the status did not change:\n%s", body)
	}

	// A status the project does not declare is the CLI's own refusal.
	said := refuse(t, session, "set_field", map[string]any{
		"project": "Beta", "id": made.ID, "field": "status", "value": "Nonsense",
	})
	if said == "" {
		t.Error("an impossible status was refused without saying why")
	}

	// And a note goes to the inbox rather than the board.
	call[written](t, session, "create_task", map[string]any{
		"project": "Alpha", "title": "A captured thought", "draft": true,
	})
	if _, err := os.Stat(filepath.Join(root, "alpha", "backlog", "drafts")); err != nil {
		t.Fatalf("no drafts directory: %v", err)
	}
	type entity struct {
		Title string `json:"title"`
	}
	drafts := call[[]entity](t, session, "list_entities", map[string]any{
		"kind": "draft", "project": "Alpha",
	})
	if len(drafts) != 1 || drafts[0].Title != "A captured thought" {
		t.Errorf("the inbox holds %+v", drafts)
	}
}

// A tool list an agent cannot read is a tool list it will misuse.
func TestEveryToolExplainsItself(t *testing.T) {
	session, _ := bench(t)

	tools, err := session.ListTools(context.Background(), nil)
	if err != nil {
		t.Fatalf("ListTools: %v", err)
	}
	if len(tools.Tools) < 8 {
		t.Fatalf("only %d tools", len(tools.Tools))
	}
	for _, tool := range tools.Tools {
		if len(tool.Description) < 40 {
			t.Errorf("%s is described as %q", tool.Name, tool.Description)
		}
		if tool.InputSchema == nil {
			t.Errorf("%s has no input schema", tool.Name)
		}
	}
}

func fileWith(t *testing.T, dir, match string) string {
	t.Helper()
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("reading %s: %v", dir, err)
	}
	for _, entry := range entries {
		if strings.Contains(entry.Name(), match) {
			raw, err := os.ReadFile(filepath.Join(dir, entry.Name()))
			if err != nil {
				t.Fatal(err)
			}
			return string(raw)
		}
	}
	t.Fatalf("nothing matching %q in %s", match, dir)
	return ""
}

// What a client is told, checked from the client's side rather than by reading
// the code that sets it. An instruction the client never receives is a
// comment.
func TestTheClientIsToldWhenToUseThisAndWhenNotTo(t *testing.T) {
	session, _ := bench(t)

	said := session.InitializeResult().Instructions
	switch {
	case said == "":
		t.Fatal("the server sends no instructions at all")
	case !strings.Contains(said, "every Backlog.md project"):
		t.Errorf("the instructions do not say what this is for: %q", said)
	case !strings.Contains(said, "backlog CLI"):
		t.Errorf("the instructions do not say what to use instead: %q", said)
	case !strings.Contains(said, "list_projects"):
		t.Errorf("the instructions do not say where to start: %q", said)
	}

	tools, err := session.ListTools(context.Background(), nil)
	if err != nil {
		t.Fatalf("ListTools: %v", err)
	}

	reads := map[string]bool{
		"list_projects": true, "list_tasks": true, "get_task": true,
		"search": true, "list_milestones": true, "list_entities": true,
	}
	for _, tool := range tools.Tools {
		t.Run(tool.Name, func(t *testing.T) {
			if tool.Annotations == nil {
				t.Fatalf("%s says nothing about whether it changes anything", tool.Name)
			}
			if reads[tool.Name] {
				if !tool.Annotations.ReadOnlyHint {
					t.Errorf("%s answers a question but is not marked read-only", tool.Name)
				}
				return
			}

			// A write, then. It must not claim to be read-only, and its
			// description has to bound itself: a tool description is read on
			// its own, so relying on the instructions having been read first
			// would be relying on luck.
			if tool.Annotations.ReadOnlyHint {
				t.Errorf("%s changes files but is marked read-only", tool.Name)
			}
			if !strings.Contains(tool.Description, "backlog CLI") {
				t.Errorf("%s does not say when to prefer the project's own CLI: %q",
					tool.Name, tool.Description)
			}
		})
	}
}

// An empty answer and a missing file are not the same thing, and an agent
// cannot tell them apart. Being told "no projects" when the registry was
// simply looked for in the wrong place — which is what a sandboxed client
// does, its XDG_CONFIG_HOME redirected — is a wrong answer, not an unhelpful
// one.
func TestAMissingRegistryIsSaidOutLoud(t *testing.T) {
	nowhere := filepath.Join(t.TempDir(), "muster", "projects.yml")

	server, err := mcpserver.New(nowhere)
	if err != nil {
		t.Fatalf("the server refused to start without a registry: %v", err)
	}

	// It still starts, because a client that cannot connect cannot be told
	// anything at all.
	clientTransport, serverTransport := mcp.NewInMemoryTransports()
	go func() { _ = server.MCP().Run(context.Background(), serverTransport) }()
	client := mcp.NewClient(&mcp.Implementation{Name: "test", Version: "1"}, nil)
	session, err := client.Connect(context.Background(), clientTransport, nil)
	if err != nil {
		t.Fatalf("Connect: %v", err)
	}
	t.Cleanup(func() { _ = session.Close() })

	for _, tool := range []string{"list_projects", "search", "list_milestones", "list_entities"} {
		args := map[string]any{}
		switch tool {
		case "search":
			args["text"] = "anything"
		case "list_entities":
			args["kind"] = "document"
		}
		res, err := session.CallTool(context.Background(), &mcp.CallToolParams{
			Name: tool, Arguments: args,
		})
		if err != nil {
			t.Fatalf("%s: %v", tool, err)
		}
		if !res.IsError {
			t.Errorf("%s answered as though nothing were wrong", tool)
			continue
		}
		said := textOf(res)
		if !strings.Contains(said, nowhere) {
			t.Errorf("%s does not say where it looked: %s", tool, said)
		}
		if !strings.Contains(said, "MUSTER_REGISTRY") {
			t.Errorf("%s does not say how to point it elsewhere: %s", tool, said)
		}
	}
}
