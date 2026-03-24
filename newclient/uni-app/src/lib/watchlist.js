export const H5_WATCHLIST_KEY = "sercherai_newclient_h5_watchlist";

function getStorage() {
  if (typeof globalThis.localStorage !== "undefined" && globalThis.localStorage) {
    return globalThis.localStorage;
  }
  return {
    getItem() {
      return null;
    },
    setItem() {},
    removeItem() {}
  };
}

function parseItems(raw) {
  if (!raw) {
    return [];
  }
  try {
    const items = JSON.parse(raw);
    return Array.isArray(items) ? items : [];
  } catch {
    return [];
  }
}

function saveItems(items) {
  getStorage().setItem(H5_WATCHLIST_KEY, JSON.stringify(items));
  return items;
}

export function getWatchlistItems() {
  return parseItems(getStorage().getItem(H5_WATCHLIST_KEY));
}

export function isWatched(id) {
  return getWatchlistItems().some((item) => item.id === id);
}

export function addWatchItem(item) {
  const items = getWatchlistItems();
  if (items.some((entry) => entry.id === item.id)) {
    return items;
  }
  return saveItems([...items, item]);
}

export function removeWatchItem(id) {
  return saveItems(getWatchlistItems().filter((item) => item.id !== id));
}

export function toggleWatchItem(item) {
  if (isWatched(item.id)) {
    return removeWatchItem(item.id);
  }
  return addWatchItem(item);
}
