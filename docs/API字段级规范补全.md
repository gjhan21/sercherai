# API 字段级规范补全（v0.1）

本文件补全全量接口的字段级请求/响应说明，作为实现与联调依据。

**1. 鉴权**
- `Authorization: Bearer <token>`
- Admin 端需 `role=ADMIN`

**1.1 错误码**
1. `40001` 参数错误
2. `40101` 未登录
3. `40301` 权限不足
4. `40302` 配额超限
5. `40401` 资源不存在
6. `40901` 重复回调或重复处理
7. `50001` 服务异常

**2. 用户与认证**

**2.1 注册**
```
POST /api/v1/auth/register
```
**Request**
```json
{
  "phone": "string",
  "password": "string"
}
```
**Response**
```json
{
  "code": 0,
  "message": "ok",
  "data": {"user_id": "string"}
}
```

**2.2 登录**
```
POST /api/v1/auth/login
```
**Request**
```json
{
  "phone": "string",
  "password": "string"
}
```
**Response**
```json
{
  "code": 0,
  "message": "ok",
  "data": {"token": "string", "expires_in": 7200}
}
```

**2.3 退出**
```
POST /api/v1/auth/logout
```
**Response**
```json
{"code": 0, "message": "ok", "data": {}}
```

**2.4 获取用户信息**
```
GET /api/v1/user/profile
```
**Response**
```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "id": "string",
    "phone": "string",
    "email": "string",
    "kyc_status": "UNVERIFIED|PENDING|APPROVED|REJECTED",
    "member_level": "FREE|VIP1|VIP2"
  }
}
```

**2.5 更新用户信息**
```
PUT /api/v1/user/profile
```
**Request**
```json
{
  "email": "string"
}
```
**Response**
```json
{"code": 0, "message": "ok", "data": {}}
```

**2.6 实名状态**
```
GET /api/v1/user/kyc/status
```
**Response**
```json
{
  "code": 0,
  "message": "ok",
  "data": {"kyc_status": "UNVERIFIED|PENDING|APPROVED|REJECTED"}
}
```

**2.7 实名提交**
```
POST /api/v1/user/kyc/submit
```
**Request**
```json
{
  "real_name": "string",
  "id_number": "string"
}
```
**Response**
```json
{"code": 0, "message": "ok", "data": {"kyc_status": "PENDING"}}
```

**3. 股票推荐**

**3.1 推荐列表**
```
GET /api/v1/stocks/recommendations?page=1&page_size=20&risk=MEDIUM&sort=score_desc
```
**Response**
```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "items": [
      {
        "id": "string",
        "symbol": "string",
        "name": "string",
        "score": 0,
        "risk_level": "LOW|MEDIUM|HIGH",
        "reason_summary": "string",
        "valid_to": "datetime"
      }
    ],
    "page": 1,
    "page_size": 20,
    "total": 0
  }
}
```

**3.2 推荐详情**
```
GET /api/v1/stocks/recommendations/{id}
```
**Response（实名）**
```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "id": "string",
    "symbol": "string",
    "name": "string",
    "score": 0,
    "risk_level": "LOW|MEDIUM|HIGH",
    "position_range": "string",
    "detail": {
      "tech_score": 0,
      "fund_score": 0,
      "sentiment_score": 0,
      "money_flow_score": 0
    },
    "take_profit": "string",
    "stop_loss": "string",
    "risk_note": "string",
    "valid_to": "datetime"
  }
}
```
**Response（未实名脱敏）**
```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "id": "string",
    "symbol": "string",
    "name": "string",
    "score": 0,
    "risk_level": "LOW|MEDIUM|HIGH",
    "detail": null,
    "take_profit": null,
    "stop_loss": null
  }
}
```

**3.3 推荐表现**
```
GET /api/v1/stocks/recommendations/{id}/performance
```
**Response**
```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "points": [
      {"date": "2026-02-24", "return": 0.0}
    ]
  }
}
```

**4. 期货策略与套利**

**4.1 策略列表**
```
GET /api/v1/futures/strategies?page=1&page_size=20&direction=LONG
```
**Response**
```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "items": [
      {
        "id": "string",
        "contract": "string",
        "direction": "LONG|SHORT",
        "risk_level": "LOW|MEDIUM|HIGH",
        "position_range": "string",
        "valid_to": "datetime"
      }
    ],
    "page": 1,
    "page_size": 20,
    "total": 0
  }
}
```

**4.2 策略详情**
```
GET /api/v1/futures/strategies/{id}
```
**Response**
```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "id": "string",
    "contract": "string",
    "direction": "LONG|SHORT",
    "risk_level": "LOW|MEDIUM|HIGH",
    "position_range": "string",
    "reason_summary": "string",
    "valid_to": "datetime"
  }
}
```

**4.3 套利列表**
```
GET /api/v1/futures/arbitrage?page=1&page_size=20&type=CALENDAR
```
**Response**
```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "items": [
      {
        "id": "string",
        "type": "CALENDAR|INTERCOMMODITY|CASH_FUTURES",
        "contract_a": "string",
        "contract_b": "string",
        "spread": 0,
        "percentile": 0.0,
        "entry_point": 0,
        "exit_point": 0,
        "stop_point": 0,
        "valid_to": "datetime"
      }
    ],
    "page": 1,
    "page_size": 20,
    "total": 0
  }
}
```

**4.4 套利详情**
```
GET /api/v1/futures/arbitrage/{id}
```
**Response**
```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "id": "string",
    "type": "CALENDAR|INTERCOMMODITY|CASH_FUTURES",
    "contract_a": "string",
    "contract_b": "string",
    "spread": 0,
    "percentile": 0.0,
    "entry_point": 0,
    "exit_point": 0,
    "stop_point": 0,
    "trigger_rule": "string"
  }
}
```

**4.5 套利机会列表**
```
GET /api/v1/futures/arbitrage/opportunities?page=1&page_size=20&type=CALENDAR
```
**Response**
```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "items": [
      {
        "id": "string",
        "type": "CALENDAR|INTERCOMMODITY|CASH_FUTURES",
        "contract_a": "string",
        "contract_b": "string",
        "spread": 0,
        "percentile": 0.0,
        "z_score": 0.0,
        "half_life": 0.0,
        "risk_level": "LOW|MEDIUM|HIGH",
        "status": "WATCH|ENTRY|HOLD|EXIT"
      }
    ],
    "page": 1,
    "page_size": 20,
    "total": 0
  }
}
```

**4.6 操作指导详情**
```
GET /api/v1/futures/guidance/{contract}
```
**Response**
```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "contract": "string",
    "guidance_direction": "LONG_SPREAD|SHORT_SPREAD|NEUTRAL",
    "position_level": "LIGHT|STANDARD|CAUTIOUS",
    "entry_range": "string",
    "take_profit_range": "string",
    "stop_loss_range": "string",
    "risk_level": "LOW|MEDIUM|HIGH",
    "invalid_condition": "string",
    "valid_to": "datetime"
  }
}
```

**4.7 套利提醒订阅**
```
POST /api/v1/futures/alerts
```
**Request**
```json
{
  "contract": "string",
  "alert_type": "ENTRY|EXIT|STOP_LOSS",
  "threshold": 0.0
}
```
**Response**
```json
{"code": 0, "message": "ok", "data": {"id": "string"}}
```

**4.8 套利复盘列表**
```
GET /api/v1/futures/reviews?page=1&page_size=20
```
**Response**
```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "items": [
      {
        "id": "string",
        "strategy_id": "string",
        "hit_rate": 0.0,
        "pnl": 0.0,
        "max_drawdown": 0.0,
        "review_date": "datetime"
      }
    ]
  }
}
```

**5. 市场动态**

**5.1 动态列表**
```
GET /api/v1/market/events?page=1&page_size=20&type=PRICE
```
**Response**
```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "items": [
      {
        "id": "string",
        "event_type": "PRICE|VOLUME|OPEN_INTEREST|NEWS|ANNOUNCEMENT",
        "symbol": "string",
        "summary": "string",
        "trigger_rule": "string",
        "created_at": "datetime"
      }
    ],
    "page": 1,
    "page_size": 20,
    "total": 0
  }
}
```

**5.2 动态详情**
```
GET /api/v1/market/events/{id}
```
**Response**
```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "id": "string",
    "event_type": "PRICE|VOLUME|OPEN_INTEREST|NEWS|ANNOUNCEMENT",
    "symbol": "string",
    "summary": "string",
    "trigger_rule": "string",
    "source": "string",
    "created_at": "datetime"
  }
}
```

**6. 公开持仓/持股**

**6.1 机构持股**
```
GET /api/v1/public/holdings?page=1&page_size=20&symbol=600000
```
**Response**
```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "items": [
      {
        "id": "string",
        "holder": "string",
        "symbol": "string",
        "ratio": 0.0,
        "disclosed_at": "datetime",
        "source": "string"
      }
    ],
    "page": 1,
    "page_size": 20,
    "total": 0
  }
}
```

**6.2 期货持仓结构**
```
GET /api/v1/public/futures-positions?page=1&page_size=20&contract=RB2405
```
**Response**
```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "items": [
      {
        "id": "string",
        "contract": "string",
        "long_position": 0,
        "short_position": 0,
        "disclosed_at": "datetime",
        "source": "string"
      }
    ],
    "page": 1,
    "page_size": 20,
    "total": 0
  }
}
```

**7. 订阅与消息**

**7.1 订阅列表**
```
GET /api/v1/subscriptions
```
**Response**
```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "items": [
      {
        "id": "string",
        "type": "STOCK_RECO|FUTURES_STRATEGY|ARBITRAGE|EVENT",
        "scope": "string",
        "frequency": "INSTANT|DAILY|WEEKLY",
        "status": "ACTIVE|PAUSED"
      }
    ]
  }
}
```

**7.2 订阅创建**
```
POST /api/v1/subscriptions
```
**Request**
```json
{
  "type": "STOCK_RECO|FUTURES_STRATEGY|ARBITRAGE|EVENT",
  "scope": "string",
  "frequency": "INSTANT|DAILY|WEEKLY"
}
```
**Response**
```json
{"code": 0, "message": "ok", "data": {"id": "string"}}
```

**7.3 订阅更新**
```
PUT /api/v1/subscriptions/{id}
```
**Request**
```json
{
  "frequency": "INSTANT|DAILY|WEEKLY",
  "status": "ACTIVE|PAUSED"
}
```
**Response**
```json
{"code": 0, "message": "ok", "data": {}}
```

**7.4 消息列表**
```
GET /api/v1/messages?page=1&page_size=20
```
**Response**
```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "items": [
      {
        "id": "string",
        "title": "string",
        "type": "RECO|EVENT|SYSTEM",
        "read_status": "UNREAD|READ",
        "created_at": "datetime"
      }
    ],
    "page": 1,
    "page_size": 20,
    "total": 0
  }
}
```

**7.5 消息已读**
```
PUT /api/v1/messages/{id}/read
```
**Response**
```json
{"code": 0, "message": "ok", "data": {}}
```

**8. 会员**

**8.1 会员产品**
```
GET /api/v1/membership/products
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

**8.2 会员下单**
```
POST /api/v1/membership/orders
```
**Request**
```json
{
  "product_id": "string",
  "pay_channel": "string"
}
```
**Response**
```json
{"code": 0, "message": "ok", "data": {"order_id": "string"}}
```

**8.3 订单列表**
```
GET /api/v1/membership/orders?page=1&page_size=20
```
**Response**
```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "items": [
      {
        "id": "string",
        "product_id": "string",
        "amount": 0,
        "status": "PENDING|PAID|CANCELLED|REFUNDED",
        "paid_at": "datetime"
      }
    ]
  }
}
```

**8.4 会员配额查询**
```
GET /api/v1/membership/quota
```
**Response**
```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "member_level": "VIP1|VIP2",
    "period_key": "2026-02",
    "doc_read_limit": 100,
    "doc_read_used": 24,
    "doc_read_remaining": 76,
    "news_subscribe_limit": 50,
    "news_subscribe_used": 12,
    "news_subscribe_remaining": 38,
    "reset_cycle": "MONTHLY"
  }
}
```
**配额超限响应**
```json
{
  "code": 40302,
  "message": "quota exceeded",
  "data": {
    "quota_type": "DOC_READ|NEWS_SUBSCRIBE",
    "remaining": 0,
    "reset_at": "datetime"
  }
}
```

**9. Admin（鉴权：ADMIN）**

**9.1 策略创建**
```
POST /api/v1/admin/strategies
```
**Request**
```json
{
  "contract": "string",
  "direction": "LONG|SHORT",
  "risk_level": "LOW|MEDIUM|HIGH",
  "position_range": "string",
  "valid_from": "datetime",
  "valid_to": "datetime",
  "reason_summary": "string"
}
```
**Response**
```json
{"code": 0, "message": "ok", "data": {"id": "string"}}
```

**9.2 策略更新**
```
PUT /api/v1/admin/strategies/{id}
```
**Request**
```json
{
  "risk_level": "LOW|MEDIUM|HIGH",
  "position_range": "string",
  "valid_to": "datetime",
  "reason_summary": "string"
}
```
**Response**
```json
{"code": 0, "message": "ok", "data": {}}
```

**9.3 推荐创建**
```
POST /api/v1/admin/recommendations
```
**Request**
```json
{
  "symbol": "string",
  "score": 0,
  "risk_level": "LOW|MEDIUM|HIGH",
  "position_range": "string",
  "valid_from": "datetime",
  "valid_to": "datetime",
  "reason_summary": "string"
}
```
**Response**
```json
{"code": 0, "message": "ok", "data": {"id": "string"}}
```

**9.4 推荐发布**
```
PUT /api/v1/admin/recommendations/{id}/publish
```
**Response**
```json
{"code": 0, "message": "ok", "data": {}}
```

**9.5 实名审核**
```
PUT /api/v1/admin/kyc/{id}/review
```
**Request**
```json
{
  "status": "APPROVED|REJECTED",
  "reason": "string"
}
```
**Response**
```json
{"code": 0, "message": "ok", "data": {}}
```

**10. 新闻资讯**

**10.1 分类列表**
```
GET /api/v1/news/categories
```
**Response**
```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "items": [
      {
        "id": "string",
        "name": "string",
        "slug": "string",
        "sort": 0,
        "visibility": "PUBLIC|VIP",
        "status": "ACTIVE|INACTIVE"
      }
    ]
  }
}
```

**10.2 新闻列表**
```
GET /api/v1/news/articles?page=1&page_size=20&category_id=string
```
**Response**
```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "items": [
      {
        "id": "string",
        "category_id": "string",
        "title": "string",
        "summary": "string",
        "tags": ["string"],
        "visibility": "PUBLIC|VIP",
        "published_at": "datetime"
      }
    ],
    "page": 1,
    "page_size": 20,
    "total": 0
  }
}
```

**10.3 新闻详情**
```
GET /api/v1/news/articles/{id}
```
**Response**
```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "id": "string",
    "category_id": "string",
    "title": "string",
    "summary": "string",
    "content": "string",
    "tags": ["string"],
    "visibility": "PUBLIC|VIP",
    "published_at": "datetime"
  }
}
```
**权限规则**
- 非 VIP 用户访问 `visibility=VIP` 返回 `40301`
- VIP 用户可浏览全部新闻

**10.4 新闻附件列表**
```
GET /api/v1/news/articles/{id}/attachments
```
**Response**
```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "items": [
      {
        "id": "string",
        "file_name": "string",
        "file_url": "string",
        "file_size": 0,
        "mime_type": "string",
        "created_at": "datetime"
      }
    ]
  }
}
```

**10.5 附件签名链接**
```
GET /api/v1/news/attachments/{id}/signed-url
```
**Response**
```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "signed_url": "string",
    "expired_at": "datetime"
  }
}
```

**11. Admin 新闻管理（鉴权：ADMIN）**

**11.1 创建分类**
```
POST /api/v1/admin/news/categories
```
**Request**
```json
{
  "name": "string",
  "slug": "string",
  "sort": 0,
  "visibility": "PUBLIC|VIP",
  "status": "ACTIVE|INACTIVE"
}
```
**Response**
```json
{"code": 0, "message": "ok", "data": {"id": "string"}}
```

**11.2 更新分类**
```
PUT /api/v1/admin/news/categories/{id}
```
**Request**
```json
{
  "name": "string",
  "sort": 0,
  "visibility": "PUBLIC|VIP",
  "status": "ACTIVE|INACTIVE"
}
```
**Response**
```json
{"code": 0, "message": "ok", "data": {}}
```

**11.3 创建新闻**
```
POST /api/v1/admin/news/articles
```
**Request**
```json
{
  "category_id": "string",
  "title": "string",
  "summary": "string",
  "content": "string",
  "tags": ["string"],
  "visibility": "PUBLIC|VIP"
}
```
**Response**
```json
{"code": 0, "message": "ok", "data": {"id": "string"}}
```

**11.4 更新新闻**
```
PUT /api/v1/admin/news/articles/{id}
```
**Request**
```json
{
  "title": "string",
  "summary": "string",
  "content": "string",
  "tags": ["string"],
  "visibility": "PUBLIC|VIP",
  "status": "DRAFT|PENDING|OFFLINE"
}
```
**Response**
```json
{"code": 0, "message": "ok", "data": {}}
```

**11.5 发布新闻**
```
PUT /api/v1/admin/news/articles/{id}/publish
```
**Request**
```json
{
  "status": "PUBLISHED"
}
```
**Response**
```json
{"code": 0, "message": "ok", "data": {}}
```

**11.6 上传附件**
```
POST /api/v1/admin/news/articles/{id}/attachments
```
**Request**
```json
{
  "file_name": "string",
  "file_url": "string",
  "file_size": 0,
  "mime_type": "string"
}
```
**Response**
```json
{"code": 0, "message": "ok", "data": {"id": "string"}}
```

**11.7 删除附件**
```
DELETE /api/v1/admin/news/attachments/{id}
```
**Response**
```json
{"code": 0, "message": "ok", "data": {}}
```

**12. 客户增长与激励**

**12.1 浏览历史列表**
```
GET /api/v1/user/browse-history?page=1&page_size=20&content_type=NEWS
```
**Response**
```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "items": [
      {
        "id": "string",
        "content_type": "STOCK|FUTURES|NEWS",
        "content_id": "string",
        "title": "string",
        "source_page": "string",
        "viewed_at": "datetime"
      }
    ],
    "page": 1,
    "page_size": 20,
    "total": 0
  }
}
```

**12.2 删除单条浏览历史**
```
DELETE /api/v1/user/browse-history/{id}
```
**Response**
```json
{"code": 0, "message": "ok", "data": {}}
```

**12.3 清空浏览历史**
```
DELETE /api/v1/user/browse-history
```
**Response**
```json
{"code": 0, "message": "ok", "data": {}}
```

**12.4 充值记录列表**
```
GET /api/v1/user/recharge-records?page=1&page_size=20&status=PAID
```
**Response**
```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "items": [
      {
        "id": "string",
        "order_no": "string",
        "amount": 0,
        "pay_channel": "ALIPAY|WECHAT|CARD",
        "status": "PENDING|PAID|FAILED|CANCELLED",
        "paid_at": "datetime",
        "remark": "string"
      }
    ],
    "page": 1,
    "page_size": 20,
    "total": 0
  }
}
```

**12.5 分享链接列表**
```
GET /api/v1/user/share-links
```
**Response**
```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "items": [
      {
        "id": "string",
        "invite_code": "string",
        "url": "string",
        "channel": "string",
        "status": "ACTIVE|EXPIRED|DISABLED",
        "expired_at": "datetime"
      }
    ]
  }
}
```

**12.6 创建分享链接**
```
POST /api/v1/user/share-links
```
**Request**
```json
{
  "channel": "string",
  "expired_at": "datetime"
}
```
**Response**
```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "id": "string",
    "invite_code": "string",
    "url": "string"
  }
}
```

**12.7 邀请记录**
```
GET /api/v1/user/share/invites?page=1&page_size=20
```
**Response**
```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "items": [
      {
        "id": "string",
        "invitee_user_id": "string",
        "status": "REGISTERED|VERIFIED|FIRST_PAID|INVALID",
        "register_at": "datetime",
        "first_pay_at": "datetime"
      }
    ]
  }
}
```

**12.8 分享奖励记录**
```
GET /api/v1/user/share/rewards?page=1&page_size=20
```
**Response**
```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "items": [
      {
        "id": "string",
        "reward_type": "CASH|COUPON|VIP_DAYS",
        "reward_value": 0,
        "trigger_event": "INVITEE_FIRST_RECHARGE|INVITEE_FIRST_MEMBERSHIP_PAY",
        "status": "PENDING|ISSUED|REJECTED|FROZEN",
        "issued_at": "datetime"
      }
    ]
  }
}
```

**13. Admin 增长管理（鉴权：ADMIN）**

**13.1 邀请记录列表**
```
GET /api/v1/admin/growth/invite-records?page=1&page_size=20&status=FIRST_PAID
```
**Response**
```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "items": [
      {
        "id": "string",
        "inviter_user_id": "string",
        "invitee_user_id": "string",
        "status": "REGISTERED|VERIFIED|FIRST_PAID|INVALID",
        "risk_flag": "NORMAL|SUSPECTED"
      }
    ]
  }
}
```

**13.2 奖励记录列表**
```
GET /api/v1/admin/growth/reward-records?page=1&page_size=20&status=PENDING
```
**Response**
```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "items": [
      {
        "id": "string",
        "inviter_user_id": "string",
        "invitee_user_id": "string",
        "reward_type": "CASH|COUPON|VIP_DAYS",
        "reward_value": 0,
        "status": "PENDING|ISSUED|REJECTED|FROZEN"
      }
    ]
  }
}
```

**13.3 奖励审核**
```
PUT /api/v1/admin/growth/reward-records/{id}/review
```
**Request**
```json
{
  "status": "ISSUED|REJECTED|FROZEN",
  "reason": "string"
}
```
**Response**
```json
{"code": 0, "message": "ok", "data": {}}
```

**14. Admin 会员配额管理（鉴权：ADMIN）**

**14.1 配额配置列表**
```
GET /api/v1/admin/membership/quota-configs
```
**Response**
```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "items": [
      {
        "id": "string",
        "member_level": "VIP1|VIP2",
        "doc_read_limit": 100,
        "news_subscribe_limit": 50,
        "reset_cycle": "DAILY|MONTHLY",
        "status": "ACTIVE|INACTIVE",
        "effective_at": "datetime"
      }
    ]
  }
}
```

**14.2 新增配额配置**
```
POST /api/v1/admin/membership/quota-configs
```
**Request**
```json
{
  "member_level": "VIP1|VIP2",
  "doc_read_limit": 100,
  "news_subscribe_limit": 50,
  "reset_cycle": "DAILY|MONTHLY",
  "effective_at": "datetime"
}
```
**Response**
```json
{"code": 0, "message": "ok", "data": {"id": "string"}}
```

**14.3 更新配额配置**
```
PUT /api/v1/admin/membership/quota-configs/{id}
```
**Request**
```json
{
  "doc_read_limit": 120,
  "news_subscribe_limit": 60,
  "reset_cycle": "DAILY|MONTHLY",
  "status": "ACTIVE|INACTIVE",
  "effective_at": "datetime"
}
```
**Response**
```json
{"code": 0, "message": "ok", "data": {}}
```

**14.4 用户配额查询**
```
GET /api/v1/admin/membership/user-quotas?page=1&page_size=20&user_id=string
```
**Response**
```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "items": [
      {
        "user_id": "string",
        "member_level": "FREE|VIP1|VIP2",
        "period_key": "2026-02",
        "doc_read_limit": 100,
        "doc_read_used": 24,
        "news_subscribe_limit": 50,
        "news_subscribe_used": 12
      }
    ]
  }
}
```

**14.5 用户配额调整**
```
PUT /api/v1/admin/membership/user-quotas/{user_id}/adjust
```
**Request**
```json
{
  "period_key": "2026-02",
  "doc_read_delta": 10,
  "news_subscribe_delta": 5,
  "reason": "string"
}
```
**Response**
```json
{"code": 0, "message": "ok", "data": {}}
```

**15. 奖励账户与提现**

**15.1 奖励账户详情**
```
GET /api/v1/user/reward-wallet
```
**Response**
```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "cash_balance": 100.5,
    "cash_frozen": 20,
    "coupon_balance": 30,
    "vip_days_balance": 15
  }
}
```

**15.2 奖励流水**
```
GET /api/v1/user/reward-wallet/txns?page=1&page_size=20
```
**Response**
```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "items": [
      {
        "id": "string",
        "txn_type": "ISSUE|WITHDRAW|DEDUCT|ROLLBACK",
        "amount": 20,
        "status": "PENDING|SUCCESS|FAILED",
        "created_at": "datetime"
      }
    ]
  }
}
```

**15.3 提现申请**
```
POST /api/v1/user/reward-wallet/withdraw
```
**Request**
```json
{
  "amount": 50
}
```
**Response**
```json
{"code": 0, "message": "ok", "data": {"withdraw_id": "string"}}
```

**16. 支付回调与对账**

**16.1 支付回调**
```
POST /api/v1/payment/callbacks/{channel}
```
**Request**
```json
{
  "order_no": "string",
  "channel_txn_no": "string",
  "idempotency_key": "string",
  "sign": "string",
  "payload": {}
}
```
**Response**
```json
{"code": 0, "message": "ok", "data": {}}
```
**规则**
- 验签失败返回 `40301`
- 幂等冲突返回 `40901`

**16.2 对账批次列表（Admin）**
```
GET /api/v1/admin/payment/reconciliation?page=1&page_size=20
```
**Response**
```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "items": [
      {
        "id": "string",
        "pay_channel": "ALIPAY|WECHAT|CARD",
        "batch_date": "2026-02-24",
        "status": "DONE|PARTIAL_FAILED",
        "diff_count": 0
      }
    ]
  }
}
```

**16.3 对账重试（Admin）**
```
POST /api/v1/admin/payment/reconciliation/{batch_id}/retry
```
**Response**
```json
{"code": 0, "message": "ok", "data": {}}
```

**17. Admin 风控配置与命中管理**

**17.1 风控规则列表**
```
GET /api/v1/admin/risk/rules
```
**Response**
```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "items": [
      {
        "id": "string",
        "rule_code": "DEVICE_DUP",
        "rule_name": "同设备重复邀请",
        "threshold": 3,
        "status": "ACTIVE|INACTIVE"
      }
    ]
  }
}
```

**17.2 风控规则新增**
```
POST /api/v1/admin/risk/rules
```
**Request**
```json
{
  "rule_code": "IP_HIGH_FREQ",
  "rule_name": "同IP高频邀请",
  "threshold": 10,
  "status": "ACTIVE"
}
```
**Response**
```json
{"code": 0, "message": "ok", "data": {"id": "string"}}
```

**17.3 风控规则更新**
```
PUT /api/v1/admin/risk/rules/{id}
```
**Request**
```json
{
  "threshold": 8,
  "status": "ACTIVE|INACTIVE"
}
```
**Response**
```json
{"code": 0, "message": "ok", "data": {}}
```

**17.4 风险命中列表**
```
GET /api/v1/admin/risk/hits?page=1&page_size=20&status=PENDING_REVIEW
```
**Response**
```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "items": [
      {
        "id": "string",
        "rule_code": "DEVICE_DUP",
        "user_id": "string",
        "risk_level": "LOW|MEDIUM|HIGH",
        "status": "PENDING_REVIEW|CONFIRMED|RELEASED"
      }
    ]
  }
}
```

**17.5 风险命中复核**
```
PUT /api/v1/admin/risk/hits/{id}/review
```
**Request**
```json
{
  "status": "CONFIRMED|RELEASED",
  "reason": "string"
}
```
**Response**
```json
{"code": 0, "message": "ok", "data": {}}
```
