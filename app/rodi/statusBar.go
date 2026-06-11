package rodi

import (
	"strconv"

	"github.com/yinweli/RovingDiner/internal/cores"
)

// statusBar 狀態列組件(區 7; 【營業顯示規格書 | 6、畫面規格 | 6.9】): 兩行對齊表——第 1 行標籤、第 2 行數值,
// 欄序 = 回合 / 士氣 / 護盾 / 格擋 / 出牌點數 / 滿意 / 階段 / 模式; 數值全取世界鏡像(全域屬性 + 座標), 無自持狀態。
// 模式欄 M25 前固定顯「快速」、階段欄的「當前 > 下一」細節留 M22(M20 拍板); 回合計數與 seed 不入列(歸 M23 計數 modal)。
type statusBar struct{}

// View 渲染兩行對齊表; 超寬依預算截斷。
func (this statusBar) View(world *mirror, width int) string {
	row1, row2 := alignPair(
		[]string{"回合", "士氣", "護盾", "格擋", "出牌點數", "滿意", "階段", "模式"},
		[]string{
			num(float64(world.round)) + "/" + num(world.attr["roundMax"]),
			num(world.attr["morale"]) + "/" + num(world.attr["moraleMax"]),
			numFloor(world.attr["moraleShield"]),
			numFloor(world.attr["moraleBlock"]),
			energyText(world),
			num(world.attr["score"]),
			phaseName(world.phase),
			"快速",
		})
	return truncTo(row1, width) + "\n" + truncTo(row2, width)
}

// num 整數屬性值轉顯示字串(屬性容器為整數, 後值小數部位必為零)。
func num(value float64) string {
	return strconv.FormatInt(int64(value), 10)
}

// numFloor 護盾 / 格擋專用: <= 0 顯 0(【營業顯示規格書 | 6、畫面規格 | 6.9】固定欄)。
func numFloor(value float64) string {
	if value <= 0 {
		return "0"
	} // if

	return num(value)
}

// energyText 出牌點數欄: 當前值 / 上限; 出牌點數保留(energyKeep)鎖定中加「保」標記(標記對照依
// 【營業顯示規格書 | 4、渲染政策：ASCII + CJK only】)。
func energyText(world *mirror) string {
	text := num(world.attr["energy"]) + "/" + num(world.attr["energyMax"])

	if world.lock["energyKeep"] > 0 {
		text += "保"
	} // if

	return text
}

// phaseName 階段中文全名(【營業顯示規格書 | 6、畫面規格 | 6.10】座標用全名); 無階段 / 未知顯 -。
func phaseName(phase cores.PhaseKind) string {
	switch phase {
	case cores.PhaseGameStart:
		return "營業開始"

	case cores.PhaseRoundStart:
		return "回合開始"

	case cores.PhasePlayerAction:
		return "玩家行動"

	case cores.PhaseGuestAction:
		return "顧客行動"

	case cores.PhaseRoundEnd:
		return "回合結束"

	case cores.PhaseGameSucc:
		return "營業成功"

	case cores.PhaseGameFail:
		return "營業失敗"

	default:
		return "-"
	} // switch
}
