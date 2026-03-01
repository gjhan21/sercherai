-- Default config for invite register VIP days reward

INSERT INTO system_configs (id, config_key, config_value, description, updated_by, updated_at)
SELECT
  'cfg_invite_reward_vip_days',
  'invite.register.reward.vip_days',
  '3',
  '邀请注册奖励VIP天数',
  'system',
  NOW()
WHERE NOT EXISTS (
  SELECT 1
  FROM system_configs
  WHERE config_key = 'invite.register.reward.vip_days'
);
