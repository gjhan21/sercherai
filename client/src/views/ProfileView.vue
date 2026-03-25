<template>
  <section class="profile-page fade-up">
    <header class="account-card card">
      <div class="profile-hero-copy">
        <div class="finance-pill-row">
          <span class="finance-pill finance-pill-compact finance-pill-neutral">我的</span>
          <span class="finance-pill finance-pill-compact finance-pill-info">账户管理台</span>
          <span class="finance-pill finance-pill-compact finance-pill-info">待办优先</span>
        </div>
        <div class="identity">
          <div class="avatar">U</div>
          <div>
            <h1>{{ displayProfile.name }}</h1>
            <p>
              {{ vipInfo.level }} · {{ activationStateLabel }} · KYC {{ displayProfile.kycStatus }} · 最近更新
              {{ lastUpdatedAt || "-" }}
            </p>
          </div>
        </div>
      </div>
      <div class="actions">
        <button class="primary finance-primary-btn" type="button" @click="loadUserCenterData">刷新数据</button>
        <button class="ghost finance-ghost-btn" type="button" @click="openSecurityPanel">账户安全</button>
      </div>
      <div class="profile-hero-stats finance-hero-stat-grid">
        <article class="finance-hero-stat-card">
          <span>账户身份</span>
          <strong>{{ displayProfile.name }}</strong>
          <p>个人中心先看身份、会员与实名状态，再进入查询中心。</p>
        </article>
        <article class="finance-hero-stat-card">
          <span>会员状态</span>
          <strong>{{ vipInfo.level }} · {{ activationStateLabel }}</strong>
          <p>KYC {{ displayProfile.kycStatus }} · 当前阅读和支付链会按此状态承接。</p>
        </article>
        <article class="finance-hero-stat-card">
          <span>未读消息</span>
          <strong>{{ unreadMessageCount }} 条</strong>
          <p>高频账户确认和支付提醒优先在今日行动板处理。</p>
        </article>
        <article class="finance-hero-stat-card">
          <span>今日优先</span>
          <strong>{{ todos[0]?.title || "查看账户待办" }}</strong>
          <p>{{ todos[0]?.note || "先看待办中心，再进入订单、订阅和邀请管理。" }}</p>
        </article>
      </div>
    </header>

    <StatePanel
      :tone="profileRhythmStatus.tone"
      :eyebrow="profileRhythmStatus.eyebrow"
      :title="profileRhythmStatus.title"
      :description="profileRhythmStatus.desc"
    >
      <template #actions>
        <button type="button" class="finance-primary-btn" @click="handleAction(profileRhythmStatus.primaryAction)">
          {{ profileRhythmStatus.primaryAction.label }}
        </button>
        <button class="ghost finance-ghost-btn" type="button" @click="handleAction(profileRhythmStatus.secondaryAction)">
          {{ profileRhythmStatus.secondaryAction.label }}
        </button>
      </template>
    </StatePanel>

    <section class="profile-workbench-layout finance-dual-rail">
      <div class="profile-main-stack finance-stack-tight">
        <article class="card profile-focus-card finance-section-card">
          <header class="profile-focus-head finance-section-head-grid">
            <div>
              <p class="section-kicker">账户身份概览</p>
              <h2 class="section-title">先查看账户状态和待办，再进入查询中心。</h2>
              <p class="section-subtitle">
                查看账户状态、今日重点和待处理事项。
              </p>
            </div>
            <div class="profile-focus-actions finance-action-row">
              <button class="primary finance-primary-btn" type="button" @click="loadUserCenterData">刷新账户数据</button>
              <button class="ghost finance-ghost-btn" type="button" @click="handleAction(profileRhythmStatus.primaryAction)">
                {{ profileRhythmStatus.primaryAction.label }}
              </button>
            </div>
          </header>

          <div class="profile-overview-grid finance-card-grid finance-card-grid-2">
            <article v-for="item in profileOverviewRows" :key="item.label" class="finance-card-surface">
              <p>{{ item.label }}</p>
              <strong>{{ item.value }}</strong>
              <span>{{ item.note }}</span>
            </article>
          </div>

          <div class="profile-guide-grid finance-card-grid finance-card-grid-3">
            <article v-for="item in profileGuideRows" :key="item.title" class="finance-card-surface">
              <strong>{{ item.title }}</strong>
              <p>{{ item.desc }}</p>
            </article>
          </div>
        </article>

        <article class="card rhythm-card">
          <header class="rhythm-head">
            <div>
              <p class="section-kicker">今日行动板</p>
              <h2 class="section-title">根据今日节奏安排查看顺序。</h2>
              <p class="section-subtitle">
                08:30 看主推荐，11:30 看资讯，15:30 回关注清单，周末做历史复盘。
              </p>
            </div>
            <div class="rhythm-pill finance-summary-pill">
              <p>今日节奏</p>
              <strong>{{ vipInfo.level }}</strong>
              <small>{{ vipInfo.status }} · 未读 {{ unreadMessageCount }} 条</small>
            </div>
          </header>

          <div class="rhythm-grid">
            <article v-for="entry in profileCadenceEntries" :key="entry.slot" class="rhythm-item finance-card-surface">
              <p class="rhythm-slot">{{ entry.slot }}</p>
              <h3>{{ entry.title }}</h3>
              <p class="rhythm-desc">{{ entry.desc }}</p>
              <div class="rhythm-tags">
                <span class="finance-pill finance-pill-compact finance-pill-info">{{ entry.highlight }}</span>
                <span class="finance-pill finance-pill-compact finance-pill-neutral">{{ entry.supporting }}</span>
              </div>
              <div class="rhythm-actions">
                <button type="button" class="finance-primary-btn" @click="handleAction(entry.primaryAction)">
                  {{ entry.primaryAction.label }}
                </button>
                <button class="ghost finance-ghost-btn" type="button" @click="handleAction(entry.secondaryAction)">
                  {{ entry.secondaryAction.label }}
                </button>
              </div>
            </article>
          </div>
        </article>

        <article class="card todo-card">
          <header class="finance-copy-stack">
            <h2 class="section-title">待办中心</h2>
            <p class="section-subtitle">把高频操作前置，减少跳转。</p>
          </header>
          <ul>
            <li v-for="todo in todos" :key="todo.title" class="finance-list-card finance-list-card-panel">
              <span class="dot" :class="todo.level" />
              <div>
                <p class="title">{{ todo.title }}</p>
                <p class="note">{{ todo.note }}</p>
              </div>
              <div class="todo-actions">
                <button type="button" class="finance-mini-btn finance-mini-btn-soft" @click="handleAction(todo.action)">
                  {{ todo.actionLabel }}
                </button>
              </div>
            </li>
          </ul>
        </article>

        <article class="card query-card">
      <header class="query-head">
        <div>
          <h2 class="section-title">账户确认与管理中心</h2>
          <p class="section-subtitle">核对 VIP、支付、阅读、订阅、通知与邀请信息。</p>
        </div>
        <div class="range-switch">
          <button
            v-for="range in timeRanges"
            :key="range"
            type="button"
            class="finance-toggle-btn"
            :class="{ active: activeRange === range }"
            @click="activeRange = range"
          >
            {{ range }}
          </button>
        </div>
      </header>

      <nav class="query-nav">
        <button
          v-for="item in modules"
          :key="item.key"
          type="button"
          class="finance-toggle-btn finance-toggle-btn-block"
          :class="{ active: activeModule === item.key }"
          @click="activeModule = item.key"
        >
          {{ item.label }}
        </button>
      </nav>

      <div class="query-tip finance-info-box">
        <p>当前查询：{{ currentModule.label }}</p>
        <p>时间范围：{{ activeRange }}</p>
      </div>

      <div v-if="loading" class="state-box finance-note-strip finance-note-strip-info">正在加载账户数据...</div>
      <div v-else-if="loadError" class="state-box finance-note-strip finance-note-strip-warning">数据加载失败：{{ loadError }}</div>

      <section class="query-body">
        <template v-if="activeModule === 'vip'">
          <div class="vip-panel">
            <article class="vip-main">
              <p class="vip-level">{{ vipInfo.level }}</p>
              <h3>{{ isPaidPendingKYC ? activationPromptTitle : `VIP 有效期至 ${vipInfo.expireAt}` }}</h3>
              <p>
                {{
                  isPaidPendingKYC
                    ? activationPromptDesc
                    : `下次续费时间：${vipInfo.nextRenewAt}，剩余 ${vipInfo.remainingDays} 天`
                }}
              </p>
            </article>
            <div class="summary-grid">
              <article v-for="item in vipMetrics" :key="item.label" class="finance-summary-pill">
                <p>{{ item.label }}</p>
                <strong>{{ item.value }}</strong>
              </article>
            </div>
          </div>
          <div class="benefits-grid">
            <article v-for="item in vipBenefits" :key="item.title" class="finance-card-surface">
              <h4>{{ item.title }}</h4>
              <p>{{ item.desc }}</p>
            </article>
          </div>
          <div v-if="isPaidPendingKYC" class="activation-panel">
            <div class="activation-copy finance-card-surface">
              <p class="section-kicker">待实名激活</p>
              <h4>{{ activationPromptTitle }}</h4>
              <p>{{ activationPromptDesc }}</p>
              <div class="activation-tags">
                <span class="finance-pill finance-pill-roomy finance-pill-info">会员等级 {{ vipInfo.level }}</span>
                <span class="finance-pill finance-pill-roomy finance-pill-info">激活状态 {{ activationStateLabel }}</span>
                <span class="finance-pill finance-pill-roomy finance-pill-info">KYC {{ displayProfile.kycStatus }}</span>
              </div>
            </div>
            <div class="activation-form-wrap finance-card-surface">
              <p v-if="kycActionError" class="state-box finance-note-strip finance-note-strip-warning">{{ kycActionError }}</p>
              <p v-else-if="kycActionMessage" class="state-box finance-note-strip finance-note-strip-info">{{ kycActionMessage }}</p>
              <p v-if="!canSubmitKYC" class="state-box finance-note-strip finance-note-strip-info">
                {{
                  currentKYCStatusRaw === "PENDING"
                    ? "实名材料已提交，审核通过后会自动激活高级权益。"
                    : "当前状态无需重复提交实名材料。"
                }}
              </p>
              <form v-else class="kyc-form" @submit.prevent="handleSubmitKYC">
                <label>
                  真实姓名
                  <input v-model.trim="kycForm.real_name" placeholder="请输入真实姓名" />
                </label>
                <label>
                  身份证号
                  <input v-model.trim="kycForm.id_number" placeholder="请输入身份证号" />
                </label>
                <button type="submit" :disabled="kycSubmitting">
                  {{ kycSubmitting ? "提交中..." : "提交实名信息" }}
                </button>
              </form>
            </div>
          </div>
        </template>

        <template v-else-if="activeModule === 'payment'">
          <div class="summary-grid">
            <article v-for="item in paymentSummary" :key="item.label" class="finance-summary-pill">
              <p>{{ item.label }}</p>
              <strong>{{ item.value }}</strong>
            </article>
          </div>

          <div class="payment-table-wrap finance-table-wrap">
            <table class="payment-table finance-data-table">
              <thead>
                <tr>
                  <th>订单号</th>
                  <th>时间</th>
                  <th>项目</th>
                  <th>金额</th>
                  <th>方式</th>
                  <th>状态</th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="item in paymentRecords" :key="item.orderNo">
                  <td>{{ item.orderNo }}</td>
                  <td>{{ item.time }}</td>
                  <td>{{ item.product }}</td>
                  <td>{{ item.amount }}</td>
                  <td>{{ item.method }}</td>
                  <td>
                    <span class="status finance-pill finance-pill-compact" :class="paymentStatusClass(item.status)">
                      {{ item.status }}
                    </span>
                  </td>
                </tr>
              </tbody>
            </table>
          </div>

          <div class="payment-mobile">
            <article v-for="item in paymentRecords" :key="`m-${item.orderNo}`" class="finance-list-card finance-list-card-panel">
              <div class="top-line">
                <p>{{ item.product }}</p>
                <span>{{ item.amount }}</span>
              </div>
              <div class="meta-line finance-meta-line">
                <span>{{ item.time }}</span>
                <span>{{ item.method }}</span>
                <span class="status finance-pill finance-pill-compact" :class="paymentStatusClass(item.status)">{{ item.status }}</span>
              </div>
              <p class="order">订单号：{{ item.orderNo }}</p>
            </article>
          </div>
        </template>

        <template v-else-if="activeModule === 'reading'">
          <div class="summary-grid">
            <article v-for="item in readingStats" :key="item.label" class="finance-summary-pill">
              <p>{{ item.label }}</p>
              <strong>{{ item.value }}</strong>
            </article>
          </div>

          <div class="log-list">
            <article v-for="item in readingLogs" :key="item.id" class="finance-list-card finance-list-card-panel">
              <div class="top-line">
                <p>{{ item.title }}</p>
                <span>{{ item.type }}</span>
              </div>
              <p class="desc">{{ item.desc }}</p>
              <div class="meta-line finance-meta-line">
                <span>{{ item.time }}</span>
                <span>阅读时长 {{ item.duration }}</span>
                <span>完成度 {{ item.progress }}</span>
              </div>
            </article>
          </div>
        </template>

        <template v-else-if="activeModule === 'subscription'">
          <div class="subscription-create finance-card-surface">
            <label>
              订阅类型
              <select v-model="newSubscriptionForm.type">
                <option v-for="item in subscriptionTypeOptions" :key="item.value" :value="item.value">
                  {{ item.label }}
                </option>
              </select>
            </label>
            <label>
              推送频率
              <select v-model="newSubscriptionForm.frequency">
                <option v-for="item in subscriptionFrequencyOptions" :key="item.value" :value="item.value">
                  {{ item.label }}
                </option>
              </select>
            </label>
            <label class="scope-input">
              订阅范围
              <input v-model.trim="newSubscriptionForm.scope" placeholder="如：ALL / A股 / 沪深300" />
            </label>
            <button
              type="button"
              class="finance-mini-btn finance-mini-btn-primary"
              :disabled="creatingSubscription"
              @click="handleCreateSubscription"
            >
              {{ creatingSubscription ? "创建中..." : "新增订阅" }}
            </button>
          </div>
          <p v-if="subscriptionActionError" class="state-box finance-note-strip finance-note-strip-warning">{{ subscriptionActionError }}</p>
          <p
            v-else-if="subscriptionActionMessage"
            class="state-box finance-note-strip finance-note-strip-info"
          >
            {{ subscriptionActionMessage }}
          </p>

          <div class="subscription-grid">
            <article v-for="item in subscriptionItems" :key="item.id" class="subscription-item finance-list-card">
              <div class="top-line">
                <p>{{ item.name }}</p>
                <span class="status finance-pill finance-pill-compact" :class="subscriptionStatusClass(item.status)">
                  {{ item.status }}
                </span>
              </div>
              <p class="desc">{{ item.desc }}</p>
              <div class="meta-line finance-meta-line">
                <span>周期：{{ item.cycle }}</span>
                <span>范围：{{ item.scope }}</span>
                <span>{{ item.price }}</span>
              </div>
              <div class="subscription-actions">
                <button
                  type="button"
                  class="secondary finance-mini-btn finance-mini-btn-soft"
                  :disabled="item.saving"
                  @click="handleRotateSubscriptionFrequency(item)"
                >
                  {{ item.saving ? "处理中..." : "切换频率" }}
                </button>
                <button
                  type="button"
                  class="finance-mini-btn finance-mini-btn-primary"
                  :disabled="item.saving"
                  @click="handleToggleSubscriptionStatus(item)"
                >
                  {{ item.statusRaw === "ACTIVE" ? "暂停订阅" : "恢复订阅" }}
                </button>
              </div>
            </article>
          </div>
        </template>

        <template v-else-if="activeModule === 'message'">
          <div class="summary-grid">
            <article v-for="item in messageStats" :key="item.label" class="finance-summary-pill">
              <p>{{ item.label }}</p>
              <strong>{{ item.value }}</strong>
            </article>
          </div>

          <div class="message-list">
            <article v-for="item in messageItems" :key="item.id" class="message-item finance-list-card">
              <div class="top-line">
                <p>{{ item.title }}</p>
                <span class="status finance-pill finance-pill-compact" :class="item.readStatusRaw === 'READ' ? 'success' : 'pending'">
                  {{ item.readStatus }}
                </span>
              </div>
              <p class="desc">{{ item.content }}</p>
              <div class="meta-line finance-meta-line">
                <span>{{ item.type }}</span>
                <span>{{ item.time }}</span>
              </div>
              <div class="message-actions">
                <button
                  type="button"
                  class="secondary finance-mini-btn finance-mini-btn-soft"
                  :disabled="item.readStatusRaw === 'READ' || item.loading"
                  @click="handleReadMessage(item)"
                >
                  {{ item.readStatusRaw === "READ" ? "已读" : item.loading ? "处理中..." : "标记已读" }}
                </button>
              </div>
            </article>
          </div>
        </template>

        <template v-else-if="activeModule === 'invite'">
          <div class="summary-grid">
            <article v-for="item in inviteStats" :key="item.label" class="finance-summary-pill">
              <p>{{ item.label }}</p>
              <strong>{{ item.value }}</strong>
            </article>
          </div>

          <div class="other-grid">
            <article class="finance-card-surface">
              <h4>我的注册来源</h4>
              <div class="kv-list">
                <p>
                  <span>注册来源</span>
                  <strong>{{ inviteSourceInfo.registrationSource }}</strong>
                </p>
                <p>
                  <span>邀请人ID</span>
                  <strong>{{ inviteSourceInfo.inviterUserID }}</strong>
                </p>
                <p>
                  <span>邀请码</span>
                  <strong>{{ inviteSourceInfo.inviteCode }}</strong>
                </p>
                <p>
                  <span>注册绑定时间</span>
                  <strong>{{ inviteSourceInfo.invitedAt }}</strong>
                </p>
              </div>
            </article>
            <article class="finance-card-surface">
              <h4>我的分享链接</h4>
              <div class="invite-create finance-card-surface">
                <label>
                  渠道
                  <select v-model="newShareLinkChannel">
                    <option v-for="item in shareChannelOptions" :key="item.value" :value="item.value">
                      {{ item.label }}
                    </option>
                  </select>
                </label>
                <button
                  type="button"
                  class="finance-mini-btn finance-mini-btn-primary"
                  :disabled="creatingShareLink"
                  @click="handleCreateShareLink"
                >
                  {{ creatingShareLink ? "创建中..." : "新增分享链接" }}
                </button>
              </div>
              <p v-if="inviteActionError" class="state-box finance-note-strip finance-note-strip-warning">{{ inviteActionError }}</p>
              <p v-else-if="inviteActionMessage" class="state-box finance-note-strip finance-note-strip-info">{{ inviteActionMessage }}</p>
              <div class="kv-list">
                <p v-if="shareLinks.length === 0">
                  <span>链接状态</span>
                  <strong>暂无分享链接</strong>
                </p>
                <p v-for="item in shareLinks" :key="item.id" class="invite-link-row">
                  <span>{{ item.code }} · {{ mapShareChannel(item.channel) }} · {{ item.status }}</span>
                  <strong>
                    <button
                      type="button"
                      class="finance-mini-btn finance-mini-btn-soft"
                      :disabled="item.copying"
                      @click="handleCopyInviteLink(item)"
                    >
                      {{ item.copying ? "复制中..." : "复制链接" }}
                    </button>
                  </strong>
                </p>
              </div>
            </article>
          </div>

          <div class="payment-table-wrap finance-table-wrap">
            <table class="payment-table finance-data-table">
              <thead>
                <tr>
                  <th>被邀请用户</th>
                  <th>注册时间</th>
                  <th>首单支付</th>
                  <th>状态</th>
                  <th>风控</th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="item in inviteRecords" :key="item.id">
                  <td>{{ item.inviteeUser }}</td>
                  <td>{{ item.registerAt }}</td>
                  <td>{{ item.firstPayAt }}</td>
                  <td>{{ item.status }}</td>
                  <td>{{ item.riskFlag }}</td>
                </tr>
              </tbody>
            </table>
          </div>

          <div class="payment-mobile">
            <article v-if="inviteRecords.length === 0" class="finance-list-card finance-list-card-panel">
              <div class="top-line">
                <p>暂无邀请记录</p>
              </div>
            </article>
            <article v-for="item in inviteRecords" :key="`invite-${item.id}`" class="finance-list-card finance-list-card-panel">
              <div class="top-line">
                <p>{{ item.inviteeUser }}</p>
                <span>{{ item.status }}</span>
              </div>
              <div class="meta-line">
                <span>注册：{{ item.registerAt }}</span>
                <span>首单：{{ item.firstPayAt }}</span>
                <span>风控：{{ item.riskFlag }}</span>
              </div>
            </article>
          </div>
        </template>

        <template v-else>
          <div class="other-grid">
            <article v-for="item in otherInfos" :key="item.title" class="finance-card-surface">
              <h4>{{ item.title }}</h4>
              <div class="kv-list">
                <p v-for="row in item.rows" :key="`${item.title}-${row.key}`">
                  <span>{{ row.key }}</span>
                  <strong>{{ row.value }}</strong>
                </p>
              </div>
            </article>
          </div>
        </template>
      </section>
        </article>

        <article class="card quick-card">
          <header class="finance-copy-stack">
            <h2 class="section-title">快捷入口</h2>
            <p class="section-subtitle">常用操作统一收敛到个人中心。</p>
          </header>
          <div class="quick-grid">
            <article v-for="item in quickActions" :key="item.title" class="quick-item finance-card-surface">
              <h3>{{ item.title }}</h3>
              <p>{{ item.desc }}</p>
              <button type="button" class="finance-mini-btn finance-mini-btn-soft" @click="handleAction(item.action)">
                {{ item.actionLabel }}
              </button>
            </article>
          </div>
        </article>
      </div>

      <aside class="profile-side-rail finance-stack-tight finance-sticky-side">
        <article class="card profile-side-card finance-section-card">
          <header class="section-head compact">
            <div>
              <h2 class="section-title">账户摘要</h2>
              <p class="section-subtitle">会员、实名和消息状态在侧栏长期可见。</p>
            </div>
          </header>
          <div class="profile-side-list finance-card-stack">
            <article v-for="item in profileAccountSummaryRows" :key="item.label" class="finance-card-surface">
              <strong>{{ item.label }}</strong>
              <p>{{ item.value }}</p>
            </article>
          </div>
        </article>

        <article class="card profile-side-card finance-section-card">
          <header class="section-head compact">
            <div>
              <h2 class="section-title">待处理事项</h2>
              <p class="section-subtitle">把今天最影响回访节奏的事项收在这里。</p>
            </div>
          </header>
          <div class="profile-side-list finance-card-stack">
            <article v-for="item in profilePendingRows" :key="item.title" class="finance-card-surface">
              <strong>{{ item.title }}</strong>
              <p>{{ item.desc }}</p>
            </article>
          </div>
        </article>

        <article class="card profile-side-card finance-section-card">
          <header class="section-head compact">
            <div>
              <h2 class="section-title">状态快照</h2>
              <p class="section-subtitle">持续盯住会员、实名、通知和订阅，不用反复切模块。</p>
            </div>
          </header>
          <div class="profile-status-grid finance-card-grid finance-card-grid-2">
            <article v-for="item in profileStatusRows" :key="item.label" class="finance-card-surface">
              <p>{{ item.label }}</p>
              <strong>{{ item.value }}</strong>
              <span>{{ item.note }}</span>
            </article>
          </div>
        </article>
      </aside>
    </section>
  </section>
</template>

<script setup>
import { computed, onMounted, ref } from "vue";
import { useRouter } from "vue-router";
import StatePanel from "../components/StatePanel.vue";
import {
  createShareLink,
  createSubscription,
  getInviteSummary,
  getMembershipQuota,
  getUserProfile,
  listBrowseHistory,
  listInviteRecords,
  listMessages,
  listMembershipOrders,
  listRechargeRecords,
  listShareLinks,
  listSubscriptions,
  readMessage,
  submitKYC,
  updateSubscription
} from "../api/userCenter";
import { shouldUseDemoFallback } from "../lib/fallback-policy";
import {
  modules,
  shareChannelOptions,
  subscriptionFrequencyOptions,
  subscriptionTypeOptions,
  timeRanges
} from "./profile/constants";
import {
  buildInviteURL,
  copyText,
  formatAmount,
  formatDateTime,
  inRange,
  mapActivationState,
  mapContentType,
  mapInviteLinkStatus,
  mapInviteRiskFlag,
  mapInviteStatus,
  mapKYCStatus,
  mapMemberLevel,
  mapMessageReadStatus,
  mapMessageType,
  mapPayChannel,
  mapPaymentStatus,
  mapProduct,
  mapRegistrationSource,
  mapResetCycle,
  mapShareChannel,
  mapSubscriptionFrequency,
  mapSubscriptionStatus,
  mapSubscriptionType,
  mapVIPStatus,
  nextSubscriptionFrequency,
  paymentStatusClass,
  subscriptionStatusClass,
  toArray,
  toTimestamp
} from "./profile/helpers";
import {
  fallbackBrowseHistory,
  fallbackInviteRecords,
  fallbackInviteSummary,
  fallbackMembershipOrders,
  fallbackMessages,
  fallbackProfile,
  fallbackQuota,
  fallbackRechargeRecords,
  fallbackShareLinks,
  fallbackSubscriptions
} from "./profile/fallback";

const useDemoFallback = shouldUseDemoFallback();
const router = useRouter();

const activeModule = ref("vip");
const activeRange = ref(timeRanges[1] || timeRanges[0] || "全部");

const loading = ref(false);
const loadError = ref("");
const lastUpdatedAt = ref("");
const creatingSubscription = ref(false);
const subscriptionActionMessage = ref("");
const subscriptionActionError = ref("");
const subscriptionSavingMap = ref({});
const messageActionLoadingMap = ref({});
const creatingShareLink = ref(false);
const inviteActionMessage = ref("");
const inviteActionError = ref("");
const copyInviteID = ref("");
const newShareLinkChannel = ref(shareChannelOptions[0]?.value || "APP");

const newSubscriptionForm = ref({
  type: subscriptionTypeOptions[0]?.value || "STOCK_RECO",
  frequency: "DAILY",
  scope: "ALL"
});
const kycForm = ref({
  real_name: "",
  id_number: ""
});
const kycSubmitting = ref(false);
const kycActionMessage = ref("");
const kycActionError = ref("");

const rawProfile = ref(useDemoFallback ? { ...fallbackProfile } : {});
const rawQuota = ref(useDemoFallback ? { ...fallbackQuota } : {});
const rawMembershipOrders = ref(useDemoFallback ? [...fallbackMembershipOrders] : []);
const rawRechargeRecords = ref(useDemoFallback ? [...fallbackRechargeRecords] : []);
const rawBrowseHistory = ref(useDemoFallback ? [...fallbackBrowseHistory] : []);
const rawSubscriptions = ref(useDemoFallback ? [...fallbackSubscriptions] : []);
const rawMessages = ref(useDemoFallback ? [...fallbackMessages] : []);
const rawShareLinks = ref(useDemoFallback ? [...fallbackShareLinks] : []);
const rawInviteRecords = ref(useDemoFallback ? [...fallbackInviteRecords] : []);
const rawInviteSummary = ref(useDemoFallback ? { ...fallbackInviteSummary } : {});

const currentModule = computed(() => modules.find((item) => item.key === activeModule.value) || modules[0]);
const currentKYCStatusRaw = computed(() =>
  String(rawQuota.value?.kyc_status || rawProfile.value?.kyc_status || "").toUpperCase()
);
const currentActivationState = computed(() => {
  const activationState = String(
    rawProfile.value?.activation_state || rawQuota.value?.activation_state || ""
  ).toUpperCase();
  if (activationState) {
    return activationState;
  }
  const level = String(rawProfile.value?.member_level || rawQuota.value?.member_level || "").toUpperCase();
  if (!level.startsWith("VIP")) {
    return "NON_MEMBER";
  }
  const vipStatus = String(rawQuota.value?.vip_status || rawProfile.value?.vip_status || "").toUpperCase();
  if (vipStatus === "EXPIRED") {
    return "NON_MEMBER";
  }
  return currentKYCStatusRaw.value === "APPROVED" || currentKYCStatusRaw.value === "VERIFIED"
    ? "ACTIVE"
    : "PAID_PENDING_KYC";
});
const activationStateLabel = computed(() => mapActivationState(currentActivationState.value));
const isPaidPendingKYC = computed(() => currentActivationState.value === "PAID_PENDING_KYC");
const canSubmitKYC = computed(() => isPaidPendingKYC.value && currentKYCStatusRaw.value !== "PENDING");
const activationPromptTitle = computed(() =>
  currentKYCStatusRaw.value === "REJECTED"
    ? "实名未通过，请重新提交后激活高级权益"
    : "会员已开通，待实名激活高级权益"
);
const activationPromptDesc = computed(() => {
  if (currentKYCStatusRaw.value === "PENDING") {
    return "实名材料已提交，审核通过后会自动激活完整策略档案、VIP 资讯、盘中跟踪与复盘能力。";
  }
  if (currentKYCStatusRaw.value === "REJECTED") {
    return "你的会员资格仍然保留，但高级权益会继续冻结，直到重新提交实名并审核通过。";
  }
  return "你已经完成会员支付，但完整策略档案、VIP 资讯和跟踪能力会在实名通过后统一激活。";
});

const displayProfile = computed(() => ({
  name: rawProfile.value?.id ? `用户 ${rawProfile.value.id}` : "当前用户",
  phone: rawProfile.value?.phone || "-",
  email: rawProfile.value?.email || "-",
  kycStatus: mapKYCStatus(currentKYCStatusRaw.value),
  memberLevel: rawProfile.value?.member_level || rawQuota.value?.member_level || "FREE"
}));

const paymentRecords = computed(() => {
  const orders = (rawMembershipOrders.value || []).map((item) => ({
    source: "membership",
    orderNo: item.order_no || item.id || "-",
    time: formatDateTime(item.paid_at || item.created_at),
    rawTime: item.paid_at || item.created_at || "",
    product: mapProduct(item.product_id),
    amount: formatAmount(item.amount),
    amountValue: Number(item.amount || 0),
    method: mapPayChannel(item.pay_channel),
    status: mapPaymentStatus(item.status)
  }));

  const recharges = (rawRechargeRecords.value || []).map((item) => ({
    source: "recharge",
    orderNo: item.order_no || item.id || "-",
    time: formatDateTime(item.paid_at || item.created_at),
    rawTime: item.paid_at || item.created_at || "",
    product: "账户充值",
    amount: formatAmount(item.amount),
    amountValue: Number(item.amount || 0),
    method: mapPayChannel(item.pay_channel),
    status: mapPaymentStatus(item.status)
  }));

  return [...orders, ...recharges]
    .filter((item) => inRange(item.rawTime, activeRange.value))
    .sort((a, b) => toTimestamp(b.rawTime) - toTimestamp(a.rawTime));
});

const readingLogs = computed(() =>
  (rawBrowseHistory.value || [])
    .filter((item) => inRange(item.viewed_at, activeRange.value))
    .sort((a, b) => toTimestamp(b.viewed_at) - toTimestamp(a.viewed_at))
    .map((item) => ({
      id: item.id,
      title: item.title || "未命名内容",
      type: mapContentType(item.content_type),
      desc: `来源页面 ${item.source_page || "-"}，内容ID ${item.content_id || "-"}`,
      time: formatDateTime(item.viewed_at),
      duration: resolveReadingDuration(item),
      progress: resolveReadingProgress(item)
    }))
);

const subscriptionItems = computed(() =>
  (rawSubscriptions.value || []).map((item) => {
    const statusRaw = String(item.status || "").toUpperCase();
    const frequencyRaw = String(item.frequency || "").toUpperCase();
    return {
      id: item.id,
      name: mapSubscriptionType(item.type),
      status: mapSubscriptionStatus(statusRaw),
      statusRaw: statusRaw || "ACTIVE",
      desc: `订阅范围：${item.scope || "ALL"}，推送频率：${mapSubscriptionFrequency(frequencyRaw)}`,
      cycle: mapSubscriptionFrequency(frequencyRaw),
      frequencyRaw: frequencyRaw || "DAILY",
      scope: item.scope || "ALL",
      price: "按会员权益",
      saving: !!subscriptionSavingMap.value[item.id]
    };
  })
);

const vipInfo = computed(() => {
  const levelText = mapMemberLevel(displayProfile.value.memberLevel, rawQuota.value?.member_level);
  const expireRaw = rawProfile.value?.vip_expire_at || rawQuota.value?.vip_expire_at || "";
  const expireTs = toTimestamp(expireRaw);
  const computedRemaining = expireTs > 0 ? Math.max(0, Math.ceil((expireTs - Date.now()) / (24 * 3600 * 1000))) : 0;
  const serverRemaining = Number(rawQuota.value?.vip_remaining_days ?? rawProfile.value?.vip_remaining_days ?? 0);
  const remainingDays = Number.isFinite(serverRemaining) && serverRemaining > 0 ? serverRemaining : computedRemaining;
  const expireAt = formatDateTime(expireRaw);
  const nextRenewAt = expireAt;
  const status = isPaidPendingKYC.value
    ? "待实名激活"
    : mapVIPStatus(rawQuota.value?.vip_status || rawProfile.value?.vip_status, displayProfile.value.memberLevel);

  return {
    level: levelText,
    status,
    expireAt,
    nextRenewAt,
    remainingDays
  };
});

const vipMetrics = computed(() => [
  { label: "会员等级", value: vipInfo.value.level },
  { label: "激活状态", value: activationStateLabel.value },
  { label: "会员状态", value: vipInfo.value.status },
  {
    label: "文档阅读配额",
    value: `${rawQuota.value?.doc_read_used ?? 0}/${rawQuota.value?.doc_read_limit ?? 0}`
  },
  {
    label: "资讯订阅余量",
    value: `${rawQuota.value?.news_subscribe_remaining ?? 0}`
  },
  { label: "KYC状态", value: displayProfile.value.kycStatus }
]);

const vipBenefits = computed(() => [
  {
    title: "文档阅读额度",
    desc: `周期 ${rawQuota.value?.period_key || "-"}，已用 ${
      rawQuota.value?.doc_read_used ?? 0
    } / ${rawQuota.value?.doc_read_limit ?? 0}。`
  },
  {
    title: "资讯订阅额度",
    desc: `已用 ${rawQuota.value?.news_subscribe_used ?? 0} / ${
      rawQuota.value?.news_subscribe_limit ?? 0
    }，剩余 ${rawQuota.value?.news_subscribe_remaining ?? 0}。`
  },
  {
    title: "重置周期",
    desc: `周期 ${mapResetCycle(rawQuota.value?.reset_cycle)}，下次重置 ${formatDateTime(
      rawQuota.value?.reset_at
    )}。`
  },
  {
    title: "会员续费状态",
    desc: isPaidPendingKYC.value
      ? `${activationPromptDesc.value} 当前到期时间：${vipInfo.value.expireAt}。`
      : `当前会员状态：${vipInfo.value.status}，到期时间：${vipInfo.value.expireAt}。`
  }
]);

const vipLevelRaw = computed(() => String(displayProfile.value.memberLevel || rawQuota.value?.member_level || "").toUpperCase());
const isVIPActive = computed(() => {
  if (currentActivationState.value) {
    return currentActivationState.value === "ACTIVE";
  }
  const quotaStatus = String(rawQuota.value?.vip_status || rawProfile.value?.vip_status || "").toUpperCase();
  if (quotaStatus === "ACTIVE") {
    return true;
  }
  return vipLevelRaw.value.startsWith("VIP") && quotaStatus !== "EXPIRED";
});

const paymentSummary = computed(() => {
  const paid = paymentRecords.value.filter((item) => item.status === "已支付");
  const pending = paymentRecords.value.filter((item) => item.status === "处理中");
  const refund = paymentRecords.value.filter((item) => item.status === "已退款");

  const paidAmount = paid.reduce((acc, item) => acc + item.amountValue, 0);
  const refundAmount = refund.reduce((acc, item) => acc + item.amountValue, 0);

  return [
    { label: `${activeRange.value}支付总额`, value: formatAmount(paidAmount) },
    { label: "成功支付笔数", value: `${paid.length} 笔` },
    { label: "待处理订单", value: `${pending.length} 笔` },
    { label: "退款金额", value: formatAmount(refundAmount) }
  ];
});

const pendingPaymentCount = computed(
  () => paymentRecords.value.filter((item) => item.status === "处理中").length
);

const readingStats = computed(() => {
  const items = readingLogs.value;
  const reportCount = items.filter((item) => item.type === "研报").length;
  const journalCount = items.filter((item) => item.type === "期刊").length;
  const newsCount = items.filter((item) => item.type === "新闻").length;

  return [
    { label: `${activeRange.value}阅读总量`, value: `${items.length} 篇` },
    { label: "研报阅读", value: `${reportCount} 篇` },
    { label: "期刊阅读", value: `${journalCount} 篇` },
    { label: "新闻阅读", value: `${newsCount} 篇` }
  ];
});

const messageItems = computed(() =>
  (rawMessages.value || [])
    .filter((item) => inRange(item.created_at, activeRange.value))
    .sort((a, b) => toTimestamp(b.created_at) - toTimestamp(a.created_at))
    .map((item) => {
      const readStatusRaw = String(item.read_status || "UNREAD").toUpperCase();
      return {
        id: item.id,
        title: item.title || "未命名通知",
        content: item.content || "-",
        type: mapMessageType(item.type),
        time: formatDateTime(item.created_at),
        readStatus: mapMessageReadStatus(readStatusRaw),
        readStatusRaw,
        loading: !!messageActionLoadingMap.value[item.id]
      };
    })
);

const messageStats = computed(() => {
  const list = messageItems.value;
  const unread = list.filter((item) => item.readStatusRaw !== "READ").length;
  const read = list.length - unread;
  const strategy = list.filter((item) => item.type === "策略提醒").length;
  return [
    { label: `${activeRange.value}通知总量`, value: `${list.length} 条` },
    { label: "未读通知", value: `${unread} 条` },
    { label: "已读通知", value: `${read} 条` },
    { label: "策略提醒", value: `${strategy} 条` }
  ];
});

const unreadMessageCount = computed(
  () => messageItems.value.filter((item) => item.readStatusRaw !== "READ").length
);

const shareLinks = computed(() =>
  (rawShareLinks.value || []).map((item) => ({
    id: item.id,
    code: item.invite_code || "-",
    channel: item.channel || "-",
    status: mapInviteLinkStatus(item.status),
    rawStatus: item.status || "",
    expiredAt: formatDateTime(item.expired_at),
    url: item.url || "",
    shareURL: buildInviteURL(item.url, item.invite_code),
    copying: copyInviteID.value === item.id
  }))
);

const inviteRecords = computed(() =>
  (rawInviteRecords.value || [])
    .filter((item) => inRange(item.register_at, activeRange.value))
    .sort((a, b) => toTimestamp(b.register_at) - toTimestamp(a.register_at))
    .map((item) => ({
      id: item.id,
      inviteeUser: item.invitee_user_id || "-",
      registerAt: formatDateTime(item.register_at),
      firstPayAt: formatDateTime(item.first_pay_at),
      status: mapInviteStatus(item.status),
      riskFlag: mapInviteRiskFlag(item.risk_flag),
      statusRaw: String(item.status || "").toUpperCase()
    }))
);

const inviteSourceInfo = computed(() => ({
  registrationSource: mapRegistrationSource(rawProfile.value?.registration_source),
  inviterUserID: rawProfile.value?.inviter_user_id || "-",
  inviteCode: rawProfile.value?.invite_code || "-",
  invitedAt: formatDateTime(rawProfile.value?.invited_at)
}));

const inviteStats = computed(() => {
  const summary = rawInviteSummary.value || {};
  let invitedCount = Number(summary.registered_count || 0);
  let convertedCount = Number(summary.first_paid_count || 0);
  let conversionRate = Number(summary.conversion_rate || 0);

  if (activeRange.value === "近7天") {
    invitedCount = Number(summary.last_7d_registered_count || 0);
    convertedCount = Number(summary.last_7d_first_paid_count || 0);
    conversionRate = Number(summary.last_7d_conversion_rate || 0);
  } else if (activeRange.value === "近30天") {
    invitedCount = Number(summary.last_30d_registered_count || 0);
    convertedCount = Number(summary.last_30d_first_paid_count || 0);
    conversionRate = Number(summary.last_30d_conversion_rate || 0);
  }

  const activeShareLinks = Number(summary.share_link_count || 0);
  const window7Rate = Number(summary.last_7d_conversion_rate || 0);
  const window30Rate = Number(summary.last_30d_conversion_rate || 0);
  return [
    { label: `${activeRange.value}邀请注册`, value: `${invitedCount} 人` },
    { label: "首单转化", value: `${convertedCount} 人` },
    { label: "生效分享链接", value: `${activeShareLinks} 条` },
    { label: `${activeRange.value}转化率`, value: `${(conversionRate * 100).toFixed(1)}%` },
    { label: "近7天转化率", value: `${(window7Rate * 100).toFixed(1)}%` },
    { label: "近30天转化率", value: `${(window30Rate * 100).toFixed(1)}%` },
    { label: "我的注册来源", value: inviteSourceInfo.value.registrationSource }
  ];
});

const activeSubscriptionCount = computed(
  () => subscriptionItems.value.filter((item) => item.statusRaw === "ACTIVE").length
);

const profileRhythmStatus = computed(() => {
  if (loading.value) {
    return {
      tone: "info",
      eyebrow: "同步中",
      title: "正在刷新个人中心数据",
      desc: "同步完成后，会按你当前状态给出下一步动作。",
      primaryAction: { type: "route", value: "/strategies", label: "先看主推荐" },
      secondaryAction: { type: "module", value: "vip", label: "查看 VIP 情况" }
    };
  }
  if (loadError.value) {
    return {
      tone: "warning",
      eyebrow: "需处理",
      title: "个人中心部分数据加载失败",
      desc: loadError.value,
      primaryAction: { type: "route", value: "/membership", label: "回会员中心核对" },
      secondaryAction: { type: "module", value: "payment", label: "查看支付情况" }
    };
  }
  if (pendingPaymentCount.value > 0) {
    return {
      tone: "warning",
      eyebrow: "待完成",
      title: `还有 ${pendingPaymentCount.value} 笔订单待处理，别让会员权益断档`,
      desc: "先回会员中心完成支付，再回这里核对支付记录和通知提醒。",
      primaryAction: { type: "route", value: "/membership", label: "去完成支付" },
      secondaryAction: { type: "module", value: "payment", label: "查看支付明细" }
    };
  }
  if (isPaidPendingKYC.value) {
    return {
      tone: "warning",
      eyebrow: "待实名激活",
      title: activationPromptTitle.value,
      desc: `${activationPromptDesc.value} 当前实名状态：${displayProfile.value.kycStatus}。`,
      primaryAction: {
        type: "module",
        value: "vip",
        label: currentKYCStatusRaw.value === "PENDING" ? "查看激活进度" : "提交实名信息"
      },
      secondaryAction: { type: "route", value: "/archive", label: "先看公开历史样本" }
    };
  }
  if (unreadMessageCount.value > 0) {
    return {
      tone: "info",
      eyebrow: "有提醒",
      title: `你还有 ${unreadMessageCount.value} 条未读通知，建议在 15:30 集中处理`,
      desc: "先去关注页看盘后变化，再回个人中心把提醒清掉。",
      primaryAction: { type: "route", value: "/watchlist", label: "先去我的关注" },
      secondaryAction: { type: "module", value: "message", label: "处理未读通知" }
    };
  }
  if (!isVIPActive.value) {
    return {
      tone: "info",
      eyebrow: "升级前节奏",
      title: "先用公开入口形成回访习惯，再决定是否升级会员",
      desc: "今天建议先跑一遍 08:30 / 11:30 / 15:30 / 周末 的完整节奏。",
      primaryAction: { type: "route", value: "/strategies", label: "开始今日节奏" },
      secondaryAction: { type: "route", value: "/membership", label: "查看套餐方案" }
    };
  }
  return {
    tone: "success",
    eyebrow: "会员已开通",
    title: "个人中心是你的账户总览和待办入口",
    desc: "可在这里安排节奏、处理提醒、检查订阅和查看复盘进度。",
    primaryAction: { type: "route", value: "/news", label: "去看午盘资讯" },
    secondaryAction: { type: "module", value: "subscription", label: "管理订阅频率" }
  };
});

const profileCadenceEntries = computed(() => {
  if (isPaidPendingKYC.value) {
    return [
      {
        slot: "08:30",
        title: "先把实名激活动作补齐",
        desc: "完整主推荐解释链会在实名通过后开启，先把激活动作放在今天的第一步。",
        highlight: "入口：个人中心 VIP 模块",
        supporting: `实名状态 ${displayProfile.value.kycStatus}`,
        primaryAction: {
          type: "module",
          value: "vip",
          label: currentKYCStatusRaw.value === "PENDING" ? "查看激活进度" : "提交实名信息"
        },
        secondaryAction: { type: "route", value: "/strategies", label: "先看公开主推荐" }
      },
      {
        slot: "11:30",
        title: "午盘继续看公开资讯",
        desc: `当前有效订阅 ${activeSubscriptionCount.value} 项，实名完成前仍可先用公开资讯查看盘中变化。`,
        highlight: "入口：资讯页",
        supporting: "高级资讯权限待实名后激活",
        primaryAction: { type: "route", value: "/news", label: "进入资讯页" },
        secondaryAction: { type: "module", value: "subscription", label: "调整订阅" }
      },
      {
        slot: "15:30",
        title: "先保留收盘后的回访习惯",
        desc:
          unreadMessageCount.value > 0
            ? `先去我的关注，再回来处理 ${unreadMessageCount.value} 条通知，避免节奏断掉。`
            : "先去我的关注保留收盘回访习惯，等实名后再接回完整解释能力。",
        highlight: "入口：我的关注",
        supporting: "高级跟踪能力待实名后激活",
        primaryAction: { type: "route", value: "/watchlist", label: "进入我的关注" },
        secondaryAction: { type: "module", value: "message", label: "处理通知" }
      },
      {
        slot: "周末",
        title: "先用公开样本继续复盘",
        desc: "周末先去历史档案看公开兑现样本，再回个人中心确认实名激活进度。",
        highlight: "入口：历史档案",
        supporting: `支付 ${paymentRecords.value.length} 条 · 激活状态 ${activationStateLabel.value}`,
        primaryAction: { type: "route", value: "/archive", label: "进入历史档案" },
        secondaryAction: { type: "module", value: "vip", label: "回看激活状态" }
      }
    ];
  }

  return [
    {
      slot: "08:30",
      title: "开盘前看主推荐",
      desc: isVIPActive.value
        ? "先去策略页确认今日主推，再回个人中心确认会员状态和剩余有效期。"
        : "先用公开主推荐建立每日首访理由，看完再评估是否升级。",
      highlight: "入口：策略页",
      supporting: `${vipInfo.value.level} · ${vipInfo.value.status}`,
      primaryAction: { type: "route", value: "/strategies", label: "进入策略页" },
      secondaryAction: { type: "module", value: "vip", label: "查看 VIP 情况" }
    },
    {
      slot: "11:30",
      title: "午盘回来看资讯",
      desc: `当前有效订阅 ${activeSubscriptionCount.value} 项，可先用资讯页查看盘中变化，再回这里调订阅频率。`,
      highlight: "入口：资讯页",
      supporting: `订阅 ${activeSubscriptionCount.value} 项`,
      primaryAction: { type: "route", value: "/news", label: "进入资讯页" },
      secondaryAction: { type: "module", value: "subscription", label: "调整订阅" }
    },
    {
      slot: "15:30",
      title: "收盘回关注清单",
      desc:
        unreadMessageCount.value > 0
          ? `收盘后先去我的关注，再回来处理 ${unreadMessageCount.value} 条未读通知，形成闭环。`
          : "收盘后先去我的关注看跟踪结果，再回个人中心补齐消息和阅读记录。",
      highlight: "入口：我的关注",
      supporting: `未读通知 ${unreadMessageCount.value} 条`,
      primaryAction: { type: "route", value: "/watchlist", label: "进入我的关注" },
      secondaryAction: { type: "module", value: "message", label: "处理通知" }
    },
    {
      slot: "周末",
      title: "做周度复盘",
      desc: "周末先去历史档案看推荐兑现，再回个人中心检查支付、邀请和账户投入产出。",
      highlight: "入口：历史档案",
      supporting: `支付 ${paymentRecords.value.length} 条 · 邀请 ${inviteRecords.value.length} 条`,
      primaryAction: { type: "route", value: "/archive", label: "进入历史档案" },
      secondaryAction: {
        type: "module",
        value: pendingPaymentCount.value > 0 ? "payment" : "invite",
        label: pendingPaymentCount.value > 0 ? "回看支付明细" : "查看邀请复盘"
      }
    }
  ];
});

const otherInfos = computed(() => [
  {
    title: "账户基础信息",
    rows: [
      { key: "客户编号", value: rawProfile.value?.id || "-" },
      { key: "手机号", value: rawProfile.value?.phone || "-" },
      { key: "邮箱", value: rawProfile.value?.email || "-" }
    ]
  },
  {
    title: "会员与配额",
    rows: [
      { key: "会员等级", value: vipInfo.value.level },
      { key: "激活状态", value: activationStateLabel.value },
      { key: "实名状态", value: displayProfile.value.kycStatus },
      { key: "文档配额剩余", value: `${rawQuota.value?.doc_read_remaining ?? 0}` },
      { key: "资讯订阅剩余", value: `${rawQuota.value?.news_subscribe_remaining ?? 0}` }
    ]
  },
  {
    title: "记录统计",
    rows: [
      { key: "支付记录", value: `${paymentRecords.value.length} 条` },
      { key: "阅读记录", value: `${readingLogs.value.length} 条` },
      { key: "订阅项", value: `${subscriptionItems.value.length} 条` },
      { key: "通知消息", value: `${messageItems.value.length} 条` },
      { key: "邀请记录", value: `${inviteRecords.value.length} 条` }
    ]
  },
  {
    title: "邀请关系",
    rows: [
      { key: "注册来源", value: inviteSourceInfo.value.registrationSource },
      { key: "邀请人", value: inviteSourceInfo.value.inviterUserID },
      { key: "邀请码", value: inviteSourceInfo.value.inviteCode },
      { key: "分享链接数", value: `${shareLinks.value.length} 条` }
    ]
  }
]);

const todos = computed(() => [
  pendingPaymentCount.value > 0
    ? {
        title: "完成待支付订单",
        note: `当前还有 ${pendingPaymentCount.value} 笔订单处理中，建议优先处理。`,
        level: "high",
        actionLabel: "去会员中心",
        action: { type: "route", value: "/membership" }
      }
    : isPaidPendingKYC.value
      ? {
          title: "完成实名激活",
          note: activationPromptDesc.value,
          level: "high",
          actionLabel: currentKYCStatusRaw.value === "PENDING" ? "查看激活进度" : "提交实名信息",
          action: { type: "module", value: "vip" }
        }
    : {
        title: isVIPActive.value ? "确认会员有效期" : "完成会员升级决策",
        note: isVIPActive.value
          ? `当前 ${vipInfo.value.status}，剩余 ${vipInfo.value.remainingDays} 天。`
          : "先跑一轮完整节奏，再决定是否升级会员。",
        level: "high",
        actionLabel: isVIPActive.value ? "查看 VIP 情况" : "查看套餐方案",
        action: { type: isVIPActive.value ? "module" : "route", value: isVIPActive.value ? "vip" : "/membership" }
      },
  {
    title: "检查订阅节奏",
    note: `当前有效订阅 ${activeSubscriptionCount.value} 项，确认午盘和周末频率是否匹配。`,
    level: "mid",
    actionLabel: "管理订阅",
    action: { type: "module", value: "subscription" }
  },
  {
    title: unreadMessageCount.value > 0 ? "清理未读通知" : "回顾阅读记录",
    note:
      unreadMessageCount.value > 0
        ? `还有 ${unreadMessageCount.value} 条未读，适合在收盘后一并处理。`
        : `当前范围内累计阅读 ${readingLogs.value.length} 条，可用于周度复盘。`,
    level: "low",
    actionLabel: unreadMessageCount.value > 0 ? "处理通知" : "查看阅读情况",
    action: { type: "module", value: unreadMessageCount.value > 0 ? "message" : "reading" }
  }
]);

const quickActions = computed(() => [
  {
    title: "会员权益中心",
    desc: isPaidPendingKYC.value
      ? `当前 ${vipInfo.value.level} 已开通，但还在等待实名激活。`
      : isVIPActive.value
        ? `当前 ${vipInfo.value.level} 生效中，适合随时检查续费和权益。`
      : "当前还没进入会员节奏，先看方案与权益差异。",
    actionLabel: isPaidPendingKYC.value ? "完成实名激活" : isVIPActive.value ? "查看会员页" : "去升级会员",
    action: isPaidPendingKYC.value ? { type: "module", value: "vip" } : { type: "route", value: "/membership" }
  },
  {
    title: "午盘订阅设置",
    desc: `有效订阅 ${activeSubscriptionCount.value} 项，把盘中提醒压缩成可执行节奏。`,
    actionLabel: "调整订阅",
    action: { type: "module", value: "subscription" }
  },
  {
    title: "通知与消息",
    desc: unreadMessageCount.value > 0 ? `当前未读 ${unreadMessageCount.value} 条。` : "当前消息已清空，继续保持。",
    actionLabel: "查看消息",
    action: { type: "module", value: "message" }
  },
  {
    title: "邀请与分享",
    desc: `分享链接 ${shareLinks.value.length} 条，邀请记录 ${inviteRecords.value.length} 条。`,
    actionLabel: "查看邀请关系",
    action: { type: "module", value: "invite" }
  },
  {
    title: "我的主题回看",
    desc: "集中回看自己发过的主题，适合收盘后补充结论、依据和风险变化。",
    actionLabel: "查看我的主题",
    action: { type: "route", value: { path: "/community", query: { mine: "topics" } } }
  },
  {
    title: "我的评论回看",
    desc: "集中回看自己参与过的评论，确认哪些讨论还需要继续跟进。",
    actionLabel: "查看我的评论",
    action: { type: "route", value: { path: "/community", query: { mine: "comments" } } }
  }
]);
const profileOutstandingCount = computed(() => {
  let count = 0;
  if (pendingPaymentCount.value > 0) count += 1;
  if (isPaidPendingKYC.value) count += 1;
  if (!isVIPActive.value) count += 1;
  if (unreadMessageCount.value > 0) count += 1;
  return count;
});
const profileOverviewRows = computed(() => [
  {
    label: "会员身份",
    value: vipInfo.value.level,
    note: `${vipInfo.value.status} · ${activationStateLabel.value}`
  },
  {
    label: "今日重点",
    value: profileRhythmStatus.value.eyebrow,
    note: profileRhythmStatus.value.title
  },
  {
    label: "待处理事项",
    value: `${profileOutstandingCount.value} 项`,
    note:
      profileOutstandingCount.value > 0
        ? "先把支付、实名、未读消息等阻塞清掉，再进入查询中心逐项核对。"
        : "当前主要阻塞已清空，可以直接按今日节奏执行。"
  },
  {
    label: "当前查询",
    value: currentModule.value.label,
    note: `时间范围 ${activeRange.value} · 最近更新 ${lastUpdatedAt.value || "-"}`
  }
]);
const profileGuideRows = computed(() => [
  {
    title: "先看今日重点",
    desc: "先确认今天最重要的账户事项，再进入各模块查看明细。"
  },
  {
    title: "状态清晰",
    desc: "待支付、待实名激活、未读通知、未开通会员等状态都会明确显示。"
  },
  {
    title: "常用入口集中",
    desc: "可从这里继续前往会员页、资讯页、关注页、历史档案和我的讨论。"
  }
]);
const profileAccountSummaryRows = computed(() => [
  { label: "账户身份", value: `${displayProfile.value.name} · ${vipInfo.value.level}` },
  { label: "会员状态", value: `${vipInfo.value.status} · 到期 ${vipInfo.value.expireAt}` },
  { label: "实名状态", value: `${displayProfile.value.kycStatus} · ${activationStateLabel.value}` },
  { label: "消息与订阅", value: `未读 ${unreadMessageCount.value} 条 · 生效订阅 ${activeSubscriptionCount.value} 项` }
]);
const profilePendingRows = computed(() => {
  const rows = [];
  if (pendingPaymentCount.value > 0) {
    rows.push({
      title: "先处理待支付订单",
      desc: `当前还有 ${pendingPaymentCount.value} 笔订单处理中，建议先回会员中心完成支付。`
    });
  }
  if (isPaidPendingKYC.value) {
    rows.push({
      title: "完成实名激活",
      desc: activationPromptDesc.value
    });
  }
  if (unreadMessageCount.value > 0) {
    rows.push({
      title: "统一处理未读通知",
      desc: `还有 ${unreadMessageCount.value} 条未读消息，适合在收盘后一次清掉。`
    });
  }
  if (!isVIPActive.value) {
    rows.push({
      title: "确认是否进入会员节奏",
      desc: "先跑完今日公开节奏，再决定是否升级会员能力。"
    });
  }
  if (rows.length < 3) {
    rows.push({
      title: "管理订阅频率",
      desc: `当前有效订阅 ${activeSubscriptionCount.value} 项，可继续调整午盘和周末提醒节奏。`
    });
  }
  if (rows.length < 3) {
    rows.push({
      title: "回看查询中心",
      desc: `当前查询聚焦 ${currentModule.value.label}，可继续核对 ${activeRange.value} 范围内的数据。`
    });
  }
  return rows.slice(0, 3);
});
const profileStatusRows = computed(() => [
  {
    label: "会员有效期",
    value: vipInfo.value.expireAt,
    note: `剩余 ${vipInfo.value.remainingDays} 天`
  },
  {
    label: "实名进度",
    value: displayProfile.value.kycStatus,
    note: isPaidPendingKYC.value ? activationPromptDesc.value : "当前实名状态不阻塞高级能力。"
  },
  {
    label: "未读通知",
    value: `${unreadMessageCount.value} 条`,
    note: unreadMessageCount.value > 0 ? "建议在 15:30 和收盘后统一处理。" : "当前通知已清空，可继续保持。"
  },
  {
    label: "订阅与邀请",
    value: `${activeSubscriptionCount.value} 项订阅 / ${shareLinks.value.length} 条分享`,
    note: `邀请记录 ${inviteRecords.value.length} 条，可在查询中心继续核对。`
  }
]);

async function loadUserCenterData() {
  loading.value = true;
  loadError.value = "";
  subscriptionActionMessage.value = "";
  subscriptionActionError.value = "";
  inviteActionMessage.value = "";
  inviteActionError.value = "";
  const tasks = [
    {
      key: "profile",
      label: "用户资料",
      request: getUserProfile(),
      apply: (data) => {
        rawProfile.value = data || rawProfile.value;
      }
    },
    {
      key: "quota",
      label: "会员配额",
      request: getMembershipQuota(),
      apply: (data) => {
        rawQuota.value = data || rawQuota.value;
      }
    },
    {
      key: "orders",
      label: "会员订单",
      request: listMembershipOrders({ page: 1, page_size: 50 }),
      apply: (data) => {
        rawMembershipOrders.value = toArray(data?.items, rawMembershipOrders.value);
      }
    },
    {
      key: "recharges",
      label: "充值记录",
      request: listRechargeRecords({ page: 1, page_size: 50 }),
      apply: (data) => {
        rawRechargeRecords.value = toArray(data?.items, rawRechargeRecords.value);
      }
    },
    {
      key: "browses",
      label: "阅读记录",
      request: listBrowseHistory({ page: 1, page_size: 100 }),
      apply: (data) => {
        rawBrowseHistory.value = toArray(data?.items, rawBrowseHistory.value);
      }
    },
    {
      key: "subscriptions",
      label: "订阅列表",
      request: listSubscriptions({ page: 1, page_size: 50 }),
      apply: (data) => {
        applySubscriptionItems(toArray(data?.items, rawSubscriptions.value));
      }
    },
    {
      key: "messages",
      label: "通知消息",
      request: listMessages({ page: 1, page_size: 100 }),
      apply: (data) => {
        rawMessages.value = toArray(data?.items, rawMessages.value);
        messageActionLoadingMap.value = {};
      }
    },
    {
      key: "shareLinks",
      label: "分享链接",
      request: listShareLinks(),
      apply: (data) => {
        rawShareLinks.value = toArray(data?.items, rawShareLinks.value);
      }
    },
    {
      key: "invites",
      label: "邀请记录",
      request: listInviteRecords({ page: 1, page_size: 100 }),
      apply: (data) => {
        rawInviteRecords.value = toArray(data?.items, rawInviteRecords.value);
      }
    },
    {
      key: "inviteSummary",
      label: "邀请汇总",
      request: getInviteSummary(),
      apply: (data) => {
        rawInviteSummary.value = data || rawInviteSummary.value;
      }
    }
  ];

  const results = await Promise.allSettled(tasks.map((item) => item.request));
  const errors = [];
  results.forEach((result, index) => {
    const task = tasks[index];
    if (result.status === "fulfilled") {
      task.apply(result.value);
      return;
    }
    errors.push(`${task.label}加载失败：${result.reason?.message || "unknown error"}`);
  });

  loadError.value = errors.join("；");
  lastUpdatedAt.value = formatDateTime(new Date().toISOString());
  loading.value = false;
}

async function refreshInviteData() {
  const tasks = [
    {
      label: "用户资料",
      request: getUserProfile(),
      apply: (data) => {
        rawProfile.value = data || rawProfile.value;
      }
    },
    {
      label: "分享链接",
      request: listShareLinks(),
      apply: (data) => {
        rawShareLinks.value = toArray(data?.items, rawShareLinks.value);
      }
    },
    {
      label: "邀请记录",
      request: listInviteRecords({ page: 1, page_size: 100 }),
      apply: (data) => {
        rawInviteRecords.value = toArray(data?.items, rawInviteRecords.value);
      }
    },
    {
      label: "邀请汇总",
      request: getInviteSummary(),
      apply: (data) => {
        rawInviteSummary.value = data || rawInviteSummary.value;
      }
    }
  ];

  const results = await Promise.allSettled(tasks.map((item) => item.request));
  const errors = [];
  results.forEach((result, index) => {
    const task = tasks[index];
    if (result.status === "fulfilled") {
      task.apply(result.value);
      return;
    }
    errors.push(`${task.label}刷新失败：${result.reason?.message || "unknown error"}`);
  });
  if (errors.length > 0) {
    throw new Error(errors.join("；"));
  }
}

async function handleCreateShareLink() {
  if (creatingShareLink.value) {
    return;
  }
  creatingShareLink.value = true;
  inviteActionMessage.value = "";
  inviteActionError.value = "";
  try {
    const payload = {
      channel: newShareLinkChannel.value
    };
    const result = await createShareLink(payload);
    if (result?.id) {
      rawShareLinks.value = [result, ...(rawShareLinks.value || [])];
    } else {
      await refreshInviteData();
    }
    inviteActionMessage.value = `已创建分享链接（${mapShareChannel(newShareLinkChannel.value)}）`;
  } catch (error) {
    inviteActionError.value = error?.message || "创建分享链接失败";
  } finally {
    creatingShareLink.value = false;
  }
}

async function handleCopyInviteLink(item) {
  if (!item?.id || !item?.shareURL) {
    return;
  }
  copyInviteID.value = item.id;
  inviteActionMessage.value = "";
  inviteActionError.value = "";
  try {
    await copyText(item.shareURL);
    inviteActionMessage.value = `已复制邀请码 ${item.code} 的分享链接`;
  } catch (error) {
    inviteActionError.value = error?.message || "复制失败，请手动复制";
  } finally {
    copyInviteID.value = "";
  }
}

async function handleReadMessage(item) {
  if (!item?.id || item.readStatusRaw === "READ") {
    return;
  }
  setMessageSaving(item.id, true);
  loadError.value = "";
  try {
    await readMessage(item.id);
    rawMessages.value = (rawMessages.value || []).map((row) => {
      if (row.id === item.id) {
        return { ...row, read_status: "READ" };
      }
      return row;
    });
  } catch (error) {
    loadError.value = error?.message || "标记已读失败";
  } finally {
    setMessageSaving(item.id, false);
  }
}

async function handleSubmitKYC() {
  if (!canSubmitKYC.value || kycSubmitting.value) {
    return;
  }
  const payload = {
    real_name: String(kycForm.value.real_name || "").trim(),
    id_number: String(kycForm.value.id_number || "").trim()
  };
  if (!payload.real_name || !payload.id_number) {
    kycActionError.value = "请先填写真实姓名和身份证号";
    kycActionMessage.value = "";
    return;
  }
  kycSubmitting.value = true;
  kycActionMessage.value = "";
  kycActionError.value = "";
  try {
    const result = await submitKYC(payload);
    const nextStatus = String(result?.kyc_status || "PENDING").toUpperCase();
    rawProfile.value = {
      ...rawProfile.value,
      kyc_status: nextStatus
    };
    rawQuota.value = {
      ...rawQuota.value,
      kyc_status: nextStatus
    };
    kycActionMessage.value =
      nextStatus === "PENDING"
        ? "实名材料已提交，审核通过后会自动激活高级权益。"
        : `实名状态已更新为 ${mapKYCStatus(nextStatus)}`;
    await loadUserCenterData();
  } catch (error) {
    kycActionError.value = error?.message || "提交实名失败";
  } finally {
    kycSubmitting.value = false;
  }
}

async function refreshSubscriptions() {
  const subscriptionData = await listSubscriptions({ page: 1, page_size: 50 });
  applySubscriptionItems(toArray(subscriptionData?.items, rawSubscriptions.value));
}

async function handleCreateSubscription() {
  if (creatingSubscription.value) {
    return;
  }
  creatingSubscription.value = true;
  subscriptionActionMessage.value = "";
  subscriptionActionError.value = "";
  try {
    const payload = {
      type: newSubscriptionForm.value.type,
      frequency: newSubscriptionForm.value.frequency,
      scope: newSubscriptionForm.value.scope || "ALL"
    };
    const result = await createSubscription(payload);
    await refreshSubscriptions();
    const createdID = result?.id || "-";
    subscriptionActionMessage.value = `订阅创建成功：${createdID}`;
  } catch (error) {
    subscriptionActionError.value = error?.message || "创建订阅失败";
  } finally {
    creatingSubscription.value = false;
  }
}

async function handleToggleSubscriptionStatus(item) {
  const targetStatus = item.statusRaw === "ACTIVE" ? "PAUSED" : "ACTIVE";
  await updateSubscriptionWithFeedback(
    item,
    { frequency: item.frequencyRaw, status: targetStatus },
    `订阅状态已更新为 ${mapSubscriptionStatus(targetStatus)}`
  );
}

async function handleRotateSubscriptionFrequency(item) {
  const nextFrequency = nextSubscriptionFrequency(item.frequencyRaw);
  await updateSubscriptionWithFeedback(
    item,
    { frequency: nextFrequency, status: item.statusRaw },
    `订阅频率已更新为 ${mapSubscriptionFrequency(nextFrequency)}`
  );
}

async function updateSubscriptionWithFeedback(item, payload, successMessage) {
  if (!item?.id) {
    return;
  }
  setSubscriptionSaving(item.id, true);
  subscriptionActionMessage.value = "";
  subscriptionActionError.value = "";
  try {
    await updateSubscription(item.id, payload);
    await refreshSubscriptions();
    subscriptionActionMessage.value = successMessage;
  } catch (error) {
    subscriptionActionError.value = error?.message || "更新订阅失败";
  } finally {
    setSubscriptionSaving(item.id, false);
  }
}

function setSubscriptionSaving(id, loadingState) {
  const next = { ...subscriptionSavingMap.value };
  if (loadingState) {
    next[id] = true;
  } else {
    delete next[id];
  }
  subscriptionSavingMap.value = next;
}

function setMessageSaving(id, loadingState) {
  const next = { ...messageActionLoadingMap.value };
  if (loadingState) {
    next[id] = true;
  } else {
    delete next[id];
  }
  messageActionLoadingMap.value = next;
}

function applySubscriptionItems(items) {
  rawSubscriptions.value = toArray(items, rawSubscriptions.value);
  subscriptionSavingMap.value = {};
}

function resolveReadingDuration(item) {
  const secondCandidates = [
    item?.duration_seconds,
    item?.duration_sec,
    item?.read_duration_seconds,
    item?.read_seconds
  ];
  const msCandidates = [item?.duration_ms, item?.read_duration_ms];
  for (const value of secondCandidates) {
    const seconds = Number(value);
    if (Number.isFinite(seconds) && seconds > 0) {
      return formatDuration(seconds);
    }
  }
  for (const value of msCandidates) {
    const milliseconds = Number(value);
    if (Number.isFinite(milliseconds) && milliseconds > 0) {
      return formatDuration(milliseconds / 1000);
    }
  }
  return "-";
}

function resolveReadingProgress(item) {
  const candidates = [
    item?.progress,
    item?.progress_rate,
    item?.completion_rate,
    item?.read_percent,
    item?.finish_percent
  ];
  for (const value of candidates) {
    const num = Number(value);
    if (!Number.isFinite(num) || num < 0) {
      continue;
    }
    const normalized = num > 1 ? num / 100 : num;
    const clipped = Math.max(0, Math.min(1, normalized));
    return `${Math.round(clipped * 100)}%`;
  }
  return "100%";
}

function formatDuration(secondsValue) {
  const seconds = Math.max(0, Math.round(Number(secondsValue)));
  if (!Number.isFinite(seconds) || seconds <= 0) {
    return "-";
  }
  const hour = Math.floor(seconds / 3600);
  const minute = Math.floor((seconds % 3600) / 60);
  const second = seconds % 60;
  if (hour > 0) {
    return `${hour}小时${minute}分`;
  }
  if (minute > 0) {
    return `${minute}分${second}秒`;
  }
  return `${second}秒`;
}

function openSecurityPanel() {
  focusModule("other");
}

function handleAction(action) {
  if (!action?.type) {
    return;
  }
  if (action.type === "route") {
    goToRoute(action.value);
    return;
  }
  if (action.type === "module") {
    focusModule(action.value);
  }
}

function goToRoute(path) {
  if (!path) {
    return;
  }
  router.push(path);
}

function focusModule(moduleKey) {
  if (moduleKey) {
    activeModule.value = moduleKey;
  }
  scrollToQueryCard();
}

function scrollToQueryCard() {
  if (typeof window === "undefined") {
    return;
  }
  const target = document.querySelector(".query-card");
  if (target) {
    target.scrollIntoView({ behavior: "smooth", block: "start" });
  }
}

onMounted(() => {
  loadUserCenterData();
});
</script>

<style scoped>
.profile-page {
  display: grid;
  gap: 12px;
}

.profile-page > * {
  min-width: 0;
}

.account-card {
  border-radius: 20px;
  padding: 15px;
  display: grid;
  grid-template-columns: 1fr auto;
  gap: 10px;
  align-items: start;
  background:
    radial-gradient(circle at 100% 0%, var(--color-focus-glow) 0%, transparent 36%),
    radial-gradient(circle at 0% 100%, var(--color-line-gold-soft) 0%, transparent 34%),
    rgba(255, 255, 255, 0.93);
}

.profile-hero-copy {
  min-width: 0;
  display: grid;
  gap: 10px;
}

.profile-hero-stats {
  grid-column: 1 / -1;
}

.section-kicker {
  margin: 0;
  font-size: 12px;
  color: var(--color-pine-600);
}

.identity {
  display: flex;
  align-items: center;
  gap: 10px;
  min-width: 0;
}

.avatar {
  width: 58px;
  height: 58px;
  border-radius: 16px;
  background: var(--gradient-primary);
  color: #fff;
  font-size: 24px;
  font-family: var(--font-serif);
  display: inline-flex;
  align-items: center;
  justify-content: center;
}

h1 {
  margin: 0;
  font-size: 24px;
  line-height: 1.25;
  overflow-wrap: anywhere;
}

.identity p {
  margin: 4px 0 0;
  color: var(--color-text-sub);
  font-size: 13px;
}

.actions {
  display: inline-flex;
  gap: 8px;
}

.actions button {
  width: auto;
}

.actions .primary {
  color: #fff;
}

.actions .ghost {
  color: var(--color-pine-700);
}

.profile-focus-actions button {
  width: auto;
}

.profile-focus-actions .primary {
  color: #fff;
}

.profile-focus-actions .ghost {
  color: var(--color-pine-700);
}

.profile-overview-grid article,
.profile-guide-grid article,
.profile-side-list article,
.profile-status-grid article {
  min-width: 0;
}

.profile-overview-grid p,
.profile-overview-grid strong,
.profile-overview-grid span,
.profile-guide-grid strong,
.profile-guide-grid p,
.profile-side-list strong,
.profile-side-list p,
.profile-status-grid p,
.profile-status-grid strong,
.profile-status-grid span {
  margin: 0;
}

.profile-overview-grid p,
.profile-status-grid p {
  font-size: 12px;
  color: var(--color-text-sub);
}

.profile-overview-grid strong,
.profile-guide-grid strong,
.profile-side-list strong,
.profile-status-grid strong {
  font-size: 16px;
  line-height: 1.45;
  color: var(--color-pine-700);
}

.profile-overview-grid span,
.profile-guide-grid p,
.profile-side-list p,
.profile-status-grid span {
  margin-top: 4px;
  font-size: 13px;
  line-height: 1.65;
  color: var(--color-text-sub);
}

.rhythm-card {
  padding: 14px;
}

.rhythm-head {
  display: grid;
  grid-template-columns: 1fr auto;
  gap: 12px;
  align-items: end;
}

.rhythm-pill {
  min-width: 200px;
}

.rhythm-pill p,
.rhythm-pill small {
  margin: 0;
}

.rhythm-pill strong {
  margin: 4px 0;
}

.rhythm-grid {
  margin-top: 12px;
  display: grid;
  gap: 10px;
  grid-template-columns: repeat(4, minmax(0, 1fr));
}

.rhythm-item {
  display: grid;
  gap: 8px;
}

.rhythm-slot {
  margin: 0;
  font-size: 12px;
  font-weight: 700;
  letter-spacing: 0.12em;
  color: var(--color-pine-700);
}

.rhythm-item h3 {
  margin: 0;
  font-size: 18px;
  line-height: 1.3;
}

.rhythm-desc {
  margin: 0;
  font-size: 13px;
  line-height: 1.6;
  color: var(--color-text-sub);
}

.rhythm-tags {
  display: grid;
  gap: 6px;
}

.rhythm-actions {
  margin-top: auto;
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.rhythm-actions button {
  width: 100%;
}

.rhythm-actions button:not(.ghost) {
  color: #fff;
}

.rhythm-actions .ghost {
  color: var(--color-pine-700);
}

.query-card {
  padding: 14px;
}

.query-head {
  display: grid;
  grid-template-columns: 1fr auto;
  align-items: end;
  gap: 10px;
  min-width: 0;
}

.range-switch {
  display: inline-flex;
  gap: 6px;
}

.range-switch button {
  flex-shrink: 0;
}

.range-switch button.active {
  border-color: var(--color-border-focus-medium);
  box-shadow: inset 0 0 0 1px var(--color-focus-fill);
}

.query-nav {
  margin-top: 12px;
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(120px, 1fr));
  gap: 8px;
}

.query-nav button {
  text-align: left;
}

.query-nav button.active {
  border-color: var(--color-border-focus-medium);
  box-shadow: inset 0 0 0 1px var(--color-focus-fill);
}

.query-tip {
  margin-top: 10px;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 8px;
}

.query-tip p {
  min-width: 0;
  overflow-wrap: anywhere;
}

.state-box {
  margin-top: 10px;
}

.query-body {
  margin-top: 10px;
  display: grid;
  gap: 10px;
}

.vip-panel {
  display: grid;
  gap: 10px;
  grid-template-columns: 0.95fr 1.05fr;
}

.vip-main {
  border-radius: 13px;
  border: 1px solid var(--color-border-soft);
  background: linear-gradient(160deg, rgba(19, 54, 103, 0.95), rgba(30, 83, 161, 0.95));
  color: #f5faf7;
  padding: 12px;
}

.vip-level {
  margin: 0;
  font-size: 12px;
  color: rgba(248, 251, 255, 0.76);
}

.vip-main h3 {
  margin: 6px 0;
  font-size: 24px;
  font-family: var(--font-serif);
  line-height: 1.3;
}

.vip-main p {
  margin: 0;
  font-size: 13px;
  color: rgba(248, 251, 255, 0.82);
}

.summary-grid {
  display: grid;
  gap: 8px;
  grid-template-columns: repeat(4, minmax(0, 1fr));
}

.summary-grid article {
  min-width: 0;
}

.summary-grid p {
  margin: 0;
  color: var(--color-text-sub);
  font-size: 12px;
}

.summary-grid strong {
  margin-top: 5px;
  display: block;
  color: var(--color-pine-700);
  font-size: 18px;
}

.benefits-grid {
  display: grid;
  gap: 8px;
  grid-template-columns: repeat(2, minmax(0, 1fr));
}

.benefits-grid article {
  min-width: 0;
}

.benefits-grid h4 {
  margin: 0;
  font-size: 14px;
}

.benefits-grid p {
  margin: 5px 0 0;
  font-size: 13px;
  color: var(--color-text-sub);
  line-height: 1.55;
}

.activation-panel {
  display: grid;
  gap: 10px;
  grid-template-columns: 1fr 1fr;
}

.activation-copy,
.activation-form-wrap {
  min-width: 0;
}

.activation-copy h4 {
  margin: 6px 0 0;
  font-size: 16px;
  color: var(--color-pine-700);
}

.activation-copy p:last-of-type {
  margin: 8px 0 0;
  font-size: 13px;
  line-height: 1.6;
  color: var(--color-text-sub);
}

.activation-tags {
  margin-top: 10px;
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.kyc-form {
  display: grid;
  gap: 10px;
}

.kyc-form label {
  display: grid;
  gap: 6px;
  font-size: 13px;
  color: var(--color-text-sub);
}

.kyc-form input {
  width: 100%;
  border-radius: 10px;
  border: 1px solid rgba(176, 188, 208, 0.92);
  padding: 10px 12px;
  font: inherit;
  background: var(--color-surface-card-elevated);
  color: var(--color-text-main);
}

.kyc-form button {
  border: 0;
  border-radius: 10px;
  padding: 10px 12px;
  font-weight: 600;
  color: #fff;
  cursor: pointer;
  background: var(--gradient-primary);
}

.kyc-form button:disabled {
  cursor: not-allowed;
  opacity: 0.72;
}

.payment-table {
  min-width: 760px;
}

.status {
  font-weight: 700;
}

.status.success {
  color: var(--color-success);
  background: rgba(201, 229, 211, 0.72);
  border-color: rgba(46, 125, 50, 0.16);
}

.status.pending {
  color: var(--color-warning);
  background: rgba(243, 228, 194, 0.84);
  border-color: rgba(184, 130, 48, 0.16);
}

.status.refund,
.status.inactive,
.status.fail {
  color: var(--color-danger);
  background: rgba(237, 198, 190, 0.68);
  border-color: rgba(178, 58, 42, 0.14);
}

.payment-mobile {
  display: none;
}

.payment-mobile article,
.log-list article,
.subscription-item,
.message-item,
.other-grid article,
.todo-card li,
.quick-item {
  min-width: 0;
}

.top-line {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
  min-width: 0;
}

.top-line p {
  margin: 0;
  font-size: 14px;
  font-weight: 600;
  min-width: 0;
  overflow-wrap: anywhere;
}

.top-line span {
  font-size: 12px;
  color: var(--color-text-sub);
}

.order {
  margin: 6px 0 0;
  font-size: 12px;
  color: var(--color-text-sub);
}

.log-list {
  display: grid;
  gap: 8px;
}

.message-list {
  display: grid;
  gap: 8px;
}

.log-list .desc {
  margin: 5px 0 0;
  font-size: 13px;
  color: var(--color-text-sub);
  line-height: 1.58;
}

.message-actions {
  margin-top: 8px;
  display: flex;
  justify-content: flex-end;
}

.message-actions button {
  flex-shrink: 0;
}

.message-actions button:disabled {
  opacity: 0.72;
}

.subscription-grid {
  display: grid;
  gap: 8px;
  grid-template-columns: repeat(3, minmax(0, 1fr));
}

.subscription-create {
  margin-bottom: 8px;
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 8px;
  align-items: end;
}

.subscription-create label {
  display: grid;
  gap: 4px;
  font-size: 12px;
  color: var(--color-text-sub);
}

.subscription-create select,
.subscription-create input {
  border-radius: 8px;
  border: 1px solid var(--color-border-soft-heavy);
  background: #fff;
  padding: 7px 8px;
  color: var(--color-text-main);
}

.subscription-create .scope-input {
  grid-column: span 2;
}

.subscription-create button {
  justify-self: stretch;
}

.subscription-create button:disabled {
  opacity: 0.72;
}

.invite-create {
  margin: 8px 0;
  display: flex;
  align-items: flex-end;
  gap: 8px;
  min-width: 0;
}

.invite-create label {
  display: grid;
  gap: 4px;
  font-size: 12px;
  color: var(--color-text-sub);
}

.invite-create select {
  border-radius: 8px;
  border: 1px solid var(--color-border-soft-heavy);
  background: #fff;
  padding: 7px 8px;
  color: var(--color-text-main);
}

.invite-create button {
  flex-shrink: 0;
}

.invite-create button:disabled {
  opacity: 0.72;
}

.invite-link-row button {
  flex-shrink: 0;
}

.invite-link-row button:disabled {
  opacity: 0.72;
}

.subscription-item .desc {
  margin: 5px 0 0;
  font-size: 13px;
  color: var(--color-text-sub);
  line-height: 1.58;
}

.subscription-actions {
  margin-top: 10px;
  display: flex;
  gap: 8px;
  min-width: 0;
}

.subscription-item button {
  flex: 1;
}

.subscription-item button.secondary {
  color: var(--color-pine-700);
}

.subscription-item button:disabled {
  opacity: 0.72;
}

.other-grid {
  display: grid;
  gap: 8px;
  grid-template-columns: repeat(3, minmax(0, 1fr));
}

.other-grid h4 {
  margin: 0;
  font-size: 14px;
}

.kv-list {
  margin-top: 7px;
  display: grid;
  gap: 6px;
}

.kv-list p {
  margin: 0;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
  font-size: 12px;
  color: var(--color-text-sub);
  min-width: 0;
}

.kv-list strong {
  color: var(--color-text-main);
  font-size: 13px;
  text-align: right;
  overflow-wrap: anywhere;
  word-break: break-word;
}

.bottom-grid {
  display: grid;
  gap: 12px;
  grid-template-columns: 1fr 1fr;
}

.todo-card,
.quick-card {
  padding: 14px;
}

.todo-card ul {
  margin: 10px 0 0;
  padding: 0;
  list-style: none;
  display: grid;
  gap: 8px;
}

.todo-card li {
  display: grid;
  grid-template-columns: auto 1fr auto;
  gap: 8px;
  align-items: center;
}

.dot {
  margin-top: 5px;
  width: 8px;
  height: 8px;
  border-radius: 50%;
}

.dot.high {
  background: var(--color-priority-high);
}

.dot.mid {
  background: var(--color-priority-mid);
}

.dot.low {
  background: var(--color-priority-low);
}

.title {
  margin: 0;
  font-size: 13px;
  font-weight: 600;
}

.note {
  margin: 3px 0 0;
  font-size: 12px;
  color: var(--color-text-sub);
}

.quick-grid {
  margin-top: 10px;
  display: grid;
  gap: 8px;
  grid-template-columns: repeat(2, minmax(0, 1fr));
}

.quick-item h3 {
  margin: 0;
  font-size: 16px;
}

.quick-item p {
  margin: 6px 0 0;
  font-size: 13px;
  color: var(--color-text-sub);
  line-height: 1.56;
}

.todo-actions,
.quick-item button {
  justify-self: end;
}

.todo-actions button,
.quick-item button {
  flex-shrink: 0;
}

.quick-item button {
  margin-top: 10px;
}

@media (max-width: 1080px) {
  .profile-focus-head,
  .profile-guide-grid {
    grid-template-columns: 1fr;
  }

  .query-nav {
    grid-template-columns: repeat(3, minmax(0, 1fr));
  }

  .rhythm-head {
    grid-template-columns: 1fr;
  }

  .subscription-create {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

  .rhythm-grid,
  .vip-panel,
  .activation-panel,
  .subscription-grid,
  .other-grid {
    grid-template-columns: 1fr;
  }

  .summary-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}

@media (max-width: 980px) {
  .account-card,
  .profile-workbench-layout,
  .query-head,
  .bottom-grid {
    grid-template-columns: 1fr;
  }

  .profile-side-rail {
    position: static;
  }
}

@media (max-width: 760px) {
  .account-card,
  .profile-focus-card,
  .profile-side-card,
  .rhythm-card,
  .query-card,
  .todo-card,
  .quick-card {
    border-radius: 14px;
    padding: 12px;
  }

  .identity {
    align-items: flex-start;
    gap: 8px;
  }

  .avatar {
    width: 50px;
    height: 50px;
    border-radius: 14px;
    font-size: 21px;
  }

  h1 {
    font-size: 20px;
  }

  .identity p {
    line-height: 1.5;
    font-size: 12px;
  }

  .actions {
    width: 100%;
    display: grid;
    grid-template-columns: repeat(2, minmax(0, 1fr));
    gap: 6px;
  }

  .actions button {
    width: 100%;
    padding: 8px 10px;
  }

  .profile-focus-actions {
    justify-content: stretch;
  }

  .profile-focus-actions button {
    width: 100%;
  }

  .range-switch,
  .query-nav {
    display: flex;
    overflow-x: auto;
    padding-bottom: 2px;
    scrollbar-width: none;
  }

  .range-switch::-webkit-scrollbar,
  .query-nav::-webkit-scrollbar {
    display: none;
  }

  .range-switch button,
  .query-nav button {
    flex: 0 0 auto;
    min-width: 86px;
    padding: 8px 10px;
  }

  .query-tip {
    flex-direction: column;
    align-items: flex-start;
    gap: 4px;
  }

  .vip-main h3 {
    font-size: 20px;
  }

  .summary-grid strong {
    font-size: 17px;
  }

  .summary-grid,
  .benefits-grid,
  .activation-panel,
  .profile-overview-grid,
  .profile-status-grid,
  .quick-grid {
    grid-template-columns: 1fr;
  }

  .subscription-create {
    grid-template-columns: 1fr;
  }

  .subscription-create .scope-input {
    grid-column: auto;
  }

  .subscription-actions {
    flex-direction: column;
  }

  .invite-create {
    flex-direction: column;
    align-items: stretch;
  }

  .invite-create button {
    width: 100%;
  }

  .top-line {
    align-items: flex-start;
  }

  .top-line span {
    flex-shrink: 0;
  }

  .kv-list p {
    flex-direction: column;
    align-items: flex-start;
    gap: 4px;
  }

  .kv-list strong {
    text-align: left;
  }

  .todo-card li {
    grid-template-columns: auto 1fr;
  }

  .todo-actions {
    grid-column: 1 / -1;
    justify-self: stretch;
  }

  .todo-actions button,
  .quick-item button {
    width: 100%;
  }

  .payment-table-wrap {
    display: none;
  }

  .payment-mobile {
    display: grid;
    gap: 8px;
  }
}
</style>
