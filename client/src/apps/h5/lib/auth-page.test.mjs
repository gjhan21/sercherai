import test from "node:test";
import assert from "node:assert/strict";
import {
  buildAuthSurfaceModel,
  describeAuthScene,
  normalizeAuthRedirect,
  normalizeInviteCode,
  resolveAuthBackTarget,
  resolveAuthInitialMode
} from "./auth-page.js";

test("normalizeInviteCode reads invite aliases and trims output", () => {
  assert.equal(normalizeInviteCode({ invite_code: " abC123 " }), "ABC123");
  assert.equal(normalizeInviteCode({ code: " zx9 " }), "ZX9");
  assert.equal(normalizeInviteCode({ invite: " hello " }), "HELLO");
  assert.equal(normalizeInviteCode({}), "");
});

test("resolveAuthInitialMode switches to register when invite code exists", () => {
  assert.equal(resolveAuthInitialMode({ invite_code: "ABC123" }), "register");
  assert.equal(resolveAuthInitialMode({}), "login");
});

test("normalizeAuthRedirect keeps valid h5 paths and rewrites legacy paths", () => {
  assert.equal(normalizeAuthRedirect("/membership"), "/membership");
  assert.equal(normalizeAuthRedirect("/news?article=n1"), "/news?article=n1");
  assert.equal(normalizeAuthRedirect("/archive?article=n2"), "/archive?article=n2");
  assert.equal(normalizeAuthRedirect("/profile?section=watchlist"), "/profile?section=watchlist");
  assert.equal(normalizeAuthRedirect("/watchlist"), "/profile?section=watchlist");
  assert.equal(normalizeAuthRedirect("https://example.com"), "/home");
  assert.equal(normalizeAuthRedirect(""), "/home");
});

test("describeAuthScene maps redirect path to h5 scene copy", () => {
  const membership = describeAuthScene("/membership");
  assert.equal(membership.label, "会员中心");
  assert.match(membership.tip, /升级|续费|激活/);

  const news = describeAuthScene("/news?article=n1");
  assert.equal(news.label, "资讯正文");
  assert.match(news.tip, /正文|附件/);

  const strategies = describeAuthScene("/strategies?id=s1");
  assert.equal(strategies.label, "策略详情");
  assert.match(strategies.tip, /推荐|风险边界/);

  const watchlist = describeAuthScene("/profile?section=watchlist");
  assert.equal(watchlist.label, "我的 > 我的关注");
  assert.match(watchlist.tip, /个人中心|关注|变化/);

  const archive = describeAuthScene("/archive");
  assert.equal(archive.label, "历史档案");
  assert.match(archive.tip, /历史|复盘|样本/);

  const home = describeAuthScene("/unknown");
  assert.equal(home.label, "首页");
});

test("buildAuthSurfaceModel creates app-style register copy when invite code exists", () => {
  const model = buildAuthSurfaceModel({
    redirectPath: "/news?article=n1",
    inviteCode: "VIP2026",
    mode: "register"
  });

  assert.equal(model.heroKicker, "研究决策会员入口");
  assert.equal(model.sceneLabel, "资讯正文");
  assert.equal(model.primaryActionLabel, "注册并继续阅读");
  assert.equal(model.submitHint, "完成注册后将自动绑定邀请码，并返回当前资讯场景。"
  );
  assert.equal(model.cards[0].title, "认证后返回");
  assert.match(model.cards[2].desc, /VIP2026/);
});

test("buildAuthSurfaceModel creates membership login copy without invite", () => {
  const model = buildAuthSurfaceModel({
    redirectPath: "/membership",
    inviteCode: "",
    mode: "login"
  });

  assert.equal(model.primaryActionLabel, "登录并继续会员服务");
  assert.equal(model.agreementLabel, "登录即代表你已阅读并同意《用户服务协议》与《隐私说明》。");
  assert.equal(model.cards[1].title, "登录后可继续");
  assert.match(model.cards[1].desc, /升级|续费|激活/);
});

test("resolveAuthBackTarget avoids redirect loops for auth-only destinations", () => {
  assert.equal(resolveAuthBackTarget("/membership"), "/home");
  assert.equal(resolveAuthBackTarget("/profile"), "/home");
  assert.equal(resolveAuthBackTarget("/news?article=n1"), "/news?article=n1");
  assert.equal(resolveAuthBackTarget("/profile?section=watchlist"), "/home");
  assert.equal(resolveAuthBackTarget("/archive?article=n1"), "/archive?article=n1");
});
