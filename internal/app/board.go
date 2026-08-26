package app

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/wailsapp/wails/v3/pkg/application"

	"github.com/FMakareev/muster-backlog/internal/backlog"
	"github.com/FMakareev/muster-backlog/internal/backlogcli"
	"github.com/FMakareev/muster-backlog/internal/board"
	"github.com/FMakareev/muster-backlog/internal/registry"
	"github.com/FMakareev/muster-backlog/internal/store"
	"github.com/FMakareev/muster-backlog/internal/watcher"
)

// EventProjectChanged is emitted after a project has been reloaded because
// something changed on disk. The frontend refreshes from it rather than polling.
const EventProjectChanged = "muster:project:changed"

// EventRegistryChanged is emitted when the set of projects itself changes.
const EventRegistryChanged = "muster:registry:changed"

// ProjectChanged is the payload of EventProjectChanged.
type ProjectChanged struct {
	// Project is the registered project path, matching ProjectView.Path.
	Project string `json:"project"`
}

func init() {
	application.RegisterEvent[ProjectChanged](EventProjectChanged)
	application.RegisterEvent[struct{}](EventRegistryChanged)
}

// ProblemKind classifies a failure so the UI can decide how loudly to show it.
type ProblemKind string

const (
	// ProblemNoRegistry means there is no registry yet. This is first run, not
	// a fault, and the UI offers to add a project rather than showing an error.
	ProblemNoRegistry ProblemKind = "no_registry"
	// ProblemRegistry means the registry itself could not be read.
	ProblemRegistry ProblemKind = "registry"
	// ProblemProject means one project could not be loaded. The others are fine.
	ProblemProject ProblemKind = "project"
	// ProblemFile means one file was skipped during a scan.
	ProblemFile ProblemKind = "file"
)

// Problem is a failure the frontend can render.
//
// Go errors cross the bridge as bare strings, which leaves the UI with nothing
// to lay out and nothing to act on. A problem carries enough structure to be
// shown as something other than a log line.
type Problem struct {
	Kind ProblemKind `json:"kind"`
	// Title is one short line, suitable for a heading.
	Title string `json:"title"`
	// Detail is the underlying message.
	Detail string `json:"detail"`
	// Path is the file or folder concerned, empty when none applies.
	Path string `json:"path"`
}

// ProjectView is one registered project as the frontend sees it.
type ProjectView struct {
	Path       string `json:"path"`
	Name       string `json:"name"`
	Colour     string `json:"colour"`
	OK         bool   `json:"ok"`
	Problem    string `json:"problem"`
	TaskCount  int    `json:"taskCount"`
	DraftCount int    `json:"draftCount"`
	// Statuses is what this project declares, in its own order. The board's
	// columns are the union of these across projects.
	Statuses []string `json:"statuses"`
	// Priorities and Types are what this project configures. They are read
	// from the project rather than assumed, because both are configurable and
	// a hardcoded list would be wrong the first time someone changes one.
	Priorities []string `json:"priorities"`
	Types      []string `json:"types"`
	// Layout is how the project's data directory was found.
	Layout string `json:"layout"`
}

// TaskView is one task together with where it came from.
//
// The entity is exposed as the domain type rather than flattened into a
// hand-written shape: the TypeScript is generated from it, so a mirrored type
// would only be a type that drifts.
type TaskView struct {
	Project       string         `json:"project"`
	ProjectName   string         `json:"projectName"`
	ProjectColour string         `json:"projectColour"`
	Kind          backlog.Kind   `json:"kind"`
	Class         backlog.Class  `json:"class"`
	ID            string         `json:"id"`
	Entity        backlog.Entity `json:"entity"`
}

// QueryInput is the filter set the frontend sends. Every field is optional and
// they combine with AND.
type QueryInput struct {
	Kinds      []string `json:"kinds"`
	Classes    []string `json:"classes"`
	Projects   []string `json:"projects"`
	Statuses   []string `json:"statuses"`
	Milestones []string `json:"milestones"`
	Priorities []string `json:"priorities"`
	Types      []string `json:"types"`
	Labels     []string `json:"labels"`
	Assignees  []string `json:"assignees"`
	Text       string   `json:"text"`
}

// BoardService is the frontend's only entry point to the backend.
type BoardService struct {
	mu       sync.Mutex
	store    *store.Store
	watch    *watcher.Watcher
	problems []Problem
	// cli is the only thing in the application that writes. It is resolved
	// once at startup so a missing binary is one report rather than the same
	// failure arriving on every action.
	cli    *backlogcli.Runner
	cliErr error
	// registryPath is where the registry is read from. It is a field rather
	// than a call to registry.DefaultPath() so that tests can point the service
	// at their own file instead of manipulating the environment - the XDG
	// lookup caches its variables at process start and cannot be moved later.
	registryPath string
}

// NewBoardService returns a service reading the registry from its usual place.
func NewBoardService() *BoardService {
	return NewBoardServiceAt(registry.DefaultPath())
}

// NewBoardServiceAt returns a service reading the registry from a given path.
func NewBoardServiceAt(registryPath string) *BoardService {
	return &BoardService{store: store.New(), registryPath: registryPath}
}

// ServiceName identifies the service in Wails logs.
func (s *BoardService) ServiceName() string { return "muster.board" }

// ServiceStartup loads everything and starts following the filesystem.
//
// A failure to load the registry is not a startup failure: the application
// opens, says what is wrong, and lets the user fix it. Returning an error here
// would leave them with a window that refuses to appear.
func (s *BoardService) ServiceStartup(_ context.Context, _ application.ServiceOptions) error {
	s.reload()
	s.resolveCLI()
	return nil
}

// ServiceShutdown stops the watcher.
func (s *BoardService) ServiceShutdown() error {
	s.mu.Lock()
	w := s.watch
	s.watch = nil
	s.mu.Unlock()

	if w != nil {
		return w.Close()
	}
	return nil
}

// Reload re-reads the registry and rescans every project.
//
// This is what the user reaches for when they have edited projects.yml by hand.
func (s *BoardService) Reload() []Problem {
	s.reload()
	emit(EventRegistryChanged, struct{}{})
	return s.Problems()
}

// reload rebuilds the store and the watcher from the registry on disk.
func (s *BoardService) reload() {
	reg, err := registry.LoadFrom(s.RegistryPath())

	var problems []Problem
	switch {
	case err != nil && isNoRegistry(err):
		problems = append(problems, Problem{
			Kind:   ProblemNoRegistry,
			Title:  "No projects yet",
			Detail: "Add a folder holding a Backlog.md project to get started.",
			Path:   reg.Path,
		})
	case err != nil:
		problems = append(problems, Problem{
			Kind:   ProblemRegistry,
			Title:  "The project registry could not be read",
			Detail: err.Error(),
			Path:   reg.Path,
		})
	}

	s.store.Load(reg)

	var projects []watcher.Project
	for _, p := range s.store.Projects() {
		if p.OK() {
			projects = append(projects, watcher.Project{
				Path: p.Registry.Path, DataDir: p.Registry.Location.DataDir,
			})
			continue
		}
		problems = append(problems, Problem{
			Kind:   ProblemProject,
			Title:  fmt.Sprintf("%s could not be loaded", p.Registry.DisplayName),
			Detail: p.Err.Error(),
			Path:   p.Registry.Path,
		})
	}
	for _, d := range s.store.Diagnostics() {
		problems = append(problems, Problem{
			Kind:   ProblemFile,
			Title:  "A file was skipped",
			Detail: d.Reason,
			Path:   d.Path,
		})
	}

	s.mu.Lock()
	s.problems = problems
	old := s.watch
	s.mu.Unlock()

	if old != nil {
		_ = old.Close()
	}

	w, err := watcher.New(projects, watcher.Options{
		OnChange: s.onProjectChanged,
		OnError:  func(error) {},
	})
	if err != nil {
		s.mu.Lock()
		s.problems = append(s.problems, Problem{
			Kind:   ProblemRegistry,
			Title:  "Changes on disk will not be picked up",
			Detail: err.Error(),
		})
		s.mu.Unlock()
		return
	}

	s.mu.Lock()
	s.watch = w
	s.mu.Unlock()
}

// onProjectChanged reloads one project and tells the frontend which.
func (s *BoardService) onProjectChanged(projectPath string) {
	if !s.store.Reload(projectPath) {
		return
	}
	emit(EventProjectChanged, ProjectChanged{Project: projectPath})
}

// emit publishes an event when an application is running.
//
// Guarded so the service works in tests and in a headless context, where there
// is no application to publish to.
func emit(name string, data any) {
	if app := application.Get(); app != nil {
		app.Event.Emit(name, data)
	}
}

// RegistryPath reports where the registry is read from, so the UI can name it.
func (s *BoardService) RegistryPath() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.registryPath == "" {
		return registry.DefaultPath()
	}
	return s.registryPath
}

// Problems returns everything currently wrong, worst first by kind.
func (s *BoardService) Problems() []Problem {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]Problem, len(s.problems))
	copy(out, s.problems)
	return out
}

// Projects returns every registered project, in registry order.
func (s *BoardService) Projects() []ProjectView {
	states := s.store.Projects()
	out := make([]ProjectView, 0, len(states))
	for _, p := range states {
		view := ProjectView{
			Path:   p.Registry.Path,
			Name:   p.Registry.DisplayName,
			Colour: p.Registry.Colour,
			OK:     p.OK(),
		}
		if p.Err != nil {
			view.Problem = p.Err.Error()
		}
		if p.OK() {
			view.TaskCount = len(p.Scanned.Tasks)
			view.DraftCount = len(p.Scanned.Drafts)
			view.Statuses = p.Scanned.Config.Statuses
			view.Priorities = p.Scanned.Config.Priorities
			view.Types = p.Scanned.Config.Types
			view.Layout = string(p.Registry.Location.Layout)
		}
		out = append(out, view)
	}
	return out
}

// Tasks returns every task matching a query, across every project.
func (s *BoardService) Tasks(q QueryInput) []TaskView {
	items := s.store.Query(q.toQuery())
	out := make([]TaskView, 0, len(items))
	for _, it := range items {
		out = append(out, TaskView{
			Project:       it.Ref.Project,
			ProjectName:   it.ProjectName,
			ProjectColour: it.ProjectColour,
			Kind:          it.Ref.Kind,
			Class:         it.Ref.Class,
			ID:            it.Ref.ID,
			Entity:        it.Entity,
		})
	}
	return out
}

// Task returns one entity, or ok false when the ref names nothing.
func (s *BoardService) Task(projectPath string, kind, class, id string) (TaskView, bool) {
	item, ok := s.store.Get(store.Ref{
		Project: projectPath,
		Key: backlog.Key{
			Kind: backlog.Kind(kind), Class: backlog.Class(class), ID: id,
		},
	})
	if !ok {
		return TaskView{}, false
	}
	return TaskView{
		Project:       item.Ref.Project,
		ProjectName:   item.ProjectName,
		ProjectColour: item.ProjectColour,
		Kind:          item.Ref.Kind,
		Class:         item.Ref.Class,
		ID:            item.Ref.ID,
		Entity:        item.Entity,
	}, true
}

// FilterValues returns the distinct values of one field across every project,
// for building filter menus.
func (s *BoardService) FilterValues(field string) []string {
	return s.store.Values(field)
}

// StatusLists returns each project's declared status list, in registry order.
//
// The board's columns are the union of these. Computing that union is the
// board's own concern, with its own ordering rules.
func (s *BoardService) StatusLists() [][]string {
	return s.store.StatusLists()
}

func (q QueryInput) toQuery() store.Query {
	out := store.Query{
		Projects:   q.Projects,
		Statuses:   q.Statuses,
		Milestones: q.Milestones,
		Priorities: q.Priorities,
		Types:      q.Types,
		Labels:     q.Labels,
		Assignees:  q.Assignees,
		Text:       q.Text,
	}
	for _, k := range q.Kinds {
		out.Kinds = append(out.Kinds, backlog.Kind(k))
	}
	for _, c := range q.Classes {
		out.Classes = append(out.Classes, backlog.Class(c))
	}
	return out
}

func isNoRegistry(err error) bool {
	return errors.Is(err, registry.ErrNoRegistry)
}

// ColumnView is one column of the unified board.
type ColumnView struct {
	Name string `json:"name"`
	// Projects are the registered projects that declare this status. A project
	// absent from the list simply has no cell in this column.
	Projects []string `json:"projects"`
}

// ConflictView records a pair of statuses the projects ordered differently, so
// the interface can explain an order that would otherwise look arbitrary.
type ConflictView struct {
	Before  string `json:"before"`
	After   string `json:"after"`
	Votes   int    `json:"votes"`
	Against int    `json:"against"`
}

// BoardLayout is the board's columns and how they were decided.
type BoardLayout struct {
	Columns   []ColumnView   `json:"columns"`
	Conflicts []ConflictView `json:"conflicts"`
}

// Layout returns the unified board columns.
//
// Statuses are per-project configuration and projects do not agree, so the
// columns are the union of every declared list. Muster never edits another
// project's config to make its own view simpler.
func (s *BoardService) Layout() BoardLayout {
	layout := board.Build(s.projectStatuses())

	out := BoardLayout{}
	for _, c := range layout.Columns {
		out.Columns = append(out.Columns, ColumnView{Name: c.Name, Projects: c.Projects})
	}
	for _, c := range layout.Conflicts {
		out.Conflicts = append(out.Conflicts, ConflictView{
			Before: c.Before, After: c.After, Votes: c.Votes, Against: c.Against,
		})
	}
	return out
}

// CanMove reports whether a task in a project may take a status.
//
// A task can only ever hold a status its own project declares; anything else
// would be a value the Backlog.md CLI itself rejects.
func (s *BoardService) CanMove(projectPath, status string) bool {
	return board.Build(s.projectStatuses()).CanMove(projectPath, status)
}

// projectStatuses collects what each loaded project declares, in registry order.
func (s *BoardService) projectStatuses() []board.ProjectStatuses {
	var out []board.ProjectStatuses
	for _, p := range s.store.Projects() {
		if p.OK() {
			out = append(out, board.ProjectStatuses{
				Project:  p.Registry.Path,
				Statuses: p.Scanned.Config.Statuses,
			})
		}
	}
	return out
}
