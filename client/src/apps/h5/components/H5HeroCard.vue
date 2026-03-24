<template>
  <section class="h5-card h5-hero-card" :class="surfaceClass">
    <div class="h5-card-body hero-body">
      <div class="hero-copy">
        <p v-if="eyebrow" class="h5-eyebrow">{{ eyebrow }}</p>
        <h1 class="h5-title">{{ title }}</h1>
        <p v-if="description" class="h5-subtitle">{{ description }}</p>
        <div v-if="meta?.length" class="h5-inline-meta">
          <span v-for="item in meta" :key="item" class="h5-meta-chip">{{ item }}</span>
        </div>
        <slot name="actions" />
      </div>
      <div class="hero-side">
        <slot />
      </div>
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
  description: {
    type: String,
    default: ""
  },
  meta: {
    type: Array,
    default: () => []
  },
  tone: {
    type: String,
    default: "hero"
  }
});

const surfaceClass = computed(() => resolveSurfaceToneClasses(props.tone));
</script>

<style scoped>
.hero-body {
  display: grid;
  gap: 16px;
}

.hero-copy,
.hero-side {
  display: grid;
  gap: 12px;
}
</style>
