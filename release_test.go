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
			ReleaseType         string   `json:"release-type"`
			ExtraFiles          []string `json:"extra-files"`
			BumpMinorPreMajor   bool     `json:"bump-minor-pre-major"`
			BumpPatchForMinor   bool     `json:"bump-patch-for-minor-pre-major"`
			IncludeVInTag       bool     `json:"include-v-in-tag"`
			IncludeComponentTag bool     `json:"include-component-in-tag"`
		} `json:"packages"`
	}
	if err := json.Unmarshal(content, &config); err != nil {
		t.Fatalf("the release config is not valid JSON: %v", err)
	}

	root, ok := config.Packages["."]
	if !ok {
		t.Fatal("the release config has no package for the repository root")
	}
	found := false
	for _, file := range root.ExtraFiles {
		if file == "build/config.yml" {
			found = true
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
