package backlog

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/goccy/go-yaml"
)

// frontmatterFence delimits frontmatter. It is only ever recognised at the very
// start of a file: seven task files in the author's projects contain bare ---
// lines inside comment envelopes, so a reader that scans the whole file for a
// fence takes half a comment for frontmatter.
const frontmatterFence = "---"

// sectionMarkers are the seven literal marker pairs, exactly as the 1.48.0
// binary writes them. Two naming conventions coexist and there is no rule to
// infer - the long SECTION:<ID>: prefix and the short bare one are simply both
// real, so both are hardcoded.
var sectionMarkers = []struct {
	Section    Section
	Begin, End string
}{
	{SectionDescription, "<!-- SECTION:DESCRIPTION:BEGIN -->", "<!-- SECTION:DESCRIPTION:END -->"},
	{SectionPlan, "<!-- SECTION:PLAN:BEGIN -->", "<!-- SECTION:PLAN:END -->"},
	{SectionNotes, "<!-- SECTION:NOTES:BEGIN -->", "<!-- SECTION:NOTES:END -->"},
	{SectionFinalSummary, "<!-- SECTION:FINAL_SUMMARY:BEGIN -->", "<!-- SECTION:FINAL_SUMMARY:END -->"},
}

const (
	acBegin  = "<!-- AC:BEGIN -->"
	acEnd    = "<!-- AC:END -->"
	dodBegin = "<!-- DOD:BEGIN -->"
	dodEnd   = "<!-- DOD:END -->"

	commentsBegin = "<!-- COMMENTS:BEGIN -->"
	commentsEnd   = "<!-- COMMENTS:END -->"
	// The singular form is a newer per-comment envelope defined by 1.48.0 that
	// appears in no file yet. Tolerated so it does not become a surprise.
	commentBegin = "<!-- COMMENT:BEGIN -->"
	commentEnd   = "<!-- COMMENT:END -->"
)

// criterionWithIndex is the grammar every one of the 3367 checked items in the
// corpus conforms to, without exception.
var criterionWithIndex = regexp.MustCompile(`^- \[( |x)\] #(\d+) (.+)$`)

// criterionWithoutIndex is the legacy form 1.48.0 still ships a migration for.
var criterionWithoutIndex = regexp.MustCompile(`^- \[( |x)\] (.+)$`)

// frontmatterFields mirrors the 1.48.0 serialiser. Every field is optional
// except id and title; three of them appear in no file yet and are read anyway.
type frontmatterFields struct {
	ID             string   `yaml:"id"`
	Title          string   `yaml:"title"`
	Status         string   `yaml:"status"`
	Assignee       []string `yaml:"assignee"`
	Reporter       string   `yaml:"reporter"`
	CreatedDate    string   `yaml:"created_date"`
	UpdatedDate    string   `yaml:"updated_date"`
	Date           string   `yaml:"date"`
	Labels         []string `yaml:"labels"`
	Tags           []string `yaml:"tags"`
	Milestone      string   `yaml:"milestone"`
	Dependencies   []string `yaml:"dependencies"`
	References     []string `yaml:"references"`
	Documentation  []string `yaml:"documentation"`
	ModifiedFiles  []string `yaml:"modified_files"`
	ParentTaskID   string   `yaml:"parent_task_id"`
	Subtasks       []string `yaml:"subtasks"`
	Priority       string   `yaml:"priority"`
	Type           string   `yaml:"type"`
	Ordinal        *int     `yaml:"ordinal"`
	OnStatusChange string   `yaml:"onStatusChange"`
}

// ParseFile turns one file's bytes into an entity.
//
// kind and class come from the directory the file was found in; nothing about
// them is inferred from the content. hasMarkers says whether the task section
// grammar applies: it must not be applied to documents, which are free markdown
// and can quote the markers in prose. One document in this very repository does.
func ParseFile(path string, raw []byte, kind Kind, class Class, hasMarkers bool) (Entity, error) {
	text := normaliseNewlines(string(raw))

	front, body, err := splitFrontmatter(text)
	if err != nil {
		return Entity{}, err
	}

	var fields frontmatterFields
	if err := yaml.Unmarshal([]byte(front), &fields); err != nil {
		return Entity{}, fmt.Errorf("frontmatter is not valid YAML: %w", err)
	}
	if strings.TrimSpace(fields.ID) == "" {
		return Entity{}, fmt.Errorf("frontmatter has no id")
	}

	e := Entity{
		Key:  Key{Kind: kind, Class: class, ID: strings.TrimSpace(fields.ID)},
		Path: path,
		// The title is taken from frontmatter and never from the filename,
		// which is a lossy one-way slug.
		Title:          strings.TrimSpace(fields.Title),
		Status:         fields.Status,
		Priority:       fields.Priority,
		Type:           fields.Type,
		Assignee:       fields.Assignee,
		Labels:         fields.Labels,
		Tags:           fields.Tags,
		Milestone:      fields.Milestone,
		Dependencies:   fields.Dependencies,
		ParentTaskID:   fields.ParentTaskID,
		Subtasks:       fields.Subtasks,
		References:     fields.References,
		Documentation:  fields.Documentation,
		ModifiedFiles:  fields.ModifiedFiles,
		Reporter:       fields.Reporter,
		OnStatusChange: fields.OnStatusChange,
		Ordinal:        fields.Ordinal,
		HasOrdinal:     fields.Ordinal != nil,
		Created:        parseTimestamp(fields.CreatedDate),
		Updated:        parseTimestamp(fields.UpdatedDate),
		Date:           parseTimestamp(fields.Date),
		Body:           body,
		Sections:       map[Section]string{},
	}

	if hasMarkers {
		parseSections(&e, body)
	}
	return e, nil
}

// splitFrontmatter takes the first fenced block, anchored at the start of the
// file, and returns it with the remaining body.
func splitFrontmatter(text string) (front, body string, err error) {
	rest, ok := strings.CutPrefix(text, frontmatterFence+"\n")
	if !ok {
		return "", "", fmt.Errorf("file does not begin with frontmatter")
	}
	end := strings.Index(rest, "\n"+frontmatterFence)
	if end < 0 {
		return "", "", fmt.Errorf("frontmatter is not closed")
	}
	front = rest[:end]
	body = rest[end+len("\n"+frontmatterFence):]
	body = strings.TrimPrefix(body, "\n")
	return front, body, nil
}

// parseSections fills the delimited regions of a task body.
//
// Matching is on the markers alone. The whitespace between a heading and its
// marker is not uniform - four section kinds always have a blank line and two
// never do - so anything measured from the heading is wrong.
func parseSections(e *Entity, body string) {
	for _, m := range sectionMarkers {
		if content, ok := between(body, m.Begin, m.End); ok {
			e.Sections[m.Section] = strings.TrimSpace(content)
		}
	}
	if content, ok := between(body, acBegin, acEnd); ok {
		e.AcceptanceCriteria = parseCriteria(content)
	}
	if content, ok := between(body, dodBegin, dodEnd); ok {
		e.DefinitionOfDone = parseCriteria(content)
	}
	if content, ok := between(body, commentsBegin, commentsEnd); ok {
		e.Comments = parseComments(content)
	}
}

// between returns the text delimited by a marker pair, and whether the pair was
// present. An empty region is present-and-empty, which is a different state
// from absent: one task in the corpus has an AC block with nothing in it.
func between(text, begin, end string) (string, bool) {
	start := strings.Index(text, begin)
	if start < 0 {
		return "", false
	}
	start += len(begin)
	stop := strings.Index(text[start:], end)
	if stop < 0 {
		return "", false
	}
	return text[start : start+stop], true
}

// parseCriteria reads acceptance-criteria or definition-of-done items; one
// grammar serves both.
//
// It returns a non-nil empty slice for a present-but-empty section, so callers
// can tell "no criteria" from "no section".
func parseCriteria(content string) []Criterion {
	items := []Criterion{}
	for _, line := range strings.Split(content, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if m := criterionWithIndex.FindStringSubmatch(line); m != nil {
			index, err := strconv.Atoi(m[2])
			if err != nil {
				continue
			}
			items = append(items, Criterion{
				Index: index, Checked: m[1] == "x", Text: m[3],
			})
			continue
		}
		// The legacy form carries no index. 1.48.0 still migrates these, so
		// they can appear; number them positionally.
		if m := criterionWithoutIndex.FindStringSubmatch(line); m != nil {
			items = append(items, Criterion{
				Index: len(items) + 1, Checked: m[1] == "x", Text: m[2],
			})
		}
	}
	return items
}

// parseComments reads the comment envelope.
//
// Each comment is author / created / --- / body / ---. Those bare fences are
// the reason frontmatter must be anchored at byte 0.
func parseComments(content string) []Comment {
	// Tolerate the newer per-comment marker form, which no file uses yet, by
	// removing the markers and letting the shared envelope grammar run.
	content = strings.ReplaceAll(content, commentBegin, "")
	content = strings.ReplaceAll(content, commentEnd, "")

	var comments []Comment
	var current *Comment
	var body []string
	inBody := false

	flush := func() {
		if current == nil {
			return
		}
		current.Body = strings.TrimSpace(strings.Join(body, "\n"))
		comments = append(comments, *current)
		current, body, inBody = nil, nil, false
	}

	for _, line := range strings.Split(content, "\n") {
		trimmed := strings.TrimSpace(line)
		switch {
		case strings.HasPrefix(trimmed, "author:") && current == nil:
			current = &Comment{Author: strings.TrimSpace(strings.TrimPrefix(trimmed, "author:"))}
		case strings.HasPrefix(trimmed, "created:") && current != nil && !inBody:
			current.Created = strings.TrimSpace(strings.TrimPrefix(trimmed, "created:"))
		case trimmed == frontmatterFence && current != nil && !inBody:
			inBody = true
		case trimmed == frontmatterFence && inBody:
			flush()
		case inBody:
			body = append(body, line)
		}
	}
	flush()
	return comments
}

// timestampLayouts covers what the corpus contains. Storage is invariably UTC
// "YYYY-MM-DD HH:MM"; date_format in a project config is a display setting and
// says nothing about how dates are stored, so it is deliberately not consulted.
var timestampLayouts = []string{
	"2006-01-02 15:04",
	"2006-01-02T15:04:05Z07:00",
	"2006-01-02 15:04:05",
	"2006-01-02",
}

// parseTimestamp returns the zero time for anything unparseable. A date that
// cannot be read is not worth failing a file over.
func parseTimestamp(raw string) time.Time {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return time.Time{}
	}
	for _, layout := range timestampLayouts {
		if t, err := time.Parse(layout, raw); err == nil {
			return t
		}
	}
	return time.Time{}
}

// normaliseNewlines folds CRLF, matching what 1.48.0 does before it parses
// sections. No file in the corpus needs it; every file written on Windows will.
func normaliseNewlines(s string) string {
	if !strings.Contains(s, "\r") {
		return s
	}
	s = strings.ReplaceAll(s, "\r\n", "\n")
	return strings.ReplaceAll(s, "\r", "\n")
}
