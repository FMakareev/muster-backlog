/**
 * A project's colour.
 *
 * Hue in this interface always means "which project" — the chrome itself is
 * achromatic so that nine project colours can sit on it without competing.
 * A project may pin its colour in the registry; when it does not, one is
 * derived from its path so that it is stable across restarts and unaffected by
 * adding or removing other projects.
 *
 * The palette is a fixed set tuned for the dark ground rather than an
 * arbitrary hue, so no project ever lands on something unreadable.
 */
const palette = [
  "#7aa2f7", // blue
  "#9ece6a", // green
  "#e0af68", // amber
  "#bb9af7", // violet
  "#7dcfff", // cyan
  "#f7768e", // rose
  "#c0caf5", // pale
  "#b8bb26", // lime
  "#d3869b", // mauve
  "#83a598", // sage
  "#fabd2f", // gold
  "#8ec07c", // mint
];

export function projectColour(path: string, pinned: string): string {
  if (pinned) return pinned;

  // FNV-1a over the path: stable, cheap, and well spread for short strings.
  let hash = 2166136261;
  for (let i = 0; i < path.length; i++) {
    hash ^= path.charCodeAt(i);
    hash = Math.imul(hash, 16777619);
  }
  return palette[Math.abs(hash) % palette.length];
}
