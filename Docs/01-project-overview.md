# 01 · Project Overview

## 1. Background and Problem

The open-source community already has a large number of Agent benchmarks (SWE-bench 6068 stars, OSWorld 3229, τ²-bench 2284, AgentBench 3784, etc.), but treating them as "usable evaluation infrastructure" runs into five walls:

**Wall 1: No unified way to onboard the system under test.**
τ-bench's Agent follows a fixed loop that injects a tool message after the first tool_call; VitaBench uses `generate_next_message`; MCPMark uses the LiteLLM tool loop; AgentCompass uses `run_task`. To run the same Agent on two benchmarks you must write two adapter layers, and the behavior semantics of the two sides are not equivalent, so cross-benchmark scores are not comparable.

**Wall 2: No unified scoring abstraction.**
OSWorld uses a two-layer decoupling of "result getters (43 kinds) × judgement functions (120 kinds)"; MCPMark uses a standalone scoring script that judges by exit code; TheAgentCompany accumulates pluggable checkpoints; AgentDojo splits tasks into two orthogonal axes, "utility" and "safety"; VitaBench uses a sliding window over rubric entries. Scorers cannot be reused across projects, so evaluation platform capabilities cannot accumulate.

**Wall 3: Infrastructure failures directly contaminate scores.**
τ-bench v1 records any exception (including API errors) as a 0 score and counts it in the denominator, which drags pass^k down with data center jitter; OSWorld uses `unscored_errors` to exclude "the environment crashed" from the denominator; ClawBench does the same with `adjusted_eligible`. The three conventions are mutually incompatible, and none of them is written into the score itself — so anyone reading the leaderboard cannot tell whether a 0 means the model failed or the network dropped.

**Wall 4: Scoring logic is visible to the system under test.**
In most benchmarks the scoring script simply sits in the task directory, and the Agent under test can read "what conditions count as a pass" as soon as it has a shell. TheAgentCompany avoids this with encrypted scripts, but it is the only one that does.

**Wall 5: No quality control loop for task authoring.**
The cost of authoring tasks for a benchmark is extremely high (OSWorld's 369 tasks and terminal-bench's 241 tasks are all hand-built), yet there is generally no upfront validation of "the reference solution solves it / the scorer will not misjudge it / the difficulty label is trustworthy", so defects in the tasks themselves stay latent for a long time.

## 1.2 Project Positioning

**b-eval is a general-purpose Agent evaluation platform, not a specific Agent benchmark.**

What it provides is the complete capability of "running a set of Agents on a set of reproducible, explainable and comparable tasks to produce scores"; which tasks are actually measured is decided by the task sets inside the platform.

Division of labor with existing projects:

| Relationship | Target | b-eval's stance |
|---|---|---|
| Reuse | Existing benchmarks such as SWE-bench / OSWorld / τ²-bench / AgentBench | Onboarded as "connectable task sets"; the platform does not rebuild tasks |
| Replace | The evaluation framework bundled with each benchmark | Provide unified run orchestration, scoring and reporting capabilities |
| Fill gaps | Human rubrics, unified cross-benchmark metrics, observable scoring | The parts existing projects generally lack, provided in a unified way by the platform |
| Extend | Private task sets, regression testing of internal Agent versions | The platform must support the full lifecycle of non-public tasks |

## 1.3 Goals

**G1 · One adapter protocol takes all Agents.** Any Agent only needs to implement a minimal interface to be evaluated, and the evaluation behavior is decoupled from the Agent's internal implementation. Acceptance criterion: the integration cost for that Agent, without modifying the platform, is ≤ 1 day.

**G2 · One scoring abstraction takes all task types.** Deterministic checks, final environment state comparison, rule assertions, script-based scoring, LLM judging and human review — all six kinds of scorers are unified under the same model and composable across tasks and benchmarks.

**G3 · Scores are trustworthy and explainable.** The score of any evaluation run can answer three questions: which tasks did not score, why they did not score, and whether a "no score" is the model's problem or the platform's problem.

**G4 · Scores are comparable across Agents on the same tasks.** Under the same task set and the same budget convention, the scores of different Agents are directly comparable and are not affected by the default values of their respective frameworks.

**G5 · Evaluation cost is predictable.** Before a run you can estimate the token volume, cost and time upper bound of that evaluation; if the run exceeds them, you know immediately.

**G6 · Task quality has an upfront gate.** Before a new task enters an official task set, it must pass three validations: solvability, scoring trustworthiness and difficulty trustworthiness.

## 1.4 Non-Goals

- **Not a general-purpose Agent framework.** b-eval does not provide runtime capabilities such as Agent planning, memory or reflection; it only provides "onboarding the system under test" and "running it".
- **Not model training / reinforcement learning.** Evaluation and training are two different things. Exporting trajectories into a reusable format is within this platform's boundary, but the training loop is out of scope.
- **Not pursuing task volume.** Prefer 200 tasks that hold up to scrutiny over 2000 tasks mixed with quality problems.
- **Not general-purpose quality evaluation.** No competition with DeepEval / Ragas / promptfoo for the "LLM output quality assertion" market; b-eval only cares about Agent behavior.
- **Not pursuing a single score.** No metric of the form "one number represents Agent capability"; the platform mandates output decomposed along at least two axes: capability dimension and domain.

## 1.5 User Roles

| Role | Typical need | Core use case |
|---|---|---|
| **Agent developer** | How does my Agent perform on real tasks? Did this change make it better or worse? | Submit evaluations, read reports, compare against the previous version, inspect failed trajectories to locate problems |
| **Agent developer / platform engineer** | I need to give someone a runnable Agent — how do evaluations hook in? | Onboard an Agent, get the sample task running, review the onboarding self-check results |
| **Task author** | I want to add a batch of test tasks | Create tasks, attach scorers, run validation, submit into the set |
| **Benchmark maintainer** | Do this benchmark's tasks still hold up? Is the difficulty tiering still accurate? | Version management, task set auditing, task revision and retirement |
| **Evaluation organizer** | Which of these Agents is stronger? Can the conclusion be trusted? | Batch orchestration, cross-benchmark comparison, leaderboard publishing, result auditing |
| **Security reviewer** | Could this Agent be steered off course by prompt injection? | Red and blue team task sets, ASR reports, defense strategy comparison |

## 1.6 Core Concept Model

The platform has only seven first-class concepts; everything else is a configuration of them.

```
Benchmark
  └─ TaskSet  ── versioned, frozen, publishable
       └─ Task
            ├─ Domain         environment/business scenario the task belongs to
            ├─ CapabilityTag  capability dimension the task examines
            ├─ Difficulty     difficulty tier of the task
            ├─ InitialState   environment snapshot and data before the task starts
            ├─ Contract       action space, observation form, termination conditions, budget
            └─ Judgement      scorer set and how the scorers are combined

System under test (Agent) ── AgentAdapter ── AgentProtocol

Run ── Plan ── Trial ── Trajectory
                        └─ Verdict ── Metric
```

**Key constraints:**

- **Tasks and Agents are fully decoupled.** A task does not know an Agent exists, and an Agent does not know what the task is; only the "contract" is the single interface between the two sides.
- **A Trial is the atomic unit of evaluation.** One Trial = one complete execution of "task × Agent × random seed × budget". All metrics are ultimately aggregated from Trials.
- **Judgement and execution are separated.** Judgement can be re-executed over already saved trajectories without re-running the Agent (offline rescoring). This is the foundation for troubleshooting, changing the scoring convention and third-party audits.
- **Every score must be drillable down to Trials.** Every number in a report can be traced back to the specific Trials it was computed from.

## 1.7 Differentiation

Relative to existing work, b-eval has exactly three things it must do, and that nobody currently does:

1. **A unified Agent adapter protocol + unified cross-benchmark metrics.** The same Agent runs SWE-bench, OSWorld and τ²-bench through the same adapter layer, and the resulting scores land in the same coordinate system.
2. **Invisible scoring + auditable scoring.** The scoring logic is invisible to the Agent under test; at the same time it is fully auditable by platform operators — every score change carries an author, a reason and a timestamp.
3. **Score confidence is explicit.** Every score carries information about "how trustworthy this score is": how many Trials were excluded for platform reasons, how high the judge agreement is, how wide the confidence interval is.

## 1.8 Open Questions (Require Product Decisions)

| ID | Question | Area of impact |
|---|---|---|
| Q1 | Should the platform operate a public leaderboard? A public leaderboard requires a submission flow and anti-cheating measures; a private deployment does not | Submission flow, result visibility, review queue |
| Q2 | Should invisible scoring code be delivered "encrypted", or should "scoring run in an isolated environment where the tested party has no read access"? | Task package format, sandbox topology |
| Q3 | Are third parties allowed to submit Agent results? If so, the scoring logic is necessarily visible to the submitter, which conflicts with G-"invisible scoring" | Leaderboard design and security boundary |
| Q4 | Are task sets required to be tiered (public subset / semi-public / private)? The existence of private sets makes scores impossible to compare across the board | Score comparability, leaderboard design |
| Q5 | Is the user simulator (a persona-based user) a first-class citizen or an optional component? | Task model, interaction contract |
