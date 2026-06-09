<template>
  <div class="auth-layout">
    <div class="auth-card glass">
      <div class="auth-brand">
        <div class="auth-brand-icon">
          <svg viewBox="0 0 32 32" fill="none" stroke="currentColor" stroke-width="2"><circle cx="16" cy="16" r="14"/><path d="M10 20l4-6 4 3 4-7" stroke-width="2.5"/></svg>
        </div>
        <h1>创建账号</h1>
        <p>加入智投AI，开启智能投资之旅</p>
      </div>
      <div class="auth-form">
        <div class="auth-field">
          <label>手机号</label>
          <input type="text" placeholder="请输入手机号" v-model="phone" />
        </div>
        <div class="auth-field">
          <label>邮箱</label>
          <input type="email" placeholder="请输入电子邮箱" v-model="email" />
          <p class="field-hint">💡 邮箱为帐号找回，密码重置唯一凭证</p>
        </div>
        <div class="auth-field">
          <label>设置密码</label>
          <input type="password" placeholder="至少6位密码" v-model="password" />
        </div>
        <div class="auth-field">
          <label>确认密码</label>
          <input type="password" placeholder="再次输入密码" v-model="confirmPwd" />
        </div>
        <p v-if="errorMsg" class="auth-error">{{ errorMsg }}</p>
        <button class="auth-btn" :disabled="!phone || !password || loading" @click="handleRegister">{{ loading ? '注册中...' : '注册' }}</button>
        <p class="auth-switch">已有账号？<button class="link" @click="$router.push('/login')">立即登录</button></p>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref } from "vue";
import { useRouter } from "vue-router";
import { register } from "@/api/auth";
import { setClientAuthSession } from "@/shared/auth/client-auth";

const router = useRouter();
const phone = ref("");
const email = ref("");
const password = ref("");
const confirmPwd = ref("");
const loading = ref(false);
const errorMsg = ref("");

async function handleRegister() {
  if (!phone.value || !email.value || !password.value) { errorMsg.value = "请填写完整"; return; }
  if (password.value !== confirmPwd.value) { errorMsg.value = "两次密码不一致"; return; }
  if (password.value.length < 6) { errorMsg.value = "密码至少6位"; return; }
  loading.value = true;
  errorMsg.value = "";
  try {
    const payload = { phone: phone.value, email: email.value, password: password.value };
    const result = await register(payload);
    setClientAuthSession(result);
    router.push("/");
  } catch (e) {
    errorMsg.value = e?.message || "注册失败，请重试";
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
.field-hint { font-size: 11px; color: var(--accent-gold); margin-top: 4px; opacity: 0.9; }
</style>
