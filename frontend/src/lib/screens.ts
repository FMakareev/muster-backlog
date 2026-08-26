/**
 * The screens the shell can show.
 *
 * Unbuilt screens are listed rather than hidden. The shape of the product is
 * useful information, and a nav that quietly grows over months tells a reader
 * nothing about where they are in it. Each one says plainly that it is not
 * built yet rather than pretending to be a working destination.
 */
export type ScreenID =
  "board" | "list" | "inbox" | "docs" | "stats" | "projects";

export interface Screen {
  id: ScreenID;
  label: string;
  /** The key that opens it. Shown in the interface, not hidden in a tooltip:
   *  in a keyboard-driven tool the shortcut is the label. */
  key: string;
  /** What it is for, shown when it is not built yet. */
  purpose: string;
  available: boolean;
}

export const screens: Screen[] = [
  {
    id: "board",
    label: "Board",
    key: "b",
    purpose: "Every project's tasks in one set of columns",
    available: true,
  },
  {
    id: "list",
    label: "List",
    key: "l",
    purpose: "A sortable table across every project",
    available: true,
  },
  {
    id: "inbox",
    label: "Inbox",
    key: "i",
    purpose: "Drafts waiting to be triaged",
    available: true,
  },
  {
    id: "docs",
    label: "Docs",
    key: "d",
    purpose: "Documents and decisions, rendered",
    available: true,
  },
  {
    id: "stats",
    label: "Stats",
    key: "s",
    purpose: "Counts, stale tasks and what is blocked",
    available: true,
  },
  {
    id: "projects",
    label: "Projects",
    key: "p",
    purpose: "The registry: add, arrange, hide and initialise projects",
    available: true,
  },
];

export function screenFor(key: string): Screen | undefined {
  return screens.find((s) => s.key === key && s.available);
}
