package model

// LoginRequest 登录请求
// 实际使用 code 查询参数，此处用于文档说明

// CancelGameRequest 取消游戏请求
type CancelGameRequest struct {
	GameID int `json:"game_id" binding:"required"`
}

// UpdateUserRequest 更新用户信息请求
type UpdateUserRequest struct {
	Nickname string `json:"nickname" binding:"max=50"`
	Avatar   string `json:"avatar"` // base64编码的图片
}

// RebuildUserStatsRequest 重建用户统计请求
type RebuildUserStatsRequest struct {
	UserIDs []int `json:"user_ids"`
}

// UpdateCurrentPlayersRequest 更新当前牌桌玩家
type UpdateCurrentPlayersRequest struct {
	UserID  int   `json:"user_id" binding:"required"`
	UserIDs []int `json:"user_ids" binding:"required,min=1,max=4"`
}

// RecordMaJiangGameRequest 按旧版记牌流程直接记录一局已结算对局
type RecordMaJiangGameRequest struct {
	GameType   int                       `json:"gameType" binding:"required,min=1,max=6"`
	Players    []int                     `json:"players" binding:"required,min=1"`
	RecorderID int                       `json:"recorderId" binding:"required"`
	Winners    []*RecordMaJiangWinnerDTO `json:"winners" binding:"required,min=1"`
	Losers     []int                     `json:"losers"`
	Remark     string                    `json:"remark" binding:"max=200"`
}

// RecordMaJiangWinnerDTO 旧版记牌赢家信息
type RecordMaJiangWinnerDTO struct {
	UserID     int      `json:"userId" binding:"required"`
	BasePoints int      `json:"basePoints"`
	WinTypes   []string `json:"winTypes"`
}
