package model

// 预定义番型
type WinTypeDef struct {
	Code        string
	Name        string
	BaseMulti   int
	Description string
}

var DefaultWinTypes = []*WinTypeDef{
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
func GetWinTypeByCode(code string) (*WinTypeDef, bool) {
	for _, wt := range DefaultWinTypes {
		if wt.Code == code {
			return wt, true
		}
	}
	return nil, false
}

// GetWinTypeByName 根据中文名称获取番型
func GetWinTypeByName(name string) (*WinTypeDef, bool) {
	for _, wt := range DefaultWinTypes {
		if wt.Name == name {
			return wt, true
		}
	}
	return nil, false
}

// ResolveWinType 支持按 code 或中文名称解析番型
func ResolveWinType(codeOrName string) (*WinTypeDef, bool) {
	if wt, ok := GetWinTypeByCode(codeOrName); ok {
		return wt, true
	}
	return GetWinTypeByName(codeOrName)
}
