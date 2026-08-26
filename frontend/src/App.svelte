<script lang="ts">
  import { onMount } from "svelte";
  import { Kanban, WillowDark } from "@svar-ui/svelte-kanban";
  import { columns, connect } from "./lib/board";
  import { visibleTasks } from "./lib/ui";
  import Shell from "./lib/Shell.svelte";

  onMount(() => connect());

  // SVAR takes ownership of the arrays it is handed, and nanostores expose
  // their values as readonly, so the board gets copies.
  const boardColumns = $derived(
    $columns.map((status) => ({ id: status, label: status })),
  );

  const boardCards = $derived(
    $visibleTasks.map((task, index) => ({
      id: index + 1,
      label: task.entity.Title,
      column: task.entity.Status,
    })),
  );
</script>

<Shell>
  <div class="min-h-0 flex-1">
    <WillowDark>
      <!-- Read-only until writes go through the backlog CLI (TASK-25). A
           draggable card here would move in memory only, showing a state the
           files on disk do not have. -->
      <Kanban cards={boardCards} columns={boardColumns} readonly />
    </WillowDark>
  </div>
</Shell>
