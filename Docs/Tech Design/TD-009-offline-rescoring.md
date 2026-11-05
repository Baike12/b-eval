# TD-009: Offline Re-scoring

| Field | Value |
|---|---|
| **ID** | TD-009 |
| **Status** | Implemented |
| **Implements** | [FEAT-012 Offline Re-scoring](../Features/execution/offline-rescoring.md) |
| **Owner** | Platform team |
| **Created** | 2026-11-05 |
| **Last Updated** | 2026-11-05 |

> This document specifies *how* offline re-scoring is built. The *what* and *why* live in FEAT-012.

## 1. Problem Restatement

The platform stores a `domain.Trajectory` per trial, carrying the answer, the steps, the usage and the termination reason. A verdict is a pure function of the task's scorer configuration and that trajectory. Re-judgement is therefore a matter of applying a new function to existing data, but three questions have to be answered before it can be exposed: which run record holds the result, what happens to the execution outcome, and how the change is made auditable.

## 2. Design Goals

| Goal | Consequence for the design |
|---|---|
| The source verdict survives | The result is a new run; the source is never written. |
| The change is auditable | `rescored_from` names the source and `rescore_reason` is mandatory. |
| A corrected scorer can be applied | The call accepts another task set and version. |
| The execution outcome is not re-interpreted | The failure class is recomputed from the trajectory, not from the new verdict. |
| A missing task is not a silent skip | The whole call is refused. |

## 3. Algorithm

```
Rejudge(taskSet, agentID, trajectories, scorerRegistry):
    index tasks by task_id
    for each trajectory:
        task = index[trajectory.task_id]        or refuse the call
        failure = Classify(trajectory.termination)
        verdict = Judge(task, trajectory, registry)
        absorb the judge's usage into trajectory.usage
        if failure != completed:
            platform failure -> verdict.UNSCORED
            otherwise        -> verdict.FAIL, score 0
        outcome = buildOutcome(task, trialIndex, seed, trajectory, failure, verdict)
    return outcomes, verdicts
```

`buildOutcome` is the same function the executor uses, so a re-judged trial enters the metric layer with the same fields as a fresh one.

The trial index comes from `Trajectory.TrialIndex`. A run recorded before that field existed carries the index only in its trial identifier, so the re-judger parses `"<task>#<index>#<seed>"` as a fallback.

## 4. API Contract

```
POST /api/runs/{id}/rescore
Content-Type: application/json

{
  "reason": "the rubric changed after review",
  "task_set_id": "demo",          // optional, defaults to the source run's
  "task_set_version": "v2"        // optional, must accompany task_set_id
}
```

| Response | Condition |
|---|---|
| 201 + the new run | Re-judged and stored |
| 400 | The reason is empty, or only one of the task-set pair was given |
| 404 | The source run or the judging task set does not exist |
| 409 | The source run is not completed |
| 422 | A trajectory's task is absent from the judging task set |

## 5. Run Identity and History

The new identifier is `<source>-rejudge-<UTC timestamp>`, with a numeric suffix when that identifier is already taken. Two re-judgements in the same second therefore produce two records rather than one overwrite.

The new run copies the source plan, then sets the task set and version to the ones it was judged against, because the plan fingerprint is the statement of "what was scored". The trajectories are copied unchanged. The status is `completed`, and the start and finish times are the re-judgement's own times.

## 6. Failure Semantics

A re-judgement must not let a scoring fix turn a failed execution into a pass. The failure class is recomputed from the trajectory's termination reason:

| Termination | Failure class | Verdict after re-judgement |
|---|---|---|
| `completed` / `declared_infeasible` | completed | the scorer's verdict |
| `agent_error` / `environment_error` | model failure | FAIL, score 0 |
| `step_limit` / `turn_limit` / `budget_exhausted` | budget exhausted | FAIL, score 0 |
| `protocol_violation` | protocol violation | FAIL, score 0 |
| `platform_error` | platform failure | UNSCORED, excluded from the denominator |
| `cancelled` | cancelled | FAIL, excluded from the denominator |

## 7. Limitations

- A trajectory stores the answer and the steps, not a live environment. A J1 state-check scorer that needs an environment snapshot cannot be re-run from the trajectory alone; it must be judged from the recorded observation data or executed as a new run.
- A model judge may return a different grade on a second pass even at temperature 0, because the provider is not guaranteed deterministic. The two runs make the difference visible instead of hiding it.
- The re-judged run keeps the source plan's fingerprint inputs except the task-set version. The comparison endpoint therefore shows them as different plans, which is correct: they were judged under different criteria.
- The copied trajectories keep the usage recorded during execution; the re-judged trial row carries the new judge's cost. The trial row is the authoritative cost of the new judgement, and the trajectory remains the evidence of what the agent did.

## 8. Alternatives Considered

| Alternative | Why rejected |
|---|---|
| Overwrite the source run's verdicts | Destroys the evidence the platform promises to preserve and makes the change uncomparable. |
| Zero-byte edit of the trajectory and re-run the plan | Calls the agent, changes the seeds and produces different trials, so the result is not a re-judgement. |
| A separate judgement table keyed by run and version | Duplicates the run record and forces every report, export and comparison path to merge two stores. A new run is the same idea with no second code path. |
| Skip a trajectory whose task was retired | Silently shrinks the denominator and produces a number that does not describe the run. |

## 9. Verification

Verified by [TR-012 Offline Re-scoring](../Test Report/TR-012-offline-rescoring.md).

## 10. Open Points

- Inline scorer overrides would avoid creating a whole task-set version for a one-scorer correction.
- The comparison endpoint does not yet mark two runs as "same execution, different judgement".

## References

- Feature: [FEAT-012 Offline Re-scoring][feat-012]
- Test Report: [TR-012 Offline Re-scoring][tr-012]
- Related design: [TD-004 M0 Runtime and Quality Gate System][td-004]

[feat-012]: <../Features/execution/offline-rescoring.md>
[tr-012]: <../Test Report/TR-012-offline-rescoring.md>
[td-004]: <./TD-004-m0-runtime-and-quality-gate-system.md>
