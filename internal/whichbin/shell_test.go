package whichbin

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

// A shell that behaves like a person's: it prints a greeting, exports a PATH
// nobody could have guessed, and only then answers.
func fakeShell(t *testing.T, body string) string {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("login shells are a Unix idea")
	}
	path := filepath.Join(t.TempDir(), "shell")
	script := "#!/bin/sh\n" + body + "\n"
	if err := os.WriteFile(path, []byte(script), 0o755); err != nil {
		t.Fatalf("write the shell: %v", err)
	}
	return path
}

func forget() {
	askedMu.Lock()
	asked = map[string]shellAnswer{}
	askedMu.Unlock()
}

// The whole point: a command in a directory no list contains is still found,
// because the shell knows where it is.
func TestACommandOnlyTheShellKnowsAboutIsFound(t *testing.T) {
	forget()
	t.Cleanup(forget)

	// Somewhere nothing in dirs() would ever look.
	hidden := t.TempDir()
	tool := filepath.Join(hidden, "a-tool-nobody-guesses")
	if err := os.WriteFile(tool, []byte("#!/bin/sh\n"), 0o755); err != nil {
		t.Fatalf("write the tool: %v", err)
	}

	t.Setenv("SHELL", fakeShell(t, `
echo "Welcome to your shell"
echo "nvm: now using node v24"
command -v -- "$WHICHBIN_NAME" 2>/dev/null || echo "$HOME/`+filepath.Base(tool)+`"
`))
	// The fake answers with the real path, the way a profile-loaded shell
	// would after putting that directory on PATH.
	t.Setenv("SHELL", fakeShell(t, `
echo "Welcome to your shell"
echo "nvm: now using node v24"
echo "`+tool+`"
`))

	got, found := askShell("a-tool-nobody-guesses")
	if !found {
		t.Fatal("the shell was asked and its answer was not used")
	}
	if got != tool {
		t.Errorf("got %q, want %q", got, tool)
	}
}

// A profile prints things. The answer is the runnable path among them, not
// the last line.
func TestGreetingsAndNoiseAreNotMistakenForAnAnswer(t *testing.T) {
	forget()
	t.Cleanup(forget)

	t.Setenv("SHELL", fakeShell(t, `
echo "some banner"
echo "/definitely/not/here/backlog"
echo "backlog"
`))
	if got, found := askShell("backlog"); found {
		t.Errorf("a path that does not exist was accepted: %q", got)
	}
}

// `command -v` answers with a bare word for a function, an alias or a builtin.
// None of those is something this process can execute, and a relative path is
// worse than nothing.
func TestOnlyAnAbsolutePathToSomethingRunnableIsAccepted(t *testing.T) {
	forget()
	t.Cleanup(forget)

	for _, answer := range []string{"backlog", "./backlog", "backlog: aliased to backlog --plain", ""} {
		forget()
		t.Setenv("SHELL", fakeShell(t, "echo '"+answer+"'"))
		if got, found := askShell("backlog"); found {
			t.Errorf("the shell answered %q and it was accepted as %q", answer, got)
		}
	}
}

// A profile that hangs must not take the application with it.
func TestAShellThatHangsIsGivenUpOn(t *testing.T) {
	forget()
	t.Cleanup(forget)

	t.Setenv("SHELL", fakeShell(t, "sleep 60"))

	start := time.Now()
	_, found := askShell("backlog")
	elapsed := time.Since(start)

	if found {
		t.Error("a shell that never answered was treated as an answer")
	}
	if elapsed > shellTimeout+3*time.Second {
		t.Errorf("waited %s for a shell that hangs", elapsed)
	}
}

// A second and a half is far too long to spend twice on the same question.
func TestTheShellIsAskedOnce(t *testing.T) {
	forget()
	t.Cleanup(forget)

	counter := filepath.Join(t.TempDir(), "asked")
	t.Setenv("SHELL", fakeShell(t, "echo x >> "+counter+"\necho nothing-here"))

	for range 3 {
		askShell("backlog")
	}

	content, err := os.ReadFile(counter)
	if err != nil {
		t.Fatalf("the shell was never run at all: %v", err)
	}
	if runs := strings.Count(string(content), "x"); runs != 1 {
		t.Errorf("the shell was run %d times, want once", runs)
	}
}

// The name reaches the shell as data, never as syntax.
func TestTheNameIsNotShellSyntax(t *testing.T) {
	forget()
	t.Cleanup(forget)

	marker := filepath.Join(t.TempDir(), "should-not-exist")
	t.Setenv("SHELL", fakeShell(t, `
# Whatever the name is, it arrives in a variable rather than in the command.
echo "asked for: $WHICHBIN_NAME"
`))

	askShell("x; touch " + marker + "; echo y")

	if _, err := os.Stat(marker); err == nil {
		t.Error("the name was interpreted as shell syntax and ran a command")
	}
}

// With no SHELL there is nothing to ask, and nothing should be run.
func TestNoShellMeansNoAnswer(t *testing.T) {
	forget()
	t.Cleanup(forget)

	t.Setenv("SHELL", "")
	if _, found := askShell("backlog"); found {
		t.Error("something answered with no shell set")
	}

	forget()
	t.Setenv("SHELL", filepath.Join(t.TempDir(), "not-a-real-shell"))
	if _, found := askShell("backlog"); found {
		t.Error("a shell that does not exist answered")
	}
}
