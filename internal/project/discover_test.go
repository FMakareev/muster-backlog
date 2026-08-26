package project_test

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/FMakareev/muster-backlog/internal/project"
)

// corpus is the pinned reference corpus committed by the format-contract spike.
const corpus = "../../testdata/backlog-format"

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

func TestDiscoverStandardLayout(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "backlog", "config.yml"), "project_name: Sample\n")

	got, err := project.Discover(root)
	if err != nil {
		t.Fatalf("Discover: %v", err)
	}
	if got.Layout != project.LayoutStandard {
		t.Errorf("Layout = %q, want %q", got.Layout, project.LayoutStandard)
	}
	if want := filepath.Join(root, "backlog"); got.DataDir != want {
		t.Errorf("DataDir = %q, want %q", got.DataDir, want)
	}
	if want := filepath.Join(root, "backlog", "tasks"); got.Sub("tasks") != want {
		t.Errorf("Sub(tasks) = %q, want %q", got.Sub("tasks"), want)
	}
}

// The hidden layout is not hypothetical: one of the nine projects in the
// corpus uses it, and probing only "backlog" would make it invisible.
func TestDiscoverHiddenLayoutFromCorpus(t *testing.T) {
	root := filepath.Join(corpus, "04-layout-dot-backlog")

	got, err := project.Discover(root)
	if err != nil {
		t.Fatalf("Discover: %v", err)
	}
	if got.Layout != project.LayoutHidden {
		t.Errorf("Layout = %q, want %q", got.Layout, project.LayoutHidden)
	}
	if filepath.Base(got.DataDir) != ".backlog" {
		t.Errorf("DataDir = %q, want it to end in .backlog", got.DataDir)
	}
	if _, err := os.Stat(got.ConfigPath); err != nil {
		t.Errorf("ConfigPath %q does not exist: %v", got.ConfigPath, err)
	}
}

func TestDiscoverPrefersStandardOverHidden(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "backlog", "config.yml"), "project_name: Standard\n")
	writeFile(t, filepath.Join(root, ".backlog", "config.yml"), "project_name: Hidden\n")

	got, err := project.Discover(root)
	if err != nil {
		t.Fatalf("Discover: %v", err)
	}
	if got.Layout != project.LayoutStandard {
		t.Errorf("Layout = %q, want the standard layout to win", got.Layout)
	}
}

func TestDiscoverConfigYamlExtension(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "backlog", "config.yaml"), "project_name: Sample\n")

	got, err := project.Discover(root)
	if err != nil {
		t.Fatalf("Discover: %v", err)
	}
	if filepath.Base(got.ConfigPath) != "config.yaml" {
		t.Errorf("ConfigPath = %q, want the .yaml spelling to be accepted", got.ConfigPath)
	}
}

func TestDiscoverCustomDirectoryFromRootConfig(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "backlog.config.yml"),
		"project_name: Sample\nbacklog_directory: \"planning/work\"\n")
	if err := os.MkdirAll(filepath.Join(root, "planning", "work"), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	got, err := project.Discover(root)
	if err != nil {
		t.Fatalf("Discover: %v", err)
	}
	if got.Layout != project.LayoutCustom {
		t.Errorf("Layout = %q, want %q", got.Layout, project.LayoutCustom)
	}
	if want := filepath.Join(root, "planning", "work"); got.DataDir != want {
		t.Errorf("DataDir = %q, want %q", got.DataDir, want)
	}
}

// A root config that points outside its own project would let a registered
// folder reach arbitrary parts of the filesystem.
func TestDiscoverRejectsEscapingBacklogDirectory(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "backlog.config.yml"),
		"backlog_directory: ../../elsewhere\n")

	_, err := project.Discover(root)
	if !errors.Is(err, project.ErrNotFound) {
		t.Fatalf("err = %v, want ErrNotFound", err)
	}
}

func TestDiscoverNotAProject(t *testing.T) {
	cases := map[string]func(t *testing.T) string{
		"empty directory": func(t *testing.T) string { return t.TempDir() },
		"missing directory": func(t *testing.T) string {
			return filepath.Join(t.TempDir(), "nope")
		},
		"a file, not a directory": func(t *testing.T) string {
			path := filepath.Join(t.TempDir(), "file.txt")
			writeFile(t, path, "")
			return path
		},
		"backlog directory without a config": func(t *testing.T) string {
			root := t.TempDir()
			if err := os.MkdirAll(filepath.Join(root, "backlog", "tasks"), 0o755); err != nil {
				t.Fatalf("mkdir: %v", err)
			}
			return root
		},
		"root config without backlog_directory": func(t *testing.T) string {
			root := t.TempDir()
			writeFile(t, filepath.Join(root, "backlog.config.yml"), "project_name: Sample\n")
			return root
		},
	}

	for name, setup := range cases {
		t.Run(name, func(t *testing.T) {
			_, err := project.Discover(setup(t))
			if !errors.Is(err, project.ErrNotFound) {
				t.Fatalf("err = %v, want ErrNotFound", err)
			}
		})
	}
}
