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
	world := newMirror()
	world.round = 3
	world.phase = cores.PhasePlayerAction
	world.attr = map[string]float64{
		"roundMax": 10, "morale": 25, "moraleMax": 30, "moraleShield": 5, "moraleBlock": 3,
		"energy": 2, "energyMax": 10, "score": 1250,
	}
	this.Equal("回合  士氣   護盾  格擋  出牌點數  滿意  階段      模式\n"+
		"3/10  25/30  5     3     2/10      1250  玩家行動  快速", statusBar{}.View(world, 100))
	this.Equal("回合\n3/10", statusBar{}.View(world, 4)) // 超寬截斷
}

// TestNum 驗證整數屬性值轉字串。
func (this *SuiteStatusBar) TestNum() {
	this.Equal("0", num(0))
	this.Equal("1250", num(1250))
	this.Equal("-3", num(-3))
}

// TestNumFloor 驗證護盾 / 格擋顯示下限: <= 0 顯 0。
func (this *SuiteStatusBar) TestNumFloor() {
	this.Equal("5", numFloor(5))
	this.Equal("0", numFloor(0))
	this.Equal("0", numFloor(-2))
}

// TestEnergyText 驗證出牌點數欄: 當前 / 上限; 出牌點數保留鎖定中加「保」。
func (this *SuiteStatusBar) TestEnergyText() {
	world := newMirror()
	world.attr["energy"] = 2
	world.attr["energyMax"] = 10
	this.Equal("2/10", energyText(world))

	world.lock["energyKeep"] = 1
	this.Equal("2/10保", energyText(world))
}

// TestPhaseName 驗證階段中文全名; 無階段 / 未知顯 -。
func (this *SuiteStatusBar) TestPhaseName() {
	this.Equal("營業開始", phaseName(cores.PhaseGameStart))
	this.Equal("回合開始", phaseName(cores.PhaseRoundStart))
	this.Equal("玩家行動", phaseName(cores.PhasePlayerAction))
	this.Equal("顧客行動", phaseName(cores.PhaseGuestAction))
	this.Equal("回合結束", phaseName(cores.PhaseRoundEnd))
	this.Equal("營業成功", phaseName(cores.PhaseGameSucc))
	this.Equal("營業失敗", phaseName(cores.PhaseGameFail))
	this.Equal("-", phaseName(cores.PhaseNone))
}
