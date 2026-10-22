# FEAT-002: Capability Taxonomy and Task System

| Field | Value |
|---|---|
| **ID** | FEAT-002 |
| **Status** | Draft |
| **Owner** | Platform team |
| **Priority** | P0 |
| **Implements** | `F-TSK-01`, `F-TSK-02`, `F-TSK-03`, `F-TSK-04`, `F-TSK-05`, `F-TSK-06`, `F-TSK-07`, `F-TSK-08`, `F-TSK-09`, `F-TSK-10`, `F-TSK-11`, `F-TSK-12`, `F-TSK-13` |
| **Created** | 2026-10-22 |
| **Last Updated** | 2026-10-22 |

## Summary

The platform introduces a dual-axis capability taxonomy — six level-1 capability dimensions (`C1`–`C6`) with 18 level-2 labels, plus a domain axis of seven domains (`D1`–`D7`) — as the labelling scheme every stored task must carry. On top of that taxonomy it specifies the complete task unit: metadata contract, templating and parameterization, task-set partitioning, difficulty tiering, isolated execution, task sources, and admission validation gates. Every task declares exactly one **form** — `environment-agentic` or `static-qa` — and the form decides which field groups the task must carry and which admission gates apply to it, so that a batch of plain question-and-answer data is a first-class task set rather than a second-class one. The outcome is a task store in which every score is attributable to a labelled, versioned, reproducible task, and in which an untested capability is distinguishable from a failed one.

## Problem

**Current behaviour:** Not specified in the source design; the source chapter defines the capability taxonomy and task system as new platform structure rather than describing an existing implementation.

**Pain:** Without the dual-axis taxonomy a task cannot answer "which capability is the Agent missing" or "in which class of real-world scenarios the Agent fails", so platform results cannot be decomposed, compared or regression-tracked. Without a fixed task unit, immutable task IDs, declared partitions, a difficulty tier, isolated execution and admission gates, a score cannot be attributed to the task it came from, and a failure cannot be separated into a platform fault, a model failure or a judgement fault. Without a task **form**, a batch of plain question-and-answer data must be disguised as an environment task: the author has to invent an environment snapshot, an action space and an observation form that the task never uses, and the platform has to pay for environment bring-up, isolation and teardown machinery on every Trial to produce a verdict that is really just a text comparison.

**Evidence:** None cited in the source design; the chapter is a normative design specification rather than a measurement report. The form split is justified by the survey recorded in [`Docs/README.md` §2](../../README.md): of the twenty-three surveyed benchmarks, the large majority are environment-agentic, and none of them models a static question-and-answer task set as a first-class form.

## Proposed Solution

Define the taxonomy and the task system as platform-level rules that every stored task must satisfy: a two-axis labelling scheme (capability dimensions and domains) with automatic label checks, a weighted score composition model that refuses to hide missing coverage, a mandatory task metadata contract enforced per task form, template instantiation with declared distributions, five independently versioned task sets, a four-tier difficulty scheme with baseline calibration, isolated per-Trial environments with a four-stage lifecycle, and eight admission validation gates whose environment-dependent subset applies only to `environment-agentic` tasks.

## Scope

### In Scope

- Dual-axis task taxonomy: a capability dimension axis (6 level-1 dimensions `C1`–`C6`, 18 level-2 labels) and a domain axis (`D1`–`D7`), labelled independently, with labels allowed to overlap across both axes.
- The task **form** dimension: `environment-agentic` and `static-qa`, the form-dependent metadata contract, the form-dependent subset of the admission gates, and the rejection of fields a form does not use.
- Labeling rules: minimum and recommended level-1 label counts, a level-2 label must hit the task's actual failure point, multi-select labels, and an automatic mutual-exclusion check before a task enters the store.
- Capability score composition: two-level ("dimension → domain") weights normalized to sum to 1, weights declared explicitly and persisted alongside the results, explicit "no coverage" marking, and unweighted raw per-domain scores shown alongside weighted views on public leaderboards.
- The complete task unit: identity, capability annotation, task description, initial state, interaction contract, judgement conditions, reference material, and meta information, with the specified required/optional split.
- Task templating and parameterization: templates, parameter instances, independent per-instance scoring with stable instance IDs, an explicitly declared post-instantiation distribution, and deduplication plus coverage checks on the parameter domain.
- Task set partitioning: five coexisting task sets (`smoke`, `dev`, `test-public`, `test-private`, `stress`) under one benchmark, each independently versioned, with declared visibility, purpose and constraints.
- Difficulty tiering: a four-tier scheme (`L1`–`L4`), baseline measurement before publishing a task set, conflict alerts between annotation and measurement, and six-monthly re-calibration.
- Environment and state isolation: an isolated environment instance per task, the four lifecycle stages (Preparation, Execution, Judgement, Cleanup) with per-stage failure attribution, and rollback-capable re-scoring.
- Task sources: external benchmark import, internal construction, and community submission, each with its own admission requirement and additional constraints.
- Task admission validation: eight validation gates that every new task must pass before entering `dev`, with gates 2–5 additionally required before entering `test`.

### Out of Scope

Non-goals stated by the source design:

- A single "overall capability score" for the platform — the platform does not produce one, only configurable weighted views.
- Weighted composed scores on public leaderboards — weighted scores are only used for internal regression comparison; public leaderboards must simultaneously show the unweighted raw per-domain scores.
- The `stress` set inside the total score — it is scored separately and not merged into the total score, because its difficulty distribution differs from the regular sets and mixing it in would destroy comparability.
- Judgement conditions and judgement logic inside the Agent-accessible environment — scoring must not be visible to the Agent (see [Chapter 06][feat-006]).
- Failed community submissions in the official task set — submissions that fail review do not enter the official set.

## Requirements

Numbered, independently testable, and each traceable to an acceptance criterion.

- **R1** Every task is labelled on two independently labelled axes — a capability dimension axis and a domain axis — and a single task may fall into multiple capability dimensions and multiple domains at the same time.

  **2.1 Capability Taxonomy Model**

  The platform uses a **dual-axis taxonomy**: the capability dimension axis answers "which capability is the Agent missing", and the domain axis answers "in which class of real-world scenarios the Agent fails". The two axes are labeled independently; any single task may fall into multiple capability dimensions and multiple domains at the same time.

- **R2** The capability dimension axis is divided into exactly 6 level-1 dimensions, `C1`–`C6`.

  **2.1.1 Capability Dimension Axis (Level 1)**

  Merged from the four foundational capability threads of the Agent evaluation survey (planning / tool use / self-reflection / memory) and AgentBoard's six-dimension weighted score, divided into 6 level-1 dimensions:

  | Dimension | Connotation | Typical assessment method |
  |---|---|---|
  | **C1 Planning and Reasoning** | Decomposing tasks, choosing paths, making decisions under constraints | Multi-step tasks, partial-completion scoring, step efficiency |
  | **C2 Tool and Interface Use** | Calling tools correctly, correct arguments, combining multiple tools, handling call failures | Tool-call match rate, argument correctness rate, final-environment-state comparison |
  | **C3 State and Memory Retention** | Tracking state across turns/steps, remembering historical information, resolving conflicting information | Progress rate, memory-type datasets, incremental multi-turn follow-ups |
  | **C4 Interaction and Collaboration** | Dialoguing with users/other Agents, clarifying requirements, handling changes and interruptions | User simulator tasks, multi-Agent collaboration tasks |
  | **C5 Error Handling and Self-Correction** | Locating the cause after failure, correcting, retrying without diverging | Tasks with injected failure branches, post-failure recovery rate |
  | **C6 Robustness and Safety** | Resisting prompt injection, unauthorized operations, inappropriate content, recognizing infeasible tasks | Attack-defense task sets, infeasible tasks |

- **R3** The capability dimension axis carries 18 level-2 labels in total, distributed under the 6 level-1 dimensions, applied under the stated labeling rules.

  **2.1.2 Capability Dimension Axis (Level 2, 18 items in total)**

  | Level 1 | Level 2 labels |
  |---|---|
  | C1 Planning and Reasoning | Task decomposition, constraint adherence, path selection, long-horizon planning, goal clarification |
  | C2 Tool and Interface Use | Tool selection, argument construction, tool combination, tool-failure recovery, protocol adaptation (incl. MCP) |
  | C3 State and Memory Retention | Working memory, session state tracking, long-term memory, information conflict resolution, context compression |
  | C4 Interaction and Collaboration | Requirement clarification, proactive information synchronization, response to user intent changes, multi-Agent collaboration |
  | C5 Error Handling and Self-Correction | Failure localization, self-correction, retry strategy, hallucination suppression |
  | C6 Robustness and Safety | Prompt injection resistance, unauthorized-request refusal, infeasible task identification, content compliance |

  **Labeling rules:**

  - Every task carries **at least 1** level-1 dimension label, **2 is recommended**; more than 3 is treated as over-broad labeling and requires review.
  - A level-2 label must hit the task's actual failure point; do not infer it from "what this kind of task usually tests".
  - Labels are **multi-select**, and the **mutual-exclusion check** is executed automatically before a task enters the store (the same level-2 label must not repeat).

- **R4** The domain axis is divided into 7 domains, `D1`–`D7`, and a task is allowed to span multiple domains.

  **2.1.3 Domain Axis**

  Divided by "what the Agent operates on", consistent with the domain-bundling approach of τ²-bench:

  | Domain | Coverage | Reference sources |
  |---|---|---|
  | **D1 Web and Browser** | Information retrieval, form filling, cross-site operations, visibility-sensitive web tasks | WebArena, VisualWebArena, ClawBench |
  | **D2 Desktop and GUI Applications** | Operating systems, office software, image/media, IDEs | OSWorld, WindowsAgentArena |
  | **D3 Code and Software Engineering** | Repository-level bug fixing, feature implementation, terminal operations, training and tuning | SWE-bench, terminal-bench, MLE-bench |
  | **D4 Tools and APIs** | Function calling, REST/service integration, MCP services, enterprise systems | BFCL, MCPMark, τ-bench |
  | **D5 Real Business Processes** | Customer service, e-commerce, travel, government affairs, and other end-to-end business processes | τ²-bench, VitaBench |
  | **D6 Multi-Agent Collaboration** | Division of labor, handoff, collaborative delivery | TheAgentCompany, AgentVerse |
  | **D7 Knowledge and Reasoning** | Closed-book question answering, reading comprehension over supplied material, reasoning and calculation over given data | New in b-eval — none of the surveyed benchmarks models closed-book QA as a first-class domain |

  **Cross-domain tasks:** A task is allowed to span multiple domains (e.g. "compare prices on the web → produce a report in office software"), used to assess C1/C4.

  **Domain and form are independent axes.** A domain says *what the task is about*; the form (2.9) says *how the verdict is obtained*. A closed-book knowledge question is `D7` + `static-qa`. A question answered by navigating a live web page is `D1` + `environment-agentic`. A question whose page content is pasted into the prompt is `D1` + `static-qa`.

- **R5** Capability scores are exposed only as configurable weighted views with explicitly declared and persisted weights normalized at each of the two levels, and missing coverage is never silently zero-filled.

  **2.1.4 Capability Score Composition Rules**

  The platform **does not produce a single "overall capability score"**, but provides configurable weighted views. Weights must be declared explicitly and persisted alongside the results; the rules are:

  - Weights are given at two levels, "dimension → domain", and the weights at each level are normalized to sum to 1.
  - **Missing coverage is never silently zero-filled**: when no task under a given dimension was evaluated, that dimension's score is marked "no coverage", excluded from the weighting, and highlighted in red in the report.
  - Weighted scores are only used for internal regression comparison; **public leaderboards must simultaneously show the unweighted raw per-domain scores**.

- **R6** Every task declares exactly one **form**, the metadata contract is enforced per form, task IDs are immutable and never reused, and judgement conditions are kept out of the Agent-accessible environment.

  **2.2 Task Unit and Metadata**

  The complete definition of a task contains the following fields. Field names are functional naming only; the concrete storage format is specified in `[pending technical design]`. The **form** (2.9) decides which field groups are required: a task may not omit a group its form requires, and may not carry a group its form does not use.

  | Field group | Content | `environment-agentic` | `static-qa` |
  |---|---|:---:|:---:|
  | **Identity** | Task ID (stable, globally unique, never reused), display name, owning benchmark, owning task set version, task form | ✔ | ✔ |
  | **Capability annotation** | Level-1 dimension list, level-2 label list, difficulty level, domain list | ✔ | ✔ |
  | **Task description** | Natural-language instruction aimed at the Agent; explanation for evaluators other than the Agent (optional) | ✔ | ✔ |
  | **Budget** | Step cap, token cap, duration cap, cost cap, error-count cap | ✔ | ✔ |
  | **Initial state** | Environment snapshot identifier, initial data, initial files, initial credentials, random seed | ✔ | — |
  | **Interaction contract** | Action space description, observation form, termination conditions, whether multi-turn is allowed | ✔ | partial — termination conditions and the multi-turn flag only |
  | **Judgement conditions** | Scorer list, combination method, partial-credit rules, primary criterion, target state **or** reference answer | ✔ | ✔ |
  | **Reference material** | Reference solution (optional), reference trajectory (optional), domain policy document (optional) | ✔ | — |
  | **Meta information** | Author, source, difficulty annotation basis, estimated human time cost, estimated Agent time cost, creation time, last revision time, deprecation status | ✔ | ✔ |

  **Design points:**

  - **Task IDs are immutable and never reused.** Deprecated tasks are retained in the store and marked as deprecated; an ID is never assigned to a new task again—otherwise historical scores would point to a different task.
  - **Difficulty is a required field**, derived from a combination of manual annotation + estimated time cost + actual failure rate; the decision rules are in [§2.5][feat-002].
  - **"Instructions aimed at the Agent" and "explanations aimed at evaluators" are strictly separated.** The former participates in the run; the latter appears only in reports and review interfaces, preventing task leakage.
  - **Judgement conditions are stored separately from the task description.** Judgement conditions must not appear inside the Agent-accessible environment (scoring must not be visible; see [Chapter 06][feat-006]). For a `static-qa` task the reference answer is a judgement condition and is likewise withheld from the Agent.
  - **A field the form does not use is a field nobody maintains.** The platform rejects a task that carries an environment-only field group under `static-qa`, because an unused field silently diverges from reality and gives a false impression of what the task controls.

- **R7** Task templating and parameterization follow the template/instance model, score each instance independently with a stable instance ID, declare the post-instantiation distribution explicitly, and subject the parameter domain to deduplication and coverage checks.

  **2.3 Task Templating and Parameterization**

  Following WebArena's approach (190 intent templates → 812 task instances), the platform supports task templating:

  - A **template** defines the task skeleton: instruction template, action space, judgement method, difficulty.
  - **Parameter instances** fill in concrete values, generating multiple task instances.
  - **Instances must each be scored independently**, and each instance ID must itself be stable.
  - **The distribution after template instantiation must be declared explicitly**: if one template yields 20 instances, those 20 instances carry 20 units of weight in the overall denominator—the template group must be marked in the report, to prevent inflating the denominator share of some task category via "multi-instance templates".
  - The **parameter domain of template instantiation must undergo deduplication and coverage checks**, to avoid generating semantically duplicated instances (following τ²-bench's stratified sampling by `write_profile` / `issue_persona`: stratify → fix the seed → freeze → generate digest).

- **R8** Multiple task sets coexist under one benchmark, each independently versioned, and the partition rules for `dev`, `test-public`, `test-private` and `stress` are enforced.

  **2.4 Task Set Partitioning**

  Multiple task sets coexist under one benchmark, each independently versioned:

  | Task set | Visibility | Purpose | Constraints |
  |---|---|---|---|
  | **Smoke set (smoke)** | Public | Integration self-check, CI smoke test | ≤ 10 items per domain, elapsed time < 10 minutes |
  | **Development set (dev)** | Public | Task authors debugging scorers, trial runs of tasks | May be called repeatedly; counted in no leaderboard |
  | **Public test set (test-public)** | Public | Routine external evaluation | Not modified after publication |
  | **Private test set (test-private)** | Not public | Official leaderboard, trusted evaluation | Scoring details and parts of the task content are not visible |
  | **Stress set (stress)** | Public | Long context, ultra long-horizon, robustness-specific | Scored separately, not merged into the total score |

  **Rules:**

  - **dev / test task IDs must not overlap.** Task authors must declare the partition when a task enters the store; the platform automatically verifies that the same task ID does not appear in two partitions.
  - **A public test set is frozen upon release.** Fixing task defects can only be done by publishing a new version of the task set; old-version scores are retained and annotated with the version number.
  - **The existence of a private test set means scores are not comparable across platforms**, so the report must prominently mark "this score comes from a private task set".
  - **The stress set is not merged into the total score**—its difficulty distribution differs from the regular sets, and mixing it in would destroy comparability.

- **R9** Difficulty uses a four-tier scheme derived from measurable criteria, and difficulty labels are validated against the baseline Agent before publication and re-calibrated every six months.

  **2.5 Difficulty Tiering**

  Difficulty uses a **four-tier** scheme, defined as follows (each tier gives an actionable criterion; no subjective adjectives are used):

  | Tier | Criterion (meeting any one places the task in that tier; the highest tier applies) |
  |---|---|
  | **L1 Simple** | Completed in one or two steps; no external dependencies; the Agent can answer without reading additional information |
  | **L2 Medium** | 3–8 steps; requires more than one tool; a single correct path exists; the failure cause is unique |
  | **L3 Hard** | 8–20 steps; requires combining multiple tools or crossing applications; multiple viable paths exist; requires reading material that is not explicitly specified |
  | **L4 Extremely Hard** | > 20 steps or requires long-horizon memory; the reference solution is itself fairly complex; the task has a clear time threshold for a real human (estimated human time cost > 20 minutes) |

  **Automatic validation of difficulty labels (prevents annotation drift):**

  - Before publishing a task set, run it once with the baseline Agent to obtain the actual success rate.
  - **Conflict alert between annotation and measurement**: labeled L1 but baseline success rate > 90% → downgrade alert; labeled L4 but baseline success rate > 70% → upgrade alert.
  - Re-run the baseline calibration every six months, producing a difficulty calibration report.

- **R10** Each task runs on an isolated environment instance through the four lifecycle stages, with per-stage failure attribution and read-only scoring against the final state snapshot.

  **2.6 Environment and State Isolation**

  Each task runs on an **isolated environment instance**; instances share no mutable state.

  **Four stages of the task lifecycle** (following MCPMark's Setup→Execute→Verify→Cleanup):

  | Stage | Content | Failure handling |
  |---|---|---|
  | **Preparation** | Bring up the environment instance, load initial data and snapshots, verify environment health | Failure → recorded as a platform fault, excluded from the denominator |
  | **Execution** | The Agent interacts with the environment until termination | Agent crash → recorded as a model failure, included in the denominator |
  | **Judgement** | Run the scorers, collect judgement evidence | Scorer exception → recorded as a judgement fault, counted separately |
  | **Cleanup** | Destroy the environment instance, reclaim credentials and temporary data | Cleanup failure → alert, manual intervention |

  **Isolation requirements:**

  - Each Trial uses an isolated environment instance and isolated credentials; reusing stateful resources across Trials is forbidden.
  - The final state snapshot required for scoring is produced in the "Judgement" stage; the scoring process reads the snapshot read-only and **must not be executed directly on the Agent's environment**.
  - For rollback-capable environments (VM/container snapshots), support "roll back to the initial snapshot and re-run scoring".

- **R11** Tasks enter the platform only through the three specified sources, each with its own admission requirement and additional constraints.

  **2.7 Task Sources and Production Process**

  Task sources fall into three categories, each with its own admission criteria:

  | Source | Admission requirement | Additional constraints |
  |---|---|---|
  | **External benchmark import** | The original benchmark's data and scoring semantics must map completely onto the platform model | After import, **the baseline must be re-run once** and a comparison score generated; the original benchmark score and the platform score are displayed side by side |
  | **Internal construction** | Goes through the full new-task validation process (see [§2.8][feat-002]) | Scoring code is invisible to the system under test |
  | **Community submission** | Goes through the full new-task validation process + manual review | The submitter is credited; submissions that fail review do not enter the official set |

- **R12** Every new task must pass all eight validation gates before entering `dev`, and must additionally pass 2–5 before entering `test`; tasks that fail validation are never allowed into the store.

  **2.8 Task Admission Validation Gates**

  **Any new task must pass all of the following before entering dev, and must additionally pass 2–5 before entering test.**

  | # | Validation item | Pass criterion | Basis |
  |---|---|---|---|
  | 1 | **Solvability** | A reference solution exists and runs through successfully; judged as a pass | Based on the oracle agent self-check of terminal-bench |
  | 2 | **Scorer credibility** | The reference solution passes + several "wrong solutions" (deliberately incorrect) all fail | Based on SWE-bench's fail-to-pass / pass-to-pass dual sets |
  | 3 | **Scorer non-triviality** | At least one scorer is not always-true; a no-op must not pass | Based on SWE-bench's evidence validation that "the test really ran" |
  | 4 | **Platform fault separation** | Deliberately induced environment faults are correctly marked as "excluded from the denominator" | Based on OSWorld `unscored_errors` and ClawBench `adjusted_eligible` |
  | 5 | **Difficulty credibility** | The measured success rate does not conflict with the annotated difficulty (2.5 rules) | Based on AgentBoard's difficulty tier statistics |
  | 6 | **Scoring invisibility** | The judgement conditions and judgement logic cannot be found within the Agent-accessible scope | Based on TheAgentCompany's encrypted scoring scripts |
  | 7 | **Annotation completeness** | Capability labels, domain labels, budget, and initial state have no missing items; the label mutual-exclusion check passes | Platform built-in check |
  | 8 | **Reproducibility** | Running twice with the same seed yields consistent judgement results (cost/duration variance allowed) | Platform built-in check |

  **Tasks that fail validation are not allowed into the store** and may only remain in draft state. The validation report is saved together with the task as part of the task metadata.

- **R13** Every task declares exactly one **form**, the platform rejects a task whose declared fields do not match its declared form, and the admission gates that depend on a live environment are applied only to `environment-agentic` tasks.

  **2.9 Task Forms**

  Two forms exist, and they differ in what the platform must bring up and what the verdict is read from:

  | Form | The Agent does | The platform brings up | The verdict is read from |
  |---|---|---|---|
  | **`environment-agentic`** | Acts through an action space on a live environment | An isolated environment instance per Trial | The environment's final state, or the artifact the Agent produced |
  | **`static-qa`** | Answers a prompt | Nothing | The answer text, compared with the task's matchers and rubric |

  **Why two forms rather than one.** A task set of plain question-and-answer items — a support FAQ, a knowledge check, a single-document reasoning question — has no environment to snapshot, no action space to declare and no environment failure mode to attribute. Forcing it into the environment form makes the task author invent a dummy environment and makes the platform's environment-isolation machinery pure overhead on every Trial, to produce a verdict that is really a text comparison. The `static-qa` form removes that cost while keeping the same Task → Trial → Trajectory → Verdict → Metric pipeline, so a static-QA score and an environment-agentic score stay on one scale and can be reported in the same table.

  **What a `static-qa` task declares:** the prompt given to the Agent; the termination condition (normally: the Agent has produced an answer, or the token budget is exhausted); its judgement conditions, which are matchers (exact / contains / regex) over the answer text, rubric items judged over the answer text, or both; and its budget.

  **What a `static-qa` task does not declare:** an environment snapshot, initial data, initial files, initial credentials, an action space, or an observation form. The platform rejects a `static-qa` task that carries any of them.

  **Form-dependent admission gates.** The eight admission gates of 2.8 are applied with the following form rule:

  | Gate | `environment-agentic` | `static-qa` |
  |---|---|---|
  | 1 Solvability | A reference solution runs through successfully | A reference answer exists and the scorer accepts it |
  | 2 Scorer credibility | The reference passes, deliberately wrong solutions fail | The reference answer passes, a deliberately wrong answer fails |
  | 3 Scorer non-triviality | ✔ | ✔ |
  | 4 Platform fault separation | ✔ | ✘ — a static-QA trial has no environment that can fail |
  | 5 Difficulty credibility | ✔ | ✔ |
  | 6 Scoring invisibility | ✔ | ✘ — there is no environment to hide the scoring logic in |
  | 7 Annotation completeness | ✔ | ✔ |
  | 8 Reproducibility | ✔ | ✔ |

  **Form and failure classification.** A `static-qa` Trial can still end in `model_failure`, `budget_exhausted`, `protocol_violation` or `cancelled`. It can never end in `platform_failure` for an environment reason, because no environment is brought up; a platform failure remains possible if the platform itself errors.

  **Domain assignment.** A `static-qa` task still carries a domain label. A closed-book knowledge question belongs to `D7 Knowledge and Reasoning`; a question answered by navigating a live web page is `D1` + `environment-agentic`; a question whose page content is pasted into the prompt is `D1` + `static-qa`.

## Acceptance Criteria

Written as verifiable outcomes.

- [ ] **AC1** (R3) Every task in the store carries at least 1 level-1 dimension label, 2 is recommended, and a task with more than 3 level-1 labels is treated as over-broad labeling and requires review; the label mutual-exclusion check is executed automatically before a task enters the store and a repeated level-2 label is rejected; labels are multi-select.
- [ ] **AC2** (R2, R3) Every level-2 label used on a task hits the task's actual failure point and is never inferred from "what this kind of task usually tests".
- [ ] **AC3** (R4) A task that spans several domains carries all of them in its domain list and is used to assess C1/C4.
- [ ] **AC4** (R5) Weights are declared explicitly and persisted alongside the results, and the weights at each of the two levels ("dimension → domain") are normalized to sum to 1.
- [ ] **AC5** (R5) A dimension for which no task was evaluated is marked "no coverage", is excluded from the weighting, and is highlighted in red in the report — it is never silently zero-filled — and a public leaderboard simultaneously shows the unweighted raw per-domain scores.
- [ ] **AC6** (R6) A task entering the store has every field group its declared form requires populated — for `environment-agentic`: identity, capability annotation, task description, budget, initial state, interaction contract, judgement conditions, reference material, meta information; for `static-qa`: identity, capability annotation, task description, budget, interaction contract (termination conditions and the multi-turn flag only), judgement conditions, meta information — and carries no field group its form does not use; a task ID is never assigned to a second task, and deprecated tasks are retained in the store marked as deprecated.
- [ ] **AC7** (R6) Judgement conditions and judgement logic are never found within the Agent-accessible environment, and "instructions aimed at the Agent" and "explanations aimed at evaluators" remain strictly separated, with the latter absent from the run.
- [ ] **AC8** (R7) Every instance generated from a template is scored independently and carries a stable instance ID; the post-instantiation distribution is declared, so a template yielding 20 instances contributes 20 units of weight to the overall denominator and the template group is marked in the report; the parameter domain passes deduplication and coverage checks, so no two generated instances are semantically duplicated.
- [ ] **AC9** (R8) The `smoke` set contains at most 10 items per domain and completes in under 10 minutes; the `dev` set is counted in no leaderboard; `test-public` is not modified after publication, and old-version scores are retained and annotated with the version number; the `stress` set is scored separately and is not merged into the total score.
- [ ] **AC10** (R8) The platform automatically verifies that the same task ID does not appear in two partitions, so `dev` / `test` task IDs never overlap, and a report built on the private task set prominently marks "this score comes from a private task set".
- [ ] **AC11** (R9) Before publishing a task set it is run once with the baseline Agent; a task labeled L1 with a baseline success rate > 90% raises a downgrade alert, and a task labeled L4 with a baseline success rate > 70% raises an upgrade alert; the baseline calibration is re-run every six months and produces a difficulty calibration report.
- [ ] **AC12** (R10) Each Trial uses an isolated environment instance and isolated credentials, and no stateful resource is reused across Trials; a preparation failure is recorded as a platform fault and excluded from the denominator, an Agent crash is recorded as a model failure and included in the denominator, a scorer exception is recorded as a judgement fault and counted separately, and a cleanup failure raises an alert requiring manual intervention.
- [ ] **AC13** (R10) The final state snapshot required for scoring is produced in the "Judgement" stage, the scoring process reads that snapshot read-only and is never executed directly on the Agent's environment, and for rollback-capable environments (VM/container snapshots) "roll back to the initial snapshot and re-run scoring" succeeds.
- [ ] **AC14** (R11) An externally imported benchmark has its data and scoring semantics mapped completely onto the platform model, the baseline is re-run once after import, a comparison score is generated, and the original benchmark score and the platform score are displayed side by side; a community submission that fails review does not enter the official set, and a submitting author of an accepted submission is credited.
- [ ] **AC15** (R12) A new task passes all 8 validation gates before entering `dev` — a reference solution that runs through successfully, a reference solution that passes while deliberately incorrect "wrong solutions" all fail, at least one scorer that is not always-true so a no-op must not pass, deliberately induced environment faults correctly marked "excluded from the denominator", measured success rate not conflicting with the annotated difficulty, judgement conditions and judgement logic not findable within the Agent-accessible scope, no missing items in capability labels / domain labels / budget / initial state with the label mutual-exclusion check passing, and consistent judgement results across two runs with the same seed (cost/duration variance allowed) — and additionally passes validation items 2–5 before entering `test`.
- [ ] **AC16** (R12) A task that fails validation never enters the store, remains in draft state only, and its validation report is saved together with the task as part of the task metadata.
- [ ] **AC17** (R13) A `static-qa` task with no environment snapshot, no initial data, no initial files, no initial credentials, no action space and no observation form is admitted into `dev` without being forced to declare any of them, and a `static-qa` task that declares any of those six fields is rejected with the offending field named.
- [ ] **AC18** (R13) The admission gate report for a `static-qa` task lists gates 1, 2, 3, 5, 7 and 8 as executed and gates 4 and 6 as "not applicable — no environment", and the report for an `environment-agentic` task lists all eight gates as executed.
- [ ] **AC19** (R13) A `static-qa` task set produces verdicts through the same Trial → Trajectory → Verdict → Metric pipeline as an `environment-agentic` task set, and the resulting metrics appear in the same report table with their task counts and coverage, so that the two forms are directly comparable within one run.

## Impact

| Area | Effect |
|---|---|
| Users | Task authors choose the form that matches their data. A batch of question-and-answer items is a task set with its own contract and gates, not a disguised environment task. Authors of environment tasks keep the full contract and lose nothing. |
| Existing behaviour | The metadata contract gains a form column, and the ingestion report gains a "not applicable — no environment" outcome for gates 4 and 6 on static-QA tasks. |
| Operational cost | A static-QA Trial brings up no environment, so it costs one Agent call and one judgement pass instead of a full environment lifecycle. |

## Risks and Mitigations

| Risk | Likelihood | Mitigation |
|---|---|---|
| The two forms drift apart, so a static-QA score and an environment-agentic score stop being comparable. | Med | Both forms run through the same Trial → Trajectory → Verdict → Metric pipeline; AC19 requires their metrics to appear in the same report table with their coverage. |
| A task author hides a real environment dependency inside a `static-qa` task, so the platform reports a text comparison while the Agent actually needed a browser. | Med | A static-QA task declares its prompt as the complete input; the Agent receives nothing else, so a missing environment dependency shows up as a failed Trial rather than as a wrong verdict. |
| The form becomes a dumping ground, and authors pick `static-qa` for tasks that need an environment, weakening the whole benchmark. | Med | Gates 1, 2 and 8 apply to both forms: a static-QA task must have a reference answer that the scorer accepts, must fail on a deliberately wrong answer, and must produce the same verdict on two runs with the same seed. |

## Rollout

The form column is introduced together with the metric registry and the M0 runtime. Existing task sets are migrated by labelling each task with the form it already satisfies; a task that satisfies neither form is rejected rather than guessed at. The rollout is reverted by removing the form column, which is safe because no existing task depends on it.

## References

- Tech Design: [TD-002 Task Forms and the Task Metadata Contract][td-002]
- Test Report: [TR-003 M0 Vertical Slice][tr-003]
- Task storage format: `[pending technical design]`.
- Sibling feature chapters cited by this document: [FEAT-006 Security, Robustness and Quality Assurance][feat-006].

## Open Questions

Not specified in the source design.

<!-- Link definitions stay at the bottom of the file. Angle brackets are required
     because "Tech Design" and "Test Report" contain spaces. -->

[feat-002]: <./capability-taxonomy-and-task-system.md>
[feat-006]: <../security/security-robustness-and-quality-assurance.md>
[td-002]: <../../Tech Design/TD-002-task-forms-and-metadata-contract.md>
[tr-003]: <../../Test Report/TR-003-m0-vertical-slice.md>
