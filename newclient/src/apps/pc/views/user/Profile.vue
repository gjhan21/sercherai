<template>
  <div class="profile-page">
    <section class="user-card glass fade-in-up">
      <div class="user-avatar">{{ (profile?.nickname || '股票投资者')[0] }}</div>
      <div class="user-info">
        <h2>{{ profile?.nickname || '股票投资者' }}</h2>
        <div class="user-meta">
          <span class="user-level">{{ profile?.level || 'Lv.1' }}</span>
          <span class="user-vip" v-if="quota">{{ quota.member_level || '免费版' }}</span>
          <span v-if="quota">分析剩余 {{ quota.doc_read_remaining || '-' }}/{{ quota.doc_read_limit || '-' }}</span>
        </div>
      </div>
      <div class="user-actions">
        <button class="btn-primary" @click="$router.push('/user/vip')">VIP 中心</button>
      </div>
    </section>

    <!-- Portfolio Summary -->
    <section class="section fade-in-up fade-in-up-delay-1">
      <div class="section-header"><h2 class="section-title">投资组合</h2></div>
      <div class="portfolio-summary">
        <div class="portfolio-total glass">
          <span class="pt-label">总资产 (模拟)</span>
          <span class="pt-value">¥{{ (portfolioTotal).toLocaleString() }}</span>
        </div>
        <div class="portfolio-empty"><p>API 未返回持仓数据时显示模拟数据</p></div>
        <div class="portfolio-holdings">
          <div v-for="h in holdings" :key="h.symbol" class="holding-item" @click="$router.push('/identify/' + h.symbol)">
            <div class="holding-left">
              <span class="holding-name">{{ h.name }}</span>
              <span class="holding-shares">{{ h.shares || '--' }} 股</span>
            </div>
            <div class="holding-right">
              <span class="holding-value">¥{{ (h.marketValue || 0).toLocaleString() }}</span>
              <span class="holding-profit" :class="(h.profitPct || 0) >= 0 ? 'up' : 'down'">{{ h.profitPct >= 0 ? '+' : '' }}{{ h.profitPct }}%</span>
            </div>
          </div>
        </div>
      </div>
    </section>

    <!-- Menu -->
    <section class="section fade-in-up fade-in-up-delay-2">
      <div class="profile-menu">
        <button class="menu-item" @click="$router.push('/user/vip')">
          <span class="menu-icon" style="color:var(--accent-gold)">👑</span><span>VIP 会员</span><span class="menu-arrow">→</span>
        </button>
        <button class="menu-item" @click="$router.push('/user/history')">
          <span class="menu-icon" style="color:var(--accent-cyan)">📊</span><span>使用记录</span><span class="menu-arrow">→</span>
        </button>
        <button class="menu-item" @click="$router.push('/user/settings')">
          <span class="menu-icon" style="color:var(--text-secondary)">⚙️</span><span>设置</span><span class="menu-arrow">→</span>
        </button>
      </div>
    </section>
  </div>
</template>

<script setup>
import { computed, onMounted, ref } from "vue"
import { useRouter } from "vue-router"
import { USER_PROFILE as MOCK_PROFILE } from "@/mock/user.js"
import { getUserProfile } from "@/api/membership.js"
import { getMembershipQuota } from "@/api/membership.js"
import { useClientAuth } from "@/shared/auth/client-auth"



const profile = ref(MOCK_PROFILE);
const quota = ref(null);
const portfolioTotal = ref(MOCK_PROFILE.portfolio.totalAssets);
const holdings = ref(MOCK_PROFILE.portfolio.holdings);

async function loadProfile() {
  if (!isLoggedIn.value) return;
  try {
    const [userResult, quotaResult] = await Promise.allSettled([getUserProfile(), getMembershipQuota()]);
    if (userResult.status === 'fulfilled' && userResult.value) {
      profile.value = { nickname: userResult.value.phone || userResult.value.email || '用户', level: 'Lv.1', ...userResult.value };
    }
    if (quotaResult.status === 'fulfilled' && quotaResult.value) {
      quota.value = quotaResult.value;
    }
  } catch { /* use mock */ }
}

onMounted(loadProfile);
</script>

<style scoped>
.profile-page { display: grid; gap: 16px; max-width: 1000px; }
.user-card { display: flex; align-items: center; gap: 16px; padding: 20px; border-radius: var(--radius-lg); }
.user-avatar { width: 56px; height: 56px; border-radius: 50%; background: var(--accent-gold-glow); color: var(--accent-gold); display: flex; align-items: center; justify-content: center; font-size: 24px; font-weight: 700; }
.user-info { flex: 1; }
.user-info h2 { font-size: 18px; font-weight: 700; margin-bottom: 4px; }
.user-meta { display: flex; gap: 10px; font-size: 12px; color: var(--text-secondary); flex-wrap: wrap; }
.user-level { color: var(--accent-gold); }
.user-vip { color: var(--positive); }
.btn-primary { padding: 8px 18px; border-radius: var(--radius-full); background: linear-gradient(135deg,var(--accent-gold),var(--accent-gold-dim)); color: #000; font-size: 13px; font-weight: 600; border: none; cursor: pointer; }
.section { background: var(--bg-card); border: 1px solid var(--border); border-radius: var(--radius-lg); padding: 24px; }
.section-header { margin-bottom: 16px; }
.section-title { font-size: 20px; font-weight: 700; }
.portfolio-total { padding: 16px; border-radius: var(--radius-md); margin-bottom: 12px; }
.pt-label { display: block; font-size: 12px; color: var(--text-secondary); margin-bottom: 4px; }
.pt-value { display: block; font-size: 28px; font-weight: 800; margin-bottom: 4px; }
.portfolio-holdings { display: grid; gap: 8px; }
.holding-item { display: flex; justify-content: space-between; padding: 10px; border-radius: var(--radius-sm); border: 1px solid var(--border); cursor: pointer; }
.holding-item:hover { background: rgba(255,255,255,.03); }
.holding-name { display: block; font-weight: 600; font-size: 14px; }
.holding-shares { font-size: 11px; color: var(--text-secondary); }
.holding-value { display: block; text-align: right; font-weight: 600; font-size: 14px; }
.holding-profit { font-size: 13px; font-weight: 600; }
.up { color: var(--positive); }
.down { color: var(--negative); }
.profile-menu { display: grid; gap: 4px; }
.menu-item { display: flex; align-items: center; gap: 12px; padding: 14px; border-radius: var(--radius-sm); width: 100%; text-align: left; font-size: 14px; }
.menu-item:hover { background: rgba(255,255,255,.03); }
.menu-icon { font-size: 18px; width: 24px; }
.menu-arrow { margin-left: auto; color: var(--text-muted); }
</style>
