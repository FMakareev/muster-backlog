import { atom, computed } from "nanostores";
import { BoardService } from "../../bindings/github.com/FMakareev/muster-backlog/internal/app";
import { TaskView as TaskViewMode } from "../../bindings/github.com/FMakareev/muster-backlog/internal/settings/models";
import { drafts, projects, settings as settingsStore, tasks } from "./board";
import type { ScreenID } from "./screens";

/** Which screen is showing. */
export const screen = atom<ScreenID>("board");

/**
 * Which project is focused in the roll, by path.
 *
 * Empty means all of them, which is the point of the application and therefore
 * the default. Focusing one is a temporary narrowing, not a mode.
 */
export const focusedProject = atom<string>("");

export function toggleProject(path: string): void {
  focusedProject.set(focusedProject.get() === path ? "" : path);
}

/**
 * Which project a form starts on.
 *
 * The rule, wherever something has to be created somewhere: the project the
 * person is already looking at, and the first registered one when they are
 * looking at all of them. Never nothing - a picker that starts empty makes
 * every creation two decisions instead of one, and the second is a decision
 * the interface already knows the answer to.
 *
 * Broken projects are skipped: they cannot be written to.
 *
 * Anything else that has to choose a project should call this rather than
 * decide again.
 */
export function defaultProject(): string {
  const healthy = projects.get().filter((p) => p.ok);
  const usable = (path: string) =>
    path !== "" && healthy.some((p) => p.path === path);

  const focused = focusedProject.get();
  if (usable(focused)) return focused;

  // Then wherever the last note went. Capture only pays if it costs nothing,
  // and choosing the same project again is part of that - but it sits below
  // the focused one, because someone looking at a project means that one.
  const last = settingsStore.get().lastProject;
  if (usable(last)) return last;

  return healthy[0]?.path ?? "";
}

/**
 * Remember where a note was captured, for the next one.
 *
 * Written to the preferences rather than held in memory: this is a tool that
 * gets restarted, and a habit that resets every morning is not a habit.
 */
export async function rememberProject(path: string): Promise<void> {
  const current = settingsStore.get();
  if (!path || current.lastProject === path) return;
  const next = { ...current, lastProject: path };
  settingsStore.set(next);
  await BoardService.SaveSettings(next);
}

/**
 * The task the panel is showing, as a reference rather than a copy.
 *
 * Holding a reference is what keeps the panel live: the content is derived
 * from the same task list the board reads, so when the watcher reloads a
 * project the panel follows without asking for anything.
 */
export interface TaskRef {
  project: string;
  kind: string;
  class: string;
  id: string;
}

/**
 * What cards are grouped by inside each column: nothing, project or milestone.
 *
 * SVAR's open edition has no swimlanes, so grouping is expressed by ordering:
 * a column's cards clump instead of interleaving. On a board over nine
 * projects that is the difference between reading a column and scanning it.
 * Milestones are the other axis this backlog is actually planned on.
 */
export const groupBy = computed(settingsStore, (s) => s.groupBy);

const groupCycle = ["", "project", "milestone"] as const;

/** Cycle through the grouping options and remember the choice. */
export async function cycleGrouping(): Promise<void> {
  const current = settingsStore.get();
  const at = groupCycle.indexOf(current.groupBy as (typeof groupCycle)[number]);
  const next = {
    ...current,
    groupBy: groupCycle[(at + 1) % groupCycle.length],
  };
  settingsStore.set(next);
  await BoardService.SaveSettings(next);
}

/** Whether the preferences panel is open. */
export const showSettings = atom<boolean>(false);

/** Whether the create-task form is open. */
export const showNewTask = atom<boolean>(false);

export function openNewTask(): void {
  showNewTask.set(true);
}

export function closeNewTask(): void {
  showNewTask.set(false);
}

/** Whether the problems panel is open. */
export const showProblems = atom<boolean>(false);

export const selected = atom<TaskRef | null>(null);

export function openTask(ref: TaskRef): void {
  selected.set(ref);
}

/**
 * The document or decision the viewer is showing.
 *
 * Kept here rather than inside the viewer so that a search result can open one
 * from anywhere: a hit is a hit whether it is a task or a document, and having
 * only tasks openable made half the search results dead ends.
 */
export const selectedDoc = atom<TaskRef | null>(null);

/**
 * Open whatever a reference points at, on the screen that shows it.
 *
 * Tasks and drafts belong in the panel; documents, decisions and milestones
 * belong in the viewer. Routing by kind here means every caller - search, a
 * dependency link, a list row - can just say "open this".
 */
export function openEntity(ref: TaskRef): void {
  if (
    ref.kind === "document" ||
    ref.kind === "decision" ||
    ref.kind === "milestone"
  ) {
    selectedDoc.set(ref);
    screen.set("docs");
    return;
  }
  openTask(ref);
}

export function closeTask(): void {
  selected.set(null);
}

/**
 * Switch between the side panel and the centred reading view.
 *
 * On a wide screen a column against the edge puts the text a long way from
 * where the eye rests, so which one is used is a preference rather than a mode
 * to rediscover each time.
 */
export async function toggleTaskView(): Promise<void> {
  const current = settingsStore.get();
  const next = {
    ...current,
    taskView:
      current.taskView === TaskViewMode.ViewCentred
        ? TaskViewMode.ViewPanel
        : TaskViewMode.ViewCentred,
  };
  settingsStore.set(next);
  await BoardService.SaveSettings(next);
}

/** The open task, or null. Derived, so it follows a reload. */
export const selectedTask = computed(
  [tasks, drafts, selected],
  (all, waiting, ref) => {
    if (!ref) return null;
    // Drafts are searched too, so a note opens in the panel and can be read
    // whole rather than as the two lines the inbox has room for.
    const found = [...all, ...waiting].find(
      (task) =>
        task.project === ref.project &&
        task.kind === ref.kind &&
        task.class === ref.class &&
        task.id === ref.id,
    );
    return found ?? null;
  },
);

/**
 * Find a task by id inside one project.
 *
 * Ids collide across projects - 200 of 351 in the author's own repositories -
 * so a reference resolves inside its own project and nowhere else.
 */
export function findInProject(project: string, id: string) {
  return (
    tasks
      .get()
      .find(
        (task) =>
          task.project === project &&
          task.id.toLowerCase() === id.toLowerCase(),
      ) ?? null
  );
}

/**
 * Board filters.
 *
 * Without them the board shows every open task at once, which is the dump
 * problem that disqualified the table alternative. Kept in one object so the
 * whole set clears in one action and can be saved as a view.
 */
export interface Filters {
  statuses: string[];
  priorities: string[];
  types: string[];
  labels: string[];
  milestones: string[];
  text: string;
}

export const emptyFilters: Filters = {
  statuses: [],
  priorities: [],
  types: [],
  labels: [],
  milestones: [],
  text: "",
};

export const filters = atom<Filters>({ ...emptyFilters });

export const filtersActive = computed(
  filters,
  (f) =>
    f.statuses.length +
      f.priorities.length +
      f.types.length +
      f.labels.length +
      f.milestones.length >
      0 || f.text.trim() !== "",
);

export function clearFilters(): void {
  filters.set({ ...emptyFilters });
}

export function toggleFilter(field: keyof Filters, value: string): void {
  const current = filters.get();
  const list = current[field];
  if (!Array.isArray(list)) return;
  filters.set({
    ...current,
    [field]: list.includes(value)
      ? list.filter((v) => v !== value)
      : [...list, value],
  });
}

/**
 * Bumped whenever a project has been reloaded.
 *
 * Screens that fetch their own data - the documents viewer, the overview -
 * watch this instead of subscribing to every task, so they stay as live as the
 * board without holding a second copy of it.
 */
export const projectChanged = atom<number>(0);

/** Whether search is open. */
export const showSearch = atom<boolean>(false);

/** Whether the filter panel is open. */
export const showFilters = atom<boolean>(false);

/**
 * The tasks currently on screen.
 *
 * Derived in one place so that nothing can disagree about it: the board and the
 * status strip both read this, rather than each filtering for themselves and
 * drifting apart the moment one of them changes.
 */
export const visibleTasks = computed(
  [tasks, focusedProject, filters],
  (all, focused, f) => {
    const matches = (values: string[], value: string) =>
      values.length === 0 ||
      values.some((v) => v.toLowerCase() === (value ?? "").toLowerCase());

    const needle = f.text.trim().toLowerCase();

    return all.filter((task) => {
      if (focused && task.project !== focused) return false;
      const e = task.entity;
      if (!matches(f.statuses, e.Status)) return false;
      if (!matches(f.priorities, e.Priority)) return false;
      if (!matches(f.types, e.Type)) return false;
      if (!matches(f.milestones, e.Milestone)) return false;
      if (
        f.labels.length > 0 &&
        !(e.Labels ?? []).some((l) =>
          f.labels.some((w) => w.toLowerCase() === l.toLowerCase()),
        )
      ) {
        return false;
      }
      if (needle && !e.Title.toLowerCase().includes(needle)) return false;
      return true;
    });
  },
);
