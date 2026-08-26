<script lang="ts">
  import { BoardService } from "../../bindings/github.com/FMakareev/muster-backlog/internal/app";
  import { refresh } from "./board";
  import { notify } from "./notices";
  import Markdown from "./Markdown.svelte";

  interface Props {
    project: string;
    taskID: string;
    /** One of the sections the parser exposes: description, plan, notes. */
    section: string;
    heading: string;
    source: string;
  }
  let { project, taskID, section, heading, source }: Props = $props();

  /**
   * One body section, read or edited.
   *
   * Editing shows the markdown source and saves it verbatim: this is a file
   * somebody else's tool wrote and will read again, so reformatting it on the
   * way through would be a change nobody asked for.
   */

  let editing = $state(false);
  let draft = $state("");
  let busy = $state(false);
  let area: HTMLTextAreaElement | undefined = $state();

  // What the file said when editing started. If it changes underneath - an
  // agent writing to the same task - saving would silently discard that, so
  // the conflict is surfaced instead.
  let basis = $state("");
  const changedUnderneath = $derived(editing && source !== basis);

  function start(): void {
    draft = source;
    basis = source;
    editing = true;
    queueMicrotask(() => area?.focus());
  }

  function cancel(): void {
    editing = false;
    draft = "";
  }

  async function save(): Promise<void> {
    if (busy) return;
    busy = true;
    const result = await BoardService.SetSection(
      project,
      taskID,
      section,
      draft,
    );
    busy = false;

    if (!result.ok) {
      // The text stays on screen; losing what someone just wrote because a
      // command failed is the least forgivable thing an editor can do.
      notify(
        result.problem
          ? `${result.problem.title}: ${result.problem.detail}`
          : "The change could not be saved.",
      );
      return;
    }
    editing = false;
    await refresh();
  }

  function onKeydown(event: KeyboardEvent): void {
    event.stopPropagation();
    if (event.key === "Escape") {
      cancel();
    } else if (event.key === "Enter" && (event.metaKey || event.ctrlKey)) {
      event.preventDefault();
      void save();
    }
  }
</script>

<section>
  <div class="flex items-baseline gap-2">
    <h3
      class="text-micro font-medium tracking-[0.14em] text-chalk-faint uppercase"
    >
      {heading}
    </h3>
    {#if editing}
      <button
        type="button"
        class="ml-auto font-mono text-micro text-chalk-faint hover:text-chalk"
        onclick={cancel}
        disabled={busy}
      >
        cancel esc
      </button>
      <button
        type="button"
        class="font-mono text-micro text-chalk hover:underline"
        onclick={save}
        disabled={busy}
      >
        {busy ? "saving…" : "save ⌃↵"}
      </button>
    {:else}
      <button
        type="button"
        class="ml-auto font-mono text-micro text-chalk-faint hover:text-chalk"
        onclick={start}
      >
        edit
      </button>
    {/if}
  </div>

  {#if editing}
    {#if changedUnderneath}
      <p class="mt-1 text-body text-chalk">
        This section changed on disk while you were typing. Saving will replace
        what is there now.
      </p>
    {/if}
    <textarea
      bind:this={area}
      bind:value={draft}
      onkeydown={onKeydown}
      disabled={busy}
      class="mt-1 min-h-40 w-full font-mono"
      aria-label="{heading} source"></textarea>
  {:else if source.trim()}
    <div class="mt-1">
      <Markdown {source} {project} />
    </div>
  {:else}
    <p class="mt-1 text-body text-chalk-faint">Empty.</p>
  {/if}
</section>
