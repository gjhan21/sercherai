INSERT INTO system_configs (id, config_key, config_value, description, updated_by, updated_at)
VALUES
  ('cfg_payment_yolkpay_enabled', 'payment.channel.yolkpay.enabled', 'false', '蛋黄支付通道开关', 'system', NOW()),
  ('cfg_payment_yolkpay_pid', 'payment.channel.yolkpay.pid', '', '蛋黄支付商户ID', 'system', NOW()),
  ('cfg_payment_yolkpay_key', 'payment.channel.yolkpay.key', '', '蛋黄支付商户密钥', 'system', NOW()),
  ('cfg_payment_yolkpay_gateway', 'payment.channel.yolkpay.gateway', 'https://www.yolkpay.net', '蛋黄支付网关域名', 'system', NOW()),
  ('cfg_payment_yolkpay_mapi_path', 'payment.channel.yolkpay.mapi_path', '/mapi.php', '蛋黄支付下单路径', 'system', NOW()),
  ('cfg_payment_yolkpay_notify_url', 'payment.channel.yolkpay.notify_url', '', '蛋黄支付异步通知地址', 'system', NOW()),
  ('cfg_payment_yolkpay_return_url', 'payment.channel.yolkpay.return_url', '', '蛋黄支付同步跳转地址', 'system', NOW()),
  ('cfg_payment_yolkpay_pay_type', 'payment.channel.yolkpay.pay_type', 'airpay', '蛋黄支付类型', 'system', NOW()),
  ('cfg_payment_yolkpay_device', 'payment.channel.yolkpay.device', 'pc', '蛋黄支付设备类型', 'system', NOW())
ON DUPLICATE KEY UPDATE
  description = VALUES(description),
  updated_by = VALUES(updated_by),
  updated_at = VALUES(updated_at);
