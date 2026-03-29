INSERT INTO users (id, phone, email, password_hash, status, kyc_status, member_level, created_at, updated_at)
VALUES
  ('u_demo_002', '13800000002', 'demo2@sercherai.local', 'a03c32fcd351cba2d9738622b083bed022ef07793bd92b59faea0207653f371d', 'ACTIVE', 'PENDING', 'FREE', NOW(), NOW()),
  ('u_demo_003', '13800000003', 'demo3@sercherai.local', 'a03c32fcd351cba2d9738622b083bed022ef07793bd92b59faea0207653f371d', 'DISABLED', 'REJECTED', 'VIP2', NOW(), NOW()),
  ('u_demo_004', '13800000004', 'demo4@sercherai.local', 'a03c32fcd351cba2d9738622b083bed022ef07793bd92b59faea0207653f371d', 'ACTIVE', 'APPROVED', 'VIP2', NOW(), NOW()),
  ('admin_002', '19900000002', 'admin2@sercherai.local', 'a03c32fcd351cba2d9738622b083bed022ef07793bd92b59faea0207653f371d', 'ACTIVE', 'APPROVED', 'VIP3', NOW(), NOW()),
  ('admin_003', '13800000011', 'admin3@sercherai.local', 'a03c32fcd351cba2d9738622b083bed022ef07793bd92b59faea0207653f371d', 'ACTIVE', 'APPROVED', 'VIP3', NOW(), NOW())
ON DUPLICATE KEY UPDATE
  email = VALUES(email),
  status = VALUES(status),
  kyc_status = VALUES(kyc_status),
  member_level = VALUES(member_level),
  updated_at = VALUES(updated_at);

INSERT INTO membership_products (id, name, price, status, member_level, duration_days)
VALUES
  ('mp_demo_002', 'VIP2 季卡', 268.00, 'ACTIVE', 'VIP2', 90),
  ('mp_demo_003', 'VIP2 年卡', 899.00, 'DISABLED', 'VIP2', 365)
ON DUPLICATE KEY UPDATE
  name = VALUES(name),
  price = VALUES(price),
  status = VALUES(status),
  member_level = VALUES(member_level),
  duration_days = VALUES(duration_days);

INSERT INTO membership_orders (id, order_no, user_id, product_id, amount, pay_channel, status, paid_at, created_at, updated_at)
VALUES
  ('mo_demo_002', 'mo_demo_002', 'u_demo_002', 'mp_demo_001', 99.00, 'WECHAT', 'PENDING', NULL, NOW(), NOW()),
  ('mo_demo_003', 'mo_demo_003', 'u_demo_003', 'mp_demo_002', 268.00, 'ALIPAY', 'FAILED', NULL, NOW(), NOW()),
  ('mo_demo_004', 'mo_demo_004', 'u_demo_001', 'mp_demo_002', 268.00, 'CARD', 'REFUNDED', DATE_SUB(NOW(), INTERVAL 2 DAY), DATE_SUB(NOW(), INTERVAL 4 DAY), NOW()),
  ('mo_demo_005', 'mo_demo_005', 'u_demo_004', 'mp_demo_002', 268.00, 'ALIPAY', 'PAID', DATE_SUB(NOW(), INTERVAL 1 DAY), DATE_SUB(NOW(), INTERVAL 1 DAY), NOW())
ON DUPLICATE KEY UPDATE
  status = VALUES(status),
  amount = VALUES(amount),
  pay_channel = VALUES(pay_channel),
  paid_at = VALUES(paid_at),
  updated_at = VALUES(updated_at);

INSERT INTO vip_quota_configs (id, member_level, doc_read_limit, news_subscribe_limit, reset_cycle, status, effective_at, updated_at)
VALUES
  ('vqc_vip3', 'VIP3', 1000, 600, 'MONTHLY', 'ACTIVE', NOW(), NOW()),
  ('vqc_vip2_weekly', 'VIP2', 160, 80, 'WEEKLY', 'DISABLED', DATE_SUB(NOW(), INTERVAL 30 DAY), NOW())
ON DUPLICATE KEY UPDATE
  doc_read_limit = VALUES(doc_read_limit),
  news_subscribe_limit = VALUES(news_subscribe_limit),
  reset_cycle = VALUES(reset_cycle),
  status = VALUES(status),
  updated_at = VALUES(updated_at);

INSERT INTO user_quota_usages (id, user_id, member_level, period_key, doc_read_used, news_subscribe_used, updated_at)
VALUES
  ('uqu_demo_002', 'u_demo_002', 'FREE', DATE_FORMAT(NOW(), '%Y-%m'), 7, 3, NOW()),
  ('uqu_demo_003', 'u_demo_003', 'VIP2', DATE_FORMAT(NOW(), '%Y-%m'), 66, 28, NOW()),
  ('uqu_demo_004', 'u_demo_004', 'VIP2', DATE_FORMAT(NOW(), '%Y-%m'), 31, 14, NOW())
ON DUPLICATE KEY UPDATE
  member_level = VALUES(member_level),
  doc_read_used = VALUES(doc_read_used),
  news_subscribe_used = VALUES(news_subscribe_used),
  updated_at = VALUES(updated_at);

INSERT INTO news_categories (id, name, slug, sort, visibility, status, created_at, updated_at)
VALUES
  ('nc_demo_002', '策略跟踪', 'strategy-track', 2, 'VIP', 'PUBLISHED', NOW(), NOW())
ON DUPLICATE KEY UPDATE
  name = VALUES(name),
  status = VALUES(status),
  updated_at = VALUES(updated_at);

INSERT INTO news_articles (id, category_id, title, summary, content, cover_url, tags, visibility, status, published_at, author_id, created_at, updated_at)
VALUES
  ('article_demo_002', 'nc_demo_002', '量化策略周报', '示例周报摘要', '用于 admin 管理端联调的策略周报正文。', 'https://images.unsplash.com/photo-1556157382-97eda2d62296?auto=format&fit=crop&w=1200&q=80', '["策略","周报"]', 'VIP', 'PUBLISHED', NOW(), 'admin_001', NOW(), NOW()),
  ('article_demo_003', 'nc_demo_001', '热点快评', '待审核稿件', '该稿件用于审核中心待处理演示。', 'https://images.unsplash.com/photo-1454165804606-c3d57bc86b40?auto=format&fit=crop&w=1200&q=80', '["热点","快评"]', 'PUBLIC', 'DRAFT', NULL, 'admin_001', NOW(), NOW())
ON DUPLICATE KEY UPDATE
  summary = VALUES(summary),
  content = VALUES(content),
  cover_url = VALUES(cover_url),
  status = VALUES(status),
  updated_at = VALUES(updated_at);

INSERT INTO stock_recommendations (id, symbol, name, score, risk_level, position_range, valid_from, valid_to, status, reason_summary, created_at)
VALUES
  ('sr_demo_002', '000001.SZ', '平安银行', 76.40, 'LOW', '5%-10%', NOW(), DATE_ADD(NOW(), INTERVAL 3 DAY), 'PUBLISHED', '估值中枢回归', NOW()),
  ('sr_demo_003', '300750.SZ', '宁德时代', 84.10, 'HIGH', '8%-12%', NOW(), DATE_ADD(NOW(), INTERVAL 2 DAY), 'DRAFT', '行业景气度提升', NOW())
ON DUPLICATE KEY UPDATE
  score = VALUES(score),
  risk_level = VALUES(risk_level),
  position_range = VALUES(position_range),
  valid_to = VALUES(valid_to),
  status = VALUES(status),
  reason_summary = VALUES(reason_summary);

INSERT INTO futures_strategies (id, contract, name, direction, risk_level, position_range, valid_from, valid_to, status, reason_summary)
VALUES
  ('fs_demo_002', 'RB2601', '螺纹短线策略', 'SHORT', 'HIGH', '15%-25%', NOW(), DATE_ADD(NOW(), INTERVAL 2 DAY), 'ACTIVE', '库存去化预期走弱'),
  ('fs_demo_003', 'AU2604', '黄金波段策略', 'LONG', 'LOW', '10%-20%', NOW(), DATE_ADD(NOW(), INTERVAL 5 DAY), 'DRAFT', '避险偏好抬升')
ON DUPLICATE KEY UPDATE
  direction = VALUES(direction),
  risk_level = VALUES(risk_level),
  position_range = VALUES(position_range),
  valid_to = VALUES(valid_to),
  status = VALUES(status),
  reason_summary = VALUES(reason_summary);

INSERT INTO review_tasks (id, module, target_id, submitter_id, reviewer_id, status, submit_note, review_note, submitted_at, reviewed_at, created_at, updated_at)
VALUES
  ('rt_demo_001', 'NEWS', 'article_demo_003', 'admin_001', 'admin_002', 'PENDING', '请优先审核热点稿件', NULL, DATE_SUB(NOW(), INTERVAL 5 HOUR), NULL, NOW(), NOW()),
  ('rt_demo_002', 'STOCK', 'sr_demo_002', 'admin_001', 'admin_001', 'APPROVED', '策略可发布', '通过，风险提示已补充', DATE_SUB(NOW(), INTERVAL 1 DAY), DATE_SUB(NOW(), INTERVAL 20 HOUR), NOW(), NOW()),
  ('rt_demo_003', 'FUTURES', 'fs_demo_003', 'admin_002', 'admin_001', 'REJECTED', '请补充止损规则', '止损区间不明确', DATE_SUB(NOW(), INTERVAL 2 DAY), DATE_SUB(NOW(), INTERVAL 40 HOUR), NOW(), NOW())
ON DUPLICATE KEY UPDATE
  reviewer_id = VALUES(reviewer_id),
  status = VALUES(status),
  submit_note = VALUES(submit_note),
  review_note = VALUES(review_note),
  submitted_at = VALUES(submitted_at),
  reviewed_at = VALUES(reviewed_at),
  updated_at = VALUES(updated_at);

INSERT INTO workflow_messages (id, review_id, target_id, module, receiver_id, sender_id, event_type, title, content, is_read, created_at, read_at)
VALUES
  ('wm_demo_001', 'rt_demo_001', 'article_demo_003', 'NEWS', 'admin_002', 'admin_001', 'REVIEW_ASSIGNED', '审核任务已分配', '任务 rt_demo_001 已分配给你，请在今日内处理。', 0, DATE_SUB(NOW(), INTERVAL 3 HOUR), NULL),
  ('wm_demo_002', 'rt_demo_002', 'sr_demo_002', 'STOCK', 'admin_001', 'admin_001', 'REVIEW_APPROVED', '审核已通过', '任务 rt_demo_002 已通过，可进入发布流程。', 1, DATE_SUB(NOW(), INTERVAL 10 HOUR), DATE_SUB(NOW(), INTERVAL 9 HOUR)),
  ('wm_demo_003', NULL, 'wind', 'SYSTEM', 'admin_001', 'system', 'DATA_SOURCE_UNHEALTHY', '数据源告警', 'wind 数据源连续失败，请检查。', 0, DATE_SUB(NOW(), INTERVAL 1 HOUR), NULL)
ON DUPLICATE KEY UPDATE
  title = VALUES(title),
  content = VALUES(content),
  is_read = VALUES(is_read),
  read_at = VALUES(read_at),
  created_at = VALUES(created_at);

INSERT INTO scheduler_job_definitions (id, job_name, display_name, module, cron_expr, status, last_run_at, updated_by, created_at, updated_at)
VALUES
  ('jobdef_news_digest', 'daily_news_digest', '每日新闻摘要生成', 'NEWS', '0 0 7 * * *', 'ACTIVE', DATE_SUB(NOW(), INTERVAL 1 DAY), 'admin_001', NOW(), NOW()),
  ('jobdef_cleanup', 'nightly_cleanup', '夜间清理任务', 'SYSTEM', '0 30 2 * * *', 'DISABLED', DATE_SUB(NOW(), INTERVAL 2 DAY), 'admin_001', NOW(), NOW())
ON DUPLICATE KEY UPDATE
  display_name = VALUES(display_name),
  module = VALUES(module),
  cron_expr = VALUES(cron_expr),
  status = VALUES(status),
  last_run_at = VALUES(last_run_at),
  updated_by = VALUES(updated_by),
  updated_at = VALUES(updated_at);

INSERT INTO scheduler_job_runs (id, parent_run_id, retry_count, job_name, trigger_source, status, started_at, finished_at, result_summary, error_message, operator_id, created_at)
VALUES
  ('jr_demo_001', NULL, 0, 'daily_stock_recommendation', 'SYSTEM', 'SUCCESS', DATE_SUB(NOW(), INTERVAL 6 HOUR), DATE_SUB(NOW(), INTERVAL 6 HOUR) + INTERVAL 2 MINUTE, '生成 20 条推荐', NULL, 'system', NOW()),
  ('jr_demo_002', NULL, 0, 'daily_news_digest', 'SYSTEM', 'FAILED', DATE_SUB(NOW(), INTERVAL 4 HOUR), DATE_SUB(NOW(), INTERVAL 4 HOUR) + INTERVAL 1 MINUTE, '抓取 50 篇新闻，解析失败 8 篇', 'upstream timeout', 'system', NOW()),
  ('jr_demo_003', 'jr_demo_002', 1, 'daily_news_digest', 'MANUAL', 'RUNNING', DATE_SUB(NOW(), INTERVAL 20 MINUTE), NULL, '重试执行中', NULL, 'admin_001', NOW())
ON DUPLICATE KEY UPDATE
  parent_run_id = VALUES(parent_run_id),
  retry_count = VALUES(retry_count),
  status = VALUES(status),
  started_at = VALUES(started_at),
  finished_at = VALUES(finished_at),
  result_summary = VALUES(result_summary),
  error_message = VALUES(error_message),
  operator_id = VALUES(operator_id);

INSERT INTO risk_rule_configs (id, rule_code, rule_name, threshold, status, effective_at, updated_at)
VALUES
  ('rrc_demo_002', 'LOGIN_FAIL', '登录失败次数阈值', 5, 'ACTIVE', NOW(), NOW()),
  ('rrc_demo_003', 'IP_BRUTE_FORCE', 'IP 暴力尝试阈值', 18, 'ACTIVE', NOW(), NOW())
ON DUPLICATE KEY UPDATE
  rule_name = VALUES(rule_name),
  threshold = VALUES(threshold),
  status = VALUES(status),
  updated_at = VALUES(updated_at);

INSERT INTO risk_hit_logs (id, rule_code, user_id, event_id, risk_level, status, created_at)
VALUES
  ('rhl_demo_002', 'LOGIN_FAIL', 'u_demo_002', 'evt_login_001', 'MEDIUM', 'PENDING', DATE_SUB(NOW(), INTERVAL 90 MINUTE)),
  ('rhl_demo_003', 'IP_BRUTE_FORCE', 'u_demo_003', 'evt_login_002', 'HIGH', 'CONFIRMED', DATE_SUB(NOW(), INTERVAL 30 MINUTE))
ON DUPLICATE KEY UPDATE
  risk_level = VALUES(risk_level),
  status = VALUES(status),
  created_at = VALUES(created_at);

INSERT INTO share_reward_records (id, inviter_user_id, invitee_user_id, invite_record_id, reward_type, reward_value, trigger_event, status, issued_at, review_reason, created_at)
VALUES
  ('srr_demo_002', 'u_demo_001', 'u_demo_003', 'ir_demo_001', 'CASH', 10.00, 'INVITEE_REGISTER', 'PENDING', NULL, NULL, NOW()),
  ('srr_demo_003', 'u_demo_004', 'u_demo_002', 'ir_demo_001', 'VIP_DAYS', 15.00, 'INVITEE_FIRST_RECHARGE', 'REJECTED', NULL, '风控规则拦截', NOW())
ON DUPLICATE KEY UPDATE
  reward_value = VALUES(reward_value),
  status = VALUES(status),
  review_reason = VALUES(review_reason),
  created_at = VALUES(created_at);

INSERT INTO withdraw_requests (id, user_id, amount, status, review_reason, applied_at, reviewed_at)
VALUES
  ('wd_demo_002', 'u_demo_002', 88.00, 'APPROVED', '审核通过，待打款', DATE_SUB(NOW(), INTERVAL 1 DAY), DATE_SUB(NOW(), INTERVAL 20 HOUR)),
  ('wd_demo_003', 'u_demo_003', 60.00, 'REJECTED', '资料不完整', DATE_SUB(NOW(), INTERVAL 10 HOUR), DATE_SUB(NOW(), INTERVAL 8 HOUR))
ON DUPLICATE KEY UPDATE
  amount = VALUES(amount),
  status = VALUES(status),
  review_reason = VALUES(review_reason),
  reviewed_at = VALUES(reviewed_at);

INSERT INTO reconciliation_records (id, pay_channel, batch_date, status, diff_count, created_at)
VALUES
  ('rec_demo_002', 'WECHAT', CURRENT_DATE, 'PENDING', 2, NOW()),
  ('rec_demo_003', 'ALIPAY', DATE_SUB(CURRENT_DATE, INTERVAL 1 DAY), 'FAILED', 5, NOW())
ON DUPLICATE KEY UPDATE
  status = VALUES(status),
  diff_count = VALUES(diff_count),
  created_at = VALUES(created_at);

INSERT INTO invite_records (id, inviter_user_id, invitee_user_id, invite_link_id, register_at, kyc_at, first_pay_at, status, risk_flag)
VALUES
  ('ir_demo_002', 'u_demo_001', 'u_demo_003', 'il_demo_001', DATE_SUB(NOW(), INTERVAL 3 DAY), NULL, NULL, 'REGISTERED', 'LOW'),
  ('ir_demo_003', 'u_demo_004', 'u_demo_002', 'il_demo_001', DATE_SUB(NOW(), INTERVAL 2 DAY), DATE_SUB(NOW(), INTERVAL 1 DAY), NULL, 'KYC_PASSED', 'NORMAL')
ON DUPLICATE KEY UPDATE
  status = VALUES(status),
  risk_flag = VALUES(risk_flag),
  first_pay_at = VALUES(first_pay_at);

INSERT INTO auth_login_logs (id, user_id, phone, action, status, reason, ip, user_agent, created_at)
VALUES
  ('all_demo_001', 'admin_001', '13800000000', 'LOGIN', 'SUCCESS', NULL, '10.0.0.1', 'Mozilla/5.0', DATE_SUB(NOW(), INTERVAL 2 HOUR)),
  ('all_demo_002', 'u_demo_003', '13800000003', 'LOGIN', 'FAILED', 'bad_password', '10.0.0.2', 'Mozilla/5.0', DATE_SUB(NOW(), INTERVAL 80 MINUTE)),
  ('all_demo_003', 'admin_002', '13800000010', 'REFRESH', 'SUCCESS', NULL, '10.0.0.3', 'Mozilla/5.0', DATE_SUB(NOW(), INTERVAL 45 MINUTE))
ON DUPLICATE KEY UPDATE
  status = VALUES(status),
  reason = VALUES(reason),
  created_at = VALUES(created_at);

INSERT INTO auth_risk_config_logs (
  id, operator_user_id, old_phone_fail, old_ip_fail, old_ip_phone, old_lock_seconds,
  new_phone_fail, new_ip_fail, new_ip_phone, new_lock_seconds, created_at
)
VALUES
  ('arcl_demo_001', 'admin_001', 5, 20, 5, 900, 6, 24, 6, 1200, DATE_SUB(NOW(), INTERVAL 1 DAY))
ON DUPLICATE KEY UPDATE
  new_phone_fail = VALUES(new_phone_fail),
  new_ip_fail = VALUES(new_ip_fail),
  new_ip_phone = VALUES(new_ip_phone),
  new_lock_seconds = VALUES(new_lock_seconds),
  created_at = VALUES(created_at);

INSERT INTO auth_unlock_logs (id, operator_user_id, phone, ip, reason, created_at)
VALUES
  ('aul_demo_001', 'admin_001', '13800000003', '10.0.0.2', '客服手动解锁', DATE_SUB(NOW(), INTERVAL 40 MINUTE))
ON DUPLICATE KEY UPDATE
  reason = VALUES(reason),
  created_at = VALUES(created_at);

INSERT INTO admin_operation_logs (id, module, action, target_type, target_id, operator_user_id, before_value, after_value, reason, created_at)
VALUES
  ('aol_demo_001', 'MEMBERSHIP', 'UPDATE_ORDER_STATUS', 'MEMBERSHIP_ORDER', 'mo_demo_002', 'admin_001', 'PENDING', 'PAID', '人工补单', DATE_SUB(NOW(), INTERVAL 5 HOUR)),
  ('aol_demo_002', 'RISK', 'CREATE_RULE', 'RISK_RULE', 'rrc_demo_003', 'admin_001', '', 'ACTIVE', '新增IP风控规则', DATE_SUB(NOW(), INTERVAL 3 HOUR)),
  ('aol_demo_003', 'SYSTEM', 'UPSERT_CONFIG', 'SYSTEM_CONFIG', 'data_source.eastmoney', 'admin_001', '', '{"status":"ACTIVE"}', '新增数据源', DATE_SUB(NOW(), INTERVAL 2 HOUR))
ON DUPLICATE KEY UPDATE
  action = VALUES(action),
  after_value = VALUES(after_value),
  reason = VALUES(reason),
  created_at = VALUES(created_at);

INSERT INTO system_configs (id, config_key, config_value, description, updated_by, updated_at)
VALUES
  (
    'cfg_data_source_eastmoney',
    'data_source.eastmoney',
    '{"name":"东方财富","source_type":"STOCK","status":"ACTIVE","config":{"endpoint":"https://api.example.com/eastmoney","retry_times":2,"fail_threshold":3,"retry_interval_ms":200,"health_timeout_ms":3000,"alert_receiver_id":"admin_001"}}',
    '东方财富数据源',
    'admin_001',
    NOW()
  ),
  (
    'cfg_data_source_binance',
    'data_source.binance',
    '{"name":"Binance","source_type":"FUTURES","status":"DISABLED","config":{"endpoint":"https://api.example.com/binance","retry_times":1,"fail_threshold":2,"retry_interval_ms":300,"health_timeout_ms":2500,"alert_receiver_id":"admin_002"}}',
    'Binance数据源',
    'admin_001',
    NOW()
  ),
  ('cfg_review_sla_hours', 'workflow.review.sla_hours', '48', '审核SLA小时数', 'admin_001', NOW())
ON DUPLICATE KEY UPDATE
  config_value = VALUES(config_value),
  description = VALUES(description),
  updated_by = VALUES(updated_by),
  updated_at = VALUES(updated_at);

INSERT INTO rbac_roles (id, role_key, role_name, description, status, built_in, created_at, updated_at)
VALUES
  ('role_content_editor', 'CONTENT_EDITOR', '内容编辑', '新闻内容编辑和审核查看', 'ACTIVE', 0, NOW(), NOW())
ON DUPLICATE KEY UPDATE
  role_name = VALUES(role_name),
  description = VALUES(description),
  status = VALUES(status),
  updated_at = VALUES(updated_at);

INSERT INTO rbac_role_permissions (role_id, permission_code, created_at)
VALUES
  ('role_content_editor', 'dashboard.view', NOW()),
  ('role_content_editor', 'news.view', NOW()),
  ('role_content_editor', 'news.edit', NOW()),
  ('role_content_editor', 'community.view', NOW()),
  ('role_content_editor', 'community.review', NOW()),
  ('role_content_editor', 'review.view', NOW()),
  ('role_content_editor', 'workflow.view', NOW())
ON DUPLICATE KEY UPDATE
  created_at = VALUES(created_at);

INSERT INTO discussion_topics (
  id, user_id, title, summary, content, topic_type, stance, time_horizon, reason_text, risk_text,
  status, comment_count, like_count, favorite_count, report_count, last_active_at, created_at, updated_at
)
VALUES (
  'ct_demo_002', 'u_demo_002', '盘后研报怎么看，是继续持有还是等回踩',
  '围绕研报解读后的持仓节奏讨论，适合资讯页引流到讨论页联调。',
  '我更关注研报落地后的兑现节奏，如果只是情绪催化，第二天不一定适合追。',
  'NEWS', 'WATCH', 'SWING', '研报热度高，但价格反应已经较快。', '如果开盘直接高开过大，追入性价比会明显下降。',
  'PENDING_REVIEW', 1, 0, 0, 1, NOW(), NOW(), NOW()
)
ON DUPLICATE KEY UPDATE
  summary = VALUES(summary),
  content = VALUES(content),
  topic_type = VALUES(topic_type),
  stance = VALUES(stance),
  time_horizon = VALUES(time_horizon),
  reason_text = VALUES(reason_text),
  risk_text = VALUES(risk_text),
  status = VALUES(status),
  comment_count = VALUES(comment_count),
  report_count = VALUES(report_count),
  last_active_at = VALUES(last_active_at),
  updated_at = VALUES(updated_at);

INSERT INTO discussion_topic_links (topic_id, target_type, target_id, target_snapshot, created_at, updated_at)
VALUES ('ct_demo_002', 'NEWS_ARTICLE', 'article_demo_002', '量化策略周报', NOW(), NOW())
ON DUPLICATE KEY UPDATE
  target_type = VALUES(target_type),
  target_id = VALUES(target_id),
  target_snapshot = VALUES(target_snapshot),
  updated_at = VALUES(updated_at);

INSERT INTO discussion_comments (id, topic_id, user_id, parent_comment_id, reply_to_user_id, content, status, like_count, created_at, updated_at)
VALUES ('cc_demo_002', 'ct_demo_002', 'admin_003', NULL, NULL, '这类话题适合放资讯详情页侧栏引导，不适合做成聊天流。', 'PUBLISHED', 2, NOW(), NOW())
ON DUPLICATE KEY UPDATE
  content = VALUES(content),
  status = VALUES(status),
  like_count = VALUES(like_count),
  updated_at = VALUES(updated_at);

INSERT INTO discussion_reports (id, reporter_user_id, target_type, target_id, reason, status, review_note, created_at, updated_at)
VALUES ('cr_demo_001', 'u_demo_004', 'TOPIC', 'ct_demo_002', '标题偏情绪化，建议管理员复核。', 'PENDING', NULL, NOW(), NOW())
ON DUPLICATE KEY UPDATE
  reason = VALUES(reason),
  status = VALUES(status),
  updated_at = VALUES(updated_at);

INSERT INTO rbac_user_roles (user_id, role_id, created_at)
VALUES
  ('admin_003', 'role_content_editor', NOW())
ON DUPLICATE KEY UPDATE
  created_at = VALUES(created_at);

INSERT INTO data_source_health_logs (id, source_key, status, reachable, http_status, latency_ms, message, checked_at)
VALUES
  ('dsh_demo_001', 'eastmoney', 'HEALTHY', 1, 200, 118, 'ok', DATE_SUB(NOW(), INTERVAL 25 MINUTE)),
  ('dsh_demo_002', 'eastmoney', 'UNHEALTHY', 0, 504, 3000, 'timeout', DATE_SUB(NOW(), INTERVAL 5 MINUTE)),
  ('dsh_demo_003', 'binance', 'UNHEALTHY', 0, 503, 2100, 'service unavailable', DATE_SUB(NOW(), INTERVAL 15 MINUTE))
ON DUPLICATE KEY UPDATE
  status = VALUES(status),
  reachable = VALUES(reachable),
  http_status = VALUES(http_status),
  latency_ms = VALUES(latency_ms),
  message = VALUES(message),
  checked_at = VALUES(checked_at);

INSERT INTO scheduler_job_definitions (id, job_name, display_name, module, cron_expr, status, last_run_at, updated_by, created_at, updated_at)
VALUES
  ('jobdef_forecast_l3_dispatch_pending', 'forecast_l3_dispatch_pending', 'Forecast L3 Pending Dispatch', 'GROWTH', '0 */10 * * * *', 'DISABLED', NULL, 'admin_001', NOW(), NOW()),
  ('jobdef_forecast_l3_quality_backfill', 'forecast_l3_quality_backfill', 'Forecast L3 Quality Backfill', 'GROWTH', '0 15 * * * *', 'DISABLED', NULL, 'admin_001', NOW(), NOW())
ON DUPLICATE KEY UPDATE
  display_name = VALUES(display_name),
  module = VALUES(module),
  cron_expr = VALUES(cron_expr),
  status = VALUES(status),
  updated_by = VALUES(updated_by),
  updated_at = VALUES(updated_at);

INSERT INTO system_configs (id, config_key, config_value, description, updated_by, updated_at)
VALUES
  ('cfg_growth_forecast_l3_enabled', 'growth.forecast_l3.enabled', 'true', 'enable forecast l3 runtime', 'admin_001', NOW()),
  ('cfg_growth_forecast_l3_admin_manual_enabled', 'growth.forecast_l3.admin_manual_enabled', 'true', 'allow admin manual forecast l3 runs', 'admin_001', NOW()),
  ('cfg_growth_forecast_l3_user_request_enabled', 'growth.forecast_l3.user_request_enabled', 'true', 'allow user-request forecast l3 runs', 'admin_001', NOW()),
  ('cfg_growth_forecast_l3_auto_priority_enabled', 'growth.forecast_l3.auto_priority_enabled', 'false', 'allow auto-priority forecast l3 runs', 'admin_001', NOW()),
  ('cfg_growth_forecast_l3_client_read_enabled', 'growth.forecast_l3.client_read_enabled', 'true', 'allow client read-side forecast l3 summary', 'admin_001', NOW()),
  ('cfg_growth_forecast_l3_require_vip_for_full_report', 'growth.forecast_l3.require_vip_for_full_report', 'true', 'require vip for full forecast l3 report', 'admin_001', NOW()),
  ('cfg_growth_forecast_l3_max_active_runs', 'growth.forecast_l3.max_active_runs', '3', 'max active forecast l3 runs', 'admin_001', NOW()),
  ('cfg_growth_forecast_l3_max_runs_per_day', 'growth.forecast_l3.max_runs_per_day', '48', 'max daily forecast l3 runs', 'admin_001', NOW()),
  ('cfg_growth_forecast_l3_max_user_runs_per_day', 'growth.forecast_l3.max_user_runs_per_day', '2', 'max user-request forecast l3 runs per day', 'admin_001', NOW()),
  ('cfg_growth_forecast_l3_min_priority_threshold', 'growth.forecast_l3.min_priority_threshold', '0.72', 'min priority threshold for forecast l3', 'admin_001', NOW()),
  ('cfg_growth_forecast_l3_dispatch_enabled', 'growth.forecast_l3.dispatch.enabled', 'true', 'enable forecast l3 dispatch worker', 'admin_001', NOW()),
  ('cfg_growth_forecast_l3_dispatch_interval_minutes', 'growth.forecast_l3.dispatch.interval_minutes', '8', 'dispatch worker interval minutes', 'admin_001', NOW()),
  ('cfg_growth_forecast_l3_quality_enabled', 'growth.forecast_l3.quality.enabled', 'true', 'enable forecast l3 quality worker', 'admin_001', NOW()),
  ('cfg_growth_forecast_l3_quality_interval_minutes', 'growth.forecast_l3.quality.interval_minutes', '30', 'quality worker interval minutes', 'admin_001', NOW()),
  ('cfg_growth_forecast_l3_default_engine_key', 'growth.forecast_l3.default_engine_key', 'LOCAL_SYNTHESIS', 'default engine key for forecast l3', 'admin_001', NOW())
ON DUPLICATE KEY UPDATE
  config_value = VALUES(config_value),
  description = VALUES(description),
  updated_by = VALUES(updated_by),
  updated_at = VALUES(updated_at);

INSERT INTO strategy_forecast_l3_runs (
  id, target_type, target_id, target_key, target_label, trigger_type,
  request_user_id, operator_user_id, engine_key, status, priority_score,
  reason, failure_reason, context_meta_json, summary_json, report_ref_json,
  queued_at, started_at, finished_at, cancelled_at, created_at, updated_at
)
VALUES (
  'l3run_demo_seed_001', 'STOCK', 'reco_demo_001', '600519.SH', '贵州茅台', 'ADMIN_MANUAL',
  'admin_001', 'admin_001', 'LOCAL_SYNTHESIS', 'SUCCEEDED', 0.8800,
  'seeded demo run', NULL,
  JSON_OBJECT('source', 'seed', 'context', 'publish_history'),
  JSON_OBJECT(
    'run_id', 'l3run_demo_seed_001',
    'status', 'SUCCEEDED',
    'engine_key', 'LOCAL_SYNTHESIS',
    'trigger_type', 'ADMIN_MANUAL',
    'target_type', 'STOCK',
    'target_key', '600519.SH',
    'target_label', '贵州茅台',
    'executive_summary', '趋势延续，但需要确认成交额是否继续放大。',
    'primary_scenario', 'base',
    'action_guidance', '先等确认，再考虑加仓。',
    'confidence_label', 'MEDIUM',
    'priority_score', 0.88,
    'generated_at', DATE_FORMAT(DATE_SUB(NOW(), INTERVAL 40 MINUTE), '%Y-%m-%dT%H:%i:%sZ'),
    'report_available', TRUE
  ),
  JSON_OBJECT(
    'run_id', 'l3run_demo_seed_001',
    'report_id', 'l3report_demo_seed_001',
    'status', 'SUCCEEDED',
    'engine_key', 'LOCAL_SYNTHESIS',
    'generated_at', DATE_FORMAT(DATE_SUB(NOW(), INTERVAL 40 MINUTE), '%Y-%m-%dT%H:%i:%sZ'),
    'requires_vip', TRUE,
    'full_readable', FALSE
  ),
  DATE_SUB(NOW(), INTERVAL 55 MINUTE),
  DATE_SUB(NOW(), INTERVAL 52 MINUTE),
  DATE_SUB(NOW(), INTERVAL 40 MINUTE),
  NULL,
  DATE_SUB(NOW(), INTERVAL 55 MINUTE),
  DATE_SUB(NOW(), INTERVAL 40 MINUTE)
)
ON DUPLICATE KEY UPDATE
  status = VALUES(status),
  priority_score = VALUES(priority_score),
  reason = VALUES(reason),
  context_meta_json = VALUES(context_meta_json),
  summary_json = VALUES(summary_json),
  report_ref_json = VALUES(report_ref_json),
  queued_at = VALUES(queued_at),
  started_at = VALUES(started_at),
  finished_at = VALUES(finished_at),
  updated_at = VALUES(updated_at);

INSERT INTO strategy_forecast_l3_reports (
  id, run_id, version, executive_summary, primary_scenario,
  alternative_scenarios_json, trigger_checklist_json, invalidation_signals_json,
  role_disagreements_json, action_guidance_json, markdown_body, html_body,
  summary_json, created_at, updated_at
)
VALUES (
  'l3report_demo_seed_001', 'l3run_demo_seed_001', 1, '趋势延续，但需要确认成交额是否继续放大。', 'base',
  JSON_ARRAY(
    JSON_OBJECT('name', 'bull', 'probability', 0.24, 'thesis', '量价继续共振', 'action', '顺势跟踪'),
    JSON_OBJECT('name', 'bear', 'probability', 0.18, 'thesis', '高位分歧扩大', 'action', '收缩仓位')
  ),
  JSON_ARRAY(
    JSON_OBJECT('label', '成交额', 'status', 'WATCH', 'note', '等待放量确认', 'trigger', '成交额继续放大'),
    JSON_OBJECT('label', '资金回流', 'status', 'READY', 'note', '机构回流延续', 'trigger', '北向继续净流入')
  ),
  JSON_ARRAY('跌破关键均线', '北向资金明显转负'),
  JSON_ARRAY(
    JSON_OBJECT('role', 'RISK', 'stance', 'CAUTION', 'summary', '高位追涨赔率一般', 'veto', FALSE)
  ),
  JSON_ARRAY('先看确认', '缩短验证窗口'),
  '# Forecast L3 Demo\n\n- Executive summary: trend remains intact.\n- Action: wait for confirmation.',
  '<h1>Forecast L3 Demo</h1><p>Trend remains intact. Wait for confirmation.</p>',
  JSON_OBJECT(
    'run_id', 'l3run_demo_seed_001',
    'status', 'SUCCEEDED',
    'engine_key', 'LOCAL_SYNTHESIS',
    'trigger_type', 'ADMIN_MANUAL',
    'target_type', 'STOCK',
    'target_key', '600519.SH',
    'target_label', '贵州茅台',
    'executive_summary', '趋势延续，但需要确认成交额是否继续放大。',
    'primary_scenario', 'base',
    'action_guidance', '先等确认，再考虑加仓。',
    'confidence_label', 'MEDIUM',
    'priority_score', 0.88,
    'generated_at', DATE_FORMAT(DATE_SUB(NOW(), INTERVAL 40 MINUTE), '%Y-%m-%dT%H:%i:%sZ'),
    'report_available', TRUE
  ),
  DATE_SUB(NOW(), INTERVAL 40 MINUTE),
  DATE_SUB(NOW(), INTERVAL 40 MINUTE)
)
ON DUPLICATE KEY UPDATE
  executive_summary = VALUES(executive_summary),
  primary_scenario = VALUES(primary_scenario),
  alternative_scenarios_json = VALUES(alternative_scenarios_json),
  trigger_checklist_json = VALUES(trigger_checklist_json),
  invalidation_signals_json = VALUES(invalidation_signals_json),
  role_disagreements_json = VALUES(role_disagreements_json),
  action_guidance_json = VALUES(action_guidance_json),
  markdown_body = VALUES(markdown_body),
  html_body = VALUES(html_body),
  summary_json = VALUES(summary_json),
  updated_at = VALUES(updated_at);

INSERT INTO strategy_forecast_l3_logs (id, run_id, step_key, status, message, payload_json, created_at)
VALUES
  ('l3log_demo_seed_001', 'l3run_demo_seed_001', 'LOAD_CONTEXT', 'SUCCESS', 'publish history and explanation loaded', JSON_OBJECT('sources', 4), DATE_SUB(NOW(), INTERVAL 53 MINUTE)),
  ('l3log_demo_seed_002', 'l3run_demo_seed_001', 'RUN_DEEP_FORECAST', 'SUCCESS', 'local synthesis finished', JSON_OBJECT('engine', 'LOCAL_SYNTHESIS', 'roles', 4), DATE_SUB(NOW(), INTERVAL 45 MINUTE)),
  ('l3log_demo_seed_003', 'l3run_demo_seed_001', 'BUILD_REPORT', 'SUCCESS', 'report persisted', JSON_OBJECT('report_id', 'l3report_demo_seed_001'), DATE_SUB(NOW(), INTERVAL 40 MINUTE))
ON DUPLICATE KEY UPDATE
  status = VALUES(status),
  message = VALUES(message),
  payload_json = VALUES(payload_json),
  created_at = VALUES(created_at);

INSERT INTO strategy_forecast_l3_learning_records (
  id, run_id, target_type, target_key, scenario_hit, trigger_hit, invalidation_early,
  bias_label, role_effectiveness_json, summary_text, created_at, updated_at
)
VALUES (
  'l3learn_demo_seed_001', 'l3run_demo_seed_001', 'STOCK', '600519.SH', 1, 1, 0,
  'NONE', JSON_OBJECT('RISK', 0.64, 'SUPPLY', 0.58, 'EVENT', 0.61), 'Demo learning record for seeded forecast l3 run.',
  DATE_SUB(NOW(), INTERVAL 10 MINUTE), DATE_SUB(NOW(), INTERVAL 10 MINUTE)
)
ON DUPLICATE KEY UPDATE
  scenario_hit = VALUES(scenario_hit),
  trigger_hit = VALUES(trigger_hit),
  invalidation_early = VALUES(invalidation_early),
  bias_label = VALUES(bias_label),
  role_effectiveness_json = VALUES(role_effectiveness_json),
  summary_text = VALUES(summary_text),
  updated_at = VALUES(updated_at);
