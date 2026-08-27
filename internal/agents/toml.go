package agents

import (
	"fmt"
	"os"
	"regexp"
	"strings"
)

// Editing one table of somebody else's TOML.
//
// Codex keeps its MCP servers in ~/.codex/config.toml, beside everything else it
// knows — trusted projects, models, personality. Round-tripping that through a
// TOML library would reformat a file the user wrote by hand, so this touches one
// table and nothing else: it finds [mcp_servers.muster], replaces its lines, and
// leaves every byte outside that table exactly as it was. What it cannot do is
// pretend to understand the rest of the file, and it does not try.

func sectionHeader(section string) *regexp.Regexp {
	// [mcp_servers.muster] or [mcp_servers."muster"], with any spacing
	parts := strings.Split(section, ".")
	for i, p := range parts {
		parts[i] = fmt.Sprintf(`(?:%s|"%s")`, regexp.QuoteMeta(p), regexp.QuoteMeta(p))
	}
	return regexp.MustCompile(`^\s*\[\s*` + strings.Join(parts, `\s*\.\s*`) + `\s*\]\s*$`)
}

var anySection = regexp.MustCompile(`^\s*\[`)

/** The file as it is and as it would be, with our table set or removed. */
func tomlChange(path, section string, lines []string, remove bool) (string, string, error) {
	raw, err := os.ReadFile(path)
	if err != nil && !os.IsNotExist(err) {
		return "", "", err
	}
	before := string(raw)
	head := sectionHeader(section)

	src := strings.Split(before, "\n")
	out := make([]string, 0, len(src)+len(lines)+2)
	found := false
	i := 0
	for i < len(src) {
		if !head.MatchString(src[i]) {
			out = append(out, src[i])
			i++
			continue
		}
		found = true
		// skip the old table: its header and everything up to the next one
		i++
		for i < len(src) && !anySection.MatchString(src[i]) {
			i++
		}
		if !remove {
			out = append(out, "["+section+"]")
			out = append(out, lines...)
		}
	}
	if !found && !remove {
		if len(out) > 0 && strings.TrimSpace(out[len(out)-1]) != "" {
			out = append(out, "")
		}
		out = append(out, "["+section+"]")
		out = append(out, lines...)
	}
	// a file that did not exist should not start with a blank line
	for len(out) > 0 && strings.TrimSpace(out[0]) == "" {
		out = out[1:]
	}
	after := strings.Join(out, "\n")
	if !strings.HasSuffix(after, "\n") {
		after += "\n"
	}
	if before != "" && !strings.HasSuffix(before, "\n") {
		before += "\n"
	}
	return before, after, nil
}

/** Whether our table is already in the file. */
func tomlHas(path, section string) bool {
	raw, err := os.ReadFile(path)
	if err != nil {
		return false
	}
	head := sectionHeader(section)
	for _, line := range strings.Split(string(raw), "\n") {
		if head.MatchString(line) {
			return true
		}
	}
	return false
}
