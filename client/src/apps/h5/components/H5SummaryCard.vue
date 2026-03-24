<template>
  <article class="h5-metric-card" :class="surfaceClass">
    <p>{{ label }}</p>
    <strong>{{ value }}</strong>
    <small v-if="note">{{ note }}</small>
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
