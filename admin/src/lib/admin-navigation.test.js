import test from "node:test";
import assert from "node:assert/strict";

function mockLocalStorage(session) {
  globalThis.localStorage = {
    getItem(key) {
      if (key !== "sercherai_admin_session") {
        return null;
      }
      return JSON.stringify(session);
    },
    setItem() {},
    removeItem() {}
  };
}

test("admin navigation keeps the research workbench menus in the expected order", async () => {
  mockLocalStorage({
    accessToken: "test-token",
    permissionCodes: ["*"]
  });

  const { adminNavigationItems } = await import(`./admin-navigation.js?test=${Date.now()}`);
  const orderedNames = adminNavigationItems.map((item) => item.name);

  assert.deepEqual(
    orderedNames.filter((name) =>
      [
        "market-center",
        "stock-selection",
        "futures-selection",
        "data-sources",
        "review-center",
        "system-jobs"
      ].includes(name)
    ),
    [
      "market-center",
      "stock-selection",
      "futures-selection",
      "data-sources",
      "review-center",
      "system-jobs"
    ]
  );
});

test("resolveFirstAccessibleRoute returns the first visible route for the current session", async () => {
  mockLocalStorage({
    accessToken: "test-token",
    permissionCodes: ["dashboard.view"]
  });

  const { resolveFirstAccessibleRoute } = await import(`./admin-navigation.js?test=${Date.now()}-first`);

  assert.equal(resolveFirstAccessibleRoute(), "/dashboard");
});

test("community moderation menu is visible when community.view permission exists", async () => {
  mockLocalStorage({
    accessToken: "test-token",
    permissionCodes: ["community.view"]
  });

  const { getVisibleAdminNavigationItems, resolveFirstAccessibleRoute } = await import(
    `./admin-navigation.js?test=${Date.now()}-community`
  );

  const visibleItems = getVisibleAdminNavigationItems();
  assert.ok(visibleItems.some((item) => item.name === "community"));
  assert.equal(resolveFirstAccessibleRoute(), "/community");
});
