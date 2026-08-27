<script lang="ts">
  import { Events } from "@wailsio/runtime";
  import { BoardService } from "../../bindings/github.com/FMakareev/muster-backlog/internal/app";
  import type {
    BulkFailure,
    BulkProgress,
  } from "../../bindings/github.com/FMakareev/muster-backlog/internal/app/models";
  import { canMove, columns, milestones, projects, refresh } from "./board";
  import { chosenTasks, clearChosen } from "./ui";

  /**
   * Changing many tasks at once.
   *
   * Every write still goes through the CLI one task at a time, because that is
   * the only writer there is. What is bulk here is the choosing: the set is
   * picked once and the change described once.
   *
   * Two rules shape the form. Nothing is offered that half the selection would
   * reject, so status and priority are the intersection of what every chosen
   * project configures. And a milestone is not offered across projects at all:
   * ids belong to one project, and `task edit -m` writes an id it does not
   * recognise without complaining, so applying one across a selection would
   * plant a dangling reference everywhere but one.
   */

  // Saying nothing about a field has to be a value the select can hold, and it
  // has to be one that no status, priority or milestone id can be. The empty
  // string is already taken: clearing a milestone is a real instruction rather
  // than the absence of one. A leading NUL cannot appear in any of them.
  const KEEP = "\u0000keep";
  const CLEAR = "\u0000clear";

  let status = $state(KEEP);
  let priority = $state(KEEP);
  let milestone = $state(KEEP);
  let adding = $state("");
  let removing = $state<string[]>([]);

  let busy = $state(false);
  let done = $state(0);
  let outcome = $state<{ changed: number; failures: BulkFailure[] } | null>(
    null,
  );
  let refused = $state("");

  const chosenProjects = $derived([
    ...new Set($chosenTasks.map((t) => t.project)),
  ]);

  const named = $derived(
    chosenProjects.map(
      (path) => $projects.find((p) => p.path === path)?.name ?? path,
    ),
  );

  /** What every chosen project accepts, rather than what any one of them does. */
  const statuses = $derived(
    $columns
      .map((c) => c.name)
      .filter((name) => chosenProjects.every((p) => canMove(p, name))),
  );

  const priorities = $derived.by(() => {
    const lists = chosenProjects.map(
      (path) => $projects.find((p) => p.path === path)?.priorities ?? [],
    );
    if (lists.length === 0) return [];
    return lists[0].filter((value) => lists.every((l) => l.includes(value)));
  });

  const oneProject = $derived(chosenProjects.length === 1);
  const projectMilestones = $derived(
    oneProject
      ? $milestones.filter((m) => m.project === chosenProjects[0])
      : [],
  );

  /** Every label actually on the chosen tasks, which is what can be removed. */
  const present = $derived(
    [...new Set($chosenTasks.flatMap((t) => t.entity.Labels ?? []))].sort(),
  );

  const addLabels = $derived(
    adding
      .split(",")
      .map((l) => l.trim())
      .filter(Boolean),
  );

  const asked = $derived(
    status !== KEEP ||
      priority !== KEEP ||
      milestone !== KEEP ||
      addLabels.length > 0 ||
      removing.length > 0,
  );

  // A label can only be removed while it is still on something chosen.
  $effect(() => {
    const gone = removing.filter((l) => !present.includes(l));
    if (gone.length > 0) removing = removing.filter((l) => present.includes(l));
  });

  function toggleRemoval(label: string): void {
    removing = removing.includes(label)
      ? removing.filter((l) => l !== label)
      : [...removing, label];
  }

  function reset(): void {
    status = KEEP;
    priority = KEEP;
    milestone = KEEP;
    adding = "";
    removing = [];
  }

  async function apply(): Promise<void> {
    if (!asked || busy) return;
    const total = $chosenTasks.length;
    busy = true;
    done = 0;
    outcome = null;
    refused = "";

    // Nothing else reports how far a queue of separate writes has got, and at
    // roughly a third of a second each a large run is a long silence.
    const off = Events.On(
      "muster:bulk:progress",
      (event: { data: unknown }) => {
        const progress = (
          Array.isArray(event.data) ? event.data[0] : event.data
        ) as BulkProgress;
        if (progress) done = progress.done;
      },
    );

    const result = await BoardService.ChangeMany({
      tasks: $chosenTasks.map((t) => ({ project: t.project, taskId: t.id })),
      status: status === KEEP ? null : status,
      priority: priority === KEEP ? null : priority,
      milestone:
        milestone === KEEP ? null : milestone === CLEAR ? "" : milestone,
      addLabels: addLabels,
      removeLabels: removing,
    }).catch((err: unknown) => ({
      changed: 0,
      failures: null,
      problem: { title: "The change could not be run", detail: String(err) },
    }));

    off();
    busy = false;
    done = total;

    if (result.problem) {
      refused = `${result.problem.title}. ${result.problem.detail}`;
      return;
    }

    outcome = { changed: result.changed, failures: result.failures ?? [] };
    reset();
    await refresh();
  }

  const label =
    "text-micro font-medium tracking-[0.14em] text-chalk-faint uppercase";
  const quiet = "min-h-6 font-mono text-data text-chalk-faint hover:text-chalk";
</script>

<section
  class="flex shrink-0 flex-col gap-2 border-b border-rule bg-ink-sunken px-3 py-2"
  aria-label="Change the chosen tasks"
>
  <div class="flex flex-wrap items-baseline gap-x-3 gap-y-1">
    <span class="font-mono text-data text-chalk tabular-nums">
      {$chosenTasks.length}
      {$chosenTasks.length === 1 ? "task chosen" : "tasks chosen"}
    </span>
    <span class="text-body text-chalk-faint">
      in {named.join(", ")}
    </span>
    <button type="button" class="{quiet} ml-auto" onclick={clearChosen}>
      choose none
    </button>
  </div>

  <div class="flex flex-wrap items-end gap-x-4 gap-y-2">
    <label class="flex flex-col gap-1">
      <span class={label}>Status</span>
      <select bind:value={status} aria-label="Status for the chosen tasks">
        <option value={KEEP}>leave alone</option>
        {#each statuses as name (name)}
          <option value={name}>{name}</option>
        {/each}
      </select>
    </label>

    <label class="flex flex-col gap-1">
      <span class={label}>Priority</span>
      <select bind:value={priority} aria-label="Priority for the chosen tasks">
        <option value={KEEP}>leave alone</option>
        {#each priorities as name (name)}
          <option value={name}>{name}</option>
        {/each}
      </select>
    </label>

    <label class="flex flex-col gap-1">
      <span class={label}>Milestone</span>
      {#if oneProject}
        <select
          bind:value={milestone}
          aria-label="Milestone for the chosen tasks"
        >
          <option value={KEEP}>leave alone</option>
          <option value={CLEAR}>clear it</option>
          {#each projectMilestones as m (m.id)}
            <option value={m.id}>{m.title}</option>
          {/each}
        </select>
      {:else}
        <span class="max-w-md text-body text-chalk-faint">
          Not across projects — a milestone id belongs to one of them. Choose
          tasks from a single project to set one.
        </span>
      {/if}
    </label>

    <label class="flex flex-col gap-1">
      <span class={label}>Add labels</span>
      <input
        class="w-48"
        bind:value={adding}
        aria-label="Labels to add to the chosen tasks"
        placeholder="comma separated"
      />
    </label>
  </div>

  {#if present.length > 0}
    <div class="flex flex-wrap items-baseline gap-1.5">
      <span class="{label} self-center">Remove</span>
      {#each present as tag (tag)}
        <button
          type="button"
          class="min-h-6 rounded-[2px] border px-1.5 text-micro
                 {removing.includes(tag)
            ? 'border-chalk bg-ink text-chalk line-through'
            : 'border-rule text-chalk-dim hover:text-chalk'}"
          aria-pressed={removing.includes(tag)}
          onclick={() => toggleRemoval(tag)}
        >
          {tag}
        </button>
      {/each}
    </div>
  {/if}

  <div class="flex flex-wrap items-baseline gap-3">
    <button
      type="button"
      class="border-b border-chalk-faint text-body text-chalk
             hover:border-chalk disabled:text-chalk-faint"
      disabled={!asked || busy}
      onclick={apply}
    >
      {#if busy}
        changing {done} of {$chosenTasks.length}…
      {:else}
        Change {$chosenTasks.length}
        {$chosenTasks.length === 1 ? "task" : "tasks"}
      {/if}
    </button>

    {#if refused}
      <span class="text-body text-chalk">{refused}</span>
    {:else if outcome}
      <span class="text-body text-chalk-dim">
        {outcome.changed}
        {outcome.changed === 1 ? "task changed" : "tasks changed"}{outcome
          .failures.length > 0
          ? `, ${outcome.failures.length} did not`
          : "."}
      </span>
    {/if}
  </div>

  {#if outcome && outcome.failures.length > 0}
    <ul class="flex flex-col gap-0.5">
      {#each outcome.failures as failure (failure.project + failure.taskId)}
        <li class="text-body text-chalk-dim">
          <span class="font-mono text-data text-chalk">{failure.taskId}</span>
          in {failure.projectName} — {failure.detail}
        </li>
      {/each}
    </ul>
  {/if}
</section>
