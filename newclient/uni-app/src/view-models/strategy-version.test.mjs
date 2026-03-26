import test from "node:test";
import assert from "node:assert/strict";

import { buildExplanationView } from "./strategy-version.js";

test("buildExplanationView flattens explanation proof and version meta", () => {
  const view = buildExplanationView(
    {
      confidence_reason: "趋势与资金流共振",
      graph_summary: "产业链景气度抬升",
      market_regime: "ROTATION",
      risk_boundary: "跌破止损线后退出",
      evidence_cards: [
        { title: "趋势", value: "向上", note: "20日线拐头" },
        { title: "资金", value: "净流入", note: "近三日持续" }
      ],
      theme_tags: ["新能源", "锂电"],
      version_diff: {
        added: ["新增龙头观察"],
        promoted: ["上调优先级"],
        summary: "相较上版更积极"
      }
    },
    "默认推荐理由"
  );

  assert.equal(view.whyNow, "趋势与资金流共振");
  assert.deepEqual(view.proofTags, ["新能源", "锂电", "趋势: 向上", "资金: 净流入"]);
  assert.match(view.meta, /ROTATION/);
  assert.match(view.meta, /产业链景气度抬升/);
  assert.equal(view.riskBoundary, "跌破止损线后退出");
  assert.equal(view.versionSummary, "相较上版更积极");
});
