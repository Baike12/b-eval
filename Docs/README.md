# b-eval · Agent 测评项目设计方案

> 本目录是 **b-eval**（Agent 测评平台）的**功能层面**设计方案。
> 范围限定为「系统要提供什么能力、解决什么问题、用户如何使用」，
> **不包含**技术选型、架构分层、接口协议、存储格式、部署形态等实现细节——那部分另行成文。
> 凡涉及需要技术决策的位置，均以 `【待技术方案】` 标注。

## 文档索引

| 文档 | 内容 | 回答的问题 |
|---|---|---|
| [01-项目概述](./01-项目概述.md) | 背景、定位、目标、非目标、角色、核心概念 | 为什么要做、做成什么 |
| [02-能力维度与任务体系](./02-能力维度与任务体系.md) | 能力分类模型、任务元数据、任务集划分、难度分层、环境隔离 | 测什么、怎么组织题 |
| [03-评测执行与编排](./03-评测执行与编排.md) | 评测流程、Agent 接入协议、交互契约、预算、并发、续跑、轨迹 | 怎么跑起来 |
| [04-评分与指标体系](./04-评分与指标体系.md) | 判分器分类、组合规则、三态判定、指标公式、聚合口径 | 怎么打分、算什么 |
| [05-报告分析与对比](./05-报告分析与对比.md) | 产物清单、单次报告、对比视图、错误归因、排行榜 | 结果怎么看、怎么比 |
| [06-安全鲁棒性与质量保障](./06-安全鲁棒性与质量保障.md) | 防作弊、攻防评测、任务自检、Judge 校准、污染防护 | 怎么保证分数可信 |
| [07-用户流程与功能清单](./07-用户流程与功能清单.md) | 用户旅程、功能模块清单、界面与交互、权限模型 | 用户怎么用 |
| [08-演进路线与验收](./08-演进路线与验收.md) | 里程碑、验收标准、风险清单 | 分几步做完、怎么算做完 |

## 附：市面开源项目调研结论

调研方式：BrowserOS 打开 GitHub + Google 检索，GitHub 网页内嵌数据取 star 数（数据日期 2026-10-22）。
参考项目已全部克隆至本仓库 `vendor/`（只读参考副本，不纳入版本管理）。

### A. 直接可作为「被测对象」的基准（benchmark）

| 项目 | Star | 覆盖能力 | 对 b-eval 的可借鉴点 |
|---|---:|---|---|
| [SWE-bench](https://github.com/SWE-bench/SWE-bench) | 6068 | 代码 Agent | 任务=issue+仓库快照；fail-to-pass / pass-to-pass 双测试集；测试输出边界标记 + 真实执行证据交叉校验；防作弊规则清单 |
| [τ²-bench](https://github.com/sierra-research/tau2-bench) | 2284 | 工具-Agent-用户三方交互 | **「参考轨迹≠必需路径」**：用「重放参考轨迹得目标环境终态 → 比对哈希」代替逐步对比动作；pass^k 公式；基础设施错误与模型失败分离；冻结任务子集 + digest |
| [τ-bench](https://github.com/sierra-research/tau-bench) | 1472 | 工具调用 | 用户模拟器五策略（human/llm/react/verify/reflection）；显式停止 token；条件分支式指令承载「用户中途改主意」 |
| [OSWorld](https://github.com/xlang-ai/OSWorld) | 3229 | 桌面 GUI | **结果获取层 × 判定层二元解耦**（43 种 getter × 120 种 metric，任务文件里用字符串名声明）；复合判定 conj=and/or + 部分给分；未打分错误排除出分母 |
| [WebArena](https://github.com/web-arena-x/webarena) | 1636 | 网页 Agent | 任务模板化（190 模板 → 812 实例）；三类 evaluator 乘积组合；模板+参数化下发 |
| [VisualWebArena](https://github.com/web-arena-x/visualwebarena) | 490 | 多模态网页 | 三维难度标注（推理/视觉/综合）；图像目标；任务间依赖关系标注 |
| [BrowserGym](https://github.com/ServiceNow/BrowserGym) | 1411 | 网页环境抽象 | 环境三方法契约（setup/validate/teardown）；动作集=函数→自动生成描述/样例/工具 schema；观测模态正交组合 + 视口粒度裁剪 |
| [AgentBench](https://github.com/THUDM/AgentBench) | 3784 | 通用 8 环境 | 任务-环境-客户端三层解耦；统一状态机 + 错误分类占比统计；判分器唯一钩子（返回任意 JSON，框架不关心语义） |
| [AgentBoard](https://github.com/hkust-nlp/AgentBoard) | 450 | 通用 9 任务 | **成功率 + 进度率双轨**；6 维能力加权分；难度分层自动出图；分数变化点记录 → 逐步骤分数曲线 |
| [AgentDojo](https://github.com/ethz-spylab/agentdojo) | 948 | 攻防安全 | 攻击/防御注册表 + 可组合 pipeline；**用探针自动发现模型实际可见的注入点**；Utility / ASR 双轴报告 |
| [MCPMark](https://github.com/eval-sys/mcpmark) | 467 | MCP 生态 | 任务三件套（元信息 / 任务描述 / 独立判分脚本）；Setup→Execute→Verify→Cleanup 四阶段状态隔离；MCP server 版本钉死 |
| [VitaBench](https://github.com/meituan-longcat/vitabench) | 182 | 真实服务场景 | **rubric 条目化评分 + 滑动窗口**（10 轮窗 / 2 轮重叠，状态跨窗可推翻）；judge 提示词内置反作弊口径 |
| [MemoryAgentBench](https://github.com/HUST-AI-HYZ/MemoryAgentBench) | 477 | 长期记忆 | 「注入一次、多轮追问」；能力 × 上下文长度二维矩阵；同题换 Agent 的对照实验 |
| [TheAgentCompany](https://github.com/TheAgentCompany/TheAgentCompany) | 801 | 模拟公司协作 | **checkpoint 部分完成评分**（可插拔累加策略）；NPC 同事脚本；**判分代码对被测方不可见** |
| [ClawBench](https://github.com/TIGER-AI-Lab/ClawBench) | 1095 | 浏览器日常任务 | 两阶段评分（确定性拦截 → LLM 裁决）；拦截器天然防不可逆操作；「模型不行」与「环境挂了」分离归类 |
| [terminal-bench](https://github.com/laude-institute/terminal-bench) | 2605 | 终端任务 | 任务自带独立环境定义 + 独立测试环境 + 参考解；难度分层与人类耗时基线；数据集注册与版本指纹 |
| [MLE-bench](https://github.com/OpenAI/mle-bench) | 1783 | 机器学习工程 | 端到端真实工程产出判定 |
| [AgentCompass](https://github.com/open-compass/AgentCompass) | 170 | 统一评测基建 | **七层职责切分**（模型 / 基准 / 框架 / 环境 / 适配 / 运行时 / 分析器）；指标契约（binary_success / scalar）；可插拔分析器六分类 |
| [BenchFlow / awesome-evals](https://github.com/benchflow-ai/awesome-evals) | 994 | 资源索引 | 评测基础设施的分类学（框架 / 观测 / 裁判 / RL 环境 / 追踪标准）；benchmark 完整性检查清单（污染、饱和、标签错误、刷榜） |
| [ai-agent-benchmark-compendium](https://github.com/philschmid/ai-agent-benchmark-compendium) | 200 | 资源索引 | 四大类分类：函数调用与工具使用 / 通用助手与推理 / 编码与软件工程 / 计算机交互（GUI 与 Web） |
| [Agent 评测综述配套](https://github.com/Asaf-Yehudai/LLM-Agent-Evaluation-Survey) | 114 | 方法论 | 四大基础能力（规划 / 工具使用 / 自我反思 / 记忆）+ 四类应用域；三层视角：捕获 → 解读 → 信任 |

### B. 评测框架与执行框架

| 项目 | Star | 对 b-eval 的可借鉴点 |
|---|---:|---|
| [Inspect AI](https://github.com/UKGovernmentBEIS/inspect_ai) | 3058 | 评测六要素（任务 / 求解器 / 评分器 / 指标 / 模型 / 沙箱）；组件自描述 + 日志内可重建；预算六件套（消息/token/回合/时长/工作量/成本）；离线重打分；分数编辑审计链 |
| [OpenAI Evals](https://github.com/openai/evals) | 19626 | 「换掉被测系统」的抽象；评测集 + 评测类 + YAML 注册；命名含版本号 |
| [SWE-agent](https://github.com/princeton-nlp/SWE-agent) | 20607 | Agent 执行框架功能面：提示词模板体系、工具黑名单、上下文管理链、成本/调用次数预算、轨迹回放、钩子机制 |
| [FastChat / Chatbot Arena](https://github.com/lm-sys/FastChat) | 39549 | 众包人类偏好的对战式评测范式与 Elo 类排名 |
| [promptfoo](https://github.com/promptfoo/promptfoo) | 26245 | 提示词/Agent 回归测试的产品化形态与红队套件 |
| [DeepEval](https://github.com/confident-ai/deepeval) | 19020 | 断言式用例写法（断言作为一等公民） |
| [AgentScope](https://github.com/alibaba/AgentScope) | 33598 | 多 Agent 编排的消息级抽象 |
| [agent-lightning](https://github.com/microsoft/agent-lightning) | 18782 | 评测与训练的衔接（同一轨迹两用） |

### C. 调研中确认的空白点（b-eval 的机会）

1. **没有统一的多环境判分抽象**——OSWorld 的 getter/metric、MCPMark 的 verify.py、TheAgentCompany 的 checkpoint、AgentDojo 的 utility/security 各写各的，跨项目无法复用同一套评分器。
2. **没有统一的 Agent 接入协议**——τ-bench、VitaBench、MCPMark、AgentCompass 的 Agent 接口互不兼容，评测同一批任务要重写接入层。
3. **人工标注 rubric + LLM 判分 + 确定性校验三者融合**没有成熟范式——VitaBench 只用 LLM，TheAgentCompany 二选一，MCPMark 只用确定性。
4. **基础设施故障污染分数**——τ-bench v1 把异常记 0 分计入分母，v2 才分离；OSWorld 用 `unscored_errors` 排除；ClawBench 用 `adjusted_eligible`。三种互不兼容的口径。
5. **判分代码对被测方可见**——只有 TheAgentCompany 做了加密，主流基准普遍可被读取判分逻辑。
6. **题目生产成本高且无质检**——普遍缺「参考解能解 + 判分器可信 + 难度校准」的前置校验。

## 使用约定

- 每条功能需求给出**唯一编号**（如 `F-RUN-01`），便于后续技术方案与排期引用；编号总表见 [07-用户流程与功能清单](./07-用户流程与功能清单.md)。
- 所有指标定义给出**计算口径与分母**，不写「大概」「较好」这类描述。
- 文档中出现的开源项目均以 `vendor/<owner>__<repo>` 形式指代本地副本，便于查阅依据。
