<template>
  <div class="settings-page">
    <section class="section fade-in-up">
      <h2 class="section-title">设置</h2>

      <!-- Profile Info -->
      <div class="settings-group">
        <h3 class="settings-group-title">个人信息</h3>
        <div class="profile-info-card glass" v-if="session">
          <div class="profile-avatar">{{ (session.phone || session.email || '用')[0] }}</div>
          <div class="profile-details">
            <strong>{{ session.phone ? session.phone.slice(0,3)+'****'+session.phone.slice(-4) : session.email || '未登录' }}</strong>
            <span>ID: {{ session.userID || '--' }}</span>
          </div>
          <span class="profile-role" v-if="session.role === 'ADMIN'">管理员</span>
        </div>
      </div>

      <!-- Notifications -->
      <div class="settings-group">
        <h3 class="settings-group-title">通知设置</h3>
        <div class="settings-row">
          <div><span>股价异动提醒</span><p class="settings-desc">当关注股票出现异动时推送通知</p></div>
          <label class="toggle"><input type="checkbox" v-model="settings.priceAlert" /><span class="toggle-slider"></span></label>
        </div>
        <div class="settings-row">
          <div><span>资讯推送</span><p class="settings-desc">每日 AI 精选资讯推送</p></div>
          <label class="toggle"><input type="checkbox" v-model="settings.newsAlert" /><span class="toggle-slider"></span></label>
        </div>
        <div class="settings-row">
          <div><span>系统通知</span><p class="settings-desc">系统公告和账户变动通知</p></div>
          <label class="toggle"><input type="checkbox" v-model="settings.systemNotice" /><span class="toggle-slider"></span></label>
        </div>
      </div>

      <!-- Display -->
      <div class="settings-group">
        <h3 class="settings-group-title">显示设置</h3>
        <div class="settings-row">
          <span>主题</span>
          <span class="settings-value">深色模式 <span class="tag tag-gold">当前</span></span>
        </div>
        <div class="settings-row">
          <span>语言</span>
          <span class="settings-value">简体中文</span>
        </div>
      </div>

      <!-- Change Password -->
      <div class="settings-group" v-if="isLoggedIn">
        <h3 class="settings-group-title">修改密码</h3>
        <div class="password-form">
          <div class="settings-field">
            <label>当前密码</label>
            <input type="password" v-model="pwdForm.oldPwd" placeholder="输入当前密码" />
          </div>
          <div class="settings-field">
            <label>新密码</label>
            <input type="password" v-model="pwdForm.newPwd" placeholder="至少6位" />
          </div>
          <div class="settings-field">
            <label>确认新密码</label>
            <input type="password" v-model="pwdForm.confirmPwd" placeholder="再次输入新密码" />
          </div>
          <p v-if="pwdError" class="pwd-error">{{ pwdError }}</p>
          <p v-if="pwdSuccess" class="pwd-success">{{ pwdSuccess }}</p>
          <button class="btn-secondary" @click="handleChangePassword" :disabled="pwdLoading">{{ pwdLoading ? '提交中...' : '修改密码' }}</button>
        </div>
      </div>

      <!-- About -->
      <div class="settings-group">
        <h3 class="settings-group-title">关于</h3>
        <div class="settings-row"><span>版本</span><span class="settings-value">v0.1.0</span></div>
        <div class="settings-row"><span>AI 模型</span><span class="settings-value">多因子量化 + LLM</span></div>
      </div>

      <!-- Account Actions -->
      <div class="settings-group">
        <h3 class="settings-group-title">账户</h3>
        <div class="settings-actions">
          <button v-if="isLoggedIn" class="danger-btn" @click="handleLogout" :disabled="loggingOut">
            {{ loggingOut ? '退出中...' : '退出登录' }}
          </button>
          <button v-else class="btn-primary" @click="$router.push('/login')">登录 / 注册</button>
        </div>
      </div>
    </section>
  </div>
</template>

<script setup>
import { onMounted, ref, watch } from "vue";
import { useRouter } from "vue-router";
import { useClientAuth } from "@/shared/auth/client-auth";
import { logout, changePassword } from "@/api/auth.js";

const router = useRouter();
const { session, isLoggedIn } = useClientAuth();
const loggingOut = ref(false);

const pwdForm = ref({ oldPwd: "", newPwd: "", confirmPwd: "" });
const pwdLoading = ref(false);
const pwdError = ref("");
const pwdSuccess = ref("");

async function handleChangePassword() {
  pwdError.value = "";
  pwdSuccess.value = "";
  if (!pwdForm.value.oldPwd || !pwdForm.value.newPwd) { pwdError.value = "请填写完整"; return; }
  if (pwdForm.value.newPwd.length < 6) { pwdError.value = "新密码至少6位"; return; }
  if (pwdForm.value.newPwd !== pwdForm.value.confirmPwd) { pwdError.value = "两次密码不一致"; return; }
  pwdLoading.value = true;
  try {
    await changePassword({ old_password: pwdForm.value.oldPwd, new_password: pwdForm.value.newPwd });
    pwdSuccess.value = "密码修改成功！";
    pwdForm.value = { oldPwd: "", newPwd: "", confirmPwd: "" };
  } catch (e) {
    pwdError.value = e?.message || "修改失败，请重试";
  } finally {
    pwdLoading.value = false;
  }
}

const settings = ref({
  priceAlert: true,
  newsAlert: true,
  systemNotice: false
});

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

onMounted(() => {
  // Load saved settings from localStorage
  try {
    const saved = localStorage.getItem("ai_settings");
    if (saved) settings.value = { ...settings.value, ...JSON.parse(saved) };
  } catch { /* ignore */ }
});

// Save settings on change
watch(settings, (val) => {
  try { localStorage.setItem("ai_settings", JSON.stringify(val)); } catch { /* ignore */ }
}, { deep: true });
</script>

<style scoped>
.settings-page { max-width: 700px; }
.section { background: var(--bg-card); border: 1px solid var(--border); border-radius: var(--radius-lg); padding: 24px; }
.section-title { font-size: 20px; font-weight: 700; margin-bottom: 20px; }
.settings-group { margin-bottom: 24px; }
.settings-group-title { font-size: 13px; color: var(--text-muted); text-transform: uppercase; letter-spacing: .5px; margin-bottom: 12px; padding-bottom: 8px; border-bottom: 1px solid var(--border); }
.settings-row { display: flex; align-items: center; justify-content: space-between; padding: 14px 0; border-bottom: 1px solid rgba(255,255,255,.03); font-size: 14px; gap: 12px; }
.settings-row span { flex-shrink: 0; }
.settings-desc { font-size: 12px; color: var(--text-muted); margin-top: 2px; }
.settings-value { color: var(--text-secondary); font-size: 13px; display: flex; align-items: center; gap: 6px; }

/* Profile card */
.profile-info-card { display: flex; align-items: center; gap: 14px; padding: 16px; border-radius: var(--radius-md); }
.profile-avatar { width: 44px; height: 44px; border-radius: 50%; background: var(--accent-gold-glow); color: var(--accent-gold); display: flex; align-items: center; justify-content: center; font-size: 18px; font-weight: 700; }
.profile-details { flex: 1; }
.profile-details strong { display: block; font-size: 14px; }
.profile-details span { font-size: 12px; color: var(--text-muted); }
.profile-role { font-size: 11px; padding: 3px 8px; border-radius: var(--radius-full); background: var(--accent-gold-glow); color: var(--accent-gold); }

/* Actions */
.settings-actions { display: flex; gap: 10px; }
.danger-btn { padding: 10px 24px; border-radius: var(--radius-full); background: var(--negative-bg); color: var(--negative); font-size: 13px; font-weight: 600; cursor: pointer; border: none; }
.danger-btn:disabled { opacity: .6; }
.btn-primary { padding: 10px 24px; border-radius: var(--radius-full); background: linear-gradient(135deg,var(--accent-gold),var(--accent-gold-dim)); color: #000; font-size: 13px; font-weight: 600; cursor: pointer; border: none; text-decoration: none; }

/* Toggle */
/* Password Form */
.password-form { display: grid; gap: 12px; }
.settings-field { display: grid; gap: 4px; }
.settings-field label { font-size: 12px; color: var(--text-secondary); }
.settings-field input { padding: 10px 12px; border-radius: var(--radius-sm); background: rgba(255,255,255,.04); border: 1px solid var(--border); color: var(--text-primary); font-size: 13px; }
.settings-field input:focus { border-color: var(--accent-gold); }
.pwd-error { color: var(--negative); font-size: 12px; }
.pwd-success { color: var(--positive); font-size: 12px; }
.btn-secondary { padding: 10px 20px; border-radius: var(--radius-full); border: 1px solid var(--border); font-size: 13px; color: var(--text-primary); background: none; cursor: pointer; width: auto; }
.btn-secondary:hover { border-color: var(--border-light); }
.btn-secondary:disabled { opacity: .5; }

.toggle { position: relative; display: inline-block; width: 44px; height: 24px; flex-shrink: 0; }
.toggle input { opacity: 0; width: 0; height: 0; }
.toggle-slider { position: absolute; inset: 0; background: rgba(255,255,255,.1); border-radius: 24px; transition: .3s; cursor: pointer; }
.toggle-slider::before { content: ''; position: absolute; height: 18px; width: 18px; left: 3px; bottom: 3px; background: #fff; border-radius: 50%; transition: .3s; }
.toggle input:checked + .toggle-slider { background: var(--accent-gold); }
.toggle input:checked + .toggle-slider::before { transform: translateX(20px); }
</style>
