from __future__ import annotations

from app.core.publish_store import InMemoryPublishStore
from app.domain.reports.report_renderer import ReportRenderer
from app.schemas.job import JobRecord, JobStatus
from app.schemas.publish import PublishCompareResponse, PublishJobRequest, PublishPolicyConfig, PublishRecord, PublishReplay


class GoBackendPublisher:
    def __init__(self, store: InMemoryPublishStore, renderer: ReportRenderer) -> None:
        self._store = store
        self._renderer = renderer

    def publish_job(self, record: JobRecord, request: PublishJobRequest | None = None) -> PublishRecord:
        if record.status != JobStatus.SUCCEEDED or record.result is None:
            raise ValueError("job is not ready for publish")
        request = request or PublishJobRequest()

        report = record.result.artifacts.get("report")
        if not isinstance(report, dict):
            raise ValueError("job does not contain a publishable report artifact")

        replay = _build_replay(record.job_type, report, record.result.warnings)
        _validate_publish_policy(record.job_type, report, replay, request.policy, request.force)

        if request.force and request.override_reason.strip():
            replay.notes.append(f"人工覆盖发布原因: {request.override_reason.strip()}。")
        elif request.force:
            replay.notes.append("本次发布通过人工覆盖放行。")
        if request.policy.default_publisher.strip():
            replay.notes.append(f"策略默认发布者: {request.policy.default_publisher.strip()}。")
        if request.policy.override_note_template.strip():
            replay.notes.append(request.policy.override_note_template.strip())
        replay.notes.append(f"发布操作人: {request.requested_by or 'system'}。")

        markdown = self._renderer.render_markdown(record.job_type, report)
        html = self._renderer.render_html(record.job_type, report)
        asset_keys = _asset_keys(record.job_type, report)
        publish_payloads = [item for item in report.get("publish_payloads", []) if isinstance(item, dict)]

        return self._store.create_record(
            job_id=record.job_id,
            job_type=record.job_type,
            trade_date=str(report.get("trade_date", "")),
            report_summary=str(report.get("report_summary", record.result.summary)),
            selected_count=int(report.get("selected_count", len(asset_keys))),
            asset_keys=asset_keys,
            payload_count=len(publish_payloads),
            markdown=markdown,
            html=html,
            publish_payloads=publish_payloads,
            report_snapshot=report,
            replay=replay,
        )

    def get_record(self, publish_id: str) -> PublishRecord | None:
        return self._store.get_record(publish_id)

    def list_history(self, job_type: str):
        return self._store.list_records(job_type)

    def compare(self, left_publish_id: str, right_publish_id: str) -> PublishCompareResponse:
        left = self._store.get_record(left_publish_id)
        right = self._store.get_record(right_publish_id)
        if left is None or right is None:
            raise ValueError("publish record not found")

        left_assets = set(left.asset_keys)
        right_assets = set(right.asset_keys)
        return PublishCompareResponse(
            left_publish_id=left.publish_id,
            right_publish_id=right.publish_id,
            left_version=left.version,
            right_version=right.version,
            selected_count_delta=right.selected_count - left.selected_count,
            payload_count_delta=right.payload_count - left.payload_count,
            warning_count_delta=right.replay.warning_count - left.replay.warning_count,
            veto_count_delta=len(right.replay.vetoed_assets) - len(left.replay.vetoed_assets),
            added_assets=sorted(right_assets - left_assets),
            removed_assets=sorted(left_assets - right_assets),
            shared_assets=sorted(left_assets & right_assets),
        )


def _asset_keys(job_type: str, report: dict) -> list[str]:
    if job_type == "stock-selection":
        return [str(item.get("symbol", "")) for item in report.get("candidates", []) if item.get("symbol")]
    return [str(item.get("contract", "")) for item in report.get("strategies", []) if item.get("contract")]


def _build_replay(job_type: str, report: dict, warnings: list[str]) -> PublishReplay:
    if job_type == "stock-selection":
        items = report.get("candidates", [])
        key_name = "symbol"
    else:
        items = report.get("strategies", [])
        key_name = "contract"

    invalidated_assets = [
        str(item.get(key_name, ""))
        for item in items
        if item.get(key_name) and item.get("invalidations")
    ]
    vetoed_assets = [
        str(item.get("asset_key", ""))
        for item in report.get("simulations", [])
        if item.get("asset_key") and item.get("vetoed")
    ]

    notes = []
    if warnings:
        notes.append(f"本次发布包含 {len(warnings)} 条风控或过滤提醒。")
    if vetoed_assets:
        notes.append(f"被风险 agent 否决的标的: {'、'.join(vetoed_assets)}。")
    if invalidated_assets:
        notes.append(f"已记录失效条件的标的数: {len(invalidated_assets)}。")
    if not notes:
        notes.append("本次发布未出现额外警告，可作为后续复盘基线版本。")

    return PublishReplay(
        warning_count=len(warnings),
        warning_messages=warnings,
        vetoed_assets=vetoed_assets,
        invalidated_assets=invalidated_assets,
        notes=notes,
    )


def _validate_publish_policy(
    job_type: str,
    report: dict,
    replay: PublishReplay,
    policy: PublishPolicyConfig,
    force: bool,
) -> None:
    risk_rank = {"LOW": 1, "MEDIUM": 2, "HIGH": 3}
    allowed_rank = risk_rank.get((policy.max_risk_level or "MEDIUM").upper(), 2)
    if job_type == "stock-selection":
        items = report.get("candidates", [])
        key_name = "symbol"
    else:
        items = report.get("strategies", [])
        key_name = "contract"

    breaches: list[str] = []
    high_risk_assets = [
        str(item.get(key_name, ""))
        for item in items
        if item.get(key_name) and risk_rank.get(str(item.get("risk_level", "")).upper(), 0) > allowed_rank
    ]
    if high_risk_assets:
        breaches.append(f"存在风险等级超过 {policy.max_risk_level} 的标的: {'、'.join(high_risk_assets)}")
    if replay.warning_count > policy.max_warning_count:
        breaches.append(f"警告数量 {replay.warning_count} 超过阈值 {policy.max_warning_count}")
    if replay.vetoed_assets and not policy.allow_vetoed_publish:
        breaches.append(f"存在被 veto 的标的: {'、'.join(replay.vetoed_assets)}")

    if breaches and not force:
        raise ValueError("发布策略拦截: " + "；".join(breaches))
