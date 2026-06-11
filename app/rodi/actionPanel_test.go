package rodi

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/yinweli/RovingDiner/internal/cores"
)

func TestSuiteActionPanel(t *testing.T) {
	suite.Run(t, new(SuiteActionPanel))
}

// SuiteActionPanel 驗證行動佇列組件(actionPanel.go): 2 行項橫向並排 / 行動類型標記 / 截斷記號。
type SuiteActionPanel struct {
	suite.Suite
}

// TestActionPanelView 驗證渲染: 隊頭在左、技能 / 顧客兩行 content-fit 對齊、空佇列兩行留白、超寬補右緣 >。
func (this *SuiteActionPanel) TestActionPanelView() {
	world := newMirror(testSheet())
	this.Equal("+- 行動佇列(0) "+strings.Repeat("-", 45)+"\n\n", actionPanel{}.View(world, 60)) // 空佇列留白

	world.Apply(cores.EventData{Kind: cores.EventContainer, DataID: 501, InstanceID: 21, From: cores.ContainerNone, To: cores.ContainerWait})
	world.Apply(cores.EventData{Kind: cores.EventAction, DataID: 501, InstanceID: 21, SkillID: 301, Task: cores.TaskCalm, Alive: true})
	world.Apply(cores.EventData{Kind: cores.EventAction, DataID: 501, InstanceID: 21, SkillID: 301, Task: cores.TaskSate, Alive: true})

	this.Equal(strings.Join([]string{
		"+- 行動佇列(2) " + strings.Repeat("-", 45),
		"301@開朗     301@開朗",
		"501@老饕 耐  501@老饕 飽",
	}, "\n"), actionPanel{}.View(world, 60))

	row := strings.Split(actionPanel{}.View(world, 10), "\n") // 超寬: 補右緣 >
	this.Equal("301@開朗 >", row[1])
}

// TestTaskName 驗證行動類型標記; 越界顯 ?。
func (this *SuiteActionPanel) TestTaskName() {
	this.Equal("飽", taskName(cores.TaskSate))
	this.Equal("耐", taskName(cores.TaskCalm))
	this.Equal("?", taskName(cores.TaskKind(9)))
}
