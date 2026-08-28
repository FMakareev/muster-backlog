package agents

import (
	"encoding/json"
	"strings"
	"testing"
)

// The catalogue that ships is the part most likely to be wrong, because every
// client owns its own syntax and changes it without asking. These are the
// checks that do not depend on any one client being right.
func TestTheShippedCatalogueRegistersACommand(t *testing.T) {
	name, providers, err := Catalog()
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	if name != "muster" {
		t.Errorf("the server is registered as %q", name)
	}
	if len(providers) < 5 {
		t.Fatalf("only %d clients", len(providers))
	}

	const binary = "/opt/muster/bin/muster"
	for _, p := range providers {
		t.Run(p.ID, func(t *testing.T) {
			if p.Label == "" || p.Docs == "" {
				t.Errorf("%s has no label or no documentation link", p.ID)
			}

			// Whatever it writes or runs has to name this binary and ask for
			// the mcp subcommand. Refloft, where this came from, serves over
			// HTTP and registers a URL; nothing here should still do that.
			var text string
			switch p.Kind {
			case "cli":
				if p.Bin == "" || len(p.Add) == 0 {
					t.Fatalf("%s is a cli provider with nothing to run", p.ID)
				}
				text = strings.Join(fillArgs(p.Add, name, binary), " ")
			case "config":
				if p.configPathFor("linux") == "" || len(p.JSONPath) == 0 {
					t.Fatalf("%s is a config provider with no path", p.ID)
				}
				raw, err := json.Marshal(fillValue(p.Value, name, binary))
				if err != nil {
					t.Fatal(err)
				}
				text = string(raw)
			case "toml":
				if len(p.Lines) == 0 || p.Section == "" {
					t.Fatalf("%s is a toml provider with nothing to write", p.ID)
				}
				text = strings.Join(fillArgs(p.Lines, name, binary), " ")
			default:
				t.Fatalf("%s has kind %q", p.ID, p.Kind)
			}

			switch {
			case !strings.Contains(text, binary):
				t.Errorf("does not name the binary: %s", text)
			// No subcommand, and no arguments at all. The command registered
			// is muster-mcp, which serves the protocol on its own; the
			// desktop binary cannot, because it links a browser engine and
			// will not start in the sandbox or container an agent runs in.
			// An `mcp` argument left in one entry breaks that one client in
			// exactly the place it is hardest to notice.
			case argumentsAfterTheCommand(p, name, binary) != "":
				t.Errorf("spawns the command with arguments (%s): %s",
					argumentsAfterTheCommand(p, name, binary), text)
			case strings.Contains(text, "http://"), strings.Contains(text, "{url}"):
				t.Errorf("still registers a URL: %s", text)
			case strings.Contains(text, "{command}"), strings.Contains(text, "{name}"):
				t.Errorf("a placeholder survived substitution: %s", text)
			}
		})
	}
}

// configPathFor is configPath for one operating system, so the test can check
// the entries for a platform it is not running on.
func (p Provider) configPathFor(goos string) string {
	if v, ok := p.Paths[goos]; ok {
		return v
	}
	return ""
}

// Nothing writes another project's name into somebody's files.
//
// This module came from refloft, and the first port left backups called
// "…refloft-backup-…" in other programs' directories. A file named after a
// program the person has never installed is a small thing that reads as a bug
// in whatever they were using.
func TestNothingIsNamedAfterWhereThisCameFrom(t *testing.T) {
	name, providers, err := Catalog()
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	if strings.Contains(strings.ToLower(name), "refloft") {
		t.Errorf("the server registers itself as %q", name)
	}
	for _, p := range providers {
		raw, err := json.Marshal(p)
		if err != nil {
			t.Fatal(err)
		}
		if strings.Contains(strings.ToLower(string(raw)), "refloft") {
			t.Errorf("%s still mentions where this came from: %s", p.ID, raw)
		}
	}
	if got := backupName("/tmp/mcp.json"); !strings.Contains(got, "muster-backup") {
		t.Errorf("a backup would be called %q", got)
	}
}

// argumentsAfterTheCommand reports anything a provider would pass to the
// server binary, which should be nothing.
func argumentsAfterTheCommand(p Provider, name, binary string) string {
	switch p.Kind {
	case "cli":
		args := fillArgs(p.Add, name, binary)
		for i, a := range args {
			if a == binary && i+1 < len(args) {
				return strings.Join(args[i+1:], " ")
			}
		}
	case "config":
		raw, err := json.Marshal(fillValue(p.Value, name, binary))
		if err != nil {
			return ""
		}
		// Both shapes the clients use: a separate args array, and a command
		// that is itself a list beginning with the binary.
		var shape struct {
			Args    []string        `json:"args"`
			Command json.RawMessage `json:"command"`
		}
		if err := json.Unmarshal(raw, &shape); err == nil {
			if len(shape.Args) > 0 {
				return strings.Join(shape.Args, " ")
			}
			var list []string
			if json.Unmarshal(shape.Command, &list) == nil && len(list) > 1 {
				return strings.Join(list[1:], " ")
			}
		}
	case "toml":
		for _, line := range fillArgs(p.Lines, name, binary) {
			if strings.HasPrefix(strings.TrimSpace(line), "args") &&
				!strings.Contains(line, "[]") {
				return line
			}
		}
	}
	return ""
}
