# 专题C：期货深度因子与主数据画像

最后更新: 2026-03-23  
状态: Planned

## 1. 专题目标

把当前“可运行的期货策略研究链”继续升级为“主数据画像完整、深层因子稳定、解释更像研究报告”的期货研究底座。

专题C需要解决的核心问题：

- place / grade / brand 等结构字段已经进入主链，但仍偏上下文摘要，还不是完整主数据画像
- 期货因子仍以趋势、结构、仓单和价差为主，缺少更稳定的深层研究因子体系
- 商品链、替代关系、上下游成本传导还没有稳定接入 explanation
- 期货候选详情能展示 explanation，但证据层次还不够深

## 2. 当前基线

- Go 后端已具备 futures context、term structure、curve slope、spread、inventory 与结构维度
- strategy-engine 已能消费 place / grade / warehouse / brand 等仓单结构字段
- Admin 与 Client 已能展示期货 explanation、风险与 evidence card
- 图谱仍主要偏 run snapshot，不足以表达完整商品链和政策传导路径

## 3. 专题完成标准

专题C标记为 `Done` 时，至少满足：

- futures instrument master v2 可以稳定表达商品、交割地、仓库、品牌、等级、库存口径与合约链
- 深层因子可稳定使用库存变化、结构偏移、品牌/等级溢价、跨品种联动等指标
- 期货 explanation 可展示更深的因子证据和商品链解释
- 图谱能够表达商品链、替代关系、上下游成本传导与政策冲击路径

## 4. 非目标

- 复杂期权波动率曲面
- 高频 CTA 或实盘执行平台
- 跨交易所超高频研究链

## 5. 推荐第一轮范围

第一轮只做“主数据画像 + 深层因子 + explanation 增强”的最小闭环：

1. futures instrument master v2
2. 库存 / 仓单信号从解释层推进到因子层
3. 深化 basis / spread / term structure 因子
4. 商品链与政策冲击的基础表达
5. 期货候选详情增强

## 6. 第一轮实施清单

正式开工请直接按：

- `/Users/gjhan21/cursor/sercherai/docs/strategy-engine/专题C-第一轮-开发启动清单.md`

推进。
