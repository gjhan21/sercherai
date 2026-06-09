<template>
  <div class="auth-layout">
    <div class="auth-card glass">
      <div class="auth-brand">
        <div class="auth-brand-icon">
          <svg viewBox="0 0 32 32" fill="none" stroke="currentColor" stroke-width="2"><circle cx="16" cy="16" r="14"/><path d="M10 20l4-6 4 3 4-7" stroke-width="2.5"/></svg>
        </div>
        <h1>智投AI</h1>
        <p>智能荐股平台</p>
      </div>
      <div class="auth-form">
        <div class="auth-field">
          <label>手机号 / 邮箱</label>
          <input type="text" placeholder="请输入手机号或邮箱" v-model="account" />
        </div>
        <div class="auth-field">
          <div class="field-header">
            <label>密码</label>
            <button class="link-btn-sm" @click="$router.push('/forgot-password')">忘记密码？</button>
          </div>
          <input type="password" placeholder="请输入密码" v-model="password" @keyup.enter="handleLogin" />
        </div>
        <p v-if="errorMsg" class="auth-error">{{ errorMsg }}</p>
        <button class="auth-btn" :disabled="!account || !password || loading" @click="handleLogin">{{ loading ? '登录中...' : '登录' }}</button>
        <p class="auth-switch">还没有账号？<button class="link" @click="$router.push('/register')">立即注册</button></p>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref } from "vue";
import { useRouter } from "vue-router";
import { login } from "@/api/auth";
import { setClientAuthSession } from "@/shared/auth/client-auth";

const router = useRouter();
const account = ref("");
const password = ref("");
const loading = ref(false);
const errorMsg = ref("");

async function handleLogin() {
  if (!account.value || !password.value) return;
  loading.value = true;
  errorMsg.value = "";
  try {
    const payload = { account: account.value, password: password.value };
    const result = await login(payload);
    setClientAuthSession(result);
    router.push("/");
  } catch (e) {
    errorMsg.value = e?.message || "登录失败，请重试";
  } finally {
    loading.value = false;
  }
}
</script>

<style scoped>
.auth-card { width: 400px; max-width: 90vw; padding: 40px 32px; border-radius: var(--radius-xl); text-align: center; }
.auth-brand { margin-bottom: 32px; }
.auth-brand-icon { width: 48px; height: 48px; color: var(--accent-gold); margin: 0 auto 12px; filter: drop-shadow(0 0 8px rgba(240,185,11,.3)); }
.auth-brand h1 { font-size: 24px; font-weight: 800; background: linear-gradient(135deg,var(--accent-gold),#fff); -webkit-background-clip: text; -webkit-text-fill-color: transparent; background-clip: text; }
.auth-brand p { font-size: 13px; color: var(--text-secondary); }
.auth-form { text-align: left; }
.auth-field { margin-bottom: 16px; }
.auth-field label { display: block; font-size: 13px; color: var(--text-secondary); margin-bottom: 6px; }
.auth-field input { width: 100%; padding: 12px 14px; border-radius: var(--radius-sm); background: rgba(255,255,255,.04); border: 1px solid var(--border); color: var(--text-primary); font-size: 14px; }
.auth-field input:focus { border-color: var(--accent-gold); }
.auth-error { color: var(--negative); font-size: 12px; margin-bottom: 8px; }
.auth-btn { width: 100%; padding: 14px; border-radius: var(--radius-full); background: linear-gradient(135deg,var(--accent-gold),var(--accent-gold-dim)); color: #000; font-size: 15px; font-weight: 600; margin: 8px 0 16px; }
.auth-btn:disabled { opacity: .5; }
.auth-switch { text-align: center; font-size: 13px; color: var(--text-secondary); }
.link { color: var(--accent-gold); font-weight: 600; }
.field-header { display: flex; justify-content: space-between; align-items: center; }
.link-btn-sm { background: none; border: none; color: var(--accent-gold); font-size: 12px; cursor: pointer; padding: 0; }
.link-btn-sm:hover { text-decoration: underline; }
</style>
