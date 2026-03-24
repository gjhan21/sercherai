<template>
  <div v-if="visible" class="h5-sticky-cta">
    <div class="h5-sticky-cta-inner" :class="`is-${actionMode}`">
      <div class="h5-sticky-cta-copy">
        <strong>{{ title }}</strong>
        <p v-if="description">{{ description }}</p>
      </div>
      <div class="h5-action-row" :class="`is-${actionMode}`">
        <button v-if="secondaryLabel" type="button" class="h5-btn-secondary" @click="$emit('secondary')">
          {{ secondaryLabel }}
        </button>
        <button v-if="primaryLabel" type="button" class="h5-btn block" @click="$emit('primary')">
          {{ primaryLabel }}
        </button>
      </div>
    </div>
  </div>
  <div v-if="visible" class="h5-sticky-cta-space" />
</template>

<script setup>
import { computed } from "vue";
import { resolveStickyActionMode } from "../lib/surface-tone.js";

const props = defineProps({
  visible: {
    type: Boolean,
    default: true
  },
  title: {
    type: String,
    default: ""
  },
  description: {
    type: String,
    default: ""
  },
  primaryLabel: {
    type: String,
    default: ""
  },
  secondaryLabel: {
    type: String,
    default: ""
  }
});

const actionMode = computed(() => resolveStickyActionMode({
  primaryLabel: props.primaryLabel,
  secondaryLabel: props.secondaryLabel
}));

defineEmits(["primary", "secondary"]);
</script>
