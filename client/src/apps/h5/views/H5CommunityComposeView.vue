<template>
  <div class="h5-page fade-up community-compose-page">
    <div class="h5-page-topline">
      <button type="button" class="h5-topline-back" @click="router.push('/community')">返回广场</button>
      <span>发起讨论</span>
    </div>

    <H5HeroCard
      eyebrow="发起讨论"
      title="先写观点，再写依据和风险"
      description="移动端发帖需遵循结构化要求，确保每条观点都有扎实的判断逻辑。"
      tone="accent"
    />

    <form class="h5-form-stack" @submit.prevent="submitTopic">
      <H5SectionBlock eyebrow="基本信息" title="标题与分类" tone="soft">
        <div class="h5-field">
          <label>讨论标题</label>
          <input v-model.trim="form.title" type="text" maxlength="60" placeholder="例如：白酒龙头当前位置更适合观察" />
        </div>
        
        <div class="h5-grid-2">
          <div class="h5-field">
            <label>分类</label>
            <select v-model="form.topic_type">
              <option v-for="item in topicTypeOptions" :key="item.value" :value="item.value">{{ item.label }}</option>
            </select>
          </div>
          <div class="h5-field">
            <label>方向</label>
            <select v-model="form.stance">
              <option v-for="item in stanceOptions" :key="item.value" :value="item.value">{{ item.label }}</option>
            </select>
          </div>
        </div>

        <div class="h5-field">
          <label>预期时间</label>
          <select v-model="form.time_horizon">
            <option v-for="item in horizonOptions" :key="item.value" :value="item.value">{{ item.label }}</option>
          </select>
        </div>
      </H5SectionBlock>

      <H5SectionBlock eyebrow="关联对象" title="讨论指向的具体内容" tone="soft">
        <div class="h5-field">
          <label>类型</label>
          <select v-model="form.target_type">
            <option v-for="item in targetTypeOptions" :key="item.value" :value="item.value">{{ item.label }}</option>
          </select>
        </div>
        <div class="h5-field">
          <label>对象 ID / 快照</label>
          <input v-model.trim="form.target_id" type="text" placeholder="ID (如 600519)" />
          <input v-model.trim="form.target_snapshot" type="text" placeholder="快照标题 (如 贵州茅台)" />
        </div>
      </H5SectionBlock>

      <H5SectionBlock eyebrow="观点核心" title="摘要、理由与风险" tone="accent">
        <div class="h5-field">
          <label>摘要 (一句话观点)</label>
          <textarea v-model.trim="form.summary" maxlength="160" placeholder="概括你的核心判断..." />
        </div>

        <div class="h5-field">
          <label>为什么这样看 (依据)</label>
          <textarea v-model.trim="form.reason_text" maxlength="240" placeholder="列出你的逻辑、数据或资讯支持..." />
        </div>

        <div class="h5-field">
          <label>风险边界 (失效条件)</label>
          <textarea v-model.trim="form.risk_text" maxlength="240" placeholder="什么情况下你的观点会错误..." />
        </div>

        <div class="h5-field">
          <label>完整正文</label>
          <textarea v-model.trim="form.content" maxlength="2000" placeholder="展开描述你的详细分析..." />
        </div>
      </H5SectionBlock>

      <div class="h5-agreement-box">
        <label class="h5-checkbox-label">
          <input v-model="agreed" type="checkbox" />
          <span>我确认已写明观点依据和风险边界。</span>
        </label>
      </div>

      <p v-if="errorMessage" class="h5-error-msg">{{ errorMessage }}</p>

      <div class="h5-form-actions">
        <button type="submit" class="h5-btn" :disabled="submitting">
          {{ submitting ? "提交中..." : "确认发布" }}
        </button>
        <button type="button" class="h5-btn-ghost" @click="router.push('/community')">暂不发布</button>
      </div>
    </form>
  </div>
</template>

<script setup>
import { reactive, ref } from "vue";
import { useRouter, useRoute } from "vue-router";
import H5HeroCard from "../components/H5HeroCard.vue";
import H5SectionBlock from "../components/H5SectionBlock.vue";
import { createCommunityTopic } from "../../../api/community";

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
  if (errorMessage.value) return;

  submitting.value = true;
  try {
    const result = await createCommunityTopic({ ...form });
    router.replace(`/community/topics/${result.id}`);
  } catch (error) {
    errorMessage.value = error?.message || "主题提交失败";
  } finally {
    submitting.value = false;
  }
}

function validateForm() {
  if (!form.title) return "请填写标题";
  if (!form.target_id) return "请填写关联对象 ID";
  if (!form.reason_text) return "请填写观点依据";
  if (!form.risk_text) return "请填写风险边界";
  if (!form.content) return "请填写完整正文";
  if (!agreed.value) return "请勾选确认声明";
  return "";
}
</script>

<style scoped>
.community-compose-page {
  gap: 12px;
}
.h5-topline-back {
  background: none;
  border: 0;
  padding: 0;
  color: var(--h5-brand, #102a56);
  font-size: 13px;
  font-weight: 700;
}
.h5-form-stack {
  display: grid;
  gap: 12px;
}
.h5-field {
  display: grid;
  gap: 6px;
  margin-bottom: 12px;
}
.h5-field:last-child {
  margin-bottom: 0;
}
.h5-field label {
  font-size: 11px;
  font-weight: 700;
  color: var(--h5-text-soft, #999);
}
.h5-field input,
.h5-field select,
.h5-field textarea {
  width: 100%;
  border: 1px solid var(--h5-line, #eee);
  border-radius: 12px;
  padding: 10px 14px;
  font-size: 14px;
  background: #fff;
  color: var(--h5-text, #333);
}
.h5-field textarea {
  min-height: 80px;
  resize: none;
}
.h5-agreement-box {
  padding: 0 16px;
}
.h5-checkbox-label {
  display: flex;
  align-items: center;
  gap: 10px;
  font-size: 13px;
  color: var(--h5-text-sub, #666);
}
.h5-error-msg {
  color: var(--h5-danger, #e64545);
  font-size: 12px;
  padding: 0 16px;
}
.h5-form-actions {
  padding: 16px;
  display: grid;
  gap: 10px;
}
</style>
