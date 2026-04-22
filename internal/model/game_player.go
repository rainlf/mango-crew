package model

import (
	"encoding/json"
	"time"
)

// PlayerRole 玩家角色
type PlayerRole int

const (
	RoleWinner   PlayerRole = 1 // 赢家
	RoleLoser    PlayerRole = 2 // 输家
	RoleRecorder PlayerRole = 3 // 记录者
	RoleNeutral  PlayerRole = 4 // 参与但本局分数不变
)

func (r PlayerRole) Name() string {
	switch r {
	case RoleWinner:
		return "赢家"
	case RoleLoser:
		return "输家"
	case RoleRecorder:
		return "记录者"
	case RoleNeutral:
		return "参与者"
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
	WinTypesRaw string     `gorm:"column:win_types;type:text" json:"-"`

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

// SyncWinTypesRaw 将番型信息写入持久化字段
func (gp *GamePlayer) SyncWinTypesRaw() error {
	if len(gp.WinTypes) == 0 {
		gp.WinTypesRaw = ""
		return nil
	}

	data, err := json.Marshal(gp.WinTypes)
	if err != nil {
		return err
	}
	gp.WinTypesRaw = string(data)
	return nil
}

// LoadWinTypesFromRaw 从持久化字段恢复番型信息
func (gp *GamePlayer) LoadWinTypesFromRaw() error {
	if gp.WinTypesRaw == "" {
		gp.WinTypes = nil
		return nil
	}

	var winTypes []*GamePlayerWinType
	if err := json.Unmarshal([]byte(gp.WinTypesRaw), &winTypes); err != nil {
		return err
	}
	gp.WinTypes = winTypes
	return nil
}

// GamePlayerWinType 玩家番型信息
type GamePlayerWinType struct {
	WinTypeCode string `json:"win_type_code"`
	Multiplier  int    `json:"multiplier"`
}
