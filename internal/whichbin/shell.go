package whichbin

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"
)

// Asking the owner's shell where a command is.
//
// This is the answer to "why not just run `which`?". Running `which` from
// inside the process is what exec.LookPath already does, and it fails for the
// reason this package exists: the process was started by a launcher and has
// the system default PATH, not the one the owner sees in a terminal.
//
// Their shell does have it, because the shell reads the profile where the
// entry was written. So the last resort is to start their login shell and ask
// it. That needs no list of directories and cannot go stale — which the list
// above it can and did: a CLI installed by pnpm sat in a folder the list did
// not have, and the error then printed every folder it had looked in without
// printing that one.
//
// It is last for two reasons. It costs about a second and a half, because a
// login shell runs the whole profile — version managers, completions, the
// lot. And it runs the owner's own startup files, which is code this process
// did not choose; harmless, since it is exactly what every terminal they open
// already runs, but not something to do when a cheaper answer exists.

// shellTimeout bounds the wait. A profile that hangs — a prompt, a network
// call, a version manager fetching an index — must not take the application
// with it, and one that slow is not going to answer usefully anyway.
const shellTimeout = 5 * time.Second

// The variable the shell reads the name from, rather than interpolating it
// into the command. Nothing the caller passes is ever parsed as shell syntax.
const nameVar = "WHICHBIN_NAME"

var (
	askedMu sync.Mutex
	asked   = map[string]shellAnswer{}
)

type shellAnswer struct {
	path  string
	found bool
}

// askShell asks the login shell where name is, once per name per process.
//
// The answer is cached whether or not it found anything: a second and a half
// is far too long to spend twice on the same question, and the shell is not
// going to change its mind while the application is running.
func askShell(name string) (string, bool) {
	if runtime.GOOS == "windows" {
		// A login shell that reads a profile is a Unix idea. On Windows the
		// PATH a process gets is the PATH the owner has.
		return "", false
	}

	askedMu.Lock()
	if answer, ok := asked[name]; ok {
		askedMu.Unlock()
		return answer.path, answer.found
	}
	askedMu.Unlock()

	path, found := runShell(name)

	askedMu.Lock()
	asked[name] = shellAnswer{path: path, found: found}
	askedMu.Unlock()
	return path, found
}

func runShell(name string) (string, bool) {
	shell := strings.TrimSpace(os.Getenv("SHELL"))
	if shell == "" || !runnable(shell) {
		return "", false
	}

	ctx, cancel := context.WithTimeout(context.Background(), shellTimeout)
	defer cancel()

	// -l reads the login profile and -i the interactive one, because which of
	// the two holds the PATH entry differs by shell and by person: zsh puts
	// most people's in .zshrc, which only an interactive shell reads.
	cmd := exec.CommandContext(ctx, shell, "-lic", `command -v -- "$`+nameVar+`"`)
	cmd.Env = append(os.Environ(), nameVar+"="+name)
	cmd.Stdin = nil

	// The context alone does not bound this, and the first version of it did
	// not: cancelling kills the shell, but anything the profile started
	// inherits the output pipe and keeps it open, and Output waits on the
	// pipe rather than on the process. A profile that starts a background
	// process — plenty do — held the application for the full sixty seconds
	// of a test that was meant to prove the opposite.
	//
	// WaitDelay gives up on the pipes shortly after the kill. detach puts the
	// shell in its own process group where the platform allows, so the kill
	// reaches what the profile started rather than only the shell.
	cmd.WaitDelay = 2 * time.Second
	detach(cmd)

	out, err := cmd.Output()
	if err != nil && len(out) == 0 {
		return "", false
	}

	// A profile prints things — greetings, version manager notices, warnings.
	// The answer is whichever line is an absolute path to something runnable,
	// taken from the end, since the command runs last.
	lines := strings.Split(string(out), "\n")
	for i := len(lines) - 1; i >= 0; i-- {
		candidate := strings.TrimSpace(lines[i])
		if candidate == "" {
			continue
		}
		// `command -v` answers with a bare word for a function, an alias or a
		// builtin. None of those is something this process can execute, and
		// running whatever a relative path happens to resolve to would be a
		// worse bug than the one being fixed.
		if !filepath.IsAbs(candidate) || !runnable(candidate) {
			continue
		}
		return candidate, true
	}
	return "", false
}
