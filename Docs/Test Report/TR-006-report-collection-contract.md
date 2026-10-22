# TR-006: Report Collection Contract

| Field | Value |
|---|---|
| **ID** | TR-006 |
| **Status** | Verified |
| **Verifies** | [FIX-003 Report Null Collection Blanks the Report Screen][fix-003] |
| **Owner** | Platform team |
| **Created** | 2026-10-22 |
| **Last Updated** | 2026-10-22 |

> This report records what was checked, how it was checked and what the check returned. The defect it verifies is specified in [FIX-003][fix-003].

## 1. Scope

The verification covers the contract between the report API and the web client: the collections the server publishes, the key names it publishes them under, and the behaviour of the client when a payload breaks that contract.

The defect had three faces. The report published `null` where it promised a list, the metric registry published its custom kinds under keys the client did not read, and the client had no way to report a broken payload except by going blank. All three are covered here.

## 2. Method

| Aspect | How it was checked |
|---|---|
| The shape of a report | A Go test that marshals three reports and walks the JSON for null values |
| The shape of the API | A sweep that compares every field the TypeScript declarations promise against every key the API actually returns, including the nested `Task`, `TaskSet`, `Agent`, `Run` and `Report` objects |
| The key names of the custom kinds | An HTTP-level test against `GET /api/metrics` |
| The client | The frontend test suite, and the Chromium instance attached to BrowserOS neo reading the live DOM |
| The deployment | The response of the running platform on the host, and the seven screens opened in the browser |

The sweep was run against a local server holding the two preview reports, so that a report with no failed trial and a report with three metrics that have no coverage were both in the set.

## 3. The Evidence Set

| Report | Content | Why it was used |
|---|---|---|
| `run-plan-demo-demo-good-3408113d` | 8 trials, 8 passed, score 1.0000 | A run with no failed trial: the failure list is empty, so this is the report that the defect made unopenable |
| `run-plan-partial-partial-agent-7f4b3d8d` | 8 trials, 0 passed, score 0.0000, three metrics with no coverage | A run whose agent solved nothing, and a report whose readings sit at the two ends of the scale |

## 4. Checks

### 4.1 The Server Publishes Lists

| # | Check | Expected | Measured | Result |
|---|---|---|---|---|
| 1 | The failure list of a run with no failed trial is a list | `[]` | `"failed_trials": []` | Pass |
| 2 | No value of a marshalled report is null | no null anywhere | 0 null values across the whole document | Pass |
| 3 | A report with a trial that has no capability dimension publishes cells | `[]` | `"cells": []` for that dimension | Pass |
| 4 | A report with no trial publishes its collections | `[]` | metrics, cells, scorers and failed trials all `[]` | Pass |
| 5 | The custom kinds publish their required parameters as a list | `[]` for a kind that requires nothing | `"required_params": []` for `exact_reference_match` and `step_budget_respected` | Pass |
| 6 | The fix changes nothing but the empty collections | one line of difference | the report of the affected run, built by the code before the fix and by the code after it, differs in one line: `"failed_trials": null` becomes `"failed_trials": []` | Pass |

The Go test `TestReportContainsNoNullValue` covers checks 1 to 5. Run against the code before the fix, it reports 10 null values across its three cases: `$.failures.failed_trials` and `$.failures.scorers` in every case, and `$.breakdown[n].cells` for each dimension that no trial populated. Run against the fixed code, it passes.

Check 6 is a line-by-line comparison of the whole report of the run on the deployment, built twice by the same builder from the same stored run: once with the code before the fix and once with the code after it. The score, the interval, the confidence section, all 14 metric rows, every breakdown cell and every task row are identical.

### 4.2 The Server and the Client Agree

| # | Check | Expected | Measured | Result |
|---|---|---|---|---|
| 7 | Every field the TypeScript declarations promise is present in the response | no missing key | 0 missing keys across 20 interfaces and 8 endpoint responses | Pass |
| 8 | No field the declarations promise is null in the response | no null in a non-nullable field | 0 nulls | Pass |
| 9 | The custom kinds are published under the keys the client reads | `name`, `description`, `required_params` | the same three keys | Pass |
| 10 | The run list and the report declare their plans separately | the run's plan has no fingerprint | `RunPlan` and `PlanInfo` are two interfaces | Pass |

Check 7 and 8 are a sweep over `GET /api/tasksets`, `GET /api/agents`, `GET /api/runs`, `GET /api/metrics`, the detail of both task sets and the report of both runs, with the nested `Task`, `TaskSet`, `Agent`, `Run`, `Report`, `MetricRow`, `TaskResult`, `Breakdown` and `BreakdownCell` objects descended into.

Check 9 was false before the fix: the Go structure carried no JSON tags, so the keys were `Name`, `Description` and `RequiredParams`, and the client read `name`, which arrived as `undefined`.

### 4.3 The Client

| # | Check | Expected | Measured | Result |
|---|---|---|---|---|
| 11 | The report screen of a run with no failed trial renders | the report, with the failure section saying so | 6 sections, and the text "No trial failed." | Pass |
| 12 | A report whose failure list is empty renders in the test suite | the empty state | the case "renders a run in which no trial failed" passes | Pass |
| 13 | A render failure is named instead of leaving a blank document | a heading and the fault | the case "names the fault instead of leaving a blank screen" passes | Pass |
| 14 | The computation dropdown of the metric registration form carries a value on every option | 4 options with 4 values | 4 options, 0 with the value `undefined` | Pass |

Check 14 was false before the fix: every option of the dropdown had the value `undefined`, so no custom metric could be registered from the interface.

### 4.4 The Repository Gates

| # | Check | Result |
|---|---|---|
| 15 | `web-typecheck` | passed |
| 16 | `web-test` | passed, 18 tests in 5 files |
| 17 | `web-build` | passed |
| 18 | `go-fmt`, `go-vet`, `go-test`, `go-build` | passed |
| 19 | `docs-english`, `docs-links`, `docs-index` | passed |
| 20 | `e2e-smoke` | passed |

## 5. Defects Found and Fixed

| # | Defect | Cause | Fix |
|---|---|---|---|
| 1 | The report screen was blank for every run with no failed trial | The builders filled their slices with `append` and returned them uninitialised, and the JSON encoder writes a nil slice as `null` | Every collection the builders fill is initialised |
| 2 | A breakdown dimension with no trial published `cells: null` | The same pattern, in `buildBreakdown` | The cells of a dimension start as an empty list |
| 3 | A custom kind published its keys in PascalCase | The Go structure carried no JSON tags | The structure carries the tags the client reads |
| 4 | A custom kind that requires nothing published `required_params: null` | The field was left uninitialised | The field is an empty list |
| 5 | The client went blank on a payload that broke the contract | Nothing caught a render failure | The screen is wrapped in a boundary that names the fault |

## 6. Result

The defect is verified as fixed by [FIX-003][fix-003]. Five defects were found during the verification and all five were fixed. The 20 checks above pass, and the report screen of the run that was unopenable now renders.

## 7. What This Report Does Not Claim

- It does not claim that the sweep covers every field the API will ever publish. It covers the fields the client declares today, which is the set that can break the interface.
- It does not verify the platform's numbers. The report builder is covered by the Go test suite and by the end-to-end smoke gate, both of which pass.
- It does not claim that a broken payload is reported well. It claims only that it is reported at all, and that the reader is not left with a blank document.

[fix-003]: <../Fix/reporting/report-null-collection-blanks-the-report-screen.md>
