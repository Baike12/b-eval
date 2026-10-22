# b-eval

Agent evaluation platform.

## Status

| Stage | Content | Where |
|---|---|---|
| Functional design | What the system does (no implementation) | [`Docs/Features/`](./Docs/Features/) |
| Delivery documents | Features, fixes, technical design, test reports | [`Docs/Features/`](./Docs/Features/), [`Docs/Tech Design/`](<./Docs/Tech Design/>), [`Docs/Test Report/`](<./Docs/Test Report/>), [`Docs/Fix/`](./Docs/Fix/) |
| Technical design | Architecture and implementation | In progress |
| Implementation | Not started | — |

## Working Rules

Read [`Agent.md`](./Agent.md) before contributing. It defines the English-only language policy, the document templates, the mandatory cross-references between feature, technical design and test report documents, and the identifier namespaces.

[`AGENTS.md`](./AGENTS.md) adds the repository-specific notes on version control.

## Directory Layout

```
b-eval/
├── Agent.md                 # Operating rules: language policy, templates, cross-references
├── AGENTS.md                # Repository-specific working rules
├── Docs/
│   ├── README.md            # Index of all documentation
│   ├── Feature-Template.md  # Template for Docs/Features/
│   ├── Fix-Template.md      # Template for Docs/Fix/
│   ├── Features/            # FEAT-<nnn>-<slug>.md — the functional design lives here
│   │   ├── FEAT-001..008    # Platform capability areas
│   │   └── FEAT-009         # Agent adapter protocol (deep dive)
│   ├── Fix/                 # FIX-<nnn>-<slug>.md
│   ├── Tech Design/         # TD-<nnn>-<slug>.md
│   └── Test Report/         # TR-<nnn>-<slug>.md
└── vendor/                  # Read-only mirrors of third-party reference projects (not versioned)
```

## Reference Projects

`vendor/` holds 24 high-star open-source projects related to agent evaluation, cloned read-only for design comparison.
Selection criteria and per-project takeaways: [Docs/README.md §2](./Docs/README.md#2-open-source-landscape-survey).

- **Benchmarks**: SWE-bench, τ²-bench, τ-bench, OSWorld, WebArena, VisualWebArena, BrowserGym, AgentBench, AgentBoard, AgentDojo, MCPMark, VitaBench, MemoryAgentBench, TheAgentCompany, ClawBench, terminal-bench, MLE-bench, AgentCompass
- **Resource indexes**: awesome-evals, ai-agent-benchmark-compendium, agent evaluation survey companion
- **Evaluation and execution frameworks**: Inspect AI, OpenAI Evals, SWE-agent, FastChat, promptfoo, DeepEval, AgentScope, agent-lightning
