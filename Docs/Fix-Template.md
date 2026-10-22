# Fix Template

> Template for every document in `Docs/Fix/`.
> Copy this file, rename it to `FIX-<nnn>-<slug>.md`, and fill in every section.
> Delete this quote block and the guidance in each section before committing.

**Best-practice sources for this template:** bug-report intake practice (one bug per report, numbered reproduction steps, expected vs. actual, environment, objective language, severity separated from priority); root-cause-analysis practice (classify the failure mode, not just the symptom); regression-test practice (the reproduction steps become an automated test that runs on every change).

**Design rules of the template:**

- **One defect per document.** Two unrelated symptoms in one report break tracking, assignment and verification.
- **Severity and priority are different axes.** Severity is technical impact; priority is business urgency. A cosmetic issue in the leaderboard can be high priority; a crash in an internal path can be low priority.
- **Objective language.** "The task list shows the wrong result" — not "the UI is broken".
- **Root cause, not symptom.** Patching the symptom without naming the cause guarantees a recurrence.
- **The fix ships with a regression test**, and that test is the proof the fix works.
- **Impact must be quantified.** "Some scores were wrong" is not an impact statement; "scores for 3 tasks in the 2026-09 batch were overstated by up to 20 points" is.

---

# FIX-<nnn>: <Short title stating the observable defect>

| Field | Value |
|---|---|
| **ID** | FIX-<nnn> |
| **Status** | Open \| In Progress \| Fixed \| Verified \| Closed |
| **Owner** | <name or team> |
| **Severity** | S1 Critical \| S2 High \| S3 Medium \| S4 Low |
| **Priority** | P0 \| P1 \| P2 \| P3 |
| **Affects** | `F-<module>-<nn>` (functional requirement IDs whose behaviour was wrong) |
| **Detected in** | <Version, release, or run ID.> |
| **Created** | YYYY-MM-DD |
| **Last Updated** | YYYY-MM-DD |

## Summary

<Two sentences: the defect in observable terms, and the consequence.>

## Severity and Impact

**Technical impact:** <What is wrong, and under what conditions.>

**Blast radius:** <Which data, which results, which users, which time window. Quantify it.>

**Detection:** <How it was found — monitoring, user report, test failure, audit.>

## Symptom

<What is observed, stated factually. Include the exact error text where one exists.>

## Reproduction

**Environment:** <Version, configuration, dataset or task set, model, seed.>

**Preconditions:** <State the starting condition required for the defect.>

**Steps:**

1. <Step.>
2. <Step.>
3. <Step.>

**Expected:** <What should happen, per the specification.>

**Actual:** <What happens instead.>

## Root Cause

**Cause:** <The specific mechanism that produces the symptom.>

**Failure class:** Logic error \| State handling \| Race condition \| Configuration \| Data \| Integration \| Resource exhaustion

**Why it was not caught:** <Which guard was missing, or which assumption was wrong.>

## Fix

**Change:** <What is changed. The implementation detail lives in the Tech Design document if one is needed.>

**Rationale:** <Why this change addresses the cause rather than the symptom.>

**Guardrail added:** <What now prevents this class of defect from recurring silently — a check, an assertion, a threshold, a test.>

## Verification

| Check | Method | Result |
|---|---|---|
| Original reproduction no longer reproduces | <How it was re-run> | <Result> |
| Regression test added | Test ID / name | <Result> |
| Adjacent behaviour unaffected | <What was checked> | <Result> |
| Blast radius re-scored | <How historical results were re-checked> | <Result> |

## Regression Prevention

- **Test added:** <Test ID, where it lives, what it asserts.>
- **Run gate:** <Where this test runs — every commit, every pull request, nightly.>
- **Detection change:** <Any new monitoring, assertion, or health metric added.>

## References

Replace the block below with a live link to this fix's test report.
The link is mandatory; a fix document without it is incomplete.

```markdown
## References

- Test Report: [TR-001 <Title>][tr-001]

<!-- Link definitions stay at the bottom of the file. Angle brackets are required
     because "Test Report" contains a space. -->

[tr-001]: <../Test Report/TR-001-<slug>.md>
```

## Open Questions

| # | Question | Owner | Blocking? |
|---|---|---|---|
| 1 | <Question.> | <Name> | Yes/No |
