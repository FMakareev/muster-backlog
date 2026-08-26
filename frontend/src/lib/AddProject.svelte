<script lang="ts">
  import { BoardService } from "../../bindings/github.com/FMakareev/muster-backlog/internal/app";
  import type {
    FolderView,
    InitFolder,
  } from "../../bindings/github.com/FMakareev/muster-backlog/internal/app/models";
  import { refresh } from "./board";
  import { notify } from "./notices";

  /**
   * Adding a folder.
   *
   * The folder is inspected first and what it is decides what is offered: one
   * that already holds a backlog is registered, one that does not is
   * initialised, and one that is not a git repository has that answered for it
   * rather than being asked a question `backlog init` would fail on.
   */

  interface Props {
    onDone: () => void;
  }
  let { onDone }: Props = $props();

  let path = $state("");
  let found = $state<FolderView | null>(null);
  let looking = $state(false);
  let busy = $state(false);
  let token = 0;

  let name = $state("");
  let colour = $state("");

  // Whether the desktop can show a directory chooser. False in the server
  // build, which has no dialogs, so the button is absent rather than broken.
  let canBrowse = $state(false);
  $effect(() => {
    void BoardService.CanChooseFolder().then((yes) => (canBrowse = yes));
  });

  // The init form. Everything empty means "let Backlog.md decide": the CLI is
  // run with --defaults and only what is filled in here is passed to it.
  let backlogDir = $state("");
  let configLocation = $state("");
  let taskPrefix = $state("");
  let zeroPadded = $state("");
  let integrationMode = $state("");
  let noGit = $state(false);
  let agents = $state<string[]>([]);

  const AGENTS = ["claude", "agents", "gemini", "copilot", "cursor"];

  // 1.48.0 refuses agent instructions when AI integration is off, so the form
  // does not offer the combination rather than letting it fail after the fact.
  const agentsAllowed = $derived(integrationMode !== "none");

  async function look(value: string): Promise<void> {
    const mine = ++token;
    if (!value.trim()) {
      found = null;
      return;
    }
    looking = true;
    const view = await BoardService.InspectFolder(value);
    if (mine !== token) return;
    looking = false;
    found = view;
    // The folder name is the obvious project name; only fill it while the
    // person has not typed one of their own.
    if (!name && view?.name) name = view.name;
  }

  function onPath(value: string): void {
    path = value;
    void look(value);
  }

  /**
   * Pick a folder with the desktop's own chooser.
   *
   * The chosen path goes through the same field and the same inspection as one
   * that was typed, so there is one path into the form rather than two.
   * Cancelling returns nothing and leaves the form exactly as it was.
   */
  async function browse(): Promise<void> {
    const chosen = await BoardService.ChooseFolder();
    if (!chosen) return;
    onPath(chosen);
  }

  async function add(): Promise<void> {
    if (!found || busy) return;
    busy = true;
    const result = await BoardService.AddProject(path, {
      name,
      colour,
      hidden: false,
    });
    busy = false;
    if (!result.ok) {
      notify(
        result.problem
          ? `${result.problem.title}: ${result.problem.detail}`
          : "The project could not be added.",
      );
      return;
    }
    await refresh();
    onDone();
  }

  async function initialise(): Promise<void> {
    if (!found || busy) return;
    busy = true;
    const form: InitFolder = {
      path,
      name,
      colour,
      backlogDir,
      configLocation,
      taskPrefix,
      zeroPaddedIds: Number(zeroPadded) || 0,
      // A folder that is not a repository takes this path whatever the form
      // says; the backend enforces the same thing.
      noGit: noGit || !found.isGit,
      agentInstructions: agentsAllowed ? agents.join(",") : "",
      integrationMode,
    };
    const result = await BoardService.InitProject(form);
    busy = false;
    if (!result.ok) {
      notify(
        result.problem
          ? `${result.problem.title}: ${result.problem.detail}`
          : "The folder could not be initialised.",
      );
      return;
    }
    await refresh();
    onDone();
  }

  function toggleAgent(value: string): void {
    agents = agents.includes(value)
      ? agents.filter((a) => a !== value)
      : [...agents, value];
  }

  const label =
    "text-micro font-medium uppercase tracking-[0.14em] text-chalk-faint";
  const note = "text-body text-chalk-faint";
  const usable = $derived(!!found && !found.problem && !found.registered);
</script>

<section
  class="flex flex-col gap-3 border-b border-rule bg-ink-sunken px-4 py-3"
  aria-label="Add a folder"
>
  <div class="flex flex-col gap-1">
    <span class={label}>Folder</span>
    <div class="flex items-center gap-2">
      <!-- svelte-ignore a11y_autofocus -->
      <input
        class="min-w-0 flex-1 font-mono"
        autofocus
        placeholder="/path/to/a/project, or ~/Dev/thing"
        value={path}
        aria-label="Folder to add"
        oninput={(e) => onPath(e.currentTarget.value)}
      />
      {#if canBrowse}
        <button
          type="button"
          class="min-h-6 shrink-0 border-b border-chalk-faint font-mono text-data
                 text-chalk hover:border-chalk"
          onclick={browse}
        >
          browse
        </button>
      {/if}
    </div>
  </div>

  {#if looking && !found}
    <p class={note}>Looking…</p>
  {:else if found && path.trim()}
    <p class="text-body">
      {#if found.problem}
        <span class="text-chalk">{found.problem}</span>
      {:else if found.registered}
        <span class="text-chalk">Muster already holds this folder.</span>
      {:else if found.hasBacklog}
        <span class="text-chalk"
          >A Backlog.md project is already here, in {found.layout}.</span
        >
        <span class={note}>It will be registered as it is.</span>
      {:else}
        <span class="text-chalk">There is no backlog here yet.</span>
        <span class={note}>
          {found.isGit
            ? "This is a git repository, so Backlog.md will use git."
            : "This is not a git repository, so it will be initialised without git integration."}
        </span>
      {/if}
    </p>
  {/if}

  {#if usable}
    <div class="flex flex-wrap gap-3">
      <label class="flex flex-col gap-1">
        <span class={label}>Name</span>
        <input
          class="w-56"
          bind:value={name}
          aria-label="Project name"
          placeholder={found?.name ?? ""}
        />
      </label>
      <label class="flex flex-col gap-1">
        <span class={label}>Colour</span>
        <span class="flex items-center gap-2">
          <input
            class="w-28 font-mono"
            bind:value={colour}
            aria-label="Project colour"
            placeholder="#7aa2f7"
          />
          <span
            class="h-4 w-4 shrink-0 rounded-[2px] border border-rule"
            style="background-color: {colour || 'transparent'}"
          ></span>
        </span>
      </label>
    </div>
  {/if}

  {#if usable && found && !found.hasBacklog}
    <div class="flex flex-col gap-3 border-t border-rule pt-3">
      <p class={note}>
        Everything below is optional. Anything left alone is answered by
        Backlog.md's own default.
      </p>

      <div class="flex flex-wrap gap-3">
        <label class="flex flex-col gap-1">
          <span class={label}>Backlog directory</span>
          <input
            class="w-40 font-mono"
            bind:value={backlogDir}
            aria-label="Backlog directory"
            placeholder="backlog"
          />
          <span class={note}>Where tasks are kept. Default: backlog</span>
        </label>

        <label class="flex flex-col gap-1">
          <span class={label}>Config location</span>
          <select bind:value={configLocation} aria-label="Config location">
            <option value="">Default (folder)</option>
            <option value="folder">Inside the backlog directory</option>
            <option value="root">At the project root</option>
          </select>
          <span class={note}>Where config.yml goes.</span>
        </label>

        <label class="flex flex-col gap-1">
          <span class={label}>Task prefix</span>
          <input
            class="w-28 font-mono"
            bind:value={taskPrefix}
            aria-label="Task prefix"
            placeholder="task"
          />
          <span class={note}>Letters only. Default: task</span>
        </label>

        <label class="flex flex-col gap-1">
          <span class={label}>Zero-padded ids</span>
          <input
            class="w-20 font-mono"
            bind:value={zeroPadded}
            aria-label="Zero-padded ids"
            placeholder="0"
          />
          <span class={note}
            >Digits to pad to. 0 turns it off, the default.</span
          >
        </label>

        <label class="flex flex-col gap-1">
          <span class={label}>AI integration</span>
          <select bind:value={integrationMode} aria-label="AI integration mode">
            <option value="">Default (cli)</option>
            <option value="cli">CLI instructions</option>
            <option value="mcp">MCP server</option>
            <option value="none">None</option>
          </select>
          <span class={note}>How agents talk to Backlog.md.</span>
        </label>
      </div>

      <div class="flex flex-col gap-1">
        <span class={label}>Agent instruction files</span>
        <div class="flex flex-wrap gap-1.5">
          {#each AGENTS as agent (agent)}
            <button
              type="button"
              class="min-h-6 rounded-[2px] border px-2 py-0.5 text-micro
                     {agents.includes(agent)
                ? 'border-chalk bg-ink text-chalk'
                : 'border-rule text-chalk-dim hover:text-chalk'}
                     {agentsAllowed ? '' : 'opacity-40'}"
              aria-pressed={agents.includes(agent)}
              disabled={!agentsAllowed}
              onclick={() => toggleAgent(agent)}
            >
              {agent}
            </button>
          {/each}
        </div>
        <span class={note}>
          {agentsAllowed
            ? "Instruction files written into the project. None by default."
            : "Backlog.md refuses instruction files when AI integration is none."}
        </span>
      </div>

      <div class="flex flex-col gap-1">
        <span class={label}>Git</span>
        {#if found.isGit}
          <label class="flex items-center gap-2 text-body text-chalk">
            <input type="checkbox" bind:checked={noGit} />
            Initialise without git integration
          </label>
          <span class={note}>
            This is a repository, so Backlog.md will use git unless you say
            otherwise. Default: git is used.
          </span>
        {:else}
          <span class={note}>
            There is no repository here, so git integration is off. Backlog.md
            would refuse to initialise otherwise.
          </span>
        {/if}
      </div>
    </div>
  {/if}

  <div class="flex items-baseline gap-3">
    {#if usable && found?.hasBacklog}
      <button
        type="button"
        class="border-b border-chalk-faint text-body text-chalk hover:border-chalk"
        disabled={busy}
        onclick={add}
      >
        {busy ? "Adding…" : "Add this project"}
      </button>
    {:else if usable}
      <button
        type="button"
        class="border-b border-chalk-faint text-body text-chalk hover:border-chalk"
        disabled={busy}
        onclick={initialise}
      >
        {busy ? "Initialising…" : "Initialise and add"}
      </button>
    {/if}
    <button
      type="button"
      class="ml-auto font-mono text-data text-chalk-faint hover:text-chalk"
      onclick={onDone}
    >
      cancel
    </button>
  </div>
</section>
