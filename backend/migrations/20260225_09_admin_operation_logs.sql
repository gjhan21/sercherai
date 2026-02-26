-- Admin operation logs migration for MySQL 8.x

CREATE TABLE IF NOT EXISTS admin_operation_logs (
  id               varchar(32) PRIMARY KEY,
  module           varchar(32) NOT NULL,
  action           varchar(64) NOT NULL,
  target_type      varchar(32) NOT NULL,
  target_id        varchar(64) NOT NULL,
  operator_user_id varchar(32) NOT NULL,
  before_value     varchar(512),
  after_value      varchar(512),
  reason           varchar(256),
  created_at       datetime NOT NULL,
  INDEX idx_module_action_time (module, action, created_at),
  INDEX idx_operator_time (operator_user_id, created_at),
  INDEX idx_target_time (target_type, target_id, created_at)
);
