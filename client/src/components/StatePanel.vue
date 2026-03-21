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
  border-radius: 16px;
  border: 1px solid var(--color-border-soft-heavy);
  padding: 14px 16px;
  display: grid;
  gap: 6px;
  box-shadow: var(--shadow-soft);
}

.state-panel.compact {
  padding: 11px 12px;
  border-radius: 14px;
}

.state-panel.info {
  border-color: var(--color-border-focus-soft);
  background: var(--color-surface-panel-soft-strong);
}

.state-panel.success {
  border-color: rgba(46, 125, 50, 0.18);
  background: rgba(239, 247, 242, 0.96);
}

.state-panel.warning {
  border-color: rgba(184, 130, 48, 0.2);
  background: rgba(251, 245, 233, 0.96);
}

.state-panel.danger {
  border-color: rgba(178, 58, 42, 0.18);
  background: rgba(253, 242, 242, 0.96);
}

.state-eyebrow {
  margin: 0;
  font-size: 11px;
  font-weight: 700;
  letter-spacing: 0.14em;
  color: var(--color-pine-700);
}

.state-title {
  font-size: 15px;
  color: var(--color-text-main);
}

.state-description {
  margin: 0;
  font-size: 12px;
  line-height: 1.6;
  color: var(--color-text-sub);
}

.state-actions {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  margin-top: 2px;
}

:deep(.state-actions button),
:deep(.state-actions a) {
  min-height: 40px;
  border: 1px solid transparent;
  border-radius: 12px;
  padding: 8px 13px;
  font-size: 12px;
  font-weight: 600;
  text-decoration: none;
  touch-action: manipulation;
  transition:
    border-color 0.2s ease,
    box-shadow 0.2s ease,
    background 0.2s ease,
    color 0.2s ease;
}

:deep(.state-actions button),
:deep(.state-actions a.primary) {
  color: #fff;
  background: var(--gradient-primary);
}

:deep(.state-actions .ghost) {
  color: var(--color-pine-700);
  background: var(--color-surface-card-elevated);
  border-color: var(--color-border-soft);
}

:deep(.state-actions button:focus-visible),
:deep(.state-actions a:focus-visible) {
  outline: none;
  box-shadow: var(--shadow-focus);
}
</style>
