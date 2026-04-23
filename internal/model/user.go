package model

import (
	"time"
)

// User 用户模型
type User struct {
	ID          int       `gorm:"primaryKey;autoIncrement" json:"id"`
	Nickname    string    `gorm:"size:50" json:"nickname"`
	AvatarURL   string    `gorm:"size:255" json:"avatar_url"`
	Remark      string    `gorm:"size:200" json:"remark"` // 备注
	OpenID      string    `gorm:"size:64;not null;uniqueIndex:idx_open_id" json:"-"`
	SessionKey  string    `gorm:"size:64;not null" json:"-"`
	TotalPoints int       `gorm:"not null;default:0" json:"total_points"`
	TotalGames  int       `gorm:"not null;default:0" json:"total_games"`
	WinCount    int       `gorm:"not null;default:0" json:"win_count"`
	WinRate     float64   `gorm:"type:decimal(8,4);not null;default:0" json:"win_rate"`
	CreatedAt   time.Time `gorm:"not null;autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time `gorm:"not null;autoUpdateTime" json:"updated_at"`
}

func (User) TableName() string {
	return "user"
}
