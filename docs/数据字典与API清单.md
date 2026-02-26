# 数据字典与 API 清单（v0.1）

本文覆盖核心数据对象字段定义与 API 列表，用于 PC Web + 服务端 + Admin 的首版落地。字段类型使用通用约定，具体数据库类型可在实现阶段细化。

**1. 约定**
- 时间统一使用 UTC 或 Asia/Shanghai，字段后缀 `_at` 表示时间
- 主键统一为 `id`（UUID 或雪花），外键以 `_id` 结尾
- 金额使用 `decimal(18,4)` 或字符串，避免浮点误差
- 枚举值在接口返回中同时给出 `code` 与 `label`

**2. 核心数据字典**

**2.1 用户与实名**
- `user`
1. `id` string 用户 ID
2. `phone` string 手机号
3. `email` string 邮箱
4. `password_hash` string 密码哈希
5. `status` enum 账号状态 `ACTIVE|SUSPENDED|DELETED`
6. `kyc_status` enum 实名状态 `UNVERIFIED|PENDING|APPROVED|REJECTED`
7. `member_level` enum 会员等级 `FREE|VIP1|VIP2`
8. `created_at` datetime
9. `updated_at` datetime

- `kyc_record`
1. `id` string
2. `user_id` string
3. `real_name` string
4. `id_number` string
5. `provider` string 实名服务商
6. `status` enum `PENDING|APPROVED|REJECTED`
7. `reason` string 拒绝原因
8. `submitted_at` datetime
9. `reviewed_at` datetime

**2.2 股票推荐**
- `stock_reco`
1. `id` string
2. `symbol` string 股票代码
3. `name` string 股票名称
4. `score` number 推荐评分
5. `risk_level` enum `LOW|MEDIUM|HIGH`
6. `position_range` string 建议仓位区间
7. `valid_from` datetime
8. `valid_to` datetime
9. `status` enum `ACTIVE|EXPIRED|WITHDRAWN`
10. `reason_summary` string 推荐理由摘要
11. `created_at` datetime

- `stock_reco_detail`
1. `reco_id` string
2. `tech_score` number 技术面评分
3. `fund_score` number 基本面评分
4. `sentiment_score` number 情绪评分
5. `money_flow_score` number 资金流评分
6. `take_profit` string 止盈建议
7. `stop_loss` string 止损建议
8. `risk_note` string 风险提示

**2.3 期货策略与套利**
- `futures_strategy`
1. `id` string
2. `contract` string 合约代码
3. `name` string 合约名称
4. `direction` enum `LONG|SHORT`
5. `risk_level` enum `LOW|MEDIUM|HIGH`
6. `position_range` string 建议仓位区间
7. `valid_from` datetime
8. `valid_to` datetime
9. `status` enum `ACTIVE|EXPIRED|WITHDRAWN`
10. `reason_summary` string

- `arbitrage_reco`
1. `id` string
2. `type` enum `CALENDAR|INTERCOMMODITY|CASH_FUTURES`
3. `contract_a` string
4. `contract_b` string
5. `spread` number 当前价差
6. `percentile` number 历史分位
7. `entry_point` number 入场点位
8. `exit_point` number 出场点位
9. `stop_point` number 止损点位
10. `trigger_rule` string 触发条件
11. `status` enum `ACTIVE|EXPIRED|WITHDRAWN`

- `futures_guidance`
1. `id` string
2. `contract` string 合约代码
3. `guidance_direction` enum `LONG_SPREAD|SHORT_SPREAD|NEUTRAL`
4. `position_level` enum `LIGHT|STANDARD|CAUTIOUS`
5. `entry_range` string 入场区间
6. `take_profit_range` string 止盈区间
7. `stop_loss_range` string 止损区间
8. `risk_level` enum `LOW|MEDIUM|HIGH`
9. `invalid_condition` string 失效条件
10. `valid_to` datetime

**2.4 市场动态**
- `market_event`
1. `id` string
2. `event_type` enum `PRICE|VOLUME|OPEN_INTEREST|NEWS|ANNOUNCEMENT`
3. `symbol` string 关联标的
4. `summary` string 事件摘要
5. `trigger_rule` string 触发条件
6. `source` string 数据来源
7. `created_at` datetime

**2.5 公开持仓/持股**
- `public_holding`
1. `id` string
2. `holder` string 机构名称
3. `symbol` string 标的
4. `ratio` number 持股比例
5. `disclosed_at` datetime 披露日期
6. `source` string

- `futures_position_public`
1. `id` string
2. `contract` string 合约
3. `long_position` number 多头持仓
4. `short_position` number 空头持仓
5. `disclosed_at` datetime
6. `source` string

**2.6 订阅与消息**
- `subscription`
1. `id` string
2. `user_id` string
3. `type` enum `STOCK_RECO|FUTURES_STRATEGY|ARBITRAGE|EVENT`
4. `scope` string 标的范围
5. `frequency` enum `INSTANT|DAILY|WEEKLY`
6. `status` enum `ACTIVE|PAUSED`

- `message`
1. `id` string
2. `user_id` string
3. `title` string
4. `content` string
5. `type` enum `RECO|EVENT|SYSTEM`
6. `read_status` enum `UNREAD|READ`
7. `created_at` datetime

**2.7 会员与订单**
- `membership_order`
1. `id` string
2. `user_id` string
3. `product_id` string
4. `amount` decimal
5. `status` enum `PENDING|PAID|CANCELLED|REFUNDED`
6. `paid_at` datetime

- `vip_quota_config`
1. `id` string
2. `member_level` enum `VIP1|VIP2`
3. `doc_read_limit` number 阅读文档数量上限
4. `news_subscribe_limit` number 订阅资讯数量上限
5. `reset_cycle` enum `DAILY|MONTHLY`
6. `status` enum `ACTIVE|INACTIVE`
7. `effective_at` datetime
8. `updated_at` datetime

- `user_quota_usage`
1. `id` string
2. `user_id` string
3. `member_level` enum `FREE|VIP1|VIP2`
4. `period_key` string 配额周期标识，如 `2026-02`
5. `doc_read_used` number 已用阅读文档数量
6. `news_subscribe_used` number 已用订阅资讯数量
7. `updated_at` datetime

**2.8 新闻资讯**
- `news_category`
1. `id` string
2. `name` string 分类名称
3. `slug` string 分类标识
4. `sort` number 排序
5. `status` enum `ACTIVE|INACTIVE`
6. `visibility` enum `PUBLIC|VIP`
7. `created_at` datetime

- `news_article`
1. `id` string
2. `category_id` string
3. `title` string
4. `summary` string
5. `content` string
6. `tags` string[] 标签
7. `visibility` enum `PUBLIC|VIP`
8. `status` enum `DRAFT|PENDING|PUBLISHED|REJECTED|OFFLINE`
9. `published_at` datetime
10. `author_id` string
11. `created_at` datetime
12. `updated_at` datetime

- `news_attachment`
1. `id` string
2. `article_id` string
3. `file_name` string
4. `file_url` string
5. `file_size` number
6. `mime_type` string
7. `created_at` datetime
8. `signed_url_expired_at` datetime 签名链接过期时间

**2.9 客户增长与激励**
- `browse_history`
1. `id` string
2. `user_id` string
3. `content_type` enum `STOCK|FUTURES|NEWS`
4. `content_id` string
5. `source_page` string
6. `viewed_at` datetime

- `recharge_record`
1. `id` string
2. `user_id` string
3. `order_no` string
4. `amount` decimal
5. `pay_channel` enum `ALIPAY|WECHAT|CARD`
6. `status` enum `PENDING|PAID|FAILED|CANCELLED`
7. `paid_at` datetime
8. `remark` string

- `invite_link`
1. `id` string
2. `user_id` string
3. `invite_code` string
4. `url` string
5. `channel` string
6. `status` enum `ACTIVE|EXPIRED|DISABLED`
7. `expired_at` datetime
8. `created_at` datetime

- `invite_record`
1. `id` string
2. `inviter_user_id` string
3. `invitee_user_id` string
4. `invite_link_id` string
5. `register_at` datetime
6. `kyc_at` datetime
7. `first_pay_at` datetime
8. `status` enum `REGISTERED|VERIFIED|FIRST_PAID|INVALID`

- `share_reward_record`
1. `id` string
2. `inviter_user_id` string
3. `invitee_user_id` string
4. `invite_record_id` string
5. `reward_type` enum `CASH|COUPON|VIP_DAYS`
6. `reward_value` decimal
7. `trigger_event` enum `INVITEE_FIRST_RECHARGE|INVITEE_FIRST_MEMBERSHIP_PAY`
8. `status` enum `PENDING|ISSUED|REJECTED|FROZEN`
9. `issued_at` datetime
10. `risk_flag` enum `NORMAL|SUSPECTED`

- `reward_wallet`
1. `id` string
2. `user_id` string
3. `cash_balance` decimal
4. `cash_frozen` decimal
5. `coupon_balance` decimal
6. `vip_days_balance` number
7. `updated_at` datetime

- `reward_wallet_txn`
1. `id` string
2. `wallet_id` string
3. `txn_type` enum `ISSUE|WITHDRAW|DEDUCT|ROLLBACK`
4. `amount` decimal
5. `status` enum `PENDING|SUCCESS|FAILED`
6. `ref_id` string 关联订单/提现单
7. `created_at` datetime

- `withdraw_request`
1. `id` string
2. `user_id` string
3. `amount` decimal
4. `status` enum `PENDING|APPROVED|REJECTED|PAID|FAILED`
5. `review_reason` string
6. `applied_at` datetime
7. `reviewed_at` datetime

- `payment_callback_log`
1. `id` string
2. `pay_channel` enum `ALIPAY|WECHAT|CARD`
3. `order_no` string
4. `channel_txn_no` string
5. `sign_verified` bool
6. `idempotency_key` string
7. `callback_status` enum `RECEIVED|PROCESSED|IGNORED|FAILED`
8. `created_at` datetime

- `reconciliation_record`
1. `id` string
2. `pay_channel` enum `ALIPAY|WECHAT|CARD`
3. `batch_date` date
4. `status` enum `DONE|PARTIAL_FAILED`
5. `diff_count` number
6. `created_at` datetime

- `risk_rule_config`
1. `id` string
2. `rule_code` string 规则编码
3. `rule_name` string
4. `threshold` number
5. `status` enum `ACTIVE|INACTIVE`
6. `effective_at` datetime
7. `updated_at` datetime

- `risk_hit_log`
1. `id` string
2. `rule_code` string
3. `user_id` string
4. `event_id` string
5. `risk_level` enum `LOW|MEDIUM|HIGH`
6. `status` enum `PENDING_REVIEW|CONFIRMED|RELEASED`
7. `created_at` datetime

**3. API 清单**

**3.1 认证与用户**
- `POST /api/v1/auth/login`
- `POST /api/v1/auth/logout`
- `POST /api/v1/auth/register`
- `GET /api/v1/user/profile`
- `PUT /api/v1/user/profile`
- `GET /api/v1/user/kyc/status`
- `POST /api/v1/user/kyc/submit`

**3.2 股票推荐**
- `GET /api/v1/stocks/recommendations`
- `GET /api/v1/stocks/recommendations/{id}`
- `GET /api/v1/stocks/recommendations/{id}/performance`

**3.3 期货策略**
- `GET /api/v1/futures/strategies`
- `GET /api/v1/futures/strategies/{id}`
- `GET /api/v1/futures/arbitrage`
- `GET /api/v1/futures/arbitrage/{id}`
- `GET /api/v1/futures/arbitrage/opportunities`
- `GET /api/v1/futures/guidance/{contract}`
- `POST /api/v1/futures/alerts`
- `GET /api/v1/futures/reviews`

**3.4 市场动态**
- `GET /api/v1/market/events`
- `GET /api/v1/market/events/{id}`

**3.5 公开持仓/持股**
- `GET /api/v1/public/holdings`
- `GET /api/v1/public/futures-positions`

**3.6 订阅与消息**
- `GET /api/v1/subscriptions`
- `POST /api/v1/subscriptions`
- `PUT /api/v1/subscriptions/{id}`
- `GET /api/v1/messages`
- `PUT /api/v1/messages/{id}/read`

**3.7 会员**
- `GET /api/v1/membership/products`
- `POST /api/v1/membership/orders`
- `GET /api/v1/membership/orders`
- `GET /api/v1/membership/quota`

**3.8 新闻资讯**
- `GET /api/v1/news/categories`
- `GET /api/v1/news/articles`
- `GET /api/v1/news/articles/{id}`
- `GET /api/v1/news/articles/{id}/attachments`
- `GET /api/v1/news/attachments/{id}/signed-url`

**3.9 客户增长与激励**
- `GET /api/v1/user/browse-history`
- `DELETE /api/v1/user/browse-history/{id}`
- `DELETE /api/v1/user/browse-history`
- `GET /api/v1/user/recharge-records`
- `GET /api/v1/user/share-links`
- `POST /api/v1/user/share-links`
- `GET /api/v1/user/share/rewards`
- `GET /api/v1/user/share/invites`
- `GET /api/v1/user/reward-wallet`
- `GET /api/v1/user/reward-wallet/txns`
- `POST /api/v1/user/reward-wallet/withdraw`

**3.10 支付与对账**
- `POST /api/v1/payment/callbacks/{channel}`
- `GET /api/v1/admin/payment/reconciliation`
- `POST /api/v1/admin/payment/reconciliation/{batch_id}/retry`

**3.11 Admin**
- `POST /api/v1/admin/strategies`
- `PUT /api/v1/admin/strategies/{id}`
- `POST /api/v1/admin/recommendations`
- `PUT /api/v1/admin/recommendations/{id}/publish`
- `GET /api/v1/admin/users`
- `PUT /api/v1/admin/kyc/{id}/review`
- `GET /api/v1/admin/data-sources`
- `POST /api/v1/admin/data-sources`
- `PUT /api/v1/admin/data-sources/{source_key}`
- `DELETE /api/v1/admin/data-sources/{source_key}`
- `POST /api/v1/admin/data-sources/{source_key}/health-check`
- `GET /api/v1/admin/data-sources/{source_key}/health-logs`
- `POST /api/v1/admin/data-sources/health-checks`
- `POST /api/v1/admin/news/categories`
- `PUT /api/v1/admin/news/categories/{id}`
- `POST /api/v1/admin/news/articles`
- `PUT /api/v1/admin/news/articles/{id}`
- `PUT /api/v1/admin/news/articles/{id}/publish`
- `POST /api/v1/admin/news/articles/{id}/attachments`
- `DELETE /api/v1/admin/news/attachments/{id}`
- `GET /api/v1/admin/growth/invite-records`
- `GET /api/v1/admin/growth/reward-records`
- `PUT /api/v1/admin/growth/reward-records/{id}/review`
- `GET /api/v1/admin/membership/quota-configs`
- `POST /api/v1/admin/membership/quota-configs`
- `PUT /api/v1/admin/membership/quota-configs/{id}`
- `GET /api/v1/admin/membership/user-quotas`
- `PUT /api/v1/admin/membership/user-quotas/{user_id}/adjust`
- `GET /api/v1/admin/risk/rules`
- `POST /api/v1/admin/risk/rules`
- `PUT /api/v1/admin/risk/rules/{id}`
- `GET /api/v1/admin/risk/hits`
- `PUT /api/v1/admin/risk/hits/{id}/review`
- `GET /api/v1/admin/reward-wallet/withdraw-requests`
- `PUT /api/v1/admin/reward-wallet/withdraw-requests/{id}/review`

**4. 接口示例**

**4.1 股票推荐列表**
```json
GET /api/v1/stocks/recommendations
```
```json
{
  "items": [
    {
      "id": "reco_001",
      "symbol": "600000",
      "name": "浦发银行",
      "score": 84,
      "risk_level": {"code": "MEDIUM", "label": "中"},
      "reason_summary": "资金流入增强，估值修复预期",
      "valid_to": "2026-02-28T16:00:00+08:00"
    }
  ],
  "page": 1,
  "page_size": 20,
  "total": 231
}
```

**4.2 期货套利详情**
```json
GET /api/v1/futures/arbitrage/{id}
```
```json
{
  "id": "arb_1001",
  "type": {"code": "CALENDAR", "label": "跨期套利"},
  "contract_a": "RB2405",
  "contract_b": "RB2409",
  "spread": 110,
  "percentile": 0.86,
  "entry_point": 120,
  "exit_point": 60,
  "stop_point": 170,
  "trigger_rule": "价差>=120 且分位>0.85"
}
```

**5. 权限校验规则**
- 未实名用户：`/stocks/recommendations/{id}` 返回脱敏字段
- 非会员用户：高频推荐字段置空或返回 `403`
- 非会员用户：仅可访问 `visibility=PUBLIC` 的新闻资讯
- 分享奖励：同一被邀请用户首充奖励仅触发一次
- 会员用户：阅读文档数量与订阅资讯数量受会员配额控制
- 文档/订阅配额超限返回统一错误码 `40302`
- 管理接口仅限 `ADMIN` 角色访问
