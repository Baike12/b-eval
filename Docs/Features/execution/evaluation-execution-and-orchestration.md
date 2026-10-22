# FEAT-003: Evaluation Execution and Orchestration

| Field | Value |
|---|---|
| **ID** | FEAT-003 |
| **Status** | Draft |
| **Owner** | Platform team |
| **Priority** | P0 |
| **Implements** | `F-RUN-01`, `F-RUN-02`, `F-RUN-03`, `F-RUN-04`, `F-RUN-05`, `F-RUN-06`, `F-RUN-07`, `F-RUN-08`, `F-AGT-01`, `F-AGT-02`, `F-AGT-03`, `F-AGT-04` |
| **Created** | 2026-10-22 |
| **Last Updated** | 2026-10-22 |
| **Source** | Functional design chapter 03 — Evaluation Execution and Orchestration |

## Summary

The platform gains a single execution and orchestration layer that drives every evaluation through a fixed ten-step flow, from task set selection to archive / publish, with the judgement phase decoupled from the execution phase. Task authors, Agent developers and evaluation organisers use it to run Trials against Agents that are integrated through a common adapter protocol, under declared interaction contracts, budgets and concurrency controls. The single most important outcome is that an evaluation can be judged — or re-judged with new scoring criteria, a new judge model or manual review — without calling the Agent again.

## Problem

**Current behaviour:** Not specified in the source design. The source specifies the target platform behaviour only: a ten-step evaluation flow, an adapter protocol for the Agent under test, an interaction contract, budgets and limits, concurrency and rate limiting, failure classification, resume semantics, trajectory records and the Evaluation Plan.

**Pain:**

- Without a common adapter protocol, every Agent integration would define its own input and output shape, so runs would not be comparable. The source therefore keeps the protocol to "action + observation + usage" and explicitly leaves the Agent's internal planning, memory and reflection loop — the capabilities actually under test — out of the protocol.
- Without decoupling the judgement phase from the execution phase, switching the scoring criteria, switching the judge model or adding manual review would require calling the Agent again.
- Without an adapter self-check gate, integration bugs would be recorded as weak model capability.
- Without task-declared trimming rules decided by the task rather than unilaterally by the platform, the observation budget of different tasks would not be comparable.
- Without an explicit exclusion of platform failures from the denominator, an environment outage or a platform bug would be scored as if it were a model defect.
- Without a circuit-breaking threshold per task, an environment outage would burn through the whole batch budget.

**Evidence:** Not specified in the source design.

## Proposed Solution

The platform provides a unified evaluation execution and orchestration layer. Every evaluation is described by a single Evaluation Plan, executed as a fixed ten-step flow over Trials running on one isolated environment instance each, and bounded by task- and plan-declared budgets. Agents are integrated through a four-capability adapter protocol in one of three adapter shapes, and every Trial records a trajectory that carries the judgement, so scoring can be re-run offline without the Agent. Failures are classified into six categories with explicit denominator and re-run rules, and interrupted runs can be resumed because the resume unit is (task ID, Trial index, random seed) and the configuration must be unchanged.

## Scope

### In Scope

- The ten-step evaluation flow, from selecting the task set and version to archiving or publishing the result, with the judgement phase decoupled from the execution phase.
- The adapter protocol for the Agent under test: four protocol capabilities, three adapter shapes (native action, text action, black-box CLI), and the adapter self-check suite.
- The interaction contract: action space, observation shapes and observation trimming, termination conditions, and the optional user simulator.
- Budgets and limits per Trial (step, turn, token, duration, cost, call count, error count), the pre-run estimate, and budget circuit-breaking.
- Concurrency and rate limiting controls, the separation of throttling from retries, and the per-task circuit-breaking threshold.
- Failure classification into six categories, with the denominator and re-run rules for each and the platform-failure-rate threshold.
- Resume and idempotency keyed on (task ID, Trial index, random seed) with plan-fingerprint consistency checks.
- Trajectory and observability records per Trial, and the offline re-scoring, trajectory comparison and export capabilities built on them.
- The Evaluation Plan (Plan) and the plan fingerprint derived from it.

### Out of Scope

- **The Agent's internal planning, memory and reflection strategy** — the source states that the protocol does not specify these, because they are the capabilities under test, not protocol content.
- **Exposure of the Agent's internal chain of thought** — the protocol only requires exposing "action + observation + usage".
- **A particular model-invocation style** — the adapter may be "calling the model directly", "wrapping a CLI tool", "wrapping an IDE Agent" or "wrapping an MCP client".
- Any further exclusions beyond the adapter-protocol non-goals above: not specified in the source design.

## Requirements

- **R1** Every evaluation runs through the fixed ten-step flow below, and the judgement phase (step ⑦) is decoupled from the execution phase (step ⑤).

  **3.1 The Complete Flow of One Evaluation**

    ```
    ① Select task set and version  →  ② Configure evaluation plan (Plan)  →  ③ Estimate cost
            ↓
    ④ Prepare environment (one isolated instance per Trial)  →  ⑤ Execute interaction (Agent ↔ environment/user)
            ↓
    ⑥ Collect trajectory and usage  →  ⑦ Judge (can be re-run offline)  →  ⑧ Aggregate metrics
            ↓
    ⑨ Generate report and comparison  →  ⑩ Archive / publish
    ```

  **Key design: step ⑦ and step ⑤ are decoupled.** Once the trajectory is persisted, judgement can be re-run independently of execution — switching the scoring criteria, switching the judge model, or adding manual review never requires calling the Agent again. This is one of the platform's most important engineering capabilities.

- **R2** An Agent adapter must implement the four protocol capabilities below, and must report the termination reason faithfully.

  **3.2 Adapter Protocol for the Agent under Test**

  **3.2.1 Protocol Definition**

  An Agent adapter must implement the following four capabilities:

  | Capability | Semantics | Notes |
  |---|---|---|
  | **Start session** | Receive the task contract and the initial observation, and open a session | The adapter is responsible for constructing the first-turn input (the first-turn shape varies enormously across Agents) |
  | **Advance one step** | Receive the latest observation and produce the next action | The action shape is defined by the contract; the adapter maps between the Agent's native output and the contract action |
  | **Determine completion** | When the platform asks whether the run may end, the adapter returns either "continue" or "end + termination reason" | The termination reason is key data for error attribution and must be reported faithfully by the adapter |
  | **Report usage** | Report the cumulative token count, call count, elapsed time and cost spent at any time | Used for budget circuit-breaking and cost accounting |

  **What the protocol explicitly does not do:**

    - It does not specify how the Agent plans internally, whether it uses memory, or whether it has a reflection loop — those are the capabilities under test, not protocol content.
    - It does not require the Agent to expose its internal chain of thought; it only requires exposing "action + observation + usage".
    - It does not require the Agent to use a particular model-invocation style. The adapter may be "calling the model directly", "wrapping a CLI tool", "wrapping an IDE Agent" or "wrapping an MCP client".

- **R3** The platform must support the three adapter shapes below and perform the platform-side duties each shape requires.

  **3.2.2 Three Adapter Shapes**

  | Shape | Applies to | What the platform must do |
  |---|---|---|
  | **Native action adaptation** | The Agent natively supports function calls / structured actions | Provide the action schema so that the Agent connects to the environment directly; the action space is given by the task contract |
  | **Text action adaptation** | The Agent only outputs natural-language actions (e.g. `Action: X with Action Input: Y`) | The platform parses the text actions; parse failures count toward the "action format error rate" and are handled according to the contract rules |
  | **Black-box CLI adaptation** | The Agent is an off-the-shelf tool (e.g. a coding Agent in a terminal) | The platform runs its process in a controlled environment and collects its terminal interactions and file outputs |

- **R4** The platform must provide an adapter self-check suite, and must disable the evaluation capability when the self-check fails.

  **3.2.3 Adapter Self-Check**

  The platform provides an adapter self-check suite: task authors and Agent developers use the same set of sample tasks (≤ 10 items per domain, see [chapter 02][feat-002]) to verify that the integration is correct.

  The self-check outputs:

    - The difference between the expected action sequence and the actual action sequence of each sample task (provided only when the sample task ships with a reference trajectory).
    - The action format error rate and the number of contract violations.
    - Whether usage reporting is correct.
    - **When the self-check fails, the evaluation capability must be disabled** — so that integration bugs are never recorded as weak model capability.

- **R5** Each task declares the interaction shape it expects, and the platform drives execution accordingly; a task must declare its action space, and a terminating action must occupy a turn exclusively.

  **3.3 Interaction Contract**

  Each task declares the interaction shape it expects, and the platform drives execution accordingly.

  **3.3.1 Action Space**

    - A task declares an **action set**; each action includes: name, parameter structure, semantic description, whether it changes environment state, and behaviour on failure.
    - Actions fall into four categories: **read** (query, no state change), **write** (changes state), **reason** (pure internal computation), **terminate** (e.g. escalate to a human, declare completion).
    - **"Whether it changes state" is a switch independent of the four categories**, because there exist actions that are "semantically a write but actually do not change state" (see τ²-bench's separation of `mutates_state` from `ToolType`). This switch determines which actions are executed when the trajectory is replayed during the judgement phase.
    - A terminating action **must occupy a turn exclusively**: if a single response contains both a terminating action and other actions, the whole package is rejected and counted as a protocol violation (see τ²-bench's rule that the terminating tool owns the turn).

- **R6** A task must declare the observation shapes it expects, and trimming rules must be declared by the task and must not be decided unilaterally by the platform.

  **3.3.2 Observation Shapes**

  A task declares what the Agent can see at each step:

  | Shape | Description |
  |---|---|
  | Structured observation | Tool return values (JSON), state query results |
  | Text observation | Page text, terminal output, file content |
  | Visual observation | Screenshots, accessibility trees and combinations of them |
  | Mixed observation | Text and visual given together |

  **Observation trimming capability** (to keep long context from drowning out the task itself): viewport filtering, token-limit truncation, element filtering, on-demand expansion. Trimming rules are declared by the task and must not be decided unilaterally by the platform — otherwise the observation budget of different tasks would not be comparable.

- **R7** A task must declare its termination conditions, and the termination reason must be recorded.

  **3.3.3 Termination Conditions**

  A task declares its termination conditions; the platform and the adapter enforce them together:

  | Termination type | Triggered by | Semantics |
  |---|---|---|
  | Normal completion | Agent | The Agent proactively declares the task complete |
  | Declared infeasible | Agent | The Agent declares the task cannot be completed (correct only when the task allows it) |
  | Step / turn exhaustion | Platform | The step limit declared by the task is reached |
  | Budget exhaustion | Platform | The token / duration / cost limit is reached |
  | Protocol violation | Platform | Consecutive action format errors, a terminating action not owning its turn, etc. |
  | Environment failure | Platform | The environment side errors out or becomes unavailable |

  **The termination reason must be recorded**; it is the primary evidence for error attribution.

- **R8** A task must declare whether there is a human user and that user's persona, and the simulator must not fabricate information the task did not provide.

  **3.3.4 User Simulator (Optional)**

  Some tasks (e.g. business processes, requirement clarification) require a simulated user to take part. The task declares "whether there is a human user" along with that user's persona.

    - The **user persona sources** are separated into three non-overlapping layers: general behavioural guidelines / in-task persona / run-time override.
    - The **information the user must be told** is specified by the task and defaults to "user intent + information already known to the user"; it excludes scoring conditions, the tool list and the reference solution.
    - The **user's information disclosure strategy**: by default "disclose step by step on demand", giving only the information the current step needs; a task may declare "say everything in the first turn".
    - The **user may change their mind**: conditional branches are supported in task instructions (e.g. "if a certain option is missing, you will pick another").
    - **End signal**: the user emits an explicit end marker instead of relying on the model to decide on its own.
    - The **user must not fabricate** information the task did not provide; out-of-bounds simulator behaviour (fabrication, leaking the answer, refusing to answer) is recorded and counted toward the "simulator noise" metric.

- **R9** The platform must impose the budgets below on each Trial, must estimate the cost before the run, and must terminate the Trial when a budget is reached.

  **3.4 Budgets and Limits**

  The platform imposes multiple budgets on each Trial; the first one reached terminates the run:

  | Budget item | Defined by | Description |
  |---|---|---|
  | Step limit | Task declaration | Number of actions the Agent may execute |
  | Turn limit | Task declaration | Number of dialogue turns |
  | Token limit | Task or plan declaration | Total input + output tokens; supports expression forms (e.g. computed as a proportion of the input volume) |
  | Duration limit | Task or plan declaration | Wall-clock duration |
  | Cost limit | Plan declaration | Cost converted via the model price table |
  | Call count limit | Plan declaration | Number of model calls |
  | Error count limit | Task declaration | Number of consecutive tool / environment errors |

  **Pre-run estimate (F-RUN-02):** when the plan is submitted, the platform uses task scale and historical run data to give the **token volume range, cost range and duration range** of this evaluation. The estimate is based on measured data from previous runs of the same task set; when the sample is insufficient it gives a conservative upper bound and marks it "low estimate confidence".

  **Circuit-breaking behaviour:** reaching any budget → terminate the Trial, judge it as "budget exhausted", and count it in the denominator or exclude it according to the rules declared by the task (counted in the denominator by default, because budget exhaustion reflects an Agent efficiency problem; a task may declare it as an exclusion).

- **R10** The platform must apply the concurrency and rate limiting controls below, must not retry model-side logic errors, and must suspend the remaining Trials of a task that fails consecutively up to the threshold.

  **3.5 Concurrency and Rate Limiting**

  | Control | Semantics |
  |---|---|
  | **Maximum concurrent Trials** | Upper bound on the number of tasks running at the same time |
  | **Maximum concurrent environment instances** | An independent upper bound under environment resource constraints |
  | **Model call concurrency** | Divided into a static limit and an adaptive one; in adaptive mode it slows down automatically when the server throttles and speeds up automatically when the server is healthy |
  | **Intra-task concurrency** | Number of parallel sub-actions allowed within a single task |

  **Separation of responsibilities between rate limiting and retries:**

    - **Model-side throttling** (429, insufficient quota) → back off and retry; a retry counts toward the budget's "call count" but not toward "effective steps".
    - **Transient environment-side errors** → retry as above.
    - **Model-side logic errors** (malformed output, hallucinated tool names) → **must not be retried**; the step fails directly and the error is fed back to the Agent (this is itself a capability under test).

  **Circuit-breaking threshold:** when the same task fails consecutively up to the threshold → the remaining Trials of that task are suspended and an alert is raised, so that an environment outage does not burn through the whole batch budget.

- **R11** The platform must classify every incomplete Trial into one of the six failure categories below, must never count platform failures in the denominator, and must report the platform failure rate separately.

  **3.6 Failure Classification**

  The platform classifies the reasons why "a Trial did not run to completion" into six categories, each of which enters a different statistical treatment:

  | Category | Meaning | Counted in denominator | Re-run |
  |---|---|---|---|
  | **Completed** | Ran to completion and was judged | ✔ | — |
  | **Model failure** | The Agent crashed, the output is unparseable, hallucinated tools, timeout | ✔ | Re-runnable (with a different seed) |
  | **Platform failure** | The environment does not start, the scorer misbehaves, an error in the platform itself | ✘ (excluded, counted separately) | ✔ Automatically retried |
  | **Budget exhausted** | A limit was reached | ✔ (by default) | ✘ |
  | **Protocol violation** | Persistent action format errors, the contract was broken | ✔ | ✘ |
  | **Cancelled** | Manually aborted, or cancelled by the evaluation organiser | ✘ | Decided by the initiator |

  **Key constraint (F-RUN-06): platform failures must never be counted in the denominator, and the report must separately list "the platform failure rate of this evaluation".** If the platform failure rate exceeds the threshold (5% by default), the whole report is marked "insufficient confidence" and does not enter the leaderboard.

- **R12** The resume unit must be (task ID, Trial index, random seed), and the configuration must be identical when resuming.

  **3.7 Resume and Idempotency**

  **Resume unit = (task ID, Trial index, random seed).**

    - The platform records completed units by this triple and automatically skips completed units after a restart.
    - **The configuration must be identical when resuming**: if the task set version, Agent version, scoring configuration or budget changes in any respect, the platform refuses to resume and reports "this is a new evaluation". The decision is based on the plan fingerprint.
    - **A completed judgement can be re-run on its own** (without changing the trajectory) in order to switch the scoring criteria; the re-run result is recorded as a new judgement version, and the historical judgement is retained for comparison.

- **R13** Each Trial must record the trajectory categories below, and the trajectory must support offline re-scoring, trajectory comparison and export.

  **3.8 Trajectory and Observability Records**

  Each Trial records the following:

  | Category | Content |
  |---|---|
  | **Task context** | Task ID and version, initial state identifier, capability tags, budget |
  | **Interaction sequence** | The action, observation, duration and success flag of every step |
  | **Termination info** | Termination type, termination reason, actual step count |
  | **Usage** | Input / output tokens, model call count, tool call count, cost, end-to-end duration |
  | **Judgement** | The result of each scorer, the combined result, judgement evidence, scoring duration |
  | **Errors** | Error messages and stack traces of each step, environment anomalies |
  | **Environment evidence** | Environment snapshot, screenshots, logs, produced files (optional per task type) |
  | **Plan fingerprint** | Task set version, Agent identifier and version, scoring configuration, budget, platform version |

  **The trajectory is the platform's core asset:**

    - Every number in a report can be drilled down to a specific Trial.
    - It supports **offline re-scoring**: re-judging old trajectories with a new scorer, without calling the Agent.
    - It supports **trajectory comparison**: the trajectories of two runs can be diffed (actions added and removed, parameter drift, changes in cost / latency / outcome).
    - It supports **export**: trajectories can be exported to a common format for external tools to consume.

- **R14** An evaluation must be described by a single Evaluation Plan containing the configuration items below, and the plan fingerprint must be derived from them.

  **3.9 Evaluation Plan (Plan)**

  An evaluation is described by a single plan, which contains at least:

  | Configuration item | Description |
  |---|---|
  | Task set and version | Select the task set and lock its version |
  | System under test | Agent identifier + version + integration configuration |
  | Run configuration | Repetition count (number of Trials), random seed strategy, budget overrides |
  | Scoring configuration | Scorer selection, judge model, partial-credit policy |
  | Concurrency and rate limiting | The items in section 3.5 |
  | Ownership information | Who initiated it, which project / version it belongs to, whether it is used for a release |

  The **plan fingerprint** is derived from the above and is used for resume consistency checks, result archiving and leaderboard submission checks.

## Acceptance Criteria

- [ ] **AC1** A persisted trajectory can be judged, or re-judged with different scoring criteria, a different judge model or manual review, without calling the Agent again (R1).
- [ ] **AC2** An adapter implements all four protocol capabilities — start session, advance one step, determine completion and report usage — and reports the cumulative token count, call count, elapsed time and cost at any time (R2).
- [ ] **AC3** The adapter reports the termination reason faithfully whenever it returns "end + termination reason" (R2).
- [ ] **AC4** The platform supports native action adaptation, text action adaptation and black-box CLI adaptation, and performs the platform-side duty listed for each shape (R3).
- [ ] **AC5** The adapter self-check suite runs the same set of sample tasks (≤ 10 items per domain) for task authors and Agent developers, and outputs the expected-versus-actual action sequence difference, the action format error rate and contract violation count, and whether usage reporting is correct (R4).
- [ ] **AC6** When the adapter self-check fails, the evaluation capability is disabled, so that integration bugs are never recorded as weak model capability (R4).
- [ ] **AC7** A terminating action occupies a turn exclusively: a response containing both a terminating action and other actions is rejected as a whole and counted as a protocol violation (R5).
- [ ] **AC8** Trimming rules are declared by the task and are never decided unilaterally by the platform, so that the observation budget of different tasks is comparable (R6).
- [ ] **AC9** The termination reason is recorded for every Trial that ends (R7).
- [ ] **AC10** The information the user simulator is told defaults to "user intent + information already known to the user" and excludes scoring conditions, the tool list and the reference solution (R8).
- [ ] **AC11** The user simulator never fabricates information the task did not provide; out-of-bounds simulator behaviour (fabrication, leaking the answer, refusing to answer) is recorded and counted toward the "simulator noise" metric (R8).
- [ ] **AC12** The user simulator emits an explicit end marker instead of relying on the model to decide on its own (R8).
- [ ] **AC13** Reaching any budget terminates the Trial, judges it as "budget exhausted", and counts it in the denominator or excludes it according to the rules declared by the task (counted in the denominator by default) (R9).
- [ ] **AC14** On plan submission the platform returns the token volume range, cost range and duration range of the evaluation; when the sample is insufficient it returns a conservative upper bound marked "low estimate confidence" (R9).
- [ ] **AC15** Model-side logic errors (malformed output, hallucinated tool names) are never retried; the step fails directly and the error is fed back to the Agent (R10).
- [ ] **AC16** A retry after model-side throttling or a transient environment-side error counts toward the budget's "call count" but not toward "effective steps" (R10).
- [ ] **AC17** When the same task fails consecutively up to the threshold, the remaining Trials of that task are suspended and an alert is raised (R10).
- [ ] **AC18** A Trial that ran to completion and was judged is classified as **Completed**, counted in the denominator (✔), with no re-run (R11).
- [ ] **AC19** A Trial where the Agent crashed, the output is unparseable, tools were hallucinated, or a timeout occurred is classified as **Model failure**, counted in the denominator (✔), and is re-runnable with a different seed (R11).
- [ ] **AC20** A Trial where the environment does not start, the scorer misbehaves, or an error occurs in the platform itself is classified as **Platform failure**, excluded from the denominator (✘, counted separately), and is retried automatically (✔) (R11).
- [ ] **AC21** A Trial where a limit was reached is classified as **Budget exhausted**, counted in the denominator by default (✔), and is not re-run (✘) (R11).
- [ ] **AC22** A Trial with persistent action format errors or a broken contract is classified as **Protocol violation**, counted in the denominator (✔), and is not re-run (✘) (R11).
- [ ] **AC23** A Trial that was manually aborted or cancelled by the evaluation organiser is classified as **Cancelled**, excluded from the denominator (✘), with re-run decided by the initiator (R11).
- [ ] **AC24** Platform failures are never counted in the denominator, and the report separately lists "the platform failure rate of this evaluation" (R11).
- [ ] **AC25** If the platform failure rate exceeds the threshold (5% by default), the whole report is marked "insufficient confidence" and does not enter the leaderboard (R11).
- [ ] **AC26** The resume unit is (task ID, Trial index, random seed), completed units are recorded by this triple, and completed units are automatically skipped after a restart (R12).
- [ ] **AC27** If the task set version, Agent version, scoring configuration or budget changes in any respect, the platform refuses to resume and reports "this is a new evaluation"; the decision is based on the plan fingerprint (R12).
- [ ] **AC28** A completed judgement can be re-run on its own without changing the trajectory; the re-run result is recorded as a new judgement version and the historical judgement is retained for comparison (R12).
- [ ] **AC29** Every Trial records the task context, interaction sequence, termination info, usage, judgement, errors, environment evidence and plan fingerprint categories listed above (R13).
- [ ] **AC30** A report number can be drilled down to a specific Trial; old trajectories can be re-judged with a new scorer without calling the Agent; two runs' trajectories can be diffed; and trajectories can be exported to a common format for external tools to consume (R13).
- [ ] **AC31** An evaluation is described by a single Evaluation Plan containing task set and version, system under test, run configuration, scoring configuration, concurrency and rate limiting, and ownership information (R14).
- [ ] **AC32** The plan fingerprint is derived from the plan configuration and is used for resume consistency checks, result archiving and leaderboard submission checks (R14).

## Impact

| Area | Effect |
|---|---|
| Users | Task authors and Agent developers use the adapter self-check suite on the same set of sample tasks; evaluation organisers initiate plans and may cancel Trials; tasks that need a simulated user gain a declared persona, disclosure strategy and end signal. |
| Existing behaviour | Not specified in the source design. |
| Operational cost | Each Trial is bounded by the step, turn, token, duration, cost, call count and error count limits; the pre-run estimate returns the token volume range, cost range and duration range for the evaluation; usage (token count, call count, elapsed time, cost) is reported by the adapter for budget circuit-breaking and cost accounting; concurrency is capped by maximum concurrent Trials, maximum concurrent environment instances, model call concurrency and intra-task concurrency. |

## Risks and Mitigations

| Risk | Likelihood | Mitigation |
|---|---|---|
| Integration bugs are recorded as weak model capability. | Not specified in the source design. | The adapter self-check suite; **when the self-check fails, the evaluation capability must be disabled**. |
| An environment outage burns through the whole batch budget. | Not specified in the source design. | The circuit-breaking threshold: when the same task fails consecutively up to the threshold, the remaining Trials of that task are suspended and an alert is raised. |
| Platform failures are scored as model defects and distort the denominator. | Not specified in the source design. | Platform failures must never be counted in the denominator and the platform failure rate is listed separately; if it exceeds the threshold (5% by default) the report is marked "insufficient confidence" and does not enter the leaderboard. |
| The observation budget of different tasks is not comparable. | Not specified in the source design. | Trimming rules are declared by the task and must not be decided unilaterally by the platform. |
| Budget exhaustion is mistaken for an Agent capability defect. | Not specified in the source design. | Reaching any budget terminates the Trial, judges it as "budget exhausted", and counts it in the denominator by default because budget exhaustion reflects an Agent efficiency problem. |

## Rollout

Not specified in the source design.

## References

- Tech Design: [TD-004 M0 Runtime and Quality Gate System][td-004]
- Test Report: [TR-003 M0 Vertical Slice][tr-003]
- Functional design chapter 02 — Capability Taxonomy and Task System: [FEAT-002][feat-002]
- Source of this document: Functional design chapter 03 — Evaluation Execution and Orchestration (this document).

## Open Questions

| # | Question | Owner | Blocking? |
|---|---|---|---|
| 1 | Not specified in the source design. | - | - |

<!-- Link definitions stay at the bottom of the file. Angle brackets are required
     because the referenced paths may contain spaces. -->

[feat-002]: <../task-system/capability-taxonomy-and-task-system.md>
[td-004]: <../../Tech Design/TD-004-m0-runtime-and-quality-gate-system.md>
[tr-003]: <../../Test Report/TR-003-m0-vertical-slice.md>
