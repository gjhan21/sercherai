# 股票与期货预测增强文档使用与线程交接说明

## 目的

这份文档用于让其他线程、其他协作者或后续会话可以直接接手本项目，不需要重新口头解释背景。

本主题的正式文档集合是：

- 总纲：`/Users/gjhan21/cursor/sercherai/docs/superpowers/specs/2026-03-28-stock-futures-forecast-roadmap.md`
- `L1`：`/Users/gjhan21/cursor/sercherai/docs/superpowers/specs/2026-03-28-stock-futures-forecast-l1-design.md`
- `L2`：`/Users/gjhan21/cursor/sercherai/docs/superpowers/specs/2026-03-28-stock-futures-forecast-l2-design.md`
- `L3`：`/Users/gjhan21/cursor/sercherai/docs/superpowers/specs/2026-03-28-stock-futures-forecast-l3-design.md`
- `L3` 实施计划：`/Users/gjhan21/cursor/sercherai/docs/superpowers/plans/2026-03-29-stock-futures-forecast-l3-implementation.md`
- Admin：`/Users/gjhan21/cursor/sercherai/docs/superpowers/specs/2026-03-28-stock-futures-forecast-admin-config-design.md`
- 线程说明：本文件

## 当前执行结论

当前结论已经固定：

1. `L1` 主线与后置 admin 轻量承接已落地到 `main`
2. `L2` 主线与复用式 admin 可见性也已落地到 `main`
3. `L3 MVP` 已在当前工作区完成实现并通过定向验证，覆盖：
   - 异步 run 队列
   - 报告资产与步骤日志
   - 学习回写与质量摘要
   - client 读链接入
   - admin `Forecast Lab` 工作台
4. `L3` 仍然只是增强层，目标是增强现有股票推荐与期货策略体系，不另造一套平行主系统
5. 借鉴 `MiroFish` 的是预测闭环、角色化推演与长期学习方法，不是整套社交仿真产品形态
6. admin 已允许新增独立 `Forecast Lab` 工作台，但它只承接 `L3` 运行与复核，不接管主推荐链

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

1. 这次改动是否仍然围绕既定 `L1 / L2 / L3` 路线，而不是另起平行预测主链？
2. explanation / history 的增强字段是否保持非破坏性扩展，而不是改坏旧读链？
3. 这次实现是否误改了排序、权重、发布审批主链？
4. 新增页面、接口或持久化表是否只限 `L3` 已批准范围？
5. 没有 `L3` 结果时，前台和后台是否还能优雅回退到 `L2`？

如果其中任意一项回答为“是，超出边界”，需要先停下来重新确认范围。

## 当前实现状态（2026-03-29）

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
8. `L2` relationship snapshot、bull/base/bear 三情景、agent opinions/veto 已落地
9. client 已承接 `L2` 三情景与关系快照展示
10. admin 已承接 `L2` 开关、情景摘要、veto 只读展示
11. `L3` 后端主链已落地：
   - `/Users/gjhan21/cursor/sercherai/backend/internal/growth/model/strategy_forecast_l3.go`
   - `/Users/gjhan21/cursor/sercherai/backend/internal/growth/dto/strategy_forecast_l3.go`
   - `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/strategy_forecast_l3_repo.go`
   - `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/strategy_forecast_l3_orchestrator.go`
   - `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/strategy_forecast_l3_report_builder.go`
   - `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/strategy_forecast_l3_learning.go`
   - `/Users/gjhan21/cursor/sercherai/backend/internal/growth/handler/user_growth_handler_forecast_l3.go`
   - `/Users/gjhan21/cursor/sercherai/backend/internal/growth/handler/admin_growth_handler_forecast_l3.go`
12. explanation / history 已接入 `L3` 摘要与报告引用：
   - `/Users/gjhan21/cursor/sercherai/backend/internal/growth/model/strategy_client_explanation.go`
   - `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/strategy_client_explanation.go`
13. client 已承接 `L3`：
   - `/Users/gjhan21/cursor/sercherai/client/src/api/forecast.js`
   - `/Users/gjhan21/cursor/sercherai/client/src/views/ForecastRunView.vue`
   - `/Users/gjhan21/cursor/sercherai/client/src/views/StrategyView.vue`
   - `/Users/gjhan21/cursor/sercherai/client/src/views/RecommendationArchiveView.vue`
14. admin 已承接 `L3`：
   - `/Users/gjhan21/cursor/sercherai/admin/src/views/ForecastLabView.vue`
   - `/Users/gjhan21/cursor/sercherai/admin/src/views/SystemConfigsView.vue`
   - `/Users/gjhan21/cursor/sercherai/admin/src/views/MarketCenterView.vue`
   - `/Users/gjhan21/cursor/sercherai/admin/src/lib/forecast-admin.js`

### 当前验证状态

以下命令在当前工作区已经通过，可作为其他线程接手前的事实基线：

1. `cd /Users/gjhan21/cursor/sercherai/backend && go test ./internal/growth/repo -run 'ForecastL3|ExecuteForecastL3Run|BuildForecastL3Report|ForecastL3Quality|LearningRecord|BuildStockStrategyExplanationAttachesDeepForecastSummary|GetStockRecommendationVersionHistoryUsesBackfilledLocalContexts|GetStockRecommendationVersionHistoryAttachesDeepForecastSummary'`
2. `cd /Users/gjhan21/cursor/sercherai/backend && go test ./internal/growth/handler -run 'ForecastL3'`
3. `cd /Users/gjhan21/cursor/sercherai/client && node --test src/lib/strategy-version.test.js src/views/forecast-run-view.test.js`
4. `cd /Users/gjhan21/cursor/sercherai/client && npm run build`
5. `cd /Users/gjhan21/cursor/sercherai/admin && node --test src/lib/forecast-admin.test.js src/lib/admin-navigation.test.js src/views/system-configs-view.test.js src/views/market-center-view.test.js src/views/forecast-lab-view.test.js`
6. `cd /Users/gjhan21/cursor/sercherai/admin && npm run build`

### 关键提交

- `44b38ff Merge branch 'codex/stock-futures-forecast-l1'`
- `6c8ca7e feat: add optional admin forecast controls`
- `3e8c383 feat: add l2 relationship snapshot contracts`
- `d61b86c feat: add stable l2 scenario snapshots`
- `1b29758 feat: add l2 agent opinions and veto meta`
- `303c529 feat: wire l2 scenarios into insight and history`
- `aef4465 feat: surface l2 scenarios in client views`
- `f5cd039 feat: add l2 forecast visibility in admin`

### 继续推进时的事实

- 当前已经存在独立 admin `Forecast Lab`，后续线程不要再重复造第二个 `L3` 管理入口
- 当前不需要再继续补 `L1 / L2` 主线能力，除非是 bugfix、联调或展示收口
- 如果继续推进新功能，应优先在已落地的 `L3 MVP` 基线上补联调、可用性和质量校准，而不是重新解释 `L1 / L2`
- `/Users/gjhan21/cursor/sercherai/admin/src/router/index.js:155-164` 的权限跳转问题仍是独立问题，不属于本主题范围

## 其他线程的推荐阅读顺序

### 如果是做总览判断

先读：

1. `/Users/gjhan21/cursor/sercherai/docs/superpowers/specs/2026-03-28-stock-futures-forecast-roadmap.md`
2. `/Users/gjhan21/cursor/sercherai/docs/superpowers/specs/2026-03-28-stock-futures-forecast-thread-handoff.md`
3. `/Users/gjhan21/cursor/sercherai/docs/superpowers/specs/2026-03-28-stock-futures-forecast-admin-config-design.md`

### 如果是准备查看当前已落地基线

先读：

1. `/Users/gjhan21/cursor/sercherai/docs/superpowers/specs/2026-03-28-stock-futures-forecast-roadmap.md`
2. `/Users/gjhan21/cursor/sercherai/docs/superpowers/specs/2026-03-28-stock-futures-forecast-l1-design.md`
3. `/Users/gjhan21/cursor/sercherai/docs/superpowers/specs/2026-03-28-stock-futures-forecast-l2-design.md`
4. `/Users/gjhan21/cursor/sercherai/docs/superpowers/specs/2026-03-28-stock-futures-forecast-admin-config-design.md`
5. `/Users/gjhan21/cursor/sercherai/backend/internal/growth/model/strategy_client_explanation.go`
6. `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/strategy_client_explanation.go`
7. `/Users/gjhan21/cursor/sercherai/client/src/lib/strategy-version.js`
8. `/Users/gjhan21/cursor/sercherai/admin/src/lib/forecast-admin.js`

### 如果是准备开发 L3

先读：

1. `/Users/gjhan21/cursor/sercherai/docs/superpowers/specs/2026-03-28-stock-futures-forecast-roadmap.md`
2. `/Users/gjhan21/cursor/sercherai/docs/superpowers/specs/2026-03-28-stock-futures-forecast-l2-design.md`
3. `/Users/gjhan21/cursor/sercherai/docs/superpowers/specs/2026-03-28-stock-futures-forecast-admin-config-design.md`
4. `/Users/gjhan21/cursor/sercherai/docs/superpowers/specs/2026-03-28-stock-futures-forecast-l3-design.md`
5. `/Users/gjhan21/cursor/sercherai/docs/superpowers/plans/2026-03-29-stock-futures-forecast-l2-implementation.md`
6. `/Users/gjhan21/cursor/sercherai/docs/superpowers/plans/2026-03-29-stock-futures-forecast-l3-implementation.md`
7. `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/strategy_client_explanation.go`
8. `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/strategy_forecast_l3_repo.go`
9. `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/strategy_forecast_l3_orchestrator.go`
10. `/Users/gjhan21/cursor/sercherai/client/src/views/StrategyView.vue`
11. `/Users/gjhan21/cursor/sercherai/client/src/views/ForecastRunView.vue`
12. `/Users/gjhan21/cursor/sercherai/admin/src/views/MarketCenterView.vue`
13. `/Users/gjhan21/cursor/sercherai/admin/src/views/ForecastLabView.vue`

### 如果是评估 L3 是否该启动

先读：

1. 总纲
2. `L1 / L2` 当前实现情况
3. Admin 配置规划
4. 对应阶段 spec

不要绕开已落地的 `L1 / L2` 基线去另造平行系统。

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
- 不绕开现有 `Forecast Lab` 再新开另一套后台中心

## 推荐给其他线程的启动方法

### 方法一：让其他线程先看文档再做计划

可以直接把下面这段话发给新线程：

```text
请先阅读以下文档并基于文档开展工作，不要自行改题：
1. /Users/gjhan21/cursor/sercherai/docs/superpowers/specs/2026-03-28-stock-futures-forecast-roadmap.md
2. /Users/gjhan21/cursor/sercherai/docs/superpowers/specs/2026-03-28-stock-futures-forecast-l1-design.md
3. /Users/gjhan21/cursor/sercherai/docs/superpowers/specs/2026-03-28-stock-futures-forecast-thread-handoff.md
4. /Users/gjhan21/cursor/sercherai/docs/superpowers/specs/2026-03-28-stock-futures-forecast-admin-config-design.md

当前主题已经进入 L3 MVP 阶段。请特别遵守以下硬边界：
- L3 只做增强层，不能直接改排序、权重和发布审核主链
- explanation / version-history 只允许做非破坏性读链扩展
- 已有 Forecast Lab 是唯一后台工作台，不要再起第二套入口
- 外部 MiroFish 仍然只作为方法参考，不是当前事实源

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

当前如非明确要求实现代码，请先做设计审查、质量审查或联调风险评估，且不要改动既有主推荐链边界。
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

1. 由当前线程或新线程先基于本文件确认 `L3 MVP` 已落地的事实源和验证基线
2. 后续优先做联调、可用性优化、权限裁剪、质量口径校验和手工验收
3. 如果要继续扩展，只能在现有 `Forecast Lab + explanation/history + ForecastRunView` 基线上增量推进
