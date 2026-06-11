package rules

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/yinweli/RovingDiner/internal/cores"
	"github.com/yinweli/RovingDiner/internal/exprs"
)

func TestSuiteCommandCard(t *testing.T) {
	suite.Run(t, new(SuiteCommandCard))
}

// SuiteCommandCard 驗證卡牌屬性命令(commandCard.go): cardCost*(寫鎖 / 捨入 / 夾 0)、cardEffect*(增 / 刪一 / 刪全)。
type SuiteCommandCard struct {
	suite.Suite
}

func (this *SuiteCommandCard) TestCardCost() {
	game := newGame()
	card := cores.NewCard(game, 101)
	card.GetCost().Set(3)
	game.Hand = cores.CardList{card}
	id := []cores.InstanceID{card.GetInstanceID()}

	commandCardCostAdd(game, id, nums(2)) // 3 + 2
	this.Equal(int32(5), card.GetCost().GetValue())

	card.GetCost().Set(3)
	commandCardCostMul(game, id, []exprs.Value{exprs.NewNum(1.5)}) // 3 * 1.5 = 4.5 → 5(捨入)
	this.Equal(int32(5), card.GetCost().GetValue())

	commandCardCostSet(game, id, nums(7)) // = 7
	this.Equal(int32(7), card.GetCost().GetValue())

	commandCardCostSet(game, id, nums(-5)) // 夾下限 0
	this.Equal(int32(0), card.GetCost().GetValue())
}

func (this *SuiteCommandCard) TestCardCostNoop() {
	game := newGame()
	card := cores.NewCard(game, 101)
	card.GetCost().Set(3)
	card.GetCost().Lock() // 鎖定
	game.Hand = cores.CardList{card}
	id := []cores.InstanceID{card.GetInstanceID()}

	commandCardCostAdd(game, id, nums(2)) // 鎖定 → no-op
	this.Equal(int32(3), card.GetCost().GetValue())

	commandCardCostAdd(game, id, nil) // N 缺漏 → 整動作 no-op
	this.Equal(int32(3), card.GetCost().GetValue())

	commandCardCostAdd(game, []cores.InstanceID{99}, nums(2)) // 非卡牌實例 → 該項 no-op
	this.Equal(int32(3), card.GetCost().GetValue())
}

// TestCardCostEmit 驗證費用命令的屬性事件(數值型操作命令事件; M21 拍板): 逐卡包前後值、鎖定拒寫以 Before == After 表達。
func (this *SuiteCommandCard) TestCardCostEmit() {
	game, record := newGameRecord()
	card := cores.NewCard(game, 101)
	card.GetCost().Set(3)
	game.Hand = cores.CardList{card}
	id := []cores.InstanceID{card.GetInstanceID()}

	commandCardCostAdd(game, id, nums(2))
	this.Require().Len(record.Event, 1)
	this.Equal(cores.EventData{Kind: cores.EventProperty, DataID: 101, InstanceID: card.GetInstanceID(), Attr: "cost", Op: cores.AssignAdd, Operand: 2, Before: 3, After: 5}, record.Event[0])

	card.GetCost().Lock()
	commandCardCostAdd(game, id, nums(2)) // 鎖定拒寫 → Before == After
	this.Require().Len(record.Event, 2)
	this.Equal(float64(5), record.Event[1].Before)
	this.Equal(float64(5), record.Event[1].After)
}

func (this *SuiteCommandCard) TestCardEffect() {
	game := newGame()
	card := cores.NewCard(game, 101)
	card.GetEffectID().Add(20, 30, 20)
	game.Hand = cores.CardList{card}
	id := []cores.InstanceID{card.GetInstanceID()}

	commandCardEffectAdd(game, id, nums(40))
	this.Equal([]int32{20, 30, 20, 40}, card.GetEffectID().List()) // 尾端加入 40

	commandCardEffectDel(game, id, nums(20))
	this.Equal([]int32{30, 20, 40}, card.GetEffectID().List()) // 移除第一個 20

	commandCardEffectDel(game, id, nums(99))
	this.Equal([]int32{30, 20, 40}, card.GetEffectID().List()) // 卡上無 99 → no-op

	commandCardEffectDelAll(game, id, nums(20))
	this.Equal([]int32{30, 40}, card.GetEffectID().List()) // 移除全部 20

	commandCardEffectAdd(game, id, nil)    // 效果編號缺漏 → no-op
	commandCardEffectDel(game, id, nil)    // 同上
	commandCardEffectDelAll(game, id, nil) // 同上
	this.Equal([]int32{30, 40}, card.GetEffectID().List())

	commandCardEffectAdd(game, []cores.InstanceID{99}, nums(50))    // 非卡牌實例 → no-op
	commandCardEffectDel(game, []cores.InstanceID{99}, nums(30))    // 同上
	commandCardEffectDelAll(game, []cores.InstanceID{99}, nums(30)) // 同上
	this.Equal([]int32{30, 40}, card.GetEffectID().List())
}
