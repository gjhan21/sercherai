import http from "../shared/lib/http";

export function trackExperimentEvent(payload) {
  return http.post("/public/experiments/events", payload);
}
