<template>
  <div class="demo-selector-page">
    <div class="selector-shell">
      <header class="selector-hero">
        <p class="hero-kicker">PC HOME DEMO MATRIX</p>
        <h1>首页方案对比台</h1>
        <p class="hero-copy">
          当前版本已经原样保存。下面给你 3 个新方向可以直接打开对比：A 更像国际投行官网，B 更像权威官网 + 高端 SaaS 门户，C 更像精品研究机构官网。
        </p>
      </header>

      <section class="selector-grid">
        <article v-for="item in variants" :key="item.key" class="variant-card">
          <p class="variant-tag">{{ item.tag }}</p>
          <h2>{{ item.title }}</h2>
          <p class="variant-desc">{{ item.desc }}</p>
          <ul class="variant-list">
            <li v-for="point in item.points" :key="point">{{ point }}</li>
          </ul>
          <RouterLink class="variant-link" :to="item.path">打开这个方案</RouterLink>
        </article>
      </section>
    </div>
  </div>
</template>

<script setup>
const variants = [
  {
    key: "current",
    tag: "Current Baseline",
    title: "当前保存版本",
    desc: "你刚才确认过的现有首页，已备份保留，随时可以回退。",
    points: ["保留当前视觉语言", "继续作为回退基线", "路径：/demo/current"],
    path: "/demo/current"
  },
  {
    key: "a",
    tag: "Option A",
    title: "纯投行官网感",
    desc: "更克制、更权威、更像国际机构官网，强调品牌信任与旗舰观点。",
    points: ["大留白与大标题", "更少产品味", "更强机构权威感"],
    path: "/demo/a"
  },
  {
    key: "b",
    tag: "Option B",
    title: "权威官网 + SaaS 门户",
    desc: "在金融官网气质上，增强产品价值表达和入口分流能力。",
    points: ["官网感和商业化平衡", "更适合作为真实产品首页", "我当前最推荐"],
    path: "/demo/b"
  },
  {
    key: "c",
    tag: "Option C",
    title: "精品研究机构官网",
    desc: "内容感、编辑感更强，像研究机构的高端门户或研究月刊首页。",
    points: ["更有内容气质", "更像研究机构门面", "商业转化更弱一些"],
    path: "/demo/c"
  }
];
</script>

<style scoped>
.demo-selector-page {
  min-height: 100%;
  color: #f1f6ff;
  background:
    radial-gradient(circle at 20% 0%, rgba(88, 128, 219, 0.16), transparent 24%),
    radial-gradient(circle at 85% 18%, rgba(224, 171, 105, 0.14), transparent 20%),
    linear-gradient(180deg, #09111c 0%, #0d1624 100%);
}

.selector-shell {
  width: min(1220px, calc(100% - 32px));
  margin: 0 auto;
  padding: 48px 0 44px;
}

.selector-hero {
  max-width: 760px;
}

.hero-kicker,
.variant-tag {
  margin: 0;
  color: #8fb3ff;
  font-size: 12px;
  letter-spacing: 0.24em;
  text-transform: uppercase;
}

.selector-hero h1 {
  margin: 16px 0 14px;
  font-size: clamp(40px, 6vw, 76px);
  line-height: 0.98;
  letter-spacing: -0.04em;
}

.hero-copy {
  color: #a9b9cf;
  font-size: 16px;
  line-height: 1.8;
}

.selector-grid {
  margin-top: 34px;
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 18px;
}

.variant-card {
  padding: 26px;
  border: 1px solid rgba(149, 170, 205, 0.18);
  background: rgba(12, 21, 34, 0.82);
  backdrop-filter: blur(12px);
  box-shadow: 0 24px 48px rgba(0, 0, 0, 0.24);
}

.variant-card h2 {
  margin: 14px 0 12px;
  font-size: 30px;
}

.variant-desc {
  color: #a6b5c9;
  line-height: 1.78;
}

.variant-list {
  margin: 18px 0 24px;
  padding-left: 18px;
  color: #dce7f9;
  line-height: 1.9;
}

.variant-link {
  display: inline-flex;
  padding: 11px 18px;
  border-radius: 999px;
  color: #09111d;
  text-decoration: none;
  background: linear-gradient(135deg, #b7cdff, #efc997);
}

@media (max-width: 860px) {
  .selector-grid {
    grid-template-columns: 1fr;
  }
}
</style>
