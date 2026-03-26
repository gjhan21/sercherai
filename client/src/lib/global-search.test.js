import test from "node:test";
import assert from "node:assert/strict";
import {
  buildSearchPageQuery,
  normalizeGlobalSearchKeyword,
  normalizeGlobalSearchResult,
  shouldRequestGlobalSearch
} from "./global-search.js";

test("normalizeGlobalSearchKeyword trims repeated spaces", () => {
  assert.equal(normalizeGlobalSearchKeyword("  宁德   时代  "), "宁德 时代");
  assert.equal(normalizeGlobalSearchKeyword(" "), "");
});

test("shouldRequestGlobalSearch requires at least two characters", () => {
  assert.equal(shouldRequestGlobalSearch(""), false);
  assert.equal(shouldRequestGlobalSearch("中"), false);
  assert.equal(shouldRequestGlobalSearch("中国"), true);
});

test("buildSearchPageQuery keeps q and optional focus query", () => {
  assert.deepEqual(buildSearchPageQuery(" 宁德 时代 "), {
    q: "宁德 时代"
  });
  assert.deepEqual(
    buildSearchPageQuery(" 宁德 时代 ", {
      focusType: "stocks",
      focusID: "sr_001"
    }),
    {
      q: "宁德 时代",
      focus_type: "stocks",
      focus_id: "sr_001"
    }
  );
});

test("normalizeGlobalSearchResult keeps grouped totals and scope", () => {
  const result = normalizeGlobalSearchResult(
    {
      keyword: "宁德",
      scope: "user",
      stocks: { items: [{ id: "s1" }], total: 3 },
      strategies: { items: [{ id: "f1" }], total: 2 },
      news: { items: [{ id: "n1" }], total: 4 }
    },
    "宁德"
  );

  assert.equal(result.keyword, "宁德");
  assert.equal(result.scope, "USER");
  assert.equal(result.stocks.total, 3);
  assert.equal(result.strategies.total, 2);
  assert.equal(result.news.total, 4);
});
