package app

import (
	"os"
	"path/filepath"
	"strings"
)

// Whether Muster is running inside a container, and what that means for the
// command it registers with an agent client.
//
// The failure this exists to prevent: installed inside a distrobox container
// and exported to the host, os.Executable is /usr/local/bin/muster and
// /usr/local/bin/muster-mcp sits beside it, so os.Stat confirms a path that is
// perfectly real — inside the container. But a distrobox container shares
// $HOME with the host, so the file the connector edits is the host's
// ~/.claude.json, while the client reading it runs on the host, where that
// path holds nothing. ENOENT, from a program that cannot say where the path
// came from.
//
// os.Stat answers for the filesystem this process is looking at, not the one
// the client will spawn from, and in a container those differ while sharing
// the configuration file. So the container has to be detected before anything
// is written, and it can be: every runtime leaves a mark.

// containerMarks are the files a container runtime leaves behind.
//
// /run/.containerenv is podman's, and distrobox and toolbox are podman
// containers. /.dockerenv is docker's. /run/host is what distrobox and
// toolbox mount to reach the host, and its presence is the strongest signal
// that the host is a different filesystem from this one.
var containerMarks = []string{
	"/run/.containerenv",
	"/.dockerenv",
	"/run/host",
}

// containerVars are set inside the container rather than on the host.
// CONTAINER_ID is distrobox's and toolbox's; container= is set by podman and
// by systemd-nspawn.
var containerVars = []string{"CONTAINER_ID", "container"}

// inContainer reports whether this process is inside one. A variable so a
// test can be either without being run inside one.
var inContainer = func() bool {
	for _, mark := range containerMarks {
		if _, err := os.Stat(mark); err == nil {
			return true
		}
	}
	for _, name := range containerVars {
		if strings.TrimSpace(os.Getenv(name)) != "" {
			return true
		}
	}
	return false
}

// sharedBinDir is where a container's exported wrappers land.
//
// ~/.local/bin, because $HOME is what a distrobox or toolbox container shares
// with the host — the same reason the client's configuration file is reachable
// from in here at all. A wrapper written there is visible from both sides,
// which is exactly what the registered command has to be.
var sharedBinDir = func() string {
	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		return ""
	}
	return filepath.Join(home, ".local", "bin")
}

// serverLocation is what to register, and what to say about where it runs.
type serverLocation struct {
	// Command is the absolute path to spawn, empty when none would work.
	Command string
	// Where is one line for the plan, so the environment the command runs in
	// is visible before anything is written rather than discovered afterwards.
	Where string
	// Problem is why there is nothing to register.
	Problem string
}

// exportedServer is the wrapper a container's own tooling makes, if there is
// one.
//
// distrobox-export writes a small script into ~/.local/bin that runs the real
// binary through distrobox-enter when called from the host, hops back out
// through distrobox-host-exec when called from a different container, and
// execs it directly when called from its own. That is a command that resolves
// from either side, which is the whole requirement, and it is the container's
// own mechanism rather than something invented here.
//
// A plain copy of the server in the same directory works too — it is a static
// binary — and is accepted for the same reason: what matters is that the path
// is on the shared home, not what put it there.
func exportedServer() string {
	dir := sharedBinDir()
	if dir == "" {
		return ""
	}
	path := filepath.Join(dir, serverFileName())
	info, err := os.Stat(path)
	if err != nil || info.IsDir() || info.Mode().Perm()&0o111 == 0 {
		return ""
	}
	return path
}

// containerLocation decides what to register from inside a container.
func containerLocation() serverLocation {
	if exported := exportedServer(); exported != "" {
		return serverLocation{
			Command: exported,
			Where: "Muster is running in a container. This registers the " +
				"copy in your home directory, which the container shares " +
				"with the host — so it resolves whether the client runs " +
				"inside the container or outside it.",
		}
	}

	inside := besideExecutable()
	if inside == "" {
		inside = "/usr/local/bin/" + serverFileName()
	}
	return serverLocation{
		Problem: "Muster is running in a container, and " + mcpServerName +
			" exists only inside it. Registering that path would write a " +
			"command the client cannot spawn: your home directory is shared " +
			"with the host, so the file being edited is the host's, and " +
			inside + " means nothing there. Export it first, which puts a " +
			"wrapper in ~/.local/bin that works from either side:\n\n" +
			"    distrobox-export --bin " + inside,
	}
}
