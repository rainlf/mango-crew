package model

import "time"

// SessionPlayer 当前牌桌玩家
type SessionPlayer struct {
	ID        int       `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID    int       `gorm:"not null;index:idx_session_player_user" json:"user_id"`
	Seat      int       `gorm:"not null" json:"seat"`
	CreatedAt time.Time `gorm:"not null;autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"not null;autoUpdateTime" json:"updated_at"`
}

func (SessionPlayer) TableName() string {
	return "session_player"
}
