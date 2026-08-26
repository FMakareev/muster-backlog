package store

import (
	"sort"
	"strings"
	"time"

	"github.com/FMakareev/muster-backlog/internal/backlog"
)

// Count is one label and how many things carry it.
type Count struct {
	Label string
	Total int
}

// Blocked is a task waiting on a dependency that is not finished.
type Blocked struct {
	Item Item
	// On are the dependencies still outstanding, as ids.
	On []string
}

// Analytics is the cross-project overview.
//
// The vocabulary is the one `backlog overview` already established from purely
// native data - status and priority breakdowns, average age, stale tasks,
// blocked tasks - because inventing a second one would only mean two ways to
// describe the same backlog.
type Analytics struct {
	// Project is empty for the total across every project.
	Project     string
	ProjectName string

	Tasks    int
	Statuses []Count
	Priority []Count
	Types    []Count

	// Unprioritised counts tasks with no priority at all, which was the
	// symptom that started this project: more than half the backlog
	// indistinguishable item from item.
	Unprioritised int

	// AverageAgeDays is over tasks that are not in a terminal status.
	AverageAgeDays float64
	// Stale are open tasks untouched for longer than the threshold.
	Stale []Item
	// Blocked are open tasks with unfinished dependencies.
	Blocked []Blocked
}

// AnalyticsOptions tune what counts as stale.
type AnalyticsOptions struct {
	// StaleAfter is how long an open task may go untouched. Zero means 30 days.
	StaleAfter time.Duration
	// Now is the reference time; zero means time.Now.
	Now time.Time
}

// Analytics computes the overview per project and in total.
//
// The per-project entries come first in registry order, and the total is last.
func (s *Store) Analytics(opts AnalyticsOptions) []Analytics {
	if opts.StaleAfter <= 0 {
		opts.StaleAfter = 30 * 24 * time.Hour
	}
	if opts.Now.IsZero() {
		opts.Now = time.Now()
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	var out []Analytics
	total := Analytics{ProjectName: "All projects"}
	statuses, priorities, types := map[string]int{}, map[string]int{}, map[string]int{}
	var totalAge float64
	var aged int

	for _, path := range s.order {
		state, ok := s.projects[path]
		if !ok || !state.OK() {
			continue
		}

		// "Finished" is the last status a project declares: the only
		// definition the format offers.
		var terminal string
		if list := state.Scanned.Config.Statuses; len(list) > 0 {
			terminal = list[len(list)-1]
		}

		a := Analytics{Project: path, ProjectName: state.Registry.DisplayName}
		ps, pp, pt := map[string]int{}, map[string]int{}, map[string]int{}
		var age float64
		var counted int

		done := map[string]bool{}
		for _, task := range state.Scanned.Tasks {
			if task.Class == backlog.ClassActive &&
				terminal != "" && strings.EqualFold(task.Status, terminal) {
				done[strings.ToUpper(task.ID)] = true
			}
		}

		for _, task := range state.Scanned.Tasks {
			if task.Class != backlog.ClassActive {
				continue
			}
			a.Tasks++
			total.Tasks++
			bump(ps, task.Status)
			bump(statuses, task.Status)
			bump(pt, task.Type)
			bump(types, task.Type)

			if strings.TrimSpace(task.Priority) == "" {
				a.Unprioritised++
				total.Unprioritised++
			} else {
				bump(pp, strings.ToLower(task.Priority))
				bump(priorities, strings.ToLower(task.Priority))
			}

			finished := terminal != "" && strings.EqualFold(task.Status, terminal)
			if finished {
				continue
			}

			if !task.Created.IsZero() {
				days := opts.Now.Sub(task.Created).Hours() / 24
				age += days
				counted++
				totalAge += days
				aged++
			}

			touched := task.Updated
			if touched.IsZero() {
				touched = task.Created
			}
			if !touched.IsZero() && opts.Now.Sub(touched) > opts.StaleAfter {
				it := item(state, task)
				a.Stale = append(a.Stale, it)
				total.Stale = append(total.Stale, it)
			}

			var outstanding []string
			for _, dep := range task.Dependencies {
				if !done[strings.ToUpper(strings.TrimSpace(dep))] {
					outstanding = append(outstanding, dep)
				}
			}
			if len(outstanding) > 0 {
				b := Blocked{Item: item(state, task), On: outstanding}
				a.Blocked = append(a.Blocked, b)
				total.Blocked = append(total.Blocked, b)
			}
		}

		if counted > 0 {
			a.AverageAgeDays = age / float64(counted)
		}
		a.Statuses, a.Priority, a.Types = sortCounts(ps), sortCounts(pp), sortCounts(pt)
		out = append(out, a)
	}

	if aged > 0 {
		total.AverageAgeDays = totalAge / float64(aged)
	}
	total.Statuses, total.Priority, total.Types =
		sortCounts(statuses), sortCounts(priorities), sortCounts(types)
	return append(out, total)
}

func bump(counts map[string]int, label string) {
	if label = strings.TrimSpace(label); label != "" {
		counts[label]++
	}
}

// sortCounts orders by count then by label, so the same data always renders
// the same way.
func sortCounts(counts map[string]int) []Count {
	out := make([]Count, 0, len(counts))
	for label, total := range counts {
		out = append(out, Count{Label: label, Total: total})
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Total != out[j].Total {
			return out[i].Total > out[j].Total
		}
		return out[i].Label < out[j].Label
	})
	return out
}
