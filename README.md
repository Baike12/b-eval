# b-eval

Agent 测评平台（Agent Evaluation Platform）。

## 项目状态

| 阶段 | 内容 | 位置 |
|---|---|---|
| 设计方案 | 功能层面设计（不含技术实现） | [`Docs/`](./Docs/README.md) |
| 技术方案 | 待编写 | — |
| 实现 | 待启动 | — |

## 目录结构

```
b-eval/
├── Docs/          # 设计方案（功能层面）
└── vendor/        # 第三方开源参考项目的只读副本，不纳入版本管理
```

## 参考项目

`vendor/` 下克隆了 24 个高星开源 Agent 测评相关项目，供设计方案比对使用。
选型依据与逐项借鉴点见 [Docs/README.md 的调研结论](./Docs/README.md#附市面开源项目调研结论)。

按类别：

- **基准（benchmark）**：SWE-bench、τ²-bench、τ-bench、OSWorld、WebArena、VisualWebArena、BrowserGym、AgentBench、AgentBoard、AgentDojo、MCPMark、VitaBench、MemoryAgentBench、TheAgentCompany、ClawBench、terminal-bench、MLE-bench、AgentCompass
- **资源索引**：awesome-evals、ai-agent-benchmark-compendium、Agent 评测综述配套
- **评测与执行框架**：Inspect AI、OpenAI Evals、SWE-agent、FastChat、promptfoo、DeepEval、AgentScope、agent-lightning
