package cores

import (
	"testing"

	"github.com/stretchr/testify/suite"

	sheeter "github.com/yinweli/RovingDiner/sheet"
)

func TestSuiteIdent(t *testing.T) {
	suite.Run(t, new(SuiteIdent))
}

// SuiteIdent 驗證識別碼工具(ident.go): 模板組裝、主畫面省實例段、技能無實例段、查無資料顯 ?。
type SuiteIdent struct {
	suite.Suite
}

// TestIdentCard 驗證卡牌識別碼: 完整三段、NoneID 省實例段、查無資料名稱顯 ?。
func (this *SuiteIdent) TestIdentCard() {
	this.Equal("101@上菜#7", IdentCard(identSheet(), 101, 7))
	this.Equal("101@上菜", IdentCard(identSheet(), 101, NoneID)) // 主畫面省實例段
	this.Equal("999@?#7", IdentCard(identSheet(), 999, 7))     // 查無資料
}

// TestIdentGuest 驗證顧客識別碼; 規則同 IdentCard。
func (this *SuiteIdent) TestIdentGuest() {
	this.Equal("501@老饕#3", IdentGuest(identSheet(), 501, 3))
	this.Equal("501@老饕", IdentGuest(identSheet(), 501, NoneID))
	this.Equal("999@?", IdentGuest(identSheet(), 999, NoneID))
}

// TestIdentSkill 驗證技能識別碼: 靜態無實例段。
func (this *SuiteIdent) TestIdentSkill() {
	this.Equal("301@開朗", IdentSkill(identSheet(), 301))
	this.Equal("999@?", IdentSkill(identSheet(), 999))
}

// TestIdentEffect 驗證效果識別碼。
func (this *SuiteIdent) TestIdentEffect() {
	this.Equal("401@加耐#9", IdentEffect(identSheet(), 401, 9))
	this.Equal("999@?#9", IdentEffect(identSheet(), 999, 9))
}

// TestIdentTarget 驗證辨型識別碼: 卡牌表優先、未中查顧客表、兩表皆查無顯 ?。
func (this *SuiteIdent) TestIdentTarget() {
	this.Equal("101@上菜#7", IdentTarget(identSheet(), 101, 7))
	this.Equal("501@老饕#3", IdentTarget(identSheet(), 501, 3))
	this.Equal("999@?#1", IdentTarget(identSheet(), 999, 1))
}

// identSheet 識別碼測試用迷你表: 各類各一筆具名資料列。
func identSheet() *sheeter.Sheeter {
	sheet := &sheeter.Sheeter{}
	sheet.Card.Data = map[int32]*sheeter.Card{101: {ID: 101, Name: "上菜"}}
	sheet.Guest.Data = map[int32]*sheeter.Guest{501: {ID: 501, Name: "老饕"}}
	sheet.Skill.Data = map[int32]*sheeter.Skill{301: {ID: 301, Name: "開朗"}}
	sheet.Effect.Data = map[int32]*sheeter.Effect{401: {ID: 401, Name: "加耐"}}
	return sheet
}
