# 专题B：股票事件图谱与资讯语义工作台

最后更新: 2026-03-23  
状态: In Progress（第一轮闭环已完成）

## 1. 专题目标

把当前股票研究从“run 级图快照 + 浅层 related entities”升级为“事件真相源 + 图谱承接 + 审核工作台”的研究底座。

专题B需要解决的核心问题：

- 当前新闻和图谱更多服务单次 run，缺少稳定事件真相源
- 资讯语义、行业主题、政策财报等关系没有统一标准化 schema
- 图谱可查 run snapshot，但不能稳定表达 reviewed event edges
- Admin 还没有事件审核台，误判、聚类错误和主题归属无法治理
- 股票 explanation 缺少“为什么这条新闻影响这个股票/主题”的证据链

## 2. 当前基线

### 2.1 已经完成的能力

- Go 后端已具备资讯同步、多源 market news、股票 explanation、graph summary 与图服务代理
- `strategy-graph` 已支持 run 图快照写入、按实体查询一跳/两跳子图
- 股票候选详情已可展示基础 explanation 和 evidence card，但更多基于当次 run 结果
- Admin 已具备股票研究后台模块，但还没有面向事件真相源的治理工作台

### 2.2 当前基线的主要落点

- 资讯同步与真相源：`backend/internal/growth/repo/market_data_multi_source.go`
- 股票研究上下文与 explanation：`backend/internal/growth/repo/strategy_engine_context_repo.go`
- 图谱代理与图快照：`backend/internal/growth/repo/strategy_graph_*.go`
- 股票后台页：`admin/src/views/stock-*` 相关研究页面
- Client explanation helper：`client/src/lib/strategy-version.js`

## 3. 专题完成标准

专题B标记为 `Done` 时，至少满足：

- 新闻、公告、政策、财报、行业主题等可沉淀为统一事件 schema
- 图谱既能承接 run snapshot，也能承接 reviewed event truth source
- Admin 能审核事件聚类、标签和主题/行业归属
- 股票候选 explanation 可以稳定展示事件证据链
- 事件审核运营结果能回写真相源，而不是只影响展示层

## 4. 非目标

本专题不并入：

- 高频舆情情绪量化生产化
- 面向 C 端开放图谱探索社区
- 全自动无人工审核的事件生产链
- 与期货主数据画像直接混做成一个专题

## 5. 推荐第一轮范围

第一轮只做“事件能标准化、能审核、能被 explanation 使用”的最小闭环：

1. 事件 schema 与标准实体类型
2. 资讯抽取 + 事件聚类 + symbol/topic 映射
3. 图谱承接 reviewed event edges
4. Admin 事件审核台第一版
5. 股票 explanation 增加事件证据卡

## 6. 第一轮实施清单

第一轮已完成以下闭环：

- 事件真相源最小 schema / model / repo 已落地，新闻同步后可生成聚类后的待审核事件
- reviewed stock events 已可经审核流进入图服务，并兼容现有 `subgraph` 查询
- Admin 已新增股票事件审核台，并与 `ReviewCenterView` 的 `STOCK_EVENT` 审核任务联动
- 股票 explanation 已能注入 reviewed events 的证据卡，客户端已展示增强后的事件说明

后续增量仍请直接按：

- `/Users/gjhan21/cursor/sercherai/docs/strategy-engine/专题B-第一轮-开发启动清单.md`

推进。
