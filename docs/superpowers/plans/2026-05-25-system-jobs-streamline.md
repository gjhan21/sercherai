# 系统任务中心轻量重组实现计划

1. 更新 `admin/src/lib/system-jobs-admin.js`
   - tab 只保留 `overview / market-data / runs`
   - 去掉首页动作卡
   - 收缩说明卡与总览卡辅助逻辑

2. 更新 `admin/src/views/SystemJobsView.vue`
   - 移除 hero 下动作卡
   - 精简 `overview` 区块
   - 精简 `market-data` 区块
   - 删除 `config` 与 `trigger` 主视图渲染

3. 更新测试
   - `admin/src/lib/system-jobs-admin.test.js`
   - `admin/src/views/system-jobs-view.test.js`

4. 验证
   - `node --test admin/src/lib/system-jobs-admin.test.js admin/src/views/system-jobs-view.test.js admin/src/views/system-jobs-sync-model.test.js admin/src/composables/useMarketSyncConsole.test.js admin/src/views/data-sources/data-sources-sync-section-view.test.js`
   - `cd admin && npm run build`
