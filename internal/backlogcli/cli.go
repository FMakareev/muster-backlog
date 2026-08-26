// Package backlogcli runs the Backlog.md CLI.
//
// This is the only place in Muster that writes anything to a project. Reads go
// straight to the markdown, but the format is Backlog.md's to define: bodies
// are marked up with section comments, ordinal holds manual ordering, id
// generation handles collisions, and the CLI ships a doctor command precisely
// because none of that is trivial. A writer of our own would have to chase
// every CLI release and would be the first step towards owning the format.
//
// So there is no second writer. If the CLI cannot express something, that is a
// limit to work within or to raise upstream.
package backlogcli

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"sync"
	"time"
)

// BinaryName is the CLI executable Muster looks for on PATH.
const BinaryName = "backlog"

// MinimumVersion is the oldest CLI this was verified against. The on-disk
// format contract in the project documentation was derived from it.
//
// It stays at 1.48.0 rather than following the newest release, because that is
// the version the format was measured on and raising it would turn an
// unmeasured claim into a hard requirement on someone else's machine.
//
// RecommendedVersion is where the known differences stop mattering, and 1.48.0
// carries one that does: `draft promote` and `draft archive` exit 0 there when
// the id does not resolve, so a write that did nothing looks like a write that
// worked. Muster does not rely on the exit code for those - it checks that the
// note actually left the inbox - but the check exists because of this.
const (
	MinimumVersion     = "1.48.0"
	RecommendedVersion = "1.50.1"
)

// AtLeastRecommended reports whether the resolved CLI is new enough that none
// of the known older behaviours apply.
func (r *Runner) AtLeastRecommended() bool {
	return atLeast(r.version, RecommendedVersion)
}

// DefaultTimeout bounds a single command. The CLI is fast; anything that hangs
// this long is stuck, and an application that waits forever looks broken.
const DefaultTimeout = 30 * time.Second

// ErrNotInstalled reports that no usable CLI was found.
var ErrNotInstalled = errors.New("the backlog CLI was not found")

// ErrUnsupportedVersion reports that the CLI is older than Muster supports.
var ErrUnsupportedVersion = errors.New("the backlog CLI is too old")

// CommandError is a command that failed.
//
// It carries what was run and what the CLI said, because "exit status 1" tells
// a user nothing and gives the interface nothing to render.
type CommandError struct {
	// Args is the argument list, without the binary.
	Args []string
	// Dir is the project the command ran in.
	Dir string
	// ExitCode is the process exit code, or -1 when it never ran.
	ExitCode int
	// Stderr is what the CLI wrote, trimmed.
	Stderr string
	// Stdout is sometimes where the CLI explains itself instead.
	Stdout string
	// Err is the underlying failure.
	Err error
}

func (e *CommandError) Error() string {
	detail := strings.TrimSpace(e.Stderr)
	if detail == "" {
		detail = strings.TrimSpace(e.Stdout)
	}
	if detail == "" {
		detail = e.Err.Error()
	}
	return fmt.Sprintf("backlog %s: %s", strings.Join(e.Args, " "), detail)
}

func (e *CommandError) Unwrap() error { return e.Err }

// Runner executes CLI commands against projects.
//
// The zero value is not usable; call New.
type Runner struct {
	binary  string
	version string
	timeout time.Duration

	// locks serialises writes per project. Two concurrent edits inside one
	// project can collide over ids and ordinals; two different projects have no
	// reason to wait for each other.
	mu    sync.Mutex
	locks map[string]*sync.Mutex
}

// New locates the CLI and checks its version.
//
// It is called once at startup so that a missing or unsupported CLI is one
// clear report, rather than the same failure arriving again on every action a
// user tries.
func New() (*Runner, error) {
	binary, err := exec.LookPath(BinaryName)
	if err != nil {
		return nil, fmt.Errorf(
			"%w: install it and make sure %q is on PATH", ErrNotInstalled, BinaryName)
	}
	return newAt(binary)
}

// NewAt is New for an explicit binary path, used by tests and by a future
// setting for people whose CLI is not on PATH.
func NewAt(binary string) (*Runner, error) {
	return newAt(binary)
}

func newAt(binary string) (*Runner, error) {
	r := &Runner{binary: binary, timeout: DefaultTimeout, locks: map[string]*sync.Mutex{}}

	out, err := r.run(context.Background(), "", "--version")
	if err != nil {
		return nil, fmt.Errorf("%w: %s cannot be run: %w", ErrNotInstalled, binary, err)
	}

	r.version = strings.TrimSpace(out)
	if !atLeast(r.version, MinimumVersion) {
		return nil, fmt.Errorf(
			"%w: found %s, need %s or newer",
			ErrUnsupportedVersion, r.version, MinimumVersion)
	}
	return r, nil
}

// Version reports the CLI version found.
func (r *Runner) Version() string { return r.version }

// Binary reports the resolved executable path.
func (r *Runner) Binary() string { return r.binary }

// lockFor returns the write lock for one project.
func (r *Runner) lockFor(dir string) *sync.Mutex {
	r.mu.Lock()
	defer r.mu.Unlock()
	if lock, ok := r.locks[dir]; ok {
		return lock
	}
	lock := &sync.Mutex{}
	r.locks[dir] = lock
	return lock
}

// Exec runs a command in a project and returns its standard output.
//
// Arguments are passed as a slice to the process directly. Nothing is ever
// assembled into a shell string: task titles in real projects contain quotes,
// backticks, semicolons and dollar signs, and a shell would turn the next
// awkward title into a command.
func (r *Runner) Exec(ctx context.Context, dir string, args ...string) (string, error) {
	lock := r.lockFor(dir)
	lock.Lock()
	defer lock.Unlock()
	return r.run(ctx, dir, args...)
}

// run is Exec without the per-project lock, for calls that do not write.
func (r *Runner) run(ctx context.Context, dir string, args ...string) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, r.binary, args...)
	cmd.Dir = dir

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err == nil {
		return stdout.String(), nil
	}

	cmdErr := &CommandError{
		Args:     args,
		Dir:      dir,
		ExitCode: -1,
		Stderr:   stderr.String(),
		Stdout:   stdout.String(),
		Err:      err,
	}
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		cmdErr.ExitCode = exitErr.ExitCode()
	}
	if ctx.Err() != nil {
		cmdErr.Err = fmt.Errorf("timed out after %s: %w", r.timeout, ctx.Err())
	}
	return stdout.String(), cmdErr
}

// atLeast compares dotted version strings numerically, so 1.48.0 is newer than
// 1.9.0 rather than older as a string comparison would have it.
func atLeast(have, want string) bool {
	h := parseVersion(have)
	w := parseVersion(want)
	for i := range w {
		switch {
		case i >= len(h):
			return false
		case h[i] > w[i]:
			return true
		case h[i] < w[i]:
			return false
		}
	}
	return true
}

func parseVersion(v string) []int {
	v = strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(v), "v"))
	if i := strings.IndexAny(v, " -+"); i >= 0 {
		v = v[:i]
	}
	parts := strings.Split(v, ".")
	out := make([]int, 0, len(parts))
	for _, p := range parts {
		n := 0
		for _, c := range p {
			if c < '0' || c > '9' {
				break
			}
			n = n*10 + int(c-'0')
		}
		out = append(out, n)
	}
	return out
}
