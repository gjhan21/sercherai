<script setup>
import { formatScore } from "@/utils/format"; // Assuming standard score formatter exist

defineProps({
  explanation: {
    type: Object,
    default: null
  },
  explanationSummary: {
    type: String,
    default: ""
  },
  explanationCards: {
    type: Array,
    default: () => []
  },
  seedHighlights: {
    type: Array,
    default: () => []
  },
  scenarios: {
    type: Array,
    default: () => []
  },
  scenarioMeta: {
    type: Object,
    default: null
  },
  relationshipSummary: {
    type: Object,
    default: null
  },
  agentOpinions: {
    type: Array,
    default: () => []
  },
  risks: {
    type: Array,
    default: () => []
  },
  proofSourceNote: {
    type: String,
    default: "系统会从多个场景验证这次推荐。"
  },
  categoryLabel: {
    type: String,
    default: "为什么选它"
  }
});
</script>

<template>
  <div class="strategy-market-insight-stack">
    <!-- Explanation / Why Now -->
    <section v-if="explanation" class="strategy-explanation-box finance-card-pale">
      <div class="stock-news-head">
        <p>{{ categoryLabel }}</p>
        <span>{{ explanation.strategy_version || "strategy-engine" }}</span>
      </div>
      <p class="explanation-summary">
        {{ explanationSummary }}
      </p>
      <div class="reason-support-grid">
        <article
          v-for="item in explanationCards"
          :key="`explain-${item.label}`"
          class="finance-list-card finance-list-card-panel"
        >
          <p>{{ item.label }}</p>
          <strong>{{ item.value }}</strong>
          <span>{{ item.note }}</span>
        </article>
      </div>
      <div class="chip-group" v-if="seedHighlights.length > 0" style="margin-top: 16px">
        <span
          v-for="item in seedHighlights"
          :key="`seed-${item}`"
          class="finance-pill finance-pill-compact finance-pill-info"
        >
          {{ item }}
        </span>
      </div>
    </section>

    <!-- Scenarios -->
    <section class="strategy-explanation-box finance-card-pale" style="margin-top: 24px">
      <div class="stock-news-head">
        <p>多场景推演</p>
        <span>{{ scenarios.length }} 个场景</span>
      </div>
      <p class="explanation-note">{{ proofSourceNote }}</p>
      
      <div v-if="scenarioMeta || relationshipSummary" class="reason-support-grid" style="margin-bottom: 16px">
        <article v-if="scenarioMeta" class="finance-list-card finance-list-card-panel">
          <p>L2 情景摘要</p>
          <strong>{{ scenarioMeta.summary }}</strong>
          <span>{{ scenarioMeta.note || "当前未补更多主情景说明。" }}</span>
        </article>
        <article v-if="relationshipSummary" class="finance-list-card finance-list-card-panel">
          <p>关系快照</p>
          <strong>{{ relationshipSummary.summary }}</strong>
          <span>{{ relationshipSummary.note || "当前未补更多关系节点说明。" }}</span>
        </article>
      </div>

      <div class="scenario-grid">
        <article
          v-for="item in scenarios"
          :key="`scenario-${item.scenario}`"
          class="scenario-item finance-list-card finance-list-card-panel"
        >
          <p>{{ item.scenario }}</p>
          <strong>{{ item.action }}</strong>
          <span>{{ item.thesis }}</span>
          <em>
            {{ item.confirmation ? `确认 ${item.confirmation}` : "等待更多确认信号" }}
            <template v-if="item.invalidation"> · 失效 {{ item.invalidation }}</template>
            <template v-else-if="item.window"> · 窗口 {{ item.window }}</template>
          </em>
        </article>
      </div>
    </section>

    <slot name="custom-insight"></slot>

    <!-- Agent Opinions -->
    <section class="strategy-explanation-box finance-card-pale" style="margin-top: 24px">
      <div class="stock-news-head">
        <p>角色评审</p>
        <span>{{ agentOpinions.length }} 个视角</span>
      </div>
      <div class="agent-opinion-list">
        <article
          v-for="item in agentOpinions"
          :key="`agent-${item.role}`"
          class="agent-opinion-item finance-list-card finance-list-card-panel"
        >
          <p>{{ item.role }}</p>
          <strong>{{ item.stance }} · {{ item.confidence ? (item.confidence * 100).toFixed(0) + '%' : '-' }}</strong>
          <span>{{ item.summary }}</span>
          <em v-if="item.veto">已触发 veto</em>
        </article>
      </div>
    </section>

    <!-- Risks -->
    <section class="strategy-explanation-box finance-card-pale" style="margin-top: 24px">
      <div class="stock-news-head">
        <p>风险与失效条件</p>
        <span>{{ risks.length }} 条</span>
      </div>
      <div class="risk-flag-list">
        <article
          v-for="item in risks"
          :key="`risk-${item.label}-${item.text}`"
          class="risk-flag-item finance-list-card finance-list-card-panel"
          :class="{ subtle: item.subtle }"
        >
          <strong>{{ item.label }}</strong>
          <span>{{ item.text }}</span>
        </article>
      </div>
    </section>
  </div>
</template>
