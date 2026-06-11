package rodi

import (
	"github.com/yinweli/RovingDiner/internal/cores"
)

// barStatus 狀態列組件(區 7; 【營業顯示規格書 | 6、畫面規格 | 6.9】): 標題列 + 兩行對齊表——
// 第 1 行標籤、第 2 行數值, 欄序 = 回合 / 士氣 / 護盾 / 格擋 / 出牌點數 / 滿意 / 階段 / 模式;
// 數值直讀引擎盤面、無自持狀態, 唯模式欄吃 UI 狀態(model 持模式下傳; M25)。階段欄只顯當前(M22 拍板);
// 回合計數與 seed 不入列(歸 M26 計數 modal)。
type barStatus struct{}

// View 渲染標題列 + 兩行對齊表(M26 R1.5 帶框); 超寬依預算截斷(非捲動區, 不補記號)。
func (this barStatus) View(game *cores.Game, mode mode, width int) string {
	row1, row2 := alignPair(
		[]string{"回合", "士氣", "護盾", "格擋", "出牌點數", "滿意", "階段", "模式"},
		[]string{
			num(game.GetRound().GetValue()) + "/" + num(game.GetRoundMax().GetValue()),
			num(game.GetMorale().GetValue()) + "/" + num(game.GetMoraleMax().GetValue()),
			numFloor(game.GetMoraleShield().GetValue()),
			numFloor(game.GetMoraleBlock().GetValue()),
			energyText(game),
			num(game.GetScore().GetValue()),
			cores.PhaseName(game.GetPhase()),
			mode.name(),
		})
	return panelTitle("狀態列", width) + "\n" + boxTrunc(row1, width) + "\n" + boxTrunc(row2, width)
}

// energyText 出牌點數欄: 當前值 / 上限; 出牌點數保留(energyKeep)鎖定中加「保」標記(標記對照依
// 【營業顯示規格書 | 4、渲染政策：ASCII + CJK only】)。
func energyText(game *cores.Game) string {
	text := num(game.GetEnergy().GetValue()) + "/" + num(game.GetEnergyMax().GetValue())

	if game.GetEnergyKeep().IsLock() {
		text += "保"
	} // if

	return text
}
