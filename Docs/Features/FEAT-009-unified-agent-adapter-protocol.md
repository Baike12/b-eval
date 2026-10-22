# FEAT-009: Unified Agent Adapter Protocol

| Field | Value |
|---|---|
| **ID** | FEAT-009 |
| **Status** | Draft |
| **Owner** | Platform team |
| **Priority** | P0 |
| **Implements** | `F-AGT-01`, `F-AGT-02`, `F-AGT-03`, `F-AGT-04` |
| **Created** | 2026-10-22 |
| **Last Updated** | 2026-10-22 |
| **Source** | Adapter protocol deep-dive, authored with the delivery-document structure |

## Summary

Define a single protocol that any agent under test must satisfy in order to be evaluated by b-eval, plus a self-check that verifies the adapter before evaluation is enabled. The protocol makes it possible to run one agent against SWE-bench, OSWorld and τ²-bench tasks and get scores in a single comparable coordinate system.

## Problem

**Current behaviour:** Every upstream benchmark defines its own agent interface. τ-bench runs a fixed tool-call loop with tool messages injected after each call. VitaBench exposes `generate_next_message`. MCPMark drives a LiteLLM tool loop. AgentCompass exposes `run_task` and expects a `PreparedTask`. WebArena and OSWorld expect a parsed action string or a low-level action tuple.

**Pain:** An agent that already works with one benchmark needs a bespoke integration for each other one, and the behavioural differences between those integrations are not equivalent. A score produced through the VitaBench adapter is not comparable with a score produced through the τ²-bench adapter, so a single number for "how good is this agent" does not exist. Integrators also have no way to tell an adapter bug apart from an agent failure.

**Evidence:** No internal measurement yet — this is a foundational bet. The gap is documented in the open-source survey of agent evaluation projects, where no two of the twelve surveyed frameworks share an agent interface.

## Proposed Solution

Introduce an agent adapter protocol with four required capabilities — start a session, advance one step, decide whether to stop, and report usage — and a task-declared interaction contract that describes the action space, observation shape, termination rules and budget the agent must operate under. Add a self-check suite that runs a small smoke task set against an adapter and reports action-format errors, contract violations and usage-reporting correctness.

The protocol deliberately does not constrain how an agent plans, remembers or reflects. Those are the capabilities under test. It constrains only the boundary: what the agent can do, what it can see, when it stops, and what it spends.

## Scope

### In Scope

- The adapter contract: session start, step advance, stop decision, usage reporting.
- Three adapter shapes: native structured-action adapters, text-action adapters, and black-box CLI adapters.
- The task-side interaction contract: action space, observation shape, termination conditions, budget.
- The self-check suite and the readiness gate that disables evaluation for an adapter that has not passed it.
- Reporting of action-format error rate and contract-violation count as first-class metrics.

### Out of Scope

- Any agent implementation strategy (planning, memory, reflection) — these are measured, not prescribed.
- Exposing or storing an agent's internal reasoning.
- Multi-agent orchestration within a single trial — covered by a separate feature.
- User-simulator behaviour beyond what the interaction contract needs to express.

## Requirements

- **R1** An adapter must be able to start a session given a task contract and an initial observation, and return the first action request.
- **R2** An adapter must be able to advance one step given the latest observation, and return exactly one action expressible in the task's declared action space.
- **R3** An adapter must be able to answer a stop query with either "continue" or "stop" together with a termination reason drawn from a fixed vocabulary.
- **R4** An adapter must report cumulative token usage, model call count, elapsed time and incurred cost at any point during a session.
- **R5** The platform must map between the agent's native output form and the contract's action space without changing the semantics of the action.
- **R6** The platform must record every protocol violation — malformed action, action outside the declared space, termination action bundled with another action, missing usage report — and expose the counts per trial.
- **R7** The platform must offer a self-check that runs an adapter against the smoke task set and reports, per task, the expected and actual action sequence, the action-format error rate, the contract-violation count and usage-reporting correctness.
- **R8** Evaluation must be disabled for an adapter that has not passed the self-check, and the reason must be displayed.

## Acceptance Criteria

- [ ] **AC1** An adapter implementing the four capabilities runs a full smoke task set and every trial terminates with a reason from the fixed vocabulary.
- [ ] **AC2** A malformed action is rejected by the platform, counted as an action-format error, and the trial continues with the error surfaced to the agent.
- [ ] **AC3** A termination action bundled with another action in one response is rejected as a whole, and the rejection is counted as a contract violation.
- [ ] **AC4** A trial that exceeds its declared budget terminates with reason "budget exhausted", and the actual token and cost consumption appear in the trial record.
- [ ] **AC5** The self-check reports the expected-versus-actual action difference for every smoke task that ships with a reference trace.
- [ ] **AC6** Evaluation of an adapter that has not passed the self-check is refused, and the refusal message names the failing check.
- [ ] **AC7** Two different adapters for the same agent produce action sequences that differ only in representation, and their scores on the same task set fall within the trial-to-trial variance.

## Impact

| Area | Effect |
|---|---|
| Users | Agent developers get one integration surface instead of one per benchmark. |
| Existing behaviour | No existing adapter exists; this is a new surface. Upstream benchmarks are consumed through adapters. |
| Operational cost | Self-check adds a smoke run per adapter registration (minutes). No per-trial cost added. |

## Risks and Mitigations

| Risk | Likelihood | Mitigation |
|---|---|---|
| The protocol is too loose to make scores comparable across adapters | Medium | AC7 tests representation-equivalence explicitly; comparison across adapters is gated on that check. |
| The protocol is too tight and excludes useful agents | Medium | Three adapter shapes, including black-box CLI, cover agents that cannot be modified at all. |
| Protocol violations are misattributed to agent capability | Medium | Violations are counted and reported separately from task outcomes, and the failure-attribution layer places them in the "integration problem" band. |

## Rollout

The protocol ships behind the M0 milestone. Adapters are registered one at a time, each passing the self-check before evaluation is enabled for it. The first upstream benchmark consumed is SWE-bench, chosen because its judgement is state-based and therefore does not depend on the agent's internal message format.

## References

- Tech Design: [TD-001 Unified Agent Adapter Protocol][td-001]
- Test Report: [TR-001 Unified Agent Adapter Protocol][tr-001]

## Open Questions

| # | Question | Owner | Blocking? |
|---|---|---|---|
| 1 | Does the termination reason vocabulary need a "cancelled by operator" value, or is operator cancellation handled entirely outside the protocol? | Platform team | No |
| 2 | Should cost reporting be mandatory, given some model providers do not expose a price table? | Platform team | No |

[td-001]: <../Tech Design/TD-001-unified-agent-adapter-protocol.md>
[tr-001]: <../Test Report/TR-001-unified-agent-adapter-protocol.md>
