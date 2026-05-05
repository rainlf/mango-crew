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
	TotalPoints int      `json:"total_points"` // 用户总积分
	TotalGames  int      `json:"total_games"`  // 参与游戏数
	WinCount    int      `json:"win_count"`    // 赢的次数
	WinRate     float64  `json:"win_rate"`     // 胜率，范围 0-1
	Tags        []string `json:"tags,omitempty"`
}

// UserFitnessStats 用户健身统计
type UserFitnessStats struct {
	UserID      int
	TotalPoints int
	TotalGames  int
}

// FromUser 从 User 创建带统计信息 DTO
func (dto *UserWithStatsDTO) FromUser(user *User) *UserWithStatsDTO {
	if user == nil {
		return nil
	}
	return &UserWithStatsDTO{
		UserDTO:     (&UserDTO{}).FromUser(user),
		TotalPoints: user.TotalPoints,
		TotalGames:  user.TotalGames,
		WinCount:    user.WinCount,
		WinRate:     user.WinRate,
	}
}

// GameDTO 游戏记录DTO
type GameDTO struct {
	ID        int              `json:"id"`
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

// PrizePoolDTO 奖池信息
type PrizePoolDTO struct {
	PoolType string `json:"pool_type"`
	Balance  int    `json:"balance"`
}

// PrizePoolContributorDTO 奖池贡献明细
type PrizePoolContributorDTO struct {
	User              *UserDTO `json:"user"`
	ContributedPoints int      `json:"contributed_points"`
}

// PrizePoolJackpotEventDTO 奖池中奖事件
type PrizePoolJackpotEventDTO struct {
	GameID    int      `json:"game_id"`
	User      *UserDTO `json:"user"`
	Points    int      `json:"points"`
	CreatedAt string   `json:"created_at"`
}

// PrizePoolDetailDTO 奖池明细
type PrizePoolDetailDTO struct {
	PoolType      string                      `json:"pool_type"`
	Contributors  []*PrizePoolContributorDTO  `json:"contributors"`
	JackpotEvents []*PrizePoolJackpotEventDTO `json:"jackpot_events"`
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
