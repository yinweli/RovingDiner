package rodi

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

func TestSuiteMode(t *testing.T) {
	suite.Run(t, new(SuiteMode))
}

// SuiteMode 驗證執行模式(mode.go): 顯示名稱 / 排拍間隔 / 循環切換。
type SuiteMode struct {
	suite.Suite
}

// TestModeName 驗證顯示名稱(狀態列模式欄文字); 未知值防禦顯 "-"。
func (this *SuiteMode) TestModeName() {
	this.Equal("快速", modeFast.name())
	this.Equal("慢速", modeSlow.name())
	this.Equal("步進", modeStep.name())
	this.Equal("-", mode(99).name())
}

// TestModeInterval 驗證排拍間隔: 快 0.2 秒 / 慢 1 秒(M25 拍板數值); 步進不排拍回 0。
func (this *SuiteMode) TestModeInterval() {
	this.Equal(200*time.Millisecond, modeFast.interval())
	this.Equal(time.Second, modeSlow.interval())
	this.Equal(time.Duration(0), modeStep.interval())
}

// TestModeNext 驗證循環切換: 快速 → 慢速 → 步進 → 快速。
func (this *SuiteMode) TestModeNext() {
	this.Equal(modeSlow, modeFast.next())
	this.Equal(modeStep, modeSlow.next())
	this.Equal(modeFast, modeStep.next())
}
