package main

import (
	"os"
	"strings"
	"testing"

	"github.com/goccy/go-yaml"
)

// Issue forms cannot be tried out before the repository exists, and a
// malformed one does not fail loudly: GitHub declines to render it and the
// person filing the bug gets a blank box instead. So the shape is checked
// here, along with the things these particular forms exist to ask for.

type formField struct {
	Type       string         `yaml:"type"`
	ID         string         `yaml:"id"`
	Attributes map[string]any `yaml:"attributes"`
	Validation map[string]any `yaml:"validations"`
}

type issueForm struct {
	Name        string      `yaml:"name"`
	Description string      `yaml:"description"`
	Labels      []string    `yaml:"labels"`
	Body        []formField `yaml:"body"`
}

func readForm(t *testing.T, path string) issueForm {
	t.Helper()
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	var form issueForm
	if err := yaml.Unmarshal(content, &form); err != nil {
		t.Fatalf("%s is not valid YAML: %v", path, err)
	}
	return form
}

// The shape GitHub requires, for every form.
func TestTheIssueFormsAreWellFormed(t *testing.T) {
	known := map[string]bool{
		"markdown": true, "input": true, "textarea": true,
		"dropdown": true, "checkboxes": true,
	}

	for _, path := range []string{
		".github/ISSUE_TEMPLATE/bug_report.yml",
		".github/ISSUE_TEMPLATE/feature_request.yml",
	} {
		form := readForm(t, path)
		if form.Name == "" || form.Description == "" {
			t.Errorf("%s needs both a name and a description to appear in the chooser", path)
		}
		if len(form.Body) == 0 {
			t.Fatalf("%s has no fields", path)
		}

		seen := map[string]bool{}
		for i, field := range form.Body {
			if !known[field.Type] {
				t.Errorf("%s field %d has type %q, which GitHub does not render", path, i, field.Type)
				continue
			}
			if field.ID != "" {
				if seen[field.ID] {
					t.Errorf("%s uses the id %q twice; the later field is dropped", path, field.ID)
				}
				seen[field.ID] = true
			}

			switch field.Type {
			case "markdown":
				if field.Attributes["value"] == nil {
					t.Errorf("%s field %d is markdown with nothing in it", path, i)
				}
			case "dropdown", "checkboxes":
				if field.Attributes["options"] == nil {
					t.Errorf("%s field %q is a %s with no options", path, field.ID, field.Type)
				}
				fallthrough
			default:
				if field.Attributes["label"] == nil {
					t.Errorf("%s field %q has no label", path, field.ID)
				}
			}

			// A field marked required outside validations is silently
			// optional, which is the failure this form is meant to prevent.
			if field.Attributes["required"] != nil {
				t.Errorf("%s field %q puts required under attributes, where it does nothing", path, field.ID)
			}
		}
	}
}

// What the bug form exists to collect. Every one of these is something a
// report is nearly useless without, and each was chosen because this
// application's failures actually turn on it.
func TestTheBugFormAsksForWhatAReportNeeds(t *testing.T) {
	form := readForm(t, ".github/ISSUE_TEMPLATE/bug_report.yml")

	fields := map[string]formField{}
	for _, field := range form.Body {
		if field.ID != "" {
			fields[field.ID] = field
		}
	}

	for _, id := range []string{"versions", "system", "steps", "projects", "single-project"} {
		field, ok := fields[id]
		if !ok {
			t.Errorf("the bug form does not ask for %q", id)
			continue
		}
		if field.Validation["required"] != true {
			t.Errorf("the bug form asks for %q but does not require it", id)
		}
	}

	// Two versions, not one: Muster writes only through the backlog CLI, so
	// which CLI is involved is half of nearly every answer.
	whole, err := os.ReadFile(".github/ISSUE_TEMPLATE/bug_report.yml")
	if err != nil {
		t.Fatalf("read the bug form: %v", err)
	}
	text := string(whole)
	for _, want := range []string{"muster --version", "backlog --version", "Backlog.md CLI"} {
		if !strings.Contains(text, want) {
			t.Errorf("the bug form never mentions %q, so a reporter has to guess where to look", want)
		}
	}
	if !strings.Contains(text, "Wayland") {
		t.Error("the bug form does not ask about the session type, which the tray and the hotkey turn on")
	}
}

// The problem before the solution, and provably so rather than by intent.
func TestTheFeatureFormAsksForTheProblemFirst(t *testing.T) {
	form := readForm(t, ".github/ISSUE_TEMPLATE/feature_request.yml")

	positions := map[string]int{}
	firstAsked := -1
	for i, field := range form.Body {
		if field.ID != "" {
			positions[field.ID] = i
		}
		// markdown blocks are prose, not questions.
		if firstAsked < 0 && field.Type != "markdown" {
			firstAsked = i
		}
	}

	problem, ok := positions["problem"]
	if !ok {
		t.Fatal("the feature form has no field asking what the problem is")
	}
	if problem != firstAsked {
		t.Errorf("the first question is at %d and the problem is at %d; the problem has to come first", firstAsked, problem)
	}
	if idea, ok := positions["idea"]; ok && idea < problem {
		t.Errorf("the proposed solution (%d) is asked before the problem (%d)", idea, problem)
	}
	if form.Body[problem].Validation["required"] != true {
		t.Error("the problem is optional, so a request can arrive as a solution with no problem attached")
	}
	if idea, ok := positions["idea"]; ok && form.Body[idea].Validation["required"] == true {
		t.Error("the proposed solution is required, which forces a person who only has a problem to invent one")
	}
}

// Questions belong in Discussions, and a blank issue is how the forms get
// skipped.
func TestTheChooserSendsQuestionsElsewhere(t *testing.T) {
	content, err := os.ReadFile(".github/ISSUE_TEMPLATE/config.yml")
	if err != nil {
		t.Fatalf("read the chooser config: %v", err)
	}
	var config struct {
		Blank *bool `yaml:"blank_issues_enabled"`
		Links []struct {
			Name  string `yaml:"name"`
			URL   string `yaml:"url"`
			About string `yaml:"about"`
		} `yaml:"contact_links"`
	}
	if err := yaml.Unmarshal(content, &config); err != nil {
		t.Fatalf("the chooser config is not valid YAML: %v", err)
	}

	if config.Blank == nil || *config.Blank {
		t.Error("blank issues are enabled, which is how both forms get skipped")
	}
	if len(config.Links) == 0 {
		t.Fatal("the chooser offers nowhere to ask a question")
	}
	discussions := false
	for _, link := range config.Links {
		if link.Name == "" || link.URL == "" || link.About == "" {
			t.Errorf("the chooser entry %q is missing a name, url or about, and GitHub drops it", link.Name)
		}
		if strings.Contains(link.URL, "/discussions") {
			discussions = true
		}
	}
	if !discussions {
		t.Error("nothing in the chooser points at Discussions, so questions land in Issues")
	}
}

// The pull request template's job is to make a reviewer able to find what the
// change was supposed to do.
func TestThePullRequestTemplateAsksForTheTask(t *testing.T) {
	content, err := os.ReadFile(".github/pull_request_template.md")
	if err != nil {
		t.Fatalf("read the pull request template: %v", err)
	}
	text := string(content)
	for _, want := range []string{"TASK-", "acceptance criteria", "verified"} {
		if !strings.Contains(strings.ToLower(text), strings.ToLower(want)) {
			t.Errorf("the pull request template never mentions %q", want)
		}
	}
	if strings.Count(text, "- [ ]") < 5 {
		t.Error("the checklist is too short to be one")
	}
	// The two rules this project would rather catch in review than in a
	// release.
	for _, want := range []string{"backlog` CLI", "Backlog.md does not support"} {
		if !strings.Contains(text, want) {
			t.Errorf("the checklist does not mention %q", want)
		}
	}
}
