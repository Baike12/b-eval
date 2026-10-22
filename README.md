# b-eval

Agent evaluation platform.

## Status

| Stage | Content | Where |
|---|---|---|
| Functional design | What the system does (no implementation) | [`Docs/Features/`](./Docs/Features) |
| Delivery documents | Features, fixes, technical design, test reports | [`Docs/Features/index.md`](./Docs/Features/index.md), [`Docs/Fix/index.md`](./Docs/Fix/index.md) |
| Technical design | Architecture and implementation | In progress |
| Implementation | Not started | — |

## Working Rules

Read [`AGENTS.md`](./AGENTS.md) before contributing. It is the single rules file for this
repository: the English-only language policy, the code-comment rules, the document templates,
the mandatory cross-references, the folder and index conventions, and the identifier namespaces.


## Directory Layout

```
b-eval/
├── AGENTS.md                # The single operating-rules file: language, comments, docs, workflow
├── Docs/
│   ├── README.md            # Index of all documentation
│   ├── Feature-Template.md  # Template for Docs/Features/
│   ├── Fix-Template.md      # Template for Docs/Fix/
│   ├── Features/            # The functional design — one category per folder
│   │   ├── index.md         # Index of every feature document
│   │   ├── platform/        # Scope, positioning, delivery plan
│   │   ├── task-system/     # Capability taxonomy, task authoring
│   │   ├── agent-integration/  # Agent adapter protocol
│   │   ├── execution/       # Evaluation plans, orchestration
│   │   ├── scoring/         # Scorers and metrics
│   │   ├── reporting/       # Reports, comparison, leaderboard
│   │   ├── security/        # Anti-cheat, quality assurance
│   │   └── experience/      # Journeys, screens, permissions
│   ├── Fix/                 # Bug fixes — one category per folder
│   │   └── index.md         # Index of every fix document
│   ├── Tech Design/         # TD-<nnn>-<slug>.md
│   └── Test Report/         # TR-<nnn>-<slug>.md
└── vendor/                  # Read-only mirrors of third-party reference projects (not versioned)
```

## Reference Projects

`vendor/` holds 24 high-star open-source projects related to agent evaluation, cloned read-only for design comparison.
Selection criteria and per-project takeaways: [Docs/README.md §2](./Docs/README.md).

- **Benchmarks**: SWE-bench, τ²-bench, τ-bench, OSWorld, WebArena, VisualWebArena, BrowserGym, AgentBench, AgentBoard, AgentDojo, MCPMark, VitaBench, MemoryAgentBench, TheAgentCompany, ClawBench, terminal-bench, MLE-bench, AgentCompass
- **Resource indexes**: awesome-evals, ai-agent-benchmark-compendium, agent evaluation survey companion
- **Evaluation and execution frameworks**: Inspect AI, OpenAI Evals, SWE-agent, FastChat, promptfoo, DeepEval, AgentScope, agent-lightning
