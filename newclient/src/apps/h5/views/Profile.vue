<template>
  <div class="h5-profile">
    <div class="h5-profile-card">
      <div class="h5-avatar">{{ isLoggedIn ? (profile?.nickname || '用')[0] : '客' }}</div>
      <div class="h5-profile-info">
        <strong>{{ isLoggedIn ? (profile?.nickname || '股票投资者') : '未登录 / 游客' }}</strong>
        <span v-if="isLoggedIn">{{ quota?.member_level || '免费版' }} · AI 剩余 {{ quota?.doc_read_remaining || '--' }} 次</span>
        <span v-else>登录后体验完整 AI 投资助手</span>
      </div>
      <button class="h5-profile-btn" v-if="!isLoggedIn" @click="$router.push('/login')">登录</button>
    </div>

    <div class="h5-quick-menu">
      <button class="h5-menu-item" @click="$router.push('/profile/vip')"><span style="color:var(--accent-gold)">👑</span><span>VIP 会员</span><span>→</span></button>
      <button class="h5-menu-item" @click="$router.push('/profile/settings')"><span>⚙️</span><span>设置</span><span>→</span></button>
      <button class="h5-menu-item" @click="$router.push('/ai-chat')"><span style="color:var(--accent-gold)">💬</span><span>AI 对话</span><span>→</span></button>
    </div>

    <div class="h5-section" v-if="isLoggedIn">
      <div class="h5-section-header"><h3>持仓概览</h3></div>
      <div class="h5-holdings">
        <div v-for="h in profile?.portfolio?.holdings || mockHoldings" :key="h.symbol" class="h5-holding-item" @click="$router.push('/markets/' + h.symbol)">
          <span class="h5-holding-name">{{ h.name }}</span>
          <span class="h5-holding-value">¥{{ (h.marketValue || 0).toLocaleString() }}</span>
          <span class="h5-holding-profit" :class="(h.profitPct || 0) >= 0 ? 'up' : 'down'">{{ h.profitPct >= 0 ? '+' : '' }}{{ h.profitPct }}%</span>
        </div>
      </div>
    </div>
    <div class="h5-portfolio-empty-login" v-else>
      <p>请登录后查看您的持仓数据与资产总览</p>
      <button class="h5-login-btn-sm" @click="$router.push('/login')">立即登录</button>
    </div>
  </div>
</template>

<script setup>
import { onMounted, ref } from "vue"
import { USER_PROFILE as MOCK } from "@/mock/user.js"
import { getUserProfile, getMembershipQuota } from "@/api/membership.js"
import { useClientAuth } from "@/shared/auth/client-auth"

const { isLoggedIn } = useClientAuth();

const profile = ref(MOCK);
const quota = ref(null);
const mockHoldings = MOCK.portfolio?.holdings || [];

async function loadProfile() {
  if (!isLoggedIn.value) return;
  try {
    const [u, q] = await Promise.allSettled([getUserProfile(), getMembershipQuota()]);
    if (u.status === 'fulfilled' && u.value) {
      profile.value = { ...profile.value, nickname: u.value.phone || u.value.email || '用户' };
    }
    if (q.status === 'fulfilled' && q.value) {
      quota.value = q.value;
    }
  } catch { /* use mock */ }
}

onMounted(loadProfile);
</script>

<style scoped>
.h5-profile { display: grid; gap: 14px; }
.h5-profile-card { display: flex; align-items: center; gap: 12px; padding: 16px; background: var(--bg-card); border: 1px solid var(--border); border-radius: var(--radius-lg); }
.h5-avatar { width: 44px; height: 44px; border-radius: 50%; background: var(--accent-gold-glow); color: var(--accent-gold); display: flex; align-items: center; justify-content: center; font-size: 18px; font-weight: 700; }
.h5-profile-info { flex: 1; }
.h5-profile-info strong { display: block; font-size: 15px; }
.h5-profile-info span { font-size: 11px; color: var(--text-secondary); }
.h5-profile-btn { padding: 8px 18px; border-radius: var(--radius-full); background: linear-gradient(135deg,var(--accent-gold),var(--accent-gold-dim)); color: #000; font-size: 13px; font-weight: 600; border: none; cursor: pointer; }
.h5-quick-menu { display: grid; gap: 4px; }
.h5-menu-item { display: flex; align-items: center; gap: 10px; padding: 14px; background: var(--bg-card); border: 1px solid var(--border); border-radius: var(--radius-md); font-size: 14px; width: 100%; text-align: left; cursor: pointer; }
.h5-menu-item:active { background: rgba(255,255,255,.03); }
.h5-menu-item span:last-child { margin-left: auto; color: var(--text-muted); }
.h5-section { background: var(--bg-card); border: 1px solid var(--border); border-radius: var(--radius-lg); padding: 16px; }
.h5-section-header { margin-bottom: 10px; }
.h5-section-header h3 { font-size: 14px; font-weight: 700; }
.h5-holdings { display: grid; gap: 8px; }
.h5-holding-item { display: flex; align-items: center; gap: 8px; padding: 10px; border-radius: var(--radius-sm); border: 1px solid var(--border); cursor: pointer; }
.h5-holding-item:active { background: rgba(255,255,255,.03); }
.h5-holding-name { flex: 1; font-size: 13px; font-weight: 600; }
.h5-holding-value { font-size: 13px; font-weight: 600; }
.h5-holding-profit { font-size: 12px; font-weight: 600; }
.up { color: var(--positive); }
.down { color: var(--negative); }

.h5-portfolio-empty-login {
  background: var(--bg-card);
  border: 1px dashed var(--border);
  border-radius: var(--radius-lg);
  padding: 30px 16px;
  text-align: center;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
}
.h5-portfolio-empty-login p {
  color: var(--text-muted);
  font-size: 13px;
  margin-bottom: 14px;
}
.h5-login-btn-sm {
  padding: 8px 20px;
  border-radius: var(--radius-full);
  background: linear-gradient(135deg,var(--accent-gold),var(--accent-gold-dim));
  color: #000;
  font-size: 12px;
  font-weight: 600;
  border: none;
  cursor: pointer;
}
</style>
