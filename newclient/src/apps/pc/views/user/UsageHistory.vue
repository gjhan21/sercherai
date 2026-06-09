<template>
  <div class="history-page">
    <section class="section fade-in-up">
      <div class="section-header"><h2 class="section-title">VIP 权益配额</h2></div>
      <div class="credits-grid">
        <div v-for="q in quotaList" :key="q.name" class="quota-card glass">
          <div class="quota-header">
            <span class="quota-name">{{ q.name }}</span>
            <span class="quota-reset" v-if="q.reset">重置: {{ q.reset }}</span>
          </div>
          <div class="quota-main">
            <span class="quota-remaining" :class="{ 'warning': q.remaining === 0 }">{{ q.remaining }}</span>
            <span class="quota-total">/ {{ q.total }}</span>
          </div>
          <div class="quota-bar-bg">
            <div class="quota-bar" :style="{width: q.percent + '%', background: q.barColor}"></div>
          </div>
        </div>
      </div>
    </section>

    <section class="section fade-in-up fade-in-up-delay-1">
      <div class="section-header"><h2 class="section-title">历史明细</h2></div>
      <div class="history-list">
        <div v-for="item in historyItems" :key="item.id" class="history-item glass" @click="goItem(item)">
          <div class="history-type-badge" :class="item.type">{{ item.typeLabel }}</div>
          <div class="history-body">
            <strong>{{ item.title }}</strong>
            <span class="history-meta">{{ (item.date || '').replace('T', ' ').slice(0, 19) }}</span>
          </div>
        </div>
        <div v-if="!historyItems.length" class="empty-tip">暂无使用记录</div>
      </div>
    </section>
  </div>
</template>

<script setup>
import { computed, onMounted, ref } from "vue";
import { useRouter } from "vue-router";
import { USER_PROFILE as MOCK } from "@/mock/user.js";
import { getMembershipQuota, listBrowseHistory } from "@/api/membership.js";
import { useClientAuth } from "@/shared/auth/client-auth";
const { isLoggedIn } = useClientAuth();

const router = useRouter();
const quota = ref(null);
const browseHistories = ref([]);

const quotaList = computed(() => {
  if (!quota.value) return [];
  const cycleText = {
    'MONTHLY': '按月',
    'WEEKLY': '按周',
    'DAILY': '按天'
  }[quota.value.reset_cycle] || '按周期';
  
  return [
    {
      name: '文档阅读',
      remaining: quota.value.doc_read_remaining ?? 0,
      total: quota.value.doc_read_limit ?? 0,
      percent: quota.value.doc_read_limit ? ((quota.value.doc_read_remaining / quota.value.doc_read_limit) * 100) : 0,
      reset: cycleText,
      barColor: 'linear-gradient(90deg, var(--accent-gold), var(--positive))'
    },
    {
      name: 'VIP资讯阅读',
      remaining: quota.value.news_subscribe_remaining ?? 0,
      total: quota.value.news_subscribe_limit ?? 0,
      percent: quota.value.news_subscribe_limit ? ((quota.value.news_subscribe_remaining / quota.value.news_subscribe_limit) * 100) : 0,
      reset: cycleText,
      barColor: 'linear-gradient(90deg, #3b82f6, #60a5fa)'
    },
    {
      name: '附件下载',
      remaining: quota.value.download_remaining ?? 0,
      total: quota.value.download_limit ?? 0,
      percent: quota.value.download_limit ? ((quota.value.download_remaining / quota.value.download_limit) * 100) : 0,
      reset: cycleText,
      barColor: 'linear-gradient(90deg, #ec4899, #f472b6)'
    },
    {
      name: '深度推演',
      remaining: quota.value.forecast_remaining ?? 0,
      total: quota.value.forecast_limit ?? 0,
      percent: quota.value.forecast_limit ? ((quota.value.forecast_remaining / quota.value.forecast_limit) * 100) : 0,
      reset: cycleText,
      barColor: 'linear-gradient(90deg, #8b5cf6, #a78bfa)'
    },
    {
      name: '股票分析',
      remaining: quota.value.stock_reco_remaining ?? 0,
      total: quota.value.stock_reco_limit ?? 0,
      percent: quota.value.stock_reco_limit ? ((quota.value.stock_reco_remaining / quota.value.stock_reco_limit) * 100) : 0,
      reset: cycleText,
      barColor: 'linear-gradient(90deg, #10b981, #34d399)'
    }
  ];
});

const historyItems = computed(() => {
  if (browseHistories.value.length) {
    return browseHistories.value.map(item => {
      let typeBadge = 'news';
      let typeLabel = '文档';
      if (item.content_type === 'NEWS') {
        typeBadge = 'news';
        typeLabel = '资讯';
      } else if (item.content_type === 'ATTACHMENT') {
        typeBadge = 'attachment';
        typeLabel = '下载';
      } else if (item.content_type === 'FORECAST') {
        typeBadge = 'forecast';
        typeLabel = '推演';
      } else if (item.content_type === 'STOCK') {
        typeBadge = 'stock';
        typeLabel = '分析';
      }
      return {
        id: item.id,
        title: item.title,
        type: typeBadge,
        typeLabel: typeLabel,
        date: item.viewed_at || '',
        stock: item.target_key || null,
        contentType: item.content_type,
        contentId: item.content_id
      };
    });
  }
  return [];
});

function goItem(item) {
  if (item.contentType === 'NEWS') {
    router.push('/news/' + item.contentId);
  } else if (item.contentType === 'STOCK' && item.stock) {
    router.push('/identify/' + item.stock);
  } else if (item.contentType === 'FORECAST' && item.contentId) {
    router.push('/forecast/' + item.contentId);
  } else if (item.contentType === 'ATTACHMENT') {
    // Open news article detail for download
    router.push('/news/' + item.contentId);
  } else {
    if (item.stock) {
      router.push('/identify/' + item.stock);
    }
  }
}

async function loadData() {
  if (!isLoggedIn.value) return;
  try {
    const [q, h] = await Promise.allSettled([
      getMembershipQuota(),
      listBrowseHistory({ page: 1, page_size: 100 })
    ]);
    if (q.status === 'fulfilled' && q.value) quota.value = q.value;
    if (h.status === 'fulfilled' && h.value?.items?.length) browseHistories.value = h.value.items;
  } catch { /* ignored */ }
}

onMounted(loadData);
</script>

<style scoped>
.history-page { display: grid; gap: 24px; max-width: 1000px; }
.section { background: var(--bg-card); border: 1px solid var(--border); border-radius: var(--radius-lg); padding: 24px; }
.section-header { margin-bottom: 20px; }
.section-title { font-size: 20px; font-weight: 700; color: var(--text-primary); }

.credits-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(180px, 1fr));
  gap: 16px;
  margin-bottom: 8px;
}

.quota-card {
  padding: 16px;
  border-radius: var(--radius-md);
  border: 1px solid var(--border);
  display: flex;
  flex-direction: column;
}

.quota-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 8px;
}

.quota-name {
  font-size: 13px;
  color: var(--text-secondary);
  font-weight: 600;
}

.quota-reset {
  font-size: 11px;
  color: var(--text-muted);
}

.quota-main {
  margin-bottom: 12px;
  display: flex;
  align-items: baseline;
}

.quota-remaining {
  font-size: 28px;
  font-weight: 800;
  color: var(--text-primary);
}

.quota-remaining.warning {
  color: var(--negative);
}

.quota-total {
  font-size: 14px;
  color: var(--text-secondary);
  margin-left: 4px;
}

.quota-bar-bg {
  height: 6px;
  background: rgba(255,255,255,.06);
  border-radius: 3px;
  overflow: hidden;
  margin-top: auto;
}

.quota-bar {
  height: 100%;
  border-radius: 3px;
  transition: width 0.3s ease;
}

.history-list { display: grid; gap: 10px; }
.history-item { display: flex; align-items: center; gap: 12px; padding: 12px 16px; border-radius: var(--radius-md); cursor: pointer; border: 1px solid var(--border); transition: all 0.2s ease; }
.history-item:hover { background: rgba(255,255,255,.02); border-color: var(--accent-gold); }

.history-type-badge { width: 44px; height: 22px; border-radius: var(--radius-sm); display: flex; align-items: center; justify-content: center; font-size: 11px; font-weight: 700; flex-shrink: 0; }
.history-type-badge.news { background: rgba(59,130,246,.1); color: var(--accent-blue); }
.history-type-badge.attachment { background: rgba(236,72,153,.1); color: #ec4899; }
.history-type-badge.forecast { background: rgba(139,92,246,.1); color: #a78bfa; }
.history-type-badge.stock { background: rgba(16,185,129,.1); color: #10b981; }

.history-body { flex: 1; }
.history-body strong { display: block; font-size: 14px; color: var(--text-primary); margin-bottom: 2px; }
.history-meta { font-size: 12px; color: var(--text-muted); }

.empty-tip {
  text-align: center;
  padding: 40px 0;
  color: var(--text-muted);
  font-size: 14px;
}
</style>

