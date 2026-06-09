<template>
  <div class="kline-wrapper" ref="wrapperRef">
    <canvas ref="canvasRef" class="kline-canvas"></canvas>
    <div v-if="hoverData" class="kline-tooltip" :style="tooltipStyle">
      <div class="kt-date">{{ hoverData.date }}</div>
      <div class="kt-row"><span>开</span><strong>{{ hoverData.open.toFixed(2) }}</strong></div>
      <div class="kt-row"><span>收</span><strong :style="{color:hoverData.close >= hoverData.open?'var(--positive)':'var(--negative)'}">{{ hoverData.close.toFixed(2) }}</strong></div>
      <div class="kt-row"><span>高</span><strong>{{ hoverData.high.toFixed(2) }}</strong></div>
      <div class="kt-row"><span>低</span><strong>{{ hoverData.low.toFixed(2) }}</strong></div>
      <div class="kt-row" v-if="!hoverData.isPredict"><span>量</span><strong>{{ (hoverData.volume / 10000).toFixed(0) }}万</strong></div>
      <div class="kt-row" v-else><span>类型</span><strong style="color: #8b5cf6">AI 7D 预测</strong></div>
    </div>
  </div>
</template>

<script setup>
import { computed, onMounted, onUnmounted, ref, watch } from "vue";

const props = defineProps({
  data: { type: Array, default: () => [] },
  width: { type: Number, default: 600 },
  height: { type: Number, default: 380 },
  showMa: { type: Boolean, default: true },
  prediction: { type: Object, default: null },
  matches: { type: Array, default: () => [] }
});

const emit = defineEmits(["loaded"]);

const canvasRef = ref(null);
const wrapperRef = ref(null);
const hoverData = ref(null);
const tooltipStyle = ref({});
let mouseX = -1;
let mouseY = -1;
let animFrame = null;

const PADDING = { top: 20, right: 20, bottom: 45, left: 55 };
const VOLUME_HEIGHT = 60;

const chartW = computed(() => props.width - PADDING.left - PADDING.right);
const chartH = computed(() => props.height - PADDING.top - PADDING.bottom - VOLUME_HEIGHT - 10);
const volumeTop = computed(() => PADDING.top + chartH.value + 10);

function draw() {
  const canvas = canvasRef.value;
  if (!canvas || !props.data.length) return;
  const ctx = canvas.getContext("2d");
  const dpr = window.devicePixelRatio || 1;
  canvas.width = props.width * dpr;
  canvas.height = props.height * dpr;
  ctx.scale(dpr, dpr);
  ctx.clearRect(0, 0, props.width, props.height);

  const hasPred = props.prediction && props.prediction.median && props.prediction.median.length;
  const predN = hasPred ? props.prediction.median.length : 0;
  const reservePx = hasPred ? predN * 12 + 20 : 0; // reserve space for prediction candles
  const pts = hasPred ? props.data.slice(-Math.max(20, props.data.length - Math.floor((chartW.value - reservePx) / (chartW.value / props.data.length)))) : props.data;
  const n = pts.length;
  const cw = chartW.value - (hasPred ? reservePx : 0);
  const ch = chartH.value;
  const barWidth = Math.max(3, Math.min(10, cw / n * 0.65));
  const gap = cw / n;

  // Price range (include prediction values)
  let minP = Infinity, maxP = -Infinity, maxV = 0;
  pts.forEach(p => {
    if (p.low < minP) minP = p.low;
    if (p.high > maxP) maxP = p.high;
    if (p.volume > maxV) maxV = p.volume;
  });
  const predRange = props.prediction;
  if (predRange && predRange.median && pts.length > 0) {
    const last = pts[pts.length - 1].close;
    [predRange.median, predRange.upper_75, predRange.lower_25].forEach(arr => {
      (arr || []).forEach(pct => {
        const price = last * (1 + pct / 100);
        if (price < minP) minP = price;
        if (price > maxP) maxP = price;
      });
    });
  }
  const pad = (maxP - minP) * 0.08 || 5;
  minP -= pad; maxP += pad;

  function px(i) { return PADDING.left + i * gap + gap / 2; }
  function py(v) { return PADDING.top + ch - ((v - minP) / (maxP - minP)) * ch; }
  function vy(v) { return volumeTop.value + VOLUME_HEIGHT - (v / maxV) * VOLUME_HEIGHT; }

  // Grid
  ctx.strokeStyle = "rgba(255,255,255,0.04)";
  ctx.lineWidth = 0.5;
  for (let i = 0; i <= 4; i++) {
    const y = PADDING.top + (ch / 4) * i;
    ctx.beginPath(); ctx.moveTo(PADDING.left, y); ctx.lineTo(props.width - PADDING.right, y); ctx.stroke();
    // Price labels
    const price = maxP - ((maxP - minP) / 4) * i;
    ctx.fillStyle = "rgba(255,255,255,0.35)";
    ctx.font = "10px sans-serif";
    ctx.textAlign = "right";
    ctx.fillText(price.toFixed(2), PADDING.left - 5, y + 3);
  }

  // Volume bars
  pts.forEach((p, i) => {
    const x = px(i) - barWidth / 2;
    const h = vy(p.volume / (maxV || 1) * maxV);
    ctx.fillStyle = p.close >= p.open ? "rgba(0,200,151,0.3)" : "rgba(255,71,87,0.3)";
    ctx.fillRect(x, h, barWidth, volumeTop.value + VOLUME_HEIGHT - h);
  });

  // MA lines
  if (props.showMa) {
    drawMA(ctx, pts, "ma5", "#f0b90b", px, py);
    drawMA(ctx, pts, "ma10", "#3b82f6", px, py);
    drawMA(ctx, pts, "ma20", "#8b5cf6", px, py);
  }

  // Candlesticks + hover
  pts.forEach((p, i) => {
    const x = px(i);
    const x1 = x - barWidth / 2;
    const wickH = py(p.high) - py(p.low);
    const rawBody = py(Math.min(p.open, p.close)) - py(Math.max(p.open, p.close));	    const bodyH = Math.max(4, Math.min(rawBody * 3, rawBody + 6));	    const bodyY = py(Math.max(p.open, p.close)) - (bodyH - rawBody) / 2;

    const isUp = p.close >= p.open;
    const color = isUp ? "#00c897" : "#ff4757";

    // Wick
    ctx.strokeStyle = color;
    ctx.lineWidth = 1;
    ctx.beginPath(); ctx.moveTo(x, py(p.high)); ctx.lineTo(x, py(p.low)); ctx.stroke();

    // Body
    ctx.fillStyle = isUp ? "rgba(0,200,151,0.85)" : "rgba(255,71,87,0.85)";
    ctx.fillRect(x1, bodyY, barWidth, bodyH);

    // Highlight on hover
    if (mouseX >= 0) {
      const dist = Math.abs(x - mouseX);
      if (dist < gap / 2 && dist < 15) {
        ctx.strokeStyle = "rgba(255,255,255,0.15)";
        ctx.lineWidth = 1;
        ctx.setLineDash([3, 3]);
        ctx.beginPath(); ctx.moveTo(x, PADDING.top); ctx.lineTo(x, PADDING.top + ch); ctx.stroke();
        ctx.setLineDash([]);
        hoverData.value = p;
        tooltipStyle.value = { left: Math.min(x + 15, props.width - 160) + "px", top: Math.max(PADDING.top, mouseY - 80) + "px" };
      }
    }
  });

  // Date labels
  const step = Math.max(1, Math.floor(n / 8));
  pts.forEach((p, i) => {
    if (i % step !== 0 && i !== n - 1) return;
    ctx.fillStyle = "rgba(255,255,255,0.35)";
    ctx.font = "10px sans-serif";
    ctx.textAlign = "center";
    ctx.fillText(p.date.slice(5), px(i), props.height - 10);
  });

  // Prediction overlay as candlesticks
  const pred = props.prediction;
  if (pred && pred.median && pred.median.length && pts.length > 0) {
    const lastCandle = pts[pts.length - 1];
    const lastClose = lastCandle.close;
    const lastX = px(pts.length - 1);
    const predCount = pred.median.length;
    const availPx = (props.width - PADDING.right) - lastX - 8;
    const predStep = Math.max(6, Math.min(14, availPx / (predCount + 0.5)));
    const predBarW = Math.max(3, predStep * 0.45);
    ctx.strokeStyle = "rgba(139,92,246,0.4)";
    ctx.lineWidth = 1.5;
    ctx.setLineDash([4, 4]);
    ctx.beginPath(); ctx.moveTo(lastX, PADDING.top); ctx.lineTo(lastX, PADDING.top + ch);
    ctx.stroke();
    ctx.setLineDash([]);
    ctx.fillStyle = "rgba(139,92,246,0.8)";
    ctx.font = "10px sans-serif";
    ctx.textAlign = "left";
    ctx.fillText("▎预测区", lastX + 4, PADDING.top + 12);

    let prevClose = lastClose;
    let maxHighVal = -Infinity, minLowVal = Infinity;
    let maxHighIdx = -1, minLowIdx = -1;
    const predPts = [];

    for (let d = 0; d < predCount; d++) {
      const cumRet = pred.median[d];
      const open = prevClose;
      const close = lastClose * (1 + cumRet / 100);
      const upPct = pred.upper_75 && pred.upper_75[d] !== undefined ? pred.upper_75[d] : cumRet;
      const lowPct = pred.lower_25 && pred.lower_25[d] !== undefined ? pred.lower_25[d] : cumRet;
      let high = lastClose * (1 + Math.max(upPct, cumRet) / 100);
      let low = lastClose * (1 + Math.min(lowPct, cumRet) / 100);
      high = Math.max(high, open, close);
      low = Math.min(low, open, close);
      prevClose = close;

      const cx = lastX + (d + 1) * predStep;
      const cw2 = predBarW / 2;

      predPts.push({ d: d + 1, date: `D${d+1} (预测)`, open, close, high, low, cx });

      if (high > maxHighVal) {
        maxHighVal = high;
        maxHighIdx = d;
      }
      if (low < minLowVal) {
        minLowVal = low;
        minLowIdx = d;
      }

      ctx.strokeStyle = "rgba(139,92,246,0.6)";
      ctx.lineWidth = 1;
      ctx.beginPath();
      ctx.moveTo(cx, Math.max(PADDING.top, Math.min(PADDING.top + ch, py(high))));
      ctx.lineTo(cx, Math.max(PADDING.top, Math.min(PADDING.top + ch, py(low))));
      ctx.stroke();

      const bodyTop = py(Math.max(open, close));
      const bodyBtm = py(Math.min(open, close));
      const bodyH = Math.max(2, bodyBtm - bodyTop);
      const isUp = close >= open;
      ctx.fillStyle = isUp ? "rgba(139,92,246,0.55)" : "rgba(99,52,206,0.55)";
      ctx.fillRect(cx - cw2, bodyTop, predBarW, bodyH);

      ctx.fillStyle = "rgba(139,92,246,0.5)";
      ctx.font = "9px sans-serif";
      ctx.textAlign = "center";
      ctx.fillText("D" + (d + 1), cx, props.height - 10);
    }

    // 标出预测期的极高极低点
    if (maxHighIdx !== -1) {
      const p = predPts[maxHighIdx];
      const y = py(p.high);
      ctx.fillStyle = "#ff4757";
      ctx.beginPath(); ctx.arc(p.cx, y, 3, 0, Math.PI * 2); ctx.fill();
      ctx.font = "bold 10px sans-serif";
      ctx.textAlign = p.cx > props.width - 100 ? "right" : "left";
      ctx.fillText(`▲最高: ${maxHighVal.toFixed(2)}`, p.cx > props.width - 100 ? p.cx - 6 : p.cx + 6, y - 2);
    }
    if (minLowIdx !== -1) {
      const p = predPts[minLowIdx];
      const y = py(p.low);
      ctx.fillStyle = "#00c897";
      ctx.beginPath(); ctx.arc(p.cx, y, 3, 0, Math.PI * 2); ctx.fill();
      ctx.font = "bold 10px sans-serif";
      ctx.textAlign = p.cx > props.width - 100 ? "right" : "left";
      ctx.fillText(`▼最低: ${minLowVal.toFixed(2)}`, p.cx > props.width - 100 ? p.cx - 6 : p.cx + 6, y + 8);
    }

    // 预测 K 线的 Hover 交互
    predPts.forEach(p => {
      if (mouseX >= 0) {
        const dist = Math.abs(p.cx - mouseX);
        if (dist < predStep / 2 && dist < 15) {
          ctx.strokeStyle = "rgba(139,92,246,0.3)";
          ctx.lineWidth = 1;
          ctx.setLineDash([3, 3]);
          ctx.beginPath(); ctx.moveTo(p.cx, PADDING.top); ctx.lineTo(p.cx, PADDING.top + ch); ctx.stroke();
          ctx.setLineDash([]);
          
          hoverData.value = {
            date: p.date,
            open: p.open,
            close: p.close,
            high: p.high,
            low: p.low,
            isPredict: true
          };
          tooltipStyle.value = { 
            left: Math.min(p.cx + 15, props.width - 160) + "px", 
            top: Math.max(PADDING.top, mouseY - 80) + "px" 
          };
        }
      }
    });

    // 绘制历史相似印证均线折线
    const hasMatches = props.matches && props.matches.length;
    const matchAvgPath = [];
    if (hasMatches) {
      for (let d = 0; d < 7; d++) {
        let sumPct = 0;
        let count = 0;
        props.matches.forEach(m => {
          if (m.next_7d && m.next_7d[d] !== undefined) {
            sumPct += m.next_7d[d];
            count++;
          }
        });
        matchAvgPath.push(count > 0 ? sumPct / count : 0);
      }
    }

    if (hasMatches && matchAvgPath.length > 0) {
      ctx.strokeStyle = "rgba(59,130,246,0.75)";
      ctx.lineWidth = 1.5;
      ctx.setLineDash([3, 3]);
      ctx.beginPath();
      ctx.moveTo(lastX, py(lastClose));
      for (let d = 0; d < 7; d++) {
        const cx = lastX + (d + 1) * predStep;
        const price = lastClose * (1 + matchAvgPath[d] / 100);
        ctx.lineTo(cx, py(price));
      }
      ctx.stroke();
      ctx.setLineDash([]);

      const endX = lastX + 7 * predStep;
      const endPrice = lastClose * (1 + matchAvgPath[6] / 100);
      ctx.fillStyle = "rgba(59,130,246,0.95)";
      ctx.font = "9px sans-serif";
      ctx.textAlign = "left";
      ctx.fillText(" ── 历史相似均线", endX, py(endPrice) + 3);
    }
  }

  // MA legend
  if (props.showMa) {
    ctx.font = "11px sans-serif";
    ctx.textAlign = "left";
    let lx = PADDING.left + 10;
    const legends = [
      { label: "MA5", color: "#f0b90b" },
      { label: "MA10", color: "#3b82f6" },
      { label: "MA20", color: "#8b5cf6" }
    ];
    legends.forEach(l => {
      ctx.fillStyle = l.color;
      ctx.fillRect(lx, PADDING.top + 4, 12, 2);
      ctx.fillText(l.label, lx + 16, PADDING.top + 8);
      lx += 60;
    });
  }
}

function drawMA(ctx, pts, key, color, px, py) {
  ctx.strokeStyle = color;
  ctx.lineWidth = 1;
  ctx.beginPath();
  pts.forEach((p, i) => {
    const val = p[key];
    if (val === undefined || val === null) return;
    const x = px(i), y = py(val);
    if (i === 0 || pts[i - 1][key] === undefined) ctx.moveTo(x, y);
    else ctx.lineTo(x, y);
  });
  ctx.stroke();
}

function handleMouse(e) {
  const rect = canvasRef.value?.getBoundingClientRect();
  if (!rect) return;
  mouseX = e.clientX - rect.left;
  mouseY = e.clientY - rect.top;
  draw();
}

function handleLeave() {
  mouseX = -1; hoverData.value = null; draw();
}

onMounted(() => {
  const el = canvasRef.value;
  if (el) {
    el.addEventListener("mousemove", handleMouse);
    el.addEventListener("mouseleave", handleLeave);
  }
  draw();
});

onUnmounted(() => {
  const el = canvasRef.value;
  if (el) {
    el.removeEventListener("mousemove", handleMouse);
    el.removeEventListener("mouseleave", handleLeave);
  }
});

watch(() => props.data, draw, { deep: true });
watch(() => props.prediction, draw, { deep: true });
watch(() => props.matches, draw, { deep: true });
</script>

<style scoped>
.kline-wrapper { position: relative; }
.kline-canvas { display: block; width: 100%; cursor: crosshair; }
.kline-tooltip { position: absolute; background: rgba(11,14,26,0.92); border: 1px solid rgba(255,255,255,0.1); border-radius: 8px; padding: 10px 14px; font-size: 12px; pointer-events: none; z-index: 10; min-width: 120px; }
.kt-date { font-size: 11px; color: rgba(255,255,255,0.5); margin-bottom: 4px; }
.kt-row { display: flex; justify-content: space-between; gap: 16px; line-height: 1.6; }
.kt-row span { color: rgba(255,255,255,0.5); }
.kt-row strong { color: var(--text-primary); font-weight: 600; }
</style>
