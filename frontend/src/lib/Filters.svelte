<script lang="ts">
  import { dismissable } from "./overlay";
  import { milestoneLabel, milestones, tasks } from "./board";
  import {
    clearFilters,
    filters,
    filtersActive,
    showFilters,
    toggleFilter,
    visibleTasks,
    type Filters,
  } from "./ui";

  /**
   * Board filters.
   *
   * The choices are the values actually in use rather than everything the
   * projects configure: a filter for a priority nothing carries is a filter
   * that can only ever return nothing.
   */

  function distinct(pick: (t: (typeof $tasks)[number]) => string[]): string[] {
    // A plain array rather than a Set: this is derived once per render from
    // the task list and never mutated afterwards, so reactive collection
    // semantics would buy nothing.
    const seen: string[] = [];
    for (const task of $tasks) {
      for (const raw of pick(task)) {
        const value = raw?.trim();
        if (value && !seen.includes(value)) seen.push(value);
      }
    }
    return seen.sort((a, b) => a.localeCompare(b));
  }

  const statuses = $derived(distinct((t) => [t.entity.Status]));
  const priorities = $derived(distinct((t) => [t.entity.Priority]));
  const types = $derived(distinct((t) => [t.entity.Type]));
  const labels = $derived(distinct((t) => t.entity.Labels ?? []));
  const milestoneValues = $derived(distinct((t) => [t.entity.Milestone]));

  let field: HTMLInputElement | undefined = $state();

  // Opening a filter panel and then having to reach for the mouse to type in
  // it defeats the point of a keyboard shortcut.
  $effect(() => {
    if ($showFilters) field?.focus();
  });

  const groups: [keyof Filters, string, () => string[]][] = [
    ["statuses", "Status", () => statuses],
    ["priorities", "Priority", () => priorities],
    ["types", "Type", () => types],
    ["milestones", "Milestone", () => milestoneValues],
    ["labels", "Label", () => labels],
  ];

  function label(field: keyof Filters, value: string): string {
    if (field !== "milestones") return value;
    // A milestone reads as its title here for the same reason it does on a
    // card: the id alone says nothing.
    const project = $tasks.find((t) => t.entity.Milestone === value)?.project;
    return project ? milestoneLabel(project, value, $milestones) : value;
  }

  function chosen(field: keyof Filters, value: string): boolean {
    const list = $filters[field];
    return Array.isArray(list) && list.includes(value);
  }
</script>

{#if $showFilters}
  <!-- svelte-ignore a11y_no_noninteractive_element_to_interactive_role -->
  <aside
    role="dialog"
    aria-label="Filters"
    tabindex="-1"
    class="flex max-h-64 shrink-0 flex-col gap-2 overflow-y-auto border-b border-rule
           bg-ink-sunken px-3 py-2"
    use:dismissable={() => showFilters.set(false)}
  >
    <div class="sticky top-0 flex items-baseline gap-3 bg-ink-sunken pb-1">
      <input
        bind:this={field}
        class="w-64"
        placeholder="Filter by title"
        value={$filters.text}
        aria-label="Filter by title"
        oninput={(e) =>
          filters.set({ ...$filters, text: e.currentTarget.value })}
      />
      <span class="font-mono text-data text-chalk-faint tabular-nums">
        {$visibleTasks.length} of {$tasks.length}
      </span>
      {#if $filtersActive}
        <button
          type="button"
          class="font-mono text-data text-chalk-dim hover:text-chalk"
          onclick={clearFilters}
        >
          clear all
        </button>
      {/if}
      <button
        type="button"
        class="ml-auto font-mono text-data text-chalk-faint hover:text-chalk"
        onclick={() => showFilters.set(false)}
      >
        close esc
      </button>
    </div>

    {#each groups as [field, heading, values] (field)}
      {@const list = values()}
      {#if list.length > 0}
        <div class="flex flex-wrap items-baseline gap-1.5 gap-y-2">
          <span
            class="w-16 shrink-0 text-micro font-medium tracking-[0.14em] text-chalk-faint uppercase"
          >
            {heading}
          </span>
          {#each list as value (value)}
            <button
              type="button"
              class="min-h-6 rounded-[2px] border px-2 py-0.5 text-micro
                     {chosen(field, value)
                ? 'border-chalk bg-ink text-chalk'
                : 'border-rule text-chalk-dim hover:text-chalk'}"
              aria-pressed={chosen(field, value)}
              onclick={() => toggleFilter(field, value)}
            >
              {label(field, value)}
            </button>
          {/each}
        </div>
      {/if}
    {/each}
  </aside>
{/if}
