<script lang="ts">
  import type { Snippet } from "svelte";
  import { loading, problems, projects, registryPath } from "./board";
  import { screenFor, screens } from "./screens";
  import { groupByProject, screen, toggleGrouping } from "./ui";
  import { dismiss, notices } from "./notices";
  import ProjectRoll from "./ProjectRoll.svelte";
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
      toggleGrouping();
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
          {#if item.available}
            <kbd class="font-mono text-micro text-chalk-faint">{item.key}</kbd>
          {/if}
        </button>
      {/each}
    </nav>

    <button
      type="button"
      class="ml-auto flex items-baseline gap-1.5 rounded-sm px-2 py-1 text-body
             {$groupByProject
        ? 'bg-ink text-chalk'
        : 'text-chalk-dim hover:text-chalk'}"
      aria-pressed={$groupByProject}
      title="Keep each project's cards together inside a column"
      onclick={toggleGrouping}
    >
      Group by project
      <kbd class="font-mono text-micro text-chalk-faint">g</kbd>
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

        {@render children()}
      {/if}
    </main>
  </div>

  <StatusStrip />
</div>
