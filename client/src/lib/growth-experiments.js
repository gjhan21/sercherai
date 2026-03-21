const STORAGE_KEY = "sercherai_growth_experiments";

function getSearchParams() {
  if (typeof window === "undefined") {
    return null;
  }
  try {
    return new URLSearchParams(window.location.search || "");
  } catch {
    return null;
  }
}

function readExperimentMap() {
  if (typeof window === "undefined") {
    return {};
  }
  try {
    const raw = window.localStorage.getItem(STORAGE_KEY);
    return raw ? JSON.parse(raw) : {};
  } catch {
    return {};
  }
}

function saveExperimentMap(map) {
  if (typeof window === "undefined") {
    return;
  }
  try {
    window.localStorage.setItem(STORAGE_KEY, JSON.stringify(map));
  } catch {
    // Ignore storage failures and fall back to in-memory selection.
  }
}

function hashString(text) {
  let hash = 0;
  const source = String(text || "");
  for (let index = 0; index < source.length; index += 1) {
    hash = (hash * 31 + source.charCodeAt(index)) >>> 0;
  }
  return hash;
}

export function getExperimentVariant(experimentKey, variants = ["A", "B"]) {
  const keys = Array.isArray(variants) ? variants.filter(Boolean) : [];
  if (!experimentKey || keys.length === 0) {
    return "A";
  }

  const overrideKey = `exp_${experimentKey}`;
  const params = getSearchParams();
  const overrideValue = params?.get(overrideKey);
  if (overrideValue && keys.includes(overrideValue)) {
    const nextMap = {
      ...readExperimentMap(),
      [experimentKey]: overrideValue
    };
    saveExperimentMap(nextMap);
    return overrideValue;
  }

  const storedMap = readExperimentMap();
  if (keys.includes(storedMap[experimentKey])) {
    return storedMap[experimentKey];
  }

  const assigned = keys[hashString(experimentKey) % keys.length];
  saveExperimentMap({
    ...storedMap,
    [experimentKey]: assigned
  });
  return assigned;
}
