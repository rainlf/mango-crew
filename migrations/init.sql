-- Mango Crew 数据库初始化脚本
-- 创建数据库
CREATE DATABASE IF NOT EXISTS mango_crew 
    CHARACTER SET utf8mb4 
    COLLATE utf8mb4_unicode_ci;

USE mango_crew;

-- 用户表
CREATE TABLE IF NOT EXISTS `user` (
    `id` INT UNSIGNED AUTO_INCREMENT PRIMARY KEY COMMENT '用户ID',
    `nickname` VARCHAR(50) DEFAULT '' COMMENT '昵称',
    `avatar_url` VARCHAR(255) DEFAULT '' COMMENT '头像URL',
    `remark` VARCHAR(200) DEFAULT '' COMMENT '备注',
    `open_id` VARCHAR(64) NOT NULL COMMENT '微信OpenID',
    `session_key` VARCHAR(64) NOT NULL COMMENT '微信SessionKey',
    `total_points` INT NOT NULL DEFAULT 0 COMMENT '总积分',
    `total_games` INT NOT NULL DEFAULT 0 COMMENT '总场次',
    `win_count` INT NOT NULL DEFAULT 0 COMMENT '赢的场次',
    `win_rate` DECIMAL(8,4) NOT NULL DEFAULT 0 COMMENT '胜率，范围 0-1',
    `created_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL COMMENT '创建时间',
    `updated_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP NOT NULL COMMENT '更新时间',
    UNIQUE KEY `idx_open_id` (`open_id`),
    KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户表';

-- 游戏记录表
CREATE TABLE IF NOT EXISTS `game` (
    `id` INT UNSIGNED AUTO_INCREMENT PRIMARY KEY COMMENT '游戏ID',
    `type` TINYINT NOT NULL COMMENT '类型: 1-平胡 2-自摸 3-一炮双响 4-一炮三响 5-相公',
    `status` TINYINT DEFAULT 0 NOT NULL COMMENT '状态: 0-进行中 1-已结算 2-已取消',
    `remark` VARCHAR(200) DEFAULT '' COMMENT '备注',
    `created_by` INT UNSIGNED NOT NULL COMMENT '创建者ID',
    `created_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL COMMENT '创建时间',
    `settled_at` TIMESTAMP NULL DEFAULT NULL COMMENT '结算时间',
    KEY `idx_status` (`status`),
    KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='游戏记录表';

-- 当前牌桌玩家表
CREATE TABLE IF NOT EXISTS `game_player` (
    `id` INT UNSIGNED AUTO_INCREMENT PRIMARY KEY COMMENT '记录ID',
    `user_id` INT UNSIGNED NOT NULL COMMENT '用户ID',
    `seat` TINYINT NOT NULL COMMENT '当前位置 1-4',
    `created_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL COMMENT '创建时间',
    `updated_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP NOT NULL COMMENT '更新时间',
    KEY `idx_game_player_user` (`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='当前牌桌玩家表';

-- 对局记录表
CREATE TABLE IF NOT EXISTS `game_record` (
    `id` INT UNSIGNED AUTO_INCREMENT PRIMARY KEY COMMENT '记录ID',
    `game_id` INT UNSIGNED NOT NULL COMMENT '游戏ID',
    `user_id` INT UNSIGNED NOT NULL COMMENT '用户ID',
    `seat` TINYINT NOT NULL COMMENT '座位号 1-4',
    `role` TINYINT NOT NULL COMMENT '角色: 1-赢家 2-输家 3-记录者 4-参与者',
    `base_points` INT DEFAULT 0 COMMENT '基础分',
    `final_points` INT DEFAULT 0 COMMENT '最终分数',
    `win_types` TEXT NULL COMMENT '赢家番型JSON，非赢家可为空',
    `is_settled` TINYINT(1) DEFAULT 0 NOT NULL COMMENT '是否已结算',
    `created_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL COMMENT '创建时间',
    `updated_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP NOT NULL COMMENT '更新时间',
    KEY `idx_game_record_game` (`game_id`),
    KEY `idx_game_record_user` (`user_id`),
    KEY `idx_game_record_game_user` (`game_id`, `user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='对局记录表';
