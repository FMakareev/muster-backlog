package backlogcli

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/FMakareev/muster-backlog/internal/whichbin"
)

// Find locates the Backlog.md CLI and reports everywhere it looked.
//
// The looking itself is whichbin's: PATH first, because a PATH entry is the
// owner saying where a name should come from, then the short list of places
// installers actually use. That matters here because an application started
// from a desktop launcher inherits the session environment rather than a
// shell's, so a CLI installed by pnpm, npm or bun into a directory only a shell
// rc file adds to PATH is invisible to it - which is exactly where Backlog.md
// installs, and why a packaged build reported it missing while the same binary
// ran in a terminal on the same machine.
//
// More than one candidate comes back because a file can be there and still not
// run, and the caller has to be able to try the next one.
func Find() ([]string, []string) {
	searched := []string{"PATH"}

	var found []string
	if binary, _ := whichbin.Find(BinaryName); binary != "" {
		found = append(found, binary)
	}

	// The explicit list is kept for two reasons whichbin cannot serve: telling
	// someone where the search went when it failed, and offering a second
	// candidate when the first is a wrapper that cannot run.
	for _, dir := range candidateDirs() {
		candidate := filepath.Join(dir, BinaryName)
		searched = append(searched, candidate)
		if executable(candidate) && !contains(found, candidate) {
			found = append(found, candidate)
		}
	}

	// Said last because it is done last, and said at all because a message
	// listing only directories reads as though a fixed list was the whole
	// search — which is the complaint it earned when the list turned out to be
	// missing somebody's directory. Claimed only when there is a shell to ask.
	if shell := strings.TrimSpace(os.Getenv("SHELL")); shell != "" {
		searched = append(searched, "your login shell ("+shell+")")
	}
	return found, searched
}

func contains(list []string, want string) bool {
	for _, item := range list {
		if item == want {
			return true
		}
	}
	return false
}

// Environment is what the CLI is run with: our own, plus every directory a
// package manager installs binaries into.
//
// This is not belt and braces. The binary pnpm installs is a shell script
// ending in `exec node …`, and node is itself usually installed by a version
// manager into a directory only a shell rc file adds to PATH. So an
// application started from a desktop launcher can find the CLI, run it, and
// watch it fail with "exec: node: not found" - which is what happens on the
// author's machine, and why finding the binary was only half the problem.
func Environment() []string {
	env := os.Environ()

	var dirs []string
	seen := map[string]bool{}
	for _, dir := range append(candidateDirs(), nodeBinDirs()...) {
		if dir != "" && !seen[dir] {
			seen[dir] = true
			dirs = append(dirs, dir)
		}
	}
	if len(dirs) == 0 {
		return env
	}

	extra := strings.Join(dirs, string(os.PathListSeparator))
	for i, entry := range env {
		if after, ok := strings.CutPrefix(entry, "PATH="); ok {
			env[i] = "PATH=" + after + string(os.PathListSeparator) + extra
			return env
		}
	}
	return append(env, "PATH="+extra)
}

// nodeBinDirs are the directories a Node version manager keeps interpreters
// in, which is where `exec node` has to find one.
func nodeBinDirs() []string {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil
	}
	var dirs []string
	dirs = append(dirs, versionedNodeBins(filepath.Join(home, ".nvm", "versions", "node"))...)
	dirs = append(dirs, versionedNodeBins(filepath.Join(home, ".local", "share", "fnm", "node-versions"))...)
	dirs = append(dirs, filepath.Join(home, ".volta", "bin"))
	return dirs
}

// candidateDirs are the directories package managers install global binaries
// into, in the order they are worth trying.
//
// Environment variables first, since a person who set one meant it, then the
// defaults those tools use, then the ordinary system locations.
func candidateDirs() []string {
	var dirs []string
	add := func(paths ...string) {
		for _, path := range paths {
			if strings.TrimSpace(path) != "" {
				dirs = append(dirs, path)
			}
		}
	}

	// Both, because pnpm keeps its own binaries directly in PNPM_HOME and the
	// ones it installs globally in bin underneath it. Only the first was
	// listed, so a CLI installed with `pnpm add -g` was reported missing and
	// the message then printed every directory except the one it was in.
	if pnpm := os.Getenv("PNPM_HOME"); pnpm != "" {
		add(pnpm, filepath.Join(pnpm, "bin"))
	}
	if bun := os.Getenv("BUN_INSTALL"); bun != "" {
		add(filepath.Join(bun, "bin"))
	}
	if npm := os.Getenv("NPM_CONFIG_PREFIX"); npm != "" {
		add(filepath.Join(npm, "bin"))
	}

	home, err := os.UserHomeDir()
	if err == nil {
		data := os.Getenv("XDG_DATA_HOME")
		if data == "" {
			data = filepath.Join(home, ".local", "share")
		}
		add(
			filepath.Join(data, "pnpm"),
			filepath.Join(data, "pnpm", "bin"),
			filepath.Join(home, ".local", "share", "pnpm"),
			filepath.Join(home, ".local", "share", "pnpm", "bin"),
			filepath.Join(home, ".bun", "bin"),
			filepath.Join(home, ".local", "bin"),
			filepath.Join(home, ".npm-global", "bin"),
			filepath.Join(home, ".cargo", "bin"),
			filepath.Join(home, "bin"),
			// Node installed through nvm or fnm puts global bins beside the
			// version it belongs to, which no session environment knows about.
			filepath.Join(home, ".volta", "bin"),
		)
		add(versionedNodeBins(filepath.Join(home, ".nvm", "versions", "node"))...)
		add(versionedNodeBins(filepath.Join(home, ".local", "share", "fnm", "node-versions"))...)
	}

	if runtime.GOOS == "darwin" {
		add("/opt/homebrew/bin", "/usr/local/bin")
	}
	add("/usr/local/bin", "/usr/bin")
	return dirs
}

// versionedNodeBins lists the bin directories under a Node version manager's
// install root, newest name last, since that is the order the filesystem gives
// and the newest is the likeliest.
func versionedNodeBins(root string) []string {
	entries, err := os.ReadDir(root)
	if err != nil {
		return nil
	}
	var dirs []string
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		dirs = append(dirs,
			filepath.Join(root, entry.Name(), "bin"),
			filepath.Join(root, entry.Name(), "installation", "bin"))
	}
	return dirs
}

// executable reports whether a path is a file that can be run.
func executable(path string) bool {
	info, err := os.Stat(path)
	if err != nil || info.IsDir() {
		return false
	}
	if runtime.GOOS == "windows" {
		return true
	}
	return info.Mode()&0o111 != 0
}
