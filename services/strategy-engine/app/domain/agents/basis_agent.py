from __future__ import annotations

from app.schemas.simulation import AgentOpinion


class BasisAgent:
    name = "basis"

    def evaluate(self, context: dict) -> AgentOpinion:
        basis_pct = float(context.get("basis_pct", 0))
        carry_pct = float(context.get("carry_pct", 0))
        term_structure_pct = float(context.get("term_structure_pct", 0))
        curve_slope_pct = float(context.get("curve_slope_pct", 0))
        inventory_pressure = float(context.get("inventory_pressure", 0))
        inventory_focus_area = str(context.get("inventory_focus_area", "")).strip()
        inventory_focus_warehouse = str(context.get("inventory_focus_warehouse", "")).strip()
        inventory_focus_brand = str(context.get("inventory_focus_brand", "")).strip()
        inventory_focus_place = str(context.get("inventory_focus_place", "")).strip()
        inventory_focus_grade = str(context.get("inventory_focus_grade", "")).strip()
        inventory_area_share = float(context.get("inventory_area_share", 0))
        inventory_warehouse_share = float(context.get("inventory_warehouse_share", 0))
        inventory_brand_share = float(context.get("inventory_brand_share", 0))
        inventory_place_share = float(context.get("inventory_place_share", 0))
        inventory_grade_share = float(context.get("inventory_grade_share", 0))
        spread_pressure = float(context.get("spread_pressure", 0))
        pe_ttm = float(context.get("pe_ttm", 0))
        if basis_pct or carry_pct or term_structure_pct or curve_slope_pct or inventory_pressure or spread_pressure:
            inventory_focus_share = max(
                inventory_area_share,
                inventory_warehouse_share,
                inventory_brand_share,
                inventory_place_share,
                inventory_grade_share,
            )
            inventory_focus_label = (
                inventory_focus_area
                or inventory_focus_warehouse
                or inventory_focus_brand
                or inventory_focus_place
                or inventory_focus_grade
            )
            if ((carry_pct > 0 and term_structure_pct >= 0 and curve_slope_pct >= 0) or basis_pct < 0) and spread_pressure > -0.12 and inventory_pressure >= -0.1:
                summary = "期限结构、仓单变化与价差压力共同对当前方向形成确认。"
                if inventory_focus_label and inventory_focus_share >= 0.35:
                    summary = f"期限结构与仓单结构共同确认，重点集中在{inventory_focus_label}。"
                return AgentOpinion(agent=self.name, stance="POSITIVE", confidence=71, summary=summary)
            if inventory_pressure <= -0.18:
                summary = "仓单累库对结构方向形成压制，需防守处理。"
                if inventory_focus_label and inventory_focus_share >= 0.35:
                    summary = f"{inventory_focus_label}仓单集中且累库，结构压力偏空。"
                return AgentOpinion(agent=self.name, stance="NEGATIVE", confidence=64, summary=summary)
            if spread_pressure <= -0.15:
                return AgentOpinion(agent=self.name, stance="NEGATIVE", confidence=64, summary="关联跨期价差处于不利腿，结构压力偏空。")
            if carry_pct and term_structure_pct and ((carry_pct > 0 > term_structure_pct) or (carry_pct < 0 < term_structure_pct)):
                return AgentOpinion(agent=self.name, stance="CAUTION", confidence=60, summary="期限结构与跨期斜率存在分歧，需降低结构因子权重。")
            summary = "基差与跨期结构一般，策略需依赖其他因子。"
            if inventory_focus_label and inventory_focus_share >= 0.4:
                summary = f"仓单主要集中在{inventory_focus_label}，但结构信号仍需其他因子确认。"
            return AgentOpinion(agent=self.name, stance="CAUTION", confidence=58, summary=summary)
        if 0 < pe_ttm < 30:
            return AgentOpinion(agent=self.name, stance="POSITIVE", confidence=63, summary="估值锚相对稳定，可支持观点延续。")
        if pe_ttm >= 45:
            return AgentOpinion(agent=self.name, stance="NEGATIVE", confidence=64, summary="估值锚偏贵，安全边际不足。")
        return AgentOpinion(agent=self.name, stance="CAUTION", confidence=55, summary="锚点信号中性，维持观察。")
