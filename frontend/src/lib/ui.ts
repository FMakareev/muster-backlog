import { atom, computed } from "nanostores";
import { tasks } from "./board";
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
 * The tasks currently on screen.
 *
 * Derived in one place so that nothing can disagree about it: the board and the
 * status strip both read this, rather than each filtering for themselves and
 * drifting apart the moment one of them changes.
 */
export const visibleTasks = computed([tasks, focusedProject], (all, focused) =>
  focused ? all.filter((task) => task.project === focused) : all,
);
