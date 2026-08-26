// Package settings holds the preferences Muster keeps for itself.
//
// Separate from the project registry: that file says where projects are and a
// person edits it by hand. This one holds how the application behaves and is
// written by the application, so mixing them would mean rewriting a
// hand-edited file every time a toggle is flipped.
package settings

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/adrg/xdg"
	"github.com/goccy/go-yaml"
)

// FileName is the settings file inside the Muster config directory.
const FileName = "settings.yml"

// AppDir is Muster's directory under the XDG config home.
const AppDir = "muster"

// WindowBehaviour is what happens when the window is closed.
type WindowBehaviour string

const (
	// BehaviourQuit closes the application, like an ordinary window.
	BehaviourQuit WindowBehaviour = "quit"
	// BehaviourTray leaves it resident in the system tray.
	BehaviourTray WindowBehaviour = "tray"
)

// TaskView is where a task is shown when it is opened.
type TaskView string

const (
	// ViewPanel is the column against the edge of the window.
	ViewPanel TaskView = "panel"
	// ViewCentred is a reading view over the middle of the board, which is
	// easier on a wide screen where the edge is a long way from the eye.
	ViewCentred TaskView = "centred"
)

// Settings is the whole of Muster's own preferences.
type Settings struct {
	// OnWindowClose is quit or tray. Neither is right for everyone, so this is
	// a preference rather than a decision.
	OnWindowClose WindowBehaviour `yaml:"on_window_close" json:"onWindowClose"`
	// TaskView is how an opened task is shown.
	TaskView TaskView `yaml:"task_view" json:"taskView"`
	// GroupBy is what the board groups cards by: empty, project or milestone.
	GroupBy string `yaml:"group_by" json:"groupBy"`
}

// Defaults are what a first run gets: ordinary window behaviour and the side
// panel, because both are the least surprising.
func Defaults() Settings {
	return Settings{
		OnWindowClose: BehaviourQuit,
		TaskView:      ViewPanel,
		GroupBy:       "",
	}
}

// Path is where the settings file lives.
func Path() string {
	return filepath.Join(xdg.ConfigHome, AppDir, FileName)
}

// Load reads the settings, falling back to the defaults for anything unset.
func Load() (Settings, error) { return LoadFrom(Path()) }

// LoadFrom reads settings from an explicit path.
//
// A missing file is the ordinary first-run case and is not an error.
func LoadFrom(path string) (Settings, error) {
	s := Defaults()

	raw, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return s, nil
		}
		return s, fmt.Errorf("reading %s: %w", path, err)
	}
	if err := yaml.Unmarshal(raw, &s); err != nil {
		return Defaults(), fmt.Errorf("parsing %s: %w", path, err)
	}
	return s.normalised(), nil
}

// Save writes the settings, creating the directory if needed.
func (s Settings) Save() error { return s.SaveTo(Path()) }

// SaveTo writes the settings to an explicit path.
func (s Settings) SaveTo(path string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("creating %s: %w", filepath.Dir(path), err)
	}
	raw, err := yaml.Marshal(s.normalised())
	if err != nil {
		return fmt.Errorf("encoding settings: %w", err)
	}
	if err := os.WriteFile(path, raw, 0o644); err != nil {
		return fmt.Errorf("writing %s: %w", path, err)
	}
	return nil
}

// normalised replaces anything unrecognised with the default, so a hand-edited
// typo degrades to sane behaviour rather than to an unusable window.
func (s Settings) normalised() Settings {
	out := s
	if out.OnWindowClose != BehaviourTray {
		out.OnWindowClose = BehaviourQuit
	}
	if out.TaskView != ViewCentred {
		out.TaskView = ViewPanel
	}
	if out.GroupBy != "project" && out.GroupBy != "milestone" {
		out.GroupBy = ""
	}
	return out
}
