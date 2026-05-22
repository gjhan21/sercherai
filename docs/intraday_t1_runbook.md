# 超短线 T+1 选股 —— 操作手册

最后更新: 2026-05-05

## 概述

本手册涵盖超短线 T+1 选股功能上线前的 3 项操作任务：
1. 同步 Tushare 涨停板/龙虎榜数据
2. 运行回测标定策略权重
3. Go 后端编译验证

所有操作需在**已配置好 Go 编译环境和数据库**的开发机上执行。

---

## 前置条件

- Go 1.20+ 编译环境
- MySQL 数据库（已执行 `backend/migrations/20260505_00_intraday_t1_tables.sql`）
- Tushare Pro token（积分 >= 2120）
- Python 3.10+ 环境（用于回测和烟幕测试）
- `backend/` 和 `services/strategy-engine/` 目录在同一个项目根下

环境变量（在 `backend/` 和 `services/strategy-engine/` 目录各需要一个 `.env`）：

```bash
# backend/.env
TUSHARE_TOKEN=你的token
MYSQL_DSN=user:pass@tcp(127.0.0.1:3306)/sercherai?charset=utf8mb4&parseTime=true

# services/strategy-engine/.env
STRATEGY_ENVIRONMENT=development
STRATEGY_ENGINE_GO_BACKEND_BASE_URL=http://localhost:8080
STRATEGY_ENGINE_GO_BACKEND_TIMEOUT_MS=8000
```

---

## 任务 1: 同步 Tushare 涨停板/龙虎榜数据

### 1.1 启动 Go 后端

```bash
cd backend
go build -o sercherai-server ./cmd/server
./sercherai-server &
```

### 1.2 同步涨停板榜单 (kpl_list)

```bash
# 同步最近 5 个交易日（默认）
curl -X POST http://localhost:8080/admin/stocks/kpl-list/sync \
  -H "Content-Type: application/json" \
  -d '{"source_key": "tushare", "days": 5}'

# 同步最近 30 个交易日
curl -X POST http://localhost:8080/admin/stocks/kpl-list/sync \
  -H "Content-Type: application/json" \
  -d '{"source_key": "tushare", "days": 30}'
```

### 1.3 同步龙虎榜 (top_list)

```bash
# 同步最近 5 个交易日
curl -X POST http://localhost:8080/admin/stocks/top-list/sync \
  -H "Content-Type: application/json" \
  -d '{"source_key": "tushare", "days": 5}'

# 同步最近 30 个交易日
curl -X POST http://localhost:8080/admin/stocks/top-list/sync \
  -H "Content-Type: application/json" \
  -d '{"source_key": "tushare", "days": 30}'
```

### 1.4 验证数据

```sql
-- 检查涨停板数据量
SELECT COUNT(*), MAX(trade_date) FROM stock_limit_up_daily;

-- 检查龙虎榜数据量
SELECT COUNT(*), MAX(trade_date) FROM stock_top_list_daily;

-- 按日期统计涨停板
SELECT trade_date, COUNT(*) as cnt FROM stock_limit_up_daily
WHERE trade_date >= DATE_SUB(CURDATE(), INTERVAL 30 DAY)
GROUP BY trade_date ORDER BY trade_date DESC;

-- 按日期统计龙虎榜
SELECT trade_date, COUNT(*) as cnt FROM stock_top_list_daily
WHERE trade_date >= DATE_SUB(CURDATE(), INTERVAL 30 DAY)
GROUP BY trade_date ORDER BY trade_date DESC;
```

**预期结果**:
- kpl_list: 每交易日约 50-120 条（涨停股票数），30 天约 1500-3600 条
- top_list: 每交易日约 20-50 条（龙虎榜上榜股票），30 天约 600-1500 条

---

## 任务 2: 运行回测标定

### 2.1 烟幕测试（无 Go 后端依赖，验证管道正确性）

```bash
cd services/strategy-engine

# 90 天回测（合成数据，仅验证管道端到端正确性）
python3 -m app.tools.smoke_test_backtest \
  --days 90 --seeds 30 --limit 5 --min-score 10 \
  --format summary

# 输出 JSON 结果
python3 -m app.tools.smoke_test_backtest \
  --days 90 --seeds 30 --limit 5 --min-score 10 \
  --format json --output data/smoke_test_90d.json

# 输出 CSV 交易记录
python3 -m app.tools.smoke_test_backtest \
  --days 90 --seeds 30 --limit 5 --min-score 10 \
  --format csv --output data/smoke_test_90d.csv
```

### 2.2 真实数据回测（需要 Go 后端运行中）

**方式 A: CLI**

```bash
cd services/strategy-engine

python3 -m app.tools.backtest_runner \
  --start 2025-11-01 \
  --end 2026-05-01 \
  --limit 5 \
  --min-score 65.0 \
  --max-risk MEDIUM \
  --max-per-strategy 2 \
  --max-per-sector 1 \
  --format summary
```

**方式 B: API**

```bash
# 启动策略引擎服务
cd services/strategy-engine
pip3 install -r requirements.txt  # 首次运行需要安装依赖
python3 -m uvicorn app.main:app --host 0.0.0.0 --port 8000 &

# 查看可用日期范围
curl http://localhost:8000/backtest/intraday-t1/available-dates

# 运行回测
curl -X POST http://localhost:8000/backtest/intraday-t1 \
  -H "Content-Type: application/json" \
  -d '{
    "start_date": "2025-11-01",
    "end_date": "2026-05-01",
    "limit": 5,
    "min_score": 65.0,
    "max_risk_level": "MEDIUM",
    "max_per_strategy": 2,
    "max_per_sector": 1,
    "format": "summary"
  }'
```

### 2.3 分析回测结果

重点关注以下指标进行策略权重标定：

1. **策略胜率**: 各策略在不同市场环境下的 hit_rate
2. **平均收益**: 单策略 avg_return > 0.2% 为优
3. **市场环境适配**: 对比 `_REGIME_WEIGHTS` 中预设权重与实际表现
4. **多策略共振**: multi_strategy_hit_rate 应显著高于单策略

在 `services/strategy-engine/app/domain/decision/intraday_decision_fusion.py` 中调整 `_REGIME_WEIGHTS`，然后重新运行回测验证。

---

## 任务 3: Go 编译验证

### 3.1 编译

```bash
cd backend

# 完整编译检查
go build ./...

# 单独编译 server
go build -o /dev/null ./cmd/server

# 运行 vet 检查
go vet ./...

# 检查是否有未使用的导入
go build ./... 2>&1 | grep -E "imported and not used|undefined"
```

### 3.2 运行单元测试

```bash
cd backend

# 所有测试
go test ./... -count=1

# 仅 growth/repo 包
go test ./internal/growth/repo/... -v -count=1

# 仅 migration 相关
go test ./internal/growth/... -run TestSync -v
```

### 3.3 静态分析脚本

```bash
cd backend

python3 ../scripts/go_static_verify.py
```

该脚本会检查：
- 所有修改文件是否存在
- 结构体定义
- 跨文件引用一致性
- 路由注册
- Handler-Service-Repo 链完整性
- Go 1.20 兼容性

---

## 附录 A: 项目文件路径速查

| 用途 | 路径 |
|------|------|
| DB 表迁移 | `backend/migrations/20260505_00_intraday_t1_tables.sql` |
| 模板注册 | `backend/migrations/20260505_01_intraday_t1_template.sql` |
| MySQL repo (kpl/top) | `backend/internal/growth/repo/mysql_repo.go` |
| 多数据源调度 | `backend/internal/growth/repo/market_data_multi_source.go` |
| 接口定义 | `backend/internal/growth/repo/interfaces.go` |
| 内存 repo stub | `backend/internal/growth/repo/inmemory_repo.go` |
| Service 层 | `backend/internal/growth/service/service.go` |
| Handler 层 | `backend/internal/growth/handler/market_data_sync_handler.go` |
| 路由注册 | `backend/router/admin.go` |
| 策略 A (超跌反弹) | `services/strategy-engine/app/domain/strategies/oversold_bounce.py` |
| 策略 B (缩量止跌) | `services/strategy-engine/app/domain/strategies/volume_exhaustion.py` |
| 策略 C (资金背离) | `services/strategy-engine/app/domain/strategies/capital_divergence.py` |
| 策略 D (均线支撑) | `services/strategy-engine/app/domain/strategies/ma_support.py` |
| 策略 E (看涨吞没) | `services/strategy-engine/app/domain/strategies/bullish_engulfing.py` |
| 策略 F (涨停次日) | `services/strategy-engine/app/domain/strategies/limit_up_next_day.py` |
| 策略 G (强势回踩) | `services/strategy-engine/app/domain/strategies/strong_pullback.py` |
| 种子挖掘器 | `services/strategy-engine/app/domain/seeds/intraday_seed_miner.py` |
| 决策融合器 | `services/strategy-engine/app/domain/decision/intraday_decision_fusion.py` |
| 组合风控 | `services/strategy-engine/app/domain/risk/portfolio_guard.py` |
| 回测引擎 | `services/strategy-engine/app/domain/backtesting/intraday_backtest.py` |
| 回测报告器 | `services/strategy-engine/app/domain/backtesting/backtest_reporter.py` |
| 回测 API | `services/strategy-engine/app/api/routes_backtest.py` |
| 回测 CLI | `services/strategy-engine/app/tools/backtest_runner.py` |
| 烟幕测试 | `services/strategy-engine/app/tools/smoke_test_backtest.py` |
| 设计文档 | `docs/超短线T1选股开发方案.md` |
| 操作手册 | `docs/intraday_t1_runbook.md` |

## 附录 B: 关键参数说明

| 参数 | 默认值 | 说明 |
|------|--------|------|
| `limit` | 5 | 每日推荐股票数量 |
| `min_score` | 65 | 最低评分阈值（合成数据测试建议 10） |
| `max_risk_level` | MEDIUM | 最大风险等级 (LOW/MEDIUM/HIGH) |
| `max_per_strategy` | 2 | 每策略最大输出（分散化约束） |
| `max_per_sector` | 1 | 每板块最大输出（分散化约束） |
| `watchlist_limit` | 10 | 备选观察列表长度 |

**注意**: `min_score=65` 是设计规范中的阈值，但由于融合层会对归一化后的分数施加市场环境权重（0.15-0.30），实际加权分数通常不超过 30。在融合层中，当 `min_score` 过滤无结果时会自动下调 10 分。使用真实数据的回测需要根据实际分数分布调整此阈值。

## 附录 C: 常见问题

**Q: Tushare 同步返回空数据？**
A: 检查 token 积分是否足够（kpl_list + top_list 共需约 2120 分），确认交易日非节假日。

**Q: 回测报 "go backend returned empty stock seeds"？**
A: Go 后端未启动或数据库无数据。先执行任务 1 同步数据。

**Q: 策略 F (limit_up_next_day) 无候选？**
A: 策略 F 依赖 kpl_list/top_list 数据。确保已执行任务 1，且数据中有涨停记录。

**Q: 合成回测收益为 0？**
A: 确保 `--min-score` 参数设置合理（合成数据建议 10-15）。
