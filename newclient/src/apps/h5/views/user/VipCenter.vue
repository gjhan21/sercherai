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
        <button class="h5-tier-btn">{{ loading ? '...' : tier.price === 0 ? '当前' : '升级' }}</button>
      </div>
    </div>
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
        tier: (p.member_level || 't' + i).toLowerCase(), name: p.name,
        price: p.price, period: p.duration_days >= 360 ? '年' : '月',
        popular: i === 1,
        benefits: p.description ? [p.description] : [p.name + '权益']
      }));
      apiTiers.unshift(MOCK_TIERS[0]);
      tierList.value = apiTiers;
    }
  } catch { /* use mock */ }
  finally { loading.value = false; }
}

onMounted(loadProducts);
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
</style>
