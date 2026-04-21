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
    `created_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL COMMENT '创建时间',
    `updated_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP NOT NULL COMMENT '更新时间',
    UNIQUE KEY `idx_open_id` (`open_id`),
    KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户表';

-- 游戏场次表
CREATE TABLE IF NOT EXISTS `game_session` (
    `id` INT UNSIGNED AUTO_INCREMENT PRIMARY KEY COMMENT '场次ID',
    `name` VARCHAR(100) DEFAULT '' COMMENT '场次名称',
    `status` TINYINT DEFAULT 0 NOT NULL COMMENT '状态: 0-进行中 1-已结束',
    `created_by` INT UNSIGNED NOT NULL COMMENT '创建者ID',
    `created_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL COMMENT '创建时间',
    `ended_at` TIMESTAMP NULL DEFAULT NULL COMMENT '结束时间',
    KEY `idx_status` (`status`),
    KEY `idx_created_by` (`created_by`),
    KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='游戏场次表';

-- 游戏记录表
CREATE TABLE IF NOT EXISTS `game` (
    `id` INT UNSIGNED AUTO_INCREMENT PRIMARY KEY COMMENT '游戏ID',
    `session_id` INT UNSIGNED NOT NULL COMMENT '场次ID',
    `type` TINYINT NOT NULL COMMENT '类型: 1-平胡 2-自摸 3-一炮双响 4-一炮三响 5-相公 6-运动',
    `status` TINYINT DEFAULT 0 NOT NULL COMMENT '状态: 0-进行中 1-已结算 2-已取消',
    `remark` VARCHAR(200) DEFAULT '' COMMENT '备注',
    `created_by` INT UNSIGNED NOT NULL COMMENT '创建者ID',
    `created_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL COMMENT '创建时间',
    `settled_at` TIMESTAMP NULL DEFAULT NULL COMMENT '结算时间',
    KEY `idx_session` (`session_id`),
    KEY `idx_status` (`status`),
    KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='游戏记录表';

-- 当前牌桌玩家表
CREATE TABLE IF NOT EXISTS `session_player` (
    `id` INT UNSIGNED AUTO_INCREMENT PRIMARY KEY COMMENT '记录ID',
    `session_id` INT UNSIGNED NOT NULL COMMENT '场次ID',
    `user_id` INT UNSIGNED NOT NULL COMMENT '用户ID',
    `seat` TINYINT NOT NULL COMMENT '当前位置 1-4',
    `created_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL COMMENT '创建时间',
    `updated_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP NOT NULL COMMENT '更新时间',
    KEY `idx_session_player_session` (`session_id`),
    KEY `idx_session_player_user` (`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='当前牌桌玩家表';

-- 游戏玩家表
CREATE TABLE IF NOT EXISTS `game_player` (
    `id` INT UNSIGNED AUTO_INCREMENT PRIMARY KEY COMMENT '记录ID',
    `game_id` INT UNSIGNED NOT NULL COMMENT '游戏ID',
    `user_id` INT UNSIGNED NOT NULL COMMENT '用户ID',
    `seat` TINYINT NOT NULL COMMENT '座位号 1-4',
    `role` TINYINT NOT NULL COMMENT '角色: 1-赢家 2-输家 3-记录者 4-参与者',
    `base_points` INT DEFAULT 0 COMMENT '基础分',
    `final_points` INT DEFAULT 0 COMMENT '最终分数',
    `is_settled` TINYINT(1) DEFAULT 0 NOT NULL COMMENT '是否已结算',
    `created_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL COMMENT '创建时间',
    `updated_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP NOT NULL COMMENT '更新时间',
    KEY `idx_game` (`game_id`),
    KEY `idx_user` (`user_id`),
    KEY `idx_game_user` (`game_id`, `user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='游戏玩家表';

-- 玩家番型记录表
CREATE TABLE IF NOT EXISTS `game_player_win_type` (
    `id` INT UNSIGNED AUTO_INCREMENT PRIMARY KEY COMMENT '记录ID',
    `game_player_id` INT UNSIGNED NOT NULL COMMENT '游戏玩家记录ID',
    `win_type_code` VARCHAR(20) NOT NULL COMMENT '番型代码',
    `multiplier` INT NOT NULL COMMENT '倍数',
    KEY `idx_game_player` (`game_player_id`),
    KEY `idx_win_type_code` (`win_type_code`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='玩家番型记录表';

-- 番型字典表
CREATE TABLE IF NOT EXISTS `win_type` (
    `code` VARCHAR(20) PRIMARY KEY COMMENT '番型代码',
    `name` VARCHAR(20) NOT NULL COMMENT '番型名称',
    `base_multi` INT NOT NULL COMMENT '基础倍数',
    `description` VARCHAR(100) DEFAULT '' COMMENT '描述'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='番型字典表';

-- 插入默认番型数据
INSERT INTO `win_type` (`code`, `name`, `base_multi`, `description`) VALUES
('wu_hua_guo', '无花果', 1, '无番型'),
('peng_peng_hu', '碰碰胡', 2, '全部由碰牌组成'),
('yi_tiao_long', '一条龙', 2, '同一花色1-9'),
('hun_yi_se', '混一色', 2, '同一花色加字牌'),
('qing_yi_se', '清一色', 4, '同一花色'),
('xiao_qi_dui', '小七对', 4, '七个对子'),
('long_qi_dui', '龙七对', 8, '小七对加一根'),
('da_diao_che', '大吊车', 2, '单吊将牌'),
('men_qian_qing', '门前清', 2, '未碰未吃'),
('gang_kai_hua', '杠开花', 2, '杠牌后自摸');
