# 今日推荐回落设计

- 目标：当 /recommendations 指定今日 trade_date 但今日无推荐时，自动回落到最近一个有推荐的交易日。
- 范围：仅修改 ListStockRecommendations 后端查询回落逻辑，不改前端调用协议。
- 测试：补 repo 测试覆盖 今日命中 / 指定日回落 / 历史为空。
