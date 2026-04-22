USE mango_crew;

-- 先删除依赖 game_session 的字段和索引，再删除场次表本身。
ALTER TABLE `game`
    DROP INDEX `idx_session`,
    DROP COLUMN `session_id`;

ALTER TABLE `session_player`
    DROP INDEX `idx_session_player_session`,
    DROP COLUMN `session_id`;

DROP TABLE IF EXISTS `game_session`;
