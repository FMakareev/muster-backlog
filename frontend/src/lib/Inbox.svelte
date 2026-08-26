<script lang="ts">
  import { BoardService } from "../../bindings/github.com/FMakareev/muster-backlog/internal/app";
  import type {
    DraftView,
    WriteResult,
  } from "../../bindings/github.com/FMakareev/muster-backlog/internal/app/models";
  import { drafts, projects, refresh } from "./board";
  import { projectColour } from "./colour";
  import Markdown from "./Markdown.svelte";
  import { notify } from "./notices";
  import { focusedProject } from "./ui";

  /**
   * The inbox.
   *
   * Backlog.md keeps drafts off the board, which is what makes capture cheap
   * and also what makes an unread drafts folder the new pile nobody looks at.
   * This is where the pile is emptied: promote a note, rewrite it, send it to
   * another project, or discard it.
   *
   * Oldest first, and every note says how long it has been waiting, because a
   * growing inbox is only a problem you can act on if you can see it grow.
   */

  let editing = $state("");
  let confirming = $state("");
  let busy = $state("");

  let title = $state("");
  let description = $state("");
  let labels = $state("");
  let assignee = $state("");
  let target = $state("");

  const healthy = $derived($projects.filter((p) => p.ok && !p.hidden));
  const shown = $derived(
    $focusedProject
      ? $drafts.filter((d) => d.project === $focusedProject)
      : $drafts,
  );

  const key = (draft: DraftView) => draft.project + " " + draft.id;

  /** How long a note has waited, in words rather than a bare number. */
  function waited(days: number): string {
    if (days < 0) return "no date";
    if (days === 0) return "today";
    if (days === 1) return "1 day";
    return `${days} days`;
  }

  function startEditing(draft: DraftView): void {
    editing = key(draft);
    confirming = "";
    title = draft.entity.Title;
    description = draft.entity.Sections?.description ?? "";
    labels = (draft.entity.Labels ?? []).join(", ");
    assignee = (draft.entity.Assignee ?? []).join(", ");
    target = draft.project;
  }

  async function run(
    draft: DraftView,
    failure: string,
    action: () => Promise<WriteResult>,
  ): Promise<void> {
    busy = key(draft);
    const result = await action();
    busy = "";
    if (!result.ok) {
      notify(
        result.problem
          ? `${result.problem.title}: ${result.problem.detail}`
          : failure,
      );
      return;
    }
    await refresh();
  }

  const promote = (draft: DraftView) =>
    run(draft, "The draft could not be promoted", () =>
      BoardService.PromoteDraft(draft.project, draft.id),
    );

  const discard = (draft: DraftView) => {
    confirming = "";
    return run(draft, "The draft could not be discarded", () =>
      BoardService.DiscardDraft(draft.project, draft.id),
    );
  };

  async function saveEdit(draft: DraftView): Promise<void> {
    await run(draft, "The draft could not be rewritten", () =>
      BoardService.ReviseDraft(draft.project, draft.id, {
        title,
        description,
        labels: labels
          .split(",")
          .map((l) => l.trim())
          .filter(Boolean),
        assignee,
        project: target,
      }),
    );
    editing = "";
  }

  const label =
    "text-micro font-medium uppercase tracking-[0.14em] text-chalk-faint";
</script>

<div class="flex min-h-0 min-w-0 flex-1 flex-col overflow-y-auto">
  <div
    class="flex shrink-0 items-baseline gap-3 border-b border-rule px-3 py-1.5"
  >
    <span class="font-mono text-data tabular-nums text-chalk-faint">
      {shown.length}
      {shown.length === 1 ? "draft" : "drafts"} waiting
    </span>
    {#if shown.length > 0 && shown[0].waitingDays > 0}
      <span class="font-mono text-data text-chalk-faint">
        · the oldest for {waited(shown[0].waitingDays)}
      </span>
    {/if}
  </div>

  {#if shown.length === 0}
    <p class="px-3 py-3 text-body text-chalk-faint">
      {$drafts.length === 0
        ? "Nothing waiting. Notes captured into any project land here."
        : "Nothing waiting in this project."}
    </p>
  {/if}

  <ul class="flex flex-col">
    {#each shown as item (key(item))}
      {@const colour = projectColour(item.project, item.projectColour)}
      {@const open = editing === key(item)}
      <li class="flex flex-col gap-1.5 border-b border-rule px-3 py-2">
        <div class="flex items-baseline gap-2">
          <span
            class="h-2 w-2 shrink-0 self-center rounded-[1px]"
            style="background-color: {colour}"
            title={item.projectName}
          ></span>
          <span class="shrink-0 font-mono text-data text-chalk-faint">
            {item.projectName}
          </span>
          <span class="shrink-0 font-mono text-micro text-chalk-faint">
            {item.id}
          </span>
          <span class="min-w-0 flex-1 truncate text-body text-chalk">
            {item.entity.Title}
          </span>
          <span
            class="shrink-0 font-mono text-data tabular-nums
                   {item.waitingDays >= 30 ? 'text-chalk' : 'text-chalk-faint'}"
            title="Waiting {item.waitingDays < 0
              ? 'since a date the file does not record'
              : `${item.waitingDays} days`}"
          >
            {waited(item.waitingDays)}
          </span>
        </div>

        {#if !open}
          {#if item.entity.Sections?.description?.trim()}
            <div class="max-w-3xl text-body text-chalk-dim">
              <Markdown
                source={item.entity.Sections.description}
                project={item.project}
              />
            </div>
          {/if}

          <div class="flex flex-wrap items-baseline gap-x-3 gap-y-1">
            {#each item.entity.Labels ?? [] as tag (tag)}
              <span
                class="rounded-[2px] bg-ink-sunken px-1 text-micro text-chalk-dim"
              >
                {tag}
              </span>
            {/each}
            <span class="ml-auto flex shrink-0 items-baseline gap-3">
              <button
                type="button"
                class="min-h-6 font-mono text-data text-chalk hover:underline"
                disabled={busy === key(item)}
                onclick={() => promote(item)}
              >
                promote
              </button>
              <button
                type="button"
                class="min-h-6 font-mono text-data text-chalk-faint hover:text-chalk"
                disabled={busy === key(item)}
                onclick={() => startEditing(item)}
              >
                edit or move
              </button>
              {#if confirming === key(item)}
                <button
                  type="button"
                  class="min-h-6 font-mono text-data text-chalk hover:underline"
                  onclick={() => discard(item)}
                >
                  discard, really
                </button>
                <button
                  type="button"
                  class="min-h-6 font-mono text-data text-chalk-faint hover:text-chalk"
                  onclick={() => (confirming = "")}
                >
                  keep
                </button>
              {:else}
                <button
                  type="button"
                  class="min-h-6 font-mono text-data text-chalk-faint hover:text-chalk"
                  disabled={busy === key(item)}
                  onclick={() => (confirming = key(item))}
                >
                  discard
                </button>
              {/if}
            </span>
          </div>
        {:else}
          <div class="flex max-w-3xl flex-col gap-2">
            <label class="flex flex-col gap-1">
              <span class={label}>Title</span>
              <input
                class="w-full"
                bind:value={title}
                aria-label="Draft title"
              />
            </label>
            <label class="flex flex-col gap-1">
              <span class={label}>Description</span>
              <textarea
                class="min-h-24 w-full font-mono"
                bind:value={description}
                aria-label="Draft description"></textarea>
            </label>
            <div class="flex flex-wrap gap-3">
              <label class="flex flex-col gap-1">
                <span class={label}>Labels</span>
                <input
                  class="w-56"
                  bind:value={labels}
                  aria-label="Draft labels"
                  placeholder="comma separated"
                />
              </label>
              <label class="flex flex-col gap-1">
                <span class={label}>Assignee</span>
                <input
                  class="w-40"
                  bind:value={assignee}
                  aria-label="Draft assignee"
                />
              </label>
              <label class="flex flex-col gap-1">
                <span class={label}>Project</span>
                <select bind:value={target} aria-label="Project for this draft">
                  {#each healthy as project (project.path)}
                    <option value={project.path}>{project.name}</option>
                  {/each}
                </select>
              </label>
            </div>

            <p class="text-body text-chalk-faint">
              Backlog.md has no way to edit a draft, so this captures a new note
              and discards this one. The rewritten note is captured now, so its
              wait starts again from today.
            </p>

            <div class="flex items-baseline gap-3">
              <button
                type="button"
                class="border-b border-chalk-faint text-body text-chalk hover:border-chalk"
                disabled={busy === key(item) || !title.trim()}
                onclick={() => saveEdit(item)}
              >
                {busy === key(item) ? "Saving" : "Save as a new note"}
              </button>
              <button
                type="button"
                class="font-mono text-data text-chalk-faint hover:text-chalk"
                onclick={() => (editing = "")}
              >
                cancel
              </button>
            </div>
          </div>
        {/if}
      </li>
    {/each}
  </ul>
</div>
