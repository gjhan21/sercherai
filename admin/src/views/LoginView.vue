<script setup>
import { reactive, ref } from "vue";
import { useRouter } from "vue-router";
import { login, mockLogin } from "../api/auth";
import { saveSession } from "../lib/session";

const router = useRouter();
const submitting = ref(false);
const errorMessage = ref("");
const mode = ref("mock");

const form = reactive({
  phone: "",
  password: "",
  expire_seconds: 86400,
  user_id: "admin_001",
  role: "ADMIN"
});

async function handleSubmit() {
  errorMessage.value = "";
  submitting.value = true;
  try {
    const payload =
      mode.value === "password"
        ? await login({
            phone: form.phone.trim(),
            password: form.password,
            expire_seconds: form.expire_seconds
          })
        : await mockLogin({
            user_id: form.user_id.trim(),
            role: form.role,
            expire_seconds: form.expire_seconds
          });

    if ((payload.role || "").toUpperCase() !== "ADMIN") {
      throw new Error("当前账号不是 ADMIN 角色");
    }
    saveSession(payload);
    await router.replace("/dashboard");
  } catch (error) {
    errorMessage.value = error.message || "登录失败";
  } finally {
    submitting.value = false;
  }
}
</script>

<template>
  <div class="login-page">
    <div class="login-card">
      <h1 class="login-title">SercherAI Admin</h1>
      <p class="muted">先完成登录，进入管理端</p>

      <div class="login-mode">
        <label class="mode-item">
          <input v-model="mode" type="radio" value="mock" />
          Mock 登录（开发环境）
        </label>
        <label class="mode-item">
          <input v-model="mode" type="radio" value="password" />
          密码登录
        </label>
      </div>

      <form class="grid" @submit.prevent="handleSubmit">
        <div v-if="mode === 'password'" class="form-grid">
          <div class="form-item">
            <label>手机号</label>
            <input v-model="form.phone" class="input" placeholder="13800000001" required />
          </div>
          <div class="form-item">
            <label>密码</label>
            <input v-model="form.password" class="input" type="password" placeholder="请输入密码" required />
          </div>
        </div>

        <div v-else class="form-grid">
          <div class="form-item">
            <label>用户 ID</label>
            <input v-model="form.user_id" class="input" placeholder="admin_001" required />
          </div>
          <div class="form-item">
            <label>角色</label>
            <select v-model="form.role" class="select">
              <option value="ADMIN">ADMIN</option>
            </select>
          </div>
        </div>

        <div class="form-item">
          <label>Token 有效期（秒）</label>
          <input v-model.number="form.expire_seconds" class="input" type="number" min="60" />
        </div>

        <div v-if="errorMessage" class="error-message">{{ errorMessage }}</div>

        <button class="btn btn-primary" :disabled="submitting" type="submit">
          {{ submitting ? "登录中..." : "登录" }}
        </button>
      </form>

      <p class="hint">
        若使用 Mock 登录，请确保后端启动时设置了 `ALLOW_MOCK_LOGIN=true`。
      </p>
    </div>
  </div>
</template>

<style scoped>
.login-page {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 20px;
}

.login-card {
  width: min(520px, 100%);
  background: #fff;
  border: 1px solid #e5e7eb;
  border-radius: 14px;
  padding: 22px;
}

.login-title {
  margin: 0 0 6px;
  font-size: 26px;
}

.login-mode {
  display: flex;
  flex-wrap: wrap;
  gap: 14px;
  margin: 14px 0;
}

.mode-item {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  font-size: 14px;
}
</style>
