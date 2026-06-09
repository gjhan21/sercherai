<template>
  <div class="purchase-page">
    <section class="section fade-in-up">
      <h2 class="section-title">开通会员</h2>
      <p class="section-subtitle">{{ loadingProducts ? '加载中...' : '三步完成升级' }}</p>
      <div class="wizard-steps">
        <div class="wizard-step" :class="{ active: wizardStep === 1, done: wizardStep > 1 }"><span class="ws-num">1</span><span>选择方案</span></div>
        <div class="wizard-step" :class="{ active: wizardStep === 2, done: wizardStep > 2 }"><span class="ws-num">2</span><span>确认信息</span></div>
        <div class="wizard-step" :class="{ active: wizardStep === 3, done: wizardStep > 3 }"><span class="ws-num">3</span><span>完成支付</span></div>
      </div>

      <div v-if="wizardStep === 1" class="step-content">
        <div class="tier-options">
          <button v-for="tier in paidTiers" :key="tier.tier" class="tier-option" :class="{ selected: selectedTier === tier.tier }" @click="selectedTier = tier.tier">
            <strong>{{ tier.name }}</strong>
            <span class="to-price">¥{{ tier.price }}/{{ tier.period }}</span>
          </button>
        </div>
        <button class="btn-primary next-btn" :disabled="!selectedTier" @click="wizardStep = 2">下一步</button>
      </div>

      <div v-if="wizardStep === 2" class="step-content">
        <div class="order-summary glass">
          <div class="order-row"><span>方案</span><strong>{{ selectedInfo?.name }}</strong></div>
          <div class="order-row"><span>价格</span><strong style="color:var(--accent-gold)">¥{{ selectedInfo?.price }}/{{ selectedInfo?.period }}</strong></div>
          <div class="order-row"><span>权益</span><span>{{ (selectedInfo?.benefits || []).join('、') }}</span></div>
        </div>
        <button class="btn-primary next-btn" @click="wizardStep = 3">确认并支付</button>
      </div>

      <div v-if="wizardStep === 3" class="step-content">
        <div class="payment-methods">
          <label class="payment-option glass"><input type="radio" name="pay" checked /><span>💳 微信支付</span></label>
          <label class="payment-option glass"><input type="radio" name="pay" /><span>💳 支付宝</span></label>
        </div>
        <button class="btn-primary next-btn" @click="completePurchase">💳 确认支付 ¥{{ selectedInfo?.price }}</button>
      </div>

      <div v-if="done" class="success-step glass">
        <div class="success-icon">✅</div>
        <h3>开通成功！</h3>
        <p>您现在可以享受 {{ selectedInfo?.name }} 的全部权益</p>
        <button class="btn-primary" @click="$router.push('/user/vip')">返回会员中心</button>
      </div>
    </section>
  </div>
</template>

<script setup>
import { computed, onMounted, ref, watch } from "vue"
import { useRoute } from "vue-router"
import { VIP_TIERS as MOCK_TIERS } from "@/mock/user.js"
import { listMembershipProducts, createMembershipOrder } from "@/api/membership.js"
import { useClientAuth } from "@/shared/auth/client-auth"
import { trackExposure, trackUpgradeIntent, trackPaymentSuccess } from "@/shared/lib/experiment-tracker.js"

const route = useRoute();
const { isLoggedIn } = useClientAuth();

const wizardStep = ref(1);
const selectedTier = ref(route.query.tier || null);
const done = ref(false);
const loadingProducts = ref(false);
const products = ref([]);

const paidTiers = computed(() => {
  if (products.value.length) return products.value;
  return MOCK_TIERS.filter(t => t.price > 0);
});

const selectedInfo = computed(() => {
  const all = products.value.length ? products.value : MOCK_TIERS;
  return all.find(t => t.tier === selectedTier.value);
});

async function loadProducts() {
  if (!isLoggedIn.value) return;
  loadingProducts.value = true;
  try {
    const result = await listMembershipProducts({ status: 'ACTIVE' });
    if (result?.items?.length) {
      products.value = result.items.map((p, i) => ({
        id: p.id,
        tier: (p.member_level || 't' + i).toLowerCase(), name: p.name,
        price: p.price, period: p.duration_days >= 360 ? '年' : p.duration_days >= 85 ? '季' : '月',
        benefits: p.description ? [p.description] : [p.name + '权益']
      }));
    }
  } catch (e) {
    console.error("Failed to load products:", e);
  } finally {
    loadingProducts.value = false;
  }
}

async function completePurchase() {
  if (selectedInfo.value && selectedInfo.value.price > 0) {
    try {
      await createMembershipOrder({ product_id: selectedInfo.value.id, pay_channel: 'ALIPAY' });
      trackPaymentSuccess("vip_purchase", {
        price: selectedInfo.value.price,
        pay_channel: 'ALIPAY',
        product_id: selectedInfo.value.id,
        tier: selectedInfo.value.tier
      });
    } catch (e) {
      console.error("Order creation failed:", e);
    }
  }
  done.value = true;
}

watch(wizardStep, (newStep) => {
  if (newStep === 1) {
    trackExposure("vip_purchase");
  } else if (newStep === 2) {
    trackUpgradeIntent("vip_purchase");
  }
});

onMounted(() => {
  loadProducts();
  trackExposure("vip_purchase");
});
</script>

<style scoped>
.purchase-page { max-width: 700px; }
.section { background: var(--bg-card); border: 1px solid var(--border); border-radius: var(--radius-lg); padding: 24px; }
.section-title { font-size: 20px; font-weight: 700; margin-bottom: 4px; }
.section-subtitle { font-size: 13px; color: var(--text-secondary); margin-bottom: 20px; }
.wizard-steps { display: flex; gap: 16px; margin-bottom: 24px; }
.wizard-step { display: flex; align-items: center; gap: 8px; font-size: 13px; color: var(--text-muted); }
.wizard-step.active { color: var(--accent-gold); }
.wizard-step.done { color: var(--positive); }
.ws-num { width: 24px; height: 24px; border-radius: 50%; background: rgba(255,255,255,.06); display: flex; align-items: center; justify-content: center; font-size: 12px; font-weight: 700; }
.wizard-step.active .ws-num { background: var(--accent-gold-glow); color: var(--accent-gold); }
.wizard-step.done .ws-num { background: var(--positive-bg); color: var(--positive); }
.step-content { display: grid; gap: 16px; }
.tier-options { display: grid; gap: 8px; }
.tier-option { display: flex; justify-content: space-between; align-items: center; padding: 14px; border-radius: var(--radius-md); border: 1px solid var(--border); text-align: left; width: 100%; background: none; color: var(--text-primary); cursor: pointer; }
.tier-option:hover { border-color: var(--border-light); }
.tier-option.selected { border-color: var(--accent-gold); background: var(--accent-gold-glow); }
.to-price { color: var(--accent-gold); font-weight: 600; }
.next-btn { width: 100%; padding: 14px; justify-content: center; display: flex; font-size: 15px; cursor: pointer; }
.btn-primary { border: none; border-radius: var(--radius-full); background: linear-gradient(135deg,var(--accent-gold),var(--accent-gold-dim)); color: #000; font-weight: 600; }
.btn-primary:disabled { opacity: .5; }
.order-summary { padding: 16px; border-radius: var(--radius-md); display: grid; gap: 10px; }
.order-row { display: flex; justify-content: space-between; font-size: 14px; }
.order-row span { color: var(--text-secondary); }
.order-row strong { color: var(--text-primary); }
.payment-methods { display: grid; gap: 8px; }
.payment-option { display: flex; align-items: center; gap: 10px; padding: 12px; border-radius: var(--radius-md); cursor: pointer; }
.success-step { text-align: center; padding: 40px 20px; border-radius: var(--radius-lg); }
.success-icon { font-size: 48px; margin-bottom: 16px; }
.success-step h3 { font-size: 22px; margin-bottom: 8px; }
.success-step p { font-size: 14px; color: var(--text-secondary); margin-bottom: 20px; }
</style>
