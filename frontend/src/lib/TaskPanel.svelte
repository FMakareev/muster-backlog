<script lang="ts">
  import { BoardService } from "../../bindings/github.com/FMakareev/muster-backlog/internal/app";
  import type { TaskView } from "../../bindings/github.com/FMakareev/muster-backlog/internal/app/models";
  import { refresh, settings } from "./board";
  import { projectColour } from "./colour";
  import { notify } from "./notices";
  import CriteriaEditor from "./CriteriaEditor.svelte";
  import Markdown from "./Markdown.svelte";
  import SectionEditor from "./SectionEditor.svelte";
  import TaskControls from "./TaskControls.svelte";
  import TaskLifecycle from "./TaskLifecycle.svelte";
  import TaskRelations from "./TaskRelations.svelte";
  import {
    closeTask,
    findInProject,
    openTask,
    selectedTask,
    toggleTaskView,
  } from "./ui";

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

  // Every editable section is shown whether or not it has content: an empty
  // plan you can start writing is more useful than a heading that is missing.
  const sections: [string, string][] = [
    ["description", "Description"],
    ["plan", "Implementation plan"],
    ["notes", "Implementation notes"],
  ];

  function sectionSource(key: string): string {
    const map = entity?.Sections as Record<string, string> | undefined;
    return map?.[key] ?? "";
  }

  const centred = $derived($settings.taskView === "centred");

  let renaming = $state(false);
  let titleDraft = $state("");
  let busyTitle = $state(false);

  function startRename(): void {
    titleDraft = entity?.Title ?? "";
    renaming = true;
  }

  async function saveTitle(): Promise<void> {
    if (!task || busyTitle || !titleDraft.trim()) return;
    busyTitle = true;
    const result = await BoardService.SetTitle(
      task.project,
      task.id,
      titleDraft.trim(),
    );
    busyTitle = false;
    if (!result.ok) {
      notify(
        result.problem
          ? `${result.problem.title}: ${result.problem.detail}`
          : "The title could not be changed.",
      );
      return;
    }
    renaming = false;
    await refresh();
  }

  /**
   * Subtasks are fetched when a task is opened, not carried on every card.
   *
   * The board asks for hundreds of cards at once and each shows only a count;
   * the list with titles and statuses is worth a request only here.
   */
  let subtasks = $state<TaskView[]>([]);
  let subtaskToken = 0;

  $effect(() => {
    const open = task;
    if (!open) {
      subtasks = [];
      return;
    }
    const mine = ++subtaskToken;
    void BoardService.Subtasks(
      open.project,
      open.kind,
      open.class,
      open.id,
    ).then((found) => {
      // A slower earlier request must not land on a later task.
      if (mine !== subtaskToken) return;
      subtasks = found ?? [];
    });
  });

  const parent = $derived(task?.family?.parent ?? null);

  function openRef(ref: {
    project: string;
    kind: string;
    class: string;
    id: string;
  }): void {
    openTask({
      project: ref.project,
      kind: ref.kind,
      class: ref.class,
      id: ref.id,
    });
  }

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
    class={centred
      ? "absolute inset-x-0 top-8 z-10 mx-auto flex max-h-[calc(100%-4rem)] w-full max-w-3xl flex-col overflow-y-auto rounded-sm border border-rule bg-ink-sunken shadow-2xl"
      : "flex w-[34rem] shrink-0 flex-col overflow-y-auto border-l border-rule bg-ink-sunken"}
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
        title={centred
          ? "Show this in the side panel instead"
          : "Show this centred, which reads better on a wide screen"}
        onclick={toggleTaskView}
      >
        {centred ? "to the side" : "centre"}
      </button>
      <button
        type="button"
        class="font-mono text-data text-chalk-faint hover:text-chalk"
        onclick={closeTask}
      >
        close esc
      </button>
    </header>

    <div class="flex flex-col gap-4 px-4 py-3">
      {#if entity.ParentTaskID}
        <p class="-mb-2 flex items-baseline gap-2 text-data text-chalk-faint">
          <span>Subtask of</span>
          {#if parent}
            <button
              type="button"
              class="border-b border-chalk-faint font-mono text-chalk hover:border-chalk"
              onclick={() => openRef(parent)}
            >
              {parent.id}
            </button>
            <span class="truncate text-chalk-dim"
              >{task.family?.parentTitle}</span
            >
          {:else}
            <span
              class="font-mono"
              title="No task with this id in {task.projectName}"
            >
              {entity.ParentTaskID}
            </span>
          {/if}
        </p>
      {/if}

      {#if renaming}
        <input
          class="w-full"
          bind:value={titleDraft}
          disabled={busyTitle}
          aria-label="Task title"
          onkeydown={(e) => {
            e.stopPropagation();
            if (e.key === "Escape") renaming = false;
            if (e.key === "Enter") void saveTitle();
          }}
          onblur={() => void saveTitle()}
        />
      {:else}
        <button
          type="button"
          class="text-left text-title font-semibold text-chalk hover:underline
                 decoration-chalk-faint underline-offset-4"
          title="Rename"
          onclick={startRename}
        >
          {entity.Title}
        </button>
      {/if}

      <TaskControls {task} />

      <dl class="grid grid-cols-[auto_1fr] gap-x-3 gap-y-1 text-data">
        {#snippet field(label: string, value: string)}
          {#if value}
            <dt class="text-chalk-faint">{label}</dt>
            <dd class="text-chalk-dim">{value}</dd>
          {/if}
        {/snippet}
        <!-- Status, priority, assignee, milestone and labels are not repeated
             here: the controls above already show them, and showing a value
             twice a few pixels apart only invites the two to disagree. -->
        {@render field("Type", entity.Type)}
        {@render field("Updated", shortDate(String(entity.Updated ?? "")))}
        {@render field("Created", shortDate(String(entity.Created ?? "")))}
      </dl>

      <TaskRelations
        {task}
        onOpen={follow}
        resolves={(id) => dependency(id) !== null}
      />

      {#if subtasks.length > 0}
        <section>
          <h3
            class="text-micro font-medium tracking-[0.14em] text-chalk-faint uppercase"
          >
            Subtasks {task.family?.done ?? 0}/{task.family?.total ??
              subtasks.length}
          </h3>
          <ul class="mt-1 flex flex-col">
            {#each subtasks as child (child.project + child.class + child.id)}
              <li>
                <button
                  type="button"
                  class="flex w-full items-baseline gap-2 py-1 text-left hover:bg-ink"
                  onclick={() => openRef(child)}
                >
                  <span class="font-mono text-data text-chalk-faint"
                    >{child.id}</span
                  >
                  <span class="min-w-0 flex-1 truncate text-data text-chalk"
                    >{child.entity.Title}</span
                  >
                  <span class="font-mono text-micro text-chalk-faint"
                    >{child.entity.Status}</span
                  >
                </button>
              </li>
            {/each}
          </ul>
        </section>
      {/if}

      <CriteriaEditor
        project={task.project}
        taskID={task.id}
        criteria={entity.AcceptanceCriteria}
      />

      {#each sections as [key, heading] (key)}
        <SectionEditor
          project={task.project}
          taskID={task.id}
          section={key}
          {heading}
          source={sectionSource(key)}
        />
      {/each}

      {#if sectionSource("final_summary").trim()}
        <section>
          <h3
            class="text-micro font-medium tracking-[0.14em] text-chalk-faint uppercase"
          >
            Final summary
          </h3>
          <div class="mt-1">
            <Markdown
              source={sectionSource("final_summary")}
              project={task.project}
              {onTaskRef}
            />
          </div>
        </section>
      {/if}

      {#if task.kind === "task" && task.class === "active"}
        <!-- Last, because it is what happens at the end and because two of
             the three take the task off the board. -->
        <TaskLifecycle {task} />
      {/if}

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
