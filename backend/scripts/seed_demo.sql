INSERT INTO users (id, phone, email, password_hash, status, kyc_status, member_level, created_at, updated_at)
VALUES ('u_demo_001', '13800000001', 'demo@sercherai.local', 'a03c32fcd351cba2d9738622b083bed022ef07793bd92b59faea0207653f371d', 'ACTIVE', 'APPROVED', 'VIP1', NOW(), NOW())
ON DUPLICATE KEY UPDATE updated_at = VALUES(updated_at), member_level = VALUES(member_level);

INSERT INTO users (id, phone, email, password_hash, status, kyc_status, member_level, created_at, updated_at)
VALUES ('admin_001', '13800000000', 'admin@sercherai.local', 'a03c32fcd351cba2d9738622b083bed022ef07793bd92b59faea0207653f371d', 'ACTIVE', 'APPROVED', 'VIP1', NOW(), NOW())
ON DUPLICATE KEY UPDATE updated_at = VALUES(updated_at), member_level = VALUES(member_level);

INSERT INTO membership_products (id, name, price, status, member_level, duration_days)
VALUES ('mp_demo_001', 'VIP月卡', 99.00, 'ACTIVE', 'VIP1', 30)
ON DUPLICATE KEY UPDATE price = VALUES(price), status = VALUES(status), member_level = VALUES(member_level), duration_days = VALUES(duration_days);

INSERT INTO membership_orders (id, order_no, user_id, product_id, amount, pay_channel, status, paid_at, created_at, updated_at)
VALUES ('mo_demo_001', 'mo_demo_001', 'u_demo_001', 'mp_demo_001', 99.00, 'ALIPAY', 'PAID', NOW(), NOW(), NOW())
ON DUPLICATE KEY UPDATE status = VALUES(status), paid_at = VALUES(paid_at), amount = VALUES(amount), updated_at = VALUES(updated_at);

INSERT INTO vip_quota_configs (id, member_level, doc_read_limit, news_subscribe_limit, reset_cycle, status, effective_at, updated_at)
VALUES ('vqc_vip1', 'VIP1', 200, 100, 'MONTHLY', 'ACTIVE', NOW(), NOW())
ON DUPLICATE KEY UPDATE doc_read_limit = VALUES(doc_read_limit), news_subscribe_limit = VALUES(news_subscribe_limit), updated_at = VALUES(updated_at);

INSERT INTO vip_quota_configs (id, member_level, doc_read_limit, news_subscribe_limit, reset_cycle, status, effective_at, updated_at)
VALUES ('vqc_free', 'FREE', 20, 10, 'MONTHLY', 'ACTIVE', NOW(), NOW())
ON DUPLICATE KEY UPDATE doc_read_limit = VALUES(doc_read_limit), news_subscribe_limit = VALUES(news_subscribe_limit), updated_at = VALUES(updated_at);

INSERT INTO vip_quota_configs (id, member_level, doc_read_limit, news_subscribe_limit, reset_cycle, status, effective_at, updated_at)
VALUES ('vqc_vip2', 'VIP2', 500, 300, 'MONTHLY', 'ACTIVE', NOW(), NOW())
ON DUPLICATE KEY UPDATE doc_read_limit = VALUES(doc_read_limit), news_subscribe_limit = VALUES(news_subscribe_limit), updated_at = VALUES(updated_at);

INSERT INTO user_quota_usages (id, user_id, member_level, period_key, doc_read_used, news_subscribe_used, updated_at)
VALUES ('uqu_demo_001', 'u_demo_001', 'VIP1', DATE_FORMAT(NOW(), '%Y-%m'), 13, 7, NOW())
ON DUPLICATE KEY UPDATE doc_read_used = VALUES(doc_read_used), news_subscribe_used = VALUES(news_subscribe_used), updated_at = VALUES(updated_at);

INSERT INTO browse_histories (id, user_id, content_type, content_id, source_page, viewed_at)
VALUES ('bh_demo_001', 'u_demo_001', 'NEWS', 'article_demo_001', '/news', NOW())
ON DUPLICATE KEY UPDATE viewed_at = VALUES(viewed_at);

INSERT INTO recharge_records (id, user_id, order_no, amount, pay_channel, status, paid_at, remark, created_at)
VALUES ('rc_demo_001', 'u_demo_001', 'ORDER_DEMO_001', 199.00, 'ALIPAY', 'PAID', NOW(), 'VIP开通', NOW())
ON DUPLICATE KEY UPDATE status = VALUES(status), paid_at = VALUES(paid_at), created_at = VALUES(created_at);

INSERT INTO invite_links (id, user_id, invite_code, url, channel, status, expired_at, created_at)
VALUES ('il_demo_001', 'u_demo_001', 'DEMO2026', 'https://example.com/invite/DEMO2026', 'wechat', 'ACTIVE', DATE_ADD(NOW(), INTERVAL 30 DAY), NOW())
ON DUPLICATE KEY UPDATE status = VALUES(status), expired_at = VALUES(expired_at), created_at = VALUES(created_at);

INSERT INTO invite_records (id, inviter_user_id, invitee_user_id, invite_link_id, register_at, kyc_at, first_pay_at, status, risk_flag)
VALUES ('ir_demo_001', 'u_demo_001', 'u_demo_002', 'il_demo_001', NOW(), NOW(), NOW(), 'FIRST_PAID', 'NORMAL')
ON DUPLICATE KEY UPDATE status = VALUES(status), first_pay_at = VALUES(first_pay_at), risk_flag = VALUES(risk_flag);

INSERT INTO share_reward_records (id, inviter_user_id, invitee_user_id, invite_record_id, reward_type, reward_value, trigger_event, status, issued_at, review_reason, created_at)
VALUES ('srr_demo_001', 'u_demo_001', 'u_demo_002', 'ir_demo_001', 'CASH', 20.00, 'INVITEE_FIRST_RECHARGE', 'ISSUED', NOW(), NULL, NOW())
ON DUPLICATE KEY UPDATE status = VALUES(status), issued_at = VALUES(issued_at), reward_value = VALUES(reward_value);

INSERT INTO reward_wallets (id, user_id, cash_balance, cash_frozen, coupon_balance, vip_days_balance, updated_at)
VALUES ('rw_demo_001', 'u_demo_001', 120.00, 10.00, 30.00, 7, NOW())
ON DUPLICATE KEY UPDATE cash_balance = VALUES(cash_balance), cash_frozen = VALUES(cash_frozen), coupon_balance = VALUES(coupon_balance), vip_days_balance = VALUES(vip_days_balance), updated_at = VALUES(updated_at);

INSERT INTO reward_wallet_txns (id, wallet_id, txn_type, amount, status, ref_id, created_at)
VALUES ('rwt_demo_001', 'rw_demo_001', 'ISSUE', 20.00, 'SUCCESS', 'srr_demo_001', NOW())
ON DUPLICATE KEY UPDATE status = VALUES(status), amount = VALUES(amount), created_at = VALUES(created_at);

INSERT INTO withdraw_requests (id, user_id, amount, status, review_reason, applied_at, reviewed_at)
VALUES ('wd_demo_001', 'u_demo_001', 50.00, 'PENDING', NULL, NOW(), NULL)
ON DUPLICATE KEY UPDATE status = VALUES(status), applied_at = VALUES(applied_at);

INSERT INTO news_categories (id, name, slug, sort, visibility, status, created_at, updated_at)
VALUES ('nc_demo_001', '盘前速递', 'pre-market', 1, 'PUBLIC', 'PUBLISHED', NOW(), NOW())
ON DUPLICATE KEY UPDATE updated_at = VALUES(updated_at), status = VALUES(status);

INSERT INTO news_articles (id, category_id, title, summary, content, tags, visibility, status, published_at, author_id, created_at, updated_at)
VALUES ('article_demo_001', 'nc_demo_001', 'A股盘前观察', '示例资讯', '这是用于联调的示例资讯正文。', JSON_ARRAY('A股', '盘前'), 'VIP', 'PUBLISHED', NOW(), 'admin_demo_001', NOW(), NOW())
ON DUPLICATE KEY UPDATE summary = VALUES(summary), content = VALUES(content), updated_at = VALUES(updated_at), status = VALUES(status);

INSERT INTO news_attachments (id, article_id, file_name, file_url, file_size, mime_type, created_at)
VALUES ('att_demo_001', 'article_demo_001', 'daily-note.pdf', 'https://example.com/files/daily-note.pdf', 20480, 'application/pdf', NOW())
ON DUPLICATE KEY UPDATE file_url = VALUES(file_url), file_size = VALUES(file_size), created_at = VALUES(created_at);

INSERT INTO arbitrage_recos (id, type, contract_a, contract_b, spread, percentile, entry_point, exit_point, stop_point, trigger_rule, status)
VALUES ('arb_demo_001', 'CALENDAR', 'RB2505', 'RB2510', 118.50, 0.82, 120.00, 80.00, 150.00, '跨期价差高位回归', 'WATCH')
ON DUPLICATE KEY UPDATE spread = VALUES(spread), percentile = VALUES(percentile), status = VALUES(status);

INSERT INTO futures_guidances (id, contract, guidance_direction, position_level, entry_range, take_profit_range, stop_loss_range, risk_level, invalid_condition, valid_to)
VALUES ('fg_demo_001', 'RB2505', 'LONG_SPREAD', 'LIGHT', '110-125', '75-90', '150-160', 'MEDIUM', '基差结构反转', DATE_ADD(NOW(), INTERVAL 2 DAY))
ON DUPLICATE KEY UPDATE guidance_direction = VALUES(guidance_direction), position_level = VALUES(position_level), valid_to = VALUES(valid_to);

INSERT INTO reconciliation_records (id, pay_channel, batch_date, status, diff_count, created_at)
VALUES ('rec_demo_001', 'ALIPAY', CURRENT_DATE, 'DONE', 0, NOW())
ON DUPLICATE KEY UPDATE status = VALUES(status), diff_count = VALUES(diff_count), created_at = VALUES(created_at);

INSERT INTO risk_rule_configs (id, rule_code, rule_name, threshold, status, effective_at, updated_at)
VALUES ('rrc_demo_001', 'DEVICE_DUP', '同设备重复邀请', 3, 'ACTIVE', NOW(), NOW())
ON DUPLICATE KEY UPDATE threshold = VALUES(threshold), status = VALUES(status), updated_at = VALUES(updated_at);

INSERT INTO risk_hit_logs (id, rule_code, user_id, event_id, risk_level, status, created_at)
VALUES ('rhl_demo_001', 'DEVICE_DUP', 'u_demo_002', 'evt_demo_001', 'MEDIUM', 'PENDING_REVIEW', NOW())
ON DUPLICATE KEY UPDATE risk_level = VALUES(risk_level), status = VALUES(status), created_at = VALUES(created_at);
