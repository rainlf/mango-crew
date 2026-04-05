package model

// UserDTO 用户数据传输对象
type UserDTO struct {
	ID        int    `json:"id"`
	Nickname  string `json:"nickname"`
	AvatarURL string `json:"avatar_url,omitempty"`
}

// FromUser 从 User 创建 DTO
func (dto *UserDTO) FromUser(user *User) *UserDTO {
	if user == nil {
		return nil
	}
	return &UserDTO{
		ID:        user.ID,
		Nickname:  user.Nickname,
		AvatarURL: user.AvatarURL,
	}
}

// UserWithStatsDTO 带统计信息的用户DTO
type UserWithStatsDTO struct {
	*UserDTO
	TotalPoints int      `json:"total_points"` // 实时计算的总积分
	TotalGames  int      `json:"total_games"`  // 参与游戏数
	WinCount    int      `json:"win_count"`    // 赢的次数
	Tags        []string `json:"tags,omitempty"`
}

// GameSessionDTO 场次DTO
type GameSessionDTO struct {
	ID          int      `json:"id"`
	Name        string   `json:"name"`
	Status      int      `json:"status"`
	CreatedBy   *UserDTO `json:"created_by"`
	GameCount   int      `json:"game_count"`
	PlayerCount int      `json:"player_count"`
	CreatedAt   string   `json:"created_at"`
	EndedAt     *string  `json:"ended_at,omitempty"`
}

// GameDTO 游戏记录DTO
type GameDTO struct {
	ID        int              `json:"id"`
	SessionID int              `json:"session_id"`
	Type      string           `json:"type"`
	TypeCode  int              `json:"type_code"`
	Status    int              `json:"status"`
	Remark    string           `json:"remark"`
	CreatedBy *UserDTO         `json:"created_by"`
	Players   []*GamePlayerDTO `json:"players"`
	CreatedAt string           `json:"created_at"`
	SettledAt *string          `json:"settled_at,omitempty"`
}

// GamePlayerDTO 游戏玩家DTO
type GamePlayerDTO struct {
	ID          int           `json:"id"`
	User        *UserDTO      `json:"user"`
	Seat        int           `json:"seat"`
	Role        string        `json:"role"`
	RoleCode    int           `json:"role_code"`
	BasePoints  int           `json:"base_points"`
	FinalPoints int           `json:"final_points"`
	WinTypes    []*WinTypeDTO `json:"win_types,omitempty"`
}

// WinTypeDTO 番型DTO
type WinTypeDTO struct {
	Code       string `json:"code"`
	Name       string `json:"name"`
	Multiplier int    `json:"multiplier"`
}

// PlayerSummaryDTO 玩家汇总DTO
type PlayerSummaryDTO struct {
	CurrentPlayers []*UserDTO `json:"current_players"` // 最近一场的玩家
	AllPlayers     []*UserDTO `json:"all_players"`     // 所有玩家
}

// WeixinSession 微信登录响应
type WeixinSession struct {
	OpenID     string `json:"openid"`
	SessionKey string `json:"session_key"`
	UnionID    string `json:"unionid,omitempty"`
	ErrCode    int    `json:"errcode,omitempty"`
	ErrMsg     string `json:"errmsg,omitempty"`
}

// IsValid 检查微信响应是否有效
func (s *WeixinSession) IsValid() bool {
	return s.OpenID != "" && s.ErrCode == 0
}
