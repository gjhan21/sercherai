<template>
  <div class="auth-layout">
    <div class="h5-auth-card">
      <div class="h5-auth-brand">
        <svg viewBox="0 0 32 32" fill="none" stroke="currentColor" stroke-width="2" width="36" height="36" style="color:var(--accent-gold)"><circle cx="16" cy="16" r="14"/><path d="M10 20l4-6 4 3 4-7" stroke-width="2.5"/></svg>
        <h1>找回密码</h1>
        <p style="font-size:12px;color:var(--text-secondary);margin-top:6px;">输入注册时绑定的邮箱以重置密码</p>
      </div>

      <!-- Step 1: Input Email -->
      <div v-if="step === 1" class="form-container">
        <div class="h5-auth-field"><input type="email" placeholder="注册邮箱" v-model="email" /></div>
        <p v-if="errorMsg" class="auth-error">{{ errorMsg }}</p>
        <button class="h5-auth-btn" :disabled="!email || loading" @click="handleSendCode">{{ loading ? '发送中...' : '发送验证码' }}</button>
      </div>

      <!-- Step 2: Verification and Reset -->
      <div v-else class="form-container">
        <div class="h5-auth-field"><input type="text" placeholder="输入6位验证码" v-model="code" /></div>
        <div class="h5-auth-field"><input type="password" placeholder="设置新密码" v-model="newPassword" /></div>
        <div class="h5-auth-field"><input type="password" placeholder="确认新密码" v-model="confirmPassword" /></div>
        <p v-if="errorMsg" class="auth-error">{{ errorMsg }}</p>
        <p v-if="successMsg" class="auth-success">{{ successMsg }}</p>
        <button class="h5-auth-btn" :disabled="!code || !newPassword || !confirmPassword || loading" @click="handleResetPassword">{{ loading ? '提交中...' : '重置密码' }}</button>
      </div>

      <p class="h5-auth-link"><button class="link" @click="$router.push('/login')">返回登录</button></p>
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
    if (res?.data?.code) {
      code.value = res.data.code;
    }
  } catch (e) {
    errorMsg.value = e?.message || "发送失败";
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
    successMsg.value = "重置成功，正在返回登录页...";
    setTimeout(() => {
      router.push("/login");
    }, 2000);
  } catch (e) {
    errorMsg.value = e?.message || "重置失败";
  } finally {
    loading.value = false;
  }
}
</script>

<style scoped>
.auth-layout { min-height: 100vh; display: flex; align-items: center; justify-content: center; background: var(--bg-primary); padding: 20px; }
.h5-auth-card { width: 100%; max-width: 360px; margin: 0 auto; display: grid; gap: 14px; }
.h5-auth-brand { text-align: center; margin-bottom: 20px; }
.h5-auth-brand h1 { font-size: 24px; font-weight: 800; background: linear-gradient(135deg,var(--accent-gold),#fff); -webkit-background-clip: text; -webkit-text-fill-color: transparent; background-clip: text; }
.form-container { display: grid; gap: 14px; }
.h5-auth-field input { width: 100%; padding: 14px; border-radius: var(--radius-sm); background: rgba(255,255,255,.04); border: 1px solid var(--border); color: var(--text-primary); font-size: 14px; }
.h5-auth-field input:focus { border-color: var(--accent-gold); }
.auth-error { color: var(--negative); font-size: 12px; text-align: center; }
.auth-success { color: var(--positive); font-size: 12px; text-align: center; }
.h5-auth-btn { width: 100%; padding: 14px; border-radius: var(--radius-full); background: linear-gradient(135deg,var(--accent-gold),var(--accent-gold-dim)); color: #000; font-size: 15px; font-weight: 600; border: none; cursor: pointer; }
.h5-auth-btn:disabled { opacity: .5; }
.h5-auth-link { text-align: center; font-size: 13px; color: var(--text-secondary); }
.link { color: var(--accent-gold); font-weight: 600; background: none; border: none; cursor: pointer; }
</style>
