package registry

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/goccy/go-yaml"
	"github.com/goccy/go-yaml/ast"
	"github.com/goccy/go-yaml/parser"
)

// ErrNoSuchProject reports an edit aimed at a path the registry does not hold.
var ErrNoSuchProject = errors.New("no such project in the registry")

// ErrAlreadyRegistered reports an attempt to register a folder twice.
var ErrAlreadyRegistered = errors.New("already registered")

// projectsKey is the sequence every edit works inside.
const projectsKey = "projects"

// Editing the registry goes through the YAML syntax tree rather than through
// the struct.
//
// This file is Muster's own, but it is also a file a person writes by hand and
// may have annotated. Marshalling the struct back would lose every comment,
// sort the keys alphabetically and change the list indentation - measured, not
// assumed. Printing a parsed tree keeps all three.
//
// Two limits of the library shape everything below, both found by trying it:
// a key parsed from a separate snippet and spliced into an existing mapping
// prints truncated, and replacing the whole sequence at once flattens the list
// indentation. So each operation touches one element, and an element that
// changes is rebuilt whole from its Entry.

// Add appends a project to the registry, creating the file if there is none.
//
// The path is stored as given - a leading ~ and all - because that is how a
// person writes it, and resolution already understands both forms.
func Add(path string, entry Entry) error {
	return edit(path, func(seq *ast.SequenceNode) error {
		target, err := expand(entry.Path)
		if err != nil {
			return err
		}
		for _, node := range seq.Values {
			existing, err := expand(pathOf(node))
			if err != nil {
				continue
			}
			if existing == target {
				return fmt.Errorf("%w: %s", ErrAlreadyRegistered, entry.Path)
			}
		}
		node, err := entryNode(entry)
		if err != nil {
			return err
		}
		seq.Values = append(seq.Values, node)
		return nil
	})
}

// Update replaces one project's entry, found by the path it is written under.
//
// The path identifies rather than an index: paths are unique by construction,
// and an index goes stale the moment the file is edited outside Muster.
func Update(path, projectPath string, entry Entry) error {
	return edit(path, func(seq *ast.SequenceNode) error {
		at, err := indexOf(seq, projectPath)
		if err != nil {
			return err
		}
		node, err := entryNode(entry)
		if err != nil {
			return err
		}
		// Only this entry loses a comment of its own, which is the limit of
		// preserving them: it is being rewritten.
		seq.Values[at] = node
		return nil
	})
}

// Remove deletes a project's entry. The folder on disk is untouched:
// unregistering is not deleting, and nothing in Muster deletes a backlog.
func Remove(path, projectPath string) error {
	return edit(path, func(seq *ast.SequenceNode) error {
		at, err := indexOf(seq, projectPath)
		if err != nil {
			return err
		}
		seq.Values = append(seq.Values[:at], seq.Values[at+1:]...)
		return nil
	})
}

// Move puts a project at a new position. Order in the file is the order on
// screen, so this is how the roll is arranged.
//
// The position is clamped rather than rejected: asking for one past the end is
// how a person says "last".
func Move(path, projectPath string, to int) error {
	return edit(path, func(seq *ast.SequenceNode) error {
		from, err := indexOf(seq, projectPath)
		if err != nil {
			return err
		}
		if to < 0 {
			to = 0
		}
		if to > len(seq.Values)-1 {
			to = len(seq.Values) - 1
		}
		if to == from {
			return nil
		}
		node := seq.Values[from]
		rest := append(seq.Values[:from:from], seq.Values[from+1:]...)
		seq.Values = append(rest[:to:to], append([]ast.Node{node}, rest[to:]...)...)
		return nil
	})
}

// edit applies one change to the projects sequence and writes the file back.
//
// A file that will not parse is refused rather than overwritten: an
// unparseable registry is a file someone is in the middle of editing, or one
// this version does not understand, and either way replacing it loses work.
func edit(path string, apply func(*ast.SequenceNode) error) error {
	raw, err := os.ReadFile(path)
	switch {
	case err == nil:
	case os.IsNotExist(err):
		raw = []byte(projectsKey + ":\n")
	default:
		return fmt.Errorf("reading %s: %w", path, err)
	}

	file, err := parser.ParseBytes(raw, parser.ParseComments)
	if err != nil {
		return fmt.Errorf("parsing %s: %w", path, err)
	}
	if len(file.Docs) == 0 || file.Docs[0].Body == nil {
		file, err = parser.ParseBytes([]byte(projectsKey+":\n"), parser.ParseComments)
		if err != nil {
			return fmt.Errorf("preparing %s: %w", path, err)
		}
	}

	body, ok := file.Docs[0].Body.(*ast.MappingNode)
	if !ok {
		return fmt.Errorf("%s: expected a mapping at the top level", path)
	}

	seq, err := projectsSequence(body)
	if err != nil {
		return fmt.Errorf("%s: %w", path, err)
	}
	if err := apply(seq); err != nil {
		return err
	}

	content := []byte(strings.TrimRight(file.String(), "\n") + "\n")
	// Never write a registry that cannot be read back. This is not
	// belt-and-braces: appending a block entry to a list written as
	// `projects: []` produced a file the parser rejected, and without this
	// check it went to disk.
	if err := readable(content); err != nil {
		return fmt.Errorf("%s would become unreadable: %w", path, err)
	}
	return write(path, content)
}

// projectsSequence finds the projects list, adding an empty one if the file
// has none - which is the shape of a registry that only sets wip_limits.
func projectsSequence(body *ast.MappingNode) (*ast.SequenceNode, error) {
	for _, value := range body.Values {
		if value.Key.String() != projectsKey {
			continue
		}
		switch node := value.Value.(type) {
		case *ast.SequenceNode:
			// A list written inline - `projects: []`, which is what an empty
			// registry looks like - cannot hold a block entry. Rebuilding it
			// as a block list is the only way to append one, and costs
			// nothing: an inline list is either empty or a single line.
			if node.IsFlowStyle {
				rebuilt, err := blockSequence(node)
				if err != nil {
					return nil, err
				}
				value.Value = rebuilt
				return rebuilt, nil
			}
			return node, nil
		case *ast.NullNode:
			// "projects:" with nothing under it.
			seq, err := emptySequence()
			if err != nil {
				return nil, err
			}
			value.Value = seq
			return seq, nil
		default:
			return nil, fmt.Errorf("projects is %T, not a list", value.Value)
		}
	}

	seq, err := emptySequence()
	if err != nil {
		return nil, err
	}
	pair, err := parser.ParseBytes([]byte(projectsKey+": []\n"), 0)
	if err != nil {
		return nil, err
	}
	added := pair.Docs[0].Body.(*ast.MappingNode).Values[0]
	added.Value = seq
	body.Values = append(body.Values, added)
	return seq, nil
}

// blockSequence rewrites an inline list as a block one, keeping its entries.
func blockSequence(flow *ast.SequenceNode) (*ast.SequenceNode, error) {
	var entries []Entry
	if err := yaml.NodeToValue(flow, &entries); err != nil {
		return nil, fmt.Errorf("reading the inline projects list: %w", err)
	}
	if len(entries) == 0 {
		return emptySequence()
	}
	node, err := yaml.ValueToNode(entries)
	if err != nil {
		return nil, err
	}
	seq, ok := node.(*ast.SequenceNode)
	if !ok {
		return nil, fmt.Errorf("expected a list, got %T", node)
	}
	return seq, nil
}

// readable checks that what is about to be written parses back into the shape
// the rest of the package expects.
func readable(content []byte) error {
	var file File
	return yaml.Unmarshal(content, &file)
}

func emptySequence() (*ast.SequenceNode, error) {
	node, err := yaml.ValueToNode([]Entry{})
	if err != nil {
		return nil, err
	}
	seq, ok := node.(*ast.SequenceNode)
	if !ok {
		return nil, fmt.Errorf("expected a list, got %T", node)
	}
	return seq, nil
}

// entryNode builds one list element, dropping the fields that are empty so an
// entry stays as short as a person would have written it.
func entryNode(entry Entry) (ast.Node, error) {
	written := struct {
		Path   string `yaml:"path"`
		Name   string `yaml:"name,omitempty"`
		Colour string `yaml:"color,omitempty"`
		Hidden bool   `yaml:"hidden,omitempty"`
	}{
		Path:   strings.TrimSpace(entry.Path),
		Name:   strings.TrimSpace(entry.Name),
		Colour: strings.TrimSpace(entry.Colour),
		Hidden: entry.Hidden,
	}
	if written.Path == "" {
		return nil, errors.New("a project entry needs a path")
	}
	return yaml.ValueToNode(written)
}

// indexOf finds an entry by the folder it points at, comparing resolved paths
// so ~/Dev/x and /home/me/Dev/x are the same project.
func indexOf(seq *ast.SequenceNode, projectPath string) (int, error) {
	target, err := expand(projectPath)
	if err != nil {
		return 0, err
	}
	for i, node := range seq.Values {
		written := pathOf(node)
		if written == "" {
			continue
		}
		resolved, err := expand(written)
		if err != nil {
			continue
		}
		if resolved == target {
			return i, nil
		}
	}
	return 0, fmt.Errorf("%w: %s", ErrNoSuchProject, projectPath)
}

// pathOf reads the path an element is written under, or "" if it has none.
func pathOf(node ast.Node) string {
	mapping, ok := node.(*ast.MappingNode)
	if !ok {
		if value, single := node.(*ast.MappingValueNode); single {
			// A one-key element parses as a value rather than a mapping.
			if value.Key.String() == "path" {
				return scalar(value.Value)
			}
		}
		return ""
	}
	for _, value := range mapping.Values {
		if value.Key.String() == "path" {
			return scalar(value.Value)
		}
	}
	return ""
}

func scalar(node ast.Node) string {
	if str, ok := node.(*ast.StringNode); ok {
		return str.Value
	}
	return strings.Trim(node.String(), `"'`)
}

// write replaces the file atomically, so an interrupted save cannot leave a
// half-written registry behind.
func write(path string, content []byte) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("creating %s: %w", dir, err)
	}
	temp, err := os.CreateTemp(dir, ".projects-*.yml")
	if err != nil {
		return fmt.Errorf("creating a temporary file in %s: %w", dir, err)
	}
	name := temp.Name()
	defer func() { _ = os.Remove(name) }()

	if _, err := temp.Write(content); err != nil {
		_ = temp.Close()
		return fmt.Errorf("writing %s: %w", name, err)
	}
	if err := temp.Close(); err != nil {
		return fmt.Errorf("closing %s: %w", name, err)
	}
	if err := os.Chmod(name, 0o644); err != nil {
		return fmt.Errorf("setting permissions on %s: %w", name, err)
	}
	if err := os.Rename(name, path); err != nil {
		return fmt.Errorf("replacing %s: %w", path, err)
	}
	return nil
}
