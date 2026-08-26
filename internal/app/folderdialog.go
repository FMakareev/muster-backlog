package app

import (
	"os"
	"strings"

	"github.com/wailsapp/wails/v3/pkg/application"
)

// CanChooseFolder reports whether the interface should offer a browse button.
//
// Asked rather than assumed: the same frontend runs in the server build, which
// has no dialogs at all, and a button that can only fail is worse than no
// button.
func (s *BoardService) CanChooseFolder() bool { return FolderDialogAvailable() }

// ChooseFolder opens the desktop's own directory chooser.
//
// Returns the folder, or "" when the person cancelled - which is not an error
// and must not be reported as one. Where there is no dialog at all, the
// interface has already hidden the button; this still answers "" rather than
// failing, so a stale window cannot produce a broken state.
func (s *BoardService) ChooseFolder() string {
	if !FolderDialogAvailable() {
		return ""
	}

	app := application.Get()
	if app == nil {
		return ""
	}

	dialog := app.Dialog.OpenFile().
		CanChooseDirectories(true).
		CanChooseFiles(false).
		CanCreateDirectories(true).
		SetTitle("Choose a project folder")

	// Start somewhere useful: beside the projects already registered, since a
	// second project usually lives next to the first, and the home directory
	// otherwise.
	if start := s.dialogStart(); start != "" {
		dialog.SetDirectory(start)
	}

	chosen, err := dialog.PromptForSingleSelection()
	if err != nil {
		// Cancelling is reported as an error by some backends and as an empty
		// string by others. Neither is a failure worth showing.
		return ""
	}
	return strings.TrimSpace(chosen)
}

// dialogStart is where the chooser opens.
func (s *BoardService) dialogStart() string {
	for _, p := range s.store.Projects() {
		if p.OK() {
			return parentOf(p.Registry.Path)
		}
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return home
}

// parentOf is the containing directory, or the path itself at the root.
func parentOf(path string) string {
	if at := strings.LastIndex(path, string(os.PathSeparator)); at > 0 {
		return path[:at]
	}
	return path
}
