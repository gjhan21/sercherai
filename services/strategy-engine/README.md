# Strategy Engine

`strategy-engine` is the standalone service for stock-selection and futures-strategy job orchestration.

## Current scope

- phase 1: bootable FastAPI service, health endpoint, job protocol
- phase 2: stock-selection MVP with seed loading, feature scoring, risk filtering, report output
- phase 3: futures-strategy MVP with direction, price levels, leverage guard, report output
- phase 4: scenario simulation, lightweight market graph summary, multi-agent consensus and veto
- phase 5: publish archive, Markdown/HTML rendering, history compare and replay
- phase 6: admin-facing job center list API and runtime config handoff from Go backend
  - runtime scenario template override
  - publish policy enforcement and manual override publish

## Quick start

```bash
cd /Users/gjhan21/cursor/sercherai/services/strategy-engine
python3 -m venv .venv
source .venv/bin/activate
pip install -e '.[dev]'
STRATEGY_ENGINE_GO_BACKEND_BASE_URL=http://127.0.0.1:18080 \
uvicorn app.main:app --reload --port 8081
```

## Endpoints

- `GET /internal/v1/health`
- `POST /internal/v1/jobs/stock-selection`
- `POST /internal/v1/jobs/futures-strategy`
- `GET /internal/v1/jobs`
- `GET /internal/v1/jobs/{job_id}`
- `POST /internal/v1/publish/jobs/{job_id}`
- `GET /internal/v1/publish/history/{job_type}`
- `GET /internal/v1/publish/records/{publish_id}`
- `GET /internal/v1/publish/records/{publish_id}/replay`
- `POST /internal/v1/publish/compare`

## Stock-selection payload example

```json
{
  "requested_by": "operator",
  "payload": {
    "trade_date": "2026-03-17",
    "limit": 5,
    "max_risk_level": "MEDIUM",
    "min_score": 80,
    "seed_symbols": ["600519.SH", "601318.SH", "300750.SZ"]
  }
}
```

## Futures-strategy payload example

```json
{
  "requested_by": "operator",
  "payload": {
    "trade_date": "2026-03-17",
    "limit": 3,
    "contracts": ["IF2606", "IH2606", "IC2606"],
    "max_risk_level": "HIGH",
    "min_confidence": 55
  }
}
```

The job detail response includes a structured `report` artifact with publish-ready payloads for the existing Go write models.

`POST /internal/v1/publish/jobs/{job_id}` also accepts runtime publish controls such as:

```json
{
  "requested_by": "ops-admin",
  "force": false,
  "override_reason": "",
  "policy": {
    "max_risk_level": "MEDIUM",
    "max_warning_count": 3,
    "allow_vetoed_publish": false,
    "default_publisher": "strategy-engine",
    "override_note_template": "人工覆盖后需记录复盘说明。"
  }
}
```

## Report enrichments

- stock and futures reports include `simulations` for bull/base/bear/shock scenario cards
- reports include `graph_summary` for lightweight market structure context
- reports include `consensus_summary` plus per-asset agent opinions and veto state
- reports now include `context_meta`, so stock/futures explanation chains can show the real trade date, source, and news window
- publish records include versioned Markdown / HTML report snapshots and replay metadata

## Real context source

- stock-selection now loads seeds from Go backend `POST /internal/v1/strategy-engine/context/stock-selection`
- futures-strategy now loads seeds from Go backend `POST /internal/v1/strategy-engine/context/futures-strategy`
- the Go backend reads from local truth/news tables instead of letting `strategy-engine` connect to MySQL directly
- when running locally, `STRATEGY_ENGINE_GO_BACKEND_BASE_URL` must point to the Go backend base URL, otherwise stock/futures jobs cannot fetch context
- sample seeds are kept only for explicit dev/test fallback via:
  - `STRATEGY_ENGINE_ALLOW_SAMPLE_STOCK_SEEDS=true`
  - `STRATEGY_ENGINE_ALLOW_SAMPLE_FUTURES_SEEDS=true`
