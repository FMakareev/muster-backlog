package backlog

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/FMakareev/muster-backlog/internal/project"
)

// Project is one scanned Backlog.md project, held entirely in memory. A
// thousand files parse in milliseconds, so nothing here needs a database.
type Project struct {
	Location project.Location
	Config   Config

	Tasks      []Entity
	Drafts     []Entity
	Milestones []Entity
	Documents  []Entity
	Decisions  []Entity

	// Diagnostics record files that were skipped and why. They are part of the
	// result, not a log line: the user has to be able to see them.
	Diagnostics []Diagnostic
}

// source describes one directory to walk.
type source struct {
	dir        string
	kind       Kind
	class      Class
	hasMarkers bool
	recursive  bool
}

// sources lists every directory `backlog init` creates. All of them are walked
// even though several are empty in every project the author has: an empty
// directory today is a populated one tomorrow.
//
// hasMarkers is false for milestones, documents and decisions. Their bodies are
// free markdown with no marker grammar, and a document may quote task markers
// in prose - one document in this very repository does exactly that, so
// applying the grammar to it would mis-parse it.
func sources() []source {
	return []source{
		{"tasks", KindTask, ClassActive, true, false},
		{"drafts", KindDraft, ClassActive, true, false},
		{"completed", KindTask, ClassCompleted, true, false},
		{filepath.Join("archive", "tasks"), KindTask, ClassArchived, true, false},
		{filepath.Join("archive", "drafts"), KindDraft, ClassArchived, true, false},
		{filepath.Join("archive", "milestones"), KindMilestone, ClassArchived, false, false},
		{"milestones", KindMilestone, ClassActive, false, false},
		// Documents support nested subdirectories through the CLI's -p flag.
		{"docs", KindDocument, ClassActive, false, true},
		{"decisions", KindDecision, ClassActive, false, false},
	}
}

// Scan reads a whole project.
//
// It returns an error only when the project itself cannot be read - a missing
// config, an unreadable data directory. Individual files never fail a scan:
// anything that is not an entity becomes a diagnostic and the walk continues.
func Scan(loc project.Location) (*Project, error) {
	cfg, err := LoadConfig(loc.ConfigPath)
	if err != nil {
		return nil, err
	}

	p := &Project{Location: loc, Config: cfg}
	for _, src := range sources() {
		p.walk(src)
	}
	sortEntities(p.Tasks)
	sortEntities(p.Drafts)
	sortEntities(p.Milestones)
	sortEntities(p.Documents)
	sortEntities(p.Decisions)
	return p, nil
}

func (p *Project) walk(src source) {
	dir := p.Location.Sub(src.dir)

	entries, err := os.ReadDir(dir)
	if err != nil {
		// A missing directory is ordinary: not every project has every one.
		if !os.IsNotExist(err) {
			p.Diagnostics = append(p.Diagnostics, Diagnostic{
				Path: dir, Reason: fmt.Sprintf("cannot read directory: %v", err),
			})
		}
		return
	}

	for _, entry := range entries {
		path := filepath.Join(dir, entry.Name())
		if entry.IsDir() {
			if src.recursive {
				nested := src
				nested.dir = filepath.Join(src.dir, entry.Name())
				p.walk(nested)
			}
			continue
		}
		if !strings.EqualFold(filepath.Ext(entry.Name()), ".md") {
			// .gitkeep and editor leftovers are normal and not worth reporting.
			continue
		}

		raw, err := os.ReadFile(path)
		if err != nil {
			p.Diagnostics = append(p.Diagnostics, Diagnostic{
				Path: path, Reason: fmt.Sprintf("cannot read file: %v", err),
			})
			continue
		}

		e, err := ParseFile(path, raw, src.kind, src.class, src.hasMarkers)
		if err != nil {
			// A README or a note sitting in an entity directory is not an
			// error in the project; it is simply not an entity.
			p.Diagnostics = append(p.Diagnostics, Diagnostic{
				Path: path, Reason: err.Error(),
			})
			continue
		}
		p.add(src.kind, e)
	}
}

func (p *Project) add(kind Kind, e Entity) {
	switch kind {
	case KindTask:
		p.Tasks = append(p.Tasks, e)
	case KindDraft:
		p.Drafts = append(p.Drafts, e)
	case KindMilestone:
		p.Milestones = append(p.Milestones, e)
	case KindDocument:
		p.Documents = append(p.Documents, e)
	case KindDecision:
		p.Decisions = append(p.Decisions, e)
	}
}

// sortEntities orders entities the way Backlog.md's own comparator does:
// anything with an ordinal comes before anything without, then by ordinal, then
// by id. Ordinals are neither mandatory nor unique, so both fallbacks are load
// bearing rather than defensive.
func sortEntities(list []Entity) {
	sort.SliceStable(list, func(i, j int) bool {
		a, b := list[i], list[j]
		if a.HasOrdinal != b.HasOrdinal {
			return a.HasOrdinal
		}
		if a.HasOrdinal && b.HasOrdinal && *a.Ordinal != *b.Ordinal {
			return *a.Ordinal < *b.Ordinal
		}
		return lessID(a.ID, b.ID)
	})
}

// lessID compares ids numerically where it can, so TASK-9 sorts before TASK-10
// rather than after it. Zero padding varies by project, so string comparison
// alone gives the wrong order in some projects and the right one in others.
func lessID(a, b string) bool {
	ap, an := splitID(a)
	bp, bn := splitID(b)
	if !strings.EqualFold(ap, bp) {
		return strings.ToLower(ap) < strings.ToLower(bp)
	}
	if an != bn {
		return an < bn
	}
	return a < b
}

// splitID separates an id's prefix from its leading number. Subtask ids carry a
// dotted child segment, which is compared as part of the string tail.
func splitID(id string) (prefix string, number float64) {
	i := strings.LastIndex(id, "-")
	if i < 0 {
		return id, 0
	}
	prefix = id[:i]
	tail := id[i+1:]

	var whole, frac float64
	var scale float64 = 1
	seenDot := false
	for _, r := range tail {
		switch {
		case r >= '0' && r <= '9':
			if seenDot {
				scale /= 10
				frac += float64(r-'0') * scale
			} else {
				whole = whole*10 + float64(r-'0')
			}
		case r == '.' && !seenDot:
			seenDot = true
		default:
			return prefix, whole + frac
		}
	}
	return prefix, whole + frac
}
