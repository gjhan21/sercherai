-- Up
ALTER TABLE `stock_recommendations` ADD COLUMN `ai_review_content` TEXT COMMENT 'AI生成的复盘草稿' AFTER `review_note`;

CREATE TABLE `user_virtual_sandbox` (
  `id` varchar(36) NOT NULL COMMENT '主键ID',
  `user_id` varchar(36) NOT NULL COMMENT '用户ID',
  `stock_recommendation_id` varchar(36) NOT NULL COMMENT '股票推荐ID',
  `add_price` decimal(20,4) NOT NULL DEFAULT '0.0000' COMMENT '加入沙盘时的价格快照',
  `status` varchar(32) NOT NULL DEFAULT 'ACTIVE' COMMENT '状态: ACTIVE, REMOVED',
  `added_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '加入时间',
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  KEY `idx_user_id` (`user_id`),
  KEY `idx_stock_reco_id` (`stock_recommendation_id`),
  UNIQUE KEY `uk_user_stock` (`user_id`, `stock_recommendation_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='用户虚拟沙盘表';

-- Down
DROP TABLE IF EXISTS `user_virtual_sandbox`;
ALTER TABLE `stock_recommendations` DROP COLUMN `ai_review_content`;
