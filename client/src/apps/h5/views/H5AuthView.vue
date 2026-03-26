<template>
  <div class="h5-auth-page fade-up">
    <main class="h5-auth-wrap">
      <header class="h5-auth-header">
        <button type="button" class="h5-auth-back" @click="handleBack">
          <span class="h5-auth-back-icon" aria-hidden="true">
            <svg viewBox="0 0 24 24" role="img">
              <path d="M14.75 5.75 8.5 12l6.25 6.25" />
            </svg>
          </span>
          <span>返回</span>
        </button>
        <div class="h5-auth-brand">
          <span class="h5-auth-brand-mark">S</span>
          <div>
            <strong>SercherAI</strong>
            <small>{{ surfaceModel.heroKicker }}</small>
          </div>
        </div>
      </header>

      <section class="h5-auth-hero-card">
        <div class="h5-auth-scene-row">
          <span class="h5-auth-scene-pill">{{ surfaceModel.sceneLabel }}</span>
          <span v-if="inviteCode" class="h5-auth-scene-pill soft">已识别邀请码</span>
        </div>

        <div class="h5-auth-hero-copy">
          <h1>{{ surfaceModel.heroTitle }}</h1>
          <p>{{ surfaceModel.heroTip }}</p>
        </div>

        <div class="h5-auth-highlight-row">
          <span v-for="item in surfaceModel.highlights" :key="item" class="h5-auth-highlight">{{ item }}</span>
        </div>

        <div v-if="inviteCode" class="h5-auth-invite-banner">
          <span>邀请码</span>
          <strong>{{ inviteCode }}</strong>
          <p>完成注册后将自动绑定邀请关系。</p>
        </div>
      </section>

      <section class="h5-auth-main-card">
        <div class="h5-auth-tabs">
          <button type="button" class="h5-auth-tab" :class="{ active: currentMode === 'login' }" @click="currentMode = 'login'">
            登录
          </button>
          <button type="button" class="h5-auth-tab" :class="{ active: currentMode === 'register' }" @click="currentMode = 'register'">
            注册
          </button>
        </div>

        <div class="h5-auth-panel-copy">
          <strong>{{ currentMode === 'login' ? '登录后继续当前服务' : '注册后自动进入当前场景' }}</strong>
          <p>{{ surfaceModel.submitHint }}</p>
        </div>

        <form v-if="currentMode === 'login'" class="h5-auth-form" @submit.prevent="handleSubmit('login')">
          <label class="h5-auth-field">
            <span>手机号 / 邮箱</span>
            <input
              v-model.trim="loginForm.account"
              type="text"
              placeholder="请输入手机号或邮箱"
              autocomplete="username"
            />
          </label>
          <label class="h5-auth-field">
            <span>密码</span>
            <input
              v-model="loginForm.password"
              type="password"
              placeholder="请输入登录密码"
              autocomplete="current-password"
            />
          </label>
          <p v-if="loginErrorMessage" class="h5-auth-feedback error">{{ loginErrorMessage }}</p>
          <p v-else-if="loginSuccessMessage" class="h5-auth-feedback success">{{ loginSuccessMessage }}</p>
          <button type="submit" class="h5-btn block" :disabled="submittingMode === 'login'">
            {{ submittingMode === 'login' ? '登录中...' : surfaceModel.primaryActionLabel }}
          </button>
          <p class="h5-auth-agreement">{{ surfaceModel.agreementLabel }}</p>
        </form>

        <form v-else class="h5-auth-form" @submit.prevent="handleSubmit('register')">
          <label class="h5-auth-field">
            <span>手机号 / 邮箱</span>
            <input
              v-model.trim="registerForm.account"
              type="text"
              placeholder="请输入手机号或邮箱"
              autocomplete="username"
            />
          </label>
          <label class="h5-auth-field">
            <span>密码</span>
            <input
              v-model="registerForm.password"
              type="password"
              placeholder="请设置不少于 8 位密码"
              autocomplete="new-password"
            />
          </label>
          <div class="h5-auth-invite-card">
            <span>邀请码状态</span>
            <strong>{{ inviteCode || '当前未携带邀请码' }}</strong>
            <p>{{ inviteCode ? '注册完成后会自动绑定邀请关系。' : '没有邀请码也可直接注册，后续继续正常使用。' }}</p>
          </div>
          <p v-if="registerErrorMessage" class="h5-auth-feedback error">{{ registerErrorMessage }}</p>
          <p v-else-if="registerSuccessMessage" class="h5-auth-feedback success">{{ registerSuccessMessage }}</p>
          <button type="submit" class="h5-btn block" :disabled="submittingMode === 'register'">
            {{ submittingMode === 'register' ? '注册中...' : surfaceModel.primaryActionLabel }}
          </button>
          <p class="h5-auth-agreement">{{ surfaceModel.agreementLabel }}</p>
        </form>
      </section>

      <section class="h5-auth-scene-cards">
        <article v-for="item in surfaceModel.cards" :key="item.title" class="h5-auth-scene-card">
          <span>{{ item.title }}</span>
          <strong>{{ item.desc }}</strong>
        </article>
      </section>

      <footer class="h5-auth-footer">
        <p>仅支持手机号 / 邮箱 + 密码认证，成功后将按当前来源自动返回。</p>
      </footer>
    </main>
  </div>
</template>

<script setup>
import { computed, reactive, ref } from "vue";
import { useRoute, useRouter } from "vue-router";
import { login, register } from "../../../shared/api/auth";
import { setClientAuthSession } from "../../../shared/auth/client-auth";
import {
  buildAuthSurfaceModel,
  normalizeAuthRedirect,
  normalizeInviteCode,
  resolveAuthBackTarget,
  resolveAuthInitialMode
} from "../lib/auth-page";

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

const inviteCode = computed(() => normalizeInviteCode(route.query));
const redirectPath = computed(() => normalizeAuthRedirect(route.query.redirect));
const currentMode = ref(resolveAuthInitialMode(route.query));
const surfaceModel = computed(() => buildAuthSurfaceModel({
  redirectPath: redirectPath.value,
  inviteCode: inviteCode.value,
  mode: currentMode.value
}));

function normalizeAccount(account) {
  return String(account || "").replace(/\s+/g, "").trim();
}

function validateForm(form) {
  const account = normalizeAccount(form.account);
  if (!account) {
    throw new Error("请输入手机号或邮箱");
  }
  if (String(form.password || "").length < 8) {
    throw new Error("密码长度至少为 8 位");
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
      loginSuccessMessage.value = "登录成功，正在返回当前场景...";
    } else {
      payload = await register(buildRegisterPayload(account));
      registerSuccessMessage.value = "注册成功，正在返回当前场景...";
    }
    setClientAuthSession({
      ...payload,
      phone: account.includes("@") ? "" : account,
      email: account.includes("@") ? account.toLowerCase() : ""
    });
    await router.replace(redirectPath.value);
  } catch (error) {
    const message = error?.message || "提交失败，请稍后重试";
    if (mode === "login") {
      loginErrorMessage.value = message;
    } else {
      registerErrorMessage.value = message;
    }
  } finally {
    submittingMode.value = "";
  }
}

async function handleBack() {
  await router.replace(resolveAuthBackTarget(route.query.redirect));
}
</script>

<style scoped>
.h5-auth-page {
  min-height: 100vh;
  padding:
    calc(16px + env(safe-area-inset-top, 0px))
    14px
    calc(26px + env(safe-area-inset-bottom, 0px));
  background:
    radial-gradient(circle at top right, rgba(215, 175, 91, 0.2), transparent 24%),
    linear-gradient(180deg, #eef3f9 0%, #f6f9fc 34%, #ffffff 100%);
}

.h5-auth-wrap {
  width: min(100%, 560px);
  margin: 0 auto;
  display: grid;
  gap: 14px;
}

.h5-auth-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}

.h5-auth-back {
  min-height: 40px;
  padding: 0 12px;
  border: 1px solid rgba(18, 52, 95, 0.1);
  border-radius: 999px;
  background: rgba(255, 255, 255, 0.92);
  color: var(--h5-brand);
  display: inline-flex;
  align-items: center;
  gap: 8px;
  font-size: 13px;
  font-weight: 700;
  cursor: pointer;
}

.h5-auth-back-icon {
  width: 16px;
  height: 16px;
  display: inline-flex;
}

.h5-auth-back-icon svg {
  width: 100%;
  height: 100%;
  fill: none;
  stroke: currentColor;
  stroke-linecap: round;
  stroke-linejoin: round;
  stroke-width: 2;
}

.h5-auth-brand {
  flex: 1;
  min-width: 0;
  display: inline-flex;
  align-items: center;
  gap: 10px;
}

.h5-auth-brand-mark {
  width: 38px;
  height: 38px;
  border-radius: 14px;
  display: grid;
  place-items: center;
  background: linear-gradient(180deg, #173a6e, #244f84);
  color: #fff;
  font-size: 18px;
  font-weight: 800;
  letter-spacing: 0.06em;
}

.h5-auth-brand strong,
.h5-auth-brand small,
.h5-auth-hero-copy h1,
.h5-auth-hero-copy p,
.h5-auth-invite-banner p,
.h5-auth-invite-banner strong,
.h5-auth-panel-copy strong,
.h5-auth-panel-copy p,
.h5-auth-field span,
.h5-auth-agreement,
.h5-auth-feedback,
.h5-auth-invite-card span,
.h5-auth-invite-card strong,
.h5-auth-invite-card p,
.h5-auth-scene-card span,
.h5-auth-scene-card strong,
.h5-auth-footer p {
  margin: 0;
}

.h5-auth-brand strong {
  display: block;
  color: var(--h5-text);
  font-size: 15px;
}

.h5-auth-brand small {
  display: block;
  color: var(--h5-text-soft);
  font-size: 11px;
}

.h5-auth-hero-card,
.h5-auth-main-card,
.h5-auth-scene-card {
  border: 1px solid var(--h5-line);
  border-radius: var(--h5-radius);
  box-shadow: var(--h5-shadow);
}

.h5-auth-hero-card {
  padding: 20px 18px;
  display: grid;
  gap: 16px;
  background:
    radial-gradient(circle at top right, rgba(215, 175, 91, 0.2), transparent 28%),
    linear-gradient(160deg, #173a6e 0%, #21497d 54%, #f6f9fc 54%, #ffffff 100%);
}

.h5-auth-scene-row,
.h5-auth-highlight-row,
.h5-auth-scene-cards {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
}

.h5-auth-scene-pill,
.h5-auth-highlight {
  min-height: 30px;
  padding: 0 12px;
  border-radius: 999px;
  display: inline-flex;
  align-items: center;
  font-size: 12px;
  font-weight: 700;
}

.h5-auth-scene-pill {
  background: rgba(255, 255, 255, 0.92);
  color: #173a6e;
}

.h5-auth-scene-pill.soft {
  background: rgba(255, 232, 191, 0.92);
  color: #9b6614;
}

.h5-auth-hero-copy {
  display: grid;
  gap: 8px;
}

.h5-auth-hero-copy h1 {
  color: #16263d;
  font-size: clamp(28px, 7vw, 34px);
  line-height: 1.15;
  letter-spacing: -0.03em;
}

.h5-auth-hero-copy p,
.h5-auth-invite-banner p {
  color: #5f7088;
  font-size: 13px;
  line-height: 1.72;
}

.h5-auth-highlight {
  background: rgba(255, 255, 255, 0.14);
  color: #fff;
}

.h5-auth-invite-banner {
  padding: 14px;
  border-radius: 18px;
  background: rgba(255, 255, 255, 0.88);
  display: grid;
  gap: 4px;
}

.h5-auth-invite-banner span {
  color: #7a8798;
  font-size: 11px;
}

.h5-auth-invite-banner strong {
  color: #173a6e;
  font-size: 16px;
}

.h5-auth-main-card {
  padding: 16px;
  background: rgba(255, 255, 255, 0.98);
  display: grid;
  gap: 14px;
}

.h5-auth-tabs {
  padding: 4px;
  border-radius: 18px;
  background: rgba(19, 52, 95, 0.06);
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 6px;
}

.h5-auth-tab {
  min-height: 42px;
  border: 0;
  border-radius: 14px;
  background: transparent;
  color: #5f7088;
  font-size: 14px;
  font-weight: 700;
  cursor: pointer;
}

.h5-auth-tab.active {
  background: linear-gradient(180deg, #173a6e, #22497b);
  color: #fff;
  box-shadow: 0 10px 20px rgba(16, 42, 86, 0.16);
}

.h5-auth-panel-copy,
.h5-auth-form,
.h5-auth-field,
.h5-auth-invite-card,
.h5-auth-footer {
  display: grid;
  gap: 10px;
}

.h5-auth-panel-copy strong {
  color: var(--h5-text);
  font-size: 16px;
  line-height: 1.45;
}

.h5-auth-panel-copy p,
.h5-auth-agreement,
.h5-auth-feedback,
.h5-auth-invite-card p,
.h5-auth-footer p {
  color: var(--h5-text-soft);
  font-size: 12px;
  line-height: 1.7;
}

.h5-auth-field span,
.h5-auth-invite-card span,
.h5-auth-scene-card span {
  color: #7a8798;
  font-size: 12px;
  font-weight: 700;
}

.h5-auth-field input {
  min-height: 48px;
  border: 1px solid rgba(18, 52, 95, 0.12);
  border-radius: 16px;
  padding: 0 14px;
  background: rgba(248, 250, 253, 0.96);
  color: var(--h5-text);
  font-size: 14px;
}

.h5-auth-field input:focus {
  outline: 0;
  border-color: rgba(24, 58, 110, 0.28);
  background: #fff;
}

.h5-auth-invite-card,
.h5-auth-scene-card {
  padding: 14px;
  background: rgba(247, 249, 252, 0.92);
}

.h5-auth-invite-card strong,
.h5-auth-scene-card strong {
  color: var(--h5-text);
  font-size: 14px;
  line-height: 1.6;
}

.h5-auth-feedback.error {
  color: var(--h5-danger);
}

.h5-auth-feedback.success {
  color: var(--h5-success);
}

.h5-auth-scene-cards {
  display: grid;
  gap: 10px;
}

.h5-auth-footer {
  padding: 0 4px;
}

@media (prefers-reduced-motion: reduce) {
  .h5-auth-tab {
    transition: none;
  }
}
</style>
