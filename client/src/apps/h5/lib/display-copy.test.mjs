import test from "node:test";
import assert from "node:assert/strict";
import { shapeNewsDisplayTitle, shapeStrategyDisplayTitle } from "./display-copy.js";

test("shapeNewsDisplayTitle removes noisy promo wrappers and trims long English book titles into mobile-friendly labels", () => {
  assert.equal(
    shapeNewsDisplayTitle("【畅销书】You Deserve to Be Rich: Master the Inner Game of Wealth and Claim Your Future Rashad Bilal", "图书"),
    "畅销书导读：You Deserve to Be Rich"
  );

  assert.equal(
    shapeNewsDisplayTitle("【亚马逊畅销书】Don't Believe Everything You Think", "图书"),
    "图书导读：Don't Believe Everything You Think"
  );
});

test("shapeNewsDisplayTitle keeps strong Chinese titles readable", () => {
  assert.equal(
    shapeNewsDisplayTitle("机构周报：景气板块进入估值切换窗口", "研报"),
    "机构周报：景气板块进入估值切换窗口"
  );
});

test("shapeStrategyDisplayTitle prefers the human-readable asset name over raw exchange codes", () => {
  assert.equal(
    shapeStrategyDisplayTitle({ title: "600036.SH 招商银行", symbol: "600036.SH", name: "招商银行" }),
    "招商银行"
  );

  assert.equal(
    shapeStrategyDisplayTitle({ title: "IF2603 股指趋势跟踪", contract: "IF2603", name: "股指趋势跟踪" }),
    "股指趋势跟踪"
  );
});
