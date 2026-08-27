<script lang="ts">
  import { BoardService } from "../../bindings/github.com/FMakareev/muster-backlog/internal/app";
  import type {
    Plan,
    Result,
    Status,
  } from "../../bindings/github.com/FMakareev/muster-backlog/internal/agents/models";

  /**
   * Connecting an agent to Muster's MCP server.
   *
   * This replaces a paragraph of documentation with a button, and the whole
   * design is about being boring: it shows exactly what it will run or write,
   * in full, before doing either; it backs up every file it touches; and
   * disconnecting is the same button in reverse.
   *
   * The list arrives in two passes because the two questions cost differently.
   * Whether a client is installed is a stat call. Whether Muster is already in
   * its configuration means running it — `claude mcp list` takes seconds — so
   * the installed list is shown at once and the connected answers fill in when
   * they arrive. Saying "not connected" before asking would be a false
   * statement with a button attached to it.
   */

  let clients = $state<Status[]>([]);
  let loading = $state(true);
  let plan = $state<(Plan & { disconnect: boolean }) | null>(null);
  let result = $state<(Result & { id: string }) | null>(null);
  let busy = $state("");

  async function load(): Promise<void> {
    loading = true;
    clients = (await BoardService.AgentClients().catch(() => [])) ?? [];
    loading = false;
    const asked = await BoardService.AgentConnections().catch(() => null);
    if (asked) clients = asked;
  }

  $effect(() => {
    void load();
  });

  async function preview(id: string, disconnect: boolean): Promise<void> {
    result = null;
    busy = id;
    const next = await BoardService.AgentPlan(id, disconnect);
    busy = "";
    plan = { ...next, disconnect };
  }

  /** Nothing happens until this: what runs is the plan shown above. */
  async function apply(): Promise<void> {
    if (!plan) return;
    const { id, disconnect } = plan;
    busy = id;
    const res = await BoardService.AgentApply(id, disconnect).catch(
      (err: unknown) =>
        ({ ok: false, verified: false, error: String(err) }) as Result,
    );
    busy = "";
    plan = null;
    result = { ...res, id };
    await load();
  }

  const label =
    "text-micro font-medium uppercase tracking-[0.14em] text-chalk-faint";
  const quiet = "min-h-6 font-mono text-data text-chalk-faint hover:text-chalk";
</script>

<div class="flex flex-col gap-2">
  <span class={label}>Agent clients</span>
  <p class="text-body text-chalk-faint">
    Muster can answer an agent about every registered project at once. Connect
    one and it runs that client's own command, or writes that client's own
    configuration file — shown in full first, and undone the same way.
  </p>

  {#if loading}
    <p class="text-body text-chalk-faint">Looking…</p>
  {:else if clients.length === 0}
    <p class="text-body text-chalk-faint">No clients this knows about.</p>
  {/if}

  <ul class="flex flex-col">
    {#each clients as client (client.id)}
      <li class="flex flex-wrap items-baseline gap-x-3 gap-y-1 py-1">
        <span class="w-32 shrink-0 text-body text-chalk">{client.label}</span>

        {#if !client.installed}
          <span class="text-body text-chalk-faint">not installed here</span>
        {:else if client.asking}
          <span class="text-body text-chalk-faint">asking it…</span>
        {:else if client.connected}
          <span class="font-mono text-data text-chalk">connected</span>
        {:else}
          <span class="font-mono text-data text-chalk-faint">not connected</span
          >
        {/if}

        {#if client.detail}
          <span
            class="min-w-0 flex-1 truncate font-mono text-micro text-chalk-faint"
            title={client.path || client.detail}
          >
            {client.detail}
          </span>
        {/if}

        {#if client.installed && !client.asking}
          <span class="ml-auto flex shrink-0 items-baseline gap-3">
            <button
              type="button"
              class={quiet}
              disabled={busy === client.id}
              onclick={() => preview(client.id, client.connected)}
            >
              {client.connected ? "disconnect" : "connect"}
            </button>
          </span>
        {/if}
      </li>
    {/each}
  </ul>

  {#if plan}
    <section
      class="flex flex-col gap-2 border border-rule bg-ink p-3"
      aria-label="What this will do"
    >
      <span class={label}>
        {plan.disconnect ? "This will undo" : "This will do"}
      </span>

      {#if plan.error}
        <p class="text-body text-chalk">{plan.error}</p>
      {:else}
        {#if plan.command}
          <!-- svelte-ignore a11y_no_noninteractive_tabindex -->
          <!-- A box that scrolls has to be reachable by keyboard, which is
               what role="region" and a tab stop are for. Svelte's rule does
               not know about that; axe fails the page without it. -->
          <pre
            class="overflow-x-auto font-mono text-data text-chalk-dim"
            tabindex="0"
            role="region"
            aria-label="The command this will run">{plan.command}</pre>
        {/if}
        {#if plan.path}
          <p class="font-mono text-micro text-chalk-faint">{plan.path}</p>
        {/if}
        {#if plan.diff}
          <!-- svelte-ignore a11y_no_noninteractive_tabindex -->
          <pre
            class="max-h-64 overflow-auto font-mono text-data text-chalk-dim"
            tabindex="0"
            role="region"
            aria-label="The change this will make">{plan.diff}</pre>
        {/if}
        {#if plan.note}
          <p class="text-body text-chalk-faint">{plan.note}</p>
        {/if}
      {/if}

      <div class="flex items-baseline gap-3">
        {#if !plan.error}
          <button
            type="button"
            class="border-b border-chalk-faint text-body text-chalk hover:border-chalk"
            disabled={busy !== ""}
            onclick={apply}
          >
            {busy !== "" ? "working…" : plan.disconnect ? "undo it" : "do it"}
          </button>
        {/if}
        <button type="button" class={quiet} onclick={() => (plan = null)}>
          cancel
        </button>
      </div>
    </section>
  {/if}

  {#if result}
    <section
      class="flex flex-col gap-1 border border-rule bg-ink p-3"
      aria-label="What happened"
    >
      <span class={label}>
        {result.ok ? "Done" : "That did not work"}
      </span>
      {#if result.error}
        <p class="text-body text-chalk">{result.error}</p>
      {/if}
      {#if result.ok && !result.verified}
        <p class="text-body text-chalk-faint">
          It reported no error, but saying so is as far as it goes — this client
          offers no way to check afterwards.
        </p>
      {/if}
      {#if result.backup}
        <p class="font-mono text-micro text-chalk-faint">
          The file it had before is at {result.backup}
        </p>
      {/if}
      {#if result.output}
        <!-- svelte-ignore a11y_no_noninteractive_tabindex -->
        <pre
          class="max-h-40 overflow-auto font-mono text-data text-chalk-dim"
          tabindex="0"
          role="region"
          aria-label="What the client said">{result.output}</pre>
      {/if}
      {#if result.command}
        <p class="text-body text-chalk-faint">
          The same thing, to run yourself if you would rather:
        </p>
        <!-- svelte-ignore a11y_no_noninteractive_tabindex -->
        <pre
          class="overflow-x-auto font-mono text-data text-chalk-dim"
          tabindex="0"
          role="region"
          aria-label="The command, to run yourself">{result.command}</pre>
      {/if}
    </section>
  {/if}
</div>
