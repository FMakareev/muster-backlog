/**
 * Rasterise the application's icon into every size the build needs.
 *
 * `build/appicon.svg` is the mark. Everything else — the packaged PNG, the
 * one carried in the binary, the AppImage icon, the macOS disk image — comes
 * from it here, and the Windows `.ico` and macOS `.icns` come from those via
 * `wails3 generate icons`. Nothing is drawn twice, so the mark cannot differ
 * between platforms.
 *
 * Run from the repository root:
 *
 *     node build/icon/render.mjs
 *     wails3 task common:generate:icons
 *     wails3 generate icons -input build/darwin/dmg-file-icon.png \
 *       -macfilename build/darwin/dmg-file-icon.icns -windowsfilename /tmp/unused.ico
 *
 * A browser does the rasterising, so what ships is what a browser draws from
 * the vector rather than what an approximation of one makes of it. Playwright
 * is not a dependency of this project - the icon changes about never, and
 * putting a browser download into every contributor's install for it would be
 * a poor trade - so it is asked for by name and the script says so plainly
 * when it is missing:
 *
 *     npm install -g playwright && playwright install chromium
 */

import { readFileSync } from "fs";
import { dirname, resolve } from "path";
import { fileURLToPath } from "url";

let chromium;
try {
  ({ chromium } = await import("playwright"));
} catch {
  console.error(
    "This needs Playwright, which this project does not depend on:\n" +
      "  npm install -g playwright && playwright install chromium",
  );
  process.exit(1);
}

const repo = resolve(dirname(fileURLToPath(import.meta.url)), "../..");
const mark = readFileSync(`${repo}/build/appicon.svg`, "utf8");

/**
 * The macOS disk image background. Wails ships one with its own logo and the
 * word WAILS on it; this replaces rather than deletes it, because the darwin
 * Taskfile names the file and a build that cannot find it is a worse answer
 * than a plain one for a platform this release does not claim.
 */
const dmgBackground = `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 540 380">
  <rect width="540" height="380" fill="#14161c"/>
  <rect x="0.5" y="0.5" width="539" height="379" fill="none" stroke="#272c38"/>
</svg>`;

const targets = [
  { svg: mark, out: "build/appicon.png", w: 1024, h: 1024 },
  // The size the application itself carries: a tray draws this at about
  // twenty-two pixels and a window switcher at forty-eight, so embedding the
  // 1024px sheet would mean resampling it hard on every draw, for nothing.
  { svg: mark, out: "build/appicon-256.png", w: 256, h: 256 },
  { svg: mark, out: "build/linux/appimage/muster.png", w: 1024, h: 1024 },
  { svg: mark, out: "build/darwin/dmg-file-icon.png", w: 1254, h: 1254 },
  { svg: dmgBackground, out: "build/darwin/dmg-background.png", w: 540, h: 380 },
];

const browser = await chromium.launch();
for (const { svg, out, w, h } of targets) {
  const page = await browser.newPage({ viewport: { width: w, height: h } });
  // The viewBox carries the artwork's own coordinates; width and height are
  // set here so one vector serves every size.
  const sized = svg.replace(
    /<svg([^>]*?)(?:\s+width="[^"]*")?(?:\s+height="[^"]*")?>/,
    `<svg$1 width="${w}" height="${h}">`,
  );
  await page.setContent(
    `<style>html,body{margin:0;background:transparent}svg{display:block}</style>${sized}`,
  );
  await page.waitForTimeout(200);
  await page.screenshot({ path: `${repo}/${out}`, omitBackground: true });
  await page.close();
  console.log(`${out}  ${w}x${h}`);
}
await browser.close();
