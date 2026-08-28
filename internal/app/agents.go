package app

import (
	"os"
	"path/filepath"
	"runtime"

	"github.com/FMakareev/muster-backlog/internal/agents"
	"github.com/FMakareev/muster-backlog/internal/whichbin"
)

// Connecting an agent client is a button rather than a paragraph.
//
// Everything here passes through to internal/agents, which shows what it will
// do before doing it, backs up every file it writes, and disconnects by the
// same route it connected. A tool that edits other programs' configuration has
// to be boring and visible.

// mcpServerName is the binary that serves the protocol.
//
// A separate program from the desktop one, and it has to be: this binary
// imports the Wails application package, so the dynamic loader resolves
// libwebkit2gtk before main is entered. `muster mcp` therefore cannot start
// inside a Flatpak sandbox, a container, or on a machine with no desktop -
// which is where agents run. muster-mcp links none of that.
const mcpServerName = "muster-mcp"

// serverFileName is the server binary's name on this platform.
func serverFileName() string {
	if runtime.GOOS == "windows" {
		return mcpServerName + ".exe"
	}
	return mcpServerName
}

// besideExecutable is the server sitting next to this binary, or empty.
//
// Where every way of installing them puts the two together: the package, the
// AppImage, and a build tree's bin directory.
func besideExecutable() string {
	binary, err := os.Executable()
	if err != nil {
		return ""
	}
	if resolved, err := filepath.EvalSymlinks(binary); err == nil {
		binary = resolved
	}
	beside := filepath.Join(filepath.Dir(binary), serverFileName())
	if info, err := os.Stat(beside); err == nil && !info.IsDir() {
		return beside
	}
	return ""
}

// locateServer decides what a client should be told to spawn.
//
// An absolute path, not the word "muster-mcp": a client started from a
// launcher has no more of a shell's PATH than this application did when it
// could not find the Backlog.md CLI. Registering a bare name would reproduce
// that bug inside somebody else's program, where it is harder to diagnose.
//
// In a container the obvious answer is the wrong one, and confidently so:
// see container.go. Everywhere else it is the server beside this binary, then
// wherever a person could run it from.
func locateServer() serverLocation {
	if inContainer() {
		return containerLocation()
	}
	if beside := besideExecutable(); beside != "" {
		return serverLocation{Command: beside}
	}
	if path, ok := whichbin.Find(mcpServerName); ok {
		return serverLocation{Command: path}
	}
	return serverLocation{Problem: missingServer}
}

// AgentClients lists the clients Muster knows how to connect, and whether each
// is installed here. It starts nobody else's program, so it answers at once.
func (s *BoardService) AgentClients() ([]agents.Status, error) {
	return agents.List(serverCommand().Command)
}

// AgentConnections is the slow half of that answer: for each installed CLI
// client, whether Muster is already in its configuration. Asking means running
// the client, so the interface shows the list first and fills these in when
// they arrive.
func (s *BoardService) AgentConnections() ([]agents.Status, error) {
	return agents.Connections(serverCommand().Command)
}

// missingServer is what to say when the thing a client would be told to spawn
// is not on this machine.
//
// Refusing is the point. Writing a command that does not exist into another
// program's configuration is precisely the failure this exists to prevent: it
// looks like it worked, and fails later inside a program that cannot explain
// why.
const missingServer = mcpServerName + " is not next to this application, and " +
	"that is the program a client has to spawn — the desktop binary links a " +
	"browser engine and cannot start in a sandbox or a container. It ships " +
	"with the package and the AppImage; from a build tree, `wails3 task build` " +
	"makes it."

// serverCommand is locateServer, behind a name a test can replace.
//
// The refusal below is the behaviour worth testing and it cannot be reached
// on a machine that has the server installed - which every machine that has
// run the package does. A seam here beats a test that skips exactly where it
// matters.
var serverCommand = locateServer

// AgentPlan is exactly what connecting - or disconnecting - would do.
func (s *BoardService) AgentPlan(id string, disconnect bool) agents.Plan {
	server := serverCommand()
	// Disconnecting only needs the name, so it is still offered: somebody
	// whose server binary has gone is exactly who needs to remove the entry
	// pointing at it.
	if server.Command == "" && !disconnect {
		return agents.Plan{ID: id, Error: server.Problem}
	}
	plan := agents.PlanFor(id, server.Command, disconnect)
	// Said before anything is written, because which environment the command
	// resolves in is the thing that was invisible when this went wrong.
	if server.Where != "" && !disconnect && plan.Error == "" {
		if plan.Note == "" {
			plan.Note = server.Where
		} else {
			plan.Note = server.Where + "\n\n" + plan.Note
		}
	}
	return plan
}

// AgentApply carries out that plan. Nothing happens without this call.
func (s *BoardService) AgentApply(id string, disconnect bool) agents.Result {
	server := serverCommand()
	if server.Command == "" && !disconnect {
		return agents.Result{Error: server.Problem}
	}
	return agents.Apply(id, server.Command, disconnect)
}
