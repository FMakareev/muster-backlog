import { mount, unmount } from "svelte";
import { beforeEach, describe, expect, it, vi } from "vitest";
import type {
  ProjectEdit,
  ProjectView,
} from "../../bindings/github.com/FMakareev/muster-backlog/internal/app/models";

/**
 * The roll is where a person stands when they want a project off the board,
 * so this asks the rendered thing rather than the code behind it: is the
 * control there, does it say which project it acts on, and is there a way
 * back once the project has gone.
 *
 * Hiding shipped two days before it was asked for again by the person who
 * had it. The feature was never the problem; being unable to reach it was.
 */

const saved: { path: string; edit: ProjectEdit }[] = [];

vi.mock(
  "../../bindings/github.com/FMakareev/muster-backlog/internal/app",
  () => ({
    BoardService: {
      SaveProject: (path: string, edit: ProjectEdit) => {
        saved.push({ path, edit });
        return Promise.resolve({ ok: true, problem: null });
      },
      // refresh() runs after every write. It is answered with nothing, which
      // is what makes the store keep whatever the test seeded.
      Projects: () => Promise.resolve(null),
      Tasks: () => Promise.resolve(null),
      Problems: () => Promise.resolve(null),
      Layout: () => Promise.resolve({ columns: null, conflicts: null }),
      Milestones: () => Promise.resolve(null),
      Drafts: () => Promise.resolve(null),
    },
  }),
);

const { projects } = await import("./board");
const { focusedProject } = await import("./ui");
const ProjectRoll = (await import("./ProjectRoll.svelte")).default;

function project(name: string, hidden = false): ProjectView {
  return {
    path: `/dev/${name}`,
    displayPath: `~/dev/${name}`,
    name,
    colour: "",
    ok: true,
    problem: "",
    taskCount: 3,
    draftCount: 0,
    statuses: ["To Do"],
    priorities: null,
    types: null,
    layout: "standard",
    hidden,
  };
}

function render(): { root: HTMLElement; stop: () => void } {
  const root = document.createElement("div");
  document.body.appendChild(root);
  const component = mount(ProjectRoll, { target: root });
  return {
    root,
    stop: () => {
      unmount(component);
      root.remove();
    },
  };
}

function byLabel(root: HTMLElement, label: string): HTMLButtonElement | null {
  return root.querySelector(`button[aria-label="${label}"]`);
}

beforeEach(() => {
  saved.length = 0;
  projects.set([]);
  focusedProject.set("");
});

describe("the project roll", () => {
  it("offers to hide each project it lists, by name", () => {
    projects.set([project("Muster"), project("Treeline")]);
    const { root, stop } = render();

    for (const name of ["Muster", "Treeline"]) {
      const button = byLabel(root, `Hide ${name} from the board`);
      expect(button, `no hide control for ${name}`).not.toBeNull();
      // Reachable without a mouse: a control only a pointer can find is the
      // same as no control for anyone driving this from the keyboard.
      expect(button?.disabled).toBe(false);
      expect(button?.tagName).toBe("BUTTON");
    }
    stop();
  });

  it("hides the project the control names, and nothing else", async () => {
    projects.set([project("Muster"), project("Treeline")]);
    const { root, stop } = render();

    byLabel(root, "Hide Treeline from the board")?.click();
    await Promise.resolve();

    expect(saved).toHaveLength(1);
    expect(saved[0].path).toBe("/dev/Treeline");
    expect(saved[0].edit.hidden).toBe(true);
    // The rest of the entry goes back as it stood: the registry rewrites the
    // whole thing, so a field this control forgets is a field it erases.
    expect(saved[0].edit.name).toBe("Treeline");
    stop();
  });

  it("lets go of a project it is focused on before hiding it", async () => {
    projects.set([project("Muster")]);
    focusedProject.set("/dev/Muster");
    const { root, stop } = render();

    byLabel(root, "Hide Muster from the board")?.click();
    await Promise.resolve();

    // Otherwise the board narrows to a project that is no longer on it, and
    // shows nothing at all without saying why.
    expect(focusedProject.get()).toBe("");
    stop();
  });

  it("counts what has been put away and offers it back", async () => {
    projects.set([project("Muster"), project("Treeline", true)]);
    const { root, stop } = render();

    expect(root.textContent).toContain("Muster");
    expect(root.textContent).not.toContain("Treeline");
    expect(root.textContent).toContain("hidden");

    const opener = root.querySelector<HTMLButtonElement>(
      "button[aria-expanded]",
    );
    expect(opener, "the roll never says anything is hidden").not.toBeNull();
    opener?.click();
    await Promise.resolve();

    const back = byLabel(root, "Show Treeline on the board");
    expect(back, "no way back from the roll").not.toBeNull();
    back?.click();
    await Promise.resolve();

    expect(saved).toHaveLength(1);
    expect(saved[0].path).toBe("/dev/Treeline");
    expect(saved[0].edit.hidden).toBe(false);
    stop();
  });

  it("says nothing about hidden projects when there are none", () => {
    projects.set([project("Muster")]);
    const { root, stop } = render();
    expect(root.querySelector("button[aria-expanded]")).toBeNull();
    stop();
  });
});
