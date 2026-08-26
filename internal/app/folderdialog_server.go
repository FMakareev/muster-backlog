//go:build server

package app

// FolderDialogAvailable is false in the server build, which has no dialogs.
func FolderDialogAvailable() bool { return false }
