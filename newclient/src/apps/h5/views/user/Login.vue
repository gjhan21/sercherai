<template>
  <div class="auth-layout">
    <div class="h5-auth-card">
      <div class="h5-auth-brand">
        <svg viewBox="0 0 32 32" fill="none" stroke="currentColor" stroke-width="2" width="36" height="36" style="color:var(--accent-gold)"><circle cx="16" cy="16" r="14"/><path d="M10 20l4-6 4 3 4-7" stroke-width="2.5"/></svg>
        <h1>智投AI</h1>
      </div>
      <div class="h5-auth-field"><input type="text" placeholder="手机号/邮箱" v-model="account" /></div>
      <div class="h5-auth-field">
        <input type="password" placeholder="密码" v-model="password" @keyup.enter="handleLogin" />
        <div class="forgot-link-container">
          <button class="link-btn-sm" @click="$router.push('/profile/forgot-password')">忘记密码？</button>
        </div>
      </div>
      <p v-if="errorMsg" class="auth-error">{{ errorMsg }}</p>
      <button class="h5-auth-btn" :disabled="!account || !password || loading" @click="handleLogin">{{ loading ? '登录中...' : '登录' }}</button>
      <p class="h5-auth-link">没有账号？<button class="link" @click="$router.push('/register')">注册</button></p>
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
    const result = await login({ account: account.value, password: password.value });
    setClientAuthSession(result);
    router.push("/");
  } catch (e) {
    errorMsg.value = e?.message || "登录失败";
  } finally {
    loading.value = false;
  }
}
</script>

<style scoped>
.h5-auth-card { width: 100%; max-width: 360px; margin: 0 auto; display: grid; gap: 14px; }
.h5-auth-brand { text-align: center; margin-bottom: 20px; }
.h5-auth-brand h1 { font-size: 24px; font-weight: 800; background: linear-gradient(135deg,var(--accent-gold),#fff); -webkit-background-clip: text; -webkit-text-fill-color: transparent; background-clip: text; }
.h5-auth-field input { width: 100%; padding: 14px; border-radius: var(--radius-sm); background: rgba(255,255,255,.04); border: 1px solid var(--border); color: var(--text-primary); font-size: 14px; }
.h5-auth-field input:focus { border-color: var(--accent-gold); }
.auth-error { color: var(--negative); font-size: 12px; text-align: center; }
.h5-auth-btn { width: 100%; padding: 14px; border-radius: var(--radius-full); background: linear-gradient(135deg,var(--accent-gold),var(--accent-gold-dim)); color: #000; font-size: 15px; font-weight: 600; }
.h5-auth-btn:disabled { opacity: .5; }
.h5-auth-link { text-align: center; font-size: 13px; color: var(--text-secondary); }
.link { color: var(--accent-gold); font-weight: 600; }
.forgot-link-container { display: flex; justify-content: flex-end; margin-top: 6px; }
.link-btn-sm { background: none; border: none; color: var(--accent-gold); font-size: 12px; cursor: pointer; padding: 0; }
.link-btn-sm:active { opacity: 0.7; }
</style>
