-- Membership product/order extensions for MySQL 8.x

ALTER TABLE membership_products
  ADD COLUMN member_level varchar(16) NOT NULL DEFAULT 'VIP1',
  ADD COLUMN duration_days int NOT NULL DEFAULT 30;

ALTER TABLE membership_orders
  ADD COLUMN order_no varchar(64),
  ADD COLUMN pay_channel varchar(16),
  ADD COLUMN created_at datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  ADD COLUMN updated_at datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP;

ALTER TABLE membership_orders
  ADD UNIQUE KEY uk_membership_order_no (order_no);
