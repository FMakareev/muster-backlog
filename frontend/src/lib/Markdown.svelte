<script lang="ts">
  import {
    hasMermaid,
    highlightCode,
    renderMarkdown,
    renderMermaid,
  } from "./markdown";

  interface Props {
    /** The markdown to render. */
    source: string;
    /** The project this body belongs to; task references resolve inside it. */
    project?: string;
    /** Called when a task reference is clicked. */
    onTaskRef?: (id: string, project: string) => void;
  }
  let { source, project = "", onTaskRef }: Props = $props();

  let host: HTMLDivElement | undefined = $state();

  const html = $derived(renderMarkdown(source, { project }));

  // Highlighting and diagrams happen after the HTML is in the DOM, because
  // both work on real elements rather than on a string.
  $effect(() => {
    if (!host) return;
    // Re-run whenever the rendered html changes.
    void html;

    highlightCode(host);
    if (hasMermaid(host)) {
      void renderMermaid(host, true);
    }
  });

  function onClick(event: MouseEvent): void {
    const link = (event.target as Element | null)?.closest("[data-task-ref]");
    if (!link) return;
    event.preventDefault();
    onTaskRef?.(
      link.getAttribute("data-task-ref") ?? "",
      link.getAttribute("data-project") ?? project,
    );
  }
</script>

<!-- svelte-ignore a11y_click_events_have_key_events -->
<!-- svelte-ignore a11y_no_static_element_interactions -->
<div bind:this={host} class="prose-muster" onclick={onClick}>
  <!-- eslint-disable-next-line svelte/no-at-html-tags -->
  {@html html}
</div>
