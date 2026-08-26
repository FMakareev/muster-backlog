import { atom } from "nanostores";

/**
 * Transient messages: something the person just tried that could not happen,
 * explained where they tried it.
 *
 * Separate from problems, which are standing conditions. A notice answers "why
 * did that not work", and stops being true a moment later.
 */
export interface Notice {
  id: number;
  text: string;
}

export const notices = atom<Notice[]>([]);

let nextID = 1;

export function notify(text: string, ms = 6000): void {
  const notice = { id: nextID++, text };
  notices.set([...notices.get(), notice]);
  setTimeout(() => {
    notices.set(notices.get().filter((n) => n.id !== notice.id));
  }, ms);
}

export function dismiss(id: number): void {
  notices.set(notices.get().filter((n) => n.id !== id));
}
