-- Growth module migration for MySQL 8.x

CREATE TABLE IF NOT EXISTS browse_histories (
  id           varchar(32) PRIMARY KEY,
  user_id      varchar(32) NOT NULL,
  content_type varchar(16) NOT NULL,
  content_id   varchar(32) NOT NULL,
  source_page  varchar(64),
  viewed_at    datetime NOT NULL,
  INDEX idx_user_viewed_at (user_id, viewed_at)
);

CREATE TABLE IF NOT EXISTS recharge_records (
  id          varchar(32) PRIMARY KEY,
  user_id     varchar(32) NOT NULL,
  order_no    varchar(64) NOT NULL UNIQUE,
  amount      decimal(10,2) NOT NULL,
  pay_channel varchar(16) NOT NULL,
  status      varchar(16) NOT NULL,
  paid_at     datetime,
  remark      varchar(256),
  created_at  datetime NOT NULL,
  INDEX idx_user_created_at (user_id, created_at)
);

CREATE TABLE IF NOT EXISTS invite_links (
  id          varchar(32) PRIMARY KEY,
  user_id     varchar(32) NOT NULL,
  invite_code varchar(32) NOT NULL UNIQUE,
  url         varchar(512) NOT NULL,
  channel     varchar(64),
  status      varchar(16) NOT NULL,
  expired_at  datetime,
  created_at  datetime NOT NULL
);

CREATE TABLE IF NOT EXISTS invite_records (
  id              varchar(32) PRIMARY KEY,
  inviter_user_id varchar(32) NOT NULL,
  invitee_user_id varchar(32) NOT NULL,
  invite_link_id  varchar(32) NOT NULL,
  register_at     datetime,
  kyc_at          datetime,
  first_pay_at    datetime,
  status          varchar(16) NOT NULL,
  risk_flag       varchar(16) NOT NULL DEFAULT 'NORMAL',
  UNIQUE KEY uk_invitee_once (invitee_user_id),
  FOREIGN KEY (invite_link_id) REFERENCES invite_links(id)
);

CREATE TABLE IF NOT EXISTS share_reward_records (
  id               varchar(32) PRIMARY KEY,
  inviter_user_id  varchar(32) NOT NULL,
  invitee_user_id  varchar(32) NOT NULL,
  invite_record_id varchar(32) NOT NULL,
  reward_type      varchar(16) NOT NULL,
  reward_value     decimal(10,2) NOT NULL,
  trigger_event    varchar(32) NOT NULL,
  status           varchar(16) NOT NULL,
  issued_at        datetime,
  review_reason    varchar(256),
  created_at       datetime NOT NULL,
  FOREIGN KEY (invite_record_id) REFERENCES invite_records(id),
  INDEX idx_inviter_created_at (inviter_user_id, created_at)
);

