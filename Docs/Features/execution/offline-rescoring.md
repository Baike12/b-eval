# FEAT-012: Offline Re-scoring

| Field | Value |
|---|---|
| **ID** | FEAT-012 |
| **Status** | Implemented |
| **Owner** | Platform team |
| **Priority** | P0 |
| **Implements** | `F-JDG-07` |
| **Created** | 2026-11-05 |
| **Last Updated** | 2026-11-05 |
| **Source** | `Functional design chapter 03 - Evaluation Execution and Orchestration, section 3.7` |

## Summary

A completed run can be judged again from its stored trajectories, without calling the agent, and the result is stored as a new run that names the one it re-judged. The single most important outcome is that a scoring mistake or a changed criterion can be corrected over historical evidence while the original verdict remains readable and comparable.

## Problem

**Current behaviour:** The platform persists every trajectory, but the only way to obtain a verdict is to execute the plan again. `POST /api/runs` always drives the adapter; there is no path that reads an existing trajectory and judges it.

**Pain:** Correcting a scorer, changing a rubric or applying a corrected task-set version forces a full re-execution. That costs the agent's budget again, changes the random seeds, and produces a different set of trials, so the corrected and original numbers are not comparable. FEAT-003 AC1 and AC28 and FEAT-004 AC9 describe offline re-judgement as a platform capability; without it the trajectory is an archive rather than an asset.

**Evidence:** FEAT-003 R1 states that judgement (step 7) is decoupled from execution (step 5). TD-004 records the trajectory as the record that "carries everything needed to re-judge the trial offline". The stored `domain.Trajectory` already carries the answer, the steps, the usage and the termination reason.

## Proposed Solution

`POST /api/runs/{id}/rescore` reads a completed run, re-runs the scorers over its saved trajectories, and stores the result as a new run whose `rescored_from` field names the source and whose `rescore_reason` records why the judgement changed. The call optionally takes a different task-set version, so a corrected scoring configuration can be applied to the same agent behaviour. The source run is never modified.

## Scope

### In Scope

- A re-judgement endpoint over a completed run's trajectories.
- An optional task set and version, so a corrected scorer configuration can be applied.
- A mandatory reason for the change.
- A new run record carrying `rescored_from` and `rescore_reason`, with the source run unchanged.
- Preservation of the execution outcome: a trial that did not complete stays a failure of its original class, and a platform failure stays UNSCORED.
- Re-judged runs are reportable, comparable and exportable exactly like an executed run.
- A refusal when a trajectory's task is no longer present in the task set.

### Out of Scope

- **Changing the agent's behaviour.** Re-judgement judges the recorded behaviour; it never re-executes the agent.
- **Editing a trajectory.** The trajectory is evidence; a change to it is a new execution, not a re-judgement.
- **Automatic re-judgement on a task-set publish.** The operator triggers the re-judgement and states the reason.
- **A verdict diff view.** The two runs are compared through the existing comparison endpoint.

## Requirements

- **R1** The platform must re-judge a completed run from its persisted trajectories without calling the agent adapter.
- **R2** A re-judgement must require a non-empty reason.
- **R3** The result must be a new run record that names its source in `rescored_from`, carries the reason in `rescore_reason`, and leaves the source run unchanged.
- **R4** A re-judgement must accept an optional task set and version and judge the trajectories against that version's scorers.
- **R5** The original execution outcome must be preserved: a non-completed trial keeps its failure class and is FAIL, except a platform failure which stays UNSCORED.
- **R6** A trajectory whose task is not in the judging task set must refuse the whole re-judgement rather than be skipped.
- **R7** A re-judged run must be reportable and exportable through the same endpoints as an executed run.
- **R8** Two re-judgements of the same run must not overwrite each other.

## Acceptance Criteria

- [ ] **AC1** (R1) Re-judging a completed run returns a new run whose trials are judged from the existing trajectories, and the agent is not invoked.
- [ ] **AC2** (R2) A re-judgement without a reason is refused with status 400 and a message naming the reason.
- [ ] **AC3** (R3) The new run carries `rescored_from` and `rescore_reason`, and a subsequent read of the source run shows it unchanged.
- [ ] **AC4** (R4) Re-judging against a second task-set version applies that version's scorer: a trajectory that passed under a `contains` scorer fails under an `exact` scorer, and the new plan names the second version.
- [ ] **AC5** (R5) A trajectory that ended on the step limit is re-judged as budget-exhausted with verdict FAIL and score 0.
- [ ] **AC6** (R6) A trajectory whose task is absent from the judging task set makes the call fail with a message naming the task and the trial.
- [ ] **AC7** (R7) The re-judged run has a report and a CSV export.
- [ ] **AC8** (R8) Two re-judgements in the same second produce two distinct run identifiers.

## Impact

| Area | Effect |
|---|---|
| Users | Evaluation organisers and task authors correct a scoring decision over historical runs and compare the old and new verdicts; the agent's budget is not spent again. |
| Existing behaviour | `POST /api/runs` is unchanged. A stored run is now the input to a second judgement rather than a finished record. |
| Operational cost | One scorer pass per trajectory, including the judge calls a model scorer makes. No agent calls and no new environment instances. |

## Risks and Mitigations

| Risk | Likelihood | Mitigation |
|---|---|---|
| A re-judgement silently replaces the original verdict | Low | The source run is never modified; the result is a new record with `rescored_from` and a mandatory reason |
| Re-judging against another task-set version mislabels a trial | Med | The whole call is refused when any trajectory's task is missing; the new plan records the version it was judged under |
| A model judge returns a different grade on a second pass | Med | The judge runs at temperature 0 and the prompt version is recorded; the difference is visible in the two runs, not hidden |
| Two re-judgements collide on one run identifier | Low | The identifier carries a timestamp and a numeric suffix on collision |

## Rollout

The endpoint ships enabled and changes no existing path. Rollback removes the two routes; the runs already produced by a re-judgement remain readable because they are ordinary run records.

## References

- Tech Design: [TD-009 Offline Re-scoring][td-009]
- Test Report: [TR-012 Offline Re-scoring][tr-012]
- Related features: [FEAT-003 Evaluation Execution and Orchestration][feat-003], [FEAT-011 Model Judge Scoring][feat-011]

[td-009]: <../../Tech Design/TD-009-offline-rescoring.md>
[tr-012]: <../../Test Report/TR-012-offline-rescoring.md>
[feat-003]: <./evaluation-execution-and-orchestration.md>
[feat-011]: <../scoring/model-judge-scoring.md>

## Open Questions

| # | Question | Owner | Blocking? |
|---|---|---|---|
| 1 | Should a re-judgement be allowed to override a single scorer inline, instead of requiring a task-set version? | Platform team | No |
| 2 | Should the platform mark an evaluation comparison as "same execution" when the two runs share trajectories? | Platform team | No |
