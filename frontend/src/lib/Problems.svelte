<script lang="ts">
  import { problems } from "./board";
  import { showProblems } from "./ui";

  /**
   * Everything currently wrong, in one place.
   *
   * The status strip counts problems and the banner shows the first; without
   * this, a skipped file would be a number nobody could act on. Opened from
   * that count and closed by Escape.
   */

  const kindLabel: Record<string, string> = {
    no_registry: "No projects yet",
    registry: "Registry",
    project: "Project",
    file: "Skipped file",
    cli: "Backlog.md CLI",
  };

  function onKeydown(event: KeyboardEvent): void {
    if (event.key === "Escape") {
      event.stopPropagation();
      showProblems.set(false);
    }
  }
</script>

{#if $showProblems}
  <!-- svelte-ignore a11y_no_noninteractive_element_to_interactive_role -->
  <aside
    role="dialog"
    aria-label="Problems"
    tabindex="-1"
    class="flex max-h-[50%] shrink-0 flex-col overflow-y-auto border-t border-rule bg-ink-sunken"
    onkeydown={onKeydown}
  >
    <header
      class="sticky top-0 flex items-baseline gap-2 border-b border-rule bg-ink-sunken px-3 py-1.5"
    >
      <h2 class="text-body font-medium">
        {$problems.length}
        {$problems.length === 1 ? "problem" : "problems"}
      </h2>
      <button
        type="button"
        class="ml-auto font-mono text-data text-chalk-faint hover:text-chalk"
        onclick={() => showProblems.set(false)}
      >
        close esc
      </button>
    </header>

    {#if $problems.length === 0}
      <p class="px-3 py-2 text-body text-chalk-faint">Nothing is wrong.</p>
    {:else}
      <ul class="flex flex-col">
        {#each $problems as problem, i (problem.kind + problem.path + i)}
          <li
            class="flex items-baseline gap-2 border-b border-rule px-3 py-1.5"
          >
            <span
              class="w-24 shrink-0 font-mono text-micro tracking-wide text-chalk-faint uppercase"
            >
              {kindLabel[problem.kind] ?? problem.kind}
            </span>
            <span class="min-w-0 flex-1">
              <span class="text-body text-chalk">{problem.title}</span>
              <span class="text-body text-chalk-dim"> — {problem.detail}</span>
              {#if problem.path}
                <span class="block font-mono text-micro text-chalk-faint">
                  {problem.path}
                </span>
              {/if}
            </span>
          </li>
        {/each}
      </ul>
    {/if}
  </aside>
{/if}
