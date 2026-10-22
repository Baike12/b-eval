# FEAT-005: Reporting, Analysis and Comparison

| Field | Value |
|---|---|
| **ID** | FEAT-005 |
| **Status** | Draft |
| **Owner** | Platform team |
| **Priority** | P1 |
| **Implements** | `F-RPT-01`, `F-RPT-02`, `F-RPT-03`, `F-RPT-04`, `F-RPT-05`, `F-RPT-06`, `F-RPT-07`, `F-RPT-08`, `F-LDR-01`, `F-LDR-02`, `F-LDR-03`, `F-LDR-04`, `F-LDR-05` |
| **Created** | 2026-10-22 |
| **Last Updated** | 2026-10-22 |
| **Source** | Functional design chapter 05 — Reporting, Analysis and Comparison |

## Summary

A single evaluation is turned into four layered artifacts — trial record, run record, comparison record and leaderboard entry — each stamped with the plan fingerprint, and rendered as a confidence-first single-run report, a comparison view, a leaderboard, an exportable audit package and external publication pages. Failures are attributed automatically against the fixed five-level hierarchy ① Platform failure → ⑤ Task issue and can be overturned by human review, so a low score always resolves into what should be changed. Leaderboard publication re-computes the score from the submitted log, runs the anti-cheat and attribution verification steps and applies five confidence guardrails, so a published score is always traceable and reproducible by third parties without platform access.

## Problem

**Current behaviour:** A single evaluation produces four kinds of artifacts, ordered from the finest to the coarsest granularity. Without the plan fingerprint on every artifact it is impossible to tell whether two results are comparable, and a score that is not preceded by its confidence information is read as trustworthy on its face.

**Pain:** Readers take an untrustworthy score at face value when confidence information is not presented first, so a low score does not say "what should be changed" until it has been attributed to the platform, the integration layer, the budget, the Agent's capability or the task itself. Comparison is hurt twice over: without a baseline the difference between two runs cannot be read, and without a significance judgement noise is presented as an improvement. Developer interest concentrates on per-task differences, which are buried when only aggregate scores are exposed.

**Evidence:** Not specified in the source design.

## Proposed Solution

The platform turns one evaluation into a layered artifact chain — trial record, run record, comparison record, leaderboard entry — where every artifact carries the plan fingerprint, and renders it as a confidence-first single-run report, a comparison view, a leaderboard, an exportable audit package and external publication pages. Every failure is attributed automatically against the fixed, extensible five-level hierarchy ① Platform failure → ⑤ Task issue, and the initial attribution is appealable: human review can overturn it and the overturn is logged, so scorer and platform failures are never charged to the Agent. Leaderboard publication re-computes the score from the submitted log, verifies the anti-cheat and attribution conditions, applies the saturation, contamination, solvability, sample size and time guardrails, and displays the submitter's disclosure. Every export is designed so that third parties can recompute the scores without platform access.

## Scope

### In Scope

- Artifact inventory for a single evaluation: trial record, run record, comparison record and leaderboard entry, each carrying the plan fingerprint.
- Single-run report: confidence-first header, main score section, disaggregation section, failure analysis section and trajectory viewer.
- Comparison analysis: comparison targets (identical task set and version, explicitly enabled cross-task-set comparison), the six comparison layers with significance judgement, and baseline management including stale-baseline detection.
- Error attribution: the fixed five-level hierarchy ①–⑤, automatic classification, appeal with logged overturns, and the extensible failure mode taxonomy.
- Leaderboard and submission flow: leaderboard dimensions, the five-step submission flow with score re-computation, anti-cheat verification and attribution verification, and submission disclosure requirements.
- Leaderboard confidence guardrails: saturation, contamination, solvability, sample size and time.
- Human review and audit: verdict edit chain, task change history, plan fingerprint, submission verification records and the exportable audit package.
- External publication formats: static report page, comparison page, leaderboard page, dataset/log package and regression dashboard.

### Out of Scope

Not specified in the source design.

## Requirements

- **R1** Every evaluation produces the four artifact types defined in 5.1, ordered from the finest to the coarsest granularity, and every artifact carries the plan fingerprint.

  #### 5.1 Artifact Inventory

  A single evaluation produces four kinds of artifacts, ordered from the finest to the coarsest granularity:

  | Artifact | Granularity | Contents | Retention |
  |---|---|---|---|
  | **Trial record** | One execution of a single task | Full trajectory, verdict, usage, errors, environment evidence | Long-term |
  | **Run record** | One evaluation plan | Plan fingerprint, task set version, Agent version, aggregate metrics, disaggregated metrics, aggregated failure details | Long-term |
  | **Comparison record** | A contrast of two or more runs | Per-task differences, metric differences, significance markers | Long-term |
  | **Leaderboard entry** | One publication | Displayed score, Agent metadata, submission log, validation conclusion | Long-term |

  **All artifacts must carry the plan fingerprint**, otherwise it is impossible to tell whether two results are comparable.

- **R2** The single-run report presents all confidence information before any score, and marks the report as a whole "insufficient confidence" when the platform failure rate exceeds the threshold.

  #### 5.2 Single-Run Report

  ##### 5.2.1 Report Header (Confidence First)

  Scores must be presented after the confidence information; otherwise readers will take an untrustworthy score at face value.

  | Field | Description |
  |---|---|
  | Task set and version | |
  | Agent under test and version | |
  | Plan fingerprint | Used to compare against other runs |
  | Run time and elapsed time | |
  | **Platform failure rate** | When it exceeds the threshold, the report as a whole is marked "insufficient confidence" |
  | **Verdict failure rate** | Share of unscored items caused by scorer anomalies |
  | **Valid sample size** | Valid Trial count / planned Trial count |
  | **Scorer agreement** | Vote agreement rate when there are multiple judges; marked "not evaluated" for a single judge |
  | **Estimate vs. actual comparison** | Estimated cost/duration vs. actual |

- **R3** The main score section of the single-run report shows the total score under the selected aggregation method, pass@k and pass^k side by side, the continuous average score, a 95% confidence interval next to every ratio, the effective coverage rate and the primary criterion description.

  ##### 5.2.2 Main Score Section

  | Display item | Content |
  |---|---|
  | Total score | Given according to the selected aggregation method (default macro), with micro also given |
  | pass@k, pass^k | Displayed side by side, default k = number of repetitions |
  | Continuous average score | Score including the progress score of tasks |
  | Confidence interval | 95%, shown next to every ratio |
  | Effective coverage rate | Effective task coverage rate for pass@k-type metrics |
  | Primary criterion description | One sentence stating the definition of "success" |

- **R4** The disaggregation section produces one table/chart for each dimension of 4.3.1 with sample size, success rate, continuous average score and confidence interval in every cell, together with the five chart panels, and leaves dimensions with no coverage explicitly blank and annotated instead of filled with 0.

  ##### 5.2.3 Disaggregation Section

  For each dimension in [4.3.1][feat-004], produce one table/chart, each cell containing: sample size, success rate, continuous average score, confidence interval.

  **Chart panels:**
  - Success rate bar chart by domain
  - Capability dimension radar chart (**dimensions with no coverage are explicitly left blank and annotated, not filled with 0**)
  - Difficulty-stratified bar chart
  - Per-step score curve (cumulative average score at each step, showing "at which step it starts to improve")
  - Cost-vs-success-rate scatter plot

- **R5** The failure analysis section reports the termination reason distribution, the verdict failure details, the task failure list, the auto-clustered typical failure modes and the action quality metrics.

  ##### 5.2.4 Failure Analysis Section

  | Panel | Content |
  |---|---|
  | Termination reason distribution | Pie chart/bar chart + table |
  | Verdict failure details | PASS/FAIL/UNSCORED share by scorer, to locate unreliable criteria |
  | Task failure list | List of the tasks that did not pass, drill down to the Trial |
  | Typical failure modes | Top N high-frequency failure modes auto-clustered by the error attribution module (referencing the analyzer taxonomy of AgentCompass) |
  | Action quality | Invalid action rate, repeated action rate, malformed format rate, grounding accuracy |

- **R6** The trajectory viewer replays a Trial item by item, filters by capability tag, failure mode and termination reason, compares multiple trajectories side by side and exports a single trajectory.

  ##### 5.2.5 Trajectory Viewer

  - Per-Trial replay: actions, observations, elapsed time, and verdict evidence are shown item by item.
  - Supports filtering by capability tag, failure mode, and termination reason.
  - Supports side-by-side comparison of multiple trajectories.
  - Supports exporting a single trajectory.

- **R7** Any two run records can be compared when the task set and version are identical; comparing across different task sets requires explicitly enabling "cross-task-set comparison", in which case the report must state that scores are not directly comparable.

  #### 5.3 Comparison Analysis

  ##### 5.3.1 Comparison Targets

  Any two run records can be compared, provided that the **task set and version are identical**. Comparing across different task sets requires explicitly enabling "cross-task-set comparison", in which case the report must explicitly state "scores are not directly comparable; only the disaggregated structure can be compared".

- **R8** The comparison view shows the six comparison layers, marks a difference whose confidence interval spans 0 as "no significant difference", and supports both a "show only what improved" and a "show only what got worse" filter on the per-task difference list.

  ##### 5.3.2 Comparison Content

  | Layer | Content |
  |---|---|
  | Total score difference | Difference + confidence interval of the difference + whether it is statistically significant |
  | Disaggregation difference | Difference per domain and per capability dimension; **locate which cell the change comes from** |
  | Per-task difference | List of tasks that newly pass / newly fail / change in score |
  | Cost difference | Cost change, and whether "score-cost" improved simultaneously |
  | Behaviour difference | Change in average step count, tool call composition, and failure mode composition |
  | **Trajectory difference** | Per-task action additions/removals, parameter drift, and location of key divergence points |

  **Per-task differences are what developers care about most**; both "show only what improved" and "show only what got worse" filters must be supported.

  **Significance judgement:** when the confidence interval of the difference spans 0, mark it as "no significant difference"; presenting noise as an improvement is forbidden.

- **R9** Each task set maintains baseline runs that every new run is automatically compared against, a baseline can hold multiple entries, and a baseline whose task set version changed or whose platform failure rate exceeds the threshold is marked "stale" with a regeneration prompt.

  ##### 5.3.3 Baseline Management

  - Each task set can maintain **baseline runs** (e.g. "the run of the current primary model on test-public").
  - All new runs are automatically compared against the baseline and the difference is displayed.
  - A baseline can have multiple entries (e.g. "old and new versions", "open-source model vs. in-house model").
  - A baseline itself has failure conditions: task set version change, or the baseline platform failure rate exceeding the threshold → the baseline is automatically marked "stale" and a prompt to regenerate it is shown.

- **R10** Failures are attributed automatically against the fixed, extensible five-level hierarchy, the initial attribution is appealable and the overturn is logged, and failures of the scorer itself (the issues before ④) are never charged to the Agent.

  #### 5.4 Error Attribution

  Error attribution is the key step that turns "low score" into "what should be changed". The platform classifies according to a fixed taxonomy, and the classification items are extensible.

  **Attribution hierarchy (from outer to inner):**

  ```
  ① Platform failure   Platform/environment/scorer issue   → Not counted in the denominator; file a platform-side ticket
  ② Integration issue  Problem in the Agent adapter layer    → The integration self-check should catch it; report to the adapter layer
  ③ Budget issue       Not enough steps/time/cost            → Report "insufficient efficiency" or "budget set too tight"
  ④ Capability issue   Insufficient Agent capability         → Report capability dimension + failure mode
  ⑤ Task issue         The task itself is defective          → Report to the task maintainer
  ```

  **Key design: attribution must be automatic, and it must be appealable.** Failures of the scorer itself (the issues before ④) must never be charged to the Agent. The platform's failure mode classifier gives the initial attribution; human review can overturn it, and the overturn is logged.

  **Failure mode taxonomy examples (extensible):**
  - Action format mismatch (integration issue)
  - Hallucinated tool names / hallucinated arguments
  - Repeatedly retrying the same action without changing strategy
  - Goal achieved but not wrapped up (completion not declared)
  - Did the right thing but did not communicate it to the user
  - Failed to recognise an impossible task and forced an attempt anyway
  - Led astray by injected content
  - Early stop on the correct path

- **R11** Leaderboards are divided by task set + version + task set visibility, one entry for an Agent on the same leaderboard counts as one submission, and a leaderboard displays total score, pass@k, pass^k, continuous average score, confidence interval, platform failure rate and verdict failure rate.

  #### 5.5 Leaderboard and Submission Flow

  ##### 5.5.1 Leaderboard Dimensions

  - Leaderboards are divided by **task set + version + task set visibility**.
  - One entry for an Agent on the same leaderboard = one submission.
  - A leaderboard must display all of the following: total score, pass@k, pass^k, continuous average score, confidence interval, **platform failure rate**, **verdict failure rate**.

- **R12** A submission passes through the five-step flow (submit, recompute, anti-cheat verification, attribution verification, publish); the platform recomputes the score from the submitted log and rejects the submission when the recomputed value mismatches, when the log does not support the claimed result, when there are instances judged as failed that were actually claimed to be solved, or when the disclosure and exclusion rules are violated.

  ##### 5.5.2 Submission Flow

  | Step | Content | Validation |
  |---|---|---|
  | 1. Submit | The submitter submits the evaluation log and artifacts | Format and completeness validation |
  | 2. Recompute | The platform **recomputes the score** based on the submitted log | Mismatch with the submitted value → reject |
  | 3. Anti-cheat verification | Check whether the log supports the claimed result | Not supported → reject and record the reason |
  | 4. Attribution verification | Check whether there are instances judged as failed that were actually claimed to be solved | Exists → reject |
  | 5. Publish | Enters the leaderboard | — |

  **Specific checks in anti-cheat verification (referencing the submission validation approach of SWE-bench):**
  - A task claimed to pass must have supporting evidence in the submitted log; **pass claims without a submitted log are always rejected**.
  - For an instance claimed to pass, the verdict evidence must be self-consistent (e.g. evidence that "the tests really executed" exists).
  - The plan fingerprint must match the leaderboard requirements (task set version, scoring configuration, number of repetitions).
  - Exclusions caused by platform failures must be genuine exclusions; they must not be used to remove unfavourable samples.

  **Submission disclosure requirements:** the submitter must disclose the complete form of the Agent under test (model, prompt, tools, version), and the leaderboard entry displays that disclosure. Submissions that conceal the tool set or use unpublished tools are revoked once discovered during review.

- **R13** Leaderboard entries are subject to the saturation, contamination, solvability, sample size and time guardrails.

  ##### 5.5.3 Leaderboard Confidence Guardrails

  | Guardrail | Rule |
  |---|---|
  | **Saturation reminder** | A task passed by ≥ 90% of submissions → marked as "saturated", no longer used as a discrimination metric, and flagged in the report |
  | **Contamination reminder** | A task's instruction text appears in public corpora → marked "suspected contamination" |
  | **Solvability reminder** | A task passed by ≥ 95% of submissions and manually annotated as difficulty L4 → triggers a difficulty review |
  | **Sample size guardrail** | If the valid sample size for a given score of the same Agent is below the threshold, mark it as "insufficient samples" |
  | **Time guardrail** | A leaderboard entry carries an evaluation date; once it exceeds the validity period (default 90 days), mark it "possibly outdated" |

- **R14** Human review and audit record the verdict edit chain, the task change history, the plan fingerprint and the submission verification records, and a single evaluation can be exported as a self-contained audit package that lets third parties recompute the scores without platform access.

  #### 5.6 Human Review and Audit

  **Audit capabilities (for platform operators and third parties):**
  - Verdict edit chain: every manual overturn records author, time, and reason, and can be fully replayed.
  - Task change history: every revision of task content is logged, so one can check "under which task version a given score was obtained".
  - Plan fingerprint: prevents "same score, different tasks" from being conflated.
  - Submission verification records: rejected submissions and their reasons are retained permanently.

  **Exportable audit package:**
  A single evaluation can be exported as a self-contained audit package, containing: task set version and task content, verdict configuration, scorer implementation version, per-Trial trajectories and verdict evidence, the metric computation process, and the raw judge output (if the scoring was model-based). **The purpose is to let third parties recompute the scores without platform access.**

- **R15** Results can be published in the five external formats, each with its own key requirements.

  #### 5.7 External Publication Formats

  | Format | Use | Key requirements |
  |---|---|---|
  | **Static report page** | Sharing the result of a single evaluation | Confidence information first; private task sets must have the task content masked |
  | **Comparison page** | Contrast of two versions / two Agents | Significance markers; locate the source of changes |
  | **Leaderboard page** | Public ranking | Complete confidence columns; guardrail markers |
  | **Dataset/log package** | For external reproduction | Self-contained; includes plan fingerprint and verdict evidence |
  | **Regression dashboard** | Agent regression in CI | Time series of only the key metrics; a failure blocks the pipeline |

## Acceptance Criteria

Written as verifiable outcomes.

- [ ] **AC1** (R1) All four artifact types — trial record, run record, comparison record and leaderboard entry — expose the plan fingerprint, and an artifact without it is visibly not comparable.
- [ ] **AC2** (R2) The report header shows the task set and version, the Agent under test and version, the plan fingerprint, run time and elapsed time, the platform failure rate, the verdict failure rate, the valid sample size, the scorer agreement and the estimate vs. actual comparison, and it is rendered above the score section.
- [ ] **AC3** (R2) When the platform failure rate exceeds the threshold, the report as a whole is marked "insufficient confidence".
- [ ] **AC4** (R2) Scorer agreement is marked "not evaluated" for a single judge and shows the vote agreement rate when there are multiple judges.
- [ ] **AC5** (R3) The main score section gives the total score according to the selected aggregation method (default macro) with micro also given, shows pass@k and pass^k side by side with default k = number of repetitions, shows the continuous average score including the progress score of tasks, and states the primary criterion description in one sentence.
- [ ] **AC6** (R3) Every ratio in the main score section is shown next to a 95% confidence interval, and the effective coverage rate for pass@k-type metrics is displayed.
- [ ] **AC7** (R4) One table/chart is produced for each dimension in 4.3.1 and every cell contains sample size, success rate, continuous average score and confidence interval.
- [ ] **AC8** (R4) The disaggregation view contains the success rate bar chart by domain, the capability dimension radar chart, the difficulty-stratified bar chart, the per-step score curve and the cost-vs-success-rate scatter plot.
- [ ] **AC9** (R4) Capability dimensions with no coverage are left blank and annotated, and are never filled with 0.
- [ ] **AC10** (R5) The failure analysis section shows the termination reason distribution as a pie chart/bar chart plus table, the PASS/FAIL/UNSCORED share by scorer, the task failure list drillable down to the Trial, the top N auto-clustered typical failure modes and the action quality metrics (invalid action rate, repeated action rate, malformed format rate, grounding accuracy).
- [ ] **AC11** (R6) The trajectory viewer shows actions, observations, elapsed time and verdict evidence item by item, filters by capability tag, failure mode and termination reason, supports side-by-side comparison of multiple trajectories and exports a single trajectory.
- [ ] **AC12** (R7) Two run records with an identical task set and version can be compared; comparing run records from different task sets is only possible after "cross-task-set comparison" is explicitly enabled, and the report then explicitly states "scores are not directly comparable; only the disaggregated structure can be compared".
- [ ] **AC13** (R8) The comparison view shows the total score difference with its confidence interval and a significance marker, the disaggregation difference per domain and per capability dimension locating the originating cell, the per-task difference list, the cost difference, the behaviour difference and the trajectory difference.
- [ ] **AC14** (R8) When the confidence interval of the difference spans 0 the difference is marked "no significant difference", and no difference is presented as an improvement on that basis.
- [ ] **AC15** (R8) The per-task difference list supports both a "show only what improved" filter and a "show only what got worse" filter.
- [ ] **AC16** (R9) Every new run is automatically compared against the baseline runs of its task set and the difference is displayed, and a baseline can hold multiple entries.
- [ ] **AC17** (R9) When the task set version changes or the baseline platform failure rate exceeds the threshold, the baseline is automatically marked "stale" and a prompt to regenerate it is shown.
- [ ] **AC18** (R10) Every failure is classified into one of ① Platform failure, ② Integration issue, ③ Budget issue, ④ Capability issue or ⑤ Task issue, and the platform's failure mode classifier supplies the initial attribution automatically.
- [ ] **AC19** (R10) Failures classified as ① or ② are never charged to the Agent; a ① Platform failure is not counted in the denominator and a platform-side ticket is filed.
- [ ] **AC20** (R10) Human review can overturn the automatic attribution and the overturn is logged.
- [ ] **AC21** (R11) Leaderboards are separated by task set + version + task set visibility, and one entry for an Agent on the same leaderboard counts as one submission.
- [ ] **AC22** (R11) A leaderboard displays total score, pass@k, pass^k, continuous average score, confidence interval, platform failure rate and verdict failure rate.
- [ ] **AC23** (R12) The platform recomputes the score from the submitted log, and a mismatch with the submitted value causes the submission to be rejected.
- [ ] **AC24** (R12) A pass claim without a submitted log is always rejected, and a claimed-pass instance whose verdict evidence is not self-consistent is rejected.
- [ ] **AC25** (R12) A submission whose plan fingerprint does not match the leaderboard requirements (task set version, scoring configuration, number of repetitions) is rejected.
- [ ] **AC26** (R12) Exclusions caused by platform failures are genuine exclusions and are never used to remove unfavourable samples.
- [ ] **AC27** (R12) The attribution verification step rejects the submission when instances judged as failed were actually claimed to be solved, and the rejection reason is recorded.
- [ ] **AC28** (R12) The submitter discloses the complete form of the Agent under test (model, prompt, tools, version) and the leaderboard entry displays that disclosure; a submission that conceals the tool set or uses unpublished tools is revoked once discovered during review.
- [ ] **AC29** (R13) A task passed by ≥ 90% of submissions is marked "saturated", is no longer used as a discrimination metric, and is flagged in the report.
- [ ] **AC30** (R13) A task whose instruction text appears in public corpora is marked "suspected contamination".
- [ ] **AC31** (R13) A task passed by ≥ 95% of submissions and manually annotated as difficulty L4 triggers a difficulty review.
- [ ] **AC32** (R13) A score whose valid sample size for the same Agent is below the threshold is marked "insufficient samples".
- [ ] **AC33** (R13) A leaderboard entry carries an evaluation date and is marked "possibly outdated" once it exceeds the validity period (default 90 days).
- [ ] **AC34** (R14) Every manual overturn records author, time and reason and can be fully replayed; every revision of task content is logged; rejected submissions and their reasons are retained permanently.
- [ ] **AC35** (R14) The exported audit package is self-contained and contains the task set version and task content, the verdict configuration, the scorer implementation version, the per-Trial trajectories and verdict evidence, the metric computation process and the raw judge output (if the scoring was model-based), so that a third party can recompute the scores without platform access.
- [ ] **AC36** (R15) The static report page shows confidence information first and masks the task content of private task sets; the comparison page carries significance markers and locates the source of changes; the leaderboard page shows complete confidence columns and guardrail markers; the dataset/log package is self-contained and includes the plan fingerprint and verdict evidence; the regression dashboard shows a time series of only the key metrics and a failure blocks the pipeline.

## Impact

| Area | Effect |
|---|---|
| Users | Not specified in the source design. |
| Existing behaviour | Not specified in the source design. |
| Operational cost | Not specified in the source design. |

## Risks and Mitigations

| Risk | Likelihood | Mitigation |
|---|---|---|
| Not specified in the source design. | - | - |

## Rollout

Not specified in the source design.

## References

- Tech Design: Not yet created.
- Test Report: Not yet created.
- Sibling design chapter referenced by this feature: [FEAT-004 Scoring and Metrics][feat-004] (source §4.3.1 disaggregation dimensions, referenced from §5.2.3).

## Open Questions

| # | Question | Owner | Blocking? |
|---|---|---|---|
| 1 | Not specified in the source design. | - | - |

<!-- Link definitions stay at the bottom of the file. -->

[feat-004]: <./FEAT-004-scoring-and-metrics.md>
