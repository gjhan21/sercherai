<template>
  <div class="h5-vip">
    <div class="h5-vip-hero"><h2>VIP 会员</h2><p>解锁更多 AI 分析能力</p></div>
    <div class="h5-tier-scroll">
      <div v-for="tier in tierList" :key="tier.tier" class="h5-tier-card" :class="{ popular: tier.popular }">
        <div v-if="tier.popular" class="h5-pop-badge">推荐</div>
        <h3>{{ tier.name }}</h3>
        <div class="h5-tier-price">
          <span class="h5-tp-amount">{{ tier.price === 0 ? '免费' : '¥' + tier.price }}</span>
          <span v-if="tier.price > 0" class="h5-tp-period">/{{ tier.period }}</span>
        </div>
        <ul><li v-for="b in (tier.benefits || [])" :key="b">{{ b }}</li></ul>
        <button class="h5-tier-btn" :class="{ popular: tier.popular || tier.tier === userLevel.toLowerCase() }" @click="handleUpgrade(tier)">{{ loading ? '...' : tier.tier === userLevel.toLowerCase() ? '当前' : '升级' }}</button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { onMounted, ref } from "vue"
import { VIP_TIERS as MOCK_TIERS } from "@/mock/user.js"
import { listMembershipProducts, getMembershipQuota } from "@/api/membership.js"
import { useClientAuth } from "@/shared/auth/client-auth"
import { trackExposure, trackUpgradeIntent } from "@/shared/lib/experiment-tracker.js"

const { isLoggedIn } = useClientAuth();

const loading = ref(false);
const tierList = ref(MOCK_TIERS);
const userLevel = ref("FREE");

const levelMap = {
  'FREE': { benefits: ['每日 5 次 AI 分析', '基础市场数据'] },
  'VIP1': { benefits: ['每日 50 次 AI 分析', '进阶技术指标', 'Level-2 行情'] },
  'VIP2': { benefits: ['无限次 AI 分析', 'VIP 专属策略', 'Level-2 行情'] },
  'VIP3': { benefits: ['所有黄金权益', '私募级策略', '投顾服务'] },
  'VIP4': { benefits: ['所有铂金权益', '无限尊贵策略', '私人管家专属投顾'] }
};

async function loadProducts() {
  if (!isLoggedIn.value) return;
  loading.value = true;
  try {
    try {
      const quota = await getMembershipQuota();
      if (quota && quota.member_level) {
        userLevel.value = quota.member_level.toUpperCase();
      }
    } catch (e) {
      console.error("Failed to load user level:", e);
    }

    const result = await listMembershipProducts({ status: 'ACTIVE' });
    if (result?.items?.length) {
      // Sort result items by member level (VIP1 -> VIP2 -> VIP3)
      const sortedItems = [...result.items].sort((a, b) => {
        const lvA = (a.member_level || "").toUpperCase();
        const lvB = (b.member_level || "").toUpperCase();
        return lvA.localeCompare(lvB);
      });

      const apiTiers = sortedItems.map((p) => {
        const lvl = (p.member_level || "").toUpperCase();
        const cfg = levelMap[lvl] || { benefits: [p.description || (p.name + '权益')] };
        return {
          tier: lvl.toLowerCase(),
          name: p.name || lvl,
          price: p.price,
          period: p.duration_days >= 360 ? '年' : '月',
          popular: lvl === 'VIP2',
          benefits: p.description ? [p.description] : cfg.benefits
        };
      });
      apiTiers.unshift(MOCK_TIERS[0]);
      tierList.value = apiTiers;
    }
  } catch (e) {
    console.error("Failed to load products:", e);
  } finally {
    loading.value = false;
  }
}

function handleUpgrade(tier) {
  if (tier.tier === userLevel.value.toLowerCase()) {
    return;
  }
  if (tier.price > 0) {
    trackUpgradeIntent("membership", { metadata: { tier: tier.tier, target: "h5_upgrade_alert" } });
    alert('请前往电脑端网页版开通 ' + tier.name + ' 服务');
  }
}

onMounted(() => {
  loadProducts();
  trackExposure("membership");
});
</script>

<style scoped>
.h5-vip { display: grid; gap: 14px; }
.h5-vip-hero { text-align: center; padding: 24px 0; }
.h5-vip-hero h2 { font-size: 22px; font-weight: 800; }
.h5-vip-hero p { font-size: 13px; color: var(--text-secondary); }
.h5-tier-scroll { display: grid; gap: 12px; }
.h5-tier-card { padding: 20px; background: var(--bg-card); border: 1px solid var(--border); border-radius: var(--radius-lg); text-align: center; position: relative; }
.h5-tier-card.popular { border-color: var(--accent-gold); }
.h5-pop-badge { position: absolute; top: -8px; right: -8px; padding: 2px 8px; border-radius: var(--radius-full); background: var(--accent-gold); color: #000; font-size: 10px; font-weight: 700; }
.h5-tier-card h3 { font-size: 16px; font-weight: 700; margin-bottom: 8px; }
.h5-tier-price { margin-bottom: 12px; }
.h5-tp-amount { font-size: 24px; font-weight: 800; }
.h5-tp-period { font-size: 12px; color: var(--text-secondary); }
.h5-tier-card ul { list-style: none; padding: 0; margin: 0 0 14px; display: grid; gap: 6px; }
.h5-tier-card li { font-size: 12px; color: var(--text-secondary); }
.h5-tier-btn { width: 100%; padding: 10px; border-radius: var(--radius-full); border: 1px solid var(--border); font-size: 13px; font-weight: 600; background: none; color: var(--text-primary); cursor: pointer; }
.h5-tier-btn.popular { background: linear-gradient(135deg,var(--accent-gold),var(--accent-gold-dim)); color: #000; border: none; }
</style>
