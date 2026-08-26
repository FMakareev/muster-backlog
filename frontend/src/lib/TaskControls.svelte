<script lang="ts">
  import { BoardService } from "../../bindings/github.com/FMakareev/muster-backlog/internal/app";
  import type { TaskView } from "../../bindings/github.com/FMakareev/muster-backlog/internal/app/models";
  import { applyWrite, canMove, columns, projectConfig } from "./board";
  import { notify } from "./notices";

  interface Props {
    task: TaskView;
  }
  let { task }: Props = $props();

  /**
   * The controls that change a task.
   *
   * Every one of them writes through the backlog CLI and then re-reads, so
   * what is shown afterwards is what the file says. None of them assumes.
   */

  const config = $derived(projectConfig(task.project));

  // A task can only take a status its own project declares, so the choices are
  // that project's list rather than the whole board's columns.
  const statuses = $derived(
    $columns.map((c) => c.name).filter((name) => canMove(task.project, name)),
  );
  const priorities = $derived(config?.priorities ?? []);

  let busy = $state(false);
  let newLabel = $state("");

  async function run(action: () => Promise<{ ok: boolean; problem: unknown }>) {
    busy = true;
    await applyWrite(action as never, (message) => notify(message));
    busy = false;
  }

  const control =
    "rounded-sm border border-rule bg-ink px-1.5 py-0.5 text-data text-chalk " +
    "disabled:opacity-50";
</script>

<div class="flex flex-wrap items-center gap-2">
  <label class="flex items-center gap-1">
    <span class="text-micro tracking-[0.14em] text-chalk-faint uppercase"
      >Status</span
    >
    <select
      class={control}
      value={task.entity.Status}
      disabled={busy}
      onchange={(e) =>
        run(() =>
          BoardService.SetStatus(task.project, task.id, e.currentTarget.value),
        )}
    >
      {#each statuses as status (status)}
        <option value={status}>{status}</option>
      {/each}
    </select>
  </label>

  <label class="flex items-center gap-1">
    <span class="text-micro tracking-[0.14em] text-chalk-faint uppercase"
      >Priority</span
    >
    <select
      class={control}
      value={priorities.find(
        (p) => p.toLowerCase() === (task.entity.Priority ?? "").toLowerCase(),
      ) ?? ""}
      disabled={busy}
      onchange={(e) =>
        run(() =>
          BoardService.SetPriority(
            task.project,
            task.id,
            e.currentTarget.value,
          ),
        )}
    >
      <option value="">none</option>
      {#each priorities as priority (priority)}
        <option value={priority}>{priority}</option>
      {/each}
    </select>
  </label>

  <label class="flex items-center gap-1">
    <span class="text-micro tracking-[0.14em] text-chalk-faint uppercase"
      >Assignee</span
    >
    <input
      class="{control} w-32"
      value={(task.entity.Assignee ?? []).join(", ")}
      disabled={busy}
      placeholder="@someone"
      onchange={(e) =>
        run(() =>
          BoardService.SetAssignee(
            task.project,
            task.id,
            e.currentTarget.value,
          ),
        )}
    />
  </label>

  <span class="flex flex-wrap items-center gap-1">
    <span class="text-micro tracking-[0.14em] text-chalk-faint uppercase"
      >Labels</span
    >
    {#each task.entity.Labels ?? [] as label (label)}
      <button
        type="button"
        class="rounded-[2px] bg-ink-sunken px-1 text-micro text-chalk-dim hover:text-chalk"
        title="Remove {label}"
        disabled={busy}
        onclick={() =>
          run(() => BoardService.RemoveLabel(task.project, task.id, label))}
      >
        {label} ×
      </button>
    {/each}
    <input
      class="{control} w-24"
      bind:value={newLabel}
      disabled={busy}
      placeholder="add"
      onkeydown={(e) => {
        if (e.key !== "Enter" || !newLabel.trim()) return;
        const label = newLabel.trim();
        newLabel = "";
        void run(() => BoardService.AddLabel(task.project, task.id, label));
      }}
    />
  </span>
</div>
