package buildinfo

import (
	"strings"
	"testing"
)

// A build that was not stamped must say so rather than claim a number nobody
// released. This is the case every `go build` and `go test` hits.
func TestAnUnstampedBuildSaysDev(t *testing.T) {
	original := stamped
	t.Cleanup(func() { stamped = original })

	stamped = ""
	got := Version()
	// Under `go test` the toolchain records no revision, so this is plain
	// "dev"; a hand-built binary with VCS stamping gets the commit appended.
	if got != "dev" && !strings.HasPrefix(got, "dev+") {
		t.Errorf("an unstamped build calls itself %q", got)
	}
	if strings.Contains(got, "0.0.0") {
		t.Errorf("an unstamped build claimed a version: %q", got)
	}
}

func TestAStampedBuildReportsWhatItWasStampedWith(t *testing.T) {
	original := stamped
	t.Cleanup(func() { stamped = original })

	stamped = "1.2.3"
	if got := Version(); got != "1.2.3" {
		t.Errorf("Version() = %q, want the stamp", got)
	}

	// The ldflags value arrives from a Taskfile template, so a stray newline
	// or space is a real possibility and would end up in a bug report.
	stamped = "  1.2.3\n"
	if got := Version(); got != "1.2.3" {
		t.Errorf("Version() = %q, want it trimmed", got)
	}
}

// The line answers the two questions that follow "which version?".
func TestTheVersionLineNamesTheBuildAndThePlatform(t *testing.T) {
	original := stamped
	t.Cleanup(func() { stamped = original })

	stamped = "9.9.9"
	line := Line()
	for _, want := range []string{"muster", "9.9.9", "go1.", "/"} {
		if !strings.Contains(line, want) {
			t.Errorf("the version line %q does not mention %q", line, want)
		}
	}
}
