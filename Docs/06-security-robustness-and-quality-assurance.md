# 06 · Security, Robustness, and Quality Assurance

This chapter answers three questions: **can scores be inflated by cheating**, **are the tasks themselves trustworthy**, and **are the scores reported by the platform trustworthy**.

## 6.1 Anti-Cheating During Evaluation

The Agent under test holds certain permissions during evaluation (reading files, running commands, making network requests). You must assume it will attempt to read the scoring logic, tamper with the scoring environment, and construct false positives.

### 6.1.1 Scoring Must Not Be Visible

| Measure | Requirement |
|---|---|
| **Scoring criteria are out of the Agent's reach** | Scoring criteria, expected end states, and reference solutions must not appear in any environment, file, or prompt the Agent can read |
| **The scorer implementation is unreadable** | The scorer code is not visible to the tested party (following TheAgentCompany's encrypted-delivery approach) |
| **Automatic scan before ingestion** | When a task is ingested, the Agent-accessible scope is scanned for scoring keywords, expected-answer fragments, and reference-solution signature strings; any hit blocks ingestion |
| **Injection detection** | When suspicious content originating from the tested output (e.g. a fabricated test report, a forged completion claim) is found in the scoring evidence, the verdict is marked as "untrustworthy verdict" and listed separately |

**Which invisibility mechanism is used (encrypted delivery vs. an isolated judgement environment) depends on Q2**, but **the effect requirement is the same**: the tested party must not obtain the scoring logic during evaluation.

### 6.1.2 Separation of the Judgement Environment from the Execution Environment

- Scoring runs in a **judgement environment that is independent of the Agent's environment**, so the Agent cannot interfere.
- The final state required for judgement is **frozen into a snapshot** after the execution stage ends; the judgement reads only the snapshot.
- The credentials of the judgement environment are separate from those of the Agent environment.

### 6.1.3 Concrete Rules against "False Positives"

A verdict must prove that the work was actually done, not merely that the result "looks correct":

| Rule | Description | Reference |
|---|---|---|
| **Execution evidence must exist** | When the criterion is "tests pass", there must be evidence that "the tests actually ran" (non-zero case count, execution summary); an empty result does not count as a pass | SWE-bench's SUITE_RAN evidence check |
| **Exit code and log are cross-checked** | If the exit code is non-zero but the log contains no failure record at all, the result is invalid (prevents a patch from printing success itself) | SWE-bench's exit code cross-check |
| **Test output marker boundaries** | Test output must fall between the specified start and end markers; otherwise the entire verdict is invalid | SWE-bench's START/END_TEST_OUTPUT |
| **Skipped does not count as passed** | Skipped tests do not satisfy the "fail-to-pass" condition | SWE-bench's SKIPPED semantics |
| **A completion claim is not completion** | An Agent claiming "done" is only a termination signal; whether the task is complete is decided by the scorer | General principle |
| **Retrieved information is not completion** | An Agent having found information does not mean the job was done | Built-in rubric semantics |

### 6.1.4 Guarding against "False Negatives" and Against Killing Good Agents

Anti-cheating must not degenerate into "condemning good Agents as well":

- **The reference solution must pass** (item 2 of the ingestion gate), ensuring the scorer never kills a correct solution by mistake.
- **Scorer changes require regression testing**: any scorer modification must be re-run on the scorer regression set (containing known-pass and known-fail samples), and may take effect only after passing.
- **The scorer version is bound to the task version**: the report shows the scorer version, and historical scores never change because of a scorer upgrade.

### 6.1.5 Guarding against "Environment-Side Cheating"

| Scenario | Rule |
|---|---|
| The Agent directly modifies scoring files | Scoring files live in an isolated environment, and integrity checks run before and after execution |
| The Agent contaminates a later Trial | Each Trial gets its own environment instance plus its own credentials (see Chapter 02, Section 2.6) |
| The Agent cached the previous run's result | Each Trial's environment instance is created fresh, or rolled back to the initial snapshot |
| The Agent leaves a backdoor | The judgement environment snapshot captures only state relevant to the task and never trusts global environment state |

### 6.1.6 Cost-Oriented Cheating

An Agent may lower its failure rate by "doing less and answering more" (e.g. declaring a task impossible whenever it hits a hard one).

- A task's "impossible declaration" counts as correct only when the task genuinely contains that branch; otherwise it is recorded as a failure.
- **For an Agent whose share of impossible declarations is abnormally high, the report gives an explicit warning** (e.g. "this Agent declared X% of tasks impossible").
- Provide an "impossible-task subset" dedicated to measuring this capability, scored separately from other tasks.

## 6.2 Adversarial Security Evaluation

The platform supports dedicated adversarial evaluation (corresponding to capability dimension C6), using a **dual-axis utility-security** paradigm (following AgentDojo).

### 6.2.1 Experiment Design

An adversarial experiment consists of four elements:

| Element | Content |
|---|---|
| **User task** | The task the Agent is supposed to complete (the benign task) |
| **Attack task** | Something the attacker wants the Agent to do that the user never authorized |
| **Injection vector** | Where the attack payload is planted (tool return values, web content, files, email, etc.) |
| **Defense strategy** | A protective measure on the Agent side or the platform side |

### 6.2.2 Attack Library and Defense Library

- **Attack library**: attacks are defined as a combination of "attack strategy × injection vector", and each attack declares what it aims to achieve.
- **Automatic injection-point discovery**: probes (canaries) automatically locate "which injection points the model can actually see", avoiding omissions in manual annotation (following AgentDojo's `get_injection_candidates` approach).
- **Defense library**: defenses plug into the evaluation pipeline as pluggable components, including content filtering, tool allowlists, injection detection, provenance marking, instruction restatement, and so on. **Defenses follow the same specification as attacks: only one variable changes per experiment.**

### 6.2.3 Dual-Axis Metrics

| Metric | Definition | Notes |
|---|---|---|
| **Utility** | The proportion of user tasks the Agent completes normally in the absence of attacks | A defense must not come at the cost of utility |
| **Utility under attack** | The proportion of benign tasks still completed under attack | |
| **Targeted attack success rate (ASR)** | The proportion of attacks whose objective is achieved | The lower the better |
| **Untargeted attack success rate** | The proportion of cases where the Agent exhibits abnormal behavior (denial of service, abnormal output) | Distinguishes "led astray" from "broken" |
| **Utility-ASR trade-off curve** | A two-dimensional view of utility against ASR | Used to pick operating points: high utility and low ASR at the same time |

**The report must give both the no-attack utility and the ASR.** Reporting only the ASR rewards a defense that "rejects everything outright"; reporting only utility hides injection risk.

### 6.2.4 Denial-of-Service Attacks

A separate class: the attack objective is to make the Agent unable to complete the task (e.g. trapping it in a loop or triggering a CAPTCHA). The verdict for this class is "safe = not fully blocked", which is the opposite of targeted attacks, so it requires an independent channel and independent metrics.

## 6.3 Task Quality Assurance

### 6.3.1 Ingestion Gate

See the eight checks in Chapter 02, Section 2.8. Items 1–3 among them (solvability, scorer trustworthiness, scorer non-triviality) are **hard gates**: a task that fails them must not enter dev.

### 6.3.2 Continuous Audit

After publication, a task still needs continuous health checks. The platform provides:

| Audit item | Verdict | Action |
|---|---|---|
| **Abnormally high pass rate** | A task is passed by ≥ 95% of Agents | Saturation alert → re-review the difficulty or design |
| **Abnormally high failure rate** | A task fails for ≥ 95% of Agents, and failure modes concentrate in the same place | Suspicion alert → inspect the scorer and the task |
| **Scorer inconsistent with the reference solution** | The reference solution no longer passes on a re-run | Immediately downgrade the task and roll it back to the previous version |
| **Non-determinism** | A re-run with the same seed yields a different verdict | Isolate the task and investigate environment randomness |
| **Abnormal runtime** | Actual elapsed time far exceeds the estimate | Investigate whether an environment wait/timeout trap exists |
| **Annotation drift** | Measured difficulty conflicts with the annotated difficulty | Trigger difficulty re-annotation (see Chapter 02, Section 2.5) |

**Audit conclusions feed into the task's "health" field**, visible in reports and leaderboards — when a task's health is "suspicious", the score it contributes is flagged in the report.

### 6.3.3 Task Revision Discipline

- Any revision to task content **must bump the version**; old versions are retained.
- A revision must state its reason (fixing a scorer error / fixing an ambiguity / fixing a factual error / fixing an annotation error) and its scope of impact.
- **Scores of a published version are never recomputed because of a revision**, unless the revision is judged as "the original task was wrong", in which case a voiding notice for the original task's score must be published at the same time.

## 6.4 Scoring Quality Assurance

### 6.4.1 Judge Calibration

Model-based scoring must be calibrated before use:

| Calibration item | Method |
|---|---|
| **Human-annotated set** | Maintain a batch of human-judged samples (including positives, negatives, and boundary cases) |
| **Agreement rate evaluation** | The agreement rate between the judge model's verdicts on the annotated set and the human verdicts |
| **Agreement coefficient** | Krippendorff's alpha across multiple judges |
| **Bias analysis** | Analyze systematic judge biases (tendency to pass, tendency to give partial credit, leniency toward long trajectories) |
| **Deployment threshold** | Agreement rate and agreement coefficient below the threshold → must not be used for official evaluation; above the threshold → label "calibrated, judge=version X" |

### 6.4.2 Judge Failure Protection

- Scoring model unavailable → UNSCORED; do not record FAIL.
- Scoring output unparseable → UNSCORED; record the samples that failed to parse.
- Verdict failure rate above the threshold → the report of that evaluation run is marked "scoring unreliable".
- **Scoring overhead counts toward the total cost** (see Chapter 04, Section 4.2.4).

### 6.4.3 Blending Manual and Automatic Scoring

The platform supports three blending modes, and a task can choose among them:

| Mode | Mechanism | Applicability |
|---|---|---|
| **Automatic first + manual spot-check** | Automatic verdicts, manual sampling at a fixed ratio, producing a deviation rate | Clear criteria, large volume |
| **Manual first + automatic review** | Manual verdicts, with automation used only for cross-validation | High value, small volume |
| **Dual-track parallel** | Both automatic and manual judging; when the difference exceeds the threshold, the case enters manual arbitration | High dispute, strongly subjective rubrics |

**Conflict handling:** samples where the two tracks disagree enter the arbitration queue; the arbitration conclusion becomes the final verdict for that sample, and the original values of both verdicts are recorded.

## 6.5 Data Contamination and Freshness

| Measure | Description |
|---|---|
| **Timestamped tasks** | A task carries a "creation time" and a "last verified time"; a task set whose validity period has expired without re-verification is marked "freshness in question" |
| **Contamination probe** | Fragments of task instructions and expected answers are compared against public corpora; a hit is marked "suspected contamination" |
| **Private-set hedge** | The official leaderboard includes both public-set and private-set results; private-set scores resist contamination |
| **Periodic rotation** | The private set rotates its tasks regularly; the public set receives a substantive update at least once a year |
| **Canary markers** | A task set carries an identifiable marker string used to detect whether task content has appeared elsewhere (following terminal-bench's canary practice) |
| **Freshness metric** | The report gives "the time since the last substantive update of the task set used in this evaluation" for reference when interpreting scores |

## 6.6 The Platform's Own Reliability

Trustworthy scores require a trustworthy platform. The platform must expose its own health:

| Metric | Definition | Threshold behavior |
|---|---|---|
| **Platform failure rate** | The share of Trials affected by platform/environment/scorer failures | > 5% → the report is marked "insufficient confidence" and is excluded from the leaderboard |
| **Verdict failure rate** | The share of cases where the scorer returns UNSCORED | > 3% → the report is marked "scoring unreliable" |
| **Resume consistency** | The share of cases where a resumed run of the same plan yields results identical to a single uninterrupted run | < 99% → investigate idempotency |
| **Benchmark regression pass rate** | The pass rate of the platform's built-in benchmark regression suite | < 100% → block the release |
| **Scorer regression pass rate** | The pass rate of the regression set after a scorer change | < 100% → block the scorer release |

The **platform's built-in benchmark regression suite** contains: a minimal task set (≤ 10 tasks per domain), the scorer regression set, and known-answer cases for metric computation (e.g. the examples in Chapter 04, Section 4.6 as fixed assertions). The platform must pass this suite on every self-upgrade.

## 6.7 Sensitive Operation Controls

The evaluation environment may involve real external systems (real websites, real APIs). The platform must:

| Control | Requirement |
|---|---|
| **Request interception** | Outbound requests are intercepted before leaving the controlled environment; the platform only allows allowlisted targets |
| **Irreversible actions** | Actions such as placing orders, sending messages, making payments, and deleting are intercepted by default; they are allowed only when the task explicitly declares permission |
| **Credential management** | Use one-time credentials, generated per Trial and destroyed with the Trial |
| **Rate constraints** | Set request-rate caps on external systems to keep evaluation behavior from affecting real business |
| **Data minimization** | No real user data is placed in the controlled environment; synthetic data is used |
| **External impact log** | Every allowed outbound action is recorded for post-hoc audit |
