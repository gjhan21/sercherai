<template>
  <section class="auth-page fade-up">
    <article class="auth-focus card finance-dual-rail finance-section-card">
      <div class="auth-focus-copy">
        <div class="finance-pill-row">
          <span class="finance-pill finance-pill-compact finance-pill-neutral">身份入口</span>
          <span class="finance-pill finance-pill-compact finance-pill-info">登录 / 注册双栏</span>
          <span class="finance-pill finance-pill-compact finance-pill-info">来源回跳可见</span>
        </div>
        <p class="auth-kicker">身份验证</p>
        <h1 class="section-title">登录或注册后继续当前操作</h1>
        <p class="section-subtitle">
          支持登录、注册、邀请码绑定和返回原页面。
        </p>
        <p v-if="redirectContextTip" class="auth-redirect-tip">{{ redirectContextTip }}</p>
        <div class="auth-focus-tags">
          <span class="finance-pill finance-pill-compact finance-pill-neutral">手机号 / 邮箱登录</span>
          <span class="finance-pill finance-pill-compact finance-pill-neutral">邀请码注册</span>
          <span class="finance-pill finance-pill-compact finance-pill-neutral">仅限未登录访问</span>
          <span class="finance-pill finance-pill-compact finance-pill-neutral">登录后返回原页面</span>
        </div>
        <div class="auth-hero-stats finance-hero-stat-grid">
          <article class="finance-hero-stat-card">
            <span>当前来源</span>
            <strong>{{ redirectDestinationLabel }}</strong>
            <p>说明当前动作和阅读链或会员动作直接相关，不是随机跳转。</p>
          </article>
          <article class="finance-hero-stat-card">
            <span>登录后回跳</span>
            <strong>{{ redirectPath }}</strong>
            <p>完成身份进入后优先回原页面，而不是强制送回首页。</p>
          </article>
          <article class="finance-hero-stat-card">
            <span>邀请码</span>
            <strong>{{ inviteCode || "未携带邀请码" }}</strong>
            <p>注册时继续保留 invite_code，可直接完成邀请绑定。</p>
          </article>
          <article class="finance-hero-stat-card">
            <span>当前支持</span>
            <strong>账号密码 / 邀请注册</strong>
            <p>保持现有账户体系，不扩短信和第三方登录入口。</p>
          </article>
        </div>
      </div>

      <aside class="auth-focus-side finance-stack-tight">
        <article v-for="item in authGuideRows" :key="item.title" class="finance-card-surface">
          <strong>{{ item.title }}</strong>
          <p>{{ item.desc }}</p>
        </article>
      </aside>
    </article>

    <section class="auth-main-layout finance-dual-rail">
      <div class="auth-content-stack finance-stack-tight">
        <article class="card auth-entry-card finance-section-card">
          <header class="section-head">
            <div>
              <h2 class="section-title">登录与注册</h2>
              <p class="section-subtitle">桌面端并列展示，移动端上下排列，并保留邀请码和来源信息。</p>
            </div>
        </header>

        <div class="auth-dual-grid">
            <article class="auth-form-card finance-card-surface">
              <div class="auth-form-head">
                <div>
                  <p>登录</p>
                  <h3>手机号 / 邮箱 + 密码</h3>
                </div>
                <span class="auth-badge finance-pill finance-pill-roomy finance-pill-info">返回原页面</span>
              </div>

              <form class="auth-form" @submit.prevent="handleSubmit('login')">
                <label>
                  手机号 / 邮箱
                  <input
                    v-model.trim="loginForm.account"
                    type="text"
                    placeholder="请输入手机号或邮箱"
                    autocomplete="username"
                  />
                </label>
                <label>
                  密码
                  <input
                    v-model="loginForm.password"
                    type="password"
                    placeholder="请输入密码"
                    autocomplete="current-password"
                  />
                </label>

                <p v-if="loginErrorMessage" class="form-tip finance-note-strip finance-note-strip-danger">{{ loginErrorMessage }}</p>
                <p v-else-if="loginSuccessMessage" class="form-tip finance-note-strip finance-note-strip-success">{{ loginSuccessMessage }}</p>

                <button type="submit" class="submit-btn finance-primary-btn" :disabled="submittingMode === 'login'">
                  {{ submittingMode === "login" ? "登录中..." : "立即登录" }}
                </button>
              </form>
            </article>

            <article class="auth-form-card register-card finance-card-surface">
              <div class="auth-form-head">
                <div>
                  <p>注册</p>
                  <h3>支持邀请码注册并自动绑定邀请关系</h3>
                </div>
                <span class="auth-badge accent finance-pill finance-pill-roomy">invite_code</span>
              </div>

              <div class="invite-box finance-info-box">
                <p>当前邀请码</p>
                <strong>{{ inviteCode || "未携带邀请码" }}</strong>
                <span>
                  {{
                    inviteCode
                      ? "当前通过分享链接注册，提交后会自动绑定邀请关系。"
                      : "当前没有邀请码也可直接注册，只是不自动绑定邀请关系。"
                  }}
                </span>
              </div>

              <form class="auth-form" @submit.prevent="handleSubmit('register')">
                <label>
                  手机号 / 邮箱
                  <input
                    v-model.trim="registerForm.account"
                    type="text"
                    placeholder="请输入手机号或邮箱"
                    autocomplete="username"
                  />
                </label>
                <label>
                  密码
                  <input
                    v-model="registerForm.password"
                    type="password"
                    placeholder="请输入密码"
                    autocomplete="new-password"
                  />
                </label>

                <p v-if="registerErrorMessage" class="form-tip finance-note-strip finance-note-strip-danger">{{ registerErrorMessage }}</p>
                <p
                  v-else-if="registerSuccessMessage"
                  class="form-tip finance-note-strip finance-note-strip-success"
                >
                  {{ registerSuccessMessage }}
                </p>

                <button type="submit" class="submit-btn finance-primary-btn" :disabled="submittingMode === 'register'">
                  {{ submittingMode === "register" ? "注册中..." : "注册并登录" }}
                </button>
              </form>
            </article>
          </div>
        </article>

        <article class="card auth-block-card finance-section-card">
          <header class="section-head">
            <div>
              <h2 class="section-title">当前支持</h2>
              <p class="section-subtitle">当前支持登录、注册、邀请码绑定和原页返回。</p>
            </div>
          </header>

          <div class="auth-card-grid">
            <article v-for="item in authSupportRows" :key="item.title" class="auth-info-item finance-card-surface">
              <strong>{{ item.title }}</strong>
              <p>{{ item.desc }}</p>
            </article>
          </div>
        </article>

        <article class="card auth-block-card finance-section-card">
          <header class="section-head">
            <div>
              <h2 class="section-title">登录后返回位置</h2>
              <p class="section-subtitle">完成身份验证后优先返回原页面继续操作。</p>
            </div>
          </header>

          <div class="auth-card-grid auth-route-grid">
            <article
              v-for="item in authRedirectRows"
              :key="item.path"
              class="auth-route-item finance-card-surface"
              :class="{ active: item.path === redirectPath }"
            >
              <strong>{{ item.title }}</strong>
              <p>{{ item.desc }}</p>
            </article>
          </div>
        </article>
      </div>

      <aside class="auth-side-stack finance-stack-tight finance-sticky-side">
        <article class="card auth-side-card finance-section-card">
          <strong>当前来源</strong>
          <p>{{ redirectDestinationLabel }}</p>
          <span>登录后返回当前页面。</span>
        </article>

        <article class="card auth-side-card finance-section-card">
          <strong>会员服务</strong>
          <p>登录后可继续升级、续费、阅读研报和保存关注。</p>
        </article>

        <article class="card auth-side-card finance-section-card">
          <strong>支持范围</strong>
          <p>当前支持登录、注册、邀请码绑定和原页返回。</p>
        </article>
      </aside>
    </section>

    <article class="auth-footer-strip card finance-section-card">
      <div class="auth-footer-grid finance-card-grid finance-card-grid-3">
        <section class="auth-footer-col finance-card-surface">
          <h3>当前支持</h3>
          <p>支持登录、注册、邀请码绑定和返回原页面。</p>
        </section>
        <section class="auth-footer-col finance-card-surface">
          <h3>返回位置</h3>
          <p>当前来源是 {{ redirectDestinationLabel }}，完成身份验证后继续返回原页面。</p>
        </section>
        <section class="auth-footer-col finance-card-surface">
          <h3>使用说明</h3>
          <p>登录后可继续会员升级、资讯阅读和关注保存等操作。</p>
        </section>
      </div>
      <div class="auth-footer-meta">
        <span class="finance-pill finance-pill-compact finance-pill-neutral">当前支持：手机号 / 邮箱登录</span>
        <span class="finance-pill finance-pill-compact finance-pill-neutral">当前支持：邀请码注册</span>
        <span class="finance-pill finance-pill-compact finance-pill-neutral">当前支持：未登录访问与原页返回</span>
      </div>
    </article>
  </section>
</template>

<script setup>
import { computed, reactive, ref } from "vue";
import { useRoute, useRouter } from "vue-router";
import { login, register } from "../api/auth";
import { setClientAuthSession } from "../lib/client-auth";

const route = useRoute();
const router = useRouter();

const submittingMode = ref("");
const loginErrorMessage = ref("");
const loginSuccessMessage = ref("");
const registerErrorMessage = ref("");
const registerSuccessMessage = ref("");

const loginForm = reactive({
  account: "",
  password: ""
});

const registerForm = reactive({
  account: "",
  password: ""
});

const inviteCode = computed(() =>
  String(route.query.invite_code || route.query.code || route.query.invite || "").trim().toUpperCase()
);
const redirectPath = computed(() => resolveRedirectPath());
const redirectDestinationLabel = computed(() => mapRedirectLabel(redirectPath.value));
const redirectContextTip = computed(() => {
  if (redirectPath.value === "/membership") {
    return "登录后会直接进入会员中心，继续完成升级、续费或激活确认。";
  }
  if (redirectPath.value === "/news") {
    return "登录后会回到资讯页，继续查看摘要、正文和附件阅读链。";
  }
  if (redirectPath.value === "/archive") {
    return "登录后会回到历史档案页，继续比较样本、版本和兑现情况。";
  }
  if (redirectPath.value === "/watchlist") {
    return "登录后会回到我的关注，继续保存标的并查看变化。";
  }
  if (redirectPath.value === "/strategies") {
    return "登录后会回到策略页，继续看主推荐、理由和风险边界。";
  }
  return "登录后会按当前来源回跳，继续原来的阅读或操作。";
});
const authGuideRows = computed(() => [
  {
    title: "快速完成登录或注册",
    desc: "表单保持简洁，完成后直接返回当前操作。"
  },
  {
    title: "邀请码直接可见",
    desc: "有邀请码时可直接在注册流程中完成绑定。"
  },
  {
    title: "完成后返回原页面",
    desc: "登录或注册成功后，会优先返回当前来源页面。"
  }
]);
const authSupportRows = computed(() => [
  {
    title: "手机号 / 邮箱注册与密码登录",
    desc: "支持手机号或邮箱注册，以及密码登录。"
  },
  {
    title: "邀请码注册自动绑定邀请关系",
    desc: "注册时会识别链接中的邀请码，并自动绑定邀请关系。"
  },
  {
    title: "登录状态自动保持",
    desc: "登录成功后会保持当前会话状态。"
  },
  {
    title: "未登录访问与原页返回",
    desc: "已登录用户不会重复进入认证页，未登录用户完成认证后会返回原页面。"
  }
]);
const authRedirectRows = computed(() => [
  {
    path: "/membership",
    title: "从会员页来",
    desc: "登录后直接回会员中心，继续升级、续费或核对激活状态。"
  },
  {
    path: "/news",
    title: "从资讯页来",
    desc: "登录后回到资讯正文，继续阅读摘要、全文和附件。"
  },
  {
    path: "/archive",
    title: "从历史档案页来",
    desc: "登录后回到原档案位置，继续比较历史样本和兑现表现。"
  },
  {
    path: "/watchlist",
    title: "从关注页来",
    desc: "登录后回到关注页，继续保存标的或查看变化回访结果。"
  },
  {
    path: "/strategies",
    title: "从策略页来",
    desc: "登录后回到策略页，继续阅读主推荐、理由和风险边界。"
  },
  {
    path: "/home",
    title: "从首页或其他入口来",
    desc: "没有明确来源时回到首页，继续进入推荐、资讯或会员节奏。"
  }
]);

function mapRedirectLabel(path) {
  if (path === "/membership") return "会员中心";
  if (path === "/news") return "资讯页";
  if (path === "/archive") return "历史档案页";
  if (path === "/watchlist") return "我的关注";
  if (path === "/strategies") return "策略页";
  return "首页或默认入口";
}

function normalizeAccount(account) {
  return String(account || "").replace(/\s+/g, "").trim();
}

function validateForm(form) {
  const account = normalizeAccount(form.account);
  if (!account) {
    throw new Error("请输入手机号或邮箱");
  }
  if ((form.password || "").length < 8) {
    throw new Error("密码长度至少为8位");
  }
  return account;
}

function buildRegisterPayload(account) {
  const payload = {
    password: registerForm.password
  };
  if (account.includes("@")) {
    payload.email = account.toLowerCase();
  } else {
    payload.phone = account;
  }
  if (inviteCode.value) {
    payload.invite_code = inviteCode.value;
  }
  return payload;
}

function resolveRedirectPath() {
  const redirect = String(route.query.redirect || "");
  if (redirect.startsWith("/")) {
    return redirect;
  }
  return "/home";
}

function resetFeedback(mode) {
  if (mode === "login") {
    loginErrorMessage.value = "";
    loginSuccessMessage.value = "";
    return;
  }
  registerErrorMessage.value = "";
  registerSuccessMessage.value = "";
}

async function handleSubmit(mode) {
  resetFeedback(mode);
  submittingMode.value = mode;
  try {
    const form = mode === "login" ? loginForm : registerForm;
    const account = validateForm(form);
    let payload = null;
    if (mode === "login") {
      payload = await login({
        account,
        password: form.password
      });
      loginSuccessMessage.value = "登录成功，正在返回原页面...";
    } else {
      payload = await register(buildRegisterPayload(account));
      registerSuccessMessage.value = "注册成功，正在返回原页面...";
    }
    setClientAuthSession({
      ...payload,
      phone: account.includes("@") ? "" : account,
      email: account.includes("@") ? account.toLowerCase() : ""
    });
    await router.replace(resolveRedirectPath());
  } catch (error) {
    const message = error?.message || "提交失败";
    if (mode === "login") {
      loginErrorMessage.value = message;
    } else {
      registerErrorMessage.value = message;
    }
  } finally {
    submittingMode.value = "";
  }
}
</script>

<style scoped>
.auth-page {
  display: grid;
  gap: 12px;
  padding: 20px 14px 28px;
}

.auth-page > * {
  width: min(1180px, 100%);
  margin: 0 auto;
}

.auth-focus {
  --finance-main-column: minmax(0, 1.7fr);
  border-radius: 22px;
}

.auth-focus-copy {
  border-radius: 16px;
  padding: 14px;
  background:
    radial-gradient(circle at 100% 0%, rgba(36, 87, 167, 0.15) 0%, transparent 34%),
    var(--color-surface-panel-soft-subtle);
  border: 1px solid rgba(214, 222, 234, 0.85);
}

.auth-kicker {
  margin: 0;
  font-size: 12px;
  color: var(--color-pine-600);
}

.auth-focus-copy h1 {
  margin-top: 8px;
}

.auth-redirect-tip {
  margin: 10px 0 0;
  padding: 10px 12px;
  border-radius: 12px;
  border: 1px solid var(--color-focus-glow);
  background: var(--color-focus-fill);
  color: var(--color-pine-700);
  font-size: 13px;
  line-height: 1.6;
}

.auth-focus-tags {
  margin-top: 12px;
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.auth-hero-stats {
  margin-top: 12px;
}

.auth-badge.accent {
  color: var(--color-accent);
  background: var(--color-surface-gold-strong);
  border-color: var(--color-line-gold);
}

.auth-focus-side strong,
.auth-focus-side p,
.auth-side-card strong,
.auth-side-card p,
.auth-side-card span {
  margin: 0;
}

.auth-focus-side strong,
.auth-side-card strong {
  font-size: 16px;
  line-height: 1.45;
  color: var(--color-pine-700);
}

.auth-focus-side p,
.auth-side-card p,
.auth-side-card span {
  margin-top: 6px;
  font-size: 13px;
  line-height: 1.65;
  color: var(--color-text-sub);
}

.auth-dual-grid,
.auth-card-grid {
  margin-top: 12px;
  display: grid;
  gap: 12px;
  grid-template-columns: repeat(2, minmax(0, 1fr));
}

.auth-route-grid {
  gap: 10px;
}

.auth-form-card,
.auth-info-item,
.auth-route-item {
  min-width: 0;
}

.auth-form-card.register-card {
  background: var(--gradient-card-active);
}

.auth-form-head {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 10px;
}

.auth-form-head p,
.auth-form-head h3,
.auth-info-item strong,
.auth-info-item p,
.auth-route-item strong,
.auth-route-item p {
  margin: 0;
}

.auth-form-head p {
  font-size: 12px;
  color: var(--color-text-sub);
}

.auth-form-head h3 {
  margin-top: 4px;
  font-size: 18px;
  line-height: 1.4;
  color: var(--color-text-main);
}

.auth-form {
  display: grid;
  gap: 10px;
}

.auth-form label {
  display: grid;
  gap: 5px;
  font-size: 13px;
  color: var(--color-text-sub);
}

.auth-form input {
  border: 1px solid var(--color-border-soft-heavy);
  border-radius: 10px;
  padding: 9px 10px;
  background: var(--color-surface-card-elevated);
  color: var(--color-text-main);
}

.auth-form input:focus {
  outline: none;
  border-color: var(--color-pine-500);
  box-shadow: 0 0 0 3px rgba(36, 87, 167, 0.15);
}

.submit-btn {
  width: 100%;
}

.submit-btn:disabled {
  opacity: 0.7;
}

.quick-login {
  display: grid;
  gap: 6px;
}

.quick-login p {
  margin: 0;
  font-size: 12px;
  color: var(--color-text-sub);
}

.quick-login button {
  border-style: dashed;
  width: 100%;
  text-align: left;
}

.auth-info-item,
.auth-route-item {
  display: grid;
  gap: 6px;
}

.auth-info-item strong,
.auth-route-item strong {
  font-size: 16px;
  line-height: 1.45;
  color: var(--color-pine-700);
}

.auth-info-item p,
.auth-route-item p {
  font-size: 13px;
  line-height: 1.65;
  color: var(--color-text-sub);
}

.auth-route-item.active {
  border-color: var(--color-border-focus);
  background: var(--gradient-card-active);
  box-shadow: var(--shadow-card-active);
}

.auth-footer-col {
  min-width: 0;
}

.auth-footer-col h3,
.auth-footer-col p {
  margin: 0;
}

.auth-footer-col h3 {
  font-size: 16px;
  line-height: 1.45;
  color: var(--color-pine-700);
}

.auth-footer-col p {
  margin-top: 6px;
  font-size: 13px;
  line-height: 1.65;
  color: var(--color-text-sub);
}

.auth-footer-meta {
  margin-top: 12px;
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

@media (max-width: 980px) {
  .auth-focus,
  .auth-main-layout,
  .auth-dual-grid,
  .auth-card-grid,
  .auth-footer-grid {
    grid-template-columns: 1fr;
  }
}

@media (max-width: 760px) {
  .auth-page {
    padding: 16px 10px 24px;
  }

  .auth-focus,
  .auth-entry-card,
  .auth-block-card,
  .auth-side-card {
    border-radius: 16px;
  }

  .auth-card-grid {
    grid-template-columns: 1fr;
  }

  .auth-form-head {
    flex-direction: column;
  }

  .submit-btn,
  .quick-login button {
    width: 100%;
  }
}
</style>
