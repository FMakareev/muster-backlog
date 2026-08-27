<script lang="ts">
  import { milestoneLabel, milestones } from "./board";
  import BulkBar from "./BulkBar.svelte";
  import { projectColour } from "./colour";
  import {
    chosen,
    chosenTasks,
    chooseAll,
    clearChosen,
    openTask,
    refKey,
    toggleChosen,
    visibleTasks,
  } from "./ui";

  /**
   * The task list.
   *
   * The mode for scanning many tasks at once, where the board is the mode for
   * moving a few. Sortable on every column, and the visible column set is
   * remembered so a working view survives a restart.
   */

  /** Blank for the zero time, which is what "no date" parses to. */
  function shortDate(value: string): string {
    if (!value || value.startsWith("0001-01-01")) return "";
    return value.slice(0, 10);
  }

  type Column = {
    key: string;
    label: string;
    value: (t: (typeof $visibleTasks)[number]) => string;
    mono?: boolean;
    width?: string;
  };

  const allColumns: Column[] = [
    {
      key: "project",
      label: "Project",
      value: (t) => t.projectName,
      width: "w-32",
    },
    { key: "id", label: "ID", value: (t) => t.id, mono: true, width: "w-24" },
    { key: "title", label: "Title", value: (t) => t.entity.Title },
    {
      key: "status",
      label: "Status",
      value: (t) => t.entity.Status,
      width: "w-28",
    },
    {
      key: "priority",
      label: "Priority",
      value: (t) => t.entity.Priority ?? "",
      width: "w-20",
    },
    {
      key: "type",
      label: "Type",
      value: (t) => t.entity.Type ?? "",
      width: "w-24",
    },
    {
      key: "milestone",
      label: "Milestone",
      value: (t) =>
        t.entity.Milestone
          ? milestoneLabel(t.project, t.entity.Milestone, $milestones)
          : "",
      width: "w-40",
    },
    {
      key: "assignee",
      label: "Assignee",
      value: (t) => (t.entity.Assignee ?? []).join(", "),
      width: "w-28",
    },
    {
      key: "labels",
      label: "Labels",
      value: (t) => (t.entity.Labels ?? []).join(", "),
      width: "w-32",
    },
    {
      key: "updated",
      label: "Updated",
      // A task with no updated_date has a zero time, which reads as
      // 0001-01-01 and looks like a date somebody meant.
      value: (t) => shortDate(String(t.entity.Updated ?? "")),
      mono: true,
      width: "w-24",
    },
  ];

  const STORAGE_KEY = "muster.list.columns";
  const defaultVisible = [
    "project",
    "id",
    "title",
    "status",
    "priority",
    "milestone",
    "updated",
  ];

  function loadVisible(): string[] {
    try {
      const raw = localStorage.getItem(STORAGE_KEY);
      if (raw) return JSON.parse(raw) as string[];
    } catch {
      // A browser that refuses storage is not a reason to show no columns.
    }
    return defaultVisible;
  }

  let visible = $state<string[]>(loadVisible());
  let sortKey = $state("project");
  let ascending = $state(true);
  let picking = $state(false);

  const columns = $derived(allColumns.filter((c) => visible.includes(c.key)));

  function toggleColumn(key: string): void {
    visible = visible.includes(key)
      ? visible.filter((k) => k !== key)
      : [
          ...allColumns
            .map((c) => c.key)
            .filter((k) => visible.includes(k) || k === key),
        ];
    try {
      localStorage.setItem(STORAGE_KEY, JSON.stringify(visible));
    } catch {
      // Not being able to remember the choice is not a reason to refuse it.
    }
  }

  function sortBy(key: string): void {
    if (sortKey === key) ascending = !ascending;
    else {
      sortKey = key;
      ascending = true;
    }
  }

  type Task = (typeof $visibleTasks)[number];
  type Row = { task: Task; depth: number };

  /**
   * Subtasks sit under their parent rather than sorting away from it.
   *
   * The sort still decides the order of everything at the top level; a
   * subtask simply follows the task it belongs to, and its siblings keep the
   * same sort among themselves. A subtask whose parent is not in the list —
   * filtered out, or archived — keeps its own place at the top level rather
   * than disappearing with it.
   */
  const rows = $derived.by(() => {
    const column = allColumns.find((c) => c.key === sortKey) ?? allColumns[0];
    const sorted = [...$visibleTasks].sort((a, b) =>
      column
        .value(a)
        .localeCompare(column.value(b), undefined, { numeric: true }),
    );
    const ordered = ascending ? sorted : sorted.reverse();

    const present = new Set(ordered.map(refKey));
    const parentOf = (t: Task): string | null => {
      const parent = t.family?.parent;
      if (!parent) return null;
      // The backend resolved the parent to where it actually lives, so this
      // is an exact ref rather than an id to guess a class for.
      const key = refKey(parent);
      return present.has(key) ? key : null;
    };

    /* eslint-disable svelte/prefer-svelte-reactivity --
       These are scratch values inside one computation, thrown away when it
       ends. The reactive collections exist for state that outlives a render
       and has to notify on mutation; neither is true here, and using them
       would only mean tracking writes nothing ever reads. */
    const children = new Map<string, Task[]>();
    for (const task of ordered) {
      const key = parentOf(task);
      if (!key) continue;
      const list = children.get(key);
      if (list) list.push(task);
      else children.set(key, [task]);
    }

    const out: Row[] = [];
    const seen = new Set<string>();
    const emit = (task: Task, depth: number): void => {
      const key = refKey(task);
      // A task cannot be its own ancestor, but a hand-edited file could say
      // so, and an interface that hangs on one is worse than one that does
      // not draw it twice.
      if (seen.has(key)) return;
      seen.add(key);
      out.push({ task, depth });
      for (const child of children.get(key) ?? []) emit(child, depth + 1);
    };
    for (const task of ordered) {
      if (!parentOf(task)) emit(task, 0);
    }
    // Anything left is part of a cycle; it still belongs on screen.
    for (const task of ordered) emit(task, 0);
    /* eslint-enable svelte/prefer-svelte-reactivity */
    return out;
  });

  /**
   * Ticking tasks to change together.
   *
   * The box is the only thing that selects; a click on the row still opens the
   * task, because the list is read far more often than it is edited.
   *
   * Shift extends from the last box clicked rather than from the top of the
   * table, which is what makes choosing forty tasks one action instead of
   * forty. It only ever adds: a range that unticked whatever it crossed would
   * make a mis-aimed shift-click destroy a selection that took work to build.
   */
  let anchor = $state(-1);

  function pick(index: number, event: MouseEvent | KeyboardEvent): void {
    if (event.shiftKey && anchor >= 0 && anchor < rows.length) {
      const [from, to] = anchor < index ? [anchor, index] : [index, anchor];
      /* eslint-disable-next-line svelte/prefer-svelte-reactivity --
         A scratch value inside one call, thrown away when it returns. The
         reactive collections are for state that outlives a render; this
         never reaches one. */
      const keys = new Set($chosen);
      for (const row of rows.slice(from, to + 1)) keys.add(refKey(row.task));
      chosen.set([...keys]);
    } else {
      toggleChosen(rows[index].task);
    }
    anchor = index;
  }

  const allChosen = $derived(
    rows.length > 0 && $chosenTasks.length === rows.length,
  );
</script>

<div class="flex min-h-0 min-w-0 flex-1 flex-col">
  <div
    class="flex shrink-0 items-baseline gap-3 border-b border-rule px-3 py-1.5"
  >
    <span class="font-mono text-data text-chalk-faint tabular-nums">
      {rows.length}
      {rows.length === 1 ? "task" : "tasks"}
    </span>
    <button
      type="button"
      class="ml-auto font-mono text-data text-chalk-faint hover:text-chalk"
      aria-expanded={picking}
      aria-label="Choose which columns to show"
      onclick={() => (picking = !picking)}
    >
      columns
    </button>
  </div>

  {#if picking}
    <div class="flex shrink-0 flex-wrap gap-1.5 border-b border-rule px-3 py-2">
      {#each allColumns as column (column.key)}
        <button
          type="button"
          class="rounded-[2px] border px-1.5 text-micro
                 {visible.includes(column.key)
            ? 'border-chalk bg-ink text-chalk'
            : 'border-rule text-chalk-dim hover:text-chalk'}"
          aria-pressed={visible.includes(column.key)}
          onclick={() => toggleColumn(column.key)}
        >
          {column.label}
        </button>
      {/each}
    </div>
  {/if}

  {#if $chosenTasks.length > 0}
    <BulkBar />
  {/if}

  <div class="min-h-0 flex-1 overflow-auto">
    <table class="w-full border-collapse text-body">
      <thead class="sticky top-0 bg-ink-raised">
        <tr>
          <th class="w-8 border-b border-rule px-2 py-1">
            <input
              type="checkbox"
              class="align-middle"
              checked={allChosen}
              indeterminate={$chosenTasks.length > 0 && !allChosen}
              aria-label="Choose every task shown"
              onchange={() =>
                allChosen ? clearChosen() : chooseAll(rows.map((r) => r.task))}
            />
          </th>
          {#each columns as column (column.key)}
            <th
              class="border-b border-rule px-2 py-1 text-left text-micro font-medium
                     tracking-[0.14em] text-chalk-faint uppercase {column.width ??
                ''}"
            >
              <button
                type="button"
                class="hover:text-chalk"
                aria-label="Sort by {column.label}"
                onclick={() => sortBy(column.key)}
              >
                {column.label}{sortKey === column.key
                  ? ascending
                    ? " ↑"
                    : " ↓"
                  : ""}
              </button>
            </th>
          {/each}
        </tr>
      </thead>
      <tbody>
        {#each rows as { task, depth }, index (task.project + task.kind + task.class + task.id)}
          <tr
            class="cursor-pointer border-b border-rule hover:bg-ink-raised
                   {$chosen.includes(refKey(task)) ? 'bg-ink-raised' : ''}"
            onclick={() =>
              openTask({
                project: task.project,
                kind: task.kind,
                class: task.class,
                id: task.id,
              })}
          >
            <td class="w-8 px-2 py-1">
              <input
                type="checkbox"
                class="align-middle"
                checked={$chosen.includes(refKey(task))}
                aria-label="Choose {task.id} in {task.projectName}"
                onclick={(e) => {
                  e.stopPropagation();
                  pick(index, e);
                }}
              />
            </td>
            {#each columns as column, i (column.key)}
              <td
                class="max-w-0 truncate px-2 py-1 {column.mono
                  ? 'font-mono text-data'
                  : ''} {column.key === 'title'
                  ? 'text-chalk'
                  : 'text-chalk-dim'}"
                style={i === 0 && depth > 0
                  ? `padding-left: ${0.5 + depth * 1.25}rem`
                  : ""}
              >
                {#if column.key === "project"}
                  <span class="flex items-baseline gap-1.5">
                    {#if i === 0 && depth > 0}
                      <span class="text-chalk-faint" aria-hidden="true">↳</span>
                    {/if}
                    <span
                      class="h-2 w-2 shrink-0 self-center rounded-[1px]"
                      style="background-color: {projectColour(
                        task.project,
                        task.projectColour,
                      )}"
                    ></span>
                    {column.value(task)}
                  </span>
                {:else if i === 0 && depth > 0}
                  <span class="text-chalk-faint" aria-hidden="true">↳</span>
                  {column.value(task)}
                {:else}
                  {column.value(task)}
                {/if}
              </td>
            {/each}
          </tr>
        {/each}
      </tbody>
    </table>
  </div>
</div>
