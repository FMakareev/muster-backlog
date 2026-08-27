package app_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/FMakareev/muster-backlog/internal/app"
)

// draft writes one draft file directly, so the list and its ordering can be
// tested without the CLI. Writing is always tested through the CLI instead.
func draft(t *testing.T, root, id, title string, captured time.Time) {
	t.Helper()
	dir := filepath.Join(root, "backlog", "drafts")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	date := ""
	if !captured.IsZero() {
		date = "created_date: '" + captured.Format("2006-01-02 15:04") + "'\n"
	}
	body := "---\nid: " + id + "\ntitle: " + title + "\nstatus: Draft\n" + date +
		"---\n\n## Description\n\n<!-- SECTION:DESCRIPTION:BEGIN -->\n" +
		title + " body\n<!-- SECTION:DESCRIPTION:END -->\n"
	name := strings.ToLower(id) + " - " + strings.ReplaceAll(title, " ", "-") + ".md"
	if err := os.WriteFile(filepath.Join(dir, name), []byte(body), 0o644); err != nil {
		t.Fatalf("write draft: %v", err)
	}
}

func titles(drafts []app.DraftView) []string {
	var out []string
	for _, d := range drafts {
		out = append(out, d.Entity.Title)
	}
	return out
}

// The inbox is emptied oldest first, so that is the order it is shown in.
func TestDraftsAreListedOldestFirstAcrossProjects(t *testing.T) {
	one := newProject(t, "one", nil)
	two := newProject(t, "two", nil)
	now := time.Now()

	draft(t, one, "DRAFT-1", "Three days old", now.AddDate(0, 0, -3))
	draft(t, one, "DRAFT-2", "Undated", time.Time{})
	draft(t, two, "DRAFT-1", "Ten days old", now.AddDate(0, 0, -10))
	draft(t, two, "DRAFT-2", "Captured today", now)

	s := startService(t, withRegistry(t, one, two))
	got := s.Drafts()

	want := []string{"Ten days old", "Three days old", "Captured today", "Undated"}
	if strings.Join(titles(got), ",") != strings.Join(want, ",") {
		t.Fatalf("order is %v, want %v", titles(got), want)
	}
	if got[0].WaitingDays != 10 || got[1].WaitingDays != 3 || got[2].WaitingDays != 0 {
		t.Errorf("waits are %d, %d, %d", got[0].WaitingDays, got[1].WaitingDays, got[2].WaitingDays)
	}
	// Nothing is known about an undated note, and -1 says so rather than
	// pretending it arrived today.
	if got[3].WaitingDays != -1 {
		t.Errorf("an undated draft reports a wait of %d", got[3].WaitingDays)
	}
	// Ids collide across projects; both are here.
	if got[0].Project == got[1].Project {
		t.Error("the two projects' drafts were not both listed")
	}
}

// A discarded draft is archived, not deleted, and must not come back in the
// inbox afterwards.
func TestDiscardedDraftsLeaveTheInbox(t *testing.T) {
	if _, err := exec.LookPath("backlog"); err != nil {
		t.Skip("the backlog CLI is not installed")
	}
	one := newProject(t, "one", nil)
	draft(t, one, "DRAFT-1", "A passing thought", time.Now().AddDate(0, 0, -1))
	s := startService(t, withRegistry(t, one))

	if len(s.Drafts()) != 1 {
		t.Fatalf("started with %d drafts", len(s.Drafts()))
	}
	if result := s.DiscardDraft(one, "DRAFT-1"); !result.OK {
		t.Fatalf("DiscardDraft: %+v", result.Problem)
	}
	if got := s.Drafts(); len(got) != 0 {
		t.Errorf("a discarded draft is still in the inbox: %v", titles(got))
	}
	// Archived, not gone: Backlog.md has no delete and neither does Muster.
	entries, err := os.ReadDir(filepath.Join(one, "backlog", "archive", "drafts"))
	if err != nil || len(entries) != 1 {
		t.Errorf("the note was not archived: %v %v", entries, err)
	}
}

func TestPromotingADraftMakesItATask(t *testing.T) {
	if _, err := exec.LookPath("backlog"); err != nil {
		t.Skip("the backlog CLI is not installed")
	}
	one := newProject(t, "one", nil)
	draft(t, one, "DRAFT-1", "Worth doing", time.Now().AddDate(0, 0, -2))
	s := startService(t, withRegistry(t, one))

	result := s.PromoteDraft(one, "DRAFT-1")
	if !result.OK {
		t.Fatalf("PromoteDraft: %+v", result.Problem)
	}
	// The note has to say what it became, or promoting it means going to look
	// for it.
	if result.TaskID == "" {
		t.Error("promotion did not report the task it produced")
	}
	if got := s.Drafts(); len(got) != 0 {
		t.Errorf("the draft is still in the inbox: %v", titles(got))
	}
	tasks := s.Tasks(app.QueryInput{})
	if len(tasks) != 1 || tasks[0].Entity.Title != "Worth doing" {
		t.Fatalf("got %d tasks: %+v", len(tasks), tasks)
	}
	// The board's first status, not Draft.
	if tasks[0].Entity.Status == "Draft" {
		t.Error("the promoted task kept the draft status")
	}
}

// Backlog.md has no draft edit, so revising is capture-and-discard. The note
// must survive that whether or not it stays in the same project.
func TestRevisingADraftRewritesItInPlace(t *testing.T) {
	if _, err := exec.LookPath("backlog"); err != nil {
		t.Skip("the backlog CLI is not installed")
	}
	one := newProject(t, "one", nil)
	draft(t, one, "DRAFT-1", "vaguely worded thing", time.Now().AddDate(0, 0, -5))
	s := startService(t, withRegistry(t, one))

	result := s.ReviseDraft(one, "DRAFT-1", app.DraftEdit{
		Title:       "Clearly worded thing",
		Description: "With the context I remembered later.",
		Labels:      []string{"idea", "later"},
	})
	if !result.OK {
		t.Fatalf("ReviseDraft: %+v", result.Problem)
	}

	got := s.Drafts()
	if len(got) != 1 {
		t.Fatalf("got %d drafts, want the one rewritten: %v", len(got), titles(got))
	}
	if got[0].Entity.Title != "Clearly worded thing" {
		t.Errorf("title is %q", got[0].Entity.Title)
	}
	if !strings.Contains(got[0].Entity.Sections["description"], "remembered later") {
		t.Errorf("the body did not carry over: %q", got[0].Entity.Sections["description"])
	}
	if strings.Join(got[0].Entity.Labels, ",") != "idea,later" {
		t.Errorf("labels are %v", got[0].Entity.Labels)
	}
	// The wait restarts, which is the cost of having no draft edit.
	if got[0].WaitingDays != 0 {
		t.Errorf("the rewritten note reports a wait of %d days", got[0].WaitingDays)
	}
}

func TestRevisingADraftCanMoveItToAnotherProject(t *testing.T) {
	if _, err := exec.LookPath("backlog"); err != nil {
		t.Skip("the backlog CLI is not installed")
	}
	one := newProject(t, "one", nil)
	two := newProject(t, "two", nil)
	draft(t, one, "DRAFT-1", "Belongs elsewhere", time.Now().AddDate(0, 0, -1))
	s := startService(t, withRegistry(t, one, two))

	result := s.ReviseDraft(one, "DRAFT-1", app.DraftEdit{
		Title: "Belongs elsewhere", Project: two,
	})
	if !result.OK {
		t.Fatalf("ReviseDraft: %+v", result.Problem)
	}
	got := s.Drafts()
	if len(got) != 1 {
		t.Fatalf("got %d drafts: %v", len(got), titles(got))
	}
	if got[0].Project != two {
		t.Errorf("the note stayed in %s", got[0].Project)
	}
}

func TestRevisingRefusesWhatCannotWork(t *testing.T) {
	if _, err := exec.LookPath("backlog"); err != nil {
		t.Skip("the backlog CLI is not installed")
	}
	one := newProject(t, "one", nil)
	draft(t, one, "DRAFT-1", "Something", time.Now())
	s := startService(t, withRegistry(t, one))

	if result := s.ReviseDraft(one, "DRAFT-1", app.DraftEdit{Title: "  "}); result.OK {
		t.Error("an empty title was accepted")
	}
	if result := s.ReviseDraft(one, "DRAFT-1", app.DraftEdit{
		Title: "Something", Project: "/nowhere",
	}); result.OK {
		t.Error("a project that is not registered was accepted")
	}
	// Neither refusal touched the note.
	if got := s.Drafts(); len(got) != 1 || got[0].Entity.Title != "Something" {
		t.Errorf("a refused revision changed the inbox: %v", titles(got))
	}
}

// A write whose outcome can be checked is checked, because a zero exit is not
// proof: on 1.48.0 both promote and archive exit 0 for an id they cannot find.
func TestDraftWritesAreConfirmedByOutcome(t *testing.T) {
	if _, err := exec.LookPath("backlog"); err != nil {
		t.Skip("the backlog CLI is not installed")
	}
	one := newProject(t, "one", nil)
	draft(t, one, "DRAFT-1", "Still here", time.Now())
	s := startService(t, withRegistry(t, one))

	for name, ok := range map[string]bool{
		"promote": s.PromoteDraft(one, "DRAFT-404").OK,
		"discard": s.DiscardDraft(one, "DRAFT-404").OK,
	} {
		if ok {
			t.Errorf("%s of an id that does not resolve reported success", name)
		}
	}
	// And the real note is untouched by any of it.
	if got := s.Drafts(); len(got) != 1 || got[0].Entity.Title != "Still here" {
		t.Errorf("the inbox changed: %v", titles(got))
	}
}

// Capture is the other half of a usable inbox, and a captured note may carry
// everything the CLI lets a draft hold.
func TestCaptureNoteCarriesTheWholeFieldSurface(t *testing.T) {
	if _, err := exec.LookPath("backlog"); err != nil {
		t.Skip("the backlog CLI is not installed")
	}
	one := newProject(t, "one", nil)
	s := startService(t, withRegistry(t, one))

	result := s.CaptureNote(one, app.DraftEdit{
		Title:              "A thought with detail",
		Description:        "The context, written down while it was fresh.",
		Labels:             []string{"idea", "later"},
		Assignee:           "@me",
		Priority:           "high",
		AcceptanceCriteria: []string{"It is testable", "It is finishable"},
	})
	if !result.OK {
		t.Fatalf("CaptureNote: %+v", result.Problem)
	}

	got := s.Drafts()
	if len(got) != 1 {
		t.Fatalf("got %d drafts", len(got))
	}
	note := got[0].Entity
	switch {
	case note.Title != "A thought with detail":
		t.Errorf("title is %q", note.Title)
	case note.Status != "Draft":
		t.Errorf("status is %q, want Draft", note.Status)
	case note.Priority != "high":
		t.Errorf("priority is %q - the field a draft could not hold before", note.Priority)
	case strings.Join(note.Labels, ",") != "idea,later":
		t.Errorf("labels are %v", note.Labels)
	case len(note.AcceptanceCriteria) != 2:
		t.Errorf("got %d acceptance criteria", len(note.AcceptanceCriteria))
	case !strings.Contains(note.Sections["description"], "while it was fresh"):
		t.Errorf("body is %q", note.Sections["description"])
	}

	if result := s.CaptureNote(one, app.DraftEdit{Title: "   "}); result.OK {
		t.Error("a note with no title was accepted")
	}
}

// Rewriting keeps the fields a draft can hold, which is the point of doing it
// through task create --draft rather than draft create.
func TestRevisingKeepsEveryFieldADraftCanHold(t *testing.T) {
	if _, err := exec.LookPath("backlog"); err != nil {
		t.Skip("the backlog CLI is not installed")
	}
	one := newProject(t, "one", nil)
	draft(t, one, "DRAFT-1", "Rough", time.Now().AddDate(0, 0, -2))
	s := startService(t, withRegistry(t, one))

	result := s.ReviseDraft(one, "DRAFT-1", app.DraftEdit{
		Title:              "Sharpened",
		Description:        "Now with reasons.",
		Priority:           "medium",
		Labels:             []string{"kept"},
		AcceptanceCriteria: []string{"Still true afterwards"},
	})
	if !result.OK {
		t.Fatalf("ReviseDraft: %+v", result.Problem)
	}
	got := s.Drafts()
	if len(got) != 1 {
		t.Fatalf("got %d drafts: %v", len(got), titles(got))
	}
	note := got[0].Entity
	if note.Priority != "medium" || len(note.AcceptanceCriteria) != 1 {
		t.Errorf("the rewrite dropped fields: priority %q, %d criteria",
			note.Priority, len(note.AcceptanceCriteria))
	}
}

// Notes open in the panel, which used to mean every task-editing control was
// offered against one — and every one of them failed, because Backlog.md has
// no way to edit a draft. The interface no longer offers them; this is the
// floor under that.
func TestTaskEditsAreRefusedOnANote(t *testing.T) {
	if _, err := exec.LookPath("backlog"); err != nil {
		t.Skip("the backlog CLI is not installed")
	}
	one := newProject(t, "one", map[string]string{
		"task-1 - a.md": task("TASK-1", "A real task", "To Do"),
	})
	draft(t, one, "DRAFT-4", "A captured thought", time.Now())
	s := startService(t, withRegistry(t, one))

	attempts := map[string]func() app.WriteResult{
		"description": func() app.WriteResult {
			return s.SetSection(one, "DRAFT-4", "description", "Rewritten.")
		},
		"title":    func() app.WriteResult { return s.SetTitle(one, "DRAFT-4", "Renamed") },
		"status":   func() app.WriteResult { return s.SetStatus(one, "DRAFT-4", "To Do") },
		"priority": func() app.WriteResult { return s.SetPriority(one, "DRAFT-4", "high") },
		"label":    func() app.WriteResult { return s.AddLabel(one, "DRAFT-4", "idea") },
		"deps": func() app.WriteResult {
			return s.SetDependencies(one, "DRAFT-4", []string{"TASK-1"})
		},
	}
	for name, attempt := range attempts {
		result := attempt()
		if result.OK {
			t.Errorf("editing the %s of a note was accepted", name)
			continue
		}
		if result.Problem == nil || !strings.Contains(result.Problem.Title, "not a task") {
			t.Errorf("%s: refused with %+v, want it named as a note", name, result.Problem)
		}
		// And the CLI's own confusing answer never reaches the interface.
		if result.Problem != nil && strings.Contains(result.Problem.Detail, "not found") {
			t.Errorf("%s: the CLI's message leaked: %q", name, result.Problem.Detail)
		}
	}

	// The note is untouched, and the same edits still work on a real task.
	if got := s.Drafts(); len(got) != 1 || got[0].Entity.Title != "A captured thought" {
		t.Errorf("the note changed: %v", titles(got))
	}
	if result := s.SetTitle(one, "TASK-1", "Still editable"); !result.OK {
		t.Errorf("editing a real task broke: %+v", result.Problem)
	}
}
