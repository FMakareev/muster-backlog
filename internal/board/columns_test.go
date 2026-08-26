package board_test

import (
	"strings"
	"testing"

	"github.com/FMakareev/muster-backlog/internal/board"
)

func names(l board.Layout) string {
	var out []string
	for _, c := range l.Columns {
		out = append(out, c.Name)
	}
	return strings.Join(out, " | ")
}

func lists(pairs ...[2]string) []board.ProjectStatuses {
	out := make([]board.ProjectStatuses, 0, len(pairs))
	for _, p := range pairs {
		out = append(out, board.ProjectStatuses{
			Project:  p[0],
			Statuses: strings.Split(p[1], ","),
		})
	}
	return out
}

// The real corpus: eight projects on three statuses, one on four. The union is
// the superset, and because one list contains the other the order is not in
// dispute at all.
func TestTheActualCorpus(t *testing.T) {
	var projects []board.ProjectStatuses
	for _, name := range []string{"a", "b", "c", "d", "e", "f", "g", "h"} {
		projects = append(projects, board.ProjectStatuses{
			Project: name, Statuses: []string{"To Do", "In Progress", "Done"},
		})
	}
	projects = append(projects, board.ProjectStatuses{
		Project: "muster", Statuses: []string{"To Do", "In Progress", "In Review", "Done"},
	})

	got := board.Build(projects)
	if want := "To Do | In Progress | In Review | Done"; names(got) != want {
		t.Errorf("columns = %q, want %q", names(got), want)
	}
	if len(got.Conflicts) != 0 {
		t.Errorf("conflicts = %+v, want none", got.Conflicts)
	}

	// In Review belongs to one project only. The others have no cell there.
	review, ok := got.Column("In Review")
	if !ok {
		t.Fatal("no In Review column")
	}
	if len(review.Projects) != 1 || review.Projects[0] != "muster" {
		t.Errorf("In Review declared by %v", review.Projects)
	}
	if review.Declares("a") {
		t.Error("a project that does not declare a status must not claim the column")
	}
}

// Lists that share no statuses still have to produce one board.
func TestDisjointLists(t *testing.T) {
	got := board.Build(lists(
		[2]string{"alpha", "To Do,Done"},
		[2]string{"beta", "Backlog,Doing,Shipped"},
	))
	if len(got.Columns) != 5 {
		t.Fatalf("columns = %q, want all five", names(got))
	}
	// Each project's own order must survive within the union.
	if before(got, "To Do", "Done") == false {
		t.Errorf("columns = %q, want To Do before Done", names(got))
	}
	if !before(got, "Backlog", "Doing") || !before(got, "Doing", "Shipped") {
		t.Errorf("columns = %q, want beta's order intact", names(got))
	}
}

// Eight projects agreeing must outvote one that disagrees.
func TestMajorityWinsAConflict(t *testing.T) {
	var projects []board.ProjectStatuses
	for _, name := range []string{"a", "b", "c", "d", "e", "f", "g", "h"} {
		projects = append(projects, board.ProjectStatuses{
			Project: name, Statuses: []string{"To Do", "Done"},
		})
	}
	projects = append(projects, board.ProjectStatuses{
		Project: "odd", Statuses: []string{"Done", "To Do"},
	})

	got := board.Build(projects)
	if !before(got, "To Do", "Done") {
		t.Errorf("columns = %q, want the majority order", names(got))
	}
	if len(got.Conflicts) != 1 {
		t.Fatalf("conflicts = %+v, want the disagreement recorded", got.Conflicts)
	}
	c := got.Conflicts[0]
	if c.Votes != 8 || c.Against != 1 {
		t.Errorf("conflict = %+v, want 8 against 1", c)
	}
}

// Three projects can disagree in a loop, and no order satisfies all of them.
// The board still has to be drawable.
func TestCycleIsBrokenDeterministically(t *testing.T) {
	projects := lists(
		[2]string{"alpha", "A,B"},
		[2]string{"beta", "B,C"},
		[2]string{"gamma", "C,A"},
	)

	first := board.Build(projects)
	if len(first.Columns) != 3 {
		t.Fatalf("columns = %q, want all three", names(first))
	}
	// Whatever it decides, it must decide the same way every time.
	for range 20 {
		if got := names(board.Build(projects)); got != names(first) {
			t.Fatalf("order is not deterministic: %q then %q", names(first), got)
		}
	}
}

// Nothing may depend on Go's map iteration order.
func TestOrderIsStableAcrossRuns(t *testing.T) {
	projects := lists(
		[2]string{"a", "To Do,In Progress,In Review,Done"},
		[2]string{"b", "To Do,Blocked,In Progress,Done"},
		[2]string{"c", "Triage,To Do,Done,Archived"},
	)
	want := names(board.Build(projects))
	for range 50 {
		if got := names(board.Build(projects)); got != want {
			t.Fatalf("order is not stable: %q then %q", want, got)
		}
	}
}

// Enum case is not normalised on disk, so two spellings are one column, and
// the spelling shown is the first one seen in registry order.
func TestCaseInsensitiveWithFirstSpellingKept(t *testing.T) {
	got := board.Build(lists(
		[2]string{"alpha", "To Do,Done"},
		[2]string{"beta", "TO DO,done"},
	))
	if len(got.Columns) != 2 {
		t.Fatalf("columns = %q, want two", names(got))
	}
	if got.Columns[0].Name != "To Do" {
		t.Errorf("name = %q, want the first project's spelling", got.Columns[0].Name)
	}
	column, _ := got.Column("to do")
	if len(column.Projects) != 2 {
		t.Errorf("declared by %v, want both projects", column.Projects)
	}
}

// A task may only take a status its own project declares. Writing anything
// else would produce a task the CLI itself considers invalid.
func TestCanMove(t *testing.T) {
	layout := board.Build(lists(
		[2]string{"alpha", "To Do,In Progress,Done"},
		[2]string{"muster", "To Do,In Progress,In Review,Done"},
	))

	if !layout.CanMove("alpha", "In Progress") {
		t.Error("alpha declares In Progress")
	}
	if layout.CanMove("alpha", "In Review") {
		t.Error("alpha does not declare In Review and must not be moved there")
	}
	if !layout.CanMove("muster", "In Review") {
		t.Error("muster declares In Review")
	}
	if layout.CanMove("alpha", "Nonexistent") {
		t.Error("a status no project declares is not a column at all")
	}
	// Case follows the same rule as everywhere else.
	if !layout.CanMove("alpha", "in progress") {
		t.Error("status comparison must be case-insensitive")
	}
}

func TestEmptyAndDegenerateInput(t *testing.T) {
	if got := board.Build(nil); len(got.Columns) != 0 {
		t.Errorf("columns = %q, want none", names(got))
	}
	if got := board.Build(lists([2]string{"alpha", ""})); len(got.Columns) != 0 {
		t.Errorf("columns = %q, want none for a project declaring nothing", names(got))
	}
	// A project that repeats a status must not produce two columns.
	got := board.Build(lists([2]string{"alpha", "To Do,To Do,Done"}))
	if len(got.Columns) != 2 {
		t.Errorf("columns = %q, want two", names(got))
	}
}

// A project whose config gains a status must gain a column, without anything
// being restarted: Build is called again with the new lists.
func TestAddingAStatusAddsAColumn(t *testing.T) {
	before := board.Build(lists([2]string{"alpha", "To Do,Done"}))
	if len(before.Columns) != 2 {
		t.Fatalf("columns = %q", names(before))
	}

	after := board.Build(lists([2]string{"alpha", "To Do,In Review,Done"}))
	if len(after.Columns) != 3 {
		t.Fatalf("columns = %q, want the new status", names(after))
	}
	if !after.CanMove("alpha", "In Review") {
		t.Error("the new status must immediately be a valid target")
	}

	removed := board.Build(lists([2]string{"alpha", "To Do,Done"}))
	if _, ok := removed.Column("In Review"); ok {
		t.Error("a removed status must stop being a column")
	}
}

// before reports whether one column comes before another.
func before(l board.Layout, a, b string) bool {
	ai, bi := -1, -1
	for i, c := range l.Columns {
		if strings.EqualFold(c.Name, a) {
			ai = i
		}
		if strings.EqualFold(c.Name, b) {
			bi = i
		}
	}
	return ai >= 0 && bi >= 0 && ai < bi
}
