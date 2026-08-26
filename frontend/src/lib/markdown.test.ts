import { describe, expect, it } from "vitest";
import { renderMarkdown } from "./markdown";

/**
 * These bodies come from files the application did not write. The tests that
 * matter here are the hostile ones: what a task description must not be able
 * to do to the window it is opened in.
 */
describe("renderMarkdown", () => {
  it("renders ordinary markdown", () => {
    const html = renderMarkdown("# Title\n\nSome **bold** text.\n");
    expect(html).toContain("<h1>Title</h1>");
    expect(html).toContain("<strong>bold</strong>");
  });

  it("strips Backlog.md section markers rather than leaking them", () => {
    const body = [
      "## Description",
      "",
      "<!-- SECTION:DESCRIPTION:BEGIN -->",
      "The description.",
      "<!-- SECTION:DESCRIPTION:END -->",
      "",
      "## Acceptance Criteria",
      "<!-- AC:BEGIN -->",
      "- [x] #1 Done",
      "<!-- AC:END -->",
    ].join("\n");

    const html = renderMarkdown(body);
    expect(html).toContain("The description.");
    expect(html).not.toContain("SECTION:DESCRIPTION");
    expect(html).not.toContain("AC:BEGIN");
  });

  it("renders acceptance criteria as checkboxes reflecting the file", () => {
    const html = renderMarkdown("- [x] #1 Checked\n- [ ] #2 Unchecked\n");
    const checked = html.match(/<input[^>]*checked[^>]*>/g) ?? [];
    const boxes = html.match(/<input[^>]*type="checkbox"[^>]*>/g) ?? [];
    expect(boxes).toHaveLength(2);
    expect(checked).toHaveLength(1);
    // Nothing is editable until writes go through the CLI.
    expect(html).toContain("disabled");
  });

  it("links task references in prose", () => {
    const html = renderMarkdown("Blocked by TASK-42 until then.", {
      project: "/home/me/dev/alpha",
    });
    expect(html).toContain('data-task-ref="TASK-42"');
    expect(html).toContain('data-project="/home/me/dev/alpha"');
  });

  it("links ids with any prefix, since the prefix is per project", () => {
    const html = renderMarkdown("See STORY-7 and TASK-3.1.");
    expect(html).toContain('data-task-ref="STORY-7"');
    expect(html).toContain('data-task-ref="TASK-3.1"');
  });

  it("leaves references inside code alone", () => {
    const html = renderMarkdown("Run `backlog task view TASK-42` first.");
    const code = html.slice(html.indexOf("<code>"), html.indexOf("</code>"));
    expect(code).toContain("TASK-42");
    expect(code).not.toContain("data-task-ref");
  });

  it("leaves references inside fenced blocks alone", () => {
    const html = renderMarkdown("```sh\nbacklog task view TASK-42\n```\n");
    expect(html).not.toContain("data-task-ref");
  });

  it("removes script tags", () => {
    const html = renderMarkdown("Hello <script>alert(1)</script> world");
    expect(html).not.toContain("<script");
    expect(html).not.toContain("alert(1)");
  });

  it("removes event handler attributes", () => {
    const html = renderMarkdown('<img src="./x.png" onerror="alert(1)">');
    expect(html).not.toContain("onerror");
    expect(html).not.toContain("alert(1)");
  });

  it("strips javascript: links", () => {
    const html = renderMarkdown("[click](javascript:alert(1))");
    expect(html).not.toContain("javascript:");
  });

  it("blocks remote images so opening a task cannot make a request", () => {
    const html = renderMarkdown("![beacon](https://example.com/pixel.gif)");
    // The url survives only as an inert record of what was blocked; no
    // attribute the browser would fetch may carry it.
    expect(html).not.toMatch(/\ssrc="https:/);
    expect(html).toContain('data-blocked-src="https://example.com/pixel.gif"');
  });

  it("keeps local and inline images", () => {
    const local = renderMarkdown("![diagram](./diagram.png)");
    expect(local).toContain('src="./diagram.png"');

    const inline = renderMarkdown("![dot](data:image/png;base64,iVBORw0KGgo=)");
    expect(inline).toContain("data:image/png;base64");
  });

  it("resolves relative links inside a project", () => {
    const html = renderMarkdown(
      "[the spec](./docs/spec.md) and [up](../README.md)",
    );
    expect(html).toContain('href="./docs/spec.md"');
    expect(html).toContain('href="../README.md"');
  });

  it("sends external links to the system browser", () => {
    const html = renderMarkdown("[docs](https://example.com)");
    expect(html).toContain('target="_blank"');
    expect(html).toContain("noopener");
  });

  it("removes iframes and forms", () => {
    const html = renderMarkdown(
      '<iframe src="./x"></iframe><form><input></form>',
    );
    expect(html).not.toContain("<iframe");
    expect(html).not.toContain("<form");
  });

  it("keeps a mermaid block for the diagram renderer", () => {
    const html = renderMarkdown("```mermaid\ngraph TD;\nA-->B;\n```\n");
    expect(html).toContain("language-mermaid");
    expect(html).toContain("A--&gt;B");
  });

  it("handles an empty body", () => {
    expect(renderMarkdown("")).toBe("");
  });

  it("renders Cyrillic unchanged", () => {
    const html = renderMarkdown("Риг байка на 2D-физике: рама и колёса");
    expect(html).toContain("Риг байка на 2D-физике");
  });
});
