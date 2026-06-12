package roditool

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

func TestSuiteField(t *testing.T) {
	suite.Run(t, new(SuiteField))
}

// SuiteField 驗證單筆檢查引擎(field.go): 三種文法各自的通過與文法 / 詞彙錯誤(與表單檢查共用同一路徑,
// 細項分支由 SuiteCheckEffect 與 games / cores 各層測試釘住, 此處驗分派)。
type SuiteField struct {
	suite.Suite
}

// TestCheckExpr 驗證運算式單筆檢查: 通過 / 文法錯 / 詞彙錯。
func (this *SuiteField) TestCheckExpr() {
	this.NoError(CheckExpr("morale > 0 AND self.calm >= 1"))
	this.Error(CheckExpr("1 +"))         // 文法錯
	this.Error(CheckExpr("morale2 > 1")) // 詞彙錯(R4A)
}

// TestCheckCommand 驗證命令單筆檢查: 通過 / 文法錯 / arity 錯 / 內嵌詞彙錯。
func (this *SuiteField) TestCheckCommand() {
	this.NoError(CheckCommand("handAdd(none, 10031, 1)"))
	this.NoError(CheckCommand("morale -= 2"))
	this.Error(CheckCommand("handAdd(none"))               // 文法錯
	this.Error(CheckCommand("guestExit(guestPick, 1, 1)")) // 命令對象 arity 錯(R1)
	this.Error(CheckCommand("morale = morale2 + 1"))       // 內嵌算術式詞彙錯(R4A)
}

// TestCheckThreshold 驗證門檻配對單筆檢查: 通過 / 格式錯(技能參照屬表單檢查不在此)。
func (this *SuiteField) TestCheckThreshold() {
	this.NoError(CheckThreshold("6^301"))
	this.Error(CheckThreshold("x^301")) // 格式錯
}
