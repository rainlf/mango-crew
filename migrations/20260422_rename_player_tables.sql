USE mango_crew;

RENAME TABLE
    `game_player` TO `game_record`,
    `session_player` TO `game_player`;

ALTER TABLE `game_player`
    RENAME INDEX `idx_session_player_user` TO `idx_game_player_user`;

ALTER TABLE `game_record`
    RENAME INDEX `idx_game` TO `idx_game_record_game`,
    RENAME INDEX `idx_user` TO `idx_game_record_user`,
    RENAME INDEX `idx_game_user` TO `idx_game_record_game_user`;
