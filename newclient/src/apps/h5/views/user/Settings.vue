<template>
  <div class="h5-settings-page">
    <!-- Profile -->
    <div class="h5-settings-profile glass" v-if="session">
      <div class="h5-avatar">{{ (session.phone || session.email || '用')[0] }}</div>
      <div>
        <strong>{{ session.phone ? session.phone.slice(0,3)+'****'+session.phone.slice(-4) : session.email || '未登录' }}</strong>
        <span>ID: {{ session.userID || '--' }}</span>
      </div>
    </div>

    <!-- Notifications -->
    <div class="h5-settings-group">
      <div class="h5-settings-row">
        <div><span>股价异动提醒</span><p>关注股票异动时推送</p></div>
        <label class="toggle"><input type="checkbox" v-model="settings.priceAlert" /><span class="toggle-slider"></span></label>
      </div>
      <div class="h5-settings-row">
        <div><span>资讯推送</span><p>每日 AI 精选资讯</p></div>
        <label class="toggle"><input type="checkbox" v-model="settings.newsAlert" /><span class="toggle-slider"></span></label>
      </div>
      <div class="h5-settings-row">
        <div><span>深色模式</span><p>跟随系统</p></div>
        <label class="toggle"><input type="checkbox" checked /><span class="toggle-slider"></span></label>
      </div>
    </div>

    <!-- Change Password -->
    <div class="h5-settings-group" v-if="isLoggedIn">
      <div class="h5-settings-row" style="flex-direction:column;align-items:stretch;gap:10px;">
        <span style="font-weight:600">修改密码</span>
        <input class="h5-pwd-input" type="password" placeholder="当前密码" v-model="pwd.oldPwd" />
        <input class="h5-pwd-input" type="password" placeholder="新密码（至少6位）" v-model="pwd.newPwd" />
        <input class="h5-pwd-input" type="password" placeholder="确认新密码" v-model="pwd.confirmPwd" />
        <p v-if="pwdError" class="pwd-error">{{ pwdError }}</p>
        <p v-if="pwdSuccess" class="pwd-success">{{ pwdSuccess }}</p>
        <button class="h5-pwd-btn" @click="handleChangePassword">修改密码</button>
      </div>
    </div>

    <!-- Info -->
    <div class="h5-settings-group">
      <div class="h5-settings-row"><span>版本</span><span class="h5-settings-val">v0.1.0</span></div>
      <div class="h5-settings-row"><span>AI 模型</span><span class="h5-settings-val">多因子+LLM</span></div>
    </div>

    <!-- Logout -->
    <button v-if="isLoggedIn" class="h5-logout-btn" @click="handleLogout">退出登录</button>
    <button v-else class="h5-login-btn" @click="$router.push('/login')">登录 / 注册</button>
  </div>
</template>

<script setup>
import { ref } from "vue";
import { useRouter } from "vue-router";
import { useClientAuth } from "@/shared/auth/client-auth";
import { logout, changePassword } from "@/api/auth.js";

const router = useRouter();
const { session, isLoggedIn } = useClientAuth();
const loggingOut = ref(false);

const settings = ref({ priceAlert: true, newsAlert: true });
const pwd = ref({ oldPwd: "", newPwd: "", confirmPwd: "" });
const pwdError = ref("");
const pwdSuccess = ref("");

async function handleChangePassword() {
  pwdError.value = ""; pwdSuccess.value = "";
  if (!pwd.value.oldPwd || !pwd.value.newPwd) { pwdError.value = "请填写完整"; return; }
  if (pwd.value.newPwd.length < 6) { pwdError.value = "新密码至少6位"; return; }
  if (pwd.value.newPwd !== pwd.value.confirmPwd) { pwdError.value = "两次密码不一致"; return; }
  try {
    await changePassword({ old_password: pwd.value.oldPwd, new_password: pwd.value.newPwd });
    pwdSuccess.value = "密码修改成功！";
    pwd.value = { oldPwd: "", newPwd: "", confirmPwd: "" };
  } catch (e) {
    pwdError.value = e?.message || "修改失败";
  }
}

async function handleLogout() {
  if (loggingOut.value) return;
  loggingOut.value = true;
  try {
    const refreshToken = session.value?.refreshToken || "";
    if (refreshToken) await logout(refreshToken);
  } catch { /* ignore */ }
  const { clearClientAuthSession } = await import("@/shared/auth/client-auth");
  clearClientAuthSession();
  loggingOut.value = false;
  router.push("/login");
}
</script>

<style scoped>
.h5-settings-page { display: grid; gap: 14px; }
.h5-settings-profile { display: flex; align-items: center; gap: 12px; padding: 16px; border-radius: var(--radius-lg); }
.h5-avatar { width: 44px; height: 44px; border-radius: 50%; background: var(--accent-gold-glow); color: var(--accent-gold); display: flex; align-items: center; justify-content: center; font-size: 18px; font-weight: 700; }
.h5-settings-profile strong { display: block; font-size: 14px; }
.h5-settings-profile span { font-size: 11px; color: var(--text-muted); }
.h5-settings-group { background: var(--bg-card); border: 1px solid var(--border); border-radius: var(--radius-lg); padding: 8px 16px; }
.h5-settings-row { display: flex; align-items: center; justify-content: space-between; padding: 14px 0; border-bottom: 1px solid rgba(255,255,255,.03); font-size: 14px; gap: 8px; }
.h5-settings-row:last-child { border-bottom: none; }
.h5-settings-row p { font-size: 11px; color: var(--text-muted); margin-top: 2px; }
.h5-settings-val { color: var(--text-secondary); font-size: 13px; }
.pwd-error { color: var(--negative); font-size: 12px; }
.pwd-success { color: var(--positive); font-size: 12px; }
.h5-pwd-input { padding: 10px 12px; border-radius: var(--radius-sm); background: rgba(255,255,255,.04); border: 1px solid var(--border); color: var(--text-primary); font-size: 13px; }
.h5-pwd-input:focus { border-color: var(--accent-gold); }
.h5-pwd-btn { padding: 10px; border-radius: var(--radius-full); border: 1px solid var(--border); font-size: 13px; background: none; color: var(--text-primary); cursor: pointer; width: 100%; }
.h5-logout-btn { width: 100%; padding: 14px; border-radius: var(--radius-full); background: var(--negative-bg); color: var(--negative); font-size: 14px; font-weight: 600; border: none; cursor: pointer; }
.h5-login-btn { width: 100%; padding: 14px; border-radius: var(--radius-full); background: linear-gradient(135deg,var(--accent-gold),var(--accent-gold-dim)); color: #000; font-size: 14px; font-weight: 600; border: none; cursor: pointer; text-align: center; }

.toggle { position: relative; display: inline-block; width: 44px; height: 24px; flex-shrink: 0; }
.toggle input { opacity: 0; width: 0; height: 0; }
.toggle-slider { position: absolute; inset: 0; background: rgba(255,255,255,.1); border-radius: 24px; transition: .3s; cursor: pointer; }
.toggle-slider::before { content: ''; position: absolute; height: 18px; width: 18px; left: 3px; bottom: 3px; background: #fff; border-radius: 50%; transition: .3s; }
.toggle input:checked + .toggle-slider { background: var(--accent-gold); }
.toggle input:checked + .toggle-slider::before { transform: translateX(20px); }
</style>
