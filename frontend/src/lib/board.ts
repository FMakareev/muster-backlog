import { atom, computed } from "nanostores";
import { Events } from "@wailsio/runtime";
import { BoardService } from "../../bindings/github.com/FMakareev/muster-backlog/internal/app";
import type {
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

/** Columns for the board: the statuses every loaded project declares.
 *
 * This is a placeholder union that simply preserves first-seen order. The real
 * algorithm — which has to cope with projects whose status lists share no
 * order at all — belongs to TASK-27.
 */
export const columns = computed(projects, (list) => {
  const seen: string[] = [];
  for (const project of list) {
    for (const status of project.statuses ?? []) {
      if (!seen.includes(status)) seen.push(status);
    }
  }
  return seen;
});

/** Reload everything the UI shows from the backend. */
export async function refresh(): Promise<void> {
  const [nextProjects, nextTasks, nextProblems] = await Promise.all([
    BoardService.Projects(),
    BoardService.Tasks(allTasks()),
    BoardService.Problems(),
  ]);
  projects.set(nextProjects ?? []);
  tasks.set(nextTasks ?? []);
  problems.set(nextProblems ?? []);
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
