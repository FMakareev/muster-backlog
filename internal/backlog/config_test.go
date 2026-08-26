package backlog_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/FMakareev/muster-backlog/internal/backlog"
)

func loadCorpusConfig(t *testing.T, rel string) backlog.Config {
	t.Helper()
	cfg, err := backlog.LoadConfig(filepath.Join(corpus, rel))
	if err != nil {
		t.Fatalf("LoadConfig(%s): %v", rel, err)
	}
	return cfg
}

func TestConfigStatusListsFromCorpus(t *testing.T) {
	defaults := loadCorpusConfig(t, "03-config/statuses-default/config.yml")
	if got, want := strings.Join(defaults.Statuses, "|"), "To Do|In Progress|Done"; got != want {
		t.Errorf("statuses = %q, want %q", got, want)
	}

	withReview := loadCorpusConfig(t, "03-config/statuses-with-in-review/config.yml")
	if len(withReview.Statuses) != 4 {
		t.Errorf("statuses = %v, want four", withReview.Statuses)
	}

	// The synthetic config exists precisely to prove the union algorithm will
	// have to cope with lists that share no order with the common one.
	custom := loadCorpusConfig(t, "03-config/statuses-custom-synthetic/config.yml")
	if len(custom.Statuses) == 0 {
		t.Fatal("custom config declares no statuses")
	}
	if strings.Join(custom.Statuses, "|") == strings.Join(defaults.Statuses, "|") {
		t.Error("the custom fixture should differ from the default list")
	}
}

// Ordering within statuses is meaningful: the board's columns are built from
// these lists, so a parser that returned a set rather than a sequence would
// quietly lose the only ordering information that exists.
func TestConfigPreservesStatusOrder(t *testing.T) {
	cfg, err := backlog.ParseConfig("statuses: [\"Shipped\", \"Doing\", \"Backlog\"]\n")
	if err != nil {
		t.Fatalf("ParseConfig: %v", err)
	}
	if got := strings.Join(cfg.Statuses, "|"); got != "Shipped|Doing|Backlog" {
		t.Errorf("statuses = %q, want the file's order", got)
	}
}

func TestConfigZeroPaddedAndEditor(t *testing.T) {
	cfg := loadCorpusConfig(t, "03-config/zero-padded-and-editor/config.yml")
	if cfg.ZeroPaddedIDs == 0 {
		t.Error("zero_padded_ids not read")
	}
}

func TestRootConfigDeclaresBacklogDirectory(t *testing.T) {
	cfg := loadCorpusConfig(t, "03-config/root-config-synthetic/backlog.config.yml")
	if cfg.BacklogDirectory == "" {
		t.Error("backlog_directory not read from a root config")
	}
}

// task_prefix is per-project configuration, so "task-" must never be assumed.
func TestConfigDefaultsAndOverrides(t *testing.T) {
	cfg, err := backlog.ParseConfig("project_name: Sample\n")
	if err != nil {
		t.Fatalf("ParseConfig: %v", err)
	}
	if cfg.TaskPrefix != "task" {
		t.Errorf("TaskPrefix = %q, want the default", cfg.TaskPrefix)
	}
	if len(cfg.Types) != 7 || len(cfg.Priorities) != 3 {
		t.Errorf("defaults not applied: %d types, %d priorities",
			len(cfg.Types), len(cfg.Priorities))
	}

	custom, err := backlog.ParseConfig("task_prefix: \"story\"\ntypes: [\"a\", \"b\"]\n")
	if err != nil {
		t.Fatalf("ParseConfig: %v", err)
	}
	if custom.TaskPrefix != "story" {
		t.Errorf("TaskPrefix = %q, want story", custom.TaskPrefix)
	}
	if len(custom.Types) != 2 {
		t.Errorf("Types = %v, want the override to replace the defaults", custom.Types)
	}
}

// Backlog.md parses its own config with a hand-rolled line reader, so files
// exist that it accepts and strict YAML rejects. Refusing to read a project the
// tool that owns the format is happy with is not a defensible position.
func TestConfigFallsBackToLineReader(t *testing.T) {
	// A tab-indented block list is invalid YAML but survives a line reader.
	raw := "project_name: Sample\nstatuses:\n\t- To Do\n\t- Done\n"
	cfg, err := backlog.ParseConfig(raw)
	if err != nil {
		t.Fatalf("ParseConfig: %v", err)
	}
	if cfg.ProjectName != "Sample" {
		t.Errorf("ProjectName = %q, want the fallback reader to have read it", cfg.ProjectName)
	}
}

func TestConfigBlockAndInlineListsAgree(t *testing.T) {
	inline, err := backlog.ParseConfig("statuses: [\"To Do\", \"Done\"]\n")
	if err != nil {
		t.Fatalf("ParseConfig: %v", err)
	}
	block, err := backlog.ParseConfig("statuses:\n  - \"To Do\"\n  - \"Done\"\n")
	if err != nil {
		t.Fatalf("ParseConfig: %v", err)
	}
	if strings.Join(inline.Statuses, "|") != strings.Join(block.Statuses, "|") {
		t.Errorf("inline %v and block %v disagree", inline.Statuses, block.Statuses)
	}
}

func TestConfigMissingFile(t *testing.T) {
	_, err := backlog.LoadConfig(filepath.Join(t.TempDir(), "nope.yml"))
	if err == nil {
		t.Fatal("want an error for a missing config")
	}
	if !os.IsNotExist(err) && !strings.Contains(err.Error(), "nope.yml") {
		t.Errorf("err = %v, want it to name the file", err)
	}
}
