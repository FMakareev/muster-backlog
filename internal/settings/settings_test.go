package settings_test

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/FMakareev/muster-backlog/internal/settings"
)

// same compares settings, which hold a map and so cannot use ==.
func same(a, b settings.Settings) bool { return reflect.DeepEqual(a, b) }

// A first run has no settings file, and that is not a failure.
func TestMissingFileGivesDefaults(t *testing.T) {
	s, err := settings.LoadFrom(filepath.Join(t.TempDir(), "settings.yml"))
	if err != nil {
		t.Fatalf("LoadFrom: %v", err)
	}
	if !same(s, settings.Defaults()) {
		t.Errorf("settings = %+v, want the defaults", s)
	}
}

func TestRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nested", "settings.yml")
	want := settings.Settings{
		OnWindowClose:  settings.BehaviourTray,
		TaskView:       settings.ViewCentred,
		GroupBy:        "milestone",
		WIPLimits:      map[string]int{"In Progress": 3},
		StaleAfterDays: 14,
	}
	if err := want.SaveTo(path); err != nil {
		t.Fatalf("SaveTo: %v", err)
	}

	got, err := settings.LoadFrom(path)
	if err != nil {
		t.Fatalf("LoadFrom: %v", err)
	}
	if !same(got, want) {
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
	if !same(got, settings.Defaults()) {
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
	if !same(got, settings.Defaults()) {
		t.Errorf("settings = %+v, want the defaults alongside the error", got)
	}
}

// A limit of zero or less is no limit, not a column nobody may use.
func TestNonPositiveWIPLimitsAreDropped(t *testing.T) {
	path := filepath.Join(t.TempDir(), "settings.yml")
	content := "wip_limits:\n  In Progress: 3\n  Done: 0\n  To Do: -1\n"
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	got, err := settings.LoadFrom(path)
	if err != nil {
		t.Fatalf("LoadFrom: %v", err)
	}
	if got.WIPLimits["In Progress"] != 3 {
		t.Errorf("limits = %v, want In Progress kept", got.WIPLimits)
	}
	if _, ok := got.WIPLimits["Done"]; ok {
		t.Errorf("limits = %v, want a zero limit dropped", got.WIPLimits)
	}
	if _, ok := got.WIPLimits["To Do"]; ok {
		t.Errorf("limits = %v, want a negative limit dropped", got.WIPLimits)
	}
}

func TestStaleThresholdFallsBack(t *testing.T) {
	path := filepath.Join(t.TempDir(), "settings.yml")
	if err := os.WriteFile(path, []byte("stale_after_days: 0\n"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	got, err := settings.LoadFrom(path)
	if err != nil {
		t.Fatalf("LoadFrom: %v", err)
	}
	if got.StaleAfterDays != 30 {
		t.Errorf("StaleAfterDays = %d, want the default", got.StaleAfterDays)
	}
}
