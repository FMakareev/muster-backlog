<script lang="ts">
  import { BoardService } from "../../bindings/github.com/FMakareev/muster-backlog/internal/app";
  import type { TaskView } from "../../bindings/github.com/FMakareev/muster-backlog/internal/app/models";
  import { refresh } from "./board";
  import { notify } from "./notices";

  /**
   * What a task points at: dependencies, references, documentation, and the
   * files it touched.
   *
   * All four are lists the CLI replaces wholesale rather than patches, so each
   * is edited as text — one per line — and sent as a whole. That also makes
   * the one asymmetry visible: the file list can be set but not emptied,
   * because Backlog.md has no command that empties it.
   */

  interface Props {
    task: TaskView;
    /** Called with a task id to open it, for the dependency links. */
    onOpen: (id: string) => void;
    /** Whether a dependency id resolves inside this project. */
    resolves: (id: string) => boolean;
  }
  let { task, onOpen, resolves }: Props = $props();

  type Field = {
    key: string;
    label: string;
    hint: string;
    read: (t: TaskView) => string[];
    write: (values: string[]) => Promise<{
      ok: boolean;
      problem?: { title: string; detail: string } | null;
    }>;
    /** Dependencies are ids in this project, so they are shown as links. */
    links?: boolean;
  };

  const fields: Field[] = [
    {
      key: "dependencies",
      label: "Depends on",
      hint: "Task ids in this project, one per line",
      read: (t) => t.entity.Dependencies ?? [],
      write: (v) => BoardService.SetDependencies(task.project, task.id, v),
      links: true,
    },
    {
      key: "references",
      label: "References",
      hint: "Links or paths, one per line",
      read: (t) => t.entity.References ?? [],
      write: (v) => BoardService.SetReferences(task.project, task.id, v),
    },
    {
      key: "documentation",
      label: "Documentation",
      hint: "Documents this task belongs to, one per line",
      read: (t) => t.entity.Documentation ?? [],
      write: (v) => BoardService.SetDocumentation(task.project, task.id, v),
    },
    {
      key: "modified",
      label: "Files touched",
      hint: "Paths from the project root, one per line. Backlog.md cannot empty this list.",
      read: (t) => t.entity.ModifiedFiles ?? [],
      write: (v) => BoardService.SetModifiedFiles(task.project, task.id, v),
    },
  ];

  let editing = $state("");
  let draft = $state("");
  let busy = $state("");

  function start(field: Field): void {
    editing = field.key;
    draft = field.read(task).join("\n");
  }

  async function save(field: Field): Promise<void> {
    busy = field.key;
    const values = draft
      .split("\n")
      .map((line) => line.trim())
      .filter(Boolean);
    const result = await field.write(values);
    busy = "";
    if (!result.ok) {
      // The draft stays on screen, because whatever was typed is what needs
      // fixing.
      notify(
        result.problem
          ? `${result.problem.title}: ${result.problem.detail}`
          : `${field.label} could not be changed.`,
      );
      return;
    }
    editing = "";
    await refresh();
  }

  const empty = $derived(
    fields.filter((f) => f.read(task).length === 0 && editing !== f.key),
  );

  const heading =
    "text-micro font-medium tracking-[0.14em] text-chalk-faint uppercase";
</script>

{#each fields as field (field.key)}
  {@const values = field.read(task)}
  {#if values.length > 0 || editing === field.key}
    <section>
      <div class="flex items-baseline gap-3">
        <h3 class={heading}>{field.label}</h3>
        {#if editing !== field.key}
          <button
            type="button"
            class="ml-auto font-mono text-data text-chalk-faint hover:text-chalk"
            onclick={() => start(field)}
          >
            edit
          </button>
        {/if}
      </div>

      {#if editing === field.key}
        <div class="mt-1 flex flex-col gap-1">
          <textarea
            class="min-h-16 w-full font-mono"
            bind:value={draft}
            aria-label={field.label}></textarea>
          <span class="text-body text-chalk-faint">{field.hint}</span>
          <div class="flex items-baseline gap-3">
            <button
              type="button"
              class="border-b border-chalk-faint text-body text-chalk hover:border-chalk"
              disabled={busy === field.key}
              onclick={() => save(field)}
            >
              {busy === field.key ? "saving…" : "save"}
            </button>
            <button
              type="button"
              class="font-mono text-data text-chalk-faint hover:text-chalk"
              onclick={() => (editing = "")}
            >
              cancel
            </button>
          </div>
        </div>
      {:else}
        <ul class="mt-1 flex flex-wrap gap-2">
          {#each values as value (value)}
            <li>
              {#if field.links && resolves(value)}
                <button
                  type="button"
                  class="border-b border-chalk-faint font-mono text-data text-chalk hover:border-chalk"
                  onclick={() => onOpen(value)}
                >
                  {value}
                </button>
              {:else if field.links}
                <span
                  class="font-mono text-data text-chalk-faint"
                  title="No task with this id in {task.projectName}"
                >
                  {value}
                </span>
              {:else}
                <span class="font-mono text-data text-chalk-dim">{value}</span>
              {/if}
            </li>
          {/each}
        </ul>
      {/if}
    </section>
  {/if}
{/each}

<!-- Whatever the task has none of yet, offered as one line rather than as
     four empty sections. A task with dependencies but no references still
     needs somewhere to add one. -->
{#if empty.length > 0}
  <section class="flex flex-wrap items-baseline gap-x-3 gap-y-1">
    <h3 class={heading}>Add</h3>
    {#each empty as field (field.key)}
      <button
        type="button"
        class="font-mono text-data text-chalk-faint hover:text-chalk"
        onclick={() => start(field)}
      >
        {field.label.toLowerCase()}
      </button>
    {/each}
  </section>
{/if}
