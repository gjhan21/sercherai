<template>
  <div class="history-page">
    <section class="section fade-in-up">
      <div class="section-header"><h2 class="section-title">使用记录</h2></div>
      <div class="credits-card glass">
        <div class="credits-main">
          <span class="credits-label">AI 分析次数</span>
          <span class="credits-remaining">{{ remainingCredits }}</span>
          <span class="credits-total">/ {{ totalCredits }}</span>
          <span class="credits-reset">下次重置：{{ resetDate }}</span>
        </div>
        <div class="credits-bar-bg">
          <div class="credits-bar" :style="{width: usedPercent + '%'}"></div>
        </div>
      </div>

      <div class="usage-chart" v-if="usageData.length">
        <div v-for="day in usageData" :key="day.day" class="usage-bar-wrap">
          <div class="usage-bar" :style="{height: (day.used * 15) + 'px'}"></div>
          <span class="usage-label">{{ day.day.slice(3) }}</span>
          <span class="usage-value">{{ day.used }}</span>
        </div>
      </div>
    </section>

    <section class="section fade-in-up fade-in-up-delay-1">
      <div class="section-header"><h2 class="section-title">历史明细</h2></div>
      <div class="history-list">
        <div v-for="item in historyItems" :key="item.id" class="history-item glass" @click="goStock(item.stock)">
          <div class="history-type-badge" :class="item.type || 'analysis'">{{ (item.type === 'chat' ? '聊' : 'AI') }}</div>
          <div class="history-body">
            <strong>{{ item.title }}</strong>
            <span class="history-meta">{{ (item.date || item.created_at || '').replace('T', ' ') }} · 消耗 {{ item.creditsUsed || 1 }} 次</span>
          </div>
        </div>
      </div>
    </section>
  </div>
</template>

<script setup>
import { computed, onMounted, ref } from "vue";
import { useRouter } from "vue-router";
import { USER_PROFILE as MOCK, CREDIT_USAGE_BY_DAY } from "@/mock/user.js";
import { getMembershipQuota, listMembershipOrders } from "@/api/membership.js";
import { useClientAuth } from "@/shared/auth/client-auth";
const { isLoggedIn } = useClientAuth();

const router = useRouter();
const quota = ref(null);
const orders = ref([]);
const usageData = ref(CREDIT_USAGE_BY_DAY);

const remainingCredits = computed(() => quota.value?.doc_read_remaining ?? MOCK.credits.remaining);
const totalCredits = computed(() => quota.value?.doc_read_limit ?? MOCK.credits.total);
const resetDate = computed(() => quota.value?.reset_cycle || MOCK.credits.nextReset);
const usedPercent = computed(() => (remainingCredits.value / totalCredits.value) * 100);

const historyItems = computed(() => {
  if (orders.value.length) {
    return orders.value.map(o => ({
      id: o.id, title: '会员订单', type: 'analysis',
      date: o.created_at || '', creditsUsed: 1, stock: null
    }));
  }
  return MOCK.history;
});

function goStock(symbol) { if (symbol) router.push('/identify/' + symbol); }

async function loadData() {
  if (!isLoggedIn.value) return;
  try {
    const [q, o] = await Promise.allSettled([getMembershipQuota(), listMembershipOrders({ page: 1, page_size: 20 })]);
    if (q.status === 'fulfilled' && q.value) quota.value = q.value;
    if (o.status === 'fulfilled' && o.value?.items?.length) orders.value = o.value.items;
  } catch { /* use mock */ }
}

onMounted(loadData);
</script>

<style scoped>
.history-page { display: grid; gap: 16px; max-width: 1000px; }
.section { background: var(--bg-card); border: 1px solid var(--border); border-radius: var(--radius-lg); padding: 24px; }
.section-header { margin-bottom: 16px; }
.section-title { font-size: 20px; font-weight: 700; }
.credits-card { padding: 20px; border-radius: var(--radius-md); margin-bottom: 20px; }
.credits-main { margin-bottom: 12px; }
.credits-label { display: block; font-size: 12px; color: var(--text-secondary); margin-bottom: 4px; }
.credits-remaining { font-size: 42px; font-weight: 800; color: var(--accent-gold); }
.credits-total { font-size: 18px; color: var(--text-secondary); }
.credits-reset { display: block; font-size: 12px; color: var(--text-muted); margin-top: 4px; }
.credits-bar-bg { height: 8px; background: rgba(255,255,255,.06); border-radius: 4px; overflow: hidden; }
.credits-bar { height: 100%; border-radius: 4px; background: linear-gradient(90deg,var(--accent-gold),var(--positive)); }
.usage-chart { display: flex; align-items: flex-end; gap: 12px; height: 120px; padding: 12px 0; }
.usage-bar-wrap { flex: 1; display: flex; flex-direction: column; align-items: center; }
.usage-bar { width: 20px; background: linear-gradient(to top,var(--accent-gold),var(--positive)); border-radius: 4px 4px 0 0; min-height: 4px; }
.usage-label { font-size: 10px; color: var(--text-muted); margin-top: 6px; }
.usage-value { font-size: 11px; color: var(--text-secondary); margin-top: 2px; }
.history-list { display: grid; gap: 8px; }
.history-item { display: flex; align-items: center; gap: 12px; padding: 12px; border-radius: var(--radius-sm); cursor: pointer; }
.history-item:hover { background: rgba(255,255,255,.02); }
.history-type-badge { width: 32px; height: 32px; border-radius: 50%; display: flex; align-items: center; justify-content: center; font-size: 11px; font-weight: 700; flex-shrink: 0; }
.history-type-badge.analysis { background: var(--accent-gold-glow); color: var(--accent-gold); }
.history-type-badge.chat { background: rgba(59,130,246,.1); color: var(--accent-blue); }
.history-body { flex: 1; }
.history-body strong { display: block; font-size: 13px; }
.history-meta { font-size: 11px; color: var(--text-muted); }
</style>
