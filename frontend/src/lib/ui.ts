import { atom, computed } from "nanostores";
import { BoardService } from "../../bindings/github.com/FMakareev/muster-backlog/internal/app";
import { TaskView as TaskViewMode } from "../../bindings/github.com/FMakareev/muster-backlog/internal/settings/models";
import { settings as settingsStore, tasks } from "./board";
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
export const selectedTask = computed([tasks, selected], (all, ref) => {
  if (!ref) return null;
  return (
    all.find(
      (task) =>
        task.project === ref.project &&
        task.kind === ref.kind &&
        task.class === ref.class &&
        task.id === ref.id,
    ) ?? null
  );
});

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
 * The tasks currently on screen.
 *
 * Derived in one place so that nothing can disagree about it: the board and the
 * status strip both read this, rather than each filtering for themselves and
 * drifting apart the moment one of them changes.
 */
export const visibleTasks = computed([tasks, focusedProject], (all, focused) =>
  focused ? all.filter((task) => task.project === focused) : all,
);
