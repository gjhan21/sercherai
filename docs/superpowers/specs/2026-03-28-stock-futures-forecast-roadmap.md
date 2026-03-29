# 股票与期货预测增强路线图与总体设计规范

## Summary

本组文档用于收口“股票分析预测 / 期货分析预测”能力的总体路线图，并明确 `L1 / L2 / L3` 三个阶段的范围、边界与交付顺序。

本路线图的目标不是再造一套独立推荐系统，而是在现有 `strategy-engine + insight + performance + version-history + evaluation` 体系之上，引入更强的研究编排、走势推演与后验学习能力，持续提升：

- 股票推荐的准确性
- 期货策略的方向判断质量
- 单标的后续走势推演能力
- 推荐理由与失效条件的可解释性
- 系统从历史结果中吸取经验的能力

## 当前系统基线

### 已有真实能力

当前系统并不是从零开始，已经具备以下基础：

- 股票推荐、详情、表现、解释、版本历史
- 期货策略、guidance、解释、版本历史、绩效摘要
- `StrategyClientExplanation` 的结构化解释字段
- `strategy-engine` 的股票与期货上下文构建
- `1 / 3 / 5 / 10 / 20` 日评估回补与表现追踪
- 资讯、研报、事件、相关内容承接链路

关键代码事实源：

- `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/strategy_engine_context_repo.go`
- `/Users/gjhan21/cursor/sercherai/backend/internal/growth/model/strategy_client_explanation.go`
- `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/strategy_client_explanation.go`
- `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/stock_selection_run_evaluation_backfill.go`
- `/Users/gjhan21/cursor/sercherai/client/src/api/market.js`

### 当前核心短板

当前系统更像“会推荐、会解释、会回看”，但还不够“会研究、会推演、会吸取错误”。具体缺口主要有四类：

1. 缺少“单标的研究编排层”，当前解释更多是结果拼装，还不是主动拆解式研究过程。
2. 缺少“当前有效理由 vs 历史失效理由”的清晰分层，容易让旧逻辑残留。
3. 缺少“多情景推演层”，当前更偏结论和理由，未来路径推演仍偏轻。
4. 缺少“后验记忆回灌”，评估结果存在，但还没有充分变成下一轮推荐的记忆输入。

## 对 MiroFish 的借鉴结论

本路线图参考 `MiroFish` 的思路，但不照搬其产品形态。最值得借鉴的是预测闭环，而不是整套社交仿真世界。

重点借鉴能力：

1. 子问题拆解式深度检索
2. 当前有效事实与历史失效事实分层
3. 多视角意见与共识 / veto 机制
4. 情景推演的配置化生成
5. 模拟或推演结果的记忆回灌
6. 受工具约束的研究生成流程

明确不照搬的部分：

- 不把社交世界模拟直接当作股票价格主预测器
- 不先做重型 Twitter / Reddit 世界仿真
- 不做过重的人设系统
- 不在全市场全量标的上直接跑深度模拟

## 阶段划分

### L1：研究编排增强层

目标是先提升“推荐准确率”和“解释质量”，不直接上重型模拟。

核心方向：

- 子问题检索与证据编排
- 有效理由 / 失效理由分层
- 基于真实评估结果的记忆回灌
- 置信度与风险边界校准
- 继续沿用现有 insight / version-history 接口承接

对应文档：

- `/Users/gjhan21/cursor/sercherai/docs/superpowers/specs/2026-03-28-stock-futures-forecast-l1-design.md`

### L2：轻量金融图谱与多情景推演层

目标是增强“后续走势推演能力”，让系统不只给结论，还能给路径和触发条件。

核心方向：

- 轻量金融知识图快照
- 牛 / 基 / 熊三情景推演
- 多角色观点分歧与 veto 机制
- 市场环境感知的推演模板
- 与现有 insight / version-history / 复盘链承接
- `archive / analysis` 只作为未来可承接面，非 `L2` 本阶段实现前提

对应文档：

- `/Users/gjhan21/cursor/sercherai/docs/superpowers/specs/2026-03-28-stock-futures-forecast-l2-design.md`

### L3：高价值标的深度模拟与长期学习层

目标是对少量高价值标的做深推演，形成真正差异化的预测壁垒。

核心方向：

- 选择性深度模拟
- 金融专用角色库
- 推演行为回灌长期记忆
- 异步报告与可审计研究日志
- 预测质量校准与长期学习体系

对应文档：

- `/Users/gjhan21/cursor/sercherai/docs/superpowers/specs/2026-03-28-stock-futures-forecast-l3-design.md`

## 总体实施原则

### 1. 先增强现有推荐链，不另起一套主系统

所有阶段都优先复用现有：

- 推荐与策略结果
- 上下文构建
- 解释结构
- 版本历史
- 评估回补
- 前台 insight 展示链

### 2. 先做能提升质量的中台机制，再做重产品形态

本路线图优先级固定为：

1. 研究编排
2. 失效事实管理
3. 记忆回灌
4. 多情景推演
5. 深度模拟

### 3. MiroFish 只作为方法参考，不是必须全量接入对象

L1 不要求直接集成 MiroFish 引擎。  
L2 才开始引入图谱化和情景模板思想。  
L3 才考虑选择性深度模拟或外部引擎联动。

### 4. 股票与期货共享框架，但不强行共享全部细节

股票更强调：

- 公司 / 行业 / 主题 / 事件 / 资金 / 估值 / 财报

期货更强调：

- 供需 / 库存 / 基差 / 期限结构 / 宏观 / 产业链 / 政策

因此三阶段都采用“共享骨架 + 分品类模板”的设计。

## 各阶段依赖关系

### L1 是 L2 / L3 的地基

没有 L1 的研究编排和记忆回灌，L2 / L3 会变成“更会讲故事，但不一定更准”。

### L2 是 L3 的前提

没有 L2 的轻量金融图谱和情景推演模板，L3 的深度模拟就很难和现有系统稳定融合。

## 不在本路线图本阶段处理的内容

- 不重写现有 `client / newclient / uni-app` 全部页面
- 不先做全市场任意标的一步到位预测页
- 不把聊天式 AI 作为主交互入口
- 不新增 BFF 或大规模协议改造
- 不先做第三方支付、社区扩展、直播、聊天室等无关能力

## 成功标准

### L1 成功标准

- 推荐解释更结构化
- 旧逻辑失效点能明确呈现
- 历史评估结果已被消费为下一轮 explanation / confidence 输入
- 该反馈在 `L1` 中仅作为 advisory input，不直接改动排序、发布审批或组合权重
- 不破坏现有推荐与 insight 链路

### L2 成功标准

- 能输出可信的多情景路径
- 情景中包含触发条件、验证点、失效点
- 股票和期货都有各自合理的推演模板

### L3 成功标准

- 少量高价值标的具备明显更强的深度推演能力
- 推演结果具备日志、可解释性和复盘链
- 系统开始形成长期预测学习壁垒

## 当前执行结论

当前正式执行顺序固定为：

1. `L1` 已完成实施并落地到 `main`
2. `L2` 已完成实施并落地到 `main`
3. `L3` 仍以正式 spec 形态冻结设计，尚未进入实现
4. 后续线程统一按 handoff 文档衔接，并以 `L3` 计划为下一阶段入口

## 当前状态更新（2026-03-29）

- `L1` 已完成：研究编排、active/historical thesis、memory feedback 消费、confidence calibration、四个真实入口接入、client 展示承接、admin 轻量嵌入均已在主分支落地。
- `L2` 已完成：relationship snapshot、稳定 bull/base/bear 三情景、agent opinions/veto、四个真实入口接入、client 展示承接、admin 复用式可见性均已在主分支落地。
- 当前未完成阶段为 `L3`。
- 因此本路线图的下一步不再是解释是否要做 `L2`，而是决定是否开始 `L3` 的实施计划与开发。

配套线程交接说明见：

- `/Users/gjhan21/cursor/sercherai/docs/superpowers/specs/2026-03-28-stock-futures-forecast-thread-handoff.md`

配套 admin 配置规划见：

- `/Users/gjhan21/cursor/sercherai/docs/superpowers/specs/2026-03-28-stock-futures-forecast-admin-config-design.md`
