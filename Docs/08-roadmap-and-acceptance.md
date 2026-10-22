# 08 · Evolution Roadmap and Acceptance

## 8.1 Phase Breakdown

The roadmap advances with **"first make it run, then make it trustworthy, finally make it operable"**. Each phase ends with an acceptance-verifiable functional closed loop; no half-finished work is stacked up.

### M0 · Skeleton and Integration Bring-up

**Goal:** An Agent can run one task on the platform and obtain a score.

| Deliverable | Acceptance criteria |
|---|---|
| Subject-under-test registration and integration adapter | The Agent requiring the least modification completes integration in ≤ 1 day |
| Integration self-check | The self-check can report action format errors and contract violations; the evaluation capability is disabled when the self-check fails |
| Minimal task model | Supports the five elements "instruction + initial state + action space + judgement criteria + budget" |
| Two scorer types (state validation, rule-based criteria) | At least 10 sample tasks, with verdicts matching human judgement 100% |
| Single run and trajectory | The trajectory fully records action/observation/usage/termination information and can be replayed |
| Basic metrics | SR, pass@k, pass^k, avg@k are computed correctly (asserted against known-answer cases) |

**Out of scope:** Task set management, report breakdown, leaderboard.

### M1 · Task System and Scorer System Come of Age

**Goal:** Tasks can be managed systematically, and scoring capability can be reused.

| Deliverable | Acceptance criteria |
|---|---|
| Task CRUD + capability/domain/difficulty tagging | Mutually exclusive label checks take effect; the four difficulty tiers are defined |
| Task set version management and freezing | Modifications after release are blocked; version diffs are inspectable |
| All six scorer types available | At least 2 real scorer implementations per type; each type has "known-pass/known-fail" regression samples |
| Scorer registration and composition | Supports conjunction/disjunction/weighting/partial credit; partial-credit scores map item by item to the criteria and are inspectable |
| Task ingestion validation gate | All eight checks are implemented; any failing check blocks ingestion and provides remediation guidance |
| Environment isolation and four-phase lifecycle | Two concurrent Trials do not contaminate each other; no residual resources remain after cleanup |
| Scoring invisibility | Judgement criteria cannot be retrieved within the Agent's accessible scope; ingestion scanning intercepts leaked tasks |
| Introduce ≥ 2 external benchmarks | At least the SWE-bench and τ²-bench subsets are imported, the mapping is completed, and baselines are re-run |

### M2 · Execution Orchestration and Report Analysis

**Goal:** Runs reliably, is understandable, and is comparable.

| Deliverable | Acceptance criteria |
|---|---|
| Evaluation plan and cost estimate | The estimated token/cost/duration deviates from the actual by ≤ 30% (when sufficient historical data for the same task set exists) |
| Concurrency and adaptive rate limiting | It automatically slows down when the server throttles and automatically speeds up after recovery; no request storms |
| Budget circuit breaker | All six budget types take effect; once triggered, the Trial terminates correctly and the termination reason is marked |
| Failure classification | The six failure categories are classified accurately; platform failures are excluded from the denominator and listed separately in the report |
| Checkpoint resume | After an interruption, the resumed run yields results identical to a run completed in one pass (metric consistency rate ≥ 99%) |
| Offline re-scoring | Re-judging historical trajectories with a different scorer succeeds; the original verdict is retained and comparable |
| Single-run report | The confidence section comes first; the breakdown section covers domain/capability/difficulty/split; failure analysis includes termination reasons and scorer details |
| Comparative analysis | Six layers of difference: total score/breakdown/per-task/cost/behavior/trajectory; significance marking is correct (no improvement is marked when the confidence interval of the difference crosses 0) |
| Baseline management | Baselines are compared automatically; stale baselines are automatically flagged |

### M3 · Quality Assurance and Safety Evaluation

**Goal:** Scores are trustworthy, cheating can be prevented, and adversarial evaluation is possible.

| Deliverable | Acceptance criteria |
|---|---|
| Scorer regression and version binding | Releases are blocked when scorer changes fail the regression set; the report displays the scorer version |
| Task health audit | The five audits can run on a schedule and produce alerts; alerts can enter the remediation workflow |
| Human review workbench | Re-judgement is fully recorded (author/time/reason); history cannot be deleted |
| Review sampling and deviation rate | Mandatory review items automatically enter the queue; the deviation rate of automatic scoring can be computed |
| Judge calibration | A human-annotated set is established; judges with an agreement rate below the threshold are banned from official evaluation |
| Adversarial evaluation | Automatic discovery of attack libraries/defense libraries/injection points is available; a dual-axis report is produced; the attack-defense tradeoff curve is correct |
| External request governance | Allowlist, interception, irreversible-action control, and rate caps take effect |
| Platform health | The six metrics and their threshold behaviors take effect; when the failure rate exceeds the threshold, the report is marked "insufficient confidence" and does not participate in the leaderboard |

### M4 · Leaderboard and Operations

**Goal:** Can be published externally and can be audited by third parties.

| Deliverable | Acceptance criteria |
|---|---|
| Leaderboard publication | Partitioned by task set and version; leaderboard entries display the complete confidence column |
| Submission flow | The platform recomputes the score; submissions are rejected when the recomputed value does not match the submitted value |
| Anti-cheating verification | All three verifications (evidence support, verdict self-consistency, fingerprint consistency) are implemented; pass claims without submitted logs are rejected |
| Disclosure requirements | Submissions missing the Agent form disclosure are blocked |
| Leaderboard guardrails | The four guardrails — saturation, contamination, sample size, timeliness — take effect |
| Audit package export | A self-contained audit package allows a third party to recompute the same score |
| Permission model enforcement | Separation of re-judgement rights and submission rights takes effect |

## 8.2 Cross-Phase Continuous Acceptance Items

The following items must pass at the end of every phase, as part of phase acceptance:

| Item | Standard |
|---|---|
| **Benchmark regression suite** | The platform's built-in minimal set + the scorer regression set + metric known-answer cases, 100% passing |
| **Attribution correctness spot check** | Randomly sample 20 failed Trials; the agreement rate between platform attribution and human attribution is ≥ 85% |
| **Score drill-down** | Every aggregate number in the report can be traced back to the Trials that compose it |
| **Untrustworthy numbers are explicit** | All excluded/NA/uncovered values are distinguished from normal values in both the UI and exports |
| **No integration regression** | Existing subjects under test can continue to be evaluated on a new version without modification |

## 8.3 Risk Register

| Risk | Impact | Mitigation |
|---|---|---|
| **Scoring invisibility conflicts with a public leaderboard** | Third-party submitters can see the scoring logic → score gaming | See Q3. Preferred approach: the public leaderboard uses only scorers from private task sets; public sets are used only for self-evaluation |
| **judge noise** | The error introduced by model-based judging exceeds the differences being measured | Judge calibration + confidence intervals + mandatory "conclusions must include evidence"; key conclusions require human spot checks |
| **Task production cost** | Building high-quality task sets takes a long time | Task templating + external benchmark import + community submissions + the ingestion gate guarantees quality rather than quantity |
| **The adapter protocol is too permissive** | Various Agent behaviors are not comparable | The protocol only fixes the boundaries (action/observation/termination/usage); comparability is guaranteed by "contract + unified budget accounting", and the report annotates each Agent's deviations |
| **Platform failures contaminate scores** | Scores are untrustworthy | Failures are excluded from the denominator + reports are vetoed when the failure rate exceeds the threshold + platform health is published |
| **Tasks leak into training sets** | Scores are inflated | Private task sets + contamination probes + canary + periodic rotation |
| **Scorers are over-optimized against** | Real capability is measured inaccurately | Scorer regression set + scorer version persisted alongside the score + periodic replacement of scoring implementations |
| **Evaluation cost out of control** | The project is unsustainable | Pre-run cost estimation + budget circuit breaker + budget exhaustion counted in the denominator (surfacing efficiency problems rather than hiding them) |
| **Capability taxonomy disputes** | Breakdown results are questioned | Public taxonomy, annotation rationale recorded, disputes can be appealed and trigger re-annotation |

## 8.4 Content Complementary to This Design but Out of Scope for This Document

The following content belongs to the **technical design** and must be written up separately:

- The technical architecture and layered implementation of the platform
- The interface definitions and reference implementation of the Agent adapter protocol
- The storage formats and serialization of tasks, scorers, and metrics
- The concrete mechanism for scoring invisibility (encrypted distribution vs. an independent judgement environment, see Q2)
- The implementation selection and topology of the environment sandbox
- The mechanism implementation of concurrency scheduling, rate limiting, and circuit breaking
- The front-end form of reports and leaderboards
- The concrete authorization implementation of the permission model
- The packaging format of the audit package and the recomputation tool

The positions marked `[pending technical design]` in this document correspond one-to-one with the above content.
