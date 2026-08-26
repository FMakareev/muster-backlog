<script lang="ts">
  import { onMount } from "svelte";
  import { Kanban, WillowDark } from "@svar-ui/svelte-kanban";
  import { cards, columns } from "./lib/board";
  // The Go type is app.Service; aliased here so the call site reads clearly.
  import { Service as AppService } from "../bindings/github.com/FMakareev/muster-backlog/internal/app";

  // nanostores are Svelte-store compatible, so the $ prefix works directly and
  // no adapter package is needed.
  let version = $state("");

  // nanostores expose their values as readonly, while SVAR takes ownership of
  // the arrays it is given, so the board gets copies rather than the store's
  // own arrays.
  const boardCards = $derived([...$cards]);
  const boardColumns = $derived([...$columns]);

  onMount(async () => {
    version = await AppService.AppVersion();
  });
</script>

<div class="flex h-full flex-col">
  <header
    class="flex items-center gap-3 border-b border-shell-border bg-shell-raised px-4 py-2"
  >
    <h1 class="text-sm font-semibold tracking-wide">Muster</h1>
    <span class="text-xs text-shell-muted">
      placeholder board — no project is loaded yet
    </span>
    <span class="ml-auto font-mono text-xs text-shell-muted">{version}</span>
  </header>

  <main class="min-h-0 flex-1">
    <WillowDark>
      <Kanban cards={boardCards} columns={boardColumns} />
    </WillowDark>
  </main>
</div>
