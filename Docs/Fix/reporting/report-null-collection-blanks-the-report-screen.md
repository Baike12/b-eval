# FIX-003: The run report screen rendered blank when a report carried an empty collection

| Field | Value |
|---|---|
| **ID** | FIX-003 |
| **Status** | Fixed |
| **Owner** | Platform team |
| **Severity** | S1 Critical |
| **Priority** | P0 |
| **Affects** | `F-RPT-01`, `F-RPT-02`, `F-MTR-01` |
| **Detected in** | Deployment `20261022-ui`, run `run-plan-demo-demo-good-3408113d`, 2026-10-22 |
| **Created** | 2026-10-22 |
| **Last Updated** | 2026-10-22 |

## Summary

The report builder returned Go nil slices for the collections it had nothing to fill, and the JSON encoder wrote them as `null`. The web client read one of those `null` values as a list, so the run report screen threw during render and the page stayed empty. Every report of a run in which no trial failed was unopenable.

## Severity and Impact

**Technical impact:** `encoding/json` writes a nil slice as `null` and an empty slice as `[]`. The report builders created their slices with `append` and never initialised them, so a collection with no members was `null` in the response. The client reads `failed_trials.length`, `scorers` and `cells` as arrays, and reading a length off `null` throws `TypeError: Cannot read properties of null (reading 'length')`. The exception escaped the render, React unmounted the whole tree, and the document body was empty: no navigation, no report, no message.

**Blast radius:** Every report whose run had no failed trial, and every report whose run had a trial without a capability dimension. On the deployment `20261022-ui` the single stored run `run-plan-demo-demo-good-3408113d` passed all 8 of its trials, so 1 of 1 published reports — 100% — was unopenable. The report API itself was unaffected: the numbers, the interval and the breakdown were all correct in the response, and only the presentation of an empty collection was wrong. No published score was wrong.

**Detection:** Found during the visual verification of the deployed interface, when the report screen of the stored run was checked in the browser and rendered an empty document.

## Symptom

Opening `http://<host>/eval/runs/run-plan-demo-demo-good-3408113d` produced a page with no content. The console held:

```
TypeError: Cannot read properties of null (reading 'length')
    at Hp (http://<host>/eval/assets/index-DpkwXGo_.js:60:44908)
```

The response body of `GET /api/runs/run-plan-demo-demo-good-3408113d/report` carried `"failed_trials": null` while every other collection in the same document was a list.

## Reproduction

**Environment:** Deployment `20261022-ui` on the cloud host, Go 1.24, run `run-plan-demo-demo-good-3408113d` (8 trials, 8 passed, 0 failed).

**Preconditions:** A run in which every trial passed, so the failure list is empty.

**Steps:**

1. Open the run report screen of a run in which no trial failed.
2. Read the console.
3. Request the report from the API and inspect the `failures` object.

**Expected:** The report screen renders the report, and the failure section states that no trial failed. The API returns `"failed_trials": []`.

**Actual:** The document body is empty. The API returns `"failed_trials": null`.

## Root Cause

**Cause:** Three builders in `internal/report/report.go` filled their slices with `append` and returned them without initialising them: `buildFailures` for `FailedTrials` and `Scorers`, and `buildBreakdown` for the cells of a dimension that no trial populated. `Build` did the same for `Metrics`. A nil slice in Go is a valid empty slice, so the builders were correct Go; the defect is that the JSON contract they produce was not the contract the client was written against. The TypeScript definitions declare `failed_trials: FailedTrial[]` and `cells: BreakdownCell[]`, so the client read them as arrays without a guard.

**Failure class:** Data

**Why it was not caught:** The report was always produced by a run that had at least one failed trial, so the nil slice never appeared in the evidence set the tests used. The frontend fixture in `web/src/routes/RunReport.test.tsx` listed one failed trial, so the empty case was not in the suite either. Nothing asserted the shape of the response, so the contract between the two sides was never checked against itself.

## Fix

**Change:** The report builders now initialise every collection they fill, so a collection with no members is an empty list rather than a nil slice. `buildFailures` starts with `FailedTrials` and `Scorers` as empty lists, `buildBreakdown` starts each dimension with empty cells, and `Build` starts with empty metrics. `internal/metrics/custom.go` does the same for the required parameters of a custom metric kind, so a kind that requires nothing reports an empty list.

**Rationale:** The client was written against a contract that says these fields are lists, and the server was producing `null` for them. Fixing the server restores the contract instead of teaching every client to defend against a shape the server does not intend to publish. The client-side change that accompanies it does not relax the contract: it is a boundary that keeps one bad payload from blanking the whole interface.

**Guardrail added:** `TestReportContainsNoNullValue` in `internal/report/report_test.go` builds three reports and asserts that the marshalled document holds no null value anywhere, so a collection that regresses to nil fails the gate. The frontend gained the case that was missing: `RunReport.test.tsx` renders a report whose failure list is empty.

## Verification

| Check | Method | Result |
|---|---|---|
| Original reproduction no longer reproduces | Opened the report screen of the stored run on the deployment | **Pass** — the report renders, and the failure section states that no trial failed |
| Regression test added | `internal/report/report_test.go` → `TestReportContainsNoNullValue` | **Pass** — fails on the three null sites without the fix, passes with it |
| Adjacent behaviour unaffected | `go test ./...`, `npm run test`, and the full gate | **Pass** — 18 frontend tests and the whole Go suite pass; every collection in a built report is a list |
| Blast radius re-scored | The report of the affected run built by the code before the fix and by the code after it, compared line by line | **Pass** — one line differs: `"failed_trials": null` becomes `"failed_trials": []`. Every other value of the report, the score, the interval, the confidence section, the metrics, the breakdown and the task rows, is byte for byte the same |

## Regression Prevention

- **Test added:** `TestReportContainsNoNullValue`, in `internal/report/report_test.go`, asserts that no value of a marshalled report is null. The test covers a passing trial with a scorer, a trial with neither a scorer nor a capability dimension, and a run with no trial at all.
- **Run gate:** The test runs in the `go-test` gate of `scripts/gate.sh`, which every change must pass.
- **Detection change:** The web client now renders a render failure as a named fault inside the page frame instead of an empty document, so a payload that breaks the contract is visible to the reader and to the console rather than silent.

## References

- Test Report: [TR-006 Report Collection Contract][tr-006]
- Related Feature: [FEAT-005 Reporting, Analysis and Comparison][feat-005]
- Related Feature: [FEAT-004 Scoring and Metrics][feat-004]

[tr-006]: <../../Test Report/TR-006-report-collection-contract.md>
[feat-005]: <../../Features/reporting/reporting-analysis-and-comparison.md>
[feat-004]: <../../Features/scoring/scoring-and-metrics.md>
