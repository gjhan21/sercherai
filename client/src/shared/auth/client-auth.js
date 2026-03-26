import { computed, ref } from "vue";
import { clearClientSession, getClientSession, saveClientSession } from "./session";

const session = ref(getClientSession());
const isLoggedIn = computed(() => Boolean(session.value?.accessToken));

export function useClientAuth() {
  return {
    session,
    isLoggedIn
  };
}

export function syncClientAuthSession() {
  session.value = getClientSession();
  return session.value;
}

export function setClientAuthSession(payload) {
  session.value = saveClientSession(payload);
  return session.value;
}

export function clearClientAuthSession() {
  clearClientSession();
  session.value = null;
}
