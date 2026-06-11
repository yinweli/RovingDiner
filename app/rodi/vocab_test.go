package rodi

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/yinweli/RovingDiner/internal/cores"
)

func TestSuiteVocab(t *testing.T) {
	suite.Run(t, new(SuiteVocab))
}

// SuiteVocab 驗證範圍轉換(vocab.go): 詞彙對照本體的測試隨下沉移居 cores。
type SuiteVocab struct {
	suite.Suite
}

// TestScopeText 驗證範圍事件名: 六種範圍 + 時機帶時機名 + 未知顯 ?。
func (this *SuiteVocab) TestScopeText() {
	this.Equal("玩家出牌", scopeText(cores.ScopePlay, ""))
	this.Equal("顧客行動", scopeText(cores.ScopeGuest, ""))
	this.Equal("前置技能", scopeText(cores.ScopePrefix, ""))
	this.Equal("時機:玩家出牌", scopeText(cores.ScopeTrigger, cores.TriggerCardPlay))
	this.Equal("執行結算", scopeText(cores.ScopeSettle, ""))
	this.Equal("手動結束", scopeText(cores.ScopeManual, ""))
	this.Equal("?", scopeText(cores.ScopeNone, ""))
}
