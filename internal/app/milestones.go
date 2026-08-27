package app

import (
	"context"
	"strings"

	"github.com/FMakareev/muster-backlog/internal/backlog"
	"github.com/FMakareev/muster-backlog/internal/backlogcli"
)

// AddMilestone creates a milestone in a project.
func (s *BoardService) AddMilestone(projectPath, name, description string) CreateResult {
	if strings.TrimSpace(name) == "" {
		return CreateResult{Problem: &Problem{
			Kind: ProblemCLI, Title: "A milestone needs a name", Path: projectPath,
		}}
	}

	var id string
	result := s.write(projectPath, "The milestone could not be created",
		func(cli *backlogcli.Runner) error {
			created, err := cli.AddMilestone(context.Background(),
				s.dataDirFor(projectPath), name, description)
			id = created
			return err
		})
	if !result.OK {
		return CreateResult{Problem: result.Problem}
	}
	if id == "" || !s.milestoneLive(projectPath, id) {
		return CreateResult{Problem: &Problem{
			Kind:  ProblemCLI,
			Title: "Backlog.md reported success but wrote no milestone",
			Path:  projectPath,
		}}
	}
	return CreateResult{OK: true, TaskID: id}
}

// RenameMilestone renames a milestone and the tasks that point at it.
func (s *BoardService) RenameMilestone(projectPath, from, to string) WriteResult {
	if strings.TrimSpace(to) == "" {
		return WriteResult{Problem: &Problem{
			Kind: ProblemCLI, Title: "A milestone needs a name", Path: projectPath,
		}}
	}
	return s.write(projectPath, "The milestone could not be renamed",
		func(cli *backlogcli.Runner) error {
			return cli.RenameMilestone(context.Background(),
				s.dataDirFor(projectPath), from, to)
		})
}

// RetireMilestone takes a milestone out of the active list.
//
// Two commands behind one idea, and the difference is what happens to the
// tasks rather than to the file: both archive/milestones it. Keeping the tasks
// as they are is `milestone archive`; clearing or reassigning them is
// `milestone remove`. The caller says which, because both rewrite other files.
func (s *BoardService) RetireMilestone(
	projectPath, name, handling, reassignTo string,
) WriteResult {
	mode := backlogcli.TaskHandling(strings.TrimSpace(handling))
	if mode == "" {
		mode = backlogcli.HandlingKeep
	}

	if mode == backlogcli.HandlingReassign {
		if !s.milestoneNamed(projectPath, reassignTo) {
			return WriteResult{Problem: &Problem{
				Kind:  ProblemCLI,
				Title: "No such milestone to reassign to",
				Detail: "The tasks would have nowhere to go, so nothing was " +
					"changed.",
				Path: projectPath,
			}}
		}
	}

	return s.write(projectPath, "The milestone could not be retired",
		func(cli *backlogcli.Runner) error {
			if mode == backlogcli.HandlingKeep {
				// Leaving the tasks alone is what archive does, and it says so
				// in one word rather than through a flag.
				return cli.ArchiveMilestone(context.Background(),
					s.dataDirFor(projectPath), name)
			}
			return cli.RemoveMilestone(context.Background(),
				s.dataDirFor(projectPath), name, mode, reassignTo)
		})
}

// milestoneLive reports whether a project holds a milestone by id.
func (s *BoardService) milestoneLive(projectPath, id string) bool {
	for _, item := range s.store.Entities(backlog.KindMilestone) {
		if item.Ref.Project == projectPath && item.Ref.Class == backlog.ClassActive &&
			strings.EqualFold(item.Ref.ID, id) {
			return true
		}
	}
	return false
}

// milestoneNamed reports whether a project holds a milestone under an id or a
// title, which is how the CLI accepts one.
func (s *BoardService) milestoneNamed(projectPath, name string) bool {
	name = strings.TrimSpace(name)
	if name == "" {
		return false
	}
	for _, item := range s.store.Entities(backlog.KindMilestone) {
		if item.Ref.Project != projectPath || item.Ref.Class != backlog.ClassActive {
			continue
		}
		if strings.EqualFold(item.Ref.ID, name) ||
			strings.EqualFold(item.Entity.Title, name) {
			return true
		}
	}
	return false
}
