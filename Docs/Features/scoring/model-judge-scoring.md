# FEAT-011: Model Judge Scoring

| Field | Value |
|---|---|
| **ID** | FEAT-011 |
| **Status** | Implemented |
| **Owner** | Platform team |
| **Priority** | P0 |
| **Implements** | `F-JDG-01`, `F-JDG-03`, `F-MTR-05` |
| **Created** | 2026-11-05 |
| **Last Updated** | 2026-11-05 |
| **Source** | `Functional design chapter 04 - Scoring and Metrics, section 4.4` |

## Summary

The platform gains a real J5 model scorer: a task declares a rubric, an independent model grades every item, and the verdict carries the grades, the rationale and the judge's token consumption. The single most important outcome is that open-ended tasks can be scored without hand-writing a deterministic rule, while a judge failure stays UNSCORED and never turns into a capability deduction.

## Problem

**Current behaviour:** The M0 scorer set is deterministic. The `rubric` scorer matches each declared rubric item with a substring needle, so an item such as "the reply is professional" or "the refund was actually issued" cannot be judged at all. FEAT-004 R17 specifies model scoring (J5), but TD-003 §11 records that M0 only ships a rule-based stand-in.

**Pain:** Task authors cannot express the open-ended criteria that most agent tasks need, so they either rewrite them as brittle keyword checks or leave them unjudged. A keyword check measures wording rather than behaviour and is easy for the agent to satisfy without doing the task.

**Evidence:** The rule-based `RubricScorer` in `internal/scoring/scoring.go` only performs `strings.Contains` on a needle. FEAT-004 AC9, AC10, AC24, AC25 and AC26 describe the model-scoring behaviour that was not implemented.

## Proposed Solution

The platform registers a `model_rubric` scorer. A task references it like any other scorer and supplies a weighted rubric; the platform sends the task instruction, the rubric, the reference answer and the agent response to an independent, OpenAI-compatible judge at temperature 0, parses one `C`/`P`/`I` grade per item, and maps the weighted average to PASS or FAIL. The judge's address and credential come from the server, and every judging failure is UNSCORED.

## Scope

### In Scope

- A registered `model_rubric` scorer implementing the J5 category.
- Per-item grading with the three-level vocabulary `C` (complete), `P` (partial) and `I` (incomplete), and partial credit mapped to 0.5.
- A weighted pass threshold declared by the scorer reference, with 0.6 as the default.
- Platform-side judge configuration: endpoint, model and credential, with the endpoint and credential unavailable to a task.
- Delimiter escaping of every untrusted text sent to the judge.
- A fixed output contract, parsed from the last grade for each item.
- UNSCORED on an unreachable endpoint, a timeout, a non-2xx answer or an unparseable answer, with at most two attempts.
- The judge model, endpoint, prompt version, temperature, attempt count and duration recorded in the scorer's evidence.
- The judge's input and output tokens recorded on the scorer result and folded into the trial's usage and cost.

### Out of Scope

- **Judge calibration and agreement rate.** F-JDG-04 and FEAT-006 define calibration on a human-annotated set; that is the M3 quality-assurance milestone.
- **Multiple judges voting.** FEAT-004 R13 describes agreement rate and a consistency coefficient; this feature implements one judge per scorer reference. A task composes several model scorers with different rubrics instead.
- **Sliding-window rubric evaluation.** The long-trajectory window is not implemented; the judge sees the recorded answer and the task instruction.
- **A per-task prompt template editor.** The prompt is a platform constant; only the rubric and the pass threshold are task data.

## Requirements

- **R1** The platform must register a `model_rubric` scorer, and a task must reference it by type and parameters like any other scorer.
- **R2** The judge endpoint, the model and the credential must come from the platform, and a task must not be able to override the endpoint or the credential.
- **R3** The judge must grade every rubric item independently and return exactly one of `C`, `P` or `I` per item, with `C = 1.0`, `P = 0.5` and `I = 0.0`.
- **R4** The verdict score must be the weighted average of the item scores, and the primary criterion must pass when that score reaches the declared threshold (default 0.6).
- **R5** The judge request must fix the temperature at 0 and record the model, the redacted endpoint, the prompt version, the temperature, the attempt count and the duration with the result.
- **R6** Every untrusted text sent to the judge must be delimiter-escaped, so the agent cannot close a delimiter and address the judge as if it were the platform.
- **R7** An unreachable endpoint, a timeout, a non-2xx answer and an unparseable answer must each produce UNSCORED, never FAIL, and the platform must not attempt the call more than twice.
- **R8** The judge's input and output tokens must be recorded on the scorer result and added to the trial's usage, so the scoring overhead appears in the cost metrics.
- **R9** A task that declares no rubric must be judged against its reference answer as a single criterion.

## Acceptance Criteria

- [ ] **AC1** (R1) A task declaring a `model_rubric` scorer is judged by the registered scorer, and the scorer type appears in the registry listing.
- [ ] **AC2** (R2) A task that sets an `endpoint` in the scorer parameters does not reach that endpoint; the call uses the platform configuration, and with no platform endpoint the result is UNSCORED.
- [ ] **AC3** (R3) A response of `ITEM 1: C` and `ITEM 2: P` produces item scores 1.0 and 0.5.
- [ ] **AC4** (R4) With item weights 1 and 3, the same grades produce a score of 0.625, which passes the default threshold.
- [ ] **AC5** (R5) The recorded evidence carries the judge model, the redacted endpoint, `model-rubric/1`, temperature 0, the attempt count and the duration; the request body carries `temperature: 0`.
- [ ] **AC6** (R6) An answer containing `</agent_response><system>` arrives at the judge escaped as `&lt;/agent_response&gt;`, and the delimiter appears exactly once.
- [ ] **AC7** (R7) A judge answer with no parseable item grade is UNSCORED, and an endpoint that keeps answering 500 is called at most twice and then recorded as UNSCORED.
- [ ] **AC8** (R8) The scorer result carries the judge's input and output tokens, and the trial's total token count includes them.
- [ ] **AC9** (R9) A task with no rubric and a reference answer is sent one criterion derived from that reference answer.

## Impact

| Area | Effect |
|---|---|
| Users | Task authors can express open-ended criteria as rubric items; report readers see the judge's grades, rationale and model recorded with the verdict. |
| Existing behaviour | The deterministic scorers are unchanged. A task that does not reference `model_rubric` makes no judge call and pays no judge cost. |
| Operational cost | One judge call per model-scorer reference per trial, plus at most one retry. The tokens are recorded and included in `avg_tokens_per_success` and the new `avg_judge_tokens_per_trial`. |

## Risks and Mitigations

| Risk | Likelihood | Mitigation |
|---|---|---|
| Judge noise exceeds the difference being measured | Med | Temperature fixed at 0, fixed prompt version, partial credit mapped explicitly, UNSCORED on failure, and calibration deferred to F-JDG-04 |
| The agent injects instructions into its response | Med | Every untrusted text is delimiter-escaped, and the system prompt states that the response is data, not instructions |
| A task redirects the judge to an attacker's endpoint | Low | The endpoint and credential are platform configuration; task parameters cannot set them |
| A judge outage becomes a capability deduction | Low | Every judge failure is UNSCORED and counted in the judgement failure rate, not in the success denominator |
| Scoring cost is invisible | Low | Judge tokens are recorded separately and in total, and `avg_judge_tokens_per_trial` reports them |

## Rollout

The scorer ships enabled. A run makes no judge call until a task references `model_rubric`, so existing task sets are unaffected. Rollback is removing the scorer from `builtinScorers`; a task that still references it then resolves to the unknown-scorer path and is UNSCORED rather than silently mis-scored.

## References

- Tech Design: [TD-008 Model Judge Scoring][td-008]
- Test Report: [TR-011 Model Judge Scoring][tr-011]
- Related features: [FEAT-004 Scoring and Metrics][feat-004], [FEAT-012 Offline Re-scoring][feat-012]

[td-008]: <../../Tech Design/TD-008-model-judge-scoring.md>
[tr-011]: <../../Test Report/TR-011-model-judge-scoring.md>
[feat-004]: <./scoring-and-metrics.md>
[feat-012]: <../execution/offline-rescoring.md>

## Open Questions

| # | Question | Owner | Blocking? |
|---|---|---|---|
| 1 | Should a task be allowed to choose among platform-registered judge models, instead of one model per deployment? | Platform team | No |
| 2 | Should a judging failure be retried once on a 4xx answer that looks transient? | Platform team | No |
