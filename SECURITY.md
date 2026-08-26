# Security Policy

## Supported versions

Muster is pre-1.0 and pre-release. There is no released version yet, and therefore no maintained release branch: only the latest commit on the default branch is supported.

| Version | Supported |
| :-- | :-- |
| default branch | yes |
| anything else | no |

Once 1.0 ships, this table will name the release lines that receive fixes.

## Reporting a vulnerability

**Do not open a public issue for a security problem.**

Report it privately through GitHub: open the repository's **Security** tab and choose **Report a vulnerability**. That opens a private advisory visible only to the maintainer. If private reporting is unavailable to you, send a private message to [@FMakareev](https://github.com/FMakareev) asking for a channel, without details of the problem.

Please include what you would need yourself to reproduce it: the version or commit, the operating system, the steps, and what happens versus what should. A proof of concept helps; an exploit for someone else's machine does not.

Expect an acknowledgement within a week. This is a personal project maintained in spare time — that is the honest figure, not a service level agreement.

## What is in scope

Muster is a single-user, local-first desktop application. It has no server, no accounts, no network service and no cloud component, so the usual web attack surface does not apply. What does matter:

- **Arbitrary code or command execution.** Every write goes through the `backlog` CLI as a subprocess. Anything that allows argument injection, shell interpolation, or execution of an attacker-chosen binary through that path is a vulnerability.
- **Path traversal outside registered projects.** Muster reads and writes only within the folders listed in its registry. Anything that escapes that boundary is a vulnerability.
- **Untrusted markdown.** Task bodies, documents and decisions are rendered as markdown from files the application did not write. Script execution, network requests, or filesystem access triggered by rendering a task is a vulnerability.
- **Data destruction.** Muster operates on real repositories. Anything that loses or corrupts a task file beyond what the CLI itself would have produced is a vulnerability.

## What is out of scope

- A malicious `backlog` binary already present on the user's `PATH`. Muster trusts the CLI it is configured to call; a compromised machine is outside the threat model.
- Content in a project a user deliberately registered. Registering a folder is an act of trust in that folder, the same way opening a repository in an editor is.
- Anything requiring physical or already-privileged access to the machine.
