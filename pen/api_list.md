# SercherAI Client 接口清单

本清单根据 `/client/src/api` 及相关共享 API 定义整理，涵盖了客户端应用所需的所有核心业务接口。

---

## 1. 认证模块 (Auth)
文件路径: `src/shared/api/auth.js` (由 `src/api/auth.js` 导出)

| 接口名称 | 方法 | 请求路径 | 功能描述 |
| :--- | :--- | :--- | :--- |
| `register` | POST | `/auth/register` | 用户注册 |
| `login` | POST | `/auth/login` | 用户登录 |
| `refreshToken` | POST | `/auth/refresh` | 刷新访问令牌 (Access Token) |
| `logout` | POST | `/auth/logout` | 用户退出登录 |
| `logoutAll` | POST | `/auth/logout-all` | 在所有设备/会话上退出登录 |
| `getAuthProfile` | GET | `/auth/me` | 获取当前认证用户的个人资料简要信息 |

---

## 2. 社区模块 (Community)
文件路径: `src/api/community.js`

| 接口名称 | 方法 | 请求路径 | 功能描述 |
| :--- | :--- | :--- | :--- |
| `listCommunityTopics` | GET | `/community/topics` / `/public/community/topics` | 获取社区话题列表 (支持公共访问) |
| `listMyCommunityTopics` | GET | `/community/me/topics` | 获取当前用户发布的话题列表 |
| `listMyCommunityComments` | GET | `/community/me/comments` | 获取当前用户发布的评论列表 |
| `getCommunityTopicDetail` | GET | `/community/topics/{id}` / `/public/community/topics/{id}` | 获取特定话题的详细信息 |
| `listCommunityComments` | GET | `/community/topics/{topicID}/comments` | 获取特定话题下的评论列表 |
| `createCommunityTopic` | POST | `/community/topics` | 发布新话题 |
| `createCommunityComment` | POST | `/community/topics/{topicID}/comments` | 在话题下发布新评论 |
| `createCommunityReaction` | POST | `/community/reactions` | 创建互动反馈 (如点赞、收藏) |
| `deleteCommunityReaction` | DELETE | `/community/reactions` | 删除互动反馈 |
| `createCommunityReport` | POST | `/community/reports` | 提交话题或评论的举报 |

---

## 3. 预测模块 (Forecast)
文件路径: `src/api/forecast.js`

| 接口名称 | 方法 | 请求路径 | 功能描述 |
| :--- | :--- | :--- | :--- |
| `createForecastRun` | POST | `/forecast/runs` | 发起一个新的预测任务 |
| `listForecastRuns` | GET | `/forecast/runs` | 获取预测任务历史列表 |
| `getForecastRunDetail` | GET | `/forecast/runs/{id}` | 获取特定预测任务的详情及结果 |

---

## 4. 市场模块 (Market)
文件路径: `src/api/market.js`

| 接口名称 | 方法 | 请求路径 | 功能描述 |
| :--- | :--- | :--- | :--- |
| `listStockRecommendations` | GET | `/stocks/recommendations` | 获取股票推荐列表 |
| `getStockRecommendationDetail` | GET | `/stocks/recommendations/{id}` | 获取股票推荐详情 |
| `getStockRecommendationPerformance` | GET | `/stocks/recommendations/{id}/performance` | 获取推荐股票的回测/表现数据 |
| `getStockRecommendationInsight` | GET | `/stocks/recommendations/{id}/insight` | 获取推荐股票的深入分析/见解 |
| `getStockRecommendationVersionHistory` | GET | `/stocks/recommendations/{id}/version-history` | 获取推荐版本的历史变更 |
| `listFuturesArbitrage` | GET | `/futures/arbitrage` | 获取期货套利机会列表 |
| `getFuturesGuidance` | GET | `/futures/guidance/{contract}` | 获取特定期货合约的操作指导 |
| `listFuturesStrategies` | GET | `/futures/strategies` | 获取期货策略列表 |
| `getFuturesStrategyInsight` | GET | `/futures/strategies/{id}/insight` | 获取期货策略的深入分析 |
| `getFuturesStrategyVersionHistory` | GET | `/futures/strategies/{id}/version-history` | 获取期货策略的版本历史 |
| `listMarketEvents` | GET | `/market/events` | 获取市场重大事件列表 |
| `getMarketEventDetail` | GET | `/market/events/{id}` | 获取市场事件的详细分析 |

---

## 5. 会员与订单 (Membership)
文件路径: `src/api/membership.js`

| 接口名称 | 方法 | 请求路径 | 功能描述 |
| :--- | :--- | :--- | :--- |
| `listMembershipProducts` | GET | `/membership/products` | 获取可用的会员产品/方案列表 |
| `listMembershipOrders` | GET | `/membership/orders` | 获取会员购买订单历史 |
| `getMembershipQuota` | GET | `/membership/quota` | 获取当前用户的会员额度/剩余使用次数 |
| `createMembershipOrder` | POST | `/membership/orders` | 创建会员购买订单 |
| `triggerPaymentCallback` | POST | `/payment/callbacks/{channel}` | 触发支付渠道回调 (通常用于测试或特定支付流) |

---

## 6. 新闻模块 (News)
文件路径: `src/api/news.js`

| 接口名称 | 方法 | 请求路径 | 功能描述 |
| :--- | :--- | :--- | :--- |
| `listNewsCategories` | GET | `/public/news/categories` | 获取新闻分类列表 |
| `listNewsArticles` | GET | `/public/news/articles` | 获取新闻文章列表 |
| `getNewsArticleDetail` | GET | `/news/articles/{id}` / `/public/news/articles/{id}` | 获取文章正文详情 |
| `listNewsAttachments` | GET | `/news/articles/{id}/attachments` | 获取文章关联的附件列表 |
| `getAttachmentSignedURL` | GET | `/news/attachments/{id}/signed-url` | 获取附件的限时签名下载地址 |

---

## 7. 搜索模块 (Search)
文件路径: `src/api/search.js`

| 接口名称 | 方法 | 请求路径 | 功能描述 |
| :--- | :--- | :--- | :--- |
| `searchGlobal` | GET | `/search/global` | 全局搜索 (需要登录) |
| `searchGlobalPublic` | GET | `/public/search/global` | 全局搜索 (公共访问) |

---

## 8. 个人中心与消息 (User Center)
文件路径: `src/api/userCenter.js`

| 接口名称 | 方法 | 请求路径 | 功能描述 |
| :--- | :--- | :--- | :--- |
| `getUserProfile` | GET | `/user/profile` | 获取用户详细个人信息 |
| `listRechargeRecords` | GET | `/user/recharge-records` | 获取充值记录列表 |
| `listBrowseHistory` | GET | `/user/browse-history` | 获取浏览历史记录 |
| `listMessages` | GET | `/messages` | 获取站内信/通知消息列表 |
| `readMessage` | PUT | `/messages/{id}/read` | 将消息标记为已读 |
| `listShareLinks` | GET | `/user/share-links` | 获取用户生成的邀请/分享链接列表 |
| `createShareLink` | POST | `/user/share-links` | 创建新的分享链接 |
| `listInviteRecords` | GET | `/user/share/invites` | 获取受邀用户列表 |
| `getInviteSummary` | GET | `/user/share/invite-summary` | 获取邀请奖励汇总信息 |
| `listSubscriptions` | GET | `/subscriptions` | 获取订阅通知配置列表 |
| `createSubscription` | POST | `/subscriptions` | 创建新的订阅配置 |
| `updateSubscription` | PUT | `/subscriptions/{id}` | 更新订阅配置 |

---

> [!NOTE]
> 1. 大部分接口通过 `src/lib/http` 进行调用，其中会自动处理 BaseURL 和认证 Header。
> 2. 以 `/public` 开头的接口通常支持免登录访问。
> 3. `resolveReadPath` 等辅助函数会根据用户当前的登录状态 (是否存在 Access Token) 自动切换公共路径与私有路径。
