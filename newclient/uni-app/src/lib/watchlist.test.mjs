import test from "node:test";
import assert from "node:assert/strict";

import {
  H5_WATCHLIST_KEY,
  addWatchItem,
  getWatchlistItems,
  removeWatchItem
} from "./watchlist.js";

function createStorage() {
  const bucket = new Map();
  return {
    getItem(key) {
      return bucket.has(key) ? bucket.get(key) : null;
    },
    setItem(key, value) {
      bucket.set(key, String(value));
    },
    removeItem(key) {
      bucket.delete(key);
    }
  };
}

test("addWatchItem stores unique items in H5 local watchlist", () => {
  globalThis.localStorage = createStorage();

  addWatchItem({ id: "stock-1", type: "stock", symbol: "sh600519", name: "贵州茅台" });
  addWatchItem({ id: "stock-1", type: "stock", symbol: "sh600519", name: "贵州茅台" });
  addWatchItem({ id: "future-1", type: "futures", symbol: "RB2510", name: "螺纹钢主连" });

  assert.equal(getWatchlistItems().length, 2);
  assert.match(globalThis.localStorage.getItem(H5_WATCHLIST_KEY), /贵州茅台/);
});

test("removeWatchItem deletes target item only", () => {
  globalThis.localStorage = createStorage();

  addWatchItem({ id: "stock-1", type: "stock", symbol: "sh600519", name: "贵州茅台" });
  addWatchItem({ id: "future-1", type: "futures", symbol: "RB2510", name: "螺纹钢主连" });
  removeWatchItem("stock-1");

  assert.deepEqual(getWatchlistItems(), [
    {
      id: "future-1",
      type: "futures",
      symbol: "RB2510",
      name: "螺纹钢主连"
    }
  ]);
});
