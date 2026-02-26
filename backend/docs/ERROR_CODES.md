# Error Codes

## Response Envelope

```json
{
  "code": 0,
  "message": "ok",
  "data": {}
}
```

- `code=0`: success
- `code!=0`: business error

## Common Codes

- `40001`: invalid request payload / parameter binding error
- `40002`: missing required business field (for example unlock request missing `phone` and `ip`)
- `40003`: invalid `date_from` format, expected `YYYY-MM-DD`
- `40004`: invalid `date_to` format, expected `YYYY-MM-DD`

- `40101`: missing authorization header
- `40102`: invalid authorization format (must be `Bearer <token>`)
- `40103`: invalid token
- `40104`: invalid credentials
- `40105`: invalid token type (non-access token used for protected APIs)

- `40301`: role missing in context
- `40302`: insufficient permission
- `40303`: user status invalid (for example disabled or banned)

- `40401`: article not found / no permission to view detail
- `40402`: attachment not found
- `40403`: stock recommendation not found
- `40404`: futures strategy not found

- `40901`: duplicate callback
- `40902`: phone already exists

- `42901`: too many failed attempts (risk control lock)

- `50001`: internal server error
- `50301`: dependency unavailable (for example auth DB unavailable)

## Notes

- Protected APIs require `Authorization: Bearer <access_token>`.
- Admin APIs require role `ADMIN`.
- CSV export APIs return `text/csv` on success and JSON error envelope on failure.
