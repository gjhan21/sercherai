# New Client Homepage Demo Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 在 `newclient/h5demo` 产出一套已确认方向的首页 demo，包括 PC 首页、H5 首页与统一入口页。

**Architecture:** 采用纯静态 HTML + 内嵌 CSS 的方式输出高保真 demo，不接入真实 API，只以静态样例内容表达当前真实能力边界。PC 与 H5 分文件设计，共享同一套蓝金金融视觉系统，但分别收敛到中文金融站首页和国内金融 App 首页。

**Tech Stack:** HTML5、CSS3、少量原生 JavaScript、静态文档输出

---

### Task 1: 固化设计边界

**Files:**
- Create: `docs/superpowers/specs/2026-03-24-newclient-homepage-demo-design.md`
- Modify: `newclient/h5demo/现有client接口与功能盘点.md`

- [ ] Step 1: 根据已确认内容整理首页设计边界
- [ ] Step 2: 明确 PC 与 H5 首页模块顺序
- [ ] Step 3: 明确不允许扩展的能力边界
- [ ] Step 4: 写入 spec 文档，作为 demo 的约束来源

### Task 2: 规划 demo 输出文件

**Files:**
- Create: `docs/superpowers/plans/2026-03-24-newclient-homepage-demo.md`
- Create: `newclient/h5demo/home-demo-index.html`
- Create: `newclient/h5demo/home-pc-demo.html`
- Create: `newclient/h5demo/home-h5-demo.html`

- [ ] Step 1: 明确 demo 文件命名
- [ ] Step 2: 规划入口页职责
- [ ] Step 3: 规划 PC / H5 页面分工
- [ ] Step 4: 明确交付审查方式

### Task 3: 输出 PC 首页 demo

**Files:**
- Create: `newclient/h5demo/home-pc-demo.html`

- [ ] Step 1: 写顶部导航与页面壳层
- [ ] Step 2: 写推荐主 Hero 与操作区
- [ ] Step 3: 写推荐理由与执行边界模块
- [ ] Step 4: 写历史推荐样本预览模块
- [ ] Step 5: 写焦点研报与市场资讯侧栏
- [ ] Step 6: 写底部信息区
- [ ] Step 7: 手动检查 PC 首页信息顺序与统一色系

### Task 4: 输出 H5 首页 demo

**Files:**
- Create: `newclient/h5demo/home-h5-demo.html`

- [ ] Step 1: 写 H5 顶部品牌头与主线 Hero
- [ ] Step 2: 写快捷入口宫格与数据条
- [ ] Step 3: 写推荐主卡与理由说明
- [ ] Step 4: 写焦点研报卡
- [ ] Step 5: 写历史样本与资讯列表
- [ ] Step 6: 写精简底部信息与底部 Tab
- [ ] Step 7: 手动检查页面是否足够像国内金融 App 首页

### Task 5: 输出统一入口页并做静态检查

**Files:**
- Create: `newclient/h5demo/home-demo-index.html`

- [ ] Step 1: 写入口说明与两个 demo 的链接卡片
- [ ] Step 2: 补充“设计边界说明”
- [ ] Step 3: 本地检查三个 HTML 文件存在且可打开
- [ ] Step 4: 整理交付说明，供后续继续设计其他页面
