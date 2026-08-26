<script lang="ts">
  import { projectColour } from "./colour";
  import Markdown from "./Markdown.svelte";
  import TaskControls from "./TaskControls.svelte";
  import { closeTask, findInProject, openTask, selectedTask } from "./ui";

  /**
   * The task panel.
   *
   * The whole point of this application over a table of tasks is fast access
   * to the body — description, acceptance criteria, plan, notes — so the panel
   * sits beside the board rather than over it, and shows everything at once
   * rather than behind tabs.
   */

  let panel: HTMLElement | undefined = $state();

  const task = $derived($selectedTask);
  const entity = $derived(task?.entity);
  const colour = $derived(
    task ? projectColour(task.project, task.projectColour) : "transparent",
  );

  // Focus moves to the panel when it opens, so a keyboard user is not left
  // behind on the board.
  $effect(() => {
    if (task && panel) panel.focus();
  });

  function onKeydown(event: KeyboardEvent): void {
    if (event.key === "Escape") {
      event.stopPropagation();
      closeTask();
    }
  }

  /** A dependency resolves inside its own project, or not at all. */
  function dependency(id: string) {
    return task ? findInProject(task.project, id) : null;
  }

  function follow(id: string): void {
    const found = dependency(id);
    if (found) {
      openTask({
        project: found.project,
        kind: found.kind,
        class: found.class,
        id: found.id,
      });
    }
  }

  function onTaskRef(id: string, project: string): void {
    const found = findInProject(project || (task?.project ?? ""), id);
    if (found) {
      openTask({
        project: found.project,
        kind: found.kind,
        class: found.class,
        id: found.id,
      });
    }
  }

  const sections = $derived(
    entity
      ? ([
          ["Description", entity.Sections?.description],
          ["Implementation plan", entity.Sections?.plan],
          ["Implementation notes", entity.Sections?.notes],
          ["Final summary", entity.Sections?.final_summary],
        ].filter(([, body]) => (body ?? "").trim() !== "") as [
          string,
          string,
        ][])
      : [],
  );

  function shortDate(value: string): string {
    if (!value || value.startsWith("0001-01-01")) return "";
    return value.slice(0, 16).replace("T", " ");
  }
</script>

{#if task && entity}
  <!-- svelte-ignore a11y_no_noninteractive_element_to_interactive_role -->
  <aside
    bind:this={panel}
    tabindex="-1"
    role="dialog"
    aria-label="Task details"
    class="flex w-[34rem] shrink-0 flex-col overflow-y-auto border-l border-rule bg-ink-sunken"
    onkeydown={onKeydown}
  >
    <header
      class="sticky top-0 flex items-baseline gap-2 border-b border-rule bg-ink-sunken px-4 py-2"
    >
      <span
        class="h-2 w-2 shrink-0 self-center rounded-[1px]"
        style="background-color: {colour}"
      ></span>
      <span class="font-mono text-data text-chalk-faint"
        >{task.projectName}</span
      >
      <span class="font-mono text-data text-chalk">{task.id}</span>
      <button
        type="button"
        class="ml-auto font-mono text-data text-chalk-faint hover:text-chalk"
        onclick={closeTask}
      >
        close esc
      </button>
    </header>

    <div class="flex flex-col gap-4 px-4 py-3">
      <h2 class="text-title font-semibold text-chalk">{entity.Title}</h2>

      <TaskControls {task} />

      <dl class="grid grid-cols-[auto_1fr] gap-x-3 gap-y-1 text-data">
        {#snippet field(label: string, value: string)}
          {#if value}
            <dt class="text-chalk-faint">{label}</dt>
            <dd class="text-chalk-dim">{value}</dd>
          {/if}
        {/snippet}
        <!-- Status, priority, assignee and labels are not repeated here: the
             controls above already show them, and showing a value twice a few
             pixels apart only invites the two to disagree. -->
        {@render field("Type", entity.Type)}
        {@render field("Milestone", entity.Milestone)}
        {@render field("Updated", shortDate(String(entity.Updated ?? "")))}
        {@render field("Created", shortDate(String(entity.Created ?? "")))}
      </dl>

      {#if (entity.Dependencies?.length ?? 0) > 0}
        <section>
          <h3
            class="text-micro font-medium tracking-[0.14em] text-chalk-faint uppercase"
          >
            Depends on
          </h3>
          <ul class="mt-1 flex flex-wrap gap-2">
            {#each entity.Dependencies ?? [] as id (id)}
              {@const found = dependency(id)}
              <li>
                {#if found}
                  <button
                    type="button"
                    class="border-b border-chalk-faint font-mono text-data text-chalk hover:border-chalk"
                    title={found.entity.Title}
                    onclick={() => follow(id)}
                  >
                    {id}
                  </button>
                {:else}
                  <span
                    class="font-mono text-data text-chalk-faint"
                    title="No task with this id in {task.projectName}"
                  >
                    {id}
                  </span>
                {/if}
              </li>
            {/each}
          </ul>
        </section>
      {/if}

      {#if entity.AcceptanceCriteria}
        <section>
          <h3
            class="text-micro font-medium tracking-[0.14em] text-chalk-faint uppercase"
          >
            Acceptance criteria
          </h3>
          {#if entity.AcceptanceCriteria.length === 0}
            <p class="mt-1 text-body text-chalk-faint">None recorded.</p>
          {:else}
            <ul class="mt-1 flex flex-col gap-1">
              {#each entity.AcceptanceCriteria as criterion (criterion.Index)}
                <li class="flex items-baseline gap-2">
                  <input
                    type="checkbox"
                    checked={criterion.Checked}
                    disabled
                    class="self-center accent-chalk-dim"
                    aria-label={criterion.Text}
                  />
                  <span class="font-mono text-micro text-chalk-faint">
                    #{criterion.Index}
                  </span>
                  <span
                    class="text-body {criterion.Checked
                      ? 'text-chalk-faint line-through'
                      : 'text-chalk-dim'}"
                  >
                    {criterion.Text}
                  </span>
                </li>
              {/each}
            </ul>
          {/if}
        </section>
      {/if}

      {#each sections as [heading, body] (heading)}
        <section>
          <h3
            class="text-micro font-medium tracking-[0.14em] text-chalk-faint uppercase"
          >
            {heading}
          </h3>
          <div class="mt-1">
            <Markdown source={body} project={task.project} {onTaskRef} />
          </div>
        </section>
      {/each}

      {#if (entity.Comments?.length ?? 0) > 0}
        <section>
          <h3
            class="text-micro font-medium tracking-[0.14em] text-chalk-faint uppercase"
          >
            Comments
          </h3>
          <ul class="mt-1 flex flex-col gap-2">
            {#each entity.Comments ?? [] as comment, i (i)}
              <li class="border-l border-rule pl-2">
                <p class="font-mono text-micro text-chalk-faint">
                  {comment.Author}
                  {comment.Created}
                </p>
                <Markdown
                  source={comment.Body}
                  project={task.project}
                  {onTaskRef}
                />
              </li>
            {/each}
          </ul>
        </section>
      {/if}

      <p class="font-mono text-micro text-chalk-faint">{entity.Path}</p>
    </div>
  </aside>
{/if}
