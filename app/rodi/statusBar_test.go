package rodi

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/yinweli/RovingDiner/internal/cores"
)

func TestSuiteStatusBar(t *testing.T) {
	suite.Run(t, new(SuiteStatusBar))
}

// SuiteStatusBar 驗證狀態列組件(statusBar.go): 兩行對齊表欄序與各欄取值。
type SuiteStatusBar struct {
	suite.Suite
}

// TestStatusBarView 驗證渲染: 欄序 / 對齊與【營業顯示規格書 | 6、畫面規格 | 6.1】草圖一致; 超寬依預算截斷。
func (this *SuiteStatusBar) TestStatusBarView() {
	game := testGame()
	game.SetPhase(cores.PhasePlayerAction)
	game.GetRound().Set(3)
	game.GetRoundMax().Set(10)
	game.GetMorale().Set(25)
	game.GetMoraleMax().Set(30)
	game.GetMoraleShield().Set(5)
	game.GetMoraleBlock().Set(3)
	game.GetEnergy().Set(2)
	game.GetEnergyMax().Set(10)
	game.GetScore().Set(1250)

	this.Equal("回合  士氣   護盾  格擋  出牌點數  滿意  階段      模式\n"+
		"3/10  25/30  5     3     2/10      1250  玩家行動  快速", statusBar{}.View(game, 100))
	this.Equal("回合\n3/10", statusBar{}.View(game, 4)) // 超寬截斷
}

// TestEnergyText 驗證出牌點數欄: 當前 / 上限; 出牌點數保留鎖定中加「保」。
func (this *SuiteStatusBar) TestEnergyText() {
	game := testGame()
	game.GetEnergy().Set(2)
	game.GetEnergyMax().Set(10)
	this.Equal("2/10", energyText(game))

	game.GetEnergyKeep().Lock()
	this.Equal("2/10保", energyText(game))
}
