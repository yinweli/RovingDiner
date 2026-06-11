package rodi

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/yinweli/RovingDiner/internal/cores"
)

func TestSuiteSeatPanel(t *testing.T) {
	suite.Run(t, new(SuiteSeatPanel))
}

// SuiteSeatPanel 驗證座位組件(seatPanel.go): 桌 strip / 2 行摘要 / 旗標列 / 固定桌欄寬 / 截斷記號。
type SuiteSeatPanel struct {
	suite.Suite
}

// TestSeatPanelView 驗證渲染: 桌號列 + 桌內 2 座各 2 行、桌欄固定寬 25、空位顯 空、超寬桌號列補右緣 >。
func (this *SuiteSeatPanel) TestSeatPanelView() {
	world := newMirror(testSheet())
	world.Apply(cores.EventData{Kind: cores.EventContainer, DataID: 501, InstanceID: 21, From: cores.ContainerNone, To: cores.ContainerWait})
	world.Apply(cores.EventData{Kind: cores.EventContainer, DataID: 501, InstanceID: 21, From: cores.ContainerWait, To: cores.ContainerSeat, SeatID: 1})
	world.Apply(cores.EventData{Kind: cores.EventEffect, DataID: 501, InstanceID: 21, EffectID: 401, EffectInstanceID: 41, Stage: cores.EffectStageJoin, Stack: 1, Alive: true})

	this.Equal(strings.Join([]string{ // 顧客 501: 封(資料 SateSeal)+ 效(佇列效果)命中、免 留白
		"+- 座位 " + strings.Repeat("-", 52),
		"桌1" + strings.Repeat(" ", 24) + "桌2",
		"501@老饕 飽0耐3" + strings.Repeat(" ", 12) + "空",
		strings.Repeat(" ", 9) + "封  效",
		"空",
		"",
	}, "\n"), seatPanel{}.View(world, 60))

	row := strings.Split(seatPanel{}.View(world, 10), "\n") // 超寬: 桌號列補右緣 >
	this.Equal("桌1"+strings.Repeat(" ", 6)+">", row[1])

	world.seat[2] = 99 // 防禦: 座位表有編號但視圖缺 → 顯 空
	this.Contains(seatPanel{}.View(world, 60), "空")
}
