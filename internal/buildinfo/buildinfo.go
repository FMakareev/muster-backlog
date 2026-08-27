// Package buildinfo answers what this build is.
//
// It sits in its own package because two callers need it and neither can
// import the other: the command, for `muster --version`, and the board
// service, so the interface can show it to somebody writing a bug report.
package buildinfo

import (
	"fmt"
	"runtime"
	"runtime/debug"
	"strings"
)

// stamped is set at build time from build/config.yml, which is the one
// place the version is written and the file the release automation updates.
//
//	-ldflags "-X github.com/FMakareev/muster-backlog/internal/buildinfo.stamped=1.2.3"
//
// It is deliberately not a constant: a build that forgets the flag should say
// so rather than claim a number nobody released.
var stamped = ""

// Version is what the binary reports about itself.
//
// A bug report is worth what its version line is worth, so this never invents
// one. With no stamp it answers "dev" — and, when the toolchain recorded a
// revision, the commit as well, which is the most a build from a working tree
// can honestly say.
func Version() string {
	if v := strings.TrimSpace(stamped); v != "" {
		return v
	}

	// The build passes -buildvcs=false, so this is usually absent. When
	// somebody runs `go build` or `go install` by hand it is not, and a
	// commit is far better than nothing to a person reading a bug report.
	if info, ok := debug.ReadBuildInfo(); ok {
		for _, setting := range info.Settings {
			if setting.Key != "vcs.revision" || setting.Value == "" {
				continue
			}
			revision := setting.Value
			if len(revision) > 12 {
				revision = revision[:12]
			}
			return "dev+" + revision
		}
	}
	return "dev"
}

// Line is what `muster --version` prints.
//
// The Go version and the platform are on it because the two questions that
// follow "which version?" in a bug report are "built with what?" and "running
// where?", and asking a person to find those out is asking them to give up.
func Line() string {
	return fmt.Sprintf("muster %s (%s, %s/%s)",
		Version(), runtime.Version(), runtime.GOOS, runtime.GOARCH)
}
