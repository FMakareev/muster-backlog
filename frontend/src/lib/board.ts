import { atom } from "nanostores";
import type { ColumnConfig, KanbanCard } from "@svar-ui/svelte-kanban";

/**
 * Placeholder board state.
 *
 * This exists only to prove that the three integrations hold together: a
 * nanostore feeds SVAR Kanban, which renders inside the Tailwind shell. The
 * real store is built in TASK-17 and fed by the Go backend; the shape below
 * deliberately mirrors what a multi-project board will need, so replacing the
 * source does not mean rewriting the consumer.
 */
export type BoardCard = KanbanCard & {
  /** Which registered project the task came from. */
  project: string;
};

/**
 * Columns are the union of the status lists declared by the registered
 * projects (see TASK-27). Hardcoded here until projects are actually loaded.
 */
export const columns = atom<ColumnConfig[]>([
  { id: "todo", label: "To Do" },
  { id: "doing", label: "In Progress" },
  { id: "review", label: "In Review" },
  { id: "done", label: "Done" },
]);

export const cards = atom<BoardCard[]>([
  {
    id: 1,
    label: "Scaffold Wails v3 application skeleton",
    column: "doing",
    project: "muster-backlog",
  },
  {
    id: 2,
    label: "Map the Backlog.md on-disk format contract",
    column: "doing",
    project: "muster-backlog",
  },
  {
    id: 3,
    label: "Parse Backlog.md task files into a domain model",
    column: "todo",
    project: "muster-backlog",
  },
]);
