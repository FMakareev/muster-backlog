package agents

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

// A catalog with one of each kind, pointed at a temp directory: real clients are
// not installed on a build machine, and a test that only runs on the author's
// laptop tests nothing.
func setup(t *testing.T, home string) {
	t.Helper()
	bin := filepath.Join(home, "bin")
	if err := os.MkdirAll(bin, 0o755); err != nil {
		t.Fatal(err)
	}
	// a stub client that records what it was told and answers "list"
	script := "#!/bin/sh\n" +
		"echo \"$@\" >> " + filepath.Join(home, "calls.log") + "\n" +
		"case \"$1 $2\" in\n" +
		"  'mcp list') cat " + filepath.Join(home, "servers.txt") + " 2>/dev/null || true;;\n" +
		"  'mcp add') echo added; echo muster >> " + filepath.Join(home, "servers.txt") + ";;\n" +
		"  'mcp remove') echo removed; rm -f " + filepath.Join(home, "servers.txt") + ";;\n" +
		"  *) echo 'unknown command'; exit 2;;\n" +
		"esac\n"
	if err := os.WriteFile(filepath.Join(bin, "stubclient"), []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))
	t.Setenv("HOME", home)

	cat := map[string]any{
		"name": "muster",
		"providers": []map[string]any{
			{
				"id": "stub-cli", "label": "Stub CLI", "kind": "cli", "bin": "stubclient",
				"add":    []string{"mcp", "add", "{name}", "--", "{command}", "mcp"},
				"remove": []string{"mcp", "remove", "{name}"},
				"check":  []string{"mcp", "list"}, "checkContains": "{name}",
			},
			{
				"id": "stub-config", "label": "Stub Config", "kind": "config",
				"paths": map[string]string{
					"linux": filepath.Join(home, "client.json"), "darwin": filepath.Join(home, "client.json"),
					"windows": filepath.Join(home, "client.json"),
				},
				"jsonPath": []string{"mcpServers", "{name}"},
				"value":    map[string]any{"command": "{command}"},
			},
			{
				"id": "stub-missing", "label": "Not Installed", "kind": "cli", "bin": "no-such-client-anywhere",
				"add": []string{"add"},
			},
		},
	}
	b, _ := json.MarshalIndent(cat, "", " ")
	file := filepath.Join(home, "agents.json")
	if err := os.WriteFile(file, b, 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("MUSTER_AGENTS_FILE", file)
}

// What a client is told to spawn: this binary, by its full path.
const command = "/opt/muster/bin/muster"

// The full answer, slow question included: most tests here are about whether the
// board ended up in somebody's configuration, and only Connections asks that.
func statusOf(t *testing.T, id string) Status {
	t.Helper()
	list, err := Connections(command)
	if err != nil {
		t.Fatal(err)
	}
	for _, s := range list {
		if s.ID == id {
			return s
		}
	}
	t.Fatalf("no status for %q", id)
	return Status{}
}

func TestTheBuiltInCatalogIsUsable(t *testing.T) {
	// the file we ship must parse and describe every provider well enough to act
	name, providers, err := Catalog()
	if err != nil {
		t.Fatal(err)
	}
	if name == "" || len(providers) == 0 {
		t.Fatal("the shipped catalog is empty")
	}
	for _, p := range providers {
		if p.ID == "" || p.Label == "" {
			t.Errorf("a provider has no id or label: %+v", p)
		}
		switch p.Kind {
		case "cli":
			if p.Bin == "" || len(p.Add) == 0 {
				t.Errorf("%s: a cli provider needs a binary and an add command", p.ID)
			}
		case "config":
			if len(p.Paths) == 0 || len(p.JSONPath) == 0 || len(p.Value) == 0 {
				t.Errorf("%s: a config provider needs a path, a place in the JSON and a value", p.ID)
			}
			if _, ok := p.Paths[runtime.GOOS]; !ok && runtime.GOOS != "js" {
				t.Errorf("%s: no config location for %s", p.ID, runtime.GOOS)
			}
		case "toml":
			if len(p.Paths) == 0 || p.Section == "" || len(p.Lines) == 0 {
				t.Errorf("%s: a toml provider needs a path, a table and its lines", p.ID)
			}
			if _, ok := p.Paths[runtime.GOOS]; !ok && runtime.GOOS != "js" {
				t.Errorf("%s: no config location for %s", p.ID, runtime.GOOS)
			}
		default:
			t.Errorf("%s: unknown kind %q", p.ID, p.Kind)
		}
	}
}

func TestDetectionSaysWhatItLookedFor(t *testing.T) {
	setup(t, t.TempDir())
	installed := statusOf(t, "stub-cli")
	if !installed.Installed || installed.Connected {
		t.Errorf("a client on PATH with nothing registered: %+v", installed)
	}
	missing := statusOf(t, "stub-missing")
	if missing.Installed {
		t.Error("a client that is not there must not be reported as installed")
	}
	if !strings.Contains(missing.Detail, "PATH") {
		t.Errorf("the reason should say where we looked: %q", missing.Detail)
	}
}

// The failure a packaged build actually had: a launcher hands the app a PATH of
// system directories, the client is installed under $HOME, and the settings
// window calls it "not installed" while the person is looking at it in their
// terminal.
func TestAClientOffPATHIsStillFoundAndStillWorks(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("the stub client is a shell script")
	}
	home := t.TempDir()
	setup(t, home)
	// exactly what a .desktop entry gives you: no ~/bin, no ~/.local/bin
	t.Setenv("PATH", "/usr/bin"+string(os.PathListSeparator)+"/bin")

	want := filepath.Join(home, "bin", "stubclient")
	s := statusOf(t, "stub-cli")
	if !s.Installed {
		t.Fatalf("a client under $HOME must still be found: %+v", s)
	}
	if s.Path != want || s.Detail != want {
		t.Errorf("the status must say where it was found, in full: %+v", s)
	}
	// and what is shown must be runnable as shown, which the bare name is not
	plan := PlanFor("stub-cli", command, false)
	if !strings.HasPrefix(plan.Command, want+" mcp add") {
		t.Errorf("the command shown must be the one that will run: %q", plan.Command)
	}
	if res := Apply("stub-cli", command, false); !res.OK || !res.Verified {
		t.Fatalf("connecting to a client off PATH failed: %+v", res)
	}
	if !statusOf(t, "stub-cli").Connected {
		t.Error("after connecting, the status must say connected")
	}
}

func TestThePlanIsExactlyWhatWillRun(t *testing.T) {
	home := t.TempDir()
	setup(t, home)
	plan := PlanFor("stub-cli", command, false)
	if plan.Error != "" {
		t.Fatal(plan.Error)
	}
	if !strings.Contains(plan.Command, "stubclient mcp add") || !strings.Contains(plan.Command, command) {
		t.Errorf("the command shown must be the command run: %q", plan.Command)
	}
	// showing a plan must not do anything
	if _, err := os.Stat(filepath.Join(home, "calls.log")); err == nil {
		t.Error("planning ran the client")
	}
}

func TestConnectingRunsTheClientAndChecksAfterwards(t *testing.T) {
	home := t.TempDir()
	setup(t, home)
	res := Apply("stub-cli", command, false)
	if !res.OK || !res.Verified {
		t.Fatalf("connect failed: %+v", res)
	}
	calls, _ := os.ReadFile(filepath.Join(home, "calls.log"))
	if !strings.Contains(string(calls), "mcp add muster -- "+command+" mcp") {
		t.Errorf("the client was called with something else: %q", calls)
	}
	if !statusOf(t, "stub-cli").Connected {
		t.Error("after connecting, the status must say connected")
	}
	// and disconnecting is the same button in reverse
	out := Apply("stub-cli", command, true)
	if !out.OK || !out.Verified {
		t.Fatalf("disconnect failed: %+v", out)
	}
	if statusOf(t, "stub-cli").Connected {
		t.Error("after disconnecting, the status must say not connected")
	}
}

func TestAFailedCommandIsReportedWithItsOutput(t *testing.T) {
	home := t.TempDir()
	setup(t, home)
	// point the add command at something the stub refuses
	file := filepath.Join(home, "agents.json")
	b, err := os.ReadFile(file)
	if err != nil {
		t.Fatal(err)
	}
	var cat map[string]any
	if err := json.Unmarshal(b, &cat); err != nil {
		t.Fatal(err)
	}
	provs := cat["providers"].([]any)
	provs[0].(map[string]any)["add"] = []any{"nope", "nope"}
	nb, err := json.Marshal(cat)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(file, nb, 0o644); err != nil {
		t.Fatal(err)
	}

	res := Apply("stub-cli", command, false)
	if res.OK {
		t.Fatal("a failing command must not be reported as success")
	}
	if res.Error == "" || res.Command == "" {
		t.Errorf("a failure must carry the reason and the command to try by hand: %+v", res)
	}
}

func TestConfigWriteShowsADiffFirstAndKeepsABackup(t *testing.T) {
	home := t.TempDir()
	setup(t, home)
	path := filepath.Join(home, "client.json")
	existing := `{
  "mcpServers": {
    "somethingElse": { "command": "http://example.com" }
  },
  "theme": "dark"
}`
	if err := os.WriteFile(path, []byte(existing), 0o644); err != nil {
		t.Fatal(err)
	}

	plan := PlanFor("stub-config", command, false)
	if plan.Diff == "" || !strings.Contains(plan.Diff, command) {
		t.Fatalf("the diff must show what will be added: %q", plan.Diff)
	}
	if strings.Contains(plan.Diff, "- ") && strings.Contains(plan.Diff, "theme") {
		t.Errorf("the diff claims to touch settings it does not: %q", plan.Diff)
	}
	after, _ := os.ReadFile(path)
	if string(after) != existing {
		t.Error("planning wrote to the file")
	}

	res := Apply("stub-config", command, false)
	if !res.OK || !res.Verified {
		t.Fatalf("config write failed: %+v", res)
	}
	if res.Backup == "" {
		t.Fatal("somebody else's config was rewritten with no backup")
	}
	old, err := os.ReadFile(res.Backup)
	if err != nil || string(old) != existing {
		t.Errorf("the backup does not hold the previous file: %v", err)
	}
	doc, err := readJSON(path)
	if err != nil {
		t.Fatal(err)
	}
	servers := doc["mcpServers"].(map[string]any)
	if _, ok := servers["somethingElse"]; !ok {
		t.Error("their other server was dropped")
	}
	if doc["theme"] != "dark" {
		t.Error("their unrelated settings were dropped")
	}
	if servers["muster"].(map[string]any)["command"] != command {
		t.Errorf("our server is not in the file: %v", servers["muster"])
	}
}

func TestConfigWriteCreatesTheFileWhenThereIsNone(t *testing.T) {
	home := t.TempDir()
	setup(t, home)
	res := Apply("stub-config", command, false)
	if !res.OK || !res.Verified {
		t.Fatalf("%+v", res)
	}
	if res.Backup != "" {
		t.Error("nothing existed, so nothing should have been backed up")
	}
	if !statusOf(t, "stub-config").Connected {
		t.Error("the new file does not report as connected")
	}
}

func TestConnectingTwiceChangesNothingTheSecondTime(t *testing.T) {
	home := t.TempDir()
	setup(t, home)
	Apply("stub-config", command, false)
	before, _ := os.ReadFile(filepath.Join(home, "client.json"))
	plan := PlanFor("stub-config", command, false)
	if plan.Diff != "" {
		t.Errorf("a second connect wants to change something: %q", plan.Diff)
	}
	res := Apply("stub-config", command, false)
	after, _ := os.ReadFile(filepath.Join(home, "client.json"))
	if string(before) != string(after) {
		t.Error("the file was rewritten for no reason")
	}
	if !res.OK || res.Backup != "" {
		t.Errorf("an unnecessary write left a backup: %+v", res)
	}
}

func TestDisconnectingLeavesTheRestOfTheirConfigAlone(t *testing.T) {
	home := t.TempDir()
	setup(t, home)
	path := filepath.Join(home, "client.json")
	if err := os.WriteFile(path,
		[]byte(`{"mcpServers":{"other":{"command":"http://x"}},"fontSize":14}`),
		0o644); err != nil {
		t.Fatal(err)
	}
	if res := Apply("stub-config", command, false); !res.OK {
		t.Fatalf("connecting first: %+v", res)
	}
	res := Apply("stub-config", command, true)
	if !res.OK || !res.Verified {
		t.Fatalf("%+v", res)
	}
	doc, _ := readJSON(path)
	servers := doc["mcpServers"].(map[string]any)
	if _, ok := servers["muster"]; ok {
		t.Error("we are still in their config")
	}
	if _, ok := servers["other"]; !ok {
		t.Error("their other server went with us")
	}
	if doc["fontSize"] != float64(14) {
		t.Error("their settings went with us")
	}
}

func TestAConfigWeCannotParseIsNotTouched(t *testing.T) {
	home := t.TempDir()
	setup(t, home)
	path := filepath.Join(home, "client.json")
	// JSONC: legal in several of these clients, and not something to rewrite
	weird := "{\n  // my servers\n  \"mcpServers\": {}\n}"
	if err := os.WriteFile(path, []byte(weird), 0o644); err != nil {
		t.Fatal(err)
	}

	plan := PlanFor("stub-config", command, false)
	if plan.Error == "" {
		t.Fatal("a file we cannot read must be refused, not rewritten")
	}
	res := Apply("stub-config", command, false)
	if res.OK {
		t.Fatal("we wrote a file we could not read")
	}
	now, _ := os.ReadFile(path)
	if string(now) != weird {
		t.Error("their file was changed anyway")
	}
	if !strings.Contains(res.Error, "by hand") {
		t.Errorf("the refusal should tell them what to do instead: %q", res.Error)
	}
}

func TestUnknownProviderIsAnError(t *testing.T) {
	setup(t, t.TempDir())
	if PlanFor("no-such-client", command, false).Error == "" {
		t.Error("planning for a client we do not know must fail loudly")
	}
	if Apply("no-such-client", command, false).Error == "" {
		t.Error("applying for a client we do not know must fail loudly")
	}
}

func TestDiffShowsOnlyWhatChanges(t *testing.T) {
	before := "{\n  \"a\": 1,\n  \"b\": 2,\n  \"c\": 3\n}\n"
	after := "{\n  \"a\": 1,\n  \"b\": 22,\n  \"c\": 3\n}\n"
	d := diff(before, after)
	if !strings.Contains(d, "- ") || !strings.Contains(d, "+ ") {
		t.Fatalf("no change shown: %q", d)
	}
	if strings.Count(d, "-") > 2 {
		t.Errorf("the diff exaggerates: %q", d)
	}
	if diff(before, before) != "" {
		t.Error("identical files must produce no diff")
	}
}

// ---- TOML: one table in a file full of somebody else's settings ------------

func tomlSetup(t *testing.T, home, existing string) string {
	t.Helper()
	path := filepath.Join(home, "config.toml")
	if existing != "" {
		if err := os.WriteFile(path, []byte(existing), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	cat := map[string]any{
		"name": "muster",
		"providers": []map[string]any{{
			"id": "stub-toml", "label": "Stub TOML", "kind": "toml",
			"paths":   map[string]string{"linux": path, "darwin": path, "windows": path},
			"section": "mcp_servers.{name}",
			"lines":   []string{`command = "{command}"`},
		}},
	}
	b, _ := json.MarshalIndent(cat, "", " ")
	file := filepath.Join(home, "agents.json")
	if err := os.WriteFile(file, b, 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("MUSTER_AGENTS_FILE", file)
	t.Setenv("HOME", home)
	return path
}

const codexish = `personality = "pragmatic"
model = "gpt-5.3-codex"

[projects."/home/me/work"]
trust_level = "trusted"

[mcp_servers.figma]
command = "https://mcp.figma.com/mcp"
`

func TestTomlAddsOurTableAndLeavesTheRestAlone(t *testing.T) {
	home := t.TempDir()
	path := tomlSetup(t, home, codexish)

	plan := PlanFor("stub-toml", command, false)
	if plan.Error != "" {
		t.Fatal(plan.Error)
	}
	if !strings.Contains(plan.Diff, "mcp_servers.muster") || !strings.Contains(plan.Diff, command) {
		t.Fatalf("the diff does not show the table being added: %q", plan.Diff)
	}
	if strings.Contains(plan.Diff, "- ") {
		t.Errorf("adding a table should remove nothing: %q", plan.Diff)
	}

	res := Apply("stub-toml", command, false)
	if !res.OK || !res.Verified {
		t.Fatalf("%+v", res)
	}
	now, _ := os.ReadFile(path)
	text := string(now)
	for _, keep := range []string{`personality = "pragmatic"`, `model = "gpt-5.3-codex"`,
		`[projects."/home/me/work"]`, `trust_level = "trusted"`,
		"[mcp_servers.figma]", `command = "https://mcp.figma.com/mcp"`} {
		if !strings.Contains(text, keep) {
			t.Errorf("their own settings were lost: %q is gone", keep)
		}
	}
	if !strings.Contains(text, "[mcp_servers.muster]") || !strings.Contains(text, command) {
		t.Errorf("our table is not in the file:\n%s", text)
	}
	if res.Backup == "" {
		t.Error("their file was rewritten with no backup")
	}
}

func TestTomlReplacesOurOwnTableRatherThanRepeatingIt(t *testing.T) {
	home := t.TempDir()
	path := tomlSetup(t, home, codexish+"\n[mcp_servers.muster]\nurl = \"http://127.0.0.1:1111\"\n")

	res := Apply("stub-toml", command, false)
	if !res.OK || !res.Verified {
		t.Fatalf("%+v", res)
	}
	text, _ := os.ReadFile(path)
	if n := strings.Count(string(text), "[mcp_servers.muster]"); n != 1 {
		t.Errorf("the table appears %d times", n)
	}
	if strings.Contains(string(text), "1111") {
		t.Errorf("the old address survived:\n%s", text)
	}
	if !strings.Contains(string(text), "[mcp_servers.figma]") {
		t.Error("their other server was eaten")
	}
}

func TestTomlDisconnectRemovesOnlyOurTable(t *testing.T) {
	home := t.TempDir()
	path := tomlSetup(t, home, codexish)
	Apply("stub-toml", command, false)
	res := Apply("stub-toml", command, true)
	if !res.OK || !res.Verified {
		t.Fatalf("%+v", res)
	}
	text, _ := os.ReadFile(path)
	if strings.Contains(string(text), "muster") {
		t.Errorf("we are still in their config:\n%s", text)
	}
	if !strings.Contains(string(text), "[mcp_servers.figma]") || !strings.Contains(string(text), `personality = "pragmatic"`) {
		t.Errorf("something else went with us:\n%s", text)
	}
}

func TestTomlCreatesTheFileWhenThereIsNone(t *testing.T) {
	home := t.TempDir()
	path := tomlSetup(t, home, "")
	if s := statusOf(t, "stub-toml"); s.Installed {
		t.Error("a client with no config file at all should not read as installed")
	}
	res := Apply("stub-toml", command, false)
	if !res.OK || !res.Verified {
		t.Fatalf("%+v", res)
	}
	text, _ := os.ReadFile(path)
	if !strings.HasPrefix(string(text), "[mcp_servers.muster]") {
		t.Errorf("a fresh file should hold just our table:\n%s", text)
	}
}

func TestTomlNoticesTheTableInEitherSpelling(t *testing.T) {
	home := t.TempDir()
	path := tomlSetup(t, home, "[mcp_servers.\"muster\"]\nurl = \"http://127.0.0.1:8765\"\n")
	if !tomlHas(path, "mcp_servers.muster") {
		t.Error("a quoted table name was not recognised")
	}
}

// ---- the fast half and the slow half -----------------------------------------

func listOf(t *testing.T, id string) Status {
	t.Helper()
	list, err := List(command)
	if err != nil {
		t.Fatal(err)
	}
	for _, s := range list {
		if s.ID == id {
			return s
		}
	}
	t.Fatalf("no status for %q", id)
	return Status{}
}

// Listing the clients must not run any of them.
//
// Measured on a plain machine before this split: eight of the nine providers
// cost nothing and `claude mcp list` cost 4.1 seconds, and the whole list waited
// for it. The stub client writes a line every time it is invoked, so the file it
// writes to is the evidence.
func TestListingClientsRunsNoneOfThem(t *testing.T) {
	home := t.TempDir()
	setup(t, home)
	if _, err := List(command); err != nil {
		t.Fatal(err)
	}
	if b, err := os.ReadFile(filepath.Join(home, "calls.log")); err == nil && len(b) > 0 {
		t.Fatalf("List ran the client: %s", b)
	}
	s := listOf(t, "stub-cli")
	if !s.Installed {
		t.Fatal("the client is installed and List must say so without running it")
	}
	if !s.Asking {
		t.Fatal("List must say the connection question has not been asked yet")
	}
	if s.Connected {
		t.Fatal("List must not claim to know whether the board is connected")
	}
}

func TestAskingTheClientsIsWhatRunsThem(t *testing.T) {
	home := t.TempDir()
	setup(t, home)
	list, err := Connections(command)
	if err != nil {
		t.Fatal(err)
	}
	b, readErr := os.ReadFile(filepath.Join(home, "calls.log"))
	if readErr != nil || len(b) == 0 {
		t.Fatal("Connections must ask the client, by running it")
	}
	if !strings.Contains(string(b), "mcp list") {
		t.Fatalf("expected the check command, got %q", b)
	}
	for _, s := range list {
		if s.Asking {
			t.Fatalf("%s: Connections leaves nothing unasked", s.ID)
		}
	}
}

// A client that is not installed is never run, and never left "checking".
func TestAMissingClientIsNotAsked(t *testing.T) {
	setup(t, t.TempDir())
	for _, s := range []Status{listOf(t, "stub-missing"), statusOf(t, "stub-missing")} {
		if s.Installed || s.Asking || s.Connected {
			t.Fatalf("a missing client should be plainly missing: %+v", s)
		}
	}
}

// A config client is a file read, so the fast half answers it in full.
func TestAConfigClientIsAnsweredByTheFastHalf(t *testing.T) {
	setup(t, t.TempDir())
	if s := listOf(t, "stub-config"); s.Asking {
		t.Fatal("reading a file is not a question worth deferring")
	}
	if res := Apply("stub-config", command, false); !res.OK {
		t.Fatalf("could not connect the config client: %s", res.Error)
	}
	if !listOf(t, "stub-config").Connected {
		t.Fatal("List must answer for a client it can answer for")
	}
}

// The clients are asked at the same time, not one after another.
//
// Serially the cost was the sum: somebody else's four seconds multiplied by
// however many clients they happen to have installed. Three stubs that each
// sleep 300ms make the difference visible without making the suite slow.
func TestClientsAreAskedAtTheSameTime(t *testing.T) {
	home := t.TempDir()
	setup(t, home)
	bin := filepath.Join(home, "bin")

	providers := []map[string]any{}
	for i := 0; i < 3; i++ {
		name := fmt.Sprintf("slowclient%d", i)
		if err := os.WriteFile(filepath.Join(bin, name),
			[]byte("#!/bin/sh\nsleep 0.3\necho muster\n"), 0o755); err != nil {
			t.Fatal(err)
		}
		providers = append(providers, map[string]any{
			"id": name, "label": name, "kind": "cli", "bin": name,
			"add":   []string{"mcp", "add"},
			"check": []string{"mcp", "list"}, "checkContains": "muster",
		})
	}
	b, _ := json.MarshalIndent(map[string]any{"name": "muster", "providers": providers}, "", " ")
	file := filepath.Join(home, "slow.json")
	if err := os.WriteFile(file, b, 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("MUSTER_AGENTS_FILE", file)

	start := time.Now()
	list, err := Connections(command)
	if err != nil {
		t.Fatal(err)
	}
	took := time.Since(start)
	if len(list) != 3 {
		t.Fatalf("expected three clients, got %d", len(list))
	}
	for _, s := range list {
		if !s.Connected {
			t.Fatalf("%s: the stub says connected", s.ID)
		}
	}
	// three at 300ms each: together well under 900ms, in turn well over
	if took > 700*time.Millisecond {
		t.Fatalf("asked one after another: %s for three clients of 300ms", took.Round(time.Millisecond))
	}
	t.Logf("three clients of 300ms answered in %s", took.Round(time.Millisecond))
}
