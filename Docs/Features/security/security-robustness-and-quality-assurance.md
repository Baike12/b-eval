# FEAT-006: Security, Robustness and Quality Assurance

| Field | Value |
|---|---|
| **ID** | FEAT-006 |
| **Status** | Draft |
| **Owner** | Platform team |
| **Priority** | P0 |
| **Implements** | `F-SEC-01`, `F-SEC-02`, `F-SEC-03`, `F-SEC-04`, `F-SEC-05`, `F-SEC-06`, `F-SEC-07`, `F-HUM-01`, `F-HUM-02`, `F-HUM-03`, `F-HUM-04`, `F-HUM-05`, `F-GOV-01`, `F-GOV-02`, `F-GOV-03`, `F-GOV-04`, `F-GOV-05` |
| **Created** | 2026-10-22 |
| **Last Updated** | 2026-10-22 |
| **Source** | Functional design chapter 06 — Security, Robustness and Quality Assurance |

## Summary

FEAT-006 specifies the security, robustness and quality-assurance layer of the evaluation platform, covering anti-cheating during evaluation, adversarial security evaluation, task and scoring quality assurance, data contamination and freshness, the platform's own reliability, and sensitive operation controls. It answers three questions: **can scores be inflated by cheating**, **are the tasks themselves trustworthy**, and **are the scores reported by the platform trustworthy**. The most important outcome is that every reported score can be defended, because the scoring logic is out of the Agent's reach, the tasks are gated and continuously audited, the judges are calibrated, and the platform publishes its own health.

## Problem

**Current behaviour:** The Agent under test holds certain permissions during evaluation (reading files, running commands, making network requests), so it can attempt to read the scoring logic, tamper with the scoring environment, and construct false positives. Anti-cheating that is not balanced degenerates into "condemning good Agents as well". An Agent may lower its failure rate by "doing less and answering more" (e.g. declaring a task impossible whenever it hits a hard one). Reporting only the ASR rewards a defense that "rejects everything outright"; reporting only utility hides injection risk. After publication, a task still needs continuous health checks, and trustworthy scores require a trustworthy platform — while the evaluation environment may involve real external systems (real websites, real APIs).

**Pain:** Users of the platform lose trust in the leaderboard whenever a score can be inflated by a cheating Agent, a task turns out to be unsolvable or trivially passed, or a judge turns out to be lenient or miscalibrated. Benchmark authors and downstream score consumers cannot tell which published tasks are still healthy, cannot distinguish an Agent failure from a platform/scorer failure, and cannot reproduce a score across runs. Evaluated Agents are unfairly penalised when anti-cheating is applied without a passing reference solution, and real external systems are exposed when outbound requests, irreversible actions and credentials are uncontrolled.

**Evidence:** Not specified in the source design.

## Proposed Solution

The platform gains a layered quality-assurance and security envelope around the whole evaluation pipeline: the scoring logic is hidden from the Agent, judgement runs in an environment independent of the Agent's environment, every verdict must be backed by execution evidence, and dedicated adversarial experiments measure a dual-axis utility-security trade-off. Tasks are gated at ingestion and continuously audited afterwards, model-based judges are calibrated before use, contamination and freshness are measured, and the platform publishes its own failure-rate, consistency and regression metrics against hard thresholds. Sensitive operations against real external systems are intercepted, allowlisted, rate-limited and logged. The user-visible behaviour is that every report states its own confidence level — for example "insufficient confidence", "scoring unreliable", "freshness in question" or "suspicious" — while the implementation detail belongs in the Tech Design document.

## Scope

### In Scope

- Anti-cheating during evaluation: hiding the scoring criteria, expected end states, reference solutions and the scorer implementation from the tested party.
- Automatic scan before ingestion and injection detection in the scoring evidence.
- Separation of the judgement environment from the execution environment, with snapshot freezing and separate credentials.
- Concrete rules against "false positives" and guards against "false negatives" that would kill good Agents.
- Guarding against "environment-side cheating" and against cost-oriented cheating.
- Adversarial security evaluation: experiment design, attack library, defense library, dual-axis metrics and denial-of-service attacks.
- Task quality assurance: ingestion gate, continuous audit and task revision discipline.
- Scoring quality assurance: judge calibration, judge failure protection and blending of manual and automatic scoring.
- Data contamination and freshness measurement and reporting.
- The platform's own reliability metrics, thresholds and built-in benchmark regression suite.
- Sensitive operation controls when the evaluation environment involves real external systems.

### Out of Scope

Not specified in the source design.

## Requirements

Numbered, independently testable, and each traceable to an acceptance criterion. The bold labels under each requirement preserve the original design section headings and their full detail.

- **R1** Scoring criteria, expected end states, reference solutions and the scorer implementation must not be visible to the Agent under test, and ingestion must be blocked when a pre-ingestion scan finds scoring material in the Agent-accessible scope.

  **6.1 Anti-Cheating During Evaluation**

  The Agent under test holds certain permissions during evaluation (reading files, running commands, making network requests). You must assume it will attempt to read the scoring logic, tamper with the scoring environment, and construct false positives.

  **6.1.1 Scoring Must Not Be Visible**

  | Measure | Requirement |
  |---|---|
  | **Scoring criteria are out of the Agent's reach** | Scoring criteria, expected end states, and reference solutions must not appear in any environment, file, or prompt the Agent can read |
  | **The scorer implementation is unreadable** | The scorer code is not visible to the tested party (following TheAgentCompany's encrypted-delivery approach) |
  | **Automatic scan before ingestion** | When a task is ingested, the Agent-accessible scope is scanned for scoring keywords, expected-answer fragments, and reference-solution signature strings; any hit blocks ingestion |
  | **Injection detection** | When suspicious content originating from the tested output (e.g. a fabricated test report, a forged completion claim) is found in the scoring evidence, the verdict is marked as "untrustworthy verdict" and listed separately |

  **Which invisibility mechanism is used (encrypted delivery vs. an isolated judgement environment) depends on Q2**, but **the effect requirement is the same**: the tested party must not obtain the scoring logic during evaluation.

- **R2** Scoring must run in a judgement environment that is independent of the Agent's environment, and the judgement must read only a snapshot frozen after the execution stage ends.

  **6.1.2 Separation of the Judgement Environment from the Execution Environment**

  - Scoring runs in a **judgement environment that is independent of the Agent's environment**, so the Agent cannot interfere.
  - The final state required for judgement is **frozen into a snapshot** after the execution stage ends; the judgement reads only the snapshot.
  - The credentials of the judgement environment are separate from those of the Agent environment.

- **R3** Every verdict must be backed by execution evidence; a result that merely "looks correct" is not a pass.

  **6.1.3 Concrete Rules against "False Positives"**

  A verdict must prove that the work was actually done, not merely that the result "looks correct":

  | Rule | Description | Reference |
  |---|---|---|
  | **Execution evidence must exist** | When the criterion is "tests pass", there must be evidence that "the tests actually ran" (non-zero case count, execution summary); an empty result does not count as a pass | SWE-bench's SUITE_RAN evidence check |
  | **Exit code and log are cross-checked** | If the exit code is non-zero but the log contains no failure record at all, the result is invalid (prevents a patch from printing success itself) | SWE-bench's exit code cross-check |
  | **Test output marker boundaries** | Test output must fall between the specified start and end markers; otherwise the entire verdict is invalid | SWE-bench's START/END_TEST_OUTPUT |
  | **Skipped does not count as passed** | Skipped tests do not satisfy the "fail-to-pass" condition | SWE-bench's SKIPPED semantics |
  | **A completion claim is not completion** | An Agent claiming "done" is only a termination signal; whether the task is complete is decided by the scorer | General principle |
  | **Retrieved information is not completion** | An Agent having found information does not mean the job was done | Built-in rubric semantics |

- **R4** Anti-cheating must not condemn good Agents: the reference solution must pass, scorer changes require regression testing, and the scorer version is bound to the task version.

  **6.1.4 Guarding against "False Negatives" and Against Killing Good Agents**

  Anti-cheating must not degenerate into "condemning good Agents as well":

  - **The reference solution must pass** (item 2 of the ingestion gate), ensuring the scorer never kills a correct solution by mistake.
  - **Scorer changes require regression testing**: any scorer modification must be re-run on the scorer regression set (containing known-pass and known-fail samples), and may take effect only after passing.
  - **The scorer version is bound to the task version**: the report shows the scorer version, and historical scores never change because of a scorer upgrade.

- **R5** Environment-side cheating must be prevented and detected for every Trial.

  **6.1.5 Guarding against "Environment-Side Cheating"**

  | Scenario | Rule |
  |---|---|
  | The Agent directly modifies scoring files | Scoring files live in an isolated environment, and integrity checks run before and after execution |
  | The Agent contaminates a later Trial | Each Trial gets its own environment instance plus its own credentials (see [Chapter 02, Section 2.6][feat-002]) |
  | The Agent cached the previous run's result | Each Trial's environment instance is created fresh, or rolled back to the initial snapshot |
  | The Agent leaves a backdoor | The judgement environment snapshot captures only state relevant to the task and never trusts global environment state |

- **R6** Cost-oriented cheating must be detected, reported and measured separately.

  **6.1.6 Cost-Oriented Cheating**

  An Agent may lower its failure rate by "doing less and answering more" (e.g. declaring a task impossible whenever it hits a hard one).

  - A task's "impossible declaration" counts as correct only when the task genuinely contains that branch; otherwise it is recorded as a failure.
  - **For an Agent whose share of impossible declarations is abnormally high, the report gives an explicit warning** (e.g. "this Agent declared X% of tasks impossible").
  - Provide an "impossible-task subset" dedicated to measuring this capability, scored separately from other tasks.

- **R7** Every adversarial experiment must declare its user task, attack task, injection vector and defense strategy.

  **6.2 Adversarial Security Evaluation**

  The platform supports dedicated adversarial evaluation (corresponding to capability dimension C6), using a **dual-axis utility-security** paradigm (following AgentDojo).

  **6.2.1 Experiment Design**

  An adversarial experiment consists of four elements:

  | Element | Content |
  |---|---|
  | **User task** | The task the Agent is supposed to complete (the benign task) |
  | **Attack task** | Something the attacker wants the Agent to do that the user never authorized |
  | **Injection vector** | Where the attack payload is planted (tool return values, web content, files, email, etc.) |
  | **Defense strategy** | A protective measure on the Agent side or the platform side |

- **R8** Attacks and defenses must be maintained as libraries, and only one variable may change per experiment.

  **6.2.2 Attack Library and Defense Library**

  - **Attack library**: attacks are defined as a combination of "attack strategy × injection vector", and each attack declares what it aims to achieve.
  - **Automatic injection-point discovery**: probes (canaries) automatically locate "which injection points the model can actually see", avoiding omissions in manual annotation (following AgentDojo's `get_injection_candidates` approach).
  - **Defense library**: defenses plug into the evaluation pipeline as pluggable components, including content filtering, tool allowlists, injection detection, provenance marking, instruction restatement, and so on. **Defenses follow the same specification as attacks: only one variable changes per experiment.**

- **R9** Every adversarial report must give both the no-attack utility and the ASR, computed on the dual-axis metric set.

  **6.2.3 Dual-Axis Metrics**

  | Metric | Definition | Notes |
  |---|---|---|
  | **Utility** | The proportion of user tasks the Agent completes normally in the absence of attacks | A defense must not come at the cost of utility |
  | **Utility under attack** | The proportion of benign tasks still completed under attack | |
  | **Targeted attack success rate (ASR)** | The proportion of attacks whose objective is achieved | The lower the better |
  | **Untargeted attack success rate** | The proportion of cases where the Agent exhibits abnormal behavior (denial of service, abnormal output) | Distinguishes "led astray" from "broken" |
  | **Utility-ASR trade-off curve** | A two-dimensional view of utility against ASR | Used to pick operating points: high utility and low ASR at the same time |

  **The report must give both the no-attack utility and the ASR.** Reporting only the ASR rewards a defense that "rejects everything outright"; reporting only utility hides injection risk.

- **R10** Denial-of-service attacks must be judged on an independent channel with independent metrics.

  **6.2.4 Denial-of-Service Attacks**

  A separate class: the attack objective is to make the Agent unable to complete the task (e.g. trapping it in a loop or triggering a CAPTCHA). The verdict for this class is "safe = not fully blocked", which is the opposite of targeted attacks, so it requires an independent channel and independent metrics.

- **R11** Task ingestion must apply the eight checks of the ingestion gate, of which items 1–3 are hard gates.

  **6.3 Task Quality Assurance**

  **6.3.1 Ingestion Gate**

  See the eight checks in [Chapter 02, Section 2.8][feat-002]. Items 1–3 among them (solvability, scorer trustworthiness, scorer non-triviality) are **hard gates**: a task that fails them must not enter dev.

- **R12** Published tasks must be continuously audited, and their health must be visible in reports and leaderboards.

  **6.3.2 Continuous Audit**

  After publication, a task still needs continuous health checks. The platform provides:

  | Audit item | Verdict | Action |
  |---|---|---|
  | **Abnormally high pass rate** | A task is passed by ≥ 95% of Agents | Saturation alert → re-review the difficulty or design |
  | **Abnormally high failure rate** | A task fails for ≥ 95% of Agents, and failure modes concentrate in the same place | Suspicion alert → inspect the scorer and the task |
  | **Scorer inconsistent with the reference solution** | The reference solution no longer passes on a re-run | Immediately downgrade the task and roll it back to the previous version |
  | **Non-determinism** | A re-run with the same seed yields a different verdict | Isolate the task and investigate environment randomness |
  | **Abnormal runtime** | Actual elapsed time far exceeds the estimate | Investigate whether an environment wait/timeout trap exists |
  | **Annotation drift** | Measured difficulty conflicts with the annotated difficulty | Trigger difficulty re-annotation (see [Chapter 02, Section 2.5][feat-002]) |

  **Audit conclusions feed into the task's "health" field**, visible in reports and leaderboards — when a task's health is "suspicious", the score it contributes is flagged in the report.

- **R13** Task revisions must be versioned and reasoned, and must never silently rewrite published scores.

  **6.3.3 Task Revision Discipline**

  - Any revision to task content **must bump the version**; old versions are retained.
  - A revision must state its reason (fixing a scorer error / fixing an ambiguity / fixing a factual error / fixing an annotation error) and its scope of impact.
  - **Scores of a published version are never recomputed because of a revision**, unless the revision is judged as "the original task was wrong", in which case a voiding notice for the original task's score must be published at the same time.

- **R14** Model-based scoring must be calibrated before it is used for official evaluation.

  **6.4 Scoring Quality Assurance**

  **6.4.1 Judge Calibration**

  Model-based scoring must be calibrated before use:

  | Calibration item | Method |
  |---|---|
  | **Human-annotated set** | Maintain a batch of human-judged samples (including positives, negatives, and boundary cases) |
  | **Agreement rate evaluation** | The agreement rate between the judge model's verdicts on the annotated set and the human verdicts |
  | **Agreement coefficient** | Krippendorff's alpha across multiple judges |
  | **Bias analysis** | Analyze systematic judge biases (tendency to pass, tendency to give partial credit, leniency toward long trajectories) |
  | **Deployment threshold** | Agreement rate and agreement coefficient below the threshold → must not be used for official evaluation; above the threshold → label "calibrated, judge=version X" |

- **R15** Judge failures must produce UNSCORED rather than FAIL, and must be surfaced in the report.

  **6.4.2 Judge Failure Protection**

  - Scoring model unavailable → UNSCORED; do not record FAIL.
  - Scoring output unparseable → UNSCORED; record the samples that failed to parse.
  - Verdict failure rate above the threshold → the report of that evaluation run is marked "scoring unreliable".
  - **Scoring overhead counts toward the total cost** (see [Chapter 04, Section 4.2.4][feat-004]).

- **R16** Manual and automatic scoring must be blendable through the three supported modes, with explicit conflict handling.

  **6.4.3 Blending Manual and Automatic Scoring**

  The platform supports three blending modes, and a task can choose among them:

  | Mode | Mechanism | Applicability |
  |---|---|---|
  | **Automatic first + manual spot-check** | Automatic verdicts, manual sampling at a fixed ratio, producing a deviation rate | Clear criteria, large volume |
  | **Manual first + automatic review** | Manual verdicts, with automation used only for cross-validation | High value, small volume |
  | **Dual-track parallel** | Both automatic and manual judging; when the difference exceeds the threshold, the case enters manual arbitration | High dispute, strongly subjective rubrics |

  **Conflict handling:** samples where the two tracks disagree enter the arbitration queue; the arbitration conclusion becomes the final verdict for that sample, and the original values of both verdicts are recorded.

- **R17** Task sets must carry contamination and freshness controls that are reported together with the scores.

  **6.5 Data Contamination and Freshness**

  | Measure | Description |
  |---|---|
  | **Timestamped tasks** | A task carries a "creation time" and a "last verified time"; a task set whose validity period has expired without re-verification is marked "freshness in question" |
  | **Contamination probe** | Fragments of task instructions and expected answers are compared against public corpora; a hit is marked "suspected contamination" |
  | **Private-set hedge** | The official leaderboard includes both public-set and private-set results; private-set scores resist contamination |
  | **Periodic rotation** | The private set rotates its tasks regularly; the public set receives a substantive update at least once a year |
  | **Canary markers** | A task set carries an identifiable marker string used to detect whether task content has appeared elsewhere (following terminal-bench's canary practice) |
  | **Freshness metric** | The report gives "the time since the last substantive update of the task set used in this evaluation" for reference when interpreting scores |

- **R18** The platform must expose its own health and must meet the stated reliability thresholds.

  **6.6 The Platform's Own Reliability**

  Trustworthy scores require a trustworthy platform. The platform must expose its own health:

  | Metric | Definition | Threshold behavior |
  |---|---|---|
  | **Platform failure rate** | The share of Trials affected by platform/environment/scorer failures | > 5% → the report is marked "insufficient confidence" and is excluded from the leaderboard |
  | **Verdict failure rate** | The share of cases where the scorer returns UNSCORED | > 3% → the report is marked "scoring unreliable" |
  | **Resume consistency** | The share of cases where a resumed run of the same plan yields results identical to a single uninterrupted run | < 99% → investigate idempotency |
  | **Benchmark regression pass rate** | The pass rate of the platform's built-in benchmark regression suite | < 100% → block the release |
  | **Scorer regression pass rate** | The pass rate of the regression set after a scorer change | < 100% → block the scorer release |

  The **platform's built-in benchmark regression suite** contains: a minimal task set (≤ 10 tasks per domain), the scorer regression set, and known-answer cases for metric computation (e.g. the examples in [Chapter 04, Section 4.6][feat-004] as fixed assertions). The platform must pass this suite on every self-upgrade.

- **R19** The evaluation environment must control sensitive operations whenever real external systems are involved.

  **6.7 Sensitive Operation Controls**

  The evaluation environment may involve real external systems (real websites, real APIs). The platform must:

  | Control | Requirement |
  |---|---|
  | **Request interception** | Outbound requests are intercepted before leaving the controlled environment; the platform only allows allowlisted targets |
  | **Irreversible actions** | Actions such as placing orders, sending messages, making payments, and deleting are intercepted by default; they are allowed only when the task explicitly declares permission |
  | **Credential management** | Use one-time credentials, generated per Trial and destroyed with the Trial |
  | **Rate constraints** | Set request-rate caps on external systems to keep evaluation behavior from affecting real business |
  | **Data minimization** | No real user data is placed in the controlled environment; synthetic data is used |
  | **External impact log** | Every allowed outbound action is recorded for post-hoc audit |

## Acceptance Criteria

Written as verifiable outcomes, derived only from the mandatory rules, thresholds and gates stated in the source design.

- [ ] **AC1** (R1) Scoring criteria, expected end states and reference solutions are absent from every environment, file and prompt the Agent can read, and the scorer code is not visible to the tested party.
- [ ] **AC2** (R1) Ingestion is blocked whenever the pre-ingestion scan of the Agent-accessible scope hits scoring keywords, expected-answer fragments or reference-solution signature strings.
- [ ] **AC3** (R1) A verdict whose scoring evidence contains suspicious content originating from the tested output is marked as "untrustworthy verdict" and listed separately.
- [ ] **AC4** (R2) Scoring runs in a judgement environment independent of the Agent's environment, with separate credentials, and the judgement reads only the snapshot frozen after the execution stage ends.
- [ ] **AC5** (R3) A "tests pass" verdict always carries evidence that the tests actually ran (non-zero case count, execution summary); an empty result never counts as a pass.
- [ ] **AC6** (R3) A result whose exit code is non-zero while its log contains no failure record at all is marked invalid.
- [ ] **AC7** (R3) Test output that does not fall between the specified start and end markers invalidates the entire verdict.
- [ ] **AC8** (R3) Skipped tests never satisfy the "fail-to-pass" condition.
- [ ] **AC9** (R3) An Agent's claim of "done" is treated only as a termination signal; the scorer decides whether the task is complete.
- [ ] **AC10** (R3) An Agent having retrieved information is never treated as completion.
- [ ] **AC11** (R4) The reference solution passes the ingestion gate, so the scorer never kills a correct solution by mistake.
- [ ] **AC12** (R4) Every scorer modification is re-run on the scorer regression set (known-pass and known-fail samples) and takes effect only after passing.
- [ ] **AC13** (R4) A scorer upgrade never changes historical scores, and the report shows the scorer version.
- [ ] **AC14** (R5) Scoring files live in an isolated environment and integrity checks run before and after execution.
- [ ] **AC15** (R5) Every Trial gets its own environment instance and its own credentials, the instance being created fresh or rolled back to the initial snapshot.
- [ ] **AC16** (R6) An "impossible declaration" counts as correct only when the task genuinely contains that branch; otherwise it is recorded as a failure.
- [ ] **AC17** (R6) The report gives an explicit warning for an Agent whose share of impossible declarations is abnormally high.
- [ ] **AC18** (R6) An "impossible-task subset" is provided and is scored separately from other tasks.
- [ ] **AC19** (R7) Every adversarial experiment declares all four elements: user task, attack task, injection vector and defense strategy.
- [ ] **AC20** (R8) Canaries automatically locate which injection points the model can actually see.
- [ ] **AC21** (R8) Defenses plug into the evaluation pipeline as pluggable components and, following the same specification as attacks, change only one variable per experiment.
- [ ] **AC22** (R9) The report gives both the no-attack utility and the ASR.
- [ ] **AC23** (R9) The Utility-ASR trade-off curve is available so that operating points with high utility and low ASR at the same time can be picked.
- [ ] **AC24** (R10) Denial-of-service verdicts are produced on an independent channel with independent metrics, using the "safe = not fully blocked" rule.
- [ ] **AC25** (R11) Ingestion applies the eight checks in [Chapter 02, Section 2.8][feat-002], and a task that fails hard gate items 1–3 (solvability, scorer trustworthiness, scorer non-triviality) does not enter dev.
- [ ] **AC26** (R12) A task passed by ≥ 95% of Agents raises a saturation alert, and a task failing for ≥ 95% of Agents with concentrated failure modes raises a suspicion alert.
- [ ] **AC27** (R12) A reference solution that no longer passes on a re-run causes the task to be downgraded immediately and rolled back to the previous version.
- [ ] **AC28** (R12) A non-deterministic re-run isolates the task and investigates environment randomness, an abnormal runtime investigates an environment wait/timeout trap, and an annotation drift triggers difficulty re-annotation (see [Chapter 02, Section 2.5][feat-002]).
- [ ] **AC29** (R12) Audit conclusions populate the task's "health" field, and the score contributed by a task whose health is "suspicious" is flagged in the report.
- [ ] **AC30** (R13) Every revision to task content bumps the version, retains the old versions, and states its reason and its scope of impact.
- [ ] **AC31** (R13) Scores of a published version are never recomputed because of a revision; when the revision is judged as "the original task was wrong", a voiding notice for the original task's score is published at the same time.
- [ ] **AC32** (R14) A judge whose agreement rate and agreement coefficient are below the threshold is not used for official evaluation; above the threshold it is labeled "calibrated, judge=version X".
- [ ] **AC33** (R15) An unavailable scoring model or an unparseable scoring output is recorded as UNSCORED and never as FAIL, and the samples that failed to parse are recorded.
- [ ] **AC34** (R15) A verdict failure rate above the threshold marks the report of that evaluation run as "scoring unreliable".
- [ ] **AC35** (R16) In dual-track parallel blending, a difference above the threshold sends the case into manual arbitration; the arbitration conclusion becomes the final verdict and the original values of both verdicts are recorded.
- [ ] **AC36** (R17) A task set whose validity period has expired without re-verification is marked "freshness in question", and a contamination probe hit is marked "suspected contamination".
- [ ] **AC37** (R17) The private set rotates its tasks regularly, the public set receives a substantive update at least once a year, and every task set carries an identifiable canary marker string.
- [ ] **AC38** (R18) A platform failure rate > 5% marks the report "insufficient confidence" and excludes it from the leaderboard.
- [ ] **AC39** (R18) A verdict failure rate > 3% marks the report "scoring unreliable".
- [ ] **AC40** (R18) Resume consistency below 99% triggers an idempotency investigation.
- [ ] **AC41** (R18) A benchmark regression pass rate below 100% blocks the release, and a scorer regression pass rate below 100% blocks the scorer release.
- [ ] **AC42** (R18) The platform passes its built-in benchmark regression suite (a minimal task set of ≤ 10 tasks per domain, the scorer regression set and known-answer cases for metric computation, e.g. the examples in [Chapter 04, Section 4.6][feat-004]) on every self-upgrade.
- [ ] **AC43** (R19) Outbound requests are intercepted before leaving the controlled environment and only allowlisted targets are allowed.
- [ ] **AC44** (R19) Irreversible actions (placing orders, sending messages, making payments, deleting) are intercepted by default and are allowed only when the task explicitly declares permission.
- [ ] **AC45** (R19) One-time credentials are generated per Trial and destroyed with the Trial, and no real user data is placed in the controlled environment — synthetic data is used.
- [ ] **AC46** (R19) Request-rate caps are set on external systems, and every allowed outbound action is recorded for post-hoc audit.

## Impact

Not specified in the source design.

## Risks and Mitigations

| Risk | Likelihood | Mitigation |
|---|---|---|
| Not specified in the source design. | - | - |

## Rollout

Not specified in the source design.

## References

- Tech Design: Not yet created.
- Test Report: Not yet created.
- Capability Taxonomy and Task System (Chapter 02): [FEAT-002 Capability Taxonomy and Task System][feat-002]
- Scoring and Metrics (Chapter 04): [FEAT-004 Scoring and Metrics][feat-004]

## Open Questions

| # | Question | Owner | Blocking? |
|---|---|---|---|
| 1 | Which invisibility mechanism is used to hide the scoring logic — encrypted delivery or an isolated judgement environment? (tracked as Q2 in the source design) | Platform team | Yes |

<!-- Link definitions stay at the bottom of the file. Angle brackets are required
     because the linked file names contain no spaces, but the pattern is kept
     consistent with the template. -->

[feat-002]: <../task-system/capability-taxonomy-and-task-system.md>
[feat-004]: <../scoring/scoring-and-metrics.md>
