import { atom, computed } from "nanostores";
import { Events } from "@wailsio/runtime";
import { BoardService } from "../../bindings/github.com/FMakareev/muster-backlog/internal/app";
import type {
  BoardLayout,
  DraftView,
  MilestoneView,
  Problem,
  ProjectView,
  QueryInput,
  TaskView,
} from "../../bindings/github.com/FMakareev/muster-backlog/internal/app/models";
import {
  TaskView as TaskViewMode,
  WindowBehaviour,
} from "../../bindings/github.com/FMakareev/muster-backlog/internal/settings/models";
import type { Settings } from "../../bindings/github.com/FMakareev/muster-backlog/internal/settings/models";

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

/** Every milestone across every project, so a card can show a name rather
 *  than a bare id that reads exactly like a task id. */
export const milestones = atom<MilestoneView[]>([]);

/**
 * Every waiting draft, oldest first.
 *
 * Kept alongside the tasks rather than fetched by the inbox screen, because
 * the navigation shows the depth from everywhere: an inbox you have to open to
 * discover is the inbox nobody reads.
 */
export const drafts = atom<DraftView[]>([]);

/** Muster's own preferences. */
export const settings = atom<Settings>({
  onWindowClose: WindowBehaviour.BehaviourQuit,
  taskView: TaskViewMode.ViewPanel,
  groupBy: "",
  wipLimits: {},
  staleAfterDays: 30,
  scalePercent: 100,
  backlogPath: "",
  lastProject: "",
});

/**
 * Apply the interface scale.
 *
 * Every size in the stylesheet is in rem, so setting the root font size moves
 * type, spacing, controls and the fixed chrome heights together. Changing only
 * the font size would leave large text in small boxes.
 */
settings.subscribe((prefs) => {
  const scale = (prefs.scalePercent || 100) / 100;
  document.documentElement.style.setProperty("--muster-scale", String(scale));
});

/**
 * A milestone by id or title, inside its own project.
 *
 * The list is passed in rather than read from the store here, so that a
 * component calling this subscribes to it and re-renders when the milestones
 * arrive. Reading the store inside would capture whatever was loaded at first
 * render and never update.
 */
export function milestoneOf(
  project: string,
  value: string,
  list: readonly MilestoneView[],
): MilestoneView | undefined {
  if (!value) return undefined;
  const needle = value.toLowerCase();
  return list.find(
    (m) =>
      m.project === project &&
      (m.id.toLowerCase() === needle || m.title.toLowerCase() === needle),
  );
}

/** How a milestone should read on screen: its title, or the raw value when it
 *  names nothing this project has. */
export function milestoneLabel(
  project: string,
  value: string,
  list: readonly MilestoneView[],
): string {
  return milestoneOf(project, value, list)?.title ?? value;
}
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
  const [
    nextProjects,
    nextTasks,
    nextProblems,
    nextLayout,
    nextMilestones,
    nextDrafts,
  ] = await Promise.all([
    BoardService.Projects(),
    BoardService.Tasks(allTasks()),
    BoardService.Problems(),
    BoardService.Layout(),
    BoardService.Milestones(),
    BoardService.Drafts(),
  ]);
  projects.set(nextProjects ?? []);
  tasks.set(nextTasks ?? []);
  problems.set(nextProblems ?? []);
  layout.set(nextLayout);
  milestones.set(nextMilestones ?? []);
  drafts.set(nextDrafts ?? []);
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
    bumpProjectChanged();
  });
  const offRegistry = Events.On("muster:registry:changed", () => {
    void refresh();
  });

  void BoardService.RegistryPath().then((path) => registryPath.set(path));
  void BoardService.Settings().then((prefs) => settings.set(prefs));
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

/** Let screens that fetch their own data know a project was reloaded. */
function bumpProjectChanged(): void {
  void import("./ui").then((ui) =>
    ui.projectChanged.set(ui.projectChanged.get() + 1),
  );
}
