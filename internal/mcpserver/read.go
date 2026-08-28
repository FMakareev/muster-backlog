package mcpserver

import (
	"context"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/FMakareev/muster-backlog/internal/backlog"
	"github.com/FMakareev/muster-backlog/internal/store"
)

// ProjectSummary is one registered project as an agent needs it: enough to
// decide where to look next without asking again.
type ProjectSummary struct {
	Name       string         `json:"name"`
	Path       string         `json:"path"`
	OK         bool           `json:"ok"`
	Problem    string         `json:"problem,omitempty"`
	Hidden     bool           `json:"hidden,omitempty"`
	Tasks      int            `json:"tasks"`
	Drafts     int            `json:"drafts"`
	ByStatus   map[string]int `json:"byStatus,omitempty"`
	Statuses   []string       `json:"statuses,omitempty"`
	Priorities []string       `json:"priorities,omitempty"`
	Types      []string       `json:"types,omitempty"`
}

// TaskSummary is a task without its body, for lists.
type TaskSummary struct {
	Project      string   `json:"project"`
	ID           string   `json:"id"`
	Title        string   `json:"title"`
	Status       string   `json:"status"`
	Priority     string   `json:"priority,omitempty"`
	Type         string   `json:"type,omitempty"`
	Milestone    string   `json:"milestone,omitempty"`
	Labels       []string `json:"labels,omitempty"`
	Assignee     []string `json:"assignee,omitempty"`
	Parent       string   `json:"parent,omitempty"`
	Dependencies []string `json:"dependencies,omitempty"`
	Class        string   `json:"class"`
	Updated      string   `json:"updated,omitempty"`
}

// TaskDetail is one task in full, which is the thing worth a second call.
type TaskDetail struct {
	TaskSummary
	Description        string   `json:"description,omitempty"`
	Plan               string   `json:"plan,omitempty"`
	Notes              string   `json:"notes,omitempty"`
	FinalSummary       string   `json:"finalSummary,omitempty"`
	AcceptanceCriteria []string `json:"acceptanceCriteria,omitempty"`
	References         []string `json:"references,omitempty"`
	Documentation      []string `json:"documentation,omitempty"`
	Path               string   `json:"path"`
}

func summarise(item store.Item) TaskSummary {
	return TaskSummary{
		Project:      item.ProjectName,
		ID:           item.Entity.ID,
		Title:        item.Entity.Title,
		Status:       item.Entity.Status,
		Priority:     item.Entity.Priority,
		Type:         item.Entity.Type,
		Milestone:    item.Entity.Milestone,
		Labels:       item.Entity.Labels,
		Assignee:     item.Entity.Assignee,
		Parent:       item.Entity.ParentTaskID,
		Dependencies: item.Entity.Dependencies,
		Class:        string(item.Ref.Class),
		Updated:      shortDate(item.Entity.Updated.String()),
	}
}

// shortDate drops the zero time, which reads as a date somebody meant.
func shortDate(value string) string {
	if value == "" || strings.HasPrefix(value, "0001-01-01") {
		return ""
	}
	return value[:min(10, len(value))]
}

// reads marks a tool as answering without changing anything, so a client can
// stop asking permission for questions and keep asking for changes.
func reads() *mcp.ToolAnnotations {
	no := false
	return &mcp.ToolAnnotations{
		ReadOnlyHint:   true,
		IdempotentHint: true,
		// The answer comes from files on this machine, not from anywhere the
		// client would have to reason about.
		OpenWorldHint: &no,
	}
}

// Every listing answers with an object around its list, because the protocol
// says structuredContent is a JSON object and a client that validates the
// result rejects a bare array outright — the whole response, before the agent
// sees any of it. Five tools did exactly that: the data was read correctly and
// thrown away at the client.
//
// One named field each rather than a shared envelope, so the name says what
// the list holds when an agent reads the schema.

type projectList struct {
	Projects []ProjectSummary `json:"projects" jsonschema:"every registered project"`
}

type taskList struct {
	Tasks []TaskSummary `json:"tasks" jsonschema:"the tasks that matched"`
}

func (s *Server) addReadTools(server *mcp.Server) {
	mcp.AddTool(server, &mcp.Tool{
		Annotations: reads(),
		Name:        "list_projects",
		Description: "Every registered Backlog.md project, with how many tasks " +
			"it holds, how they are spread across its statuses, and the " +
			"statuses, priorities and types it declares. Start here: ids and " +
			"statuses are per-project and nothing else in this server makes " +
			"sense without knowing which projects exist.",
	}, func(_ context.Context, _ *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, projectList, error) {
		s.Reload()
		// The one tool an agent starts with, and the one place a missing
		// registry has to be said out loud rather than answered with an empty
		// list that reads as "you have no projects".
		if s.missing {
			return nil, projectList{}, s.noRegistry()
		}
		var out []ProjectSummary
		for _, p := range s.store.Projects() {
			summary := ProjectSummary{
				Name:   p.Registry.DisplayName,
				Path:   p.Registry.Path,
				OK:     p.OK(),
				Hidden: p.Registry.Hidden,
			}
			if p.Err != nil {
				summary.Problem = p.Err.Error()
			}
			if p.OK() {
				summary.ByStatus = s.store.CountByStatus(p.Registry.Path)
				for _, count := range summary.ByStatus {
					summary.Tasks += count
				}
				summary.Drafts = len(p.Scanned.Drafts)
				summary.Statuses = p.Scanned.Config.Statuses
				summary.Priorities = p.Scanned.Config.Priorities
				summary.Types = p.Scanned.Config.Types
			}
			out = append(out, summary)
		}
		return ok(projectList{Projects: out})
	})

	type listTasksIn struct {
		Project         string   `json:"project,omitempty" jsonschema:"limit to one project by name or path; omit for every project"`
		Statuses        []string `json:"statuses,omitempty" jsonschema:"keep only these statuses"`
		Priorities      []string `json:"priorities,omitempty" jsonschema:"keep only these priorities"`
		Types           []string `json:"types,omitempty" jsonschema:"keep only these task types"`
		Milestones      []string `json:"milestones,omitempty" jsonschema:"keep only tasks in these milestones, by id or title"`
		Labels          []string `json:"labels,omitempty" jsonschema:"keep tasks carrying any of these labels"`
		Assignees       []string `json:"assignees,omitempty" jsonschema:"keep tasks assigned to any of these"`
		Text            string   `json:"text,omitempty" jsonschema:"keep tasks whose title or description contains this"`
		IncludeArchived bool     `json:"includeArchived,omitempty" jsonschema:"include completed and archived tasks; they are left out by default"`
		Limit           int      `json:"limit,omitempty" jsonschema:"stop after this many; 0 means all of them"`
	}
	mcp.AddTool(server, &mcp.Tool{
		Annotations: reads(),
		Name:        "list_tasks",
		Description: "Tasks across every registered project, or one of them, " +
			"with the same filters the board has. Archived and completed " +
			"tasks are left out unless asked for. Bodies are not included: " +
			"use get_task for one.\n\n" +
			"This is the tool for looking beyond the project you are in. " +
			"Inside one, its own backlog CLI lists tasks faster and knows " +
			"more.",
	}, func(_ context.Context, _ *mcp.CallToolRequest, in listTasksIn) (*mcp.CallToolResult, taskList, error) {
		s.Reload()
		var projects []string
		if strings.TrimSpace(in.Project) != "" {
			p, err := s.project(in.Project)
			if err != nil {
				return fail[taskList](err)
			}
			projects = []string{p.Registry.Path}
		}

		items := s.store.Query(store.Query{
			Projects:   projects,
			Classes:    classOf(in.IncludeArchived),
			Statuses:   in.Statuses,
			Priorities: in.Priorities,
			Types:      in.Types,
			Milestones: in.Milestones,
			Labels:     in.Labels,
			Assignees:  in.Assignees,
			Text:       in.Text,
		})

		out := make([]TaskSummary, 0, len(items))
		for _, item := range items {
			if in.Limit > 0 && len(out) >= in.Limit {
				break
			}
			out = append(out, summarise(item))
		}
		return ok(taskList{Tasks: out})
	})

	mcp.AddTool(server, &mcp.Tool{
		Annotations: reads(),
		Name:        "get_task",
		Description: "One task in full: description, acceptance criteria, " +
			"implementation plan, notes, final summary, references and " +
			"documentation. Ids are unique only inside a project, so the " +
			"project is required.",
	}, func(_ context.Context, _ *mcp.CallToolRequest, in entityRef) (*mcp.CallToolResult, TaskDetail, error) {
		s.Reload()
		p, err := s.project(in.Project)
		if err != nil {
			return fail[TaskDetail](err)
		}
		for _, kind := range []backlog.Kind{backlog.KindTask, backlog.KindDraft} {
			for _, item := range s.store.Entities(kind) {
				if item.Ref.Project != p.Registry.Path ||
					!strings.EqualFold(item.Entity.ID, in.ID) {
					continue
				}
				detail := TaskDetail{
					TaskSummary:   summarise(item),
					Description:   item.Entity.Sections[backlog.SectionDescription],
					Plan:          item.Entity.Sections[backlog.SectionPlan],
					Notes:         item.Entity.Sections[backlog.SectionNotes],
					FinalSummary:  item.Entity.Sections[backlog.SectionFinalSummary],
					References:    item.Entity.References,
					Documentation: item.Entity.Documentation,
					Path:          item.Entity.Path,
				}
				for _, criterion := range item.Entity.AcceptanceCriteria {
					mark := "[ ] "
					if criterion.Checked {
						mark = "[x] "
					}
					detail.AcceptanceCriteria = append(detail.AcceptanceCriteria,
						mark+criterion.Text)
				}
				return ok(detail)
			}
		}
		return fail[TaskDetail](notFound(in, p.Registry.DisplayName))
	})

	type searchIn struct {
		Text  string `json:"text" jsonschema:"what to look for, in titles and bodies"`
		Limit int    `json:"limit,omitempty" jsonschema:"stop after this many; 30 by default"`
	}
	type hit struct {
		Project string `json:"project"`
		Kind    string `json:"kind"`
		ID      string `json:"id"`
		Title   string `json:"title"`
		Excerpt string `json:"excerpt,omitempty"`
	}
	type hitList struct {
		Hits []hit `json:"hits" jsonschema:"what matched, best first"`
	}
	mcp.AddTool(server, &mcp.Tool{
		Annotations: reads(),
		Name:        "search",
		Description: "Find anything across every project by text: tasks, " +
			"drafts, documents, decisions and milestones. Titles rank above " +
			"bodies.",
	}, func(_ context.Context, _ *mcp.CallToolRequest, in searchIn) (*mcp.CallToolResult, hitList, error) {
		s.Reload()
		if s.missing {
			return nil, hitList{}, s.noRegistry()
		}
		limit := in.Limit
		if limit <= 0 {
			limit = 30
		}
		var out []hit
		for _, found := range s.store.Search(in.Text, limit) {
			out = append(out, hit{
				Project: found.Item.ProjectName,
				Kind:    string(found.Item.Ref.Kind),
				ID:      found.Item.Entity.ID,
				Title:   found.Item.Entity.Title,
				Excerpt: found.Excerpt,
			})
		}
		return ok(hitList{Hits: out})
	})

	type milestone struct {
		Project string `json:"project"`
		ID      string `json:"id"`
		Title   string `json:"title"`
		Done    int    `json:"done"`
		Total   int    `json:"total"`
	}
	type milestoneList struct {
		Milestones []milestone `json:"milestones" jsonschema:"every milestone, across every project"`
	}
	mcp.AddTool(server, &mcp.Tool{
		Annotations: reads(),
		Name:        "list_milestones",
		Description: "Milestones across every project with their progress. " +
			"This is the axis these backlogs are planned on, so it is usually " +
			"the answer to what is being worked towards.",
	}, func(_ context.Context, _ *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, milestoneList, error) {
		s.Reload()
		if s.missing {
			return nil, milestoneList{}, s.noRegistry()
		}
		var out []milestone
		for _, p := range s.store.Projects() {
			if !p.OK() {
				continue
			}
			terminal := terminalOf(p)
			counts := map[string][2]int{}
			for _, task := range p.Scanned.Tasks {
				if task.Class != backlog.ClassActive || task.Milestone == "" {
					continue
				}
				key := strings.ToLower(task.Milestone)
				c := counts[key]
				c[0]++
				if terminal != "" && strings.EqualFold(task.Status, terminal) {
					c[1]++
				}
				counts[key] = c
			}
			for _, m := range p.Scanned.Milestones {
				if m.Class != backlog.ClassActive {
					continue
				}
				c := counts[strings.ToLower(m.ID)]
				if t := counts[strings.ToLower(m.Title)]; t[0] > c[0] {
					c = t
				}
				out = append(out, milestone{
					Project: p.Registry.DisplayName, ID: m.ID, Title: m.Title,
					Done: c[1], Total: c[0],
				})
			}
		}
		return ok(milestoneList{Milestones: out})
	})

	type entityIn struct {
		Kind    string `json:"kind" jsonschema:"one of document, decision, draft or milestone"`
		Project string `json:"project,omitempty" jsonschema:"limit to one project; omit for every project"`
	}
	type entityOut struct {
		Project string `json:"project"`
		ID      string `json:"id"`
		Title   string `json:"title"`
		Status  string `json:"status,omitempty"`
		Type    string `json:"type,omitempty"`
		Body    string `json:"body,omitempty"`
		Path    string `json:"path"`
	}
	type entityList struct {
		Entities []entityOut `json:"entities" jsonschema:"what was found, of the kind asked for"`
	}
	mcp.AddTool(server, &mcp.Tool{
		Annotations: reads(),
		Name:        "list_entities",
		Description: "Documents, decisions, drafts or milestones, with their " +
			"bodies. Decisions are where a project records why a choice was " +
			"made, which is often what an agent is missing.",
	}, func(_ context.Context, _ *mcp.CallToolRequest, in entityIn) (*mcp.CallToolResult, entityList, error) {
		s.Reload()
		if s.missing {
			return nil, entityList{}, s.noRegistry()
		}
		kind := backlog.Kind(strings.ToLower(strings.TrimSpace(in.Kind)))
		switch kind {
		case backlog.KindDocument, backlog.KindDecision, backlog.KindDraft, backlog.KindMilestone:
		default:
			return fail[entityList](badKind(in.Kind))
		}

		var only string
		if strings.TrimSpace(in.Project) != "" {
			p, err := s.project(in.Project)
			if err != nil {
				return fail[entityList](err)
			}
			only = p.Registry.Path
		}

		var out []entityOut
		for _, item := range s.store.Entities(kind) {
			if only != "" && item.Ref.Project != only {
				continue
			}
			if item.Ref.Class != backlog.ClassActive {
				continue
			}
			out = append(out, entityOut{
				Project: item.ProjectName,
				ID:      item.Entity.ID,
				Title:   item.Entity.Title,
				Status:  item.Entity.Status,
				Type:    item.Entity.Type,
				Body:    item.Entity.Body,
				Path:    item.Entity.Path,
			})
		}
		return ok(entityList{Entities: out})
	})
}

// terminalOf is a project's last declared status, the only definition of
// finished the format offers.
func terminalOf(p store.ProjectState) string {
	list := p.Scanned.Config.Statuses
	if len(list) == 0 {
		return ""
	}
	return list[len(list)-1]
}
