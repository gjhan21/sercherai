from __future__ import annotations

import json
from urllib import error, request

from app.schemas.research import ResearchGraphSnapshot, StrategyGraphWriteResult
from app.settings import Settings


class StrategyGraphClient:
    def __init__(self, settings: Settings) -> None:
        self._base_url = settings.graph_service_base_url.rstrip("/")
        self._timeout_seconds = max(settings.graph_service_timeout_ms / 1000, 0.5)

    def enabled(self) -> bool:
        return bool(self._base_url)

    def write_snapshot(self, snapshot: ResearchGraphSnapshot) -> StrategyGraphWriteResult:
        if not self.enabled():
            return StrategyGraphWriteResult(status="SKIPPED")

        payload = snapshot.model_dump(mode="json")
        endpoint = f"{self._base_url}/internal/v1/graph/snapshots"
        body = json.dumps(payload, ensure_ascii=False).encode("utf-8")
        req = request.Request(
            endpoint,
            data=body,
            headers={"Content-Type": "application/json"},
            method="POST",
        )
        try:
            with request.urlopen(req, timeout=self._timeout_seconds) as resp:
                raw = resp.read().decode("utf-8")
            parsed = json.loads(raw or "{}")
            return StrategyGraphWriteResult(
                snapshot_id=str(parsed.get("snapshot_id", "") or ""),
                backend=str(parsed.get("backend", "") or ""),
                node_count=int(parsed.get("node_count", 0) or 0),
                relation_count=int(parsed.get("relation_count", 0) or 0),
                status="WRITTEN",
            )
        except error.HTTPError as exc:
            detail = exc.read().decode("utf-8", errors="ignore")
            return StrategyGraphWriteResult(
                status="FAILED",
                error_message=f"http {exc.code}: {detail or exc.reason}",
            )
        except Exception as exc:  # pragma: no cover - defensive guard
            return StrategyGraphWriteResult(
                status="FAILED",
                error_message=str(exc),
            )
