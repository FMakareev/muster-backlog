package backlog_test

import (
	"os"
	"strings"
	"testing"
	"time"

	"github.com/FMakareev/muster-backlog/internal/backlog"
	"github.com/FMakareev/muster-backlog/internal/project"
)

// realProjectsEnv names a colon-separated list of Backlog.md project roots to
// smoke-test the parser against.
//
// The committed fixtures cover every variation deliberately; this covers the
// long tail nobody thought to isolate. It is opt-in because those projects live
// on one machine and must never be a prerequisite for a green build elsewhere.
const realProjectsEnv = "MUSTER_SMOKE_PROJECTS"

func TestParseRealProjects(t *testing.T) {
	raw := os.Getenv(realProjectsEnv)
	if raw == "" {
		t.Skipf("set %s to a colon-separated list of project roots to run this",
			realProjectsEnv)
	}

	var totalFiles, totalDiagnostics int
	start := time.Now()

	for _, root := range strings.Split(raw, ":") {
		root = strings.TrimSpace(root)
		if root == "" {
			continue
		}
		t.Run(root, func(t *testing.T) {
			loc, err := project.Discover(root)
			if err != nil {
				t.Fatalf("Discover: %v", err)
			}
			p, err := backlog.Scan(loc)
			if err != nil {
				t.Fatalf("Scan: %v", err)
			}

			count := len(p.Tasks) + len(p.Drafts) + len(p.Milestones) +
				len(p.Documents) + len(p.Decisions)
			totalFiles += count
			totalDiagnostics += len(p.Diagnostics)

			if count == 0 {
				t.Fatalf("parsed nothing from %s", root)
			}
			if len(p.Config.Statuses) == 0 {
				t.Error("project declares no statuses")
			}

			// Every entity must have the two things the whole application
			// depends on: an identity and a title that came from frontmatter.
			for _, e := range allEntities(p) {
				if e.ID == "" {
					t.Errorf("%s has no id", e.Path)
				}
				if strings.TrimSpace(e.Title) == "" {
					t.Errorf("%s has no title", e.Path)
				}
				if strings.Contains(e.Title, "\n") {
					t.Errorf("%s has a multi-line title: %q", e.Path, e.Title)
				}
			}
			// Task statuses must be valid against the project's own list.
			valid := map[string]bool{}
			for _, s := range p.Config.Statuses {
				valid[strings.ToLower(s)] = true
			}
			for _, task := range p.Tasks {
				if task.Status == "" || task.Class != backlog.ClassActive {
					continue
				}
				if !valid[strings.ToLower(task.Status)] {
					t.Errorf("%s has status %q, which is not in %v",
						task.Path, task.Status, p.Config.Statuses)
				}
			}

			for _, d := range p.Diagnostics {
				t.Logf("diagnostic: %s: %s", d.Path, d.Reason)
			}
		})
	}

	t.Logf("parsed %d entities with %d diagnostics in %s",
		totalFiles, totalDiagnostics, time.Since(start))
}

func allEntities(p *backlog.Project) []backlog.Entity {
	out := make([]backlog.Entity, 0,
		len(p.Tasks)+len(p.Drafts)+len(p.Milestones)+len(p.Documents)+len(p.Decisions))
	out = append(out, p.Tasks...)
	out = append(out, p.Drafts...)
	out = append(out, p.Milestones...)
	out = append(out, p.Documents...)
	out = append(out, p.Decisions...)
	return out
}
