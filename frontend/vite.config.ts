import { writeFileSync } from "node:fs";
import { resolve } from "node:path";
import { defineConfig, type Plugin } from "vite";
import { svelte } from "@sveltejs/vite-plugin-svelte";
import tailwindcss from "@tailwindcss/vite";
import wails from "@wailsio/runtime/plugins/vite";

// dist holds one tracked file. main.go embeds all:frontend/dist and Go
// resolves that at compile time, so without something in the directory nothing
// that compiles the main package works on a clone with no build history —
// linting and testing included. The build empties dist before writing, which
// takes the placeholder with it: every build then leaves a deleted file in
// `git status`, and the first person to commit that deletion puts the
// repository back to a clone that cannot be built.
const keepThePlaceholder = (): Plugin => ({
  name: "muster:keep-the-dist-placeholder",
  apply: "build",
  closeBundle() {
    writeFileSync(resolve(import.meta.dirname, "dist/.gitkeep"), "");
  },
});

export default defineConfig({
  server: {
    host: "127.0.0.1",
    port: Number(process.env.WAILS_VITE_PORT) || 9245,
    strictPort: true,
  },
  plugins: [svelte(), tailwindcss(), wails("./bindings"), keepThePlaceholder()],
});
