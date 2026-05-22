# Newclient 客户端深度推演接入设计规范

## Summary

`newclient` 需要把后台已经具备的 `L3` 深度推演能力接到用户客户端，但必须保持当前产品里的真实关系：

- 主推荐 / 主策略链继续承担“今天看什么、怎么做”的主职责
- 深度推演承担“为什么成立、接下来可能怎么演化”的增强职责

首版采用 `双层接入`：

1. 在现有股票 / 期货详情链中展示一张 `深度推演摘要卡`
2. 提供独立的 `深度推演详情页`，承接完整报告、备选情景、运行日志和 `VIP` 正文

这不是新增一级导航模块，也不是把管理端的 `Forecast Lab` 直接搬到客户端，而是把 `L3` 作为客户端阅读链中的增强层。

## Background

当前仓库和后端能力已经具备以下基础：

- `/Users/gjhan21/cursor/sercherai/newclient/src/api/forecast.js` 已经提供用户侧 `forecast run` 读写接口
- 后端用户侧接口已支持：
  - `POST /forecast/runs`
  - `GET /forecast/runs`
  - `GET /forecast/runs/:id`
- 后端在读侧已支持把最新可读的深推演摘要挂到 explanation 上，核心字段包括：
  - `deep_forecast_summary`
  - `deep_forecast_report_ref`
- 旧版 `/Users/gjhan21/cursor/sercherai/client` 已经存在成熟的 `PcForecastRunView`、`H5ForecastRunView` 和深推演摘要格式化逻辑，可作为行为参考

但 `newclient` 目前仍处于“接口已有、用户体验未接入”的状态：

- `newclient` 还没有用户可见的深推演详情页
- 现有股票分析、股票详情、期货详情页还没有把深推演摘要嵌入进去
- `newclient` 的页面结构是分散的，不存在一个天然统一的“推荐详情页”可直接承接所有深推演信息

因此本设计的目标不是发明新能力，而是把已有后端能力和旧版交互经验，以符合 `newclient` 结构的方式落到正式客户端。

## Goal

在 `newclient` 中新增一条完整、可理解、可扩展的客户端深度推演阅读链，使用户能够：

1. 在现有股票 / 期货页面中直接看到深推演是否存在以及核心结论
2. 在需要时进入独立详情页查看完整推演结果
3. 在 `PC` 与 `H5` 上以一致语义使用该能力
4. 在 `STOCK` 与 `FUTURES` 两种标的类型上共享同一套深推演模型
5. 在 `VIP` 权限边界下保留“摘要免费、正文按权限”的产品平衡

## Non-Goals

本设计明确不做以下内容：

- 不新增首页一级导航“深度推演”
- 不新增客户端“深度推演列表页”
- 不让用户在首版中主动创建新的深推演 `run`
- 不把深度推演做成与主推荐链并列的主模块
- 不要求先重构 `newclient` 的所有详情页为统一详情框架
- 不沿用管理端的 `Forecast Lab` 命名与工作台语义
- 不把旧 `/Users/gjhan21/cursor/sercherai/client` 的页面样式整体照搬到 `newclient`

## User Decisions Locked In

用户在设计阶段已明确确认以下边界：

- 模块要接入 `http://127.0.0.1:5275/` 对应的正式客户端，也就是 `/Users/gjhan21/cursor/sercherai/newclient`
- 产品形态采用 `摘要卡 + 独立详情页` 的双层结构
- `PC` 与 `H5` 都要纳入本次设计
- `股票` 与 `期货` 都要支持
- 深度推演在客户端中是 `增强层`，不是新的一级主模块

## Product Contract

### 1. 客户端名称与定位

客户端对用户统一展示名称为 `深度推演`。

可以在页面小字中保留 `Forecast L3` 作为能力标签，但主文案不直接使用 `Forecast Lab`，避免管理端实验工作台语义进入客户端。

### 2. 页面层级

首版页面层级固定为两层：

- `第一层：摘要卡`
  - 嵌入到现有股票 / 期货详情链中
  - 负责告诉用户“有没有深推演、现在是什么状态、核心结论是什么”
- `第二层：深度推演详情页`
  - 展示完整结构化报告
  - 展示备选剧本、触发清单、角色分歧和运行日志
  - 承担 `VIP` 正文边界

首版不再继续拆第三层，例如“推演历史页”或“推演列表页”。

### 3. 客户端路由命名

客户端独立详情页不沿用管理端的 `forecast-lab` 命名，而使用更产品化的路由：

- `PC`：`/forecast/:id`
- `H5`：`/m/forecast/:id`

这样做的原因是：

- `lab` 更像后台工作台，不像客户端用户语言
- 客户端页面本质是“深度推演报告详情”，不是实验室控制台
- 后续若要承接分享、消息通知或历史跳转，`/forecast/:id` 更自然

### 4. 阅读链原则

客户端中的深度推演始终遵循：

1. 先读主链
2. 再读摘要
3. 最后决定是否进入深推演详情页

换句话说，`L3` 负责增强解释，不负责替代主推荐链。

## Information Architecture

## 1. 模块不是一级导航

深度推演不作为新的一级导航入口出现。

原因：

- 它不是一个独立内容宇宙，而是依附于推荐 / 策略阅读链的增强层
- 强行独立成一级模块会破坏“主链 + 增强链”的产品关系
- 首版没有列表页与主动发起能力，独立导航会造成空心模块

## 2. 首版接入入口

### PC 端

首版优先接入以下页面：

- `/Users/gjhan21/cursor/sercherai/newclient/src/apps/pc/views/analysis/StockAnalysis.vue`
- `/Users/gjhan21/cursor/sercherai/newclient/src/apps/pc/views/futures/FuturesStrategyDetail.vue`

其中：

- 股票分析页承接 `STOCK` 深推演摘要
- 期货策略详情页承接 `FUTURES` 深推演摘要

### H5 端

首版优先接入以下页面：

- `/Users/gjhan21/cursor/sercherai/newclient/src/apps/h5/views/StockDetail.vue`

H5 股票详情页承接 `STOCK` 深推演摘要。

对于 `FUTURES`：

- 能力层面必须支持 `FUTURES`
- 独立详情页必须支持 `FUTURES run`
- 但当前 `newclient` 尚不存在与 `PC` 对齐的 `H5` 期货策略详情壳

因此首版策略是：

- 先保证 `FUTURES` 可在独立详情页中被完整阅读
- `H5` 侧的摘要入口优先使用现有期货详情宿主页面承接
- 当前代码库中如果只有 `/Users/gjhan21/cursor/sercherai/newclient/src/apps/h5/views/futures/FuturesArbitrageDetail.vue` 这类期货详情壳，则可先放置最小摘要入口或深链接按钮
- 一旦后续出现正式 `H5` 期货策略详情页，复用相同的摘要组件迁移即可

这保证了“股票与期货都支持”的产品边界与当前代码现状不冲突。

## 3. 摘要卡放置位置

### PC 股票分析页

摘要卡应插入到解释性内容区，而不是行情或输入区。

推荐位置：

- 放在“AI 分析师意见 / 场景推演”附近
- 与当前 explanation 内容形成连续阅读流

### PC 期货策略详情页

摘要卡放在当前“AI 深度分析”区块之后，作为进一步增强的结论层。

### H5 股票详情页

摘要卡放在“AI 快评”之后，以轻量单列卡片形式展示。

### H5 期货详情页

摘要卡只做最小承接，不在首版追求与 `PC` 同级的信息密度。

## Data Contract

## 1. 摘要不是单独首跳接口

首版不为摘要额外设计新的首屏接口。

各页面继续使用自己的主接口：

- 股票分析页继续使用现有股票 insight / recommendation 接口
- 期货策略详情继续使用现有 futures strategy detail / insight 接口
- H5 股票详情继续使用现有推荐 / 详情接口

然后从这些已有返回中提取深推演增强字段：

- `deep_forecast_summary`
- `deep_forecast_report_ref`

这符合后端当前设计：深推演是附着在主链 explanation 上的增强信息，不是独立主链。

## 2. 独立详情页才拉 run detail

只有当用户进入 `/forecast/:id` 或 `/m/forecast/:id` 时，前端才调用：

- `getForecastRunDetail(id)`

这样设计有三个直接收益：

1. 首屏不需要挂复杂轮询逻辑
2. 主链页面不会被 `run` 状态机污染
3. `QUEUED / RUNNING / FAILED / VIP` 等复杂状态集中在一个页面处理

## 3. 首版不主动使用 create/list 接口

虽然 `/Users/gjhan21/cursor/sercherai/newclient/src/api/forecast.js` 已经提供：

- `createForecastRun`
- `listForecastRuns`
- `getForecastRunDetail`

但首版客户端阅读链仅以 `getForecastRunDetail` 为主。

原因：

- 首版目标是读能力接入，不是创建能力接入
- 一旦开放主动创建 run，会带来配额、成本、等待预期与队列管理问题
- 列表页不在首版范围内，因此 `listForecastRuns` 暂不作为主路径能力

这并不否定后续扩展，而是对首版进行明确收口。

## 4. 无摘要时的处理

当以下任一情况出现时，摘要模块应优雅降级：

- 后端未开启 client read side
- 当前对象没有可读的深推演摘要
- explanation 中没有 `deep_forecast_summary`
- explanation 中没有可用的 `report_ref`

降级行为应为：

- 不显示摘要卡
- 不破坏原有推荐 / 分析 / 期货详情页布局
- 不额外显示重提示错误

只有在“已知存在 run 但仍在生成 / 失败”的场景下，才显示状态型摘要卡。

## Frontend Module Design

`newclient` 当前并不是统一详情框架，因此首版采用 `可插拔模块` 设计。

## 1. 共享格式化层

建议在 `newclient` 内新增一层共享的深推演格式化工具，用于把不同来源的数据统一成摘要视图模型。

建议职责包括：

- 解析 `deep_forecast_summary`
- 解析 `deep_forecast_report_ref`
- 输出统一字段：
  - `status`
  - `statusLabel`
  - `summary`
  - `scenario`
  - `actionGuidance`
  - `runId`
  - `requiresVip`
  - `generatedAt`
  - `targetType`

该层在语义上应继承旧版 `/Users/gjhan21/cursor/sercherai/client/src/lib/strategy-version.js` 中的深推演摘要构建规则，但不要求直接照搬实现。

## 2. 共享摘要组件

建议新增一个共享摘要卡组件，例如：

- `DeepForecastSummaryCard`

职责：

- 只负责展示摘要内容
- 不负责发请求
- 不关心股票还是期货
- 通过 props 控制 `PC / H5` 的显示密度

摘要卡需要支持至少 4 种展示态：

- 可读摘要
- 排队中 / 推演中
- 推演失败
- `VIP` 正文可锁但摘要可读

## 3. 页面适配层

由于股票分析页、期货策略详情页、H5 股票详情页的数据形状不同，建议使用一层轻量适配逻辑，例如：

- `useDeepForecastEntry`

职责：

- 接受页面原始数据
- 调用共享格式化层生成摘要对象
- 生成摘要卡 CTA 跳转链接
- 处理“当前页面返回哪里”的上下文参数

这样每个页面只需要提供自己的原始 source，而不需要各自重复理解深推演字段。

## 4. 独立详情容器

建议新增独立深推演详情容器，负责：

- 拉取 `run detail`
- 处理轮询
- 处理 `VIP` 正文门槛
- 渲染结构化报告
- 渲染运行日志

该容器在行为上参考旧版：

- `/Users/gjhan21/cursor/sercherai/client/src/apps/pc/views/PcForecastRunView.vue`
- `/Users/gjhan21/cursor/sercherai/client/src/apps/h5/views/H5ForecastRunView.vue`

但视觉风格必须服从 `newclient` 当前风格，不直接复制旧版样式。

## Interaction States

## 1. 摘要卡状态

### 可读摘要

这是首选状态。

摘要卡展示：

- 当前状态
- 主情景
- 动作建议
- 更新时间
- CTA：`查看完整深度推演`

### 排队中 / 推演中

摘要卡只做轻提示：

- `AI 正在深度推演中`
- 展示状态文案
- 提供 `查看进度` 或 `查看详情` CTA

主链页面不在该状态下轮询。

### 推演失败

摘要卡只展示：

- `本次深度推演未成功完成`

并提供进入详情页查看失败信息或刷新结果的入口。

主链页面不承担失败解释职责。

### VIP 正文锁定

只要摘要存在，摘要卡就可以正常展示。

锁定的只有正文区域，不是摘要区域。

## 2. 详情页状态机

深推演详情页负责处理完整状态机：

### 初始加载

- 首次进入时调用 `getForecastRunDetail(id)`

### 运行中

如果状态为：

- `QUEUED`
- `RUNNING`

则详情页应轮询刷新，推荐间隔为 `5s` 左右。

### 失败

如果状态为 `FAILED`：

- 展示失败说明
- 展示刷新按钮
- 保留返回原始推荐 / 策略页面的能力

### 成功

如果状态已稳定且存在 report：

- 先展示结构化执行摘要
- 再展示备选情景
- 再展示触发清单
- 再展示角色分歧
- 最后展示正文与日志

## VIP Contract

首版 `VIP` 设计采用：

- `摘要免费`
- `正文按权限`

### 免费用户可见内容

- 摘要
- 主情景
- 动作建议
- 状态
- 关键日志摘要

### VIP 用户额外可见内容

- 完整正文
- 更完整的结构化细节
- 更完整的阅读链承接

### 权限判定

前端沿用现有：

- `/Users/gjhan21/cursor/sercherai/newclient/src/api/membership.js`
  - `getMembershipQuota()`

不在首版新增新的权限接口。

## Platform-Specific Behavior

## 1. PC

`PC` 是首版完整度最高的宿主。

应完整支持：

- 股票摘要入口
- 期货策略摘要入口
- 独立深推演详情页
- `VIP` 正文边界
- 运行中轮询
- 失败态与日志态

## 2. H5

`H5` 语义与 `PC` 一致，但密度更轻。

应支持：

- 股票摘要入口
- 独立深推演详情页
- `FUTURES run` 的详情阅读能力

对于 `H5` 期货入口，首版允许采取“最小可用承接”：

- 如果当前只有已有期货详情壳可挂入口，则放置轻量摘要 / CTA
- 后续出现正式 `H5` 期货策略详情壳后，直接复用相同摘要组件迁移

这样既满足“股票 + 期货”都支持，又不强迫首版先补一整套新的 `H5` 期货详情体系。

## Old Client Reuse Policy

旧 `/Users/gjhan21/cursor/sercherai/client` 中已有成熟经验，但 `newclient` 接入必须遵守以下复用策略：

- 复用 `行为语义`
- 复用 `字段语义`
- 复用 `状态机经验`
- 不直接复制旧版样式与页面壳

特别是以下旧版经验应被明确继承：

- 深推演摘要字段的标准化逻辑
- `QUEUED / RUNNING / FAILED / SUCCESS` 的前端解释方式
- `VIP` 正文锁定而摘要可读的交互边界
- 独立详情页中的轮询与日志承接方式

## Risks

## 1. 页面数据形状不一致

股票分析、股票详情、期货策略详情的返回结构不完全相同。

如果不加适配层，深推演逻辑会在多个页面中重复出现，后续难以维护。

## 2. H5 期货宿主不足

当前 `newclient` 中没有与 `PC` 对齐的 `H5` 期货策略详情宿主。

如果不在设计中明确“最小承接位”策略，实施阶段很容易在“必须支持期货”和“没有合适页面可挂”的矛盾之间反复拉扯。

## 3. 把深推演做成主模块的诱惑

一旦新增独立详情页，容易继续膨胀出列表页、一级导航和主动发起能力。

这会把首版从“读侧增强接入”膨胀成另一条产品线，偏离当前目标。

## 4. VIP 边界处理不当

如果连摘要一起锁掉，用户没有点击动机；

如果全文都免费，又会削弱会员价值。

因此必须守住“摘要免费、正文按权限”的中线设计。

## Test Plan

首版至少需要覆盖以下验证：

### 1. 共享格式化层

- `STOCK` 摘要可正确标准化
- `FUTURES` 摘要可正确标准化
- `QUEUED / RUNNING / FAILED / SUCCESS` 标签正确
- `requires_vip` 语义正确

### 2. 摘要组件

- 可读摘要态渲染正确
- 运行中态渲染正确
- 失败态渲染正确
- 无摘要时页面不应被破坏

### 3. 独立详情页

- `getForecastRunDetail` 正常渲染
- 运行中状态可轮询
- 失败态可展示
- `VIP` 正文锁定正确
- 返回路径正确

### 4. 路由与入口

- `PC` `/forecast/:id` 可进入详情页
- `H5` `/m/forecast/:id` 可进入详情页
- 股票入口可跳转
- 期货入口可跳转

### 5. 回归验证

- 没有深推演数据时，股票分析 / 股票详情 / 期货详情页面不出现破坏性回归
- `newclient` 现有主链页面依旧能在未登录 / 无摘要数据时正常工作

## Recommended Implementation Direction

建议按以下顺序实施：

1. 在 `newclient` 中补共享深推演摘要格式化层
2. 新增 `PC` / `H5` 独立深推演详情页路由
3. 复刻旧版详情页的状态机语义到 `newclient`
4. 把摘要卡接到 `PC` 股票分析页
5. 把摘要卡接到 `PC` 期货策略详情页
6. 把摘要卡接到 `H5` 股票详情页
7. 给 `H5` 期货链补最小可用入口
8. 增加共享单测与页面级回归测试

这个顺序可以优先把最关键的“共享模型 + 独立详情页”立起来，再逐步接入各宿主页面，降低页面分散结构带来的实施风险。

## Exit Criteria

当以下条件全部满足时，可认为该设计达成目标：

1. `newclient` 中存在用户可访问的深推演详情页
2. `PC` 与 `H5` 都能进入深推演详情页
3. `STOCK` 与 `FUTURES` 都能被统一深推演模型承接
4. 至少一个股票宿主和一个期货宿主已接入摘要卡
5. 摘要卡不会在无数据时破坏原页面
6. `VIP` 正文边界与摘要可读边界清晰成立
7. 深推演在客户端中仍然是增强层，而不是新的一级主模块
