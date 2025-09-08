create database mgtt;
use mgtt;

create table mgtt_config
(
    id           int auto_increment
        primary key,
    `key`        varchar(512)                       not null,
    value        varchar(512)                       not null,
    created_time datetime default CURRENT_TIMESTAMP not null,
    updated_time datetime default CURRENT_TIMESTAMP not null on update CURRENT_TIMESTAMP
);

create table mgtt_log
(
    id           int auto_increment
        primary key,
    level        varchar(16)                        null,
    thread       varchar(128)                       null,
    message      text                               null,
    stack_trace  text                               null,
    created_time datetime default CURRENT_TIMESTAMP not null,
    updated_time datetime default CURRENT_TIMESTAMP not null on update CURRENT_TIMESTAMP
);

create table mgtt_user
(
    id              int auto_increment
        primary key,
    username        varchar(50)                        null,
    points          int      default 0                 not null,
    real_name       varchar(64)                        null,
    avatar          blob                               null,
    open_id         varchar(64)                        not null,
    session_key     varchar(64)                        not null,
    is_deleted      tinyint  default 0                 not null,
    last_login_time datetime                           not null,
    created_time    datetime default CURRENT_TIMESTAMP not null,
    updated_time    datetime default CURRENT_TIMESTAMP not null on update CURRENT_TIMESTAMP,
    constraint open_id_idx
        unique (open_id)
);

create table mgtt_majiang_game
(
    id           int auto_increment
        primary key,
    type         int                                not null,
    player1      int                                not null,
    player2      int                                not null,
    player3      int                                not null,
    player4      int                                not null,
    is_deleted   tinyint  default 0                 not null,
    created_time datetime default CURRENT_TIMESTAMP not null,
    updated_time datetime default CURRENT_TIMESTAMP not null on update CURRENT_TIMESTAMP
);

create table mgtt_majiang_game_item
(
    id           int auto_increment
        primary key,
    game_id      int                                not null,
    user_id      int                                not null,
    type         int                                not null,
    base_point   int      default 0                 null,
    win_types    varchar(512)                       null,
    multi        int                                null,
    points       int      default 0                 not null,
    created_time datetime default CURRENT_TIMESTAMP not null,
    updated_time datetime default CURRENT_TIMESTAMP not null on update CURRENT_TIMESTAMP
);

create table mgtt_api_monitor
(
    id              bigint auto_increment
        primary key,
    uri             varchar(500)                       not null,
    http_method     varchar(10)                        not null,
    class_method    varchar(255)                       not null,
    request_params  text                               null,
    response_body   text                               null,
    status_code     int                                not null,
    cost_time       int                                not null,
    has_exception   tinyint    default 0               not null,
    exception_stack longtext                           null,
    ip_address      varchar(45)                        null,
    biz_id          varchar(128)                       null,
    create_time     datetime default CURRENT_TIMESTAMP not null
);