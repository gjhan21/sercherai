<template>
  <div class="chat-page">
    <div class="chat-layout">
      <!-- Sidebar conversations -->
      <aside class="chat-sidebar glass">
        <div class="chat-sidebar-header">
          <h2>对话记录</h2>
          <button class="btn-icon" @click="startNewChat" title="新建对话">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" width="18" height="18"><line x1="12" y1="5" x2="12" y2="19"/><line x1="5" y1="12" x2="19" y2="12"/></svg>
          </button>
        </div>
        <div class="chat-history">
          <div v-for="conv in conversations" :key="conv.id" class="chat-history-item" :class="{ active: conv.id === activeConversation }" @click="activeConversation = conv.id">
            <div class="history-icon">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" width="16" height="16"><path d="M21 15a2 2 0 01-2 2H7l-4 4V5a2 2 0 012-2h14a2 2 0 012 2z"/></svg>
            </div>
            <div class="history-text">
              <strong>{{ conv.title }}</strong>
              <span>{{ conv.time }}</span>
            </div>
          </div>
        </div>
        <div class="chat-ai-status">
          <span class="status-dot"></span>
          <span>AI 在线</span>
        </div>
      </aside>

      <!-- Main Chat Area -->
      <div class="chat-main glass">
        <!-- Messages -->
        <div class="chat-messages" ref="messagesRef">
          <div v-for="(msg, i) in messages" :key="i" class="message" :class="'msg-' + msg.role">
            <div class="msg-avatar">
              <span v-if="msg.role === 'ai'">AI</span>
              <span v-else>我</span>
            </div>
            <div class="msg-content">
              <div class="msg-bubble" v-html="renderMessage(msg.content)"></div>
              <div v-if="msg.stocks?.length" class="msg-stocks">
                <div v-for="s in msg.stocks" :key="s.symbol" class="msg-stock-chip" @click="$router.push('/analysis')">
                  <span class="chip-symbol">{{ s.symbol }}</span>
                  <span class="chip-name">{{ s.name }}</span>
                  <span class="chip-price">{{ s.price }}</span>
                  <span class="chip-change" :class="s.change >= 0 ? 'up' : 'down'">{{ s.change >= 0 ? '+' : '' }}{{ s.change }}%</span>
                </div>
              </div>
              <span class="msg-time">{{ msg.time }}</span>
            </div>
          </div>

          <!-- Typing indicator -->
          <div v-if="isTyping" class="message msg-ai">
            <div class="msg-avatar"><span>AI</span></div>
            <div class="msg-content">
              <div class="typing-indicator">
                <span></span><span></span><span></span>
              </div>
            </div>
          </div>
        </div>

        <!-- Suggestions -->
        <div class="chat-suggestions" v-if="suggestions.length && !isTyping">
          <button v-for="s in suggestions" :key="s" class="suggestion-chip" @click="sendQuickQuestion(s)">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" width="12" height="12"><path d="M12 2v4M12 18v4M4.93 4.93l2.83 2.83M16.24 16.24l2.83 2.83M2 12h4M18 12h4"/></svg>
            {{ s }}
          </button>
        </div>

        <!-- Input -->
        <div class="chat-input-area">
          <div class="chat-input-wrap">
            <input
              ref="inputRef"
              v-model="inputText"
              type="text"
              placeholder="向 AI 询问股票分析、市场趋势..."
              @keydown.enter.prevent="sendMessage"
            />
          </div>
          <button class="chat-send-btn" :disabled="!inputText.trim() || isTyping" @click="sendMessage">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" width="18" height="18"><line x1="22" y1="2" x2="11" y2="13"/><polygon points="22 2 15 22 11 13 2 9 22 2"/></svg>
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { computed, nextTick, ref } from "vue";

const inputText = ref("");
const isTyping = ref(false);
const activeConversation = ref("conv-1");
const messagesRef = ref(null);
const inputRef = ref(null);

const conversations = ref([
  { id: "conv-1", title: "今日市场分析", time: "10:32" },
  { id: "conv-2", title: "宁德时代深度分析", time: "昨天" },
  { id: "conv-3", title: "半导体板块机会", time: "昨天" },
  { id: "conv-4", title: "投资组合建议", time: "3天前" },
  { id: "conv-5", title: "比亚迪 vs 特斯拉", time: "5天前" }
]);

const suggestions = ref([
  "今日市场热点分析",
  "宁德时代值得买入吗",
  "半导体板块怎么看",
  "推荐3只AI精选股票"
]);

const messages = ref([
  {
    role: "ai",
    content: "你好！我是 智投AI 助手，可以帮你分析股票、解读市场、提供投资建议。请问有什么可以帮你的？",
    time: "10:30"
  },
  {
    role: "user",
    content: "帮我分析一下今天市场情况",
    time: "10:31"
  },
  {
    role: "ai",
    content: "📊 <strong>今日市场概览</strong><br><br>今日市场整体偏强，三大指数集体上涨。AI 模型综合评分 <strong>78.6</strong>，市场情绪偏积极。<br><br>• <strong>上证指数</strong> 3,286.54 <span style='color:var(--positive)'>+1.28%</span><br>• <strong>深证成指</strong> 10,842.63 <span style='color:var(--positive)'>+1.86%</span><br>• <strong>创业板指</strong> 2,168.74 <span style='color:var(--positive)'>+2.35%</span><br><br>今日重点关注 <strong>半导体</strong>、<strong>人工智能</strong> 板块，资金流入明显。",
    time: "10:31",
    stocks: [
      { symbol: "300750.SZ", name: "宁德时代", price: "198.62", change: 3.45 },
      { symbol: "002415.SZ", name: "海康威视", price: "35.27", change: 2.65 },
      { symbol: "688981.SH", name: "中芯国际", price: "56.78", change: 4.12 }
    ]
  }
]);

function renderMessage(content) {
  return content;
}

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

  // Simulate AI response
  isTyping.value = true;
  setTimeout(() => {
    const responses = [
      `根据 AI 模型分析，<strong>${text}</strong> 相关问题，我的建议如下：<br><br>1️⃣ 该标的目前技术面呈现多头排列，趋势向好<br>2️⃣ 资金面持续流入，主力资金近5日净流入显著<br>3️⃣ 建议关注量能变化，若持续放量可考虑分批建仓<br><br>⚠️ 风险提示：以上分析由 AI 生成，仅供参考，不构成投资建议。`,
      `关于 <strong>${text}</strong>，AI 模型综合分析如下：<br><br>✅ <strong>积极因素</strong>：行业景气度提升，政策面利好<br>⚠️ <strong>风险因素</strong>：估值处于历史中高水平，短期存在回调可能<br>🎯 <strong>建议策略</strong>：中长期看好，短期可等待回调建仓`,
      `AI 深度分析完成：<br><br>📈 <strong>趋势判断</strong>：中期上升趋势确立<br>💰 <strong>资金信号</strong>：主力资金连续流入，北向资金加速配置<br>🎯 <strong>目标价位</strong>：AI 预测1个月内目标区间 215-228<br>🛡️ <strong>风控建议</strong>：建议仓位控制在总资金 15% 以内，跌破支撑位止损`
    ];
    const reply = responses[Math.floor(Math.random() * responses.length)];

    messages.value.push({
      role: "ai",
      content: reply,
      time: new Date().toLocaleTimeString("zh-CN", { hour: "2-digit", minute: "2-digit" }),
      stocks: [
        { symbol: "300750.SZ", name: "宁德时代", price: "198.62", change: 3.45 },
        { symbol: "600941.SH", name: "中国移动", price: "106.80", change: 1.82 }
      ]
    });
    isTyping.value = false;
    scrollToBottom();
  }, 1500);
}

function sendQuickQuestion(q) {
  inputText.value = q;
  sendMessage();
}

function startNewChat() {
  const id = `conv-${conversations.value.length + 1}`;
  conversations.value.unshift({ id, title: "新对话", time: "刚刚" });
  activeConversation.value = id;
  messages.value = [
    {
      role: "ai",
      content: "你好！我是 智投AI 助手，可以帮你分析股票、解读市场、提供投资建议。请问有什么可以帮你的？",
      time: "刚刚"
    }
  ];
  suggestions.value = [
    "今日市场热点分析",
    "宁德时代值得买入吗",
    "半导体板块怎么看",
    "推荐3只AI精选股票"
  ];
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
.chat-page {
  height: calc(100vh - 140px);
}

.chat-layout {
  display: grid;
  grid-template-columns: 240px 1fr;
  gap: 16px;
  height: 100%;
}

.chat-sidebar {
  padding: 16px;
  border-radius: var(--radius-lg);
  display: flex;
  flex-direction: column;
}

.chat-sidebar-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 16px;
}

.chat-sidebar-header h2 {
  font-size: 15px;
}

.btn-icon {
  width: 32px;
  height: 32px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--text-secondary);
  transition: all 0.2s;
}

.btn-icon:hover {
  background: rgba(255, 255, 255, 0.06);
  color: var(--text-primary);
}

.chat-history {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 4px;
  overflow-y: auto;
}

.chat-history-item {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 10px;
  border-radius: var(--radius-sm);
  cursor: pointer;
  transition: all 0.15s;
}

.chat-history-item:hover {
  background: rgba(255, 255, 255, 0.04);
}

.chat-history-item.active {
  background: var(--accent-gold-glow);
  border: 1px solid var(--border-gold);
}

.history-icon {
  color: var(--text-muted);
  flex-shrink: 0;
}

.history-text strong {
  display: block;
  font-size: 13px;
  line-height: 1.3;
}

.history-text span {
  font-size: 11px;
  color: var(--text-muted);
}

.chat-ai-status {
  display: flex;
  align-items: center;
  gap: 8px;
  padding-top: 12px;
  border-top: 1px solid var(--border);
  font-size: 12px;
  color: var(--positive);
}

.chat-ai-status .status-dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: var(--positive);
  animation: pulse-glow 2s infinite;
}

/* Main Chat */
.chat-main {
  display: flex;
  flex-direction: column;
  border-radius: var(--radius-lg);
  overflow: hidden;
}

.chat-messages {
  flex: 1;
  overflow-y: auto;
  padding: 20px;
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.message {
  display: flex;
  gap: 10px;
  max-width: 85%;
}

.msg-ai { align-self: flex-start; }
.msg-user { align-self: flex-end; flex-direction: row-reverse; }

.msg-avatar {
  width: 32px;
  height: 32px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 11px;
  font-weight: 700;
  flex-shrink: 0;
}

.msg-ai .msg-avatar {
  background: var(--accent-gold-glow);
  color: var(--accent-gold);
  border: 1px solid var(--border-gold);
}

.msg-user .msg-avatar {
  background: rgba(59, 130, 246, 0.15);
  color: var(--accent-blue);
}

.msg-bubble {
  padding: 12px 16px;
  border-radius: var(--radius-md);
  font-size: 13px;
  line-height: 1.7;
}

.msg-ai .msg-bubble {
  background: rgba(255, 255, 255, 0.03);
  border: 1px solid var(--border);
}

.msg-user .msg-bubble {
  background: var(--accent-gold-glow);
  border: 1px solid var(--border-gold);
}

.msg-time {
  display: block;
  font-size: 11px;
  color: var(--text-muted);
  margin-top: 4px;
}

.msg-stocks {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  margin-top: 8px;
}

.msg-stock-chip {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 12px;
  border-radius: var(--radius-sm);
  border: 1px solid var(--border);
  cursor: pointer;
  transition: all 0.2s;
  background: rgba(255, 255, 255, 0.02);
}

.msg-stock-chip:hover {
  border-color: var(--border-gold);
}

.chip-symbol {
  font-family: var(--font-mono);
  font-size: 11px;
  color: var(--text-secondary);
}

.chip-name {
  font-size: 12px;
  font-weight: 600;
}

.chip-price {
  font-size: 12px;
  font-weight: 600;
}

.chip-change {
  font-size: 12px;
  font-weight: 600;
}

.chip-change.up { color: var(--positive); }
.chip-change.down { color: var(--negative); }

/* Typing */
.typing-indicator {
  display: flex;
  gap: 4px;
  padding: 12px 16px;
}

.typing-indicator span {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background: var(--text-muted);
  animation: typing 1.4s infinite;
}

.typing-indicator span:nth-child(2) { animation-delay: 0.2s; }
.typing-indicator span:nth-child(3) { animation-delay: 0.4s; }

@keyframes typing {
  0%, 100% { opacity: 0.3; transform: translateY(0); }
  50% { opacity: 1; transform: translateY(-4px); }
}

/* Suggestions */
.chat-suggestions {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  padding: 0 20px 12px;
}

.suggestion-chip {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 8px 14px;
  border-radius: var(--radius-full);
  border: 1px solid var(--border);
  font-size: 12px;
  color: var(--text-secondary);
  transition: all 0.2s;
}

.suggestion-chip:hover {
  border-color: var(--border-gold);
  color: var(--accent-gold);
  background: var(--accent-gold-glow);
}

/* Input */
.chat-input-area {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 12px 20px;
  border-top: 1px solid var(--border);
}

.chat-input-wrap {
  flex: 1;
  padding: 10px 16px;
  border-radius: var(--radius-full);
  background: rgba(255, 255, 255, 0.04);
  border: 1px solid var(--border);
}

.chat-input-wrap input {
  width: 100%;
  background: none;
  color: var(--text-primary);
  font-size: 14px;
}

.chat-input-wrap input::placeholder {
  color: var(--text-muted);
}

.chat-send-btn {
  width: 40px;
  height: 40px;
  border-radius: 50%;
  background: linear-gradient(135deg, var(--accent-gold), var(--accent-gold-dim));
  color: #000;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: opacity 0.2s;
  flex-shrink: 0;
}

.chat-send-btn:disabled {
  opacity: 0.4;
  cursor: default;
}

.chat-send-btn:not(:disabled):hover {
  opacity: 0.9;
}

@media (max-width: 900px) {
  .chat-sidebar {
    display: none;
  }
  .chat-layout {
    grid-template-columns: 1fr;
  }
}
</style>
