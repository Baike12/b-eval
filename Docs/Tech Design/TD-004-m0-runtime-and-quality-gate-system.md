# TD-004: M0 Runtime and Quality Gate System

| Field | Value |
|---|---|
| **ID** | TD-004 |
| **Status** | Draft |
| **Implements** | [FEAT-003 Evaluation Execution and Orchestration](../Features/execution/evaluation-execution-and-orchestration.md), [FEAT-008 Delivery Roadmap and Acceptance Criteria](../Features/platform/delivery-roadmap-and-acceptance-criteria.md) |
| **Owner** | Platform team |
| **Created** | 2026-10-22 |
| **Last Updated** | 2026-10-22 |

> This document specifies *how* the M0 runtime and the quality gate system are built. The *what* and *why* live in the Feature documents it implements.

## 1. Problem Restatement

The functional design names a ten-step evaluation flow, an agent adapter protocol, a scoring layer, a metric layer, a report and a set of user-facing screens. None of it exists. The M0 milestone needs a running system on which one task runs and produces a score, plus a gate that refuses to let a broken version pass.

Two decisions had to be made before any code: the shape of the runtime, and the shape of the gate.

## 2. Design Goals

| Goal | Consequence for the design |
|---|---|
| The whole slice runs | Go backend, TypeScript and React frontend, one HTTP surface between them |
| The gate is executable, not advisory | `b-eval gate` exits non-zero on a failure and the pipeline stops there |
| The gate leaves a trail | Every gate appends an evidence record that is never rewritten |
| The slice is testable without a model | The mock adapter is deterministic, so a failure means the code is wrong |
| The audit package is real | The store writes readable JSON, so a third party can recompute the numbers |

## 3. Non-Goals

- The M0 runtime does not bring up a live environment. `environment-agentic` tasks are modelled and validated, and their execution path is the same as a static-QA task minus the environment bring-up.
- The M0 runtime does not serve a public leaderboard, a human review workbench or an adversarial evaluation. Those are M3 and M4.
- The M0 runtime does not accept arbitrary code as a custom metric. A registration names one of the computations the platform offers and supplies its parameters.

## 4. Runtime Architecture

```
cmd/b-eval                 serve · gate · seed · version
internal/domain            the data model: task, task set, plan, trial, trajectory, verdict
internal/domain/validate.go the form contract and its validator, declared together
internal/adapter           the four-capability protocol, the HTTP adapter, the mock adapter, the self-check
internal/scoring           the scorers and the three-state verdict
internal/metrics           the metric registry, the built-in metrics, the confidence intervals
internal/runner            the plan execution, the budgets, the failure classification
internal/report            the single-run report
internal/store             the file-backed record store
internal/api               the HTTP surface
internal/gate              the gate runner, the checks, the evidence
internal/web               the embedded single-page application
web/                       the frontend source
```

The dependency direction is one-way: `domain` depends on nothing, and every other package depends on it. The runner depends on the adapter, the scoring and the metric layers; the API depends on the store and the runner; the gate depends on none of the runtime packages except the API, which it exercises over HTTP.

**Why the gate depends on the API and not on the internals.** A gate that calls the packages directly proves the packages work. A gate that starts the HTTP server and drives the real endpoints proves the product works. The gate uses the second one.

## 5. The Store

One JSON file per record under the data directory:

```
data/
├── tasksets/<task-set-id>__<version>.json
├── agents/<agent-id>.json
└── runs/<run-id>.json
```

A write goes to a temporary file and is renamed into place, so a crash during a write cannot leave a half-written record that fails to parse on the next start.

**Why JSON and not a database.** The audit package promises that a third party can recompute a score without the platform. A JSON record is readable without the platform; a database file is not. The migration path is a repository interface with a SQL implementation behind it, which M1 introduces when the query load justifies it.

## 6. The HTTP API

| Method | Path | Purpose |
|---|---|---|
| GET | `/api/healthz` | Liveness and the record counts |
| GET/POST | `/api/tasksets` | List and create task sets |
| GET | `/api/tasksets/{id}/{version}` | Read one task set |
| POST | `/api/tasksets/{id}/{version}/tasks` | Add one task to a task set |
| GET/POST | `/api/agents` | List and register agents |
| POST | `/api/agents/{id}/selfcheck` | Run the adapter self-check |
| GET/POST | `/api/metrics` | Read the registry and register a metric |
| GET | `/api/scorers` | The registered scorer types |
| GET/POST | `/api/runs` | List and start runs |
| GET | `/api/runs/{id}` | Read one run |
| GET | `/api/runs/{id}/report` | The report of one run |
| GET | `/api/runs/{id}/trials/{trialID}` | One trajectory |
| GET | `/api/compare?a=&b=` | Two reports and their difference |

The request bodies are decoded with unknown fields rejected, so a typo in a field name is an error rather than a silently dropped value.

**The readiness gate.** A run is refused with `409` when the agent has no self-check result, or when the self-check did not pass. The refusal names the reason, because "your run was rejected" without a reason sends the user looking in the wrong place.

## 7. The Frontend

TypeScript and React, built with Vite, embedded into the binary. The screens are:

| Screen | Content |
|---|---|
| Dashboard | The API state, the run form, the recent runs |
| Task sets | The stored task sets and their tasks, with the form of every task |
| Task | One task: its instruction, its contract fields, its judgement |
| Agents | The registered agents and their self-check results |
| Runs | The executed evaluations |
| Run report | The confidence section, the main score, the metrics, the breakdown, the tasks, the failure analysis |
| Metrics | The metric registry and the registration form |

**Two rules the frontend enforces.** The confidence section is the first block of a report and has no control that hides it. A metric without coverage is rendered as an explicit label and never as a zero, because a metric of zero and a metric with no data are different facts.

## 8. The Quality Gate System

### 8.1 The two references

The gate combines two ideas.

The **skeleton** comes from bblog: a check is an executable command, the pipeline runs numbered stages, and a failing stage blocks the next one. bblog's `cmd/bblog check` validates the content tree and exits 1 on an error; `scripts/build.sh` runs numbered stages under `set -euo pipefail`; `scripts/publish.sh` validates, then publishes, then health-checks.

The **interior** comes from loopx: a gate is a first-class operation that produces a classification and a recommended action, and every gate appends an evidence record to a file that is only ever added to. loopx's `promotion_gate.py` records a readiness event with a classification and an action; the loopx examples carry one smoke test per gate.

### 8.2 The gate set

| Gate | Protects | Implementation |
|---|---|---|
| `go-fmt` | Every Go file is gofmt-clean | A dedicated check, because `gofmt -l` exits zero and only prints |
| `go-vet` | No suspicious constructs | `go vet ./...` |
| `go-test` | The unit tests pass | `go test ./...` |
| `go-build` | The platform compiles | `go build ./...` |
| `web-typecheck` | The frontend type-checks | `npm run typecheck` |
| `web-test` | The frontend tests pass | `npm run test` |
| `web-build` | The frontend builds | `npm run build` |
| `docs-english` | No non-English text entered the tree | A CJK scan over the tracked file types |
| `docs-links` | Every local link resolves | A Markdown link and reference-definition scan |
| `docs-index` | Every document appears in its index | A coverage scan of `Docs/Features` and `Docs/Fix` |
| `e2e-smoke` | The vertical slice works | The API driven over HTTP from a temporary data directory |

A frontend gate is **skipped**, not passed, when `web/node_modules` is absent. A skipped gate produces the `insufficient_evidence` classification, which is what keeps a skipped gate from being read as a green one.

### 8.3 The verdict model

| Status | Classification | Recommended action |
|---|---|---|
| `passed` | `ready` | proceed |
| `failed` | `blocked` | fix the failing check and re-run the gate |
| `skipped` | `insufficient_evidence` | run the check manually; the gate produced no verdict |

The overall verdict is `blocked` when any gate is `blocked`, `insufficient_evidence` when any gate was skipped and none blocked, and `ready` otherwise.

### 8.4 The evidence

```
.gates/evidence.jsonl    append-only, one JSON record per line
.gates/latest.json       the latest state of every gate
```

Each record carries the schema version, the generation time, the run identifier, the gate identifier, the status, the classification, the recommended action, the duration and the detail. The file is opened in append mode and each record is one line, so a reader always sees whole records even while a run is in progress.

**Why the evidence is never rewritten.** A gate run that could edit its own history would make the history worthless: the record of what was known when a decision was made would be the record of what the writer wanted it to be.

### 8.5 The negative smoke

`scripts/gate-smoke.sh` builds broken fixtures in a temporary directory and asserts that the matching gate refuses them: a broken link, non-English text and a document missing from its index. The end-to-end smoke proves the platform works; the negative smoke proves the checks notice when it does not. A gate that never fails is not a gate.

## 9. Verification

Verified by [TR-003 M0 Vertical Slice][tr-003].

## 10. Open Points

- The store is file-backed. A concurrent write to the same task set is serialised by an in-process lock, which is enough for one process and not enough for several. M1 replaces the store behind a repository interface.
- The environment bring-up of an `environment-agentic` task is not implemented. Such a task is modelled, validated and admitted; running it needs an environment provider that M1 introduces.
- The scorer registry declares six categories; `J1` (state check) and `J6` (human review) have no implementation in M0, and a task that names them is reported as unscored rather than as failed.
- The bootstrap interval uses a package-level seed. It should become a per-run setting so that two runs with different seeds can be compared.

## References

- Feature: [FEAT-003 Evaluation Execution and Orchestration][feat-003]
- Feature: [FEAT-008 Delivery Roadmap and Acceptance Criteria][feat-008]
- Test Report: [TR-003 M0 Vertical Slice][tr-003]
- Adapter protocol: [TD-001 Unified Agent Adapter Protocol][td-001]

[feat-003]: <../Features/execution/evaluation-execution-and-orchestration.md>
[feat-008]: <../Features/platform/delivery-roadmap-and-acceptance-criteria.md>
[td-001]: <./TD-001-unified-agent-adapter-protocol.md>
[tr-003]: <../Test Report/TR-003-m0-vertical-slice.md>
