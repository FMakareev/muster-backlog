#!/usr/bin/env bash
# A demo corpus for the screenshots.
#
# Entirely invented: five projects that do not exist, about work nobody is
# doing. Screenshots of a real backlog would put someone's private planning on
# a public README, and that is the criterion this has to satisfy.
#
# Shaped to show what the product is for rather than to look busy: several
# projects at once, statuses that differ between them, ids that collide,
# milestones, subtasks, drafts waiting, and enough finished work for the
# figures to have something to say.
set -euo pipefail

# The corpus is built inside a home directory of its own, and the application
# is run with HOME pointed at it. That is not tidiness: the status bar and the
# Projects screen both show paths, so a screenshot taken any other way puts a
# scratch directory or a username on the README.
B="${1:?usage: build.sh <directory>}"
rm -rf "$B"
mkdir -p "$B/projects" "$B/.config/muster"

new() { # new <dir> <name> [statuses]
  mkdir -p "$B/projects/$1"
  (
    cd "$B/projects/$1"
    git init -q .
    backlog init "$2" --defaults >/dev/null 2>&1
  )
}

task() { # task <dir> <title> [args...]
  local dir="$1"; shift
  local title="$1"; shift
  (cd "$B/projects/$dir" && backlog task create "$title" "$@" >/dev/null 2>&1)
}

draft() {
  local dir="$1"; shift
  (cd "$B/projects/$dir" && backlog task create "$1" --draft "${@:2}" >/dev/null 2>&1)
}

# ---------------------------------------------------------------- atlas
new atlas "Atlas"
(cd "$B/projects/atlas"
  backlog milestone add "Public beta" -d "Everything that has to be true before strangers use it" >/dev/null 2>&1
  backlog milestone add "Offline first" >/dev/null 2>&1
)
task atlas "Import a route from a GPX file" --priority high -m m-0 -l import,parser -a @rosa
task atlas "Cache map tiles for a saved route" --priority high -m m-1 -l offline
task atlas "Show elevation alongside the track" --priority medium -m m-0 -l ui
task atlas "Decide how far ahead to prefetch" --priority medium -m m-1 -l offline,spike
task atlas "Rename a saved route" --priority low -l ui
task atlas "Handle a GPX file with no timestamps" --priority high -m m-0 -l import,bug
task atlas "Warn before overwriting a route" -l ui -a @rosa
(cd "$B/projects/atlas"
  backlog task edit TASK-1 -s "In Progress" >/dev/null 2>&1
  backlog task edit TASK-6 -s "In Progress" >/dev/null 2>&1
  backlog task edit TASK-3 -s Done >/dev/null 2>&1
  backlog task edit TASK-5 -s Done >/dev/null 2>&1
  backlog task create "Read the header without loading the whole file" -p TASK-1 -l parser >/dev/null 2>&1
  backlog task create "Reject a file that is not GPX at all" -p TASK-1 -l parser >/dev/null 2>&1
  backlog doc create "Tile cache sizing" -t specification >/dev/null 2>&1
  backlog decision create "Store tiles on disk rather than in SQLite" >/dev/null 2>&1
)
draft atlas "Something about sharing a route by link"
draft atlas "Elevation profile is unreadable on a phone"

# ---------------------------------------------------------------- beacon
new beacon "Beacon"
(cd "$B/projects/beacon"
  backlog milestone add "First alerts" >/dev/null 2>&1
)
task beacon "Send a digest instead of one message per event" --priority high -m m-0 -l notifications
task beacon "Let a rule be muted for an hour" --priority medium -m m-0 -l notifications
task beacon "Stop retrying a webhook that keeps refusing" --priority high -l reliability,bug
task beacon "Show which rule fired, on the alert itself" --priority medium -l ui
task beacon "Back off when the upstream is rate limiting" --priority high -l reliability
(cd "$B/projects/beacon"
  backlog task edit TASK-1 -s "In Progress" -a @dai >/dev/null 2>&1
  backlog task edit TASK-3 -s "In Progress" >/dev/null 2>&1
  backlog task edit TASK-4 -s Done >/dev/null 2>&1
)
draft beacon "Digest should probably be per rule, not per channel"

# ------------------------------------------------------- cinder (its own statuses)
new cinder "Cinder"
(cd "$B/projects/cinder"
  # `backlog config set statuses` refuses, and says so: statuses are edited in
  # the project config file directly. One project with a status the others do
  # not have is the point - the board's columns are the union across projects.
  sed -i 's/statuses: \["To Do", "In Progress", "Done"\]/statuses: ["To Do", "In Progress", "In Review", "Done"]/' backlog/config.yml
)
task cinder "Replace the hand-written CSV reader" --priority high -l refactor
task cinder "Column types should follow the data, not the header" --priority high -l parser
task cinder "Report the row a failure happened on" --priority medium -l errors
task cinder "Accept a file whose separator is a semicolon" --priority low -l parser
(cd "$B/projects/cinder"
  backlog task edit TASK-1 -s "In Progress" >/dev/null 2>&1
  backlog task edit TASK-2 -s "In Review" >/dev/null 2>&1
  backlog task edit TASK-3 -s Done >/dev/null 2>&1
  backlog doc create "What the type inference does" -t guide >/dev/null 2>&1
)

# ---------------------------------------------------------------- drift
new drift "Drift"
task drift "Keep the sidebar width across restarts" --priority low -l ui
task drift "Two windows must not fight over the same file" --priority high -l bug
task drift "Remember the last folder opened" --priority medium -l ui
(cd "$B/projects/drift"
  backlog task edit TASK-2 -s "In Progress" >/dev/null 2>&1
  backlog task edit TASK-1 -s Done >/dev/null 2>&1
)
draft drift "Sidebar could collapse to icons"

# ---------------------------------------------------------------- ember
new ember "Ember"
(cd "$B/projects/ember"
  backlog milestone add "Cut the boot time" >/dev/null 2>&1
)
task ember "Measure where the first second goes" --priority high -m m-0 -l perf,spike
task ember "Load the theme before anything paints" --priority medium -m m-0 -l perf
task ember "Stop parsing the whole index at startup" --priority high -m m-0 -l perf
(cd "$B/projects/ember"
  backlog task edit TASK-1 -s Done >/dev/null 2>&1
  backlog task edit TASK-3 -s "In Progress" >/dev/null 2>&1
)

cat > "$B/.config/muster/projects.yml" <<YML
projects:
  - path: $B/projects/atlas
    name: Atlas
  - path: $B/projects/beacon
    name: Beacon
  - path: $B/projects/cinder
    name: Cinder
  - path: $B/projects/drift
    name: Drift
  - path: $B/projects/ember
    name: Ember
YML

echo "built:"
for p in atlas beacon cinder drift ember; do
  printf "  %-8s %s tasks\n" "$p" "$(ls "$B/projects/$p/backlog/tasks" 2>/dev/null | wc -l)"
done

# ------------------------------------------------------------------ age
#
# The figures measure age, staleness and what is blocked. A corpus created
# five minutes ago reports zero for all three, which demonstrates nothing.
#
# Dependencies go in through the CLI, which has a flag for them. Dates do not:
# no command sets a creation date, so they are rewritten in the files here.
# That is only defensible because this corpus is disposable scaffolding for a
# screenshot — never do it to a real backlog.
(cd "$B/projects/atlas"
  backlog task edit TASK-2 --dep TASK-4 >/dev/null 2>&1
  backlog task edit TASK-7 --dep TASK-1 >/dev/null 2>&1
)
(cd "$B/projects/beacon" && backlog task edit TASK-2 --dep TASK-1 >/dev/null 2>&1)

python3 - "$B" <<'PY'
import datetime, pathlib, random, re, sys

root = pathlib.Path(sys.argv[1]) / "projects"
# A fixed seed, so the same corpus comes out of the same script.
rng = random.Random(20260827)
today = datetime.date(2026, 8, 27)

for task in sorted(root.glob("*/backlog/*/*.md")):
    text = task.read_text()
    age = rng.choice([2, 5, 9, 14, 21, 34, 47, 63, 78, 96])
    created = today - datetime.timedelta(days=age)
    # Updated later than created, and sometimes long enough ago to be stale.
    updated = created + datetime.timedelta(days=rng.choice([0, 1, 3, 8, 20]))
    if updated > today:
        updated = today
    text = re.sub(r"created_date: '[^']*'",
                  f"created_date: '{created} {rng.randrange(9, 18):02d}:{rng.randrange(0, 60):02d}'", text)
    text = re.sub(r"updated_date: '[^']*'",
                  f"updated_date: '{updated} {rng.randrange(9, 18):02d}:{rng.randrange(0, 60):02d}'", text)
    task.write_text(text)
print("dated the corpus")
PY
