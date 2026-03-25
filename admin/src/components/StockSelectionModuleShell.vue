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
  { label: "总览", to: "/stock-selection/overview" },
  { label: "运行中心", to: "/stock-selection/runs" },
  { label: "策略设计", to: "/stock-selection/profiles" },
  { label: "候选与审核发布", to: "/stock-selection/candidates" },
  { label: "事件审核", to: "/stock-selection/events" },
  { label: "评估复盘", to: "/stock-selection/evaluation" }
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
    <div class="page-header stock-selection-header">
      <div>
        <h1 class="page-title">{{ title }}</h1>
        <p class="muted">{{ description }}</p>
      </div>
      <div class="stock-selection-tabs">
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
.stock-selection-header {
  gap: 14px;
}

.stock-selection-tabs {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}
</style>
