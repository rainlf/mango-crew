package model

import "time"

// GamePlayer 当前牌桌玩家
type GamePlayer struct {
	ID        int       `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID    int       `gorm:"not null;index:idx_game_player_user" json:"user_id"`
	Seat      int       `gorm:"not null" json:"seat"`
	CreatedAt time.Time `gorm:"not null;autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"not null;autoUpdateTime" json:"updated_at"`
}

func (GamePlayer) TableName() string {
	return "game_player"
}
