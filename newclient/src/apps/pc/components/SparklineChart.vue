<template>
  <canvas ref="canvasRef" :width="width" :height="height" class="sparkline-canvas"></canvas>
</template>

<script setup>
import { onMounted, ref, watch } from "vue";

const props = defineProps({
  data: { type: Array, default: () => [] },
  width: { type: Number, default: 80 },
  height: { type: Number, default: 28 },
  color: { type: String, default: '' }
});

const canvasRef = ref(null);

function draw() {
  const canvas = canvasRef.value;
  if (!canvas || !props.data.length) return;
  const ctx = canvas.getContext("2d");
  const w = props.width, h = props.height;
  ctx.clearRect(0, 0, w, h);

  const vals = props.data;
  const min = Math.min(...vals);
  const max = Math.max(...vals);
  const range = max - min || 1;
  const isUp = vals[vals.length - 1] >= vals[0];
  const strokeColor = props.color || (isUp ? "#00c897" : "#ff4757");

  const points = vals.map((v, i) => ({
    x: (i / (vals.length - 1)) * (w - 4) + 2,
    y: h - 4 - ((v - min) / range) * (h - 8)
  }));

  ctx.beginPath();
  ctx.strokeStyle = strokeColor;
  ctx.lineWidth = 1.5;
  ctx.lineJoin = "round";
  ctx.moveTo(points[0].x, points[0].y);
  for (let i = 1; i < points.length; i++) {
    ctx.lineTo(points[i].x, points[i].y);
  }
  ctx.stroke();

  // Fill gradient
  const grad = ctx.createLinearGradient(0, 0, 0, h);
  grad.addColorStop(0, strokeColor + "33");
  grad.addColorStop(1, strokeColor + "00");
  ctx.lineTo(points[points.length - 1].x, h - 4);
  ctx.lineTo(points[0].x, h - 4);
  ctx.closePath();
  ctx.fillStyle = grad;
  ctx.fill();
}

onMounted(draw);
watch(() => props.data, draw, { deep: true });
</script>

<style scoped>
.sparkline-canvas { display: block; }
</style>
