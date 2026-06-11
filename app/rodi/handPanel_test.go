package rodi

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/yinweli/RovingDiner/internal/cores"
)

func TestSuiteHandPanel(t *testing.T) {
	suite.Run(t, new(SuiteHandPanel))
}

// SuiteHandPanel 驗證手牌組件(handPanel.go): 3 行卡塊 / cardify 來源段 / 旗標命中才顯 / 截斷記號。
type SuiteHandPanel struct {
	suite.Suite
}

// TestHandPanelView 驗證渲染: 標題含上限、卡名行帶費用與 cardify 來源、flag A / B 命中才顯、超寬補右緣 >。
func (this *SuiteHandPanel) TestHandPanelView() {
	world := newMirror(testSheet())
	world.Apply(cores.EventData{Kind: cores.EventProperty, Attr: "handMax", Op: cores.AssignSet, After: 5})
	world.Apply(cores.EventData{Kind: cores.EventContainer, DataID: 101, InstanceID: 11, From: cores.ContainerNone, To: cores.ContainerHand})
	world.Apply(cores.EventData{Kind: cores.EventContainer, DataID: 103, InstanceID: 12, From: cores.ContainerNone, To: cores.ContainerHand})
	world.Apply(cores.EventData{Kind: cores.EventContainer, DataID: 501, InstanceID: 31, From: cores.ContainerNone, To: cores.ContainerCardify})
	world.Apply(cores.EventData{Kind: cores.EventContainer, DataID: 101, InstanceID: 32, From: cores.ContainerNone, To: cores.ContainerHand, BindID: 501, BindInstanceID: 31})

	this.Equal(strings.Join([]string{ // 手牌序 = 新進入者在前: [32 12 11]
		"+- 手牌(3/5) " + strings.Repeat("-", 47),
		"101@上菜[501@老饕] (2)  103@結帳 (1)  101@上菜 (2)",
		"不棄 封印" + strings.Repeat(" ", 15) + "不棄" + strings.Repeat(" ", 10) + "封印",
		"",
	}, "\n"), handPanel{}.View(world, 60)) // 32: 綁定即不棄 + 資料封印; 12: 資料不棄; 11: 資料封印; flag B 全空留白

	row := strings.Split(handPanel{}.View(world, 12), "\n") // 超寬: 補右緣 >
	this.Equal("101@上菜[5 >", row[1])

	world.zone[cores.ContainerHand] = append(world.zone[cores.ContainerHand], 99) // 防禦: 視圖缺 → 該卡略過、計數照實
	this.Contains(handPanel{}.View(world, 60), "手牌(4/5)")
}

// TestHandFlagB 驗證 flag B: 出放 / 未放命中並列。
func (this *SuiteHandPanel) TestHandFlagB() {
	view := newCardView(testSheet(), 101)
	this.Equal("", handFlagB(view))

	view.lock["playExile"] = 1
	this.Equal("出放", handFlagB(view))

	view.lock["unplayExile"] = 2
	this.Equal("出放 未放", handFlagB(view))
}
