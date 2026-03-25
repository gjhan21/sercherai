<template>
  <section class="community-compose-page fade-up">
    <header class="card finance-section-card compose-hero">
      <div class="finance-copy-stack">
        <p class="hero-kicker">发起讨论</p>
        <h1 class="section-title">先写观点，再写依据和风险</h1>
        <p class="section-subtitle">
          这不是聊天室发言框，而是结构化主题帖。当前只支持与股票、期货、资讯、策略四类对象关联。
        </p>
      </div>
      <div class="compose-hero-actions">
        <button type="button" class="finance-ghost-btn" @click="router.push('/community')">返回广场</button>
      </div>
    </header>

    <StatePanel
      tone="warning"
      eyebrow="发帖提醒"
      title="请同时写清判断、依据和风险边界。"
      description="生产页不保留“只有结论、没有原因”的主题帖。提交后当前版本会先直接进入主题详情。"
    />

    <form class="compose-form" @submit.prevent="submitTopic">
      <section class="card finance-section-card compose-section">
        <header class="section-head compact">
          <div>
            <h2 class="section-title">主题信息</h2>
            <p class="section-subtitle">确定讨论对象属于哪一类内容。</p>
          </div>
        </header>

        <div class="compose-grid">
          <label class="compose-field">
            <span>标题</span>
            <input v-model.trim="form.title" type="text" maxlength="60" placeholder="例如：白酒龙头当前位置更适合观察而不是追高" />
          </label>

          <label class="compose-field">
            <span>主题类型</span>
            <select v-model="form.topic_type">
              <option v-for="item in topicTypeOptions" :key="item.value" :value="item.value">{{ item.label }}</option>
            </select>
          </label>

          <label class="compose-field">
            <span>观点方向</span>
            <select v-model="form.stance">
              <option v-for="item in stanceOptions" :key="item.value" :value="item.value">{{ item.label }}</option>
            </select>
          </label>

          <label class="compose-field">
            <span>时间范围</span>
            <select v-model="form.time_horizon">
              <option v-for="item in horizonOptions" :key="item.value" :value="item.value">{{ item.label }}</option>
            </select>
          </label>
        </div>
      </section>

      <section class="card finance-section-card compose-section">
        <header class="section-head compact">
          <div>
            <h2 class="section-title">关联对象</h2>
            <p class="section-subtitle">当前后端只支持对象类型、对象 ID 和对象快照文本。</p>
          </div>
        </header>

        <div class="compose-grid">
          <label class="compose-field">
            <span>对象类型</span>
            <select v-model="form.target_type">
              <option v-for="item in targetTypeOptions" :key="item.value" :value="item.value">{{ item.label }}</option>
            </select>
          </label>

          <label class="compose-field">
            <span>对象 ID</span>
            <input v-model.trim="form.target_id" type="text" maxlength="60" placeholder="例如：600519 / article_demo_001 / sr_demo_001" />
          </label>

          <label class="compose-field compose-field-span">
            <span>对象快照</span>
            <input v-model.trim="form.target_snapshot" type="text" maxlength="120" placeholder="例如：600519 贵州茅台 / 量化策略周报 / 今日主推荐" />
          </label>
        </div>
      </section>

      <section class="card finance-section-card compose-section">
        <header class="section-head compact">
          <div>
            <h2 class="section-title">观点内容</h2>
            <p class="section-subtitle">先补摘要，再写依据、风险和完整正文。</p>
          </div>
        </header>

        <div class="compose-stack">
          <label class="compose-field">
            <span>摘要</span>
            <textarea
              v-model.trim="form.summary"
              maxlength="160"
              placeholder="用一两句话概括你的观点，让用户进入详情前能快速判断是否值得继续看。"
            />
          </label>

          <div class="compose-grid">
            <label class="compose-field">
              <span>为什么这样看</span>
              <textarea
                v-model.trim="form.reason_text"
                maxlength="240"
                placeholder="例如：量价结构仍在确认，当前位置追高性价比偏低。"
              />
            </label>

            <label class="compose-field">
              <span>风险边界</span>
              <textarea
                v-model.trim="form.risk_text"
                maxlength="240"
                placeholder="例如：如果量能承接失败或消息证伪，短线回撤会明显放大。"
              />
            </label>
          </div>

          <label class="compose-field">
            <span>完整正文</span>
            <textarea
              v-model.trim="form.content"
              maxlength="2000"
              placeholder="补充你的判断逻辑、参考依据和执行条件。当前版本不支持附件与图片。"
            />
          </label>
        </div>
      </section>

      <section class="card finance-section-card compose-section">
        <header class="section-head compact">
          <div>
            <h2 class="section-title">发布确认</h2>
            <p class="section-subtitle">提交前再确认一次表达是否完整。</p>
          </div>
        </header>

        <label class="compose-agreement">
          <input v-model="agreed" type="checkbox" />
          <span>我确认已写明观点依据和风险边界，且内容仅代表个人研究判断。</span>
        </label>

        <p v-if="errorMessage" class="compose-error">{{ errorMessage }}</p>

        <div class="compose-submit-row">
          <button type="submit" class="finance-primary-btn" :disabled="submitting">
            {{ submitting ? "提交中..." : "提交主题" }}
          </button>
          <button type="button" class="finance-ghost-btn" :disabled="submitting" @click="router.push('/community')">
            暂不发布
          </button>
        </div>
      </section>
    </form>
  </section>
</template>

<script setup>
import { reactive, ref } from "vue";
import { useRouter, useRoute } from "vue-router";
import StatePanel from "../components/StatePanel.vue";
import { createCommunityTopic } from "../api/community";

const router = useRouter();
const route = useRoute();

const topicTypeOptions = [
  { value: "STOCK", label: "股票" },
  { value: "FUTURES", label: "期货" },
  { value: "NEWS", label: "资讯" },
  { value: "STRATEGY", label: "策略" }
];

const stanceOptions = [
  { value: "BULLISH", label: "看多" },
  { value: "BEARISH", label: "看空" },
  { value: "WATCH", label: "观察" }
];

const horizonOptions = [
  { value: "SHORT", label: "短线" },
  { value: "SWING", label: "波段" },
  { value: "MID", label: "中线" },
  { value: "LONG", label: "中长线" }
];

const targetTypeOptions = [
  { value: "STOCK", label: "股票对象" },
  { value: "FUTURES", label: "期货对象" },
  { value: "NEWS_ARTICLE", label: "资讯文章" },
  { value: "STRATEGY_ITEM", label: "策略条目" }
];

const form = reactive({
  title: String(route.query.title || ""),
  topic_type: String(route.query.topic_type || "STOCK"),
  stance: String(route.query.stance || "WATCH"),
  time_horizon: String(route.query.time_horizon || "SHORT"),
  target_type: String(route.query.target_type || "STOCK"),
  target_id: String(route.query.target_id || ""),
  target_snapshot: String(route.query.target_snapshot || ""),
  summary: "",
  reason_text: "",
  risk_text: "",
  content: ""
});

const agreed = ref(false);
const submitting = ref(false);
const errorMessage = ref("");

async function submitTopic() {
  errorMessage.value = validateForm();
  if (errorMessage.value) {
    return;
  }

  submitting.value = true;
  try {
    const result = await createCommunityTopic({
      title: form.title,
      summary: form.summary,
      content: form.content,
      topic_type: form.topic_type,
      stance: form.stance,
      time_horizon: form.time_horizon,
      reason_text: form.reason_text,
      risk_text: form.risk_text,
      target_type: form.target_type,
      target_id: form.target_id,
      target_snapshot: form.target_snapshot
    });
    router.replace(`/community/topics/${result.id}`);
  } catch (error) {
    errorMessage.value = error?.message || "主题提交失败";
  } finally {
    submitting.value = false;
  }
}

function validateForm() {
  if (!form.title) {
    return "请先填写标题";
  }
  if (!form.target_id) {
    return "请先填写关联对象 ID";
  }
  if (!form.reason_text) {
    return "请先填写观点依据";
  }
  if (!form.risk_text) {
    return "请先填写风险边界";
  }
  if (!form.content) {
    return "请先填写完整正文";
  }
  if (!agreed.value) {
    return "请先确认发帖说明";
  }
  return "";
}
</script>

<style scoped>
.community-compose-page {
  display: grid;
  gap: 16px;
}

.compose-hero {
  display: grid;
  grid-template-columns: minmax(0, 1fr) auto;
  gap: 16px;
  align-items: start;
}

.compose-hero-actions {
  min-width: 148px;
}

.compose-form,
.compose-section,
.compose-stack {
  display: grid;
  gap: 14px;
}

.compose-grid {
  display: grid;
  gap: 12px;
  grid-template-columns: repeat(2, minmax(0, 1fr));
}

.compose-field {
  display: grid;
  gap: 8px;
}

.compose-field-span {
  grid-column: 1 / -1;
}

.compose-field span {
  font-size: 12px;
  font-weight: 700;
  color: var(--color-text-sub);
}

.compose-field input,
.compose-field select,
.compose-field textarea {
  width: 100%;
  border: 1px solid var(--color-border-soft);
  border-radius: 14px;
  padding: 12px;
  font: inherit;
  color: var(--color-text-main);
  background: rgba(255, 255, 255, 0.94);
}

.compose-field textarea {
  min-height: 118px;
  resize: vertical;
}

.compose-agreement {
  display: flex;
  align-items: flex-start;
  gap: 10px;
  font-size: 13px;
  line-height: 1.68;
  color: var(--color-text-sub);
}

.compose-error {
  margin: 0;
  color: var(--color-danger);
  font-size: 13px;
}

.compose-submit-row {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
}

@media (max-width: 768px) {
  .compose-hero,
  .compose-grid {
    grid-template-columns: minmax(0, 1fr);
  }

  .compose-hero-actions {
    min-width: 0;
  }
}
</style>
