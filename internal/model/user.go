package model

import (
	"time"
)

type User struct {
	ID           int       `gorm:"primaryKey;autoIncrement" json:"id"`
	Username     string    `gorm:"size:50" json:"username"`
	Points       int       `gorm:"default:0;not null" json:"points"`
	RealName     string    `gorm:"size:64" json:"real_name"`
	Avatar       []byte    `gorm:"type:blob" json:"-"`
	OpenID       string    `gorm:"size:64;not null;uniqueIndex:open_id_idx" json:"-"`
	SessionKey   string    `gorm:"size:64;not null" json:"-"`
	IsDeleted    bool      `gorm:"default:false;not null" json:"-"`
	LastLoginTime time.Time `gorm:"not null" json:"-"`
	CreatedTime  time.Time `gorm:"default:CURRENT_TIMESTAMP;not null" json:"created_time"`
	UpdatedTime  time.Time `gorm:"default:CURRENT_TIMESTAMP;autoUpdateTime" json:"updated_time"`

	// 非数据库字段
	LastTags []string `gorm:"-" json:"last_tags,omitempty"`
}

func (User) TableName() string {
	return "mgtt_user"
}
