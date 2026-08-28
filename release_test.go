package main

import (
	"encoding/json"
	"os"
	"regexp"
	"strings"
	"testing"
)

// The version has one home, and three things read it: the build stamps the
// binary from it, the .deb takes its version from it, and release-please
// bumps it. Each of those is a silent failure if the line stops looking the
// way the reader expects — a package with an empty version, a binary that
// cannot name itself, a release that changes nothing.

// The same expression the Taskfile uses to pull the version out.
var versionLine = regexp.MustCompile(`(?m)^\s+version:\s*"([^"]*)"`)

func configVersion(t *testing.T) string {
	t.Helper()
	content, err := os.ReadFile("build/config.yml")
	if err != nil {
		t.Fatalf("read build/config.yml: %v", err)
	}
	match := versionLine.FindSubmatch(content)
	if match == nil {
		t.Fatal("build/config.yml has no indented version line; the build reads it with a regex and would stamp an empty string")
	}
	return string(match[1])
}

func TestTheVersionIsSemver(t *testing.T) {
	version := configVersion(t)
	if !regexp.MustCompile(`^\d+\.\d+\.\d+`).MatchString(version) {
		t.Errorf("the version is %q, which is not a SemVer number", version)
	}
}

// release-please finds the line to bump by an annotation in the file itself.
// Without it a release would tag and write a changelog while every build kept
// reporting the old number.
func TestTheVersionLineIsAnnotatedForTheReleaseAutomation(t *testing.T) {
	content, err := os.ReadFile("build/config.yml")
	if err != nil {
		t.Fatalf("read build/config.yml: %v", err)
	}
	annotated := 0
	for _, line := range strings.Split(string(content), "\n") {
		if strings.Contains(line, "x-release-please-version") &&
			!strings.HasPrefix(strings.TrimSpace(line), "#") {
			annotated++
			if !versionLine.MatchString(line + "\n") {
				t.Errorf("the annotation is on a line that is not the version: %q", line)
			}
		}
	}
	if annotated != 1 {
		t.Errorf("%d annotated version lines in build/config.yml, want exactly one", annotated)
	}
}

// The manifest is where release-please believes the last release was. If it
// disagrees with the file the build reads, one of them is lying about what is
// installed.
func TestTheReleaseManifestAgreesWithTheBuild(t *testing.T) {
	content, err := os.ReadFile(".release-please-manifest.json")
	if err != nil {
		t.Fatalf("read the release manifest: %v", err)
	}
	var manifest map[string]string
	if err := json.Unmarshal(content, &manifest); err != nil {
		t.Fatalf("the release manifest is not valid JSON: %v", err)
	}
	if got, want := manifest["."], configVersion(t); got != want {
		t.Errorf("the release manifest says %q and build/config.yml says %q", got, want)
	}
}

func TestTheReleaseConfigBumpsTheFileTheBuildReads(t *testing.T) {
	content, err := os.ReadFile("release-please-config.json")
	if err != nil {
		t.Fatalf("read the release config: %v", err)
	}
	var config struct {
		Packages map[string]struct {
			ReleaseType string `json:"release-type"`
			// Raw, because an entry may be written either way and the
			// difference between the two is the whole of this check.
			ExtraFiles          []json.RawMessage `json:"extra-files"`
			BumpMinorPreMajor   bool              `json:"bump-minor-pre-major"`
			BumpPatchForMinor   bool              `json:"bump-patch-for-minor-pre-major"`
			IncludeVInTag       bool              `json:"include-v-in-tag"`
			IncludeComponentTag bool              `json:"include-component-in-tag"`
		} `json:"packages"`
	}
	if err := json.Unmarshal(content, &config); err != nil {
		t.Fatalf("the release config is not valid JSON: %v", err)
	}

	root, ok := config.Packages["."]
	if !ok {
		t.Fatal("the release config has no package for the repository root")
	}
	// Named as an object with an explicit updater, never as a bare string.
	//
	// A bare string is not a line edit. For a .yml file release-please pairs
	// its annotation-driven updater with one that parses the YAML and writes
	// it back out, and the second one wins the file: measured against this
	// very config, thirty-nine lines became twelve. Every comment goes,
	// including the x-release-please-version annotation the first updater
	// needs; the top-level `version: '3'` — the Wails schema version, not the
	// application's — is overwritten with the release number; and
	// info.version keeps its old value and loses its quotes, so the Taskfile's
	// sed matches nothing and every binary in that release reports an empty
	// version.
	//
	// `generic` is the line edit: it changes the annotated line and nothing
	// else.
	found := false
	for _, entry := range root.ExtraFiles {
		var bare string
		if err := json.Unmarshal(entry, &bare); err == nil {
			if bare == "build/config.yml" {
				found = true
				t.Error("build/config.yml is listed as a bare string, which pairs the line edit with a YAML round-trip that wins the file")
			}
			continue
		}
		var file struct {
			Type string `json:"type"`
			Path string `json:"path"`
		}
		if err := json.Unmarshal(entry, &file); err != nil {
			t.Errorf("an extra-files entry is neither a string nor an object: %s", entry)
			continue
		}
		if file.Path != "build/config.yml" {
			continue
		}
		found = true
		if file.Type != "generic" {
			t.Errorf("build/config.yml is updated as %q; anything but \"generic\" re-serialises the file and destroys the annotation, the comments and the schema version", file.Type)
		}
	}
	if !found {
		t.Error("release-please does not update build/config.yml, so a release would not change the version any build reports")
	}
	// The documented policy before 1.0: a minor may break, a patch does not.
	// Both settings are what make that true rather than aspirational.
	if !root.BumpMinorPreMajor {
		t.Error("bump-minor-pre-major is off, so a breaking change would go to 1.0 rather than a minor")
	}
	if root.BumpPatchForMinor {
		t.Error("bump-patch-for-minor-pre-major is on, so a feature would land as a patch, which the versioning policy says does not break anything")
	}
}

// The changelog is where the versioning promise is written down, and where
// everything unreleased collects.
func TestTheChangelogKeepsAnUnreleasedSection(t *testing.T) {
	content, err := os.ReadFile("CHANGELOG.md")
	if err != nil {
		t.Fatalf("read CHANGELOG.md: %v", err)
	}
	text := string(content)
	for _, want := range []string{"## Unreleased", "Keep a Changelog", "Semantic Versioning"} {
		if !strings.Contains(text, want) {
			t.Errorf("CHANGELOG.md does not mention %q", want)
		}
	}
	// What a minor bump means before 1.0 is the whole question a version
	// number raises here, and it has to be answered in the file itself.
	if !strings.Contains(text, "Versioning before 1.0") {
		t.Error("CHANGELOG.md does not say what a version number means before 1.0")
	}
}

// The desktop entry is corrected after the generator writes it, and that
// correction runs more than once in a packaging run. An unconditional append
// put a second Comment= in the file, which appimagetool rejects outright —
// the AppImage failed at the very last step, after two minutes of bundling.
func TestTheDesktopCorrectionIsIdempotent(t *testing.T) {
	content, err := os.ReadFile("build/linux/Taskfile.yml")
	if err != nil {
		t.Fatalf("read the linux Taskfile: %v", err)
	}
	text := string(content)

	appends := strings.Contains(text, "/^Name=/a Comment=")
	deletes := strings.Contains(text, "-e '/^Comment=/d'")
	if appends && !deletes {
		t.Error("the desktop entry gains a Comment line without removing any first, so running the task twice writes two and appimagetool refuses the file")
	}
}

// nfpm builds the Maintainer field from these two variables. Nothing set them
// for a long time, so every package declared "Maintainer: <>".
func TestThePackagesNameAMaintainer(t *testing.T) {
	content, err := os.ReadFile("build/linux/Taskfile.yml")
	if err != nil {
		t.Fatalf("read the linux Taskfile: %v", err)
	}
	text := string(content)
	for _, want := range []string{"GIT_COMMITTER_NAME", "GIT_COMMITTER_EMAIL"} {
		if !strings.Contains(text, want) {
			t.Errorf("nothing sets %s, so the package declares an empty maintainer", want)
		}
	}
	// And the version, which was the same kind of silence.
	if strings.Count(text, "PRODUCT_VERSION:") < 3 {
		t.Error("not every packaging task sets PRODUCT_VERSION, so some package takes an unset variable as its version")
	}
}

// A release that builds nothing reaches nobody, and the ways this goes wrong
// are all quiet ones: a job that never runs, a version published with no
// downloads, artefacts nobody can check.
func TestTheReleaseWorkflowAttachesArtefactsBeforePublishing(t *testing.T) {
	w := readWorkflow(t, ".github/workflows/release-please.yml")

	job, ok := w.Jobs["artefacts"]
	if !ok {
		t.Fatal("the release workflow builds no artefacts, so a release is a tag and a changelog")
	}

	// In this workflow rather than one of its own, and that is not a style
	// choice. A release created with GITHUB_TOKEN raises no `release` event
	// and its tag raises no `push` event, so a workflow keyed on either would
	// never run and nothing would say why. `needs` is what makes it run.
	if !strings.Contains(toText(job.Needs), "release-please") {
		t.Error("the artefacts job does not hang off the release job, so nothing gives it the tag")
	}
	if !strings.Contains(job.If, "released") {
		t.Errorf("the artefacts job is not conditioned on a release having happened: %q", job.If)
	}

	// The order is the promise: everything is attached, and only then is the
	// release made public.
	uploaded, published, checksums := -1, -1, false
	for i, step := range job.Steps {
		if strings.Contains(step.Run, "sha256sum") {
			checksums = true
		}
		if strings.Contains(step.Run, "gh release upload") {
			uploaded = i
		}
		if strings.Contains(step.Run, "--draft=false") {
			published = i
		}
	}
	if uploaded < 0 {
		t.Error("nothing is uploaded to the release")
	}
	if published < 0 {
		t.Error("the release is never taken out of draft, so it stays invisible")
	}
	if uploaded >= 0 && published >= 0 && published < uploaded {
		t.Error("the release is published before its artefacts are attached, which is the window this shape exists to close")
	}
	if !checksums {
		t.Error("no checksums are published, so a download cannot be verified")
	}

	// Cutting the release as a draft is the other half of it. Without this
	// the release is public the moment it is made and the ordering above buys
	// nothing.
	content, err := os.ReadFile("release-please-config.json")
	if err != nil {
		t.Fatalf("read the release config: %v", err)
	}
	var config struct {
		Packages map[string]struct {
			Draft bool `json:"draft"`
		} `json:"packages"`
	}
	if err := json.Unmarshal(content, &config); err != nil {
		t.Fatalf("the release config is not valid JSON: %v", err)
	}
	if !config.Packages["."].Draft {
		t.Error("the release is created public, so it is visible with nothing attached while the artefacts are still building")
	}
}

// Two workflows building the same project with different toolchains would ship
// artefacts nothing had checked.
func TestTheReleaseBuildUsesTheToolchainThePipelineChecked(t *testing.T) {
	ci := readWorkflow(t, ".github/workflows/ci.yml")
	release := readWorkflow(t, ".github/workflows/release-please.yml")

	for _, name := range []string{"WAILS_VERSION", "NODE_VERSION"} {
		pipeline, _ := ci.Env[name].(string)
		shipped, _ := release.Env[name].(string)
		if pipeline == "" {
			t.Errorf("ci.yml does not pin %s", name)
			continue
		}
		if shipped != pipeline {
			t.Errorf("the pipeline builds with %s %q and the release builds with %q", name, pipeline, shipped)
		}
	}
}
