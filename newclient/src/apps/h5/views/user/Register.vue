<template>
  <div class="auth-layout">
    <div class="h5-auth-card">
      <div class="h5-auth-brand">
        <svg viewBox="0 0 32 32" fill="none" stroke="currentColor" stroke-width="2" width="36" height="36" style="color:var(--accent-gold)"><circle cx="16" cy="16" r="14"/><path d="M10 20l4-6 4 3 4-7" stroke-width="2.5"/></svg>
        <h1>注册账号</h1>
      </div>
      <div class="h5-auth-field"><input type="text" placeholder="手机号" v-model="phone" /></div>
      <div class="h5-auth-field">
        <input type="email" placeholder="电子邮箱" v-model="email" />
        <p class="field-hint">💡 邮箱为帐号找回，密码重置唯一凭证</p>
      </div>
      <div class="h5-auth-field"><input type="password" placeholder="密码（至少6位）" v-model="password" /></div>
      <div class="h5-auth-field"><input type="password" placeholder="确认密码" v-model="confirmPwd" /></div>
      <p v-if="errorMsg" class="auth-error">{{ errorMsg }}</p>
      <button class="h5-auth-btn" :disabled="!phone || !email || !password || loading" @click="handleRegister">{{ loading ? '注册中...' : '注册' }}</button>
      <p class="h5-auth-link">已有账号？<button class="link" @click="$router.push('/login')">登录</button></p>
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
    const result = await register({ phone: phone.value, email: email.value, password: password.value });
    setClientAuthSession(result);
    router.push("/");
  } catch (e) {
    errorMsg.value = e?.message || "注册失败";
  } finally {
    loading.value = false;
  }
}
</script>

<style scoped>
.auth-layout { min-height: 100vh; display: flex; align-items: center; justify-content: center; background: var(--bg-primary); padding: 20px; }
.h5-auth-card { width: 100%; max-width: 360px; display: grid; gap: 14px; }
.h5-auth-brand { text-align: center; margin-bottom: 20px; }
.h5-auth-brand h1 { font-size: 24px; font-weight: 800; background: linear-gradient(135deg,var(--accent-gold),#fff); -webkit-background-clip: text; -webkit-text-fill-color: transparent; background-clip: text; }
.h5-auth-field input { width: 100%; padding: 14px; border-radius: var(--radius-sm); background: rgba(255,255,255,.04); border: 1px solid var(--border); color: var(--text-primary); font-size: 14px; }
.h5-auth-field input:focus { border-color: var(--accent-gold); }
.auth-error { color: var(--negative); font-size: 12px; text-align: center; }
.h5-auth-btn { width: 100%; padding: 14px; border-radius: var(--radius-full); background: linear-gradient(135deg,var(--accent-gold),var(--accent-gold-dim)); color: #000; font-size: 15px; font-weight: 600; border: none; cursor: pointer; }
.h5-auth-btn:disabled { opacity: .5; }
.h5-auth-link { text-align: center; font-size: 13px; color: var(--text-secondary); }
.link { color: var(--accent-gold); font-weight: 600; background: none; border: none; cursor: pointer; }
.field-hint { font-size: 11px; color: var(--accent-gold); margin-top: 4px; opacity: 0.9; text-align: left; padding-left: 4px; }
</style>
