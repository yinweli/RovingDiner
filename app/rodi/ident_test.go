package rodi

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/yinweli/RovingDiner/internal/cores"
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
	this.Equal("101@上菜#7", identCard(testSheet(), 101, 7))
	this.Equal("101@上菜", identCard(testSheet(), 101, cores.NoneID)) // 主畫面投影省實例段
	this.Equal("999@?#7", identCard(testSheet(), 999, 7))           // 查無資料
}

// TestIdentGuest 驗證顧客識別碼; 規則同 identCard。
func (this *SuiteIdent) TestIdentGuest() {
	this.Equal("501@老饕#3", identGuest(testSheet(), 501, 3))
	this.Equal("501@老饕", identGuest(testSheet(), 501, cores.NoneID))
	this.Equal("999@?", identGuest(testSheet(), 999, cores.NoneID))
}

// TestIdentSkill 驗證技能識別碼: 靜態無實例段。
func (this *SuiteIdent) TestIdentSkill() {
	this.Equal("301@開朗", identSkill(testSheet(), 301))
	this.Equal("999@?", identSkill(testSheet(), 999))
}

// TestIdentEffect 驗證效果識別碼。
func (this *SuiteIdent) TestIdentEffect() {
	this.Equal("401@加耐#9", identEffect(testSheet(), 401, 9))
	this.Equal("999@?#9", identEffect(testSheet(), 999, 9))
}

// TestIdentTarget 驗證辨型識別碼: 卡牌表優先、未中查顧客表、兩表皆查無顯 ?。
func (this *SuiteIdent) TestIdentTarget() {
	this.Equal("101@上菜#7", identTarget(testSheet(), 101, 7))
	this.Equal("501@老饕#3", identTarget(testSheet(), 501, 3))
	this.Equal("999@?#1", identTarget(testSheet(), 999, 1))
}
