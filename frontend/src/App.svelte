<script lang="ts">
  import { onMount } from "svelte";
  import { connect } from "./lib/board";
  import Shell from "./lib/Shell.svelte";
  import Board from "./lib/Board.svelte";

  onMount(() => connect());
</script>

<Shell>
  <Board />
</Shell>
