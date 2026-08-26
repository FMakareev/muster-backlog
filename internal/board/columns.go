// Package board turns what projects declare into what the board shows.
//
// Statuses are per-project configuration in Backlog.md and projects do not
// agree: in the author's own nine projects there are two different lists
// already. Muster never edits another project's config to make its own view
// simpler, so the board has to adapt instead - its columns are the union of
// every list, and a task can only ever move within its own project's.
package board

import (
	"sort"
	"strings"
)

// ProjectStatuses is one project's declared status list.
type ProjectStatuses struct {
	// Project is the registered project path.
	Project string
	// Statuses is the list as written in that project's config, in its order.
	Statuses []string
}

// Column is one column of the unified board.
type Column struct {
	// Name is the status, spelled as the first project to declare it wrote it.
	Name string
	// Projects are the projects that declare this status, in registry order.
	// A project absent from this list has no cell here - not an error, simply
	// a status it does not use.
	Projects []string
}

// Declares reports whether a project uses this status.
func (c Column) Declares(project string) bool {
	for _, p := range c.Projects {
		if p == project {
			return true
		}
	}
	return false
}

// Layout is the resolved board.
type Layout struct {
	Columns []Column
	// Conflicts records pairs whose order the projects disagreed about, so the
	// interface can explain an order that will otherwise look arbitrary.
	Conflicts []Conflict
}

// Conflict is one pair of statuses the projects ordered differently.
type Conflict struct {
	// Before and After are the order that won.
	Before, After string
	// Votes is how many projects put Before first; Against is how many put
	// After first.
	Votes, Against int
}

// Column returns a column by name, matched case-insensitively.
func (l Layout) Column(name string) (Column, bool) {
	for _, c := range l.Columns {
		if strings.EqualFold(c.Name, name) {
			return c, true
		}
	}
	return Column{}, false
}

// CanMove reports whether a task in a project may be moved to a status.
//
// A task can only take a status its own project declares. Writing any other
// value would produce a task the Backlog.md CLI itself considers invalid, and
// the alternative - editing that project's config so the status exists - is
// exactly what this application does not do.
func (l Layout) CanMove(project, status string) bool {
	column, ok := l.Column(status)
	return ok && column.Declares(project)
}

// Build resolves the board layout from what each project declares.
//
// Ordering is a weighted vote. Every project that declares two statuses has an
// opinion about which comes first, and each opinion counts once. For each pair
// the stronger direction wins, so eight projects agreeing outvote one that
// does not. The winning directions form a graph which is then sorted so that
// every status comes after everything that beat it.
//
// Two things have to be decided rather than left to chance:
//
//   - Ties. When a pair is drawn, or when several statuses are equally ready
//     to be placed, the one declared by more projects goes first; if still
//     level, the one seen earlier in registry order; if still level, the
//     alphabetically earlier one. Registry order is a user-visible choice, so
//     deriving from it keeps the result predictable and stable.
//   - Cycles. Three projects can disagree in a loop - A before B, B before C,
//     C before A - and no order satisfies all three. The same tie-break above
//     picks the next status and the loop is cut there. The alternative would
//     be to refuse to draw a board.
func Build(projects []ProjectStatuses) Layout {
	// Statuses are matched case-insensitively but keep the spelling of the
	// first project to declare them: enum case is not normalised on disk.
	canonical := map[string]string{}
	firstSeen := map[string]int{}
	declaredBy := map[string][]string{}
	var order []string

	for _, p := range projects {
		for _, raw := range p.Statuses {
			status := strings.TrimSpace(raw)
			if status == "" {
				continue
			}
			key := strings.ToLower(status)
			if _, known := canonical[key]; !known {
				canonical[key] = status
				firstSeen[key] = len(order)
				order = append(order, key)
			}
			if !contains(declaredBy[key], p.Project) {
				declaredBy[key] = append(declaredBy[key], p.Project)
			}
		}
	}

	if len(order) == 0 {
		return Layout{}
	}

	votes := tally(projects)
	edges, conflicts := resolve(order, votes, canonical, declaredBy, firstSeen)
	sorted := topological(order, edges, declaredBy, firstSeen)

	layout := Layout{Conflicts: conflicts}
	for _, key := range sorted {
		layout.Columns = append(layout.Columns, Column{
			Name:     canonical[key],
			Projects: declaredBy[key],
		})
	}
	return layout
}

// pair identifies an ordered pair of statuses.
type pair struct{ before, after string }

// tally counts, for every ordered pair, how many projects put one before the
// other. Every pair in a list is counted, not only adjacent ones, so a project
// that declares four statuses contributes six opinions rather than three.
func tally(projects []ProjectStatuses) map[pair]int {
	votes := map[pair]int{}
	for _, p := range projects {
		seen := make([]string, 0, len(p.Statuses))
		for _, raw := range p.Statuses {
			status := strings.ToLower(strings.TrimSpace(raw))
			if status == "" || contains(seen, status) {
				continue
			}
			for _, earlier := range seen {
				votes[pair{earlier, status}]++
			}
			seen = append(seen, status)
		}
	}
	return votes
}

// resolve keeps the winning direction of each disputed pair.
func resolve(
	order []string,
	votes map[pair]int,
	canonical map[string]string,
	declaredBy map[string][]string,
	firstSeen map[string]int,
) (map[pair]bool, []Conflict) {
	edges := map[pair]bool{}
	var conflicts []Conflict

	for i, a := range order {
		for _, b := range order[i+1:] {
			forward := votes[pair{a, b}]
			backward := votes[pair{b, a}]
			if forward == 0 && backward == 0 {
				// No project declares both, so nothing to order them by; the
				// tie-break in the sort decides.
				continue
			}

			winner, loser := a, b
			switch {
			case backward > forward:
				winner, loser = b, a
			case forward == backward:
				// A genuine draw. Fall back to the same rule the sort uses, so
				// there is one tie-break in this package rather than two.
				if !precedes(a, b, declaredBy, firstSeen) {
					winner, loser = b, a
				}
			}
			edges[pair{winner, loser}] = true

			if forward > 0 && backward > 0 {
				conflicts = append(conflicts, Conflict{
					Before:  canonical[winner],
					After:   canonical[loser],
					Votes:   max(forward, backward),
					Against: min(forward, backward),
				})
			}
		}
	}
	return edges, conflicts
}

// topological places every status after everything that beat it.
func topological(
	order []string,
	edges map[pair]bool,
	declaredBy map[string][]string,
	firstSeen map[string]int,
) []string {
	remaining := map[string]bool{}
	for _, s := range order {
		remaining[s] = true
	}

	result := make([]string, 0, len(order))
	for len(remaining) > 0 {
		// Ready statuses are those nothing remaining is required to precede.
		var ready []string
		for s := range remaining {
			blocked := false
			for other := range remaining {
				if other != s && edges[pair{other, s}] {
					blocked = true
					break
				}
			}
			if !blocked {
				ready = append(ready, s)
			}
		}

		// A cycle: the projects disagree in a loop and no order satisfies all
		// of them. Cut it with the same tie-break rather than refuse to draw a
		// board at all.
		if len(ready) == 0 {
			for s := range remaining {
				ready = append(ready, s)
			}
		}

		sort.Slice(ready, func(i, j int) bool {
			return precedes(ready[i], ready[j], declaredBy, firstSeen)
		})
		next := ready[0]
		result = append(result, next)
		delete(remaining, next)
	}
	return result
}

// precedes is the tie-break: more projects first, then earlier in registry
// order, then alphabetical. It is total, so the result never depends on map
// iteration order.
func precedes(a, b string, declaredBy map[string][]string, firstSeen map[string]int) bool {
	if len(declaredBy[a]) != len(declaredBy[b]) {
		return len(declaredBy[a]) > len(declaredBy[b])
	}
	if firstSeen[a] != firstSeen[b] {
		return firstSeen[a] < firstSeen[b]
	}
	return a < b
}

func contains(list []string, value string) bool {
	for _, v := range list {
		if v == value {
			return true
		}
	}
	return false
}
