<script lang="ts">
  import { onMount } from "svelte";
  import { connect } from "./lib/board";
  import Shell from "./lib/Shell.svelte";
  import Board from "./lib/Board.svelte";
  import TaskPanel from "./lib/TaskPanel.svelte";

  onMount(() => connect());
</script>

<Shell>
  <div class="flex min-h-0 flex-1">
    <Board />
    <TaskPanel />
  </div>
</Shell>
