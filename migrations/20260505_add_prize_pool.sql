USE mango_crew;

-- 奖池表
CREATE TABLE IF NOT EXISTS `prize_pool` (
    `id` INT UNSIGNED AUTO_INCREMENT PRIMARY KEY COMMENT '奖池ID',
    `pool_type` VARCHAR(32) NOT NULL COMMENT '奖池类型',
    `balance` INT NOT NULL DEFAULT 0 COMMENT '当前奖池余额',
    `created_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL COMMENT '创建时间',
    `updated_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP NOT NULL COMMENT '更新时间',
    UNIQUE KEY `idx_prize_pool_type` (`pool_type`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='奖池表';

-- 初始化记录人奖池
INSERT INTO `prize_pool` (`pool_type`, `balance`)
VALUES ('recorder', 0)
ON DUPLICATE KEY UPDATE
    `balance` = `balance`;
