import test from "node:test";
import assert from "node:assert/strict";
import { resolveHighlightTone, resolveMetricToneClasses, resolveSurfaceToneClasses, resolveStickyActionMode } from "./surface-tone.js";

test("resolveSurfaceToneClasses maps supported tones to consistent utility classes", () => {
  assert.deepEqual(resolveSurfaceToneClasses("default"), []);
  assert.deepEqual(resolveSurfaceToneClasses("soft"), ["h5-card-soft"]);
  assert.deepEqual(resolveSurfaceToneClasses("accent"), ["h5-card-accent"]);
  assert.deepEqual(resolveSurfaceToneClasses("brand"), ["h5-card-brand"]);
  assert.deepEqual(resolveSurfaceToneClasses("hero"), ["h5-card-brand", "h5-card-hero"]);
});

test("resolveStickyActionMode prefers stacked layout when two actions exist", () => {
  assert.equal(resolveStickyActionMode({ primaryLabel: "立即开通", secondaryLabel: "查看账户" }), "stacked");
  assert.equal(resolveStickyActionMode({ primaryLabel: "提交实名", secondaryLabel: "" }), "single");
  assert.equal(resolveStickyActionMode({ primaryLabel: "", secondaryLabel: "查看账户" }), "single");
});

test("resolveMetricToneClasses maps summary card tones to shared metric utilities", () => {
  assert.deepEqual(resolveMetricToneClasses("default"), []);
  assert.deepEqual(resolveMetricToneClasses("soft"), ["h5-metric-card-soft"]);
  assert.deepEqual(resolveMetricToneClasses("brand"), ["h5-metric-card-brand"]);
});

test("resolveHighlightTone emphasizes the leading summary card and softens follow-up cards", () => {
  assert.equal(resolveHighlightTone(0), "brand");
  assert.equal(resolveHighlightTone(1), "soft");
  assert.equal(resolveHighlightTone(2), "default");
  assert.equal(resolveHighlightTone(0, { emphasizeFirst: false }), "default");
});
