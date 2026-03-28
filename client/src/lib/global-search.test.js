import test from "node:test";
import assert from "node:assert/strict";
import {
  buildGlobalSearchGroups,
  describeSearchGroupCount,
  buildSearchPreviewGroups,
  buildSearchPageQuery,
  buildSearchItemRoute,
  dedupeVisibleSearchGroups,
  pickBestSearchMatch,
  normalizeGlobalSearchKeyword,
  normalizeGlobalSearchResult,
  resolveSearchInitialTab,
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

test("resolveSearchInitialTab prefers focused group when route focus hits", () => {
  const result = normalizeGlobalSearchResult({
    keyword: "600",
    scope: "user",
    stocks: { items: [{ id: "sr_001", symbol: "600000.SH", name: "浦发银行" }], total: 1 },
    strategies: { items: [], total: 0 },
    news: { items: [{ id: "n1", title: "600000.SH 公告" }], total: 1 }
  });

  assert.equal(
    resolveSearchInitialTab(result, {
      focusType: "stocks",
      focusID: "sr_001"
    }),
    "stocks"
  );
});

test("pickBestSearchMatch returns focused item first", () => {
  const result = normalizeGlobalSearchResult({
    keyword: "600",
    scope: "user",
    stocks: { items: [{ id: "sr_001", symbol: "600000.SH", name: "浦发银行", reason_summary: "银行修复" }], total: 1 },
    strategies: { items: [], total: 0 },
    news: { items: [{ id: "n1", title: "600000.SH 公告", summary: "新闻摘要" }], total: 1 }
  });

  const match = pickBestSearchMatch(result, {
    focusType: "stocks",
    focusID: "sr_001"
  });

  assert.equal(match.group.key, "stocks");
  assert.equal(match.item.id, "sr_001");
  assert.equal(match.isFocused, true);
});

test("buildSearchPreviewGroups moves best-match group first and trims all view", () => {
  const result = normalizeGlobalSearchResult({
    keyword: "600",
    scope: "user",
    stocks: {
      items: [
        { id: "sr_001", symbol: "600000.SH", name: "浦发银行", reason_summary: "银行修复" },
        { id: "sr_002", symbol: "600111.SH", name: "北方稀土", reason_summary: "稀土波动" },
        { id: "sr_003", symbol: "600519.SH", name: "贵州茅台", reason_summary: "消费龙头" },
        { id: "sr_004", symbol: "600036.SH", name: "招商银行", reason_summary: "金融核心" }
      ],
      total: 4
    },
    strategies: { items: [], total: 0 },
    news: {
      items: [
        { id: "n1", title: "600000.SH 公告", summary: "新闻摘要 1" },
        { id: "n2", title: "600519.SH 公告", summary: "新闻摘要 2" }
      ],
      total: 2
    }
  });

  const groups = buildGlobalSearchGroups(result);
  const previews = buildSearchPreviewGroups(groups, {
    bestMatchGroupKey: "stocks",
    perGroupLimit: 3
  });

  assert.equal(previews[0].key, "stocks");
  assert.equal(previews[0].items.length, 3);
  assert.equal(previews[1].key, "news");
});

test("buildGlobalSearchGroups avoids duplicated stock title when backend name equals symbol", () => {
  const result = normalizeGlobalSearchResult({
    keyword: "600",
    scope: "user",
    stocks: {
      items: [{ id: "sr_600036", symbol: "600036", name: "600036", reason_summary: "银行修复" }],
      total: 1
    },
    strategies: { items: [], total: 0 },
    news: { items: [], total: 0 }
  });

  const groups = buildGlobalSearchGroups(result);

  assert.equal(groups[0].items[0].title, "600036");
});

test("dedupeVisibleSearchGroups merges repeated stock results and preserves focused id", () => {
  const result = normalizeGlobalSearchResult({
    keyword: "600",
    scope: "user",
    stocks: {
      items: [
        { id: "sr_1", symbol: "600036", name: "600036", reason_summary: "首条理由" },
        { id: "sr_2", symbol: "600036", name: "600036", reason_summary: "聚焦理由" }
      ],
      total: 2
    },
    strategies: { items: [], total: 0 },
    news: { items: [], total: 0 }
  });

  const groups = buildGlobalSearchGroups(result);
  const deduped = dedupeVisibleSearchGroups(groups, {
    focusType: "stocks",
    focusID: "sr_2"
  });

  assert.equal(deduped[0].items.length, 1);
  assert.equal(deduped[0].items[0].id, "sr_2");
  assert.equal(deduped[0].items[0].summary, "聚焦理由");
});

test("describeSearchGroupCount explains when stock results are deduped for display", () => {
  assert.equal(
    describeSearchGroupCount({
      key: "stocks",
      total: 46,
      items: [{ id: "sr_1" }, { id: "sr_2" }]
    }),
    "46 条结果，去重展示 2 条"
  );
});

test("buildSearchItemRoute carries keyword when opening news result", () => {
  assert.deepEqual(
    buildSearchItemRoute("news", {
      id: "na_001",
      title: "华电能源：股票交易异常波动 14个交易日累计涨140.07%"
    }),
    {
      path: "/news",
      query: {
        article_id: "na_001",
        keyword: "华电能源：股票交易异常波动 14个交易日累计涨140.07%"
      }
    }
  );
});
