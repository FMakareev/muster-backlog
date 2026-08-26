package settings_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/FMakareev/muster-backlog/internal/settings"
)

// A first run has no settings file, and that is not a failure.
func TestMissingFileGivesDefaults(t *testing.T) {
	s, err := settings.LoadFrom(filepath.Join(t.TempDir(), "settings.yml"))
	if err != nil {
		t.Fatalf("LoadFrom: %v", err)
	}
	if s != settings.Defaults() {
		t.Errorf("settings = %+v, want the defaults", s)
	}
}

func TestRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nested", "settings.yml")
	want := settings.Settings{
		OnWindowClose: settings.BehaviourTray,
		TaskView:      settings.ViewCentred,
		GroupBy:       "milestone",
	}
	if err := want.SaveTo(path); err != nil {
		t.Fatalf("SaveTo: %v", err)
	}

	got, err := settings.LoadFrom(path)
	if err != nil {
		t.Fatalf("LoadFrom: %v", err)
	}
	if got != want {
		t.Errorf("settings = %+v, want %+v", got, want)
	}
}

// A hand-edited typo must degrade to sane behaviour rather than to a window
// that cannot be closed or a task that cannot be opened.
func TestUnrecognisedValuesFallBack(t *testing.T) {
	path := filepath.Join(t.TempDir(), "settings.yml")
	content := "on_window_close: minimise-to-the-moon\ntask_view: hologram\ngroup_by: colour\n"
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}

	got, err := settings.LoadFrom(path)
	if err != nil {
		t.Fatalf("LoadFrom: %v", err)
	}
	if got != settings.Defaults() {
		t.Errorf("settings = %+v, want the defaults for every unrecognised value", got)
	}
}

func TestMalformedFileIsReported(t *testing.T) {
	path := filepath.Join(t.TempDir(), "settings.yml")
	if err := os.WriteFile(path, []byte("on_window_close: [unclosed\n"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	got, err := settings.LoadFrom(path)
	if err == nil {
		t.Fatal("want an error for malformed YAML")
	}
	// Even then the caller gets something usable rather than a zero value.
	if got != settings.Defaults() {
		t.Errorf("settings = %+v, want the defaults alongside the error", got)
	}
}
