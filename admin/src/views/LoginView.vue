<script setup>
import { reactive, ref } from "vue";
import { useRouter } from "vue-router";
import { login, mockLogin } from "../api/auth";
import { getAccessProfile } from "../api/admin";
import { saveSession } from "../lib/session";

const router = useRouter();
const submitting = ref(false);
const errorMessage = ref("");
const mode = ref("password");

const form = reactive({
  phone: "13800000000",
  password: "abc123456",
  expire_seconds: 86400,
  user_id: "admin_001",
  role: "ADMIN"
});

async function handleSubmit() {
  errorMessage.value = "";
  submitting.value = true;
  try {
    let payload;
    if (mode.value === "password") {
      const phone = form.phone.trim();
      const password = form.password;
      if (!phone || !password) {
        throw new Error("手机号和密码不能为空");
      }
      payload = await login({
        phone,
        password,
        expire_seconds: form.expire_seconds
      });
    } else {
      const userID = form.user_id.trim();
      if (!userID) {
        throw new Error("Mock 登录时用户 ID 不能为空");
      }
      payload = await mockLogin({
        user_id: userID,
        role: form.role,
        expire_seconds: form.expire_seconds
      });
    }

    if ((payload.role || "").toUpperCase() !== "ADMIN") {
      throw new Error("当前账号不是 ADMIN 角色");
    }
    let accessProfile = { permission_codes: [], roles: [] };
    try {
      accessProfile = await getAccessProfile();
    } catch (error) {
      console.warn("load access profile failed:", error?.message || error);
    }
    saveSession({
      ...payload,
      permission_codes: accessProfile.permission_codes || [],
      roles: accessProfile.roles || []
    });
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
    <el-card class="login-card" shadow="hover">
      <div class="login-head">
        <h1 class="login-title">SercherAI Admin</h1>
        <p class="muted">先完成登录，进入管理端</p>
      </div>

      <el-radio-group v-model="mode" size="large" class="login-mode">
        <el-radio-button label="mock">Mock 登录（开发）</el-radio-button>
        <el-radio-button label="password">密码登录</el-radio-button>
      </el-radio-group>

      <el-form label-position="top" @submit.prevent="handleSubmit">
        <template v-if="mode === 'password'">
          <el-form-item label="手机号">
            <el-input v-model="form.phone" placeholder="13800000000" clearable />
          </el-form-item>
          <el-form-item label="密码">
            <el-input v-model="form.password" type="password" placeholder="请输入密码" show-password />
          </el-form-item>
          <div class="quick-account">
            <el-button text @click="() => ((form.phone = '13800000000'), (form.password = 'abc123456'))">
              使用 admin_001
            </el-button>
            <el-button text @click="() => ((form.phone = '13800000010'), (form.password = 'abc123456'))">
              使用 admin_002
            </el-button>
          </div>
          <el-alert
            title="开发测试账号：13800000000 / abc123456，13800000010 / abc123456"
            type="info"
            :closable="false"
            class="login-hint"
          />
        </template>

        <template v-else>
          <el-form-item label="用户 ID">
            <el-input v-model="form.user_id" placeholder="admin_001" clearable />
          </el-form-item>
          <el-form-item label="角色">
            <el-select v-model="form.role">
              <el-option label="ADMIN" value="ADMIN" />
            </el-select>
          </el-form-item>
        </template>

        <el-form-item label="Token 有效期（秒）">
          <el-input-number v-model="form.expire_seconds" :min="60" :max="604800" controls-position="right" />
        </el-form-item>

        <el-alert v-if="errorMessage" :title="errorMessage" type="error" show-icon class="login-alert" />

        <el-button class="login-submit" type="primary" :loading="submitting" @click="handleSubmit">
          登录
        </el-button>
      </el-form>

      <el-text type="info" size="small">
        若使用 Mock 登录，请确保后端启动时设置了 `ALLOW_MOCK_LOGIN=true`。
      </el-text>
    </el-card>
  </div>
</template>

<style scoped>
.login-page {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 24px;
  background: radial-gradient(circle at 10% 20%, #e0f2fe 0%, #eef2ff 42%, #f8fafc 100%);
}

.login-card {
  width: min(520px, 100%);
  border-radius: 14px;
}

.login-head {
  margin-bottom: 16px;
}

.login-title {
  margin: 0 0 4px;
  font-size: 26px;
  font-weight: 700;
  color: #0f172a;
}

.login-mode {
  margin-bottom: 16px;
}

.login-alert {
  margin-bottom: 12px;
}

.quick-account {
  margin-bottom: 12px;
}

.login-hint {
  margin-bottom: 12px;
}

.login-submit {
  width: 100%;
}
</style>
