# Strategy Engine Vibe Coding 开发总览

最后更新: 2026-03-25  
状态: 阶段8已完成，阶段9已完成（双资产智能策略研究平台 V2 收官）

## 文档目的

本目录用于沉淀 `strategy-engine` 相关能力的阶段规划、实现边界、验收标准与文档真相源。

当前口径已经从“股票/期货独立策略服务”继续收口为：

- Go 后端负责多源市场数据接入、上下文构建、持久化、Admin API 与兼容旧消费链路
- `strategy-engine` 负责股票与期货的研究链、报告、发布 payload 与结构化解释产物
- `strategy-graph` 负责 run 级图快照、图谱查询与图服务降级能力
- Admin 侧形成两个一级研究模块：`智能选股`、`智能期货`
- Client 侧继续保留旧消费契约，并通过共享 helper 统一解释、版本差异与评估摘要

## 开发前必读

每次继续迭代前，按顺序阅读：

1. `docs/strategy-engine/README.md`
2. `docs/strategy-engine/开发工作流.md`
3. 当前正在执行的阶段文档
4. 与该阶段直接关联的代码入口或测试

## 阶段总表

| 阶段 | 文档 | 目标 | 状态 |
| --- | --- | --- | --- |
| 阶段0 | `docs/strategy-engine/阶段0-服务边界与真相源.md` | 固化边界、目录、API、阶段规划 | Done |
| 阶段1 | `docs/strategy-engine/阶段1-服务骨架与任务协议.md` | 建立独立服务骨架、任务协议和基础可观测性 | Done |
| 阶段2 | `docs/strategy-engine/阶段2-股票MVP.md` | 跑通股票选股 MVP：输入、筛选、风险、报告 | Done |
| 阶段3 | `docs/strategy-engine/阶段3-期货MVP.md` | 跑通期货策略 MVP：方向、价位、风险、报告 | Done |
| 阶段4 | `docs/strategy-engine/阶段4-场景模拟与多代理.md` | 引入场景模拟、多代理协同与否决机制 | Done |
| 阶段5 | `docs/strategy-engine/阶段5-报告发布与复盘.md` | 完成发布适配、报告体系与历史复盘 | Done |
| 阶段6 | `docs/strategy-engine/阶段6-后台策略管理协同.md` | 让后台真正管理 seeds / agents / scenarios / publish policy | Done |
| 阶段7 | `docs/strategy-engine/阶段7-客户端策略解释展示.md` | 让客户端完整展示种子输入、模拟过程与结论解释 | Done |
| 阶段8 | `docs/strategy-engine/阶段8-真实数据接入与真相源增强.md` | 接入真实行情/资讯数据，并增强统一真相源与数据新鲜度摘要 | Done |
| 阶段9 | `docs/strategy-engine/阶段9-股票智能量化选股系统与后台独立模块.md` | 收口为股票 + 期货双资产研究平台 V2，并完成图谱、审核、评估与统一解释闭环 | Done |

## 当前已完成的主能力

### 研究与发布底座

- `strategy-engine` 已完成股票与期货双资产研究链，包含上下文读取、候选筛选、结构化 evidence、评估摘要、发布 payload 与 explanation 产物
- `strategy-graph` 已作为独立内部服务接入，支持 run 图快照写入、快照读取、按实体查询一跳/两跳子图
- Go 后端已形成 `strategy-engine -> publish -> 本地快照 -> explanation/version-history` 的兼容闭环

### 真实数据与真相源

- 股票与期货上下文已默认优先消费 Go 后端统一真相源，而不是服务内置样本
- 期货上下文已补到期限结构、跨期价差、仓单/库存及其结构维度
- Admin 已具备数据源状态、质量摘要、真相源重建与同步入口
- Client 与 Admin 已能读取 `selected_trade_date / price_source / warnings / graph_summary` 等稳定摘要

### 双资产后台研究平台

- Admin 已新增两个一级菜单：`智能选股`、`智能期货`
- 股票与期货都已形成独立模块化后台：总览、运行中心、模板、配置、规则、因子、图谱、角色、候选、评估、审核
- 图谱页不直接连图服务，统一走 Go 后端代理接口
- 股票与期货都保留人工审核发布为默认，发布后继续回写旧消费链路

### 客户端统一解释

- `client/src/lib/strategy-version.js` 已成为股票/期货 explanation 的共享 helper
- 股票与期货已统一消费 `market_regime`、`portfolio_role`、`graph_summary`、`evidence_cards`、`version_diff`、`evaluation_meta`
- 老记录在缺少新字段时仍保持 fallback，不要求同步改旧客户端契约

## 当前文档结构

- 阶段8收官文档：`docs/strategy-engine/阶段8-真实数据接入与真相源增强.md`
- 阶段9最终收官文档：`docs/strategy-engine/阶段9-股票智能量化选股系统与后台独立模块.md`
- 后续专题总路线图：`docs/strategy-engine/专题路线图-多源市场与研究治理平台.md`
- 专题A正式文档：`docs/strategy-engine/专题A-多源市场数据统一模型与供应商治理.md`
- 专题A / 专题E 第一轮开发清单：`docs/strategy-engine/专题A-第一轮-开发启动清单.md`、`docs/strategy-engine/专题E-第一轮-开发启动清单.md`
- 阶段9第一轮历史蓝图：`docs/strategy-engine/阶段9-第一轮-自动选股与后台骨架.md`
- 阶段9第一轮历史启动清单：`docs/strategy-engine/阶段9-第一轮-开发启动清单.md`

其中阶段9第一轮相关文档继续保留，但仅作为“股票单资产首轮落地”的历史参考，不再作为当前阶段9验收真相源。

当前专题化迭代里，`专题E` 已推进到 `E4` 深化第五刀：统一消息中心现已支持从审计事件深链到发布策略、种子集、代理配置与场景模板，旧操作日志也开始复用同一对象跳转规则。

## 后续专题

阶段8、阶段9已经收官，后续继续迭代时默认从专题化能力切入，而不是回到阶段定义层反复漂移。当前明确保留到后续专题的方向包括：

- 更大的多源市场数据统一模型与供应商治理
- 更细的股票事件图谱、行业/主题事件真相源
- 更重的期货研究因子、交割地/品牌/等级深层画像
- 更复杂的回测、cohort、实验与自动化运营平台
- 对旧入口和旧页面的进一步收缩或替换

这些方向的总顺序与专题拆解，统一以：

- `docs/strategy-engine/专题路线图-多源市场与研究治理平台.md`

作为后续真相源。

当前第一个已正式展开的后续专题为：

- `docs/strategy-engine/专题A-多源市场数据统一模型与供应商治理.md`
- `docs/strategy-engine/专题A-第一轮-开发启动清单.md`

## 统一验收口径

每个阶段标记为 `Done` 时，至少满足：

- 有完整的输入、处理、输出闭环
- 对外接口、文档、页面和测试口径一致
- 兼容旧路径，不破坏现有客户端和后台稳定能力
- 关键测试与构建命令可通过
- 阶段文档已回写为当前真相源，而不是只保留中间过程
