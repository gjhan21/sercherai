<template>
  <section class="auth-page fade-up">
    <article class="auth-shell card">
      <div class="auth-copy">
        <p class="auth-kicker">账户中心</p>
        <h1 class="section-title">登录后可同步会员、订阅与阅读数据。</h1>
        <p class="section-subtitle">
          已接入注册、登录、退出全流程，支持手机号/邮箱账号与分享邀请注册。
        </p>
        <ul class="auth-points">
          <li>手机号/邮箱注册与密码登录</li>
          <li>分享邀请码注册自动绑定邀请关系</li>
          <li>Token 自动续期与失效清理</li>
          <li>支持退出当前设备登录状态</li>
        </ul>
      </div>

      <div class="auth-panel">
        <div class="mode-tabs">
          <button
            type="button"
            :class="{ active: mode === 'login' }"
            @click="mode = 'login'"
          >
            登录
          </button>
          <button
            type="button"
            :class="{ active: mode === 'register' }"
            @click="mode = 'register'"
          >
            注册
          </button>
        </div>

        <form class="auth-form" @submit.prevent="handleSubmit">
          <label>
            手机号 / 邮箱
            <input
              v-model.trim="form.account"
              type="text"
              placeholder="请输入手机号或邮箱"
              autocomplete="username"
            />
          </label>
          <label>
            密码
            <input
              v-model="form.password"
              type="password"
              placeholder="请输入密码"
              autocomplete="current-password"
            />
          </label>
          <p v-if="mode === 'register' && inviteCode" class="form-tip">
            当前通过分享链接注册，邀请码：{{ inviteCode }}
          </p>

          <p v-if="errorMessage" class="form-tip error">{{ errorMessage }}</p>
          <p v-else-if="successMessage" class="form-tip success">{{ successMessage }}</p>

          <button type="submit" class="submit-btn" :disabled="submitting">
            {{
              submitting
                ? "提交中..."
                : mode === "login"
                  ? "立即登录"
                  : "注册并登录"
            }}
          </button>
        </form>

        <div class="quick-login">
          <p>测试账号</p>
          <button type="button" @click="fillDemoAccount">13800000000 / abc123456</button>
        </div>
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

const mode = ref("login");
const submitting = ref(false);
const errorMessage = ref("");
const successMessage = ref("");

const form = reactive({
  account: "",
  password: ""
});

const inviteCode = computed(() =>
  String(route.query.invite_code || route.query.code || route.query.invite || "").trim().toUpperCase()
);

function normalizeAccount(account) {
  return String(account || "").replace(/\s+/g, "").trim();
}

function validateForm() {
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
    password: form.password
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

function fillDemoAccount() {
  form.account = "13800000000";
  form.password = "abc123456";
}

async function handleSubmit() {
  errorMessage.value = "";
  successMessage.value = "";
  submitting.value = true;
  try {
    const account = validateForm();
    let payload = null;
    if (mode.value === "login") {
      payload = await login({
        account,
        password: form.password
      });
    } else {
      payload = await register(buildRegisterPayload(account));
      successMessage.value = "注册成功，正在进入首页...";
    }
    setClientAuthSession({
      ...payload,
      phone: account.includes("@") ? "" : account,
      email: account.includes("@") ? account.toLowerCase() : ""
    });
    await router.replace(resolveRedirectPath());
  } catch (error) {
    errorMessage.value = error?.message || "提交失败";
  } finally {
    submitting.value = false;
  }
}
</script>

<style scoped>
.auth-page {
  min-height: 100vh;
  display: grid;
  place-items: center;
  padding: 20px 14px;
}

.auth-shell {
  width: min(980px, 100%);
  border-radius: 22px;
  padding: 16px;
  display: grid;
  grid-template-columns: 1fr minmax(300px, 390px);
  gap: 14px;
}

.auth-copy {
  border-radius: 16px;
  padding: 14px;
  background:
    radial-gradient(circle at 100% 0%, rgba(63, 127, 113, 0.15) 0%, transparent 34%),
    rgba(246, 244, 239, 0.82);
  border: 1px solid rgba(216, 223, 216, 0.85);
}

.auth-kicker {
  margin: 0;
  font-size: 12px;
  color: var(--color-pine-600);
}

.auth-copy h1 {
  margin-top: 8px;
}

.auth-points {
  margin: 12px 0 0;
  padding-left: 18px;
  color: var(--color-text-sub);
  display: grid;
  gap: 6px;
  font-size: 13px;
}

.auth-panel {
  border-radius: 16px;
  border: 1px solid rgba(216, 223, 216, 0.9);
  background: rgba(255, 255, 255, 0.9);
  padding: 12px;
  display: grid;
  gap: 12px;
}

.mode-tabs {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 8px;
}

.mode-tabs button {
  border: 1px solid rgba(216, 223, 216, 0.9);
  border-radius: 10px;
  background: rgba(246, 244, 239, 0.8);
  padding: 8px 10px;
  color: var(--color-text-sub);
  cursor: pointer;
}

.mode-tabs button.active {
  border-color: transparent;
  color: #fff;
  background: linear-gradient(145deg, var(--color-pine-700), var(--color-pine-500));
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
  border: 1px solid rgba(216, 223, 216, 0.95);
  border-radius: 10px;
  padding: 9px 10px;
  background: rgba(255, 255, 255, 0.94);
  color: var(--color-text-main);
}

.auth-form input:focus {
  outline: none;
  border-color: var(--color-pine-500);
  box-shadow: 0 0 0 3px rgba(63, 127, 113, 0.15);
}

.form-tip {
  margin: 0;
  border-radius: 8px;
  padding: 8px 9px;
  font-size: 12px;
}

.form-tip.error {
  background: rgba(178, 58, 42, 0.1);
  color: var(--color-danger);
}

.form-tip.success {
  background: rgba(46, 125, 50, 0.1);
  color: var(--color-success);
}

.submit-btn {
  border: 0;
  border-radius: 11px;
  padding: 10px 12px;
  color: #fff;
  background: linear-gradient(145deg, var(--color-pine-700), var(--color-pine-500));
  font-weight: 600;
  cursor: pointer;
}

.submit-btn:disabled {
  opacity: 0.7;
  cursor: not-allowed;
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
  border: 1px dashed rgba(63, 127, 113, 0.5);
  border-radius: 10px;
  padding: 8px 10px;
  background: rgba(236, 245, 242, 0.9);
  color: var(--color-pine-700);
  cursor: pointer;
  text-align: left;
}

@media (max-width: 900px) {
  .auth-shell {
    grid-template-columns: 1fr;
    border-radius: 18px;
    padding: 12px;
  }

  .auth-copy {
    padding: 12px;
  }
}
</style>
