package app_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/FMakareev/muster-backlog/internal/app"
	"github.com/FMakareev/muster-backlog/internal/settings"
)

func commentsOn(t *testing.T, s *app.BoardService, project, id string) []struct {
	Author string
	Body   string
} {
	t.Helper()
	view, ok := s.Task(project, "task", "active", id)
	if !ok {
		t.Fatalf("%s is gone", id)
	}
	var out []struct {
		Author string
		Body   string
	}
	for _, c := range view.Entity.Comments {
		out = append(out, struct {
			Author string
			Body   string
		}{c.Author, c.Body})
	}
	return out
}

func TestACommentIsWrittenAndReadBack(t *testing.T) {
	if _, err := exec.LookPath("backlog"); err != nil {
		t.Skip("the backlog CLI is not installed")
	}
	one := newProject(t, "one", map[string]string{
		"task-1 - a.md": task("TASK-1", "Something to discuss", "To Do"),
	})
	s := startService(t, withRegistry(t, one))

	if result := s.AddComment(one, "TASK-1", "The first thing to say."); !result.OK {
		t.Fatalf("AddComment: %+v", result.Problem)
	}
	if result := s.AddComment(one, "TASK-1", "And the second."); !result.OK {
		t.Fatalf("AddComment: %+v", result.Problem)
	}

	got := commentsOn(t, s, one, "TASK-1")
	if len(got) != 2 {
		t.Fatalf("got %d comments, want both", len(got))
	}
	// In the order the file holds them: a conversation read backwards is not
	// a conversation.
	if !strings.Contains(got[0].Body, "first thing") ||
		!strings.Contains(got[1].Body, "second") {
		t.Errorf("comments are %+v", got)
	}

	if result := s.AddComment(one, "TASK-1", "   "); result.OK {
		t.Error("an empty comment was accepted")
	}
	if got := commentsOn(t, s, one, "TASK-1"); len(got) != 2 {
		t.Errorf("a refused comment was written anyway: %d", len(got))
	}
}

// The name comes from the person, not from a constant. A preference first,
// then git, then nothing at all.
func TestWhoACommentIsSignedBy(t *testing.T) {
	if _, err := exec.LookPath("backlog"); err != nil {
		t.Skip("the backlog CLI is not installed")
	}
	one := newProject(t, "one", map[string]string{
		"task-1 - a.md": task("TASK-1", "Something to discuss", "To Do"),
	})
	s := startService(t, withRegistry(t, one))

	// A preference wins, and is written as these files write a name.
	prefs := settings.Defaults()
	prefs.Author = "someone"
	if problems := s.SaveSettings(prefs); len(problems) > 0 {
		t.Fatalf("SaveSettings: %+v", problems)
	}
	if got := s.CommentAuthor(one); got != "@someone" {
		t.Errorf("author is %q, want it written as a handle", got)
	}
	if result := s.AddComment(one, "TASK-1", "Said by someone."); !result.OK {
		t.Fatalf("AddComment: %+v", result.Problem)
	}
	got := commentsOn(t, s, one, "TASK-1")
	if len(got) != 1 || got[0].Author != "@someone" {
		t.Errorf("the comment is %+v", got)
	}
	// A name already written as a handle is not given a second @.
	prefs.Author = "@already"
	if problems := s.SaveSettings(prefs); len(problems) > 0 {
		t.Fatalf("SaveSettings: %+v", problems)
	}
	if got := s.CommentAuthor(one); got != "@already" {
		t.Errorf("author is %q", got)
	}
}

// Where a project is a git repository and nothing was set, git is asked -
// which is where the name already is, and what the files already agree with.
func TestTheGitIdentityIsUsedWhenNothingIsSet(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git is not installed")
	}
	one := newProject(t, "one", map[string]string{
		"task-1 - a.md": task("TASK-1", "Something", "To Do"),
	})
	for _, args := range [][]string{
		{"init", "-q"},
		{"config", "user.name", "Someone Local"},
		{"config", "user.email", "someone@example.test"},
	} {
		cmd := exec.Command("git", args...)
		cmd.Dir = one
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
	if _, err := os.Stat(filepath.Join(one, ".git")); err != nil {
		t.Fatalf("no repository: %v", err)
	}

	s := startService(t, withRegistry(t, one))
	if got := s.CommentAuthor(one); got != "@Someone Local" {
		t.Errorf("author is %q, want the git identity", got)
	}
}

// With no preference and no git to ask, a comment is written unsigned rather
// than signed with something invented. The format has that state and the CLI
// produces it: the author line is simply absent.
func TestACommentIsUnsignedWhenNobodyKnowsTheName(t *testing.T) {
	if _, err := exec.LookPath("backlog"); err != nil {
		t.Skip("the backlog CLI is not installed")
	}
	// git answers for a folder that is not a repository by falling back to the
	// global configuration, which is usually right - it is still the person.
	// Silence both so the last fallback is the one under test.
	t.Setenv("GIT_CONFIG_GLOBAL", os.DevNull)
	t.Setenv("GIT_CONFIG_SYSTEM", os.DevNull)

	one := newProject(t, "one", map[string]string{
		"task-1 - a.md": task("TASK-1", "Something", "To Do"),
	})
	s := startService(t, withRegistry(t, one))

	if got := s.CommentAuthor(one); got != "" {
		t.Fatalf("a name appeared from nowhere: %q", got)
	}
	if result := s.AddComment(one, "TASK-1", "Said by nobody in particular."); !result.OK {
		t.Fatalf("AddComment: %+v", result.Problem)
	}
	got := commentsOn(t, s, one, "TASK-1")
	if len(got) != 1 || got[0].Author != "" {
		t.Errorf("expected one unsigned comment, got %+v", got)
	}
}
