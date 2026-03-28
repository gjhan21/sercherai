# 股票与期货预测增强文档使用与线程交接说明

## 目的

这份文档用于让其他线程、其他协作者或后续会话可以直接接手本项目，不需要重新口头解释背景。

本主题的正式文档集合是：

- 总纲：`/Users/gjhan21/cursor/sercherai/docs/superpowers/specs/2026-03-28-stock-futures-forecast-roadmap.md`
- `L1`：`/Users/gjhan21/cursor/sercherai/docs/superpowers/specs/2026-03-28-stock-futures-forecast-l1-design.md`
- `L2`：`/Users/gjhan21/cursor/sercherai/docs/superpowers/specs/2026-03-28-stock-futures-forecast-l2-design.md`
- `L3`：`/Users/gjhan21/cursor/sercherai/docs/superpowers/specs/2026-03-28-stock-futures-forecast-l3-design.md`
- Admin：`/Users/gjhan21/cursor/sercherai/docs/superpowers/specs/2026-03-28-stock-futures-forecast-admin-config-design.md`
- 线程说明：本文件

## 当前执行结论

当前结论已经固定：

1. 先开发 `L1`
2. `L2 / L3` 只先冻结设计，不提前实现
3. 目标是增强现有股票推荐与期货策略体系，不另造一套平行主系统
4. 借鉴 `MiroFish` 的是预测闭环思想，不是整套社交仿真产品形态
5. admin 侧优先复用现有后台模块，不在 `L1` 新开独立预测增强后台中心
6. `L1` 主线与后置 admin 承接均已落地到 `main`

## 已定硬边界

以下内容已经确定，其他线程不要再自行解释：

1. `L1` 的 `memory_feedback` 与 `confidence calibration` 必须被 explanation 生成过程消费。
2. `L1` 中上述能力只作为 advisory input，不直接改动排序、发布审批或组合权重。
3. `L1` 首版只覆盖 4 个入口：
   - 股票推荐 `insight`
   - 股票推荐 `version-history`
   - 期货策略 `insight`
   - 期货策略 `version-history`
4. `L1` 首版默认不新增新的持久化业务表作为前提。
5. `L1` 首版只走现有 explanation 生成链与现有评估回补消费链，不另起独立异步研究系统。

## 一页式边界检查表

任何新线程准备开工前，必须先回答这 5 个问题：

1. 这次改动是否仍然只落在 `L1` 范围？
2. `memory_feedback` 是否被真正消费，而不只是新增返回字段？
3. 这次实现是否误改了排序、权重、发布审批主链？
4. 这次实现是否超出了 4 个首版入口？
5. 是否引入了新的页面、接口或持久化业务表作为前提？

如果其中任意一项回答为“是，超出边界”，需要先停下来重新确认范围。

## 当前实现状态（2026-03-28）

### 已完成

1. `L1` explanation contract 与 history contract 扩展
2. 股票/期货研究编排、active/historical thesis、watch signals
3. `memory_feedback` 真正参与 explanation 生成
4. advisory-only `confidence_calibration`
5. 四个真实入口打通：
   - 股票 `insight`
   - 股票 `version-history`
   - 期货 `insight`
   - 期货 `version-history`
6. client 已在现有页面承接：
   - `/Users/gjhan21/cursor/sercherai/client/src/views/StrategyView.vue`
   - `/Users/gjhan21/cursor/sercherai/client/src/views/RecommendationArchiveView.vue`
   - `/Users/gjhan21/cursor/sercherai/client/src/views/HomeView.vue`
7. optional admin 已按轻量嵌入方式落地：
   - `/Users/gjhan21/cursor/sercherai/admin/src/views/SystemConfigsView.vue`
   - `/Users/gjhan21/cursor/sercherai/admin/src/views/MarketCenterView.vue`
   - `/Users/gjhan21/cursor/sercherai/admin/src/views/ReviewCenterView.vue`

### 关键提交

- `44b38ff Merge branch 'codex/stock-futures-forecast-l1'`
- `6c8ca7e feat: add optional admin forecast controls`

### 继续推进时的事实

- 当前不需要再补独立 admin 规划入口
- 如果继续做 `L1` 收尾，优先做文档、联调或远端集成，不要把范围扩成 `L2 / L3`
- `/Users/gjhan21/cursor/sercherai/admin/src/router/index.js:155-164` 的权限跳转问题仍是独立问题，不属于本主题范围

## 其他线程的推荐阅读顺序

### 如果是做总览判断

先读：

1. `/Users/gjhan21/cursor/sercherai/docs/superpowers/specs/2026-03-28-stock-futures-forecast-roadmap.md`
2. `/Users/gjhan21/cursor/sercherai/docs/superpowers/specs/2026-03-28-stock-futures-forecast-thread-handoff.md`
3. `/Users/gjhan21/cursor/sercherai/docs/superpowers/specs/2026-03-28-stock-futures-forecast-admin-config-design.md`

### 如果是准备开发 L1

先读：

1. `/Users/gjhan21/cursor/sercherai/docs/superpowers/specs/2026-03-28-stock-futures-forecast-roadmap.md`
2. `/Users/gjhan21/cursor/sercherai/docs/superpowers/specs/2026-03-28-stock-futures-forecast-l1-design.md`
3. `/Users/gjhan21/cursor/sercherai/docs/superpowers/specs/2026-03-28-stock-futures-forecast-admin-config-design.md`
4. `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/strategy_engine_context_repo.go`
5. `/Users/gjhan21/cursor/sercherai/backend/internal/growth/model/strategy_client_explanation.go`
6. `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/strategy_client_explanation.go`
7. `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/stock_selection_run_evaluation_backfill.go`
8. `/Users/gjhan21/cursor/sercherai/client/src/api/market.js`

### 如果是评估 L2 / L3 是否该启动

先读：

1. 总纲
2. `L1` 当前实现情况
3. Admin 配置规划
4. 对应阶段 spec

不要跳过 `L1` 直接实现 `L2 / L3`。

## 其他线程必须遵守的边界

### 允许做的事

- 优先增强现有 `strategy-engine / insight / evaluation / version-history`
- 做非破坏性字段扩展
- 在 `backend/internal/growth` 内补内部编排与校准能力
- 为前端现有页面增加更强 explanation 展示

### 不允许擅自扩的范围

- 不先重做全站页面
- 不先做聊天式 AI 问答主入口
- 不直接把 MiroFish 整套产品接进来
- 不新增和当前主推荐链平行的第二套推荐主系统
- 不跳过 `L1` 直接做深度模拟
- 不在 `L1` 新开独立预测增强后台中心

## 推荐给其他线程的启动方法

### 方法一：让其他线程先看文档再做计划

可以直接把下面这段话发给新线程：

```text
请先阅读以下文档并基于文档开展工作，不要自行改题：
1. /Users/gjhan21/cursor/sercherai/docs/superpowers/specs/2026-03-28-stock-futures-forecast-roadmap.md
2. /Users/gjhan21/cursor/sercherai/docs/superpowers/specs/2026-03-28-stock-futures-forecast-l1-design.md
3. /Users/gjhan21/cursor/sercherai/docs/superpowers/specs/2026-03-28-stock-futures-forecast-thread-handoff.md
4. /Users/gjhan21/cursor/sercherai/docs/superpowers/specs/2026-03-28-stock-futures-forecast-admin-config-design.md

当前只允许推进 L1。请特别遵守以下硬边界：
- memory_feedback / confidence calibration 必须被消费，但不能直接改排序和发布链
- 首版只覆盖股票与期货的 insight / version-history 四个入口
- 不新增新的持久化业务表、页面或独立异步研究系统作为前提
- admin 侧优先复用现有后台模块，不新开独立预测增强后台中心

请先结合这些文档和现有代码，输出一份实施计划，再开始开发。
```

### 方法二：如果是做 L2 / L3 预研

可以发：

```text
请先阅读：
1. /Users/gjhan21/cursor/sercherai/docs/superpowers/specs/2026-03-28-stock-futures-forecast-roadmap.md
2. /Users/gjhan21/cursor/sercherai/docs/superpowers/specs/2026-03-28-stock-futures-forecast-l1-design.md
3. /Users/gjhan21/cursor/sercherai/docs/superpowers/specs/2026-03-28-stock-futures-forecast-l2-design.md
4. /Users/gjhan21/cursor/sercherai/docs/superpowers/specs/2026-03-28-stock-futures-forecast-l3-design.md
5. /Users/gjhan21/cursor/sercherai/docs/superpowers/specs/2026-03-28-stock-futures-forecast-admin-config-design.md
6. /Users/gjhan21/cursor/sercherai/docs/superpowers/specs/2026-03-28-stock-futures-forecast-thread-handoff.md

当前不要实现代码，只做设计审查或风险评估，且不能改变 L1 优先级。
```

### 方法三：如果是做代码审查

可以发：

```text
请根据以下文档审查当前实现是否偏离方案：
1. /Users/gjhan21/cursor/sercherai/docs/superpowers/specs/2026-03-28-stock-futures-forecast-roadmap.md
2. /Users/gjhan21/cursor/sercherai/docs/superpowers/specs/2026-03-28-stock-futures-forecast-l1-design.md
3. /Users/gjhan21/cursor/sercherai/docs/superpowers/specs/2026-03-28-stock-futures-forecast-thread-handoff.md

重点检查是否出现越过 L1 范围提前做 L2 / L3、是否破坏现有接口、是否引入了与当前系统平行的新主链。
```

## 给其他线程的事实源提醒

本主题的“系统已具备哪些真实能力”，不要靠猜，统一以下列代码为准：

- `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/strategy_engine_context_repo.go`
- `/Users/gjhan21/cursor/sercherai/backend/internal/growth/model/strategy_client_explanation.go`
- `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/strategy_client_explanation.go`
- `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/stock_selection_run_evaluation_backfill.go`
- `/Users/gjhan21/cursor/sercherai/client/src/api/market.js`

如果需要确认 `L1` 到底有哪些硬边界，以 `L1` spec 的 `L1 Authoritative Contract` 为准。

如果需要参考 `MiroFish`，请把它当“方法参考”，不是当前项目的事实源。

## 文档维护规则

后续如果进入开发阶段：

- 总纲只更新全局方向，不记录实现细节
- `L1 / L2 / L3` 各自只维护本阶段内容
- 线程交接说明只维护“怎么接手”和“当前优先级”
- 不要把临时讨论直接写进正式 spec

## 当前建议

现在最合适的下一步是：

1. 由当前线程或新线程基于 `L1` spec 写实施计划
2. 审查计划是否严格落在 `L1`
3. 再进入 `L1` 实际开发
