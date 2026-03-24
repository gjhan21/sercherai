# New Client Strategies Demo Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 在 `newclient/h5demo` 产出 `strategies` 页的 PC / H5 正式 HTML demo，并提供统一入口页用于审查。

**Architecture:** 本页 demo 采用纯静态 HTML + 内嵌 CSS + 极少量原生 JavaScript 的方式实现，不接入真实 API，而是用静态样例内容准确表达当前 `client` 已有的策略能力边界。PC 页承接“中文金融站 / 推荐工作台”结构，H5 页承接“国内金融 App / 观点流 + 详情流”结构，两端共享蓝金金融视觉系统但不做镜像复制。

**Tech Stack:** HTML5、CSS3、少量原生 JavaScript、静态本地 HTTP 预览

---

### Task 1: 固化 strategies 页面设计边界

**Files:**
- Modify: `docs/superpowers/specs/2026-03-24-newclient-fullsite-demo-design.md`
- Create: `docs/superpowers/specs/2026-03-24-newclient-strategies-demo-design.md`

- [ ] **Step 1: 从全站 spec 中提取 strategies 页面约束**

确认以下设计边界必须沿用：

- PC：页面定位区、主推荐工作台、推荐理由、执行区间与风险边界、股票 / 期货 / 市场事件分层区、右侧阅读顺序说明
- H5：页面标题与筛选切换、推荐主卡、推荐理由卡、风险边界卡、观点流、期货机会与事件卡、底部 Tab
- 不新增大盘条、热榜、社区、短视频、直播和假行情图

- [ ] **Step 2: 写 strategies 单页 spec**

写入以下内容：

- 页面主任务
- PC 页面结构
- H5 页面结构
- 首页如何导流过来
- 本页不能做成什么样

- [ ] **Step 3: 校对 strategies spec 与全站 spec 一致**

检查 `strategies` 的定位、顺序、边界是否与全站 spec 完全一致，不出现新的导航项或新内容能力。

### Task 2: 规划 demo 输出文件与职责

**Files:**
- Create: `newclient/h5demo/strategies-pc-demo.html`
- Create: `newclient/h5demo/strategies-h5-demo.html`
- Create: `newclient/h5demo/strategies-demo-index.html`

- [ ] **Step 1: 明确三个输出文件职责**

约束如下：

- `strategies-pc-demo.html`：单独审查 PC 页面
- `strategies-h5-demo.html`：单独审查 H5 页面
- `strategies-demo-index.html`：统一入口、预览与边界说明

- [ ] **Step 2: 定义共享视觉规则**

至少统一以下内容：

- 蓝金金融系变量
- 标题与正文字体
- 卡片圆角、阴影、边框
- 标签、CTA、状态色的层级

- [ ] **Step 3: 定义每个文件的模块顺序**

PC 文件至少包括：

1. 顶部导航与页面定位
2. 主推荐 Hero
3. 推荐理由与执行边界
4. 股票 / 期货 / 市场事件分层区
5. 右侧阅读顺序与页面关系
6. 底部站点信息

H5 文件至少包括：

1. 顶部信息头
2. 栏目 / 过滤切换
3. 推荐主卡
4. 理由 / 风控卡
5. 观点流
6. 期货 / 事件承接卡
7. 底部 Tab

### Task 3: 输出 PC strategies demo

**Files:**
- Create: `newclient/h5demo/strategies-pc-demo.html`

- [ ] **Step 1: 写页面壳层与顶部导航**

要求：

- 继承首页 PC demo 的蓝金金融壳层
- 页面标题明确是“策略中心 / 推荐工作台”
- 保留首页、关注、资讯、档案、会员、我的导航语义

- [ ] **Step 2: 写主推荐 Hero**

内容要求：

- 股票主推荐作为第一主位
- 一眼看到结论、当前动作、风险等级和主理由
- CTA 保持为“看完整观点 / 去关注页 / 去资讯页”

- [ ] **Step 3: 写推荐理由与执行边界模块**

至少包含：

- 核心结论
- 证据来源
- 执行区间
- 风险边界

- [ ] **Step 4: 写股票 / 期货 / 市场事件分层区**

要求：

- 三类内容视觉分层，不混成统一列表
- 期货与事件是承接能力，不抢股票主推荐主位
- 不使用虚构行情图和热力图

- [ ] **Step 5: 写右侧导读区**

包含：

- 阅读顺序
- 风险提醒
- 与首页 / 关注页 / 资讯页 / 档案页的关系

- [ ] **Step 6: 写 PC 页面底部信息区**

要求：

- 结构和首页 PC demo 保持同一套站点语言
- 不出现“页面做到一半断掉”的感受

- [ ] **Step 7: 运行静态校验**

Run: `rg -n "<title>|推荐工作台|风险边界|市场事件|期货" /Users/gjhan21/cursor/sercherai/newclient/h5demo/strategies-pc-demo.html`

Expected:

- 输出包含页面标题
- 输出包含“推荐工作台”
- 输出包含“风险边界”
- 输出包含“市场事件”
- 输出包含“期货”

### Task 4: 输出 H5 strategies demo

**Files:**
- Create: `newclient/h5demo/strategies-h5-demo.html`

- [ ] **Step 1: 写顶部信息头与筛选区**

要求：

- 顶部保留 H5 客户端语气
- 设置股票 / 期货 / 事件切换或栏目提示
- 不做复杂浮层筛选

- [ ] **Step 2: 写推荐主卡**

内容要求：

- 直接给结论、主理由、当前动作
- 维持“国内金融 App 推荐主卡”感
- 不做纯封面视觉卡

- [ ] **Step 3: 写理由与风控卡**

至少包含：

- 为什么今天先看它
- 风险边界
- 止盈止损相关说明

- [ ] **Step 4: 写观点流**

要求：

- 做成单列卡片流
- 每条卡片都应先给结论，再给一句摘要
- 不像资讯列表

- [ ] **Step 5: 写期货与事件承接卡**

要求：

- 仍然明显属于策略页
- 只是次主线，不压过主推荐

- [ ] **Step 6: 写底部 Tab**

要求：

- 底部 Tab 必须与全站 spec 保持一致
- 固定为：首页 / 策略 / 关注 / 资讯 / 我的

- [ ] **Step 7: 运行静态校验**

Run: `rg -n "<title>|推荐主卡|风险边界|期货|事件|底部 Tab|策略" /Users/gjhan21/cursor/sercherai/newclient/h5demo/strategies-h5-demo.html`

Expected:

- 输出包含页面标题
- 输出包含“风险边界”
- 输出包含“期货”
- 输出包含“事件”
- 页面语义明确属于策略页

### Task 5: 输出统一入口页并完成本地预览校验

**Files:**
- Create: `newclient/h5demo/strategies-demo-index.html`

- [ ] **Step 1: 写入口说明**

包含：

- 本页 demo 的范围
- PC 与 H5 的角色区别
- 本页只使用真实策略能力的边界说明

- [ ] **Step 2: 写两个预览入口**

要求：

- 提供打开 `strategies-pc-demo.html` 的入口
- 提供打开 `strategies-h5-demo.html` 的入口
- 可加入 iframe 预览区

- [ ] **Step 3: 本地 HTTP 校验**

Run: `curl -s http://127.0.0.1:8365/strategies-demo-index.html | head -n 12`

Expected:

- 返回 HTML 文档头
- 标题与内容说明正常

- [ ] **Step 4: 页面关键字联动校验**

Run: `curl -s http://127.0.0.1:8365/strategies-pc-demo.html | rg -n "推荐工作台|风险边界|市场事件" && curl -s http://127.0.0.1:8365/strategies-h5-demo.html | rg -n "推荐主卡|风险边界|期货|事件"`

Expected:

- PC 页面关键模块文字存在
- H5 页面关键模块文字存在
- 证明文件已被静态服务正确读取

- [ ] **Step 5: 补充到后续总入口计划**

记录后续在 `newclient/h5demo/index.html` 中需要汇总本页入口，但本任务先不修改全站总入口。
