package app_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/FMakareev/muster-backlog/internal/app"
)

// initialised builds a project through the CLI itself, so it has the whole
// eleven-directory skeleton. The hand-made fixtures elsewhere carry only
// tasks/, and the CLI writes a decision into a decisions/ directory it expects
// to find rather than creating one.
func initialised(t *testing.T, name string) string {
	t.Helper()
	root := filepath.Join(t.TempDir(), name)
	if err := os.MkdirAll(root, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	cmd := exec.Command("backlog", "init", name, "--defaults", "--no-git",
		"--integration-mode", "none")
	cmd.Dir = root
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("backlog init: %v\n%s", err, out)
	}
	return root
}

func read(t *testing.T, dir, match string) string {
	t.Helper()
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("reading %s: %v", dir, err)
	}
	for _, entry := range entries {
		if !strings.Contains(entry.Name(), match) {
			continue
		}
		raw, err := os.ReadFile(filepath.Join(dir, entry.Name()))
		if err != nil {
			t.Fatalf("reading %s: %v", entry.Name(), err)
		}
		return string(raw)
	}
	var names []string
	for _, entry := range entries {
		names = append(names, entry.Name())
	}
	t.Fatalf("no file matching %q in %s: %v", match, dir, names)
	return ""
}

// Creating a document with a body is two CLI calls - the create command takes
// no content - and one act to the person doing it.
func TestCreatingADocumentWithItsBody(t *testing.T) {
	if _, err := exec.LookPath("backlog"); err != nil {
		t.Skip("the backlog CLI is not installed")
	}
	one := initialised(t, "one")
	s := startService(t, withRegistry(t, one))

	result := s.CreateDocument(app.NewDocumentInput{
		Project: one,
		Title:   "How this works",
		Type:    "guide",
		Content: "# How this works\n\nIn **two** calls, as it happens.\n",
	})
	if !result.OK {
		t.Fatalf("CreateDocument: %+v", result.Problem)
	}
	if result.TaskID == "" {
		t.Error("the new document's id was not reported")
	}

	body := read(t, filepath.Join(one, "backlog", "docs"), "How-this-works")
	switch {
	case !strings.Contains(body, "type: guide"):
		t.Errorf("the type did not reach the CLI:\n%s", body)
	case !strings.Contains(body, "In **two** calls"):
		t.Errorf("the content was not written:\n%s", body)
	}

	// And it is readable through the viewer's own path immediately.
	found := false
	for _, item := range s.Entities("document") {
		if item.Entity.Title == "How this works" {
			found = true
			if !strings.Contains(item.Entity.Body, "two") {
				t.Errorf("the body did not load: %q", item.Entity.Body)
			}
		}
	}
	if !found {
		t.Error("the new document is not in the viewer")
	}

	if result := s.CreateDocument(app.NewDocumentInput{Project: one, Title: "  "}); result.OK {
		t.Error("a document with no title was accepted")
	}
}

// doc update --content takes nothing smaller than the whole body, so editing
// sends the entire document back.
func TestEditingADocumentReplacesTheWholeBody(t *testing.T) {
	if _, err := exec.LookPath("backlog"); err != nil {
		t.Skip("the backlog CLI is not installed")
	}
	one := initialised(t, "one")
	s := startService(t, withRegistry(t, one))

	created := s.CreateDocument(app.NewDocumentInput{
		Project: one, Title: "Draft notes", Content: "First version.\n",
	})
	if !created.OK {
		t.Fatalf("CreateDocument: %+v", created.Problem)
	}

	result := s.UpdateDocument(one, created.TaskID, app.DocumentUpdate{
		Title:   "Settled notes",
		Type:    "specification",
		Content: "Second version, entirely.\n",
	})
	if !result.OK {
		t.Fatalf("UpdateDocument: %+v", result.Problem)
	}

	body := read(t, filepath.Join(one, "backlog", "docs"), "doc-1")
	switch {
	case strings.Contains(body, "First version"):
		t.Errorf("the old body survived:\n%s", body)
	case !strings.Contains(body, "Second version, entirely"):
		t.Errorf("the new body is missing:\n%s", body)
	case !strings.Contains(body, "title: Settled notes"):
		t.Errorf("the title did not change:\n%s", body)
	case !strings.Contains(body, "type: specification"):
		t.Errorf("the type did not change:\n%s", body)
	}
}

// A decision can be created but not written: the CLI has no decision update,
// so what arrives is a skeleton with its headings.
func TestCreatingADecisionWritesTheSkeleton(t *testing.T) {
	if _, err := exec.LookPath("backlog"); err != nil {
		t.Skip("the backlog CLI is not installed")
	}
	one := initialised(t, "one")
	s := startService(t, withRegistry(t, one))

	result := s.CreateDecision(one, "Write the registry through the syntax tree", "accepted")
	if !result.OK {
		t.Fatalf("CreateDecision: %+v", result.Problem)
	}
	if result.TaskID == "" {
		t.Error("the new decision's id was not reported")
	}

	body := read(t, filepath.Join(one, "backlog", "decisions"), "decision-1")
	for _, want := range []string{"status: accepted", "## Context", "## Decision", "## Consequences"} {
		if !strings.Contains(body, want) {
			t.Errorf("missing %q from the decision:\n%s", want, body)
		}
	}

	if result := s.CreateDecision(one, "   ", "accepted"); result.OK {
		t.Error("a decision with no title was accepted")
	}
}
