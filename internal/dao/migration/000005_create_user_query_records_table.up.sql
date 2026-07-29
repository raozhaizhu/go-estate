CREATE TABLE IF NOT EXISTS `user_query_records`(
        `id` INT PRIMARY KEY AUTO_INCREMENT,
        `username` VARCHAR(50) UNIQUE NOT NULL,
        `query_type` TINYINT NOT NULL,
        `start_date` DATE NOT NULL,
        `end_date` DATE NOT NULL,
        `created_at` TIMESTAMP NOT NULL DEFAULT (NOW()),
        -- 索引: 每个用户只能购买特定日期的数据, 避免重复
        UNIQUE KEY `idx_user_date`(
                `username`,
                `query_type`,
                `start_date`,
                `end_date`
        )
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4;