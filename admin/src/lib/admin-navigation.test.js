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
        "forecast-lab",
        "stock-selection",
        "futures-selection",
        "data-sources",
        "review-center",
        "system-jobs"
      ].includes(name)
    ),
    [
      "market-center",
      "forecast-lab",
      "stock-selection",
      "futures-selection",
      "data-sources",
      "review-center",
      "system-jobs"
    ]
  );
});

test("data-sources navigation entry points to the governance child route", async () => {
  mockLocalStorage({
    accessToken: "test-token",
    permissionCodes: ["*"]
  });

  const { adminNavigationItems } = await import(`./admin-navigation.js?test=${Date.now()}-data-sources`);
  const item = adminNavigationItems.find((entry) => entry.name === "data-sources");

  assert.equal(item?.to, "/data-sources/governance");
});

test("forecast lab navigation entry points to the dedicated L3 workbench route", async () => {
  mockLocalStorage({
    accessToken: "test-token",
    permissionCodes: ["forecast_l3.view"]
  });

  const { adminNavigationItems, resolveFirstAccessibleRoute } = await import(
    `./admin-navigation.js?test=${Date.now()}-forecast-lab`
  );
  const item = adminNavigationItems.find((entry) => entry.name === "forecast-lab");

  assert.equal(item?.to, "/forecast-lab");
  assert.equal(resolveFirstAccessibleRoute(), "/forecast-lab");
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
