import { atom, computed } from "nanostores";
import { Events } from "@wailsio/runtime";
import { BoardService } from "../../bindings/github.com/FMakareev/muster-backlog/internal/app";
import type {
  BoardLayout,
  Problem,
  ProjectView,
  QueryInput,
  TaskView,
} from "../../bindings/github.com/FMakareev/muster-backlog/internal/app/models";

/**
 * Backend state, mirrored into nanostores.
 *
 * Nothing here polls. The backend watches the filesystem and emits an event
 * when a project has been reloaded; this module listens and refreshes. Types
 * come from the generated bindings, so there is one definition of a task rather
 * than one on each side of the bridge waiting to drift apart.
 */

/** An unfiltered query. Every field of the generated type is required, so the
 * empty case is spelled out once here rather than cast away at each call. */
export function allTasks(): QueryInput {
  return {
    kinds: null,
    classes: null,
    projects: null,
    statuses: null,
    milestones: null,
    priorities: null,
    types: null,
    labels: null,
    assignees: null,
    text: "",
  };
}

export const projects = atom<ProjectView[]>([]);
export const tasks = atom<TaskView[]>([]);
export const problems = atom<Problem[]>([]);
export const registryPath = atom<string>("");
export const loading = atom<boolean>(true);

/**
 * The board's columns, resolved by the backend.
 *
 * Statuses are per-project configuration and projects do not agree, so the
 * columns are the union of every declared list and the ordering is a weighted
 * vote across projects. That logic lives in Go where it is tested, rather than
 * being re-derived here from the raw lists.
 */
export const layout = atom<BoardLayout>({ columns: null, conflicts: null });

export const columns = computed(layout, (l) => l.columns ?? []);

/** What one project configures, for the controls that edit a task. */
export function projectConfig(path: string): ProjectView | undefined {
  return projects.get().find((p) => p.path === path);
}

/** Whether a task in a project may take a status. */
export function canMove(project: string, status: string): boolean {
  const column = (layout.get().columns ?? []).find((c) => c.name === status);
  return column ? (column.projects ?? []).includes(project) : false;
}

/** Reload everything the UI shows from the backend. */
export async function refresh(): Promise<void> {
  const [nextProjects, nextTasks, nextProblems, nextLayout] = await Promise.all(
    [
      BoardService.Projects(),
      BoardService.Tasks(allTasks()),
      BoardService.Problems(),
      BoardService.Layout(),
    ],
  );
  projects.set(nextProjects ?? []);
  tasks.set(nextTasks ?? []);
  problems.set(nextProblems ?? []);
  layout.set(nextLayout);
  loading.set(false);
}

/**
 * Start mirroring backend state.
 *
 * Returns a function that stops listening, so a caller can tear this down
 * without leaving a subscription behind.
 */
export function connect(): () => void {
  const offProject = Events.On("muster:project:changed", () => {
    void refresh();
  });
  const offRegistry = Events.On("muster:registry:changed", () => {
    void refresh();
  });

  void BoardService.RegistryPath().then((path) => registryPath.set(path));
  void refresh();

  return () => {
    offProject();
    offRegistry();
  };
}

/**
 * Apply a write and take whatever the files say afterwards.
 *
 * Every editing control goes through this, so none of them can quietly assume
 * the write succeeded. A failure is reported where the person is working.
 */
export async function applyWrite(
  run: () => Promise<{ ok: boolean; problem: Problem | null }>,
  whenFailed: (message: string) => void,
): Promise<void> {
  const result = await run();
  if (!result.ok && result.problem) {
    whenFailed(`${result.problem.title}: ${result.problem.detail}`);
  }
  await refresh();
}
