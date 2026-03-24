<template>
  <section class="h5-card h5-section-block" :class="surfaceClass">
    <div class="h5-card-body h5-section-block-body">
      <header v-if="eyebrow || title || subtitle" class="h5-section-head">
        <p v-if="eyebrow" class="h5-eyebrow">{{ eyebrow }}</p>
        <h2 v-if="title" class="h5-title">{{ title }}</h2>
        <p v-if="subtitle" class="h5-subtitle">{{ subtitle }}</p>
      </header>
      <slot />
    </div>
  </section>
</template>

<script setup>
import { computed } from "vue";
import { resolveSurfaceToneClasses } from "../lib/surface-tone.js";

const props = defineProps({
  eyebrow: {
    type: String,
    default: ""
  },
  title: {
    type: String,
    default: ""
  },
  subtitle: {
    type: String,
    default: ""
  },
  tone: {
    type: String,
    default: "default"
  }
});

const surfaceClass = computed(() => resolveSurfaceToneClasses(props.tone));
</script>
