import { marked } from "marked";
import DOMPurify from "dompurify";
import hljs from "highlight.js/lib/common";

/**
 * One markdown pipeline for task bodies, documents and decisions.
 *
 * These files are written by other tools and by agents, not by this
 * application, so the content is untrusted in the ordinary sense: not
 * malicious by assumption, but not ours to vouch for either. Everything is
 * sanitised, and nothing rendered here may reach the network — a local-first
 * tool must not let a task description phone home.
 */

/** A task reference the application can navigate to. */
export interface TaskLink {
  /** The raw id as written, for example TASK-42. */
  id: string;
}

/** Section markers Backlog.md writes into task bodies.
 *
 * They are stripped before parsing rather than sanitised afterwards: they are
 * HTML comments, and leaving them to the sanitiser means depending on its
 * configuration for whether they vanish or leak into the output. */
const SECTION_MARKER =
  /<!--\s*(?:SECTION:[A-Z_]+|AC|DOD|COMMENTS?):(?:BEGIN|END)\s*-->/g;

/** A Backlog.md id: a prefix, a number, and optionally a dotted subtask part.
 *
 * The prefix is configurable per project, so this matches any uppercase word
 * rather than assuming TASK. Word boundaries keep it from matching inside a
 * longer token. */
const TASK_REFERENCE = /\b([A-Z][A-Z0-9]*-\d+(?:\.\d+)*)\b/g;

/** Schemes a link may use. Anything else — javascript:, data: in an href,
 * a remote image — is removed. */
const SAFE_LINK_SCHEMES = /^(https?:|mailto:|muster:|#|\/|\.\/|\.\.\/)/i;

let configured = false;

function configure(): void {
  if (configured) return;
  configured = true;

  marked.setOptions({
    gfm: true,
    breaks: false,
  });

  DOMPurify.addHook("afterSanitizeAttributes", (node) => {
    const el = node as Element;

    // Acceptance criteria render as checkboxes, so inputs cannot simply be
    // forbidden. Only a checkbox survives, and always disabled: the state
    // shown is the state on disk, and changing it has to go through the CLI.
    if (el.tagName === "INPUT") {
      if (el.getAttribute("type") !== "checkbox") {
        el.remove();
        return;
      }
      el.setAttribute("disabled", "disabled");
    }

    // Nothing may fetch. An image or media element pointing anywhere but at
    // this application would turn opening a task into a network request.
    if (el.hasAttribute("src")) {
      const src = el.getAttribute("src") ?? "";
      if (!/^(data:image\/|\/|\.\/|\.\.\/)/i.test(src)) {
        el.removeAttribute("src");
        el.setAttribute("data-blocked-src", src);
      }
    }

    if (el.hasAttribute("href")) {
      const href = el.getAttribute("href") ?? "";
      if (!SAFE_LINK_SCHEMES.test(href)) {
        el.removeAttribute("href");
      } else if (/^https?:/i.test(href)) {
        // External links open in the system browser, never in the window.
        el.setAttribute("target", "_blank");
        el.setAttribute("rel", "noreferrer noopener");
      }
    }
  });
}

/** Options for one render. */
export interface RenderOptions {
  /**
   * The project a body belongs to. Task references resolve inside it and
   * nowhere else: ids collide freely across projects, so a reference that
   * searched every project would resolve to the wrong task.
   */
  project?: string;
}

/**
 * Render markdown to sanitised HTML.
 *
 * Returns a string rather than mounting anything, so the same pipeline serves
 * the task panel, the documents viewer and anything else that needs it.
 */
export function renderMarkdown(
  source: string,
  options: RenderOptions = {},
): string {
  configure();

  const stripped = (source ?? "").replace(SECTION_MARKER, "");
  const parsed = marked.parse(stripped, { async: false });
  const withRefs = linkTaskReferences(parsed, options.project ?? "");

  return DOMPurify.sanitize(withRefs, {
    ADD_ATTR: ["target", "data-task-ref", "data-project", "data-blocked-src"],
    FORBID_TAGS: ["style", "form", "iframe", "object", "embed"],
    FORBID_ATTR: ["style"],
  });
}

/**
 * Turn task ids in the rendered HTML into links the application handles.
 *
 * This runs on the HTML rather than the markdown so that ids inside code
 * spans and fenced blocks are left alone — a reference in prose is a link, a
 * reference in a code sample is a code sample.
 */
function linkTaskReferences(html: string, project: string): string {
  const skip = /<(code|pre|a)\b[^>]*>[\s\S]*?<\/\1>|<[^>]+>/gi;

  let out = "";
  let last = 0;
  for (const match of html.matchAll(skip)) {
    out += linkInText(html.slice(last, match.index), project) + match[0];
    last = match.index + match[0].length;
  }
  return out + linkInText(html.slice(last), project);
}

function linkInText(text: string, project: string): string {
  return text.replace(TASK_REFERENCE, (id) => {
    const attrs = project ? ` data-project="${escapeAttribute(project)}"` : "";
    return `<a href="#" data-task-ref="${escapeAttribute(id)}"${attrs}>${id}</a>`;
  });
}

function escapeAttribute(value: string): string {
  return value
    .replace(/&/g, "&amp;")
    .replace(/"/g, "&quot;")
    .replace(/</g, "&lt;");
}

/** Highlight a code element in place. Called after the HTML is in the DOM. */
export function highlightCode(root: ParentNode): void {
  for (const block of root.querySelectorAll("pre code")) {
    // Mermaid blocks are diagrams, not code, and are rendered separately.
    if (block.className.includes("language-mermaid")) continue;
    hljs.highlightElement(block as HTMLElement);
  }
}

/** Whether a body contains a diagram, so mermaid is loaded only when needed. */
export function hasMermaid(root: ParentNode): boolean {
  return root.querySelector("pre code.language-mermaid") !== null;
}

/**
 * Render mermaid diagrams found in the output.
 *
 * mermaid is a large dependency and most task bodies have no diagram, so it is
 * imported on demand rather than carried by every window that opens.
 */
export async function renderMermaid(
  root: ParentNode,
  dark: boolean,
): Promise<void> {
  const blocks = [...root.querySelectorAll("pre code.language-mermaid")];
  if (blocks.length === 0) return;

  const { default: mermaid } = await import("mermaid");
  mermaid.initialize({
    startOnLoad: false,
    securityLevel: "strict",
    theme: dark ? "dark" : "default",
    fontFamily: "var(--font-sans)",
    // Labels must be SVG text, not HTML inside a foreignObject. The sanitiser
    // below allows an SVG profile only, so foreignObject labels are stripped
    // and every node comes out an empty box.
    htmlLabels: false,
    flowchart: { htmlLabels: false },
  });

  for (const [index, block] of blocks.entries()) {
    const source = block.textContent ?? "";
    const target = block.closest("pre") ?? block;
    try {
      const { svg } = await mermaid.render(
        `mermaid-${Date.now()}-${index}`,
        source,
      );
      const figure = document.createElement("figure");
      figure.className = "mermaid";
      figure.innerHTML = DOMPurify.sanitize(svg, {
        USE_PROFILES: { svg: true, svgFilters: true },
      });
      target.replaceWith(figure);
    } catch (error) {
      // A diagram that will not parse is shown as the code it is, with the
      // reason. Silently dropping it would leave a hole in the document.
      const note = document.createElement("p");
      note.className = "mermaid-error";
      note.textContent = `Diagram could not be drawn: ${String(error)}`;
      target.after(note);
    }
  }
}
