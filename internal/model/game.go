package model

import (
	"time"
)

// GameType 游戏类型
type GameType int

const (
	PingHu           GameType = 1 // 平胡
	ZiMo             GameType = 2 // 自摸
	YiPaoShuangXiang GameType = 3 // 一炮双响
	YiPaoSanXiang    GameType = 4 // 一炮三响
	XiangGong        GameType = 5 // 相公
)

func (t GameType) Name() string {
	switch t {
	case PingHu:
		return "平胡"
	case ZiMo:
		return "自摸"
	case YiPaoShuangXiang:
		return "一炮双响"
	case YiPaoSanXiang:
		return "一炮三响"
	case XiangGong:
		return "相公"
	default:
		return "未知"
	}
}

func GameTypeFromCode(code int) GameType {
	switch code {
	case 1:
		return PingHu
	case 2:
		return ZiMo
	case 3:
		return YiPaoShuangXiang
	case 4:
		return YiPaoSanXiang
	case 5:
		return XiangGong
	default:
		return GameType(code)
	}
}

// GameStatus 游戏状态
type GameStatus int

const (
	GameStatusPending  GameStatus = 0 // 待确认
	GameStatusSettled  GameStatus = 1 // 已确认
	GameStatusCanceled GameStatus = 2 // 已取消
)

// Game 游戏记录（一盘）
type Game struct {
	ID        int        `gorm:"primaryKey;autoIncrement" json:"id"`
	Type      GameType   `gorm:"not null" json:"type"`
	Status    GameStatus `gorm:"default:0;not null" json:"status"`
	Remark    string     `gorm:"size:200" json:"remark"`
	CreatedBy int        `gorm:"not null" json:"created_by"`
	CreatedAt time.Time  `gorm:"not null;autoCreateTime" json:"created_at"`
	SettledAt *time.Time `json:"settled_at,omitempty"`
}

func (Game) TableName() string {
	return "game"
}

// GameWithRecords 游戏及其对局记录信息
type GameWithRecords struct {
	Game    *Game
	Records []*GameRecord
}
