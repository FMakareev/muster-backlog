<script lang="ts">
  import { onMount } from "svelte";
  import { Kanban, WillowDark } from "@svar-ui/svelte-kanban";
  import {
    columns,
    connect,
    loading,
    problems,
    projects,
    registryPath,
    tasks,
  } from "./lib/board";

  onMount(() => connect());

  // SVAR takes ownership of the arrays it is given, and nanostores expose their
  // values as readonly, so the board gets copies.
  const boardColumns = $derived(
    $columns.map((status) => ({ id: status, label: status })),
  );

  const boardCards = $derived(
    $tasks.map((task, index) => ({
      id: index + 1,
      label: task.entity.Title,
      column: task.entity.Status,
      project: task.projectName,
    })),
  );
</script>

<div class="flex h-full flex-col">
  <header
    class="flex items-center gap-3 border-b border-shell-border bg-shell-raised px-4 py-2"
  >
    <h1 class="text-sm font-semibold tracking-wide">Muster</h1>
    <span class="text-xs text-shell-muted">
      {#if $loading}
        loading…
      {:else}
        {$tasks.length} tasks across {$projects.length}
        {$projects.length === 1 ? "project" : "projects"}
      {/if}
    </span>
  </header>

  {#if $problems.length > 0}
    <section class="border-b border-shell-border bg-shell-raised px-4 py-2">
      {#each $problems.slice(0, 3) as problem (problem.title + problem.path)}
        <p class="text-xs">
          <span class="font-medium">{problem.title}</span>
          <span class="text-shell-muted"> — {problem.detail}</span>
          {#if problem.path}
            <span class="font-mono text-shell-muted">{problem.path}</span>
          {/if}
        </p>
      {/each}
      {#if $problems.length > 3}
        <p class="text-xs text-shell-muted">
          and {$problems.length - 3} more
        </p>
      {/if}
    </section>
  {/if}

  <main class="min-h-0 flex-1">
    {#if !$loading && $projects.length === 0}
      <div class="p-6 text-sm text-shell-muted">
        <p>No projects are registered yet.</p>
        <p class="mt-2">
          Create <span class="font-mono">{$registryPath}</span> listing the folders
          that hold your Backlog.md projects.
        </p>
      </div>
    {:else}
      <WillowDark>
        <Kanban cards={boardCards} columns={boardColumns} />
      </WillowDark>
    {/if}
  </main>
</div>
