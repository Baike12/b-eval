# FEAT-004: Scoring and Metrics

| Field | Value |
|---|---|
| **ID** | FEAT-004 |
| **Status** | Draft |
| **Owner** | Platform team |
| **Priority** | P0 |
| **Implements** | `F-JDG-01`, `F-JDG-02`, `F-JDG-03`, `F-JDG-04`, `F-JDG-05`, `F-JDG-06`, `F-JDG-07`, `F-MTR-01`, `F-MTR-02`, `F-MTR-03`, `F-MTR-04`, `F-MTR-05`, `F-MTR-06` |
| **Created** | 2026-10-22 |
| **Last Updated** | 2026-10-22 |
| **Source** | Functional design chapter 04 — Scoring and Metrics |

## Summary

FEAT-004 defines the platform's scoring and metrics layer: six scorer categories (J1–J6) behind one unified three-state result contract, one unique set of metric definitions (SR, pass@k, pass^k, avg@k, progress score, efficiency, cost, failure attribution, reliability), and the decomposition and aggregation conventions that every report, leaderboard, and export format must reference. It also fixes the usage rules for model judging, the human review workbench, and a worked computation example that pins down the macro/micro and platform-failure semantics. The single most important outcome is that every task, trial, and report gets exactly one unambiguous, reproducible verdict and one canonical number.

## Problem

**Current behaviour:** Judgement and reporting rules are spread across individual benchmark implementations, each with its own scorer semantics, its own metric formulas, and its own aggregation choice. There is no single place where the computation conventions of all metrics are defined, so reports, leaderboards, and export formats can each adopt a different reading of the same score. Infrastructure failures are also easily mistaken for capability failures, because a scorer failure and a failed criterion are not distinguished.

**Pain:** Task authors cannot reuse scorer implementations across tasks and across benchmarks, because a task has to be written against a concrete implementation rather than a declared reference. Agents are hurt whenever a judging model timeout, a platform failure, or an equivalently correct path is scored as a failure — the platform's own stability problems turn into capability deductions for the Agent. Report consumers are hurt when only one of pass@k and pass^k is reported, when micro and macro aggregation are mixed without being labelled, or when cells with no coverage are silently filled with 0, because the published number then does not describe what happened. A report whose platform failure rate exceeds 5% is not distinguishable from a reliable report at all, which directly invalidates the credibility of the whole evaluation.

**Evidence:** The source design takes each construct from an existing reference implementation rather than inventing one: OSWorld's metric layer, getter layer, and `_match_value_to_rule`; Inspect AI's `Score.unscored(reason="grader_failed")`; τ²-bench and τ²-bench v2 (ACTION criteria, `actions` redefined as a reference trajectory, exclusion of infra errors from the denominator, simultaneous reporting of pass^1..pass^k, cell equal-weight aggregation); terminal-bench's test parsing; VitaBench and TheAgentCompany's LLM judgement and `unscored` handling; AgentBoard's grounding_acc and dual-track approach; WebArena's early-stop rule for repeated actions; AgentCompass's `category_mean`; Chatbot Arena's human preference. The worked example in 4.6 is itself the evidence that the conventions are in tension: macro and micro yield different conclusions (0.444 vs 0.500) and a platform failure rate of 11.1% exceeds the 5% threshold, so the report must be flagged as "insufficient confidence".

## Proposed Solution

The platform exposes a fixed catalogue of six scorer categories (J1–J6) and a single unified result contract — exactly one of PASS, FAIL, or UNSCORED per scorer call — that tasks reference by "a scorer reference + combination rules" instead of by implementation, so scorers become reusable across tasks and benchmarks. Judgement is outcome-oriented by default: the final state of the environment under test is compared against the final state of a golden environment, and an explicitly declared action sequence is treated as one optional equivalent path rather than a required path. All metrics are then computed from one authoritative definition set, decomposed along fixed dimensions, aggregated with an explicitly labelled method, and reported with confidence intervals, failure attribution, and the cost of the scoring process itself included.

## Scope

### In Scope

- Six scorer categories with a fixed judgement basis, typical implementation, and reference source: `J1` state check, `J2` result retrieval, `J3` action assertion, `J4` rule criteria, `J5` model scoring, `J6` human review.
- The unified three-state scorer result contract (PASS / FAIL / UNSCORED) and the treatment of UNSCORED as a first-class state, counted in the "judgement failure rate" and excluded from the success denominator.
- The decoupling of the judgement basis ("what to fetch") from the judgement method ("how to compare"), so a task declares a reference and not an implementation.
- The separation of state judgement (outcome-oriented) from action judgement (process-oriented), presented separately in reports, with only state judgement performed by default and action judgement enabled only when the task explicitly declares it.
- The final-state comparison paradigm "Reference Trajectory ≠ Required Path" via a golden environment and an environment under test.
- Rubric-item-based scoring for open-ended tasks: itemization, independent judging, state flips, sliding-window evaluation with overlap, UNSCORED on judge unavailability, and the rubric anti-cheating criteria.
- Scorer composition by conjunction / disjunction, partial credit, declared scorer weights, and a mandatory primary criterion per task.
- The metric definition set: notation, success metrics (`SR`, `pass@k`, `pass^k`, `avg@k`, task pass rate), progress score forms, efficiency and cost metrics, failure attribution metrics, and reliability / confidence intervals.
- Decomposition dimensions, micro and macro aggregation conventions with the macro default, group equal-weighting, and missing coverage handling.
- Usage rules for judge (model scoring) and the human review workbench, revision trail, sampling rules, and judge-calibration feedback loop.
- The metric computation example that serves as a convention consistency check, including the 5% platform-failure threshold and the "insufficient confidence" flag.

### Out of Scope

- Not specified in the source design.

## Requirements

- **R1** The platform shall abstract scoring logic into six scorer categories, and every task shall declare its judgement conditions through "a reference to a scorer + combination rules" so that scorer implementations stay decoupled from tasks and remain reusable across tasks and across benchmarks.

    #### 4.1 Scorer Categories

    The platform abstracts scoring logic into six categories of scorers. A task declares its judgement conditions through "a reference to a scorer + combination rules", decoupling scorer implementations from tasks so that they are reusable across tasks and across benchmarks.

    | Category | ID | Judgement basis | Typical implementation | Reference source |
    |---|---|---|---|---|
    | **State check** | J1 | Whether the final environment state matches expectations | Read designated fields from the environment snapshot and compare | The metric layer of OSWorld |
    | **Result retrieval** | J2 | Whether the artifact content (file, database rows, page state) matches expectations | Retrieve the artifact, then apply comparison rules | The getter layer of OSWorld |
    | **Action assertion** | J3 | Whether the Agent invoked the required action sequence and whether the parameters are correct | Match action names and parameters one by one | The ACTION criteria of τ²-bench |
    | **Rule criteria** | J4 | A declarative rule set (numeric comparison, regex, set relations, structured comparison) | Rule declaration + generic executor | OSWorld's `_match_value_to_rule`, test parsing in terminal-bench |
    | **Model scoring** | J5 | Judge trajectory quality using natural-language rubric items | Rubric items + judging model | LLM judgement in VitaBench and TheAgentCompany |
    | **Human review** | J6 | A human gives a verdict on the trajectory/artifact | Annotation workbench | Expert review in AgentBoard, human preference in Chatbot Arena |

- **R2** Every scorer shall return exactly one of the three unified result states, and UNSCORED shall be a first-class state that is never equated with a failed judgement.

    #### 4.1.1 Unified Scorer Contract

    Every scorer must return exactly one of the three unified result states:

    | Result | Meaning | Counted in the denominator |
    |---|---|---|
    | **Pass (PASS)** | The criteria are satisfied | ✔ |
    | **Fail (FAIL)** | The criteria are not satisfied | ✔ |
    | **Unscored (UNSCORED)** | The scorer itself failed, the judging model is unavailable, or the criteria cannot be evaluated | ✘, and counted in the "judgement failure rate" |

    **"Unscored" is a first-class state.** A judging model timeout must not be equated with "judged as not passing" — otherwise the platform's own stability problems turn into capability deductions for the Agent. See Inspect AI's `Score.unscored(reason="grader_failed")` and τ²-bench v2's practice of excluding infra errors from the denominator.

- **R3** The platform shall elevate the decoupling of the judgement basis from the judgement method to a platform-wide principle, so that a task declares a reference and never an implementation.

    #### 4.1.2 Key Design: Decoupling the Judgement Basis from the Judgement Method

    Following OSWorld's binary decoupling, elevate it to a platform-wide principle:

    - **"What to fetch" and "how to compare" are two separate actions.** J2 fetches the result (a file, a database row, a page state), J4 compares it by rule (equality, containment, numeric approximation, structured matching). Combined, they cover the vast majority of deterministic verdicts.
    - A task declares a **reference** ("fetch X, compare Y"), not an implementation. Adding a new comparison rule requires no framework code change.

- **R4** The platform shall clearly distinguish state judgement from action judgement, shall present the two separately in reports, and shall perform only state judgement by default, enabling action judgement only when the task explicitly declares it.

    #### 4.1.3 State Judgement vs Action Judgement

    The platform clearly distinguishes two types of judgement and presents them separately in reports:

    | | State judgement (outcome-oriented) | Action judgement (process-oriented) |
    |---|---|---|
    | The question asked | Did the Agent achieve the goal? | Did the Agent take the right path? |
    | Multiple solutions allowed | ✔ Yes | ✘ No (unless explicitly declared as a "required action") |
    | Judgement method | Environment final state / artifact content comparison | Action sequence and parameter matching |
    | Risk | May misjudge (the Agent did something else useful) | May be too strict (equivalent paths judged wrong) |

    **Only state judgement is performed by default.** Action judgement is enabled only when the task explicitly declares it, and the explicitly declared action sequence is labelled a "reference trajectory", whose semantics are **one optional equivalent path**, unless the task additionally declares it a "required action" (see τ²-bench's redefinition of `actions` from "the path that must be taken" to "a reference trajectory").

- **R5** The platform shall implement the final-state comparison paradigm "Reference Trajectory ≠ Required Path" as the default judgement method, by comparing the final state of a golden environment with the final state of an environment under test.

    #### 4.1.4 The Final-State Comparison Paradigm of "Reference Trajectory ≠ Required Path"

    This is the default judgement method recommended by the platform and is especially important for tasks with multiple valid solutions:

    1. Construct a **golden environment**: clean initial state + replay of the reference trajectory → obtain the target final state.
    2. Construct an **environment under test**: clean initial state + replay of the Agent's actual trajectory → obtain the actual final state.
    3. Compare the final states of the two environments.

    Benefits:
    - Any equivalent path taken by the Agent counts as a pass (different correct solutions are not penalised).
    - No need to hand-write a scoring script for each task.
    - During replay, **only actions that change state are executed**, and read-only actions are skipped — extra queries by the Agent do not affect the final-state comparison; tool names hallucinated by the Agent are treated as no-ops and skipped, consistent with what actually happened.

- **R6** The platform shall provide rubric-item-based scoring for open-ended tasks, judging each rubric item independently, allowing rubric state to flip, supporting sliding-window evaluation, and building the rubric anti-cheating criteria into the judging prompt template.

    #### 4.1.5 Rubric-Item-Based Scoring

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

- **R7** The platform shall compose multiple scorers by conjunction or disjunction, shall support partial credit and declared scorer weights, and shall require every task to designate at least one primary criterion.

    #### 4.1.6 Scorer Composition and Partial Credit

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

- **R8** All metric computation conventions shall be defined uniquely in one place, which reports, leaderboards, and export formats all reference, and the platform notation shall be fixed as specified.

    #### 4.2 Core Metric Definitions

    The computation conventions of all metrics are defined uniquely in this section; reports, leaderboards, and export formats all reference this section.

    #### 4.2.1 Notation

    | Notation | Meaning |
    |---|---|
    | `T` | Total number of tasks in the task set |
    | `t` | A single task |
    | `n_t` | Number of valid trials for task t (platform failures already excluded) |
    | `c_t` | Number of trials judged as passing for task t |
    | `k` | Number of repetitions (default 1) |
    | `S_t` | Set of success-criterion results for task t (the boolean "primary criterion passes" when there are multiple criteria) |
    | `P_t` | Progress score of task t (0–1, see 4.2.3) |

- **R9** The platform shall compute and report the success metrics `SR`, `pass@k`, `pass^k`, `avg@k`, and the task pass rate using the defined formulas and denominators, and shall always report `pass@k` and `pass^k` together.

    #### 4.2.2 Success Metrics

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

- **R10** The platform shall support the progress score (continuous score) in three standard forms, and shall enforce its hard requirements on monotonicity, the primary-criterion cap, and dual-track reporting.

    #### 4.2.3 Progress Score (Continuous Score)

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

- **R11** The platform shall compute the efficiency and cost metrics, and shall account for cost in full, including the scoring overhead.

    #### 4.2.4 Efficiency and Cost Metrics

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

- **R12** The platform shall compute the failure attribution metrics, including the judgement failure rate, the platform failure rate, and the per-scorer failure breakdown, and shall report platform failures even though they are excluded from the denominators.

    #### 4.2.5 Failure Attribution Metrics

    | Metric | Definition |
    |---|---|
    | **Termination reason distribution** | Share of each termination type (completed / budget exhausted / protocol violation / model failure / platform failure / cancelled) |
    | **Judgement failure rate** | Share of trials where the scorer returns UNSCORED |
    | **Platform failure rate** | Share of trials with platform failures (**not counted in the denominator, but must be reported**) |
    | **Per-scorer failure breakdown** | Share of PASS / FAIL / UNSCORED for each scorer, used to locate which criterion is unreliable |

- **R13** All ratio-type metrics shall carry a confidence interval (95% by default), and the platform shall report variance across repeated runs and judge consistency.

    #### 4.2.6 Reliability and Confidence Intervals

    - **All ratio-type metrics must carry a confidence interval** (95% by default), and the interval computation method is configurable (binomial normal, Wilson, bootstrap resampling).
    - **Variance across repeated runs**: `std` and `stderr`, used to judge whether "A is 1 point higher than B" is merely noise.
    - **Judge consistency**: when multiple judging models vote, report the agreement rate and a consistency coefficient (e.g. Krippendorff's alpha); task sets below the threshold are marked as "unreliable scoring".

- **R14** The same score shall be decomposable along the defined dimensions, and every decomposition cell shall report sample size, success rate, continuous mean score, and confidence interval.

    #### 4.3 Decomposition and Aggregation Conventions

    #### 4.3.1 Decomposition Dimensions

    The same score must be decomposable along the following dimensions:

    | Dimension | Basis |
    |---|---|
    | Domain | The task's domain label ([2.1.3][feat-002]) |
    | Level-1 capability dimension | The task's capability label ([2.1.1][feat-002]) |
    | Level-2 capability tag | The task's level-2 label ([2.1.2][feat-002]) |
    | Difficulty | The task's difficulty tier ([2.5][feat-002]) |
    | Task set split | smoke / dev / test-public / test-private / stress |

    **Report on every decomposition cell:** sample size, success rate, continuous mean score, confidence interval.

- **R15** The platform shall support exactly two aggregation methods, shall have one of them explicitly selected and labelled in the report, and shall use macro (task-equal-weighted) by default while also giving the micro figures.

    #### 4.3.2 Aggregation Conventions

    Two aggregation methods, of which **one must be explicitly selected and labelled in the report**:

    | Method | Definition | Applicable to |
    |---|---|---|
    | **Trial-weighted (micro)** | All trials have equal weight | When all tasks are equally important |
    | **Task-equal-weighted (macro)** | Compute each task's score first, then average across tasks | Avoids a task with many trials dominating the result |

    **macro (task-equal-weighted) is used by default.** Rationale: the number of trials is determined by the evaluator's configuration and must not be allowed to determine task weights. The report must also give the micro figures for comparison.

    **Group equal-weighting principle:** when performing a Cartesian decomposition across multiple dimensions, every cell has equal weight, avoiding cells with many tasks from dominating the total score (see τ²-bench's cell equal-weight aggregation and AgentCompass's `category_mean`).

- **R16** The platform shall handle missing coverage explicitly, so that uncovered cells, excluded trials, and NA pass@k items are never silently filled or silently dropped.

    #### 4.3.3 Missing Coverage Handling

    - **Cells with no coverage do not participate in weighting**, are marked as "no coverage", and are shown explicitly in the report (not silently filled with 0).
    - **Excluded trials (platform failures) are not counted in the denominator**, but their number and reasons must be displayed.
    - **NA pass@k items do not participate in that average**, and the valid task coverage for that item must be reported.

- **R17** The platform shall regulate the use of the judge (model scoring) according to the rules of 4.4, covering judging-model independence, configuration recording, temperature, injection resistance, output parsing, partial credit, failure handling, and calibration.

    #### 4.4 Usage Rules for Judge (Model Scoring)

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

- **R18** The platform shall provide a human review workbench with an append-only revision trail, mandatory and random review sampling rules, and a feedback path from reviews into judge calibration.

    #### 4.5 Human Review

    - **Review workbench**: displays the trajectory, scoring evidence, rubric items, and the current verdict, and supports "uphold / revise / flag as disputed".
    - **Revision trail**: each revision records the reviser, the time, the reason, the original verdict, and the new verdict. Verdict history cannot be deleted, only appended.
    - **Review sampling rules**:
        - **Mandatory review**: all UNSCORED trials, all boundary samples (scores close to the threshold), and the first run of every high-value task.
        - **Random review**: randomly sampled at a configurable ratio. Review results are used to estimate the deviation between automatic scoring and human scoring, producing the "automatic scoring deviation rate" metric.
    - **Reviews are an input to scorer calibration**: reviewed samples enter the human-annotated set and are used for the judge calibration in 4.4.

- **R19** The platform shall treat the metric computation example as a convention consistency check, and shall flag a report as "insufficient confidence" when the platform failure rate exceeds the 5% threshold.

    #### 4.6 Metric Computation Example (Convention Consistency Check)

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

## Acceptance Criteria

- [ ] **AC1** The platform exposes exactly the six scorer categories `J1`, `J2`, `J3`, `J4`, `J5`, and `J6`, each with the judgement basis, typical implementation, and reference source given in 4.1. (R1)
- [ ] **AC2** A task declares its judgement conditions as "a reference to a scorer + combination rules" and never as an implementation; adding a new comparison rule requires no framework code change. (R1, R3)
- [ ] **AC3** Every scorer call returns exactly one of PASS, FAIL, or UNSCORED; no fourth state, empty verdict, or ambiguous verdict is possible. (R2)
- [ ] **AC4** A trial whose scorer returns UNSCORED is excluded from the success denominator and counted in the "judgement failure rate"; a judging model timeout yields UNSCORED and never FAIL. (R2)
- [ ] **AC5** A task that does not explicitly declare an action sequence is judged by state judgement only; no action judgement is executed for it. (R4)
- [ ] **AC6** An explicitly declared action sequence is labelled a "reference trajectory" and treated as one optional equivalent path, unless the task additionally declares it a "required action". (R4)
- [ ] **AC7** Judgement of a task with a golden trajectory compares the golden final state against the actual final state, and any equivalent path reaching the same final state is a pass. (R5)
- [ ] **AC8** During replay, only actions that change state are executed; read-only actions and tool names hallucinated by the Agent are skipped and do not affect the final-state comparison. (R5)
- [ ] **AC9** Each rubric item is judged independently and yields a boolean conclusion, a rationale, and supporting evidence citations, and a rubric item may flip from "satisfied" to "overturned" later in the trajectory. (R6)
- [ ] **AC10** Rubric state is preserved across the windows of the rolling window ("window size W, overlap O"), and the overlap prevents decisive information landing exactly on a window boundary from being missed. (R6)
- [ ] **AC11** When the judging model is unavailable, rubric items are judged UNSCORED and never FAIL. (R6)
- [ ] **AC12** The judging prompt template treats only the Agent's replies and tool calls as evidence, refuses query-type tool return values as completion evidence, requires rubric items involving state changes to confirm the change actually happened, and requires rationales to cite specific turns. (R6)
- [ ] **AC13** Every task designates at least one primary criterion used for the binary success rate; auxiliary criteria contribute only to the continuous score. (R7)
- [ ] **AC14** For a conjunction task the total score equals the number of satisfied scorers divided by the total number of scorers, and declared scorer weights are stored with the task and displayed. (R7)
- [ ] **AC15** Metric computation conventions exist in exactly one definition set, and every report, leaderboard, and export format takes its formulas and denominators from that set; no second definition of the same metric is in use. (R8)
- [ ] **AC16** `SR`, `pass@k`, `pass^k`, `avg@k`, and the task pass rate are computed with the formulas and denominators of 4.2.2; `pass@k` and `pass^k` are always reported together; when `n_t < k` the item is recorded as NA and the valid task coverage is reported. (R9)
- [ ] **AC17** The progress score never regresses within the same trial; for a task whose primary criterion does not pass, the final score is still capped by the primary criterion; every report presents both a binary success rate and a continuous average score curve. (R10)
- [ ] **AC18** Recorded cost equals "model input tokens + output tokens + judging model consumption + user simulator consumption" plus the scoring overhead, and the price table is persisted alongside the results. (R11)
- [ ] **AC19** The platform failure rate is reported for every run even though platform failures are excluded from the denominators, and the per-scorer PASS / FAIL / UNSCORED breakdown identifies which criterion is unreliable. (R12)
- [ ] **AC20** Every ratio-type metric in a report carries a confidence interval (95% by default) computed with a configurable method; `std` and `stderr` across repeated runs are reported; and a task set below the judge-consistency threshold is marked "unreliable scoring". (R13)
- [ ] **AC21** Every decomposition cell reports sample size, success rate, continuous mean score, and confidence interval. (R14)
- [ ] **AC22** Each report explicitly selects and labels one aggregation method, the default is macro (task-equal-weighted) with the micro figures also given, and in a multi-dimensional Cartesian decomposition every cell has equal weight. (R15)
- [ ] **AC23** Cells with no coverage are shown as "no coverage" and do not participate in weighting; excluded trials are displayed with their number and reasons; NA pass@k items do not participate in that average and their valid task coverage is reported. (R16)
- [ ] **AC24** The judging model is independent of the model under test by default, or the report prominently flags the exception; the judging model identifier, version, judging temperature, and prompt template version are persisted alongside the results; temperature is 0, or the report notes that the judging model does not support it. (R17)
- [ ] **AC25** Scorer input text is wrapped in delimiters with the delimiters escaped inside the text; the judging conclusion is extracted from the fixed format (e.g. `GRADE: C/I/P`) and only the final conclusion is accepted; partial maps to 0.5. (R17)
- [ ] **AC26** A judging model error, timeout, or unparseable output is recorded as UNSCORED and never FAIL, with no retry beyond the permitted number of attempts; a judging model is calibrated on a human-annotated set before deployment and is not used for formal evaluation below the agreement threshold. (R17)
- [ ] **AC27** The review workbench displays the trajectory, scoring evidence, rubric items, and the current verdict and supports "uphold / revise / flag as disputed"; each revision records the reviser, the time, the reason, the original verdict, and the new verdict; verdict history can be appended but never deleted. (R18)
- [ ] **AC28** All UNSCORED trials, all boundary samples, and the first run of every high-value task are reviewed; random review runs at the configurable ratio; the "automatic scoring deviation rate" metric is produced; reviewed samples enter the human-annotated set used for the judge calibration in 4.4. (R18)
- [ ] **AC29** The worked example of 4.6 reproduces exactly: `SR` = 0.500, `pass@3` = 1.000, `pass^3` = 0.500, `avg@3` = 0.606, macro success rate = 0.444, platform failure rate = 11.1%; and a report whose platform failure rate exceeds the 5% threshold is flagged "insufficient confidence". (R19)

## Impact

| Area | Effect |
|---|---|
| Users | Task authors declare a reference plus combination rules instead of a scoring script, so scorers are reusable across tasks and across benchmarks. Agents are judged on the outcome by default, so equivalent correct solutions are no longer penalised, and platform-side failures are no longer deducted as capability failures. Report readers receive one canonical, reproducible set of numbers, with `pass@k` and `pass^k` always shown together and one aggregation method explicitly labelled. |
| Existing behaviour | Metric computation conventions move to a single definition set that reports, leaderboards, and export formats all reference. Aggregation becomes macro (task-equal-weighted) by default with the micro figures also given; cells with no coverage are shown as "no coverage" instead of being silently filled with 0; platform failures stay excluded from the denominators but their number and reasons are now displayed; the automatic scoring deviation rate metric is added. |
| Operational cost | Cost is measured in full as "model input tokens + output tokens + judging model consumption + user simulator consumption" and the scoring overhead must also be counted, so the price table is persisted alongside the results. Every judging model is calibrated on a human-annotated set before deployment, and human review load is mandatory for all UNSCORED trials, all boundary samples, and the first run of every high-value task, plus random review at a configurable ratio. |

## Risks and Mitigations

| Risk | Likelihood | Mitigation |
|---|---|---|
| The platform's own stability problems turn into capability deductions for the Agent when a judging model times out. | Med | UNSCORED is a first-class state: a judging failure is never equated with "judged as not passing", and infra errors are excluded from the denominator and counted in the judgement failure rate (see Inspect AI's `Score.unscored(reason="grader_failed")` and τ²-bench v2's practice). |
| State judgement may misjudge a task in which the Agent did something else useful. | Med | State judgement is the default and is accompanied by the action-assertion category `J3`; the risk is stated explicitly in the 4.1.3 comparison table, and action judgement is only enabled when the task explicitly declares it. |
| Action judgement may be too strict, judging equivalent paths wrong. | Med | The explicitly declared action sequence is labelled a "reference trajectory" with the semantics of one optional equivalent path, unless the task additionally declares it a "required action" (τ²-bench's redefinition of `actions`). |
| Reporting only one of `pass@k` and `pass^k` misleads the reader about capability versus reliability. | Med | Both must be reported together (τ²-bench reports pass^1..pass^k simultaneously). |
| The Agent farms score through regressive actions, or a partial score is credited on a task whose primary criterion failed. | Med | The progress score must be monotonically non-decreasing within a trial, and the final score stays capped by the primary criterion; reports present both a binary success rate and a continuous average score curve. |
| The Agent forges scorer delimiters in its output to fool the scorer. | Low | The text under test is wrapped in delimiters and the delimiters are escaped within the text; the judging output is parsed from a fixed format and only the final conclusion is accepted. |
| Model scoring itself is a source of noise, so judge verdicts are not reproducible. | Med | The judging model is independent of the model under test, its identifier/version/temperature/prompt template version are persisted, temperature is fixed at 0, judging failure is UNSCORED and never retried beyond the permitted attempts, and judge quality is calibrated on a human-annotated set before deployment. |
| Trial count, set by the evaluator's configuration, distorts the published score; cells with many tasks dominate a multi-dimensional decomposition. | Med | macro (task-equal-weighted) is used by default with the micro figures given for comparison, and every cell in a Cartesian decomposition has equal weight (τ²-bench's cell equal-weight aggregation, AgentCompass's `category_mean`). |
| Platform failures silently inflate the denominator or silently drop tasks, and a platform failure rate above 5% invalidates the report's credibility. | High | Platform failures are excluded from the denominator but their number and reasons are displayed, the platform failure rate is reported, uncovered cells are shown as "no coverage", NA `pass@k` items are reported with their valid task coverage, and a report whose platform failure rate exceeds the 5% threshold is flagged "insufficient confidence" (the 4.6 example yields 11.1%). |
| Human review is skipped or its history is altered, and the automatic scoring deviation stays unknown. | Med | Mandatory review of all UNSCORED trials, all boundary samples, and the first run of every high-value task; random review at a configurable ratio producing the "automatic scoring deviation rate"; verdict history can only be appended, never deleted; reviews feed the judge calibration in 4.4. |

## Rollout

Not specified in the source design.

## References

- Tech Design: Not yet created.
- Test Report: Not yet created.
- Source: Functional design chapter 04 — Scoring and Metrics.
- Sibling design chapter cited by this document: [FEAT-002 Capability Taxonomy and Task System][feat-002] (cited as 2.1.1, 2.1.2, 2.1.3, and 2.5 for the decomposition dimensions).

## Open Questions

| # | Question | Owner | Blocking? |
|---|---|---|---|
| 1 | Not specified in the source design. | - | - |

<!-- Link definitions stay at the bottom of the file. Angle brackets are required
     because the path contains no spaces but the reference form is kept consistent
     with the template. -->

[feat-002]: <./FEAT-002-capability-taxonomy-and-task-system.md>
