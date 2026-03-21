const API_BASE_URL = (import.meta.env.VITE_API_BASE_URL || "/api/v1").replace(/\/$/, "");
const ANALYTICS_ENDPOINT = `${API_BASE_URL}/public/experiments/events`;
const ANONYMOUS_ID_KEY = "sercherai_growth_anonymous_id";
const SESSION_ID_KEY = "sercherai_growth_session_id";
const EXPOSURE_CACHE_KEY = "sercherai_growth_exposure_cache";
const ATTRIBUTION_SOURCE_KEY = "sercherai_growth_attribution_sources";
const PENDING_ATTRIBUTION_SOURCE_KEY = "sercherai_growth_pending_attribution_sources";

function createID(prefix) {
  if (typeof crypto !== "undefined" && typeof crypto.randomUUID === "function") {
    return `${prefix}_${crypto.randomUUID()}`;
  }
  return `${prefix}_${Date.now().toString(36)}_${Math.random().toString(36).slice(2, 10)}`;
}

function readStorage(storage, key) {
  if (!storage) {
    return "";
  }
  try {
    return storage.getItem(key) || "";
  } catch {
    return "";
  }
}

function writeStorage(storage, key, value) {
  if (!storage || !value) {
    return;
  }
  try {
    storage.setItem(key, value);
  } catch {
    // Ignore storage failures and fall back to runtime values.
  }
}

function ensurePersistentID(storage, key, prefix) {
  const current = readStorage(storage, key);
  if (current) {
    return current;
  }
  const nextID = createID(prefix);
  writeStorage(storage, key, nextID);
  return nextID;
}

function readExposureCache() {
  if (typeof window === "undefined") {
    return {};
  }
  try {
    const raw = window.sessionStorage.getItem(EXPOSURE_CACHE_KEY);
    return raw ? JSON.parse(raw) : {};
  } catch {
    return {};
  }
}

function writeExposureCache(cache) {
  if (typeof window === "undefined") {
    return;
  }
  try {
    window.sessionStorage.setItem(EXPOSURE_CACHE_KEY, JSON.stringify(cache));
  } catch {
    // Ignore session cache write errors.
  }
}

function readAttributionSources() {
  if (typeof window === "undefined") {
    return [];
  }
  try {
    const raw = window.sessionStorage.getItem(ATTRIBUTION_SOURCE_KEY);
    const parsed = raw ? JSON.parse(raw) : [];
    return Array.isArray(parsed) ? parsed : [];
  } catch {
    return [];
  }
}

function writeAttributionSources(items) {
  if (typeof window === "undefined") {
    return;
  }
  try {
    window.sessionStorage.setItem(ATTRIBUTION_SOURCE_KEY, JSON.stringify(items));
  } catch {
    // Ignore attribution cache write errors.
  }
}

function readPendingAttributionSources() {
  if (typeof window === "undefined") {
    return [];
  }
  try {
    const raw = window.sessionStorage.getItem(PENDING_ATTRIBUTION_SOURCE_KEY);
    const parsed = raw ? JSON.parse(raw) : [];
    return Array.isArray(parsed) ? parsed : [];
  } catch {
    return [];
  }
}

function writePendingAttributionSources(items) {
  if (typeof window === "undefined") {
    return;
  }
  try {
    window.sessionStorage.setItem(PENDING_ATTRIBUTION_SOURCE_KEY, JSON.stringify(items));
  } catch {
    // Ignore pending attribution cache write errors.
  }
}

function normalizeText(value, maxLength = 255) {
  const text = String(value || "").trim();
  if (!text) {
    return "";
  }
  return text.slice(0, maxLength);
}

function resolveDeviceType() {
  if (typeof navigator === "undefined") {
    return "UNKNOWN";
  }
  const userAgent = String(navigator.userAgent || "").toLowerCase();
  const viewportWidth = typeof window !== "undefined" ? Number(window.innerWidth || 0) : 0;
  if (/ipad|tablet/.test(userAgent)) {
    return "TABLET";
  }
  if (/mobile|android|iphone|ipod/.test(userAgent)) {
    return "MOBILE";
  }
  if (viewportWidth > 0 && viewportWidth < 768) {
    return "MOBILE";
  }
  if (viewportWidth >= 768 && viewportWidth < 1100) {
    return "TABLET";
  }
  return "DESKTOP";
}

function buildExperimentContext(options = {}) {
  if (typeof window === "undefined") {
    return null;
  }
  const experimentKey = normalizeText(options.experimentKey, 64);
  const variantKey = normalizeText(options.variantKey, 32);
  const pageKey = normalizeText(options.pageKey, 64);
  if (!experimentKey || !variantKey || !pageKey) {
    return null;
  }

  const anonymousID = ensurePersistentID(window.localStorage, ANONYMOUS_ID_KEY, "anon");
  const sessionID = ensurePersistentID(window.sessionStorage, SESSION_ID_KEY, "sess");
  const rawMetadata = options.metadata && typeof options.metadata === "object" ? options.metadata : {};
  return {
    experiment_key: experimentKey,
    variant_key: variantKey,
    page_key: pageKey,
    target_key: normalizeText(options.targetKey, 64),
    user_stage: normalizeText(options.userStage || "UNKNOWN", 32).toUpperCase(),
    anonymous_id: normalizeText(options.anonymousID || anonymousID, 64),
    session_id: normalizeText(options.sessionID || sessionID, 64),
    pathname: normalizeText(options.pathname || window.location?.pathname || "", 255),
    referrer: normalizeText(options.referrer || document.referrer || "", 255),
    metadata: {
      device_type: normalizeText(rawMetadata.device_type || resolveDeviceType(), 32).toUpperCase() || "UNKNOWN",
      ...rawMetadata
    }
  };
}

export function trackExperimentEvent(options = {}) {
  const context = buildExperimentContext(options);
  const eventType = normalizeText(options.eventType, 32).toUpperCase();
  if (!context || !eventType) {
    return Promise.resolve(false);
  }
  const payload = {
    ...context,
    event_type: eventType
  };
  return sendExperimentPayload(payload);
}

export function createExperimentContext(options = {}) {
  return buildExperimentContext(options);
}

export function rememberExperimentAttributionSource(options = {}) {
  const context = buildExperimentContext(options);
  if (!context) {
    return null;
  }
  const now = Date.now();
  const nextSource = {
    ...context,
    saved_at: now
  };
  const dedupeKey = [
    context.experiment_key,
    context.variant_key,
    context.page_key,
    context.target_key || "",
    context.user_stage || ""
  ].join(":");
  const nextItems = readAttributionSources()
    .filter((item) => {
      const itemKey = [
        item?.experiment_key || "",
        item?.variant_key || "",
        item?.page_key || "",
        item?.target_key || "",
        item?.user_stage || ""
      ].join(":");
      return itemKey !== dedupeKey;
    })
    .concat(nextSource)
    .slice(-10);
  writeAttributionSources(nextItems);
  return nextSource;
}

export function rememberPendingExperimentAttributionSource(options = {}) {
  const context = buildExperimentContext(options);
  if (!context) {
    return null;
  }
  const now = Date.now();
  const nextSource = {
    ...context,
    saved_at: now
  };
  const dedupeKey = [
    context.experiment_key,
    context.variant_key,
    context.page_key,
    context.target_key || ""
  ].join(":");
  const nextItems = readPendingAttributionSources()
    .filter((item) => {
      const itemKey = [
        item?.experiment_key || "",
        item?.variant_key || "",
        item?.page_key || "",
        item?.target_key || ""
      ].join(":");
      return itemKey !== dedupeKey;
    })
    .concat(nextSource)
    .slice(-10);
  writePendingAttributionSources(nextItems);
  return nextSource;
}

export function rememberPendingExperimentJourneySource(options = {}) {
  const redirectPath = normalizeText(options.redirectPath, 255);
  const metadata = {
    ...(options.metadata && typeof options.metadata === "object" ? options.metadata : {}),
    journey_type: normalizeText(options.journeyType || "membership", 32) || "membership"
  };
  if (redirectPath) {
    metadata.destination_after_auth = redirectPath;
  }
  return rememberPendingExperimentAttributionSource({
    ...options,
    metadata
  });
}

export function listExperimentAttributionSources(options = {}) {
  const maxAgeMs = Number(options.maxAgeMs) > 0 ? Number(options.maxAgeMs) : 30 * 60 * 1000;
  const excludeExperimentKey = normalizeText(options.excludeExperimentKey, 64);
  const now = Date.now();
  const filtered = readAttributionSources().filter((item) => {
    const savedAt = Number(item?.saved_at || 0);
    if (!savedAt || now - savedAt > maxAgeMs) {
      return false;
    }
    if (excludeExperimentKey && item?.experiment_key === excludeExperimentKey) {
      return false;
    }
    return item?.experiment_key && item?.variant_key && item?.page_key;
  });
  if (filtered.length !== readAttributionSources().length) {
    writeAttributionSources(filtered);
  }
  return filtered.map(({ saved_at, ...item }) => item);
}

export function promotePendingExperimentAttributionSources(options = {}) {
  const maxAgeMs = Number(options.maxAgeMs) > 0 ? Number(options.maxAgeMs) : 30 * 60 * 1000;
  const experimentKey = normalizeText(options.experimentKey, 64);
  const pageKey = normalizeText(options.pageKey, 64);
  const targetKey = normalizeText(options.targetKey, 64);
  const userStage = normalizeText(options.userStage || "UNKNOWN", 32).toUpperCase();
  const metadata = options.metadata && typeof options.metadata === "object" ? options.metadata : {};
  const now = Date.now();
  const promotedItems = [];
  const remainingItems = [];

  readPendingAttributionSources().forEach((item) => {
    const savedAt = Number(item?.saved_at || 0);
    if (!savedAt || now - savedAt > maxAgeMs) {
      return;
    }
    if (experimentKey && item?.experiment_key !== experimentKey) {
      remainingItems.push(item);
      return;
    }
    if (pageKey && item?.page_key !== pageKey) {
      remainingItems.push(item);
      return;
    }
    const promoted = rememberExperimentAttributionSource({
      experimentKey: item?.experiment_key || experimentKey,
      variantKey: item?.variant_key || "",
      pageKey: item?.page_key || pageKey,
      targetKey: targetKey || item?.target_key || "",
      userStage: userStage || item?.user_stage || "UNKNOWN",
      anonymousID: item?.anonymous_id || "",
      sessionID: item?.session_id || "",
      pathname: item?.pathname || "",
      referrer: item?.referrer || "",
      metadata: {
        ...(item?.metadata && typeof item.metadata === "object" ? item.metadata : {}),
        ...metadata,
        attribution_flow: "post_auth_return"
      }
    });
    if (promoted) {
      promotedItems.push(promoted);
    }
  });

  writePendingAttributionSources(remainingItems);
  return promotedItems;
}

export function promotePendingExperimentJourneySources(options = {}) {
  const metadata = {
    ...(options.metadata && typeof options.metadata === "object" ? options.metadata : {}),
    journey_type: normalizeText(options.journeyType || "membership", 32) || "membership"
  };
  return promotePendingExperimentAttributionSources({
    ...options,
    metadata
  });
}

export function clearExperimentAttributionSources(experimentKey = "") {
  const normalizedExperimentKey = normalizeText(experimentKey, 64);
  if (typeof window === "undefined") {
    return;
  }
  if (!normalizedExperimentKey) {
    writeAttributionSources([]);
    return;
  }
  const nextItems = readAttributionSources().filter((item) => item?.experiment_key !== normalizedExperimentKey);
  writeAttributionSources(nextItems);
}

export function clearPendingExperimentAttributionSources(experimentKey = "") {
  const normalizedExperimentKey = normalizeText(experimentKey, 64);
  if (typeof window === "undefined") {
    return;
  }
  if (!normalizedExperimentKey) {
    writePendingAttributionSources([]);
    return;
  }
  const nextItems = readPendingAttributionSources().filter((item) => item?.experiment_key !== normalizedExperimentKey);
  writePendingAttributionSources(nextItems);
}

function sendExperimentPayload(payload) {
  if (!payload) {
    return Promise.resolve(false);
  }
  const body = JSON.stringify(payload);

  if (typeof navigator !== "undefined" && typeof navigator.sendBeacon === "function") {
    try {
      const blob = new Blob([body], { type: "application/json" });
      if (navigator.sendBeacon(ANALYTICS_ENDPOINT, blob)) {
        return Promise.resolve(true);
      }
    } catch {
      // Fall through to fetch.
    }
  }

  return fetch(ANALYTICS_ENDPOINT, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body,
    keepalive: true
  })
    .then(() => true)
    .catch(() => false);
}

export function trackExperimentExposureOnce(options = {}) {
  const context = buildExperimentContext(options);
  if (!context || typeof window === "undefined") {
    return Promise.resolve(false);
  }
  const payload = {
    ...context,
    event_type: "EXPOSURE"
  };
  const exposureKey = [
    payload.experiment_key,
    payload.variant_key,
    payload.page_key,
    payload.target_key || "page",
    payload.user_stage || "UNKNOWN"
  ].join(":");
  const exposureCache = readExposureCache();
  if (exposureCache[exposureKey]) {
    return Promise.resolve(false);
  }
  exposureCache[exposureKey] = Date.now();
  writeExposureCache(exposureCache);
  return sendExperimentPayload(payload);
}
