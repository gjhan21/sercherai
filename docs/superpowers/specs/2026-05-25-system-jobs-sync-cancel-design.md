# System Jobs 同步任务取消能力设计

## Summary

当前 admin 端 `system-jobs` 的同步任务列表已经支持：

- 创建同步任务
- 查看同步任务列表
- 查看同步任务详情与批次明细
- 重试失败批次

但缺少一个关键运维动作：**取消正在运行或尚未完成的同步任务**。

这导致当操作员误发大范围同步、发现参数错误、或某个长历史同步已经明显不需要继续跑时，只能等待其自然完成或失败，无法主动止损。

本次能力补齐的目标是：

- 在 `system-jobs` 的同步任务列表和详情中新增“取消任务”按钮
- 为市场同步任务新增正式取消接口
- 让后端执行链在阶段/批次边界感知取消状态并优雅停止
- 保留已完成的阶段与明细，不回滚已落地数据
- 取消后仍允许通过既有重试机制继续处理该任务

## 背景

### 当前前端能力

当前 admin 页面：

- [admin/src/views/SystemJobsView.vue](/Users/gjhan21/cursor/sercherai/admin/src/views/SystemJobsView.vue)

在“同步任务列表”里，操作列只提供：

- `查看详情`
- `重试失败批次`（条件展示）

前端 API 仅包含：

- [createMarketDataBackfillRun](/Users/gjhan21/cursor/sercherai/admin/src/api/admin.js)
- [listMarketDataBackfillRuns](/Users/gjhan21/cursor/sercherai/admin/src/api/admin.js)
- [getMarketDataBackfillRun](/Users/gjhan21/cursor/sercherai/admin/src/api/admin.js)
- [listMarketDataBackfillRunDetails](/Users/gjhan21/cursor/sercherai/admin/src/api/admin.js)
- [retryMarketDataBackfillRun](/Users/gjhan21/cursor/sercherai/admin/src/api/admin.js)

不存在 `cancelMarketDataBackfillRun()`。

### 当前后端能力

admin market data 路由当前仅包含：

- `POST /admin/market-data/backfill`
- `GET /admin/market-data/backfill-runs`
- `GET /admin/market-data/backfill-runs/:id`
- `GET /admin/market-data/backfill-runs/:id/details`
- `POST /admin/market-data/backfill-runs/:id/retry`

对应文件：
- [backend/router/admin.go](/Users/gjhan21/cursor/sercherai/backend/router/admin.go)

不存在 `cancel` 路由。

repo/service 当前仅暴露：

- `AdminCreateMarketDataBackfillRun`
- `AdminListMarketDataBackfillRuns`
- `AdminGetMarketDataBackfillRun`
- `AdminListMarketDataBackfillRunDetails`
- `AdminRetryMarketDataBackfillRun`

对应文件：
- [backend/internal/growth/repo/interfaces.go](/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/interfaces.go)
- [backend/internal/growth/service/market_data_backfill_service.go](/Users/gjhan21/cursor/sercherai/backend/internal/growth/service/market_data_backfill_service.go)
- [backend/internal/growth/repo/market_data_backfill.go](/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/market_data_backfill.go)

### 当前执行链行为

执行入口在：

- [backend/internal/growth/repo/market_backfill_execution.go](/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/market_backfill_execution.go)

`executeMarketDataBackfillRun()` 当前会串行跑：

- `MASTER`
- `QUOTES`
- `DAILY_BASIC`
- `MONEYFLOW`
- `TRUTH`
- `COVERAGE_SUMMARY`

以及长历史 `QUOTES` 的 chunk/batch 流程。

当前执行链没有任何取消检查点，所以即使将 run 标成 `CANCELLED`，也不会自动停止。

## 问题定义

当前缺失取消能力会带来以下问题：

1. **误触发无法止损**
   操作员一旦发起错误的全量或长历史同步，只能等待任务跑完。

2. **运行中任务不可控**
   任务中心已有创建、查看、重试，但没有“终止”能力，运维闭环不完整。

3. **任务状态模型不完整**
   虽然状态词典中已有 `CANCELLED`，但没有真正的取消入口和执行链配合。

4. **用户体验和审计链断裂**
   没有取消按钮，操作员也无法通过操作日志清楚表达“为什么停止这次同步”。

## 目标

### 产品目标

1. 在同步任务列表提供可见的取消入口
2. 在同步任务详情中也提供取消入口
3. 仅对 `PENDING` / `RUNNING` 任务允许取消
4. 取消后任务状态改为 `CANCELLED`
5. 取消后仍允许未来通过既有重试机制继续处理

### 技术目标

1. 新增正式取消 API
2. 为 repo/service/handler 打通取消链路
3. 在执行链关键边界点检查取消状态并优雅停下
4. 不回滚已成功执行的阶段，不破坏已落地数据
5. 保持现有详情、批次明细、重试模型不变

## 非目标

本次不做以下内容：

- 不实现“强制立即中断”数据库写入
- 不做已完成阶段的数据回滚
- 不新增“从取消点继续”的新模型
- 不改变现有重试接口语义
- 不重构整个 `system-jobs` 页面信息架构

## 方案对比

### 方案 A：软取消（推荐）

做法：
- 新增取消接口
- 将 run 标记为 `CANCELLED`
- 执行链在 stage / batch 边界检查取消状态
- 发现已取消则停止后续执行

优点：
- 风险最低
- 与现有同步任务执行链兼容性最好
- 不会产生“半事务强中断”问题

缺点：
- 不能保证在极小粒度上即时停止，只能在边界点停下

### 方案 B：仅允许取消 PENDING

做法：
- 只允许取消尚未运行的任务
- 运行中的任务不能取消

优点：
- 实现简单

缺点：
- 用户最需要取消的往往是长时间运行中的任务
- 运维价值有限

### 方案 C：强制硬中断

做法：
- 在当前执行点尝试立刻打断
- 尽可能中止正在运行的 stage

优点：
- 表面上“取消响应最快”

缺点：
- 风险最大
- 易留下半写入、半统计、半 truth 状态
- 不适合当前这套同步执行模型

### 结论

采用 **方案 A：软取消**。

## 最终设计

### 1. 前端交互设计

#### 列表入口

在 [admin/src/views/SystemJobsView.vue](/Users/gjhan21/cursor/sercherai/admin/src/views/SystemJobsView.vue) 的“同步任务列表”操作列新增：

- `取消`

展示条件：
- `item.status === "PENDING" || item.status === "RUNNING"`

点击后：
- 弹出 `ElMessageBox.confirm`
- 确认文案强调：
  - 已完成批次会保留
  - 未开始批次将停止
  - 任务未来仍可重试

成功后：
- 刷新同步任务列表
- 若当前详情抽屉打开且是同一个任务，也刷新详情

#### 详情入口

在同步任务详情抽屉头部增加：

- `取消任务`

条件同上，只对 `PENDING / RUNNING` 展示。

#### 前端 API

在 [admin/src/api/admin.js](/Users/gjhan21/cursor/sercherai/admin/src/api/admin.js) 新增：

- `cancelMarketDataBackfillRun(id, payload)`

payload 第一版仅支持：

```json
{
  "reason": "manual cancel"
}
```

### 2. 后端接口设计

新增 admin 路由：

- `POST /api/v1/admin/market-data/backfill-runs/:id/cancel`

新增 handler：
- `CancelMarketDataBackfillRun`

新增 service：
- `AdminCancelMarketDataBackfillRun(runID string, operator string, reason string)`

新增 repo：
- `AdminCancelMarketDataBackfillRun(runID string, operator string, reason string)`

### 3. 状态机设计

允许取消的状态：
- `PENDING`
- `RUNNING`

拒绝取消的状态：
- `SUCCESS`
- `FAILED`
- `PARTIAL_SUCCESS`
- `CANCELLED`

取消成功后，run 更新为：
- `status = CANCELLED`
- `finished_at = now`
- `updated_at = now`
- `error_message = "任务已取消"` 或含 reason 的说明

如果需要，也可把取消原因写入 summary 扩展字段；第一阶段不是必须。

### 4. 执行链取消检查

在 [backend/internal/growth/repo/market_backfill_execution.go](/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/market_backfill_execution.go) 增加统一检查方法，例如：

- `checkMarketBackfillRunCancelled(runID string) error`

检查时机：

1. 每个 stage 开始前
2. 每个 assetType 大循环开始前
3. 长历史 quotes 的每个 chunk/batch 开始前

一旦发现已取消：
- 停止后续 stage / batch 执行
- 以 `CANCELLED` 状态收尾
- 返回可识别的取消错误或取消结果

### 5. 已完成数据与重试关系

取消不是回滚：
- 已执行成功的 detail 保留
- 已写入的行情/事实/truth 数据保留
- 未开始的 detail 不会新增成功结果

重试策略按你确认的口径保留：
- 取消后仍允许使用既有重试能力
- 本次不新增“从取消点恢复”的独立模型

## API 约定

### Request

`POST /api/v1/admin/market-data/backfill-runs/:id/cancel`

```json
{
  "reason": "参数选择错误，停止当前同步"
}
```

### Response

保持与 create/retry 风格一致：

```json
{
  "code": 0,
  "message": "OK",
  "data": {
    "run_id": "mbr_xxx",
    "scheduler_run_id": "jr_xxx",
    "universe_snapshot_id": "mus_xxx",
    "status": "CANCELLED"
  }
}
```

### Error cases

- run 不存在：404
- 当前状态不可取消：400
- DB/执行链异常：500

## 审计日志

新增操作日志：

- module: `MARKET_DATA`
- action: `CANCEL_BACKFILL_RUN`
- targetType: `MARKET_BACKFILL_RUN`
- targetID: `run.ID`
- reason: `取消原因`

这样后续排查时可以明确知道是谁、为什么取消了同步任务。

## 测试计划

### 后端 handler

扩展 [backend/internal/growth/handler/market_data_admin_handler_test.go](/Users/gjhan21/cursor/sercherai/backend/internal/growth/handler/market_data_admin_handler_test.go)：

- 取消成功
- 非法状态拒绝
- 缺少 reason 也可成功（如采用非必填）或按最终规则校验

### 后端 repo

扩展 [backend/internal/growth/repo/market_data_backfill_repo_test.go](/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/market_data_backfill_repo_test.go)：

- `PENDING` 可取消
- `RUNNING` 可取消
- `SUCCESS / FAILED / CANCELLED` 不可取消
- 执行链遇到取消后停止后续 stage

### 前端

扩展 [admin/src/views/system-jobs-view.test.js](/Users/gjhan21/cursor/sercherai/admin/src/views/system-jobs-view.test.js)：

- 同步任务列表存在取消按钮逻辑
- 只对 `PENDING / RUNNING` 展示
- 页面源码引用 `cancelMarketDataBackfillRun`

## 风险与缓解

### 风险 1：取消不够“即时”

说明：
软取消只能在 stage/batch 边界停下。

缓解：
- 在长历史批次内增加更细粒度检查
- UI 文案明确“会在当前批次结束后停止”

### 风险 2：取消后任务与明细状态不一致

缓解：
- run 状态统一收尾为 `CANCELLED`
- 已完成 detail 保持成功
- 不新增后续 detail
- 测试覆盖阶段中途取消场景

### 风险 3：取消后重试行为边界不清

缓解：
- 第一阶段明确：取消后仍可沿用现有重试入口
- 不额外发明“取消恢复”语义

## 实施顺序

1. 新增取消 API 设计与 repo/service/handler 链路
2. 在执行链加入取消检查点
3. 为 handler/repo 补测试
4. 在 admin 同步任务列表与详情中加取消按钮
5. 跑后端测试、前端测试和 admin build 验证
