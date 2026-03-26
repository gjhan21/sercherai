# Client PC/H5 Demo Alignment Design

**日期：** 2026-03-24
**范围：** `/Users/gjhan21/cursor/sercherai/client`

## 目标

将 `client/newh5demo` 中的 PC 与 H5 demo 页面落到现有 `client` 项目中，使 PC 端与 H5 端都具备统一的蓝金金融终端视觉、清晰的页面节奏，以及与 demo 对齐的信息组织方式，同时保留现有真实 API、认证、会员、实名、资讯权限和推荐数据链路。

这次改造的重点不是“复制静态 demo”，而是“让真实项目的页面结构、风格和交互节奏与 demo 对齐”。

## 背景

当前 `client` 已经拆分出双端结构：

- PC 端入口与路由位于 `client/src/apps/pc`
- H5 端入口与路由位于 `client/src/apps/h5`
- 现有共享 PC 页面主体位于 `client/src/views`
- H5 已经有独立布局、页面组件和一组数据整形 helper

现状上，H5 端已经比 PC 端更接近独立产品形态；PC 端仍更像“共享页面 + 桌面壳子”。同时，`client/newh5demo` 已经给出了首页、策略、关注、资讯、档案、会员、我的、登录注册八类页面在 PC/H5 双端的目标风格和信息顺序。

## 用户确认过的范围

- 需要同时改 PC 端和 H5 端
- H5 端不仅要改已有页面，还要补齐独立的 `watchlist` 和 `archive` 页面
- 不接受仅换肤或仅保留当前 H5 重定向结构

## 设计原则

### 1. Demo 对齐，但不复制静态 HTML

- 不将 `client/newh5demo/*.html` 原样搬入 Vue 页面
- 保留当前项目的数据流、路由、鉴权和状态判断
- 通过重构页面骨架、模块分组和样式体系，达到“真实项目长成 demo 的样子”

### 2. 真实能力优先

- 不为贴合 demo 而改动后端接口协议
- 不虚构当前系统并不支持的登录、会员或资讯权限能力
- 对于 demo 中存在但当前数据不足的展示，只做安全降级，不补假数据链路

### 3. 双端分开设计，统一品牌语言

- PC 与 H5 保持同一蓝金金融品牌气质
- PC 强调“终端 / 工作台 / 双栏深读”
- H5 强调“单列阅读 / 卡片节奏 / 底部导航 / 固定操作区”
- 不强行让两端结构完全一致

### 4. 信息组织优先于功能堆叠

- 首屏优先显示用户当下最需要理解和处理的内容
- 将“状态、待办、下一步动作”前置
- 把“历史记录、次级说明、管理明细”后置

## 非目标

- 不改 backend API
- 不改 shared auth/session/http 协议
- 不新增第三方登录、短信登录、社交入口
- 不改会员规则、实名规则、权限判定本身
- 不对这次目标无关的页面做大规模重构

## 产品与结构设计

## 整体结构

### PC 壳层

将现有 PC 端壳层从简化骨架升级为 demo 风格的金融研究终端：

- 顶部为深蓝渐变品牌栏
- 主导航使用圆角 pill 激活态
- 右侧展示用户状态、快捷操作或登录入口
- 去掉当前“第一阶段骨架已独立”的提示条
- 主内容区使用更宽的桌面画布、明确的主列/侧栏关系和更强的卡片层级

### H5 壳层

将现有 H5 壳层进一步对齐 demo 的移动端风格：

- 顶部保留品牌与场景说明，但视觉更贴近蓝金金融 App
- 页面容器强化 hero 卡、内容卡、操作区和 sticky CTA 的节奏
- 底部 tabbar 保持 5 项，不扩张成全页面目录
- `watchlist` 与 `archive` 作为独立页面存在，但不强制进入底部 tabbar
- 新增独立页必须有明确入口，不接受“只有路由、没有自然访问路径”的实现

## 页面映射

### PC 页面

#### `/home`

对应 demo：`home-pc-demo.html`

设计：

- 首屏为“今日主推荐 + 今日状态 + 主要动作”
- 主列承接推荐清单、资讯焦点和继续阅读入口
- 侧栏承接今日主推荐摘要、任务、关键状态
- 现有数据保留，但减少首屏并列块数量，避免信息过散

#### `/strategies`

对应 demo：`strategies-pc-demo.html`

设计：

- 首屏突出今日主推荐、核心理由、风控边界、执行参数
- 中段展示版本变化、支撑证据、跟踪状态和历史记录
- 页面从“信息密集控制台”调整为“主线解释 + 执行面板 + 复盘面板”

#### `/watchlist`

对应 demo：`watchlist-pc-demo.html`

设计：

- 页面定位为变化工作台 / 回访中心
- 强调加入关注后的持续跟踪、变化信号、下一步处理建议
- 保留已有关注逻辑和本地 watch 状态能力

#### `/news`

对应 demo：`news-pc-demo.html`

设计：

- 采用更明显的双栏结构：栏目 / 列表 与 正文 / 附件 / 权限承接
- 首屏突出焦点文章和“继续深读”链路
- 权限说明从技术态提示转成产品化阅读说明

#### `/archive`

对应 demo：`archive-pc-demo.html`

设计：

- 页面定位为历史推荐复盘中心
- 优先展示历史理由、结果表现、后续处理、版本变化
- 让“为什么当时推荐、后来发生了什么、今天怎么复盘”形成闭环

#### `/membership`

对应 demo：`membership-pc-demo.html`

设计：

- 首屏顺序调整为：当前状态、待办、升级价值、套餐与订单
- 价格与套餐不再抢占最前面的位置
- 强化“未实名 / 未激活 / 需要返回阅读链继续使用”的说明

#### `/profile`

对应 demo：`profile-pc-demo.html`

设计：

- 页面改造成账户管理台
- 首屏展示账户身份、会员状态、未读消息、邀请进度、今日待办
- 订单、阅读记录、消息、订阅、邀请等模块收束为查询/管理中心

#### `/auth`

对应 demo：`auth-pc-demo.html`

设计：

- 桌面端采用登录 / 注册双栏布局
- 来源说明、`redirect` 回跳、`invite_code` 信息前置展示
- 保留当前账号密码认证方式，不扩展第三方登录

### H5 页面

#### `/m/home`

对应 demo：`home-h5-demo.html`

设计：

- 继续使用“今日主线 + 核心观点 + 相关资讯 + 今日行动”节奏
- 对齐 demo 的 hero 卡视觉和区块顺序
- 强化单列浏览的节奏与卡片层次

#### `/m/strategies`

对应 demo：`strategies-h5-demo.html`

设计：

- 主打单列观点流
- 将核心结论、理由、风险边界压缩为移动端一眼可读的卡片段落
- 保留操作：看详情、加入关注、回到阅读链

#### `/m/watchlist`

对应 demo：`watchlist-h5-demo.html`

设计：

- 新增独立页面，不再重定向到 `/m/strategies`
- 页面定位为关注后的变化回访页
- 以单列列表展示已关注对象、状态变化、风险提示和推荐动作
- 主入口定义为：从 `/m/home` 的“延伸观察”区块进入 `watchlist`
- 次入口可来自 `/m/strategies` 的相关 CTA，但不是本次最小实现要求

#### `/m/news`

对应 demo：`news-h5-demo.html`

设计：

- 突出“焦点一条 + 后续内容流 + 权限承接”
- 保持文章列表与详情的移动端阅读节奏
- 让登录/会员/附件承接更自然地嵌入阅读链

#### `/m/archive`

对应 demo：`archive-h5-demo.html`

设计：

- 新增独立页面，不再重定向到 `/m/news`
- 页面定位为单列历史样本复盘页
- 展示历史推荐、结果、后续处理和版本变化摘要
- 主入口定义为：从 `/m/news` 的历史样本 / 深读延伸 CTA 进入 `archive`
- 次入口可来自策略页或我的页，但不是本次最小实现要求

#### `/m/membership`

对应 demo：`membership-h5-demo.html`

设计：

- 延续移动收银台结构
- 首屏展示当前状态、主推方案和主要动作
- 将权益说明和订单记录后置，但保留真实购买、激活和继续支付能力

#### `/m/profile`

对应 demo：`profile-h5-demo.html`

设计：

- 优先展示账户头部、会员状态、今日待办和高频入口
- 消息、邀请、实名、账户服务继续保留，但按移动端单列顺序重排

#### `/m/auth`

对应 demo：`auth-h5-demo.html`

设计：

- 保持当前 H5 登录/注册双模式切换
- 强化来源说明、邀请码状态和返回原页面的感知
- 文案与信息顺序更贴近 demo

## 技术设计

## 路由设计

### PC

PC 路由继续沿用现有结构，不新增页面类型，只对现有页面进行重排和重设计。

### H5

H5 路由调整如下：

- 文档中的 `/m/...` 表示线上 H5 访问地址；具体到 H5 router 与 auth redirect 处理时，内部仍使用 `/home`、`/watchlist`、`/archive` 这类 H5 应用内路径
- 移除 `/archive -> /news` 重定向
- 移除 `/watchlist -> /strategies` 重定向
- 新增独立 `H5ArchiveView`
- 新增独立 `H5WatchlistView`
- 更新 `client/src/apps/h5/lib/auth-page.js` 中的 legacy redirect 处理与场景文案，使未登录用户在认证完成后能正确返回新的 `/archive` 与 `/watchlist` 页面
- 更新 `client/src/apps/h5/lib/auth-page.test.mjs`，覆盖新的回跳与场景文案行为

同时更新 H5 壳层元信息，使新的页面拥有匹配的标题、副标题和入口文案。

## 组件与文件组织

### PC

主要改动文件：

- `client/src/apps/pc/layouts/PcLayout.vue`
- `client/src/apps/pc/styles/pc-shell.css`
- `client/src/views/HomeView.vue`
- `client/src/views/StrategyView.vue`
- `client/src/views/MyWatchlistView.vue`
- `client/src/views/NewsView.vue`
- `client/src/views/RecommendationArchiveView.vue`
- `client/src/views/MembershipView.vue`
- `client/src/views/ProfileView.vue`
- `client/src/views/AuthView.vue`

策略：

- 优先在页面内部做结构重排
- 尽量复用现有数据计算、行为处理和 API 调用
- 仅在必要时抽取小型局部展示组件，不做与目标无关的大拆分

### H5

主要改动文件：

- `client/src/apps/h5/layouts/H5Layout.vue`
- `client/src/apps/h5/styles/h5-shell.css`
- `client/src/apps/h5/styles/h5-ui.css`
- `client/src/apps/h5/router/index.js`
- `client/src/apps/h5/lib/auth-page.js`
- `client/src/apps/h5/lib/auth-page.test.mjs`
- `client/src/apps/h5/lib/shell-meta.js`
- `client/src/apps/h5/views/H5HomeView.vue`
- `client/src/apps/h5/views/H5StrategyView.vue`
- `client/src/apps/h5/views/H5NewsView.vue`
- `client/src/apps/h5/views/H5MembershipView.vue`
- `client/src/apps/h5/views/H5ProfileView.vue`
- `client/src/apps/h5/views/H5AuthView.vue`
- `client/src/apps/h5/views/H5WatchlistView.vue`（新增）
- `client/src/apps/h5/views/H5ArchiveView.vue`（新增）

配套新增 H5 视图模型文件，用于承接新增页面和重排后的映射逻辑，沿用现有 `home-feed.js`、`news-feed.js`、`profile-center.js` 的组织方式。

## 数据与状态设计

### 保留现有数据链路

- 推荐相关：继续使用现有市场推荐、详情、洞察与版本说明数据
- 资讯相关：继续使用现有新闻/研报/期刊拉取与权限判定
- 会员相关：继续使用当前会员产品、配额、订单、支付、实名状态链路
- 账户相关：继续使用现有账户、消息、邀请、订阅和本地会话能力

### 新增 H5 页面数据来源

#### `watchlist`

- 优先复用已有关注列表、本地关注状态和推荐补充信息
- 页面模型输出关注对象摘要、变化说明、风险状态和推荐动作

#### `archive`

- 优先复用现有档案页已存在的历史推荐、结果表现、版本变化和后续处理说明
- 页面模型输出适合移动端单列复盘的摘要结构

### 异常与降级

- 未登录时展示当前可读范围与后续解锁说明，而非空白或技术错误
- 接口失败时保留刷新按钮和卡片内告警提示
- 数据不足时使用真实空态与 fallback 数据，不虚构新的业务能力

## 视觉设计

- 主色调为深蓝、浅蓝灰与金色点缀
- PC 采用更明显的品牌顶栏、桌面画布和双栏信息组织
- H5 采用更强的 hero 卡、内容卡和 sticky CTA 组合
- 统一使用更克制的金融终端文案与状态表达
- 保留现有可访问性焦点态，避免仅靠颜色表达状态

## 测试与验证

### 单元测试

- 为新增 H5 视图模型补充 `.test.mjs`
- 对新增的 `watchlist` / `archive` 映射逻辑做数据分支验证

### 路由验证

- 验证 PC 端页面路由仍能正常访问
- 验证 `/m/watchlist` 与 `/m/archive` 成为独立页面
- 验证 `/m` H5 入口与 PC 入口仍能通过路径正确分流
- 验证未登录状态下从 H5 入口触发认证后，`redirect=/watchlist` 与 `redirect=/archive` 不再被改写到旧页面

### 构建验证

- 运行 `client` 相关测试和构建命令
- 至少确认 PC 与 H5 两端能成功编译
- 预期命令至少包含：
  - `npm --prefix client run build:all`
  - `node --test client/src/apps/h5/lib/auth-page.test.mjs`
  - `node --test` 运行新增 H5 视图模型测试文件

### 回归重点

- 登录注册来源回跳
- 邀请码展示与注册流程
- 会员页主 CTA、继续支付与实名激活链路
- H5 底部导航激活态与二级页面进入路径
- 资讯页正文 / 附件 / 权限承接

## 实施顺序

1. 改造 PC 与 H5 壳层视觉
2. 补齐 H5 `watchlist` 与 `archive` 路由、视图和视图模型
3. 重排 PC 端八类页面结构
4. 重排 H5 端现有页面结构与文案节奏
5. 统一细节样式、状态表达与 CTA
6. 跑测试与构建验证，修正回归

## 风险与取舍

### 风险 1：PC 页面体量较大

现有 PC 页面信息与逻辑较多，若一次性大拆容易引入回归。

取舍：

- 先保留数据与行为，优先重排模板结构
- 只在必要时做小范围抽取

### 风险 2：H5 新增独立页的数据适配不足

现有 H5 端没有独立 `watchlist` / `archive` 页面，可能需要从已有 PC 逻辑或共享逻辑中抽取部分数据组织方式。

取舍：

- 优先做视图模型层适配
- 不因为数据不足而退回重定向方案

### 风险 3：demo 与真实能力边界不一致

部分 demo 更偏展示稿，真实项目需要遵守当前权限和状态逻辑。

取舍：

- 真实能力优先
- 在视觉和信息组织上尽量贴近 demo，在交互与状态上尊重现有系统

## 结论

本次改造将以“壳层升级 + 页面重排 + H5 独立页补齐”为主线完成 `client` 的 PC/H5 demo 对齐。

推荐采用“复用真实数据与业务逻辑、重建页面骨架与视觉语言”的方式实施，而不是复制 demo 静态页面。这是当前成本、风险和长期维护性之间最稳妥的方案。
