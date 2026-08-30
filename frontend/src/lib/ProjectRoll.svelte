<script lang="ts">
  import type { ProjectView } from "../../bindings/github.com/FMakareev/muster-backlog/internal/app/models";
  import { projects, setProjectHidden } from "./board";
  import { projectColour } from "./colour";
  import { focusedProject, toggleProject } from "./ui";

  /**
   * The muster roll.
   *
   * Every registered project, always visible, always counted. This is the one
   * thing the application exists to show, so it never collapses and never
   * scrolls out of reach — if a project is registered, it is on screen.
   */

  // Hidden projects are left out here, which is what hiding one means. They
  // are still registered and still on the Projects screen.
  const shown = $derived($projects.filter((p) => !p.hidden));

  // And they are counted at the foot of the roll, because a project that
  // leaves without a trace is a project nobody finds their way back to. This
  // is also the only place that says hiding exists at all.
  const putAway = $derived($projects.filter((p) => p.hidden));

  // Closed by default: they were set aside, and a roll that lists them anyway
  // has not put anything away.
  let listing = $state(false);
  let busy = $state("");

  async function set(project: ProjectView, hidden: boolean): Promise<void> {
    // Hiding the project you are looking through empties the board and says
    // nothing about why. Focusing is a temporary narrowing, so it is the part
    // that gives way.
    if (hidden && $focusedProject === project.path) toggleProject(project.path);
    busy = project.path;
    await setProjectHidden(project, hidden);
    busy = "";
  }

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
    {#each shown as project (project.path)}
      {@const focused = $focusedProject === project.path}
      <li class="group flex items-stretch {focused ? 'bg-ink-raised' : ''}">
        <button
          type="button"
          class="flex min-w-0 flex-1 items-baseline gap-2 border-l-2 py-1 pl-2.5 text-left
                 group-hover:bg-ink-raised"
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

        <!-- The room this takes is held open whether or not the control is
             drawn. A control that appears on hover and pushes the name over
             makes the roll twitch under the cursor; an empty 2.5rem never
             moves. The button itself is always here for anything that is not
             a mouse — it is only the ink that waits. -->
        <button
          type="button"
          class="inline-flex w-10 shrink-0 items-center justify-center pr-1
                 font-mono text-micro text-chalk-faint opacity-0
                 group-hover:bg-ink-raised group-hover:opacity-100
                 hover:text-chalk focus-visible:opacity-100"
          aria-label="Hide {project.name} from the board"
          title="Take {project.name} off the board. It stays registered."
          disabled={busy === project.path}
          onclick={() => set(project, true)}
        >
          hide
        </button>
      </li>
    {/each}
  </ul>

  {#if shown.length === 0}
    <p class="px-3 py-2 text-body text-chalk-faint">
      {$projects.length === 0 ? "None registered." : "All of them are hidden."}
    </p>
  {/if}

  {#if putAway.length > 0}
    <div class="mt-auto border-t border-rule">
      <button
        type="button"
        class="flex min-h-6 w-full items-baseline gap-1.5 px-2.5 py-1 text-left
               font-mono text-micro text-chalk-faint hover:text-chalk"
        aria-expanded={listing}
        onclick={() => (listing = !listing)}
      >
        <span class="tabular-nums">{putAway.length}</span>
        hidden
        <span class="ml-auto" aria-hidden="true">{listing ? "×" : "…"}</span>
      </button>

      {#if listing}
        <ul class="flex flex-col pb-1">
          {#each putAway as project (project.path)}
            <li class="group flex items-baseline gap-2 pr-1 pl-2.5">
              <span class="min-w-0 flex-1 truncate text-body text-chalk-dim">
                {project.name}
              </span>
              <button
                type="button"
                class="inline-flex min-h-6 shrink-0 items-center font-mono text-micro
                       text-chalk-faint hover:text-chalk"
                aria-label="Show {project.name} on the board"
                disabled={busy === project.path}
                onclick={() => set(project, false)}
              >
                show
              </button>
            </li>
          {/each}
        </ul>
      {/if}
    </div>
  {/if}
</nav>
