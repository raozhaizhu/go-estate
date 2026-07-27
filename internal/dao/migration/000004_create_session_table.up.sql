SET
        NAMES utf8mb4;

CREATE TABLE IF NOT EXISTS `sessions`(
        `id` VARCHAR(36) PRIMARY KEY,
        `username` VARCHAR(50) NOT NULL,
        -- `refresh_token` VARCHAR(255) NOT NULL,
        `device_id` VARCHAR(255) NOT NULL,
        `user_agent` VARCHAR(255) NOT NULL,
        `client_ip` VARCHAR(45) NOT NULL,
        `is_blocked` BOOLEAN NOT NULL DEFAULT false,
        `expires_at` TIMESTAMP NOT NULL,
        `created_at` TIMESTAMP NOT NULL DEFAULT (NOW()),
        -- 为执行DeleteExpiredSessions, 定期删除过期 token
        INDEX `idx_expires_at` (`expires_at`),
        -- 外键: users_username
        CONSTRAINT `fk_sessions_users` FOREIGN KEY (`username`) REFERENCES `users`(`username`)
);