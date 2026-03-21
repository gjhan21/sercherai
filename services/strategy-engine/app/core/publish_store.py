from __future__ import annotations

from collections import defaultdict
from datetime import UTC, datetime
from threading import Lock
from uuid import uuid4

from app.schemas.publish import PublishRecord, PublishRecordSummary, PublishReplay


class InMemoryPublishStore:
    def __init__(self) -> None:
        self._records: dict[str, PublishRecord] = {}
        self._versions: dict[str, int] = defaultdict(int)
        self._job_type_index: dict[str, list[str]] = defaultdict(list)
        self._lock = Lock()

    def create_record(
        self,
        *,
        job_id: str,
        job_type: str,
        trade_date: str,
        report_summary: str,
        selected_count: int,
        asset_keys: list[str],
        payload_count: int,
        markdown: str,
        html: str,
        publish_payloads: list[dict],
        report_snapshot: dict,
        replay: PublishReplay,
    ) -> PublishRecord:
        with self._lock:
            version = self._versions[job_type] + 1
            self._versions[job_type] = version
            record = PublishRecord(
                publish_id=f"publish_{uuid4().hex}",
                job_id=job_id,
                job_type=job_type,
                version=version,
                created_at=datetime.now(UTC),
                trade_date=trade_date,
                report_summary=report_summary,
                selected_count=selected_count,
                asset_keys=asset_keys,
                payload_count=payload_count,
                markdown=markdown,
                html=html,
                publish_payloads=publish_payloads,
                report_snapshot=report_snapshot,
                replay=replay,
            )
            self._records[record.publish_id] = record
            self._job_type_index[job_type].append(record.publish_id)
            return record.model_copy(deep=True)

    def get_record(self, publish_id: str) -> PublishRecord | None:
        with self._lock:
            record = self._records.get(publish_id)
            if record is None:
                return None
            return record.model_copy(deep=True)

    def list_records(self, job_type: str) -> list[PublishRecordSummary]:
        with self._lock:
            publish_ids = list(self._job_type_index.get(job_type, []))
            records = [self._records[item].model_copy(deep=True) for item in publish_ids]

        records.sort(key=lambda item: item.version, reverse=True)
        return [
            PublishRecordSummary(
                publish_id=item.publish_id,
                job_id=item.job_id,
                job_type=item.job_type,
                version=item.version,
                created_at=item.created_at,
                trade_date=item.trade_date,
                report_summary=item.report_summary,
                selected_count=item.selected_count,
                asset_keys=item.asset_keys,
                payload_count=item.payload_count,
            )
            for item in records
        ]
