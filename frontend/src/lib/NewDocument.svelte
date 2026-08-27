<script lang="ts">
  import { BoardService } from "../../bindings/github.com/FMakareev/muster-backlog/internal/app";
  import type { ProjectView } from "../../bindings/github.com/FMakareev/muster-backlog/internal/app/models";
  import { refresh } from "./board";
  import { notify } from "./notices";
  import { defaultProject } from "./ui";

  /**
   * Writing a document or a decision.
   *
   * Two things the CLI treats quite differently. A document is created empty
   * and filled by a second call, which is one act here. A decision is created
   * with a title and a status and a skeleton body — Backlog.md has no command
   * that fills it in — so the form says where the file is instead of
   * pretending to be an editor.
   */

  interface Props {
    projects: ProjectView[];
    types: string[];
    onDone: (
      ref: { project: string; kind: string; class: string; id: string } | null,
    ) => void;
  }
  let { projects, types, onDone }: Props = $props();

  let kind = $state<"document" | "decision">("document");
  let project = $state("");
  let title = $state("");
  let type = $state("");
  let path = $state("");
  let content = $state("");
  let status = $state("proposed");
  let busy = $state(false);

  // The rule everywhere something has to be created somewhere: the project
  // being looked at, or the first registered one.
  $effect(() => {
    if (!project) project = defaultProject();
  });

  async function create(): Promise<void> {
    if (!project || !title.trim() || busy) return;
    busy = true;

    const result =
      kind === "document"
        ? await BoardService.CreateDocument({
            project,
            title: title.trim(),
            type,
            path,
            content,
          })
        : await BoardService.CreateDecision(project, title.trim(), status);
    busy = false;

    if (!result.ok) {
      notify(
        result.problem
          ? `${result.problem.title}: ${result.problem.detail}`
          : `The ${kind} could not be created.`,
      );
      return;
    }

    await refresh();
    onDone({ project, kind, class: "active", id: result.taskId });
  }

  const label =
    "text-micro font-medium uppercase tracking-[0.14em] text-chalk-faint";
</script>

<section
  class="flex shrink-0 flex-col gap-2 border-b border-rule bg-ink-sunken px-3 py-2"
  aria-label="Write something"
>
  <div class="flex gap-3">
    {#each ["document", "decision"] as choice (choice)}
      <label class="flex items-center gap-1.5 text-body text-chalk-dim">
        <input
          type="radio"
          name="kind"
          value={choice}
          checked={kind === choice}
          onchange={() => (kind = choice as "document" | "decision")}
        />
        {choice}
      </label>
    {/each}
  </div>

  <label class="flex flex-col gap-1">
    <span class={label}>Project</span>
    <select bind:value={project} aria-label="Project for this document">
      {#each projects as p (p.path)}
        <option value={p.path}>{p.name}</option>
      {/each}
    </select>
  </label>

  <label class="flex flex-col gap-1">
    <span class={label}>Title</span>
    <input class="w-full" bind:value={title} aria-label="Title" />
  </label>

  {#if kind === "document"}
    <label class="flex flex-col gap-1">
      <span class={label}>Type</span>
      <select bind:value={type} aria-label="Type">
        <option value="">Backlog.md's default</option>
        {#each types as value (value)}
          <option {value}>{value}</option>
        {/each}
      </select>
    </label>

    <label class="flex flex-col gap-1">
      <span class={label}>Sub-path</span>
      <input
        class="w-full font-mono"
        bind:value={path}
        aria-label="Sub-path inside docs"
        placeholder="optional, inside docs/"
      />
    </label>

    <label class="flex flex-col gap-1">
      <span class={label}>Markdown</span>
      <textarea
        class="min-h-24 w-full font-mono"
        bind:value={content}
        aria-label="Document markdown"></textarea>
    </label>
  {:else}
    <label class="flex flex-col gap-1">
      <span class={label}>Status</span>
      <input
        class="w-full"
        bind:value={status}
        aria-label="Decision status"
        placeholder="proposed"
      />
      <span class="text-body text-chalk-faint">
        Free-form. Backlog.md writes a skeleton with Context, Decision and
        Consequences headings, and has no command that fills them in — so the
        decision itself is written in the file.
      </span>
    </label>
  {/if}

  <div class="flex items-baseline gap-3">
    <button
      type="button"
      class="border-b border-chalk-faint text-body text-chalk hover:border-chalk"
      disabled={busy || !title.trim() || !project}
      onclick={create}
    >
      {busy ? "writing…" : `create the ${kind}`}
    </button>
    <button
      type="button"
      class="ml-auto font-mono text-data text-chalk-faint hover:text-chalk"
      onclick={() => onDone(null)}
    >
      cancel
    </button>
  </div>
</section>
