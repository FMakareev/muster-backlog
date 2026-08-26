<script lang="ts">
  import { BoardService } from "../../bindings/github.com/FMakareev/muster-backlog/internal/app";
  import type { SearchHit } from "../../bindings/github.com/FMakareev/muster-backlog/internal/app/models";
  import { projectColour } from "./colour";
  import { openEntity, showSearch } from "./ui";

  /**
   * Cross-project search.
   *
   * With hundreds of tasks across nine projects, finding one by remembering
   * its title is the commonest navigation there is. Opens on a key from
   * anywhere, and the first result is selected so Enter alone is enough.
   */

  let query = $state("");
  let hits = $state<SearchHit[]>([]);
  let at = $state(0);
  let field: HTMLInputElement | undefined = $state();
  let token = 0;

  $effect(() => {
    if ($showSearch) field?.focus();
  });

  async function run(text: string): Promise<void> {
    const mine = ++token;
    const found = await BoardService.Search(text, 60);
    // A slower earlier search must not overwrite a later one's results.
    if (mine !== token) return;
    hits = found ?? [];
    at = 0;
  }

  function onInput(value: string): void {
    query = value;
    void run(value);
  }

  function choose(hit: SearchHit): void {
    // Routed by kind, so a document hit opens the viewer rather than looking
    // for a task that does not exist.
    openEntity({
      project: hit.project,
      kind: hit.kind,
      class: hit.class,
      id: hit.id,
    });
    showSearch.set(false);
  }

  /**
   * Escape closes search from anywhere, not only from the field.
   *
   * Clicking a result list or losing focus for any other reason used to leave
   * the window stuck open until the caret was put back in the input, which is
   * exactly when a person reaches for Escape.
   */
  function onWindowKeydown(event: KeyboardEvent): void {
    if (!$showSearch || event.key !== "Escape") return;
    event.preventDefault();
    event.stopPropagation();
    showSearch.set(false);
  }

  function onKeydown(event: KeyboardEvent): void {
    event.stopPropagation();
    switch (event.key) {
      case "Escape":
        showSearch.set(false);
        break;
      case "ArrowDown":
        event.preventDefault();
        at = Math.min(at + 1, hits.length - 1);
        break;
      case "ArrowUp":
        event.preventDefault();
        at = Math.max(at - 1, 0);
        break;
      case "Enter":
        event.preventDefault();
        if (hits[at]) choose(hits[at]);
        break;
    }
  }
</script>

<svelte:window onkeydown={onWindowKeydown} />

{#if $showSearch}
  <!-- Clicking away closes it, which is what a person expects of something
       floating over everything else. -->
  <!-- svelte-ignore a11y_click_events_have_key_events -->
  <!-- svelte-ignore a11y_no_static_element_interactions -->
  <div
    class="absolute inset-0 z-30 flex items-start justify-center bg-ink/70 pt-16"
    onclick={(e) => {
      if (e.target === e.currentTarget) showSearch.set(false);
    }}
  >
    <div
      role="dialog"
      aria-label="Search"
      tabindex="-1"
      class="flex max-h-[70%] w-[46rem] flex-col overflow-hidden rounded-sm border
             border-rule bg-ink-sunken"
      onkeydown={onKeydown}
    >
      <input
        bind:this={field}
        class="m-3 shrink-0"
        placeholder="Search every project"
        aria-label="Search every project"
        value={query}
        oninput={(e) => onInput(e.currentTarget.value)}
      />

      {#if query.trim() && hits.length === 0}
        <p class="px-3 pb-3 text-body text-chalk-faint">
          Nothing matches “{query.trim()}”.
        </p>
      {/if}

      <ul class="min-h-0 flex-1 overflow-y-auto">
        {#each hits as hit, i (hit.project + hit.kind + hit.class + hit.id)}
          <li>
            <button
              type="button"
              class="flex w-full items-baseline gap-2 border-b border-rule px-3 py-1.5
                     text-left {i === at
                ? 'bg-ink-raised'
                : 'hover:bg-ink-raised'}"
              onclick={() => choose(hit)}
              onmouseenter={() => (at = i)}
            >
              <span
                class="h-2 w-2 shrink-0 self-center rounded-[1px]"
                style="background-color: {projectColour(hit.project, '')}"
                title={hit.projectName}
              ></span>
              <span class="w-16 shrink-0 font-mono text-micro text-chalk-faint">
                {hit.id}
              </span>
              <span class="min-w-0 flex-1">
                <span class="text-body text-chalk">{hit.title}</span>
                {#if hit.field === "body"}
                  <span class="block text-micro text-chalk-faint"
                    >{hit.excerpt}</span
                  >
                {/if}
              </span>
              <span class="shrink-0 font-mono text-micro text-chalk-faint">
                {hit.kind === "task"
                  ? hit.projectName
                  : `${hit.kind} · ${hit.projectName}`}
              </span>
            </button>
          </li>
        {/each}
      </ul>
    </div>
  </div>
{/if}
