<script lang="ts">
  import { BoardService } from "../../bindings/github.com/FMakareev/muster-backlog/internal/app";
  import type { TaskView } from "../../bindings/github.com/FMakareev/muster-backlog/internal/app/models";
  import { projectConfig, refresh } from "./board";
  import { notify } from "./notices";
  import { closeTask } from "./ui";

  /**
   * What happens to a task at the end.
   *
   * Three things Backlog.md does that the application could not: file a
   * finished task into completed/, archive one, and send one back to the
   * inbox. All three move the file somewhere the board does not look, so each
   * says where it is going, and the two that are not the ordinary end of a
   * task ask first.
   */

  interface Props {
    task: TaskView;
  }
  let { task }: Props = $props();

  let busy = $state("");
  let confirming = $state("");

  const config = $derived(projectConfig(task.project));

  /**
   * Finished means the project's own last declared status.
   *
   * The CLI applies the same rule and refuses anything else, so offering the
   * control only here is offering it exactly where it will work.
   */
  const terminal = $derived((config?.statuses ?? []).at(-1) ?? "");
  const finished = $derived(
    terminal !== "" &&
      (task.entity.Status ?? "").toLowerCase() === terminal.toLowerCase(),
  );

  async function run(
    what: string,
    action: () => Promise<{
      ok: boolean;
      problem?: { title: string; detail: string } | null;
    }>,
    { closes = true }: { closes?: boolean } = {},
  ): Promise<void> {
    busy = what;
    const result = await action();
    busy = "";
    confirming = "";
    if (!result.ok) {
      notify(
        result.problem
          ? `${result.problem.title}: ${result.problem.detail}`
          : `${task.id} could not be ${what}.`,
      );
      return;
    }
    await refresh();
    // The panel is showing a task that is no longer where it was, so it
    // closes rather than sitting there empty.
    if (closes) closeTask();
  }

  const action = "min-h-6 font-mono text-data";
  const quiet = `${action} text-chalk-faint hover:text-chalk`;
</script>

<section class="flex flex-col gap-1">
  <h3
    class="text-micro font-medium tracking-[0.14em] text-chalk-faint uppercase"
  >
    When it is over
  </h3>

  <div class="flex flex-wrap items-baseline gap-x-4 gap-y-1">
    {#if finished}
      <button
        type="button"
        class="{action} text-chalk hover:underline"
        title="Move it into completed/, where Backlog.md keeps finished work"
        disabled={busy !== ""}
        onclick={() =>
          run("completed", () =>
            BoardService.CompleteTask(task.project, task.id),
          )}
      >
        {busy === "completed" ? "filing…" : "file as completed"}
      </button>
    {:else}
      <span class="text-body text-chalk-faint">
        A task is filed once it reaches {terminal || "its last status"}.
      </span>
    {/if}

    <span class="ml-auto flex shrink-0 items-baseline gap-4">
      {#if confirming === "archived"}
        <button
          type="button"
          class="{action} text-chalk hover:underline"
          onclick={() =>
            run("archived", () =>
              BoardService.ArchiveTask(task.project, task.id),
            )}
        >
          archive, really
        </button>
        <button type="button" class={quiet} onclick={() => (confirming = "")}>
          keep
        </button>
      {:else if confirming === "sent back"}
        <button
          type="button"
          class="{action} text-chalk hover:underline"
          onclick={() =>
            run("sent back", () =>
              BoardService.DemoteTask(task.project, task.id),
            )}
        >
          send it back, really
        </button>
        <button type="button" class={quiet} onclick={() => (confirming = "")}>
          keep
        </button>
      {:else}
        <button
          type="button"
          class={quiet}
          title="Back to the inbox as a note. Its id becomes a DRAFT id."
          disabled={busy !== ""}
          onclick={() => (confirming = "sent back")}
        >
          back to the inbox
        </button>
        <button
          type="button"
          class={quiet}
          title="Into archive/tasks. Nothing is deleted, and the board stops showing it."
          disabled={busy !== ""}
          onclick={() => (confirming = "archived")}
        >
          archive
        </button>
      {/if}
    </span>
  </div>
</section>
