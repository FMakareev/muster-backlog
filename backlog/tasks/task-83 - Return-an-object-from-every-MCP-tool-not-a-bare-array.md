---
id: TASK-83
title: 'Return an object from every MCP tool, not a bare array'
status: Done
assignee:
  - '@claude'
created_date: '2026-08-28 03:44'
updated_date: '2026-08-28 03:49'
labels: []
milestone: m-4
dependencies: []
priority: high
type: bug
ordinal: 83000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Reported from Claude Code running on the host against the installed server. The server connects and five of the ten tools fail before their answer reaches the agent:

  Invalid result for tools/call: path: ["structuredContent"] expected: "record" received: "array"

The Model Context Protocol says structuredContent in a tools/call result is a JSON object. Five handlers return a Go slice, so the SDK infers an output schema of type ["null","array"] and puts a bare array there. A client that validates the result rejects the whole response, and the agent never sees data that the server read correctly - measured by hand, list_projects produces 7215 bytes of perfectly good JSON before it is thrown away.

The split is exactly along that line. Broken: list_projects, list_tasks, list_milestones, list_entities, search. Working: get_task and the four write tools, all of which return a struct.

Nothing in the test suite noticed, because the tests call the tools through a session that reads the structured result directly rather than validating it against the declared schema the way a strict client does.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Every tool declares an output schema of type object
- [x] #2 The five listing tools answer a client that validates the result
- [x] #3 A tool that answers with a bare array again is caught by the tests
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
The report was exactly right, including the diagnosis. The protocol says structuredContent in a tools/call result is a JSON object; five handlers returned a Go slice, so the SDK inferred an output schema of ["null","array"] and put a bare array there. A client that validates the result rejected the whole response before the agent saw any of it - and the data was fine, which is what made it look like a server fault rather than a shape fault.

Confirmed before changing anything by dumping every declared schema: the split fell exactly along the line reported, five arrays and five objects, and the five objects were the ones that worked.

Each listing now answers with an object around its list, with a named field rather than a shared envelope, so the name says what the list holds when an agent reads the schema: projects, tasks, milestones, entities, hits. Three of the wrapper types had to be declared after the types they hold, because a type declared inside a function body cannot refer forward.

The reason nothing caught it is worth more than the fix. These tests decode structuredContent straight into a Go type, which happily accepts an array; a strict client validates it against the schema the server declares, and never gets that far. So there is now a test that reads the declared schema of every tool and requires type object. It needs no project data, so it runs anywhere, and putting a bare array back on list_milestones made it fail with the tool named.

Verified through the protocol against a rebuilt server, doing what the client does: all ten schemas are objects, and calling the five listings returns structuredContent that is an object rather than an array, carrying five projects, twenty-four tasks, four milestones, two entities and two search hits.

Worth knowing for anything downstream: this changes the shape of five tools' answers. Nothing could have been consuming them, since every strict client rejected them outright, so there is nobody to break.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
Five tools returned a bare array as structuredContent, which the protocol says must be an object, so a client that validates the result threw away answers the server had read correctly. Each listing now answers with an object around its list, named for what it holds.

The reason nothing caught it matters more than the fix: the tests decode the structured answer straight into a Go type, which accepts an array, while a strict client validates against the declared schema and never gets that far. A test now reads every tool's schema and requires type object; putting a bare array back on one tool makes it fail with that tool named.

Verified through the protocol against a rebuilt server: ten schemas of type object, and the five listings returning objects carrying five projects, twenty-four tasks, four milestones, two entities and two hits.
<!-- SECTION:FINAL_SUMMARY:END -->
