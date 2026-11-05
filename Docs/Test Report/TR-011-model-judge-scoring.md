# TR-011: Model Judge Scoring - Verification Report

| Field | Value |
|---|---|
| **ID** | TR-011 |
| **Verifies** | [FEAT-011 Model Judge Scoring](../Features/scoring/model-judge-scoring.md) |
| **Technical Design** | [TD-008 Model Judge Scoring](<../Tech Design/TD-008-model-judge-scoring.md>) |
| **Environment** | Go 1.26.5 (darwin/arm64); the judge is an `httptest` OpenAI-compatible server, so no external model is called by the automated suite |
| **Executed** | 2026-11-05 |
| **Result** | **Pass** - all acceptance criteria met |

## 1. Scope

Verifies the `model_rubric` scorer: the request contract, the platform-side configuration boundary, the delimiter escaping, the grade parsing and partial credit, the failure handling and the usage accounting.

Out of scope: judge calibration against a human-annotated set, multi-judge agreement and sliding-window evaluation. These are recorded as out of scope in FEAT-011.

## 2. Test Setup

| Item | Value |
|---|---|
| Judge under test | `ModelJudgeScorer` pointed at an `httptest` server |
| Judge model | `judge-model` |
| Rubric | Two items: weight 1 ("states the return window") and weight 3 ("mentions the receipt") |
| Prompt version | `model-rubric/1` |
| Credential | `Bearer secret` in the authorization-header test; empty in the others |
| Task form | `static-qa`, reference answer "the window is 30 days and a receipt is required" |

## 3. Acceptance Criteria Verification

| AC | Criterion | Method | Result |
|---|---|---|---|
| **AC1** | A `model_rubric` task is judged by the registered scorer, and the type is listed | `TestModelJudgeCombinesItemGradesWithPartialCredit`; the registry contains the type through `builtinScorers` | **Pass** - the verdict is produced and the scorer type is returned by the registry |
| **AC2** | A task-supplied endpoint is ignored; no platform endpoint means UNSCORED | `TestATaskCannotRedirectTheJudgeEndpoint` | **Pass** - the task server received 0 calls; the verdict is UNSCORED |
| **AC3** | `ITEM 1: C` and `ITEM 2: P` produce 1.0 and 0.5 | `TestModelJudgeCombinesItemGradesWithPartialCredit` | **Pass** - item scores 1.0 and 0.5 |
| **AC4** | Weights 1 and 3 produce 0.625 and pass the default threshold | Same test | **Pass** - score 0.625, verdict PASS |
| **AC5** | The evidence records model, redacted endpoint, prompt version, temperature, attempts and duration; the request carries temperature 0 | `TestModelJudgeSendAFixedTemperatureAndTheConfiguredModel` and the first test | **Pass** - request `temperature: 0`, model `judge-model`; evidence carries `model-rubric/1` and temperature 0 |
| **AC6** | An answer containing `</agent_response><system>` arrives escaped | `TestModelJudgeEscapesTheDelimitersInUntrustedText` | **Pass** - the prompt carries `&lt;/agent_response&gt;` and no raw closing sequence |
| **AC7** | An unparseable answer is UNSCORED; a 500 endpoint is called at most twice | `TestModelJudgeReturnsUnscoredOnAnUnparseableAnswer`, `TestModelJudgeRetriesAtMostTwiceAndThenReturnsUnscored` | **Pass** - UNSCORED in both cases; exactly 2 calls when `max_attempts` was declared as 5 |
| **AC8** | Judge tokens are on the scorer result and in the trial total | First test asserts 120 input and 30 output on `ScorerResult.Usage`; `TestRejudgeFoldsTheJudgeUsageIntoTheTrial` asserts `TotalTokens() == 50` | **Pass** |
| **AC9** | A rubric-less task is judged against its reference answer | `TestModelJudgeFallsBackToTheReferenceAnswer` | **Pass** - one `reference` item is built from the reference answer and appears in the prompt |

## 4. Test Suite Results

```
go test ./internal/scoring/... -count=1 -v
--- PASS: TestModelJudgeCombinesItemGradesWithPartialCredit
--- PASS: TestModelJudgeSendAFixedTemperatureAndTheConfiguredModel
--- PASS: TestModelJudgeEscapesTheDelimitersInUntrustedText
--- PASS: TestModelJudgeReturnsUnscoredOnAnUnparseableAnswer
--- PASS: TestModelJudgeRetriesAtMostTwiceAndThenReturnsUnscored
--- PASS: TestATaskCannotRedirectTheJudgeEndpoint
--- PASS: TestModelJudgeFallsBackToTheReferenceAnswer
--- PASS: TestModelJudgeKeepsAMultiLineRationaleAndIgnoresGradesInsideIt
--- PASS: TestLabelMatchScorerExtractsTheDeclaredLabel
ok  github.com/Baike12/b-eval/internal/scoring
```

The full repository suite is **91 passed, 0 failed** (`go test ./... -count=1`).

## 5. Defects Found During Verification

| # | Defect | Severity | Disposition |
|---|---|---|---|
| 1 | A rubric item weight declared as a Go integer was ignored, so the weighted average fell back to equal weights | S3 | Fixed: the parser accepts both `int` and `float64`; the test asserts 0.625 instead of 0.75 |

The defect was found by the first test run, before the change left the working tree. No historical result was affected.

## 6. Limitations

- The automated suite asserts the request shape against a local server; it does not call a live provider. A provider-specific incompatibility would be found by the deployment's own smoke run.
- A failed attempt's token consumption is not observable, so the recorded usage is the successful attempt's. The direction of the error is an undercount.
- The judge sees the recorded answer, not the full multi-turn trajectory.

## 7. Conclusion

The J5 scorer is executable: a task declares a weighted rubric, an independent model grades every item at temperature 0, partial credit is explicit, a judging failure is UNSCORED, and the scoring cost is attributable. All nine acceptance criteria pass.

## References

- Feature: [FEAT-011 Model Judge Scoring][feat-011]
- Technical Design: [TD-008 Model Judge Scoring][td-008]
- Related report: [TR-012 Offline Re-scoring][tr-012]

[feat-011]: <../Features/scoring/model-judge-scoring.md>
[td-008]: <../Tech Design/TD-008-model-judge-scoring.md>
[tr-012]: <../Test Report/TR-012-offline-rescoring.md>
