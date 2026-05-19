<template>
  <div class="h5-chat">
    <!-- Messages -->
    <div class="h5-chat-messages" ref="messagesRef">
      <div v-for="(msg, i) in messages" :key="i" class="h5-msg" :class="'msg-' + msg.role">
        <div class="h5-msg-avatar">{{ msg.role === 'ai' ? 'AI' : '我' }}</div>
        <div class="h5-msg-content">
          <div class="h5-msg-bubble" v-html="msg.content"></div>
          <div v-if="msg.stocks?.length" class="h5-msg-stocks">
            <div v-for="s in msg.stocks" :key="s.symbol" class="h5-msg-stock">
              <span class="h5-stock-name">{{ s.symbol.split('.')[0] }}</span>
              <span class="h5-stock-change" :class="s.change >= 0 ? 'up' : 'down'">{{ s.change >= 0 ? '+' : '' }}{{ s.change }}%</span>
            </div>
          </div>
          <span class="h5-msg-time">{{ msg.time }}</span>
        </div>
      </div>

      <div v-if="isTyping" class="h5-msg msg-ai">
        <div class="h5-msg-avatar">AI</div>
        <div class="h5-msg-bubble">
          <div class="typing-dots"><span></span><span></span><span></span></div>
        </div>
      </div>
    </div>

    <!-- Suggestions -->
    <div class="h5-chat-suggestions" v-if="suggestions.length && !isTyping">
      <button v-for="s in suggestions" :key="s" class="h5-suggestion" @click="sendQuickQuestion(s)">
        {{ s }}
      </button>
    </div>

    <!-- Input -->
    <div class="h5-chat-input">
      <div class="h5-chat-input-wrap">
        <input v-model="inputText" type="text" placeholder="输入股票名称或问题..." @keydown.enter.prevent="sendMessage" />
      </div>
      <button class="h5-send-btn" :disabled="!inputText.trim() || isTyping" @click="sendMessage">
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" width="18" height="18"><line x1="22" y1="2" x2="11" y2="13"/><polygon points="22 2 15 22 11 13 2 9 22 2"/></svg>
      </button>
    </div>
  </div>
</template>

<script setup>
import { nextTick, ref } from "vue";

const inputText = ref("");
const isTyping = ref(false);
const messagesRef = ref(null);

const suggestions = ref([
  "今日热点分析",
  "推荐一只AI精选股",
  "宁德时代怎么样",
  "半导体板块如何"
]);

const messages = ref([
  {
    role: "ai",
    content: "你好！我是 智投AI，可以帮你分析股票、解读市场、提供投资建议。请问有什么可以帮你的？",
    time: "10:30"
  }
]);

function sendMessage() {
  const text = inputText.value.trim();
  if (!text || isTyping.value) return;

  messages.value.push({
    role: "user",
    content: text,
    time: new Date().toLocaleTimeString("zh-CN", { hour: "2-digit", minute: "2-digit" })
  });
  inputText.value = "";
  suggestions.value = [];
  scrollToBottom();

  isTyping.value = true;
  setTimeout(() => {
    const replies = [
      `根据 AI 分析，<strong>${text}</strong> 技术面呈现多头排列，资金面持续流入，建议关注量能变化。风险提示：AI 分析仅供参考。`,
      `关于 <strong>${text}</strong>：<br><br>✅ 行业景气度提升<br>⚠️ 估值处于中高水平<br>🎯 中长期看好，短期等待回调`,
      `AI 分析完成：趋势向好，目标价位 215-228，建议仓位 15% 以内。详细报告可查看桌面版。`
    ];
    messages.value.push({
      role: "ai",
      content: replies[Math.floor(Math.random() * replies.length)],
      time: new Date().toLocaleTimeString("zh-CN", { hour: "2-digit", minute: "2-digit" })
    });
    isTyping.value = false;
    scrollToBottom();
  }, 1200);
}

function sendQuickQuestion(q) {
  inputText.value = q;
  sendMessage();
}

function scrollToBottom() {
  nextTick(() => {
    if (messagesRef.value) {
      messagesRef.value.scrollTop = messagesRef.value.scrollHeight;
    }
  });
}
</script>

<style scoped>
.h5-chat {
  display: flex;
  flex-direction: column;
  height: calc(100vh - 140px);
}

.h5-chat-messages {
  flex: 1;
  overflow-y: auto;
  display: flex;
  flex-direction: column;
  gap: 12px;
  padding: 4px 0;
}

.h5-msg {
  display: flex;
  gap: 8px;
  max-width: 90%;
}

.msg-ai { align-self: flex-start; }
.msg-user { align-self: flex-end; flex-direction: row-reverse; }

.h5-msg-avatar {
  width: 28px;
  height: 28px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 10px;
  font-weight: 700;
  flex-shrink: 0;
}

.msg-ai .h5-msg-avatar {
  background: var(--accent-gold-glow);
  color: var(--accent-gold);
  border: 1px solid var(--border-gold);
}

.msg-user .h5-msg-avatar {
  background: rgba(59, 130, 246, 0.15);
  color: var(--accent-blue);
}

.h5-msg-bubble {
  padding: 10px 14px;
  border-radius: var(--radius-md);
  font-size: 13px;
  line-height: 1.7;
}

.msg-ai .h5-msg-bubble {
  background: rgba(255, 255, 255, 0.03);
  border: 1px solid var(--border);
}

.msg-user .h5-msg-bubble {
  background: var(--accent-gold-glow);
  border: 1px solid var(--border-gold);
}

.h5-msg-stocks {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
  margin-top: 8px;
}

.h5-msg-stock {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 6px 10px;
  border-radius: var(--radius-sm);
  border: 1px solid var(--border);
}

.h5-stock-name {
  font-size: 12px;
  font-weight: 600;
}

.h5-stock-change {
  font-size: 11px;
  font-weight: 600;
}

.h5-stock-change.up { color: var(--positive); }
.h5-stock-change.down { color: var(--negative); }

.h5-msg-time {
  display: block;
  font-size: 10px;
  color: var(--text-muted);
  margin-top: 4px;
}

/* Typing */
.typing-dots {
  display: flex;
  gap: 4px;
}

.typing-dots span {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: var(--text-muted);
  animation: typing 1.4s infinite;
}

.typing-dots span:nth-child(2) { animation-delay: 0.2s; }
.typing-dots span:nth-child(3) { animation-delay: 0.4s; }

@keyframes typing {
  0%, 100% { opacity: 0.3; }
  50% { opacity: 1; }
}

/* Suggestions */
.h5-chat-suggestions {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
  margin-bottom: 8px;
}

.h5-suggestion {
  padding: 6px 12px;
  border-radius: var(--radius-full);
  border: 1px solid var(--border);
  font-size: 12px;
  color: var(--text-secondary);
}

.h5-suggestion:active {
  border-color: var(--border-gold);
  color: var(--accent-gold);
  background: var(--accent-gold-glow);
}

/* Input */
.h5-chat-input {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 0;
  padding-bottom: max(8px, env(safe-area-inset-bottom));
}

.h5-chat-input-wrap {
  flex: 1;
  padding: 10px 14px;
  border-radius: var(--radius-full);
  background: rgba(255, 255, 255, 0.04);
  border: 1px solid var(--border);
}

.h5-chat-input-wrap input {
  width: 100%;
  background: none;
  color: var(--text-primary);
  font-size: 14px;
}

.h5-chat-input-wrap input::placeholder {
  color: var(--text-muted);
}

.h5-send-btn {
  width: 40px;
  height: 40px;
  border-radius: 50%;
  background: linear-gradient(135deg, var(--accent-gold), var(--accent-gold-dim));
  color: #000;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.h5-send-btn:disabled {
  opacity: 0.4;
}
</style>
