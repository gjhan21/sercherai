<template>
  <div class="variant-switcher">
    <RouterLink class="switch-link home-link" to="/demo">方案列表</RouterLink>
    <RouterLink :class="['switch-link', { active: current === 'current' }]" to="/demo/current">当前版本</RouterLink>
    <RouterLink :class="['switch-link', { active: current === 'a' }]" to="/demo/a">A</RouterLink>
    <RouterLink :class="['switch-link', { active: current === 'b' }]" to="/demo/b">B</RouterLink>
    <RouterLink :class="['switch-link', { active: current === 'c' }]" to="/demo/c">C</RouterLink>
  </div>
</template>

<script setup>
defineProps({
  current: {
    type: String,
    default: ""
  }
});
</script>

<style scoped>
.variant-switcher {
  position: fixed;
  top: 14px;
  right: 14px;
  z-index: 50;
  display: inline-flex;
  gap: 8px;
  padding: 8px;
  border: 1px solid rgba(135, 154, 184, 0.22);
  border-radius: 999px;
  background: rgba(6, 12, 21, 0.86);
  backdrop-filter: blur(14px);
  box-shadow: 0 18px 34px rgba(0, 0, 0, 0.3);
}

.switch-link {
  padding: 7px 12px;
  border-radius: 999px;
  color: #c9d4e6;
  text-decoration: none;
  font-size: 12px;
  line-height: 1;
  transition: background-color 180ms ease, color 180ms ease, transform 180ms ease;
}

.switch-link:hover {
  color: #f7fbff;
  background: rgba(255, 255, 255, 0.08);
  transform: translateY(-1px);
}

.switch-link.active {
  color: #08131e;
  background: linear-gradient(135deg, #b7ccff, #f4d2a5);
}

.home-link {
  border: 1px solid rgba(163, 182, 215, 0.18);
}

@media (max-width: 860px) {
  .variant-switcher {
    left: 12px;
    right: 12px;
    justify-content: center;
    flex-wrap: wrap;
    border-radius: 18px;
  }
}
</style>
