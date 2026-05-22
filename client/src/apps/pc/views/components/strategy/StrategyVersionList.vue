<script setup>
defineProps({
  items: {
    type: Array,
    required: true
  },
  activeKey: {
    type: String,
    default: ""
  },
  compareData: {
    type: Object,
    default: () => ({})
  }
});

const emit = defineEmits(["select", "reset"]);
</script>

<template>
  <div v-if="items.length" class="strategy-history-box finance-card-surface">
    <div class="stock-news-head">
      <p>历史版本</p>
      <div class="strategy-history-head-actions">
        <span>{{ items.length }} 条</span>
        <button
          v-if="compareData?.isCustomSelected"
          type="button"
          class="strategy-history-reset finance-mini-btn finance-mini-btn-soft"
          @click="emit('reset')"
        >
          回到默认对比
        </button>
      </div>
    </div>
    <p class="strategy-history-hint">点一条历史版本，直接和当前 explanation 对比。</p>
    <div class="strategy-history-list">
      <button
        v-for="item in items"
        :key="`history-${item.key}`"
        type="button"
        class="strategy-history-item finance-list-card finance-list-card-interactive"
        :class="{ active: activeKey === item.key }"
        @click="emit('select', item.key)"
      >
        <p>{{ item.title }}</p>
        <strong>{{ item.version }}</strong>
        <span>{{ item.note }}</span>
        <em v-if="item.deepForecast" class="strategy-history-forecast-tag">
          深推演 {{ item.deepForecast.statusLabel }}
        </em>
      </button>
    </div>

    <!-- Comparison Detail -->
    <div v-if="compareData?.diff" class="strategy-version-box compact finance-card-pale" style="margin-top: 16px">
      <div class="stock-news-head">
        <p>版本对照</p>
        <span>{{ compareData.selectedTitle }}</span>
      </div>
      <p class="explanation-note">{{ compareData.selectedNote }}</p>
      <div class="strategy-version-grid">
        <article class="finance-list-card finance-list-card-panel">
          <p>{{ compareData.isCustomSelected ? "所选版本" : "默认对比版本" }}</p>
          <strong>{{ compareData.diff.beforeLabel }}</strong>
          <span>{{ compareData.diff.beforeNote }}</span>
        </article>
        <article class="finance-list-card finance-list-card-panel">
          <p>当前解释</p>
          <strong>{{ compareData.diff.afterLabel }}</strong>
          <span>{{ compareData.diff.afterNote }}</span>
        </article>
        <article class="finance-list-card finance-list-card-panel">
          <p>变化总结</p>
          <strong>{{ compareData.diff.diffLabel }}</strong>
          <span>{{ compareData.diff.diffNote }}</span>
        </article>
      </div>
    </div>
  </div>
</template>
