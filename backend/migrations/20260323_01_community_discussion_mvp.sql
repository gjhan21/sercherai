CREATE TABLE IF NOT EXISTS discussion_topics (
  id VARCHAR(64) PRIMARY KEY,
  user_id VARCHAR(64) NOT NULL,
  title VARCHAR(200) NOT NULL,
  summary TEXT NOT NULL,
  content MEDIUMTEXT NOT NULL,
  topic_type VARCHAR(32) NOT NULL,
  stance VARCHAR(32) NOT NULL,
  time_horizon VARCHAR(32) NOT NULL DEFAULT '',
  reason_text TEXT NOT NULL,
  risk_text TEXT NOT NULL,
  status VARCHAR(32) NOT NULL,
  comment_count INT NOT NULL DEFAULT 0,
  like_count INT NOT NULL DEFAULT 0,
  favorite_count INT NOT NULL DEFAULT 0,
  report_count INT NOT NULL DEFAULT 0,
  last_active_at DATETIME NOT NULL,
  created_at DATETIME NOT NULL,
  updated_at DATETIME NOT NULL,
  KEY idx_discussion_topics_status_created (status, created_at),
  KEY idx_discussion_topics_type_active (topic_type, last_active_at),
  KEY idx_discussion_topics_user_created (user_id, created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS discussion_topic_links (
  topic_id VARCHAR(64) PRIMARY KEY,
  target_type VARCHAR(32) NOT NULL,
  target_id VARCHAR(64) NOT NULL,
  target_snapshot TEXT NOT NULL,
  created_at DATETIME NOT NULL,
  updated_at DATETIME NOT NULL,
  KEY idx_discussion_topic_links_target (target_type, target_id),
  CONSTRAINT fk_discussion_topic_links_topic
    FOREIGN KEY (topic_id) REFERENCES discussion_topics(id)
    ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS discussion_comments (
  id VARCHAR(64) PRIMARY KEY,
  topic_id VARCHAR(64) NOT NULL,
  user_id VARCHAR(64) NOT NULL,
  parent_comment_id VARCHAR(64) NULL,
  reply_to_user_id VARCHAR(64) NULL,
  content TEXT NOT NULL,
  status VARCHAR(32) NOT NULL,
  like_count INT NOT NULL DEFAULT 0,
  created_at DATETIME NOT NULL,
  updated_at DATETIME NOT NULL,
  KEY idx_discussion_comments_topic_created (topic_id, created_at),
  KEY idx_discussion_comments_user_created (user_id, created_at),
  KEY idx_discussion_comments_status_created (status, created_at),
  CONSTRAINT fk_discussion_comments_topic
    FOREIGN KEY (topic_id) REFERENCES discussion_topics(id)
    ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS discussion_reactions (
  id VARCHAR(64) PRIMARY KEY,
  user_id VARCHAR(64) NOT NULL,
  target_type VARCHAR(32) NOT NULL,
  target_id VARCHAR(64) NOT NULL,
  reaction_type VARCHAR(32) NOT NULL,
  created_at DATETIME NOT NULL,
  UNIQUE KEY uniq_discussion_reactions_user_target (user_id, target_type, target_id, reaction_type),
  KEY idx_discussion_reactions_target (target_type, target_id, reaction_type)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS discussion_reports (
  id VARCHAR(64) PRIMARY KEY,
  reporter_user_id VARCHAR(64) NOT NULL,
  target_type VARCHAR(32) NOT NULL,
  target_id VARCHAR(64) NOT NULL,
  reason TEXT NOT NULL,
  status VARCHAR(32) NOT NULL,
  review_note TEXT NULL,
  created_at DATETIME NOT NULL,
  updated_at DATETIME NOT NULL,
  KEY idx_discussion_reports_status_created (status, created_at),
  KEY idx_discussion_reports_target (target_type, target_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

INSERT INTO rbac_permissions (code, name, module, action, description, status, created_at, updated_at)
VALUES
  ('community.view', 'Community View', 'COMMUNITY', 'VIEW', 'view community moderation module', 'ACTIVE', NOW(), NOW()),
  ('community.edit', 'Community Edit', 'COMMUNITY', 'EDIT', 'edit community topics and comments', 'ACTIVE', NOW(), NOW()),
  ('community.review', 'Community Review', 'COMMUNITY', 'REVIEW', 'review community reports', 'ACTIVE', NOW(), NOW())
ON DUPLICATE KEY UPDATE
  name = VALUES(name),
  module = VALUES(module),
  action = VALUES(action),
  description = VALUES(description),
  status = VALUES(status),
  updated_at = VALUES(updated_at);

INSERT INTO rbac_role_permissions (role_id, permission_code, created_at)
SELECT 'role_super_admin', p.code, NOW()
FROM rbac_permissions p
WHERE p.code IN ('community.view', 'community.edit', 'community.review')
ON DUPLICATE KEY UPDATE created_at = VALUES(created_at);

INSERT INTO rbac_role_permissions (role_id, permission_code, created_at)
SELECT 'role_ops_admin', p.code, NOW()
FROM rbac_permissions p
WHERE p.code IN ('community.view', 'community.edit', 'community.review')
ON DUPLICATE KEY UPDATE created_at = VALUES(created_at);

INSERT INTO rbac_role_permissions (role_id, permission_code, created_at)
SELECT 'role_auditor', p.code, NOW()
FROM rbac_permissions p
WHERE p.code = 'community.view'
ON DUPLICATE KEY UPDATE created_at = VALUES(created_at);
