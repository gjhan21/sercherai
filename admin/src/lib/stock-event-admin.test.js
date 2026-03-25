import test from "node:test";
import assert from "node:assert/strict";

import {
  buildStockEventSubgraphViewModel
} from "./stock-event-admin.js";

test("buildStockEventSubgraphViewModel maps graph subgraph entities and relations to nodes and edges", () => {
  const model = buildStockEventSubgraphViewModel({
    entity: {
      entity_type: "StockEvent",
      entity_key: "sec_demo_001",
      label: "白酒景气事件"
    },
    entities: [
      {
        entity_type: "StockEvent",
        entity_key: "sec_demo_001",
        label: "白酒景气事件"
      },
      {
        entity_type: "Stock",
        entity_key: "600519.SH",
        label: "贵州茅台"
      }
    ],
    relations: [
      {
        relation_type: "AFFECTS",
        source_type: "StockEvent",
        source_key: "sec_demo_001",
        target_type: "Stock",
        target_key: "600519.SH"
      }
    ],
    matched_snapshot_ids: ["reviewed-event-sec_demo_001"]
  });

  assert.equal(model.nodes.length, 2);
  assert.equal(model.edges.length, 1);
  assert.equal(model.nodes[0].entity_key, "sec_demo_001");
  assert.equal(model.edges[0].relation_type, "AFFECTS");
  assert.equal(model.warning_message, "");
});

test("buildStockEventSubgraphViewModel preserves warning-only payloads", () => {
  const model = buildStockEventSubgraphViewModel({
    warning_message: "图谱增强失败，可稍后重试"
  });

  assert.equal(model.nodes.length, 0);
  assert.equal(model.edges.length, 0);
  assert.equal(model.warning_message, "图谱增强失败，可稍后重试");
});
