/**
 * Dismissing an overlay: Escape from anywhere, and a click on the backdrop.
 *
 * Every overlay used to listen for Escape on its own element, which works only
 * while focus happens to be inside it. Preferences never moved focus at all, so
 * its handler could not fire and the window could not be closed from the
 * keyboard; search became unclosable the moment a click landed anywhere else.
 * The key is listened for on the window instead, so dismissal does not depend
 * on where the caret went.
 *
 * Overlays are kept in a stack and only the topmost one closes, because Escape
 * over a preferences window opened above a task should close the preferences
 * and leave the task where it was. Listening in the capture phase and stopping
 * propagation is what keeps the layers underneath from acting on the same key.
 */

interface Entry {
  close: () => void;
}

const stack: Entry[] = [];
let listening = false;

function onWindowKeydown(event: KeyboardEvent): void {
  if (event.key !== "Escape" || stack.length === 0) return;
  event.preventDefault();
  event.stopPropagation();
  stack[stack.length - 1].close();
}

/** What can hold focus inside an overlay, in the order a person would tab. */
const FOCUSABLE = [
  "input:not([disabled])",
  "select:not([disabled])",
  "textarea:not([disabled])",
  "button:not([disabled])",
  "[href]",
  '[tabindex]:not([tabindex="-1"])',
].join(",");

/**
 * Svelte action for an overlay's outermost element.
 *
 * Its lifetime is the overlay's: the element exists only while the overlay is
 * open, so there is no open state to track separately. `close` is whatever
 * closes this overlay.
 *
 * Clicking the element itself dismisses; clicking anything inside it does not.
 * Put it on the backdrop to get click-away, or on the panel to get Escape
 * alone.
 */
export function dismissable(node: HTMLElement, close: () => void) {
  let current = close;
  const entry: Entry = { close: () => current() };

  stack.push(entry);
  if (!listening) {
    window.addEventListener("keydown", onWindowKeydown, true);
    listening = true;
  }

  function onPointerdown(event: PointerEvent): void {
    if (event.target === node) current();
  }
  node.addEventListener("pointerdown", onPointerdown);

  // Focus moves into the overlay unless it has already placed the caret
  // somewhere better itself, which the ones with a text field do.
  queueMicrotask(() => {
    if (!node.isConnected || node.contains(document.activeElement)) return;
    const target =
      node.querySelector<HTMLElement>(FOCUSABLE) ??
      node.querySelector<HTMLElement>('[role="dialog"]') ??
      node;
    target.focus();
  });

  return {
    update(next: () => void): void {
      current = next;
    },
    destroy(): void {
      node.removeEventListener("pointerdown", onPointerdown);
      const at = stack.indexOf(entry);
      if (at >= 0) stack.splice(at, 1);
    },
  };
}
