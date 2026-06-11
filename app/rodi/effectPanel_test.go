package rodi

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/yinweli/RovingDiner/internal/cores"
)

func TestSuiteEffectPanel(t *testing.T) {
	suite.Run(t, new(SuiteEffectPanel))
}

// SuiteEffectPanel 驗證效果佇列組件(effectPanel.go): 作用順序排序 / 層數與剩餘回合 / self 辨型 / 類型標記。
type SuiteEffectPanel struct {
	suite.Suite
}

// TestEffectPanelView 驗證渲染: 作用順序大者優先、層數 1 省 xN、整場保留顯 永、剩餘回合 = 結束回合 - 當前回合。
func (this *SuiteEffectPanel) TestEffectPanelView() {
	world := newMirror(testSheet())
	world.Apply(cores.EventData{Kind: cores.EventContainer, DataID: 501, InstanceID: 21, From: cores.ContainerNone, To: cores.ContainerWait})
	world.Apply(cores.EventData{Kind: cores.EventEffect, Round: 2, DataID: 501, InstanceID: 21, EffectID: 401, EffectInstanceID: 41, Stage: cores.EffectStageJoin, Stack: 2, Expire: 5, Alive: true})
	world.Apply(cores.EventData{Kind: cores.EventEffect, Round: 2, EffectID: 402, EffectInstanceID: 42, Stage: cores.EffectStageJoin, Stack: 1, Expire: 0, Alive: true})
	world.Apply(cores.EventData{Kind: cores.EventEffect, Round: 2, EffectID: 403, EffectInstanceID: 43, Stage: cores.EffectStageJoin, Stack: 1, Expire: 0, Alive: true})

	this.Equal(strings.Join([]string{ // 作用順序 9 > 5 排前; 同序 402 < 403 編號小者優先
		"+- 效果佇列(3) " + strings.Repeat("-", 45),
		"402@護盾 (永)  403@立即 (永)  401@加耐x2 (3)",
		"空 常駐" + strings.Repeat(" ", 8) + "空 ?" + strings.Repeat(" ", 11) + "501@老饕 觸發",
	}, "\n"), effectPanel{}.View(world, 60))

	row := strings.Split(effectPanel{}.View(world, 12), "\n") // 超寬: 補右緣 >
	this.Equal("402@護盾 ( >", row[1])
}

// TestEffectOrder 驗證作用順序查表; 查無回 0。
func (this *SuiteEffectPanel) TestEffectOrder() {
	this.Equal(int32(9), effectOrder(testSheet(), 402))
	this.Equal(int32(0), effectOrder(testSheet(), 999))
}

// TestEffectSelf 驗證 self 主畫面投影: 顧客 / 卡牌辨型、空物件與失蹤實例顯 空。
func (this *SuiteEffectPanel) TestEffectSelf() {
	world := newMirror(testSheet())
	world.Apply(cores.EventData{Kind: cores.EventContainer, DataID: 501, InstanceID: 21, From: cores.ContainerNone, To: cores.ContainerWait})
	world.Apply(cores.EventData{Kind: cores.EventContainer, DataID: 101, InstanceID: 11, From: cores.ContainerNone, To: cores.ContainerHand})

	this.Equal("501@老饕", effectSelf(world, &effectView{selfID: 501, selfInstanceID: 21}))
	this.Equal("101@上菜", effectSelf(world, &effectView{selfID: 101, selfInstanceID: 11}))
	this.Equal("空", effectSelf(world, &effectView{}))                                // 空物件
	this.Equal("空", effectSelf(world, &effectView{selfID: 501, selfInstanceID: 99})) // 失蹤實例(防禦)
}

// TestEffectKindName 驗證類型標記: 觸發 / 常駐; 立即與查無顯 ?。
func (this *SuiteEffectPanel) TestEffectKindName() {
	this.Equal("觸發", effectKindName(testSheet(), 401))
	this.Equal("常駐", effectKindName(testSheet(), 402))
	this.Equal("?", effectKindName(testSheet(), 403)) // 立即不入列
	this.Equal("?", effectKindName(testSheet(), 999))
}
