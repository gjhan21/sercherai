<template>
  <article class="h5-metric-card h5-summary-card" :class="surfaceClass">
    <p class="h5-summary-label">{{ label }}</p>
    <strong class="h5-summary-value">{{ value }}</strong>
    <small v-if="note" class="h5-summary-note">{{ note }}</small>
  </article>
</template>

<script setup>
import { computed } from "vue";
import { resolveMetricToneClasses } from "../lib/surface-tone.js";

const props = defineProps({
  label: {
    type: String,
    default: ""
  },
  value: {
    type: String,
    default: "-"
  },
  note: {
    type: String,
    default: ""
  },
  tone: {
    type: String,
    default: "default"
  }
});

const surfaceClass = computed(() => resolveMetricToneClasses(props.tone));
</script>
