# 阶段2：股票 MVP

最后更新: 2026-03-17
状态: Done

## 阶段目标

跑通股票选股最小闭环：输入市场种子 -> 生成候选 -> 风险过滤 -> 输出结构化报告 -> 结果可发布。

## 核心结果

- 股票种子加载可用
- 股票特征工程可用
- 股票 selector 可输出候选清单
- 每条候选都有原因、风险、失效条件
- 结果可映射回现有股票推荐数据结构

## 主要改动点

- `domain/seeds/market_seed_loader.py`
- `domain/features/stock_feature_factory.py`
- `domain/selectors/stock_selector.py`
- `domain/risk/portfolio_guard.py`
- `domain/reports/stock_report_builder.py`

## 验收标准

- 可以输出 5~10 只候选股票
- 每条候选至少包含：symbol、score、risk_level、reason_summary、invalidations
- 报告可读
- 可回写现有股票推荐表结构

## 风险点

- 输入口径如果不稳定，会让结果不一致
- 如果先做过多复杂因子，会拖慢 MVP

## 当前状态

- 已完成阶段2股票 MVP 最小闭环

## 已完成

- 已建立 `/Users/gjhan21/cursor/sercherai/services/strategy-engine/app/schemas/stock.py`
- 已建立 `/Users/gjhan21/cursor/sercherai/services/strategy-engine/app/domain/seeds/market_seed_loader.py`
- 已建立 `/Users/gjhan21/cursor/sercherai/services/strategy-engine/app/domain/features/stock_feature_factory.py`
- 已建立 `/Users/gjhan21/cursor/sercherai/services/strategy-engine/app/domain/selectors/stock_selector.py`
- 已建立 `/Users/gjhan21/cursor/sercherai/services/strategy-engine/app/domain/risk/portfolio_guard.py`
- 已建立 `/Users/gjhan21/cursor/sercherai/services/strategy-engine/app/domain/reports/stock_report_builder.py`
- 已建立 `/Users/gjhan21/cursor/sercherai/services/strategy-engine/app/domain/pipelines/stock_selection_pipeline.py`
- 已将 `/Users/gjhan21/cursor/sercherai/services/strategy-engine/app/core/job_runner.py` 接入真实股票选股流水线
- 已将股票 job 输出升级为结构化 `report + publish_payloads`
- 已建立 `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/strategy_engine_client.go`
- 已将 `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/mysql_repo.go` 接入 Go -> strategy-engine 的 stock adapter
- 已在 `/Users/gjhan21/cursor/sercherai/backend/internal/platform/config/config.go` 增加 `STRATEGY_ENGINE_BASE_URL` / `STRATEGY_ENGINE_TIMEOUT_MS` / `STRATEGY_ENGINE_POLL_MS`
- 已补充 `/Users/gjhan21/cursor/sercherai/services/strategy-engine/tests/test_stock_pipeline.py`
- 已补充 `/Users/gjhan21/cursor/sercherai/services/strategy-engine/tests/test_jobs.py` 对股票 job 结果的断言
- 已补充 `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/strategy_engine_client_test.go`

## 实际完成范围

本阶段实际采用了“内置市场种子 + 量化因子评分 + 风险门槛 + 发布映射”的轻量 MVP：
- 市场种子先使用服务内置样本集，支持 `seed_symbols` / `excluded_symbols` 控制
- 特征工程复用了现有系统量化推荐的核心思路：趋势、资金、估值、资讯四类评分
- selector 先完成排序与 shortlist，portfolio guard 负责风险门槛、风险等级和分散化控制
- report builder 负责生成候选清单、失效条件和可回写到 Go 后端的数据 payload

## 验收结果

- `POST /internal/v1/jobs/stock-selection` 已可生成真实股票报告
- 报告中每条候选均包含 `symbol`、`score`、`risk_level`、`reason_summary`、`invalidations`
- `publish_payloads` 已映射到现有股票推荐与详情写入结构
- 本地测试通过：`pytest`，4 个用例全部通过
- 本地启动冒烟通过：提交股票 job 后，`GET /internal/v1/jobs/{job_id}` 返回 `SUCCEEDED`，并带回 `selected_count = 5`

## 偏差说明

- 阶段2第一版没有直接接入真实 MySQL / Redis / 行情 API，而是先用内置样本集稳定算法和输出结构
- 阶段2结果先由 strategy-engine 产出 `artifacts.report`，再由 Go adapter 负责落库
- 目前仅在配置 `STRATEGY_ENGINE_BASE_URL` 时走新链路，未配置时仍回退本地生成逻辑

## 遗留问题

- 还未接入真实股票行情、资金流、资讯信号数据源
- 还未把 strategy-engine adapter 接到更多股票管理流转和发布动作中
- 当前报告没有历史回测段落，仍偏向当日推荐解释
- 当前种子池规模有限，尚未覆盖更完整的市场 universe

## 进入下一阶段条件

- 阶段3可以开始：基于阶段2的任务协议与报告结构，继续完成期货策略 MVP
