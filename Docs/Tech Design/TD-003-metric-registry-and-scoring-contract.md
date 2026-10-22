# TD-003: Metric Registry and Scoring Contract

| Field | Value |
|---|---|
| **ID** | TD-003 |
| **Status** | Draft |
| **Implements** | [FEAT-004 Scoring and Metrics](../Features/scoring/scoring-and-metrics.md) |
| **Owner** | Platform team |
| **Created** | 2026-10-22 |
| **Last Updated** | 2026-10-22 |

> This document specifies *how* the metric registry and the scoring contract are built. The *what* and *why* live in the Feature document it implements.

## 1. Problem Restatement

FEAT-004 requires every metric to be resolved through one registry, with a locked built-in set and an open path for registered custom metrics. A registry that is only a lookup table does not satisfy the requirement: the requirement is that a metric's definition lives in exactly one place, that a built-in metric cannot be redefined by a run or a task, and that a malformed registration is refused with a message naming what is missing.

The registry therefore needs three things the plain lookup table does not have: a **completeness gate** on entry, a **lock** on the built-in subset, and a **version** on every entry.

## 2. Design Goals

| Goal | Consequence for the design |
|---|---|
| One definition per metric | The registry is the only resolution point; no call site recomputes a formula. |
| Built-in metrics cannot move | The built-in set is registered once at startup and the registry refuses to overwrite a built-in ID. |
| A registration is either complete or refused | A completeness gate rejects an entry missing a mandatory field, naming the field. |
| A custom metric is never mistaken for a platform metric | `builtin: false` is carried on the entry and rendered everywhere the entry appears. |
| Old results stay readable after a metric changes | Every entry is versioned; a change creates a new version and old results keep the version they were computed under. |

## 3. Non-Goals

- The registry does not execute user-supplied expressions. A custom metric is a Go function registered through the API; it is not a formula string parsed at report time.
- The registry does not version the *computation code* of a metric. A code change that alters a metric's output is a platform change and is covered by the metric known-answer cases in the regression suite.

## 4. Entry Model

```go
type MetricDef struct {
    ID          string
    Name        string
    Description string
    Builtin     bool
    Unit        Unit    // UnitRatio | UnitCount | UnitScore
    Aggregation string  // AggMacro | AggMicro | AggNone
    Numerator   string
    Denominator string
    Formula     string
    CIMethod    string  // "wilson" | "normal" | "bootstrap" | ""
    Version     string
    Compute     func(in MetricInput) (MetricValue, error)
}
```

The five mandatory fields are `Unit`, `Aggregation`, `Numerator`, `Denominator` and `Formula`. `CIMethod` is additionally mandatory when `Unit == UnitRatio`.

## 5. Registry

```go
type Registry struct {
    mu      sync.RWMutex
    entries map[string]MetricDef
    order   []string
}

func (r *Registry) Register(def MetricDef) error
func (r *Registry) Get(id string) (MetricDef, bool)
func (r *Registry) List() []MetricDef
```

`Register` enforces, in order:

1. **ID is non-empty** and contains no whitespace.
2. **No collision.** If the ID exists and the existing entry is built-in, the registration is refused with `built-in metric ID is reserved`. If the existing entry is custom, the registration is refused with `metric ID already registered`; updating an existing metric is a versioned operation, not an overwrite.
3. **All five mandatory fields are present and non-empty.** The error names every missing field, not just the first, so an author fixes the registration in one pass.
4. **A ratio metric declares a CI method**, and the CI method is one of the three supported values.
5. **A custom metric declares a non-empty `Formula` and `Denominator` in prose**, because those two fields are what a third party reads to recompute the number. A custom metric whose `Formula` is a placeholder such as `TBD` is refused.

The built-in entries are registered through the exact same `Register` path. There is no privileged setter. This is deliberate: if the built-in set skipped the completeness gate, the gate would be untested by the entries that matter most.

## 6. Built-in Metrics

| ID | Unit | Aggregation | Definition |
|---|---|---|---|
| `success_rate` | ratio | macro | `Σ c_t / Σ n_t` over valid trials |
| `pass_at_k` | ratio | macro | `1 − Π (n_t − c_t − i)/(n_t − i)`; NA when `n_t < k` |
| `pass_power_k` | ratio | macro | `C(c_t, k) / C(n_t, k)`; NA when `n_t < k` |
| `avg_at_k` | score | micro | `Σ reward / Σ n_t` |
| `task_pass_rate` | ratio | macro | share of tasks that succeeded at least once |
| `progress_score` | score | micro | mean progress score over valid trials |
| `platform_failure_rate` | ratio | micro | share of trials excluded for platform reasons |
| `judgement_failure_rate` | ratio | micro | share of trials whose verdict is UNSCORED |
| `avg_steps_per_success` | score | macro | mean step count over trials judged as passing |
| `avg_tokens_per_success` | score | macro | mean input+output tokens over trials judged as passing |
| `avg_cost_per_success` | score | macro | mean cost over trials judged as passing |
| `invalid_action_rate` | ratio | micro | malformed actions / total responses |
| `repeated_action_rate` | ratio | micro | consecutive equivalent actions / total actions |
| `grounding_accuracy` | ratio | micro | accepted actions / total actions |

`success_rate` is the primary metric: it is the one the leaderboard and the total score are built from.

## 7. Computation Input and Output

```go
type MetricInput struct {
    Trials       []Trial
    K            int
    Aggregation  string
    CIMethod     string
}

type MetricValue struct {
    Value        float64
    HasValue     bool   // false => "no coverage", never 0
    Numerator    float64
    Denominator  float64
    SampleSize   int
    CI           *Interval
}
```

`HasValue == false` means **no coverage**. A metric with no coverage is excluded from aggregation and rendered as "no coverage"; it is never coerced to 0. The distinction matters: a metric of 0 and a metric with no data are different facts, and a report that conflates them understates coverage.

## 8. Aggregation

Two methods, both implemented over the same trial set:

| Method | Rule |
|---|---|
| `macro` | Compute the metric per task, then average over tasks. Each task weighs the same regardless of how many trials it has. |
| `micro` | Pool all trials of all tasks and compute the metric once. Each trial weighs the same. |

`macro` is the default. The report always shows both and labels them.

For a metric whose `Aggregation` is `none`, the value is returned as computed and is not aggregated further.

## 9. Confidence Intervals

| Method | Definition |
|---|---|
| `wilson` | Wilson score interval for a binomial proportion. The default: it stays inside `[0, 1]` and remains sensible at small `n`. |
| `normal` | Normal approximation `p ± z·√(p(1−p)/n)`. |
| `bootstrap` | Percentile bootstrap over the task-level values, 1000 resamples, fixed seed so the interval is reproducible. |

The interval method is chosen per run and recorded in the run record. A run that does not choose one gets `wilson`.

## 10. Custom Metrics

A custom metric is registered with a compute function and the mandatory metadata:

```go
reg.Register(MetricDef{
    ID:          "keyword_coverage",
    Name:        "Keyword coverage",
    Description: "Share of trials whose answer contains every required keyword.",
    Builtin:     false,
    Unit:        UnitRatio,
    Aggregation: AggMacro,
    Numerator:   "Trials whose answer contains every required keyword",
    Denominator: "Valid trials",
    Formula:     "count(answer contains all needles) / count(valid trials)",
    CIMethod:    "wilson",
    Version:     "1",
    Compute:     computeKeywordCoverage,
})
```

The compute function receives the trials and returns a `MetricValue`. It cannot change the aggregation method or the CI method: those come from the entry and from the run's settings, so a custom metric is aggregated exactly like a built-in one.

The registration path is exposed over HTTP (`POST /api/metrics`) so that a task set or an operator can register a metric without recompiling the platform.

## 11. Scoring Contract

The registry answers "which numbers". The scorer answers "what happened". The two are separate:

```go
type Verdict string

const (
    VerdictPass     Verdict = "PASS"
    VerdictFail     Verdict = "FAIL"
    VerdictUnscored Verdict = "UNSCORED"
)
```

A scorer returns exactly one of the three. `UNSCORED` is counted in the judgement failure rate and excluded from the success denominator — it is never equated with `FAIL`.

Scorer categories:

| Category | ID | Judgement basis |
|---|---|---|
| State check | `J1` | The environment's final state matches the expectation |
| Result retrieval | `J2` | A retrieved artifact matches the expectation |
| Action assertion | `J3` | The action sequence matches, when the task declares it |
| Rule criteria | `J4` | A declarative rule over the answer or the artifact |
| Model scoring | `J5` | A rubric item judged by a model |
| Human review | `J6` | A rubric item judged by a person |

M0 implements `J2`, `J3`, `J4` and `J5` with a rule-based judge for `J5`. `J1` needs a live environment and `J6` needs a human queue; both are declared in the registry of scorer types but are not executable in M0, and the platform refuses to report a metric that depends on them rather than reporting a wrong number.

## 12. Verdict Composition

A task's scorers are composed by `and` (conjunction) or `or` (disjunction). With weights:

```
score = Σ (weight_i · satisfied_i) / Σ weight_i
```

The task's verdict is `PASS` when the primary criterion is satisfied, `FAIL` when it is not, and `UNSCORED` when the primary criterion's scorer returned `UNSCORED` — an unscored primary criterion makes the whole task unscored, because a task cannot be judged on an auxiliary criterion alone.

## 13. Metric Provenance in Reports

Every metric rendered in a report carries:

- its registry ID,
- its version,
- whether it is built-in,
- the aggregation method used,
- the CI method used,
- its sample size,
- its coverage state.

A report consumer can therefore recompute the number from the exported audit package without knowing anything about the platform internals.

## 14. Alternatives Considered

| Alternative | Why rejected |
|---|---|
| A plain map of metric ID to function | Carries no completeness guarantee, no lock and no version; the requirement to refuse a malformed registration cannot be met. |
| Letting a run redefine a built-in metric | Two platforms could then publish the same metric ID with different meanings; cross-platform comparison becomes meaningless. |
| A user-supplied formula language | Adds a second, unreviewed definition path and an injection surface at report time. |
| Storing the metric definitions in a database row | The definitions are code; a data row that claims to define a metric is a second definition that can drift from the code. |

## 15. Verification

Verified by [TR-003 M0 Vertical Slice][tr-003].

## 16. Open Points

- The bootstrap interval needs a fixed seed to be reproducible; the seed is currently a package constant and should become a per-run setting.
- `pass_at_k` and `pass_power_k` need `k`; the run supplies it from the plan's repetition count. A task with fewer valid trials than `k` yields `no coverage`, which is currently the only place a built-in metric can be without a value.

## References

- Feature: [FEAT-004 Scoring and Metrics][feat-004]
- Test Report: [TR-003 M0 Vertical Slice][tr-003]

[feat-004]: <../Features/scoring/scoring-and-metrics.md>
[tr-003]: <../Test Report/TR-003-m0-vertical-slice.md>
