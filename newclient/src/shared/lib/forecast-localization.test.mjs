import test from "node:test";
import assert from "node:assert/strict";

import {
  localizeForecastChecklistStatus,
  localizeForecastProbability,
  localizeForecastScenarioName,
  localizeForecastText
} from "./forecast-localization.js";

test("forecast localization converts common report english into chinese", () => {
  const text = localizeForecastText(
    "# Forecast L3 Demo\nAlternative Scenarios\n- Executive summary: trend remains intact.\n- Action: wait for confirmation.\n- Follow the trend with tighter risk control.\n- Observe and confirm.\n- Reduce exposure quickly."
  );

  assert.match(text, /深度推演/);
  assert.match(text, /备选情景/);
  assert.match(text, /趋势主线仍然完整/);
  assert.match(text, /控制风险的前提下顺势跟踪/);
  assert.match(text, /先观察，再确认/);
  assert.match(text, /快速降低仓位暴露/);
  assert.doesNotMatch(text, /Forecast L3 Demo|Alternative Scenarios|Executive summary|Action:|Follow the trend with tighter risk control|Observe and confirm|Reduce exposure quickly/);
});

test("forecast localization formats scenario, checklist and probability helpers", () => {
  assert.equal(localizeForecastScenarioName("bull"), "乐观情景");
  assert.equal(localizeForecastScenarioName("base"), "基准情景");
  assert.equal(localizeForecastScenarioName("bear"), "悲观情景");
  assert.equal(localizeForecastChecklistStatus("READY"), "已就绪");
  assert.equal(localizeForecastChecklistStatus("watch"), "观察中");
  assert.equal(localizeForecastProbability(0.24), "24%");
  assert.equal(localizeForecastProbability("0.5"), "50%");
});
