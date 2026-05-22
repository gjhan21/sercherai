<template>
  <div class="blurry-paywall-container">
    <!-- 原内容，通过插槽传入，并根据解锁状态决定是否模糊 -->
    <div :class="['paywall-content', { 'is-blurred': !isUnlocked }]">
      <slot></slot>
    </div>

    <!-- 锁定遮罩层：引导订阅 -->
    <div v-if="!isUnlocked" class="paywall-overlay">
      <div class="paywall-box">
        <h3 class="paywall-title">解锁进阶分析与操作</h3>
        <p class="paywall-desc">开通高级会员，获取完整的 AI 智能复盘及多维数据指标，把握关键投资机会。</p>
        <button class="paywall-btn" @click="$emit('unlock')">立即开通高级会员</button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { defineProps, defineEmits } from 'vue'

const props = defineProps({
  isUnlocked: {
    type: Boolean,
    default: false
  }
})

const emit = defineEmits(['unlock'])
</script>

<style scoped>
.blurry-paywall-container {
  position: relative;
  overflow: hidden;
  border-radius: 8px;
}

.paywall-content {
  transition: filter 0.3s ease;
}

.is-blurred {
  filter: blur(6px) grayscale(50%);
  user-select: none;
  pointer-events: none;
}

.paywall-overlay {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  display: flex;
  justify-content: center;
  align-items: center;
  background: rgba(255, 255, 255, 0.4);
  z-index: 10;
}

/* 适配暗色模式 */
@media (prefers-color-scheme: dark) {
  .paywall-overlay {
    background: rgba(18, 18, 18, 0.6);
  }
}

.paywall-box {
  background: var(--bg-color, #ffffff);
  border: 1px solid var(--border-color, #e5e7eb);
  padding: 32px 24px;
  border-radius: 12px;
  text-align: center;
  box-shadow: 0 10px 25px -5px rgba(0, 0, 0, 0.1), 0 8px 10px -6px rgba(0, 0, 0, 0.1);
  max-width: 80%;
  width: 360px;
}

/* 适配暗色模式的盒子 */
@media (prefers-color-scheme: dark) {
  .paywall-box {
    background: #1f2937;
    border-color: #374151;
    box-shadow: 0 10px 25px -5px rgba(0, 0, 0, 0.5);
  }
}

.paywall-title {
  margin: 0 0 12px 0;
  font-size: 1.25rem;
  font-weight: 600;
  color: var(--text-primary, #111827);
}

@media (prefers-color-scheme: dark) {
  .paywall-title {
    color: #f9fafb;
  }
}

.paywall-desc {
  margin: 0 0 24px 0;
  font-size: 0.875rem;
  color: var(--text-secondary, #6b7280);
  line-height: 1.5;
}

@media (prefers-color-scheme: dark) {
  .paywall-desc {
    color: #9ca3af;
  }
}

.paywall-btn {
  background-color: #3b82f6;
  color: #ffffff;
  border: none;
  border-radius: 6px;
  padding: 10px 20px;
  font-size: 0.875rem;
  font-weight: 500;
  cursor: pointer;
  transition: background-color 0.2s;
  width: 100%;
}

.paywall-btn:hover {
  background-color: #2563eb;
}
</style>
