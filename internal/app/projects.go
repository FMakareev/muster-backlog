package app

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/FMakareev/muster-backlog/internal/backlogcli"
	"github.com/FMakareev/muster-backlog/internal/project"
	"github.com/FMakareev/muster-backlog/internal/registry"
)

// FolderView is what is known about a folder before it is registered.
//
// The answer decides what is offered rather than what is asked: a folder that
// already holds a backlog wants registering, one that does not wants
// initialising, and one that is not a repository wants the no-git path chosen
// for it instead of being asked a question it would fail on.
type FolderView struct {
	// Path is the folder, resolved the way the registry would resolve it.
	Path string `json:"path"`
	// Name is what the project would be called if nothing overrode it.
	Name string `json:"name"`
	// Exists and IsDir are separate: a file is a different mistake from a
	// path that is not there at all.
	Exists bool `json:"exists"`
	IsDir  bool `json:"isDir"`
	// IsGit reports a git repository at the folder itself.
	IsGit bool `json:"isGit"`
	// HasBacklog reports a Backlog.md project already in place.
	HasBacklog bool `json:"hasBacklog"`
	// Layout is how that project's data directory was found, when there is one.
	Layout string `json:"layout"`
	// Registered reports that Muster already holds this folder.
	Registered bool `json:"registered"`
	// Problem explains a folder that cannot be used at all.
	Problem string `json:"problem"`
}

// InspectFolder reports what a folder is, before anything is written to it.
func (s *BoardService) InspectFolder(path string) FolderView {
	view := FolderView{}

	trimmed := strings.TrimSpace(path)
	if trimmed == "" {
		view.Problem = "No folder given."
		return view
	}

	resolved, err := registry.Expand(trimmed)
	if err != nil {
		view.Problem = err.Error()
		return view
	}
	view.Path = resolved
	view.Name = filepath.Base(resolved)

	info, err := os.Stat(resolved)
	switch {
	case err == nil:
		view.Exists = true
		view.IsDir = info.IsDir()
		if !view.IsDir {
			view.Problem = "That is a file, not a folder."
		}
	case os.IsNotExist(err):
		view.Problem = "There is nothing at that path."
	default:
		view.Problem = err.Error()
	}

	for _, p := range s.store.Projects() {
		if p.Registry.Path == resolved {
			view.Registered = true
		}
	}

	if !view.IsDir {
		return view
	}

	if info, err := os.Stat(filepath.Join(resolved, ".git")); err == nil {
		// A worktree keeps a file rather than a directory at .git, and both
		// count: backlog init runs in either.
		view.IsGit = info.IsDir() || info.Mode().IsRegular()
	}

	if loc, err := project.Discover(resolved); err == nil {
		view.HasBacklog = true
		view.Layout = string(loc.Layout)
	}
	return view
}

// ProjectEdit is the whole of what the registry holds about one project.
//
// Sent complete rather than as a change, because the entry is rewritten as a
// whole: a partial edit would have to guess what the missing fields meant.
type ProjectEdit struct {
	Name   string `json:"name"`
	Colour string `json:"colour"`
	Hidden bool   `json:"hidden"`
}

// AddProject registers an existing folder.
//
// The folder must already hold a Backlog.md project: registering a folder that
// does not is how a board fills with entries that explain themselves as broken.
// Initialising is a separate, deliberate act.
func (s *BoardService) AddProject(path string, edit ProjectEdit) WriteResult {
	view := s.InspectFolder(path)
	switch {
	case view.Problem != "":
		return WriteResult{Problem: &Problem{
			Kind: ProblemRegistry, Title: "That folder cannot be added",
			Detail: view.Problem, Path: path,
		}}
	case view.Registered:
		return WriteResult{Problem: &Problem{
			Kind: ProblemRegistry, Title: "Already registered",
			Detail: "Muster already holds this folder.", Path: view.Path,
		}}
	case !view.HasBacklog:
		return WriteResult{Problem: &Problem{
			Kind:  ProblemRegistry,
			Title: "There is no Backlog.md project in that folder",
			Detail: "Initialise one here first, or choose the folder that " +
				"holds the backlog.",
			Path: view.Path,
		}}
	}

	return s.registryWrite("The project could not be added", view.Path,
		registry.Add(s.RegistryPath(), registry.Entry{
			// Written as given, so a person's ~ stays a ~ in their own file.
			Path:   strings.TrimSpace(path),
			Name:   strings.TrimSpace(edit.Name),
			Colour: strings.TrimSpace(edit.Colour),
			Hidden: edit.Hidden,
		}))
}

// SaveProject rewrites one project's entry.
func (s *BoardService) SaveProject(path string, edit ProjectEdit) WriteResult {
	written, err := s.writtenPath(path)
	if err != nil {
		return s.registryWrite("The project could not be changed", path, err)
	}
	return s.registryWrite("The project could not be changed", path,
		registry.Update(s.RegistryPath(), path, registry.Entry{
			Path:   written,
			Name:   strings.TrimSpace(edit.Name),
			Colour: strings.TrimSpace(edit.Colour),
			Hidden: edit.Hidden,
		}))
}

// RemoveProject unregisters a folder. Nothing on disk is touched: Muster does
// not delete backlogs, and a person who wanted the tasks gone would say so to
// the CLI.
func (s *BoardService) RemoveProject(path string) WriteResult {
	return s.registryWrite("The project could not be removed", path,
		registry.Remove(s.RegistryPath(), path))
}

// MoveProject puts a project at a position in the registry, which is the order
// it appears in everywhere.
func (s *BoardService) MoveProject(path string, to int) WriteResult {
	return s.registryWrite("The project could not be moved", path,
		registry.Move(s.RegistryPath(), path, to))
}

// writtenPath returns the path as the registry file spells it, so rewriting an
// entry does not quietly expand a person's ~ into an absolute path.
func (s *BoardService) writtenPath(path string) (string, error) {
	reg, err := registry.LoadFrom(s.RegistryPath())
	if err != nil && !isNoRegistry(err) {
		return "", err
	}
	resolved, err := registry.Expand(path)
	if err != nil {
		return "", err
	}
	for _, p := range reg.Projects {
		if p.Path == resolved || p.Written == path {
			return p.Written, nil
		}
	}
	return "", fmt.Errorf("%w: %s", registry.ErrNoSuchProject, path)
}

// registryWrite turns a writer error into a report and reloads on success.
func (s *BoardService) registryWrite(title, path string, err error) WriteResult {
	if err != nil {
		return WriteResult{Problem: &Problem{
			Kind: ProblemRegistry, Title: title, Detail: err.Error(), Path: path,
		}}
	}
	s.reload()
	emit(EventRegistryChanged, struct{}{})
	return WriteResult{OK: true}
}

// InitFolder is the form behind `backlog init`.
//
// Every field is optional: the CLI is run with --defaults, so anything left
// empty is answered by Backlog.md itself rather than by a default of Muster's
// own invention.
type InitFolder struct {
	Path              string `json:"path"`
	Name              string `json:"name"`
	BacklogDir        string `json:"backlogDir"`
	ConfigLocation    string `json:"configLocation"`
	TaskPrefix        string `json:"taskPrefix"`
	ZeroPaddedIDs     int    `json:"zeroPaddedIds"`
	NoGit             bool   `json:"noGit"`
	AgentInstructions string `json:"agentInstructions"`
	IntegrationMode   string `json:"integrationMode"`
	// Colour is Muster's own, applied when the new project is registered.
	Colour string `json:"colour"`
}

// InitProject initialises a folder as a Backlog.md project and registers it.
//
// The two halves are one action deliberately: a folder initialised and then
// not registered is work the person would have to repeat, and the CLI has
// already written to the folder by then.
func (s *BoardService) InitProject(form InitFolder) WriteResult {
	view := s.InspectFolder(form.Path)
	switch {
	case view.Problem != "":
		return WriteResult{Problem: &Problem{
			Kind: ProblemRegistry, Title: "That folder cannot be initialised",
			Detail: view.Problem, Path: form.Path,
		}}
	case view.HasBacklog:
		return WriteResult{Problem: &Problem{
			Kind: ProblemRegistry, Title: "There is already a backlog here",
			Detail: "Add the folder instead of initialising it.", Path: view.Path,
		}}
	}

	s.mu.Lock()
	cli := s.cli
	s.mu.Unlock()
	if cli == nil {
		return WriteResult{Problem: s.cliProblem()}
	}

	// A folder that is not a repository takes the no-git path whether or not
	// the form asked for it: init would fail on it otherwise, and the question
	// has only one possible answer.
	noGit := form.NoGit || !view.IsGit

	// The name is always passed, never left to the CLI's own default.
	//
	// --defaults does not cover it: the project name is a positional argument,
	// and without one 1.48.0 prompts for it even with --defaults set. With no
	// terminal on the other end that prompt cannot be answered, and the
	// command exits 0 having created nothing at all - a silent success that
	// leaves an empty folder behind.
	name := strings.TrimSpace(form.Name)
	if name == "" {
		name = filepath.Base(view.Path)
	}

	err := cli.Init(context.Background(), view.Path, backlogcli.InitOptions{
		Name:              name,
		BacklogDir:        strings.TrimSpace(form.BacklogDir),
		ConfigLocation:    strings.TrimSpace(form.ConfigLocation),
		TaskPrefix:        strings.TrimSpace(form.TaskPrefix),
		ZeroPaddedIDs:     form.ZeroPaddedIDs,
		NoGit:             noGit,
		AgentInstructions: strings.TrimSpace(form.AgentInstructions),
		IntegrationMode:   strings.TrimSpace(form.IntegrationMode),
	})
	if err != nil {
		return WriteResult{Problem: &Problem{
			Kind:  ProblemCLI,
			Title: "Backlog.md could not initialise that folder",
			// The CLI's own output, not a summary of it: init writes files,
			// and a half-written folder needs the real reason.
			Detail: initDetail(err),
			Path:   view.Path,
		}}
	}

	// Trust nothing: the CLI has exited without creating a backlog before, and
	// letting the registration step discover that would report it as the wrong
	// problem entirely.
	if _, err := project.Discover(view.Path); err != nil {
		return WriteResult{Problem: &Problem{
			Kind:  ProblemCLI,
			Title: "Backlog.md reported success but created nothing",
			Detail: "The folder is still empty. This is what happens when the " +
				"CLI asks a question it cannot ask here: " + err.Error(),
			Path: view.Path,
		}}
	}

	return s.AddProject(form.Path, ProjectEdit{
		Name: name, Colour: strings.TrimSpace(form.Colour),
	})
}

// initDetail surfaces what the CLI printed, falling back to the error itself.
func initDetail(err error) string {
	var cmd *backlogcli.CommandError
	if errors.As(err, &cmd) {
		// stderr first, then stdout: init explains itself in both, depending
		// on which check it failed.
		for _, out := range []string{cmd.Stderr, cmd.Stdout} {
			if trimmed := strings.TrimSpace(out); trimmed != "" {
				return trimmed
			}
		}
	}
	return err.Error()
}
