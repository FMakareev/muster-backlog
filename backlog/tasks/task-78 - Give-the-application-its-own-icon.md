---
id: TASK-78
title: Give the application its own icon
status: Done
assignee:
  - '@claude'
created_date: '2026-08-27 20:13'
updated_date: '2026-08-27 20:28'
labels: []
milestone: m-4
dependencies: []
priority: high
type: chore
ordinal: 78000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Every icon the application ships is Wails' own W: build/appicon.png and the vector behind it, the Windows .ico, the macOS .icns, the AppImage icon and whatever the desktop shows in a launcher or a tray. It went out in the .deb that was installed for testing. Distributing another project's mark as our own is wrong on its face, and it has to be settled before release artefacts are produced or the repository is public.

The mark is chosen: the letter M built from the same coloured squares the interface uses to say which project something belongs to. It follows the rule the interface states about itself - there is deliberately no accent hue, because every colour belongs to a project - so the mark is a muster of project colours rather than a brand colour. It was picked over two alternatives because it is the only one still legible at 16 pixels, which is the size in the tray.

The source is one vector; every other format is generated from it, so the mark cannot drift between platforms.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Every icon the application ships is the new mark
- [x] #2 No Wails artwork or naming is left anywhere in the build assets
- [x] #3 The icon is generated from a single source rather than maintained per platform
- [x] #4 The mark is legible at 16 pixels on a light, a dark and a mid-grey ground
<!-- AC:END -->

## Definition of Done
<!-- DOD:BEGIN -->
- [x] #1 Linters and formatters pass across Go and frontend
- [x] #2 Automated tests cover the change and the suite is green
- [x] #3 User-facing behaviour change is reflected in README or docs
- [x] #4 Commits follow Conventional Commits and are scoped to this task
<!-- DOD:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
The mark is the letter M built from the same squares the interface uses to say which project something belongs to. Four strokes, four colours, meeting on one chalk-coloured block. It follows the rule the interface states about itself - there is deliberately no accent hue, because every colour belongs to a project - so it is a muster of project colours rather than a brand colour.

Three candidates were drawn and shown at 64, 32 and 16 pixels on a dark, a light and a mid-grey ground, because that is where an icon is judged rather than at 1024. A fourth was drawn and dropped on the evidence: scattered squares drawn into one aligned column, the literal reading of the word, whose connecting rules disappear by 32px and leave an arrangement of dots. The letter was chosen because it was the only one still unmistakably itself at 16, which is the size in the tray.

build/appicon.svg is the source and build/icon/render.mjs rasterises it into every PNG the build needs; wails3 generates the .ico and .icns from those. Proven rather than asserted: every PNG was deleted and rebuilt from the committed SVG, then the containers regenerated, then unpacked and looked at. Playwright is deliberately not a dependency - the icon changes about never and a browser download in every contributor's install would be a poor trade - so the script asks for it by name and says so when it is missing.

Three places in the running application had no icon at all, which is why nobody had noticed a Wails logo there: the tray sets none on Linux, and neither did the window or the about box. All three now carry a 256px render, embedded in the binary from build/appicon-256.png. The 1024px sheet would have been resampled hard on every draw; a tray draws at about twenty-two pixels.

The window icon lives under Linux options rather than beside Title. LinuxWindow is a value and not a pointer, so setting it changes nothing else - the GPU policy keeps the same zero value it already had. The doc comment warning that a nil Linux block changes that policy is stale v2 phrasing.

Four things came out of looking that were not what the task was about, all in assets that ship.

The macOS disk image carried Wails' own artwork with the word WAILS across it. Replaced rather than deleted, because the darwin Taskfile names both files and a build that cannot find them is worse than a plain one for a platform this release does not claim.

The .desktop entry said Name=muster, had no description at all, and set Keywords=wails - the framework this happens to be built with, which is not a word anyone types into a launcher looking for this. It is generated on every build, so hand-editing would not have held; the generate step now corrects the three lines after the generator runs.

The manifests were full of template placeholders that had been committed and shipped: CFBundleName "My Product", CFBundleIdentifier com.example.muster, copyright "My Company", and the .deb's own desktop template naming the product "My Product". Regenerating the build assets from build/config.yml fixed all of them, once companyName and comments were actually filled in.

But that same regeneration is a trap and has to be recorded: wails3 task common:update:build-assets puts the template's defaults back over three values in nfpm.yaml - the version becomes a literal 0.0.0 instead of the build's, the vendor becomes "My Company", and the homepage becomes wails.io. It had already reverted a hand-correction once. The file now carries a comment naming exactly those three lines.

Two files keep Wails in their names and stay: build/windows/nsis/wails_tools.nsh and build/windows/wails.exe.manifest. They are the framework's own build tooling, referenced by those names from the Taskfile and project.nsi, and neither carries a picture or a name a user ever sees. The check that proves no Wails naming remains lists them as deliberate exceptions rather than passing silently over them.

Verified: the .ico carries six members including the 16px one a taskbar draws, both .icns carry eight PNGs, all of them unpacked and looked at; the mark holds at 48, 32, 22 and 16 on all three grounds in every file that ships, read from those files and not from the design-time vector. wails3 task lint clean, the whole suite green.

Left for TASK-13 rather than done here: build/linux/appimage/muster is a 14MB compiled binary committed to the repository. It should not be tracked, and the history needs reviewing before the first public push, which is that task's first acceptance criterion.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
The application had been shipping Wails' own W as its icon — in the .deb, the AppImage, and everywhere a desktop draws one. It now has its own: the letter M built from the same coloured squares the interface uses to say which project something belongs to, following the rule the interface states about itself, that there is no accent hue because every colour belongs to a project.

Three candidates were shown at 64, 32 and 16 pixels on three grounds, because that is where an icon is judged. A fourth was drawn and dropped on the evidence rather than argued about. The letter won for being the only one still unmistakably itself at 16, which is the size in the tray.

build/appicon.svg is the source and build/icon/render.mjs makes every PNG from it; the .ico and .icns come from those. Proven by deleting every PNG and rebuilding from the committed vector, then unpacking the containers and looking.

Three places in the running application had no icon at all — the tray, the window and the about box — and all three now carry it, embedded in the binary.

Four things turned up in assets that ship: Wails artwork on the macOS disk image with the word WAILS across it; a desktop entry named after the binary with no description and Keywords=wails; template placeholders that had been committed and shipped, including a .deb calling the product 'My Product' and a bundle identifier of com.example.muster; and the discovery that regenerating build assets silently reverts three hand-corrected values in nfpm.yaml, which that file now warns about by name.

Five Go tests keep it that way — the containers carry the small sizes a taskbar draws, the binary carries a 256px mark, no framework branding and no template placeholder survives in a shipped asset — and each was made to fail before it was trusted.
<!-- SECTION:FINAL_SUMMARY:END -->
