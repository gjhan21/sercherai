<script setup>
defineProps({
  title: {
    type: String,
    default: "历史推荐业绩"
  },
  summary: {
    type: String,
    default: ""
  },
  note: {
    type: String,
    default: ""
  },
  rows: {
    type: Array,
    required: true
  },
  activeId: {
    type: [String, Number],
    default: ""
  }
});
</script>

<template>
  <div class="stock-performance-box finance-card-surface">
    <div class="stock-performance-head">
      <p>{{ title }}</p>
      <span>{{ summary }}</span>
    </div>
    <p class="stock-performance-note">{{ note }}</p>
    <div v-if="rows.length > 0" class="performance-table-wrap finance-table-wrap">
      <table class="performance-table finance-data-table finance-data-table-compact">
        <thead>
          <tr>
            <th>日期</th>
            <th>单日收益</th>
            <th>累计收益</th>
            <th>基准累计</th>
            <th>累计超额</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="item in rows" :key="`${activeId}-${item.date}`">
            <td>{{ item.date }}</td>
            <td :class="item.dailyClass">{{ item.dailyReturn }}</td>
            <td :class="item.cumulativeClass">{{ item.cumulativeReturn }}</td>
            <td :class="item.benchmarkClass">{{ item.benchmarkReturn }}</td>
            <td :class="item.excessClass">{{ item.excessReturn }}</td>
          </tr>
        </tbody>
      </table>
    </div>
    <div v-else class="empty-inline finance-empty-inline">暂无历史推荐业绩</div>
  </div>
</template>

<style scoped>
/* Styles will be inherited from the global finance system or specific view styles, 
   but we can add component-specific scope here if needed. */
</style>
