package model

// LoginRequest 登录请求
// 实际使用 code 查询参数，此处用于文档说明

// CreateSessionRequest 创建场次请求
type CreateSessionRequest struct {
	Name string `json:"name" binding:"required,max=100"` // 场次名称
}

// EndSessionRequest 结束场次请求
type EndSessionRequest struct {
	SessionID int `json:"session_id" binding:"required"`
}

// CreateGameRequest 创建游戏请求
type CreateGameRequest struct {
	SessionID int                  `json:"session_id" binding:"required"`
	GameType  int                  `json:"game_type" binding:"required,min=1,max=6"`
	Remark    string               `json:"remark" binding:"max=200"`
	Players   []*GamePlayerRequest `json:"players" binding:"required,min=1"`
}

// GamePlayerRequest 游戏玩家请求
type GamePlayerRequest struct {
	UserID     int      `json:"user_id" binding:"required"`
	Seat       int      `json:"seat" binding:"required,min=1,max=4"`
	Role       int      `json:"role" binding:"required,min=1,max=4"` // 1:赢家 2:输家 3:记录者 4:参与者
	BasePoints int      `json:"base_points"`                         // 基础分，赢家必填
	WinTypes   []string `json:"win_types"`                           // 番型code列表
}

// SettleGameRequest 结算游戏请求
type SettleGameRequest struct {
	GameID int `json:"game_id" binding:"required"`
}

// CancelGameRequest 取消游戏请求
type CancelGameRequest struct {
	GameID int `json:"game_id" binding:"required"`
}

// UpdateUserRequest 更新用户信息请求
type UpdateUserRequest struct {
	Nickname string `json:"nickname" binding:"max=50"`
	Avatar   string `json:"avatar"` // base64编码的图片
}

// UpdateCurrentPlayersRequest 更新当前牌桌玩家
type UpdateCurrentPlayersRequest struct {
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
