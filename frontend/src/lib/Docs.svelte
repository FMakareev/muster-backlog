<script lang="ts">
  import { BoardService } from "../../bindings/github.com/FMakareev/muster-backlog/internal/app";
  import type { TaskView } from "../../bindings/github.com/FMakareev/muster-backlog/internal/app/models";
  import { projectColour } from "./colour";
  import Markdown from "./Markdown.svelte";
  import { findInProject, openTask, projectChanged } from "./ui";

  /**
   * Documents and decisions.
   *
   * A Backlog.md project carries these next to its tasks and they are read far
   * more often than they are written. Rendering them here is what makes the
   * application the single place a project is read from rather than only
   * tracked in.
   */

  let docs = $state<TaskView[]>([]);
  let decisions = $state<TaskView[]>([]);
  let selected = $state<TaskView | null>(null);
  let loading = $state(true);

  async function load(): Promise<void> {
    const [d, k] = await Promise.all([
      BoardService.Entities("document"),
      BoardService.Entities("decision"),
    ]);
    docs = d ?? [];
    decisions = k ?? [];
    loading = false;

    // Keep whatever is open showing the current file rather than a stale copy.
    if (selected) {
      const all = [...docs, ...decisions];
      selected =
        all.find(
          (e) =>
            e.project === selected!.project &&
            e.kind === selected!.kind &&
            e.id === selected!.id,
        ) ?? null;
    }
  }

  $effect(() => {
    // Re-read whenever any project changes, so the viewer is as live as the
    // board is.
    void $projectChanged;
    void load();
  });

  function onTaskRef(id: string, project: string): void {
    const found = findInProject(project, id);
    if (found) {
      openTask({
        project: found.project,
        kind: found.kind,
        class: found.class,
        id: found.id,
      });
    }
  }

  const groups = $derived([
    ["Documents", docs] as const,
    ["Decisions", decisions] as const,
  ]);
</script>

<div class="flex min-h-0 min-w-0 flex-1">
  <nav
    class="flex w-72 shrink-0 flex-col overflow-y-auto border-r border-rule"
    aria-label="Documents and decisions"
  >
    {#if loading}
      <p class="px-3 py-2 text-body text-chalk-faint">Reading…</p>
    {:else if docs.length + decisions.length === 0}
      <p class="px-3 py-2 text-body text-chalk-faint">
        No documents or decisions in any registered project.
      </p>
    {/if}

    {#each groups as [heading, entries] (heading)}
      {#if entries.length > 0}
        <h2
          class="px-3 pt-3 pb-1 text-micro font-medium tracking-[0.14em] text-chalk-faint uppercase"
        >
          {heading}
        </h2>
        <ul>
          {#each entries as entry (entry.project + entry.id)}
            {@const open =
              selected?.project === entry.project &&
              selected?.kind === entry.kind &&
              selected?.id === entry.id}
            <li>
              <button
                type="button"
                class="flex w-full items-baseline gap-2 px-3 py-1 text-left
                       {open ? 'bg-ink-raised' : 'hover:bg-ink-raised'}"
                onclick={() => (selected = entry)}
              >
                <span
                  class="h-2 w-2 shrink-0 self-center rounded-[1px]"
                  style="background-color: {projectColour(
                    entry.project,
                    entry.projectColour,
                  )}"
                  title={entry.projectName}
                ></span>
                <span
                  class="min-w-0 flex-1 truncate text-body
                         {open ? 'text-chalk' : 'text-chalk-dim'}"
                >
                  {entry.entity.Title}
                </span>
              </button>
            </li>
          {/each}
        </ul>
      {/if}
    {/each}
  </nav>

  <main class="min-h-0 min-w-0 flex-1 overflow-y-auto p-6">
    {#if selected}
      <div class="mx-auto max-w-3xl">
        <p class="font-mono text-data text-chalk-faint">
          {selected.projectName} · {selected.id}
          {#if selected.entity.Status}· {selected.entity.Status}{/if}
        </p>
        <h1 class="mt-1 text-title font-semibold text-chalk">
          {selected.entity.Title}
        </h1>
        <div class="mt-3">
          <Markdown
            source={selected.entity.Body}
            project={selected.project}
            {onTaskRef}
          />
        </div>
        <p class="mt-6 font-mono text-micro text-chalk-faint">
          {selected.entity.Path}
        </p>
      </div>
    {:else if !loading}
      <p class="text-body text-chalk-faint">Choose something on the left.</p>
    {/if}
  </main>
</div>
