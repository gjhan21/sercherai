import { computed, ref } from "vue";

import { clearClientSession, getClientSession, saveClientSession } from "./session.js";

const session = ref(getClientSession());
const isLoggedIn = computed(() => Boolean(session.value?.accessToken));

export function useAuthState() {
  return {
    session,
    isLoggedIn
  };
}

export function syncClientAuthState() {
  session.value = getClientSession();
  return session.value;
}

export function setClientAuthState(payload) {
  session.value = saveClientSession(payload);
  return session.value;
}

export function clearClientAuthState() {
  clearClientSession();
  session.value = null;
}
