from __future__ import annotations

import logging
import time

from app.domain.pipelines.futures_strategy_pipeline import FuturesStrategyPipeline
from app.domain.pipelines.stock_selection_pipeline import StockSelectionPipeline
from app.core.job_store import InMemoryJobStore
from app.schemas.job import JobResult

logger = logging.getLogger(__name__)


class JobRunner:
    def __init__(
        self,
        store: InMemoryJobStore,
        simulate_delay_seconds: float,
        stock_selection_pipeline: StockSelectionPipeline,
        futures_strategy_pipeline: FuturesStrategyPipeline,
    ) -> None:
        self._store = store
        self._simulate_delay_seconds = simulate_delay_seconds
        self._stock_selection_pipeline = stock_selection_pipeline
        self._futures_strategy_pipeline = futures_strategy_pipeline

    def run(self, job_id: str) -> None:
        record = self._store.mark_running(job_id)
        if record is None:
            logger.warning("job not found when starting: %s", job_id)
            return

        logger.info("job started", extra={"job_id": job_id, "job_type": record.job_type})

        try:
            result = self._dispatch(record.job_type, record.payload)
            self._store.mark_succeeded(job_id, result)
            logger.info("job finished", extra={"job_id": job_id, "job_type": record.job_type})
        except Exception as exc:  # pragma: no cover - defensive guard
            logger.exception("job failed: %s", job_id)
            self._store.mark_failed(job_id, str(exc))

    def _dispatch(self, job_type: str, payload: dict) -> JobResult:
        if job_type == "stock-selection":
            report, warnings = self._stock_selection_pipeline.run(payload)
            return JobResult(
                summary=f"stock-selection completed with {report.selected_count} publish-ready candidates",
                payload_echo=payload,
                artifacts={"report": report.model_dump(mode="json")},
                warnings=warnings,
            )

        if job_type == "futures-strategy":
            report, warnings = self._futures_strategy_pipeline.run(payload)
            return JobResult(
                summary=f"futures-strategy completed with {report.selected_count} publish-ready strategies",
                payload_echo=payload,
                artifacts={"report": report.model_dump(mode="json")},
                warnings=warnings,
            )

        if self._simulate_delay_seconds > 0:
            time.sleep(self._simulate_delay_seconds)

        return JobResult(
            summary=f"{job_type} phase-1 pipeline accepted and simulated successfully",
            payload_echo=payload,
        )
