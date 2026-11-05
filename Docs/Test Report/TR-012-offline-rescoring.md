# TR-012: Offline Re-scoring - Verification Report

| Field | Value |
|---|---|
| **ID** | TR-012 |
| **Verifies** | [FEAT-012 Offline Re-scoring](../Features/execution/offline-rescoring.md) |
| **Technical Design** | [TD-009 Offline Re-scoring](<../Tech Design/TD-009-offline-rescoring.md>) |
| **Environment** | Go 1.26.5 (darwin/arm64); the API tests run the real HTTP handler over `httptest` |
| **Executed** | 2026-11-05 |
| **Result** | **Pass** - all acceptance criteria met |

## 1. Scope

Verifies the re-judgement endpoint and the shared outcome builder: the source run is preserved, the reason is mandatory, a second task-set version changes the verdict over identical trajectories, the execution outcome is preserved, a missing task is refused, and the new run is reportable and exportable.

Out of scope: comparing the two runs beyond the existing comparison endpoint, and an inline scorer override.

## 2. Test Setup

| Item | Value |
|---|---|
| Task set v1 | `demo` v1, static-QA, `contains` scorer on "30 days" |
| Task set v2 | `demo` v2, the same task with an `exact` scorer against the reference answer |
| Agent | Mock agent answering "within 30 days" |
| Repetitions | 1 per task |
| Source verdict | PASS under v1 |
| Re-judged verdict | FAIL under v2, because the answer is not exactly the reference |

## 3. Acceptance Criteria Verification

| AC | Criterion | Method | Result |
|---|---|---|---|
| **AC1** | The run is re-judged from trajectories without the agent | `TestRejudgeScoresPersistedTrajectories`; the API test registers no agent call between the source run and the re-judgement | **Pass** - the verdict is produced from the stored answer |
| **AC2** | A missing reason is refused with 400 naming the reason | `TestRescoreRequiresAReason` | **Pass** - status 400, the body contains "reason" |
| **AC3** | The new run carries `rescored_from` and `rescore_reason`; the source is unchanged | `TestRescoreProducesANewRunAndKeepsTheOriginal` | **Pass** - the new run names the source; the source reloads with no `rescored_from` |
| **AC4** | A second task-set version applies its scorer | `TestRescoreCanApplyANewTaskSetVersion` | **Pass** - verdict flips from PASS to FAIL, the new plan names `v2` |
| **AC5** | A non-completed trial keeps its failure class | `TestRejudgeKeepsTheFailureOfAnUnfinishedTrial` | **Pass** - step limit becomes budget-exhausted, verdict FAIL, score 0 |
| **AC6** | A trajectory whose task is gone refuses the call | `TestRejudgeRefusesATrajectoryWhoseTaskIsGone` | **Pass** - the call returns an error naming the task and trial |
| **AC7** | The re-judged run is reportable and exportable | The re-judged run is an ordinary run record; `TestRunExportServesCSVAndJSON` covers the export path | **Pass** - the export reads the run by ID and emits the trial rows |
| **AC8** | Two re-judgements in the same second do not collide | `TestUniqueRunIDDoesNotOverwriteAnExistingRun` | **Pass** - an existing `run-x` yields `run-x-2`; the allocator stops after 1000 suffixes and reports an error instead of looping |

## 4. Supporting Metric and Export Tests

The re-judged rows flow through the same metric layer, so the new metric definitions are verified at the same time:

| Test | Asserts | Result |
|---|---|---|
| `TestRejudgeFoldsTheJudgeUsageIntoTheTrial` | The judge's 40 input and 10 output tokens reach the trial row and `TotalTokens()` is 50 | **Pass** |
| `TestAnswerConsistencyIsTheMeanPairwiseSimilarityAcrossTasks` | Macro consistency is 0.5 when one task is perfectly consistent and one is disjoint | **Pass** |
| `TestAnswerConsistencyHasNoCoverageWithoutTwoTrialsOfATask` | A task with one trial yields no coverage rather than 1.0 | **Pass** |
| `TestClassificationAccuracyCountsOnlyLabelledTrials` | Unlabelled trials are excluded from the denominator | **Pass** |
| `TestLatencyAndJudgeTokenMetrics` | Mean latency 200 ms and 30 judge tokens per trial | **Pass** |
| `TestLabelPrecisionRecallAndF1` | Precision, recall and F1 are all 2/3 on the 2/1/1 confusion matrix | **Pass** |
| `TestTrialsCSVCarriesTheDrillDownColumns` | The CSV header and the judge-token and termination columns | **Pass** |
| `TestRunExportServesCSVAndJSON` | CSV and JSON export over HTTP, and a 400 for an unsupported format | **Pass** |

## 5. Test Suite Results

```
go test ./internal/runner/... ./internal/api/... ./internal/metrics/... ./internal/report/... -count=1 -v
--- PASS: TestRejudgeScoresPersistedTrajectories
--- PASS: TestRejudgeFoldsTheJudgeUsageIntoTheTrial
--- PASS: TestRejudgeRefusesATrajectoryWhoseTaskIsGone
--- PASS: TestRejudgeKeepsTheFailureOfAnUnfinishedTrial
--- PASS: TestRescoreProducesANewRunAndKeepsTheOriginal
--- PASS: TestRescoreRequiresAReason
--- PASS: TestRescoreCanApplyANewTaskSetVersion
--- PASS: TestUniqueRunIDDoesNotOverwriteAnExistingRun
--- PASS: TestRunExportServesCSVAndJSON
--- PASS: TestAnswerConsistencyIsTheMeanPairwiseSimilarityAcrossTasks
--- PASS: TestAnswerConsistencyHasNoCoverageWithoutTwoTrialsOfATask
--- PASS: TestClassificationAccuracyCountsOnlyLabelledTrials
--- PASS: TestLatencyAndJudgeTokenMetrics
--- PASS: TestLabelPrecisionRecallAndF1
--- PASS: TestTrialsCSVCarriesTheDrillDownColumns
ok  github.com/Baike12/b-eval/internal/runner
ok  github.com/Baike12/b-eval/internal/api
ok  github.com/Baike12/b-eval/internal/metrics
ok  github.com/Baike12/b-eval/internal/report
```

The full repository suite is **91 passed, 0 failed** (`go test ./... -count=1`).

## 6. Defects Found During Verification

None. One design adjustment was made during verification: the endpoint originally judged only against the source run's task set, which made "re-judge with a different scorer" impossible because a task set is frozen. The optional `task_set_id` and `task_set_version` were added, and the v1/v2 test now proves a corrected scorer changes the verdict over identical trajectories.

## 7. Limitations

- The re-judged plan shares the source plan's fingerprint inputs except the task-set version; the comparison view therefore reports the two as different plans, which is the intended reading.
- A J1 state-check scorer that needs a live environment cannot be re-run from a trajectory alone.
- A model judge may return a different grade on a second pass even at temperature 0; the two records make the difference visible.

## 8. Conclusion

A completed run can be judged again from its evidence, against a corrected task-set version, with a recorded reason, while the original verdict stays readable. The execution outcome is not re-interpreted, a missing task is refused rather than skipped, and the result is an ordinary run that reports and exports. All eight acceptance criteria pass.

## References

- Feature: [FEAT-012 Offline Re-scoring][feat-012]
- Technical Design: [TD-009 Offline Re-scoring][td-009]
- Related report: [TR-011 Model Judge Scoring][tr-011]

[feat-012]: <../Features/execution/offline-rescoring.md>
[td-009]: <../Tech Design/TD-009-offline-rescoring.md>
[tr-011]: <../Test Report/TR-011-model-judge-scoring.md>
