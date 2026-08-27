import { chromium } from "playwright";

/**
 * The screenshots for the README, taken from the running application against
 * the invented corpus in build.sh.
 *
 * Two rules. Nothing here touches a real project — the server is started with
 * HOME and XDG_CONFIG_HOME pointed at the demo corpus, so the paths shown in
 * the status bar and on the Projects screen are the demo's and no real
 * registry is reachable. And each shot is of a screen doing the thing it
 * exists for, rather than an empty one that happens to be tidy.
 *
 * To retake them, from the repository root:
 *
 *     bash .github/media/demo-corpus.sh /tmp/muster-demo
 *     go build -tags "server,production,gtk3" -o /tmp/muster-srv .
 *     HOME=/tmp/muster-demo XDG_CONFIG_HOME=/tmp/muster-demo/.config /tmp/muster-srv &
 *     node .github/media/demo-shots.mjs .github/media
 *
 * It needs Playwright, which this project does not depend on; see
 * build/icon/render.mjs for the same note.
 */

const out = process.argv[2];
const shots = [];

const browser = await chromium.launch();
const page = await browser.newPage({
  viewport: { width: 1440, height: 900 },
  deviceScaleFactor: 1,
});
await page.goto("http://localhost:8080", { waitUntil: "networkidle" });
await page.waitForTimeout(3000);

async function shoot(name, note) {
  await page.waitForTimeout(700);
  await page.screenshot({ path: `${out}/${name}.png` });
  shots.push(`${name}  ${note}`);
}

// The board: five projects in one set of columns.
await shoot("board", "the board over five projects");

// Grouped by project, which is what makes a column readable across many.
await page.keyboard.press("g");
await page.waitForTimeout(900);
await shoot("board-grouped", "the same board, cards clumped by project");
await page.keyboard.press("g");
await page.keyboard.press("g");
await page.waitForTimeout(600);

// The list, with a task panel open would hide the table; leave it clean.
await page.keyboard.press("l");
await shoot("list", "the sortable list across every project");

// The inbox.
await page.keyboard.press("i");
await shoot("inbox", "drafts waiting, oldest first");

// The figures. A summary screen is wide and short; a 900px window would be
// two thirds empty and read as an unfinished screen rather than a summary.
await page.keyboard.press("s");
await page.setViewportSize({ width: 1440, height: 560 });
await shoot("stats", "the figures across every project");
await page.setViewportSize({ width: 1440, height: 900 });

// The documents viewer.
await page.keyboard.press("d");
await page.waitForTimeout(900);
await shoot("docs", "documents and decisions");

// The projects screen.
await page.keyboard.press("p");
await shoot("projects", "the registry");

// Search, over everything at once.
await page.keyboard.press("b");
await page.waitForTimeout(500);
await page.keyboard.press("/");
await page.waitForTimeout(400);
await page.keyboard.type("cache");
await page.waitForTimeout(900);
await shoot("search", "one search over every project");
await page.keyboard.press("Escape");

// A task open in the panel, which is where most of the editing lives.
await page.waitForTimeout(400);
const card = page.locator("[data-id]").first();
await card.click();
await page.waitForTimeout(1200);
await shoot("task", "a task in the panel");

await browser.close();
console.log(shots.join("\n"));
