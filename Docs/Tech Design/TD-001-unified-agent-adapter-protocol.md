# TD-001: Unified Agent Adapter Protocol

| Field | Value |
|---|---|
| **ID** | TD-001 |
| **Status** | Draft |
| **Implements** | [FEAT-009 Unified Agent Adapter Protocol](../Features/agent-integration/unified-agent-adapter-protocol.md) |
| **Owner** | Platform team |
| **Created** | 2026-10-22 |
| **Last Updated** | 2026-10-22 |

> This document specifies *how* the adapter protocol is built. The *what* and *why* live in the Feature document it implements.

## 1. Problem Restatement

Eleven upstream agent-evaluation frameworks were surveyed. Their agent interfaces are mutually incompatible, and three of them (τ-bench, MCPMark, WebArena) disagree even on where a "step" begins. A single agent therefore cannot be scored on two benchmarks without two adapters, and the two scores are not on the same scale.

## 2. Design Goals

| Goal | Consequence for the design |
|---|---|
| One agent, many benchmarks | The protocol is defined against the *task*, not against any framework. |
| Scores comparable across adapters | The protocol fixes the observation a step receives and the action a step returns. |
| Integration bugs separable from capability failures | Every protocol interaction is recorded and counted independently of task outcome. |
| Agents that cannot be modified still supported | A black-box CLI shape exists alongside in-process shapes. |
| Budget observable | Usage reporting is part of the protocol, not an optional extra. |

## 3. Non-Goals

- The protocol does not prescribe an agent's internal architecture.
- The protocol does not carry a model-specific message schema; it carries an *action* schema declared by the task.
- The protocol does not define scoring. Scoring belongs to the judgement layer.

## 4. Protocol Shape

### 4.1 Capabilities

An adapter exposes four capabilities:

| Capability | Direction | Purpose |
|---|---|---|
| `start_session` | platform → adapter | Hand over the task contract and the first observation; obtain the first agent state. |
| `advance` | platform → adapter | Hand over the latest observation; obtain exactly one action. |
| `should_stop` | platform → adapter | Ask whether the session may end; obtain "continue" or "stop" plus a reason. |
| `report_usage` | adapter → platform | Report cumulative tokens, model calls, elapsed time and cost. |

The adapter additionally *receives* the observations and *returns* the action. No other callback is required.

### 4.2 Interaction Contract (task side)

The task declares, and the platform enforces:

| Element | Content |
|---|---|
| **Action space** | Named actions, each with a parameter schema and a semantic description. |
| **Action classes** | Read / Write / Reason / Terminate. |
| **State effect** | Per action, a separate boolean: does executing it change environment state. Independent of the class, because a semantically "write" action may not mutate state. |
| **Observation shape** | Structured / Text / Visual / Mixed, plus the trimming rules that apply. |
| **Termination conditions** | Agent completion, agent-declared-infeasible, step limit, budget exhausted, protocol violation, environment error. |
| **Budget** | Step, turn, token, duration, cost and error-count limits. |

**Termination actions must occupy a turn alone.** A response containing a termination action together with any other action is rejected as a package, and the rejection is counted as a contract violation. This prevents a terminating call from carrying side effects that the trial never observes.

**State-effect flag drives replay.** During judgement replay, only actions whose state-effect flag is true are executed. Read actions are skipped, so an agent that queries more often does not change the outcome, and an action the agent hallucinated is treated as a no-op — matching what actually happened.

### 4.3 Termination Reason Vocabulary

Fixed set, closed:

`completed` · `declared_infeasible` · `step_limit` · `turn_limit` · `budget_exhausted` · `protocol_violation` · `agent_error` · `environment_error` · `platform_error` · `cancelled`

Every trial ends with exactly one reason. The reason is the primary input to failure attribution.

### 4.4 Adapter Shapes

| Shape | Agent form | Platform responsibilities |
|---|---|---|
| **Native structured action** | Agent already emits function calls or a structured action object | Publish the action schema to the agent; translate the agent's action into the contract's action. |
| **Text action** | Agent emits free text in a documented action format | Parse the text into an action; a parse failure produces an action-format error and an error message returned to the agent. |
| **Black-box CLI** | Agent is an existing executable driven in a controlled environment | Run the process in the task environment; capture its terminal interaction and file output; synthesise actions from the observable interaction. |

### 4.5 Action Mapping

For the native shape, mapping is identity: the action the agent produces is the contract action. For the text shape, mapping is a parse step whose failure mode is counted. For the black-box shape, mapping is observation-only: the platform does not constrain what the agent does inside the environment, and the contract's action space applies only to what the platform can observe and judge.

### 4.6 Protocol Violations

| Violation | Detection | Consequence |
|---|---|---|
| Malformed action | Action does not parse into the contract schema | Counted; error returned to the agent; trial continues. |
| Action outside the declared space | Action name not in the task's action space | Counted; error returned to the agent; trial continues. |
| Bundled termination action | Termination action appears with others in one response | Counted; the whole package is rejected. |
| Empty response | No action and no content | Counted; treated as a malformed action. |
| Missing usage report | No usage reported after a step that consumed model calls | Counted; the step's consumption is estimated from the recorded messages and flagged as estimated. |
| Repeated invalid action | Same violation more than `max_consecutive_violations` times in a row | Trial terminates with `protocol_violation`. |

## 5. Self-Check

The self-check runs the smoke task set (at most 10 tasks per domain) against a registered adapter and reports:

| Output | Definition |
|---|---|
| Action sequence diff | Expected versus actual, per task, for tasks that ship a reference trace. |
| Action-format error rate | Malformed-action count over total responses. |
| Contract-violation count | Count per violation type. |
| Usage-reporting correctness | Trials where reported usage matches the platform's own observation of consumption, expressed as a ratio. |
| Terminal-reason validity | Share of trials ending with a reason from the closed vocabulary. |

**Readiness gate.** Evaluation is refused for an adapter that has not passed the self-check. The refusal names the failing check. The gate can be overridden by a platform operator, and the override is recorded in the adapter's audit trail.

## 6. Self-Check Execution Order

```
1. Register adapter  →  2. Run self-check  →  3. Inspect report
        ↓
4. Passed?  ── yes ──→  Evaluation enabled
        ↓ no
5. Operator override (audited) or fix adapter and re-run
```

## 7. Data Recorded Per Trial

Protocol-level records, kept separate from task-outcome records:

`protocol_violations[]` (type, step index, raw payload) · `action_format_errors` · `terminal_reason` · `reported_usage` vs `observed_usage` · `adapter_shape` · `adapter_version` · `self_check_status`

## 8. Alternatives Considered

| Alternative | Why rejected |
|---|---|
| Adopt one upstream framework's agent interface | It is coupled to that framework's message and task model; using it makes every other framework an adapter case anyway. |
| Require every agent to be wrapped in the platform's own agent runtime | Excludes agents that cannot be modified, which is a large and important class. |
| Let the task define a bespoke protocol per task | Makes cross-benchmark comparability impossible, which is the primary goal of the protocol. |
| Ignore protocol violations and score only task outcome | Makes integration bugs indistinguishable from capability failures, and lets a broken adapter produce a plausible-looking score. |

## 9. Verification

Verified by [TR-001 Unified Agent Adapter Protocol][tr-001].

## 10. Open Points

- The `max_consecutive_violations` threshold is task-configurable; the default value needs load data before it is fixed.
- The black-box shape's action-observation model needs a second pass once the first upstream benchmark (SWE-bench) is consumed, because SWE-bench exposes no observable action stream.

## References

- Feature: [FEAT-009 Unified Agent Adapter Protocol][feat-009]
- Test Report: [TR-001 Unified Agent Adapter Protocol][tr-001]

[feat-009]: <../Features/agent-integration/unified-agent-adapter-protocol.md>
[tr-001]: <../Test Report/TR-001-unified-agent-adapter-protocol.md>
