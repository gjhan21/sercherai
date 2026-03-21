from __future__ import annotations

from html import escape


class ReportRenderer:
    def render_markdown(self, job_type: str, report: dict) -> str:
        title = _title(job_type)
        lines = [
            f"# {title}",
            "",
            f"- 交易日: {report.get('trade_date', '')}",
            f"- 结论: {report.get('report_summary', '')}",
            f"- 风险概览: {report.get('risk_summary', '')}",
        ]

        if report.get("graph_summary"):
            lines.append(f"- 市场图谱: {report['graph_summary']}")
        if report.get("consensus_summary"):
            lines.append(f"- 多代理收敛: {report['consensus_summary']}")

        lines.extend(["", "## 核心清单"])
        for item in _entries(job_type, report):
            lines.extend(
                [
                    "",
                    f"### {item['key']} {item['name']}",
                    f"- 动作: {item['action']}",
                    f"- 风险: {item['risk_level']}",
                    f"- 仓位: {item['position_range']}",
                    f"- 说明: {item['reason_summary']}",
                ]
            )
            if item["controls"]:
                lines.append(f"- 控制位: {item['controls']}")
            if item["invalidations"]:
                lines.append(f"- 失效条件: {'；'.join(item['invalidations'])}")

        return "\n".join(lines)

    def render_html(self, job_type: str, report: dict) -> str:
        title = escape(_title(job_type))
        metadata = [
            f"<li><strong>交易日:</strong> {escape(str(report.get('trade_date', '')))}</li>",
            f"<li><strong>结论:</strong> {escape(str(report.get('report_summary', '')))}</li>",
            f"<li><strong>风险概览:</strong> {escape(str(report.get('risk_summary', '')))}</li>",
        ]
        if report.get("graph_summary"):
            metadata.append(f"<li><strong>市场图谱:</strong> {escape(str(report['graph_summary']))}</li>")
        if report.get("consensus_summary"):
            metadata.append(f"<li><strong>多代理收敛:</strong> {escape(str(report['consensus_summary']))}</li>")

        items = []
        for item in _entries(job_type, report):
            rows = [
                f"<li><strong>动作:</strong> {escape(item['action'])}</li>",
                f"<li><strong>风险:</strong> {escape(item['risk_level'])}</li>",
                f"<li><strong>仓位:</strong> {escape(item['position_range'])}</li>",
                f"<li><strong>说明:</strong> {escape(item['reason_summary'])}</li>",
            ]
            if item["controls"]:
                rows.append(f"<li><strong>控制位:</strong> {escape(item['controls'])}</li>")
            if item["invalidations"]:
                rows.append(f"<li><strong>失效条件:</strong> {escape('；'.join(item['invalidations']))}</li>")
            items.append(
                "<section>"
                f"<h3>{escape(item['key'])} {escape(item['name'])}</h3>"
                f"<ul>{''.join(rows)}</ul>"
                "</section>"
            )

        return (
            "<html><body>"
            f"<h1>{title}</h1>"
            f"<ul>{''.join(metadata)}</ul>"
            "<h2>核心清单</h2>"
            f"{''.join(items)}"
            "</body></html>"
        )


def _title(job_type: str) -> str:
    if job_type == "stock-selection":
        return "股票推荐发布报告"
    return "期货策略发布报告"


def _entries(job_type: str, report: dict) -> list[dict[str, str | list[str]]]:
    if job_type == "stock-selection":
        return [
            {
                "key": str(item.get("symbol", "")),
                "name": str(item.get("name", "")),
                "action": "生成推荐",
                "risk_level": str(item.get("risk_level", "")),
                "position_range": str(item.get("position_range", "")),
                "reason_summary": str(item.get("reason_summary", "")),
                "controls": f"止盈 {item.get('take_profit', '')} / 止损 {item.get('stop_loss', '')}",
                "invalidations": [str(value) for value in item.get("invalidations", [])],
            }
            for item in report.get("candidates", [])
        ]

    return [
        {
            "key": str(item.get("contract", "")),
            "name": str(item.get("name", "")),
            "action": str(item.get("direction", "")),
            "risk_level": str(item.get("risk_level", "")),
            "position_range": str(item.get("position_range", "")),
            "reason_summary": str(item.get("reason_summary", "")),
            "controls": (
                f"入场 {item.get('entry_price', '')} / "
                f"止盈 {item.get('take_profit_price', '')} / "
                f"止损 {item.get('stop_loss_price', '')}"
            ),
            "invalidations": [str(value) for value in item.get("invalidations", [])],
        }
        for item in report.get("strategies", [])
    ]
