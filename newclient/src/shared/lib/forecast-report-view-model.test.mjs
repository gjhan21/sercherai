import test from "node:test";
import assert from "node:assert/strict";
import { buildForecastEvidenceSections } from "./forecast-report-view-model.js";

test("buildForecastEvidenceSections maps stock dimensions to research labels", () => {
  const sections = buildForecastEvidenceSections({
    targetType: "STOCK",
    dimensionEvidence: [
      { dimension: "FLOW", stance: "CONSTRUCTIVE", confidence: 0.72, summary: "资金回流修复。", supporting_points: ["成交额回暖"], risk_points: ["放量不足"] },
      { dimension: "FUNDAMENTAL", stance: "NEUTRAL", confidence: 0.63, summary: "盈利质量稳定。", supporting_points: ["回撤可控"], risk_points: ["增速仍需确认"] },
      { dimension: "TECHNICAL", stance: "BULLISH", confidence: 0.76, summary: "趋势仍在。", supporting_points: ["20日动量为正"], risk_points: ["若跌破支撑则转弱"] },
      { dimension: "VALUATION", stance: "WATCH", confidence: 0.58, summary: "估值在可跟踪区间。", supporting_points: ["PB未失控"], risk_points: ["PE继续扩张"] },
      { dimension: "EVENT", stance: "CONSTRUCTIVE", confidence: 0.66, summary: "事件面偏正向。", supporting_points: ["资讯热度回升"], risk_points: ["催化持续性不明"] }
    ]
  });

  assert.deepEqual(
    sections.map((item) => item.label),
    ["基本面", "技术面", "资金面", "估值面", "事件面"]
  );
  assert.equal(sections[0].stanceLabel, "中性");
  assert.equal(sections[1].confidenceLabel, "76%");
});

test("buildForecastEvidenceSections maps futures dimensions to research labels", () => {
  const sections = buildForecastEvidenceSections({
    targetType: "FUTURES",
    dimensionEvidence: [
      { dimension: "SUPPLY_DEMAND", stance: "CONSTRUCTIVE", confidence: 0.71, summary: "库存延续去化。", supporting_points: ["库存压力缓和"], risk_points: ["若库存反弹则修复放缓"] },
      { dimension: "TERM_STRUCTURE", stance: "BULLISH", confidence: 0.75, summary: "近月结构占优。", supporting_points: ["基差与期限结构同向"], risk_points: ["若结构背离则主线减弱"] },
      { dimension: "TAPE_TECHNICAL", stance: "WATCH", confidence: 0.6, summary: "盘面趋势仍在。", supporting_points: ["波动受控"], risk_points: ["趋势斜率放缓"] },
      { dimension: "POSITION_FLOW", stance: "CONSTRUCTIVE", confidence: 0.67, summary: "持仓资金偏多。", supporting_points: ["流向偏置仍偏多"], risk_points: ["拥挤度抬升"] },
      { dimension: "MACRO_EVENT", stance: "NEUTRAL", confidence: 0.55, summary: "宏观事件中性偏稳。", supporting_points: ["风险偏好未恶化"], risk_points: ["突发事件扰动"] }
    ]
  });

  assert.deepEqual(
    sections.map((item) => item.label),
    ["供需库存", "期限结构", "盘面技术", "持仓资金", "宏观事件"]
  );
  assert.equal(sections[2].stanceLabel, "等待确认");
  assert.equal(sections[4].riskText, "突发事件扰动");
});
