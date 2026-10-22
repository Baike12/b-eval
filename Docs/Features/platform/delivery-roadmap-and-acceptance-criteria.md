# FEAT-008: Delivery Roadmap and Acceptance Criteria

| Field | Value |
|---|---|
| **ID** | FEAT-008 |
| **Status** | Draft |
| **Owner** | Platform team |
| **Priority** | P1 |
| **Implements** | `N/A — this feature defines the delivery plan; it implements no functional requirement from the inventory.` |
| **Created** | 2026-10-22 |
| **Last Updated** | 2026-10-22 |
| **Source** | `Functional design chapter 08 — Roadmap and Acceptance` |

## Summary

The evaluation platform is delivered as five ordered milestones, M0 through M4, each ending with an acceptance-verifiable functional closed loop. Every milestone is gated by the same set of cross-phase continuous acceptance items, and the plan ships with a risk register that names the impact of each risk and its mitigation. The single most important outcome is a phased delivery order in which no half-finished work is stacked up and each phase is independently verifiable before the next one begins.

## Problem

**Current behaviour:** The functional design specifies the platform chapter by chapter (see [FEAT-001 Platform Scope and Positioning][feat-001] through [FEAT-007 User Flows and Feature Inventory][feat-007]), but nothing in the source design sequences them. There is no ordering of the capabilities, no definition of what a phase is, no statement of what makes a phase finished, and no record of which work belongs to a later phase.

**Pain:** Without an ordered, gated plan, "first make it run, then make it trustworthy, finally make it operable" cannot be enforced — half-finished work gets stacked up and no phase closes as a functional loop. Each phase would then be judged on partial capability, and the risks named in the register below (scoring invisibility against a public leaderboard, judge noise, platform failures contaminating scores, tasks leaking into training sets) would surface only after the capability was built.

**Evidence:** None recorded in the source design — no tickets, logs, user reports, or measurements accompany it.

## Proposed Solution

Publish a five-milestone roadmap — M0 Skeleton and Integration Bring-up, M1 Task System and Scorer System Come of Age, M2 Execution Orchestration and Report Analysis, M3 Quality Assurance and Safety Evaluation, M4 Leaderboard and Operations — in which every milestone states its goal, its deliverables and the acceptance criterion each deliverable must meet. A shared set of continuous acceptance items must pass at the end of every phase, so each phase closes as an acceptance-verifiable functional loop rather than a partially finished one. A risk register with an impact and a mitigation for each risk accompanies the plan, and the technical-design material that this document deliberately excludes is listed explicitly so that it can be written up separately.

## Scope

### In Scope

- Milestone **M0 · Skeleton and Integration Bring-up**: goal, deliverables and the acceptance criterion of each deliverable, so that an Agent can run one task on the platform and obtain a score.
- Milestone **M1 · Task System and Scorer System Come of Age**: goal, deliverables and the acceptance criterion of each deliverable, so that tasks can be managed systematically and scoring capability can be reused.
- Milestone **M2 · Execution Orchestration and Report Analysis**: goal, deliverables and the acceptance criterion of each deliverable, so that runs are reliable, understandable and comparable.
- Milestone **M3 · Quality Assurance and Safety Evaluation**: goal, deliverables and the acceptance criterion of each deliverable, so that scores are trustworthy, cheating can be prevented, and adversarial evaluation is possible.
- Milestone **M4 · Leaderboard and Operations**: goal, deliverables and the acceptance criterion of each deliverable, so that results can be published externally and audited by third parties.
- The cross-phase continuous acceptance items that must pass at the end of every phase as part of phase acceptance.
- The risk register that accompanies the plan, with the impact and the mitigation of each risk.
- The explicit list of technical-design content that is complementary to this design but out of scope for this document.

### Out of Scope

The following content belongs to the **technical design** and must be written up separately:

- The technical architecture and layered implementation of the platform — belongs to the technical design.
- The interface definitions and reference implementation of the Agent adapter protocol — belongs to the technical design.
- The storage formats and serialization of tasks, scorers, and metrics — belongs to the technical design.
- The concrete mechanism for scoring invisibility (encrypted distribution vs. an independent judgement environment, see Q2) — belongs to the technical design.
- The implementation selection and topology of the environment sandbox — belongs to the technical design.
- The mechanism implementation of concurrency scheduling, rate limiting, and circuit breaking — belongs to the technical design.
- The front-end form of reports and leaderboards — belongs to the technical design.
- The concrete authorization implementation of the permission model — belongs to the technical design.
- The packaging format of the audit package and the recomputation tool — belongs to the technical design.

The positions marked `[pending technical design]` in this document correspond one-to-one with the above content.

## Requirements

Numbered, independently testable, and each traceable to an acceptance criterion.

- **R1** The delivery plan must deliver milestone **M0 · Skeleton and Integration Bring-up** — an Agent can run one task on the platform and obtain a score — with every deliverable meeting its stated acceptance criterion.

  **M0 · Skeleton and Integration Bring-up**

  **Goal:** An Agent can run one task on the platform and obtain a score.

  Related chapters: [FEAT-003 Evaluation Execution and Orchestration][feat-003].

    | Deliverable | Acceptance criteria |
    |---|---|
    | Subject-under-test registration and integration adapter | The Agent requiring the least modification completes integration in ≤ 1 day |
    | Integration self-check | The self-check can report action format errors and contract violations; the evaluation capability is disabled when the self-check fails |
    | Minimal task model, both forms | An `environment-agentic` task supports the five elements "instruction + initial state + action space + judgement criteria + budget"; a `static-qa` task supports "prompt + termination condition + judgement criteria + budget" and is admitted without an environment |
    | Task form enforcement | A task that declares a form and carries a field its form does not use is rejected, and the rejection names the offending field |
    | Two scorer types (state validation, rule-based criteria) | At least 10 sample tasks, with verdicts matching human judgement 100% |
    | Single run and trajectory | The trajectory fully records action/observation/usage/termination information and can be replayed |
    | Metric registry | The built-in metrics are registered and locked; a registration missing a mandatory field is rejected with the missing field named; a registered custom metric appears in the report marked `builtin: false` |
    | Basic metrics | SR, pass@k, pass^k, avg@k are computed correctly (asserted against known-answer cases) |
    | Repository quality gates | `b-eval gate` runs the whole gate set; a failing gate blocks the pipeline and appends an evidence record; a deliberately broken fixture makes the corresponding gate fail |

  **Out of scope:** Task set management, report breakdown, leaderboard. (Deferred to later milestones of this same roadmap — M1, M2 and M4 — rather than excluded from the roadmap.)

- **R2** The delivery plan must deliver milestone **M1 · Task System and Scorer System Come of Age** — tasks can be managed systematically and scoring capability can be reused — with every deliverable meeting its stated acceptance criterion.

  **M1 · Task System and Scorer System Come of Age**

  **Goal:** Tasks can be managed systematically, and scoring capability can be reused.

  Related chapters: [FEAT-002 Capability Taxonomy and Task System][feat-002], [FEAT-004 Scoring and Metrics][feat-004].

  | Deliverable | Acceptance criteria |
  |---|---|
  | Task CRUD + capability/domain/difficulty tagging | Mutually exclusive label checks take effect; the four difficulty tiers are defined |
  | Task set version management and freezing | Modifications after release are blocked; version diffs are inspectable |
  | All six scorer types available | At least 2 real scorer implementations per type; each type has "known-pass/known-fail" regression samples |
  | Scorer registration and composition | Supports conjunction/disjunction/weighting/partial credit; partial-credit scores map item by item to the criteria and are inspectable |
    | Task ingestion validation gate | All eight checks are implemented for `environment-agentic` tasks and the six applicable checks for `static-qa` tasks; any failing check blocks ingestion and provides remediation guidance; a check that does not apply to the task's form is reported as "not applicable — no environment" rather than as a pass |
  | Environment isolation and four-phase lifecycle | Two concurrent Trials do not contaminate each other; no residual resources remain after cleanup |
  | Scoring invisibility | Judgement criteria cannot be retrieved within the Agent's accessible scope; ingestion scanning intercepts leaked tasks |
  | Introduce ≥ 2 external benchmarks | At least the SWE-bench and τ²-bench subsets are imported, the mapping is completed, and baselines are re-run |

- **R3** The delivery plan must deliver milestone **M2 · Execution Orchestration and Report Analysis** — runs reliably, are understandable, and are comparable — with every deliverable meeting its stated acceptance criterion.

  **M2 · Execution Orchestration and Report Analysis**

  **Goal:** Runs reliably, is understandable, and is comparable.

  Related chapters: [FEAT-003 Evaluation Execution and Orchestration][feat-003], [FEAT-005 Reporting, Analysis and Comparison][feat-005].

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

- **R4** The delivery plan must deliver milestone **M3 · Quality Assurance and Safety Evaluation** — scores are trustworthy, cheating can be prevented, and adversarial evaluation is possible — with every deliverable meeting its stated acceptance criterion.

  **M3 · Quality Assurance and Safety Evaluation**

  **Goal:** Scores are trustworthy, cheating can be prevented, and adversarial evaluation is possible.

  Related chapters: [FEAT-006 Security, Robustness and Quality Assurance][feat-006].

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

- **R5** The delivery plan must deliver milestone **M4 · Leaderboard and Operations** — results can be published externally and audited by third parties — with every deliverable meeting its stated acceptance criterion.

  **M4 · Leaderboard and Operations**

  **Goal:** Can be published externally and can be audited by third parties.

  Related chapters: [FEAT-005 Reporting, Analysis and Comparison][feat-005].

  | Deliverable | Acceptance criteria |
  |---|---|
  | Leaderboard publication | Partitioned by task set and version; leaderboard entries display the complete confidence column |
  | Submission flow | The platform recomputes the score; submissions are rejected when the recomputed value does not match the submitted value |
  | Anti-cheating verification | All three verifications (evidence support, verdict self-consistency, fingerprint consistency) are implemented; pass claims without submitted logs are rejected |
  | Disclosure requirements | Submissions missing the Agent form disclosure are blocked |
  | Leaderboard guardrails | The four guardrails — saturation, contamination, sample size, timeliness — take effect |
  | Audit package export | A self-contained audit package allows a third party to recompute the same score |
  | Permission model enforcement | Separation of re-judgement rights and submission rights takes effect |

- **R6** Every milestone of the roadmap must satisfy the following cross-phase continuous acceptance items at the end of the phase, as part of phase acceptance.

  **Cross-Phase Continuous Acceptance Items**

  The following items must pass at the end of every phase, as part of phase acceptance:

  | Item | Standard |
  |---|---|
  | **Benchmark regression suite** | The platform's built-in minimal set + the scorer regression set + metric known-answer cases, 100% passing |
  | **Attribution correctness spot check** | Randomly sample 20 failed Trials; the agreement rate between platform attribution and human attribution is ≥ 85% |
  | **Score drill-down** | Every aggregate number in the report can be traced back to the Trials that compose it |
  | **Untrustworthy numbers are explicit** | All excluded/NA/uncovered values are distinguished from normal values in both the UI and exports |
  | **No integration regression** | Existing subjects under test can continue to be evaluated on a new version without modification |

- **R7** The delivery plan must carry the following risk register, with an impact and a mitigation recorded for every risk.

  **Risk Register**

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

## Acceptance Criteria

Written as verifiable outcomes, derived only from the acceptance criteria and continuous acceptance items stated in the requirements above.

- [ ] **AC1** (R1) Subject-under-test registration and integration adapter — the Agent requiring the least modification completes integration in ≤ 1 day.
- [ ] **AC2** (R1) Integration self-check — the self-check can report action format errors and contract violations, and the evaluation capability is disabled when the self-check fails.
- [ ] **AC3** (R1) Minimal task model — an `environment-agentic` task supports the five elements "instruction + initial state + action space + judgement criteria + budget"; a `static-qa` task supports "prompt + termination condition + judgement criteria + budget" and is admitted into `dev` without an environment snapshot, initial data, initial files, initial credentials, action space or observation form.
- [ ] **AC3a** (R1) Task form enforcement — a `static-qa` task that declares any of those six environment-only fields is rejected with the offending field named, and a task that declares no form at all is rejected.
- [ ] **AC3b** (R1) Metric registry — the built-in metrics are registered and locked; a registration missing its unit, aggregation method, numerator, denominator or formula is rejected with the missing field named; a registration reusing a built-in ID is rejected; a registered custom metric appears in the report marked `builtin: false` and carries a confidence interval when its unit is `ratio`.
- [ ] **AC3c** (R1) Repository quality gates — `b-eval gate` runs the whole gate set and exits non-zero when any gate fails; every gate appends an evidence record that is never rewritten; a deliberately broken fixture makes the corresponding gate fail, proving the gate can block.
- [ ] **AC4** (R1) Two scorer types (state validation, rule-based criteria) — at least 10 sample tasks, with verdicts matching human judgement 100%.
- [ ] **AC5** (R1) Single run and trajectory — the trajectory fully records action/observation/usage/termination information and can be replayed.
- [ ] **AC6** (R1) Basic metrics — SR, pass@k, pass^k, avg@k are computed correctly (asserted against known-answer cases).
- [ ] **AC7** (R2) Task CRUD + capability/domain/difficulty tagging — mutually exclusive label checks take effect, and the four difficulty tiers are defined.
- [ ] **AC8** (R2) Task set version management and freezing — modifications after release are blocked, and version diffs are inspectable.
- [ ] **AC9** (R2) All six scorer types available — at least 2 real scorer implementations per type, and each type has "known-pass/known-fail" regression samples.
- [ ] **AC10** (R2) Scorer registration and composition — supports conjunction/disjunction/weighting/partial credit, and partial-credit scores map item by item to the criteria and are inspectable.
- [ ] **AC11** (R2) Task ingestion validation gate — all eight checks are implemented for `environment-agentic` tasks and the six applicable checks for `static-qa` tasks, any failing check blocks ingestion and provides remediation guidance, and a check that does not apply to the task's form is reported as "not applicable — no environment" rather than as a pass.
- [ ] **AC12** (R2) Environment isolation and four-phase lifecycle — two concurrent Trials do not contaminate each other, and no residual resources remain after cleanup.
- [ ] **AC13** (R2) Scoring invisibility — judgement criteria cannot be retrieved within the Agent's accessible scope, and ingestion scanning intercepts leaked tasks.
- [ ] **AC14** (R2) Introduce ≥ 2 external benchmarks — at least the SWE-bench and τ²-bench subsets are imported, the mapping is completed, and baselines are re-run.
- [ ] **AC15** (R3) Evaluation plan and cost estimate — the estimated token/cost/duration deviates from the actual by ≤ 30% (when sufficient historical data for the same task set exists).
- [ ] **AC16** (R3) Concurrency and adaptive rate limiting — it automatically slows down when the server throttles and automatically speeds up after recovery; no request storms.
- [ ] **AC17** (R3) Budget circuit breaker — all six budget types take effect, and once triggered, the Trial terminates correctly and the termination reason is marked.
- [ ] **AC18** (R3) Failure classification — the six failure categories are classified accurately, and platform failures are excluded from the denominator and listed separately in the report.
- [ ] **AC19** (R3) Checkpoint resume — after an interruption, the resumed run yields results identical to a run completed in one pass (metric consistency rate ≥ 99%).
- [ ] **AC20** (R3) Offline re-scoring — re-judging historical trajectories with a different scorer succeeds, and the original verdict is retained and comparable.
- [ ] **AC21** (R3) Single-run report — the confidence section comes first, the breakdown section covers domain/capability/difficulty/split, and failure analysis includes termination reasons and scorer details.
- [ ] **AC22** (R3) Comparative analysis — six layers of difference: total score/breakdown/per-task/cost/behavior/trajectory, and significance marking is correct (no improvement is marked when the confidence interval of the difference crosses 0).
- [ ] **AC23** (R3) Baseline management — baselines are compared automatically, and stale baselines are automatically flagged.
- [ ] **AC24** (R4) Scorer regression and version binding — releases are blocked when scorer changes fail the regression set, and the report displays the scorer version.
- [ ] **AC25** (R4) Task health audit — the five audits can run on a schedule and produce alerts, and alerts can enter the remediation workflow.
- [ ] **AC26** (R4) Human review workbench — re-judgement is fully recorded (author/time/reason), and history cannot be deleted.
- [ ] **AC27** (R4) Review sampling and deviation rate — mandatory review items automatically enter the queue, and the deviation rate of automatic scoring can be computed.
- [ ] **AC28** (R4) Judge calibration — a human-annotated set is established, and judges with an agreement rate below the threshold are banned from official evaluation.
- [ ] **AC29** (R4) Adversarial evaluation — automatic discovery of attack libraries/defense libraries/injection points is available, a dual-axis report is produced, and the attack-defense tradeoff curve is correct.
- [ ] **AC30** (R4) External request governance — allowlist, interception, irreversible-action control, and rate caps take effect.
- [ ] **AC31** (R4) Platform health — the six metrics and their threshold behaviors take effect, and when the failure rate exceeds the threshold, the report is marked "insufficient confidence" and does not participate in the leaderboard.
- [ ] **AC32** (R5) Leaderboard publication — partitioned by task set and version, and leaderboard entries display the complete confidence column.
- [ ] **AC33** (R5) Submission flow — the platform recomputes the score, and submissions are rejected when the recomputed value does not match the submitted value.
- [ ] **AC34** (R5) Anti-cheating verification — all three verifications (evidence support, verdict self-consistency, fingerprint consistency) are implemented, and pass claims without submitted logs are rejected.
- [ ] **AC35** (R5) Disclosure requirements — submissions missing the Agent form disclosure are blocked.
- [ ] **AC36** (R5) Leaderboard guardrails — the four guardrails — saturation, contamination, sample size, timeliness — take effect.
- [ ] **AC37** (R5) Audit package export — a self-contained audit package allows a third party to recompute the same score.
- [ ] **AC38** (R5) Permission model enforcement — separation of re-judgement rights and submission rights takes effect.
- [ ] **AC39** (R6) Benchmark regression suite — the platform's built-in minimal set + the scorer regression set + metric known-answer cases, 100% passing, at the end of every phase.
- [ ] **AC40** (R6) Attribution correctness spot check — randomly sample 20 failed Trials; the agreement rate between platform attribution and human attribution is ≥ 85%.
- [ ] **AC41** (R6) Score drill-down — every aggregate number in the report can be traced back to the Trials that compose it.
- [ ] **AC42** (R6) Untrustworthy numbers are explicit — all excluded/NA/uncovered values are distinguished from normal values in both the UI and exports.
- [ ] **AC43** (R6) No integration regression — existing subjects under test can continue to be evaluated on a new version without modification.

## Impact

| Area | Effect |
|---|---|
| Users | Agent authors gain a staged capability set: from M0 (one task runs and scores on the platform) through M4 (results publishable externally and auditable by third parties). Task-set authors gain systematic task CRUD, tagging, versioning and freezing in M1. Report readers gain the single-run report in M2 and the leaderboard in M4. |
| Existing behaviour | No runtime behaviour of the platform changes on account of this document; it sequences the capabilities defined in the sibling chapters into M0–M4. Inside M0 specifically, task set management, report breakdown and leaderboard are deferred to later milestones. |
| Operational cost | M2 adds a pre-run cost estimate (estimated token/cost/duration deviating from the actual by ≤ 30%), a budget circuit breaker over all six budget types, and budget exhaustion counted in the denominator rather than hidden. M3 adds human review load (mandatory review queue, deviation rate, judge calibration) plus adversarial-evaluation runs. M4 adds score recomputation on every submission and audit package export. |

## Risks and Mitigations

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

## Rollout

The roadmap advances with **"first make it run, then make it trustworthy, finally make it operable"**. Each phase ends with an acceptance-verifiable functional closed loop; no half-finished work is stacked up. The milestones are released strictly in this order:

- **M0 · Skeleton and Integration Bring-up** — Goal: An Agent can run one task on the platform and obtain a score.
- **M1 · Task System and Scorer System Come of Age** — Goal: Tasks can be managed systematically, and scoring capability can be reused.
- **M2 · Execution Orchestration and Report Analysis** — Goal: Runs reliably, is understandable, and is comparable.
- **M3 · Quality Assurance and Safety Evaluation** — Goal: Scores are trustworthy, cheating can be prevented, and adversarial evaluation is possible.
- **M4 · Leaderboard and Operations** — Goal: Can be published externally and can be audited by third parties.

**Phase gate:** a milestone is accepted only when every deliverable in its table meets its stated acceptance criterion **and** all five cross-phase continuous acceptance items pass at the end of the phase — the benchmark regression suite (the platform's built-in minimal set + the scorer regression set + metric known-answer cases) at 100% passing, the attribution correctness spot check (20 randomly sampled failed Trials, agreement rate ≥ 85%), score drill-down, explicit untrustworthy numbers, and no integration regression.

**Rollback condition:** the source design defines the phase-acceptance gate above and names no separate rollback trigger. Delivery therefore holds at the last accepted milestone (M0 → M1 → M2 → M3 → M4) whenever a milestone fails any of its own stated acceptance criteria or any of the five cross-phase continuous acceptance items.

## References

- Tech Design: [TD-004 M0 Runtime and Quality Gate System][td-004]
- Test Report: [TR-003 M0 Vertical Slice][tr-003]
- Platform scope and positioning: [FEAT-001 Platform Scope and Positioning][feat-001]
- Capability taxonomy and task system: [FEAT-002 Capability Taxonomy and Task System][feat-002]
- Evaluation execution and orchestration: [FEAT-003 Evaluation Execution and Orchestration][feat-003]
- Scoring and metrics: [FEAT-004 Scoring and Metrics][feat-004]
- Reporting, analysis and comparison: [FEAT-005 Reporting, Analysis and Comparison][feat-005]
- Security, robustness and quality assurance: [FEAT-006 Security, Robustness and Quality Assurance][feat-006]
- User flows and feature inventory: [FEAT-007 User Flows and Feature Inventory][feat-007]
- This feature document: [FEAT-008 Delivery Roadmap and Acceptance Criteria][feat-008]

## Open Questions

Not specified in the source design.

<!-- Link definitions stay at the bottom of the file. Angle brackets are required
     because the link targets contain no spaces but the pattern keeps the
     reference definitions uniform with the feature template. -->

[feat-001]: <./platform-scope-and-positioning.md>
[feat-002]: <../task-system/capability-taxonomy-and-task-system.md>
[feat-003]: <../execution/evaluation-execution-and-orchestration.md>
[feat-004]: <../scoring/scoring-and-metrics.md>
[feat-005]: <../reporting/reporting-analysis-and-comparison.md>
[feat-006]: <../security/security-robustness-and-quality-assurance.md>
[feat-007]: <../experience/user-flows-and-feature-inventory.md>
[feat-008]: <./delivery-roadmap-and-acceptance-criteria.md>
[td-004]: <../../Tech Design/TD-004-m0-runtime-and-quality-gate-system.md>
[tr-003]: <../../Test Report/TR-003-m0-vertical-slice.md>
