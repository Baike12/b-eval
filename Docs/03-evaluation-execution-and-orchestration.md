# 03 · Evaluation Execution and Orchestration

## 3.1 The Complete Flow of One Evaluation

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

## 3.2 Adapter Protocol for the Agent under Test

### 3.2.1 Protocol Definition

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

### 3.2.2 Three Adapter Shapes

| Shape | Applies to | What the platform must do |
|---|---|---|
| **Native action adaptation** | The Agent natively supports function calls / structured actions | Provide the action schema so that the Agent connects to the environment directly; the action space is given by the task contract |
| **Text action adaptation** | The Agent only outputs natural-language actions (e.g. `Action: X with Action Input: Y`) | The platform parses the text actions; parse failures count toward the "action format error rate" and are handled according to the contract rules |
| **Black-box CLI adaptation** | The Agent is an off-the-shelf tool (e.g. a coding Agent in a terminal) | The platform runs its process in a controlled environment and collects its terminal interactions and file outputs |

### 3.2.3 Adapter Self-Check

The platform provides an adapter self-check suite: task authors and Agent developers use the same set of sample tasks (≤ 10 items per domain, see 2.4) to verify that the integration is correct.

The self-check outputs:
- The difference between the expected action sequence and the actual action sequence of each sample task (provided only when the sample task ships with a reference trajectory).
- The action format error rate and the number of contract violations.
- Whether usage reporting is correct.
- **When the self-check fails, the evaluation capability must be disabled** — so that integration bugs are never recorded as weak model capability.

## 3.3 Interaction Contract

Each task declares the interaction shape it expects, and the platform drives execution accordingly.

### 3.3.1 Action Space

- A task declares an **action set**; each action includes: name, parameter structure, semantic description, whether it changes environment state, and behaviour on failure.
- Actions fall into four categories: **read** (query, no state change), **write** (changes state), **reason** (pure internal computation), **terminate** (e.g. escalate to a human, declare completion).
- **"Whether it changes state" is a switch independent of the four categories**, because there exist actions that are "semantically a write but actually do not change state" (see τ²-bench's separation of `mutates_state` from `ToolType`). This switch determines which actions are executed when the trajectory is replayed during the judgement phase.
- A terminating action **must occupy a turn exclusively**: if a single response contains both a terminating action and other actions, the whole package is rejected and counted as a protocol violation (see τ²-bench's rule that the terminating tool owns the turn).

### 3.3.2 Observation Shapes

A task declares what the Agent can see at each step:

| Shape | Description |
|---|---|
| Structured observation | Tool return values (JSON), state query results |
| Text observation | Page text, terminal output, file content |
| Visual observation | Screenshots, accessibility trees and combinations of them |
| Mixed observation | Text and visual given together |

**Observation trimming capability** (to keep long context from drowning out the task itself): viewport filtering, token-limit truncation, element filtering, on-demand expansion. Trimming rules are declared by the task and must not be decided unilaterally by the platform — otherwise the observation budget of different tasks would not be comparable.

### 3.3.3 Termination Conditions

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

### 3.3.4 User Simulator (Optional)

Some tasks (e.g. business processes, requirement clarification) require a simulated user to take part. The task declares "whether there is a human user" along with that user's persona.

- The **user persona sources** are separated into three non-overlapping layers: general behavioural guidelines / in-task persona / run-time override.
- The **information the user must be told** is specified by the task and defaults to "user intent + information already known to the user"; it excludes scoring conditions, the tool list and the reference solution.
- The **user's information disclosure strategy**: by default "disclose step by step on demand", giving only the information the current step needs; a task may declare "say everything in the first turn".
- The **user may change their mind**: conditional branches are supported in task instructions (e.g. "if a certain option is missing, you will pick another").
- **End signal**: the user emits an explicit end marker instead of relying on the model to decide on its own.
- The **user must not fabricate** information the task did not provide; out-of-bounds simulator behaviour (fabrication, leaking the answer, refusing to answer) is recorded and counted toward the "simulator noise" metric.

## 3.4 Budgets and Limits

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

## 3.5 Concurrency and Rate Limiting

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

## 3.6 Failure Classification

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

## 3.7 Resume and Idempotency

**Resume unit = (task ID, Trial index, random seed).**

- The platform records completed units by this triple and automatically skips completed units after a restart.
- **The configuration must be identical when resuming**: if the task set version, Agent version, scoring configuration or budget changes in any respect, the platform refuses to resume and reports "this is a new evaluation". The decision is based on the plan fingerprint.
- **A completed judgement can be re-run on its own** (without changing the trajectory) in order to switch the scoring criteria; the re-run result is recorded as a new judgement version, and the historical judgement is retained for comparison.

## 3.8 Trajectory and Observability Records

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

## 3.9 Evaluation Plan (Plan)

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
