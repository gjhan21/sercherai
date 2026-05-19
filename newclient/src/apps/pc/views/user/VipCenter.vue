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
          <button class="tier-btn" :class="{ popular: tier.popular }" @click="handleUpgrade(tier)">{{ loading ? '...' : tier.price === 0 ? '当前方案' : '升级方案' }}</button>
        </div>
      </div>
    </section>
  </div>
</template>

<script setup>
import { onMounted, ref } from "vue"
import { VIP_TIERS as MOCK_TIERS } from "@/mock/user.js"
import { listMembershipProducts } from "@/api/membership.js"
import { useClientAuth } from "@/shared/auth/client-auth"
const { isLoggedIn } = useClientAuth();



const loading = ref(false);
const tierList = ref(MOCK_TIERS);

async function loadProducts() {
  if (!isLoggedIn.value) return;
  loading.value = true;
  try {
    const result = await listMembershipProducts({ status: 'ACTIVE' });
    if (result?.items?.length) {
      const apiTiers = result.items.map((p, i) => ({
        tier: (p.member_level || 'tier' + i).toLowerCase(),
        name: p.name, price: p.price, period: p.duration_days >= 360 ? '年' : p.duration_days >= 85 ? '季' : '月',
        popular: i === 1, icon: ['free','silver','gold','platinum'][i] || 'silver',
        benefits: p.description ? [p.description, 'AI 智能分析', '深度数据'] : [p.name + '权益']
      }));
      apiTiers.unshift(MOCK_TIERS[0]);
      tierList.value = apiTiers;
    }
  } catch { /* use mock */ }
  finally { loading.value = false; }
}

function handleUpgrade(tier) {
  if (tier.price > 0) alert('升级 ' + tier.name + ': ¥' + tier.price + '/' + tier.period);
}

onMounted(loadProducts);
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
