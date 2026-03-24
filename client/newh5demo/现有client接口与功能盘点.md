# 现有 client 接口与功能盘点

## 1. 文档用途

这份文档不是 PRD，也不是未来能力清单，而是给 `/newclient/h5demo` 做设计时使用的“真相源”。

用途只有一个：在开始设计新客户端 demo 之前，先把当前主项目 `client` 已经真实存在的接口、路由、页面职责和移动端压缩承接关系整理清楚，避免后续 demo 设计时凭感觉扩能力。

本文档默认遵守以下边界：

- 只盘点当前主项目 `client` 中已经存在的前端能力。
- 只盘点已经存在的 API 包装函数、真实页面和真实路由。
- 不把 worktree / 分支中的实验模块算进当前能力。
- 不推测后端未来会返回什么字段；这里只以当前前端已调用的能力为准。
- 不把“适合做 demo 的想法”写成“当前已经支持的能力”。

明确说明：

- 社区 / 聊天室 / 互动讨论模块 **不属于当前主 client 的既有能力**。
- 当前 H5 端并不是一套与 PC 完全对等的产品，而是对主站能力做过压缩承接。

## 2. 本文档的源码依据

本次整理主要核对了以下代码：

- `client/src/router/index.js`
- `client/src/apps/pc/router/index.js`
- `client/src/apps/h5/router/index.js`
- `client/src/api/auth.js`
- `client/src/shared/api/auth.js`
- `client/src/api/market.js`
- `client/src/api/news.js`
- `client/src/api/membership.js`
- `client/src/api/userCenter.js`
- `client/src/views/RecommendationArchiveView.vue`
- `client/src/views/MyWatchlistView.vue`
- `client/src/apps/h5/views/H5AuthView.vue`
- `client/src/apps/h5/views/H5HomeView.vue`
- `client/src/apps/h5/views/H5StrategyView.vue`
- `client/src/apps/h5/views/H5NewsView.vue`
- `client/src/apps/h5/views/H5MembershipView.vue`
- `client/src/apps/h5/views/H5ProfileView.vue`
- `client/src/apps/h5/lib/*`

## 3. 当前 client 的真实前端结构

### 3.1 路由入口关系

- `client/src/router/index.js` 当前直接 `export { default } from "../apps/pc/router";`
- 这表示当前主入口默认接到 **PC 路由**
- H5 路由已经单独存在于 `client/src/apps/h5/router/index.js`
- 因此现在的真实结构不是“一套路由兼容 PC/H5”，而是“PC / H5 两套路由层，复用同一批 API 和部分共享逻辑”

### 3.2 PC 路由总览

| 路由 | 页面 | 路由级鉴权 | 说明 |
| --- | --- | --- | --- |
| `/auth` | 登录/注册页 | `guestOnly` | 已登录后会按 `redirect` 回跳 |
| `/home` | 首页 | 无 | 主站首页 |
| `/strategies` | 策略页 | 无 | 股票推荐、期货策略、市场事件主工作台 |
| `/archive` | 历史档案页 | 无 | 历史推荐结果、版本轨迹、复盘 |
| `/watchlist` | 我的关注页 | 无 | 本地关注、变化追踪、快速移除 |
| `/news` | 资讯页 | 无 | 分类、文章、附件、权限承接 |
| `/membership` | 会员页 | `requiresAuth` | 会员产品、订单、支付、实名激活链路 |
| `/profile` | 个人中心 | `requiresAuth` | 账户、实名、消息、邀请、订阅等 |

补充：

- `/invite/:inviteCode` 会跳到 `/auth`，并带上 `invite_code` 与默认 `redirect=/home`
- 未命中路由会回 `/home`

### 3.3 H5 路由总览

| 路由 | 页面 | 路由级鉴权 | 说明 |
| --- | --- | --- | --- |
| `/auth` | H5 登录/注册页 | `guestOnly` | 已登录后会按 `redirect` 回跳 |
| `/home` | H5 首页 | 无 | 首页主线压缩版 |
| `/strategies` | H5 策略页 | 无 | 策略流 + 详情单列阅读 |
| `/news` | H5 资讯页 | 无 | 资讯流、搜索、正文、附件、权限 |
| `/membership` | H5 会员页 | `requiresAuth` | H5 收银台 |
| `/profile` | H5 我的页 | `requiresAuth` | H5 账户中心 |
| `/archive` | 重定向 | 无 | 直接跳转 `/news` |
| `/watchlist` | 重定向 | 无 | 直接跳转 `/strategies` |

补充：

- H5 端也支持 `/invite/:inviteCode -> /auth`
- H5 当前 **没有独立的历史档案页和关注页**
- 这两个能力在 H5 中被压缩承接到 `/news` 和 `/strategies`

## 4. API 接口清单

说明：

- 下表只写当前前端已经封装出来的 API 函数。
- “前端鉴权特征”写的是前端代码里的使用方式，不等同于后端完整权限规则。
- 大多数列表接口都复用了 `buildParams`，会自动过滤 `undefined / null / ""`。

### 4.1 认证 API

`client/src/api/auth.js` 只是转出 `client/src/shared/api/auth.js`。

| 函数 | 方法 + 路径 | 前端鉴权特征 | 典型用途 |
| --- | --- | --- | --- |
| `register(payload)` | `POST /auth/register` | 游客态调用 | 注册 |
| `login(payload)` | `POST /auth/login` | 游客态调用 | 登录 |
| `refreshToken(refreshToken)` | `POST /auth/refresh` | 会话续期 | 刷新 token |
| `logout(refreshToken)` | `POST /auth/logout` | 登录态调用 | 当前设备退出 |
| `logoutAll()` | `POST /auth/logout-all` | 登录态调用 | 全部设备退出 |
| `getAuthProfile()` | `GET /auth/me` | 登录态调用 | 获取当前认证资料 |

设计结论：

- 当前认证能力就是“账号密码 + 邀请码绑定 + redirect 回跳”
- 没有第三方登录
- 没有短信验证码登录

### 4.2 市场 / 策略 / 推荐 API

来源文件：`client/src/api/market.js`

| 函数 | 方法 + 路径 | 前端鉴权特征 | 典型用途 |
| --- | --- | --- | --- |
| `listStockRecommendations(params)` | `GET /stocks/recommendations` | 首页、策略页、会员页、档案页直接调用 | 股票推荐列表 |
| `getStockRecommendationDetail(id)` | `GET /stocks/recommendations/:id` | 详情读取 | 股票推荐详情 |
| `getStockRecommendationPerformance(id)` | `GET /stocks/recommendations/:id/performance` | 详情读取 | 推荐表现、涨跌/收益相关展示 |
| `getStockRecommendationInsight(id)` | `GET /stocks/recommendations/:id/insight` | 详情读取 | 推荐理由、解释、证据等 |
| `getStockRecommendationVersionHistory(id)` | `GET /stocks/recommendations/:id/version-history` | 详情读取 | 版本轨迹、理由变化 |
| `listFuturesArbitrage(params)` | `GET /futures/arbitrage` | 首页直接调用 | 期货套利列表 |
| `getFuturesGuidance(contract)` | `GET /futures/guidance/:contract` | 策略详情读取 | 合约 guidance / 执行说明 |
| `listFuturesStrategies(params)` | `GET /futures/strategies` | 策略页、H5 策略页直接调用 | 期货策略列表 |
| `getFuturesStrategyInsight(id)` | `GET /futures/strategies/:id/insight` | 详情读取 | 期货策略解释 |
| `getFuturesStrategyVersionHistory(id)` | `GET /futures/strategies/:id/version-history` | 详情读取 | 期货策略版本历史 |
| `listMarketEvents(params)` | `GET /market/events` | 策略页、H5 策略页直接调用 | 市场事件列表 |
| `getMarketEventDetail(id)` | `GET /market/events/:id` | 详情读取 | 市场事件详情 |

设计结论：

- 当前 `client` 的“策略能力”不是单一股票列表，而是三类内容并存：
  - 股票推荐
  - 期货策略 / guidance
  - 市场事件
- 首页还额外用了 `futures/arbitrage`
- 因此后续 `h5demo` 设计时，策略页可以做信息架构压缩，但不能只剩“股票推荐”一种内容

### 4.3 资讯 API

来源文件：`client/src/api/news.js`

| 函数 | 方法 + 路径 | 前端鉴权特征 | 典型用途 |
| --- | --- | --- | --- |
| `listNewsCategories()` | `GET /public/news/categories` | 公共接口 | 获取资讯分类 |
| `listNewsArticles(params)` | `GET /public/news/articles` | 公共接口 | 资讯列表、分类筛选、关键词搜索 |
| `getNewsArticleDetail(id)` | 已登录：`GET /news/articles/:id`；未登录：`GET /public/news/articles/:id` | 根据登录态自动切换 public/private path | 文章详情 |
| `listNewsAttachments(articleID)` | 已登录：`GET /news/articles/:id/attachments`；未登录：`GET /public/news/articles/:id/attachments` | 根据登录态自动切换 public/private path | 附件列表 |
| `getAttachmentSignedURL(attachmentID)` | `GET /news/attachments/:id/signed-url` | 通常在详情内下载附件时调用 | 获取附件下载签名地址 |

补充：

- H5 资讯页当前已经真实接入“关键词搜索”，调用的仍然是 `listNewsArticles(params)`，前端透传参数中已使用：
  - `page`
  - `page_size`
  - `keyword`
  - `category_id`
- 前端的权限承接逻辑非常明确：
  - 游客可看公开列表和公开详情
  - 登录后继续原阅读位置
  - VIP 内容需要更高权限才解锁正文和附件

设计结论：

- 资讯搜索目前是“资讯域内搜索”，不是全站搜索
- 资讯页天然支持“新闻 / 研报 / 期刊等栏目切换”，但具体栏目仍以后端返回分类为准
- “研报”可以在 demo 中强化视觉层级，但不能设计成需要新接口的新内容体系

### 4.4 会员 API

来源文件：`client/src/api/membership.js`

| 函数 | 方法 + 路径 | 前端鉴权特征 | 典型用途 |
| --- | --- | --- | --- |
| `listMembershipProducts(params)` | `GET /membership/products` | 会员页直接调用 | 套餐产品列表 |
| `listMembershipOrders(params)` | `GET /membership/orders` | 会员页、个人中心调用 | 订单列表 |
| `getMembershipQuota()` | `GET /membership/quota` | 会员页、首页、资讯页、个人中心等调用 | 会员等级、实名状态、阅读/权限状态 |
| `createMembershipOrder(payload)` | `POST /membership/orders` | 登录后发起 | 创建订单 |
| `triggerPaymentCallback(channel, payload)` | `POST /payment/callbacks/:channel` | 支付链路使用 | 支付回调模拟 / 触发支付结果处理 |

设计结论：

- 当前会员页不是纯展示价格页，而是已经包含：
  - 套餐列表
  - 当前状态
  - 下单
  - 支付回调
  - 实名激活承接
  - 最近订单
- 后续 demo 可以改信息层级，但不能把支付、订单、实名激活链拆掉

### 4.5 用户中心 API

来源文件：`client/src/api/userCenter.js`

| 函数 | 方法 + 路径 | 前端鉴权特征 | 典型用途 |
| --- | --- | --- | --- |
| `getUserProfile()` | `GET /user/profile` | 登录后调用 | 获取个人资料 |
| `submitKYC(payload)` | `POST /user/kyc/submit` | 登录后调用 | 实名提交 |
| `getMembershipQuota()` | `GET /membership/quota` | 登录后调用 | 账户页读取会员/实名状态 |
| `listMembershipOrders(params)` | `GET /membership/orders` | 登录后调用 | 账户页订单列表 |
| `listRechargeRecords(params)` | `GET /user/recharge-records` | 登录后调用 | 充值记录 |
| `listBrowseHistory(params)` | `GET /user/browse-history` | 登录后调用 | 浏览历史 |
| `listMessages(params)` | `GET /messages` | 登录后调用 | 站内消息 |
| `readMessage(id)` | `PUT /messages/:id/read` | 登录后调用 | 消息已读 |
| `listShareLinks()` | `GET /user/share-links` | 登录后调用 | 我的分享链接 |
| `createShareLink(payload)` | `POST /user/share-links` | 登录后调用 | 创建邀请链接 |
| `listInviteRecords(params)` | `GET /user/share/invites` | 登录后调用 | 邀请记录 |
| `getInviteSummary()` | `GET /user/share/invite-summary` | 登录后调用 | 邀请概览统计 |
| `listSubscriptions(params)` | `GET /subscriptions` | 登录后调用 | 订阅列表 |
| `createSubscription(payload)` | `POST /subscriptions` | 登录后调用 | 创建订阅 |
| `updateSubscription(id, payload)` | `PUT /subscriptions/:id` | 登录后调用 | 更新订阅 |

补充：

- `userCenter.js` 中重复封装了 `membership/quota` 与 `membership/orders`，说明当前账户中心直接复用了会员域接口
- 当前个人中心不是“只改昵称头像”的轻资料页，而是一个账户管理台

## 5. 页面功能盘点

## 5.1 PC 页面职责

### `/auth`

- 登录 / 注册双场景
- 支持邀请码注册
- 支持 `redirect` 回跳
- 已登录用户进入时会自动跳回目标页或首页

### `/home`

- 首页不是纯门户，而是“推荐与阅读入口页”
- 当前真实数据来源至少包含：
  - 股票推荐
  - 期货套利
  - 资讯列表
  - 会员产品
- 作用更偏向：
  - 展示主推荐
  - 给出简化理由与行动入口
  - 承接资讯、策略、会员

### `/strategies`

- 当前 PC 策略页是主工作台之一
- 真实承载内容至少包括：
  - 股票推荐列表
  - 期货策略列表
  - 市场事件列表
  - 股票 insight / performance / version history
  - 期货 insight / version history
  - 市场事件详情
- 这页不能在新 demo 里被误设计成单一 feed 流产品

### `/archive`

- 历史推荐档案页
- 真实能力包括：
  - 历史推荐列表
  - 状态筛选
  - 来源标签
  - 历史表现与复盘
  - 版本差异
  - 版本轨迹
- 这是 PC 独有的完整档案工作台能力

### `/watchlist`

- 我的关注页
- 当前能力不是服务端自建组合，而是 **本地轻量关注 + 变化跟踪**
- 真实能力包括：
  - 保存的关注标的列表
  - 最新状态变化
  - 资讯变化
  - 风险边界变化
  - 结论变化
  - 版本/时间线对照
  - 快速移除

关键设计边界：

- 当前关注逻辑不是后端持久化 watchlist 产品
- 不能在新 demo 中直接扩写成“云端自选股社区”

### `/news`

- PC 资讯页是“列表 + 正文 + 权限承接”的完整阅读页
- 真实能力包括：
  - 分类切换
  - 列表读取
  - 详情正文
  - 附件
  - 权限锁
  - 游客 / 登录 / VIP 承接

### `/membership`

- 会员产品、配额、订单、支付、实名激活
- 当前不是营销落地页，而是功能性购买页

### `/profile`

- 当前更接近“账户管理台”
- 真实能力包括：
  - 用户资料
  - 会员状态 / quota
  - 订单
  - 浏览历史
  - 消息
  - 订阅
  - 分享链接
  - 邀请记录与邀请汇总
  - 实名提交

## 5.2 H5 页面职责

### `/home`

H5 首页已经不是 PC 首页的简单缩放，而是单列压缩结构。当前页面职责包括：

- 今日核心观点 Hero
- 今日脉冲摘要
- 主推荐理由三段式压缩
- 观察清单预览
- 一条焦点资讯入口
- 今日行动 CTA
- 主推荐一键加入 / 取消关注

真实数据来源包括：

- `listStockRecommendations`
- `getStockRecommendationDetail`
- `getStockRecommendationInsight`
- `listNewsArticles`
- `getMembershipQuota`

### `/strategies`

H5 策略页当前是“内容流 + 单列详情”的移动化承接：

- 股票 / 期货 / 事件栏目切换
- lead 卡片
- 继续阅读 feed
- 详情摘要
- 核心理由
- 执行区间
- 风险边界
- 版本与事件时间线

这说明 H5 端已经把 PC 的多面板工作台压缩成“先看结论，再进详情”的单列阅读逻辑。

### `/news`

H5 资讯页当前已经具备完整移动阅读链路：

- 搜索栏
- 分类切换
- 主阅读卡
- 内容流
- 正文详情
- 附件列表
- 附件下载
- 权限锁
- 登录后回原阅读位置

这页是后续 `h5demo` 最重要的真实设计基础之一。

### `/membership`

H5 会员页当前是移动收银台，不是营销长页：

- 当前状态
- 主推套餐
- 价值解释
- 套餐切换
- 支付通道
- 发起下单
- 最近订单
- 激活待办

### `/profile`

H5 我的页当前是移动账户中心：

- 账户身份概览
- 今日待办
- 常用入口
- 实名激活区
- 账户服务卡
- 消息提醒
- 邀请中心

### `/auth`

H5 登录注册页当前已具备：

- 登录 / 注册切换
- 邀请码识别
- redirect 回跳说明
- 来源场景提示
- 手机号 / 邮箱 + 密码认证

### H5 的压缩结论

H5 目前已经明确做了这两个取舍：

- `/archive` 压缩进 `/news`
- `/watchlist` 压缩进 `/strategies`

这代表后续做 `h5demo` 时要先分清：

- 哪些是“当前 H5 真实支持的页面”
- 哪些是“PC 有，但 H5 只是承接能力，没有独立页面”

## 6. H5 已存在的内容建模 / 展示层基础

`client/src/apps/h5/lib/` 已经有一批纯函数层，并且带有测试文件。这部分非常适合作为后续 `h5demo` 的设计基础，因为它们已经把真实接口数据压成移动端展示模型。

| 文件 | 当前作用 |
| --- | --- |
| `auth-page.js` | 认证页场景文案、邀请码与 redirect 归一化、初始模式判断 |
| `display-copy.js` | H5 标题与显示文案裁剪 / 格式化 |
| `formatters.js` | 日期、文案、VIP 阶段、文本裁剪等格式化工具 |
| `home-feed.js` | 首页 Hero、脉冲、观察清单、资讯简报建模 |
| `membership-cashier.js` | 会员方案、收银台、CTA 相关建模 |
| `news-feed.js` | 资讯栏目排序、lead 卡、ticker、feed 列表建模 |
| `page-state.js` | 页面权限状态、锁定态等 |
| `profile-center.js` | 账户 Hero、待办、快捷入口、服务卡、邀请与消息模型 |
| `shell-meta.js` | H5 壳层元信息 |
| `strategy-feed.js` | 策略列表、详情、版本/事件流建模 |
| `surface-tone.js` | 卡片高亮与视觉 tone 归类 |
| `mock-data.js` | demo fallback 数据 |

设计结论：

- 后续做 `h5demo` 时，优先复用这些“已被移动端验证过的信息压缩逻辑”
- 不建议一开始就重新发明一套完全不同的数据组织方式

## 7. 后续做 `h5demo` 时必须遵守的边界

### 7.1 只能基于当前真实能力设计

- 可以重做布局、结构、阅读顺序、视觉语言
- 不能把当前没有的后端能力直接画成“已支持功能”

### 7.2 资讯搜索目前只属于资讯域

- 当前真实已存在的是 `news` 域内搜索
- 不能在 demo 基础能力说明里写成“全站统一搜索”

### 7.3 关注页不是云端自选股系统

- 当前真实能力是本地关注 + 变化追踪
- 不是社区自选、不是多人共享组合、不是服务器持久 watchlist

### 7.4 H5 不是 PC 全功能镜像

- H5 当前没有独立 `/archive`
- H5 当前没有独立 `/watchlist`
- 如果后续 demo 想把这两个页面重新独立做出来，需要明确标记为“新规划”，不能写成“当前已有”

### 7.5 会员与个人中心都是真实业务页

- 会员页不仅有价格，还有订单、支付、实名激活链
- 个人中心不仅有资料，还有消息、邀请、订阅、实名、订单等账户管理能力

### 7.6 资讯权限链已经明确存在

- 游客：公开摘要 / 公开内容
- 登录：继续原阅读位置
- VIP：解锁更深正文与附件

后续设计时应强化这条链路，但不要重写成与现有接口不匹配的新权限模型。

### 7.7 社区模块不在当前盘点范围

再次强调：

- 聊天室
- 社区讨论
- 用户发帖
- 股票观点互动
- 讨论串

这些都不属于当前主 `client` 的既有能力，本文件不把它们算作 demo 设计基础。

## 8. 给 `h5demo` 的直接结论

如果接下来要在 `/newclient/h5demo` 先做新客户端 demo，最稳妥的设计基础应当是：

1. 以当前 `news`、`strategies`、`membership`、`profile`、`auth`、`home` 六类真实页面能力为主。
2. 以当前真实 API 分组为内容边界：
   - 认证
   - 策略 / 市场
   - 资讯
   - 会员
   - 用户中心
3. 以当前 H5 已存在的 `lib` 纯函数层为信息压缩基础。
4. 把 `/archive`、`/watchlist` 明确视为：
   - PC 有独立能力
   - H5 当前为压缩承接
5. 把“社区互动模块”视为未来新增模块，而不是当前 client 能力。

一句话总结：

后续 `h5demo` 的设计基础，不应当是“我们想做什么”，而应当是“当前 client 已经真实支持什么，以及移动端现在如何压缩承接这些能力”。
