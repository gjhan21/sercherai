# 股票与期货预测增强 Admin 配置规划设计规范

## Summary

本规范用于定义“股票与期货预测增强”能力在 `admin` 管理端的承接方式，目标不是新建一套平行后台，而是在现有 `admin` 体系上，按 `L1 / L2 / L3` 三个阶段逐步接入必要的配置、运营与审计能力。

本规范解决的问题是：

- 现有后台已经具备哪些真实能力
- 预测增强首期是否需要新增 admin 配置开发
- 哪些配置应该复用现有后台模块
- 哪些能力要等到 `L2 / L3` 再做
- 如何避免后台过早碎片化、重复造轮子

本规范的核心结论固定为：

1. 当前系统已经具备较完整的 admin 基础，不是空白后台。
2. 预测增强 `L1` 不以新增 admin 配置页面为前提。
3. `L1` 如需补后台，只允许做轻量嵌入式配置，不新开大模块。
4. `L2` 才开始系统化扩展策略引擎相关配置。
5. `L3` 才考虑独立“深推演 / 预测实验室”后台工作台。

## 当前状态更新（2026-03-29）

- `L1` 的 admin 轻量嵌入已完成并落地到 `main`：
  - `/Users/gjhan21/cursor/sercherai/admin/src/views/SystemConfigsView.vue`
  - `/Users/gjhan21/cursor/sercherai/admin/src/views/MarketCenterView.vue`
  - `/Users/gjhan21/cursor/sercherai/admin/src/views/ReviewCenterView.vue`
- `L2` 的 admin 复用式可见性与配置也已完成并落地到 `main`，继续沿用现有模块承接：
  - `SystemConfigsView`
  - `MarketCenterView`
  - `ReviewCenterView`
  - `stock-selection/*`
  - `futures-selection/*`
- 因此当前 admin 侧未完成部分主要只剩 `L3` 的更深运营与实验承接，不需要回头再讨论是否为 `L1 / L2` 新开后台中心。

## 当前 Admin 真实基线

### 已有页面级后台能力

当前 `admin` 已存在以下主页面：

- 用户中心
- 资讯后台
- 社区后台
- 审核中心
- 市场中心
- 实验分析
- 股票选股模块
- 期货选策模块
- 会员中心
- 风控中心
- 任务中心
- 配置中心
- 权限中心
- 审计日志

关键路由事实源：

- `/Users/gjhan21/cursor/sercherai/admin/src/router/index.js`

### 已有接口级后台能力

当前后端已经存在多组可复用的 admin 接口：

- 市场数据同步 / 回补 / truth 重建
- 数据源治理、质量日志、覆盖率摘要
- 股票与期货 strategy-engine 发布历史与对比
- strategy-engine seed sets / agents / scenarios / publish policies / jobs
- 市场事件、节奏任务
- 用户状态、会员等级、实名状态、密码重置
- 系统任务定义 / 运行 / 重试
- 系统配置中心
- 登录安全与风控配置

关键事实源：

- `/Users/gjhan21/cursor/sercherai/backend/router/router.go`
- `/Users/gjhan21/cursor/sercherai/admin/src/api/admin.js`

### 已有最相关的页面落点

和预测增强最相关的现有后台页面主要是：

- `stock-selection/*`
- `futures-selection/*`
- `MarketCenterView`
- `ReviewCenterView`
- `SystemConfigsView`
- `SystemJobsView`

这意味着预测增强后续 admin 承接，应优先复用这些页面，不应一上来新建第 2 套后台架构。

## Problem Statement

虽然后台基础已经存在，但预测增强相关能力目前还没有形成一套明确的 admin 规划，具体问题有三类：

1. 不清楚 `L1` 是否应该同步做后台配置，容易造成范围膨胀。
2. 不清楚哪些配置应放在模板 / profile 层，哪些应放系统配置层。
3. 不清楚什么时候才值得做独立“预测实验室”后台工作台。

如果不先定规则，后续很容易出现：

- 在 `SystemConfigs` 堆过多业务配置
- 在 `stock-selection` / `futures-selection` 重复加相似表单
- 提前为 `L2 / L3` 开独立后台页面，导致维护复杂度快速上升

## Design Goals

### 目标一：L1 不让 admin 变成阻塞项

`L1` 的重点是增强 explanation 链，不是先搭后台运营面。后台只能补最必要的开关和运营可见性。

### 目标二：后台配置按层级落点

不同层级的配置要落在不同后台模块：

- 模板 / profile 级配置
- 系统级开关与阈值
- 只读运营视图
- 深推演任务与长期记忆治理

### 目标三：最大复用现有后台模块

优先在现有模块中承接：

- `stock-selection / futures-selection`
- `SystemConfigs`
- `MarketCenter`
- `ReviewCenter`
- `SystemJobs`

## 方案比较

### 方案 A：预测增强一开始就独立开后台中心

做法：

- 新增独立“预测增强配置中心”
- 单独管理子问题模板、记忆反馈、情景模板、深推演准入

优点：

- 结构看起来完整
- 页面心智统一

缺点：

- `L1` 过度设计
- 会和现有 `stock-selection / futures-selection / system-configs` 重叠
- 后台工作流会被人为拆碎

### 方案 B：完全不做后台承接

做法：

- `L1 / L2 / L3` 全部只做后端逻辑，不给后台配置或运营面

优点：

- 开发最省
- 不会引入后台复杂度

缺点：

- 运营不可见
- 调参与治理能力弱
- 到 `L2 / L3` 时会欠下较大的后台债务

### 方案 C：分阶段复用现有后台模块

做法：

- `L1`：只做轻量嵌入式后台承接
- `L2`：扩展现有 strategy-engine 相关后台配置
- `L3`：再独立深推演工作台

优点：

- 最符合当前系统现实
- 复用最多，成本最低
- 不会过早制造后台碎片
- 和 `L1 / L2 / L3` 主路线节奏一致

缺点：

- 需要明确每阶段边界
- 页面入口初期会分散在多个现有模块中

### 推荐结论

推荐采用 **方案 C：分阶段复用现有后台模块**。

## 分阶段 Admin 规划

### L1：轻量嵌入式承接，不新增大模块

`L1` 的 admin 规划固定为：

- 不以新增 admin 页面为前提
- 不新开“预测增强配置中心”
- 只允许在现有页面中补少量必要配置与只读运营信息

#### L1 当前已落地状态

截至 `2026-03-28`，`L1` 的 admin 轻量承接已按该规范落地，且保持“嵌入式、非独立后台中心”边界不变。

已落地页面与文件如下：

1. `SystemConfigsView`

- 文件：`/Users/gjhan21/cursor/sercherai/admin/src/views/SystemConfigsView.vue`
- helper：`/Users/gjhan21/cursor/sercherai/admin/src/lib/forecast-admin.js`
- 已落地配置项：
  - `growth.forecast_l1.enabled`
  - `growth.forecast_l1.explanation_enabled`
  - `growth.forecast_l1.memory_feedback_min_samples`
  - `growth.forecast_l1.advisory_priority_threshold`

2. `MarketCenterView`

- 文件：`/Users/gjhan21/cursor/sercherai/admin/src/views/MarketCenterView.vue`
- 已落地能力：
  - 在股票/期货发布详情中展示“预测增强摘要”
  - 直接回读现有 `publish_payloads / report_snapshot`
  - 展示覆盖率、研究编排数、观察信号数、记忆反馈数、高 advisory 样本数

3. `ReviewCenterView`

- 文件：`/Users/gjhan21/cursor/sercherai/admin/src/views/ReviewCenterView.vue`
- 已落地能力：
  - 展示 advisory-only 审核提示
  - 回读 `growth.forecast_l1.*` 配置
  - 明确提示“不改变现有审核主流程”

对应测试文件：

- `/Users/gjhan21/cursor/sercherai/admin/src/lib/forecast-admin.test.js`
- `/Users/gjhan21/cursor/sercherai/admin/src/views/system-configs-view.test.js`
- `/Users/gjhan21/cursor/sercherai/admin/src/views/market-center-view.test.js`
- `/Users/gjhan21/cursor/sercherai/admin/src/views/review-center-view.test.js`

#### L1 可新增的后台项

1. 放在 `stock-selection / futures-selection profiles / templates` 的配置

- 研究编排增强开关
- 子问题模板版本
- 是否启用记忆反馈增强
- 是否启用置信度校准
- 市场环境模板选择策略

2. 放在 `SystemConfigs` 的全局项

- 全局 kill switch
- `memory_feedback` 样本阈值
- 评估回补参与窗口
- advisory priority 全局阈值
- explanation 增强字段显示总开关

3. 放在 `ReviewCenter / MarketCenter` 的只读运营项

- 当前增强 explanation 命中数量
- 当前高 advisory priority 标的
- 近期高失败复发模板摘要
- explanation 增强覆盖率

#### L1 明确不做

- 独立预测增强配置中心
- 独立分析预测后台页
- 深推演任务中心
- 长期记忆治理面板
- 后台参数直接改排序 / 权重 / 发布审批主链

### L2：正式扩展 strategy-engine 配置面

`L2` admin 规划固定为：复用现有 `strategy-engine` 相关后台接口与页面，系统化承接情景推演配置。

#### L2 重点承接面

1. `scenarios`

扩展现有 scenario templates，用于承接：

- 牛 / 基 / 熊三情景模板
- trigger / confirmation / invalidation 模板
- 股票 / 期货分域模板
- 市场环境绑定模板

2. `agents`

扩展现有 agent profiles，用于承接：

- 股票金融角色函数
- 期货金融角色函数
- 角色 veto 能力
- 角色启用条件

3. `publish-policies`

扩展现有 publish policy，用于承接：

- advisory priority 到 review 的映射
- 某些风险信号命中后的人工复核建议
- 某些模板在特定环境下的校准规则

#### L2 明确不做

- 独立深推演后台中心
- 全量任务化的深推演运营面
- 长期记忆治理台

### L3：独立深推演与长期学习工作台

只有进入 `L3`，才建议考虑新开独立后台模块。

#### L3 建议独立页面能力

1. 深推演任务中心

- 创建任务
- 准入规则
- 执行状态
- 重试与终止
- 成本与频次控制

2. 金融角色库

- 股票角色模板
- 期货角色模板
- 角色权重与 veto 能力
- 角色适用市场环境

3. 长期记忆治理中心

- 记忆聚合规则
- 过期规则
- 误记忆清理
- 记忆质量校准

4. 推演报告与审计台

- 中间日志
- 工具调用
- 证据链
- 后验命中回看

## 模块落点规则

### 放在 `stock-selection / futures-selection` 的内容

适合放：

- 模板级配置
- profile 级配置
- 研究编排开关
- 资产类型专属规则

### 放在 `SystemConfigs` 的内容

适合放：

- 全局开关
- 默认阈值
- 跨模块统一策略
- 测试与降级相关配置

### 放在 `ReviewCenter / MarketCenter` 的内容

适合放：

- 运营只读摘要
- review priority 承接
- 增强覆盖率与命中情况

### 只有到 `L3` 才单独开新模块的内容

适合放：

- 深推演任务流
- 长期记忆治理
- 重型研究日志与审计

## Public APIs / Interfaces / Types

### 当前阶段不要求新增的接口

`L1` 不要求为了 admin 单独新增后端接口，只允许：

- 复用现有 strategy-engine / selection / system-configs / review 相关接口
- 在现有接口上做非破坏性扩展

### `L2 / L3` 才可能新增的接口方向

- 情景模板专用配置字段
- 金融角色模板专用配置字段
- 深推演任务与报告接口
- 长期记忆治理接口

## Test Plan

### L1 admin 验收

- 不新增独立后台大模块
- 轻量配置能复用现有页面承接
- 后台配置不会直接改排序 / 权重 / 发布审批主链
- 当前实现已满足以上验收项

验证命令：

```bash
cd /Users/gjhan21/cursor/sercherai/admin
node --test src/lib/forecast-admin.test.js src/views/system-configs-view.test.js src/views/market-center-view.test.js src/views/review-center-view.test.js
npm run build
```

### L2 admin 验收

- 情景模板能在现有 strategy-engine 配置面中维护
- 角色函数配置能区分股票与期货
- publish policy 承接 advisory 逻辑而非直接接管推荐主链

### L3 admin 验收

- 独立深推演工作台只服务高价值对象
- 任务、报告、日志、治理具备清晰边界
- 不与现有后台页面职责冲突

## Out Of Scope

- 本规范不直接要求立刻开发 admin 新页面
- 本规范不改变当前 `L1` 首期范围
- 本规范不让后台配置直接接管推荐主链

## Recommended Next Step

当前推荐执行顺序固定为：

1. 先按 `L1` 主线开发 explanation 增强能力
2. `L1` 稳定后，再补一轮轻量 admin 配置规划落地
3. `L2` 再把 scenario / agent / publish policy 正式接入后台
4. `L3` 再规划独立深推演后台中心
