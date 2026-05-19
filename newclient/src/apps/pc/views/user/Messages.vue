<template>
  <div class="messages-page">
    <section class="section fade-in-up">
      <div class="section-header">
        <h2 class="section-title">消息通知</h2>
        <span class="section-subtitle">{{ loading ? '加载中...' : messages.length + ' 条消息' }}</span>
      </div>
      <div class="message-list">
        <div v-for="msg in messages" :key="msg.id" class="message-item glass" :class="{ unread: !msg.is_read }" @click="markRead(msg)">
          <div class="msg-dot" :class="{ active: !msg.is_read }"></div>
          <div class="msg-body">
            <strong>{{ msg.title }}</strong>
            <p>{{ msg.content }}</p>
            <span class="msg-time">{{ msg.created_at?.replace('T', ' ').slice(0, 16) || '' }}</span>
          </div>
        </div>
        <div v-if="!messages.length && !loading" class="msg-empty">暂无消息</div>
      </div>
    </section>
  </div>
</template>

<script setup>
import { onMounted, ref } from "vue";
import { listMessages, readMessage } from "@/api/membership.js";
import { useClientAuth } from "@/shared/auth/client-auth";

const loading = ref(false);
const messages = ref([]);
const { isLoggedIn } = useClientAuth();

async function loadMessages() {
  if (!isLoggedIn.value) return;
  loading.value = true;
  try {
    const r = await listMessages({ page: 1, page_size: 20 });
    if (r?.items) messages.value = r.items;
  } catch { /* empty */ }
  finally { loading.value = false; }
}

async function markRead(msg) {
  if (msg.is_read) return;
  msg.is_read = true;
  try { await readMessage(msg.id); } catch { msg.is_read = false; }
}

onMounted(loadMessages);
</script>

<style scoped>
.messages-page { max-width: 900px; }
.section { background: var(--bg-card); border: 1px solid var(--border); border-radius: var(--radius-lg); padding: 24px; }
.section-header { display: flex; align-items: baseline; gap: 12px; margin-bottom: 16px; }
.section-title { font-size: 20px; font-weight: 700; }
.section-subtitle { font-size: 13px; color: var(--text-secondary); }
.message-list { display: grid; gap: 8px; }
.message-item { display: flex; gap: 12px; padding: 14px; border-radius: var(--radius-sm); cursor: pointer; align-items: flex-start; }
.message-item.unread { border-left: 3px solid var(--accent-gold); }
.msg-dot { width: 8px; height: 8px; border-radius: 50%; background: transparent; margin-top: 6px; flex-shrink: 0; }
.msg-dot.active { background: var(--accent-gold); }
.msg-body { flex: 1; }
.msg-body strong { display: block; font-size: 14px; margin-bottom: 4px; }
.msg-body p { font-size: 13px; color: var(--text-secondary); line-height: 1.5; margin-bottom: 4px; }
.msg-time { font-size: 11px; color: var(--text-muted); }
.msg-empty { text-align: center; padding: 40px; color: var(--text-secondary); }
</style>
