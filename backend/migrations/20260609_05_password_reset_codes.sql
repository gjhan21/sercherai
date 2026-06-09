CREATE TABLE IF NOT EXISTS `password_reset_codes` (
    `id` VARCHAR(64) NOT NULL,
    `email` VARCHAR(128) NOT NULL,
    `code` VARCHAR(10) NOT NULL,
    `status` VARCHAR(20) NOT NULL DEFAULT 'UNUSED', -- UNUSED, USED
    `expired_at` TIMESTAMP NOT NULL,
    `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    INDEX `idx_email_code` (`email`, `code`, `status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
