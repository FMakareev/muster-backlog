package mcpserver

import (
	"context"
	"fmt"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/FMakareev/muster-backlog/internal/backlogcli"
	"github.com/FMakareev/muster-backlog/internal/store"
)

// notFound explains an id that does not resolve, naming the project, since ids
// are unique only inside one.
func notFound(ref entityRef, project string) error {
	return fmt.Errorf("%s has no %q. Ids are unique only inside a project",
		project, ref.ID)
}

// badKind lists what list_entities takes.
func badKind(given string) error {
	return fmt.Errorf("%q is not a kind. Use document, decision, draft or milestone",
		given)
}

// Written is what every write answers with: what changed, and where to look.
type Written struct {
	Project string `json:"project"`
	ID      string `json:"id,omitempty"`
	Did     string `json:"did"`
}

// runner returns the CLI, or the reason there is none.
//
// Writing is the half of this server that can be unavailable, and it says so
// at the moment it is asked rather than refusing to start.
func (s *Server) runner() (*backlogcli.Runner, error) {
	if s.cli != nil {
		return s.cli, nil
	}
	// Worth one retry: the person may have installed it since the agent
	// connected, and a long-lived session should not stay broken for that.
	if cli, err := backlogcli.New(); err == nil {
		s.cli, s.cliErr = cli, nil
		return cli, nil
	}
	return nil, fmt.Errorf("the Backlog.md CLI is not available, so nothing "+
		"can be written: %v", s.cliErr)
}

// write runs one mutation against a project and re-reads it afterwards.
//
// The same rule the interface follows: nothing assumes the CLI did what was
// asked. The store is refreshed from the files, so the next read is what is on
// disk rather than what was intended.
func (s *Server) write(
	name string, did string, do func(*backlogcli.Runner, store.ProjectState) (string, error),
) (*mcp.CallToolResult, Written, error) {
	s.Reload()
	p, err := s.project(name)
	if err != nil {
		return fail[Written](err)
	}
	cli, err := s.runner()
	if err != nil {
		return fail[Written](err)
	}

	id, err := do(cli, p)
	s.store.Reload(p.Registry.Path)
	if err != nil {
		return fail[Written](err)
	}
	return ok(Written{Project: p.Registry.DisplayName, ID: id, Did: did})
}

// writes marks a tool as changing something.
//
// destructive is for the ones that overwrite what was there: replacing a
// section throws away the text it replaces, and setting a field forgets the
// old value. Creating a task adds without taking anything away.
//
// None of them is idempotent in the sense the hint means: calling create_task
// twice makes two tasks.
func writes(destructive bool) *mcp.ToolAnnotations {
	no := false
	return &mcp.ToolAnnotations{
		ReadOnlyHint:    false,
		DestructiveHint: &destructive,
		IdempotentHint:  false,
		OpenWorldHint:   &no,
	}
}

// insideAProject is the sentence every write carries.
//
// A tool description is read on its own, out of the order they were written
// in, so each one has to say where it does and does not belong rather than
// relying on the server's instructions having been read first.
const insideAProject = "\n\nUse this to change something in a project you " +
	"are not working inside. If you are in the project already, its own " +
	"backlog CLI is the authority: it has the whole command surface, and " +
	"this offers a small part of it."

func (s *Server) addWriteTools(server *mcp.Server) {
	type createIn struct {
		Project            string   `json:"project" jsonschema:"the registered project to create it in"`
		Title              string   `json:"title" jsonschema:"what the task is called"`
		Description        string   `json:"description,omitempty"`
		Status             string   `json:"status,omitempty" jsonschema:"one the project declares; its first by default"`
		Priority           string   `json:"priority,omitempty"`
		Type               string   `json:"type,omitempty"`
		Milestone          string   `json:"milestone,omitempty" jsonschema:"a milestone id or title in the same project"`
		Assignee           string   `json:"assignee,omitempty"`
		Labels             []string `json:"labels,omitempty"`
		AcceptanceCriteria []string `json:"acceptanceCriteria,omitempty"`
		Parent             string   `json:"parent,omitempty" jsonschema:"a task id in the same project, making this a subtask"`
		DependsOn          []string `json:"dependsOn,omitempty" jsonschema:"task ids in the same project"`
		Draft              bool     `json:"draft,omitempty" jsonschema:"capture it into the inbox instead of onto the board"`
	}
	mcp.AddTool(server, &mcp.Tool{
		Annotations: writes(false),
		Name:        "create_task",
		Description: "Write a new task into a project, or a note into its " +
			"inbox. Everything goes through the Backlog.md CLI, so the file " +
			"is whatever that would have written." + insideAProject,
	}, func(_ context.Context, _ *mcp.CallToolRequest, in createIn) (*mcp.CallToolResult, Written, error) {
		did := "created a task"
		if in.Draft {
			did = "captured a note"
		}
		return s.write(in.Project, did, func(cli *backlogcli.Runner, p store.ProjectState) (string, error) {
			return cli.CreateTask(context.Background(), p.Registry.Location.DataDir,
				backlogcli.NewTask{
					Title:              in.Title,
					Description:        in.Description,
					Status:             in.Status,
					Priority:           in.Priority,
					Type:               in.Type,
					Milestone:          in.Milestone,
					Assignee:           in.Assignee,
					Labels:             in.Labels,
					AcceptanceCriteria: in.AcceptanceCriteria,
					Parent:             in.Parent,
					DependsOn:          in.DependsOn,
					Draft:              in.Draft,
				})
		})
	})

	type setIn struct {
		entityRef
		Field string `json:"field" jsonschema:"one of status, priority, assignee or milestone"`
		Value string `json:"value" jsonschema:"the new value; empty clears the field where the CLI allows it"`
	}
	mcp.AddTool(server, &mcp.Tool{
		Annotations: writes(true),
		Name:        "set_field",
		Description: "Change one field of one task: status, priority, " +
			"assignee or milestone. A status must be one the project declares " +
			"- they differ between projects and the CLI rejects anything " +
			"else. The old value is not kept." + insideAProject,
	}, func(_ context.Context, _ *mcp.CallToolRequest, in setIn) (*mcp.CallToolResult, Written, error) {
		field := strings.ToLower(strings.TrimSpace(in.Field))
		return s.write(in.Project, "set "+field, func(cli *backlogcli.Runner, p store.ProjectState) (string, error) {
			dir := p.Registry.Location.DataDir
			ctx := context.Background()
			switch field {
			case "status":
				return in.ID, cli.SetStatus(ctx, dir, in.ID, in.Value)
			case "priority":
				return in.ID, cli.SetPriority(ctx, dir, in.ID, in.Value)
			case "assignee":
				return in.ID, cli.SetAssignee(ctx, dir, in.ID, in.Value)
			case "milestone":
				return in.ID, cli.SetMilestone(ctx, dir, in.ID, in.Value)
			default:
				return "", fmt.Errorf(
					"%q is not a field this can set. Use status, priority, assignee or milestone",
					in.Field)
			}
		})
	})

	type labelIn struct {
		entityRef
		Label  string `json:"label" jsonschema:"the label to add or remove"`
		Remove bool   `json:"remove,omitempty" jsonschema:"remove it instead of adding it"`
	}
	mcp.AddTool(server, &mcp.Tool{
		Annotations: writes(false),
		Name:        "set_label",
		Description: "Add a label to a task, or take one off. Labels are " +
			"free-form and shared across a project, so list_projects and " +
			"list_tasks are worth reading before inventing a new one." +
			insideAProject,
	}, func(_ context.Context, _ *mcp.CallToolRequest, in labelIn) (*mcp.CallToolResult, Written, error) {
		did := "added a label"
		if in.Remove {
			did = "removed a label"
		}
		return s.write(in.Project, did, func(cli *backlogcli.Runner, p store.ProjectState) (string, error) {
			dir := p.Registry.Location.DataDir
			if in.Remove {
				return in.ID, cli.RemoveLabel(context.Background(), dir, in.ID, in.Label)
			}
			return in.ID, cli.AddLabel(context.Background(), dir, in.ID, in.Label)
		})
	})

	type sectionIn struct {
		entityRef
		Section string `json:"section" jsonschema:"one of description, plan or notes"`
		Text    string `json:"text" jsonschema:"the whole section; it is replaced, not appended to"`
	}
	mcp.AddTool(server, &mcp.Tool{
		Annotations: writes(true),
		Name:        "set_section",
		Description: "Replace a task's description, implementation plan or " +
			"implementation notes. The whole section is replaced, because " +
			"that is what the CLI takes - read it with get_task first, or " +
			"whatever was there is gone." + insideAProject,
	}, func(_ context.Context, _ *mcp.CallToolRequest, in sectionIn) (*mcp.CallToolResult, Written, error) {
		section := strings.ToLower(strings.TrimSpace(in.Section))
		return s.write(in.Project, "set the "+section, func(cli *backlogcli.Runner, p store.ProjectState) (string, error) {
			dir := p.Registry.Location.DataDir
			ctx := context.Background()
			switch section {
			case "description":
				return in.ID, cli.SetDescription(ctx, dir, in.ID, in.Text)
			case "plan":
				return in.ID, cli.SetPlan(ctx, dir, in.ID, in.Text)
			case "notes":
				return in.ID, cli.SetNotes(ctx, dir, in.ID, in.Text)
			default:
				return "", fmt.Errorf(
					"%q is not a section this can set. Use description, plan or notes",
					in.Section)
			}
		})
	})
}
