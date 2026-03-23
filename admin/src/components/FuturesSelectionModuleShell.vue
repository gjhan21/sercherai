<script setup>
import { computed } from "vue";
import { useRoute, useRouter } from "vue-router";

const props = defineProps({
  title: {
    type: String,
    required: true
  },
  description: {
    type: String,
    default: ""
  }
});

const route = useRoute();
const router = useRouter();

const tabs = [
  { label: "总览", to: "/futures-selection/overview" },
  { label: "运行中心", to: "/futures-selection/runs" },
  { label: "策略设计", to: "/futures-selection/profiles" },
  { label: "候选与审核发布", to: "/futures-selection/candidates" },
  { label: "评估复盘", to: "/futures-selection/evaluation" }
];

const activeTab = computed(() => route.path);

function go(to) {
  if (to === route.path) {
    return;
  }
  router.push(to);
}
</script>

<template>
  <div class="page">
    <div class="page-header futures-selection-header">
      <div>
        <h1 class="page-title">{{ title }}</h1>
        <p class="muted">{{ description }}</p>
      </div>
      <div class="futures-selection-tabs">
        <el-button
          v-for="tab in tabs"
          :key="tab.to"
          :type="activeTab === tab.to ? 'primary' : 'default'"
          @click="go(tab.to)"
        >
          {{ tab.label }}
        </el-button>
      </div>
      <slot name="actions" />
    </div>

    <slot />
  </div>
</template>

<style scoped>
.futures-selection-header {
  gap: 14px;
}

.futures-selection-tabs {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}
</style>
