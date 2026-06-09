import { trackExperimentEvent } from "@/api/experiment.js";
import { getMembershipQuota } from "@/api/membership.js";
import { hasClientSession } from "../auth/session.js";

const ANONYMOUS_ID_KEY = "sercherai_anonymous_id";
const SESSION_ID_KEY = "sercherai_session_id";
const USER_STAGE_KEY = "sercherai_user_stage";
const EXPERIMENT_PREFIX = "sercherai_exp_";

function generateUUID() {
  return Math.random().toString(36).substring(2, 15) + Math.random().toString(36).substring(2, 15);
}

export function getOrGenerateAnonymousID() {
  let id = localStorage.getItem(ANONYMOUS_ID_KEY);
  if (!id) {
    id = "anon_" + generateUUID();
    localStorage.setItem(ANONYMOUS_ID_KEY, id);
  }
  return id;
}

export function getSessionID() {
  let id = sessionStorage.getItem(SESSION_ID_KEY);
  if (!id) {
    id = "sess_" + generateUUID();
    sessionStorage.setItem(SESSION_ID_KEY, id);
  }
  return id;
}

export function getExperimentVariant(experimentKey) {
  const cacheKey = EXPERIMENT_PREFIX + experimentKey;
  let variant = localStorage.getItem(cacheKey);
  if (!variant) {
    variant = Math.random() < 0.5 ? "control" : "test_b";
    localStorage.setItem(cacheKey, variant);
  }
  return variant;
}

export async function resolveUserStage() {
  if (!hasClientSession()) {
    localStorage.removeItem(USER_STAGE_KEY);
    return "VISITOR";
  }

  let stage = localStorage.getItem(USER_STAGE_KEY);
  if (stage) {
    return stage;
  }

  try {
    const quota = await getMembershipQuota();
    const level = (quota?.member_level || "").toUpperCase();
    if (level.startsWith("VIP")) {
      stage = "VIP";
    } else {
      stage = "REGISTERED";
    }
  } catch (e) {
    console.error("Failed to resolve user stage from quota:", e);
    stage = "REGISTERED";
  }

  localStorage.setItem(USER_STAGE_KEY, stage);
  return stage;
}

export function clearUserStageCache() {
  localStorage.removeItem(USER_STAGE_KEY);
}

export async function track(eventType, pageKey, options = {}) {
  const experimentKey = options.experimentKey || "vip_pricing_v1";
  const variantKey = options.variantKey || getExperimentVariant(experimentKey);
  const userStage = await resolveUserStage();
  const anonymousID = getOrGenerateAnonymousID();
  const sessionID = getSessionID();

  const payload = {
    experiment_key: experimentKey,
    variant_key: variantKey,
    event_type: eventType,
    page_key: pageKey,
    target_key: options.targetKey || "",
    user_stage: userStage,
    anonymous_id: anonymousID,
    session_id: sessionID,
    pathname: window.location.pathname,
    referrer: document.referrer || "",
    metadata: options.metadata || {}
  };

  try {
    await trackExperimentEvent(payload);
  } catch (e) {
    console.error("[Tracker] Failed to send experiment event:", e);
  }
}

export function trackExposure(pageKey, options = {}) {
  return track("EXPOSURE", pageKey, options);
}

export function trackClick(pageKey, targetKey, options = {}) {
  return track("CLICK", pageKey, { ...options, targetKey });
}

export function trackUpgradeIntent(pageKey, options = {}) {
  return track("UPGRADE_INTENT", pageKey, options);
}

export function trackPaymentSuccess(pageKey, metadata = {}, options = {}) {
  return track("PAYMENT_SUCCESS", pageKey, { ...options, metadata });
}
