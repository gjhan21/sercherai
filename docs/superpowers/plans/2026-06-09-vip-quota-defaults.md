# VIP 配额默认数据设置实施计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 为本地运行中的 MySQL 数据库及种子脚本文件填充新增的下载配额、深度推演配额、股票分析配额默认值，以避免默认配额限制为 0 导致的使用记录报错及页面展示缺陷。

**Architecture:** 物理修改 SQL 种子文件（`seed_demo.sql` 和 `seed_admin_extra.sql`），并且通过终端直接在本地 MySQL 服务上执行数据更新脚本对当前数据进行修正。

**Tech Stack:** MySQL / Shell

---

### Task 1: 修改后端 SQL 种子数据文件

**Files:**
- Modify: `backend/scripts/seed_demo.sql`
- Modify: `backend/scripts/seed_admin_extra.sql`

- [ ] **Step 1: 修改 backend/scripts/seed_demo.sql 中的 vip_quota_configs 配置**
  将 `vip_quota_configs` 的插入和更新逻辑中补齐 `download_limit`, `forecast_limit`, `stock_reco_limit`。
  
  修改后的 SQL 区块应该为：
  ```sql
  INSERT INTO vip_quota_configs (id, member_level, doc_read_limit, news_subscribe_limit, download_limit, forecast_limit, stock_reco_limit, reset_cycle, status, effective_at, updated_at)
  VALUES ('vqc_vip1', 'VIP1', 200, 100, 20, 10, 30, 'MONTHLY', 'ACTIVE', NOW(), NOW())
  ON DUPLICATE KEY UPDATE doc_read_limit = VALUES(doc_read_limit), news_subscribe_limit = VALUES(news_subscribe_limit), download_limit = VALUES(download_limit), forecast_limit = VALUES(forecast_limit), stock_reco_limit = VALUES(stock_reco_limit), updated_at = VALUES(updated_at);

  INSERT INTO vip_quota_configs (id, member_level, doc_read_limit, news_subscribe_limit, download_limit, forecast_limit, stock_reco_limit, reset_cycle, status, effective_at, updated_at)
  VALUES ('vqc_free', 'FREE', 20, 10, 0, 0, 0, 'MONTHLY', 'ACTIVE', NOW(), NOW())
  ON DUPLICATE KEY UPDATE doc_read_limit = VALUES(doc_read_limit), news_subscribe_limit = VALUES(news_subscribe_limit), download_limit = VALUES(download_limit), forecast_limit = VALUES(forecast_limit), stock_reco_limit = VALUES(stock_reco_limit), updated_at = VALUES(updated_at);

  INSERT INTO vip_quota_configs (id, member_level, doc_read_limit, news_subscribe_limit, download_limit, forecast_limit, stock_reco_limit, reset_cycle, status, effective_at, updated_at)
  VALUES ('vqc_vip2', 'VIP2', 500, 300, 50, 30, 80, 'MONTHLY', 'ACTIVE', NOW(), NOW())
  ON DUPLICATE KEY UPDATE doc_read_limit = VALUES(doc_read_limit), news_subscribe_limit = VALUES(news_subscribe_limit), download_limit = VALUES(download_limit), forecast_limit = VALUES(forecast_limit), stock_reco_limit = VALUES(stock_reco_limit), updated_at = VALUES(updated_at);
  ```

- [ ] **Step 2: 修改 backend/scripts/seed_demo.sql 中的 user_quota_usages 默认已使用项**
  修改后的 SQL 区块应该为：
  ```sql
  INSERT INTO user_quota_usages (id, user_id, member_level, period_key, doc_read_used, news_subscribe_used, download_used, forecast_used, stock_reco_used, updated_at)
  VALUES ('uqu_demo_001', 'u_demo_001', 'VIP1', DATE_FORMAT(NOW(), '%Y-%m'), 13, 7, 0, 0, 0, NOW())
  ON DUPLICATE KEY UPDATE doc_read_used = VALUES(doc_read_used), news_subscribe_used = VALUES(news_subscribe_used), download_used = VALUES(download_used), forecast_used = VALUES(forecast_used), stock_reco_used = VALUES(stock_reco_used), updated_at = VALUES(updated_at);
  ```

- [ ] **Step 3: 修改 backend/scripts/seed_admin_extra.sql 中的 vip_quota_configs 配置**
  修改后的 SQL 区块应该为：
  ```sql
  INSERT INTO vip_quota_configs (id, member_level, doc_read_limit, news_subscribe_limit, download_limit, forecast_limit, stock_reco_limit, reset_cycle, status, effective_at, updated_at)
  VALUES
    ('vqc_vip3', 'VIP3', 1000, 600, 100, 50, 150, 'MONTHLY', 'ACTIVE', NOW(), NOW()),
    ('vqc_vip2_weekly', 'VIP2', 160, 80, 50, 30, 80, 'WEEKLY', 'DISABLED', DATE_SUB(NOW(), INTERVAL 30 DAY), NOW())
  ON DUPLICATE KEY UPDATE
    doc_read_limit = VALUES(doc_read_limit),
    news_subscribe_limit = VALUES(news_subscribe_limit),
    download_limit = VALUES(download_limit),
    forecast_limit = VALUES(forecast_limit),
    stock_reco_limit = VALUES(stock_reco_limit),
    reset_cycle = VALUES(reset_cycle),
    status = VALUES(status),
    updated_at = VALUES(updated_at);
  ```

- [ ] **Step 4: 修改 backend/scripts/seed_admin_extra.sql 中的 user_quota_usages 配置**
  修改后的 SQL 区块应该为：
  ```sql
  INSERT INTO user_quota_usages (id, user_id, member_level, period_key, doc_read_used, news_subscribe_used, download_used, forecast_used, stock_reco_used, updated_at)
  VALUES
    ('uqu_demo_002', 'u_demo_002', 'FREE', DATE_FORMAT(NOW(), '%Y-%m'), 7, 3, 0, 0, 0, NOW()),
    ('uqu_demo_003', 'u_demo_003', 'VIP2', DATE_FORMAT(NOW(), '%Y-%m'), 66, 28, 0, 0, 0, NOW()),
    ('uqu_demo_004', 'u_demo_004', 'VIP2', DATE_FORMAT(NOW(), '%Y-%m'), 31, 14, 0, 0, 0, NOW())
  ON DUPLICATE KEY UPDATE
    member_level = VALUES(member_level),
    doc_read_used = VALUES(doc_read_used),
    news_subscribe_used = VALUES(news_subscribe_used),
    download_used = VALUES(download_used),
    forecast_used = VALUES(forecast_used),
    stock_reco_used = VALUES(stock_reco_used),
    updated_at = VALUES(updated_at);
  ```

- [ ] **Step 5: 提交更改**
  ```bash
  git add backend/scripts/seed_demo.sql backend/scripts/seed_admin_extra.sql
  git commit -m "db: update vip quota default limits in sql seed files"
  ```

---

### Task 2: 执行运行数据库数据修正 (DB Patch)

- [ ] **Step 1: 对当前运行的本地 MySQL 数据库执行更新**
  运行命令：
  ```bash
  mysql -h 127.0.0.1 -P 3306 -u root -pabc123 -D sercherai -e "
  UPDATE vip_quota_configs SET download_limit = 20, forecast_limit = 10, stock_reco_limit = 30 WHERE member_level = 'VIP1';
  UPDATE vip_quota_configs SET download_limit = 50, forecast_limit = 30, stock_reco_limit = 80 WHERE member_level = 'VIP2';
  UPDATE vip_quota_configs SET download_limit = 100, forecast_limit = 50, stock_reco_limit = 150 WHERE member_level = 'VIP3';
  UPDATE vip_quota_configs SET download_limit = 200, forecast_limit = 100, stock_reco_limit = 300 WHERE member_level = 'VIP4';
  UPDATE vip_quota_configs SET download_limit = 0, forecast_limit = 0, stock_reco_limit = 0 WHERE member_level = 'FREE';
  "
  ```

- [ ] **Step 2: 验证数据更正结果**
  运行命令：
  ```bash
  mysql -h 127.0.0.1 -P 3306 -u root -pabc123 -D sercherai -e "SELECT id, member_level, doc_read_limit, news_subscribe_limit, download_limit, forecast_limit, stock_reco_limit FROM vip_quota_configs;"
  ```
  预期输出：所有行的 `download_limit`、`forecast_limit`、`stock_reco_limit` 都已更改为正确值。

---

### Task 3: 启动服务并端到端验证

- [ ] **Step 1: 启动服务端和客户端开发环境**
  在 `backend` 目录下通过 `make run` 或后台进程启动 Go API 服务。
  在 `newclient` 目录下执行 `npm run dev` 启动前端。

- [ ] **Step 2: 登录网页进行手动校验**
  访问网页的个人历史使用中心（`http://127.0.0.1:5275/user/history`），确认配额条的上限不再是 `0/0`，而是我们推荐的数值（如 VIP1 用户附件下载显示 `/ 20`，深度推演显示 `/ 10` 等）。
