<script lang="ts">
  import { milestoneLabel, milestones } from "./board";
  import { projectColour } from "./colour";
  import { openTask, visibleTasks } from "./ui";

  /**
   * The task list.
   *
   * The mode for scanning many tasks at once, where the board is the mode for
   * moving a few. Sortable on every column, and the visible column set is
   * remembered so a working view survives a restart.
   */

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
      value: (t) => String(t.entity.Updated ?? "").slice(0, 10),
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

  const rows = $derived.by(() => {
    const column = allColumns.find((c) => c.key === sortKey) ?? allColumns[0];
    const sorted = [...$visibleTasks].sort((a, b) =>
      column
        .value(a)
        .localeCompare(column.value(b), undefined, { numeric: true }),
    );
    return ascending ? sorted : sorted.reverse();
  });
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

  <div class="min-h-0 flex-1 overflow-auto">
    <table class="w-full border-collapse text-body">
      <thead class="sticky top-0 bg-ink-raised">
        <tr>
          {#each columns as column (column.key)}
            <th
              class="border-b border-rule px-2 py-1 text-left text-micro font-medium
                     tracking-[0.14em] text-chalk-faint uppercase {column.width ??
                ''}"
            >
              <button
                type="button"
                class="hover:text-chalk"
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
        {#each rows as task (task.project + task.kind + task.class + task.id)}
          <tr
            class="cursor-pointer border-b border-rule hover:bg-ink-raised"
            onclick={() =>
              openTask({
                project: task.project,
                kind: task.kind,
                class: task.class,
                id: task.id,
              })}
          >
            {#each columns as column (column.key)}
              <td
                class="max-w-0 truncate px-2 py-1 {column.mono
                  ? 'font-mono text-data'
                  : ''} {column.key === 'title'
                  ? 'text-chalk'
                  : 'text-chalk-dim'}"
              >
                {#if column.key === "project"}
                  <span class="flex items-baseline gap-1.5">
                    <span
                      class="h-2 w-2 shrink-0 self-center rounded-[1px]"
                      style="background-color: {projectColour(
                        task.project,
                        task.projectColour,
                      )}"
                    ></span>
                    {column.value(task)}
                  </span>
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
