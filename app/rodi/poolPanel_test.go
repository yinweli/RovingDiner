package rodi

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/yinweli/RovingDiner/internal/cores"
)

func TestSuitePoolPanel(t *testing.T) {
	suite.Run(t, new(SuitePoolPanel))
}

// SuitePoolPanel 驗證場外組件(poolPanel.go): 三列格式 / 空列留白 / 截斷記號。
type SuitePoolPanel struct {
	suite.Suite
}

// TestPoolPanelView 驗證渲染: 排隊 / 遊蕩 / 卡牌化三列(隊頭在左)、空列冒號後留空、超寬補右緣 >。
func (this *SuitePoolPanel) TestPoolPanelView() {
	world := newMirror(testSheet())
	world.Apply(cores.EventData{Kind: cores.EventContainer, DataID: 501, InstanceID: 21, From: cores.ContainerNone, To: cores.ContainerWait})
	world.Apply(cores.EventData{Kind: cores.EventContainer, DataID: 501, InstanceID: 22, From: cores.ContainerNone, To: cores.ContainerWait})
	world.Apply(cores.EventData{Kind: cores.EventContainer, DataID: 501, InstanceID: 31, From: cores.ContainerNone, To: cores.ContainerCardify})

	this.Equal(strings.Join([]string{
		"+- 場外 " + strings.Repeat("-", 52),
		"排隊(2): 501@老饕 501@老饕",
		"遊蕩(0):",
		"卡牌化(1): 501@老饕",
	}, "\n"), poolPanel{}.View(world, 60))

	row := strings.Split(poolPanel{}.View(world, 12), "\n") // 超寬: 補右緣 >
	this.Equal("排隊(2): 5 >", row[1])

	world.zone[cores.ContainerRoam] = append(world.zone[cores.ContainerRoam], 99) // 防禦: 視圖缺 → 識別碼略過、計數照實
	this.Contains(poolPanel{}.View(world, 60), "遊蕩(1):")
}
