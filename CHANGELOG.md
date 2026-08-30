# Changelog

## [1.1.0](https://github.com/FMakareev/muster-backlog/compare/v1.0.0...v1.1.0) (2026-08-30)


### Added

* **projects:** offer hiding where the projects are listed ([d0e31bd](https://github.com/FMakareev/muster-backlog/commit/d0e31bd8075f498eb65f698cdee4e063d92ec41b))


### Fixed

* **projects:** stop a write from freezing the folder name ([d22ec4e](https://github.com/FMakareev/muster-backlog/commit/d22ec4e02b3ea60e15ecd0f658b2ecd4e9bf5e01))
* **release:** publish the release outright, not as a draft ([f44add0](https://github.com/FMakareev/muster-backlog/commit/f44add093af1abcf629c6751c8584eeac8aaf263))


### Documentation

* state the promise a major version now makes ([d55903e](https://github.com/FMakareev/muster-backlog/commit/d55903eed9cb6ec2238fe318e9f30f30bb4adde2))

## 1.0.0 (2026-08-28)


### Added

* **app:** bridge the store to the frontend ([513992c](https://github.com/FMakareev/muster-backlog/commit/513992caa697182fd827ab6688e0841f64ebc4ca))
* **board:** edit tasks from the panel and move them by keyboard ([8fe32a1](https://github.com/FMakareev/muster-backlog/commit/8fe32a1b391ffbded7cc30571ec665ca4e2fafe7))
* **board:** group cards by project and finish the board ([acf76cc](https://github.com/FMakareev/muster-backlog/commit/acf76cc56e9b7c24a37db83fceaf2e2937b970f3))
* **board:** show subtasks without nesting the board ([48da8f9](https://github.com/FMakareev/muster-backlog/commit/48da8f9af77e82683a458f37cf045bfbf0baa55e))
* **board:** unify columns across projects and move cards through the CLI ([a71af9f](https://github.com/FMakareev/muster-backlog/commit/a71af9ffb61c7dfb40ba4f1aee8ee5fea9d58a8d))
* **cli:** add the backlog CLI write adapter ([a707fd2](https://github.com/FMakareev/muster-backlog/commit/a707fd21d92f9b394515b143cb1b509e869ad75d))
* **docs-view:** write documents and decisions, not only read them ([b5a70d7](https://github.com/FMakareev/muster-backlog/commit/b5a70d7ae41de6ce4990c9335aba1826b37b762c))
* **inbox:** build the drafts triage view ([f0b8d49](https://github.com/FMakareev/muster-backlog/commit/f0b8d4946bfa067c12ea3ae44d839d19e00d2e6e))
* **inbox:** make a note as workable as a task ([956df4f](https://github.com/FMakareev/muster-backlog/commit/956df4f7523450bc5357e8b6178fbc93cd68319d))
* **inbox:** make capture cost less ([372c68a](https://github.com/FMakareev/muster-backlog/commit/372c68af7a177a70407245d4c8136dbc1ef174eb))
* **list:** change many tasks at once ([5827f30](https://github.com/FMakareev/muster-backlog/commit/5827f30f9a1f504f0f53fcf1e29bd3edcf4c37c5))
* **list:** group subtasks under their parent in the table ([81583e8](https://github.com/FMakareev/muster-backlog/commit/81583e830e141338e19135301e8821b69dd417cf))
* **mcp:** connect an agent client in one click ([02df3b5](https://github.com/FMakareev/muster-backlog/commit/02df3b555f650fd7c019bac31337ffb84c8fbb0e))
* **mcp:** say where this server's boundary is ([32c54ed](https://github.com/FMakareev/muster-backlog/commit/32c54ede11fa82e62888d2b348f9486fa53013cc))
* **mcp:** serve every project at once over MCP ([0919791](https://github.com/FMakareev/muster-backlog/commit/09197915440970efe0b2aa08cf9709de6e7f3ceb))
* **parser:** read Backlog.md projects from disk ([6a518e9](https://github.com/FMakareev/muster-backlog/commit/6a518e97782705e3bc4e1425fbb58c654cae9e9e))
* **projects:** add, rename and retire milestones ([74e9f9b](https://github.com/FMakareev/muster-backlog/commit/74e9f9bbe8cf2556e1a8f428d9f7defabaef4a06))
* **projects:** choose a folder with the system dialog ([10061c8](https://github.com/FMakareev/muster-backlog/commit/10061c8573b6aa5115ff5752cf5780fcecdfb64d))
* **projects:** manage the registry from a screen instead of a text editor ([509f943](https://github.com/FMakareev/muster-backlog/commit/509f94337817d59f080f7e263f6ffcf543a33433))
* **release:** add issue forms and a pull request template ([95ac328](https://github.com/FMakareev/muster-backlog/commit/95ac328923b0270e36fc4c9d547a9e8b5fb34c72))
* **release:** attach built packages to the release, then publish it ([d58818d](https://github.com/FMakareev/muster-backlog/commit/d58818d1972b77b4ee9cbe1a8f331fa8cf4bbb4c))
* **release:** give the application its own icon ([e7568bc](https://github.com/FMakareev/muster-backlog/commit/e7568bcba1991b63c1be1613c810eed0001cd821))
* **release:** give the build a version it can report ([e9c0eb3](https://github.com/FMakareev/muster-backlog/commit/e9c0eb313218b0e9e029713cf67a77d55dcaf3b6))
* **store:** aggregate every project and follow the filesystem ([c9bff34](https://github.com/FMakareev/muster-backlog/commit/c9bff345023c6fd887a8693e80baeeceea7b665e))
* **task:** edit what a task points at, and order by dragging ([e9e3122](https://github.com/FMakareev/muster-backlog/commit/e9e3122e0f516e955636eb63f252b5c0ddcbd415))
* **task:** file, archive or send a task back from the panel ([140fb90](https://github.com/FMakareev/muster-backlog/commit/140fb905d8b23ae5e8d92c00f5d12611a5cb8d25))
* **task:** set a milestone from the panel, and make one from any picker ([6842d67](https://github.com/FMakareev/muster-backlog/commit/6842d67f43e58a8e75f55171a02d28069cea201f))
* **task:** write comments, not only read them ([641cd89](https://github.com/FMakareev/muster-backlog/commit/641cd8995cc365cc4b7a2428d37651799abd4b06))
* **ui:** add search, filters, the list, the overview and the docs viewer ([2e14c80](https://github.com/FMakareev/muster-backlog/commit/2e14c80f2711a2192f024e637370cc04292055e8))
* **ui:** add the task detail panel ([b154bb0](https://github.com/FMakareev/muster-backlog/commit/b154bb0e73ebfa2bcbe2f3a4b62f9a358de98962))
* **ui:** build the application shell ([a7cd281](https://github.com/FMakareev/muster-backlog/commit/a7cd281841eae3992f1496ac1fb067c41d2f695b))
* **ui:** create and edit tasks, name milestones, read centred, tray ([98d2d2a](https://github.com/FMakareev/muster-backlog/commit/98d2d2aa05a52c23cefdc07f86fa5ec0edd75f73))
* **ui:** make every degraded state readable ([d1aeefe](https://github.com/FMakareev/muster-backlog/commit/d1aeefee0e138c362f9ddd7f3b238fd43a40af9f))
* **ui:** render markdown for task bodies and documents ([bd411a5](https://github.com/FMakareev/muster-backlog/commit/bd411a580cfad4ab750f2b6d22af148872f7d408))


### Fixed

* **app:** stop the tray icon deadlocking the window open ([6d0bcc3](https://github.com/FMakareev/muster-backlog/commit/6d0bcc3e59d065ede48d5cb6978735d4506fd94f))
* **board:** make the plus on a column create a task ([22bebfc](https://github.com/FMakareev/muster-backlog/commit/22bebfc4d037b14a953a8b28d26c9e1c3df16f48))
* **cli:** ask the shell where the CLI is, instead of only guessing ([7a388cc](https://github.com/FMakareev/muster-backlog/commit/7a388cc3e32b7b956a1e5c90d3132fcdc4376f13))
* **cli:** find and run Backlog.md when the launcher's PATH cannot ([854e93f](https://github.com/FMakareev/muster-backlog/commit/854e93ff618346b62acfcb99da38fa217addc511))
* **deps:** declare lefthook, instead of hoping the machine has one ([722094a](https://github.com/FMakareev/muster-backlog/commit/722094abe4163ba6c87d175791d62ef823d76871))
* **mcp:** answer with an object, not a bare array ([7cdc66f](https://github.com/FMakareev/muster-backlog/commit/7cdc66f05dfae12d13d6ad7115104576e55d5f28))
* **mcp:** find the registry from a sandbox, and say when there is none ([2cacbe3](https://github.com/FMakareev/muster-backlog/commit/2cacbe35c7b022268252f1cb07c579877b89676f))
* **mcp:** register a command the client can spawn, not one only the container can see ([eebc1bf](https://github.com/FMakareev/muster-backlog/commit/eebc1bf7ef1e049334be24f1886516d60007147a))
* **mcp:** serve the protocol from a binary that runs where agents do ([0d9cc84](https://github.com/FMakareev/muster-backlog/commit/0d9cc8429411caa8f3e3b954312a17bfa355e115))
* **release:** run the workflows on the branch that exists ([592badf](https://github.com/FMakareev/muster-backlog/commit/592badfe0b23a3f84dd383995f6d1532710c6f1f))
* **release:** stop the AppImage failing on its own desktop entry ([8232b36](https://github.com/FMakareev/muster-backlog/commit/8232b3616e4e465b3245c63a158ad221b47d1673))
* **release:** update the version by line, not by rewriting the file ([a11e962](https://github.com/FMakareev/muster-backlog/commit/a11e96266ece2d18bbc11c2ff10ec54557637081))
* **task:** stop offering a note the edits it cannot take ([dc226a6](https://github.com/FMakareev/muster-backlog/commit/dc226a66a17a6649e95d58c282641efb722faac5))
* **ui:** close every overlay from the keyboard, wherever focus went ([876b4db](https://github.com/FMakareev/muster-backlog/commit/876b4db04629803ba247c57a70fe4bd363c42453))
* **ui:** let Preferences scroll when it outgrows the window ([400750c](https://github.com/FMakareev/muster-backlog/commit/400750cfe95d109684079cdc83e790c3f9e8df29))
* **ui:** make form controls readable and stop drags selecting text ([ea8fa55](https://github.com/FMakareev/muster-backlog/commit/ea8fa5511d5d01a356a985c044216e8b505608fc))
* **ui:** scale the interface and clear every accessibility violation ([57cd387](https://github.com/FMakareev/muster-backlog/commit/57cd387cbd9237c434e4f813683e081bb00afe8a))
* **ui:** start a form on the project you are looking at ([e97bd83](https://github.com/FMakareev/muster-backlog/commit/e97bd839b00ef735e417bd15c8771ca8aef6d06a))


### Documentation

* complete the open-source foundation ([df7e678](https://github.com/FMakareev/muster-backlog/commit/df7e678428700dfd08bd3b8e52ad464c225e0631))
* document the v0.1 build and add a smoke checklist ([fa356ee](https://github.com/FMakareev/muster-backlog/commit/fa356ee2965d4ed7a4922672fa0735757e106552))
* record the provisional MVP verdict and what it changed ([0759ff9](https://github.com/FMakareev/muster-backlog/commit/0759ff953f521ac33cadc50259dc21c8aec4cfbb))
* show what it looks like, and say what it is not ([7336324](https://github.com/FMakareev/muster-backlog/commit/7336324fb334d3950b42e0837167d8002fde2b71))
* **task:** record TASK-86 ([45a9217](https://github.com/FMakareev/muster-backlog/commit/45a9217364bee813c63f8f78f29f7c2e78df7e9c))

## Changelog

Every entry above is derived from the commit history rather than written by
hand: [release-please](https://github.com/googleapis/release-please) reads the
[Conventional Commits](https://www.conventionalcommits.org/) on the default
branch, opens a release pull request, and writes a new section directly under
the title when that pull request is merged. Edit a commit message, not this
file. These closing sections stay at the bottom because releases accumulate at
the top.

The format is [Keep a Changelog](https://keepachangelog.com/en/1.1.0/) and the
versioning is [Semantic Versioning](https://semver.org/spec/v2.0.0.html), read
the way the section below says.

## Versioning

- **A major bump — `1.0.0` to `2.0.0` — may break something.** It is the only
  release that may. What counts as breaking is listed below, and every one of
  them appears in the release notes under its own heading.
- **A minor bump — `1.0.0` to `1.1.0` — adds.** New capability, nothing taken
  away.
- **A patch bump — `1.0.0` to `1.0.1` — fixes.** Nothing added, nothing taken
  away.

Breaking means any of these:

- The registry file `projects.yml` stops being readable by the previous
  version, or a field in it changes meaning.
- The preferences file `settings.yml` loses a setting or changes what one
  does.
- The minimum Backlog.md CLI version goes up. Muster writes only through that
  CLI, so requiring a newer one is a requirement on the person running it.
- A command-line interface changes: the `mcp` subcommand, or the tools it
  serves over the Model Context Protocol, since an agent's configuration
  points at them.

Two things are deliberately **not** breaking changes, at any version:

- Anything about the Backlog.md files themselves. Muster reads and writes that
  format and does not own it; it adds no field of its own, so there is nothing
  there for it to break. See the guarantee in the README.
- The layout of the interface. Where a button lives is not an interface anyone
  can depend on programmatically.

### How 1.0 was reached

Not deliberately. release-please cuts the first release of a package as
`1.0.0` unless the configuration says otherwise, and this one did not, so the
first release the automation produced carried the number the project had
planned to arrive at later. It was kept rather than withdrawn, and the promise
above is now in force from that release onwards. The work the 1.0 milestone
was going to gate is tracked in the backlog and is still to be done.
