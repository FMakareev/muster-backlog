package main

import (
	"encoding/json"
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/goccy/go-yaml"
)

// A workflow cannot be run before the repository exists, and the two ways it
// fails quietly are both checkable here: a version pinned to something that
// moves, and a toolchain version that has drifted from what the README tells
// a person to install. The second is what makes "works on my machine"
// arguments unresolvable.

type workflow struct {
	Name string         `yaml:"name"`
	On   map[string]any `yaml:"on"`
	Env  map[string]any `yaml:"env"`
	Jobs map[string]struct {
		RunsOn         string            `yaml:"runs-on"`
		TimeoutMinutes int               `yaml:"timeout-minutes"`
		Needs          any               `yaml:"needs"`
		If             string            `yaml:"if"`
		Env            map[string]string `yaml:"env"`
		Steps          []struct {
			Name string            `yaml:"name"`
			Uses string            `yaml:"uses"`
			Run  string            `yaml:"run"`
			With map[string]any    `yaml:"with"`
			Env  map[string]string `yaml:"env"`
			If   string            `yaml:"if"`
		} `yaml:"steps"`
	} `yaml:"jobs"`
}

func readWorkflow(t *testing.T, path string) workflow {
	t.Helper()
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	var w workflow
	if err := yaml.Unmarshal(content, &w); err != nil {
		t.Fatalf("%s is not valid YAML: %v", path, err)
	}
	return w
}

// Pull requests and the default branch, which is what makes a required check
// mean anything.
func TestContinuousIntegrationRunsWhereItMatters(t *testing.T) {
	w := readWorkflow(t, ".github/workflows/ci.yml")
	if _, ok := w.On["pull_request"]; !ok {
		t.Error("the pipeline does not run on pull requests, so nothing can be required to pass before merge")
	}
	push, ok := w.On["push"]
	if !ok {
		t.Fatal("the pipeline does not run on pushes to the default branch")
	}
	// The default branch here is master. Naming the wrong one is not an error
	// GitHub reports — the workflow simply never runs.
	if !strings.Contains(strings.ToLower(toText(push)), "master") {
		t.Errorf("the push trigger does not name the default branch: %v", push)
	}
}

func toText(v any) string {
	out, _ := json.Marshal(v)
	return string(out)
}

// The pipeline runs the same commands a person runs. A separate list is a
// list that drifts.
func TestThePipelineRunsWhatADeveloperRuns(t *testing.T) {
	w := readWorkflow(t, ".github/workflows/ci.yml")
	job, ok := w.Jobs["check"]
	if !ok {
		t.Fatal("the workflow has no check job")
	}

	var script strings.Builder
	requiresCLI := false
	for _, step := range job.Steps {
		script.WriteString(step.Run)
		script.WriteString("\n")
		if step.Env["MUSTER_REQUIRE_BACKLOG_CLI"] != "" && strings.Contains(step.Run, "test") {
			requiresCLI = true
		}
	}
	body := script.String()

	for _, want := range []string{"wails3 task lint", "wails3 task test", "wails3 task build"} {
		if !strings.Contains(body, want) {
			t.Errorf("the pipeline never runs %q", want)
		}
	}

	// Thirty-eight tests exercise real writes through the backlog CLI and
	// skip without it. A pipeline that does not install it, or lets them
	// skip, reports green having tested none of them.
	if !strings.Contains(body, "backlog.md@") {
		t.Error("the pipeline does not install the Backlog.md CLI, so every write test would skip")
	}
	if !requiresCLI {
		t.Errorf("the test step does not set MUSTER_REQUIRE_BACKLOG_CLI, so a missing CLI would skip silently instead of failing")
	}

	// A build that produces a file proves less than a binary that can say
	// which version it is.
	if !strings.Contains(body, "--version") {
		t.Error("the pipeline never asks the built binary what version it is")
	}
}

// Every version CI installs is pinned. An unpinned tool turns the pipeline red
// on somebody else's release day, which teaches people to ignore it.
func TestTheContinuousIntegrationToolsArePinned(t *testing.T) {
	content, err := os.ReadFile(".github/workflows/ci.yml")
	if err != nil {
		t.Fatalf("read the workflow: %v", err)
	}
	text := string(content)
	if strings.Contains(text, "@latest") {
		t.Error("the workflow installs something at @latest")
	}
	for _, action := range regexp.MustCompile(`uses:\s*(\S+)`).FindAllStringSubmatch(text, -1) {
		if !strings.Contains(action[1], "@") {
			t.Errorf("the action %q is not pinned to a version", action[1])
		}
	}
}

// What CI installs and what the README tells a person to install have to be
// the same thing, or a disagreement between the two becomes unarguable.
func TestTheContinuousIntegrationToolchainMatchesTheDocumentedOne(t *testing.T) {
	w := readWorkflow(t, ".github/workflows/ci.yml")
	readme, err := os.ReadFile("README.md")
	if err != nil {
		t.Fatalf("read README.md: %v", err)
	}
	text := string(readme)

	for _, pair := range []struct{ env, describes string }{
		{"WAILS_VERSION", "the Wails CLI"},
		{"GOLANGCI_VERSION", "golangci-lint"},
		{"BACKLOG_VERSION", "the Backlog.md CLI"},
	} {
		value, _ := w.Env[pair.env].(string)
		if value == "" {
			t.Errorf("the workflow does not pin %s", pair.env)
			continue
		}
		// A moving target is not a pin, and "latest" would otherwise satisfy
		// the README check below by appearing in some unrelated sentence.
		if !regexp.MustCompile(`^v?\d+\.\d+`).MatchString(value) {
			t.Errorf("%s is %q, which is not a version", pair.env, value)
			continue
		}
		if !strings.Contains(text, strings.TrimPrefix(value, "v")) {
			t.Errorf("CI installs %s %s and the README never mentions that version", pair.describes, value)
		}
	}

	// The Go version comes from go.mod rather than being written twice.
	job := w.Jobs["check"]
	fromModFile := false
	for _, step := range job.Steps {
		if strings.Contains(step.Uses, "setup-go") && step.With["go-version-file"] != nil {
			fromModFile = true
		}
	}
	if !fromModFile {
		t.Error("CI pins a Go version of its own instead of taking it from go.mod, which is a second place for it to drift")
	}

	// And pnpm comes from packageManager, for the same reason.
	pkg, err := os.ReadFile("package.json")
	if err != nil {
		t.Fatalf("read package.json: %v", err)
	}
	var manifest struct {
		PackageManager string `json:"packageManager"`
	}
	if err := json.Unmarshal(pkg, &manifest); err != nil {
		t.Fatalf("package.json is not valid JSON: %v", err)
	}
	if !strings.HasPrefix(manifest.PackageManager, "pnpm@") {
		t.Error("package.json has no packageManager field, so pnpm/action-setup has no version to install")
	}
}

// A Go tool installed with `go install` is compiled with the Go version this
// job takes from go.mod, which quietly makes that tool's own go directive a
// floor on this module's. golangci-lint v2.13.0 raised its to 1.26.0 while
// go.mod declares 1.25.0, and the pipeline then died on "requires go >=
// 1.26.0" before a single line of this project was compiled. Pinning an older
// linter would have fixed it until the next time the linter moved.
//
// Nothing in the version pin shows that relationship, so this is what says it.
func TestTheLinterIsNotBuiltFromSource(t *testing.T) {
	content, err := os.ReadFile(".github/workflows/ci.yml")
	if err != nil {
		t.Fatalf("read the workflow: %v", err)
	}
	text := string(content)

	for _, line := range strings.Split(text, "\n") {
		if strings.Contains(line, "go install") && strings.Contains(line, "golangci-lint") {
			t.Errorf("CI compiles golangci-lint with this project's Go, which ties the linter it can pin to the Go version go.mod declares: %s", strings.TrimSpace(line))
		}
	}

	// And it does obtain it, at the pinned version rather than whatever the
	// project happens to publish today.
	if !strings.Contains(text, "releases/download/${GOLANGCI_VERSION}") {
		t.Error("CI does not download the pinned golangci-lint release")
	}

	// A download is only the release that was pinned if something checks.
	if !strings.Contains(text, "sha256sum --check") || !strings.Contains(text, "checksums.txt") {
		t.Error("CI installs a downloaded golangci-lint without verifying it against the checksums published for that release")
	}
}

// release-please only ever runs on a push to the branch it names, so naming
// the wrong one means no release pull request is ever opened and nothing says
// so. This is the check that would have caught it.
func TestTheReleaseWorkflowRunsOnTheDefaultBranch(t *testing.T) {
	w := readWorkflow(t, ".github/workflows/release-please.yml")
	push, ok := w.On["push"]
	if !ok {
		t.Fatal("release-please does not run on a push at all")
	}
	if !strings.Contains(strings.ToLower(toText(push)), "master") {
		t.Errorf("release-please does not run on the default branch: %v", push)
	}
}

// Caching is what keeps the pipeline inside the budget the README states.
func TestTheSlowThingsAreCached(t *testing.T) {
	w := readWorkflow(t, ".github/workflows/ci.yml")
	job := w.Jobs["check"]

	cached := map[string]bool{}
	for _, step := range job.Steps {
		switch {
		case strings.Contains(step.Uses, "setup-go"):
			if step.With["cache"] == true {
				cached["go modules and build"] = true
			}
		case strings.Contains(step.Uses, "setup-node"):
			if step.With["cache"] != nil {
				cached["the pnpm store"] = true
			}
		case strings.Contains(step.Uses, "actions/cache"):
			cached["the compiled Go tools"] = true
		}
	}
	for _, want := range []string{"go modules and build", "the pnpm store", "the compiled Go tools"} {
		if !cached[want] {
			t.Errorf("%s is not cached, and it is one of the slow things", want)
		}
	}

	if job.TimeoutMinutes == 0 {
		t.Error("the job has no timeout, so a hang costs a runner hour rather than failing")
	}

	// The budget is a number a person can hold the pipeline to.
	readme, err := os.ReadFile("README.md")
	if err != nil {
		t.Fatalf("read README.md: %v", err)
	}
	if !strings.Contains(string(readme), "time budget") {
		t.Error("the README does not state a time budget for the pipeline")
	}
}
