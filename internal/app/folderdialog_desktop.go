//go:build !server

package app

import "github.com/wailsapp/wails/v3/pkg/application"

// FolderDialogAvailable reports whether the desktop can show a directory
// chooser.
//
// It is a build-time fact rather than a runtime probe: the server build has no
// dialog at all - it answers every request with "file dialogs not available in
// server mode" - and the interface should not offer a button that can only
// fail. Before the application exists there is nothing to open a dialog over
// either, which is the case during tests.
func FolderDialogAvailable() bool { return application.Get() != nil }
