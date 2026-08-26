package backlog

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/goccy/go-yaml"
)

// configFields mirrors the keys 1.48.0 understands. Only what Muster actually
// needs is surfaced on Config; the rest is read so that an unknown key never
// causes a failure.
type configFields struct {
	ProjectName      string   `yaml:"project_name"`
	DefaultStatus    string   `yaml:"default_status"`
	Statuses         []string `yaml:"statuses"`
	Labels           []string `yaml:"labels"`
	Types            []string `yaml:"types"`
	Priorities       []string `yaml:"priorities"`
	TaskPrefix       string   `yaml:"task_prefix"`
	DefinitionOfDone []string `yaml:"definition_of_done"`
	ZeroPaddedIDs    *int     `yaml:"zero_padded_ids"`
	HideEmptyColumns *bool    `yaml:"hide_empty_columns"`
	BacklogDirectory string   `yaml:"backlog_directory"`
}

// LoadConfig reads a project's config file.
//
// It tries YAML first and falls back to a tolerant line reader, because
// Backlog.md 1.48.0 parses its own config with a hand-rolled line reader rather
// than a YAML parser. Files therefore exist that the CLI is perfectly happy
// with and a strict parser would reject, and refusing to read a project the
// tool that owns the format accepts is not a defensible position.
//
// Unset keys keep Backlog.md's own defaults.
func LoadConfig(path string) (Config, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return Config{}, fmt.Errorf("reading %s: %w", path, err)
	}
	return ParseConfig(normaliseNewlines(string(raw)))
}

// ParseConfig parses config file contents.
func ParseConfig(text string) (Config, error) {
	cfg := DefaultConfig()

	var fields configFields
	if err := yaml.Unmarshal([]byte(text), &fields); err != nil {
		return parseConfigLines(text)
	}
	applyConfig(&cfg, fields)
	return cfg, nil
}

func applyConfig(cfg *Config, f configFields) {
	if f.ProjectName != "" {
		cfg.ProjectName = f.ProjectName
	}
	if f.DefaultStatus != "" {
		cfg.DefaultStatus = f.DefaultStatus
	}
	if len(f.Statuses) > 0 {
		cfg.Statuses = f.Statuses
	}
	if f.Labels != nil {
		cfg.Labels = f.Labels
	}
	if len(f.Types) > 0 {
		cfg.Types = f.Types
	}
	if len(f.Priorities) > 0 {
		cfg.Priorities = f.Priorities
	}
	if f.TaskPrefix != "" {
		cfg.TaskPrefix = f.TaskPrefix
	}
	if f.DefinitionOfDone != nil {
		cfg.DefinitionOfDone = f.DefinitionOfDone
	}
	if f.ZeroPaddedIDs != nil {
		cfg.ZeroPaddedIDs = *f.ZeroPaddedIDs
	}
	if f.HideEmptyColumns != nil {
		cfg.HideEmptyColumns = *f.HideEmptyColumns
	}
	if f.BacklogDirectory != "" {
		cfg.BacklogDirectory = f.BacklogDirectory
	}
}

// parseConfigLines is the fallback: a line reader shaped like the CLI's own.
// It accepts inline ["a", "b"] and block "- a" lists and strips quotes.
func parseConfigLines(text string) (Config, error) {
	cfg := DefaultConfig()
	var fields configFields

	lines := strings.Split(text, "\n")
	for i := 0; i < len(lines); i++ {
		key, value, ok := strings.Cut(lines[i], ":")
		key = strings.TrimSpace(key)
		if !ok || key == "" || strings.HasPrefix(key, "#") || strings.HasPrefix(key, "-") {
			continue
		}
		value = strings.TrimSpace(value)

		var list []string
		if value == "" {
			// A block list continues on the following lines.
			for j := i + 1; j < len(lines); j++ {
				item := strings.TrimSpace(lines[j])
				if !strings.HasPrefix(item, "- ") {
					break
				}
				list = append(list, unquote(strings.TrimPrefix(item, "- ")))
				i = j
			}
		} else if strings.HasPrefix(value, "[") {
			list = splitInlineList(value)
		}

		switch key {
		case "project_name":
			fields.ProjectName = unquote(value)
		case "default_status":
			fields.DefaultStatus = unquote(value)
		case "statuses":
			fields.Statuses = list
		case "labels":
			fields.Labels = orEmpty(list, value)
		case "types":
			fields.Types = list
		case "priorities":
			fields.Priorities = list
		case "task_prefix":
			fields.TaskPrefix = unquote(value)
		case "definition_of_done":
			fields.DefinitionOfDone = orEmpty(list, value)
		case "backlog_directory":
			fields.BacklogDirectory = unquote(value)
		case "zero_padded_ids":
			if n, err := strconv.Atoi(unquote(value)); err == nil {
				fields.ZeroPaddedIDs = &n
			}
		case "hide_empty_columns":
			if b, err := strconv.ParseBool(unquote(value)); err == nil {
				fields.HideEmptyColumns = &b
			}
		}
	}

	applyConfig(&cfg, fields)
	return cfg, nil
}

// orEmpty distinguishes an explicitly empty list from an absent key, so that
// `labels: []` clears the defaults rather than being ignored.
func orEmpty(list []string, value string) []string {
	if len(list) > 0 {
		return list
	}
	if strings.TrimSpace(value) == "[]" {
		return []string{}
	}
	return nil
}

func splitInlineList(value string) []string {
	value = strings.TrimSuffix(strings.TrimPrefix(strings.TrimSpace(value), "["), "]")
	if strings.TrimSpace(value) == "" {
		return nil
	}
	parts := strings.Split(value, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if item := unquote(strings.TrimSpace(p)); item != "" {
			out = append(out, item)
		}
	}
	return out
}

func unquote(s string) string {
	s = strings.TrimSpace(s)
	// Strip a trailing inline comment only when the value is not quoted.
	if !strings.HasPrefix(s, `"`) && !strings.HasPrefix(s, `'`) {
		if i := strings.Index(s, " #"); i >= 0 {
			s = strings.TrimSpace(s[:i])
		}
	}
	return strings.Trim(s, `"'`)
}
