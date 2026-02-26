# Backend Setup

Docs:
- error codes: `/Users/gjhan21/cursor/sercherai/backend/docs/ERROR_CODES.md`
- admin openapi draft: `/Users/gjhan21/cursor/sercherai/backend/docs/openapi-admin.yaml`

## Initialize MySQL schema

Default connection:
- host: `127.0.0.1`
- port: `3306`
- user: `root`
- password: `abc123`
- database: `sercherai`

Run:

```bash
cd /Users/gjhan21/cursor/sercherai/backend
make init-db
```

Or customize:

```bash
cd /Users/gjhan21/cursor/sercherai/backend
MYSQL_HOST=127.0.0.1 MYSQL_PORT=3306 MYSQL_USER=root MYSQL_PWD=abc123 MYSQL_DB=sercherai make init-db
```

Migrations are executed in lexicographical order from `backend/migrations`.

## Seed demo data

```bash
cd /Users/gjhan21/cursor/sercherai/backend
make seed-db
```

It seeds demo records for:
- user `u_demo_001`
- growth module data
- news attachment
- futures arbitrage opportunities and guidance

## Run backend

```bash
cd /Users/gjhan21/cursor/sercherai/backend
GOPROXY=https://proxy.golang.org,direct go mod tidy
APP_PORT=8080 go run .
```

Quick health check:

```bash
curl http://127.0.0.1:8080/healthz
```

Pagination guard:
- `page_size` is capped at `200` on server side.

## Auth (JWT)

Environment:
- `JWT_SECRET` default: `sercherai_dev_secret_change_me`
- `JWT_EXPIRE_SECONDS` default: `86400`
- `JWT_REFRESH_EXPIRE_SECONDS` default: `604800`
- `LOGIN_FAIL_THRESHOLD` default: `5`
- `LOGIN_IP_FAIL_THRESHOLD` default: `20`
- `LOGIN_IP_PHONE_THRESHOLD` default: `5`
- `LOGIN_LOCK_SECONDS` default: `900`
- `APP_ENV` default: `dev`
- `ALLOW_MOCK_LOGIN` default: `false`
- `ALLOW_JOB_SIMULATION` default: `true` (non-production)
- `PAYMENT_SIGNING_SECRET` default: empty (required to verify payment callbacks)
- `ATTACHMENT_SIGNING_SECRET` default: empty (disable signed download)
- `ATTACHMENT_SIGNING_TTL_SECONDS` default: `300`
- `PUBLIC_BASE_URL` default: `http://127.0.0.1:8080`
- role rule in current demo: user id with prefix `admin_` gets `ADMIN`, otherwise `USER`

Register:

```bash
curl -X POST http://127.0.0.1:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"phone":"13800000009","password":"abc123456","email":"u9@demo.local"}'
```

Get token (user):

```bash
curl -X POST http://127.0.0.1:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"phone":"13800000001","password":"abc123456"}'
```

Get token (admin):

```bash
curl -X POST http://127.0.0.1:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"phone":"13800000000","password":"abc123456"}'
```

Dev-only mock token:

```bash
curl -X POST http://127.0.0.1:8080/api/v1/auth/mock-login \
  -H "Content-Type: application/json" \
  -d '{"user_id":"u_demo_001","role":"USER"}'
```

Refresh token:

```bash
curl -X POST http://127.0.0.1:8080/api/v1/auth/refresh \
  -H "Content-Type: application/json" \
  -d '{"refresh_token":"<refresh_token>"}'
```

Logout:

```bash
curl -X POST http://127.0.0.1:8080/api/v1/auth/logout \
  -H "Content-Type: application/json" \
  -d '{"refresh_token":"<refresh_token>"}'
```

Logout all sessions (requires access token):

```bash
curl -X POST http://127.0.0.1:8080/api/v1/auth/logout-all \
  -H "Authorization: Bearer <access_token>"
```

Admin audit logs:

```bash
curl "http://127.0.0.1:8080/api/v1/admin/auth/login-logs?page=1&page_size=20" \
  -H "Authorization: Bearer <admin_access_token>"
```

```bash
curl "http://127.0.0.1:8080/api/v1/admin/auth/login-logs/export.csv?action=LOGIN&status=FAILED&date_from=2026-02-01&date_to=2026-02-25" \
  -H "Authorization: Bearer <admin_access_token>" -o auth_login_logs.csv
```

Admin risk config:

```bash
curl "http://127.0.0.1:8080/api/v1/admin/auth/risk-config" \
  -H "Authorization: Bearer <admin_access_token>"
```

```bash
curl -X PUT "http://127.0.0.1:8080/api/v1/admin/auth/risk-config" \
  -H "Authorization: Bearer <admin_access_token>" \
  -H "Content-Type: application/json" \
  -d '{"phone_fail_threshold":5,"ip_fail_threshold":20,"ip_phone_threshold":5,"lock_seconds":900}'
```

```bash
curl "http://127.0.0.1:8080/api/v1/admin/auth/risk-config-logs?page=1&page_size=20" \
  -H "Authorization: Bearer <admin_access_token>"
```

```bash
curl -X POST "http://127.0.0.1:8080/api/v1/admin/auth/unlock" \
  -H "Authorization: Bearer <admin_access_token>" \
  -H "Content-Type: application/json" \
  -d '{"phone":"13800000001","ip":"10.66.1.9","reason":"manual unlock for support"}'
```

```bash
curl "http://127.0.0.1:8080/api/v1/admin/auth/unlock-logs?page=1&page_size=20" \
  -H "Authorization: Bearer <admin_access_token>"
```

## News APIs

User:

```bash
curl "http://127.0.0.1:8080/api/v1/news/categories" \
  -H "Authorization: Bearer <access_token>"
```

```bash
curl "http://127.0.0.1:8080/api/v1/news/articles?page=1&page_size=20&category_id=<cid>&keyword=A股" \
  -H "Authorization: Bearer <access_token>"
```

Admin:

```bash
curl -X POST "http://127.0.0.1:8080/api/v1/admin/news/categories" \
  -H "Authorization: Bearer <admin_access_token>" \
  -H "Content-Type: application/json" \
  -d '{"name":"盘后复盘","slug":"post-market","sort":10,"visibility":"PUBLIC","status":"PUBLISHED"}'
```

```bash
curl -X POST "http://127.0.0.1:8080/api/v1/admin/news/articles" \
  -H "Authorization: Bearer <admin_access_token>" \
  -H "Content-Type: application/json" \
  -d '{"category_id":"<cid>","title":"A股复盘","summary":"摘要","content":"正文","visibility":"VIP","status":"PUBLISHED"}'
```

## Stock/Futures APIs

User:

```bash
curl "http://127.0.0.1:8080/api/v1/stocks/recommendations?trade_date=2026-02-25&page=1&page_size=10" \
  -H "Authorization: Bearer <access_token>"
```

```bash
curl "http://127.0.0.1:8080/api/v1/futures/strategies?page=1&page_size=20" \
  -H "Authorization: Bearer <access_token>"
```

Admin:

```bash
curl -X POST "http://127.0.0.1:8080/api/v1/admin/stocks/recommendations/generate-daily?trade_date=2026-02-25" \
  -H "Authorization: Bearer <admin_access_token>"
```

```bash
curl -X POST "http://127.0.0.1:8080/api/v1/admin/futures/strategies" \
  -H "Authorization: Bearer <admin_access_token>" \
  -H "Content-Type: application/json" \
  -d '{"contract":"IF2603","name":"股指趋势","direction":"LONG","risk_level":"MEDIUM","position_range":"20%-30%","valid_from":"2026-02-25T09:00:00+08:00","valid_to":"2026-02-26T15:00:00+08:00","status":"PUBLISHED","reason_summary":"趋势一致"}'
```

Admin users/dashboard:

```bash
curl "http://127.0.0.1:8080/api/v1/admin/users?page=1&page_size=20&status=ACTIVE&kyc_status=APPROVED&member_level=VIP1" \
  -H "Authorization: Bearer <admin_access_token>"
```

```bash
curl "http://127.0.0.1:8080/api/v1/admin/users/export.csv?status=ACTIVE" \
  -H "Authorization: Bearer <admin_access_token>" -o admin_users.csv
```

```bash
curl -X PUT "http://127.0.0.1:8080/api/v1/admin/users/u_demo_001/member-level" \
  -H "Authorization: Bearer <admin_access_token>" \
  -H "Content-Type: application/json" \
  -d '{"member_level":"VIP2"}'
```

```bash
curl "http://127.0.0.1:8080/api/v1/admin/dashboard/overview" \
  -H "Authorization: Bearer <admin_access_token>"
```

```bash
curl "http://127.0.0.1:8080/api/v1/admin/audit/operation-logs?page=1&page_size=20&module=USER&action=UPDATE_STATUS" \
  -H "Authorization: Bearer <admin_access_token>"
```

```bash
curl "http://127.0.0.1:8080/api/v1/admin/audit/operation-logs/export.csv?module=MEMBERSHIP" \
  -H "Authorization: Bearer <admin_access_token>" -o admin_operation_logs.csv
```

Membership admin:

```bash
curl "http://127.0.0.1:8080/api/v1/admin/membership/products?page=1&page_size=20" \
  -H "Authorization: Bearer <admin_access_token>"
```

```bash
curl -X POST "http://127.0.0.1:8080/api/v1/admin/membership/products" \
  -H "Authorization: Bearer <admin_access_token>" \
  -H "Content-Type: application/json" \
  -d '{"name":"VIP季卡","price":269,"status":"ACTIVE"}'
```

```bash
curl "http://127.0.0.1:8080/api/v1/admin/membership/orders/export.csv?status=PAID" \
  -H "Authorization: Bearer <admin_access_token>" -o membership_orders.csv
```

```bash
curl "http://127.0.0.1:8080/api/v1/admin/membership/quota-configs?page=1&page_size=20&member_level=VIP2&status=ACTIVE" \
  -H "Authorization: Bearer <admin_access_token>"
```

User membership:

```bash
curl "http://127.0.0.1:8080/api/v1/membership/products?page=1&page_size=20" \
  -H "Authorization: Bearer <access_token>"
```

```bash
curl -X POST "http://127.0.0.1:8080/api/v1/membership/orders" \
  -H "Authorization: Bearer <access_token>" \
  -H "Content-Type: application/json" \
  -d '{"product_id":"mp_demo_001","pay_channel":"ALIPAY"}'
```

```bash
curl "http://127.0.0.1:8080/api/v1/membership/orders?page=1&page_size=20" \
  -H "Authorization: Bearer <access_token>"
```

```bash
curl -X POST "http://127.0.0.1:8080/api/v1/admin/membership/quota-configs" \
  -H "Authorization: Bearer <admin_access_token>" \
  -H "Content-Type: application/json" \
  -d '{"member_level":"VIP2","doc_read_limit":200,"news_subscribe_limit":100,"reset_cycle":"MONTHLY","status":"ACTIVE","effective_at":"2026-03-01T00:00:00+08:00"}'
```

System config admin:

```bash
curl "http://127.0.0.1:8080/api/v1/admin/system/configs?page=1&page_size=20&keyword=model" \
  -H "Authorization: Bearer <admin_access_token>"
```

```bash
curl -X PUT "http://127.0.0.1:8080/api/v1/admin/system/configs" \
  -H "Authorization: Bearer <admin_access_token>" \
  -H "Content-Type: application/json" \
  -d '{"config_key":"stock.model.version","config_value":"v2","description":"升级到v2模型"}'
```

```bash
curl "http://127.0.0.1:8080/api/v1/admin/system/job-definitions?page=1&page_size=20&module=STOCK&status=ACTIVE" \
  -H "Authorization: Bearer <admin_access_token>"
```

```bash
curl -X POST "http://127.0.0.1:8080/api/v1/admin/system/job-definitions" \
  -H "Authorization: Bearer <admin_access_token>" \
  -H "Content-Type: application/json" \
  -d '{"job_name":"daily_stock_recommendation_v2","display_name":"每日股票推荐生成V2","module":"STOCK","cron_expr":"0 40 8 * * *","status":"ACTIVE"}'
```

```bash
curl -X PUT "http://127.0.0.1:8080/api/v1/admin/system/job-definitions/jobdef_stock_daily/status" \
  -H "Authorization: Bearer <admin_access_token>" \
  -H "Content-Type: application/json" \
  -d '{"status":"DISABLED"}'
```

```bash
curl "http://127.0.0.1:8080/api/v1/admin/system/job-runs?page=1&page_size=20&job_name=daily_stock_recommendation&status=SUCCESS" \
  -H "Authorization: Bearer <admin_access_token>"
```

```bash
curl "http://127.0.0.1:8080/api/v1/admin/system/job-runs/metrics?job_name=daily_stock_recommendation" \
  -H "Authorization: Bearer <admin_access_token>"
```

```bash
curl "http://127.0.0.1:8080/api/v1/admin/system/job-runs/export.csv?job_name=daily_stock_recommendation" \
  -H "Authorization: Bearer <admin_access_token>" -o scheduler_job_runs.csv
```

```bash
curl -X POST "http://127.0.0.1:8080/api/v1/admin/system/job-runs/trigger" \
  -H "Authorization: Bearer <admin_access_token>" \
  -H "Content-Type: application/json" \
  -d '{"job_name":"daily_stock_recommendation","trigger_source":"MANUAL","simulate_status":"SUCCESS","result_summary":"generated 10 recommendations"}'
```

```bash
curl -X POST "http://127.0.0.1:8080/api/v1/admin/system/job-runs/jr_001/retry" \
  -H "Authorization: Bearer <admin_access_token>" \
  -H "Content-Type: application/json" \
  -d '{"simulate_status":"FAILED","result_summary":"retry run","error_message":"upstream timeout"}'
```

Review workflow admin:

```bash
curl "http://127.0.0.1:8080/api/v1/admin/workflow/reviews?page=1&page_size=20&module=STOCK&status=PENDING" \
  -H "Authorization: Bearer <admin_access_token>"
```

```bash
curl "http://127.0.0.1:8080/api/v1/admin/workflow/metrics?module=STOCK&receiver_id=admin_002" \
  -H "Authorization: Bearer <admin_access_token>"
```

```bash
curl "http://127.0.0.1:8080/api/v1/admin/workflow/reviews/export.csv?module=STOCK&reviewer_id=admin_002" \
  -H "Authorization: Bearer <admin_access_token>" -o review_tasks.csv
```

```bash
curl -X POST "http://127.0.0.1:8080/api/v1/admin/workflow/reviews/submit" \
  -H "Authorization: Bearer <admin_access_token>" \
  -H "Content-Type: application/json" \
  -d '{"module":"STOCK","target_id":"sr_001","reviewer_id":"admin_002","submit_note":"提交审核"}'
```

```bash
curl -X PUT "http://127.0.0.1:8080/api/v1/admin/workflow/reviews/rt_001/assign" \
  -H "Authorization: Bearer <admin_access_token>" \
  -H "Content-Type: application/json" \
  -d '{"reviewer_id":"admin_002"}'
```

```bash
curl -X PUT "http://127.0.0.1:8080/api/v1/admin/workflow/reviews/rt_001/decision" \
  -H "Authorization: Bearer <admin_access_token>" \
  -H "Content-Type: application/json" \
  -d '{"status":"APPROVED","review_note":"通过"}'
```

```bash
curl "http://127.0.0.1:8080/api/v1/admin/workflow/messages?page=1&page_size=20&module=STOCK&is_read=false&receiver_id=admin_002" \
  -H "Authorization: Bearer <admin_access_token>"
```

```bash
curl "http://127.0.0.1:8080/api/v1/admin/workflow/messages/export.csv?module=STOCK&event_type=REVIEW_SUBMITTED&receiver_id=admin_002" \
  -H "Authorization: Bearer <admin_access_token>" -o workflow_messages.csv
```

```bash
curl "http://127.0.0.1:8080/api/v1/admin/workflow/messages/unread-count?module=STOCK&receiver_id=admin_002" \
  -H "Authorization: Bearer <admin_access_token>"
```

```bash
curl -X PUT "http://127.0.0.1:8080/api/v1/admin/workflow/messages/wm_001/read" \
  -H "Authorization: Bearer <admin_access_token>" \
  -H "Content-Type: application/json" \
  -d '{"is_read":true}'
```

```bash
curl -X PUT "http://127.0.0.1:8080/api/v1/admin/workflow/messages/read-all" \
  -H "Authorization: Bearer <admin_access_token>" \
  -H "Content-Type: application/json" \
  -d '{"module":"STOCK","event_type":"REVIEW_SUBMITTED","receiver_id":"admin_002"}'
```

Call protected API:

```bash
curl http://127.0.0.1:8080/api/v1/futures/arbitrage/opportunities \
  -H "Authorization: Bearer <your_token>"
```
