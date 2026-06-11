package rodi

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/yinweli/RovingDiner/internal/cores"
)

func TestSuitePanelSeat(t *testing.T) {
	suite.Run(t, new(SuitePanelSeat))
}

// SuitePanelSeat 驗證座位組件(panelSeat.go): 桌 strip / 2 行摘要 / 旗標列 / 固定桌欄寬 / 截斷記號。
type SuitePanelSeat struct {
	suite.Suite
}

// TestPanelSeatView 驗證渲染: 桌號列 + 桌內 2 座各 2 行、桌欄固定寬 25、空位顯 空、超寬桌號列補右緣 >。
func (this *SuitePanelSeat) TestPanelSeatView() {
	game := testGame()
	guest := cores.NewGuest(game, 501)
	game.Seat.Place(1, guest)
	game.Effect.Push(cores.NewEffect(game, 401, cores.NewRefGuest(guest), 1))

	this.Equal(strings.Join([]string{ // 顧客 501: 封(資料 SateSeal)+ 效(佇列效果)命中、免 留白
		"+- 座位 " + strings.Repeat("-", 52),
		"桌1" + strings.Repeat(" ", 24) + "桌2",
		"501@老饕 飽0耐3" + strings.Repeat(" ", 12) + "空",
		strings.Repeat(" ", 9) + "封  效",
		"空",
		"",
	}, "\n"), panelSeat{}.View(game, 60))

	guest.GetEffectImmune().Add(5) // 免疫計數 > 0 → 免 點亮(M22 拍板)
	this.Contains(panelSeat{}.View(game, 60), strings.Repeat(" ", 9)+"封免效")

	row := strings.Split(panelSeat{}.View(game, 10), "\n") // 超寬: 桌號列補右緣 >
	this.Equal("桌1"+strings.Repeat(" ", 6)+">", row[1])

	game.Seat[2] = nil // 防禦: 座位表 nil 項 → 顯 空
	this.Contains(panelSeat{}.View(game, 60), "空")

	game.Effect = nil // 無效果 → 效 槽熄滅(hasEffect 掃完未命中)
	this.NotContains(panelSeat{}.View(game, 60), "效")
}
