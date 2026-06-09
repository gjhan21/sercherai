# VIP 配额默认配置与数据订正设计说明书

## 背景与问题陈述
系统在先前的升级中，通过迁移脚本 [20260609_04_vip_quota_extensions.sql](file:///Users/gjhan21/cursor/sercherai/backend/migrations/20260609_04_vip_quota_extensions.sql) 为会员配额表 `vip_quota_configs` 新增了三个核心权限配额字段：
* `download_limit` (附件下载上限)
* `forecast_limit` (深度推演上限)
* `stock_reco_limit` (股票分析上限)

并在 `user_quota_usages` 中新增了对应的使用量记录字段。

然而，现有的数据库种子文件（`seed_demo.sql` 与 `seed_admin_extra.sql`）以及当前运行中的本地数据库，均未对上述三个新增的配额上限进行填充，默认值均为 `0`。
这导致用户即使拥有 VIP 等级，其下载附件、使用深度推演和查看股票分析时的配额也显示为 0/0，并且操作时会触发“配额超限 (quota exceeded)”的报错。

---

## 解决方案

本方案将对本地种子文件进行更新，并在当前运行的本地 MySQL 数据库中直接运行数据修正 SQL，确保所有的会员等级有合理、非零的初始默认值。

### 1. 会员配额默认值设定表
基于沟通，各等级的月度（`MONTHLY`）配额具体分配如下：

| 会员等级 | 文档阅读上限 | VIP 资讯阅读上限 | 附件下载上限 | 深度推演上限 | 股票分析上限 |
| :--- | :--- | :--- | :--- | :--- | :--- |
| **FREE** | 20 | 0 | 0 | 0 | 0 |
| **VIP1** | 200 | 100 | 20 | 10 | 30 |
| **VIP2** | 500 | 300 | 50 | 30 | 80 |
| **VIP3** | 1000 | 600 | 100 | 50 | 150 |
| **VIP4** | 2000 | 1200 | 200 | 100 | 300 |

---

## 修改细节

### 1. 种子文件修改

#### [MODIFY] [seed_demo.sql](file:///Users/gjhan21/cursor/sercherai/backend/scripts/seed_demo.sql)
在插入 `vip_quota_configs` 和 `user_quota_usages` 的语句中，补全新增的字段：
* 修改 `vqc_free`、`vqc_vip1`、`vqc_vip2` 的 SQL 插入行，增加 `download_limit`, `forecast_limit`, `stock_reco_limit` 对应数值。
* 修改 `uqu_demo_001` 的 SQL 插入行，补全 `download_used`, `forecast_used`, `stock_reco_used` 的默认值 `0`。

#### [MODIFY] [seed_admin_extra.sql](file:///Users/gjhan21/cursor/sercherai/backend/scripts/seed_admin_extra.sql)
* 修改 `vqc_vip3`、`vqc_vip2_weekly` 的 SQL 插入行，补全新增的三个配额上限。
* 修改 `uqu_demo_002`、`uqu_demo_003`、`uqu_demo_004`，补全新增的三个使用量字段为 `0`。

### 2. 运行数据库数据修正 (DB Patch)
连接本地 MySQL，执行以下 SQL 语句直接更新当前数据库的数据：
```sql
-- 订正 VIP1 配额
UPDATE vip_quota_configs 
SET download_limit = 20, forecast_limit = 10, stock_reco_limit = 30 
WHERE member_level = 'VIP1';

-- 订正 VIP2 配额
UPDATE vip_quota_configs 
SET download_limit = 50, forecast_limit = 30, stock_reco_limit = 80 
WHERE member_level = 'VIP2';

-- 订正 VIP3 配额
UPDATE vip_quota_configs 
SET download_limit = 100, forecast_limit = 50, stock_reco_limit = 150 
WHERE member_level = 'VIP3';

-- 订正 VIP4 配额
UPDATE vip_quota_configs 
SET download_limit = 200, forecast_limit = 100, stock_reco_limit = 300 
WHERE member_level = 'VIP4';

-- 订正 FREE 配额 (保持 0 即可)
UPDATE vip_quota_configs 
SET download_limit = 0, forecast_limit = 0, stock_reco_limit = 0 
WHERE member_level = 'FREE';
```

---

## 验证计划
1. **直接查询数据库**：执行 `SELECT * FROM vip_quota_configs;` 验证配额是否生效。
2. **测试页面数据加载**：启动客户端与服务端，访问个人使用记录页面（`/user/history`），确认配额条的上限不再是 0，且显示正确。
