# API 详细规范（v0.1）

本规范在 `数据字典与API清单.md` 的基础上补充字段级请求/响应、分页、错误码与鉴权要求。后端技术栈为 Go + Gin，数据库为 MySQL + Redis。

**1. 通用约定**
- Base Path：`/api/v1`
- 鉴权：`Authorization: Bearer <token>`
- 分页参数：`page`、`page_size`
- 返回统一结构：
```json
{
  "code": 0,
  "message": "ok",
  "data": {},
  "request_id": "req_123"
}
```
- 通用错误码：
1. `0` 成功
2. `40001` 参数错误
3. `40101` 未登录
4. `40301` 权限不足
5. `40302` 配额超限
6. `40401` 资源不存在
7. `40901` 重复回调或重复处理
8. `50001` 服务异常

**2. 认证与用户**

**2.1 注册**
```
POST /auth/register
```
**Request**
```json
{
  "phone": "13800000000",
  "password": "******"
}
```
**Response**
```json
{
  "code": 0,
  "message": "ok",
  "data": {"user_id": "u_001"}
}
```

**2.2 登录**
```
POST /auth/login
```
**Request**
```json
{
  "phone": "13800000000",
  "password": "******"
}
```
**Response**
```json
{
  "code": 0,
  "message": "ok",
  "data": {"token": "jwt_xxx", "expires_in": 7200}
}
```

**2.3 获取用户信息**
```
GET /user/profile
```
**Response**
```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "id": "u_001",
    "phone": "13800000000",
    "kyc_status": "APPROVED",
    "member_level": "VIP1"
  }
}
```

**2.4 实名提交**
```
POST /user/kyc/submit
```
**Request**
```json
{
  "real_name": "张三",
  "id_number": "110101199001010010"
}
```
**Response**
```json
{
  "code": 0,
  "message": "ok",
  "data": {"kyc_status": "PENDING"}
}
```

**3. 股票推荐**

**3.1 推荐列表**
```
GET /stocks/recommendations?page=1&page_size=20&risk=MEDIUM&sort=score_desc
```
**Response**
```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "items": [
      {
        "id": "reco_001",
        "symbol": "600000",
        "name": "浦发银行",
        "score": 84,
        "risk_level": "MEDIUM",
        "reason_summary": "资金流入增强，估值修复预期",
        "valid_to": "2026-02-28T16:00:00+08:00"
      }
    ],
    "page": 1,
    "page_size": 20,
    "total": 231
  }
}
```

**3.2 推荐详情**
```
GET /stocks/recommendations/{id}
```
**Response（实名用户）**
```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "id": "reco_001",
    "symbol": "600000",
    "score": 84,
    "risk_level": "MEDIUM",
    "position_range": "10%-20%",
    "detail": {
      "tech_score": 82,
      "fund_score": 78,
      "sentiment_score": 74,
      "money_flow_score": 88
    },
    "take_profit": "建议分批止盈",
    "stop_loss": "跌破前低",
    "risk_note": "短期波动较大",
    "valid_to": "2026-02-28T16:00:00+08:00"
  }
}
```
**Response（未实名脱敏）**
```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "id": "reco_001",
    "symbol": "600000",
    "score": 84,
    "risk_level": "MEDIUM",
    "detail": null,
    "take_profit": null,
    "stop_loss": null
  }
}
```

**4. 期货策略**

**4.1 策略列表**
```
GET /futures/strategies?page=1&page_size=20&direction=LONG
```
**Response**
```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "items": [
      {
        "id": "fs_001",
        "contract": "RB2405",
        "direction": "LONG",
        "risk_level": "MEDIUM",
        "position_range": "5%-10%",
        "valid_to": "2026-02-28T16:00:00+08:00"
      }
    ],
    "page": 1,
    "page_size": 20,
    "total": 80
  }
}
```

**4.2 套利详情**
```
GET /futures/arbitrage/{id}
```
**Response**
```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "id": "arb_1001",
    "type": "CALENDAR",
    "contract_a": "RB2405",
    "contract_b": "RB2409",
    "spread": 110,
    "percentile": 0.86,
    "entry_point": 120,
    "exit_point": 60,
    "stop_point": 170,
    "trigger_rule": "价差>=120 且分位>0.85"
  }
}
```

**4.3 套利机会列表**
```
GET /futures/arbitrage/opportunities?page=1&page_size=20&type=CALENDAR
```

**4.4 操作指导**
```
GET /futures/guidance/{contract}
```

**4.5 提醒订阅**
```
POST /futures/alerts
```

**4.6 复盘列表**
```
GET /futures/reviews?page=1&page_size=20
```

**5. 市场动态**

**5.1 动态列表**
```
GET /market/events?page=1&page_size=20&type=PRICE
```
**Response**
```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "items": [
      {
        "id": "evt_001",
        "event_type": "PRICE",
        "symbol": "600000",
        "summary": "价格快速上涨",
        "trigger_rule": "5分钟涨幅>3%",
        "created_at": "2026-02-24T10:02:00+08:00"
      }
    ],
    "page": 1,
    "page_size": 20,
    "total": 300
  }
}
```

**6. 新闻资讯**

**6.1 分类列表**
```
GET /news/categories
```
**Response**
```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "items": [
      {
        "id": "cat_001",
        "name": "宏观",
        "slug": "macro",
        "visibility": "PUBLIC"
      }
    ]
  }
}
```

**6.2 新闻列表**
```
GET /news/articles?page=1&page_size=20&category_id=cat_001
```
**Response**
```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "items": [
      {
        "id": "news_001",
        "category_id": "cat_001",
        "title": "政策观察",
        "summary": "最新政策速览",
        "visibility": "PUBLIC",
        "published_at": "2026-02-24T11:00:00+08:00"
      }
    ],
    "page": 1,
    "page_size": 20,
    "total": 120
  }
}
```

**6.3 新闻详情**
```
GET /news/articles/{id}
```
**Response**
```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "id": "news_001",
    "category_id": "cat_001",
    "title": "政策观察",
    "summary": "最新政策速览",
    "content": "正文内容",
    "visibility": "VIP",
    "published_at": "2026-02-24T11:00:00+08:00"
  }
}
```
**权限说明**
- 非 VIP 访问 `visibility=VIP` 返回 `40301`
- VIP 可浏览全部新闻

**6.4 新闻附件列表**
```
GET /news/articles/{id}/attachments
```
**Response**
```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "items": [
      {
        "id": "att_001",
        "file_name": "report.pdf",
        "file_url": "https://example.com/report.pdf",
        "file_size": 102400
      }
    ]
  }
}
```
**附件安全说明**
- 附件 `file_url` 返回签名 URL，默认有效期 5 分钟
- 非 VIP 或未实名用户访问受限附件返回 `40301`
- 超过有效期返回 `40401`

**6.5 附件签名链接**
```
GET /news/attachments/{id}/signed-url
```
**Response**
```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "signed_url": "https://example.com/file.pdf?sign=xxx",
    "expired_at": "2026-02-24T12:05:00+08:00"
  }
}
```

**7. 订阅与消息**

**7.1 订阅列表**
```
GET /subscriptions
```

**7.2 订阅创建**
```
POST /subscriptions
```
**Request**
```json
{
  "type": "STOCK_RECO",
  "scope": "BANKING",
  "frequency": "DAILY"
}
```
**Response**
```json
{
  "code": 0,
  "message": "ok",
  "data": {"id": "sub_001"}
}
```

**7.3 订阅更新**
```
PUT /subscriptions/{id}
```

**7.4 消息列表**
```
GET /messages?page=1&page_size=20
```
**Response**
```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "items": [
      {
        "id": "msg_001",
        "title": "新的股票推荐",
        "type": "RECO",
        "read_status": "UNREAD",
        "created_at": "2026-02-24T09:30:00+08:00"
      }
    ],
    "page": 1,
    "page_size": 20,
    "total": 40
  }
}
```

**7.5 消息已读**
```
PUT /messages/{id}/read
```

**8. 会员**

**8.1 会员产品**
```
GET /membership/products
```
**Response**
```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "items": [
      {"id": "vip1", "name": "VIP1 月度", "price": 99}
    ]
  }
}
```

**8.2 会员配额**
```
GET /membership/quota
```
**Response**
```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "member_level": "VIP1",
    "period_key": "2026-02",
    "doc_read_limit": 100,
    "doc_read_used": 24,
    "news_subscribe_limit": 50,
    "news_subscribe_used": 12,
    "reset_cycle": "MONTHLY"
  }
}
```
**超限返回示例**
```json
{
  "code": 40302,
  "message": "quota exceeded",
  "data": {
    "quota_type": "DOC_READ|NEWS_SUBSCRIBE",
    "remaining": 0,
    "reset_at": "2026-03-01T00:00:00+08:00"
  }
}
```

**9. 客户增长与激励**

**9.1 浏览历史列表**
```
GET /user/browse-history?page=1&page_size=20&content_type=NEWS
```

**9.2 充值记录列表**
```
GET /user/recharge-records?page=1&page_size=20&status=PAID
```

**9.3 分享链接**
```
GET /user/share-links
POST /user/share-links
```

**9.4 邀请与奖励**
```
GET /user/share/invites?page=1&page_size=20
GET /user/share/rewards?page=1&page_size=20
```

**示例：分享奖励列表**
```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "items": [
      {
        "id": "rw_001",
        "reward_type": "CASH",
        "reward_value": 20,
        "trigger_event": "INVITEE_FIRST_RECHARGE",
        "status": "ISSUED",
        "issued_at": "2026-02-24T12:00:00+08:00"
      }
    ]
  }
}
```

**9.5 奖励账户**
```
GET /user/reward-wallet
GET /user/reward-wallet/txns?page=1&page_size=20
POST /user/reward-wallet/withdraw
```
**提现 Request**
```json
{
  "amount": 50
}
```

**10. 支付与对账**
- `POST /payment/callbacks/{channel}`
- `GET /admin/payment/reconciliation`
- `POST /admin/payment/reconciliation/{batch_id}/retry`

**回调处理规则**
- 必须验签后入账
- 同一 `idempotency_key` 重复回调返回 `40901`
- 回调处理结果写入回调日志

**11. Admin（鉴权：ADMIN）**
- `POST /admin/strategies`
- `PUT /admin/strategies/{id}`
- `POST /admin/recommendations`
- `PUT /admin/recommendations/{id}/publish`
- `PUT /admin/kyc/{id}/review`
- `GET /admin/data-sources`
- `POST /admin/data-sources`
- `PUT /admin/data-sources/{source_key}`
- `DELETE /admin/data-sources/{source_key}`
- `POST /admin/data-sources/{source_key}/health-check`
- `GET /admin/data-sources/{source_key}/health-logs`
- `POST /admin/data-sources/health-checks`
- `POST /admin/news/categories`
- `PUT /admin/news/categories/{id}`
- `POST /admin/news/articles`
- `PUT /admin/news/articles/{id}`
- `PUT /admin/news/articles/{id}/publish`
- `POST /admin/news/articles/{id}/attachments`
- `DELETE /admin/news/attachments/{id}`
- `GET /admin/growth/invite-records`
- `GET /admin/growth/reward-records`
- `PUT /admin/growth/reward-records/{id}/review`
- `GET /admin/membership/quota-configs`
- `POST /admin/membership/quota-configs`
- `PUT /admin/membership/quota-configs/{id}`
- `GET /admin/membership/user-quotas`
- `PUT /admin/membership/user-quotas/{user_id}/adjust`
- `GET /admin/risk/rules`
- `POST /admin/risk/rules`
- `PUT /admin/risk/rules/{id}`
- `GET /admin/risk/hits`
- `PUT /admin/risk/hits/{id}/review`
- `GET /admin/reward-wallet/withdraw-requests`
- `PUT /admin/reward-wallet/withdraw-requests/{id}/review`
