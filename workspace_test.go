package main

import (
	"encoding/json"
	"os"
	"os/exec"
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

// An embed pattern is resolved by the compiler, so one that matches nothing is
// not a missing asset at runtime — it is a package that does not build.
// main.go embeds all:frontend/dist, which the frontend build generates and
// .gitignore ignores, so on a clone with no build history every Go command
// that touches package main failed outright: "pattern all:frontend/dist: no matching files
// found". CI hit it in the lint step, and the test command CONTRIBUTING tells
// a contributor to run before pushing failed the same way.
//
// The question this asks is the one that matters: not whether the pattern
// resolves on this machine, where a previous build left the directory behind,
// but whether it resolves for somebody who has just cloned. So it asks git
// what is in the repository rather than asking the filesystem what is here.
func TestEveryEmbeddedPathIsInTheRepository(t *testing.T) {
	for _, file := range goFiles(t) {
		content, err := os.ReadFile(file)
		if err != nil {
			t.Fatalf("read %s: %v", file, err)
		}
		for _, line := range strings.Split(string(content), "\n") {
			directive := strings.TrimSpace(line)
			if !strings.HasPrefix(directive, "//go:embed ") {
				continue
			}
			for _, pattern := range strings.Fields(strings.TrimPrefix(directive, "//go:embed ")) {
				// all: and $ only change which files inside a match are taken.
				pattern = strings.TrimPrefix(strings.TrimPrefix(pattern, "all:"), "$")
				// Patterns are relative to the file that carries them.
				path := filepath.Join(filepath.Dir(file), pattern)

				tracked, err := exec.Command("git", "ls-files", "--", path).Output()
				if err != nil {
					t.Skipf("git is not answering here: %v", err)
				}
				if len(strings.TrimSpace(string(tracked))) == 0 {
					t.Errorf("%s embeds %q, and the repository contains nothing that matches it: a fresh clone cannot compile this package until something has been built", file, pattern)
				}
			}
		}
	}
}

// goFiles lists the module's own Go files, skipping the directories that hold
// other people's.
func goFiles(t *testing.T) []string {
	t.Helper()
	var files []string
	err := filepath.WalkDir(".", func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			switch entry.Name() {
			case "node_modules", ".git", "frontend", "bin":
				return filepath.SkipDir
			}
			return nil
		}
		if strings.HasSuffix(path, ".go") {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk the repository: %v", err)
	}
	return files
}
