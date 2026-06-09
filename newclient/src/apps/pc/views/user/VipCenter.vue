<template>
  <div class="vip-page">
    <section class="vip-hero fade-in-up">
      <h1 class="vip-hero-title">VIP 会员服务</h1>
      <p class="vip-hero-desc">选择最适合您的方案，解锁更强大的 AI 分析能力</p>
    </section>

    <section class="section fade-in-up fade-in-up-delay-1">
      <div class="tier-grid">
        <div v-for="tier in tierList" :key="tier.tier" class="tier-card glass" :class="{ popular: tier.popular }">
          <div v-if="tier.popular" class="popular-badge">推荐</div>
          <div class="tier-icon" :class="'icon-' + (tier.icon || tier.tier)">$</div>
          <h3 class="tier-name">{{ tier.name }}</h3>
          <div class="tier-price">
            <span class="price-amount">{{ tier.price === 0 ? '免费' : '¥' + tier.price }}</span>
            <span v-if="tier.price > 0" class="price-period">/{{ tier.period }}</span>
          </div>
          <ul class="tier-benefits">
            <li v-for="b in (tier.benefits || [])" :key="b">{{ b }}</li>
          </ul>
          <button class="tier-btn" :class="{ popular: tier.popular || tier.tier === userLevel.toLowerCase() }" @click="handleUpgrade(tier)">{{ loading ? '...' : tier.tier === userLevel.toLowerCase() ? '当前方案' : '升级方案' }}</button>
        </div>
      </div>
    </section>
  </div>
</template>

<script setup>
import { onMounted, ref } from "vue"
import { useRouter } from "vue-router"
import { VIP_TIERS as MOCK_TIERS } from "@/mock/user.js"
import { listMembershipProducts, getMembershipQuota } from "@/api/membership.js"
import { useClientAuth } from "@/shared/auth/client-auth"
import { trackExposure, trackClick } from "@/shared/lib/experiment-tracker.js"

const router = useRouter();
const { isLoggedIn } = useClientAuth();

const loading = ref(false);
const tierList = ref(MOCK_TIERS);
const userLevel = ref("FREE");

const levelMap = {
  'FREE': { icon: 'free', benefits: ['每日 5 次 AI 分析', '基础市场数据', '基础行情浏览'] },
  'VIP1': { icon: 'silver', benefits: ['每日 50 次 AI 分析', '进阶技术指标', 'Level-2 行情', 'VIP 交流圈'] },
  'VIP2': { icon: 'gold', benefits: ['无限次 AI 分析', 'VIP 专属策略', 'Level-2 行情', '实时盯盘提醒'] },
  'VIP3': { icon: 'platinum', benefits: ['所有黄金权益', '私募级策略', '一对一投顾服务', '优先体验新功能'] },
  'VIP4': { icon: 'diamond', benefits: ['所有铂金权益', '无限尊贵策略', '私人投顾顾问', '独立分析服务器'] }
};

async function loadProducts() {
  if (!isLoggedIn.value) return;
  loading.value = true;
  try {
    // Load current user level
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
        const cfg = levelMap[lvl] || { icon: 'silver', benefits: [p.description || (p.name + '权益')] };
        return {
          id: p.id,
          tier: lvl.toLowerCase(),
          name: p.name || lvl,
          price: p.price,
          period: p.duration_days >= 360 ? '年' : p.duration_days >= 85 ? '季' : '月',
          popular: lvl === 'VIP2',
          icon: cfg.icon,
          benefits: p.description ? [p.description, ...cfg.benefits.slice(1)] : cfg.benefits
        };
      });
      apiTiers.unshift(MOCK_TIERS[0]); // Keep free tier first
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
    trackClick("membership", "upgrade_" + tier.tier);
    router.push({ path: '/user/vip/purchase', query: { tier: tier.tier } });
  }
}

onMounted(() => {
  loadProducts();
  trackExposure("membership");
});
</script>

<style scoped>
.vip-page { display: grid; gap: 20px; max-width: 1200px; }
.vip-hero { text-align: center; padding: 40px 20px; }
.vip-hero-title { font-size: 32px; font-weight: 800; margin-bottom: 10px; }
.vip-hero-desc { font-size: 14px; color: var(--text-secondary); }
.section { background: var(--bg-card); border: 1px solid var(--border); border-radius: var(--radius-lg); padding: 24px; }
.tier-grid { display: grid; grid-template-columns: repeat(4,1fr); gap: 12px; }
.tier-card { padding: 24px 16px; border-radius: var(--radius-lg); text-align: center; position: relative; border: 1px solid var(--border); }
.tier-card.popular { border-color: var(--accent-gold); box-shadow: 0 0 20px rgba(240,185,11,.15); transform: scale(1.03); }
.popular-badge { position: absolute; top: -10px; right: -10px; background: var(--accent-gold); color: #000; padding: 3px 10px; border-radius: var(--radius-full); font-size: 11px; font-weight: 700; }
.tier-icon { width: 48px; height: 48px; border-radius: 50%; display: flex; align-items: center; justify-content: center; font-size: 20px; font-weight: 700; margin: 0 auto 12px; }
.icon-free { background: rgba(139,147,176,.1); color: var(--text-secondary); }
.icon-silver { background: rgba(192,192,192,.15); color: #c0c0c0; }
.icon-gold { background: var(--accent-gold-glow); color: var(--accent-gold); }
.icon-platinum { background: rgba(139,92,246,.1); color: #8b5cf6; }
.icon-diamond { background: rgba(244,63,94,.1); color: #f43f5e; }
.tier-name { font-size: 18px; font-weight: 700; margin-bottom: 8px; }
.tier-price { margin-bottom: 16px; }
.price-amount { font-size: 28px; font-weight: 800; }
.price-period { font-size: 14px; color: var(--text-secondary); }
.tier-benefits { list-style: none; padding: 0; margin: 0 0 16px; display: grid; gap: 8px; }
.tier-benefits li { font-size: 13px; color: var(--text-secondary); padding: 4px 0; border-bottom: 1px solid rgba(255,255,255,.03); }
.tier-btn { width: 100%; padding: 10px; border-radius: var(--radius-full); font-size: 13px; font-weight: 600; border: 1px solid var(--border); color: var(--text-secondary); background: none; cursor: pointer; }
.tier-btn.popular { background: linear-gradient(135deg,var(--accent-gold),var(--accent-gold-dim)); color: #000; border: none; }
@media (max-width: 1000px) { .tier-grid { grid-template-columns: repeat(2,1fr); } }
</style>
