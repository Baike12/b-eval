# 04 · Scoring and Metrics

## 4.1 Scorer Categories

The platform abstracts scoring logic into six categories of scorers. A task declares its judgement conditions through "a reference to a scorer + combination rules", decoupling scorer implementations from tasks so that they are reusable across tasks and across benchmarks.

| Category | ID | Judgement basis | Typical implementation | Reference source |
|---|---|---|---|---|
| **State check** | J1 | Whether the final environment state matches expectations | Read designated fields from the environment snapshot and compare | The metric layer of OSWorld |
| **Result retrieval** | J2 | Whether the artifact content (file, database rows, page state) matches expectations | Retrieve the artifact, then apply comparison rules | The getter layer of OSWorld |
| **Action assertion** | J3 | Whether the Agent invoked the required action sequence and whether the parameters are correct | Match action names and parameters one by one | The ACTION criteria of τ²-bench |
| **Rule criteria** | J4 | A declarative rule set (numeric comparison, regex, set relations, structured comparison) | Rule declaration + generic executor | OSWorld's `_match_value_to_rule`, test parsing in terminal-bench |
| **Model scoring** | J5 | Judge trajectory quality using natural-language rubric items | Rubric items + judging model | LLM judgement in VitaBench and TheAgentCompany |
| **Human review** | J6 | A human gives a verdict on the trajectory/artifact | Annotation workbench | Expert review in AgentBoard, human preference in Chatbot Arena |

### 4.1.1 Unified Scorer Contract

Every scorer must return exactly one of the three unified result states:

| Result | Meaning | Counted in the denominator |
|---|---|---|
| **Pass (PASS)** | The criteria are satisfied | ✔ |
| **Fail (FAIL)** | The criteria are not satisfied | ✔ |
| **Unscored (UNSCORED)** | The scorer itself failed, the judging model is unavailable, or the criteria cannot be evaluated | ✘, and counted in the "judgement failure rate" |

**"Unscored" is a first-class state.** A judging model timeout must not be equated with "judged as not passing" — otherwise the platform's own stability problems turn into capability deductions for the Agent. See Inspect AI's `Score.unscored(reason="grader_failed")` and τ²-bench v2's practice of excluding infra errors from the denominator.

### 4.1.2 Key Design: Decoupling the Judgement Basis from the Judgement Method

Following OSWorld's binary decoupling, elevate it to a platform-wide principle:

- **"What to fetch" and "how to compare" are two separate actions.** J2 fetches the result (a file, a database row, a page state), J4 compares it by rule (equality, containment, numeric approximation, structured matching). Combined, they cover the vast majority of deterministic verdicts.
- A task declares a **reference** ("fetch X, compare Y"), not an implementation. Adding a new comparison rule requires no framework code change.

### 4.1.3 State Judgement vs Action Judgement

The platform clearly distinguishes two types of judgement and presents them separately in reports:

| | State judgement (outcome-oriented) | Action judgement (process-oriented) |
|---|---|---|
| The question asked | Did the Agent achieve the goal? | Did the Agent take the right path? |
| Multiple solutions allowed | ✔ Yes | ✘ No (unless explicitly declared as a "required action") |
| Judgement method | Environment final state / artifact content comparison | Action sequence and parameter matching |
| Risk | May misjudge (the Agent did something else useful) | May be too strict (equivalent paths judged wrong) |

**Only state judgement is performed by default.** Action judgement is enabled only when the task explicitly declares it, and the explicitly declared action sequence is labelled a "reference trajectory", whose semantics are **one optional equivalent path**, unless the task additionally declares it a "required action" (see τ²-bench's redefinition of `actions` from "the path that must be taken" to "a reference trajectory").

### 4.1.4 The Final-State Comparison Paradigm of "Reference Trajectory ≠ Required Path"

This is the default judgement method recommended by the platform and is especially important for tasks with multiple valid solutions:

1. Construct a **golden environment**: clean initial state + replay of the reference trajectory → obtain the target final state.
2. Construct an **environment under test**: clean initial state + replay of the Agent's actual trajectory → obtain the actual final state.
3. Compare the final states of the two environments.

Benefits:
- Any equivalent path taken by the Agent counts as a pass (different correct solutions are not penalised).
- No need to hand-write a scoring script for each task.
- During replay, **only actions that change state are executed**, and read-only actions are skipped — extra queries by the Agent do not affect the final-state comparison; tool names hallucinated by the Agent are treated as no-ops and skipped, consistent with what actually happened.

### 4.1.5 Rubric-Item-Based Scoring

For open-ended tasks that cannot be fully captured by state (e.g. "whether the reply is professional and covers all requirements"), the platform provides a rubric mechanism:

- **Rubric itemization**: a task may declare multiple rubric items, each an atomic requirement that can be judged independently.
- **Each rubric item judged independently**: a boolean conclusion + a rationale + supporting evidence citations (which turn, which tool call).
- **State can flip**: a rubric item may change from "satisfied" to "overturned" later in the trajectory (e.g. the user changed their mind, voiding an earlier order).
- **Sliding window**: when the trajectory is too long for a one-shot evaluation to be reliable, a rolling window with "window size W, overlap O" updates rubric state incrementally window by window; rubric state is preserved across windows. The overlap exists to prevent "decisive information landing exactly on a window boundary" from being missed.
- **When the judging model is unavailable, judge as UNSCORED, never as FAIL** (see the `unscored` handling in VitaBench and τ²-bench).

**Rubric anti-cheating criteria (built into the judging prompt template, not reliant on the task author's conscientiousness):**
- Only the Agent's replies and tool calls are examined; what the user said serves only as guidance, never as completion evidence.
- **Return values of query-type tools do not count as completion evidence** — the Agent finding the information does not mean the Agent completed the task.
- **Rubric items involving state changes must confirm that the change actually happened** (e.g. "must actually place the order"; the user saying "I'll buy it myself" does not count).
- Judging rationales must cite specific turns.

### 4.1.6 Scorer Composition and Partial Credit

- Multiple scorers are composed by **conjunction (all satisfied) / disjunction (any satisfied)**.
- **Conjunction + partial credit**: each scorer outputs 0/1 (or a decimal), and the total score = number satisfied / total number. The task is then no longer a binary pass/fail but a continuous score.
- **Scorer weights**: weights may be declared, and the total is a weighted average; the weights are stored with the task and displayed.
- **Primary criteria (primary)**: each task must designate at least one primary criterion, used for the binary success rate. The rest are auxiliary criteria that contribute only to the continuous score.

**Example (task "help the user modify an order in a retail system"):**

| Scorer | Type | Role |
|---|---|---|
| Environment final-state hash comparison | J1 | Primary criterion: the order state matches the golden final state |
| Information that must be stated (order number, updated amount) | J5 (rubric) | Auxiliary: information delivery |
| Required action (called the order-modification tool) | J3 | Auxiliary: correct path |
| No overreach (did not modify other orders) | J1 | Auxiliary: scope constraint |

## 4.2 Core Metric Definitions

The computation conventions of all metrics are defined uniquely in this section; reports, leaderboards, and export formats all reference this section.

### 4.2.1 Notation

| Notation | Meaning |
|---|---|
| `T` | Total number of tasks in the task set |
| `t` | A single task |
| `n_t` | Number of valid trials for task t (platform failures already excluded) |
| `c_t` | Number of trials judged as passing for task t |
| `k` | Number of repetitions (default 1) |
| `S_t` | Set of success-criterion results for task t (the boolean "primary criterion passes" when there are multiple criteria) |
| `P_t` | Progress score of task t (0–1, see 4.2.3) |

### 4.2.2 Success Metrics

| Metric | Formula | Denominator | Meaning |
|---|---|---|---|
| **Success rate SR** | `Σ_t c_t / Σ_t n_t` | Total valid trials | Task-level success ratio (identical to pass@1 when k=1) |
| **pass@k** | `1 - Π_{i=0}^{k-1} (n_t - c_t - i) / (n_t - i)` (recorded as NA when n_t < k) | Number of tasks | Probability of at least one success in k attempts (unbiased estimate) |
| **pass^k** | `C(c_t, k) / C(n_t, k)` (recorded as NA when n_t < k) | Number of tasks | Probability that all k attempts succeed |
| **avg@k** | `Σ_t (Σ_i reward_{t,i}) / Σ_t n_t` | Total valid trials | Average score (especially useful for continuous-score tasks) |
| **Task pass rate** | `|{t : c_t ≥ 1}| / T` | Number of tasks | Share of tasks that succeeded at least once |

**The trade-off between pass@k and pass^k:**
- pass@k answers "could this task be solved at all", reflecting the upper bound of capability.
- pass^k answers "can this task be solved reliably", reflecting reliability.
- **Both must be reported together**; reporting only one of them is misleading (see τ²-bench's practice of reporting pass^1..pass^k simultaneously).

**Handling of k < n_t:** when some tasks have fewer than k trials, their pass@k / pass^k is recorded as NA (excluded from that average), and the **valid task coverage** is reported — so that "tasks that only succeeded once" are not silently dropped.

### 4.2.3 Progress Score (Continuous Score)

Partial completion must earn credit, otherwise the Agent is treated identically whether it "completes 80%" or "completes 0%".

The definition of the progress score is declared by the task; the platform provides three standard forms:

| Form | Definition | Applicable to |
|---|---|---|
| **Subgoal achievement rate** | Number of achieved subgoals / total number of subgoals | Tasks with explicit subgoals |
| **Environment progress** | The progress signal given by the environment itself (monotonically non-decreasing) | Environments that natively support progress feedback |
| **Judgement-weighted score** | Each scorer's score weighted by its weight | Tasks with multiple criteria |

**Hard requirements:**
- The progress score must be **monotonically non-decreasing** (it must not regress within the same trial), preventing the Agent from farming score through regressive actions.
- The upper bound of the progress score is consistent with the primary criterion: **for a task whose primary criterion does not pass, the final score is still capped by the primary criterion** (e.g. if the primary criterion requires an exactly identical final state, the progress score is merely auxiliary information and is not used for "partial credit").
- Reports must present both a **binary success rate** and a **continuous average score** curve (see AgentBoard's dual-track approach).

### 4.2.4 Efficiency and Cost Metrics

| Metric | Definition | Note |
|---|---|---|
| **Average steps per successful task** | Average number of actions over tasks judged as passing | Compared against the baseline of the same task set |
| **Average tokens per successful task** | Same, measured in tokens | |
| **Average cost per successful task** | Same, measured in cost | The price table is persisted alongside the results |
| **Success-rate–cost frontier** | A multi-point curve with cost on the x-axis and success rate on the y-axis | Referencing the proposition of "optimising accuracy and cost simultaneously" |
| **Ineffective action rate** | Number of actions producing no state change / total number of actions | Measures the degree of divergence |
| **Repeated action rate** | Share of consecutive identical or equivalent actions | Referencing WebArena's early-stop rule for repeated actions |
| **Action format error rate** | Number of responses that cannot be parsed into legal actions / total number of responses | Measures adapter quality and Agent stability |
| **Grounding accuracy** | Number of legal actions accepted by the environment / total number of actions | Referencing AgentBoard's grounding_acc |

**Cost accounting requirement:** cost is measured in full as "model input tokens + output tokens + judging model consumption + user simulator consumption"; **scoring overhead must also be counted** — otherwise the strategy of "spending more judge cost to buy a higher score" stays invisible.

### 4.2.5 Failure Attribution Metrics

| Metric | Definition |
|---|---|
| **Termination reason distribution** | Share of each termination type (completed / budget exhausted / protocol violation / model failure / platform failure / cancelled) |
| **Judgement failure rate** | Share of trials where the scorer returns UNSCORED |
| **Platform failure rate** | Share of trials with platform failures (**not counted in the denominator, but must be reported**) |
| **Per-scorer failure breakdown** | Share of PASS / FAIL / UNSCORED for each scorer, used to locate which criterion is unreliable |

### 4.2.6 Reliability and Confidence Intervals

- **All ratio-type metrics must carry a confidence interval** (95% by default), and the interval computation method is configurable (binomial normal, Wilson, bootstrap resampling).
- **Variance across repeated runs**: `std` and `stderr`, used to judge whether "A is 1 point higher than B" is merely noise.
- **Judge consistency**: when multiple judging models vote, report the agreement rate and a consistency coefficient (e.g. Krippendorff's alpha); task sets below the threshold are marked as "unreliable scoring".

## 4.3 Decomposition and Aggregation Conventions

### 4.3.1 Decomposition Dimensions

The same score must be decomposable along the following dimensions:

| Dimension | Basis |
|---|---|
| Domain | The task's domain label (2.1.3) |
| Level-1 capability dimension | The task's capability label (2.1.1) |
| Level-2 capability tag | The task's level-2 label (2.1.2) |
| Difficulty | The task's difficulty tier (2.5) |
| Task set split | smoke / dev / test-public / test-private / stress |

**Report on every decomposition cell:** sample size, success rate, continuous mean score, confidence interval.

### 4.3.2 Aggregation Conventions

Two aggregation methods, of which **one must be explicitly selected and labelled in the report**:

| Method | Definition | Applicable to |
|---|---|---|
| **Trial-weighted (micro)** | All trials have equal weight | When all tasks are equally important |
| **Task-equal-weighted (macro)** | Compute each task's score first, then average across tasks | Avoids a task with many trials dominating the result |

**macro (task-equal-weighted) is used by default.** Rationale: the number of trials is determined by the evaluator's configuration and must not be allowed to determine task weights. The report must also give the micro figures for comparison.

**Group equal-weighting principle:** when performing a Cartesian decomposition across multiple dimensions, every cell has equal weight, avoiding cells with many tasks from dominating the total score (see τ²-bench's cell equal-weight aggregation and AgentCompass's `category_mean`).

### 4.3.3 Missing Coverage Handling

- **Cells with no coverage do not participate in weighting**, are marked as "no coverage", and are shown explicitly in the report (not silently filled with 0).
- **Excluded trials (platform failures) are not counted in the denominator**, but their number and reasons must be displayed.
- **NA pass@k items do not participate in that average**, and the valid task coverage for that item must be reported.

## 4.4 Usage Rules for Judge (Model Scoring)

Model scoring is itself a step that introduces noise, so the platform regulates its use as follows:

| Rule | Content |
|---|---|
| **The judging model is independent of the model under test** | By default it is prohibited to use the model under test to score itself; if this is genuinely required, it must be prominently flagged in the report |
| **The judging model is configurable and recorded** | The judging model identifier, version, judging temperature, and prompt template version are all persisted alongside the results |
| **Temperature fixed at 0** | Judging is deterministic by default; when the judging model does not support this, it must be noted in the report |
| **Injection resistance** | The text under test that is fed into the scorer is wrapped in delimiters, and the delimiters themselves are escaped within the text, preventing the Agent from forging delimiters in its output to fool the scorer |
| **Judging output must be parseable** | A fixed format is used to extract the conclusion (e.g. `GRADE: C/I/P`); only the final conclusion is accepted, preventing parse ambiguity caused by changing one's answer mid-reasoning |
| **Partial credit** | Three-level conclusions are supported (pass / partial / fail), with partial mapped to 0.5 |
| **Judging failure = UNSCORED** | A judging model error, timeout, or unparseable output is always UNSCORED, never FAIL, and must not be retried beyond the permitted number of attempts |
| **Judging quality must be calibrated** | Before deployment, a judging model must be evaluated for agreement rate on a human-annotated set; below the threshold it must not be used for formal evaluation |

## 4.5 Human Review

- **Review workbench**: displays the trajectory, scoring evidence, rubric items, and the current verdict, and supports "uphold / revise / flag as disputed".
- **Revision trail**: each revision records the reviser, the time, the reason, the original verdict, and the new verdict. Verdict history cannot be deleted, only appended.
- **Review sampling rules**:
  - **Mandatory review**: all UNSCORED trials, all boundary samples (scores close to the threshold), and the first run of every high-value task.
  - **Random review**: randomly sampled at a configurable ratio. Review results are used to estimate the deviation between automatic scoring and human scoring, producing the "automatic scoring deviation rate" metric.
- **Reviews are an input to scorer calibration**: reviewed samples enter the human-annotated set and are used for the judge calibration in 4.4.

## 4.6 Metric Computation Example (Convention Consistency Check)

Let the task set be T = {t1, t2, t3}, k = 3:

| Task | n_t | c_t | Mean progress score |
|---|---:|---:|---:|
| t1 | 3 | 3 | 1.00 |
| t2 | 3 | 1 | 0.55 |
| t3 | 2 (1 platform failure excluded) | 0 | 0.10 |

Computation:

- **SR** = (3 + 1 + 0) / (3 + 3 + 2) = 4/8 = 0.500
- **pass@3**: t1 = 1, t2 = 1 − C(2,3)/C(3,3) = 1, t3 = NA (n_3 < 3) → valid task coverage 2/3 → pass@3 = (1 + 1)/2 = 1.000
- **pass^3**: t1 = C(3,3)/C(3,3) = 1, t2 = C(1,3)/C(3,3) = 0, t3 = NA → pass^3 = (1 + 0)/2 = 0.500
- **avg@3** = (3×1.00 + 3×0.55 + 2×0.10) / 8 = (3.00 + 1.65 + 0.20)/8 = 4.85/8 = 0.606
- **macro success rate** = (1.000 + 0.333 + 0.000)/3 = 0.444 (differs from micro's 0.500; both must be given)
- **Platform failure rate** = 1 / 9 = 11.1% → **exceeds the 5% threshold → this report is flagged as "insufficient confidence"**

This example shows that **macro and micro yield different conclusions, and the platform failure rate directly invalidates the credibility of the report**. Both points must be presented in the report simultaneously.
