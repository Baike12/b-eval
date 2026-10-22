# Features Index

> Index of every feature requirement document in b-eval.

The functional design of the platform lives here. Each capability area is one feature
document, written to [`../Feature-Template.md`](../Feature-Template.md), and filed in one
folder per category.

**Tech Design and Test Report documents are not indexed here.** Each one is reachable from
the `References` section of the feature it belongs to, so a single hop from any feature
document reaches its design and its verification.

| Category | What it covers |
|---|---|
| [Platform](#platform) | Scope, positioning, goals, roles, concept model, and the delivery plan |
| [Task System](#task-system) | Capability taxonomy, task authoring, task sets, difficulty tiers, validation gates |
| [Agent Integration](#agent-integration) | How an agent under test is plugged in, and how the plug-in is verified |
| [Execution](#execution) | Evaluation plans, orchestration, budget, concurrency, resume, traces |
| [Scoring](#scoring) | Scorer taxonomy, verdicts, metric definitions, aggregation |
| [Reporting](#reporting) | Artefacts, run reports, comparison views, error attribution, leaderboard |
| [Security](#security) | Anti-cheating, adversarial evaluation, quality assurance, contamination, platform health |
| [Experience](#experience) | User journeys, screens, interaction principles, permission model, requirement inventory |

---

## Platform

| ID | Feature | Summary |
|---|---|---|
| [FEAT-001](platform/platform-scope-and-positioning.md) | Platform Scope and Positioning | Why the platform exists, what it deliberately does not do, who uses it, and the seven core concepts everything else is expressed in. |
| [FEAT-008](platform/delivery-roadmap-and-acceptance-criteria.md) | Delivery Roadmap and Acceptance Criteria | The M0–M4 milestone sequence, the acceptance criteria that close each milestone, and the risk register. |

## Task System

| ID | Feature | Summary |
|---|---|---|
| [FEAT-002](task-system/capability-taxonomy-and-task-system.md) | Capability Taxonomy and Task System | The capability and domain taxonomy, the two task forms with their per-form metadata contract and per-form validation gates, task set partitioning, difficulty tiers, and environment isolation. |

## Agent Integration

| ID | Feature | Summary |
|---|---|---|
| [FEAT-009](agent-integration/unified-agent-adapter-protocol.md) | Unified Agent Adapter Protocol | The four adapter capabilities, three adapter shapes, the task-side interaction contract, the protocol-violation accounting, and the self-check that gates evaluation. |

## Execution

| ID | Feature | Summary |
|---|---|---|
| [FEAT-003](execution/evaluation-execution-and-orchestration.md) | Evaluation Execution and Orchestration | The end-to-end run flow, the interaction contract, the six budget classes, concurrency and rate limiting, the six failure classes, resume semantics, and the trial record. |

## Scoring

| ID | Feature | Summary |
|---|---|---|
| [FEAT-004](scoring/scoring-and-metrics.md) | Scoring and Metrics | Six scorer categories behind one three-state verdict contract, the metric registry with its locked built-in set and its custom registrations, decomposition and aggregation conventions, and the worked example that pins the semantics. |

## Reporting

| ID | Feature | Summary |
|---|---|---|
| [FEAT-005](reporting/reporting-analysis-and-comparison.md) | Reporting, Analysis and Comparison | The four artefact layers, the credibility-first run report, six levels of comparison, five-level error attribution, the submission and anti-cheat pipeline, and the five leaderboard guardrails. |

## Security

| ID | Feature | Summary |
|---|---|---|
| [FEAT-006](security/security-robustness-and-quality-assurance.md) | Security, Robustness and Quality Assurance | Anti-cheating rules, adversarial evaluation on the utility/ASR axes, task quality gates, judge calibration, contamination control, and the platform's own health metrics. |

## Experience

| ID | Feature | Summary |
|---|---|---|
| [FEAT-007](experience/user-flows-and-feature-inventory.md) | User Flows, Feature Inventory and Permissions | The five user journeys, the 69-entry functional requirement inventory, the screen inventory, the seven interaction principles, and the permission model that separates judging rights from submission rights. |
