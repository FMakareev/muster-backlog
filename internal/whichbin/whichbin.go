// Package whichbin finds a command where a person would say it is installed,
// rather than only where this process's PATH happens to point.
//
// The distinction matters because a desktop app almost never inherits the PATH
// its owner sees in a terminal. A GUI is started by a launcher — a .desktop
// entry, a dock icon, a distrobox export — and none of those read ~/.bashrc or
// ~/.zshrc, which is where every version manager and every "curl | sh" installer
// puts its line. The app ends up with the system default, /usr/bin and friends,
// and reports that a tool sitting in ~/.local/bin is "not installed".
//
// That was not a hypothetical: a packaged build showed Claude Code as missing on
// a machine where `claude` was one directory away, in ~/.local/bin, because the
// launcher handed the process a PATH of six system directories. Everything the
// app found instead — Claude Desktop, Cursor, Zed — it found by reading a file
// under $HOME, which is why the failure looked so arbitrary from outside.
//
// So: PATH first, because a PATH entry is the owner saying where they want this
// name to come from. Then the short list of places installers actually use.
package whichbin

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// ErrNotFound is what a name nobody can find comes back as.
var ErrNotFound = errors.New("not on PATH or in the usual install folders")

// Look finds an executable by name.
//
// The second return says whether it came from PATH. Callers that show a command
// to a person need it: a name found on PATH is the name to print, because that
// is what will run and what they could type themselves, while a name found
// somewhere else has to be printed in full — the bare word would not work in
// their terminal either, and printing one command while running another is the
// one thing a tool that edits other programs' configuration must never do.
func Look(name string) (string, bool, error) {
	if p, err := exec.LookPath(name); err == nil {
		return p, true, nil
	}
	// A name with a separator in it was never a PATH lookup to begin with.
	if strings.ContainsAny(name, `/\`) {
		return "", false, fmt.Errorf("%s: %w", name, ErrNotFound)
	}
	for _, dir := range dirs() {
		if p, ok := executableIn(dir, name); ok {
			return p, false, nil
		}
	}
	return "", false, fmt.Errorf("%s: %w", name, ErrNotFound)
}

// Find is Look for callers that only want the path, or nothing.
func Find(name string) (string, bool) {
	p, _, err := Look(name)
	return p, err == nil
}

// dirs is where to look after PATH: one entry per installer that people
// actually run, in the order a person would guess. Patterns may contain a *,
// because every version manager keeps its binaries one directory below a
// version number.
//
// This list is deliberately finite and boring. Walking $HOME looking for
// anything executable would find a name, eventually, and running whatever it
// found would be a much worse bug than the one this fixes.
func dirs() []string {
	var pats []string
	add := func(p ...string) {
		for _, s := range p {
			if s != "" {
				pats = append(pats, s)
			}
		}
	}
	// An installer that exports its own variable has said where it put things,
	// and that beats guessing at the default.
	env := func(key string, rest ...string) string {
		v := os.Getenv(key)
		if v == "" {
			return ""
		}
		return filepath.Join(append([]string{v}, rest...)...)
	}
	add(
		env("PNPM_HOME"),
		env("BUN_INSTALL", "bin"),
		env("VOLTA_HOME", "bin"),
		env("CARGO_HOME", "bin"),
		env("GOBIN"),
		env("NPM_CONFIG_PREFIX", "bin"),
	)
	if gp := os.Getenv("GOPATH"); gp != "" {
		for _, p := range filepath.SplitList(gp) {
			add(filepath.Join(p, "bin"))
		}
	}

	home, err := os.UserHomeDir()
	if err == nil && home != "" {
		in := func(rest ...string) string {
			return filepath.Join(append([]string{home}, rest...)...)
		}
		add(
			in(".local", "bin"), // the freedesktop answer, and where most installers land
			in("bin"),
			in(".claude", "local"), // Claude Code's own local install
			in(".local", "share", "pnpm"),
			in(".bun", "bin"),
			in(".deno", "bin"),
			in(".volta", "bin"),
			in(".cargo", "bin"),
			in("go", "bin"),
			in(".npm-global", "bin"),
			in(".yarn", "bin"),
			in(".nvm", "versions", "node", "*", "bin"),
			in(".local", "share", "fnm", "node-versions", "*", "installation", "bin"),
			in(".local", "share", "mise", "shims"),
			in(".asdf", "shims"),
		)
		switch runtime.GOOS {
		case "windows":
			add(
				filepath.Join(os.Getenv("APPDATA"), "npm"),
				filepath.Join(os.Getenv("LOCALAPPDATA"), "Microsoft", "WindowsApps"),
				filepath.Join(os.Getenv("LOCALAPPDATA"), "Programs", "*", "bin"),
				filepath.Join(os.Getenv("LOCALAPPDATA"), "Programs", "*"),
			)
		default:
			add(in(".local", "share", "flatpak", "exports", "bin"))
		}
	}

	switch runtime.GOOS {
	case "darwin":
		add("/opt/homebrew/bin", "/usr/local/bin", "/opt/local/bin")
	case "linux":
		add("/usr/local/bin", "/snap/bin", "/var/lib/flatpak/exports/bin")
	}

	out := make([]string, 0, len(pats))
	seen := map[string]bool{}
	for _, p := range pats {
		if !strings.Contains(p, "*") {
			if !seen[p] {
				seen[p] = true
				out = append(out, p)
			}
			continue
		}
		matches, _ := filepath.Glob(p)
		// Newest last for version managers, so the newest version wins: the
		// names sort lexically, which is wrong for 9 vs 10 and right often
		// enough that reversing beats not trying.
		for i := len(matches) - 1; i >= 0; i-- {
			if !seen[matches[i]] {
				seen[matches[i]] = true
				out = append(out, matches[i])
			}
		}
	}
	return out
}

// executableIn reports the full path of name inside dir, if something runnable
// is there. On Windows the name carries an extension, and PATHEXT says which.
func executableIn(dir, name string) (string, bool) {
	if runtime.GOOS != "windows" {
		p := filepath.Join(dir, name)
		if runnable(p) {
			return p, true
		}
		return "", false
	}
	exts := strings.Split(os.Getenv("PATHEXT"), ";")
	if len(exts) == 0 || os.Getenv("PATHEXT") == "" {
		exts = []string{".COM", ".EXE", ".BAT", ".CMD"}
	}
	if filepath.Ext(name) != "" {
		exts = append([]string{""}, exts...)
	}
	for _, ext := range exts {
		p := filepath.Join(dir, name+strings.ToLower(strings.TrimSpace(ext)))
		if runnable(p) {
			return p, true
		}
	}
	return "", false
}

// runnable follows symlinks, because ~/.local/bin is full of them, and asks for
// a regular file with an execute bit. On Windows the bit means nothing and
// having the right extension is the whole test.
func runnable(p string) bool {
	st, err := os.Stat(p)
	if err != nil || st.IsDir() {
		return false
	}
	if runtime.GOOS == "windows" {
		return true
	}
	return st.Mode().Perm()&0o111 != 0
}
