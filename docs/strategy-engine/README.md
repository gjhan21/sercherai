# Strategy Engine Vibe Coding 开发总览

最后更新: 2026-03-21
状态: 阶段7已完成，阶段8进行中，阶段9已规划并完成文档收口（智能选股独立后台模块与第一轮蓝图已进入文档真相源）

## 文档目的

本目录用于沉淀“选股 + 期货投资策略独立服务”开发计划、阶段目标、验收标准与更新约定。

从现在开始，`strategy-engine` 相关开发统一采用 Vibe Coding 节奏：
- 先读文档，再开工
- 先做最小可感知价值，再扩展能力
- 每一阶段独立验收
- 阶段完成后立即回写文档
- 文档是当前阶段的真相源

## 开发前必读

每次开始开发前，按顺序阅读以下文档：
1. `docs/strategy-engine/README.md`
2. `docs/strategy-engine/开发工作流.md`
3. 当前正在执行的阶段文档
4. 与当前阶段相关的现有系统代码或接口

如果阶段已切换，必须先更新本 README 中的阶段状态，再进入编码。

## 当前项目判断

当前系统已经具备以下能力：
- 用户侧股票推荐、详情、表现、解释包
- 用户侧期货策略、套利机会、操作指导
- 管理侧股票推荐管理、期货策略管理、量化评估入口
- Go 后端已经承载了股票/期货相关业务接口，但策略逻辑仍与业务逻辑耦合

本次改造不重写整个业务后端，而是新增独立的 `strategy-engine` 服务，把策略计算、场景模拟、风险约束、报告生成从现有 Go 服务中拆出来。

## 北极星目标

- 把选股与期货策略做成独立可演进的策略服务
- 输出从“打分结果”升级为“可解释、可复盘、可追问”的策略结果
- 保持现有 Go 后端稳定，继续负责权限、发布、前后台接口和业务壳层

## 阶段总表

| 阶段 | 文档 | 目标 | 状态 |
| --- | --- | --- | --- |
| 阶段0 | `docs/strategy-engine/阶段0-服务边界与真相源.md` | 固化边界、目录、API、阶段规划 | Done |
| 阶段1 | `docs/strategy-engine/阶段1-服务骨架与任务协议.md` | 建立独立服务骨架、任务协议和基础可观测性 | Done |
| 阶段2 | `docs/strategy-engine/阶段2-股票MVP.md` | 跑通股票选股 MVP：输入、筛选、风险、报告 | Done |
| 阶段3 | `docs/strategy-engine/阶段3-期货MVP.md` | 跑通期货策略 MVP：方向、价位、风险、报告 | Done |
| 阶段4 | `docs/strategy-engine/阶段4-场景模拟与多代理.md` | 引入场景模拟、多代理协同与否决机制 | Done |
| 阶段5 | `docs/strategy-engine/阶段5-报告发布与复盘.md` | 完成发布适配、报告体系与历史复盘 | Done |
| 阶段6 | `docs/strategy-engine/阶段6-后台策略管理协同.md` | 让后台真正管理 seeds / agents / scenarios / publish policy | Done |
| 阶段7 | `docs/strategy-engine/阶段7-客户端策略解释展示.md` | 让客户端完整展示种子输入、模拟过程与结论解释 | Done |
| 阶段8 | `docs/strategy-engine/阶段8-真实数据接入与真相源增强.md` | 接入真实行情/资讯数据，并继续增强策略真相源与后台稳定摘要 | In Progress |
| 阶段9 | `docs/strategy-engine/阶段9-股票智能量化选股系统与后台独立模块.md` | 让 Python 自主自动选股，并在 Admin 中落地独立的智能选股模块 | Planned（主方案与第一轮蓝图已收口） |

## 当前推荐执行顺序

严格按以下顺序推进，避免范围漂移：
1. 阶段0
2. 阶段1
3. 阶段2
4. 阶段3
5. 阶段4
6. 阶段5
7. 阶段6
8. 阶段7
9. 阶段8
10. 阶段9

## 当前系统主要落点

现有系统对接点：
- `/Users/gjhan21/cursor/sercherai/backend/router/router.go`
- `/Users/gjhan21/cursor/sercherai/backend/internal/growth/model/models.go`
- `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/interfaces.go`
- `/Users/gjhan21/cursor/sercherai/backend/internal/growth/service/service.go`
- `/Users/gjhan21/cursor/sercherai/admin/src/views/MarketCenterView.vue`

新增服务目标目录：
- `/Users/gjhan21/cursor/sercherai/services/strategy-engine/`

阶段9 目标管理模块目录：
- `/Users/gjhan21/cursor/sercherai/admin/src/views/stock-selection/`

阶段9 第一轮执行蓝图：
- `/Users/gjhan21/cursor/sercherai/docs/strategy-engine/阶段9-第一轮-自动选股与后台骨架.md`

阶段9 第一轮开发启动清单：
- `/Users/gjhan21/cursor/sercherai/docs/strategy-engine/阶段9-第一轮-开发启动清单.md`

阶段9 文档导航：
- 主方案：`/Users/gjhan21/cursor/sercherai/docs/strategy-engine/阶段9-股票智能量化选股系统与后台独立模块.md`
- 第一轮蓝图：`/Users/gjhan21/cursor/sercherai/docs/strategy-engine/阶段9-第一轮-自动选股与后台骨架.md`
- 启动清单：`/Users/gjhan21/cursor/sercherai/docs/strategy-engine/阶段9-第一轮-开发启动清单.md`
- 开发开始前必须同时阅读这三份文档，不允许只看主方案就直接开工

## 当前落地情况

阶段1、阶段2、阶段3、阶段4、阶段5已经完成以下闭环：
- 独立 FastAPI 服务骨架已建立
- `GET /internal/v1/health` 已可用
- `POST /internal/v1/jobs/stock-selection` 已具备真实股票 MVP 处理逻辑
- `POST /internal/v1/jobs/futures-strategy` 已具备真实期货 MVP 处理逻辑
- `GET /internal/v1/jobs` 已支持按任务类型 / 状态分页查看最近作业
- `GET /internal/v1/jobs/{job_id}` 可返回股票 / 期货报告、候选清单和回写 payload
- `POST /internal/v1/publish/jobs/{job_id}` 已可生成 Markdown / HTML 发布报告与发布归档
- `GET /internal/v1/publish/history/{job_type}` 与 `POST /internal/v1/publish/compare` 已可提供版本对比与复盘基础能力
- 股票结果已可映射到现有 `stock_recommendations` / `stock_reco_details` 写入结构
- 期货结果已可映射到现有 `futures_strategies` / `futures_guidances` 写入结构
- 股票 / 期货报告已增加 `simulations`、`graph_summary`、`consensus_summary`
- 多代理面板已支持趋势、事件、流动性、风险、基差五类角色协同评审
- 风险 agent 已具备否决能力，可把高风险标的直接收敛为 `REJECT`
- Go 后端已具备调用 `strategy-engine` 生成股票推荐并回写数据库的 adapter，配置 `STRATEGY_ENGINE_BASE_URL` 后生效
- Go 后端已补齐 strategy-engine 作业列表 / 详情代理接口，Admin 可直接查询最近任务
- Go 后端已新增场景模板 / 发布策略管理接口，并可在自动发布与人工发布时消费默认策略
- Admin 管理端已支持查看股票 / 期货 publish record 的 HTML / Markdown 报告正文、payload 与报告快照
- Admin 管理端“引擎配置”页已补上作业中心，可查看输入快照、结果摘要、警告与产物 JSON
- Admin 管理端已新增场景模板管理、发布策略管理，以及作业中心“按策略发布 / 人工覆盖发布”入口
- strategy-engine 已支持运行时消费后台场景模板，并在发布阶段执行风险等级 / 警告数 / veto 拦截规则
- Go 后端创建 strategy-engine job 时已开始透传 `config_refs` 与 `publish_policy_preview`，后台能看到一次任务具体消费了哪套默认配置
- Admin 管理端作业中心详情已升级为结构化标签页，可直接查看配置来源、角色评审、场景输出，以及 `READY / BLOCKED / UNKNOWN` 发布就绪状态
- Go 后端已新增 `strategy_job_runs` / `strategy_job_artifacts` / `strategy_job_replays` 三类运行时快照真相源，并在股票 / 期货生成成功后立即归档 payload、report、warnings 与发布动作
- Admin 作业列表 / 详情已切到“本地快照优先、远端回源补归档”模式，详情页可直接区分 `已归档快照`、`远端回源已归档` 与 `远端临时结果`
- Stock / Futures 发布复盘弹窗已开始展示本地 replay 审计信息，包括操作人、发布时间、覆盖原因、发布策略快照与数据来源，人工覆盖链路不再只剩 warning JSON
- Admin 作业中心列表与概览已开始展示发布次数、最近发布动作和完整发布审计时间线，后台现在能直接把“这次任务怎么生成、怎么放行、谁放行”串成一条链
- 客户端策略中心详情页已开始展示种子输入、角色评审、场景推演、风险提醒与失效条件
- 用户侧股票 / 期货 insight 接口已开始补充 explanation 结构，支撑“为什么选它”的解释展示
- 客户端首页主推荐 / 观察清单以及策略中心列表卡片，已开始展示 explanation 缩略信息，能在列表层承接“为什么选它”
- 客户端历史归档页已开始展示 explanation 摘要、关键种子、风险与失效条件，让历史页也能回答“当时为什么选”
- 客户端我的关注页已开始展示“继续跟踪原因 / 角色结论 / 风险边界变化”，让持续回访更有解释力
- 客户端会员页已开始展示“游客 / 注册 / 会员”的策略解释能力差异，把升级理由落到具体能力上
- 客户端关注页与会员页已开始动态联动，能表达“游客 -> 注册 -> 会员”在持续跟踪能力上的差异
- 客户端会员页已开始展示最近一次策略 explanation 快照，把升级价值和真实运行数据挂钩
- 客户端关注页已开始展示“加入关注时 vs 当前”的变化对比，能解释结论、风险边界和角色判断怎么变了
- 客户端关注页已开始展示“加入关注 -> 最近刷新 -> 当前判断”的变化时间线
- 客户端历史归档页和股票策略详情页已开始展示“版本差异”，能解释归档版本与当前 explanation 版本为什么不同
- 客户端期货策略详情页已开始展示“版本差异”，股票 / 期货详情页的解释表达已基本对齐
- 客户端关注页已开始沉淀“多次变化记录”，不再只对比基线和当前，而是连续记录每次刷新抓到的结论 / 风险 / 角色变化
- 客户端关注页的变化历史已开始挂接后端真实发布批次，可展示批次版本、交易日和任务标识，而不是只靠本地内容推断
- 客户端归档页 / 关注页 / 策略详情页已开始共用同一套版本历史 helper，批次版本、交易日和差异文案不再各自维护
- 客户端归档页与股票 / 期货策略详情页已开始显式展示“来自哪次生成”，让用户能直接看到批次、交易日和任务追踪信息
- 后端已补充股票 / 期货 `version-history` 接口，策略详情页已开始读取真实版本历史列表，不再只靠当前 explanation 回推
- 客户端策略详情页已开始支持“点选历史版本 vs 当前解释”对比，版本演进不再只是静态列表
- 客户端归档页已开始读取股票 `version-history` 接口，并展示后端版本轨迹摘要，历史档案不再只靠当前 explanation 推断
- 客户端关注页已开始读取股票 `version-history` 接口，并补上“版本对照 / 后端版本轨迹”模块，关注页的版本解释路径开始和详情页、归档页统一
- 客户端归档页已开始支持点击历史版本并和当前 explanation 即时对比，历史档案页的版本体验开始和策略详情页对齐
- 客户端 `strategy-version` helper 已开始统一承接详情页、归档页、关注页的版本轨迹映射与 fallback 逻辑，阶段7的多页面版本解释开始收敛到同一套前端真相源
- 客户端 `strategy-version` helper 已进一步承接 compare state，详情页与归档页的 selected/matched 版本逻辑开始从页面内实现收回共享层
- 客户端关注页的版本对照也已切到同一套 compare state helper，阶段7里详情页 / 归档页 / 关注页的版本交互路径开始真正统一
- 客户端策略详情页已补上“回到默认对比”等统一交互文案，阶段7的多页面版本体验开始从逻辑统一走向交互统一
- 客户端三类页面的版本提示语也已开始统一，阶段7的多页面版本体验正在从实现层统一推进到认知层统一
- 客户端三类页面的 CTA 与空态文案也已开始统一，阶段7 的前端承接链路正在逐步收口
- 后端 explanation / `version-history` 查询已开始切到“本地快照优先、远端回源即补归档”，客户端解释链路不再默认依赖 strategy-engine 运行时内存
- `strategy_job_replays` 已开始归档 `publish_version`，用户侧 explanation 与版本轨迹在本地快照里也能稳定拿到批次版本号
- Admin 侧 publish history / publish record / replay / compare 也已开始切到本地快照优先，后台复盘链路对远端 runtime store 的依赖继续下降
- Admin 侧 publish record / replay / compare 已进一步收口到“publish record 回填 + 本地比较 / 本地 replay 读取”，不再默认直接依赖远端 compare / replay 接口
- Admin 侧 publish history 列表摘要也已开始优先走“回填 publish record -> 本地摘要 / 回填摘要”，后台发布列表对远端 history 直出的依赖继续下降
- explanation 远端补归档也已开始优先复用本地 job snapshot，减少为补 explanation 额外回源远端 job detail 的次数
- explanation / `version-history` 仓库级测试已补到“本地直读 + 远端 publish record 回填 + 本地 job snapshot 兜底”，阶段7 的客户端解释真相源开始具备自动化回归保护
- Admin 作业中心的 job artifact snapshot 已开始归档 `payload_echo`，本地 job snapshot 对远端 runtime-only 字段的承接继续增强
- Admin 作业中心常用的 `trade_date / result_summary / selected_count / payload_count / warning_count` 已开始作为顶层 job summary 返回，前端列表与概览对嵌套 payload/result 的依赖继续下降
- 真实上下文主链已开始替换服务内置样本：股票与期货都已优先走 Go 后端内部上下文 API，服务内置种子已收口为显式开发 / 测试 fallback
- 行情接入方案已正式落地为“Go 后端对接第三方源并同步到本地标准表，`strategy-engine` 只消费统一归一后的内部行情数据，不直接绑定外部 API”
- 阶段8 第一刀已开始落地：Go 后端已新增 `POST /internal/v1/strategy-engine/context/stock-selection`，直接复用 `market_daily_bar_truth` + `market_news_items` 输出标准化股票种子上下文
- `strategy-engine` 的股票种子加载已默认切到 Go 内部上下文 API，`DEFAULT_MARKET_SEEDS` 已收口为显式开发 / 测试 fallback
- 股票上下文已开始携带 `selected_trade_date / price_source / news_window_days / warnings`，报告 artifacts 已开始保留 `context_meta`
- 用户侧 explanation 的 `seed_summary` 已开始带出真实行情日、行情源和资讯窗口；Admin 作业中心概览也已开始显示这组真实数据上下文
- 阶段8 第二刀已开始落地：Go 后端已新增 `POST /internal/v1/strategy-engine/context/futures-strategy`，直接复用 `market_daily_bar_truth` + `market_news_items` 输出标准化期货种子上下文
- `strategy-engine` 的期货种子加载已默认切到 Go 内部上下文 API，`DEFAULT_FUTURES_SEEDS` 已收口为显式开发 / 测试 fallback
- 期货 report 也已开始携带 `context_meta`，所以客户端 explanation 与 Admin 作业中心可以沿用同一套“真实交易日 / 行情源 / 资讯窗口”摘要
- 阶段8 联调收口已开始补齐：当期货任务未显式传 `contracts` 时，Go 后端上下文构建已默认优先真实来源 truth bar，不再优先落到 `MOCK` 最新样本日期；只有真实来源缺失时才会 fallback 到 `MOCK` 并返回 warning
- 本地联调已验证股票 / 期货真实上下文主链可跑通：股票日推荐已带出真实 `selected_trade_date / price_source / news_window_days`，期货上下文与期货作业 report 的 `context_meta.price_source` 也已验证不再默认回到 `MOCK`
- 阶段8 期货因子增强已开始落地：Go 后端会在 futures 上下文构建时回查 `market_daily_bars` 原始多源条目，补齐 truth bar 缺失的 `settle / prev_settle / turnover / open_interest`，让 `basis / carry / flow / regime` 不再只吃单条 truth 快照的弱字段
- 期货 `flow_bias` 与 `regime` 已开始引入 turnover confirmation，轻量因子仍维持 MVP 级别，但已经从“只看价格 + 持仓”升级为“价格 + 成交量 + 成交额 + 持仓”的真实数据组合
- 显式指定期货 `contracts` 的上下文选日也已开始优先真实来源：现在即使只传单个 contract，也会先尝试非 `MOCK` truth bar；只有真实来源完全不存在时才回退 `MOCK`
- 本地联调已确认显式 `IF2606` 不再默认命中 `2026-03-19 / MOCK`，而是先落到 `2026-03-18 / TUSHARE`；当前若真实历史不足 14 个交易日，会直接返回明确错误，而不是静默切回 `MOCK`
- 期货显式 contract 已补上“受控 fallback”开关：当真实历史不足 14 个交易日时，只有显式传入 `allow_mock_fallback_on_short_history=true` 才会退回 `MOCK`，并把 warning 带进 `context_meta`
- Go 后端 futures 每日生成链路已开始默认携带这组受控 fallback 开关，所以当前管理侧生成失败原因已经从“上下文拿不到”收口到“发布策略拦截”，上下文问题不再是主阻塞
- futures 默认 publish policy 也已补上内置回退：当后台没有持久化默认策略时，Go 后端会自动回退到 `FUTURES` 专用默认策略（`HIGH` 风险上限、允许 vetoed publish、warning 阈值 5）
- 本地联调已确认 `POST /api/v1/admin/futures/strategies/generate-daily?trade_date=2026-03-19` 现在可以成功返回 `200` 并完成发布入库；当前主链已经从“生成失败”推进到“可生成、可复盘、可继续调优策略门槛”
- 阶段8 下一刀已继续落地为“期货结构因子增强”：Go 后端开始基于同交易日多合约 truth bar 计算 `term_structure_pct`，并把 `turnover_ratio / term_structure_pct` 一起透传到 `strategy-engine`
- `strategy-engine` 的期货特征工厂、basis agent 与理由摘要已开始消费这组新增结构因子，前后台 explanation / 作业详情即使不新增页面，也能在现有理由文案里看到“成交额放大 / 近远月斜率 / 跨期结构确认”的真实来源
- 阶段8 当前最新切片已继续把期货结构因子扩到“全曲线 + 跨期价差”：Go 后端开始复用 `arbitrage_recos` 与同品种全曲线 truth bar，新增 `curve_slope_pct / spread_pressure / spread_percentile / spread_pair`
- `strategy-engine` 的 futures seed loader、feature factory、basis agent 与理由摘要已开始消费这组字段，现有 explanation / 作业详情会自然出现“全曲线斜率”“关联价差分位”“价差回归受益腿/压力腿”等更强结构文案
- 阶段8 当前最新一刀已继续把仓单/库存接入主链：Go 后端新增 `futures_inventory_snapshots` 与 `/api/v1/admin/futures/inventory/sync`，futures context 新增 `inventory_level / inventory_change_pct / inventory_pressure`
- `strategy-engine` 的 futures seed loader、feature factory、basis agent 与理由摘要已开始消费这组仓单字段；Admin `MarketCenter` 也已补上“期货仓单同步”区块
- 阶段8 当前最新切片已继续把仓单主链细化到“仓库 / 区域 / 品牌”结构：`futures_inventory_snapshots` 已新增 `brand / place / grade`，futures context 新增 `inventory_focus_*` 与 `inventory_*_share`
- `strategy-engine` 的 futures explanation / 多角色评审已开始自然带出“区域占比 / 仓库占比 / 品牌占比”文案，同一份仓单真相源开始同时服务上下文打分与解释展示
- 阶段8 当前最新切片已继续把 `place / grade` 推进到 futures explanation / agent 评审主链：现有理由文案和 basis 角色结论已开始自然带出“产地占比 / 等级占比”
- FUTURES 默认 publish policy 也已从“运行时代码内置回退”继续收口到“后台可见配置”：MySQL 仓储会在缺失默认策略时自动把 `policy_default_futures` / `policy_default_all` 落入 `system_configs`，Admin 配置页可以直接看到、编辑并通过 `updated_by / updated_at` 追踪
- 期货短历史受控 fallback 开关也已进入后台配置中心：Go 后端会物化 `agent_default_futures`，并通过 agent profile 的 `allow_mock_fallback_on_short_history` 控制 futures job 是否允许在真实历史不足时回退 `MOCK`
- Admin 管理端“多角色配置”已开始展示和编辑这组短历史回退开关，期货每日生成不再依赖 `strategy_engine_client.go` 里的硬编码 `true`
- futures 每日生成 payload 中原先“先写死 `true` 再覆盖”的冗余赋值也已移除，短历史回退开关现在只以后台 agent profile 为唯一真相源
- 当前五个阶段的最小闭环已全部打通，后续迭代以数据源增强、持久化和后台消费接入为主

## 后续专题

阶段7已经完成，后续重点已经正式收敛为阶段8：
- 真实行情与资讯数据接入
- 策略真相源与后台稳定摘要继续增强

阶段8详见：
- `/Users/gjhan21/cursor/sercherai/docs/strategy-engine/阶段8-真实数据接入与真相源增强.md`

阶段9已完成文档收口，当前默认口径为：
- `智能选股` 作为 Admin 独立一级模块建设
- 第一轮只做一个默认 profile 和 4 个后台核心页面
- `AUTO` 模式不依赖人工 `seed_symbols`
- 发布继续复用现有 `strategy-engine` 发布链路
- 完整图谱平台、复杂回测平台和重型多世界控制台不进入第一轮

## 统一验收口径

每个阶段完成时，至少满足以下要求：
- 有一个可感知的开发闭环，而不是只堆概念
- 输入、处理、输出边界清晰
- 文档、目录、接口、任务状态保持一致
- 涉及后端时至少能通过对应测试或启动验证
- 阶段文档中的“状态 / 已完成 / 偏差 / 默认约定或遗留问题 / 下阶段条件”已更新

## 更新约定

每完成一个阶段，必须更新：
- `docs/strategy-engine/README.md` 中对应阶段状态
- 当前阶段文档中的完成情况、风险和默认约定或遗留问题
- 如果范围变化，补充到 `docs/strategy-engine/开发工作流.md`

如果是 Planned 阶段先做文档收口，则必须同步更新：
- 主方案文档中的默认决策、兼容策略、验收矩阵
- 第一轮执行蓝图中的对象、接口、页面状态流与回归清单
