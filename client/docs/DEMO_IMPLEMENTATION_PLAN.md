# Client Demo Implementation Plan

## Goal

把 `client` 真实前端逐页改造成与当前 demo 套件一致的中文 PC 金融站风格，同时保持：

- 只使用当前已经存在的路由、API 和数据字段
- 不新增假后端能力
- 同一套代码同时兼容 PC 与 H5
- 按页面逐步落地，确保每一页都能单独审查和回归

当前目标页面：

- `/home` -> `client/src/views/HomeView.vue`
- `/strategies` -> `client/src/views/StrategyView.vue`
- `/archive` -> `client/src/views/RecommendationArchiveView.vue`
- `/watchlist` -> `client/src/views/MyWatchlistView.vue`
- `/news` -> `client/src/views/NewsView.vue`
- `/membership` -> `client/src/views/MembershipView.vue`
- `/profile` -> `client/src/views/ProfileView.vue`
- `/auth` -> `client/src/views/AuthView.vue`

对应 demo：

- `client/china-style-supported-demo.html`
- `client/strategy-supported-demo.html`
- `client/archive-supported-demo.html`
- `client/watchlist-supported-demo.html`
- `client/news-supported-demo.html`
- `client/membership-supported-demo.html`
- `client/profile-supported-demo.html`
- `client/auth-supported-demo.html`

## Vibe Coding Work Mode

这轮开发采用“vibe coding but controlled”的方式推进：

1. 先冻结一版页面信息架构，以 demo 为唯一视觉和结构参考。
2. 每次只改一到两个真实页面，做成可运行的垂直切片。
3. 每完成一页，就做一次 PC/H5 审查，再进入下一页。
4. 不在同一轮同时重写布局壳层、页面结构和业务逻辑，避免失控。
5. 视觉层优先收敛，业务逻辑只做“保留并重组”，不做额外能力扩展。

## Current Reality

### Shared Layer

- 路由集中在 `client/src/router/index.js`
- 统一壳层在 `client/src/components/ClientLayout.vue`
- 全局样式入口在：
  - `client/src/styles/base.css`
  - `client/src/styles/tokens.css`
- 现有壳层已经包含：
  - 顶部导航
  - 页面级 context banner
  - 移动端底部 tab

### Current Technical Constraint

- 各页面目前多为“大而全”的单文件组件
- 大部分页面已经接入真实 API，不能用纯静态页面替换
- 部分页面已有移动端逻辑，尤其是 `NewsView.vue`
- 当前视觉语言偏“产品控制台”，和新 demo 的中文站风格存在明显差异

## Design Target

### Overall Visual Direction

- 中文 PC 财经站风格
- 信息主次明确，不做欧美 SaaS dashboard 感
- 首页与各承接页形成一致的“主内容区 + 右侧辅助区 + 底部承接说明”结构
- H5 不是独立一套页面，而是在同一套页面结构上做移动端压缩和阅读顺序重排

### Shared Principles

- 首页不再做入口大全，而是双主区：
  - 推荐区
  - 研报区
- 深内容必须继续承接到具体页面，不把所有内容塞回首页
- 历史复盘、变化跟踪、研报正文、会员转化、账户管理各自承担明确角色
- 不加入以下无后端承接内容：
  - 大盘指数总览
  - 题材热榜
  - 下载二维码
  - 开放平台门户
  - 社区评论/观点流
  - 直播/短视频

## Shared Refactor Plan

### Phase 0: Shared Foundation

目标：先把真实客户端的公共壳层和样式基础改到接近 demo 语气，减少每页重复重写成本。

#### 0.1 Layout Shell Refactor

文件：

- `client/src/components/ClientLayout.vue`

动作：

- 把当前 `context-banner` 从“每页统一强主视觉”改成更克制的壳层
- 保留路由级导航，但降低它对页面首屏的抢占感
- 用更贴近 demo 的顶部结构替换现在偏 SaaS 的 `top-nav`
- 保留移动端底部 tab，但统一到新的配色、字号和触达尺寸

原因：

- 新 demo 的页面主视觉应该由各页面自己控制
- 现在的 `ClientLayout` 横向信息太强，会和每个页面的新 hero 冲突

#### 0.2 Shared Style Layer

建议新增文件：

- `client/src/styles/finance-shell.css`
- `client/src/styles/finance-pages.css`

动作：

- 不直接把 demo 的静态 CSS 原样塞进每个页面
- 抽出共享模式：
  - 顶部站点头
  - 页面 hero
  - 主区/侧栏布局
  - 表格与卡片双态
  - footer 样式
  - H5 响应式断点

依赖关系：

- `client/src/main.js`
- `client/src/styles/base.css`
- `client/src/styles/tokens.css`

#### 0.3 Common Building Blocks

建议新增组件：

- `client/src/components/page/PageHero.vue`
- `client/src/components/page/MainSideLayout.vue`
- `client/src/components/page/SectionCard.vue`
- `client/src/components/page/ResponsiveDataTable.vue`
- `client/src/components/page/PageFooterNote.vue`

说明：

- 第一阶段不追求完美抽象
- 只抽最稳定的版式层
- 复杂业务内容仍留在原 view 内，避免组件切分过早造成理解成本

## Page-by-Page Delivery Order

推荐顺序：

1. `HomeView.vue`
2. `StrategyView.vue`
3. `RecommendationArchiveView.vue`
4. `MyWatchlistView.vue`
5. `NewsView.vue`
6. `MembershipView.vue`
7. `ProfileView.vue`
8. `AuthView.vue`

这个顺序的原因：

- 先做最关键的决策链路：首页 -> 策略 -> 档案 -> 关注 -> 资讯
- 再做账户链路：会员 -> 我的 -> 登录
- 这样每一轮审查都能沿着真实用户路径往后走

## Per-Page Implementation Scope

### 1. HomeView

文件：

- `client/src/views/HomeView.vue`

目标对应 demo：

- `client/china-style-supported-demo.html`

现状：

- 已有主推荐、观察清单、资讯联动、历史样本、期货套利、会员入口
- 但首页目前内容层级偏散，转化区块偏多，主次关系还不够强

实施动作：

- 让首屏只突出两大主区：
  - 股票/期货推荐
  - 研报/市场资讯/研报解读
- 将 `conversion-grid`、`home-search-banner`、部分权益信息降级到更靠后位置
- 用当前 `newsHighlights` 与 `NewsView` 的 `report` 能力实现首页“研报解读”摘要块
- 将历史推荐表述改成“样本预览”，完整说明继续导向 `/archive`
- 保留 `arbitragePlans` 与 `methods`，但收敛为推荐区右侧辅助模块

技术注意点：

- 不改动数据加载顺序，优先改 template 和 scoped CSS
- 保持已有 watchlist、membership 跳转逻辑

验收标准：

- 首页首屏不再像综合控制台
- 用户第一眼能明确感知“先看推荐，再看研报”
- 不丢失当前已有数据能力

### 2. StrategyView

文件：

- `client/src/views/StrategyView.vue`

目标对应 demo：

- `client/strategy-supported-demo.html`

现状：

- 已有推荐矩阵、期货策略、市场事件雷达、详情弹层
- 当前更像“列表 + 弹窗”的工具页

实施动作：

- 把股票推荐详情从强弹窗体验改为更接近 in-page 深读结构
- 股票推荐区变成“列表 + 选中详情”的主内容组织
- 期货策略维持次一级重要度，放到右侧或次主区
- 市场事件继续保留，但更强调“它影响哪条判断”
- 版本、相关资讯、业绩在详情区内整合，减少用户在弹窗内来回跳

技术注意点：

- 现有 `activeStockID` / `activeFuturesID` / `activeEventID` 交互应保留
- 弹窗若短期内不移除，可先弱化为桌面增强交互，移动端仍用页内结构优先

验收标准：

- 策略页读起来像研究页，而不是只会点开弹窗的列表页

### 3. RecommendationArchiveView

文件：

- `client/src/views/RecommendationArchiveView.vue`

目标对应 demo：

- `client/archive-supported-demo.html`

现状：

- 历史时间线、版本差异、版本轨迹、来源标识、表现数据都已经很完整
- 问题主要在阅读层级，不在能力缺失

实施动作：

- 优化首屏 hero，强调“信任建立页”
- 筛选器保留，但降到工具区
- 时间线卡片内部按顺序重排：
  - 结果
  - 当时理由
  - explanation
  - 风险与失效
  - 版本变化
  - 来源说明
  - 时间线
- 强化止盈/止损/失效样本的差异表达

技术注意点：

- 不删除版本轨迹与来源系统
- 不把复杂信息再压缩成只剩收益数字

验收标准：

- 用户能快速判断一条历史样本“为什么值得信”或者“为什么该警惕”

### 4. MyWatchlistView

文件：

- `client/src/views/MyWatchlistView.vue`

目标对应 demo：

- `client/watchlist-supported-demo.html`

现状：

- 已经有非常完整的变化系统：
  - 状态变化
  - 资讯变化
  - 风险边界变化
  - 结论变化
  - 版本对照
  - 多次变化记录

实施动作：

- 把“变化优先”做成页面第一原则
- 首屏先给摘要与优先级，而不是直接堆大卡
- 关注主列表按“今天最该先处理谁”来排列
- 将变化拆层显示，减少用户需要滚完整张卡才能读懂

技术注意点：

- 当前页数据能力深，改造时主要是信息重组
- 避免因为追求样式统一而削弱变化时间线

验收标准：

- 用户回到关注页时，能在几十秒内知道先处理哪只标的

### 5. NewsView

文件：

- `client/src/views/NewsView.vue`

目标对应 demo：

- `client/news-supported-demo.html`

现状：

- 已有分类、详情、附件、VIP 锁、移动端详情面板
- 整体能力完整，结构还可以更像资讯深读页

实施动作：

- 强化“新闻 / 研报 / 期刊”的栏目分工
- 将研报放到更强的独立位置，匹配首页研报区的导流预期
- 保留当前移动端详情面板逻辑
- 优化桌面端 feed/detail 双栏布局的视觉层级

技术注意点：

- `NewsView.vue` 已有 mobile 逻辑，不要推翻重来
- 优先在现有结构上替换视觉层和 spacing

验收标准：

- 用户能明显感受到资讯页是深读页，而不是简单 feed 列表

### 6. MembershipView

文件：

- `client/src/views/MembershipView.vue`

目标对应 demo：

- `client/membership-supported-demo.html`

现状：

- 已有很强的会员价值表达：
  - 节奏承接
  - 解释能力升级
  - 配额
  - 套餐
  - 订单
  - 支付

实施动作：

- 把“为什么值得升级”与“当前状态缺哪一步”放到主区
- 把配额、订单、套餐顺序重排，减少营销页感
- 强化待实名激活状态的可见性

验收标准：

- 会员页更像明确的转化承接页，而不是产品控制面板

### 7. ProfileView

文件：

- `client/src/views/ProfileView.vue`

目标对应 demo：

- `client/profile-supported-demo.html`

现状：

- 已经有行动板、查询中心、支付表格、KYC、订阅、邀请
- 业务能力很深，但管理台感还不够统一

实施动作：

- 明确“今日行动板 -> 待办中心 -> 查询中心 -> 快捷入口”的顺序
- 保持 VIP 查询与 KYC 流程完整
- 强化个人中心和会员页之间的角色差异

验收标准：

- 个人中心像账户管理台，而不是资料页 + 功能拼盘

### 8. AuthView

文件：

- `client/src/views/AuthView.vue`

目标对应 demo：

- `client/auth-supported-demo.html`

现状：

- 登录 / 注册 / 邀请 / redirect 逻辑已经清晰
- 当前视觉仍偏基础功能页

实施动作：

- 维持功能页定位，不做营销长页
- 优化 PC 双栏布局与 H5 单栏布局
- 强化 redirect 场景文案
- 将登录和注册并列呈现，但不改变核心流程

验收标准：

- 登录页简洁、快速、能接住来源回跳

## Responsive Plan

目标：同一套代码同时兼容 PC + H5。

要求：

- 桌面端优先保持 demo 的金融站阅读层次
- 移动端不单独维护一套页面
- 所有表格必须能退化为卡片式信息结构
- 头部导航在小屏下可横向滑动或折叠
- CTA 在小屏下必须整行可点

建议做法：

- 在 shared CSS 中统一定义断点：
  - `<= 1200px`
  - `<= 768px`
  - `<= 480px`
- 优先用 CSS 重排，不重写数据逻辑
- 复杂桌面交互在移动端可以降级，但不能失去核心信息

## Engineering Strategy

### Implementation Cadence

建议每一页都按以下节奏推进：

1. 先改模板结构
2. 再改样式层
3. 再补移动端细节
4. 再做回归

### Refactor Rule

- 第一阶段不要追求把所有页面拆成很多小组件
- 先把页面结构做对
- 第二阶段再考虑把重复区块抽成组件

### Regression Checklist

每完成一页都检查：

- 路由是否仍然可进
- API 是否仍然加载成功
- loading / error / empty / locked 状态是否还在
- 登录前后路径是否正确
- PC 与 H5 是否都可读

## Milestones

### Milestone A: Shared Shell + Home

- `ClientLayout.vue`
- `base.css` / `tokens.css` / shared finance styles
- `HomeView.vue`

输出：

- 首页与 demo 基本对齐
- 新页面壳层定稿

### Milestone B: Decision Chain

- `StrategyView.vue`
- `RecommendationArchiveView.vue`
- `MyWatchlistView.vue`
- `NewsView.vue`

输出：

- 推荐 -> 解释 -> 历史 -> 变化 -> 外部信息 这条链路基本打通

### Milestone C: Account Chain

- `MembershipView.vue`
- `ProfileView.vue`
- `AuthView.vue`

输出：

- 会员转化、账户查询、登录回跳链路风格统一

### Milestone D: QA + Cleanup

- 去掉明显重复样式
- 补可复用页面组件
- 检查路由、样式、滚动、移动端

## Risks

### Risk 1: Current Views Are Too Large

很多页面是大型单文件组件，直接重写容易一次性动太多。

应对：

- 先重构模板层，不大动业务数据层

### Risk 2: Layout Shell Conflicts With New Heroes

`ClientLayout.vue` 当前 banner 很强，会和每页新的 hero 冲突。

应对：

- 第一阶段先收敛 layout shell，再动页面

### Risk 3: Scoped CSS Duplication

每页 scoped CSS 很多，后续容易造成风格碎片化。

应对：

- 第一阶段先允许少量重复
- 第二阶段再抽共享样式

### Risk 4: Over-Expanding Beyond Supported Content

开发时容易不自觉把 demo 里的“感觉”理解成新需求。

应对：

- 每页改造前都对照 API 和现有字段
- 不新增接口依赖

## Acceptance Standard

只有同时满足下面几点，才算这一轮改造完成：

- 真实 `client/src/views/*.vue` 页面结构明显向 demo 靠拢
- 所有页面仍基于当前真实 API 和字段
- PC 与 H5 都可用且可读
- 首页和承接页之间的关系变得清晰
- 会员、个人中心、登录页不再是孤立后台样式

## Recommended Next Step

下一步直接进入 Milestone A：

1. 先改 `client/src/components/ClientLayout.vue`
2. 再建立 shared finance page 样式层
3. 然后重做 `client/src/views/HomeView.vue`

原因：

- 首页是整个产品调性的锚点
- `ClientLayout` 不先处理，后面每一页都会被旧壳层拖住
- 首页完成后，策略 / 档案 / 关注 / 资讯可以沿着同一条视觉链继续往下做
