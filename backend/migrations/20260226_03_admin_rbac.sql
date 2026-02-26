-- Admin RBAC schema and baseline data

CREATE TABLE IF NOT EXISTS rbac_permissions (
  code        varchar(64) PRIMARY KEY,
  name        varchar(64) NOT NULL,
  module      varchar(32) NOT NULL,
  action      varchar(32) NOT NULL,
  description varchar(255),
  status      varchar(16) NOT NULL DEFAULT 'ACTIVE',
  created_at  datetime NOT NULL,
  updated_at  datetime NOT NULL
);

CREATE TABLE IF NOT EXISTS rbac_roles (
  id          varchar(32) PRIMARY KEY,
  role_key    varchar(64) NOT NULL UNIQUE,
  role_name   varchar(64) NOT NULL,
  description varchar(255),
  status      varchar(16) NOT NULL DEFAULT 'ACTIVE',
  built_in    tinyint(1) NOT NULL DEFAULT 0,
  created_at  datetime NOT NULL,
  updated_at  datetime NOT NULL
);

CREATE TABLE IF NOT EXISTS rbac_role_permissions (
  role_id         varchar(32) NOT NULL,
  permission_code varchar(64) NOT NULL,
  created_at      datetime NOT NULL,
  PRIMARY KEY (role_id, permission_code),
  FOREIGN KEY (role_id) REFERENCES rbac_roles(id),
  FOREIGN KEY (permission_code) REFERENCES rbac_permissions(code)
);

CREATE TABLE IF NOT EXISTS rbac_user_roles (
  user_id    varchar(32) NOT NULL,
  role_id    varchar(32) NOT NULL,
  created_at datetime NOT NULL,
  PRIMARY KEY (user_id, role_id),
  FOREIGN KEY (user_id) REFERENCES users(id),
  FOREIGN KEY (role_id) REFERENCES rbac_roles(id)
);

INSERT INTO rbac_permissions (code, name, module, action, description, status, created_at, updated_at)
VALUES
  ('dashboard.view', 'Dashboard View', 'DASHBOARD', 'VIEW', 'view admin dashboard metrics', 'ACTIVE', NOW(), NOW()),
  ('users.view', 'Users View', 'USERS', 'VIEW', 'view user management module', 'ACTIVE', NOW(), NOW()),
  ('users.edit', 'Users Edit', 'USERS', 'EDIT', 'edit user status, level and kyc', 'ACTIVE', NOW(), NOW()),
  ('membership.view', 'Membership View', 'MEMBERSHIP', 'VIEW', 'view membership products and orders', 'ACTIVE', NOW(), NOW()),
  ('membership.edit', 'Membership Edit', 'MEMBERSHIP', 'EDIT', 'edit membership products, orders and quotas', 'ACTIVE', NOW(), NOW()),
  ('news.view', 'News View', 'NEWS', 'VIEW', 'view news content module', 'ACTIVE', NOW(), NOW()),
  ('news.edit', 'News Edit', 'NEWS', 'EDIT', 'create and edit news content', 'ACTIVE', NOW(), NOW()),
  ('market.view', 'Market View', 'MARKET', 'VIEW', 'view stock and futures strategy modules', 'ACTIVE', NOW(), NOW()),
  ('market.edit', 'Market Edit', 'MARKET', 'EDIT', 'edit stock and futures strategy modules', 'ACTIVE', NOW(), NOW()),
  ('review.view', 'Review View', 'REVIEW', 'VIEW', 'view review center', 'ACTIVE', NOW(), NOW()),
  ('review.edit', 'Review Edit', 'REVIEW', 'EDIT', 'submit/assign/decision review tasks', 'ACTIVE', NOW(), NOW()),
  ('risk.view', 'Risk View', 'RISK', 'VIEW', 'view risk center modules', 'ACTIVE', NOW(), NOW()),
  ('risk.edit', 'Risk Edit', 'RISK', 'EDIT', 'edit risk rules and review risk hits', 'ACTIVE', NOW(), NOW()),
  ('audit.view', 'Audit View', 'AUDIT', 'VIEW', 'view admin operation logs', 'ACTIVE', NOW(), NOW()),
  ('auth_security.view', 'Auth Security View', 'AUTH_SECURITY', 'VIEW', 'view auth security logs and configs', 'ACTIVE', NOW(), NOW()),
  ('auth_security.edit', 'Auth Security Edit', 'AUTH_SECURITY', 'EDIT', 'update auth security configs and unlock', 'ACTIVE', NOW(), NOW()),
  ('system_config.view', 'System Config View', 'SYSTEM_CONFIG', 'VIEW', 'view system configs', 'ACTIVE', NOW(), NOW()),
  ('system_config.edit', 'System Config Edit', 'SYSTEM_CONFIG', 'EDIT', 'edit system configs', 'ACTIVE', NOW(), NOW()),
  ('system_job.view', 'System Job View', 'SYSTEM_JOB', 'VIEW', 'view scheduler jobs and runs', 'ACTIVE', NOW(), NOW()),
  ('system_job.edit', 'System Job Edit', 'SYSTEM_JOB', 'EDIT', 'edit and trigger scheduler jobs', 'ACTIVE', NOW(), NOW()),
  ('data_source.view', 'Data Source View', 'DATA_SOURCE', 'VIEW', 'view data source status and logs', 'ACTIVE', NOW(), NOW()),
  ('data_source.edit', 'Data Source Edit', 'DATA_SOURCE', 'EDIT', 'manage data sources and health checks', 'ACTIVE', NOW(), NOW()),
  ('workflow.view', 'Workflow View', 'WORKFLOW', 'VIEW', 'view workflow messages', 'ACTIVE', NOW(), NOW()),
  ('workflow.edit', 'Workflow Edit', 'WORKFLOW', 'EDIT', 'update workflow message read status', 'ACTIVE', NOW(), NOW()),
  ('growth.view', 'Growth View', 'GROWTH', 'VIEW', 'view growth referral and reward modules', 'ACTIVE', NOW(), NOW()),
  ('growth.edit', 'Growth Edit', 'GROWTH', 'EDIT', 'edit growth referral and reward modules', 'ACTIVE', NOW(), NOW()),
  ('payment.view', 'Payment View', 'PAYMENT', 'VIEW', 'view reconciliation module', 'ACTIVE', NOW(), NOW()),
  ('payment.edit', 'Payment Edit', 'PAYMENT', 'EDIT', 'trigger reconciliation retry', 'ACTIVE', NOW(), NOW()),
  ('reward_wallet.view', 'Reward Wallet View', 'REWARD_WALLET', 'VIEW', 'view reward wallet withdraw requests', 'ACTIVE', NOW(), NOW()),
  ('reward_wallet.edit', 'Reward Wallet Edit', 'REWARD_WALLET', 'EDIT', 'review reward wallet withdraw requests', 'ACTIVE', NOW(), NOW()),
  ('access.view', 'Access View', 'ACCESS', 'VIEW', 'view administrator and permission settings', 'ACTIVE', NOW(), NOW()),
  ('access.edit', 'Access Edit', 'ACCESS', 'EDIT', 'manage administrator and permission settings', 'ACTIVE', NOW(), NOW())
ON DUPLICATE KEY UPDATE
  name = VALUES(name),
  module = VALUES(module),
  action = VALUES(action),
  description = VALUES(description),
  status = VALUES(status),
  updated_at = VALUES(updated_at);

INSERT INTO rbac_roles (id, role_key, role_name, description, status, built_in, created_at, updated_at)
VALUES
  ('role_super_admin', 'SUPER_ADMIN', 'Super Admin', 'full platform access', 'ACTIVE', 1, NOW(), NOW()),
  ('role_ops_admin', 'OPS_ADMIN', 'Operations Admin', 'operations with limited access control rights', 'ACTIVE', 1, NOW(), NOW()),
  ('role_auditor', 'AUDITOR', 'Auditor', 'read focused audit role', 'ACTIVE', 1, NOW(), NOW())
ON DUPLICATE KEY UPDATE
  role_name = VALUES(role_name),
  description = VALUES(description),
  status = VALUES(status),
  built_in = VALUES(built_in),
  updated_at = VALUES(updated_at);

INSERT INTO rbac_role_permissions (role_id, permission_code, created_at)
SELECT 'role_super_admin', p.code, NOW()
FROM rbac_permissions p
ON DUPLICATE KEY UPDATE created_at = VALUES(created_at);

INSERT INTO rbac_role_permissions (role_id, permission_code, created_at)
SELECT 'role_ops_admin', p.code, NOW()
FROM rbac_permissions p
WHERE p.code <> 'access.edit'
ON DUPLICATE KEY UPDATE created_at = VALUES(created_at);

INSERT INTO rbac_role_permissions (role_id, permission_code, created_at)
SELECT 'role_auditor', p.code, NOW()
FROM rbac_permissions p
WHERE p.code IN (
  'dashboard.view',
  'users.view',
  'membership.view',
  'news.view',
  'market.view',
  'review.view',
  'risk.view',
  'audit.view',
  'auth_security.view',
  'system_config.view',
  'system_job.view',
  'data_source.view',
  'workflow.view',
  'growth.view',
  'payment.view',
  'reward_wallet.view'
)
ON DUPLICATE KEY UPDATE created_at = VALUES(created_at);

INSERT INTO rbac_user_roles (user_id, role_id, created_at)
VALUES
  ('admin_001', 'role_super_admin', NOW()),
  ('admin_002', 'role_ops_admin', NOW())
ON DUPLICATE KEY UPDATE created_at = VALUES(created_at);
