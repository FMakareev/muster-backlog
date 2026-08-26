<script lang="ts">
  import { projects } from "./board";
  import { projectColour } from "./colour";
  import { focusedProject, toggleProject } from "./ui";

  /**
   * The muster roll.
   *
   * Every registered project, always visible, always counted. This is the one
   * thing the application exists to show, so it never collapses and never
   * scrolls out of reach — if a project is registered, it is on screen.
   */

  // A project that cannot be read is drawn without hue: colour identifies a
  // project, and there is nothing here to identify.
  function rule(path: string, colour: string, ok: boolean): string {
    return ok ? projectColour(path, colour) : "var(--color-chalk-faint)";
  }
</script>

<nav
  class="flex w-roll shrink-0 flex-col overflow-y-auto border-r border-rule bg-ink-sunken pl-2"
  aria-label="Projects"
>
  <h2
    class="px-2.5 pt-3 pb-1 text-micro font-medium tracking-[0.14em] text-chalk-faint uppercase"
  >
    Projects
  </h2>

  <ul class="flex flex-col">
    {#each $projects as project (project.path)}
      {@const focused = $focusedProject === project.path}
      <li>
        <button
          type="button"
          class="group flex w-full items-baseline gap-2 border-l-2 py-1 pr-3 pl-2.5 text-left
                 hover:bg-ink-raised
                 {focused ? 'bg-ink-raised' : ''}"
          style="border-left-color: {rule(
            project.path,
            project.colour,
            project.ok,
          )}"
          aria-pressed={focused}
          title={project.ok ? project.path : project.problem}
          onclick={() => toggleProject(project.path)}
        >
          <span
            class="min-w-0 flex-1 truncate text-body
                   {focused
              ? 'text-chalk'
              : 'text-chalk-dim group-hover:text-chalk'}
                   {project.ok ? '' : 'line-through decoration-chalk-faint'}"
          >
            {project.name}
          </span>
          <span
            class="shrink-0 font-mono text-data tabular-nums text-chalk-faint"
          >
            {project.ok ? project.taskCount : "—"}
          </span>
        </button>
      </li>
    {/each}
  </ul>

  {#if $projects.length === 0}
    <p class="px-3 py-2 text-body text-chalk-faint">None registered.</p>
  {/if}
</nav>
