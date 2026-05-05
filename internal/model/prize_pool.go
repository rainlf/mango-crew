package model

import "time"

const PrizePoolTypeRecorder = "recorder"

// PrizePool 奖池余额
type PrizePool struct {
	ID        int       `gorm:"primaryKey;autoIncrement" json:"id"`
	PoolType  string    `gorm:"size:32;not null;uniqueIndex:idx_prize_pool_type" json:"pool_type"`
	Balance   int       `gorm:"not null;default:0" json:"balance"`
	CreatedAt time.Time `gorm:"not null;autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"not null;autoUpdateTime" json:"updated_at"`
}

func (PrizePool) TableName() string {
	return "prize_pool"
}
