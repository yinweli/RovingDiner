package rodi

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/yinweli/RovingDiner/internal/cores"
)

func TestSuiteModalCount(t *testing.T) {
	suite.Run(t, new(SuiteModalCount))
}

// SuiteModalCount 驗證計數 modal(modalCount.go): 標題嵌 seed、屬性配對 col2 跨群組對齊、
// 值+鎖 / 純鎖標記、空引用欄空著與附帶資訊、歷史計數樞紐表。
type SuiteModalCount struct {
	suite.Suite
}

// TestModalCountBody 驗證內容行(全量釘字串): 分隔線前綴 / 配對對齊 / 兩張表。
func (this *SuiteModalCount) TestModalCountBody() {
	game := testGame()
	game.GetRound().Set(3)
	game.GetRoundMax().Set(10)
	game.GetScore().Set(1250)
	game.SetPhase(cores.PhasePlayerAction)
	game.SetNextPhase(cores.PhaseGuestAction)
	game.GetMorale().Set(25)
	game.GetMorale().Lock()
	game.GetMorale().Lock()
	game.GetMoraleMax().Set(30)
	game.GetMoraleShield().Set(5)
	game.GetMoraleBlock().Set(3)
	game.GetHandMax().Set(5)
	game.GetDrawMax().Set(3)
	game.GetEnergy().Set(2)
	game.GetEnergyMax().Set(10)
	game.GetEnergyKeep().Lock()
	guest := cores.NewGuest(game, 501) // 實例編號自 1 起配發
	game.EventDamage(4, guest)
	game.EventSeat(guest)
	card := cores.NewCard(game, 101) // #2
	game.EventDraw(card, 3)
	game.EventPlay(card, 3)
	game.EventDrop(cores.NewCard(game, 103), 1) // #3

	title, row := modalCount{seed: 42}.Body(game)
	this.Equal("計數檢視 (seed 42)", title)
	this.Equal([]string{
		"回合 3          回合上限 10",
		"滿意值 1250     階段 玩家行動 > 顧客行動",
		"+- 士氣",
		"士氣值 25[鎖2]  士氣值上限 30",
		"士氣值護盾 5    士氣值格擋 3",
		"士氣受損 4 (501@老饕#1)",
		"+- 卡牌",
		"手牌上限 5      補牌上限 3",
		"出牌點數 2      出牌點數上限 10",
		"出牌點數保留[鎖1]",
		"+- 回合計數",
		"項目  次數  最後對象",
		"入座  1     501@老饕#1",
		"離場  0",
		"行動  0",
		"抽牌  1     101@上菜#2",
		"棄牌  1     103@結帳#3",
		"出牌  1     101@上菜#2",
		"流放  0",
		"變身  0",
		"+- 歷史計數",
		"群組  抽牌張數  棄牌張數  出牌張數  流放張數",
		"1     0         1         0         0",
		"3     1         0         1         0",
	}, row)
}

// TestModalCountRef 驗證引用顯示: 離場附座位桌號、行動附技能、變身附前後編號、空引用欄空著、
// 士氣受損空物件顯 空。
func (this *SuiteModalCount) TestModalCountRef() {
	game := testGame()
	guest := cores.NewGuest(game, 501)
	game.Seat.Place(1, guest) // 座 1 = 桌 1
	game.EventExit(guest)
	game.EventTask(guest, 301)
	card := cores.NewCard(game, 101)
	game.EventMorph(card, 101, 103)

	_, row := modalCount{}.Body(game)
	this.Contains(row, "離場  1     501@老饕#1 (桌1)")
	this.Contains(row, "行動  1     501@老饕#1 (301@開朗)")
	this.Contains(row, "變身  1     101@上菜#2 (101->103)")
	this.Contains(row, "士氣受損 0 (空)") // 空物件
}

// TestValText 驗證數值實例顯示: 無鎖顯值、鎖定計數 > 0 值後緊接 [鎖N]。
func (this *SuiteModalCount) TestValText() {
	game := testGame()
	game.GetMorale().Set(25)
	this.Equal("25", valText(game.GetMorale()))

	game.GetMorale().Lock()
	this.Equal("25[鎖1]", valText(game.GetMorale()))
}

// TestLockText 驗證純鎖屬性顯示: 一律 [鎖N] 含 [鎖0]。
func (this *SuiteModalCount) TestLockText() {
	game := testGame()
	this.Equal("[鎖0]", lockText(game.GetEnergyKeep()))

	game.GetEnergyKeep().Lock()
	this.Equal("[鎖1]", lockText(game.GetEnergyKeep()))
}
