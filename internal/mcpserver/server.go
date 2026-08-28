// Package mcpserver exposes Muster's aggregate view to an agent.
//
// Backlog.md ships its own MCP server and it is per-project: an agent talks to
// one repository at a time. What Muster has that no single project does is the
// aggregate - what is in flight elsewhere, which milestone is active, where a
// dependency points - and that is the only reason this exists. It answers
// across every registered project, and nothing else.
package mcpserver

import (
	"context"
	"fmt"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/FMakareev/muster-backlog/internal/backlog"
	"github.com/FMakareev/muster-backlog/internal/backlogcli"
	"github.com/FMakareev/muster-backlog/internal/buildinfo"
	"github.com/FMakareev/muster-backlog/internal/registry"
	"github.com/FMakareev/muster-backlog/internal/store"
)

// Version is reported to the client during initialisation.
//
// From the build rather than written here. A constant meant a client was told
// 0.1.0 by a binary that called itself 0.0.0, which is two numbers for one
// thing and exactly what stamping the version was meant to end.
func Version() string { return buildinfo.Version() }

// instructions are sent at connection, before any tool is called.
//
// An agent working inside a project has two overlapping ways to do the same
// thing: this server, and that project's own backlog CLI. The boundary is not
// subtle and it is worth stating once, at the top, rather than hoping it can
// be inferred from ten tool descriptions.
//
// It is stated here rather than in the projects' own instruction files because
// this server will usually be running outside any project. Something that
// spans repositories cannot rely on any one of them to explain it.
const instructions = `Muster answers about every Backlog.md project a person ` +
	`has registered, at once. That is the only thing it does that anything ` +
	`else cannot.

Use it to see across projects: what is in flight elsewhere, which milestone a ` +
	`project is working towards, whether something is already tracked in ` +
	`another repository, or to search every project at once. Start with ` +
	`list_projects - ids, statuses, priorities and types are per-project, and ` +
	`nothing else here makes sense without knowing which projects exist.

Do not use it as a substitute for the backlog CLI when you are working inside ` +
	`a project. That CLI is the authority there: it has the whole command ` +
	`surface, it runs where you already are, and the project's own ` +
	`instructions are written around it. The writes offered here are a small ` +
	`subset, meant for changing something in a project you are not in.

Every write goes through that same CLI, so a change made here is a change ` +
	`made the ordinary way. Nothing reaches a folder the person has not ` +
	`registered.`

// Server answers MCP calls from the same data the interface reads.
type Server struct {
	registryPath string
	store        *store.Store
	cli          *backlogcli.Runner
	// cliErr is why writing is unavailable, when it is. Read tools work
	// regardless: an agent asking what is in flight should not be stopped by a
	// missing binary it has no use for.
	cliErr error
}

// New builds a server over a registry.
//
// A missing CLI is not fatal. It makes the write tools fail with the reason
// when they are called, which is better than refusing to start and leaving an
// agent with nothing at all.
func New(registryPath string) (*Server, error) {
	reg, err := registry.LoadFrom(registryPath)
	if err != nil && !strings.Contains(err.Error(), registry.ErrNoRegistry.Error()) {
		return nil, fmt.Errorf("reading the project registry: %w", err)
	}

	s := &Server{registryPath: registryPath, store: store.New()}
	s.store.Load(reg)
	s.cli, s.cliErr = backlogcli.New()
	return s, nil
}

// Reload re-reads the registry and every project.
//
// Agents are long-lived and the files are written by other things - the
// person's own editor, the CLI, another agent - so a tool that answered from a
// snapshot taken at startup would drift further from the truth the longer the
// session lasted. Every call reloads what it needs first.
func (s *Server) Reload() {
	reg, err := registry.LoadFrom(s.registryPath)
	if err != nil && !strings.Contains(err.Error(), registry.ErrNoRegistry.Error()) {
		return
	}
	s.store.Load(reg)
}

// MCP builds the server with every tool registered.
func (s *Server) MCP() *mcp.Server {
	server := mcp.NewServer(&mcp.Implementation{
		Name:    "muster",
		Title:   "Muster: every Backlog.md project at once",
		Version: Version(),
	}, &mcp.ServerOptions{Instructions: instructions})

	s.addReadTools(server)
	s.addWriteTools(server)
	return server
}

// Run serves over stdio until the client disconnects.
func (s *Server) Run(ctx context.Context) error {
	return s.MCP().Run(ctx, &mcp.StdioTransport{})
}

// project resolves a project the way every tool must: by path or by display
// name, and only among the registered ones.
//
// This is the whole of the access rule. An agent cannot reach a folder the
// person has not registered, whatever it passes, and asking for one says so
// rather than answering emptily.
func (s *Server) project(name string) (store.ProjectState, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return store.ProjectState{}, fmt.Errorf("no project given")
	}
	var known []string
	for _, p := range s.store.Projects() {
		known = append(known, p.Registry.DisplayName)
		if strings.EqualFold(p.Registry.Path, name) ||
			strings.EqualFold(p.Registry.DisplayName, name) {
			if !p.OK() {
				return p, fmt.Errorf("%s is registered but could not be read: %v",
					p.Registry.DisplayName, p.Err)
			}
			return p, nil
		}
	}
	return store.ProjectState{}, fmt.Errorf(
		"%q is not a registered project. Registered: %s",
		name, strings.Join(known, ", "))
}

// fail turns an error into the shape MCP wants for a tool that could not do
// what was asked, rather than a transport failure.
func fail[T any](err error) (*mcp.CallToolResult, T, error) {
	var zero T
	return &mcp.CallToolResult{
		IsError: true,
		Content: []mcp.Content{&mcp.TextContent{Text: err.Error()}},
	}, zero, nil
}

// ok returns a result whose structured content is the answer.
func ok[T any](out T) (*mcp.CallToolResult, T, error) {
	return nil, out, nil
}

// entityRef is how every tool names one thing, and it is always qualified by
// project: ids collide across projects freely.
type entityRef struct {
	Project string `json:"project" jsonschema:"the registered project's name or path"`
	ID      string `json:"id" jsonschema:"the entity id, unique only inside its project"`
}

// classOf keeps the archived and completed out of an answer unless asked.
func classOf(includeArchived bool) []backlog.Class {
	if includeArchived {
		return []backlog.Class{
			backlog.ClassActive, backlog.ClassCompleted, backlog.ClassArchived,
		}
	}
	return []backlog.Class{backlog.ClassActive}
}
