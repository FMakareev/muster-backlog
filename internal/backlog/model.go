// Package backlog reads Backlog.md projects from disk.
//
// The model here carries only what Backlog.md itself defines. Muster adds no
// field, no label convention and no sidecar file of its own, so anything not
// found in a Backlog.md file has no place in these types.
//
// Everything the parser does is driven by the measured format contract in
// backlog/docs/doc-3, not by assumption. Where a rule looks arbitrary, it is
// because the corpus says so.
package backlog

import "time"

// Kind is the sort of entity a file holds.
type Kind string

// The entity kinds Backlog.md defines.
const (
	KindTask      Kind = "task"
	KindDraft     Kind = "draft"
	KindMilestone Kind = "milestone"
	KindDocument  Kind = "document"
	KindDecision  Kind = "decision"
)

// Class is the directory an entity was found in.
//
// It is part of an entity's identity, not decoration: ids are unique only
// within one directory of one project. Archiving is a soft delete and the CLI
// reuses archived ids, so the same id can name two different tasks in tasks/
// and archive/tasks/ at once. That happens in the author's own projects today.
type Class string

// The directory classes an entity can be found in.
const (
	ClassActive    Class = "active"
	ClassCompleted Class = "completed"
	ClassArchived  Class = "archived"
)

// Key identifies an entity within one project.
type Key struct {
	Kind  Kind
	Class Class
	ID    string
}

// Section is one delimited region of a task body.
type Section string

// The delimited regions a task body can carry.
const (
	SectionDescription  Section = "description"
	SectionPlan         Section = "plan"
	SectionNotes        Section = "notes"
	SectionFinalSummary Section = "final_summary"
)

// Criterion is one acceptance-criteria or definition-of-done item.
type Criterion struct {
	// Index is the 1-based handle the CLI uses for --check-ac. It is
	// renumbered on insertion and deletion, so it identifies an item only
	// within one revision of one file and must never be persisted as identity.
	Index int
	// Checked is the [x] state.
	Checked bool
	// Text is everything after the marker, verbatim.
	Text string
}

// Comment is one entry from a task's comment envelope.
type Comment struct {
	Author string
	// Created is the raw timestamp. Unlike every frontmatter date it is not
	// quoted in the file, and it is kept as written rather than parsed, since
	// nothing depends on its type.
	Created string
	Body    string
}

// Entity is any Backlog.md file: a task, draft, milestone, document or decision.
type Entity struct {
	Key
	// Path is the file this was read from.
	Path string

	// Title always comes from frontmatter. Filenames are a lossy, one-way
	// derivation - only a fifth of them survive the round trip, and five in the
	// author's projects carry no title information at all.
	Title string

	Status   string
	Priority string
	Type     string

	Assignee []string
	Labels   []string

	// Milestone is an id or a title, depending on how it was written.
	Milestone string

	// Dependencies, ParentTaskID and Subtasks reference other entities. Every
	// such reference resolves inside its own project only: ids collide across
	// projects freely.
	Dependencies []string
	ParentTaskID string
	Subtasks     []string

	// Ordinal holds manual ordering. It is optional and not unique.
	Ordinal    *int
	HasOrdinal bool

	Created time.Time
	Updated time.Time
	// Date is a decision's own date field.
	Date time.Time

	References    []string
	Documentation []string
	ModifiedFiles []string
	Tags          []string

	// Reporter and OnStatusChange are in the 1.48.0 serialiser but appear in no
	// file the corpus contains. Read, never required.
	Reporter       string
	OnStatusChange string

	// Sections holds the delimited regions of a task body. Absent sections are
	// absent from the map; a present but empty section maps to "".
	Sections map[Section]string

	// AcceptanceCriteria and DefinitionOfDone are parsed items. An empty,
	// present section yields a non-nil empty slice, because "no criteria" and
	// "no section" are different states.
	AcceptanceCriteria []Criterion
	DefinitionOfDone   []Criterion

	Comments []Comment

	// Body is everything after the frontmatter, verbatim and CRLF-normalised.
	// Non-task entities have no marker grammar at all, so for them this is the
	// whole of their content.
	Body string
}

// Section returns a body section and whether it was present.
func (e Entity) Section(s Section) (string, bool) {
	v, ok := e.Sections[s]
	return v, ok
}

// Diagnostic reports a file the parser could not treat as an entity.
//
// A diagnostic never stops a scan. A stray README, an editor backup or a
// .gitkeep in a task directory is normal, and one unreadable file must not cost
// the user the other nine hundred.
type Diagnostic struct {
	Path   string
	Reason string
}

// Config is a project's own config.yml.
//
// These values belong to the project and are read from it every time. Muster
// never copies them into its own registry, where they would go stale.
type Config struct {
	ProjectName   string
	DefaultStatus string
	// Statuses is an ordered list of arbitrary strings. The board's columns are
	// the union of these lists across projects, so order matters here.
	Statuses   []string
	Labels     []string
	Types      []string
	Priorities []string
	// TaskPrefix is configurable per project, so "task-" must never be assumed.
	TaskPrefix       string
	DefinitionOfDone []string
	ZeroPaddedIDs    int
	HideEmptyColumns bool
	// BacklogDirectory appears only in a root config.
	BacklogDirectory string
}

// DefaultConfig mirrors Backlog.md's own defaults, used for keys a project does
// not set.
func DefaultConfig() Config {
	return Config{
		DefaultStatus: "To Do",
		Statuses:      []string{"To Do", "In Progress", "Done"},
		Types: []string{
			"bug", "feature", "enhancement", "task", "chore", "docs", "spike",
		},
		Priorities: []string{"High", "Medium", "Low"},
		TaskPrefix: "task",
	}
}
