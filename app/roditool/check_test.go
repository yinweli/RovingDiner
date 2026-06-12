package roditool

import (
	"testing"

	"github.com/stretchr/testify/suite"

	sheeter "github.com/yinweli/RovingDiner/sheet"
)

func TestSuiteCheck(t *testing.T) {
	suite.Run(t, new(SuiteCheck))
}

// SuiteCheck 驗證表單檢查(check.go): 入口彙整(固定表序 / 決定性)與各表參照 / 門檻 / 設定檢查。
type SuiteCheck struct {
	suite.Suite
}

// TestIssueString 驗證單筆結果的人類可讀行格式(CLI 表單檢查輸出)。
func (this *SuiteCheck) TestIssueString() {
	this.Equal("表 card | 列 3 | 欄 SkillID | 查無技能編號:888",
		Issue{Table: tableCard, Row: "3", Column: "SkillID", Msg: "查無技能編號:888"}.String())
}

// TestCheckSheet 驗證入口: nil 防禦、跨表彙整依固定表序輸出。
func (this *SuiteCheck) TestCheckSheet() {
	this.Empty(CheckSheet(nil)) // nil 防禦

	sheet := &sheeter.Sheeter{}
	buildSetting(sheet)
	sheet.Award.Data = map[int32]*sheeter.Award{1: {ID: 1, CardID: 999}}
	sheet.Stage.Data = map[int32]*sheeter.Stage{601: {ID: 601, WaitID: []int32{888}}}
	issue := CheckSheet(sheet)
	this.Require().Len(issue, 2)
	this.Equal(Issue{Table: "award", Row: "1", Column: "CardID", Msg: "查無卡牌編號:999"}, issue[0])
	this.Equal(Issue{Table: "stage", Row: "601", Column: "WaitID", Msg: "查無顧客編號:888"}, issue[1])
}

// TestCheckAward 驗證抽獎表: 合法參照 / 0 未指派通過, 查無編號回報。
func (this *SuiteCheck) TestCheckAward() {
	sheet := &sheeter.Sheeter{}
	sheet.Card.Data = map[int32]*sheeter.Card{101: {ID: 101}}
	sheet.Award.Data = map[int32]*sheeter.Award{
		1: {ID: 1, CardID: 101}, // 合法
		2: {ID: 2, CardID: 0},   // 未指派 → 跳過
		3: {ID: 3, CardID: 999}, // 查無
	}
	issue := checkAward(sheet)
	this.Require().Len(issue, 1)
	this.Equal(Issue{Table: "award", Row: "3", Column: "CardID", Msg: "查無卡牌編號:999"}, issue[0])
}

// TestCheckCard 驗證卡牌表: 0 = 無技能通過, 查無技能回報。
func (this *SuiteCheck) TestCheckCard() {
	sheet := &sheeter.Sheeter{}
	sheet.Skill.Data = map[int32]*sheeter.Skill{301: {ID: 301}}
	sheet.Card.Data = map[int32]*sheeter.Card{
		101: {ID: 101, SkillID: 301},
		102: {ID: 102, SkillID: 0},
		103: {ID: 103, SkillID: 888},
	}
	issue := checkCard(sheet)
	this.Require().Len(issue, 1)
	this.Equal(Issue{Table: "card", Row: "103", Column: "SkillID", Msg: "查無技能編號:888"}, issue[0])
}

// TestCheckGuest 驗證顧客表門檻配對: 空字串未填跳過、格式錯誤(共用 cores.ParseThreshold)、技能參照、
// 技能編號 0 未指派跳過。
func (this *SuiteCheck) TestCheckGuest() {
	sheet := &sheeter.Sheeter{}
	sheet.Skill.Data = map[int32]*sheeter.Skill{301: {ID: 301}}
	sheet.Guest.Data = map[int32]*sheeter.Guest{
		501: {ID: 501, SateSkillID: []string{"", "6^301", "7^0"}}, // 全通過: 空筆 / 合法 / 技能未指派
		502: {ID: 502, SateSkillID: []string{"bad"}},              // 格式錯
		503: {ID: 503, CalmSkillID: []string{"5^999"}},            // 查無技能
	}
	issue := checkGuest(sheet)
	this.Require().Len(issue, 2)
	this.Equal("502", issue[0].Row)
	this.Equal("SateSkillID", issue[0].Column)
	this.Contains(issue[0].Msg, "門檻配對需為「門檻值^技能編號」兩段")
	this.Equal(Issue{Table: "guest", Row: "503", Column: "CalmSkillID", Msg: "查無技能編號:999"}, issue[1])
}

// TestCheckSetting 驗證設定表: 六鍵齊備且數字通過; 缺鍵 / 缺值 / 非數字回報。
func (this *SuiteCheck) TestCheckSetting() {
	sheet := &sheeter.Sheeter{}
	buildSetting(sheet)
	this.Empty(checkSetting(sheet))

	delete(sheet.Setting.Data, "RoundMax")
	sheet.Setting.Data["Morale"].Value = nil              // 缺值
	sheet.Setting.Data["HandMax"].Value = []string{"abc"} // 非數字
	issue := checkSetting(sheet)
	this.Require().Len(issue, 3)
	this.Equal(Issue{Table: "setting", Row: "RoundMax", Column: "ID", Msg: "缺少設定鍵:RoundMax"}, issue[0])
	this.Equal(Issue{Table: "setting", Row: "Morale", Column: "Value", Msg: "設定值未填"}, issue[1])
	this.Equal(Issue{Table: "setting", Row: "HandMax", Column: "Value", Msg: "設定值需為數字:abc"}, issue[2])
}

// TestCheckSkill 驗證技能表: 效果編號列表逐項參照, 一列可多筆。
func (this *SuiteCheck) TestCheckSkill() {
	sheet := &sheeter.Sheeter{}
	sheet.Effect.Data = map[int32]*sheeter.Effect{401: {ID: 401}}
	sheet.Skill.Data = map[int32]*sheeter.Skill{
		301: {ID: 301, EffectID: []int32{401, 0}},   // 合法 + 未指派
		302: {ID: 302, EffectID: []int32{888, 999}}, // 兩筆查無
	}
	issue := checkSkill(sheet)
	this.Require().Len(issue, 2)
	this.Equal(Issue{Table: "skill", Row: "302", Column: "EffectID", Msg: "查無效果編號:888"}, issue[0])
	this.Equal(Issue{Table: "skill", Row: "302", Column: "EffectID", Msg: "查無效果編號:999"}, issue[1])
}

// TestCheckStage 驗證關卡表: 四牌堆 → 卡牌、排隊 → 顧客、前置 → 技能, 各欄獨立回報。
func (this *SuiteCheck) TestCheckStage() {
	sheet := &sheeter.Sheeter{}
	sheet.Card.Data = map[int32]*sheeter.Card{101: {ID: 101}}
	sheet.Guest.Data = map[int32]*sheeter.Guest{501: {ID: 501}}
	sheet.Skill.Data = map[int32]*sheeter.Skill{301: {ID: 301}}
	sheet.Stage.Data = map[int32]*sheeter.Stage{
		601: {ID: 601, HandID: []int32{101}, DeckID: []int32{101}, WaitID: []int32{501}, PrefixSkillID: []int32{301}}, // 全合法
		602: {ID: 602, HandID: []int32{900}, DeckID: []int32{901}, DropID: []int32{902}, ExileID: []int32{903}, WaitID: []int32{904}, PrefixSkillID: []int32{905}},
	}
	issue := checkStage(sheet)
	this.Require().Len(issue, 6)
	this.Equal(Issue{Table: "stage", Row: "602", Column: "HandID", Msg: "查無卡牌編號:900"}, issue[0])
	this.Equal(Issue{Table: "stage", Row: "602", Column: "DeckID", Msg: "查無卡牌編號:901"}, issue[1])
	this.Equal(Issue{Table: "stage", Row: "602", Column: "DropID", Msg: "查無卡牌編號:902"}, issue[2])
	this.Equal(Issue{Table: "stage", Row: "602", Column: "ExileID", Msg: "查無卡牌編號:903"}, issue[3])
	this.Equal(Issue{Table: "stage", Row: "602", Column: "WaitID", Msg: "查無顧客編號:904"}, issue[4])
	this.Equal(Issue{Table: "stage", Row: "602", Column: "PrefixSkillID", Msg: "查無技能編號:905"}, issue[5])
}

// === 測試輔助(置尾) ===

// buildSetting 佈置設定六鍵合法值(鍵自 settingKey 列舉; 隔離 setting 檢查噪音, 聚焦各表斷言)。
func buildSetting(sheet *sheeter.Sheeter) {
	sheet.Setting.Data = map[string]*sheeter.Setting{}

	for _, itor := range settingKey {
		sheet.Setting.Data[itor] = &sheeter.Setting{ID: itor, Value: []string{"10"}}
	} // for
}
