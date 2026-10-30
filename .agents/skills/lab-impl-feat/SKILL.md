---
name: lab-impl-feat
description: Implement domain features using the active repository's documented architecture and conventions. Use for new domain behavior, business-rule changes, and domain bug fixes requiring implementation, tests, and verification; not for documentation-only or formatting changes.
---

# Lab implement feature

Deliver a focused behavior change using the repository's current documented architecture. This skill owns the procedure; the active repository supplies its language, boundaries, layout, and commands. Discover those sources on each invocation rather than carrying assumptions from an earlier architecture or feature.

## Workflow

1. **Define the behavior.** Identify the requested outcome, relevant inputs, success and failure cases, and compatibility constraints. Read applicable repository instructions and relevant product records. Resolve uncertainties that affect observable behavior before coding; otherwise state a narrow assumption and proceed.
2. **Load the applicable conventions.** Use the current documentation index or entry points to find implementation and testing rules. Follow required readings and load other guidance when it governs an affected concern. Apply the repository's documented authority and supersession rules. If missing or conflicting guidance affects behavior or an architectural boundary, ask before implementing the affected part. Obtain approval for departures from required boundaries.
3. **Trace the existing path.** Inspect the affected behavior, dependencies, assembly, and tests using the architecture and terminology just established. Use existing code as implementation evidence, not as an override for explicit rules; surface disagreements that affect the change. Identify only the files that need changes, including assembly for new behavior. Inspect the working tree and preserve unrelated work.
4. **Implement the smallest change.** Follow the established dependency direction, naming, error mapping, transaction handling, and assembly conventions. Reuse existing helpers after checking their signatures. Introduce only the structure required by the requested behavior; keep adjacent refactors outside the change.
5. **Prove the behavior.** Add or update focused tests for every new or changed behavior, including the failure paths and persisted side effects relevant to the task. A bug fix includes a regression test that fails without the fix and passes with it. Follow the repository's test isolation and dependency conventions rather than inventing a second testing strategy.
6. **Verify and inspect.** Discover verification commands in the current repository instructions and executable configuration. Run relevant tests and required project checks, inspecting any formatting changes. Run a broader suite when required by repository guidance or shared-code impact. Report missing tools or infrastructure as blockers, not successful checks. Review the final diff against the requested behavior.

Resolve references from their owning document in the active repository, not this skill's directory. Discover current locations, helpers, and commands from repository sources; report missing references rather than inventing replacements.

## Documentation updates

If behavior or a caller-facing contract changes, update the affected durable documentation. Keep reusable implementation conventions separate from product-specific rules, following the repository's existing layout. Record an architectural decision only when the change actually requires one; implementation plans and scratch notes stay outside the worktree.

## Completion

- Every requested behavior has an implementation and focused test coverage, or an explicit unresolved blocker.
- Changed contracts, transactions, integration failures, and wiring have been checked where applicable.
- Relevant tests and project checks have recorded results; unexecuted checks are clearly identified.
- The final diff contains only the requested change and directly necessary supporting changes.

Hand off with the outcome, the important changed paths, verification results, and remaining limitations. Commit or push only when requested.
