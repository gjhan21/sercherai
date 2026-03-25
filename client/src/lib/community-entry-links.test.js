import assert from "node:assert/strict";
import test from "node:test";

async function loadModule() {
  try {
    return await import("./community-entry-links.js");
  } catch (error) {
    assert.fail(`community-entry-links module should exist: ${error.message}`);
  }
}

test("buildCommunityListRoute keeps real list filters and carries article context", async () => {
  const { buildCommunityListRoute } = await loadModule();

  const route = buildCommunityListRoute({
    topicType: "news",
    sort: "latest",
    entrySource: "news_detail",
    targetType: "news_article",
    targetID: "article_001",
    targetSnapshot: "AI 产业链周报",
    title: "围绕《AI 产业链周报》继续讨论",
    stance: "watch"
  });

  assert.deepEqual(route, {
    path: "/community",
    query: {
      topic_type: "NEWS",
      sort: "LATEST",
      entry_source: "news_detail",
      target_type: "NEWS_ARTICLE",
      target_id: "article_001",
      target_snapshot: "AI 产业链周报",
      title: "围绕《AI 产业链周报》继续讨论",
      stance: "WATCH"
    }
  });
});

test("buildCommunityComposeRoute trims values and omits empty query fields", async () => {
  const { buildCommunityComposeRoute } = await loadModule();

  const route = buildCommunityComposeRoute({
    topicType: "stock",
    targetType: "stock",
    targetID: " 600519.SH ",
    targetSnapshot: " 600519.SH 贵州茅台 ",
    title: " 600519.SH 贵州茅台 当前值得继续跟踪吗 ",
    stance: "watch",
    timeHorizon: ""
  });

  assert.deepEqual(route, {
    path: "/community/new",
    query: {
      topic_type: "STOCK",
      target_type: "STOCK",
      target_id: "600519.SH",
      target_snapshot: "600519.SH 贵州茅台",
      title: "600519.SH 贵州茅台 当前值得继续跟踪吗",
      stance: "WATCH"
    }
  });
});

test("buildCommunityComposeRouteFromQuery forwards only supported draft fields", async () => {
  const { buildCommunityComposeRouteFromQuery } = await loadModule();

  const route = buildCommunityComposeRouteFromQuery({
    topic_type: "FUTURES",
    sort: "MOST_ACTIVE",
    mine: "topics",
    entry_source: "strategy_futures",
    target_type: "FUTURES",
    target_id: "IF2603",
    target_snapshot: "IF2603 股指趋势跟踪",
    title: "围绕 IF2603 股指趋势跟踪继续判断",
    stance: "WATCH",
    time_horizon: "SHORT"
  });

  assert.deepEqual(route, {
    path: "/community/new",
    query: {
      topic_type: "FUTURES",
      target_type: "FUTURES",
      target_id: "IF2603",
      target_snapshot: "IF2603 股指趋势跟踪",
      title: "围绕 IF2603 股指趋势跟踪继续判断",
      stance: "WATCH",
      time_horizon: "SHORT"
    }
  });
});

test("buildLinkedContentRoute maps linked targets back to real client pages", async () => {
  const { buildLinkedContentRoute } = await loadModule();

  assert.deepEqual(
    buildLinkedContentRoute({ targetType: "NEWS_ARTICLE", targetID: "article_001" }),
    {
      path: "/news",
      query: {
        article_id: "article_001"
      }
    }
  );

  assert.deepEqual(
    buildLinkedContentRoute({ targetType: "STOCK", targetID: "600519.SH" }),
    {
      path: "/strategies",
      query: {
        stock_id: "600519.SH"
      }
    }
  );

  assert.deepEqual(
    buildLinkedContentRoute({ targetType: "FUTURES", targetID: "IF2603" }),
    {
      path: "/strategies",
      query: {
        futures_id: "IF2603"
      }
    }
  );
});

test("findNewsArticleLocation returns category and article when feed already loaded", async () => {
  const { findNewsArticleLocation } = await loadModule();

  assert.deepEqual(
    findNewsArticleLocation(
      {
        news: [{ id: "article_001", categoryKey: "news" }],
        report: [{ id: "article_002", categoryKey: "report" }]
      },
      "article_002"
    ),
    {
      categoryKey: "report",
      articleID: "article_002"
    }
  );
});

test("findEntityIDByTarget matches either canonical id or external code", async () => {
  const { findEntityIDByTarget } = await loadModule();

  assert.equal(
    findEntityIDByTarget(
      [
        { id: "sr_001", symbol: "600519.SH", contract: "" },
        { id: "sr_002", symbol: "300750.SZ", contract: "" }
      ],
      "600519.SH",
      ["id", "symbol"]
    ),
    "sr_001"
  );

  assert.equal(
    findEntityIDByTarget(
      [{ id: "fs_001", contract: "IF2603" }],
      "fs_001",
      ["id", "contract"]
    ),
    "fs_001"
  );
});
