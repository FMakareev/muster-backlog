// Package agents connects an AI client to this board's MCP server.
//
// The point is to replace a paragraph of documentation with a button. Nothing
// here is clever: it runs the client's own command, or writes the client's own
// config file — and it shows the person exactly which one, in full, before doing
// either. A tool that edits other programs' configuration has to be boring and
// visible, and it has to be undoable: every config write leaves a backup, and
// every provider can be disconnected the same way it was connected.
package agents

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	_ "embed"

	"github.com/FMakareev/muster-backlog/internal/whichbin"
)

//go:embed agents.json
var builtinJSON []byte

// Provider is one AI client and how it is told about an MCP server.
type Provider struct {
	ID    string `json:"id"`
	Label string `json:"label"`
	// "cli": run the client's own command. "config": write its JSON config.
	Kind string `json:"kind"`

	// cli
	Bin           string   `json:"bin,omitempty"`
	Add           []string `json:"add,omitempty"`
	Remove        []string `json:"remove,omitempty"`
	Check         []string `json:"check,omitempty"`
	CheckContains string   `json:"checkContains,omitempty"`

	// config (JSON) and config (TOML): both write one entry into a file the
	// client already owns, and both leave everything else in it alone
	Paths    map[string]string `json:"paths,omitempty"`
	JSONPath []string          `json:"jsonPath,omitempty"`
	Value    map[string]any    `json:"value,omitempty"`
	Section  string            `json:"section,omitempty"`
	Lines    []string          `json:"lines,omitempty"`

	Note string `json:"note,omitempty"`
	Docs string `json:"docs,omitempty"`
}

type catalog struct {
	Name      string     `json:"name"`
	Providers []Provider `json:"providers"`
}

// Status is what the settings window shows before anything is pressed.
type Status struct {
	ID        string `json:"id"`
	Label     string `json:"label"`
	Kind      string `json:"kind"`
	Installed bool   `json:"installed"`
	Connected bool   `json:"connected"`
	/**
	 * Nobody has asked this one yet whether the board is in its configuration.
	 *
	 * Only a CLI client can be in this state, and only because asking means
	 * running it: `claude mcp list` costs four seconds on a plain machine, all of
	 * it node starting up. "Installed?" is a stat call; "connected?" is a
	 * subprocess. Reporting Connected: false before asking would be a lie with a
	 * button attached to it, so the two answers travel separately.
	 */
	Asking bool `json:"asking"`
	/** Where we looked: the binary we found, or the config file we read. */
	Detail string `json:"detail"`
	Path   string `json:"path,omitempty"`
	Note   string `json:"note,omitempty"`
	Docs   string `json:"docs,omitempty"`
}

// Plan is exactly what pressing the button will do. Shown first, always.
type Plan struct {
	ID      string `json:"id"`
	Kind    string `json:"kind"`
	Command string `json:"command,omitempty"`
	Path    string `json:"path,omitempty"`
	Diff    string `json:"diff,omitempty"`
	Note    string `json:"note,omitempty"`
	Error   string `json:"error,omitempty"`
}

// Result is what happened, including the part we could not do.
type Result struct {
	OK       bool   `json:"ok"`
	Output   string `json:"output,omitempty"`
	Verified bool   `json:"verified"`
	Backup   string `json:"backup,omitempty"`
	Error    string `json:"error,omitempty"`
	/** The same command, for a person who would rather run it themselves. */
	Command string `json:"command,omitempty"`
	/** Which config file was written, when one was. */
	Path string `json:"path,omitempty"`
}

const runTimeout = 30 * time.Second

// Catalog reads the provider list: the embedded one, or the file named by
// MUSTER_AGENTS_FILE — clients change their command syntax more often than we
// ship, and a data file can be fixed without a release.
func Catalog() (string, []Provider, error) {
	data := builtinJSON
	if p := os.Getenv("MUSTER_AGENTS_FILE"); p != "" {
		b, err := os.ReadFile(p)
		if err != nil {
			return "", nil, fmt.Errorf("agents file: %w", err)
		}
		data = b
	}
	var c catalog
	if err := json.Unmarshal(data, &c); err != nil {
		return "", nil, fmt.Errorf("agents file: %w", err)
	}
	if c.Name == "" {
		c.Name = "muster"
	}
	return c.Name, c.Providers, nil
}

// backupName is where the file a client had before is kept.
//
// Named after this application, in somebody else's directory: a stray file
// named after a program they have never installed reads as a bug in whatever
// they were using.
func backupName(path string) string {
	return fmt.Sprintf("%s.muster-backup-%s", path, time.Now().Format("20060102-150405"))
}

func expand(path string) string {
	if strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err == nil {
			return filepath.Join(home, path[2:])
		}
	}
	return path
}

func (p Provider) configPath() string {
	if len(p.Paths) == 0 {
		return ""
	}
	if v, ok := p.Paths[runtime.GOOS]; ok {
		return expand(v)
	}
	return ""
}

// fill substitutes into a catalogue entry.
//
// command is the absolute path to this binary, not a URL: refloft, where this
// module comes from, serves MCP over HTTP and registers a URL. Muster serves
// over stdio, so what a client needs is a command to spawn - and an absolute
// one, because a client will not have this on PATH any more than this had the
// Backlog.md CLI on its own.
func fill(s, name, command string) string {
	return strings.NewReplacer("{name}", name, "{command}", command).Replace(s)
}

func fillArgs(args []string, name, command string) []string {
	out := make([]string, 0, len(args))
	for _, a := range args {
		out = append(out, fill(a, name, command))
	}
	return out
}

func fillValue(v any, name, command string) any {
	switch t := v.(type) {
	case string:
		return fill(t, name, command)
	case []any:
		out := make([]any, len(t))
		for i, x := range t {
			out[i] = fillValue(x, name, command)
		}
		return out
	case map[string]any:
		out := map[string]any{}
		for k, x := range t {
			out[fill(k, name, command)] = fillValue(x, name, command)
		}
		return out
	default:
		return v
	}
}

func shellish(bin string, args []string) string {
	parts := make([]string, 0, len(args)+1)
	parts = append(parts, bin)
	for _, a := range args {
		if strings.ContainsAny(a, " \t\"'{}") {
			parts = append(parts, "'"+strings.ReplaceAll(a, "'", `'\''`)+"'")
			continue
		}
		parts = append(parts, a)
	}
	return strings.Join(parts, " ")
}

// binFor resolves a provider's command once, and answers two questions with one
// lookup: what to run, and what to show the person who is about to approve it.
//
// Those are the same word when the command is on PATH and a full path when it
// is not — see whichbin for why a packaged app so often has neither. The plan
// this window shows has to be the command it runs, and a bare "claude" printed
// on a machine where nothing can resolve "claude" would be neither true nor
// useful to copy.
func binFor(p Provider) (runBin, show string, err error) {
	path, onPATH, err := whichbin.Look(p.Bin)
	if err != nil {
		return "", p.Bin, err
	}
	if onPATH {
		return path, p.Bin, nil
	}
	return path, path, nil
}

// notFound says where we looked, because "not installed" next to a program the
// person installed themselves reads as a lie.
func notFound(bin string) string {
	return bin + " was not found — not on this app's PATH, and not where these tools usually install"
}

// List reports every provider we know about and whether it is installed here.
//
// It never runs anybody else's program. A CLI client's connection state is left
// for Connections to answer, because answering it means starting that client:
// measured on a plain machine, eight of the nine providers cost nothing and
// `claude mcp list` cost 4.1 seconds — and the whole list used to wait for it,
// including the eight answers that were already known.
func List(command string) ([]Status, error) {
	return statuses(command, false)
}

// Connections is List with the slow question asked as well: for every installed
// CLI client, whether this board is in its configuration.
//
// The clients are asked concurrently. Serially the cost was the sum, which is
// somebody else's four seconds multiplied by however many clients they happen to
// have installed; now it is the slowest one.
func Connections(command string) ([]Status, error) {
	return statuses(command, true)
}

func statuses(command string, ask bool) ([]Status, error) {
	name, providers, err := Catalog()
	if err != nil {
		return nil, err
	}
	// Sized up front and written by index, never appended to: the goroutines
	// below write into these elements, and an append that reallocated the array
	// under them would be a race — the race detector said so, and a plain test
	// run did not.
	out := make([]Status, len(providers))
	var wg sync.WaitGroup
	for i, p := range providers {
		s := Status{ID: p.ID, Label: p.Label, Kind: p.Kind, Note: p.Note, Docs: p.Docs}
		switch p.Kind {
		case "cli":
			bin, _, lookErr := binFor(p)
			s.Installed = lookErr == nil
			s.Path = bin
			if s.Installed {
				s.Detail = bin
				s.Asking = !ask
			} else {
				s.Detail = notFound(p.Bin)
			}
		case "config":
			path := p.configPath()
			s.Path = path
			if path == "" {
				s.Detail = "no known config location on this system"
				break
			}
			doc, readErr := readJSON(path)
			s.Installed = readErr == nil
			if readErr != nil {
				s.Detail = "not found: " + path
				break
			}
			s.Detail = path
			cur, _ := getPath(doc, fillArgs(p.JSONPath, name, command))
			s.Connected = cur != nil
		case "toml":
			path := p.configPath()
			s.Path = path
			if path == "" {
				s.Detail = "no known config location on this system"
				break
			}
			if _, statErr := os.Stat(path); statErr != nil {
				s.Detail = "not found: " + path
				break
			}
			s.Installed = true
			s.Detail = path
			s.Connected = tomlHas(path, fill(p.Section, name, command))
		default:
			s.Detail = "unknown provider kind: " + p.Kind
		}
		out[i] = s
		if ask && p.Kind == "cli" && s.Installed {
			// each goroutine writes the slot its own provider owns, so the
			// answers cannot land in a different order than the list they
			// belong to, and no two of them touch the same element
			at, bin, prov := i, s.Path, p
			wg.Add(1)
			go func() {
				defer wg.Done()
				out[at].Connected = cliConnected(bin, prov, name, command)
			}()
		}
	}
	wg.Wait()
	return out, nil
}

func cliConnected(bin string, p Provider, name, command string) bool {
	if len(p.Check) == 0 {
		return false
	}
	out, err := run(bin, fillArgs(p.Check, name, command))
	if err != nil {
		return false
	}
	needle := fill(p.CheckContains, name, command)
	if needle == "" {
		needle = name
	}
	return strings.Contains(out, needle)
}

func find(id string) (string, Provider, error) {
	name, providers, err := Catalog()
	if err != nil {
		return "", Provider{}, err
	}
	for _, p := range providers {
		if p.ID == id {
			return name, p, nil
		}
	}
	return "", Provider{}, fmt.Errorf("no provider called %q", id)
}

// PlanFor is what the button would do, in full, before it does it.
func PlanFor(id, command string, disconnect bool) Plan {
	name, p, err := find(id)
	if err != nil {
		return Plan{ID: id, Error: err.Error()}
	}
	plan := Plan{ID: id, Kind: p.Kind, Note: p.Note}
	switch p.Kind {
	case "cli":
		args := p.Add
		if disconnect {
			args = p.Remove
		}
		if len(args) == 0 {
			plan.Error = "this client has no command for that — see its own settings"
			return plan
		}
		_, show, lookErr := binFor(p)
		if lookErr != nil {
			plan.Error = notFound(p.Bin)
			return plan
		}
		plan.Command = shellish(show, fillArgs(args, name, command))
	case "toml", "config":
		path := p.configPath()
		plan.Path = path
		if path == "" {
			plan.Error = "no known config location on this system"
			return plan
		}
		before, after, err := changeFor(p, name, command, disconnect)
		if err != nil {
			plan.Error = err.Error()
			return plan
		}
		plan.Diff = diff(before, after)
		if plan.Diff == "" {
			plan.Note = strings.TrimSpace(plan.Note + " Nothing to change: it is already like that.")
		}
	default:
		plan.Error = "unknown provider kind: " + p.Kind
	}
	return plan
}

// Apply connects (or disconnects) one client. It never runs anything the plan
// did not show, and it verifies afterwards instead of assuming.
func Apply(id, command string, disconnect bool) Result {
	name, p, err := find(id)
	if err != nil {
		return Result{Error: err.Error()}
	}
	switch p.Kind {
	case "cli":
		args := p.Add
		if disconnect {
			args = p.Remove
		}
		if len(args) == 0 {
			return Result{Error: "this client has no command for that — see its own settings"}
		}
		runBin, show, lookErr := binFor(p)
		if lookErr != nil {
			return Result{Error: notFound(p.Bin)}
		}
		filled := fillArgs(args, name, command)
		cmd := shellish(show, filled)
		out, runErr := run(runBin, filled)
		res := Result{Output: strings.TrimSpace(out), Command: cmd}
		if runErr != nil {
			res.Error = runErr.Error()
			return res
		}
		res.OK = true
		connected := cliConnected(runBin, p, name, command)
		res.Verified = connected != disconnect
		if !res.Verified {
			res.Error = "the command finished, but the client does not list the board — run it yourself and see what it says"
		}
		return res
	case "toml", "config":
		path := p.configPath()
		if path == "" {
			return Result{Error: "no known config location on this system"}
		}
		before, after, err := changeFor(p, name, command, disconnect)
		if err != nil {
			return Result{Error: err.Error()}
		}
		res := Result{Path: path}
		if before == after {
			res.OK, res.Verified = true, true
			res.Output = "already like that — nothing was written"
			return res
		}
		backup, err := writeJSONWithBackup(path, after)
		if err != nil {
			res.Error = err.Error()
			return res
		}
		res.OK = true
		res.Backup = backup
		// read it back rather than trusting the write
		if p.Kind == "toml" {
			res.Verified = tomlHas(path, fill(p.Section, name, command)) != disconnect
			if !res.Verified {
				res.Error = "the file was written but does not contain what we expected"
			}
			return res
		}
		doc, err := readJSON(path)
		if err != nil {
			res.Error = "written, but could not read it back: " + err.Error()
			return res
		}
		cur, _ := getPath(doc, fillArgs(p.JSONPath, name, command))
		res.Verified = (cur != nil) != disconnect
		if !res.Verified {
			res.Error = "the file was written but does not contain what we expected"
		}
		return res
	}
	return Result{Error: "unknown provider kind: " + p.Kind}
}

func run(bin string, args []string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), runTimeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, bin, args...)
	out, err := cmd.CombinedOutput()
	text := string(out)
	if ctx.Err() != nil {
		return text, fmt.Errorf("%s took longer than %s and was stopped", bin, runTimeout)
	}
	if err != nil {
		var ee *exec.ExitError
		if errors.As(err, &ee) {
			return text, fmt.Errorf("%s exited with %d: %s", bin, ee.ExitCode(), strings.TrimSpace(firstLines(text, 3)))
		}
		return text, err
	}
	return text, nil
}

func firstLines(s string, n int) string {
	lines := strings.Split(s, "\n")
	if len(lines) > n {
		lines = lines[:n]
	}
	return strings.Join(lines, "\n")
}

// ---- config files ----------------------------------------------------------

func readJSON(path string) (map[string]any, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	if len(strings.TrimSpace(string(b))) == 0 {
		return map[string]any{}, nil
	}
	var doc map[string]any
	if err := json.Unmarshal(b, &doc); err != nil {
		// Someone else's file we cannot parse is one we do not touch: a config
		// with comments (JSONC) or a syntax error would be destroyed by a
		// rewrite, and the person would lose settings we never even read.
		return nil, fmt.Errorf("cannot read %s as JSON (comments and trailing commas are not supported here) — add the server by hand", filepath.Base(path))
	}
	return doc, nil
}

/** The file as it is and as it would be, whichever kind of config it is. */
func changeFor(p Provider, name, command string, remove bool) (string, string, error) {
	if p.Kind == "toml" {
		return tomlChange(p.configPath(), fill(p.Section, name, command), fillArgs(p.Lines, name, command), remove)
	}
	return configChange(p, name, command, remove)
}

/** The file as it is and as it would be, both pretty-printed. */
func configChange(p Provider, name, command string, remove bool) (string, string, error) {
	path := p.configPath()
	doc, err := readJSON(path)
	if os.IsNotExist(err) {
		doc = map[string]any{}
	} else if err != nil && !os.IsNotExist(err) {
		if _, statErr := os.Stat(path); statErr == nil {
			return "", "", err // exists but unreadable: refuse
		}
		doc = map[string]any{}
	}
	before, err := marshal(doc)
	if err != nil {
		return "", "", err
	}
	// A file that does not exist yet (or is empty) should read as additions
	// only: "- {}" in a diff looks like we are taking something away.
	if info, statErr := os.Stat(path); statErr != nil || info.Size() == 0 {
		before = ""
	}
	next := deepCopy(doc)
	keys := fillArgs(p.JSONPath, name, command)
	if remove {
		deletePath(next, keys)
	} else {
		setPath(next, keys, fillValue(p.Value, name, command))
	}
	after, err := marshal(next)
	if err != nil {
		return "", "", err
	}
	return before, after, nil
}

func marshal(v any) (string, error) {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return "", err
	}
	return string(b) + "\n", nil
}

func deepCopy(v map[string]any) map[string]any {
	b, err := json.Marshal(v)
	if err != nil {
		return map[string]any{}
	}
	var out map[string]any
	if err := json.Unmarshal(b, &out); err != nil {
		return map[string]any{}
	}
	return out
}

func getPath(doc map[string]any, keys []string) (any, bool) {
	cur := any(doc)
	for _, k := range keys {
		m, ok := cur.(map[string]any)
		if !ok {
			return nil, false
		}
		cur, ok = m[k]
		if !ok {
			return nil, false
		}
	}
	return cur, true
}

func setPath(doc map[string]any, keys []string, value any) {
	cur := doc
	for i, k := range keys {
		if i == len(keys)-1 {
			cur[k] = value
			return
		}
		next, ok := cur[k].(map[string]any)
		if !ok {
			next = map[string]any{}
			cur[k] = next
		}
		cur = next
	}
}

func deletePath(doc map[string]any, keys []string) {
	cur := doc
	for i, k := range keys {
		if i == len(keys)-1 {
			delete(cur, k)
			return
		}
		next, ok := cur[k].(map[string]any)
		if !ok {
			return
		}
		cur = next
	}
}

/**
 * Write the file, leaving the previous version beside it.
 *
 * This is somebody else's configuration. The backup is not politeness: it is the
 * difference between "undo it" and "hope you remember what was in there".
 */
func writeJSONWithBackup(path, content string) (string, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return "", err
	}
	backup := ""
	if old, err := os.ReadFile(path); err == nil {
		backup = backupName(path)
		if err := os.WriteFile(backup, old, 0o644); err != nil {
			return "", fmt.Errorf("could not back up %s: %w", filepath.Base(path), err)
		}
	}
	tmp := path + ".muster-tmp"
	if err := os.WriteFile(tmp, []byte(content), 0o644); err != nil {
		return backup, err
	}
	if err := os.Rename(tmp, path); err != nil {
		return backup, err
	}
	return backup, nil
}

// ---- a small diff ----------------------------------------------------------

/**
 * A line diff of before and after.
 *
 * Not a general one: these two texts differ in a handful of lines, so trimming
 * the common head and tail says everything that matters, and says it in a shape
 * people recognise from git.
 */
func diff(before, after string) string {
	if before == after {
		return ""
	}
	a := strings.Split(strings.TrimRight(before, "\n"), "\n")
	b := strings.Split(strings.TrimRight(after, "\n"), "\n")
	head := 0
	for head < len(a) && head < len(b) && a[head] == b[head] {
		head++
	}
	tail := 0
	for tail < len(a)-head && tail < len(b)-head && a[len(a)-1-tail] == b[len(b)-1-tail] {
		tail++
	}
	var sb strings.Builder
	ctx := 2
	from := max(0, head-ctx)
	for _, l := range a[from:head] {
		sb.WriteString("  " + l + "\n")
	}
	for _, l := range a[head : len(a)-tail] {
		sb.WriteString("- " + l + "\n")
	}
	for _, l := range b[head : len(b)-tail] {
		sb.WriteString("+ " + l + "\n")
	}
	to := min(len(a), len(a)-tail+ctx)
	for _, l := range a[len(a)-tail : to] {
		sb.WriteString("  " + l + "\n")
	}
	return sb.String()
}
