import test from "node:test";
import assert from "node:assert/strict";

test("resolveClientOrigin prefers explicit env origin", async () => {
  const { resolveClientOrigin } = await import(`./community-links.js?test=${Date.now()}-env`);

  assert.equal(
    resolveClientOrigin({
      envOrigin: " https://client.example.com/ ",
      locationOrigin: "http://127.0.0.1:5174"
    }),
    "https://client.example.com"
  );
});

test("resolveClientOrigin falls back to local client port in dev", async () => {
  const { resolveClientOrigin } = await import(`./community-links.js?test=${Date.now()}-local`);

  assert.equal(
    resolveClientOrigin({
      envOrigin: "",
      locationOrigin: "http://127.0.0.1:5174"
    }),
    "http://127.0.0.1:5175"
  );
});

test("buildCommunityTopicURL appends topic path and optional comment anchor", async () => {
  const { buildCommunityTopicURL } = await import(`./community-links.js?test=${Date.now()}-topic`);

  assert.equal(
    buildCommunityTopicURL({
      topicID: "ct_demo_001",
      commentID: "cc_demo_002",
      envOrigin: "",
      locationOrigin: "http://localhost:5174"
    }),
    "http://localhost:5175/community/topics/ct_demo_001#comment-cc_demo_002"
  );
});
