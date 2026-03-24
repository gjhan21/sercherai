<template>
  <div class="h5-empty">
    <strong>{{ title }}</strong>
    <p>{{ description }}</p>
    <slot />
  </div>
</template>

<script setup>
defineProps({
  title: {
    type: String,
    default: "暂无内容"
  },
  description: {
    type: String,
    default: "稍后再试，或先执行上方主操作。"
  }
});
</script>
