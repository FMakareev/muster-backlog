<script lang="ts">
  import { tick } from "svelte";
  import { Kanban, WillowDark } from "@svar-ui/svelte-kanban";
  import type { KanbanCard, KanbanInstanceApi } from "@svar-ui/svelte-kanban";
  import { BoardService } from "../../bindings/github.com/FMakareev/muster-backlog/internal/app";
  import type { TaskView } from "../../bindings/github.com/FMakareev/muster-backlog/internal/app/models";
  import {
    canMove,
    columns,
    milestoneLabel,
    milestones,
    refresh,
  } from "./board";
  import { notify } from "./notices";
  import { groupBy, openNewTask, openTask, visibleTasks } from "./ui";
  import Card from "./Card.svelte";

  /**
   * The multi-project board.
   *
   * Columns are the union of what every project declares; a card can only move
   * within its own project's list, and the move is written by the Backlog.md
   * CLI. The board settles on what the files say afterwards, never on where the
   * card was dropped.
   */

  const boardColumns = $derived(
    $columns.map((column) => ({ id: column.name, label: column.name })),
  );

  // A card's identity is its project, kind, class and id together: ids collide
  // across projects, and one id can name two tasks inside a project.
  const boardCards = $derived(
    $visibleTasks.map((task) => ({
      id: `${task.project}|${task.kind}|${task.class}|${task.id}`,
      label: task.entity.Title,
      column: task.entity.Status,
      task,
    })),
  );

  const byID = $derived(new Map(boardCards.map((c) => [c.id, c.task])));

  // Grouping is expressed as an ordering, because the open edition of SVAR has
  // no swimlanes. Within a project the board keeps the order the store gave it,
  // which is Backlog.md's own comparator.
  function groupKey(card: KanbanCard): string {
    const task = (card as { task?: TaskView }).task;
    if (!task) return "";
    if ($groupBy === "project") return task.projectName;
    if ($groupBy === "milestone") {
      const value = task.entity.Milestone;
      // Tasks with no milestone sort last rather than first, so an unplanned
      // pile does not push the planned work down the column.
      return value
        ? milestoneLabel(task.project, value, $milestones)
        : "\uffff";
    }
    return "";
  }

  const sort = $derived(
    $groupBy
      ? (a: KanbanCard, b: KanbanCard) => groupKey(a).localeCompare(groupKey(b))
      : null,
  );

  /** Open the task a DOM node belongs to.
   *
   * SVAR prefixes the identity it writes into data-id with a colon, so the
   * card's own id has to be recovered from it rather than read straight off. */
  function openFrom(target: EventTarget | null): void {
    const card = (target as Element | null)?.closest("[data-id]");
    const raw = card?.getAttribute("data-id");
    if (!raw) return;
    const task = byID.get(raw.replace(/^:/, ""));
    if (!task) return;
    openTask({
      project: task.project,
      kind: task.kind,
      class: task.class,
      id: task.id,
    });
  }

  function onCardClick(event: MouseEvent): void {
    openFrom(event.target);
  }

  /** Put focus back on a card after the board has been re-rendered. */
  async function restoreFocus(dataID: string | null): Promise<void> {
    if (!dataID) return;
    await tick();
    const selector = `.wx-card[data-id="${CSS.escape(dataID)}"]`;
    for (let attempt = 0; attempt < 20; attempt++) {
      const card = document.querySelector<HTMLElement>(selector);
      if (card) {
        card.focus();
        return;
      }
      await new Promise((resolve) => setTimeout(resolve, 25));
    }
  }

  // Cards are focusable, so a keyboard user reaches them by Tab and opens the
  // panel the same way they would open anything else.
  function onCardKeydown(event: KeyboardEvent): void {
    const card = (event.target as Element | null)?.closest(".wx-card");
    if (!card) return;

    if (event.key === "Enter" || event.key === " ") {
      event.preventDefault();
      openFrom(event.target);
      return;
    }

    // Square brackets move a focused card to the previous or next status its
    // own project declares, so a card can be moved without a mouse at all.
    if (event.key !== "[" && event.key !== "]") return;
    const raw = card.getAttribute("data-id");
    const task = raw ? byID.get(raw.replace(/^:/, "")) : undefined;
    if (!task) return;
    event.preventDefault();

    const allowed = $columns
      .map((c) => c.name)
      .filter((name) => canMove(task.project, name));
    const at = allowed.indexOf(task.entity.Status);
    const next = allowed[at + (event.key === "]" ? 1 : -1)];
    if (at < 0 || !next) {
      notify(
        `${task.id} is already at the ${event.key === "]" ? "last" : "first"} ` +
          `status ${task.projectName} declares.`,
      );
      return;
    }

    void BoardService.SetStatus(task.project, task.id, next).then(
      async (result) => {
        if (!result.ok && result.problem) {
          notify(`${result.problem.title}: ${result.problem.detail}`);
        }
        await refresh();
        // The refresh re-renders the card, which drops focus. Without putting
        // it back, a second keypress goes nowhere and the board stops being
        // keyboard-driven after exactly one move.
        void restoreFocus(raw);
      },
    );
  }

  function init(api: KanbanInstanceApi): void {
    /**
     * The plus on a column head, and a double-click on its empty space.
     *
     * Both ask for a task in that column, and both are answered by opening
     * the form on that status rather than by the board's own inline editor.
     * The card is still refused: the board would add it in memory, showing a
     * task the files on disk do not have, which is the one thing this
     * application must never do. The task appears when the CLI has written it.
     */
    api.intercept("add-card", (raw) => {
      const event = raw as { card?: { column?: string } };
      openNewTask(event.card?.column ?? "");
      return false;
    });

    // Editing and deleting a card in place stay refused, and the panel is
    // where both of those actually live.
    for (const action of ["update-card", "delete-card"] as const) {
      api.intercept(action, () => {
        notify(
          "Open a task to edit it. A task is filed or archived from its " +
            "panel rather than deleted from the board.",
        );
        return false;
      });
    }

    // Refuse a move a project cannot represent, and say why. Writing a status
    // the project does not declare would produce a task the CLI itself rejects,
    // and the alternative - editing that project's config - is exactly what
    // Muster does not do.
    api.intercept("move-card", (raw) => {
      // The intercept signature covers every store action, so narrow to what
      // a move actually carries rather than asserting a shape.
      const event = raw as {
        id?: string | number;
        column?: string;
        before?: string | number;
      };
      const task =
        event.id === undefined ? undefined : byID.get(String(event.id));
      const target = event.column;
      if (!task || !target) return true;

      // The card it was dropped in front of, or nothing for the end of the
      // column. That is all manual order needs: the ordinal is worked out from
      // the neighbours on the other side.
      const before =
        event.before === undefined
          ? ""
          : (byID.get(String(event.before))?.id ?? "");

      if (task.entity.Status === target) {
        void BoardService.Reorder(task.project, task.id, before).then(
          (result) => {
            if (!result.ok && result.problem) {
              notify(`${result.problem.title}: ${result.problem.detail}`);
            }
            void refresh();
          },
        );
        return true;
      }

      if (!canMove(task.project, target)) {
        notify(
          `${task.projectName} has no “${target}” status, so ${task.id} cannot move there. ` +
            `Add it to that project's config.yml if it belongs there.`,
        );
        return false;
      }

      // Write through the CLI, then take whatever the files say afterwards.
      // A card dropped at a particular place in another column is two changes,
      // and both are made: the status, then the position, or it would land
      // wherever its old ordinal happened to put it.
      void BoardService.SetStatus(task.project, task.id, target)
        .then(async (result) => {
          if (!result.ok) {
            if (result.problem) {
              notify(`${result.problem.title}: ${result.problem.detail}`);
            }
            return;
          }
          const moved = await BoardService.Reorder(
            task.project,
            task.id,
            before,
          );
          if (!moved.ok && moved.problem) {
            notify(`${moved.problem.title}: ${moved.problem.detail}`);
          }
        })
        .finally(() => void refresh());
      return true;
    });
  }
</script>

<!-- svelte-ignore a11y_no_static_element_interactions -->
<div
  class="min-h-0 min-w-0 flex-1"
  onclick={onCardClick}
  onkeydown={onCardKeydown}
>
  <WillowDark fonts={false}>
    <Kanban
      cards={boardCards}
      columns={boardColumns}
      cardContent={Card}
      {sort}
      {init}
      render={{ virtualizeCards: true, columnScroll: true }}
    />
  </WillowDark>
</div>
