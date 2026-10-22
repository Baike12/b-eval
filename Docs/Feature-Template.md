# Feature Template

> Template for every document in `Docs/Features/`.
> Copy this file, rename it to `FEAT-<nnn>-<slug>.md`, and fill in every section.
> Delete this quote block and the guidance in each section before committing.

**Best-practice sources for this template:** GitHub issue-form guidance (structured, typed, unambiguous fields); engineering-spec practice (separate problem from solution, explicit non-goals, testable acceptance criteria); feature-request intake practice (problem, current workaround, impact, evidence). This template keeps the parts that survive engineering scrutiny and drops the intake ceremony.

**Design rules of the template:**

- **Problem before solution.** The problem statement must stand on its own; the proposed solution is a hypothesis to be validated, not a given.
- **Non-goals are mandatory.** A feature without a stated boundary produces unbounded scope.
- **Requirements are numbered and testable.** `R1`, `R2`, … Each acceptance criterion maps to a requirement.
- **Acceptance criteria are observable outcomes**, written so that a tester can confirm them without asking the author what was meant.
- **One feature per document.** Split anything that needs the word "and" in its title.

---

# FEAT-<nnn>: <Short title, imperative, under 70 characters>

| Field | Value |
|---|---|
| **ID** | FEAT-<nnn> |
| **Status** | Draft \| In Review \| Approved \| In Progress \| Implemented \| Closed |
| **Owner** | <name or team> |
| **Priority** | P0 \| P1 \| P2 \| P3 |
| **Implements** | `F-<module>-<nn>`, `F-<module>-<nn>` (functional requirement IDs from the functional design) |
| **Created** | YYYY-MM-DD |
| **Last Updated** | YYYY-MM-DD |

## Summary

<Three sentences maximum: what is being built, for whom, and the single most important outcome.>

## Problem

**Current behaviour:** <What happens today.>

**Pain:** <Who is hurt, how, and how often.>

**Evidence:** <Links to tickets, logs, user reports, measurements. "None" is an acceptable answer if this is a speculative bet — say so explicitly.>

## Proposed Solution

<Two to five sentences. Describe the user-visible behaviour, not the implementation. The implementation belongs in the Tech Design document.>

## Scope

### In Scope

- <Bullet per included behaviour.>

### Out of Scope

- <Explicitly excluded, with a one-line reason where it is not obvious.>

## Requirements

Numbered, independently testable, and each traceable to an acceptance criterion.

- **R1** <Requirement stated as an observable behaviour.>
- **R2** <...>
- **R3** <...>

## Acceptance Criteria

Written as verifiable outcomes. Use Given/When/Then only where the setup is non-obvious.

- [ ] **AC1** <Maps to R1. Observable pass condition.>
- [ ] **AC2** <Maps to R2.>
- [ ] **AC3** <Maps to R3.>

## Impact

| Area | Effect |
|---|---|
| Users | <Who gains, who is affected, how.> |
| Existing behaviour | <What changes for people already using the product.> |
| Operational cost | <Tokens, wall-clock time, storage, human review load.> |

## Risks and Mitigations

| Risk | Likelihood | Mitigation |
|---|---|---|
| <Risk> | Low/Med/High | <What bounds it.> |

## Rollout

<How this reaches users, and the condition under which it is rolled back.>

## References

Replace the block below with live links to this feature's technical design and test report.
Both links are mandatory; a feature document without them is incomplete.

```markdown
## References

- Tech Design: [TD-001 <Title>][td-001]
- Test Report: [TR-001 <Title>][tr-001]

<!-- Link definitions stay at the bottom of the file. Angle brackets are required
     because "Tech Design" and "Test Report" contain spaces. -->

[td-001]: <../Tech Design/TD-001-<slug>.md>
[tr-001]: <../Test Report/TR-001-<slug>.md>
```

## Open Questions

| # | Question | Owner | Blocking? |
|---|---|---|---|
| 1 | <Question.> | <Name> | Yes/No |
