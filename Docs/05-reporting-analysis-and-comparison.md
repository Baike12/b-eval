# 05 · Reporting Analysis and Comparison

## 5.1 Artifact Inventory

A single evaluation produces four kinds of artifacts, ordered from the finest to the coarsest granularity:

| Artifact | Granularity | Contents | Retention |
|---|---|---|---|
| **Trial record** | One execution of a single task | Full trajectory, verdict, usage, errors, environment evidence | Long-term |
| **Run record** | One evaluation plan | Plan fingerprint, task set version, Agent version, aggregate metrics, disaggregated metrics, aggregated failure details | Long-term |
| **Comparison record** | A contrast of two or more runs | Per-task differences, metric differences, significance markers | Long-term |
| **Leaderboard entry** | One publication | Displayed score, Agent metadata, submission log, validation conclusion | Long-term |

**All artifacts must carry the plan fingerprint**, otherwise it is impossible to tell whether two results are comparable.

## 5.2 Single-Run Report

### 5.2.1 Report Header (Confidence First)

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

### 5.2.2 Main Score Section

| Display item | Content |
|---|---|
| Total score | Given according to the selected aggregation method (default macro), with micro also given |
| pass@k, pass^k | Displayed side by side, default k = number of repetitions |
| Continuous average score | Score including the progress score of tasks |
| Confidence interval | 95%, shown next to every ratio |
| Effective coverage rate | Effective task coverage rate for pass@k-type metrics |
| Primary criterion description | One sentence stating the definition of "success" |

### 5.2.3 Disaggregation Section

For each dimension in 4.3.1, produce one table/chart, each cell containing: sample size, success rate, continuous average score, confidence interval.

**Chart panels:**
- Success rate bar chart by domain
- Capability dimension radar chart (**dimensions with no coverage are explicitly left blank and annotated, not filled with 0**)
- Difficulty-stratified bar chart
- Per-step score curve (cumulative average score at each step, showing "at which step it starts to improve")
- Cost-vs-success-rate scatter plot

### 5.2.4 Failure Analysis Section

| Panel | Content |
|---|---|
| Termination reason distribution | Pie chart/bar chart + table |
| Verdict failure details | PASS/FAIL/UNSCORED share by scorer, to locate unreliable criteria |
| Task failure list | List of the tasks that did not pass, drill down to the Trial |
| Typical failure modes | Top N high-frequency failure modes auto-clustered by the error attribution module (referencing the analyzer taxonomy of AgentCompass) |
| Action quality | Invalid action rate, repeated action rate, malformed format rate, grounding accuracy |

### 5.2.5 Trajectory Viewer

- Per-Trial replay: actions, observations, elapsed time, and verdict evidence are shown item by item.
- Supports filtering by capability tag, failure mode, and termination reason.
- Supports side-by-side comparison of multiple trajectories.
- Supports exporting a single trajectory.

## 5.3 Comparison Analysis

### 5.3.1 Comparison Targets

Any two run records can be compared, provided that the **task set and version are identical**. Comparing across different task sets requires explicitly enabling "cross-task-set comparison", in which case the report must explicitly state "scores are not directly comparable; only the disaggregated structure can be compared".

### 5.3.2 Comparison Content

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

### 5.3.3 Baseline Management

- Each task set can maintain **baseline runs** (e.g. "the run of the current primary model on test-public").
- All new runs are automatically compared against the baseline and the difference is displayed.
- A baseline can have multiple entries (e.g. "old and new versions", "open-source model vs. in-house model").
- A baseline itself has failure conditions: task set version change, or the baseline platform failure rate exceeding the threshold → the baseline is automatically marked "stale" and a prompt to regenerate it is shown.

## 5.4 Error Attribution

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

## 5.5 Leaderboard and Submission Flow

### 5.5.1 Leaderboard Dimensions

- Leaderboards are divided by **task set + version + task set visibility**.
- One entry for an Agent on the same leaderboard = one submission.
- A leaderboard must display all of the following: total score, pass@k, pass^k, continuous average score, confidence interval, **platform failure rate**, **verdict failure rate**.

### 5.5.2 Submission Flow

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

### 5.5.3 Leaderboard Confidence Guardrails

| Guardrail | Rule |
|---|---|
| **Saturation reminder** | A task passed by ≥ 90% of submissions → marked as "saturated", no longer used as a discrimination metric, and flagged in the report |
| **Contamination reminder** | A task's instruction text appears in public corpora → marked "suspected contamination" |
| **Solvability reminder** | A task passed by ≥ 95% of submissions and manually annotated as difficulty L4 → triggers a difficulty review |
| **Sample size guardrail** | If the valid sample size for a given score of the same Agent is below the threshold, mark it as "insufficient samples" |
| **Time guardrail** | A leaderboard entry carries an evaluation date; once it exceeds the validity period (default 90 days), mark it "possibly outdated" |

## 5.6 Human Review and Audit

**Audit capabilities (for platform operators and third parties):**
- Verdict edit chain: every manual overturn records author, time, and reason, and can be fully replayed.
- Task change history: every revision of task content is logged, so one can check "under which task version a given score was obtained".
- Plan fingerprint: prevents "same score, different tasks" from being conflated.
- Submission verification records: rejected submissions and their reasons are retained permanently.

**Exportable audit package:**
A single evaluation can be exported as a self-contained audit package, containing: task set version and task content, verdict configuration, scorer implementation version, per-Trial trajectories and verdict evidence, the metric computation process, and the raw judge output (if the scoring was model-based). **The purpose is to let third parties recompute the scores without platform access.**

## 5.7 External Publication Formats

| Format | Use | Key requirements |
|---|---|---|
| **Static report page** | Sharing the result of a single evaluation | Confidence information first; private task sets must have the task content masked |
| **Comparison page** | Contrast of two versions / two Agents | Significance markers; locate the source of changes |
| **Leaderboard page** | Public ranking | Complete confidence columns; guardrail markers |
| **Dataset/log package** | For external reproduction | Self-contained; includes plan fingerprint and verdict evidence |
| **Regression dashboard** | Agent regression in CI | Time series of only the key metrics; a failure blocks the pipeline |
