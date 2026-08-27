package backlog_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/FMakareev/muster-backlog/internal/backlog"
)

const corpus = "../../testdata/backlog-format"

// parseCorpusFile parses one committed fixture as a task.
func parseCorpusFile(t *testing.T, rel string) backlog.Entity {
	t.Helper()
	path := filepath.Join(corpus, rel)
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}
	e, err := backlog.ParseFile(path, raw, backlog.KindTask, backlog.ClassActive, true)
	if err != nil {
		t.Fatalf("ParseFile(%s): %v", rel, err)
	}
	return e
}

func findFixture(t *testing.T, dir string) string {
	t.Helper()
	entries, err := os.ReadDir(filepath.Join(corpus, dir))
	if err != nil {
		t.Fatalf("read %s: %v", dir, err)
	}
	for _, e := range entries {
		if strings.HasSuffix(e.Name(), ".md") {
			return filepath.Join(dir, e.Name())
		}
	}
	t.Fatalf("no fixture in %s", dir)
	return ""
}

// A folded block scalar is the variation that disqualifies a line-oriented
// frontmatter reader: it would see "title:" with an empty value and drop it.
func TestFoldedTitleIsRead(t *testing.T) {
	e := parseCorpusFile(t, findFixture(t, "05-variants/folded-title"))
	if strings.TrimSpace(e.Title) == "" {
		t.Fatal("folded title parsed as empty")
	}
	if strings.Contains(e.Title, "\n") {
		t.Errorf("folded title kept its line breaks: %q", e.Title)
	}
}

// The filename slug is lossy and one-way; five files in the author's projects
// carry no title information in their name at all.
func TestTitleComesFromFrontmatterNotFilename(t *testing.T) {
	rel := findFixture(t, "05-variants/empty-filename-slug")
	e := parseCorpusFile(t, rel)
	if strings.TrimSpace(e.Title) == "" {
		t.Fatalf("no title parsed from %s", rel)
	}
	base := filepath.Base(rel)
	if strings.Contains(base, strings.Fields(e.Title)[0]) {
		t.Skip("this fixture's slug happens to contain the title; the point is untestable here")
	}
}

func TestCRLFIsNormalised(t *testing.T) {
	rel := findFixture(t, "05-variants/crlf-line-endings")
	raw, err := os.ReadFile(filepath.Join(corpus, rel))
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if !strings.Contains(string(raw), "\r\n") {
		t.Fatal("fixture has lost its CRLF line endings - check .gitattributes")
	}
	e := parseCorpusFile(t, rel)
	if strings.Contains(e.Body, "\r") {
		t.Error("body still contains carriage returns")
	}
	if strings.TrimSpace(e.Title) == "" {
		t.Error("no title parsed from a CRLF file")
	}
}

// A task created by the CLI and never edited has no body at all.
func TestTaskWithNoBody(t *testing.T) {
	e := parseCorpusFile(t, findFixture(t, "05-variants/no-body"))
	if len(e.Sections) != 0 {
		t.Errorf("got %d sections, want none", len(e.Sections))
	}
	if e.AcceptanceCriteria != nil {
		t.Error("want no acceptance-criteria section at all, not an empty one")
	}
	if strings.TrimSpace(e.Title) == "" {
		t.Error("a body-less task still has a title")
	}
}

// Present-and-empty is a different state from absent.
func TestEmptyAcceptanceCriteriaBlock(t *testing.T) {
	e := parseCorpusFile(t, findFixture(t, "05-variants/empty-ac-block"))
	if e.AcceptanceCriteria == nil {
		t.Fatal("want a present but empty section, got no section")
	}
	if len(e.AcceptanceCriteria) != 0 {
		t.Errorf("got %d criteria, want 0", len(e.AcceptanceCriteria))
	}
}

// The legacy index-less form is still migrated by 1.48.0, so it can appear.
func TestAcceptanceCriteriaWithoutIndex(t *testing.T) {
	e := parseCorpusFile(t, findFixture(t, "05-variants/legacy-ac-without-index"))
	if len(e.AcceptanceCriteria) == 0 {
		t.Fatal("no criteria parsed from the index-less form")
	}
	for i, c := range e.AcceptanceCriteria {
		if c.Index != i+1 {
			t.Errorf("criterion %d has index %d, want positional numbering", i, c.Index)
		}
		if strings.TrimSpace(c.Text) == "" {
			t.Errorf("criterion %d has no text", i)
		}
	}
}

func TestMixedCheckedAndUnchecked(t *testing.T) {
	e := parseCorpusFile(t, findFixture(t, "01-tasks/acceptance-criteria"))
	if len(e.AcceptanceCriteria) < 2 {
		t.Fatalf("got %d criteria, want several", len(e.AcceptanceCriteria))
	}
	var checked, unchecked int
	for _, c := range e.AcceptanceCriteria {
		if c.Checked {
			checked++
		} else {
			unchecked++
		}
		if c.Index <= 0 {
			t.Errorf("criterion %q has no index", c.Text)
		}
	}
	if checked == 0 || unchecked == 0 {
		t.Errorf("got %d checked and %d unchecked, want both states present",
			checked, unchecked)
	}
}

// Assignee is a list of opaque strings. One file in the author's projects holds
// a bare email address, so nothing may assume a leading @.
func TestAssigneeIsNotNormalised(t *testing.T) {
	e := parseCorpusFile(t, findFixture(t, "05-variants/bare-email-assignee"))
	if len(e.Assignee) == 0 {
		t.Fatal("no assignee parsed")
	}
	if strings.HasPrefix(e.Assignee[0], "@") {
		t.Errorf("assignee %q was normalised; it must be kept verbatim", e.Assignee[0])
	}
}

func TestFullyPopulatedTask(t *testing.T) {
	e := parseCorpusFile(t, findFixture(t, "01-tasks/full"))
	if e.ID == "" || e.Title == "" || e.Status == "" {
		t.Fatalf("core fields missing: %+v", e.Key)
	}
	if _, ok := e.Section(backlog.SectionDescription); !ok {
		t.Error("no description section")
	}
	if !e.HasOrdinal {
		t.Error("fixture should carry an ordinal")
	}
}

func TestDependenciesAreRead(t *testing.T) {
	e := parseCorpusFile(t, findFixture(t, "01-tasks/dependencies"))
	if len(e.Dependencies) == 0 {
		t.Fatal("no dependencies parsed")
	}
}

// Seven task files contain bare --- lines inside comment envelopes. A reader
// that scans the whole file for a fence takes half a comment for frontmatter.
func TestCommentEnvelopeDoesNotConfuseFrontmatter(t *testing.T) {
	e := parseCorpusFile(t, findFixture(t, "01-tasks/comments"))
	if e.ID == "" || e.Title == "" {
		t.Fatalf("frontmatter mis-parsed around the comment envelope: %+v", e.Key)
	}
	if len(e.Comments) == 0 {
		t.Fatal("no comments parsed")
	}
	for i, c := range e.Comments {
		if c.Author == "" {
			t.Errorf("comment %d has no author", i)
		}
		if strings.TrimSpace(c.Body) == "" {
			t.Errorf("comment %d has no body", i)
		}
	}
}

// A document may quote task markers in prose; doc-1 in this repository does.
// The marker grammar must not be applied to it.
func TestDocumentMarkersAreNotParsed(t *testing.T) {
	rel := findFixture(t, "02-entities/docs")
	raw, err := os.ReadFile(filepath.Join(corpus, rel))
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	e, err := backlog.ParseFile(rel, raw, backlog.KindDocument, backlog.ClassActive, false)
	if err != nil {
		t.Fatalf("ParseFile: %v", err)
	}
	if len(e.Sections) != 0 {
		t.Errorf("got %d sections in a document, want none parsed", len(e.Sections))
	}
	if e.Body == "" {
		t.Error("a document's whole content is its body")
	}
}

func TestNonEntityFilesAreRejected(t *testing.T) {
	cases := map[string]string{
		"no frontmatter":    "# Just a heading\n\nSome prose.\n",
		"unclosed":          "---\nid: TASK-1\n",
		"frontmatter later": "Some prose.\n\n---\nid: TASK-1\ntitle: x\n---\n",
		"no id":             "---\ntitle: Something\n---\n\nBody.\n",
		"empty file":        "",
	}
	for name, content := range cases {
		t.Run(name, func(t *testing.T) {
			_, err := backlog.ParseFile("x.md", []byte(content),
				backlog.KindTask, backlog.ClassActive, true)
			if err == nil {
				t.Fatal("want an error, got none")
			}
		})
	}
}

// A comment with no author is a state the CLI produces: `task edit --comment`
// without `--comment-author` writes the entry with no author line at all,
// rather than an empty one. Every comment in the corpus is signed, so this was
// missed until the application began writing them.
func TestAnUnsignedCommentIsStillAComment(t *testing.T) {
	body := "---\nid: TASK-1\ntitle: Talking\nstatus: To Do\n---\n\n" +
		"## Comments\n\n<!-- COMMENTS:BEGIN -->\n" +
		"author: @someone\ncreated: 2026-08-27 19:00\n---\nSigned.\n---\n\n" +
		"created: 2026-08-27 19:13\n---\nUnsigned.\n---\n" +
		"<!-- COMMENTS:END -->\n"

	entity, err := backlog.ParseFile("task-1 - talking.md", []byte(body),
		backlog.KindTask, backlog.ClassActive, true)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if len(entity.Comments) != 2 {
		t.Fatalf("got %d comments, want both: %+v", len(entity.Comments), entity.Comments)
	}
	if entity.Comments[0].Author != "@someone" || entity.Comments[0].Body != "Signed." {
		t.Errorf("the signed one is %+v", entity.Comments[0])
	}
	second := entity.Comments[1]
	switch {
	case second.Author != "":
		t.Errorf("an author appeared on an unsigned comment: %q", second.Author)
	case second.Body != "Unsigned.":
		t.Errorf("body is %q", second.Body)
	case second.Created == "":
		t.Errorf("the date was lost: %+v", second)
	}
}
