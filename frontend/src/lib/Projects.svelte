<script lang="ts">
  import { BoardService } from "../../bindings/github.com/FMakareev/muster-backlog/internal/app";
  import type { ProjectView } from "../../bindings/github.com/FMakareev/muster-backlog/internal/app/models";
  import { milestones, projects, refresh, setProjectHidden } from "./board";
  import AddProject from "./AddProject.svelte";
  import { projectColour } from "./colour";
  import Milestones from "./Milestones.svelte";
  import { notify } from "./notices";

  /**
   * The registry, as a screen rather than a file.
   *
   * projects.yml stays exactly what it was — a small file a person can read
   * and edit — and this writes to it in place, keeping the comments and the
   * order. Nothing here deletes a backlog: removing a project unregisters the
   * folder and leaves everything in it alone.
   */

  let adding = $state(false);
  let busy = $state("");
  let confirming = $state("");
  // Which project's milestones are open. Closed by default: this is the
  // registry screen, and most visits to it are not about milestones.
  let planning = $state("");

  function milestonesOf(path: string) {
    return $milestones.filter((m) => m.project === path);
  }

  async function save(
    project: ProjectView,
    change: { name?: string; colour?: string },
  ): Promise<void> {
    busy = project.path;
    const result = await BoardService.SaveProject(project.path, {
      name: change.name ?? project.name,
      colour: change.colour ?? project.colour,
      hidden: project.hidden,
    });
    busy = "";
    if (!result.ok) {
      notify(
        result.problem
          ? `${result.problem.title}: ${result.problem.detail}`
          : "The project could not be changed.",
      );
    }
    await refresh();
  }

  // Hiding goes through the shared write rather than through save(): the roll
  // offers it too, and the two must not drift apart.
  async function hide(project: ProjectView, hidden: boolean): Promise<void> {
    busy = project.path;
    await setProjectHidden(project, hidden);
    busy = "";
  }

  async function move(project: ProjectView, by: number): Promise<void> {
    const at = $projects.findIndex((p) => p.path === project.path);
    if (at < 0) return;
    busy = project.path;
    const result = await BoardService.MoveProject(project.path, at + by);
    busy = "";
    if (!result.ok && result.problem) {
      notify(`${result.problem.title}: ${result.problem.detail}`);
    }
    await refresh();
  }

  async function remove(project: ProjectView): Promise<void> {
    busy = project.path;
    const result = await BoardService.RemoveProject(project.path);
    busy = "";
    confirming = "";
    if (!result.ok && result.problem) {
      notify(`${result.problem.title}: ${result.problem.detail}`);
    }
    await refresh();
  }

  /**
   * The display name is only stored when it differs from the folder name.
   *
   * Writing the folder name into the file as an override would freeze it: the
   * folder could be renamed and the registry would keep insisting on the old
   * name, for no reason anyone could see.
   */
  function rename(project: ProjectView, value: string): void {
    const next = value.trim();
    const folder = project.path.split("/").filter(Boolean).pop() ?? "";
    void save(project, { name: next === folder ? "" : next });
  }

  const hiddenCount = $derived($projects.filter((p) => p.hidden).length);
  const label =
    "text-micro font-medium uppercase tracking-[0.14em] text-chalk-faint";
</script>

<div class="flex min-h-0 min-w-0 flex-1 flex-col overflow-y-auto">
  <div
    class="flex shrink-0 items-baseline gap-3 border-b border-rule px-3 py-1.5"
  >
    <span class="font-mono text-data tabular-nums text-chalk-faint">
      {$projects.length}
      {$projects.length === 1 ? "project" : "projects"}{hiddenCount
        ? `, ${hiddenCount} hidden`
        : ""}
    </span>
    <button
      type="button"
      class="ml-auto font-mono text-data text-chalk-faint hover:text-chalk"
      aria-expanded={adding}
      onclick={() => (adding = !adding)}
    >
      add a folder
    </button>
  </div>

  {#if adding}
    <AddProject onDone={() => (adding = false)} />
  {/if}

  {#if $projects.length === 0 && !adding}
    <p class="px-3 py-3 text-body text-chalk-faint">
      No projects yet. Add a folder that holds a Backlog.md project, or one that
      should.
    </p>
  {/if}

  <ul class="flex flex-col">
    {#each $projects as project, at (project.path)}
      {@const colour = projectColour(project.path, project.colour)}
      <li
        class="flex flex-col gap-1.5 border-b border-rule px-3 py-2
               {project.hidden ? 'opacity-60' : ''}"
      >
        <div class="flex items-baseline gap-2">
          <span
            class="h-2 w-2 shrink-0 self-center rounded-[1px]"
            style="background-color: {project.ok
              ? colour
              : 'var(--color-chalk-faint)'}"
          ></span>

          <input
            class="w-56 shrink-0"
            value={project.name}
            aria-label="Name for {project.path}"
            disabled={busy === project.path}
            onblur={(e) => rename(project, e.currentTarget.value)}
            onkeydown={(e) => {
              if (e.key === "Enter") e.currentTarget.blur();
            }}
          />

          <input
            class="w-24 shrink-0 font-mono"
            value={project.colour}
            aria-label="Colour for {project.path}"
            placeholder="auto"
            disabled={busy === project.path}
            onblur={(e) => save(project, { colour: e.currentTarget.value })}
            onkeydown={(e) => {
              if (e.key === "Enter") e.currentTarget.blur();
            }}
          />

          <!-- Shown abbreviated, with the whole path in the tooltip: this
               screen gets screenshotted, and a home directory spelled out
               carries a username for no reason. -->
          <span
            class="min-w-0 flex-1 truncate font-mono text-data text-chalk-faint"
            title={project.path}
          >
            {project.displayPath || project.path}
          </span>

          <span
            class="shrink-0 font-mono text-data tabular-nums text-chalk-dim"
          >
            {#if project.ok}
              {project.taskCount} tasks{project.draftCount
                ? ` · ${project.draftCount} drafts`
                : ""}
            {:else}
              —
            {/if}
          </span>
        </div>

        <div class="flex flex-wrap items-baseline gap-x-3 gap-y-1">
          {#if project.ok}
            <span class="font-mono text-micro text-chalk-faint">
              {project.layout} · {(project.statuses ?? []).join(" › ")}
            </span>
            {#each milestonesOf(project.path) as milestone (milestone.id)}
              <span
                class="rounded-[2px] border border-rule px-1 text-micro text-chalk-dim"
                title={milestone.title}
              >
                ◇ {milestone.id}
                {milestone.done}/{milestone.total}
              </span>
            {/each}
            <button
              type="button"
              class="min-h-6 font-mono text-micro text-chalk-faint hover:text-chalk"
              aria-expanded={planning === project.path}
              onclick={() =>
                (planning = planning === project.path ? "" : project.path)}
            >
              {planning === project.path ? "done" : "milestones"}
            </button>
          {:else}
            <span class="text-body text-chalk">{project.problem}</span>
          {/if}

          <span class="ml-auto flex shrink-0 items-baseline gap-3">
            <button
              type="button"
              class="inline-flex min-h-6 min-w-6 items-center justify-center font-mono
                     text-data text-chalk-faint hover:text-chalk"
              aria-label="Move {project.name} up"
              disabled={at === 0 || busy === project.path}
              onclick={() => move(project, -1)}
            >
              ↑
            </button>
            <button
              type="button"
              class="inline-flex min-h-6 min-w-6 items-center justify-center font-mono
                     text-data text-chalk-faint hover:text-chalk"
              aria-label="Move {project.name} down"
              disabled={at === $projects.length - 1 || busy === project.path}
              onclick={() => move(project, 1)}
            >
              ↓
            </button>
            <button
              type="button"
              class="min-h-6 font-mono text-data text-chalk-faint hover:text-chalk"
              aria-pressed={project.hidden}
              disabled={busy === project.path}
              onclick={() => hide(project, !project.hidden)}
            >
              {project.hidden ? "show" : "hide"}
            </button>
            {#if confirming === project.path}
              <button
                type="button"
                class="min-h-6 font-mono text-data text-chalk hover:underline"
                onclick={() => remove(project)}
              >
                unregister, really
              </button>
              <button
                type="button"
                class="min-h-6 font-mono text-data text-chalk-faint hover:text-chalk"
                onclick={() => (confirming = "")}
              >
                keep
              </button>
            {:else}
              <button
                type="button"
                class="min-h-6 font-mono text-data text-chalk-faint hover:text-chalk"
                disabled={busy === project.path}
                onclick={() => (confirming = project.path)}
              >
                unregister
              </button>
            {/if}
          </span>
        </div>

        {#if planning === project.path}
          <Milestones
            project={project.path}
            milestones={milestonesOf(project.path)}
          />
        {/if}
      </li>
    {/each}
  </ul>

  {#if $projects.length > 0}
    <p class="px-3 py-2 text-body text-chalk-faint">
      <span class={label}>Note</span>
      Unregistering removes the folder from Muster and leaves everything in it alone.
      Hiding keeps a project registered and out of the board, the lists, search and
      the figures.
    </p>
  {/if}
</div>
