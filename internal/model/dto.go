package model

// UserDTO 用户数据传输对象
type UserDTO struct {
	ID       int      `json:"id"`
	Username string   `json:"username"`
	Points   int      `json:"points"`
	Avatar   string   `json:"avatar,omitempty"`
	LastTags []string `json:"last_tags,omitempty"`
}

// FromUser 从 User 创建 DTO
func (dto *UserDTO) FromUser(user *User) *UserDTO {
	if user == nil {
		return nil
	}
	return &UserDTO{
		ID:       user.ID,
		Username: user.Username,
		Points:   user.Points,
		LastTags: user.LastTags,
	}
}

// MaJiangGameLogDTO 麻将游戏日志 DTO
type MaJiangGameLogDTO struct {
	ID          int               `json:"id"`
	Type        string            `json:"type"`
	Player1     *UserDTO          `json:"player1"`
	Player2     *UserDTO          `json:"player2"`
	Player3     *UserDTO          `json:"player3"`
	Player4     *UserDTO          `json:"player4"`
	CreatedTime string            `json:"created_time"`
	UpdatedTime string            `json:"updated_time"`
	Winners     []*MaJiangGameItemDTO `json:"winners"`
	Losers      []*MaJiangGameItemDTO `json:"losers"`
	Recorder    *MaJiangGameItemDTO   `json:"recorder"`
	ForOnePlayer bool             `json:"for_one_player,omitempty"`
	PlayerWin    bool             `json:"player_win,omitempty"`
}

// MaJiangGameItemDTO 游戏明细 DTO
type MaJiangGameItemDTO struct {
	User   *UserDTO `json:"user"`
	Points int      `json:"points"`
	Tags   []string `json:"tags"`
}

// PlayersDTO 玩家列表 DTO
type PlayersDTO struct {
	CurrentPlayers []*UserDTO `json:"current_players"`
	AllPlayers     []*UserDTO `json:"all_players"`
}

// SaveMaJiangGameRequest 保存麻将游戏请求
type SaveMaJiangGameRequest struct {
	GameType   int                  `json:"game_type" binding:"required"`
	Players    []int                `json:"players" binding:"required"`
	RecorderID int                  `json:"recorder_id" binding:"required"`
	Winners    []*WinnerRequest     `json:"winners" binding:"required"`
	Losers     []int                `json:"losers"`
}

// WinnerRequest 赢家请求
type WinnerRequest struct {
	UserID     int      `json:"user_id" binding:"required"`
	BasePoints int      `json:"base_points" binding:"required"`
	WinTypes   []string `json:"win_types"`
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
