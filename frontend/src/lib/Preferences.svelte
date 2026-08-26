<script lang="ts">
  import { BoardService } from "../../bindings/github.com/FMakareev/muster-backlog/internal/app";
  import {
    TaskView as TaskViewMode,
    WindowBehaviour,
  } from "../../bindings/github.com/FMakareev/muster-backlog/internal/settings/models";
  import { settings } from "./board";
  import { notify } from "./notices";
  import { showSettings } from "./ui";

  /**
   * Preferences.
   *
   * Two choices, both of which are genuinely a matter of taste rather than
   * something to decide on someone's behalf: where the application goes when
   * the window is closed, and where a task is shown when it is opened.
   */

  let busy = $state(false);

  async function save(next: typeof $settings): Promise<void> {
    busy = true;
    settings.set(next);
    const problems = await BoardService.SaveSettings(next);
    busy = false;
    for (const problem of problems ?? []) {
      notify(`${problem.title}: ${problem.detail}`);
    }
  }

  function onKeydown(event: KeyboardEvent): void {
    if (event.key === "Escape") {
      event.stopPropagation();
      showSettings.set(false);
    }
  }

  const row = "flex flex-col gap-1";
  const label =
    "text-micro font-medium uppercase tracking-[0.14em] text-chalk-faint";
</script>

{#if $showSettings}
  <div
    class="absolute inset-0 z-20 flex items-start justify-center bg-ink/70 pt-16"
  >
    <!-- svelte-ignore a11y_no_noninteractive_element_to_interactive_role -->
    <section
      role="dialog"
      aria-label="Preferences"
      tabindex="-1"
      class="flex w-[32rem] flex-col gap-4 rounded-sm border border-rule bg-ink-sunken p-4"
      onkeydown={onKeydown}
    >
      <header class="flex items-baseline gap-3">
        <h2 class="text-title font-semibold">Preferences</h2>
        <button
          type="button"
          class="ml-auto font-mono text-data text-chalk-faint hover:text-chalk"
          onclick={() => showSettings.set(false)}
        >
          close esc
        </button>
      </header>

      <div class={row}>
        <span class={label}>When the window is closed</span>
        <select
          disabled={busy}
          value={$settings.onWindowClose}
          onchange={(e) =>
            save({
              ...$settings,
              onWindowClose: e.currentTarget.value as WindowBehaviour,
            })}
        >
          <option value={WindowBehaviour.BehaviourQuit}
            >Quit, like an ordinary window</option
          >
          <option value={WindowBehaviour.BehaviourTray}
            >Stay in the system tray</option
          >
        </select>
        <p class="text-body text-chalk-faint">
          On Linux the tray is provided by the desktop and some do not have one.
          Where there is none, the window behaves ordinarily and says so rather
          than disappearing with no way back.
        </p>
      </div>

      <div class={row}>
        <span class={label}>Where a task opens</span>
        <select
          disabled={busy}
          value={$settings.taskView}
          onchange={(e) =>
            save({
              ...$settings,
              taskView: e.currentTarget.value as TaskViewMode,
            })}
        >
          <option value={TaskViewMode.ViewPanel}
            >In a panel beside the board</option
          >
          <option value={TaskViewMode.ViewCentred}
            >Centred over the board</option
          >
        </select>
        <p class="text-body text-chalk-faint">
          On a wide screen a panel against the edge puts the text a long way
          from where the eye rests.
        </p>
      </div>
    </section>
  </div>
{/if}
