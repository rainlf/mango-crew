USE mango_crew;

-- 将赢家番型数据迁移到 game_player.win_types(JSON) 中。
ALTER TABLE `game_player`
    ADD COLUMN `win_types` TEXT NULL COMMENT '赢家番型JSON，非赢家可为空' AFTER `final_points`;

UPDATE `game_player` gp
LEFT JOIN (
    SELECT
        game_player_id,
        CONCAT(
            '[',
            GROUP_CONCAT(
                JSON_OBJECT(
                    'win_type_code', win_type_code,
                    'multiplier', multiplier
                )
                ORDER BY id ASC SEPARATOR ','
            ),
            ']'
        ) AS win_types_json
    FROM `game_player_win_type`
    GROUP BY game_player_id
) wt ON wt.game_player_id = gp.id
SET gp.win_types = wt.win_types_json
WHERE wt.win_types_json IS NOT NULL;

DROP TABLE IF EXISTS `game_player_win_type`;
