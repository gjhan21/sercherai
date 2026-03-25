<script setup>
import { computed, onMounted, reactive, ref } from "vue";
import { ElMessage, ElMessageBox } from "element-plus";
import {
  listCommunityComments,
  listCommunityReports,
  listCommunityTopics,
  reviewCommunityReport,
  updateCommunityCommentStatus,
  updateCommunityTopicStatus
} from "../api/admin";
import { buildCommunityTopicURL } from "../lib/community-links";
import { hasPermission } from "../lib/session";

const activeTab = ref("topics");
const loading = ref(false);
const actionLoading = ref(false);
const errorMessage = ref("");
const successMessage = ref("");

const topicPage = ref(1);
const topicPageSize = ref(20);
const topicTotal = ref(0);
const topics = ref([]);

const commentPage = ref(1);
const commentPageSize = ref(20);
const commentTotal = ref(0);
const comments = ref([]);

const reportPage = ref(1);
const reportPageSize = ref(20);
const reportTotal = ref(0);
const reports = ref([]);

const topicFilters = reactive({
  topic_type: "",
  status: "",
  user_id: ""
});

const commentFilters = reactive({
  topic_id: "",
  status: "",
  user_id: ""
});

const reportFilters = reactive({
  status: "",
  target_type: ""
});

const canEditCommunity = hasPermission("community.edit");
const canReviewCommunity = hasPermission("community.review");
const clientOrigin = String(import.meta.env.VITE_CLIENT_ORIGIN || "").trim();

const topicTypeOptions = [
  { value: "STOCK", label: "股票" },
  { value: "FUTURES", label: "期货" },
  { value: "NEWS", label: "资讯" },
  { value: "STRATEGY", label: "策略" }
];

const topicStatusOptions = [
  { value: "PUBLISHED", label: "已发布" },
  { value: "PENDING_REVIEW", label: "待审核" },
  { value: "HIDDEN", label: "已隐藏" },
  { value: "DELETED", label: "已删除" }
];

const reportStatusOptions = [
  { value: "PENDING", label: "待处理" },
  { value: "RESOLVED", label: "已处理" },
  { value: "REJECTED", label: "已驳回" }
];

const moderationSummary = computed(() => {
  const pendingTopics = topics.value.filter((item) => String(item.status || "").toUpperCase() === "PENDING_REVIEW").length;
  const pendingComments = comments.value.filter((item) => String(item.status || "").toUpperCase() === "PENDING_REVIEW").length;
  const pendingReports = reports.value.filter((item) => String(item.status || "").toUpperCase() === "PENDING").length;
  return [
    { label: "主题总数", value: topicTotal.value, note: "用于查看当前广场主题存量。" },
    { label: "评论总数", value: commentTotal.value, note: "用于查看评论侧治理压力。" },
    { label: "待处理举报", value: pendingReports, note: "优先处理内容风险问题。" },
    { label: "待审核内容", value: pendingTopics + pendingComments, note: "包含主题与评论待审核项。" }
  ];
});

function mapTopicType(value) {
  const normalized = String(value || "").toUpperCase();
  return topicTypeOptions.find((item) => item.value === normalized)?.label || normalized || "-";
}

function mapTopicStatus(value) {
  const normalized = String(value || "").toUpperCase();
  return topicStatusOptions.find((item) => item.value === normalized)?.label || normalized || "-";
}

function mapReportStatus(value) {
  const normalized = String(value || "").toUpperCase();
  return reportStatusOptions.find((item) => item.value === normalized)?.label || normalized || "-";
}

function mapTargetType(value) {
  const normalized = String(value || "").toUpperCase();
  if (normalized === "TOPIC") return "主题";
  if (normalized === "COMMENT") return "评论";
  return normalized || "-";
}

function mapStance(value) {
  const normalized = String(value || "").toUpperCase();
  if (normalized === "BULLISH") return "看多";
  if (normalized === "BEARISH") return "看空";
  if (normalized === "WATCH") return "观察";
  return normalized || "-";
}

function formatLinkedTarget(linkedTarget) {
  return linkedTarget?.target_snapshot || linkedTarget?.target_id || "-";
}

function formatDateTime(value) {
  const ts = Date.parse(value || "");
  if (Number.isNaN(ts)) {
    return value || "-";
  }
  return new Date(ts).toLocaleString("zh-CN", { hour12: false });
}

function getTopicIDForComment(row) {
  return String(row?.topic_id || "").trim();
}

function getTopicIDForReport(row) {
  const topicID = String(row?.topic_id || "").trim();
  if (topicID) {
    return topicID;
  }
  if (String(row?.target_type || "").toUpperCase() === "TOPIC") {
    return String(row?.target_id || "").trim();
  }
  return "";
}

function openOriginalDiscussion({ topicID, commentID = "" }) {
  const url = buildCommunityTopicURL({
    topicID,
    commentID,
    envOrigin: clientOrigin,
    locationOrigin: typeof window !== "undefined" ? window.location.origin : ""
  });
  if (!url) {
    ElMessage.warning("当前无法生成讨论详情链接，请先确认客户端地址配置。");
    return;
  }
  window.open(url, "_blank", "noopener,noreferrer");
}

function openCommentTopic(row) {
  const topicID = getTopicIDForComment(row);
  if (!topicID) {
    ElMessage.warning("当前评论缺少所属主题，暂时无法跳转。");
    return;
  }
  openOriginalDiscussion({ topicID, commentID: row?.id || "" });
}

function openReportedDiscussion(row) {
  const topicID = getTopicIDForReport(row);
  if (!topicID) {
    ElMessage.warning("当前举报缺少原讨论上下文，暂时无法跳转。");
    return;
  }
  const commentID =
    String(row?.target_type || "").toUpperCase() === "COMMENT" ? String(row?.target_id || "").trim() : "";
  openOriginalDiscussion({ topicID, commentID });
}

async function fetchTopics() {
  loading.value = true;
  try {
    const data = await listCommunityTopics({
      ...topicFilters,
      page: topicPage.value,
      page_size: topicPageSize.value
    });
    topics.value = data?.items || [];
    topicTotal.value = data?.total || 0;
  } catch (error) {
    errorMessage.value = error?.message || "加载社区主题失败";
  } finally {
    loading.value = false;
  }
}

async function fetchComments() {
  loading.value = true;
  try {
    const data = await listCommunityComments({
      ...commentFilters,
      page: commentPage.value,
      page_size: commentPageSize.value
    });
    comments.value = data?.items || [];
    commentTotal.value = data?.total || 0;
  } catch (error) {
    errorMessage.value = error?.message || "加载社区评论失败";
  } finally {
    loading.value = false;
  }
}

async function fetchReports() {
  loading.value = true;
  try {
    const data = await listCommunityReports({
      ...reportFilters,
      page: reportPage.value,
      page_size: reportPageSize.value
    });
    reports.value = data?.items || [];
    reportTotal.value = data?.total || 0;
  } catch (error) {
    errorMessage.value = error?.message || "加载社区举报失败";
  } finally {
    loading.value = false;
  }
}

async function refreshAll() {
  errorMessage.value = "";
  successMessage.value = "";
  await Promise.all([fetchTopics(), fetchComments(), fetchReports()]);
}

function applyTopicFilters() {
  topicPage.value = 1;
  fetchTopics();
}

function applyCommentFilters() {
  commentPage.value = 1;
  fetchComments();
}

function applyReportFilters() {
  reportPage.value = 1;
  fetchReports();
}

function resetTopicFilters() {
  topicFilters.topic_type = "";
  topicFilters.status = "";
  topicFilters.user_id = "";
  topicPage.value = 1;
  fetchTopics();
}

function resetCommentFilters() {
  commentFilters.topic_id = "";
  commentFilters.status = "";
  commentFilters.user_id = "";
  commentPage.value = 1;
  fetchComments();
}

function resetReportFilters() {
  reportFilters.status = "";
  reportFilters.target_type = "";
  reportPage.value = 1;
  fetchReports();
}

async function handleTopicStatus(id, status) {
  if (!canEditCommunity || actionLoading.value) {
    return;
  }
  actionLoading.value = true;
  errorMessage.value = "";
  successMessage.value = "";
  try {
    await updateCommunityTopicStatus(id, status);
    successMessage.value = `主题状态已更新为${mapTopicStatus(status)}`;
    await fetchTopics();
  } catch (error) {
    errorMessage.value = error?.message || "更新主题状态失败";
  } finally {
    actionLoading.value = false;
  }
}

async function handleCommentStatus(id, status) {
  if (!canEditCommunity || actionLoading.value) {
    return;
  }
  actionLoading.value = true;
  errorMessage.value = "";
  successMessage.value = "";
  try {
    await updateCommunityCommentStatus(id, status);
    successMessage.value = `评论状态已更新为${mapTopicStatus(status)}`;
    await fetchComments();
  } catch (error) {
    errorMessage.value = error?.message || "更新评论状态失败";
  } finally {
    actionLoading.value = false;
  }
}

async function handleReportDecision(row, status) {
  if (!canReviewCommunity || actionLoading.value) {
    return;
  }
  try {
    const result = await ElMessageBox.prompt(
      status === "RESOLVED" ? "可填写处理说明，帮助后续回看。" : "可填写驳回原因，帮助后续复核。",
      status === "RESOLVED" ? "处理举报" : "驳回举报",
      {
        confirmButtonText: "提交",
        cancelButtonText: "取消",
        inputPlaceholder: status === "RESOLVED" ? "例如：已隐藏内容并记录。" : "例如：举报理由不足。"
      }
    );

    actionLoading.value = true;
    errorMessage.value = "";
    successMessage.value = "";
    await reviewCommunityReport(row.id, {
      status,
      review_note: result.value || ""
    });
    successMessage.value = `举报状态已更新为${mapReportStatus(status)}`;
    await fetchReports();
  } catch (error) {
    if (error === "cancel" || error === "close") {
      return;
    }
    errorMessage.value = error?.message || "处理举报失败";
  } finally {
    actionLoading.value = false;
  }
}

onMounted(refreshAll);
</script>

<template>
  <div class="page">
    <div class="page-header">
      <div>
        <h1 class="page-title">社区审核台</h1>
        <p class="muted">统一处理讨论广场中的主题、评论与举报，当前只围绕真实支持的社区能力治理。</p>
      </div>
      <el-button :loading="loading" @click="refreshAll">刷新</el-button>
    </div>

    <el-alert
      v-if="errorMessage"
      :title="errorMessage"
      type="error"
      show-icon
      style="margin-bottom: 12px"
    />
    <el-alert
      v-if="successMessage"
      :title="successMessage"
      type="success"
      show-icon
      style="margin-bottom: 12px"
    />

    <div class="grid grid-4" style="margin-bottom: 12px">
      <div v-for="item in moderationSummary" :key="item.label" class="metric-card">
        <div class="label">{{ item.label }}</div>
        <div class="value">{{ item.value }}</div>
        <div class="hint">{{ item.note }}</div>
      </div>
    </div>

    <div class="card">
      <el-tabs v-model="activeTab">
        <el-tab-pane :label="`主题（${topicTotal}）`" name="topics">
          <div class="toolbar">
            <el-select v-model="topicFilters.topic_type" clearable placeholder="全部主题类型" style="width: 160px">
              <el-option v-for="item in topicTypeOptions" :key="item.value" :label="item.label" :value="item.value" />
            </el-select>
            <el-select v-model="topicFilters.status" clearable placeholder="全部状态" style="width: 160px">
              <el-option v-for="item in topicStatusOptions" :key="item.value" :label="item.label" :value="item.value" />
            </el-select>
            <el-input v-model="topicFilters.user_id" clearable placeholder="按用户ID筛选" style="width: 180px" />
            <el-button type="primary" plain @click="applyTopicFilters">查询</el-button>
            <el-button @click="resetTopicFilters">重置</el-button>
          </div>

          <el-table :data="topics" border stripe v-loading="loading" empty-text="暂无社区主题">
            <el-table-column prop="id" label="主题ID" min-width="180" />
            <el-table-column label="标题" min-width="220">
              <template #default="{ row }">
                <div class="text-cell">
                  <strong>{{ row.title || "-" }}</strong>
                  <span>{{ row.summary || "-" }}</span>
                </div>
              </template>
            </el-table-column>
            <el-table-column label="类型 / 方向" min-width="120">
              <template #default="{ row }">
                {{ mapTopicType(row.topic_type) }} / {{ mapStance(row.stance) }}
              </template>
            </el-table-column>
            <el-table-column prop="user_id" label="用户ID" min-width="130" />
            <el-table-column label="关联对象" min-width="180">
              <template #default="{ row }">
                {{ row.linked_target?.target_snapshot || row.linked_target?.target_id || "-" }}
              </template>
            </el-table-column>
            <el-table-column label="互动" min-width="120">
              <template #default="{ row }">
                评 {{ row.comment_count || 0 }} / 赞 {{ row.like_count || 0 }} / 藏 {{ row.favorite_count || 0 }}
              </template>
            </el-table-column>
            <el-table-column label="状态" min-width="100">
              <template #default="{ row }">
                <el-tag>{{ mapTopicStatus(row.status) }}</el-tag>
              </template>
            </el-table-column>
            <el-table-column label="最近活跃" min-width="170">
              <template #default="{ row }">
                {{ formatDateTime(row.last_active_at) }}
              </template>
            </el-table-column>
            <el-table-column label="操作" min-width="220" fixed="right">
              <template #default="{ row }">
                <div class="action-row">
                  <el-button size="small" :disabled="!canEditCommunity || actionLoading" @click="handleTopicStatus(row.id, 'PUBLISHED')">发布</el-button>
                  <el-button size="small" :disabled="!canEditCommunity || actionLoading" @click="handleTopicStatus(row.id, 'HIDDEN')">隐藏</el-button>
                  <el-button size="small" :disabled="!canEditCommunity || actionLoading" @click="handleTopicStatus(row.id, 'DELETED')">删除</el-button>
                </div>
              </template>
            </el-table-column>
          </el-table>

          <div class="pagination">
            <el-text type="info">第 {{ topicPage }} 页，共 {{ topicTotal }} 条</el-text>
            <el-pagination
              background
              layout="prev, pager, next"
              :current-page="topicPage"
              :page-size="topicPageSize"
              :total="topicTotal"
              @current-change="(value) => { topicPage = value; fetchTopics(); }"
            />
          </div>
        </el-tab-pane>

        <el-tab-pane :label="`评论（${commentTotal}）`" name="comments">
          <div class="toolbar">
            <el-input v-model="commentFilters.topic_id" clearable placeholder="按主题ID筛选" style="width: 180px" />
            <el-select v-model="commentFilters.status" clearable placeholder="全部状态" style="width: 160px">
              <el-option v-for="item in topicStatusOptions" :key="item.value" :label="item.label" :value="item.value" />
            </el-select>
            <el-input v-model="commentFilters.user_id" clearable placeholder="按用户ID筛选" style="width: 180px" />
            <el-button type="primary" plain @click="applyCommentFilters">查询</el-button>
            <el-button @click="resetCommentFilters">重置</el-button>
          </div>

          <el-table :data="comments" border stripe v-loading="loading" empty-text="暂无社区评论">
            <el-table-column prop="id" label="评论ID" min-width="180" />
            <el-table-column prop="topic_id" label="主题ID" min-width="180" />
            <el-table-column prop="user_id" label="用户ID" min-width="130" />
            <el-table-column label="讨论上下文" min-width="280">
              <template #default="{ row }">
                <div class="text-cell">
                  <strong>{{ row.topic_title || "-" }}</strong>
                  <span>{{ row.topic_summary || "暂无主题摘要" }}</span>
                  <span>关联对象：{{ formatLinkedTarget(row.linked_target) }}</span>
                  <span>主题状态：{{ mapTopicStatus(row.topic_status) }}</span>
                </div>
              </template>
            </el-table-column>
            <el-table-column label="内容" min-width="260">
              <template #default="{ row }">
                <span class="content-cell">{{ row.content || "-" }}</span>
              </template>
            </el-table-column>
            <el-table-column label="回复对象" min-width="140">
              <template #default="{ row }">
                {{ row.reply_to_user_id || row.parent_comment_id || "-" }}
              </template>
            </el-table-column>
            <el-table-column label="点赞" min-width="90">
              <template #default="{ row }">
                {{ row.like_count || 0 }}
              </template>
            </el-table-column>
            <el-table-column label="状态" min-width="100">
              <template #default="{ row }">
                <el-tag>{{ mapTopicStatus(row.status) }}</el-tag>
              </template>
            </el-table-column>
            <el-table-column label="创建时间" min-width="170">
              <template #default="{ row }">
                {{ formatDateTime(row.created_at) }}
              </template>
            </el-table-column>
            <el-table-column label="操作" min-width="220" fixed="right">
              <template #default="{ row }">
                <div class="action-row">
                  <el-button size="small" plain @click="openCommentTopic(row)">查看原讨论</el-button>
                  <el-button size="small" :disabled="!canEditCommunity || actionLoading" @click="handleCommentStatus(row.id, 'PUBLISHED')">发布</el-button>
                  <el-button size="small" :disabled="!canEditCommunity || actionLoading" @click="handleCommentStatus(row.id, 'HIDDEN')">隐藏</el-button>
                  <el-button size="small" :disabled="!canEditCommunity || actionLoading" @click="handleCommentStatus(row.id, 'DELETED')">删除</el-button>
                </div>
              </template>
            </el-table-column>
          </el-table>

          <div class="pagination">
            <el-text type="info">第 {{ commentPage }} 页，共 {{ commentTotal }} 条</el-text>
            <el-pagination
              background
              layout="prev, pager, next"
              :current-page="commentPage"
              :page-size="commentPageSize"
              :total="commentTotal"
              @current-change="(value) => { commentPage = value; fetchComments(); }"
            />
          </div>
        </el-tab-pane>

        <el-tab-pane :label="`举报（${reportTotal}）`" name="reports">
          <div class="toolbar">
            <el-select v-model="reportFilters.status" clearable placeholder="全部状态" style="width: 160px">
              <el-option v-for="item in reportStatusOptions" :key="item.value" :label="item.label" :value="item.value" />
            </el-select>
            <el-select v-model="reportFilters.target_type" clearable placeholder="全部目标类型" style="width: 160px">
              <el-option label="主题" value="TOPIC" />
              <el-option label="评论" value="COMMENT" />
            </el-select>
            <el-button type="primary" plain @click="applyReportFilters">查询</el-button>
            <el-button @click="resetReportFilters">重置</el-button>
          </div>

          <el-table :data="reports" border stripe v-loading="loading" empty-text="暂无社区举报">
            <el-table-column prop="id" label="举报ID" min-width="180" />
            <el-table-column prop="reporter_user_id" label="举报人" min-width="130" />
            <el-table-column label="目标类型" min-width="100">
              <template #default="{ row }">
                {{ mapTargetType(row.target_type) }}
              </template>
            </el-table-column>
            <el-table-column prop="target_id" label="目标ID" min-width="180" />
            <el-table-column label="目标上下文" min-width="320">
              <template #default="{ row }">
                <div class="text-cell">
                  <strong>{{ row.topic_title || "-" }}</strong>
                  <span>{{ row.topic_summary || "暂无主题摘要" }}</span>
                  <span>被举报内容：{{ row.target_content || "-" }}</span>
                  <span>关联对象：{{ formatLinkedTarget(row.linked_target) }}</span>
                  <span>目标状态：{{ mapTopicStatus(row.target_status) }}</span>
                </div>
              </template>
            </el-table-column>
            <el-table-column label="举报原因" min-width="260">
              <template #default="{ row }">
                <span class="content-cell">{{ row.reason || "-" }}</span>
              </template>
            </el-table-column>
            <el-table-column label="处理状态" min-width="110">
              <template #default="{ row }">
                <el-tag>{{ mapReportStatus(row.status) }}</el-tag>
              </template>
            </el-table-column>
            <el-table-column label="处理说明" min-width="220">
              <template #default="{ row }">
                <span class="content-cell">{{ row.review_note || "-" }}</span>
              </template>
            </el-table-column>
            <el-table-column label="创建时间" min-width="170">
              <template #default="{ row }">
                {{ formatDateTime(row.created_at) }}
              </template>
            </el-table-column>
            <el-table-column label="操作" min-width="220" fixed="right">
              <template #default="{ row }">
                <div class="action-row">
                  <el-button size="small" plain @click="openReportedDiscussion(row)">查看原讨论</el-button>
                  <el-button size="small" type="primary" :disabled="!canReviewCommunity || actionLoading" @click="handleReportDecision(row, 'RESOLVED')">处理</el-button>
                  <el-button size="small" :disabled="!canReviewCommunity || actionLoading" @click="handleReportDecision(row, 'REJECTED')">驳回</el-button>
                </div>
              </template>
            </el-table-column>
          </el-table>

          <div class="pagination">
            <el-text type="info">第 {{ reportPage }} 页，共 {{ reportTotal }} 条</el-text>
            <el-pagination
              background
              layout="prev, pager, next"
              :current-page="reportPage"
              :page-size="reportPageSize"
              :total="reportTotal"
              @current-change="(value) => { reportPage = value; fetchReports(); }"
            />
          </div>
        </el-tab-pane>
      </el-tabs>
    </div>
  </div>
</template>

<style scoped>
.text-cell {
  display: inline-block;
  line-height: 1.6;
  white-space: normal;
}

.text-cell strong,
.text-cell span {
  display: block;
}

.action-row {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
}
</style>
