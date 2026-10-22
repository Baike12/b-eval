# TR-001: Unified Agent Adapter Protocol — Verification Report

| Field | Value |
|---|---|
| **ID** | TR-001 |
| **Verifies** | [FEAT-001 Unified Agent Adapter Protocol](../Features/FEAT-001-unified-agent-adapter-protocol.md) |
| **Technical Design** | [TD-001 Unified Agent Adapter Protocol](<../Tech Design/TD-001-unified-agent-adapter-protocol.md>) |
| **Environment** | Smoke task set: 10 tasks across 6 domains (web, desktop, code, tool, service, multi-agent) |
| **Executed** | 2026-10-22 |
| **Result** | **Pass** — all acceptance criteria met |

## 1. Scope

Verifies the four adapter capabilities, the three adapter shapes, the protocol-violation accounting, the self-check suite and the readiness gate described in [TD-001](<../Tech Design/TD-001-unified-agent-adapter-protocol.md>).

Out of scope: scoring correctness (covered by the judgement layer), multi-agent orchestration, user-simulator fidelity.

## 2. Test Setup

| Item | Value |
|---|---|
| Smoke task set | 10 tasks, 6 domains, 2 tasks per domain |
| Reference-trace tasks | 6 of 10 ship a reference action sequence |
| Adapters under test | 3 shapes — native structured action, text action, black-box CLI |
| Seeds | 5 per task (fixed set, replayable) |
| Budget per trial | 30 steps, 200k tokens, 300 s wall clock |

## 3. Acceptance Criteria Verification

| AC | Criterion | Method | Result |
|---|---|---|---|
| **AC1** | Adapter runs the full smoke set; every trial ends with a reason from the closed vocabulary | 3 adapters × 10 tasks × 5 seeds = 150 trials | **Pass** — 150/150 trials ended with a valid reason; 0 unknown values |
| **AC2** | Malformed action rejected, counted, trial continues with the error surfaced | Injected malformed action at step 3 of 10 trials per adapter | **Pass** — 30/30 rejections; error delivered to the agent in the next observation; trial continued to a valid terminal reason |
| **AC3** | Bundled termination action rejected as a package | Injected a response containing a terminate action plus a write action, 10 trials per adapter | **Pass** — 30/30 packages rejected as a whole; side effect of the bundled write action did not occur in the environment (verified by post-trial state snapshot); each rejection counted as one contract violation |
| **AC4** | Budget exhaustion produces reason "budget_exhausted" and actual consumption is recorded | 10 trials per adapter with a token limit deliberately set below expected usage | **Pass** — 30/30 trials terminated with `budget_exhausted`; token and cost present in the trial record in 30/30 |
| **AC5** | Self-check reports expected-versus-actual action difference for tasks with a reference trace | Self-check run on all 3 adapters | **Pass** — diff produced for 18 of 18 reference-trace task instances (6 tasks × 3 adapters) |
| **AC6** | Evaluation refused for an adapter that has not passed the self-check; refusal names the failing check | Registered a deliberately malformed adapter, attempted an evaluation | **Pass** — evaluation refused; refusal message named "action-format error rate exceeds threshold"; attempt logged |
| **AC7** | Two adapters for the same agent differ only in representation; scores fall within trial-to-trial variance | Same agent wrapped as native structured action and as text action; 10 tasks × 20 seeds | **Pass** — action sequences identical after mapping in 200/200 trials; success-rate difference 0.00, which is within the ±0.02 trial-to-trial band measured on the same seed set |

## 4. Protocol Accounting Tests

| Test | Expectation | Result |
|---|---|---|
| Action outside the declared action space | Counted as a violation, error returned to the agent, trial continues | Pass |
| Empty response | Counted as a malformed action | Pass |
| Missing usage report after a model-calling step | Counted; consumption estimated from recorded messages and flagged `estimated: true` | Pass |
| Repeated identical violation | Trial terminates with `protocol_violation` after the configured consecutive-violation threshold | Pass — terminated at the threshold in 3/3 injected trials |
| Termination reason outside the vocabulary | Trial rejected as an invalid trial; not counted as completed | Pass |

## 5. Adapter Shape Tests

| Shape | Test | Result |
|---|---|---|
| Native structured action | Action passes through unchanged; action schema published to the agent | Pass |
| Text action | Valid text parses; malformed text produces an action-format error and a parse-failure message returned to the agent | Pass |
| Black-box CLI | Process runs in the task environment; terminal interaction and file output captured; produced action sequence matches the reference trace on 4 of 4 code-domain reference tasks | Pass — 4/4 |

## 6. Violation Counter Consistency

Violation counts recorded by the protocol layer were compared against an independent recount from the raw trial transcript for all 150 trials.

- Mismatches: **0 / 150**.
- Counts are stable across a replay of the same seed set: **150 / 150 identical**.

## 7. Self-Check Output Verification

| Output | Expectation | Result |
|---|---|---|
| Action sequence diff | Present for every reference-trace task | Pass — 18/18 |
| Action-format error rate | Present, denominator = total responses | Pass |
| Contract-violation count | Broken down by violation type | Pass |
| Usage-reporting correctness | Ratio of matching reported vs observed usage | Pass — 1.00 across 150/150 trials |
| Terminal-reason validity | Share of valid reasons | Pass — 1.00 |

## 8. Readiness Gate Verification

| Test | Expectation | Result |
|---|---|---|
| Unregistered adapter | Evaluation refused | Pass |
| Registered adapter that has not passed self-check | Evaluation refused, failing check named | Pass |
| Adapter that passed self-check | Evaluation allowed | Pass |
| Operator override | Allowed and recorded in the audit trail | Pass — override recorded with operator identity and timestamp |

## 9. Defects Found During Verification

| # | Defect | Severity | Disposition |
|---|---|---|---|
| 1 | In the black-box CLI shape, the action sequence of a task whose agent produced no file output was recorded as empty rather than as a protocol violation | S3 | Fixed under [FIX-002][fix-002]; re-verified |
| 2 | Usage-reporting correctness was computed against a rounded token count, producing a 0.97 ratio on trials with large token counts | S4 | Fixed in the same change as the comparison logic; re-verified |

## 10. Conclusion

All seven acceptance criteria pass, with 150/150 trials valid and zero accounting mismatches. The protocol is fit for use as the agent entry surface. The black-box CLI shape's action-observation model remains an open point recorded in [TD-001 §10](<../Tech Design/TD-001-unified-agent-adapter-protocol.md>) and is revisited when the first upstream benchmark is consumed.

## References

- Feature: [FEAT-001 Unified Agent Adapter Protocol][feat-001]
- Technical Design: [TD-001 Unified Agent Adapter Protocol][td-001]

[feat-001]: <../Features/FEAT-001-unified-agent-adapter-protocol.md>
[td-001]: <../Tech Design/TD-001-unified-agent-adapter-protocol.md>
[fix-002]: <../Fix/FIX-002-blackbox-cli-empty-action-sequence-not-flagged.md>
