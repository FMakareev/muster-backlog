// Command muster-mcp serves Muster's Model Context Protocol over stdio.
//
// It exists because the desktop binary cannot run where an agent runs.
// `muster mcp` builds no window, but main imports the Wails application
// package, so the dynamic loader resolves libwebkit2gtk before main is
// entered: inside a Flatpak sandbox, a container, or on a headless machine,
// the process dies with "error while loading shared libraries" and never
// reaches the subcommand. An MCP server that needs a browser engine installed
// is the wrong shape.
//
// This links none of that. Built with CGO off it is a static binary with no
// dynamic dependencies at all, which is what makes it usable from a sandboxed
// client, a container, or a server with no desktop on it.
//
// It reads the same registry the application reads, so an agent sees exactly
// the projects the person registered and no others, and it writes the same
// way everything else does — through the backlog CLI.
package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/FMakareev/muster-backlog/internal/buildinfo"
	"github.com/FMakareev/muster-backlog/internal/mcpserver"
	"github.com/FMakareev/muster-backlog/internal/registry"
)

func main() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "--version", "-version", "-v", "version":
			fmt.Println(buildinfo.LineFor("muster-mcp"))
			return
		}
	}

	if err := serve(); err != nil && !errors.Is(err, context.Canceled) {
		// Stderr, never stdout: stdout is the protocol.
		log.Fatalf("muster-mcp: %v", err)
	}
}

func serve() error {
	server, err := mcpserver.New(registry.DefaultPath())
	if err != nil {
		return err
	}

	ctx, stop := signal.NotifyContext(context.Background(),
		syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	return server.Run(ctx)
}
