<template>
  <div class="h5-action-row" :class="{ column }">
    <slot />
  </div>
</template>

<script setup>
defineProps({
  column: {
    type: Boolean,
    default: false
  }
});
</script>

<style scoped>
.h5-action-row.column {
  display: grid;
}
</style>
