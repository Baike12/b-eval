# 07 · User Flows and Feature Inventory

## 7.1 Core User Journeys

### Journey A · Agent Developer: Onboarding and Self-Test

```
① Register the system under test (name, version, access method)  →  ② Configure the access adapter
        ↓
③ Run the access self-check (sample tasks)  →  ④ Review the self-check results: action format, contract violations, usage reporting
        ↓                                      ↓ fail
⑤ Run the smoke set (≤ 10 items per domain)  →  ⑥ Confirm the baseline score is stored  →  ⑦ Onboarding complete
```

**Key point:** When the self-check does not pass, the evaluation features are disabled for that system under test — this prevents an integration bug from being recorded as weak capability.

### Journey B · Agent Developer: Routine Regression

```
① Start an evaluation (select task set, select baseline, select repeat count)
        ↓
② Review the cost estimate, confirm the budget  →  ③ Run (can be paused/cancelled at any time)
        ↓
④ Open the report: start with the confidence section (platform failure rate, judgement failure rate)
        ↓
⑤ Locate the degraded tasks →  Drill down to the trajectory →  Review failure-mode attribution
        ↓
⑥ Export the comparison report / submit for internal review
```

### Journey C · Task Author: Adding a New Task

```
① Create a draft task (fill in identity, capability tags, difficulty, domain)
        ↓
② Write the initial state + interaction contract (action space, observations, budget, termination conditions)
        ↓
③ Write the judgement conditions (select scorer, combination method, reference solution)
        ↓
④ Run the validation gate (eight checks)  →  ⑤ Fix each failed item
        ↓
⑥ Place it in the dev set →  ⑦ Run it once with the baseline Agent, check whether the difficulty matches the label
        ↓
⑧ Submit for review →  ⑨ Once approved, it enters the test set and the version is locked
```

### Journey D · Evaluation Organizer: Batch Evaluation and Publication

```
① Create an evaluation plan (multiple Agents × multiple task sets × repeat count)
        ↓
② Estimate the total cost, confirm the resource quota
        ↓
③ Run (can be interrupted and resumed, watch progress and failure rate in real time)
        ↓
④ Result health check: platform failure rate, judgement failure rate, saturation/contamination warnings
        ↓
⑤ Generate a comparison report (horizontal across Agents, vertical across versions)
        ↓
⑥ Submit to the leaderboard →  Platform recompute + anti-cheat verification  →  ⑦ Publish or return
```

### Journey E · Security Reviewer: Adversarial Evaluation

```
① Select a user task set + an attack task set
        ↓
② Configure the attack strategy × injection vector matrix
        ↓
③ Select defense strategies (multiple allowed, forming control groups)
        ↓
④ Run →  ⑤ Review the dual-axis report: utility / targeted ASR / untargeted ASR
        ↓
⑥ Compare the attack-defense trade-off curves of different defenses →  ⑦ Export the conclusions
```

## 7.2 Feature Module Inventory

| Module | ID | Feature | Description |
|---|---|---|---|
| **System Under Test Management** | F-AGT-01 | Register / archive the system under test | Record name, version, form, disclosure information |
| | F-AGT-02 | Access adapter configuration | Configure the adaptation method, action mapping, termination judgement |
| | F-AGT-03 | Access self-check | Sample tasks + difference report + adapter protocol violation statistics |
| | F-AGT-04 | Access readiness gate | Evaluation is disabled if the self-check does not pass |
| **Tasks and Task Sets** | F-TSK-01 | Task CRUD | Three states: draft / published / deprecated |
| | F-TSK-02 | Capability and domain tagging | Tag selection + mutual-exclusion check |
| | F-TSK-03 | Difficulty tagging | Four levels + rationale record |
| | F-TSK-04 | Initial state configuration | Snapshot, data, credentials, seed |
| | F-TSK-05 | Interaction contract configuration | Action space, observation form, termination conditions, budget |
| | F-TSK-06 | Judgement condition configuration | Scorer reference + combination + weights + primary criterion |
| | F-TSK-07 | Task templates and parameterization | Template definition + parameter domain + stratified sampling + freezing |
| | F-TSK-08 | Task set version management | Publish, freeze, roll back, diff |
| | F-TSK-09 | Task split management | smoke / dev / test-public / test-private / stress |
| | F-TSK-10 | External benchmark import | Mapping + baseline re-run + tied scores |
| | F-TSK-11 | Ingestion validation gate | Eight checks + report + blocking |
| | F-TSK-12 | Task health audit | Saturation, abnormal failures, non-determinism, duration, difficulty drift |
| **Scorers** | F-JDG-01 | Scorer registration | Registration and versioning of the six scorer types |
| | F-JDG-02 | Scorer combination configuration | Conjunction / disjunction / weights / partial credit |
| | F-JDG-03 | Rubric editor | Item-based editing + weights + evidence requirements |
| | F-JDG-04 | Scorer regression | Run the regression set after a change, block versions that do not pass |
| | F-JDG-05 | Scoring invisibility configuration | Configure how the judgement conditions are isolated from the implementation |
| | F-JDG-06 | Judgement environment isolation | Independent judgement environment + snapshot freezing |
| | F-JDG-07 | Offline rescoring | Re-judge existing trajectories with a different scorer |
| **Evaluation Runs** | F-RUN-01 | Evaluation plan creation | Task sets, Agent, repeat count, budget, judgement configuration |
| | F-RUN-02 | Cost estimation | token / cost / duration range + confidence annotation |
| | F-RUN-03 | Run orchestration | Concurrency, rate limiting, budget circuit breaker, pause/cancel |
| | F-RUN-04 | Resume from checkpoint | Resume by (task, index, seed) + configuration consistency check |
| | F-RUN-05 | Real-time progress | Completed count, failure classification, failure rate |
| | F-RUN-06 | Failure classification and retry | Six failure classes + automatic retry policy |
| | F-RUN-07 | Trajectory collection | Interaction sequence, usage, termination information, environment evidence |
| | F-RUN-08 | User simulator access | Persona, disclosure policy, end signal, noise monitoring |
| **Reporting and Analysis** | F-RPT-01 | Single-run report | Confidence section first + main score + breakdown + failure analysis |
| | F-RPT-02 | Breakdown view | By domain / capability / difficulty / split |
| | F-RPT-03 | Step-by-step score curve | Reconstruct the score jump points |
| | F-RPT-04 | Comparative analysis | Total score / breakdown / per-task / cost / behaviour / trajectory differences + significance |
| | F-RPT-05 | Baseline management | Multiple baselines + automatic comparison + staleness marking |
| | F-RPT-06 | Trajectory viewer | Replay, filtering, side-by-side comparison, export |
| | F-RPT-07 | Error attribution | Five-layer attribution + failure-mode clustering + manual re-judgement |
| | F-RPT-08 | Report export | Static page, audit bundle, dataset/log bundle |
| **Metrics** | F-MTR-01 | Metric definition registration | The single source of truth for unified definitions |
| | F-MTR-02 | Aggregation method selection | macro / micro + explicit annotation |
| | F-MTR-03 | Confidence interval | Binomial normal / Wilson / bootstrap resampling |
| | F-MTR-04 | Coverage and NA handling | Effective task coverage displayed explicitly |
| | F-MTR-05 | Efficiency and cost metrics | Steps, token, cost, invalid action rate, repetition rate |
| | F-MTR-06 | Cost-success rate frontier | Two-dimensional curve |
| **Security** | F-SEC-01 | Attack library | Attack strategies × injection vectors |
| | F-SEC-02 | Automatic injection point discovery | Probes locate visible injection points |
| | F-SEC-03 | Defense library | Pluggable defense components |
| | F-SEC-04 | Attack-defense experiment run | Single-variable control |
| | F-SEC-05 | Dual-axis report | Utility / targeted ASR / untargeted ASR / trade-off curves |
| | F-SEC-06 | External request control | Allowlist, interception, irreversible action control, rate cap |
| | F-SEC-07 | Credential management | One-time credential generation and destruction |
| **Leaderboard and Collaboration** | F-LDR-01 | Leaderboard management | Segmented by task set and version |
| | F-LDR-02 | Submission and recomputation | The platform recomputes the scores |
| | F-LDR-03 | Anti-cheat verification | Log support check |
| | F-LDR-04 | Leaderboard guardrails | Saturation, contamination, solvability, sample size, timeliness |
| | F-LDR-05 | Disclosure requirements | Full disclosure of the Agent form |
| **Human Review** | F-HUM-01 | Review workbench | Trajectory + evidence + judgement + re-judgement |
| | F-HUM-02 | Re-judgement audit trail | Author, timestamp, reason, history cannot be deleted |
| | F-HUM-03 | Review sampling | Mandatory review items + random ratio |
| | F-HUM-04 | Automatic scoring bias rate | Bias metric produced by sampled review |
| | F-HUM-05 | Judge calibration set | Annotation set + agreement rate + agreement coefficient |
| **Platform Governance** | F-GOV-01 | Plan fingerprint | Basis for the global consistency check |
| | F-GOV-02 | Platform health | Six reliability metrics and threshold behaviour |
| | F-GOV-03 | Benchmark regression suite | Must run on every platform upgrade |
| | F-GOV-04 | Audit log | Full traceability of task/judgement/submission changes |
| | F-GOV-05 | Permissions and roles | Functional and data permissions of the six roles |

## 7.3 Interface and Interaction

### 7.3.1 Page Inventory

| Page | Main Content | Primary Users |
|---|---|---|
| **Home / Dashboard** | Recent runs, score trends of my systems under test, pending items (pending review, tasks awaiting validation, failures to investigate) | All |
| **System Under Test Details** | Access configuration, self-check results, historical evaluations, score trends | Agent Developer |
| **Task Set Browser** | Task list (filter by domain/capability/difficulty/split), task details, health | All |
| **Task Editor** | Step-by-step form: identity → tagging → description → initial state → contract → judgement → validation | Task Author |
| **Task Validation Report** | Item-by-item results of the eight checks and repair guidance | Task Author |
| **Evaluation Plan Editor** | Select task sets, Agent, repeat count, budget, judgement configuration, concurrency; real-time cost estimate | Initiator |
| **Run Details** | Real-time progress, failure classification distribution, failure rate, log stream, pause/cancel | Initiator |
| **Run Report** | Confidence section → main score → breakdown → failure analysis → trajectory drill down | Initiator / Reviewer |
| **Comparison Page** | Layer-by-layer comparison of two or more runs + significance markers | Initiator / Reviewer |
| **Trajectory Viewer** | Trial-by-trial replay, filtering, side-by-side comparison | Developer / Reviewer |
| **Review Workbench** | Judgement evidence, re-judgement, dispute marking | Reviewer |
| **Leaderboard Page** | Rankings, confidence column, guardrail markers, submission entry point | Public |
| **Task Author Dashboard** | My tasks, pending validation, failed validation items, health alerts | Task Author |
| **Platform Governance Page** | Health metrics, failure tickets, regression suite status | Operator |

### 7.3.2 Interaction Principles

| Principle | Concrete Behaviour |
|---|---|
| **Confidence first** | The top of the report page is always the confidence section (platform failure rate, judgement failure rate, effective sample size); collapsing or hiding it is not allowed |
| **Exceptions are explicit** | Any "excluded", "NA", or "no coverage" value must be visually distinguished from a normal value, and must be expandable to show the reason |
| **Numbers are drillable** | Every aggregate number can be clicked to open the list of Trials that compose it |
| **Differentiated annotation** | Different aggregation methods (macro/micro) for the same number are displayed side by side and labelled with the method name; showing only one is not allowed |
| **Same-batch comparison** | Multiple runs under the same task set version are displayed side by side on one page, to avoid mistaken cross-version comparison |
| **Failures are attributable** | A failed task shows its attribution conclusion by default, instead of displaying only "failed" |
| **Explicit override** | When overriding default budget, concurrency, judgement configuration, and similar behaviour, the override must be performed explicitly and left in the audit trail |

### 7.3.3 Key Interaction Flow Details

**Cost estimate feedback for the run configuration:** When the repeat count, model, or task set is changed, the estimated token / cost / duration range and the confidence annotation update in real time; when the estimate is unreliable (insufficient historical samples), the interface explicitly shows "low estimate confidence".

**Pause and cancel:** Pausing preserves the completed Trials; when cancelling, distinguish "cancel and keep the completed results" from "cancel and void this run" — the latter requires a second confirmation.

**Failure retry:** The report provides a "rerun only these" action for failed Trials; a rerun is appended as a new Trial index and does not overwrite the historical record.

**Offline rescoring:** For a completed run, the user can specify a new scorer to re-judge it; the interface explicitly states "this action does not rewrite the original results and will produce a new version of the judgement", and supports comparing the two versions.

**Re-judgement:** A re-judgement requires a reason; the reason is mandatory; after a re-judgement the metrics are recomputed automatically and the interface reports "recomputed".

## 7.4 Permission Model

| Role | System Under Test | Tasks (draft/published) | Runs | Judgements and Re-judgements | Leaderboard |
|---|---|---|---|---|---|
| Agent Developer | Own creations, editable | Read-only | Can create, run, cancel | Read-only | Can submit their own scores |
| Task Author | Read-only | Drafts are editable; published tasks require a revision request | Can run their own tasks | Read-only | Not allowed |
| Benchmark Maintainer | Read-only | Can publish/freeze/roll back | Can run | Read-only | Can maintain the leaderboard |
| Evaluation Organizer | Read-only | Read-only | Can create, pause/cancel, rerun | Read-only | Can submit, can publish |
| Reviewer | Read-only | Read-only | Read-only | Can re-judge (with audit trail) | Read-only |
| Platform Operator | All | All | All | All | All |

**Key constraint:** **No one may modify the judgement logic behind a score they submitted themselves.** Re-judgement rights and submission rights are separated, so that nobody is both the player and the referee.
