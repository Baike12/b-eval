# b-eval — Agent Evaluation Platform Design

> Functional-level design for **b-eval**, a general-purpose Agent evaluation platform.
>
> This directory describes **what the system does**, not how it is built. Architecture,
> interface protocols, storage formats and deployment choices are explicitly out of scope
> here and are documented separately under [`Tech Design/`](<./Tech Design/>).
> Every place needing a technical decision is marked `[pending technical design]`.

**Language policy:** all content in this repository is English. See [`../Agent.md`](../Agent.md).

## 1. Document Index

### 1.1 Functional design

| Document | Content | Question it answers |
|---|---|---|
| [01 Project Overview](./01-project-overview.md) | Background, positioning, goals, non-goals, roles, core concept model | Why build it, what it is |
| [02 Capability Dimensions and Task System](./02-capability-dimensions-and-task-system.md) | Capability taxonomy, task metadata, task set partitioning, difficulty tiers, environment isolation | What is measured, how tasks are organised |
| [03 Evaluation Execution and Orchestration](./03-evaluation-execution-and-orchestration.md) | Evaluation flow, agent adapter protocol, interaction contract, budget, concurrency, resume, traces | How a run works |
| [04 Scoring and Metrics](./04-scoring-and-metrics.md) | Scorer taxonomy, combination rules, three-state verdicts, metric formulas, aggregation | How results are scored and summarised |
| [05 Reporting, Analysis and Comparison](./05-reporting-analysis-and-comparison.md) | Artefact inventory, single-run report, comparison views, error attribution, leaderboard | How results are read and compared |
| [06 Security, Robustness and Quality Assurance](./06-security-robustness-and-quality-assurance.md) | Anti-cheating, adversarial evaluation, task self-checks, judge calibration, contamination | Why the scores can be trusted |
| [07 User Flows and Feature Inventory](./07-user-flows-and-feature-inventory.md) | User journeys, feature inventory, screens and interaction, permission model | How users work with it |
| [08 Roadmap and Acceptance](./08-roadmap-and-acceptance.md) | Milestones, acceptance criteria, risk register | How it gets built, how it is judged done |

### 1.2 Delivery documents

| Directory | Holds | Naming | Template |
|---|---|---|---|
| [`Features/`](./Features/) | Feature requirement documents | `FEAT-<nnn>-<slug>.md` | [`Feature-Template.md`](./Feature-Template.md) |
| [`Fix/`](./Fix/) | Bug fix documents | `FIX-<nnn>-<slug>.md` | [`Fix-Template.md`](./Fix-Template.md) |
| [`Tech Design/`](<./Tech Design/>) | Technical design documents | `TD-<nnn>-<slug>.md` | — |
| [`Test Report/`](<./Test Report/>) | Verification / test reports | `TR-<nnn>-<slug>.md` | — |

Current contents:

| Feature | Tech Design | Test Report |
|---|---|---|
| [FEAT-001 Unified Agent Adapter Protocol](./Features/FEAT-001-unified-agent-adapter-protocol.md) | [TD-001 Unified Agent Adapter Protocol](<./Tech Design/TD-001-unified-agent-adapter-protocol.md>) | [TR-001 Unified Agent Adapter Protocol](<./Test Report/TR-001-unified-agent-adapter-protocol.md>) |
| — | — | [TR-002 Black-box CLI Empty Action Sequence Fix](<./Test Report/TR-002-blackbox-cli-empty-action-sequence-fix.md>) |

| Fix | Test Report |
|---|---|
| [FIX-002 Black-box CLI Empty Action Sequence Not Flagged](./Fix/FIX-002-blackbox-cli-empty-action-sequence-not-flagged.md) | [TR-002 Black-box CLI Empty Action Sequence Fix](<./Test Report/TR-002-blackbox-cli-empty-action-sequence-fix.md>) |

### 1.3 Conventions

- Feature documents **must** reference their Tech Design document and their Test Report document, using Markdown reference-style links with the definitions at the bottom of the file.
- Fix documents **must** reference their Test Report document.
- Feature documents **must** list the functional requirement IDs they implement, in the `Implements` field.
- Paths into `Tech Design/` and `Test Report/` are wrapped in angle brackets because those directory names contain spaces.
- Identifier namespaces: `F-<module>-<nn>` (functional requirement) · `FEAT-<nnn>` (feature document) · `FIX-<nnn>` (fix document) · `TD-<nnn>` (technical design) · `TR-<nnn>` (test report).

---

## 2. Open-Source Landscape Survey

Method: Google and GitHub browsed via BrowserOS; star counts read from GitHub page data.
Survey date: 2026-10-22. All reference projects are mirrored read-only into `vendor/` at the repository root
(excluded from version control by `.gitignore`).

### 2.1 Benchmarks — the systems under test

| Project | Stars | Capability coverage | What b-eval borrows |
|---|---:|---|---|
| [SWE-bench](https://github.com/SWE-bench/SWE-bench) | 6068 | Coding agents | Task = issue + repository snapshot; fail-to-pass / pass-to-pass dual test sets; test-output boundary markers cross-checked against real execution evidence; anti-cheating rule set |
| [τ²-bench](https://github.com/sierra-research/tau2-bench) | 2284 | Tool–agent–user interaction | **"Reference trace ≠ required path"**: derive the target environment end state by replaying the reference trace, then compare hashes, instead of comparing action sequences step by step; pass^k formula; separation of infrastructure error from model failure; frozen task subsets with digests |
| [τ-bench](https://github.com/sierra-research/tau-bench) | 1472 | Tool calling | Five user-simulator strategies (human/llm/react/verify/reflection); explicit stop token; conditional-branch instructions that carry "the user changes their mind mid-call" |
| [OSWorld](https://github.com/xlang-ai/OSWorld) | 3229 | Desktop GUI | **Result-fetching layer separated from judgement layer** (43 getters × 120 metrics, referenced by string name in the task file); composite verdicts with `conj=and/or` and partial credit; unscored errors excluded from the denominator |
| [WebArena](https://github.com/web-arena-x/webarena) | 1636 | Web agents | Task templating (190 templates → 812 instances); three evaluator families combined by multiplication |
| [VisualWebArena](https://github.com/web-arena-x/visualwebarena) | 490 | Multimodal web | Three-axis difficulty labelling (reasoning / visual / overall); image targets; inter-task dependency annotation |
| [BrowserGym](https://github.com/ServiceNow/BrowserGym) | 1411 | Web environment abstraction | Three-method task contract (setup / validate / teardown); actions as functions that auto-generate descriptions, examples and tool schemas; orthogonal observation modalities with viewport-level trimming |
| [AgentBench](https://github.com/THUDM/AgentBench) | 3784 | General, 8 environments | Task / environment / client three-layer decoupling; unified status machine with error-class shares; a single scoring hook returning arbitrary JSON |
| [AgentBoard](https://github.com/hkust-nlp/AgentBoard) | 450 | General, 9 tasks | **Dual-track success rate + progress rate**; six-dimension weighted capability score; automatic difficulty-stratified charts; score-change points yielding a per-step score curve |
| [AgentDojo](https://github.com/ethz-spylab/agentdojo) | 948 | Adversarial security | Attack/defense registries with composable pipelines; **probe-based automatic discovery of which injection points the model can actually see**; utility and ASR reported on two axes |
| [MCPMark](https://github.com/eval-sys/mcpmark) | 467 | MCP ecosystem | Three-part task layout (metadata / task description / standalone verification script); Setup→Execute→Verify→Cleanup state isolation; pinned MCP server versions |
| [VitaBench](https://github.com/meituan-longcat/vitabench) | 182 | Real service workflows | **Atomic rubric entries with sliding-window evaluation** (window 10, overlap 2; a rubric's state carries across windows and can be overturned); anti-cheating rules embedded in the judge prompt |
| [MemoryAgentBench](https://github.com/HUST-AI-HYZ/MemoryAgentBench) | 477 | Long-term memory | "Inject once, query repeatedly"; capability × context-length matrix; same tasks with different agent architectures as a controlled comparison |
| [TheAgentCompany](https://github.com/TheAgentCompany/TheAgentCompany) | 801 | Simulated company work | **Checkpoint-based partial credit** with pluggable accumulation strategies; NPC colleague scripts; **judgement code hidden from the system under test** |
| [ClawBench](https://github.com/TIGER-AI-Lab/ClawBench) | 1095 | Everyday browser tasks | Two-stage scoring (deterministic interception → LLM adjudication); interception doubles as protection against irreversible actions; separation of "the model failed" from "the environment broke" |
| [terminal-bench](https://github.com/laude-institute/terminal-bench) | 2605 | Terminal tasks | Per-task isolated environment, isolated test environment and reference solution; difficulty tiers with human time baselines; dataset registry with version fingerprints |
| [MLE-bench](https://github.com/OpenAI/mle-bench) | 1783 | ML engineering | End-to-end judgement of real engineering artefacts |
| [AgentCompass](https://github.com/open-compass/AgentCompass) | 170 | Unified evaluation infrastructure | **Seven-layer responsibility split** (model / benchmark / harness / environment / recipe / runtime / analyzer); metric contracts (`binary_success` / `scalar`); six categories of pluggable analyzer |
| [awesome-evals](https://github.com/benchflow-ai/awesome-evals) | 994 | Resource index | Taxonomy of evaluation infrastructure; benchmark-completeness checklist (contamination, saturation, label errors, leaderboard gaming) |
| [ai-agent-benchmark-compendium](https://github.com/philschmid/ai-agent-benchmark-compendium) | 200 | Resource index | Four-category classification: function calling and tool use / general assistant and reasoning / coding and software engineering / computer interaction |
| [Agent evaluation survey companion](https://github.com/Asaf-Yehudai/LLM-Agent-Evaluation-Survey) | 114 | Methodology | Four foundational capabilities (planning / tool use / self-reflection / memory) plus four application domains; the capture → interpret → trust framing |

### 2.2 Evaluation and execution frameworks

| Project | Stars | What b-eval borrows |
|---|---:|---|
| [Inspect AI](https://github.com/UKGovernmentBEIS/inspect_ai) | 3058 | The six elements of an evaluation (task / solver / scorer / metric / model / sandbox); self-describing components reconstructable from a log; the six-limit budget set (message / token / turn / time / working / cost); offline re-scoring; a complete score-edit audit chain |
| [OpenAI Evals](https://github.com/openai/evals) | 19626 | The "swap out the system under test" abstraction; dataset + eval class + YAML registration; version-bearing names |
| [SWE-agent](https://github.com/princeton-nlp/SWE-agent) | 20607 | Agent runtime feature surface: prompt template system, tool blocklists, context-management chain, cost and call-count budgets, trajectory replay, hook mechanism |
| [FastChat / Chatbot Arena](https://github.com/lm-sys/FastChat) | 39549 | Crowdsourced human-preference pairwise evaluation and Elo-style ranking |
| [promptfoo](https://github.com/promptfoo/promptfoo) | 26245 | Productised shape of prompt/agent regression testing and red-teaming suites |
| [DeepEval](https://github.com/confident-ai/deepeval) | 19020 | Assertions as first-class citizens in test cases |
| [AgentScope](https://github.com/alibaba/AgentScope) | 33598 | Message-level abstraction for multi-agent orchestration |
| [agent-lightning](https://github.com/microsoft/agent-lightning) | 18782 | Handoff between evaluation and training on a shared trajectory format |

### 2.3 Gaps confirmed by the survey — b-eval's opportunity

1. **No unified multi-environment scoring abstraction.** OSWorld's getters and metrics, MCPMark's verification script, TheAgentCompany's checkpoints and AgentDojo's utility/security split are all bespoke and cannot be reused across projects.
2. **No unified agent interface.** τ-bench, VitaBench, MCPMark and AgentCompass expose mutually incompatible agent contracts; evaluating one task set against two of them means writing two integrations.
3. **No mature pattern for fusing human-annotated rubrics, LLM judging and deterministic assertions.** VitaBench uses only an LLM judge, TheAgentCompany one or the other, MCPMark only deterministic checks.
4. **Infrastructure failure contaminates scores.** τ-bench v1 counted exceptions as zero in the denominator; OSWorld excludes them via `unscored_errors`; ClawBench does the same via `adjusted_eligible`. Three incompatible conventions, none visible in the score itself.
5. **Judgement logic is visible to the system under test.** Only TheAgentCompany encrypts its scoring scripts.
6. **Task production has no quality gate.** A generic oracle-solvability check, a scorer-trust check and a difficulty-calibration check are absent almost everywhere.

---

## 3. Usage Conventions

- Every functional requirement carries a **unique identifier** (for example `F-RUN-01`); the full table lives in [07 User Flows and Feature Inventory](./07-user-flows-and-feature-inventory.md).
- Every metric definition gives its **computation and its denominator**. No "roughly", no "reasonably good".
- Upstream open-source projects referenced in this directory are named as `vendor/<owner>__<repo>` in the local mirror, so the cited evidence can be checked.
