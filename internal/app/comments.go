package app

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/FMakareev/muster-backlog/internal/backlogcli"
)

// Comments are the one part of a task that is a conversation rather than a
// field, and Muster could read them and not write one.

// AddComment appends a comment to a task, signed with whoever is using the
// application.
func (s *BoardService) AddComment(projectPath, taskID, text string) WriteResult {
	if strings.TrimSpace(text) == "" {
		return WriteResult{Problem: &Problem{
			Kind: ProblemCLI, Title: "A comment needs something in it",
			Path: projectPath,
		}}
	}
	author := s.CommentAuthor(projectPath)
	return s.editTask(projectPath, taskID,
		fmt.Sprintf("the comment on %s could not be added", taskID),
		func(cli *backlogcli.Runner) error {
			return cli.AddComment(context.Background(),
				s.dataDirFor(projectPath), taskID, text, author)
		})
}

// CommentAuthor is the name a comment will be signed with.
//
// The preference first, because someone who set one meant it. Then whatever
// git answers for that folder - the repository's identity if it has one, the
// global one otherwise, which is usually still the right person. That is where
// a name already is, and the files already agree with it: the assignees in the
// author's own projects are the name git reports. Then nothing - a comment is
// written unsigned rather than signed with something invented, which is a
// state the format has and the CLI produces when asked for no author.
//
// Exposed so the interface can show which name will be used before it is used.
func (s *BoardService) CommentAuthor(projectPath string) string {
	s.mu.Lock()
	author := strings.TrimSpace(s.prefs.Author)
	s.mu.Unlock()
	if author != "" {
		return handle(author)
	}
	if name := gitName(projectPath); name != "" {
		return handle(name)
	}
	return ""
}

// handle writes a name the way these files write one.
func handle(name string) string {
	name = strings.TrimSpace(name)
	if name == "" || strings.HasPrefix(name, "@") {
		return name
	}
	return "@" + name
}

// gitNames caches one answer per project.
//
// Asking git is a subprocess, and a panel that shows the author under every
// comment would otherwise start one every time a task is opened. Cached for
// the life of the process: a person renaming themselves mid-session is not
// worth a process per render.
var (
	gitNamesMu sync.Mutex
	gitNames   = map[string]string{}
)

// gitName asks git who the person is, for a folder.
//
// Reading, not writing: this asks git who the person is, which git already
// knows, rather than adding a second place to keep it. A folder that is not a
// repository, or a git that is not installed, simply answers nothing.
func gitName(projectPath string) string {
	gitNamesMu.Lock()
	if name, ok := gitNames[projectPath]; ok {
		gitNamesMu.Unlock()
		return name
	}
	gitNamesMu.Unlock()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "git", "-C", projectPath, "config", "user.name")
	out, err := cmd.Output()
	name := ""
	if err == nil {
		name = strings.TrimSpace(string(out))
	}

	gitNamesMu.Lock()
	gitNames[projectPath] = name
	gitNamesMu.Unlock()
	return name
}
