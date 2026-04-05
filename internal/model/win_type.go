package model

// WinType 番型字典
type WinType struct {
	Code        string `gorm:"primaryKey;size:20" json:"code"`
	Name        string `gorm:"size:20;not null" json:"name"`
	BaseMulti   int    `gorm:"not null" json:"base_multi"`
	Description string `gorm:"size:100" json:"description"`
}

func (WinType) TableName() string {
	return "win_type"
}

// 预定义番型
var DefaultWinTypes = []*WinType{
	{Code: "wu_hua_guo", Name: "无花果", BaseMulti: 1, Description: "无番型"},
	{Code: "peng_peng_hu", Name: "碰碰胡", BaseMulti: 2, Description: "全部由碰牌组成"},
	{Code: "yi_tiao_long", Name: "一条龙", BaseMulti: 2, Description: "同一花色1-9"},
	{Code: "hun_yi_se", Name: "混一色", BaseMulti: 2, Description: "同一花色加字牌"},
	{Code: "qing_yi_se", Name: "清一色", BaseMulti: 4, Description: "同一花色"},
	{Code: "xiao_qi_dui", Name: "小七对", BaseMulti: 4, Description: "七个对子"},
	{Code: "long_qi_dui", Name: "龙七对", BaseMulti: 8, Description: "小七对加一根"},
	{Code: "da_diao_che", Name: "大吊车", BaseMulti: 2, Description: "单吊将牌"},
	{Code: "men_qian_qing", Name: "门前清", BaseMulti: 2, Description: "未碰未吃"},
	{Code: "gang_kai_hua", Name: "杠开花", BaseMulti: 2, Description: "杠牌后自摸"},
}

// GetWinTypeByCode 根据code获取番型
func GetWinTypeByCode(code string) (*WinType, bool) {
	for _, wt := range DefaultWinTypes {
		if wt.Code == code {
			return wt, true
		}
	}
	return nil, false
}
