package app

import (
	"context"
	"fmt"
	"strings"

	"github.com/FMakareev/muster-backlog/internal/backlog"
	"github.com/FMakareev/muster-backlog/internal/backlogcli"
)

// CompleteTask moves a finished task into completed/.
//
// This is the ordinary end of a task rather than a destructive act, which is
// why it asks nothing: the file moves to where Backlog.md keeps finished work
// and the board stops carrying it. Across the author's nine projects, 591 of
// 875 live task files are already finished and none has ever been moved, which
// is what the absence of this button costs.
func (s *BoardService) CompleteTask(projectPath, taskID string) WriteResult {
	result := s.write(projectPath, fmt.Sprintf("%s could not be completed", taskID),
		func(cli *backlogcli.Runner) error {
			return cli.CompleteTask(context.Background(), s.dataDirFor(projectPath), taskID)
		})
	return s.confirmTaskLeft(result, projectPath, taskID, "completed")
}

// ArchiveTask moves a task into archive/tasks.
func (s *BoardService) ArchiveTask(projectPath, taskID string) WriteResult {
	result := s.write(projectPath, fmt.Sprintf("%s could not be archived", taskID),
		func(cli *backlogcli.Runner) error {
			return cli.ArchiveTask(context.Background(), s.dataDirFor(projectPath), taskID)
		})
	return s.confirmTaskLeft(result, projectPath, taskID, "archived")
}

// DemoteTask sends a task back to the drafts inbox.
func (s *BoardService) DemoteTask(projectPath, taskID string) WriteResult {
	result := s.write(projectPath, fmt.Sprintf("%s could not be sent back", taskID),
		func(cli *backlogcli.Runner) error {
			return cli.DemoteTask(context.Background(), s.dataDirFor(projectPath), taskID)
		})
	return s.confirmTaskLeft(result, projectPath, taskID, "sent back to the inbox")
}

// confirmTaskLeft checks that a task actually left the board.
//
// All three of these move a file, and all three are things whose outcome can
// be read back, so none of them is taken on the strength of an exit code. This
// CLI has reported success for a write that did nothing before.
func (s *BoardService) confirmTaskLeft(
	result WriteResult, projectPath, taskID, what string,
) WriteResult {
	if !result.OK || !s.taskLive(projectPath, taskID) {
		return result
	}
	return WriteResult{Problem: &Problem{
		Kind:   ProblemCLI,
		Title:  fmt.Sprintf("%s was not %s", taskID, what),
		Detail: "Backlog.md reported no error and the task is still on the board.",
		Path:   projectPath,
	}}
}

// taskLive reports whether a task is still in a project's tasks/ directory.
func (s *BoardService) taskLive(projectPath, taskID string) bool {
	for _, item := range s.store.Entities(backlog.KindTask) {
		if item.Ref.Class != backlog.ClassActive {
			continue
		}
		if item.Ref.Project == projectPath && strings.EqualFold(item.Ref.ID, taskID) {
			return true
		}
	}
	return false
}
