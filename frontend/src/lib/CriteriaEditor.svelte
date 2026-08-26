<script lang="ts">
  import { BoardService } from "../../bindings/github.com/FMakareev/muster-backlog/internal/app";
  import type { Criterion } from "../../bindings/github.com/FMakareev/muster-backlog/internal/backlog/models";
  import { refresh } from "./board";
  import { notify } from "./notices";

  interface Props {
    project: string;
    taskID: string;
    criteria: Criterion[] | null;
  }
  let { project, taskID, criteria }: Props = $props();

  /**
   * Acceptance criteria.
   *
   * Ticking goes through the CLI's own index, which is what it renumbers on
   * every insertion. Editing the list replaces it wholesale, which makes
   * adding, removing and reordering one operation and keeps those indexes out
   * of the interface entirely.
   */

  let editing = $state(false);
  let draft = $state("");
  let busy = $state(false);

  const items = $derived(criteria ?? []);

  function start(): void {
    draft = items.map((c) => c.Text).join("\n");
    editing = true;
  }

  async function tick(index: number, checked: boolean): Promise<void> {
    if (busy) return;
    busy = true;
    const result = await BoardService.CheckCriterion(
      project,
      taskID,
      index,
      checked,
    );
    busy = false;
    if (!result.ok && result.problem) {
      notify(`${result.problem.title}: ${result.problem.detail}`);
    }
    await refresh();
  }

  async function save(): Promise<void> {
    if (busy) return;
    busy = true;
    const lines = draft
      .split("\n")
      .map((line) => line.trim())
      .filter(Boolean);
    const result = await BoardService.SetAcceptanceCriteria(
      project,
      taskID,
      lines,
    );
    busy = false;
    if (!result.ok) {
      notify(
        result.problem
          ? `${result.problem.title}: ${result.problem.detail}`
          : "The criteria could not be saved.",
      );
      return;
    }
    editing = false;
    await refresh();
  }

  function onKeydown(event: KeyboardEvent): void {
    event.stopPropagation();
    if (event.key === "Escape") {
      editing = false;
    } else if (event.key === "Enter" && (event.metaKey || event.ctrlKey)) {
      event.preventDefault();
      void save();
    }
  }
</script>

<section>
  <div class="flex items-baseline gap-2">
    <h3
      class="text-micro font-medium tracking-[0.14em] text-chalk-faint uppercase"
    >
      Acceptance criteria
    </h3>
    {#if editing}
      <button
        type="button"
        class="ml-auto font-mono text-micro text-chalk-faint hover:text-chalk"
        onclick={() => (editing = false)}
        disabled={busy}
      >
        cancel esc
      </button>
      <button
        type="button"
        class="font-mono text-micro text-chalk hover:underline"
        onclick={save}
        disabled={busy}
      >
        {busy ? "saving…" : "save ⌃↵"}
      </button>
    {:else}
      <button
        type="button"
        class="ml-auto font-mono text-micro text-chalk-faint hover:text-chalk"
        onclick={start}
      >
        edit
      </button>
    {/if}
  </div>

  {#if editing}
    <p class="mt-1 text-body text-chalk-faint">
      One criterion per line. Order here is the order on disk.
    </p>
    <textarea
      bind:value={draft}
      onkeydown={onKeydown}
      disabled={busy}
      class="mt-1 min-h-32 w-full font-mono"
      aria-label="Acceptance criteria, one per line"></textarea>
  {:else if items.length === 0}
    <p class="mt-1 text-body text-chalk-faint">None recorded.</p>
  {:else}
    <ul class="mt-1 flex flex-col gap-2">
      {#each items as criterion (criterion.Index)}
        <li class="flex items-baseline gap-2">
          <input
            type="checkbox"
            checked={criterion.Checked}
            disabled={busy}
            class="self-center accent-chalk-dim"
            aria-label={criterion.Text}
            onchange={(e) => tick(criterion.Index, e.currentTarget.checked)}
          />
          <span class="font-mono text-micro text-chalk-faint"
            >#{criterion.Index}</span
          >
          <span
            class="text-body {criterion.Checked
              ? 'text-chalk-faint line-through'
              : 'text-chalk-dim'}"
          >
            {criterion.Text}
          </span>
        </li>
      {/each}
    </ul>
  {/if}
</section>
