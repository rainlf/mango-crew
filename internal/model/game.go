package model

import (
	"time"
)

// MaJiangGameType 麻将游戏类型
type MaJiangGameType int

const (
	PingHu          MaJiangGameType = 1 // 平胡
	ZiMo            MaJiangGameType = 2 // 自摸
	YiPaoShuangXiang MaJiangGameType = 3 // 一炮双响
	YiPaoSanXiang   MaJiangGameType = 4 // 一炮三响
	XiangGong       MaJiangGameType = 5 // 相公
	YunDong         MaJiangGameType = 6 // 运动
)

func (t MaJiangGameType) Name() string {
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

func MaJiangGameTypeFromCode(code int) MaJiangGameType {
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

// MaJiangUserType 用户在牌局中的角色
type MaJiangUserType int

const (
	Winner  MaJiangUserType = 1 // 赢家
	Loser   MaJiangUserType = 2 // 输家
	Recorder MaJiangUserType = 3 // 记录者
)

func (t MaJiangUserType) Name() string {
	switch t {
	case Winner:
		return "赢家"
	case Loser:
		return "输家"
	case Recorder:
		return "记录者"
	default:
		return "未知"
	}
}

// MaJiangWinType 赢牌类型（番型）
type MaJiangWinType struct {
	Name  string `json:"name"`
	Code  int    `json:"code"`
	Multi int    `json:"multi"` // 倍数
}

var WinTypes = map[string]*MaJiangWinType{
	"无花果":  {Name: "无花果", Code: 1, Multi: 1},
	"碰碰胡":  {Name: "碰碰胡", Code: 2, Multi: 2},
	"一条龙":  {Name: "一条龙", Code: 3, Multi: 2},
	"混一色":  {Name: "混一色", Code: 4, Multi: 2},
	"清一色":  {Name: "清一色", Code: 5, Multi: 4},
	"小七对":  {Name: "小七对", Code: 6, Multi: 4},
	"龙七对":  {Name: "龙七对", Code: 7, Multi: 8},
	"大吊车":  {Name: "大吊车", Code: 8, Multi: 2},
	"门前清":  {Name: "门前清", Code: 9, Multi: 2},
	"杠开花":  {Name: "杠开花", Code: 10, Multi: 2},
}

func GetWinTypeByName(name string) (*MaJiangWinType, bool) {
	wt, ok := WinTypes[name]
	return wt, ok
}

// MaJiangGame 麻将游戏记录
type MaJiangGame struct {
	ID          int             `gorm:"primaryKey;autoIncrement" json:"id"`
	Type        MaJiangGameType `gorm:"not null" json:"type"`
	Player1     int             `gorm:"not null" json:"player1"`
	Player2     int             `gorm:"not null" json:"player2"`
	Player3     int             `gorm:"not null" json:"player3"`
	Player4     int             `gorm:"not null" json:"player4"`
	IsDeleted   bool            `gorm:"default:false;not null" json:"-"`
	CreatedTime time.Time       `gorm:"default:CURRENT_TIMESTAMP;not null" json:"created_time"`
	UpdatedTime time.Time       `gorm:"default:CURRENT_TIMESTAMP;autoUpdateTime" json:"updated_time"`
}

func (MaJiangGame) TableName() string {
	return "mgtt_majiang_game"
}

// MaJiangGameItem 麻将游戏明细
type MaJiangGameItem struct {
	ID          int             `gorm:"primaryKey;autoIncrement" json:"id"`
	GameID      int             `gorm:"not null;index" json:"game_id"`
	UserID      int             `gorm:"not null" json:"user_id"`
	Type        MaJiangUserType `gorm:"not null" json:"type"`
	BasePoint   int             `json:"base_point"`
	WinTypes    string          `gorm:"size:512" json:"win_types"` // 逗号分隔的番型名称
	Multi       int             `json:"multi"`
	Points      int             `gorm:"default:0;not null" json:"points"`
	CreatedTime time.Time       `gorm:"default:CURRENT_TIMESTAMP;not null" json:"created_time"`
	UpdatedTime time.Time       `gorm:"default:CURRENT_TIMESTAMP;autoUpdateTime" json:"updated_time"`
}

func (MaJiangGameItem) TableName() string {
	return "mgtt_majiang_game_item"
}

// CalculatePoints 计算分数
func (item *MaJiangGameItem) CalculatePoints() {
	multi := 1
	if item.WinTypes != "" {
		// 解析番型并计算倍数
		// 这里简化处理，实际应该解析逗号分隔的字符串
	}
	item.Multi = multi
	item.Points = item.BasePoint * multi
}

// MaJiangGameInfo 游戏完整信息（包含明细）
type MaJiangGameInfo struct {
	Game  *MaJiangGame
	Items []*MaJiangGameItem
}

// FindUserGameItem 查找指定用户的游戏记录
func (info *MaJiangGameInfo) FindUserGameItem(userID int) *MaJiangGameItem {
	for _, item := range info.Items {
		if item.UserID == userID {
			return item
		}
	}
	return nil
}
