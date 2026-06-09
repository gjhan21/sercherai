<template>
  <div class="auth-layout">
    <div class="auth-card glass">
      <div class="auth-brand">
        <div class="auth-brand-icon">
          <svg viewBox="0 0 32 32" fill="none" stroke="currentColor" stroke-width="2"><circle cx="16" cy="16" r="14"/><path d="M10 20l4-6 4 3 4-7" stroke-width="2.5"/></svg>
        </div>
        <h1>找回密码</h1>
        <p>请输入注册时绑定的邮箱以重置密码</p>
      </div>
      <div class="auth-form">
        <!-- Step 1: Input Email -->
        <div v-if="step === 1">
          <div class="auth-field">
            <label>注册邮箱</label>
            <input type="email" placeholder="请输入电子邮箱" v-model="email" />
          </div>
          <p v-if="errorMsg" class="auth-error">{{ errorMsg }}</p>
          <button class="auth-btn" :disabled="!email || loading" @click="handleSendCode">{{ loading ? '发送中...' : '发送验证码' }}</button>
        </div>

        <!-- Step 2: Verification and Reset -->
        <div v-else>
          <div class="auth-field">
            <label>验证码</label>
            <input type="text" placeholder="输入6位验证码" v-model="code" />
          </div>
          <div class="auth-field">
            <label>设置新密码</label>
            <input type="password" placeholder="至少8位密码" v-model="newPassword" />
          </div>
          <div class="auth-field">
            <label>确认新密码</label>
            <input type="password" placeholder="再次输入新密码" v-model="confirmPassword" />
          </div>
          <p v-if="errorMsg" class="auth-error">{{ errorMsg }}</p>
          <p v-if="successMsg" class="auth-success">{{ successMsg }}</p>
          <button class="auth-btn" :disabled="!code || !newPassword || !confirmPassword || loading" @click="handleResetPassword">{{ loading ? '提交中...' : '重置密码' }}</button>
        </div>

        <p class="auth-switch"><button class="link" @click="$router.push('/login')">返回登录</button></p>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref } from "vue";
import { useRouter } from "vue-router";
import { forgotPassword, resetPassword } from "@/api/auth";

const router = useRouter();
const step = ref(1);
const email = ref("");
const code = ref("");
const newPassword = ref("");
const confirmPassword = ref("");
const loading = ref(false);
const errorMsg = ref("");
const successMsg = ref("");

async function handleSendCode() {
  if (!email.value) return;
  loading.value = true;
  errorMsg.value = "";
  try {
    const res = await forgotPassword(email.value);
    step.value = 2;
    // Dev helper: auto-fill code if present in response
    if (res?.data?.code) {
      code.value = res.data.code;
    }
  } catch (e) {
    errorMsg.value = e?.message || "发送验证码失败";
  } finally {
    loading.value = false;
  }
}

async function handleResetPassword() {
  if (!code.value || !newPassword.value) return;
  if (newPassword.value !== confirmPassword.value) { errorMsg.value = "两次密码不一致"; return; }
  if (newPassword.value.length < 8) { errorMsg.value = "新密码至少8位"; return; }
  loading.value = true;
  errorMsg.value = "";
  successMsg.value = "";
  try {
    await resetPassword({
      email: email.value,
      code: code.value,
      new_password: newPassword.value
    });
    successMsg.value = "密码重置成功，正在返回登录页...";
    setTimeout(() => {
      router.push("/login");
    }, 2000);
  } catch (e) {
    errorMsg.value = e?.message || "密码重置失败";
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
.auth-success { color: var(--positive); font-size: 12px; margin-bottom: 8px; }
.auth-btn { width: 100%; padding: 14px; border-radius: var(--radius-full); background: linear-gradient(135deg,var(--accent-gold),var(--accent-gold-dim)); color: #000; font-size: 15px; font-weight: 600; margin: 8px 0 16px; }
.auth-btn:disabled { opacity: .5; }
.auth-switch { text-align: center; font-size: 13px; color: var(--text-secondary); }
.link { color: var(--accent-gold); font-weight: 600; background: none; border: none; cursor: pointer; }
</style>
