-- Mango Crew 一键初始化脚本（整合历史 migrations 的最终结构）
CREATE DATABASE IF NOT EXISTS mango_crew
    CHARACTER SET utf8mb4
    COLLATE utf8mb4_unicode_ci;

USE mango_crew;

-- 为了方便重复初始化，先清理历史/遗留表（不存在也不会报错）
DROP TABLE IF EXISTS `game_player_win_type`;
DROP TABLE IF EXISTS `win_type`;
DROP TABLE IF EXISTS `game_session`;
DROP TABLE IF EXISTS `session_player`;

-- 核心业务表（按依赖顺序先删后建）
DROP TABLE IF EXISTS `api_audit_log`;
DROP TABLE IF EXISTS `prize_pool`;
DROP TABLE IF EXISTS `game_record`;
DROP TABLE IF EXISTS `game_player`;
DROP TABLE IF EXISTS `game`;
DROP TABLE IF EXISTS `user`;

CREATE TABLE `user` (
    `id` INT AUTO_INCREMENT PRIMARY KEY COMMENT '用户ID',
    `nickname` VARCHAR(50) DEFAULT '' COMMENT '昵称',
    `avatar_url` VARCHAR(255) DEFAULT '' COMMENT '头像URL',
    `remark` VARCHAR(200) DEFAULT '' COMMENT '备注',
    `open_id` VARCHAR(64) NOT NULL COMMENT '微信OpenID',
    `session_key` VARCHAR(64) NOT NULL COMMENT '微信SessionKey',
    `total_points` INT NOT NULL DEFAULT 0 COMMENT '总积分',
    `total_games` INT NOT NULL DEFAULT 0 COMMENT '总场次',
    `win_count` INT NOT NULL DEFAULT 0 COMMENT '赢的场次',
    `win_rate` DECIMAL(8,4) NOT NULL DEFAULT 0 COMMENT '胜率，范围 0-1',
    `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    UNIQUE KEY `idx_open_id` (`open_id`),
    KEY `idx_user_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户表';

CREATE TABLE `game` (
    `id` INT AUTO_INCREMENT PRIMARY KEY COMMENT '游戏ID',
    `type` TINYINT NOT NULL COMMENT '类型: 1-平胡 2-自摸 3-一炮双响 4-一炮三响 5-相公 6-深蹲兑换',
    `status` TINYINT NOT NULL DEFAULT 0 COMMENT '状态: 0-待确认 1-已结算 2-已取消',
    `remark` VARCHAR(200) DEFAULT '' COMMENT '备注',
    `created_by` INT NOT NULL COMMENT '创建者ID',
    `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `settled_at` TIMESTAMP NULL DEFAULT NULL COMMENT '结算时间',
    KEY `idx_game_status` (`status`),
    KEY `idx_game_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='游戏记录表';

CREATE TABLE `game_player` (
    `id` INT AUTO_INCREMENT PRIMARY KEY COMMENT '记录ID',
    `user_id` INT NOT NULL COMMENT '用户ID',
    `seat` TINYINT NOT NULL COMMENT '当前位置 1-4',
    `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    KEY `idx_game_player_user` (`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='当前牌桌玩家表';

CREATE TABLE `game_record` (
    `id` INT AUTO_INCREMENT PRIMARY KEY COMMENT '记录ID',
    `game_id` INT NOT NULL COMMENT '游戏ID',
    `user_id` INT NOT NULL COMMENT '用户ID',
    `seat` TINYINT NOT NULL COMMENT '座位号 1-4',
    `role` TINYINT NOT NULL COMMENT '角色: 1-赢家 2-输家 3-记录者 4-参与者 5-深蹲兑换',
    `base_points` INT NOT NULL DEFAULT 0 COMMENT '基础分',
    `final_points` INT NOT NULL DEFAULT 0 COMMENT '最终分数',
    `is_settled` TINYINT(1) NOT NULL DEFAULT 0 COMMENT '是否已结算',
    `win_types` TEXT NULL COMMENT '赢家番型JSON，非赢家可为空',
    `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    KEY `idx_game_record_game` (`game_id`),
    KEY `idx_game_record_user` (`user_id`),
    KEY `idx_game_record_game_user` (`game_id`, `user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='对局记录表';

CREATE TABLE `prize_pool` (
    `id` INT AUTO_INCREMENT PRIMARY KEY COMMENT '奖池ID',
    `pool_type` VARCHAR(32) NOT NULL COMMENT '奖池类型',
    `balance` INT NOT NULL DEFAULT 0 COMMENT '当前奖池余额',
    `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    UNIQUE KEY `idx_prize_pool_type` (`pool_type`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='奖池表';

INSERT INTO `prize_pool` (`pool_type`, `balance`)
VALUES ('recorder', 0)
ON DUPLICATE KEY UPDATE
    `balance` = `balance`;

CREATE TABLE `api_audit_log` (
    `id` INT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    `user_id` INT NULL,
    `http_method` VARCHAR(16) NOT NULL,
    `path` VARCHAR(1024) NOT NULL,
    `http_status` INT NOT NULL,
    `latency_ms` BIGINT NOT NULL DEFAULT 0,
    `client_ip` VARCHAR(64) NULL,
    `user_agent` VARCHAR(255) NULL,
    `request` LONGTEXT NULL,
    `response` LONGTEXT NULL,
    `error` TEXT NULL,
    `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='API审计日志表';
