package app

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/wailsapp/wails/v3/pkg/application"

	"github.com/FMakareev/muster-backlog/internal/backlog"
	"github.com/FMakareev/muster-backlog/internal/backlogcli"
	"github.com/FMakareev/muster-backlog/internal/board"
	"github.com/FMakareev/muster-backlog/internal/registry"
	"github.com/FMakareev/muster-backlog/internal/settings"
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
	application.RegisterEvent[BulkProgress](EventBulkProgress)
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
	Path string `json:"path"`
	// DisplayPath is Path with a home directory written as ~, for showing.
	//
	// Separate from Path on purpose: Path is the identity every write, filter
	// and selection is keyed on, and abbreviating that would send a write to a
	// literal ~ no filesystem expands. This one is only ever printed - and it
	// is printed on a screen people screenshot, where the full form carries a
	// username for no reason.
	DisplayPath string `json:"displayPath"`
	Name        string `json:"name"`
	Colour      string `json:"colour"`
	OK          bool   `json:"ok"`
	Problem     string `json:"problem"`
	TaskCount   int    `json:"taskCount"`
	DraftCount  int    `json:"draftCount"`
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
	// Hidden keeps the project out of the board, the lists, search and the
	// figures without unregistering it. The Projects screen still shows it,
	// with everything it holds.
	Hidden bool `json:"hidden"`
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
	// Family is present only for a task that has a parent or subtasks, which
	// is 92 of the 712 in the author's projects. Nil for everything else, so
	// the relationship costs nothing on the cards that have none.
	Family *FamilyView `json:"family,omitempty"`
}

// EntityRef names one entity, in the shape the interface uses to open it.
type EntityRef struct {
	Project string        `json:"project"`
	Kind    backlog.Kind  `json:"kind"`
	Class   backlog.Class `json:"class"`
	ID      string        `json:"id"`
}

// FamilyView is what a card needs to show a parent or subtask relationship
// without asking for anything else.
//
// Counts rather than the subtasks themselves: the board asks for hundreds of
// cards at once and a card only ever shows how many there are. The panel asks
// for the list separately, once, when a task is opened.
type FamilyView struct {
	// Parent is where the parent lives, which is not always the same
	// directory the child is in. Nil when the task has no parent, and also
	// when it declares one that no file answers to - the declared id is on the
	// entity either way.
	Parent      *EntityRef `json:"parent,omitempty"`
	ParentTitle string     `json:"parentTitle,omitempty"`
	// Done and Total count subtasks. Archived subtasks are in neither, since
	// the board does not show them.
	Done  int `json:"done"`
	Total int `json:"total"`
}

// familyOf turns a resolved relationship into what the interface needs, or nil
// when there is no relationship to show.
func familyOf(kin store.Kin, ok bool) *FamilyView {
	if !ok || (kin.Parent == nil && len(kin.Children) == 0) {
		return nil
	}
	out := &FamilyView{
		ParentTitle: kin.ParentTitle,
		Done:        kin.Done,
		Total:       len(kin.Children),
	}
	if kin.Parent != nil {
		out.Parent = &EntityRef{
			Project: kin.Parent.Project,
			Kind:    kin.Parent.Kind,
			Class:   kin.Parent.Class,
			ID:      kin.Parent.ID,
		}
	}
	return out
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
	mu    sync.Mutex
	store *store.Store
	watch *watcher.Watcher
	// scanProblems come from reading the registry and the projects, and are
	// replaced wholesale on every reload.
	scanProblems []Problem
	// standingProblems are conditions of the machine rather than of the data -
	// a missing CLI, a desktop with no tray, unreadable preferences. They are
	// kept apart because a reload has nothing to say about them, and folding
	// them into one list meant a reload silently discarded them.
	standingProblems []Problem
	// cli is the only thing in the application that writes. It is resolved
	// once at startup so a missing binary is one report rather than the same
	// failure arriving on every action.
	cli    *backlogcli.Runner
	cliErr error
	// prefs are the application's own settings, separate from the registry.
	prefs settings.Settings
	// registryPath is where the registry is read from. It is a field rather
	// than a call to registry.DefaultPath() so that tests can point the service
	// at their own file instead of manipulating the environment - the XDG
	// lookup caches its variables at process start and cannot be moved later.
	registryPath string
	// settingsPath is where the preferences live. Injectable so that a test
	// cannot write into the developer's own configuration, which one did.
	settingsPath string
}

// NewBoardService returns a service reading the registry from its usual place.
func NewBoardService() *BoardService {
	return NewBoardServiceAt(registry.DefaultPath())
}

// NewBoardServiceAt returns a service reading the registry from a given path,
// and the preferences from where they normally live.
func NewBoardServiceAt(registryPath string) *BoardService {
	return NewBoardServiceWith(registryPath, settings.Path())
}

// NewBoardServiceWith returns a service reading both of its own files from
// given paths.
//
// The preferences path is injectable for the same reason the registry's is: a
// test that saves a preference must not write into the developer's own
// configuration. One did, and left an author in it.
func NewBoardServiceWith(registryPath, settingsPath string) *BoardService {
	return &BoardService{
		store:        store.New(),
		registryPath: registryPath,
		settingsPath: settingsPath,
	}
}

// SettingsPath is where the preferences are read from and written to.
func (s *BoardService) SettingsPath() string {
	if s.settingsPath != "" {
		return s.settingsPath
	}
	return settings.Path()
}

// ServiceName identifies the service in Wails logs.
func (s *BoardService) ServiceName() string { return "muster.board" }

// ServiceStartup loads everything and starts following the filesystem.
//
// A failure to load the registry is not a startup failure: the application
// opens, says what is wrong, and lets the user fix it. Returning an error here
// would leave them with a window that refuses to appear.
func (s *BoardService) ServiceStartup(_ context.Context, _ application.ServiceOptions) error {
	s.loadSettings()
	s.reload()
	s.resolveCLI()
	return nil
}

// loadSettings reads the preferences, keeping the defaults when the file is
// missing or unreadable rather than refusing to start.
func (s *BoardService) loadSettings() {
	prefs, err := settings.LoadFrom(s.SettingsPath())

	s.mu.Lock()
	s.prefs = prefs
	if err != nil {
		s.standingProblems = append(s.standingProblems, Problem{
			Kind:   ProblemRegistry,
			Title:  "Preferences could not be read",
			Detail: err.Error(),
			Path:   s.SettingsPath(),
		})
	}
	s.mu.Unlock()

	applyWindowBehaviour(prefs)

	if prefs.OnWindowClose == settings.BehaviourTray && TrayUnavailable() {
		s.mu.Lock()
		s.standingProblems = append(s.standingProblems, Problem{
			Kind:  ProblemRegistry,
			Title: "This desktop has no system tray",
			Detail: "Muster is set to stay in the tray, but nothing on the " +
				"session bus offers one, so the window behaves ordinarily.",
		})
		s.mu.Unlock()
	}
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
		if p.Registry.Hidden {
			// Hiding a project is saying you do not want to hear about it.
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
	s.scanProblems = problems
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
		s.scanProblems = append(s.scanProblems, Problem{
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

// RegistryDisplayPath is the same path, for showing rather than for using.
//
// It is a separate method on purpose. RegistryPath is what every write opens,
// so abbreviating it there would send registry.Add to a literal ~ that no
// filesystem expands - which is exactly what the first attempt at this did.
// This one is only ever printed: it sits in the status bar permanently, and
// under a home directory the full form puts a username into every screenshot
// anyone takes.
func (s *BoardService) RegistryDisplayPath() string {
	return registry.Abbreviate(s.RegistryPath())
}

// Problems returns everything currently wrong.
//
// Standing conditions come first: a missing CLI matters more than a stray file
// in a task directory, and it is the one a person has to act on.
func (s *BoardService) Problems() []Problem {
	s.mu.Lock()
	defer s.mu.Unlock()

	out := make([]Problem, 0, len(s.standingProblems)+len(s.scanProblems))
	out = append(out, s.standingProblems...)
	out = append(out, s.scanProblems...)
	return out
}

// Projects returns every registered project, in registry order.
func (s *BoardService) Projects() []ProjectView {
	states := s.store.Projects()
	out := make([]ProjectView, 0, len(states))
	for _, p := range states {
		view := ProjectView{
			Path:        p.Registry.Path,
			DisplayPath: registry.Abbreviate(p.Registry.Path),
			Name:        p.Registry.DisplayName,
			Colour:      p.Registry.Colour,
			OK:          p.OK(),
			Hidden:      p.Registry.Hidden,
		}
		if p.Err != nil {
			view.Problem = p.Err.Error()
		}
		if p.OK() {
			// Active tasks only. Counting archived and completed ones here
			// made the roll disagree with the board, the list and the
			// overview, all of which show what is live.
			for _, task := range p.Scanned.Tasks {
				if task.Class == backlog.ClassActive {
					view.TaskCount++
				}
			}
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
	// Resolved once for the whole answer rather than per card: the same slice
	// would otherwise be scanned once for every task on the board.
	kin := s.store.KinIndex()
	out := make([]TaskView, 0, len(items))
	for _, it := range items {
		k, ok := kin[it.Ref]
		out = append(out, taskView(it, familyOf(k, ok)))
	}
	return out
}

// Subtasks returns a task's subtasks, in the order the project scan produced.
//
// Separate from the card payload on purpose: this is asked once when a task is
// opened, where the board asks for every card at once.
func (s *BoardService) Subtasks(projectPath, kind, class, id string) []TaskView {
	ref := store.Ref{
		Project: projectPath,
		Key: backlog.Key{
			Kind: backlog.Kind(kind), Class: backlog.Class(class), ID: id,
		},
	}
	kin := s.store.KinIndex()
	out := make([]TaskView, 0, len(kin[ref].Children))
	for _, child := range kin[ref].Children {
		item, ok := s.store.Get(child)
		if !ok {
			continue
		}
		k, has := kin[child]
		out = append(out, taskView(item, familyOf(k, has)))
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
	kin := s.store.KinIndex()
	k, has := kin[item.Ref]
	return taskView(item, familyOf(k, has)), true
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

// MilestoneView is one milestone, with enough to show it by name and to see
// how far along it is.
type MilestoneView struct {
	Project     string `json:"project"`
	ProjectName string `json:"projectName"`
	ID          string `json:"id"`
	Title       string `json:"title"`
	// Total and Done count the tasks assigned to it in that project.
	Total int `json:"total"`
	Done  int `json:"done"`
}

// Milestones returns every milestone across every project, in registry order.
//
// A card carries its milestone as a bare id, which reads exactly like a task
// id and says nothing. This is what lets the interface show the title instead.
func (s *BoardService) Milestones() []MilestoneView {
	var out []MilestoneView

	for _, p := range s.store.Projects() {
		if !p.OK() {
			continue
		}
		// Terminal statuses are the last one a project declares, which is the
		// only definition of "finished" the format offers.
		var terminal string
		if statuses := p.Scanned.Config.Statuses; len(statuses) > 0 {
			terminal = statuses[len(statuses)-1]
		}

		counts := map[string][2]int{}
		for _, task := range p.Scanned.Tasks {
			if task.Class != backlog.ClassActive || task.Milestone == "" {
				continue
			}
			key := strings.ToLower(task.Milestone)
			c := counts[key]
			c[0]++
			if terminal != "" && strings.EqualFold(task.Status, terminal) {
				c[1]++
			}
			counts[key] = c
		}

		for _, m := range p.Scanned.Milestones {
			// Archived milestones are retired: off the board, out of the
			// grouping, and not offered as somewhere to put a task. They were
			// listed alongside the live ones until retiring one from the
			// application made it obvious.
			if m.Class != backlog.ClassActive {
				continue
			}
			c := counts[strings.ToLower(m.ID)]
			// A task may name a milestone by title rather than by id.
			if t := counts[strings.ToLower(m.Title)]; t[0] > c[0] {
				c = t
			}
			out = append(out, MilestoneView{
				Project:     p.Registry.Path,
				ProjectName: p.Registry.DisplayName,
				ID:          m.ID,
				Title:       m.Title,
				Total:       c[0],
				Done:        c[1],
			})
		}
	}
	return out
}

// Settings returns the application's own preferences.
func (s *BoardService) Settings() settings.Settings {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.prefs
}

// SaveSettings stores the preferences and applies what can be applied at once.
func (s *BoardService) SaveSettings(next settings.Settings) []Problem {
	if err := next.SaveTo(s.SettingsPath()); err != nil {
		return []Problem{{
			Kind:   ProblemRegistry,
			Title:  "Preferences could not be saved",
			Detail: err.Error(),
			Path:   s.SettingsPath(),
		}}
	}
	s.mu.Lock()
	changedCLI := s.prefs.BacklogPath != next.BacklogPath
	s.prefs = next
	s.mu.Unlock()

	// Saying where the CLI is has to take effect now: someone setting it is
	// looking at a banner telling them writes are unavailable, and asking them
	// to restart to find out whether they got it right is a poor answer.
	if changedCLI {
		s.resolveCLI()
		emit(EventRegistryChanged, struct{}{})
		if problem := s.cliProblem(); problem != nil {
			return []Problem{*problem}
		}
	}

	applyWindowBehaviour(next)

	// Asking for the tray on a desktop that has none has to say so. Silently
	// keeping ordinary behaviour would leave the setting looking as though it
	// worked, and silently honouring it would make the window vanish with no
	// way to get it back.
	if next.OnWindowClose == settings.BehaviourTray && TrayUnavailable() {
		return []Problem{{
			Kind:  ProblemRegistry,
			Title: "This desktop has no system tray",
			Detail: "Nothing on the session bus offers one, so the window will " +
				"keep behaving ordinarily. The preference is saved and takes " +
				"effect wherever a tray is available.",
		}}
	}
	return nil
}

// SearchHit is one search result as the frontend sees it.
type SearchHit struct {
	Project     string        `json:"project"`
	ProjectName string        `json:"projectName"`
	Kind        backlog.Kind  `json:"kind"`
	Class       backlog.Class `json:"class"`
	ID          string        `json:"id"`
	Title       string        `json:"title"`
	// Field says where the match was: title, id or body.
	Field string `json:"field"`
	// Excerpt is the matching text with a little around it.
	Excerpt string `json:"excerpt"`
}

// Search looks through every entity of every project.
func (s *BoardService) Search(text string, limit int) []SearchHit {
	hits := s.store.Search(text, limit)

	out := make([]SearchHit, 0, len(hits))
	for _, hit := range hits {
		out = append(out, SearchHit{
			Project:     hit.Item.Ref.Project,
			ProjectName: hit.Item.ProjectName,
			Kind:        hit.Item.Ref.Kind,
			Class:       hit.Item.Ref.Class,
			ID:          hit.Item.Ref.ID,
			Title:       hit.Item.Entity.Title,
			Field:       hit.Field,
			Excerpt:     hit.Excerpt,
		})
	}
	return out
}

// Entities returns every document, decision, draft or milestone across every
// project, so the viewer can list them without asking for tasks it will drop.
func (s *BoardService) Entities(kind string) []TaskView {
	items := s.store.Entities(backlog.Kind(kind))

	out := make([]TaskView, 0, len(items))
	for _, it := range items {
		// Documents and milestones have no parent or subtasks to resolve.
		out = append(out, taskView(it, nil))
	}
	return out
}

// CountView is one label and how many things carry it.
type CountView struct {
	Label string `json:"label"`
	Total int    `json:"total"`
}

// BlockedView is a task waiting on something unfinished.
type BlockedView struct {
	Task TaskView `json:"task"`
	On   []string `json:"on"`
}

// AnalyticsView is the overview for one project, or for all of them.
type AnalyticsView struct {
	// Project is empty on the entry that covers every project.
	Project        string        `json:"project"`
	ProjectName    string        `json:"projectName"`
	Tasks          int           `json:"tasks"`
	Statuses       []CountView   `json:"statuses"`
	Priority       []CountView   `json:"priority"`
	Types          []CountView   `json:"types"`
	Unprioritised  int           `json:"unprioritised"`
	AverageAgeDays float64       `json:"averageAgeDays"`
	Stale          []TaskView    `json:"stale"`
	Blocked        []BlockedView `json:"blocked"`
}

// Analytics returns the cross-project overview.
func (s *BoardService) Analytics() []AnalyticsView {
	s.mu.Lock()
	days := s.prefs.StaleAfterDays
	s.mu.Unlock()
	if days <= 0 {
		days = 30
	}

	reports := s.store.Analytics(store.AnalyticsOptions{
		StaleAfter: time.Duration(days) * 24 * time.Hour,
	})

	out := make([]AnalyticsView, 0, len(reports))
	for _, r := range reports {
		view := AnalyticsView{
			Project:        r.Project,
			ProjectName:    r.ProjectName,
			Tasks:          r.Tasks,
			Statuses:       counts(r.Statuses),
			Priority:       counts(r.Priority),
			Types:          counts(r.Types),
			Unprioritised:  r.Unprioritised,
			AverageAgeDays: r.AverageAgeDays,
		}
		for _, item := range r.Stale {
			view.Stale = append(view.Stale, taskView(item, nil))
		}
		for _, b := range r.Blocked {
			view.Blocked = append(view.Blocked, BlockedView{
				Task: taskView(b.Item, nil), On: b.On,
			})
		}
		out = append(out, view)
	}
	return out
}

// WIPStatus is one column's load against its advisory limit.
type WIPStatus struct {
	Project     string `json:"project"`
	ProjectName string `json:"projectName"`
	Status      string `json:"status"`
	Count       int    `json:"count"`
	Limit       int    `json:"limit"`
	Over        bool   `json:"over"`
}

// WIP reports where a project is at or over an advisory limit.
//
// It is a signal, never a rule: a limit that blocks a drag is a limit people
// work around rather than one they act on.
func (s *BoardService) WIP() []WIPStatus {
	s.mu.Lock()
	limits := s.prefs.WIPLimits
	s.mu.Unlock()
	if len(limits) == 0 {
		return nil
	}

	var out []WIPStatus
	for _, p := range s.store.Projects() {
		if !p.OK() {
			continue
		}
		counts := s.store.CountByStatus(p.Registry.Path)
		for status, limit := range limits {
			count := 0
			for name, n := range counts {
				if strings.EqualFold(name, status) {
					count += n
				}
			}
			if count == 0 {
				continue
			}
			out = append(out, WIPStatus{
				Project:     p.Registry.Path,
				ProjectName: p.Registry.DisplayName,
				Status:      status,
				Count:       count,
				Limit:       limit,
				Over:        count >= limit,
			})
		}
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].ProjectName != out[j].ProjectName {
			return out[i].ProjectName < out[j].ProjectName
		}
		return out[i].Status < out[j].Status
	})
	return out
}

func counts(in []store.Count) []CountView {
	out := make([]CountView, 0, len(in))
	for _, c := range in {
		out = append(out, CountView{Label: c.Label, Total: c.Total})
	}
	return out
}

// taskView assembles one task for the frontend. Family is nil wherever the
// relationship has not been resolved, which is every list that is not tasks.
func taskView(it store.Item, family *FamilyView) TaskView {
	return TaskView{
		Project:       it.Ref.Project,
		ProjectName:   it.ProjectName,
		ProjectColour: it.ProjectColour,
		Kind:          it.Ref.Kind,
		Class:         it.Ref.Class,
		ID:            it.Ref.ID,
		Entity:        it.Entity,
		Family:        family,
	}
}
