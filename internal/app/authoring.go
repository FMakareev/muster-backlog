package app

import (
	"context"
	"strings"

	"github.com/FMakareev/muster-backlog/internal/backlog"
	"github.com/FMakareev/muster-backlog/internal/backlogcli"
)

// NewDocumentInput is a document as the form fills it in.
//
// Content is here even though `doc create` cannot take it: a document created
// empty and then filled is one act to the person doing it, and making them
// press two buttons for it would be an implementation detail leaking upwards.
type NewDocumentInput struct {
	Project string `json:"project"`
	Title   string `json:"title"`
	Type    string `json:"type"`
	Path    string `json:"path"`
	Content string `json:"content"`
}

// DocumentTypes are what the CLI accepts. Read from its help rather than
// guessed, and offered as a list so a typo cannot reach it.
func (s *BoardService) DocumentTypes() []string {
	return []string{"readme", "guide", "specification", "other"}
}

// CreateDocument writes a document, with its body if one was given.
func (s *BoardService) CreateDocument(input NewDocumentInput) CreateResult {
	if strings.TrimSpace(input.Title) == "" {
		return CreateResult{Problem: &Problem{
			Kind: ProblemCLI, Title: "A document needs a title",
			Path: input.Project,
		}}
	}

	var id string
	result := s.write(input.Project, "The document could not be created",
		func(cli *backlogcli.Runner) error {
			created, err := cli.CreateDocument(context.Background(),
				s.dataDirFor(input.Project), backlogcli.NewDocument{
					Title: input.Title, Type: input.Type, Path: input.Path,
				})
			id = created
			if err != nil {
				return err
			}
			// The body is a second call because the command takes no content.
			if strings.TrimSpace(input.Content) == "" {
				return nil
			}
			return cli.UpdateDocument(context.Background(),
				s.dataDirFor(input.Project), created,
				backlogcli.DocumentEdit{Content: input.Content, SetContent: true})
		})
	if !result.OK {
		return CreateResult{Problem: result.Problem}
	}
	if id == "" || !s.entityExists(input.Project, backlog.KindDocument, id) {
		return CreateResult{Problem: &Problem{
			Kind:  ProblemCLI,
			Title: "Backlog.md reported success but wrote no document",
			Path:  input.Project,
		}}
	}
	return CreateResult{OK: true, TaskID: id}
}

// DocumentUpdate is what the viewer can change about a document.
type DocumentUpdate struct {
	Title string `json:"title"`
	Type  string `json:"type"`
	// Content replaces the whole body. `doc update --content` takes nothing
	// smaller, so the editor sends the entire document back.
	Content string `json:"content"`
}

// UpdateDocument rewrites a document.
func (s *BoardService) UpdateDocument(projectPath, docID string, edit DocumentUpdate) WriteResult {
	return s.write(projectPath, "The document could not be saved",
		func(cli *backlogcli.Runner) error {
			return cli.UpdateDocument(context.Background(),
				s.dataDirFor(projectPath), docID, backlogcli.DocumentEdit{
					Title:      edit.Title,
					Type:       edit.Type,
					Content:    edit.Content,
					SetContent: true,
				})
		})
}

// CreateDecision writes a decision.
//
// Only its title and status: the CLI writes a skeleton with Context, Decision
// and Consequences headings and has no command that fills them. What the
// interface can do is make the act cheap and then say where the file is.
func (s *BoardService) CreateDecision(projectPath, title, status string) CreateResult {
	if strings.TrimSpace(title) == "" {
		return CreateResult{Problem: &Problem{
			Kind: ProblemCLI, Title: "A decision needs a title", Path: projectPath,
		}}
	}

	var id string
	result := s.write(projectPath, "The decision could not be created",
		func(cli *backlogcli.Runner) error {
			created, err := cli.CreateDecision(context.Background(),
				s.dataDirFor(projectPath), title, status)
			id = created
			return err
		})
	if !result.OK {
		return CreateResult{Problem: result.Problem}
	}
	if id == "" || !s.entityExists(projectPath, backlog.KindDecision, id) {
		return CreateResult{Problem: &Problem{
			Kind:  ProblemCLI,
			Title: "Backlog.md reported success but wrote no decision",
			Path:  projectPath,
		}}
	}
	return CreateResult{OK: true, TaskID: id}
}

// entityExists reports whether a project holds an entity, after a reload.
func (s *BoardService) entityExists(projectPath string, kind backlog.Kind, id string) bool {
	for _, item := range s.store.Entities(kind) {
		if item.Ref.Project == projectPath && strings.EqualFold(item.Ref.ID, id) {
			return true
		}
	}
	return false
}
