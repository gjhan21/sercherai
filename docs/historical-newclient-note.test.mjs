import test from "node:test";
import assert from "node:assert/strict";
import fs from "node:fs";

const note = [
  "> Historical note:",
  "> This document describes work from the legacy `client` frontend era.",
  "> The current official frontend has moved to `/Users/gjhan21/cursor/sercherai/newclient`."
].join("\n");

const files = [
  "/Users/gjhan21/cursor/sercherai/.worktrees/newclient-formal-cutover-2/docs/superpowers/plans/2026-03-23-community-discussion-mvp.md",
  "/Users/gjhan21/cursor/sercherai/.worktrees/newclient-formal-cutover-2/docs/superpowers/plans/2026-03-24-client-pc-h5-demo-alignment.md",
  "/Users/gjhan21/cursor/sercherai/.worktrees/newclient-formal-cutover-2/docs/superpowers/plans/2026-03-24-newclient-strategies-demo.md",
  "/Users/gjhan21/cursor/sercherai/.worktrees/newclient-formal-cutover-2/docs/superpowers/plans/2026-03-24-sercherai-h5-app-refactor.md",
  "/Users/gjhan21/cursor/sercherai/.worktrees/newclient-formal-cutover-2/docs/superpowers/plans/2026-03-27-h5-community-watchlist-restructure-plan.md",
  "/Users/gjhan21/cursor/sercherai/.worktrees/newclient-formal-cutover-2/docs/superpowers/plans/2026-03-27-pc-community-watchlist-restructure-plan.md",
  "/Users/gjhan21/cursor/sercherai/.worktrees/newclient-formal-cutover-2/docs/superpowers/plans/2026-03-28-stock-futures-forecast-l1-implementation.md",
  "/Users/gjhan21/cursor/sercherai/.worktrees/newclient-formal-cutover-2/docs/superpowers/plans/2026-03-29-stock-futures-forecast-l2-implementation.md",
  "/Users/gjhan21/cursor/sercherai/.worktrees/newclient-formal-cutover-2/docs/superpowers/plans/2026-03-29-stock-futures-forecast-l3-implementation.md",
  "/Users/gjhan21/cursor/sercherai/.worktrees/newclient-formal-cutover-2/docs/superpowers/specs/2026-03-23-community-discussion-design.md",
  "/Users/gjhan21/cursor/sercherai/.worktrees/newclient-formal-cutover-2/docs/superpowers/specs/2026-03-24-client-pc-h5-demo-alignment-design.md",
  "/Users/gjhan21/cursor/sercherai/.worktrees/newclient-formal-cutover-2/docs/superpowers/specs/2026-03-24-newclient-fullsite-demo-design.md",
  "/Users/gjhan21/cursor/sercherai/.worktrees/newclient-formal-cutover-2/docs/superpowers/specs/2026-03-24-newclient-homepage-demo-design.md",
  "/Users/gjhan21/cursor/sercherai/.worktrees/newclient-formal-cutover-2/docs/superpowers/specs/2026-03-24-newclient-profile-demo-design.md",
  "/Users/gjhan21/cursor/sercherai/.worktrees/newclient-formal-cutover-2/docs/superpowers/specs/2026-03-24-newclient-strategies-demo-design.md",
  "/Users/gjhan21/cursor/sercherai/.worktrees/newclient-formal-cutover-2/docs/superpowers/specs/2026-03-24-newclient-watchlist-demo-design.md",
  "/Users/gjhan21/cursor/sercherai/.worktrees/newclient-formal-cutover-2/docs/superpowers/specs/2026-03-24-sercherai-h5-xueqiu-design.md",
  "/Users/gjhan21/cursor/sercherai/.worktrees/newclient-formal-cutover-2/docs/superpowers/specs/2026-03-26-client-h5demo-style-alignment-design.md",
  "/Users/gjhan21/cursor/sercherai/.worktrees/newclient-formal-cutover-2/docs/superpowers/specs/2026-03-27-h5-community-watchlist-restructure-design.md",
  "/Users/gjhan21/cursor/sercherai/.worktrees/newclient-formal-cutover-2/docs/superpowers/specs/2026-03-27-pc-community-watchlist-restructure-design.md",
  "/Users/gjhan21/cursor/sercherai/.worktrees/newclient-formal-cutover-2/docs/superpowers/specs/2026-03-28-stock-futures-forecast-roadmap.md",
  "/Users/gjhan21/cursor/sercherai/.worktrees/newclient-formal-cutover-2/docs/superpowers/specs/2026-03-28-stock-futures-forecast-thread-handoff.md",
  "/Users/gjhan21/cursor/sercherai/.worktrees/newclient-formal-cutover-2/docs/vibe-stock-growth/README.md",
  "/Users/gjhan21/cursor/sercherai/.worktrees/newclient-formal-cutover-2/docs/vibe-stock-growth/阶段0-画布与真相源.md",
  "/Users/gjhan21/cursor/sercherai/.worktrees/newclient-formal-cutover-2/docs/vibe-stock-growth/阶段1-今日决策首页.md",
  "/Users/gjhan21/cursor/sercherai/.worktrees/newclient-formal-cutover-2/docs/vibe-stock-growth/阶段2-推荐档案页.md",
  "/Users/gjhan21/cursor/sercherai/.worktrees/newclient-formal-cutover-2/docs/vibe-stock-growth/阶段3-历史档案与信任改造.md",
  "/Users/gjhan21/cursor/sercherai/.worktrees/newclient-formal-cutover-2/docs/vibe-stock-growth/阶段4-我的关注与回访机制.md",
  "/Users/gjhan21/cursor/sercherai/.worktrees/newclient-formal-cutover-2/docs/vibe-stock-growth/阶段6-会员转化与内容节奏.md"
];

test("historical docs with old client references carry a cutover note", () => {
  for (const file of files) {
    const text = fs.readFileSync(file, "utf8");
    assert.match(text, /client/);
    assert.match(text, new RegExp(note.replace(/[.*+?^${}()|[\]\\]/g, "\\$&")));
  }
});
