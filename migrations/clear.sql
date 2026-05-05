truncate table game;
truncate table game_record;
update user
set total_points = 0,
    total_games  = 0,
    win_count    = 0,
    win_rate     = 0
where id > 0;

