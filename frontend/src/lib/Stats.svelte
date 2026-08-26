<script lang="ts">
  import { BoardService } from "../../bindings/github.com/FMakareev/muster-backlog/internal/app";
  import type {
    AnalyticsView,
    WIPStatus,
  } from "../../bindings/github.com/FMakareev/muster-backlog/internal/app/models";
  import { projectColour } from "./colour";
  import { openTask } from "./ui";

  /**
   * The cross-project overview.
   *
   * The vocabulary is the one `backlog overview` already established from
   * native data — status and priority breakdowns, average age, stale work,
   * blocked work — because a second way to describe the same backlog would
   * only be a second thing to reconcile. What is new is that it spans every
   * project at once, and that every figure leads back to the tasks behind it.
   */

  let reports = $state<AnalyticsView[]>([]);
  let wip = $state<WIPStatus[]>([]);
  let loading = $state(true);

  $effect(() => {
    void (async () => {
      const [a, w] = await Promise.all([
        BoardService.Analytics(),
        BoardService.WIP(),
      ]);
      reports = a ?? [];
      wip = w ?? [];
      loading = false;
    })();
  });

  // The last entry is the total across every project.
  const total = $derived(reports.at(-1));
  const perProject = $derived(reports.slice(0, -1));

  function open(task: {
    project: string;
    kind: string;
    class: string;
    id: string;
  }): void {
    openTask({
      project: task.project,
      kind: task.kind,
      class: task.class,
      id: task.id,
    });
  }

  const heading =
    "text-micro font-medium uppercase tracking-[0.14em] text-chalk-faint";
</script>

<div class="min-h-0 min-w-0 flex-1 overflow-y-auto p-4">
  {#if loading}
    <p class="text-body text-chalk-faint">Counting…</p>
  {:else if !total || total.tasks === 0}
    <p class="text-body text-chalk-faint">Nothing to count yet.</p>
  {:else}
    <div class="flex flex-col gap-6">
      <section class="flex flex-wrap gap-6">
        <div>
          <p class={heading}>Open across everything</p>
          <p class="font-mono text-title text-chalk tabular-nums">
            {total.tasks}
          </p>
        </div>
        <div>
          <p class={heading}>Without a priority</p>
          <p class="font-mono text-title text-chalk tabular-nums">
            {total.unprioritised}
            <span class="text-data text-chalk-faint">
              {total.tasks
                ? `· ${Math.round((total.unprioritised / total.tasks) * 100)}%`
                : ""}
            </span>
          </p>
        </div>
        <div>
          <p class={heading}>Average age of open work</p>
          <p class="font-mono text-title text-chalk tabular-nums">
            {Math.round(total.averageAgeDays)}<span
              class="text-data text-chalk-faint"
            >
              days</span
            >
          </p>
        </div>
        <div>
          <p class={heading}>Waiting on something</p>
          <p class="font-mono text-title text-chalk tabular-nums">
            {(total.blocked ?? []).length}
          </p>
        </div>
        <div>
          <p class={heading}>Untouched too long</p>
          <p class="font-mono text-title text-chalk tabular-nums">
            {(total.stale ?? []).length}
          </p>
        </div>
      </section>

      <section>
        <h2 class={heading}>By project</h2>
        <table class="mt-1 w-full border-collapse text-body">
          <thead>
            <tr class="text-micro text-chalk-faint uppercase">
              <th class="border-b border-rule px-2 py-1 text-left">Project</th>
              <th class="border-b border-rule px-2 py-1 text-right">Open</th>
              <th class="border-b border-rule px-2 py-1 text-right"
                >No priority</th
              >
              <th class="border-b border-rule px-2 py-1 text-right">Blocked</th>
              <th class="border-b border-rule px-2 py-1 text-right">Stale</th>
              <th class="border-b border-rule px-2 py-1 text-right">Avg age</th>
              <th class="border-b border-rule px-2 py-1 text-left">Statuses</th>
            </tr>
          </thead>
          <tbody class="font-mono text-data tabular-nums">
            {#each perProject as report (report.project)}
              <tr class="border-b border-rule">
                <td class="px-2 py-1">
                  <span class="flex items-baseline gap-1.5 font-sans text-body">
                    <span
                      class="h-2 w-2 shrink-0 self-center rounded-[1px]"
                      style="background-color: {projectColour(
                        report.project,
                        '',
                      )}"
                    ></span>
                    {report.projectName}
                  </span>
                </td>
                <td class="px-2 py-1 text-right text-chalk">{report.tasks}</td>
                <td class="px-2 py-1 text-right text-chalk-dim"
                  >{report.unprioritised}</td
                >
                <td class="px-2 py-1 text-right text-chalk-dim">
                  {(report.blocked ?? []).length}
                </td>
                <td class="px-2 py-1 text-right text-chalk-dim">
                  {(report.stale ?? []).length}
                </td>
                <td class="px-2 py-1 text-right text-chalk-dim">
                  {Math.round(report.averageAgeDays)}d
                </td>
                <td class="px-2 py-1 text-chalk-faint">
                  {(report.statuses ?? [])
                    .map((c) => `${c.label} ${c.total}`)
                    .join(" · ")}
                </td>
              </tr>
            {/each}
          </tbody>
        </table>
      </section>

      {#if wip.length > 0}
        <section>
          <h2 class={heading}>Work in progress against your limits</h2>
          <ul class="mt-1 flex flex-col gap-0.5">
            {#each wip as w (w.project + w.status)}
              <li class="flex items-baseline gap-2 font-mono text-data">
                <span class="w-40 shrink-0 truncate text-chalk-dim"
                  >{w.projectName}</span
                >
                <span class="w-28 shrink-0 text-chalk-faint">{w.status}</span>
                <span
                  class="tabular-nums {w.over
                    ? 'text-chalk'
                    : 'text-chalk-dim'}"
                >
                  {w.count} / {w.limit}
                </span>
                {#if w.over}
                  <span class="text-chalk">at or over</span>
                {/if}
              </li>
            {/each}
          </ul>
        </section>
      {/if}

      {#if (total.blocked ?? []).length > 0}
        <section>
          <h2 class={heading}>Waiting on unfinished work</h2>
          <ul class="mt-1 flex flex-col">
            {#each total.blocked ?? [] as b (b.task.project + b.task.id)}
              <li>
                <button
                  type="button"
                  class="flex w-full items-baseline gap-2 border-b border-rule px-1 py-1
                         text-left hover:bg-ink-raised"
                  onclick={() => open(b.task)}
                >
                  <span
                    class="w-40 shrink-0 truncate text-data text-chalk-faint"
                  >
                    {b.task.projectName}
                  </span>
                  <span
                    class="w-20 shrink-0 font-mono text-data text-chalk-faint"
                  >
                    {b.task.id}
                  </span>
                  <span class="min-w-0 flex-1 truncate text-chalk-dim">
                    {b.task.entity.Title}
                  </span>
                  <span class="shrink-0 font-mono text-data text-chalk-faint">
                    waiting on {(b.on ?? []).join(", ")}
                  </span>
                </button>
              </li>
            {/each}
          </ul>
        </section>
      {/if}

      {#if (total.stale ?? []).length > 0}
        <section>
          <h2 class={heading}>Untouched too long</h2>
          <ul class="mt-1 flex flex-col">
            {#each total.stale ?? [] as task (task.project + task.id)}
              <li>
                <button
                  type="button"
                  class="flex w-full items-baseline gap-2 border-b border-rule px-1 py-1
                         text-left hover:bg-ink-raised"
                  onclick={() => open(task)}
                >
                  <span
                    class="w-40 shrink-0 truncate text-data text-chalk-faint"
                  >
                    {task.projectName}
                  </span>
                  <span
                    class="w-20 shrink-0 font-mono text-data text-chalk-faint"
                  >
                    {task.id}
                  </span>
                  <span class="min-w-0 flex-1 truncate text-chalk-dim">
                    {task.entity.Title}
                  </span>
                  <span class="shrink-0 font-mono text-data text-chalk-faint">
                    {String(task.entity.Updated ?? "").slice(0, 10)}
                  </span>
                </button>
              </li>
            {/each}
          </ul>
        </section>
      {/if}
    </div>
  {/if}
</div>
