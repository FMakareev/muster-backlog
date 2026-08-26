<script lang="ts">
  import { problems, projects, registryPath, tasks } from "./board";
  import { focusedProject, visibleTasks } from "./ui";

  /** One line at the foot of the window: what is on screen, and what is wrong.
   *
   * It counts what is shown, not what is loaded. Reporting the whole corpus
   * while the board is narrowed to one project would be a quietly wrong number
   * in the one place a person looks to check their bearings. */

  const loadedProjects = $derived($projects.filter((p) => p.ok).length);
  const brokenProjects = $derived($projects.length - loadedProjects);
  const narrowed = $derived($focusedProject !== "");

  // Having no registry yet is a first run, not a fault. The main area already
  // says so and offers the way out; counting it here as a problem would
  // contradict that in the same window.
  const faults = $derived($problems.filter((p) => p.kind !== "no_registry"));
</script>

<footer
  class="flex h-strip shrink-0 items-center gap-3 border-t border-rule bg-ink-sunken px-3
         font-mono text-data text-chalk-faint"
>
  <span class="tabular-nums">
    {#if narrowed}
      {$visibleTasks.length} of {$tasks.length} tasks · 1 of {loadedProjects} projects
    {:else}
      {$tasks.length} tasks · {loadedProjects} projects
    {/if}
  </span>

  {#if brokenProjects > 0}
    <span class="text-chalk-dim">
      {brokenProjects} unreadable
    </span>
  {/if}

  {#if faults.length > 0}
    <span class="text-chalk-dim">
      {faults.length}
      {faults.length === 1 ? "problem" : "problems"}
    </span>
  {/if}

  <span class="ml-auto truncate" title={$registryPath}>{$registryPath}</span>
</footer>
