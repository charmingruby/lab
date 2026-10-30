---
name: lab-sync-docs
description: Review or update documentation for consistency with the current repository and its documented decisions. Use for documentation audits, stale project facts, broken references, conflicting instructions, or alignment after implementation changes.
---

# Lab sync docs

Keep documentation aligned without redesigning it. This skill owns the procedure; the active repository supplies its conventions, decisions, layout, and verification commands. Discover those sources on each invocation rather than carrying assumptions from an earlier architecture or review.

## Workflow

1. **Bound the review.** Identify the requested files or subject, read applicable repository instructions, and inspect the working tree. Use the current documentation index or entry points to locate relevant sources. For a changed-files review, include related documents only where needed to verify a claim or reference. For a directory review, account for every document in that directory.
2. **Establish intent.** Distinguish conventions and decisions, illustrative examples, planned changes, and claims about the project's current state. Follow the repository's documented authority and supersession rules. If the applicable guidance is missing or conflicting, report the gap and ask before making a correction that depends on resolving it.
3. **Compare claims.** Check definitions, mandatory rules, exceptions, and examples against each other. Verify claims about the project's current state against relevant implementation, configuration, or tests within the agreed scope, including API signatures and command arguments. Use documented decisions to assess intended behavior; a code mismatch alone does not make a decision obsolete. Read linked material only when needed for a claim under review or required by repository instructions. Check local link targets and heading anchors; identify claims that could not be verified.
4. **Classify findings.** Separate contradictions, ambiguities, stale references, and redundant authoritative rules. Preserve useful summaries and checklists. An illustrative capability missing from the current codebase is not a defect; a claimed local function that does not exist is.
5. **Report or correct.** A review request produces findings, not edits. When correction is requested, make the smallest change that resolves each confirmed finding and retain unrelated content and user changes. Preserve architectural decisions and document organization unless the user explicitly asks to change them.
6. **Verify.** Recheck affected claims, links, and examples after edits. Follow the repository's documentation verification requirements and report any unavailable checks. Inspect the final diff for unrelated changes.

Resolve references from their owning document in the active repository, not this skill's directory. Discover current locations and commands from repository sources; report missing references rather than inventing replacements.

## Finding format

For each confirmed finding, provide:

- **Location:** file and line; both locations for conflicting claims.
- **Conflict:** the incompatible statements, stale reference, or mismatch with verified repository facts.
- **Consequence:** the concrete mistake an agent or maintainer could make.
- **Minimal correction:** the smallest wording or example change that resolves it.

Prioritize contradictions over optional cleanup. State uncertainty rather than presenting a stylistic preference as a defect. If no confirmed issues remain, say so without proposing a rewrite.

## Completion

Every document in the agreed scope is accounted for. Each reported finding has evidence and a minimal correction; each applied correction has been rechecked. The handoff states the scope, findings or edits, verification results, and any unresolved decisions. Keep temporary audit notes outside the worktree unless the user requests a durable report.
