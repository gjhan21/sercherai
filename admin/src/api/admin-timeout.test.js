import test from "node:test";
import assert from "node:assert/strict";
import fs from "node:fs";

const source = fs.readFileSync(new URL("./admin.js", import.meta.url), "utf8");

test("long-running data sync api helpers use dedicated timeout config", () => {
  assert.match(source, /const SYNC_REQUEST_CONFIG = \{\s*timeout:\s*\d+/);
  assert.match(source, /syncMarketDataMaster\(payload\)\s*\{\s*return http\.post\("\/admin\/market-data\/master\/sync", payload, SYNC_REQUEST_CONFIG\);/s);
  assert.match(source, /syncMarketDataQuotes\(payload\)\s*\{\s*return http\.post\("\/admin\/market-data\/quotes\/sync", payload, SYNC_REQUEST_CONFIG\);/s);
  assert.match(source, /syncMarketDataDailyBasic\(payload\)\s*\{\s*return http\.post\("\/admin\/market-data\/daily-basic\/sync", payload, SYNC_REQUEST_CONFIG\);/s);
  assert.match(source, /syncMarketDataMoneyflow\(payload\)\s*\{\s*return http\.post\("\/admin\/market-data\/moneyflow\/sync", payload, SYNC_REQUEST_CONFIG\);/s);
  assert.match(source, /syncStockInstrumentMaster\(payload\)\s*\{\s*return http\.post\("\/admin\/stocks\/master\/sync", payload, SYNC_REQUEST_CONFIG\);/s);
  assert.match(source, /syncStockQuotes\(payload\)\s*\{\s*return http\.post\("\/admin\/stocks\/quotes\/sync", payload, SYNC_REQUEST_CONFIG\);/s);
  assert.match(source, /syncFuturesQuotes\(payload\)\s*\{\s*return http\.post\("\/admin\/futures\/quotes\/sync", payload, SYNC_REQUEST_CONFIG\);/s);
  assert.match(source, /syncFuturesInventory\(payload\)\s*\{\s*return http\.post\("\/admin\/futures\/inventory\/sync", payload, SYNC_REQUEST_CONFIG\);/s);
  assert.match(source, /syncMarketNewsSource\(payload\)\s*\{\s*return http\.post\("\/admin\/news\/market-sync", payload, SYNC_REQUEST_CONFIG\);/s);
});
