# 推荐股票网站 Vibe Coding 开发总览

最后更新: 2026-03-18
状态: 阶段6已完成

## 文档目的

本目录用于沉淀“推荐股票吸引客户的网站”开发计划、阶段目标、验收标准与更新约定。

从现在开始，相关开发统一采用 Vibe Coding 节奏：
- 先读文档，再开工
- 先做最小可感知价值，再扩展功能
- 每一阶段独立验收
- 阶段完成后立即回写文档
- 文档是当前阶段的真相源

## 开发前必读

每次开始开发前，按顺序阅读以下文档：
1. `docs/vibe-stock-growth/README.md`
2. `docs/vibe-stock-growth/开发工作流.md`
3. 当前正在执行的阶段文档
4. 与当前阶段相关的现有系统文档或页面实现

如果阶段已切换，必须先更新本 README 中的阶段状态，再进入编码。

## 当前项目判断

当前系统已经具备以下基础能力：
- 用户端股票推荐列表、详情、表现、解释包、相关新闻
- 后台推荐管理、每日生成、量化榜单、评估视图
- 会员状态与部分访问控制

本次改造不从零重做，而是把现有功能重组为“每日投资决策入口”，提升回访率、信任感与转化率。

## 北极星目标

- 用户每天有理由访问网站
- 用户逐步信任推荐逻辑与跟踪过程
- 用户愿意注册、关注、自选并进入会员转化漏斗

## 阶段总表

| 阶段 | 文档 | 目标 | 状态 |
| --- | --- | --- | --- |
| 阶段0 | `docs/vibe-stock-growth/阶段0-画布与真相源.md` | 统一用户路径、页面画布、状态流 | Done |
| 阶段1 | `docs/vibe-stock-growth/阶段1-今日决策首页.md` | 首页改造成“今天该看什么” | Done |
| 阶段2 | `docs/vibe-stock-growth/阶段2-推荐档案页.md` | 推荐页改造成可解释、可跟踪档案 | Done |
| 阶段3 | `docs/vibe-stock-growth/阶段3-历史档案与信任改造.md` | 建立历史档案与可信度表达 | Done |
| 阶段4 | `docs/vibe-stock-growth/阶段4-我的关注与回访机制.md` | 建立个人关注与回访理由 | Done |
| 阶段5 | `docs/vibe-stock-growth/阶段5-后台生命周期管理.md` | 后台从发推荐升级为管推荐 | Done |
| 阶段6 | `docs/vibe-stock-growth/阶段6-会员转化与内容节奏.md` | 完成权益分层与运营节奏 | Done |

## 当前推荐执行顺序

严格按以下顺序推进，避免范围漂移：
1. 阶段0
2. 阶段1
3. 阶段2
4. 阶段3
5. 阶段4
6. 阶段5
7. 阶段6

## 当前系统主要落点

前台主要改造点：
- `/Users/gjhan21/cursor/sercherai/client/src/views/HomeView.vue`
- `/Users/gjhan21/cursor/sercherai/client/src/views/StrategyView.vue`
- 新增历史档案页、我的关注页与相关路由

后台主要改造点：
- `/Users/gjhan21/cursor/sercherai/admin/src/views/MarketCenterView.vue`
- `/Users/gjhan21/cursor/sercherai/backend/internal/growth/handler/admin_growth_handler.go`
- `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/mysql_repo.go`
- `/Users/gjhan21/cursor/sercherai/backend/migrations/20260225_00_core_schema.sql`

## 统一验收口径

每个阶段完成时，至少满足以下要求：
- 页面或流程上有用户可感知变化
- 相关入口、内容、动作、反馈形成闭环
- 涉及前端时构建通过
- 涉及后台时接口或数据结构同步说明清楚
- 阶段文档中的“状态/完成情况/偏差/下阶段输入”已更新

## 更新约定

每完成一个阶段，必须更新：
- `docs/vibe-stock-growth/README.md` 中对应阶段状态
- 当前阶段文档中的完成情况、风险、遗留问题
- 如果范围变化，补充到 `docs/vibe-stock-growth/开发工作流.md`
