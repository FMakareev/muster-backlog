<script lang="ts">
  import { BoardService } from "../../bindings/github.com/FMakareev/muster-backlog/internal/app";
  import type { TaskView } from "../../bindings/github.com/FMakareev/muster-backlog/internal/app/models";
  import { projectColour } from "./colour";
  import Markdown from "./Markdown.svelte";
  import {
    findInProject,
    focusedProject,
    openEntity,
    projectChanged,
    selectedDoc,
  } from "./ui";

  /**
   * Documents and decisions.
   *
   * A Backlog.md project carries these next to its tasks and they are read far
   * more often than they are written. They are listed by project, not in one
   * heap: with nine projects, an alphabetical pile of eighty-eight files tells
   * nobody whose they are.
   */

  let entries = $state<TaskView[]>([]);
  let loading = $state(true);

  async function load(): Promise<void> {
    const [docs, decisions] = await Promise.all([
      BoardService.Entities("document"),
      BoardService.Entities("decision"),
    ]);
    entries = [...(docs ?? []), ...(decisions ?? [])];
    loading = false;
  }

  $effect(() => {
    // As live as the board: re-read whenever any project changes.
    void $projectChanged;
    void load();
  });

  // Focusing a project in the roll narrows this too. It narrowing the board
  // but not the documents was a difference with no reason behind it.
  const visible = $derived(
    $focusedProject
      ? entries.filter((e) => e.project === $focusedProject)
      : entries,
  );

  /** Entries grouped by project, in the order the projects arrived. */
  const byProject = $derived.by(() => {
    // Plain arrays: this is derived once per render and never mutated
    // afterwards, so a reactive collection would buy nothing.
    const order: string[] = [];
    for (const entry of visible) {
      if (!order.includes(entry.project)) order.push(entry.project);
    }
    return order.map((project) => {
      const mine = visible.filter((e) => e.project === project);
      return {
        project,
        name: mine[0].projectName,
        colour: projectColour(project, mine[0].projectColour),
        documents: mine.filter((e) => e.kind === "document"),
        decisions: mine.filter((e) => e.kind === "decision"),
      };
    });
  });

  const current = $derived(
    $selectedDoc
      ? (entries.find(
          (e) =>
            e.project === $selectedDoc!.project &&
            e.kind === $selectedDoc!.kind &&
            e.id === $selectedDoc!.id,
        ) ?? null)
      : null,
  );

  function choose(entry: TaskView): void {
    selectedDoc.set({
      project: entry.project,
      kind: entry.kind,
      class: entry.class,
      id: entry.id,
    });
  }

  function isOpen(entry: TaskView): boolean {
    return (
      current?.project === entry.project &&
      current?.kind === entry.kind &&
      current?.id === entry.id
    );
  }

  function onTaskRef(id: string, project: string): void {
    const found = findInProject(project, id);
    if (found) {
      openEntity({
        project: found.project,
        kind: found.kind,
        class: found.class,
        id: found.id,
      });
    }
  }
</script>

<div class="flex min-h-0 min-w-0 flex-1">
  <nav
    class="flex w-72 shrink-0 flex-col overflow-y-auto border-r border-rule"
    aria-label="Documents and decisions"
  >
    {#if loading}
      <p class="px-3 py-2 text-body text-chalk-faint">Reading…</p>
    {:else if visible.length === 0}
      <p class="px-3 py-2 text-body text-chalk-faint">
        {$focusedProject
          ? "This project has no documents or decisions."
          : "No documents or decisions in any registered project."}
      </p>
    {/if}

    {#each byProject as group (group.project)}
      <h2
        class="sticky top-0 flex items-baseline gap-2 border-b border-rule bg-ink px-3
               py-1 text-micro font-medium tracking-[0.14em] text-chalk-dim uppercase"
      >
        <span
          class="h-2 w-2 shrink-0 self-center rounded-[1px]"
          style="background-color: {group.colour}"
        ></span>
        {group.name}
      </h2>

      {#each [["Documents", group.documents], ["Decisions", group.decisions]] as [heading, list] (heading)}
        {#if (list as TaskView[]).length > 0}
          <h3 class="px-3 pt-2 pb-0.5 text-micro text-chalk-faint">
            {heading}
          </h3>
          <ul>
            {#each list as TaskView[] as entry (entry.id)}
              <li>
                <button
                  type="button"
                  class="w-full truncate px-3 py-1 text-left text-body
                         {isOpen(entry)
                    ? 'bg-ink-raised text-chalk'
                    : 'text-chalk-dim hover:bg-ink-raised hover:text-chalk'}"
                  onclick={() => choose(entry)}
                >
                  {entry.entity.Title}
                </button>
              </li>
            {/each}
          </ul>
        {/if}
      {/each}
    {/each}
  </nav>

  <main class="min-h-0 min-w-0 flex-1 overflow-y-auto p-6">
    {#if current}
      <div class="mx-auto max-w-3xl">
        <p class="font-mono text-data text-chalk-faint">
          {current.projectName} · {current.id}
          {#if current.entity.Status}· {current.entity.Status}{/if}
        </p>
        <h1 class="mt-1 text-title font-semibold text-chalk">
          {current.entity.Title}
        </h1>
        <div class="mt-3">
          <Markdown
            source={current.entity.Body}
            project={current.project}
            {onTaskRef}
          />
        </div>
        <p class="mt-6 font-mono text-micro text-chalk-faint">
          {current.entity.Path}
        </p>
      </div>
    {:else if !loading}
      <p class="text-body text-chalk-faint">Choose something on the left.</p>
    {/if}
  </main>
</div>
