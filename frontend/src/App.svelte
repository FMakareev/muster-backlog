<script lang="ts">
  import { onMount } from "svelte";
  import { connect } from "./lib/board";
  import { screen } from "./lib/ui";
  import Shell from "./lib/Shell.svelte";
  import Board from "./lib/Board.svelte";
  import Docs from "./lib/Docs.svelte";
  import Stats from "./lib/Stats.svelte";
  import TaskList from "./lib/TaskList.svelte";
  import TaskPanel from "./lib/TaskPanel.svelte";

  onMount(() => connect());
</script>

<Shell>
  <div class="flex min-h-0 flex-1">
    {#if $screen === "list"}
      <TaskList />
    {:else if $screen === "docs"}
      <Docs />
    {:else if $screen === "stats"}
      <Stats />
    {:else}
      <Board />
    {/if}

    <!-- The panel rides alongside whichever screen is showing, because a task
         opened from a search result or a list row is the same task. -->
    <TaskPanel />
  </div>
</Shell>
