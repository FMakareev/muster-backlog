<script lang="ts">
  import { BoardService } from "../../bindings/github.com/FMakareev/muster-backlog/internal/app";
  import type { TaskView } from "../../bindings/github.com/FMakareev/muster-backlog/internal/app/models";
  import { projects, refresh } from "./board";
  import { projectColour } from "./colour";
  import Markdown from "./Markdown.svelte";
  import NewDocument from "./NewDocument.svelte";
  import { notify } from "./notices";
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

  /**
   * Editing a document sends the whole body back, because `doc update
   * --content` takes nothing smaller. What is shown in the editor is exactly
   * what is in the file.
   */
  let editing = $state(false);
  let draftTitle = $state("");
  let draftType = $state("");
  let draftBody = $state("");
  let saving = $state(false);
  let creating = $state(false);

  const types = ["readme", "guide", "specification", "other"];

  function startEditing(): void {
    if (!current) return;
    editing = true;
    draftTitle = current.entity.Title;
    draftType = current.entity.Type ?? "";
    draftBody = current.entity.Body;
  }

  async function saveDocument(): Promise<void> {
    if (!current || saving) return;
    saving = true;
    const result = await BoardService.UpdateDocument(
      current.project,
      current.id,
      {
        title: draftTitle.trim(),
        type: draftType,
        content: draftBody,
      },
    );
    saving = false;
    if (!result.ok) {
      notify(
        result.problem
          ? `${result.problem.title}: ${result.problem.detail}`
          : "The document could not be saved.",
      );
      return;
    }
    editing = false;
    await load();
    await refresh();
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
    <div
      class="flex shrink-0 items-baseline gap-3 border-b border-rule px-3 py-1.5"
    >
      <button
        type="button"
        class="font-mono text-data text-chalk-faint hover:text-chalk"
        aria-expanded={creating}
        onclick={() => (creating = !creating)}
      >
        write something
      </button>
    </div>

    {#if creating}
      <NewDocument
        projects={$projects.filter((p) => p.ok && !p.hidden)}
        {types}
        onDone={async (ref) => {
          creating = false;
          await load();
          if (ref) selectedDoc.set(ref);
        }}
      />
    {/if}

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
        {#if editing}
          <div class="flex flex-col gap-2">
            <label class="flex flex-col gap-1">
              <span
                class="text-micro font-medium tracking-[0.14em] text-chalk-faint uppercase"
                >Title</span
              >
              <input
                class="w-full"
                bind:value={draftTitle}
                aria-label="Document title"
              />
            </label>
            <label class="flex flex-col gap-1">
              <span
                class="text-micro font-medium tracking-[0.14em] text-chalk-faint uppercase"
                >Type</span
              >
              <select bind:value={draftType} aria-label="Document type">
                <option value="">unchanged</option>
                {#each types as value (value)}
                  <option {value}>{value}</option>
                {/each}
              </select>
            </label>
            <label class="flex flex-col gap-1">
              <span
                class="text-micro font-medium tracking-[0.14em] text-chalk-faint uppercase"
                >Markdown</span
              >
              <textarea
                class="min-h-96 w-full font-mono"
                bind:value={draftBody}
                aria-label="Document markdown"></textarea>
            </label>
            <p class="text-body text-chalk-faint">
              Backlog.md replaces a document wholesale, so this is the entire
              file and it is sent as it stands.
            </p>
            <div class="flex items-baseline gap-3">
              <button
                type="button"
                class="border-b border-chalk-faint text-body text-chalk hover:border-chalk"
                disabled={saving}
                onclick={saveDocument}
              >
                {saving ? "saving…" : "save"}
              </button>
              <button
                type="button"
                class="font-mono text-data text-chalk-faint hover:text-chalk"
                onclick={() => (editing = false)}
              >
                cancel
              </button>
            </div>
          </div>
        {:else}
          <div class="flex items-baseline gap-3">
            <h1 class="mt-1 text-title font-semibold text-chalk">
              {current.entity.Title}
            </h1>
            {#if current.kind === "document"}
              <button
                type="button"
                class="ml-auto font-mono text-data text-chalk-faint hover:text-chalk"
                onclick={startEditing}
              >
                edit
              </button>
            {/if}
          </div>
          <div class="mt-3">
            <Markdown
              source={current.entity.Body}
              project={current.project}
              {onTaskRef}
            />
          </div>
          {#if current.kind === "decision"}
            <!-- Backlog.md creates a decision and has no command that fills
                 one in, so this is where the writing happens. -->
            <p class="mt-4 text-body text-chalk-faint">
              Backlog.md has no way to edit a decision, so this one is written
              in the file itself.
            </p>
          {/if}
          <p class="mt-6 font-mono text-micro text-chalk-faint">
            {current.entity.Path}
          </p>
        {/if}
      </div>
    {:else if !loading}
      <p class="text-body text-chalk-faint">Choose something on the left.</p>
    {/if}
  </main>
</div>
