import { svelte } from "@sveltejs/vite-plugin-svelte";
import { defineConfig } from "vitest/config";

export default defineConfig({
  // Components are tested by mounting them, so the same plugin that builds
  // them has to compile them here. Without the browser condition Svelte
  // resolves to its server build, which renders to a string and never
  // produces the DOM a keyboard control has to be found in.
  plugins: [svelte()],
  resolve: { conditions: ["browser"] },
  test: {
    environment: "jsdom",
    include: ["src/**/*.test.ts"],
  },
});
