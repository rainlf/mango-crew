package model

import (
	"time"
)

// PlayerRole 玩家角色
type PlayerRole int

const (
	RoleWinner   PlayerRole = 1 // 赢家
	RoleLoser    PlayerRole = 2 // 输家
	RoleRecorder PlayerRole = 3 // 记录者
)

func (r PlayerRole) Name() string {
	switch r {
	case RoleWinner:
		return "赢家"
	case RoleLoser:
		return "输家"
	case RoleRecorder:
		return "记录者"
	default:
		return "未知"
	}
}

// GamePlayer 游戏玩家记录
type GamePlayer struct {
	ID          int        `gorm:"primaryKey;autoIncrement" json:"id"`
	GameID      int        `gorm:"not null;index:idx_game" json:"game_id"`
	UserID      int        `gorm:"not null;index:idx_user" json:"user_id"`
	Seat        int        `gorm:"not null" json:"seat"`          // 座位号 1-4
	Role        PlayerRole `gorm:"not null" json:"role"`          // 角色
	BasePoints  int        `gorm:"default:0" json:"base_points"`  // 基础分
	FinalPoints int        `gorm:"default:0" json:"final_points"` // 最终分数
	IsSettled   bool       `gorm:"default:false;not null" json:"is_settled"`
	CreatedAt   time.Time  `gorm:"not null;autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time  `gorm:"not null;autoUpdateTime" json:"updated_at"`

	// 关联
	WinTypes []*GamePlayerWinType `gorm:"-" json:"win_types,omitempty"`
}

func (GamePlayer) TableName() string {
	return "game_player"
}

// CalculatePoints 计算最终分数（根据番型倍数）
func (gp *GamePlayer) CalculatePoints() {
	multi := 1
	for _, wt := range gp.WinTypes {
		multi *= wt.Multiplier
	}
	if multi < 1 {
		multi = 1
	}
	gp.FinalPoints = gp.BasePoints * multi
}

// GamePlayerWinType 玩家番型记录
type GamePlayerWinType struct {
	ID           int    `gorm:"primaryKey;autoIncrement" json:"id"`
	GamePlayerID int    `gorm:"not null;index:idx_game_player" json:"game_player_id"`
	WinTypeCode  string `gorm:"size:20;not null" json:"win_type_code"`
	Multiplier   int    `gorm:"not null" json:"multiplier"`
}

func (GamePlayerWinType) TableName() string {
	return "game_player_win_type"
}
