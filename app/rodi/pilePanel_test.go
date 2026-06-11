package rodi

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/yinweli/RovingDiner/internal/cores"
)

func TestSuitePilePanel(t *testing.T) {
	suite.Run(t, new(SuitePilePanel))
}

// SuitePilePanel 驗證牌堆組件(pilePanel.go): 三堆列格式(堆頂在左)/ 空堆留白 / 截斷記號。
type SuitePilePanel struct {
	suite.Suite
}

// TestPilePanelView 驗證渲染: 抽 / 棄 / 流放三列、堆頂在左、空堆冒號後留空、超寬補右緣 >。
func (this *SuitePilePanel) TestPilePanelView() {
	world := newMirror(testSheet())
	world.Apply(cores.EventData{Kind: cores.EventContainer, DataID: 101, InstanceID: 11, From: cores.ContainerNone, To: cores.ContainerDeck})
	world.Apply(cores.EventData{Kind: cores.EventContainer, DataID: 103, InstanceID: 12, From: cores.ContainerNone, To: cores.ContainerDeck})
	world.Apply(cores.EventData{Kind: cores.EventContainer, DataID: 101, InstanceID: 13, From: cores.ContainerNone, To: cores.ContainerDrop})

	this.Equal(strings.Join([]string{ // 牌堆序 = 新進入者置頂: 抽 [12 11]
		"+- 牌堆 " + strings.Repeat("-", 52),
		"抽牌堆(2): 103@結帳 101@上菜",
		"棄牌堆(1): 101@上菜",
		"流放堆(0):",
	}, "\n"), pilePanel{}.View(world, 60))

	row := strings.Split(pilePanel{}.View(world, 14), "\n") // 超寬: 補右緣 >
	this.Equal("抽牌堆(2): 1 >", row[1])
}
