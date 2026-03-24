<template>
  <article class="h5-list-card h5-article-card interactive" :class="{ active }">
    <div class="h5-list-topline h5-article-head">
      <div class="h5-copy">
        <span class="h5-list-submeta">{{ meta }}</span>
        <strong>{{ title }}</strong>
      </div>
      <span v-if="badge" class="h5-badge" :class="badgeTone">{{ badge }}</span>
    </div>
    <p class="h5-list-note">{{ summary }}</p>
    <div v-if="tags?.length" class="h5-meta-list h5-article-tags">
      <span v-for="item in tags" :key="item" class="h5-meta-chip">{{ item }}</span>
    </div>
  </article>
</template>

<script setup>
import { computed } from "vue";

const props = defineProps({
  title: {
    type: String,
    default: ""
  },
  summary: {
    type: String,
    default: ""
  },
  meta: {
    type: String,
    default: ""
  },
  badge: {
    type: String,
    default: ""
  },
  tone: {
    type: String,
    default: "brand"
  },
  tags: {
    type: Array,
    default: () => []
  },
  active: {
    type: Boolean,
    default: false
  }
});

const badgeTone = computed(() => props.tone || "brand");
</script>
