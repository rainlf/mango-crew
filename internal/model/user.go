package model

import (
	"time"
)

// User 用户模型
type User struct {
	ID         int       `gorm:"primaryKey;autoIncrement" json:"id"`
	Nickname   string    `gorm:"size:50" json:"nickname"`
	AvatarURL  string    `gorm:"size:255" json:"avatar_url"`
	OpenID     string    `gorm:"size:64;not null;uniqueIndex:idx_open_id" json:"-"`
	SessionKey string    `gorm:"size:64;not null" json:"-"`
	CreatedAt  time.Time `gorm:"default:CURRENT_TIMESTAMP;not null" json:"created_at"`
	UpdatedAt  time.Time `gorm:"default:CURRENT_TIMESTAMP;autoUpdateTime" json:"updated_at"`
}

func (User) TableName() string {
	return "user"
}
