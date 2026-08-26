// Package store holds every registered project in memory and answers queries
// across all of them at once.
//
// A thousand entity files parse in about a tenth of a second, so the whole
// corpus lives in memory and no database is warranted. The store is the single
// place that knows about more than one project; everything below it deals with
// one project at a time.
package store

import (
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"github.com/FMakareev/muster-backlog/internal/backlog"
	"github.com/FMakareev/muster-backlog/internal/registry"
)

// Ref identifies an entity across every registered project.
//
// Backlog.md ids are unique only within one directory of one project: 200 of
// the 351 distinct task ids in the author's projects appear in more than one,
// and archiving is a soft delete that lets an id name two different tasks
// inside a single project. So identity is the project plus the key, never the
// id, and there is no such thing as "task TASK-1" without saying where.
type Ref struct {
	// Project is the registered project path, which is unique by construction:
	// the registry rejects a second entry for the same folder.
	Project string
	backlog.Key
}

// String renders a ref for logs and diagnostics.
func (r Ref) String() string {
	return fmt.Sprintf("%s:%s/%s/%s", r.Project, r.Kind, r.Class, r.ID)
}

// Item is one entity together with where it came from.
type Item struct {
	Ref Ref
	// ProjectName is the display name from the registry, which is what a
	// person calls the project - not the folder name and not project_name.
	ProjectName string
	// ProjectColour is the registry's colour, empty when unset.
	ProjectColour string
	Entity        backlog.Entity
}

// ProjectState is one project as currently loaded.
type ProjectState struct {
	Registry registry.Project
	// Scanned is nil when the project failed to load; Err then says why.
	Scanned *backlog.Project
	Err     error
	// LoadedAt is when this project was last scanned.
	LoadedAt time.Time
	// LoadDuration is how long that scan took.
	LoadDuration time.Duration
}

// OK reports whether the project is loaded and queryable.
func (p ProjectState) OK() bool { return p.Err == nil && p.Scanned != nil }

// Visible reports whether this project takes part in the board, the lists,
// search and the figures.
//
// A hidden project is loaded but left out of all of them. Loading it anyway is
// deliberate: the Projects screen still shows what it holds, and unhiding is
// then instant rather than a rescan. Hiding is a display choice, not an
// unregistering - which is why it is a field on the entry and not a removal.
func (p ProjectState) Visible() bool { return p.OK() && !p.Registry.Hidden }

// Store is the aggregated view. It is safe for concurrent use: the watcher
// reloads projects while the UI reads.
type Store struct {
	mu sync.RWMutex
	// order preserves the registry's ordering, which is the order the user
	// sees. Maps alone would lose it.
	order    []string
	projects map[string]*ProjectState
}

// New returns an empty store.
func New() *Store {
	return &Store{projects: map[string]*ProjectState{}}
}

// Load scans every project in a registry, replacing whatever the store held.
//
// A project that fails to scan is kept as a failed ProjectState rather than
// dropped: a project silently missing from the board is worse than one that
// says why it is broken.
func (s *Store) Load(reg registry.Registry) {
	states := make(map[string]*ProjectState, len(reg.Projects))
	order := make([]string, 0, len(reg.Projects))

	for _, p := range reg.Projects {
		order = append(order, p.Path)
		states[p.Path] = scanOne(p)
	}

	s.mu.Lock()
	s.order, s.projects = order, states
	s.mu.Unlock()
}

// Reload rescans exactly one project.
//
// This is what a file change costs: one project, not the corpus. Reloading a
// project that is not registered is a no-op and reports false.
func (s *Store) Reload(path string) bool {
	s.mu.RLock()
	state, known := s.projects[path]
	s.mu.RUnlock()
	if !known {
		return false
	}

	// Scanning happens outside the lock so a slow disk never blocks readers.
	fresh := scanOne(state.Registry)

	s.mu.Lock()
	defer s.mu.Unlock()
	if _, still := s.projects[path]; !still {
		return false
	}
	s.projects[path] = fresh
	return true
}

func scanOne(p registry.Project) *ProjectState {
	state := &ProjectState{Registry: p, LoadedAt: time.Now()}
	if !p.OK() {
		state.Err = p.Err
		return state
	}
	start := time.Now()
	scanned, err := backlog.Scan(p.Location)
	state.LoadDuration = time.Since(start)
	if err != nil {
		state.Err = err
		return state
	}
	state.Scanned = scanned
	return state
}

// Projects returns the current state of every registered project, in registry
// order.
func (s *Store) Projects() []ProjectState {
	s.mu.RLock()
	defer s.mu.RUnlock()

	out := make([]ProjectState, 0, len(s.order))
	for _, path := range s.order {
		if state, ok := s.projects[path]; ok {
			out = append(out, *state)
		}
	}
	return out
}

// StatusLists returns each loaded project's declared status list, in registry
// order.
//
// The board's columns are the union of these, but computing that union is a
// separate concern with its own ordering rules; this only hands over the raw
// material.
func (s *Store) StatusLists() [][]string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var out [][]string
	for _, path := range s.order {
		if state, ok := s.projects[path]; ok && state.Visible() {
			out = append(out, state.Scanned.Config.Statuses)
		}
	}
	return out
}

// Diagnostics returns every file skipped during a scan.
//
// A project that failed to load entirely is not a diagnostic here: it is
// already reported on its own ProjectState, and returning it in both places
// makes one broken folder look like two problems, the second of them
// mislabelled as a skipped file.
func (s *Store) Diagnostics() []backlog.Diagnostic {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var out []backlog.Diagnostic
	for _, path := range s.order {
		state, ok := s.projects[path]
		if !ok || !state.Visible() {
			continue
		}
		out = append(out, state.Scanned.Diagnostics...)
	}
	return out
}

// Get returns one entity by ref.
func (s *Store) Get(ref Ref) (Item, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	state, ok := s.projects[ref.Project]
	if !ok || !state.Visible() {
		return Item{}, false
	}
	for _, e := range entitiesOfKind(state.Scanned, ref.Kind) {
		if e.Key == ref.Key {
			return item(state, e), true
		}
	}
	return Item{}, false
}

// Query is a set of optional filters, combined with AND. A zero Query matches
// every task in every project.
//
// Values are compared case-insensitively throughout, because enum case is not
// normalised on disk: priority is written lowercase against a capitalised
// config list, and decision status appears in both cases in the same corpus.
type Query struct {
	// Kinds defaults to tasks alone when empty, since that is what a board and
	// a list view are made of.
	Kinds   []backlog.Kind
	Classes []backlog.Class

	Projects   []string
	Statuses   []string
	Milestones []string
	Priorities []string
	Types      []string
	Labels     []string
	Assignees  []string

	// Text matches the title and the description, case-insensitively.
	Text string
}

// Query returns every matching item, ordered by project registry order and
// then by the project's own ordering.
func (s *Store) Query(q Query) []Item {
	s.mu.RLock()
	defer s.mu.RUnlock()

	kinds := q.Kinds
	if len(kinds) == 0 {
		kinds = []backlog.Kind{backlog.KindTask}
	}

	var out []Item
	for _, path := range s.order {
		state, ok := s.projects[path]
		if !ok || !state.Visible() {
			continue
		}
		if len(q.Projects) > 0 && !matchesAny(path, q.Projects) &&
			!matchesAny(state.Registry.DisplayName, q.Projects) {
			continue
		}
		for _, kind := range kinds {
			for _, e := range entitiesOfKind(state.Scanned, kind) {
				if q.matches(e) {
					out = append(out, item(state, e))
				}
			}
		}
	}
	return out
}

// Count returns how many items a query matches, without building the slice.
func (s *Store) Count(q Query) int {
	return len(s.Query(q))
}

// CountByStatus returns per-status counts for one project, which is what a
// work-in-progress limit is measured against.
func (s *Store) CountByStatus(projectPath string) map[string]int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	counts := map[string]int{}
	state, ok := s.projects[projectPath]
	if !ok || !state.Visible() {
		return counts
	}
	for _, e := range state.Scanned.Tasks {
		if e.Class == backlog.ClassActive {
			counts[e.Status]++
		}
	}
	return counts
}

func (q Query) matches(e backlog.Entity) bool {
	if len(q.Classes) > 0 && !matchesAnyClass(e.Class, q.Classes) {
		return false
	}
	if len(q.Statuses) > 0 && !matchesAny(e.Status, q.Statuses) {
		return false
	}
	if len(q.Milestones) > 0 && !matchesAny(e.Milestone, q.Milestones) {
		return false
	}
	if len(q.Priorities) > 0 && !matchesAny(e.Priority, q.Priorities) {
		return false
	}
	if len(q.Types) > 0 && !matchesAny(e.Type, q.Types) {
		return false
	}
	// A label filter matches when the entity carries any of the wanted labels.
	if len(q.Labels) > 0 && !intersects(e.Labels, q.Labels) {
		return false
	}
	if len(q.Assignees) > 0 && !intersects(e.Assignee, q.Assignees) {
		return false
	}
	if q.Text != "" {
		needle := strings.ToLower(q.Text)
		description, _ := e.Section(backlog.SectionDescription)
		if !strings.Contains(strings.ToLower(e.Title), needle) &&
			!strings.Contains(strings.ToLower(description), needle) {
			return false
		}
	}
	return true
}

func item(state *ProjectState, e backlog.Entity) Item {
	return Item{
		Ref:           Ref{Project: state.Registry.Path, Key: e.Key},
		ProjectName:   state.Registry.DisplayName,
		ProjectColour: state.Registry.Colour,
		Entity:        e,
	}
}

func entitiesOfKind(p *backlog.Project, kind backlog.Kind) []backlog.Entity {
	switch kind {
	case backlog.KindTask:
		return p.Tasks
	case backlog.KindDraft:
		return p.Drafts
	case backlog.KindMilestone:
		return p.Milestones
	case backlog.KindDocument:
		return p.Documents
	case backlog.KindDecision:
		return p.Decisions
	}
	return nil
}

func matchesAny(value string, wanted []string) bool {
	for _, w := range wanted {
		if strings.EqualFold(value, w) {
			return true
		}
	}
	return false
}

func matchesAnyClass(value backlog.Class, wanted []backlog.Class) bool {
	for _, w := range wanted {
		if value == w {
			return true
		}
	}
	return false
}

func intersects(have, wanted []string) bool {
	for _, h := range have {
		if matchesAny(h, wanted) {
			return true
		}
	}
	return false
}

// Values returns the distinct values of one field across every loaded project,
// sorted, for building filter menus.
func (s *Store) Values(field string) []string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	seen := map[string]bool{}
	for _, path := range s.order {
		state, ok := s.projects[path]
		if !ok || !state.Visible() {
			continue
		}
		for _, e := range state.Scanned.Tasks {
			switch field {
			case "status":
				add(seen, e.Status)
			case "priority":
				add(seen, e.Priority)
			case "type":
				add(seen, e.Type)
			case "milestone":
				add(seen, e.Milestone)
			case "label":
				for _, l := range e.Labels {
					add(seen, l)
				}
			case "assignee":
				for _, a := range e.Assignee {
					add(seen, a)
				}
			}
		}
	}

	out := make([]string, 0, len(seen))
	for v := range seen {
		out = append(out, v)
	}
	sort.Strings(out)
	return out
}

func add(seen map[string]bool, value string) {
	if v := strings.TrimSpace(value); v != "" {
		seen[v] = true
	}
}

// Hit is one search result.
type Hit struct {
	Item Item
	// Field says where the match was found, so a result can explain itself
	// rather than looking arbitrary when the title does not contain the words.
	Field string
	// Excerpt is the matching text with a little around it.
	Excerpt string
}

// Search looks through every entity of every project.
//
// Titles rank above bodies, because someone typing a few words is far more
// often reaching for a task they remember by name than for a phrase buried in
// a description.
func (s *Store) Search(text string, limit int) []Hit {
	needle := strings.ToLower(strings.TrimSpace(text))
	if needle == "" {
		return nil
	}
	if limit <= 0 {
		limit = 100
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	var titles, bodies []Hit
	kinds := []backlog.Kind{
		backlog.KindTask, backlog.KindDraft, backlog.KindDocument,
		backlog.KindDecision, backlog.KindMilestone,
	}

	for _, path := range s.order {
		state, ok := s.projects[path]
		if !ok || !state.Visible() {
			continue
		}
		for _, kind := range kinds {
			for _, e := range entitiesOfKind(state.Scanned, kind) {
				if strings.Contains(strings.ToLower(e.Title), needle) {
					titles = append(titles, Hit{
						Item: item(state, e), Field: "title", Excerpt: e.Title,
					})
					continue
				}
				if excerpt, found := excerptAround(e.Body, needle); found {
					bodies = append(bodies, Hit{
						Item: item(state, e), Field: "body", Excerpt: excerpt,
					})
					continue
				}
				if strings.Contains(strings.ToLower(e.ID), needle) {
					titles = append(titles, Hit{
						Item: item(state, e), Field: "id", Excerpt: e.ID,
					})
				}
			}
		}
	}

	out := append(titles, bodies...)
	if len(out) > limit {
		out = out[:limit]
	}
	return out
}

// excerptAround returns the matching text with some context, so a body hit can
// show why it matched.
func excerptAround(body, needle string) (string, bool) {
	at := strings.Index(strings.ToLower(body), needle)
	if at < 0 {
		return "", false
	}

	const context = 60
	start := max(at-context, 0)
	end := min(at+len(needle)+context, len(body))

	// Do not cut a rune in half.
	for start > 0 && !utf8.RuneStart(body[start]) {
		start--
	}
	for end < len(body) && !utf8.RuneStart(body[end]) {
		end++
	}

	excerpt := strings.Join(strings.Fields(body[start:end]), " ")
	if start > 0 {
		excerpt = "…" + excerpt
	}
	if end < len(body) {
		excerpt += "…"
	}
	return excerpt, true
}

// Entities returns every entity of one kind, across every project.
func (s *Store) Entities(kind backlog.Kind) []Item {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var out []Item
	for _, path := range s.order {
		state, ok := s.projects[path]
		if !ok || !state.Visible() {
			continue
		}
		for _, e := range entitiesOfKind(state.Scanned, kind) {
			out = append(out, item(state, e))
		}
	}
	return out
}
