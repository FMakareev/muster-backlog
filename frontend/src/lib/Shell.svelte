<script lang="ts">
  import type { Snippet } from "svelte";
  import { drafts, loading, problems, projects, registryPath } from "./board";
  import { screenFor, screens } from "./screens";
  import {
    cycleGrouping,
    filtersActive,
    groupBy,
    openNewTask,
    screen,
    showFilters,
    showSearch,
    showSettings,
  } from "./ui";
  import { dismiss, notices } from "./notices";
  import ProjectRoll from "./ProjectRoll.svelte";
  import Filters from "./Filters.svelte";
  import NewTask from "./NewTask.svelte";
  import Search from "./Search.svelte";
  import Preferences from "./Preferences.svelte";
  import Problems from "./Problems.svelte";
  import StatusStrip from "./StatusStrip.svelte";

  interface Props {
    children: Snippet;
  }
  let { children }: Props = $props();

  /**
   * Screen keys work from anywhere in the window, except while typing. A tool
   * driven from the keyboard has to give the keys up the moment a text field
   * has focus, or it eats the text.
   */
  function onKeydown(event: KeyboardEvent): void {
    if (event.metaKey || event.ctrlKey || event.altKey) return;

    const target = event.target as HTMLElement | null;
    if (
      target?.isContentEditable ||
      target instanceof HTMLInputElement ||
      target instanceof HTMLTextAreaElement ||
      target instanceof HTMLSelectElement
    ) {
      return;
    }

    if (event.key.toLowerCase() === "g") {
      event.preventDefault();
      void cycleGrouping();
      return;
    }
    if (event.key.toLowerCase() === "n") {
      event.preventDefault();
      openNewTask();
      return;
    }
    if (event.key === "/") {
      event.preventDefault();
      showSearch.set(true);
      return;
    }
    if (event.key.toLowerCase() === "f") {
      event.preventDefault();
      showFilters.set(!showFilters.get());
      return;
    }
    if (event.key === ",") {
      event.preventDefault();
      showSettings.set(!showSettings.get());
      return;
    }

    const next = screenFor(event.key.toLowerCase());
    if (next) {
      event.preventDefault();
      screen.set(next.id);
    }
  }

  // A missing registry is handled by the empty state below, not by the banner.
  const faults = $derived($problems.filter((p) => p.kind !== "no_registry"));
  const firstProblem = $derived(faults[0]);
  const noProjects = $derived(!$loading && $projects.length === 0);

  // The longest a note has waited. Drafts arrive oldest first, so this is the
  // first one; it decides whether the count on the nav reads as a number or as
  // something to do something about.
  const oldest = $derived($drafts[0]?.waitingDays ?? 0);
</script>

<svelte:window onkeydown={onKeydown} />

<div class="flex h-full flex-col bg-ink">
  <header
    class="flex h-bar shrink-0 items-center gap-4 border-b border-rule bg-ink-raised px-3"
  >
    <span class="text-title font-semibold tracking-tight">Muster</span>

    <nav class="flex items-center gap-0.5" aria-label="Screens">
      {#each screens as item (item.id)}
        {@const active = $screen === item.id}
        <button
          type="button"
          class="flex items-baseline gap-1.5 rounded-sm px-2 py-1 text-body
                 {active ? 'bg-ink text-chalk' : 'text-chalk-dim'}
                 {item.available
            ? 'hover:text-chalk'
            : 'cursor-default text-chalk-faint'}"
          aria-current={active ? "page" : undefined}
          disabled={!item.available}
          title={item.available
            ? item.purpose
            : `${item.purpose} — not built yet`}
          onclick={() => item.available && screen.set(item.id)}
        >
          {item.label}
          {#if item.id === "inbox" && $drafts.length > 0}
            <!-- The depth, on the nav itself. An inbox you have to open to
                 find out whether it is filling up is the pile nobody reads. -->
            <span
              class="rounded-[2px] bg-ink-sunken px-1 font-mono text-micro tabular-nums
                     {oldest >= 30 ? 'text-chalk' : 'text-chalk-dim'}"
              title="{$drafts.length} waiting{oldest > 0
                ? `, the oldest for ${oldest} days`
                : ''}"
            >
              {$drafts.length}
            </span>
          {/if}
          {#if item.available}
            <kbd class="font-mono text-micro text-chalk-faint">{item.key}</kbd>
          {/if}
        </button>
      {/each}
    </nav>

    <button
      type="button"
      class="ml-auto flex items-baseline gap-1.5 rounded-sm px-2 py-1 text-body
             text-chalk-dim hover:text-chalk"
      title="Search every project"
      onclick={() => showSearch.set(true)}
    >
      Search
      <kbd class="font-mono text-micro text-chalk-faint">/</kbd>
    </button>

    <button
      type="button"
      class="flex items-baseline gap-1.5 rounded-sm px-2 py-1 text-body
             {$filtersActive
        ? 'bg-ink text-chalk'
        : 'text-chalk-dim hover:text-chalk'}"
      title="Narrow what the board and the list show"
      onclick={() => showFilters.set(!showFilters.get())}
    >
      Filters
      <kbd class="font-mono text-micro text-chalk-faint">f</kbd>
    </button>

    <button
      type="button"
      class="flex items-baseline gap-1.5 rounded-sm px-2 py-1 text-body
             text-chalk-dim hover:text-chalk"
      title="Create a task"
      onclick={() => openNewTask()}
    >
      New task
      <kbd class="font-mono text-micro text-chalk-faint">n</kbd>
    </button>

    <button
      type="button"
      class="flex items-baseline gap-1.5 rounded-sm px-2 py-1 text-body
             {$groupBy
        ? 'bg-ink text-chalk'
        : 'text-chalk-dim hover:text-chalk'}"
      title="Keep cards of the same project or milestone together in a column"
      onclick={() => void cycleGrouping()}
    >
      {$groupBy === "project"
        ? "Grouped by project"
        : $groupBy === "milestone"
          ? "Grouped by milestone"
          : "Not grouped"}
      <kbd class="font-mono text-micro text-chalk-faint">g</kbd>
    </button>

    <button
      type="button"
      class="rounded-sm px-2 py-1 text-body text-chalk-dim hover:text-chalk"
      title="Preferences"
      onclick={() => showSettings.set(!showSettings.get())}
    >
      Preferences
      <kbd class="font-mono text-micro text-chalk-faint">,</kbd>
    </button>

    <span class="font-mono text-data text-chalk-faint tabular-nums">
      {#if $loading}loading{/if}
    </span>
  </header>

  <div class="flex min-h-0 flex-1">
    <ProjectRoll />

    <main class="flex min-h-0 min-w-0 flex-1 flex-col">
      {#if $loading}
        <p class="p-6 text-body text-chalk-faint">Reading your projects…</p>
      {:else if noProjects}
        <div class="max-w-xl p-6">
          <h2 class="text-title font-semibold">No projects yet</h2>
          <p class="mt-2 text-body text-chalk-dim">
            Muster reads Backlog.md projects you list for it. Create this file
            and name the folders that hold them:
          </p>
          <p class="mt-3 font-mono text-data text-chalk">{$registryPath}</p>
          <pre
            class="mt-3 overflow-x-auto rounded-sm bg-ink-sunken p-3 font-mono text-data text-chalk-dim">projects:
  - path: ~/Dev/treeline
  - path: ~/Dev/refloft
    name: Refloft
    color: "#7aa2f7"</pre>
          <p class="mt-3 text-body text-chalk-dim">
            The window picks the file up as soon as you save it.
          </p>
        </div>
      {:else}
        {#if firstProblem}
          <div
            class="flex shrink-0 items-baseline gap-2 border-b border-rule bg-ink-raised px-3 py-1.5"
            role="status"
          >
            <span class="text-body text-chalk">{firstProblem.title}</span>
            <span class="min-w-0 flex-1 truncate text-body text-chalk-dim">
              {firstProblem.detail}
            </span>
            {#if faults.length > 1}
              <span class="shrink-0 font-mono text-data text-chalk-faint">
                +{faults.length - 1}
              </span>
            {/if}
          </div>
        {/if}
        {#each $notices as notice (notice.id)}
          <div
            class="flex shrink-0 items-baseline gap-2 border-b border-rule bg-ink-raised px-3 py-1.5"
            role="status"
          >
            <span class="min-w-0 flex-1 text-body text-chalk"
              >{notice.text}</span
            >
            <button
              type="button"
              class="shrink-0 font-mono text-data text-chalk-faint hover:text-chalk"
              onclick={() => dismiss(notice.id)}
            >
              dismiss
            </button>
          </div>
        {/each}

        <Filters />
        {@render children()}
      {/if}
    </main>
  </div>

  <Search />
  <NewTask />
  <Preferences />
  <Problems />
  <StatusStrip />
</div>
