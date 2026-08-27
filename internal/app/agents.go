package app

import (
	"os"
	"path/filepath"

	"github.com/FMakareev/muster-backlog/internal/agents"
)

// Connecting an agent client is a button rather than a paragraph.
//
// Everything here passes through to internal/agents, which shows what it will
// do before doing it, backs up every file it writes, and disconnects by the
// same route it connected. A tool that edits other programs' configuration has
// to be boring and visible.

// mcpCommand is what a client is told to spawn.
//
// The absolute path to this binary, not the word "muster": a client started
// from a launcher has no more of a shell's PATH than this application did when
// it could not find the Backlog.md CLI. Registering a bare name would
// reproduce that bug in somebody else's program, where it would be harder to
// diagnose.
func mcpCommand() string {
	binary, err := os.Executable()
	if err != nil {
		return "muster"
	}
	if resolved, err := filepath.EvalSymlinks(binary); err == nil {
		binary = resolved
	}
	return binary
}

// AgentClients lists the clients Muster knows how to connect, and whether each
// is installed here. It starts nobody else's program, so it answers at once.
func (s *BoardService) AgentClients() ([]agents.Status, error) {
	return agents.List(mcpCommand())
}

// AgentConnections is the slow half of that answer: for each installed CLI
// client, whether Muster is already in its configuration. Asking means running
// the client, so the interface shows the list first and fills these in when
// they arrive.
func (s *BoardService) AgentConnections() ([]agents.Status, error) {
	return agents.Connections(mcpCommand())
}

// AgentPlan is exactly what connecting - or disconnecting - would do.
func (s *BoardService) AgentPlan(id string, disconnect bool) agents.Plan {
	return agents.PlanFor(id, mcpCommand(), disconnect)
}

// AgentApply carries out that plan. Nothing happens without this call.
func (s *BoardService) AgentApply(id string, disconnect bool) agents.Result {
	return agents.Apply(id, mcpCommand(), disconnect)
}
