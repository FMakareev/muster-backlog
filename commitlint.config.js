/**
 * Commit messages are a build input, not a style preference: release-please
 * derives the version bump and the CHANGELOG from them. The vocabulary below is
 * therefore closed - a type or scope outside it is rejected rather than
 * silently ending up in a release note nobody can categorise.
 */
export default {
  extends: ["@commitlint/config-conventional"],
  rules: {
    "type-enum": [
      2,
      "always",
      [
        "feat", // a user-visible capability
        "fix", // a user-visible defect repaired
        "perf", // a change made for speed or resource use
        "refactor", // behaviour preserved, structure changed
        "docs", // documentation only
        "test", // tests only
        "build", // build system, toolchain or dependencies
        "ci", // continuous integration configuration
        "chore", // repository housekeeping with no product effect
        "revert", // undoing a previous commit
      ],
    ],
    "scope-enum": [
      2,
      "always",
      [
        "app", // the Wails shell and Go entry point
        "parser", // reading Backlog.md markdown
        "store", // the aggregated in-memory task set
        "watcher", // filesystem watching
        "cli", // the backlog CLI write adapter
        "board", // the kanban board
        "list", // the task list view
        "task", // the task detail panel
        "inbox", // drafts and capture
        "docs-view", // the documents and decisions viewer
        "analytics", // the cross-project dashboard
        "search", // cross-project search
        "projects", // the registry and Projects screen
        "mcp", // the MCP server
        "ui", // shared frontend shell, theming and components
        "deps", // dependency updates
        "release", // versioning and release plumbing
      ],
    ],
    // A scope is helpful but not always meaningful, for instance on a
    // repository-wide chore.
    "scope-empty": [0],
    "body-max-line-length": [0],
  },
};
