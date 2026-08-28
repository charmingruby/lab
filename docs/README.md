# Documentation

Everything below is for maintainers and agents working on the codebase. Agent rules live in [AGENTS.md](../AGENTS.md).

## Project

What this project is, why it exists, and the decisions behind it. Architecture, constraints, product rules, business logic.

_(No docs yet — add product and architecture decision records here.)_

## Internals

How to work in this codebase. Patterns, guidelines, integration rules, terminology.

- [Architecture overview](./internals/architecture.md) — runtime flow, request lifecycle, domain structure
- [Glossary](./internals/glossary.md) — domain terms with source file links
- [Coding patterns](./internals/coding-patterns.md) — complete implementation reference per layer
- [Cross-module reads](./internals/cross-module-reads.md) — the client-port pattern for inter-domain data
- [External integrations](./internals/external-integrations.md) — storage, email, cache, third-party APIs
- [Versioning](./internals/versioning.md) — endpoint-level versioning with v1-to-v2 example
- [Testing](./internals/testing.md) — test conventions and verification steps
- [Commands](./internals/commands.md) — available Taskfile commands
