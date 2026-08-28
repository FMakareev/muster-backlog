package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// A lifecycle script names a program, and pnpm runs it with the workspace's
// own node_modules/.bin at the front of PATH. If nothing in the manifest
// installs that program, the name falls through to whatever the machine
// happens to have — which is not a dependency, it is a coincidence.
//
// The root prepare script ran `lefthook install` with lefthook declared
// nowhere. It worked for years on a machine with a system-wide copy, and
// `pnpm install` failed on CI and on every fresh clone with "lefthook: not
// found", at the install step, before a single check had a chance to run.
//
// This needs the workspace installed, which is what makes it meaningful:
// looking at the manifest alone cannot tell you whether the name resolves,
// because one package can provide a differently named binary — @commitlint/cli
// provides `commitlint`.
func TestEveryScriptRunsSomethingTheWorkspaceInstalls(t *testing.T) {
	for _, pkg := range []string{".", "frontend"} {
		bin := filepath.Join(pkg, "node_modules", ".bin")
		if _, err := os.Stat(bin); err != nil {
			t.Skipf("%s is not installed, so there is nothing to resolve against", pkg)
		}

		manifest, err := os.ReadFile(filepath.Join(pkg, "package.json"))
		if err != nil {
			t.Fatalf("read %s/package.json: %v", pkg, err)
		}
		var m struct {
			Scripts map[string]string `json:"scripts"`
		}
		if err := json.Unmarshal(manifest, &m); err != nil {
			t.Fatalf("%s/package.json is not valid JSON: %v", pkg, err)
		}

		for name, script := range m.Scripts {
			for _, program := range programsIn(script) {
				// node and pnpm come with the toolchain the README states,
				// not from the workspace.
				switch program {
				case "node", "pnpm", "npx", "sh":
					continue
				}
				if _, err := os.Stat(filepath.Join(bin, program)); err != nil {
					t.Errorf("%s/package.json script %q runs %q, which the workspace does not install: it will find whatever the machine happens to have, or nothing", pkg, name, program)
				}
			}
		}
	}
}

// programsIn returns the leading word of each command in a script, so that
// both halves of `eslint . && prettier --check .` are checked.
func programsIn(script string) []string {
	fields := strings.FieldsFunc(script, func(r rune) bool {
		return r == '&' || r == '|' || r == ';'
	})
	var programs []string
	for _, command := range fields {
		if word := strings.Fields(command); len(word) > 0 {
			programs = append(programs, word[0])
		}
	}
	return programs
}
