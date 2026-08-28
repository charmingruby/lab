# Lab

A lean Go foundation for proving ideas fast.

## Principles

This template encodes a single idea: **ports and adapters, nothing else.** Every domain follows the same dependency spine, every layer has exactly one job, every file lives where the pattern says it lives. Do not introduce machinery because it looks architecturally impressive. Understand the real constraint, then fight for the smallest model that makes the correct behavior unsurprising.

Channel both "measure twice, cut once" and "yagni." Fight scope creep. Try to honor the dev's intent in both a minimal and realistic fashion.

The rest of this document is meant to help you navigate the codebase and make changes effectively. Think of these instructions less as "hard rules," more as "good defaults." The developer's preferences should be able to override anything here.

Use the terminology defined in [docs/internals/glossary.md](docs/internals/glossary.md).

## Architectural rules

1. **Never skip the usecase.** The usecase owns business validation, transaction boundaries, and error mapping. Skipping it means scattered logic, no transaction safety, and tests that cannot isolate behavior.
2. **Never import another domain's internals.** Use its `client` port instead. Cross-module imports create invisible coupling and make it impossible to change one domain without breaking another.
3. **Never add layers that do not exist.** The pattern has exactly four layers — protocol, usecase, port (repository/client), adapter (postgres/console). Extra layers do not add safety; they add surface area for bugs.

See [docs/internals/architecture.md](docs/internals/architecture.md) for delivery mechanisms, module boundaries, repositories, clients, and external integrations.

## Domain structure

Domains use ports and adapters:

```
<protocol> → usecase → port → adapter
                       ↓
                     model
```

Each domain lives under `internal/<domain>/`. Cross-cutting concerns live in `internal/shared/`. Reusable infrastructure goes in `pkg/` — zero domain awareness, never imports from `internal/`.

## Development

Use the Taskfile for project commands. See [docs/internals/commands.md](docs/internals/commands.md) for available commands.

## Progressive Documentation Loading

CRITICAL: Only load the reference below that matches your current task. Do NOT read all docs up front — most work needs nothing beyond this file.

| Task | Load |
| --- | --- |
| Implement a domain feature (model → repository → usecase → endpoint) | [docs/internals/coding-patterns.md](docs/internals/coding-patterns.md) + mirror `internal/ticket/` |
| Add a layer, protocol, or module boundary | [docs/internals/architecture.md](docs/internals/architecture.md) + mirror `internal/ticket/` |
| Add or consume an external integration (storage, email, cache, third-party API) | [docs/internals/external-integrations.md](docs/internals/external-integrations.md) |
| Expose an application boundary — new/handled protocol, or a versioned HTTP/gRPC/queue interface | [docs/internals/versioning.md](docs/internals/versioning.md) |
| Read another module's data | [docs/internals/cross-module-reads.md](docs/internals/cross-module-reads.md) |
| Understand the architecture or runtime flow | [docs/internals/architecture.md](docs/internals/architecture.md) |
| Look up a term or find where code lives | [docs/internals/glossary.md](docs/internals/glossary.md) |
| Tests and verification | [docs/internals/testing.md](docs/internals/testing.md) |

## Plans and work artifacts

- Do not commit implementation plans, research notes, or agent scratch files. Keep temporary working material outside the worktree.
- Put durable architecture, constraints, and decisions in `docs/project/`. Update those docs when the product changes so agents find current facts instead of abandoned intentions.
- Nothing in `docs/internals/` should reference `docs/project/`. They are separate audiences.

## Taste

- Complexity belongs at the adapter boundary. Usecases stay pure, endpoints stay thin.
- Inferred types over annotations. `any` is the enemy.
- Comments describe how a thing is used, and move when the code moves. Use them mostly to describe functions, not to annotate every line of behavior.
- If a rule here fights the task in front of you, say so loudly and get a human sign-off before breaking it.

## When in Doubt

1. Copy the pattern from `internal/ticket/`.
2. If `internal/ticket/` doesn't cover the case, pick the option that adds the least new structure.
