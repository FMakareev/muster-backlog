// Package registry reads Muster's own configuration: which folders hold
// Backlog.md projects and how they are displayed.
//
// This is the only configuration Muster owns. Everything about how a project
// works - its statuses, priorities, types, labels and task prefix - is declared
// by that project's own config.yml and is never copied in here, because a copy
// would silently go stale the moment the project changed.
package registry

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/adrg/xdg"
	"github.com/goccy/go-yaml"

	"github.com/FMakareev/muster-backlog/internal/project"
)

// FileName is the registry's name inside the Muster config directory.
const FileName = "projects.yml"

// AppDir is Muster's directory under the XDG config home.
const AppDir = "muster"

// ErrNoRegistry reports that no registry file exists yet. This is a first-run
// state rather than a failure: the caller offers to add a first project.
var ErrNoRegistry = errors.New("no registry file")

// Entry is one registered folder, as written by the user.
type Entry struct {
	// Path is the project directory. It may be written with a leading ~.
	Path string `yaml:"path"`
	// Name overrides the display name. Empty means the folder name is used.
	Name string `yaml:"name"`
	// Colour distinguishes the project on a shared board. Empty means the UI
	// picks one.
	Colour string `yaml:"color"`
	// Hidden keeps a project out of the board, list, search and figures
	// without removing it from the registry. It is still loaded, so the
	// Projects screen can show what it holds and unhiding costs nothing.
	//
	// This is Muster's own file, so a field of its own is allowed here; the
	// rule against inventing fields is about the Backlog.md format, which this
	// is not.
	Hidden bool `yaml:"hidden,omitempty"`
}

// File is the on-disk shape of projects.yml.
type File struct {
	Projects []Entry `yaml:"projects"`
	// WIPLimits are advisory per-status ceilings, keyed by status name. They
	// are a display signal only and never block anything.
	WIPLimits map[string]int `yaml:"wip_limits"`
}

// Project is a registry entry after resolution.
type Project struct {
	Entry
	// Written is the path exactly as the file spells it, before the leading ~
	// was expanded. Resolution overwrites Entry.Path with the absolute form,
	// which is what everything else keys on; this is kept so that rewriting an
	// entry does not quietly turn a person's ~/Dev/thing into an absolute
	// path in their own file.
	Written string
	// DisplayName is Name, or the folder name when Name is empty.
	DisplayName string
	// Location is where the Backlog.md data was found. Zero when Err is set.
	Location project.Location
	// Err explains why this entry could not be resolved. The project is shown
	// as degraded rather than dropped, because a folder silently missing from
	// the board is worse than one that says why it is broken.
	Err error
}

// OK reports whether the project resolved to a Backlog.md project on disk.
func (p Project) OK() bool { return p.Err == nil }

// Registry is the resolved registry.
type Registry struct {
	// Path is the file this was read from.
	Path string
	// Projects preserves the order written in the file: that order is the
	// order the user sees, and reordering is done by editing the file.
	Projects []Project
	// WIPLimits are advisory per-status ceilings.
	WIPLimits map[string]int
}

// DefaultPath is where the registry lives: $XDG_CONFIG_HOME/muster/projects.yml,
// falling back to ~/.config/muster/projects.yml when XDG_CONFIG_HOME is unset,
// which is what the xdg package resolves for us.
func DefaultPath() string {
	return filepath.Join(xdg.ConfigHome, AppDir, FileName)
}

// Load reads and resolves the registry at its default location.
func Load() (Registry, error) {
	return LoadFrom(DefaultPath())
}

// LoadFrom reads and resolves the registry at an explicit path.
//
// One bad entry never sinks the rest: a folder that has moved, or that was
// never initialised, is reported on its own Project and the others still load.
// A missing file returns ErrNoRegistry with an otherwise usable empty registry.
func LoadFrom(path string) (Registry, error) {
	reg := Registry{Path: path, WIPLimits: map[string]int{}}

	raw, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return reg, fmt.Errorf("%s: %w", path, ErrNoRegistry)
		}
		return reg, fmt.Errorf("reading %s: %w", path, err)
	}

	var file File
	if err := yaml.Unmarshal(raw, &file); err != nil {
		return reg, fmt.Errorf("parsing %s: %w", path, err)
	}

	if file.WIPLimits != nil {
		reg.WIPLimits = file.WIPLimits
	}

	seen := make(map[string]int, len(file.Projects))
	for i, entry := range file.Projects {
		reg.Projects = append(reg.Projects, resolve(i, entry, seen))
	}
	return reg, nil
}

// resolve turns one written entry into a Project, located or explained.
func resolve(index int, entry Entry, seen map[string]int) Project {
	p := Project{Entry: entry, Written: entry.Path}

	if strings.TrimSpace(entry.Path) == "" {
		p.Err = fmt.Errorf("entry %d has no path", index+1)
		return p
	}

	expanded, err := expand(entry.Path)
	if err != nil {
		p.Err = err
		return p
	}
	p.Path = expanded
	p.DisplayName = displayName(entry.Name, expanded)

	if first, dup := seen[expanded]; dup {
		p.Err = fmt.Errorf("%s is already registered as entry %d", expanded, first+1)
		return p
	}
	seen[expanded] = index

	loc, err := project.Discover(expanded)
	if err != nil {
		p.Err = err
		return p
	}
	p.Location = loc
	return p
}

// Expand resolves a leading ~ and makes a path absolute, the same way the
// registry does when it reads one. Exported so that anything checking a folder
// before it is registered agrees with what registering it would mean.
func Expand(path string) (string, error) { return expand(path) }

// expand resolves a leading ~ and makes the path absolute, so the registry can
// be written the way a person would write it by hand.
func expand(path string) (string, error) {
	path = strings.TrimSpace(path)
	if path == "~" || strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("expanding %q: %w", path, err)
		}
		path = filepath.Join(home, strings.TrimPrefix(path, "~"))
	}
	abs, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("resolving %q: %w", path, err)
	}
	return abs, nil
}

// displayName falls back to the folder name, which is what a person would call
// the project anyway.
func displayName(name, path string) string {
	if trimmed := strings.TrimSpace(name); trimmed != "" {
		return trimmed
	}
	return filepath.Base(path)
}

// Abbreviate is expand's inverse, for showing a path rather than using one.
//
// The registry path sits in the status bar of every screenshot anybody takes,
// and under a home directory it carries a username there for no reason. `~`
// is how a person writes the path anyway, and it still pastes into a shell.
func Abbreviate(path string) string {
	path = strings.TrimSpace(path)
	home, err := os.UserHomeDir()
	if err != nil || home == "" || path == "" {
		return path
	}
	if path == home {
		return "~"
	}
	if rest, ok := strings.CutPrefix(path, home+string(filepath.Separator)); ok {
		return "~" + string(filepath.Separator) + rest
	}
	return path
}
