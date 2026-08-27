<script lang="ts">
  import { BoardService } from "../../bindings/github.com/FMakareev/muster-backlog/internal/app";
  import { columns, projects, refresh } from "./board";
  import { dismissable } from "./overlay";
  import MilestonePicker from "./MilestonePicker.svelte";
  import { notify } from "./notices";
  import {
    closeNewTask,
    defaultProject,
    openTask,
    rememberProject,
    showNewTask,
  } from "./ui";

  /**
   * Creating a task.
   *
   * A task manager that cannot create a task sends you to a terminal, which
   * breaks the loop this application exists to close. Everything offered here
   * comes from the target project's own configuration.
   */

  let project = $state("");
  let title = $state("");
  let description = $state("");
  let status = $state("");
  let priority = $state("");
  let type = $state("");
  let milestone = $state("");
  let assignee = $state("");
  let labels = $state("");
  let criteria = $state("");
  // Capture into the inbox instead of onto the board. Same form, because the
  // decision of whether a note is ready to be a task is exactly what triage is
  // for - it should not also decide how much you are allowed to write down.
  let asNote = $state(false);
  let busy = $state(false);
  let titleField: HTMLInputElement | undefined = $state();

  const open = $derived($showNewTask);
  const healthy = $derived($projects.filter((p) => p.ok));
  const config = $derived(healthy.find((p) => p.path === project));

  // Latched rather than derived: the project is chosen when the form opens,
  // and a project list that reloads underneath it must not throw away what
  // the person has since picked.
  let opened = false;

  $effect(() => {
    if (!open) {
      opened = false;
      return;
    }
    if (opened) return;
    opened = true;
    // Set on every opening, not only when it is empty. The select binds to
    // the first option the moment it renders, which used to count as a
    // choice and left the focused project ignored.
    project = defaultProject();
    titleField?.focus();
  });

  const statuses = $derived(config?.statuses ?? []);
  const priorities = $derived(config?.priorities ?? []);
  const types = $derived(config?.types ?? []);

  function reset(): void {
    title = "";
    description = "";
    status = "";
    priority = "";
    type = "";
    milestone = "";
    assignee = "";
    labels = "";
    criteria = "";
  }

  function splitLines(value: string): string[] {
    return value
      .split("\n")
      .map((line) => line.trim())
      .filter(Boolean);
  }

  async function submit(): Promise<void> {
    if (!project || !title.trim() || busy) return;
    busy = true;

    const common = {
      title: title.trim(),
      description,
      priority,
      type,
      milestone,
      assignee,
      labels: labels
        .split(",")
        .map((l) => l.trim())
        .filter(Boolean),
      acceptanceCriteria: splitLines(criteria),
    };

    // A draft's status is Draft and the CLI takes no other, so the one field
    // that does not cross over is left behind rather than sent and refused.
    // The two calls are kept apart rather than merged: only one of them
    // answers with the id of what it made.
    let ok: boolean;
    let problem: { title: string; detail: string } | null | undefined;
    let created = "";

    if (asNote) {
      const result = await BoardService.CaptureNote(project, {
        ...common,
        project: "",
      });
      ok = result.ok;
      problem = result.problem;
    } else {
      const result = await BoardService.CreateTask({
        ...common,
        project,
        status,
      });
      ok = result.ok;
      problem = result.problem;
      created = result.taskId;
    }
    busy = false;

    if (!ok) {
      // Nothing typed is thrown away: the form stays as it was so the person
      // can fix whatever the CLI objected to.
      notify(
        problem
          ? `${problem.title}: ${problem.detail}`
          : asNote
            ? "The note could not be captured."
            : "The task could not be created.",
      );
      return;
    }

    await refresh();
    if (created) {
      openTask({ project, kind: "task", class: "active", id: created });
    }

    if (asNote) {
      // Confirmed where the eye already is, and the form clears itself rather
      // than closing: capture is worth nothing if the next thought has to
      // reopen it. Nothing takes focus away from the field.
      const name = healthy.find((p) => p.path === project)?.name ?? "the inbox";
      notify(`Captured into ${name}.`, 3000);
      await rememberProject(project);
      reset();
      titleField?.focus();
      return;
    }

    reset();
    closeNewTask();
  }

  /**
   * Switch project without leaving the text.
   *
   * Reaching for a select with the mouse is exactly what a capture form is
   * for avoiding, and tabbing to it loses the caret.
   */
  function cycleProject(by: number): void {
    if (healthy.length < 2) return;
    const at = healthy.findIndex((p) => p.path === project);
    const next = (at + by + healthy.length) % healthy.length;
    project = healthy[next].path;
  }

  function onKeydown(event: KeyboardEvent): void {
    event.stopPropagation();

    // Alt with a bracket moves between projects from anywhere in the form,
    // including the middle of a sentence.
    if (event.altKey && (event.key === "[" || event.key === "]")) {
      event.preventDefault();
      cycleProject(event.key === "]" ? 1 : -1);
      return;
    }
    // The whole form submits from anywhere in it, since reaching for a button
    // with the mouse is exactly what this is meant to avoid.
    if (event.key === "Enter" && (event.metaKey || event.ctrlKey)) {
      event.preventDefault();
      void submit();
    }
  }

  const field = "w-full";
  const label =
    "text-micro font-medium uppercase tracking-[0.14em] text-chalk-faint";
</script>

{#if open}
  <div
    class="absolute inset-0 z-20 flex items-start justify-center bg-ink/70 pt-16"
    use:dismissable={closeNewTask}
  >
    <!-- svelte-ignore a11y_no_noninteractive_element_to_interactive_role -->
    <form
      role="dialog"
      aria-label="New task"
      class="flex max-h-[80%] w-[42rem] flex-col gap-3 overflow-y-auto rounded-sm
             border border-rule bg-ink-sunken p-4"
      onkeydown={onKeydown}
      onsubmit={(e) => {
        e.preventDefault();
        void submit();
      }}
    >
      <header class="flex items-baseline gap-3">
        <h2 class="text-title font-semibold">
          {asNote ? "New note" : "New task"}
        </h2>
        <label class="flex items-center gap-2 text-body text-chalk-dim">
          <input type="checkbox" bind:checked={asNote} disabled={busy} />
          capture into the inbox
        </label>
        <span class="ml-auto flex items-center gap-2">
          {#if healthy.length > 1}
            <kbd class="font-mono text-micro text-chalk-faint">alt+[ ]</kbd>
          {/if}
          <select
            aria-label="Project this task belongs to"
            bind:value={project}
            disabled={busy}
          >
            {#each healthy as p (p.path)}
              <option value={p.path}>{p.name}</option>
            {/each}
          </select>
        </span>
      </header>

      <label class="flex flex-col gap-1">
        <span class={label}>Title</span>
        <input
          bind:this={titleField}
          class={field}
          bind:value={title}
          disabled={busy}
          placeholder="What needs doing"
        />
      </label>

      <label class="flex flex-col gap-1">
        <span class={label}>Description</span>
        <textarea
          class="{field} min-h-24 font-mono"
          bind:value={description}
          disabled={busy}
          placeholder="Why it matters, and what done looks like"></textarea>
      </label>

      <label class="flex flex-col gap-1">
        <span class={label}>Acceptance criteria — one per line</span>
        <textarea
          class="{field} min-h-20 font-mono"
          bind:value={criteria}
          disabled={busy}></textarea>
      </label>

      <div class="grid grid-cols-2 gap-3">
        <label class="flex flex-col gap-1">
          <span class={label}>Status</span>
          {#if asNote}
            <!-- Said rather than silently dropped: a draft's status is Draft
                 and Backlog.md accepts no other for one. -->
            <span class="text-body text-chalk-faint">
              A note is a draft until it is promoted.
            </span>
          {:else}
            <select class={field} bind:value={status} disabled={busy}>
              <option value="">project default</option>
              {#each statuses as s (s)}
                <option value={s}>{s}</option>
              {/each}
            </select>
          {/if}
        </label>

        <label class="flex flex-col gap-1">
          <span class={label}>Priority</span>
          <select class={field} bind:value={priority} disabled={busy}>
            <option value="">none</option>
            {#each priorities as p (p)}
              <option value={p}>{p}</option>
            {/each}
          </select>
        </label>

        <label class="flex flex-col gap-1">
          <span class={label}>Type</span>
          <select class={field} bind:value={type} disabled={busy}>
            <option value="">none</option>
            {#each types as t (t)}
              <option value={t}>{t}</option>
            {/each}
          </select>
        </label>

        <MilestonePicker
          {project}
          value={milestone}
          disabled={busy}
          onChange={(next) => (milestone = next)}
        />

        <label class="flex flex-col gap-1">
          <span class={label}>Assignee</span>
          <input
            class={field}
            bind:value={assignee}
            disabled={busy}
            placeholder="@someone"
          />
        </label>

        <label class="flex flex-col gap-1">
          <span class={label}>Labels — comma separated</span>
          <input class={field} bind:value={labels} disabled={busy} />
        </label>
      </div>

      <footer class="flex items-baseline gap-3">
        <span class="font-mono text-micro text-chalk-faint">
          {$columns.length ? "ctrl+enter to create · esc to cancel" : ""}
        </span>
        <button
          type="button"
          class="ml-auto text-body text-chalk-dim hover:text-chalk"
          onclick={closeNewTask}
          disabled={busy}
        >
          Cancel
        </button>
        <button
          type="submit"
          class="rounded-sm border border-rule bg-ink px-3 py-1 text-body text-chalk
                 disabled:opacity-50"
          disabled={busy || !title.trim() || !project}
        >
          {busy ? "Creating…" : "Create task"}
        </button>
      </footer>
    </form>
  </div>
{/if}
