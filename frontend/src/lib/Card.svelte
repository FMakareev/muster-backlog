<script lang="ts">
  import type { KanbanCard } from "@svar-ui/svelte-kanban";
  import type { TaskView } from "../../bindings/github.com/FMakareev/muster-backlog/internal/app/models";
  import { projectColour } from "./colour";

  /** SVAR passes its own card through, with the task the board attached to it. */
  interface Props {
    card: KanbanCard & { task?: TaskView };
  }
  let { card }: Props = $props();

  const task = $derived(card.task);

  /**
   * Every card carries its project's colour.
   *
   * With nine projects in one column, a card that does not say where it came
   * from is a card you have to click to understand. The roll on the left is
   * the legend for this.
   */
  const colour = $derived(
    task ? projectColour(task.project, task.projectColour) : "transparent",
  );
</script>

{#if task}
  <div class="flex flex-col gap-1 py-0.5" style="--project: {colour}">
    <div class="flex items-baseline gap-2">
      <span
        class="h-2 w-2 shrink-0 self-center rounded-[1px]"
        style="background-color: {colour}"
        title={task.projectName}
      ></span>
      <span class="font-mono text-micro text-chalk-faint">{task.id}</span>
      {#if task.entity.Priority}
        <span
          class="font-mono text-micro tracking-wide text-chalk-faint uppercase"
        >
          {task.entity.Priority}
        </span>
      {/if}
      {#if task.entity.Type}
        <span class="ml-auto font-mono text-micro text-chalk-faint">
          {task.entity.Type}
        </span>
      {/if}
    </div>

    <span class="text-body text-chalk">{task.entity.Title}</span>

    {#if task.entity.Milestone || (task.entity.Labels?.length ?? 0) > 0 || (task.entity.Assignee?.length ?? 0) > 0}
      <div class="flex flex-wrap items-baseline gap-x-2 gap-y-0.5">
        {#if task.entity.Milestone}
          <span class="font-mono text-micro text-chalk-faint">
            {task.entity.Milestone}
          </span>
        {/if}
        {#each task.entity.Labels ?? [] as label (label)}
          <span
            class="rounded-[2px] bg-ink-sunken px-1 text-micro text-chalk-dim"
          >
            {label}
          </span>
        {/each}
        {#each task.entity.Assignee ?? [] as who (who)}
          <span class="text-micro text-chalk-faint">{who}</span>
        {/each}
      </div>
    {/if}
  </div>
{/if}
