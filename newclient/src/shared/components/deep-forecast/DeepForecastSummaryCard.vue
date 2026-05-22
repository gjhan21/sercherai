<template>
  <section v-if="summary" class="forecast-summary-card" :class="[`mode-${mode}`, `tone-${summary.tone || 'muted'}`]">
    <div class="forecast-summary-head">
      <div>
        <p class="forecast-summary-kicker">深度推演</p>
        <strong>{{ heading }}</strong>
      </div>
      <span class="forecast-summary-status">{{ summary.statusLabel }}</span>
    </div>
    <p class="forecast-summary-copy">{{ summary.summary }}</p>
    <div v-if="summary.note" class="forecast-summary-note">{{ summary.note }}</div>
    <div class="forecast-summary-meta">
      <span v-if="summary.scenario">主情景 {{ summary.scenario }}</span>
      <span v-if="summary.actionGuidance">动作 {{ summary.actionGuidance }}</span>
      <span v-if="summary.requiresVip">VIP 正文</span>
    </div>
    <RouterLink v-if="to" class="forecast-summary-link" :to="to">查看完整深度推演</RouterLink>
  </section>
</template>

<script setup>
import { computed } from "vue";

const props = defineProps({
  summary: {
    type: Object,
    default: null
  },
  to: {
    type: String,
    default: ""
  },
  mode: {
    type: String,
    default: "pc"
  },
  heading: {
    type: String,
    default: "L3 深推演摘要"
  }
});

const summary = computed(() => props.summary);
</script>

<style scoped>
.forecast-summary-card { display: grid; gap: 10px; padding: 16px; border-radius: var(--radius-md); border: 1px solid var(--border); background: rgba(255,255,255,.02); }
.forecast-summary-head { display: flex; justify-content: space-between; gap: 12px; align-items: flex-start; }
.forecast-summary-kicker { font-size: 11px; color: var(--accent-gold); text-transform: uppercase; letter-spacing: .08em; margin-bottom: 4px; }
.forecast-summary-status { font-size: 12px; color: var(--accent-gold); }
.forecast-summary-copy, .forecast-summary-note { color: var(--text-secondary); line-height: 1.7; }
.forecast-summary-meta { display: flex; gap: 10px; flex-wrap: wrap; color: var(--text-muted); font-size: 12px; }
.forecast-summary-link { display: inline-flex; width: fit-content; padding: 10px 14px; border-radius: var(--radius-full); border: 1px solid var(--border-gold); color: var(--accent-gold); text-decoration: none; font-size: 13px; font-weight: 600; }
.tone-success { border-color: rgba(34,197,94,.22); }
.tone-running, .tone-queued { border-color: rgba(96,165,250,.22); }
.tone-failed { border-color: rgba(248,113,113,.22); }
.mode-h5 .forecast-summary-link { width: 100%; justify-content: center; }
</style>
