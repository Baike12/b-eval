# FIX-002: Black-box CLI adapter recorded an empty action sequence instead of a protocol violation

| Field | Value |
|---|---|
| **ID** | FIX-002 |
| **Status** | Fixed |
| **Owner** | Platform team |
| **Severity** | S3 Medium |
| **Priority** | P1 |
| **Affects** | `F-AGT-02`, `F-AGT-03` |
| **Detected in** | TR-001 verification run, 2026-10-22 |
| **Created** | 2026-10-22 |
| **Last Updated** | 2026-10-22 |

## Summary

In the black-box CLI adapter shape, a trial in which the agent produced no observable output was recorded with an empty action sequence and no violation flag. The trial therefore looked like a legitimate zero-action run instead of a protocol failure, and such trials were silently counted as completed.

## Severity and Impact

**Technical impact:** The protocol layer treated "the agent produced nothing observable" as a valid trial with zero actions. The trial reached the end of the task with a terminal reason that did not reflect a protocol failure, so failure attribution placed it in the wrong band.

**Blast radius:** Affects every trial using the black-box CLI shape in which the agent exits without observable output. In the TR-001 run this was 1 trial out of 50 black-box trials (2%); in the earlier SWE-bench consumption spike it reached 4 of 96 trials (4.2%). Those 4 trials were reported as completed without action, which overstated the adapter's task success rate by 1 task-equivalent in that batch.

**Detection:** Found during the TR-001 verification run, test case "black-box CLI, agent produces no file output".

## Symptom

The trial record showed `action_sequence: []`, `protocol_violations: []` and terminal reason `agent_error`. The zero-length action sequence was accepted without a matching violation record.

## Reproduction

**Environment:** Black-box CLI adapter, smoke task set, seed set S1.

**Preconditions:** A task whose agent terminates the session without producing observable output.

**Steps:**

1. Register the black-box CLI adapter.
2. Run the smoke task set.
3. Open the trial record for the affected task.
4. Inspect `action_sequence` and `protocol_violations`.

**Expected:** The trial records a protocol violation of type "empty response" and the terminal reason reflects the protocol failure.

**Actual:** `action_sequence` is empty, `protocol_violations` is empty, terminal reason is `agent_error`.

## Root Cause

**Cause:** The black-box shape synthesises actions from observed interaction. The synthesis step returned an empty list when nothing was observed, and the empty list was passed to the violation check. The violation check only evaluated actions that were present, so an empty list produced no violations. The two shapes — native and text — build the action list from the agent's own output, where an empty response is a distinguishable event, so the defect existed only in the black-box path.

**Failure class:** Logic error

**Why it was not caught:** The violation check was specified against a non-empty action list, so the empty-list case was never in scope of the check itself. The self-check's usage-reporting and terminal-reason checks both passed, because the trial still ended with a valid reason — the defect was in the accounting, not in the trial's termination.

## Fix

**Change:** The black-box synthesis step now distinguishes "the agent produced nothing observable" from "the agent produced no action in this step". The former raises an "empty response" protocol violation before the action list reaches the violation check; the latter remains a valid zero-action step. The violation check additionally rejects any trial whose action list is empty.

**Rationale:** The defect was a missing case in the violation check, not a wrong check. Fixing the check makes the defect unreachable regardless of which adapter shape produced the list, so the three shapes share one invariant.

**Guardrail added:** A trial with an empty action sequence and no matching "empty response" violation is now rejected at the point the trial is closed, and the rejection is recorded as a platform error rather than a completed trial.

## Verification

| Check | Method | Result |
|---|---|---|
| Original reproduction no longer reproduces | Re-ran the black-box smoke set with seed set S1 | **Pass** — the affected trial now carries an "empty response" violation and terminal reason `protocol_violation` |
| Regression test added | Test `blackbox_cli_empty_output_raises_empty_response_violation` | **Pass** |
| Adjacent behaviour unaffected | Native and text shapes re-ran over the full smoke set, 150 trials | **Pass** — 150/150 trials unchanged in violation counts and terminal reasons; 0 new violations introduced |
| Blast radius re-scored | The 4 affected SWE-bench-spike trials re-scored under the fixed logic | **Pass** — 4 trials reclassified from "completed" to "protocol violation"; the batch's reported success rate corrected downward by 1 task-equivalent |

## Regression Prevention

- **Test added:** `blackbox_cli_empty_output_raises_empty_response_violation`, in the adapter self-check suite, so it runs on every adapter registration.
- **Run gate:** The self-check suite runs on every adapter registration and on every change to the adapter protocol layer.
- **Detection change:** The trial-close validation now rejects an empty action sequence without a matching violation, and logs a platform error when it does.

## References

- Test Report: [TR-002 Black-box CLI Empty Action Sequence Fix][tr-002]
- Related Feature: [FEAT-009 Unified Agent Adapter Protocol][feat-009]

## Open Questions

| # | Question | Owner | Blocking? |
|---|---|---|---|
| 1 | Should a black-box agent that legitimately produces no observable output on a given step be distinguishable from one that fails to start? | Platform team | No |

[tr-002]: <../Test Report/TR-002-blackbox-cli-empty-action-sequence-fix.md>
[feat-009]: <../Features/FEAT-009-unified-agent-adapter-protocol.md>
