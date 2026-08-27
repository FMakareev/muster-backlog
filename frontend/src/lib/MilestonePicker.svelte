<script lang="ts">
  import { BoardService } from "../../bindings/github.com/FMakareev/muster-backlog/internal/app";
  import { milestones, refresh } from "./board";
  import { notify } from "./notices";

  /**
   * Choosing a milestone, and making one when the right one does not exist.
   *
   * Discovering you need a milestone happens while you are assigning a task,
   * not while you are on the projects screen — so the picker makes one rather
   * than sending you somewhere else and back.
   *
   * One component for every place a milestone is chosen: the create form, the
   * task panel and the inbox. Three copies of a select would drift, and this
   * one has behaviour in it now.
   */

  interface Props {
    project: string;
    value: string;
    /** Called with the milestone id, or "" for none. */
    onChange: (milestone: string) => void;
    disabled?: boolean;
    label?: string;
    /**
     * Inline sits in a row of controls; stacked sits in a form. The shape
     * follows where it is put rather than the other way round.
     */
    inline?: boolean;
  }
  let {
    project,
    value,
    onChange,
    disabled = false,
    label = "Milestone",
    inline = false,
  }: Props = $props();

  // A sentinel that cannot be a milestone id, which are m-<n>.
  const NEW = "__new__";

  const options = $derived($milestones.filter((m) => m.project === project));

  let creating = $state(false);
  let name = $state("");
  let description = $state("");
  let busy = $state(false);

  function onSelect(next: string): void {
    if (next === NEW) {
      creating = true;
      return;
    }
    onChange(next);
  }

  async function create(): Promise<void> {
    if (!name.trim() || busy) return;
    busy = true;
    const result = await BoardService.AddMilestone(
      project,
      name.trim(),
      description,
    );
    busy = false;
    if (!result.ok) {
      notify(
        result.problem
          ? `${result.problem.title}: ${result.problem.detail}`
          : "The milestone could not be created.",
      );
      return;
    }
    // Everything that lists milestones reads the same store, so one refresh
    // puts it on the board, the cards and the other pickers at once.
    await refresh();
    creating = false;
    name = "";
    description = "";
    onChange(result.taskId);
  }

  const heading =
    "text-micro font-medium uppercase tracking-[0.14em] text-chalk-faint";
  const inlineHeading =
    "text-micro tracking-[0.14em] text-chalk-faint uppercase";
</script>

<div class={inline ? "flex items-center gap-1" : "flex flex-col gap-1"}>
  <span class={inline ? inlineHeading : heading}>{label}</span>

  {#if creating}
    <div class={inline ? "flex items-center gap-1" : "flex flex-col gap-1"}>
      <!-- svelte-ignore a11y_autofocus -->
      <input
        class={inline ? "w-48" : "w-full"}
        autofocus
        bind:value={name}
        aria-label="Name of the new milestone"
        placeholder="What it is called"
        onkeydown={(e) => {
          if (e.key === "Enter") {
            e.preventDefault();
            void create();
          }
          if (e.key === "Escape") creating = false;
        }}
      />
      <input
        class={inline ? "w-48" : "w-full"}
        bind:value={description}
        aria-label="Description of the new milestone"
        placeholder="optional description"
      />
      <div class="flex items-baseline gap-3">
        <button
          type="button"
          class="min-h-6 font-mono text-data text-chalk hover:underline"
          disabled={busy || !name.trim()}
          onclick={create}
        >
          {busy ? "creating…" : "create it"}
        </button>
        <button
          type="button"
          class="min-h-6 font-mono text-data text-chalk-faint hover:text-chalk"
          onclick={() => (creating = false)}
        >
          cancel
        </button>
      </div>
    </div>
  {:else}
    <select
      class={inline ? "" : "w-full"}
      aria-label={label}
      {disabled}
      {value}
      onchange={(e) => onSelect(e.currentTarget.value)}
    >
      <option value="">none</option>
      {#each options as milestone (milestone.id)}
        <option value={milestone.id}>{milestone.title}</option>
      {/each}
      <option value={NEW}>+ a new milestone…</option>
    </select>
  {/if}
</div>
