<template>
  <section class="state-panel" :class="[tone, { compact }]">
    <p v-if="eyebrow" class="state-eyebrow">{{ eyebrow }}</p>
    <strong class="state-title">{{ title }}</strong>
    <p v-if="description" class="state-description">{{ description }}</p>
    <div v-if="$slots.actions" class="state-actions">
      <slot name="actions" />
    </div>
  </section>
</template>

<script setup>
defineProps({
  tone: {
    type: String,
    default: "info"
  },
  eyebrow: {
    type: String,
    default: ""
  },
  title: {
    type: String,
    default: ""
  },
  description: {
    type: String,
    default: ""
  },
  compact: {
    type: Boolean,
    default: false
  }
});
</script>

<style scoped>
.state-panel {
  border-radius: 18px;
  border: 1px solid #dbe4ee;
  padding: 16px 18px;
  display: grid;
  gap: 6px;
  background: #fff;
}

.state-panel.compact {
  border-radius: 14px;
  padding: 12px 14px;
}

.state-panel.info {
  border-color: rgba(59, 130, 246, 0.18);
  background: rgba(239, 246, 255, 0.94);
}

.state-panel.success {
  border-color: rgba(16, 185, 129, 0.2);
  background: rgba(236, 253, 245, 0.96);
}

.state-panel.warning {
  border-color: rgba(245, 158, 11, 0.22);
  background: rgba(255, 251, 235, 0.96);
}

.state-panel.danger {
  border-color: rgba(239, 68, 68, 0.22);
  background: rgba(254, 242, 242, 0.96);
}

.state-eyebrow {
  margin: 0;
  font-size: 11px;
  font-weight: 700;
  letter-spacing: 0.16em;
  color: #2563eb;
}

.state-title {
  color: #0f172a;
  font-size: 18px;
}

.state-description {
  margin: 0;
  color: #475569;
  line-height: 1.65;
  font-size: 13px;
}

.state-actions {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
  margin-top: 4px;
}
</style>
