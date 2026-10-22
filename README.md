# b-eval

Agent evaluation platform.

Give it a batch of task data, point it at an agent under test, and it runs the agent over
the data, judges every trial on its outcome, resolves the numbers from one metric registry
and reports the result with its confidence attached.

```
task data  →  agent under test  →  verdict per trial  →  metrics  →  report
```

## Status

| Stage | Content | Where |
|---|---|---|
| Functional design | What the system does | [`Docs/Features/`](./Docs/Features) |
| Delivery documents | Features, fixes, technical design, test reports | [`Docs/Features/index.md`](./Docs/Features/index.md), [`Docs/Fix/index.md`](./Docs/Fix/index.md) |
| Technical design | Architecture and implementation | [`Docs/Tech Design/`](<./Docs/Tech Design/>) |
| Implementation | The M0 vertical slice — Go backend, React frontend, quality gates | [`cmd/`](./cmd), [`internal/`](./internal), [`web/`](./web) |
| Deployment | The container deployment to the target host | [`deploy/`](./deploy), [`scripts/deploy.sh`](./scripts/deploy.sh) |

The M0 slice is verified by [TR-003 M0 Vertical Slice](<./Docs/Test Report/TR-003-m0-vertical-slice.md>).

## Running

```bash
# Build the frontend bundle and the binary with it embedded.
bash scripts/build.sh

# Write the demonstration task set and two agents into ./data.
go run ./cmd/b-eval seed -data data

# Start the platform.
go run ./cmd/b-eval serve -addr 127.0.0.1:8091 -data data
```

The API is under `/api`; the browser interface is at the root.

## Deploying

The platform runs as a container on the target host and is served under the `/eval/` prefix by
the host's existing nginx. The binary is cross-compiled on the developer machine, the runtime
image is built on the server, and the container is published on the loopback interface.

```bash
# Name the deployment target. The address is deliberately not kept in the
# repository, so it has to be supplied by whoever deploys.
export B_EVAL_HOST=root@your-host

# Build the frontend and the linux/amd64 binary, ship it, and start the container.
bash scripts/deploy.sh --tag $(date -u +%Y%m%d-%H%M)

# Check that the deployed platform is alive.
ssh "$B_EVAL_HOST" 'curl -fsS http://127.0.0.1/eval/api/healthz'
```

The platform is at **http://<host>/eval/** on that host. The server needs neither a Go
toolchain nor a Node toolchain, and the transfer is the size of the binary rather than the
size of an image. The records live in `/opt/b-eval/data`, a bind mount, so they survive a
container that is replaced.

No security group rule is needed: the platform is reached on port 80, which the host's
security group already opens. The base path is a build parameter, because the bundle carries
the asset, API and route prefixes; `--base-path` sets it on both the bundle and the nginx
route.

The design is in [TD-005 Container Deployment](<./Docs/Tech Design/TD-005-container-deployment.md>)
and the verification is in [TR-004 Container Deployment](<./Docs/Test Report/TR-004-container-deployment.md>).

## The Quality Gate

```bash
# Run every gate.
go run ./cmd/b-eval gate -root .

# Run the staged pipeline.
bash scripts/gate.sh

# Prove that the gates can block.
bash scripts/gate-smoke.sh
```

Every gate produces a verdict, a classification and a recommended action, and appends an
evidence record to `.gates/evidence.jsonl`. The evidence file is only ever added to.

## Working Rules

Read [`AGENTS.md`](./AGENTS.md) before contributing. It is the single rules file for this
repository: the English-only language policy, the code-comment rules, the document templates,
the mandatory cross-references, the folder and index conventions, and the identifier namespaces.

## Directory Layout

```
b-eval/
├── AGENTS.md                # The single operating-rules file: language, comments, docs, workflow
├── cmd/b-eval/              # serve · gate · seed · version
├── internal/
│   ├── domain/              # The data model and the task-form contract
│   ├── adapter/             # The four-capability agent protocol, the HTTP and mock adapters
│   ├── scoring/             # The scorers and the three-state verdict
│   ├── metrics/             # The metric registry, the built-in metrics, the intervals
│   ├── runner/              # Plan execution, budgets, failure classification
│   ├── report/              # The single-run report
│   ├── store/               # The file-backed record store
│   ├── api/                 # The HTTP surface
│   ├── gate/                # The quality gate runner and the checks
│   └── web/                 # The embedded single-page application
├── web/                     # The frontend source: TypeScript and React
├── scripts/                 # build.sh · deploy.sh · gate.sh · gate-smoke.sh
├── deploy/                  # The container deployment: image, compose, host scripts, nginx
├── Docs/
│   ├── README.md            # Index of all documentation
│   ├── Feature-Template.md  # Template for Docs/Features/
│   ├── Fix-Template.md      # Template for Docs/Fix/
│   ├── Features/            # The functional design — one category per folder
│   │   ├── index.md         # Index of every feature document
│   │   ├── platform/        # Scope, positioning, delivery plan
│   │   ├── task-system/     # Capability taxonomy, task authoring, task forms
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
