CREATE TABLE IF NOT EXISTS stock_selection_publish_reviews (
  id varchar(64) NOT NULL,
  run_id varchar(64) NOT NULL,
  review_status varchar(16) NOT NULL DEFAULT 'PENDING',
  reviewer varchar(64) DEFAULT NULL,
  review_note text,
  override_reason text,
  publish_id varchar(64) DEFAULT NULL,
  publish_version int NOT NULL DEFAULT 0,
  approved_at datetime DEFAULT NULL,
  rejected_at datetime DEFAULT NULL,
  created_at datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  UNIQUE KEY uk_stock_selection_publish_reviews_run_id (run_id),
  KEY idx_stock_selection_publish_reviews_status (review_status, updated_at),
  CONSTRAINT fk_stock_selection_publish_reviews_run_id FOREIGN KEY (run_id) REFERENCES stock_selection_runs (run_id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

INSERT INTO rbac_permissions (code, name, module, action, description, status, created_at, updated_at)
VALUES
  ('stock_selection.view', 'Stock Selection View', 'STOCK_SELECTION', 'VIEW', 'view stock selection module', 'ACTIVE', NOW(), NOW()),
  ('stock_selection.manage', 'Stock Selection Manage', 'STOCK_SELECTION', 'MANAGE', 'manage stock selection profiles, runs and reviews', 'ACTIVE', NOW(), NOW()),
  ('stock_selection.run', 'Stock Selection Run', 'STOCK_SELECTION', 'RUN', 'trigger stock selection runs', 'ACTIVE', NOW(), NOW()),
  ('stock_selection.profile.view', 'Stock Selection Profile View', 'STOCK_SELECTION', 'PROFILE_VIEW', 'view stock selection profiles', 'ACTIVE', NOW(), NOW()),
  ('stock_selection.profile.manage', 'Stock Selection Profile Manage', 'STOCK_SELECTION', 'PROFILE_MANAGE', 'manage stock selection profiles', 'ACTIVE', NOW(), NOW()),
  ('stock_selection.review', 'Stock Selection Review', 'STOCK_SELECTION', 'REVIEW', 'review stock selection publish decisions', 'ACTIVE', NOW(), NOW()),
  ('stock_selection.publish', 'Stock Selection Publish', 'STOCK_SELECTION', 'PUBLISH', 'publish stock selection outputs', 'ACTIVE', NOW(), NOW())
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
WHERE p.code IN (
  'stock_selection.view',
  'stock_selection.manage',
  'stock_selection.run',
  'stock_selection.profile.view',
  'stock_selection.profile.manage',
  'stock_selection.review',
  'stock_selection.publish'
)
ON DUPLICATE KEY UPDATE created_at = VALUES(created_at);

INSERT INTO rbac_role_permissions (role_id, permission_code, created_at)
SELECT 'role_ops_admin', p.code, NOW()
FROM rbac_permissions p
WHERE p.code IN (
  'stock_selection.view',
  'stock_selection.manage',
  'stock_selection.run',
  'stock_selection.profile.view',
  'stock_selection.profile.manage',
  'stock_selection.review',
  'stock_selection.publish'
)
ON DUPLICATE KEY UPDATE created_at = VALUES(created_at);

INSERT INTO rbac_role_permissions (role_id, permission_code, created_at)
SELECT 'role_auditor', p.code, NOW()
FROM rbac_permissions p
WHERE p.code IN (
  'stock_selection.view',
  'stock_selection.profile.view'
)
ON DUPLICATE KEY UPDATE created_at = VALUES(created_at);
