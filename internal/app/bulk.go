package app

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/FMakareev/muster-backlog/internal/backlogcli"
)

// EventBulkProgress reports how far a bulk change has got.
//
// One `task edit` takes about 280ms, so forty tasks is eleven seconds of a
// button saying nothing. The count is the only honest thing to show: the work
// is a queue of separate writes, and how many of them are done is exactly
// what someone waiting wants to know.
const EventBulkProgress = "muster:bulk:progress"

// BulkProgress is the payload of EventBulkProgress.
type BulkProgress struct {
	Done  int `json:"done"`
	Total int `json:"total"`
}

// BulkTarget names one task to change.
type BulkTarget struct {
	Project string `json:"project"`
	TaskID  string `json:"taskId"`
}

// BulkChange is a set of tasks and one change to make to all of them.
//
// The three single-valued fields are pointers because leaving a field alone
// and clearing it are different instructions, and a bulk form has to be able
// to say both. Null means the form said nothing about it.
type BulkChange struct {
	Tasks     []BulkTarget `json:"tasks"`
	Status    *string      `json:"status"`
	Priority  *string      `json:"priority"`
	Milestone *string      `json:"milestone"`

	AddLabels    []string `json:"addLabels"`
	RemoveLabels []string `json:"removeLabels"`
}

// BulkFailure is one task that did not take the change, in the CLI's own words.
type BulkFailure struct {
	Project     string `json:"project"`
	ProjectName string `json:"projectName"`
	TaskID      string `json:"taskId"`
	Detail      string `json:"detail"`
}

// BulkResult is what happened, task by task.
//
// A run of twenty writes can partly fail, and there is no transaction to hide
// behind: this reports the count that landed and names every one that did not,
// because a summary that only said "done" would be wrong more often than it
// was useful.
type BulkResult struct {
	// Changed is how many tasks took the change.
	Changed int `json:"changed"`
	// Failures names the ones that did not, with the reason each gave.
	Failures []BulkFailure `json:"failures"`
	// Problem is set when nothing was attempted at all.
	Problem *Problem `json:"problem"`
}

// ChangeMany applies one change to many tasks.
//
// Every write still goes through the CLI, one task at a time, because that is
// the only writer there is. What is bulk about it is the choosing: the set is
// picked once and the change described once. Each task gets a single `task
// edit` carrying the whole change, which is also what makes a task either take
// all of it or none of it - the CLI validates every field before writing.
func (s *BoardService) ChangeMany(change BulkChange) BulkResult {
	if len(change.Tasks) == 0 {
		return BulkResult{Problem: &Problem{
			Kind:   ProblemCLI,
			Title:  "Nothing was selected",
			Detail: "Choose the tasks to change first.",
		}}
	}

	edit := backlogcli.TaskChange{
		Status:       change.Status,
		Priority:     change.Priority,
		Milestone:    change.Milestone,
		AddLabels:    trimmed(change.AddLabels),
		RemoveLabels: trimmed(change.RemoveLabels),
	}
	if edit.Empty() {
		return BulkResult{Problem: &Problem{
			Kind:   ProblemCLI,
			Title:  "Nothing was asked for",
			Detail: "Choose at least one thing to change.",
		}}
	}

	if p := s.milestoneRefusal(change); p != nil {
		return BulkResult{Problem: p}
	}

	s.mu.Lock()
	cli := s.cli
	s.mu.Unlock()
	if cli == nil {
		if p := s.cliProblem(); p != nil {
			return BulkResult{Problem: p}
		}
	}

	result := BulkResult{Failures: []BulkFailure{}}
	touched := map[string]bool{}
	total := len(change.Tasks)

	for i, target := range change.Tasks {
		emit(EventBulkProgress, BulkProgress{Done: i, Total: total})

		if err := s.changeOne(cli, target, edit); err != nil {
			result.Failures = append(result.Failures, BulkFailure{
				Project:     target.Project,
				ProjectName: s.projectName(target.Project),
				TaskID:      target.TaskID,
				Detail:      err.Error(),
			})
		} else {
			result.Changed++
		}
		touched[target.Project] = true
	}

	// One rescan per project rather than one per task. The files are the
	// answer either way, and re-reading a project twenty times to learn the
	// same thing would be most of the run.
	for _, project := range sortedKeys(touched) {
		if s.store.Reload(project) {
			emit(EventProjectChanged, ProjectChanged{Project: project})
		}
	}
	emit(EventBulkProgress, BulkProgress{Done: total, Total: total})

	return result
}

// changeOne runs the whole change against a single task.
func (s *BoardService) changeOne(
	cli *backlogcli.Runner, target BulkTarget, edit backlogcli.TaskChange,
) error {
	if s.draftPresent(target.Project, target.TaskID) {
		return fmt.Errorf("this is a note, not a task; Backlog.md cannot edit a draft")
	}
	return cli.EditTask(context.Background(), s.dataDirFor(target.Project), target.TaskID, edit)
}

// milestoneRefusal blocks a milestone change across more than one project.
//
// Not a matter of taste. A milestone id belongs to the project it was created
// in, and `task edit -m` does not check: giving a task an id that project does
// not have succeeds, writes the id into the file, and the milestone list then
// shows an entry with no file behind it. Applied across a selection, that
// plants a dangling reference in every project but one, silently, and every
// write reports success. So the whole run is refused before any of it starts
// rather than half of it being wrong.
//
// Everything else in a change carries across projects perfectly well, so the
// refusal is of the milestone and not of the selection.
func (s *BoardService) milestoneRefusal(change BulkChange) *Problem {
	if change.Milestone == nil {
		return nil
	}
	projects := map[string]bool{}
	for _, t := range change.Tasks {
		projects[t.Project] = true
	}
	if len(projects) < 2 {
		return nil
	}

	names := make([]string, 0, len(projects))
	for _, path := range sortedKeys(projects) {
		names = append(names, s.projectName(path))
	}
	return &Problem{
		Kind:  ProblemCLI,
		Title: "A milestone cannot be set across projects",
		Detail: fmt.Sprintf(
			"These tasks are in %s. Milestone ids belong to one project, and "+
				"Backlog.md writes an id it does not recognise without complaining "+
				"— which would leave the other projects pointing at a milestone "+
				"that does not exist. Change the milestone one project at a time. "+
				"Status, priority and labels can be changed across all of them.",
			english(names)),
	}
}

// projectName is the registered name for a path, falling back to the path.
func (s *BoardService) projectName(path string) string {
	for _, p := range s.store.Projects() {
		if p.Registry.Path == path {
			return p.Registry.DisplayName
		}
	}
	return path
}

// trimmed drops blanks and surrounding space from a list of values.
func trimmed(values []string) []string {
	out := make([]string, 0, len(values))
	for _, v := range values {
		if v = strings.TrimSpace(v); v != "" {
			out = append(out, v)
		}
	}
	return out
}

func sortedKeys(set map[string]bool) []string {
	out := make([]string, 0, len(set))
	for k := range set {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// english joins names the way a sentence does, because this one is read as a
// sentence rather than scanned as a list.
func english(names []string) string {
	switch len(names) {
	case 0:
		return "more than one project"
	case 1:
		return names[0]
	case 2:
		return names[0] + " and " + names[1]
	default:
		return strings.Join(names[:len(names)-1], ", ") + " and " + names[len(names)-1]
	}
}
