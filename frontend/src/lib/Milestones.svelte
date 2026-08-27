<script lang="ts">
  import { BoardService } from "../../bindings/github.com/FMakareev/muster-backlog/internal/app";
  import type { MilestoneView } from "../../bindings/github.com/FMakareev/muster-backlog/internal/app/models";
  import { refresh } from "./board";
  import { notify } from "./notices";

  /**
   * A project's milestones.
   *
   * Retiring one is where the care is needed: whichever command does it, the
   * file ends up in archive/milestones, and the real choice is what becomes of
   * the tasks that named it. Leaving them alone, clearing them and moving them
   * elsewhere are three different things to a plan, so the choice is made in
   * the open rather than behind a word like "remove".
   */

  interface Props {
    project: string;
    milestones: MilestoneView[];
  }
  let { project, milestones }: Props = $props();

  let adding = $state(false);
  let name = $state("");
  let description = $state("");

  let renaming = $state("");
  let renameTo = $state("");

  let retiring = $state("");
  let handling = $state("keep");
  let reassignTo = $state("");

  let busy = $state("");

  const others = $derived(milestones.filter((m) => m.id !== retiring));

  async function run(
    what: string,
    action: () => Promise<{
      ok: boolean;
      problem?: { title: string; detail: string } | null;
    }>,
  ): Promise<boolean> {
    busy = what;
    const result = await action();
    busy = "";
    if (!result.ok) {
      notify(
        result.problem
          ? `${result.problem.title}: ${result.problem.detail}`
          : "That could not be done.",
      );
      return false;
    }
    await refresh();
    return true;
  }

  async function add(): Promise<void> {
    if (!name.trim()) return;
    if (
      await run("add", () =>
        BoardService.AddMilestone(project, name.trim(), description),
      )
    ) {
      name = "";
      description = "";
      adding = false;
    }
  }

  async function rename(milestone: MilestoneView): Promise<void> {
    if (!renameTo.trim()) return;
    if (
      await run("rename", () =>
        BoardService.RenameMilestone(project, milestone.id, renameTo.trim()),
      )
    ) {
      renaming = "";
    }
  }

  async function retire(milestone: MilestoneView): Promise<void> {
    if (
      await run("retire", () =>
        BoardService.RetireMilestone(
          project,
          milestone.id,
          handling,
          reassignTo,
        ),
      )
    ) {
      retiring = "";
      handling = "keep";
      reassignTo = "";
    }
  }

  const quiet = "min-h-6 font-mono text-data text-chalk-faint hover:text-chalk";
</script>

<div class="flex flex-col gap-1.5 border-l border-rule pl-3">
  {#each milestones as milestone (milestone.id)}
    <div class="flex flex-wrap items-baseline gap-x-3 gap-y-1">
      <span class="font-mono text-data text-chalk-faint">{milestone.id}</span>

      {#if renaming === milestone.id}
        <input
          class="w-56"
          bind:value={renameTo}
          aria-label="New name for {milestone.id}"
          onkeydown={(e) => {
            if (e.key === "Enter") void rename(milestone);
            if (e.key === "Escape") renaming = "";
          }}
        />
        <button
          type="button"
          class="min-h-6 font-mono text-data text-chalk hover:underline"
          disabled={busy === "rename"}
          onclick={() => rename(milestone)}
        >
          save
        </button>
        <button type="button" class={quiet} onclick={() => (renaming = "")}>
          cancel
        </button>
      {:else}
        <span class="text-body text-chalk">{milestone.title}</span>
        <span class="font-mono text-data tabular-nums text-chalk-dim">
          {milestone.done}/{milestone.total}
        </span>
      {/if}

      {#if retiring !== milestone.id && renaming !== milestone.id}
        <span class="ml-auto flex shrink-0 items-baseline gap-3">
          <button
            type="button"
            class={quiet}
            onclick={() => {
              renaming = milestone.id;
              renameTo = milestone.title;
            }}
          >
            rename
          </button>
          <button
            type="button"
            class={quiet}
            onclick={() => {
              retiring = milestone.id;
              handling = "keep";
            }}
          >
            retire
          </button>
        </span>
      {/if}
    </div>

    {#if retiring === milestone.id}
      <div class="flex flex-col gap-1 pb-1">
        <!-- A sentence, not a label: it is telling you what is at stake, and
             the uppercase micro style is for field names. -->
        <span class="text-body text-chalk">
          {milestone.total === 1
            ? "One task names this milestone."
            : `${milestone.total} tasks name this milestone.`}
        </span>
        <div class="flex flex-wrap items-baseline gap-3">
          {#each [["keep", "leave them naming it"], ["clear", "clear their milestone"], ["reassign", "move them to"]] as [value, text] (value)}
            <label class="flex items-center gap-1.5 text-body text-chalk-dim">
              <input
                type="radio"
                name="handling-{milestone.id}"
                {value}
                checked={handling === value}
                onchange={() => (handling = value)}
              />
              {text}
            </label>
          {/each}
          {#if handling === "reassign"}
            <select
              bind:value={reassignTo}
              aria-label="Milestone to move them to"
            >
              <option value="">choose one</option>
              {#each others as other (other.id)}
                <option value={other.id}>{other.title}</option>
              {/each}
            </select>
          {/if}
        </div>
        <p class="text-body text-chalk-faint">
          The milestone file is archived either way, not deleted.
        </p>
        <div class="flex items-baseline gap-3">
          <button
            type="button"
            class="min-h-6 font-mono text-data text-chalk hover:underline"
            disabled={busy === "retire" ||
              (handling === "reassign" && !reassignTo)}
            onclick={() => retire(milestone)}
          >
            {busy === "retire" ? "retiring…" : "retire it"}
          </button>
          <button type="button" class={quiet} onclick={() => (retiring = "")}>
            keep it
          </button>
        </div>
      </div>
    {/if}
  {/each}

  {#if adding}
    <div class="flex flex-wrap items-baseline gap-2">
      <input
        class="w-56"
        bind:value={name}
        aria-label="Name of the new milestone"
        placeholder="What it is called"
        onkeydown={(e) => {
          if (e.key === "Enter") void add();
          if (e.key === "Escape") adding = false;
        }}
      />
      <input
        class="w-72"
        bind:value={description}
        aria-label="Description of the new milestone"
        placeholder="optional description"
      />
      <button
        type="button"
        class="min-h-6 font-mono text-data text-chalk hover:underline"
        disabled={busy === "add" || !name.trim()}
        onclick={add}
      >
        add it
      </button>
      <button type="button" class={quiet} onclick={() => (adding = false)}>
        cancel
      </button>
    </div>
  {:else}
    <button
      type="button"
      class="{quiet} self-start"
      onclick={() => (adding = true)}
    >
      add a milestone
    </button>
  {/if}
</div>
