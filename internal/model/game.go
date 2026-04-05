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
	YunDong          GameType = 6 // 运动
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
	case YunDong:
		return "运动"
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
	case 6:
		return YunDong
	default:
		return PingHu
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
	SessionID int        `gorm:"not null;index:idx_session" json:"session_id"`
	Type      GameType   `gorm:"not null" json:"type"`
	Status    GameStatus `gorm:"default:0;not null" json:"status"`
	Remark    string     `gorm:"size:200" json:"remark"`
	CreatedBy int        `gorm:"not null" json:"created_by"`
	CreatedAt time.Time  `gorm:"default:CURRENT_TIMESTAMP;not null" json:"created_at"`
	SettledAt *time.Time `json:"settled_at,omitempty"`
}

func (Game) TableName() string {
	return "game"
}

// GameWithPlayers 游戏及其玩家信息
type GameWithPlayers struct {
	Game    *Game
	Players []*GamePlayer
}
