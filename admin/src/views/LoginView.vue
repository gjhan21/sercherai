<script setup>
import { reactive, ref } from "vue";
import { useRouter } from "vue-router";
import { login, mockLogin } from "../api/auth";
import { getAccessProfile } from "../api/admin";
import { resolveFirstAccessibleRoute } from "../lib/admin-navigation";
import { clearSession, formatSessionRole, saveSession } from "../lib/session";

const router = useRouter();
const submitting = ref(false);
const errorMessage = ref("");
const allowMockLogin = import.meta.env.DEV && String(import.meta.env.VITE_ADMIN_ALLOW_MOCK_LOGIN || "true").toLowerCase() !== "false";
const showDevLoginAssist = import.meta.env.DEV;
const mode = ref("password");

const form = reactive({
  phone: "",
  password: "",
  expire_seconds: 86400,
  user_id: "admin_001",
  role: "ADMIN"
});

function formatLockedUntil(lockedUntil) {
  if (!lockedUntil) {
    return "";
  }
  const time = new Date(lockedUntil);
  if (Number.isNaN(time.getTime())) {
    return lockedUntil;
  }
  return time.toLocaleString("zh-CN", {
    hour12: false,
    year: "numeric",
    month: "2-digit",
    day: "2-digit",
    hour: "2-digit",
    minute: "2-digit",
    second: "2-digit"
  });
}

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
      if (!allowMockLogin) {
        throw new Error("当前环境未开启 Mock 登录");
      }
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
      throw new Error("当前账号不是管理员角色");
    }

    // Persist the access token first so the profile request can attach Authorization.
    saveSession(payload);

    let accessProfile = { permission_codes: [], roles: [] };
    try {
      accessProfile = await getAccessProfile();
    } catch (error) {
      clearSession();
      throw new Error(error?.message || "权限加载失败，请稍后重试");
    }
    saveSession({
      ...payload,
      permission_codes: accessProfile.permission_codes || [],
      roles: accessProfile.roles || []
    });
    await router.replace(resolveFirstAccessibleRoute());
  } catch (error) {
    if (error?.code === 42901) {
      const lockedUntil = formatLockedUntil(error?.payload?.locked_until);
      errorMessage.value = lockedUntil ? `登录失败次数过多，请在 ${lockedUntil} 后重试` : "登录失败次数过多，请稍后重试";
    } else {
      errorMessage.value = error.message || "登录失败";
    }
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
        <el-radio-button value="password">密码登录</el-radio-button>
        <el-radio-button v-if="allowMockLogin" value="mock">Mock 登录（开发）</el-radio-button>
      </el-radio-group>

      <el-form label-position="top" @submit.prevent="handleSubmit">
        <template v-if="mode === 'password'">
          <el-form-item label="手机号">
            <el-input v-model="form.phone" placeholder="19900000001" clearable />
          </el-form-item>
          <el-form-item label="密码">
            <el-input v-model="form.password" type="password" placeholder="请输入密码" show-password />
          </el-form-item>
          <div v-if="showDevLoginAssist" class="quick-account">
            <el-button text @click="() => ((form.phone = '19900000001'), (form.password = 'abc123456'))">
              使用 admin_001
            </el-button>
            <el-button text @click="() => ((form.phone = '19900000002'), (form.password = 'abc123456'))">
              使用 admin_002
            </el-button>
          </div>
          <el-alert
            v-if="showDevLoginAssist"
            title="开发测试账号：19900000001 / abc123456，19900000002 / abc123456"
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
              <el-option :label="formatSessionRole('ADMIN')" value="ADMIN" />
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

      <el-text v-if="allowMockLogin" type="info" size="small">
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
