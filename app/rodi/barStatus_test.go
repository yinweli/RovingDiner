package rodi

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/yinweli/RovingDiner/internal/cores"
)

func TestSuiteBarStatus(t *testing.T) {
	suite.Run(t, new(SuiteBarStatus))
}

// SuiteBarStatus 驗證狀態列組件(barStatus.go): 兩行對齊表欄序與各欄取值。
type SuiteBarStatus struct {
	suite.Suite
}

// TestBarStatusView 驗證渲染: 欄序 / 對齊與【營業顯示規格書 | 6、畫面規格 | 6.1】草圖一致; 超寬依預算截斷。
func (this *SuiteBarStatus) TestBarStatusView() {
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

	this.Equal(strings.Join([]string{ // 模式欄吃下傳的 UI 狀態; M26 R1.5 帶框 + 標題列
		"+- 狀態列 " + strings.Repeat("-", 89) + "+",
		"| " + padTo("回合  士氣   護盾  格擋  出牌點數  滿意  階段      模式", 96) + " |",
		"| " + padTo("3/10  25/30  5     3     2/10      1250  玩家行動  步進", 96) + " |",
	}, "\n"), barStatus{}.View(game, modeStep, 100))

	row := strings.Split(barStatus{}.View(game, modeFast, 10), "\n") // 超寬: 內容寬 6 純截斷(不補記號)
	this.Equal("| 回合   |", row[1])
	this.Equal("| 3/10   |", row[2])
}

// TestEnergyText 驗證出牌點數欄: 當前 / 上限; 出牌點數保留鎖定中加「保」。
func (this *SuiteBarStatus) TestEnergyText() {
	game := testGame()
	game.GetEnergy().Set(2)
	game.GetEnergyMax().Set(10)
	this.Equal("2/10", energyText(game))

	game.GetEnergyKeep().Lock()
	this.Equal("2/10保", energyText(game))
}
