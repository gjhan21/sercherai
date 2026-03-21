# 阶段3：期货 MVP

最后更新: 2026-03-17
状态: Done

## 阶段目标

跑通期货策略最小闭环：输入合约与市场上下文 -> 生成方向与价位 -> 风险约束 -> 输出结构化策略报告。

## 核心结果

- 期货种子加载可用
- 期货特征工程可用
- futures selector 可输出策略草稿
- 每条策略都有方向、价位、风险、失效条件
- 结果可映射回现有期货策略数据结构

## 主要改动点

- `domain/features/futures_feature_factory.py`
- `domain/selectors/futures_selector.py`
- `domain/risk/leverage_guard.py`
- `domain/reports/futures_report_builder.py`

## 验收标准

- 每条策略至少包含：contract、direction、entry_price、take_profit_price、stop_loss_price、risk_level、reason_summary
- 可输出结构化期货策略列表
- 可回写现有期货策略表结构

## 风险点

- 合约、基差、波动口径不统一会影响策略稳定性
- 如果阶段3过早与股票共用太多细节，会损失期货场景特性

## 当前状态

- 已完成阶段3期货 MVP 最小闭环

## 已完成

- 已建立 `/Users/gjhan21/cursor/sercherai/services/strategy-engine/app/schemas/futures.py`
- 已建立 `/Users/gjhan21/cursor/sercherai/services/strategy-engine/app/domain/seeds/futures_seed_loader.py`
- 已建立 `/Users/gjhan21/cursor/sercherai/services/strategy-engine/app/domain/features/futures_feature_factory.py`
- 已建立 `/Users/gjhan21/cursor/sercherai/services/strategy-engine/app/domain/selectors/futures_selector.py`
- 已建立 `/Users/gjhan21/cursor/sercherai/services/strategy-engine/app/domain/risk/leverage_guard.py`
- 已建立 `/Users/gjhan21/cursor/sercherai/services/strategy-engine/app/domain/reports/futures_report_builder.py`
- 已建立 `/Users/gjhan21/cursor/sercherai/services/strategy-engine/app/domain/pipelines/futures_strategy_pipeline.py`
- 已将 `/Users/gjhan21/cursor/sercherai/services/strategy-engine/app/core/job_runner.py` 接入真实期货策略流水线
- 已补充 `/Users/gjhan21/cursor/sercherai/services/strategy-engine/tests/test_futures_pipeline.py`
- 已补充 `/Users/gjhan21/cursor/sercherai/services/strategy-engine/tests/test_jobs.py` 对期货 job 结果的断言

## 实际完成范围

本阶段实际采用了“内置期货种子 + 方向判断 + 价位建议 + 杠杆约束 + 发布映射”的轻量 MVP：
- 期货种子先以内置股指 / 黄金样本合约为主，支持 `contracts` 定向选择
- 特征工程围绕趋势、持仓、量比、基差 / carry、资讯偏置生成多空方向和置信度
- selector 负责排序与 shortlist，leverage guard 负责风险等级、仓位范围和最小置信度过滤
- report builder 负责生成进场价、止盈价、止损价、失效条件和可回写到 Go 后端的 payload

## 验收结果

- `POST /internal/v1/jobs/futures-strategy` 已可生成真实期货策略报告
- 报告中每条策略均包含 `contract`、`direction`、`entry_price`、`take_profit_price`、`stop_loss_price`、`risk_level`、`reason_summary`
- `publish_payloads` 已映射到现有期货策略与 guidance 写入结构
- 本地测试通过：`pytest`，5 个用例全部通过
- 本地启动冒烟通过：提交期货 job 后，`GET /internal/v1/jobs/{job_id}` 返回 `SUCCEEDED`，并带回 `selected_count = 3`

## 偏差说明

- 阶段3第一版没有接入真实期货行情、基差和盘口数据，而是先用内置样本集稳定方向与价位结构
- 阶段3暂未补 Go 后端到 `strategy-engine` 的期货 adapter，本轮先完成服务内真实策略输出
- 阶段3聚焦趋势 / 方向型策略，不提前做套利、多腿配对和复杂回测仿真

## 遗留问题

- 还未接入真实期货行情、基差、持仓和资讯数据源
- 还未建立 Go 后端调用 `strategy-engine` 并回写 `futures_strategies` / `futures_guidances` 的 adapter
- 当前策略仍以日级建议为主，尚未覆盖更细粒度的时段管理
- 当前合约样本池较小，仍需扩展到更多商品和期限结构

## 进入下一阶段条件

- 阶段4可以开始：基于股票 / 期货双 MVP，引入场景模拟、多代理协同与否决机制
