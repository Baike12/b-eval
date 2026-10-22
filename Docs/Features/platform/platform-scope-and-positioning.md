# FEAT-001: Platform Scope and Positioning

| Field | Value |
|---|---|
| **ID** | FEAT-001 |
| **Status** | Draft |
| **Owner** | Platform team |
| **Priority** | P0 |
| **Implements** | `N/A — this feature defines the platform scope; it implements no functional requirement from the inventory.` |
| **Created** | 2026-10-22 |
| **Last Updated** | 2026-10-22 |
| **Source** | `Functional design chapter 01 — Project Overview` |

## Summary

b-eval is specified here as a general-purpose Agent evaluation platform — not a specific Agent benchmark — that runs a set of Agents over a set of reproducible, explainable and comparable tasks in order to produce scores, with the tasks actually measured decided by the task sets inside the platform. This feature fixes the platform scope and positioning: the five walls that stop existing benchmarks from serving as usable evaluation infrastructure, the division of labor against existing projects, goals G1–G6, the non-goals, the user roles, the core concept model, the differentiation and the product decisions still open. The outcome is a bounded platform definition against which every later feature chapter is checked.

## Problem

**1. Background and Problem**

**Current behaviour:** The open-source community already has a large number of Agent benchmarks (SWE-bench 6068 stars, OSWorld 3229, τ²-bench 2284, AgentBench 3784, etc.), but treating them as "usable evaluation infrastructure" runs into five walls:

- **Wall 1: No unified way to onboard the system under test.**
  τ-bench's Agent follows a fixed loop that injects a tool message after the first tool_call; VitaBench uses `generate_next_message`; MCPMark uses the LiteLLM tool loop; AgentCompass uses `run_task`. To run the same Agent on two benchmarks you must write two adapter layers, and the behavior semantics of the two sides are not equivalent, so cross-benchmark scores are not comparable.
- **Wall 2: No unified scoring abstraction.**
  OSWorld uses a two-layer decoupling of "result getters (43 kinds) × judgement functions (120 kinds)"; MCPMark uses a standalone scoring script that judges by exit code; TheAgentCompany accumulates pluggable checkpoints; AgentDojo splits tasks into two orthogonal axes, "utility" and "safety"; VitaBench uses a sliding window over rubric entries. Scorers cannot be reused across projects, so evaluation platform capabilities cannot accumulate.
- **Wall 3: Infrastructure failures directly contaminate scores.**
  τ-bench v1 records any exception (including API errors) as a 0 score and counts it in the denominator, which drags pass^k down with data center jitter; OSWorld uses `unscored_errors` to exclude "the environment crashed" from the denominator; ClawBench does the same with `adjusted_eligible`. The three conventions are mutually incompatible, and none of them is written into the score itself — so anyone reading the leaderboard cannot tell whether a 0 means the model failed or the network dropped.
- **Wall 4: Scoring logic is visible to the system under test.**
  In most benchmarks the scoring script simply sits in the task directory, and the Agent under test can read "what conditions count as a pass" as soon as it has a shell. TheAgentCompany avoids this with encrypted scripts, but it is the only one that does.
- **Wall 5: No quality control loop for task authoring.**
  The cost of authoring tasks for a benchmark is extremely high (OSWorld's 369 tasks and terminal-bench's 241 tasks are all hand-built), yet there is generally no upfront validation of "the reference solution solves it / the scorer will not misjudge it / the difficulty label is trustworthy", so defects in the tasks themselves stay latent for a long time.

**Pain:** Agent developers are hurt first and hardest: they cannot tell whether a 0 on the leaderboard means the model failed or the network dropped, and they cannot compare their Agent across benchmarks because each benchmark demands its own adapter layer. Platform engineers cannot reuse scorers across projects, so evaluation platform capabilities cannot accumulate. Task authors pay the highest cost — authoring is entirely hand-built — and get no upfront validation that the reference solution solves a task, that the scorer will not misjudge it, or that the difficulty label is trustworthy, so defects stay latent. Task sets are also exposed to prompt-injection-style leakage of the scoring logic whenever the scoring script sits in the task directory and the Agent under test has a shell.

**Evidence:** The source design cites the following concrete artifacts and measurements as evidence for the walls above.

- Benchmark scale cited: SWE-bench 6068 stars, OSWorld 3229, τ²-bench 2284, AgentBench 3784, etc.
- Agent onboarding loops cited: τ-bench's fixed loop that injects a tool message after the first tool_call; VitaBench's `generate_next_message`; MCPMark's LiteLLM tool loop; AgentCompass's `run_task`.
- Scoring implementations cited: OSWorld's two-layer decoupling of "result getters (43 kinds) × judgement functions (120 kinds)"; MCPMark's standalone scoring script that judges by exit code; TheAgentCompany's pluggable checkpoints; AgentDojo's two orthogonal axes "utility" and "safety"; VitaBench's sliding window over rubric entries.
- Error conventions cited: τ-bench v1 recording any exception (including API errors) as a 0 score counted in the denominator, dragging pass^k down with data center jitter; OSWorld's `unscored_errors`; ClawBench's `adjusted_eligible`.
- Scored-script protection cited: TheAgentCompany's encrypted scripts — the only project among those cited that avoids the scoring script being visible to the Agent under test.
- Task authoring scale cited: OSWorld's 369 tasks and terminal-bench's 241 tasks, all hand-built.

## Proposed Solution

b-eval is a general-purpose Agent evaluation platform, not a specific Agent benchmark. It provides the complete capability of "running a set of Agents on a set of reproducible, explainable and comparable tasks to produce scores"; which tasks are actually measured is decided by the task sets inside the platform. Its division of labor with existing projects is reuse / replace / fill gaps / extend, and it differentiates itself through a unified Agent adapter protocol with unified cross-benchmark metrics, scoring that is invisible to the Agent under test but fully auditable by platform operators, and an explicit confidence attached to every score.

## Scope

### In Scope

- A unified Agent adapter protocol that takes all Agents: any Agent only needs to implement a minimal interface, and evaluation behavior is decoupled from the Agent's internal implementation.
- A unified scoring abstraction that takes all task types: deterministic checks, final environment state comparison, rule assertions, script-based scoring, LLM judging and human review — all six kinds of scorers unified under the same model and composable across tasks and benchmarks.
- Trustworthy and explainable scores: every evaluation run can answer which tasks did not score, why they did not score, and whether a "no score" is the model's problem or the platform's problem.
- Scores comparable across Agents on the same tasks: under the same task set and the same budget convention, scores of different Agents are directly comparable and are not affected by the default values of their respective frameworks.
- Predictable evaluation cost: before a run the token volume, cost and time upper bound of that evaluation can be estimated, and exceeding them is known immediately.
- An upfront task quality gate: every new task passes solvability, scoring trustworthiness and difficulty trustworthiness before it enters an official task set.
- The core concept model: Benchmark, TaskSet, Task, Domain, CapabilityTag, Difficulty, InitialState, Contract, Judgement, AgentAdapter, AgentProtocol, Run, Plan, Trial, Trajectory, Verdict, Metric — with Trial as the atomic unit of evaluation, Judgement and execution separated (offline rescoring), and every score drillable down to Trials.
- Invisible-to-the-Agent, auditable-by-operator scoring: every score change carries an author, a reason and a timestamp.
- Explicit score confidence: how many Trials were excluded for platform reasons, how high the judge agreement is, how wide the confidence interval is.
- Reuse of existing benchmarks (SWE-bench / OSWorld / τ²-bench / AgentBench) as "connectable task sets", and replacement of the evaluation framework bundled with each benchmark with unified run orchestration, scoring and reporting.
- Gap filling: human rubrics, unified cross-benchmark metrics and observable scoring — the parts existing projects generally lack — provided in a unified way.
- The full lifecycle of non-public tasks: private task sets and regression testing of internal Agent versions.
- The six user roles and their core use cases: Agent developer, Agent developer / platform engineer, Task author, Benchmark maintainer, Evaluation organizer, Security reviewer.

### Out of Scope

**1.4 Non-Goals**

- **Not a general-purpose Agent framework.** b-eval does not provide runtime capabilities such as Agent planning, memory or reflection; it only provides "onboarding the system under test" and "running it".
- **Not model training / reinforcement learning.** Evaluation and training are two different things. Exporting trajectories into a reusable format is within this platform's boundary, but the training loop is out of scope.
- **Not pursuing task volume.** Prefer 200 tasks that hold up to scrutiny over 2000 tasks mixed with quality problems.
- **Not general-purpose quality evaluation.** No competition with DeepEval / Ragas / promptfoo for the "LLM output quality assertion" market; b-eval only cares about Agent behavior.
- **Not pursuing a single score.** No metric of the form "one number represents Agent capability"; the platform mandates output decomposed along at least two axes: capability dimension and domain.

## Requirements

Numbered, independently testable, and each traceable to an acceptance criterion.

- **R1** The platform must be positioned as a general-purpose Agent evaluation platform rather than a specific Agent benchmark, and must keep the following division of labor with existing projects.

  **1.2 Project Positioning**

  **b-eval is a general-purpose Agent evaluation platform, not a specific Agent benchmark.**

  What it provides is the complete capability of "running a set of Agents on a set of reproducible, explainable and comparable tasks to produce scores"; which tasks are actually measured is decided by the task sets inside the platform.

  Division of labor with existing projects:

  | Relationship | Target | b-eval's stance |
  |---|---|---|
  | Reuse | Existing benchmarks such as SWE-bench / OSWorld / τ²-bench / AgentBench | Onboarded as "connectable task sets"; the platform does not rebuild tasks |
  | Replace | The evaluation framework bundled with each benchmark | Provide unified run orchestration, scoring and reporting capabilities |
  | Fill gaps | Human rubrics, unified cross-benchmark metrics, observable scoring | The parts existing projects generally lack, provided in a unified way by the platform |
  | Extend | Private task sets, regression testing of internal Agent versions | The platform must support the full lifecycle of non-public tasks |

  The concrete task and taxonomy model behind these stances is specified in [FEAT-002 Capability Taxonomy and Task System][feat-002]; the run orchestration, scoring and reporting capabilities are specified in [FEAT-003 Evaluation Execution and Orchestration][feat-003], [FEAT-004 Scoring and Metrics][feat-004] and [FEAT-005 Reporting, Analysis and Comparison][feat-005].

- **R2** The platform must accept every Agent through one adapter protocol, at an integration cost of ≤ 1 day without modifying the platform.

  **1.3 Goals**

  The source section `1.3 Goals` states six goals, each carried as one numbered requirement below: R2 = G1, R3 = G2, R4 = G3, R5 = G4, R6 = G5, R7 = G6.

  - **G1 · One adapter protocol takes all Agents.** Any Agent only needs to implement a minimal interface to be evaluated, and the evaluation behavior is decoupled from the Agent's internal implementation. Acceptance criterion: the integration cost for that Agent, without modifying the platform, is ≤ 1 day.

- **R3** The platform must accept all task types through one scoring abstraction, unifying all six scorer kinds under the same composable model.

  - **G2 · One scoring abstraction takes all task types.** Deterministic checks, final environment state comparison, rule assertions, script-based scoring, LLM judging and human review — all six kinds of scorers are unified under the same model and composable across tasks and benchmarks.

- **R4** The platform must make every score trustworthy and explainable, and must attribute a "no score" to the model or to the platform.

  - **G3 · Scores are trustworthy and explainable.** The score of any evaluation run can answer three questions: which tasks did not score, why they did not score, and whether a "no score" is the model's problem or the platform's problem.

- **R5** The platform must make scores comparable across Agents on the same tasks, independently of their frameworks' default values.

  - **G4 · Scores are comparable across Agents on the same tasks.** Under the same task set and the same budget convention, the scores of different Agents are directly comparable and are not affected by the default values of their respective frameworks.

- **R6** The platform must make evaluation cost predictable, with a pre-run upper bound on token volume, cost and time.

  - **G5 · Evaluation cost is predictable.** Before a run you can estimate the token volume, cost and time upper bound of that evaluation; if the run exceeds them, you know immediately.

- **R7** The platform must gate every new task behind three upfront validations before it enters an official task set.

  - **G6 · Task quality has an upfront gate.** Before a new task enters an official task set, it must pass three validations: solvability, scoring trustworthiness and difficulty trustworthiness.

- **R8** The platform must serve the six user roles below, each with its stated need and core use case.

  **1.5 User Roles**

  | Role | Typical need | Core use case |
  |---|---|---|
  | **Agent developer** | How does my Agent perform on real tasks? Did this change make it better or worse? | Submit evaluations, read reports, compare against the previous version, inspect failed trajectories to locate problems |
  | **Agent developer / platform engineer** | I need to give someone a runnable Agent — how do evaluations hook in? | Onboard an Agent, get the sample task running, review the onboarding self-check results |
  | **Task author** | I want to add a batch of test tasks | Create tasks, attach scorers, run validation, submit into the set |
  | **Benchmark maintainer** | Do this benchmark's tasks still hold up? Is the difficulty tiering still accurate? | Version management, task set auditing, task revision and retirement |
  | **Evaluation organizer** | Which of these Agents is stronger? Can the conclusion be trusted? | Batch orchestration, cross-benchmark comparison, leaderboard publishing, result auditing |
  | **Security reviewer** | Could this Agent be steered off course by prompt injection? | Red and blue team task sets, ASR reports, defense strategy comparison |

  The end-to-end flows behind these use cases are specified in [FEAT-007 User Flows and Feature Inventory][feat-007].

- **R9** The platform must expose only the seven first-class concepts of the core concept model, and every other entity must be a configuration of them.

  **1.6 Core Concept Model**

  The platform has only seven first-class concepts; everything else is a configuration of them.

  ```
  Benchmark
    └─ TaskSet  ── versioned, frozen, publishable
         └─ Task
              ├─ Domain         environment/business scenario the task belongs to
              ├─ CapabilityTag  capability dimension the task examines
              ├─ Difficulty     difficulty tier of the task
              ├─ InitialState   environment snapshot and data before the task starts
              ├─ Contract       action space, observation form, termination conditions, budget
              └─ Judgement      scorer set and how the scorers are combined

  System under test (Agent) ── AgentAdapter ── AgentProtocol

  Run ── Plan ── Trial ── Trajectory
                          └─ Verdict ── Metric
  ```

  **Key constraints:**

  - **Tasks and Agents are fully decoupled.** A task does not know an Agent exists, and an Agent does not know what the task is; only the "contract" is the single interface between the two sides.
  - **A Trial is the atomic unit of evaluation.** One Trial = one complete execution of "task × Agent × random seed × budget". All metrics are ultimately aggregated from Trials.
  - **Judgement and execution are separated.** Judgement can be re-executed over already saved trajectories without re-running the Agent (offline rescoring). This is the foundation for troubleshooting, changing the scoring convention and third-party audits.
  - **Every score must be drillable down to Trials.** Every number in a report can be traced back to the specific Trials it was computed from.

  The Trial, Trajectory, Verdict and Metric execution model is specified in [FEAT-003 Evaluation Execution and Orchestration][feat-003]; the metric aggregation over Trials is specified in [FEAT-004 Scoring and Metrics][feat-004].

- **R10** The platform must deliver the following three differentiators relative to existing work.

  **1.7 Differentiation**

  Relative to existing work, b-eval has exactly three things it must do, and that nobody currently does:

  1. **A unified Agent adapter protocol + unified cross-benchmark metrics.** The same Agent runs SWE-bench, OSWorld and τ²-bench through the same adapter layer, and the resulting scores land in the same coordinate system.
  2. **Invisible scoring + auditable scoring.** The scoring logic is invisible to the Agent under test; at the same time it is fully auditable by platform operators — every score change carries an author, a reason and a timestamp.
  3. **Score confidence is explicit.** Every score carries information about "how trustworthy this score is": how many Trials were excluded for platform reasons, how high the judge agreement is, how wide the confidence interval is.

  The security boundary that makes the scoring logic invisible is specified in [FEAT-006 Security, Robustness and Quality Assurance][feat-006].

- **R11** The platform must stay inside the scope boundaries listed in Out of Scope; in particular it must not provide Agent runtime capabilities (planning, memory, reflection), must not contain a training / reinforcement-learning loop, must not offer general-purpose "LLM output quality assertion" evaluation, and must not collapse its output into a single number.

  **1.4 Non-Goals**

  The full non-goal wording is kept verbatim in [Scope → Out of Scope](#scope); the mandatory boundaries it states are:

  - **Not a general-purpose Agent framework.** b-eval does not provide runtime capabilities such as Agent planning, memory or reflection; it only provides "onboarding the system under test" and "running it".
  - **Not model training / reinforcement learning.** Evaluation and training are two different things. Exporting trajectories into a reusable format is within this platform's boundary, but the training loop is out of scope.
  - **Not general-purpose quality evaluation.** No competition with DeepEval / Ragas / promptfoo for the "LLM output quality assertion" market; b-eval only cares about Agent behavior.
  - **Not pursuing a single score.** No metric of the form "one number represents Agent capability"; the platform mandates output decomposed along at least two axes: capability dimension and domain.

  The one remaining non-goal — **Not pursuing task volume.** Prefer 200 tasks that hold up to scrutiny over 2000 tasks mixed with quality problems — is a stated preference rather than a prohibition, so it is not turned into an acceptance criterion below.

## Acceptance Criteria

Written as verifiable outcomes. Use Given/When/Then only where the setup is non-obvious.

- [ ] **AC1** (R2) Onboarding any new Agent requires implementing only the minimal adapter interface, requires no modification to the platform, and takes no more than 1 day; the evaluation behavior observed for that Agent is decoupled from the Agent's internal implementation.
- [ ] **AC2** (R3) All six scorer kinds — deterministic checks, final environment state comparison, rule assertions, script-based scoring, LLM judging and human review — are expressed under one scoring model, and scorer sets can be composed across tasks and across benchmarks.
- [ ] **AC3** (R4) For any evaluation run, the platform reports which tasks did not score, why they did not score, and whether a "no score" is the model's problem or the platform's problem.
- [ ] **AC4** (R5) Under the same task set and the same budget convention, scores produced for different Agents are directly comparable, and changing the default values of their respective frameworks does not change those scores.
- [ ] **AC5** (R6) Before a run starts, the platform can report the estimated token volume, cost and time upper bound of that evaluation, and reports immediately when a run exceeds them.
- [ ] **AC6** (R7) A new task cannot enter an official task set until it has passed all three validations — solvability, scoring trustworthiness and difficulty trustworthiness.
- [ ] **AC7** (R1) Existing benchmarks (SWE-bench / OSWorld / τ²-bench / AgentBench) are onboarded as "connectable task sets" and their tasks are not rebuilt, and the platform supplies the unified run orchestration, scoring and reporting that replaces each benchmark's bundled evaluation framework.
- [ ] **AC8** (R1) The platform supports the full lifecycle of non-public tasks — private task sets and regression testing of internal Agent versions.
- [ ] **AC9** (R8) Each of the six roles — Agent developer, Agent developer / platform engineer, Task author, Benchmark maintainer, Evaluation organizer, Security reviewer — can complete its listed core use case through the platform.
- [ ] **AC10** (R9) The platform's data model contains only the seven first-class concepts of the Core Concept Model, and every other platform entity is a configuration of them.
- [ ] **AC11** (R9) A task contains no reference to any Agent, and an Agent contains no reference to the task; the "contract" is the only interface between the two sides.
- [ ] **AC12** (R9) One Trial equals one complete execution of "task × Agent × random seed × budget", and every metric in a report is aggregated from Trials.
- [ ] **AC13** (R9) Judgement can be re-executed over already saved trajectories without re-running the Agent (offline rescoring).
- [ ] **AC14** (R9) Every number in a report can be traced back to the specific Trials it was computed from.
- [ ] **AC15** (R10) The scoring logic is not readable by the Agent under test.
- [ ] **AC16** (R10) Every score change carries an author, a reason and a timestamp, and a platform operator can audit the change.
- [ ] **AC17** (R10) Every score exposes its confidence information: how many Trials were excluded for platform reasons, how high the judge agreement is, and how wide the confidence interval is.
- [ ] **AC18** (R11) The platform exposes no Agent runtime capabilities such as Agent planning, memory or reflection, and contains no training / reinforcement-learning loop; trajectory export into a reusable format remains possible.
- [ ] **AC19** (R11) The platform makes no general-purpose "LLM output quality assertion" evaluation and does not compete with DeepEval / Ragas / promptfoo for that market; it only evaluates Agent behavior.
- [ ] **AC20** (R11) Platform output is decomposed along at least two axes — capability dimension and domain — and no single number is presented as representing Agent capability.

## Impact

| Area | Effect |
|---|---|
| Users | Six roles gain a platform to work in: Agent developer, Agent developer / platform engineer, Task author, Benchmark maintainer, Evaluation organizer and Security reviewer, each with the need and core use case listed under R8. Task authors gain an upfront gate instead of discovering task defects late, and evaluation organizers gain cross-benchmark comparison and result auditing; the tasks actually measured remain a choice made by the task sets inside the platform. |
| Existing behaviour | Existing benchmarks such as SWE-bench / OSWorld / τ²-bench / AgentBench are onboarded as "connectable task sets" — the platform does not rebuild tasks — and the evaluation framework bundled with each benchmark is replaced by unified run orchestration, scoring and reporting. The parts existing projects generally lack (human rubrics, unified cross-benchmark metrics, observable scoring) are provided in a unified way, while non-public task sets gain a supported lifecycle that they did not have. |
| Operational cost | Evaluation cost becomes predictable: before a run the token volume, cost and time upper bound can be estimated, and exceeding them is known immediately (G5). Task authoring cost remains the dominant human cost — OSWorld's 369 tasks and terminal-bench's 241 tasks are all hand-built — which is why G6 places three upfront validations (solvability, scoring trustworthiness, difficulty trustworthiness) in front of every new task. |

## Risks and Mitigations

| Risk | Likelihood | Mitigation |
|---|---|---|
| Not specified in the source design. | - | - |

## Rollout

Not specified in the source design.

## References

- Tech Design: Not yet created.
- Test Report: Not yet created.
- Related features: [FEAT-002 Capability Taxonomy and Task System][feat-002], [FEAT-003 Evaluation Execution and Orchestration][feat-003], [FEAT-004 Scoring and Metrics][feat-004], [FEAT-005 Reporting, Analysis and Comparison][feat-005], [FEAT-006 Security, Robustness and Quality Assurance][feat-006], [FEAT-007 User Flows and Feature Inventory][feat-007], [FEAT-008 Delivery Roadmap and Acceptance Criteria][feat-008]

## Open Questions

**1.8 Open Questions (Require Product Decisions)**

| ID | Question | Area of impact | Owner | Blocking? |
|---|---|---|---|---|
| Q1 | Should the platform operate a public leaderboard? A public leaderboard requires a submission flow and anti-cheating measures; a private deployment does not | Submission flow, result visibility, review queue | Not specified in the source design. | Not specified in the source design. |
| Q2 | Should invisible scoring code be delivered "encrypted", or should "scoring run in an isolated environment where the tested party has no read access"? | Task package format, sandbox topology | Not specified in the source design. | Not specified in the source design. |
| Q3 | Are third parties allowed to submit Agent results? If so, the scoring logic is necessarily visible to the submitter, which conflicts with G-"invisible scoring" | Leaderboard design and security boundary | Not specified in the source design. | Not specified in the source design. |
| Q4 | Are task sets required to be tiered (public subset / semi-public / private)? The existence of private sets makes scores impossible to compare across the board | Score comparability, leaderboard design | Not specified in the source design. | Not specified in the source design. |
| Q5 | Is the user simulator (a persona-based user) a first-class citizen or an optional component? | Task model, interaction contract | Not specified in the source design. | Not specified in the source design. |

<!-- Link definitions stay at the bottom of the file. -->

[feat-002]: <../task-system/capability-taxonomy-and-task-system.md>
[feat-003]: <../execution/evaluation-execution-and-orchestration.md>
[feat-004]: <../scoring/scoring-and-metrics.md>
[feat-005]: <../reporting/reporting-analysis-and-comparison.md>
[feat-006]: <../security/security-robustness-and-quality-assurance.md>
[feat-007]: <../experience/user-flows-and-feature-inventory.md>
[feat-008]: <./delivery-roadmap-and-acceptance-criteria.md>
