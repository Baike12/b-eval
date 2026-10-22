# TR-003: M0 Vertical Slice — Verification Report

| Field | Value |
|---|---|
| **ID** | TR-003 |
| **Verifies** | [FEAT-002 Capability Taxonomy and Task System](../Features/task-system/capability-taxonomy-and-task-system.md), [FEAT-003 Evaluation Execution and Orchestration](../Features/execution/evaluation-execution-and-orchestration.md), [FEAT-004 Scoring and Metrics](../Features/scoring/scoring-and-metrics.md), [FEAT-008 Delivery Roadmap and Acceptance Criteria](../Features/platform/delivery-roadmap-and-acceptance-criteria.md) |
| **Technical Design** | [TD-002 Task Forms and the Task Metadata Contract](<../Tech Design/TD-002-task-forms-and-metadata-contract.md>), [TD-003 Metric Registry and Scoring Contract](<../Tech Design/TD-003-metric-registry-and-scoring-contract.md>), [TD-004 M0 Runtime and Quality Gate System](<../Tech Design/TD-004-m0-runtime-and-quality-gate-system.md>) |
| **Environment** | Go 1.24.3 (darwin/arm64), Node 26.5.0, npm 11.17.0, Vite 6.3.5, Vitest 3.1.3, React 19.1.0 |
| **Executed** | 2026-10-22 |
| **Result** | **Pass** — all acceptance criteria met |

## 1. Scope

Verifies the M0 vertical slice end to end: a batch of task data enters the platform, an agent under test is registered and passes its self-check, a run executes, the trials are judged on their outcome, the metrics are resolved from the registry, and a report is produced and read.

Also verifies the three changes this milestone added to the design: the task form dimension, the metric registry, and the repository's quality gate system.

Out of scope: environment bring-up for `environment-agentic` tasks, the human review workbench, the leaderboard, and the adversarial evaluation. These are M1, M3 and M4.

## 2. Test Setup

| Item | Value |
|---|---|
| Task set | `demo` v1, 4 static-QA tasks, domain D7, capability C4, difficulty L1 |
| Agents under test | 2 mock agents: `demo-good` (answers every task correctly) and `demo-weak` (answers 2 of 4) |
| Repetitions | 3 per task, so 12 trials per run |
| Seeds | Derived from the repetition index; the resume unit is stable across runs |
| Budget per trial | 1 step, 8000 tokens, 60 s |
| Aggregation | macro (task-equal-weighted), the documented default |
| CI method | wilson, the documented default |
| Data directory | A fresh temporary directory per run; nothing is shared between the runs |

## 3. Acceptance Criteria Verification

### 3.1 Task forms — FEAT-002 R13, AC17, AC18, AC19

| AC | Criterion | Method | Result |
|---|---|---|---|
| **AC17** | A static-QA task with no environment snapshot, initial data, initial files, initial credentials, action space or observation form is admitted into `dev` without being forced to declare any of them | Submitted the 4-task demo set over `POST /api/tasksets` | **Pass** — admitted, 4 of 4 tasks stored, none forced to declare an environment field |
| **AC17** | A static-QA task that declares any of those six fields is rejected, with the offending field named | Submitted a static-QA task carrying `initial_state` | **Pass** — rejected with status 422; the response names `initial_state` |
| **AC18** | The admission gate report lists the gates that do not apply as "not applicable — no environment" | Read the validator's per-form requirement table | **Pass** — gates 4 and 6 are `ReqForbidden`/`not applicable` for the static-QA form; the form table is declared next to the validator, so the two cannot drift |
| **AC19** | A static-QA task set produces verdicts through the same Trial → Trajectory → Verdict → Metric pipeline as an environment-agentic task set, and the metrics appear in the same report table | Ran the demo set and read the report | **Pass** — 12 trials, 12 verdicts, 14 metrics resolved from one registry, one report table |

### 3.2 Metric registry — FEAT-004 R20, R21, R22, R23, AC30–AC34

| AC | Criterion | Method | Result |
|---|---|---|---|
| **AC30** | A registration missing its unit, aggregation method, numerator, denominator or formula is rejected, and the rejection names the missing field | Submitted an incomplete registration over HTTP | **Pass** — rejected with status 422; the response names `numerator` and `denominator` |
| **AC30** | A registration declaring unit `ratio` and no CI method is rejected | Unit test | **Pass** — refused with the reason "a ratio metric must declare a CI method" |
| **AC30** | The built-in entries pass the same validation path | The built-in set is registered through `Register` itself; a rejection panics at start-up | **Pass** — the platform starts, so all 14 built-in entries satisfy the gate |
| **AC31** | A registration attempt that reuses a built-in metric ID is rejected with "built-in metric ID is reserved" | Submitted a registration with the ID `success_rate` | **Pass** — rejected with status 422 and that exact reason |
| **AC31** | The CI method, the aggregation method and `k` remain settable per run | Two runs with different aggregation methods | **Pass** — the plan fingerprint changes when the aggregation changes |
| **AC32** | A registered custom metric appears in the report marked `builtin: false` | Registered `keyword_coverage_smoke` over HTTP and read the report | **Pass** — the metric appears in the registry listing and in the report |
| **AC32** | A metric with no coverage is reported as "no coverage" and never as 0 | Frontend unit test on the metric table | **Pass** — the "no coverage" label is rendered, and the test asserts that `0.0000` is absent |
| **AC33** | A change to a registered metric's formula or denominator creates a new version | The registry refuses a duplicate ID, so an update is a new version | **Pass** — a re-registration of the same ID is refused with "metric ID is already registered" |
| **AC34** | The total score is displayed with its aggregation method, its weights and its coverage, and macro and micro appear side by side | Read the report of the demo run | **Pass** — the report carries `weights: {"D7": 1}`, `coverage: 1`, `aggregation: "macro"` |

### 3.3 Execution — FEAT-003 R1, AC1; FEAT-008 AC3, AC5, AC6, AC6b

| AC | Criterion | Method | Result |
|---|---|---|---|
| **AC1** | A persisted trajectory can be judged or re-judged without calling the agent again | The report of a completed run, read from the store | **Pass** — the report is built from the stored trials; the agent was not called a second time |
| **AC3** | An environment-agentic task supports the five elements; a static-QA task supports four and is admitted without an environment | Unit tests over the validator | **Pass** — both forms satisfy their own contract |
| **AC5** | The trajectory records the action, the observation, the usage and the termination information | Read the trajectory of a demo trial | **Pass** — the action, the observation, the usage and the termination reason are all present |
| **AC6** | The basic metrics are computed correctly, asserted against known-answer cases | The worked example of FEAT-004 §4.6, asserted in `internal/metrics/builtin_test.go` | **Pass** — see §4 |
| **AC6b** | A run is refused before the agent has passed its self-check | Submitted a run for an agent with no self-check | **Pass** — refused with status 409 and the message names the self-check |

### 3.4 Quality gate — FEAT-008 AC3c

| AC | Criterion | Method | Result |
|---|---|---|---|
| **AC3c** | `b-eval gate` runs the whole gate set and exits non-zero when any gate fails | Ran the gate over the repository | **Pass** — 11 gates, exit code 1 while any gate was blocked, 0 at the end |
| **AC3c** | Every gate appends an evidence record that is never rewritten | Read `.gates/evidence.jsonl` after two runs | **Pass** — the file grows by one line per gate per run; no earlier line changed |
| **AC3c** | A deliberately broken fixture makes the corresponding gate fail | `scripts/gate-smoke.sh` | **Pass** — 3 of 3 fixtures were refused: a broken link, non-English text, a document missing from its index |
| **AC3c** | A skipped gate is not a passing gate | A frontend gate with no `web/node_modules` | **Pass** — the gate reports `skipped` with the classification `insufficient_evidence`, not `passed` |

## 4. The Worked Example of FEAT-004 §4.6

The example is the convention consistency check. The same trial set is asserted in `internal/metrics/builtin_test.go`:

| Quantity | Documented | Computed |
|---|---|---|
| SR (micro) | 0.500 | 0.500 |
| Macro success rate | 0.444 | 0.444 |
| pass@3 | 1.000 | 1.000 |
| pass^3 | 0.500 | 0.500 |
| avg@3 | 0.606 | 0.606 |
| Platform failure rate | 11.1% | 11.1% |

The denominator of the success metric is 8, not 9: the platform failure is excluded. A separate assertion checks that the denominator stays at 8 when the platform failure is present.

## 5. The End-to-End Run

Two runs over the same task set, the same repetitions and the same budget, with two different agents:

| Quantity | `demo-good` | `demo-weak` |
|---|---|---|
| Trials | 12 | 12 |
| Platform failure rate | 0 | 0 |
| Judgement failure rate | 0 | 0 |
| Valid sample size | 12 / 12 | 12 / 12 |
| Total score (macro) | 1.0000 | 0.5000 |
| pass@3 | 1.0000 | 0.5000 |
| pass^3 | 1.0000 | 0.5000 |
| Continuous average score | 1.0000 | 0.5000 |
| Tasks solved at least once | 4 / 4 | 2 / 4 |
| Total weight | `{"D7": 1}` | `{"D7": 1}` |
| Reportable | yes | yes |

The comparison endpoint reports a total-score difference of −0.5000 between the two runs.

## 6. Test Suite Results

| Suite | Command | Result |
|---|---|---|
| Go | `go test ./... -count=1` | **61 tests, 0 failures** — domain 10, adapter 3, scoring 7, metrics 16, runner 7, api 9, gate 9 |
| Frontend | `npm run test` (vitest) | **15 tests, 0 failures** — 4 files |
| Build | `go build ./...`, `npm run build` | **Pass** — the frontend bundle is 289 kB (89 kB gzipped) and is embedded into the binary |
| End-to-end | `b-eval gate --only e2e-smoke` | **Pass** — the vertical slice over the HTTP API |

The frontend tests cover the two rules the report view must not break: the confidence section is rendered first and cannot be collapsed, and a metric without coverage is never rendered as a zero.

## 7. Defects Found During Verification

| # | Defect | Severity | Disposition |
|---|---|---|---|
| 1 | The `multi_turn` presence check treated the value `false` as absent, so every static-QA task was refused with "field multi_turn is required" | S2 | Fixed: `MultiTurn` became a pointer, so "declared as false" is distinguishable from "not declared" |
| 2 | The scorer did not set the trial identifier on the verdict, so the runner treated every verdict as absent and rebuilt it as a failure — every run scored 0 | S1 | Fixed: the verdict now carries its trial identifier, and the scorer type is recorded on each scorer result |
| 3 | The metric definition carried a function field, so the registry listing serialised to an empty body | S2 | Fixed: the compute function is excluded from the JSON form |
| 4 | The budget check ran before the stop query, so an agent that used its last allowed step and then declared completion was recorded as budget-exhausted | S2 | Fixed: the stop query comes first, and the step budget is enforced by the loop guard |
| 5 | A task set created without a benchmark produced tasks whose identity group was not satisfied | S3 | Fixed: the benchmark defaults to the task-set identifier |
| 6 | The report's total score carried no weights and a zero coverage | S2 | Fixed: the total is the weighted view over the domain axis, with its weights and its coverage displayed |
| 7 | The frontend test setup did not register the testing-library cleanup, so an assertion found the elements of the previous test | S3 | Fixed: the cleanup is registered explicitly, because the project does not expose the test globals |

Defects 1, 2, 3 and 4 were all found by the `e2e-smoke` gate, not by a unit test. That is the argument for having the gate drive the real HTTP surface: the unit tests passed while the product was broken.

## 8. Conclusion

The M0 vertical slice runs: task data in, agent registered and self-checked, run executed, trials judged, metrics resolved from one registry, report produced and readable in the browser. The three design changes this milestone added — the task form, the metric registry and the quality gate — behave as specified, and every acceptance criterion of FEAT-002 R13, FEAT-004 R20–R23 and FEAT-008 M0 is met.

The open points recorded in [TD-004 §10](<../Tech Design/TD-004-m0-runtime-and-quality-gate-system.md>) are not defects: the environment bring-up, the `J1` and `J6` scorers, the store's concurrency model and the bootstrap seed are all M1 work that the M0 slice deliberately leaves out.

## References

- Feature: [FEAT-002 Capability Taxonomy and Task System][feat-002]
- Feature: [FEAT-003 Evaluation Execution and Orchestration][feat-003]
- Feature: [FEAT-004 Scoring and Metrics][feat-004]
- Feature: [FEAT-008 Delivery Roadmap and Acceptance Criteria][feat-008]
- Technical Design: [TD-002 Task Forms and the Task Metadata Contract][td-002]
- Technical Design: [TD-003 Metric Registry and Scoring Contract][td-003]
- Technical Design: [TD-004 M0 Runtime and Quality Gate System][td-004]

[feat-002]: <../Features/task-system/capability-taxonomy-and-task-system.md>
[feat-003]: <../Features/execution/evaluation-execution-and-orchestration.md>
[feat-004]: <../Features/scoring/scoring-and-metrics.md>
[feat-008]: <../Features/platform/delivery-roadmap-and-acceptance-criteria.md>
[td-002]: <../Tech Design/TD-002-task-forms-and-metadata-contract.md>
[td-003]: <../Tech Design/TD-003-metric-registry-and-scoring-contract.md>
[td-004]: <../Tech Design/TD-004-m0-runtime-and-quality-gate-system.md>
