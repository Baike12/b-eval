# 02 · Capability Dimensions and Task System

## 2.1 Capability Taxonomy Model

The platform uses a **dual-axis taxonomy**: the capability dimension axis answers "which capability is the Agent missing", and the domain axis answers "in which class of real-world scenarios the Agent fails". The two axes are labeled independently; any single task may fall into multiple capability dimensions and multiple domains at the same time.

### 2.1.1 Capability Dimension Axis (Level 1)

Merged from the four foundational capability threads of the Agent evaluation survey (planning / tool use / self-reflection / memory) and AgentBoard's six-dimension weighted score, divided into 6 level-1 dimensions:

| Dimension | Connotation | Typical assessment method |
|---|---|---|
| **C1 Planning and Reasoning** | Decomposing tasks, choosing paths, making decisions under constraints | Multi-step tasks, partial-completion scoring, step efficiency |
| **C2 Tool and Interface Use** | Calling tools correctly, correct arguments, combining multiple tools, handling call failures | Tool-call match rate, argument correctness rate, final-environment-state comparison |
| **C3 State and Memory Retention** | Tracking state across turns/steps, remembering historical information, resolving conflicting information | Progress rate, memory-type datasets, incremental multi-turn follow-ups |
| **C4 Interaction and Collaboration** | Dialoguing with users/other Agents, clarifying requirements, handling changes and interruptions | User simulator tasks, multi-Agent collaboration tasks |
| **C5 Error Handling and Self-Correction** | Locating the cause after failure, correcting, retrying without diverging | Tasks with injected failure branches, post-failure recovery rate |
| **C6 Robustness and Safety** | Resisting prompt injection, unauthorized operations, inappropriate content, recognizing infeasible tasks | Attack-defense task sets, infeasible tasks |

### 2.1.2 Capability Dimension Axis (Level 2, 18 items in total)

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

### 2.1.3 Domain Axis

Divided by "what the Agent operates on", consistent with the domain-bundling approach of τ²-bench:

| Domain | Coverage | Reference sources |
|---|---|---|
| **D1 Web and Browser** | Information retrieval, form filling, cross-site operations, visibility-sensitive web tasks | WebArena, VisualWebArena, ClawBench |
| **D2 Desktop and GUI Applications** | Operating systems, office software, image/media, IDEs | OSWorld, WindowsAgentArena |
| **D3 Code and Software Engineering** | Repository-level bug fixing, feature implementation, terminal operations, training and tuning | SWE-bench, terminal-bench, MLE-bench |
| **D4 Tools and APIs** | Function calling, REST/service integration, MCP services, enterprise systems | BFCL, MCPMark, τ-bench |
| **D5 Real Business Processes** | Customer service, e-commerce, travel, government affairs, and other end-to-end business processes | τ²-bench, VitaBench |
| **D6 Multi-Agent Collaboration** | Division of labor, handoff, collaborative delivery | TheAgentCompany, AgentVerse |

**Cross-domain tasks:** A task is allowed to span multiple domains (e.g. "compare prices on the web → produce a report in office software"), used to assess C1/C4.

### 2.1.4 Capability Score Composition Rules

The platform **does not produce a single "overall capability score"**, but provides configurable weighted views. Weights must be declared explicitly and persisted alongside the results; the rules are:

- Weights are given at two levels, "dimension → domain", and the weights at each level are normalized to sum to 1.
- **Missing coverage is never silently zero-filled**: when no task under a given dimension was evaluated, that dimension's score is marked "no coverage", excluded from the weighting, and highlighted in red in the report.
- Weighted scores are only used for internal regression comparison; **public leaderboards must simultaneously show the unweighted raw per-domain scores**.

## 2.2 Task Unit and Metadata

The complete definition of a task contains the following fields. Field names are functional naming only; the concrete storage format is specified in [pending technical design].

| Field group | Content | Required |
|---|---|---|
| **Identity** | Task ID (stable, globally unique, never reused), display name, owning benchmark, owning task set version | ✔ |
| **Capability annotation** | Level-1 dimension list, level-2 label list, difficulty level, domain list | ✔ |
| **Task description** | Natural-language instruction aimed at the Agent; explanation for evaluators other than the Agent (optional) | ✔ |
| **Initial state** | Environment snapshot identifier, initial data, initial files, initial credentials, random seed | ✔ |
| **Interaction contract** | Action space description, observation form, termination conditions, budget cap, whether multi-turn is allowed | ✔ |
| **Judgement conditions** | Scorer list, combination method, partial-credit rules, target state or reference solution | ✔ |
| **Reference material** | Reference solution (optional), reference trajectory (optional), domain policy document (optional) | ✖ |
| **Meta information** | Author, source, difficulty annotation basis, estimated human time cost, estimated Agent time cost, creation time, last revision time, deprecation status | ✔ |

**Design points:**

- **Task IDs are immutable and never reused.** Deprecated tasks are retained in the store and marked as deprecated; an ID is never assigned to a new task again—otherwise historical scores would point to a different task.
- **Difficulty is a required field**, derived from a combination of manual annotation + estimated time cost + actual failure rate; the decision rules are in 2.5.
- **"Instructions aimed at the Agent" and "explanations aimed at evaluators" are strictly separated.** The former participates in the run; the latter appears only in reports and review interfaces, preventing task leakage.
- **Judgement conditions are stored separately from the task description.** Judgement conditions must not appear inside the Agent-accessible environment (scoring must not be visible; see Chapter 06).

## 2.3 Task Templating and Parameterization

Following WebArena's approach (190 intent templates → 812 task instances), the platform supports task templating:

- A **template** defines the task skeleton: instruction template, action space, judgement method, difficulty.
- **Parameter instances** fill in concrete values, generating multiple task instances.
- **Instances must each be scored independently**, and each instance ID must itself be stable.
- **The distribution after template instantiation must be declared explicitly**: if one template yields 20 instances, those 20 instances carry 20 units of weight in the overall denominator—the template group must be marked in the report, to prevent inflating the denominator share of some task category via "multi-instance templates".
- The **parameter domain of template instantiation must undergo deduplication and coverage checks**, to avoid generating semantically duplicated instances (following τ²-bench's stratified sampling by `write_profile` / `issue_persona`: stratify → fix the seed → freeze → generate digest).

## 2.4 Task Set Partitioning

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

## 2.5 Difficulty Tiering

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

## 2.6 Environment and State Isolation

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

## 2.7 Task Sources and Production Process

Task sources fall into three categories, each with its own admission criteria:

| Source | Admission requirement | Additional constraints |
|---|---|---|
| **External benchmark import** | The original benchmark's data and scoring semantics must map completely onto the platform model | After import, **the baseline must be re-run once** and a comparison score generated; the original benchmark score and the platform score are displayed side by side |
| **Internal construction** | Goes through the full new-task validation process (see 2.8) | Scoring code is invisible to the system under test |
| **Community submission** | Goes through the full new-task validation process + manual review | The submitter is credited; submissions that fail review do not enter the official set |

## 2.8 Task Admission Validation Gates

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
