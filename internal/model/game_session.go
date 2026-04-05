package model

import (
	"time"
)

// GameSession 游戏场次（一局多盘）
type GameSession struct {
	ID        int        `gorm:"primaryKey;autoIncrement" json:"id"`
	Name      string     `gorm:"size:100" json:"name"`             // 场次名称，如"周五晚场"
	Status    int        `gorm:"default:0;not null" json:"status"` // 0:进行中 1:已结束
	CreatedBy int        `gorm:"not null" json:"created_by"`
	CreatedAt time.Time  `gorm:"default:CURRENT_TIMESTAMP;not null" json:"created_at"`
	EndedAt   *time.Time `json:"ended_at,omitempty"`
}

func (GameSession) TableName() string {
	return "game_session"
}
