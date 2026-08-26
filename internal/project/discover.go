// Package project locates Backlog.md projects on disk.
//
// Muster never assumes a layout: Backlog.md supports three of them, and one of
// the author's own projects uses the least common. Discovery is therefore a
// probe rather than a path join. See the format contract, section 3.1.
package project

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Layout names how a project stores its Backlog.md data.
type Layout string

const (
	// LayoutStandard is a "backlog" directory at the project root.
	LayoutStandard Layout = "backlog"
	// LayoutHidden is a ".backlog" directory at the project root.
	LayoutHidden Layout = ".backlog"
	// LayoutCustom is a directory named by backlog_directory in a root
	// backlog.config.yml, written by `backlog init --config-location root`.
	LayoutCustom Layout = "custom"
)

// rootConfigName is the only file that may declare a custom data directory.
const rootConfigName = "backlog.config.yml"

// ErrNotFound reports that a directory holds no Backlog.md project. It is an
// expected outcome, not a failure: a registered folder may simply not have been
// initialised yet, and the UI offers to initialise it.
var ErrNotFound = errors.New("no Backlog.md project found")

// Location is a resolved project on disk.
type Location struct {
	// Root is the project directory as registered.
	Root string
	// DataDir holds tasks, drafts, milestones, docs, decisions and archive.
	DataDir string
	// ConfigPath is the project's config file.
	ConfigPath string
	// Layout is how the data directory was found.
	Layout Layout
}

// Sub returns the path of one of the project's entity directories, whether or
// not it exists. `backlog init` creates all of them and leaves the unused ones
// empty, but a hand-made project may be missing some.
func (l Location) Sub(name string) string {
	return filepath.Join(l.DataDir, name)
}

// configNames are accepted inside a data directory, in probe order.
var configNames = []string{"config.yml", "config.yaml"}

// Discover locates the Backlog.md project rooted at dir.
//
// The probe order matches the CLI's own resolution: the standard directory,
// then the hidden one, then a root config naming a custom path. It returns
// ErrNotFound when none of them holds a config file.
func Discover(dir string) (Location, error) {
	root, err := filepath.Abs(dir)
	if err != nil {
		return Location{}, fmt.Errorf("resolving %q: %w", dir, err)
	}

	info, err := os.Stat(root)
	if err != nil {
		if os.IsNotExist(err) {
			return Location{}, fmt.Errorf("%q does not exist: %w", root, ErrNotFound)
		}
		return Location{}, fmt.Errorf("reading %q: %w", root, err)
	}
	if !info.IsDir() {
		return Location{}, fmt.Errorf("%q is not a directory: %w", root, ErrNotFound)
	}

	for _, layout := range []Layout{LayoutStandard, LayoutHidden} {
		dataDir := filepath.Join(root, string(layout))
		if cfg, ok := findConfig(dataDir); ok {
			return Location{Root: root, DataDir: dataDir, ConfigPath: cfg, Layout: layout}, nil
		}
	}

	// A root config is the only place a custom data directory can be declared.
	rootConfig := filepath.Join(root, rootConfigName)
	if raw, err := os.ReadFile(rootConfig); err == nil {
		custom := backlogDirectory(string(raw))
		if custom == "" {
			return Location{}, fmt.Errorf(
				"%s declares no backlog_directory: %w", rootConfigName, ErrNotFound)
		}
		dataDir := filepath.Join(root, filepath.FromSlash(custom))
		if !within(root, dataDir) {
			return Location{}, fmt.Errorf(
				"backlog_directory %q escapes the project directory: %w", custom, ErrNotFound)
		}
		if info, err := os.Stat(dataDir); err != nil || !info.IsDir() {
			return Location{}, fmt.Errorf(
				"backlog_directory %q does not exist: %w", custom, ErrNotFound)
		}
		return Location{
			Root: root, DataDir: dataDir, ConfigPath: rootConfig, Layout: LayoutCustom,
		}, nil
	}

	return Location{}, fmt.Errorf("%q: %w", root, ErrNotFound)
}

// findConfig returns the config file inside a candidate data directory.
func findConfig(dataDir string) (string, bool) {
	for _, name := range configNames {
		path := filepath.Join(dataDir, name)
		if info, err := os.Stat(path); err == nil && !info.IsDir() {
			return path, true
		}
	}
	return "", false
}

// backlogDirectory extracts backlog_directory from a root config.
//
// This is a line reader rather than a YAML parse on purpose: Backlog.md 1.48.0
// reads its own config with a hand-rolled line reader, so files exist that it
// accepts and a strict YAML parser would reject. Discovery has to be at least as
// permissive as the tool that wrote the file.
func backlogDirectory(raw string) string {
	for _, line := range strings.Split(raw, "\n") {
		line = strings.TrimSpace(strings.TrimSuffix(line, "\r"))
		key, value, found := strings.Cut(line, ":")
		if !found || strings.TrimSpace(key) != "backlog_directory" {
			continue
		}
		return strings.Trim(strings.TrimSpace(value), `"'`)
	}
	return ""
}

// within reports whether path stays inside root.
func within(root, path string) bool {
	rel, err := filepath.Rel(root, path)
	if err != nil {
		return false
	}
	return rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator))
}
