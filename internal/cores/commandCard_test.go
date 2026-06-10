package cores

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/yinweli/RovingDiner/internal/exprs"
)

func TestSuiteCommandCard(t *testing.T) {
	suite.Run(t, new(SuiteCommandCard))
}

// SuiteCommandCard 驗證卡牌屬性命令(commandCard.go):cardCost*(寫鎖 / 捨入 / 夾 0)、cardEffect*(增 / 刪一 / 刪全)。
type SuiteCommandCard struct {
	suite.Suite
}

func (this *SuiteCommandCard) TestCardCost() {
	runtime := NewRuntime(0)
	card := &Card{InstanceID: 1, Cost: NewValue(3, 0)}
	runtime.Hand = []*Card{card}
	eng := this.engine(runtime)

	commandCardCostAdd(eng, []InstanceID{1}, nums(2)) // 3 + 2
	this.Equal(int32(5), card.Cost.GetValue())

	card.Cost = NewValue(3, 0)
	commandCardCostMul(eng, []InstanceID{1}, []exprs.Value{exprs.NewNum(1.5)}) // 3 * 1.5 = 4.5 → 5(捨入)
	this.Equal(int32(5), card.Cost.GetValue())

	commandCardCostSet(eng, []InstanceID{1}, nums(7)) // = 7
	this.Equal(int32(7), card.Cost.GetValue())

	commandCardCostSet(eng, []InstanceID{1}, nums(-5)) // 夾下限 0
	this.Equal(int32(0), card.Cost.GetValue())
}

func (this *SuiteCommandCard) TestCardCostNoop() {
	runtime := NewRuntime(0)
	card := &Card{InstanceID: 1, Cost: NewValue(3, 1)} // 鎖定
	runtime.Hand = []*Card{card}
	eng := this.engine(runtime)

	commandCardCostAdd(eng, []InstanceID{1}, nums(2)) // 鎖定 → no-op
	this.Equal(int32(3), card.Cost.GetValue())

	commandCardCostAdd(eng, []InstanceID{1}, nil) // N 缺漏 → 整動作 no-op
	this.Equal(int32(3), card.Cost.GetValue())

	commandCardCostAdd(eng, []InstanceID{99}, nums(2)) // 非卡牌實例 → 該項 no-op
	this.Equal(int32(3), card.Cost.GetValue())
}

func (this *SuiteCommandCard) TestCardEffect() {
	runtime := NewRuntime(0)
	card := &Card{InstanceID: 1, EffectID: []int32{20, 30, 20}}
	runtime.Hand = []*Card{card}
	eng := this.engine(runtime)

	commandCardEffectAdd(eng, []InstanceID{1}, nums(40))
	this.Equal([]int32{20, 30, 20, 40}, card.EffectID) // 尾端加入 40

	commandCardEffectDel(eng, []InstanceID{1}, nums(20))
	this.Equal([]int32{30, 20, 40}, card.EffectID) // 移除第一個 20

	commandCardEffectDel(eng, []InstanceID{1}, nums(99))
	this.Equal([]int32{30, 20, 40}, card.EffectID) // 卡上無 99 → no-op

	commandCardEffectDelAll(eng, []InstanceID{1}, nums(20))
	this.Equal([]int32{30, 40}, card.EffectID) // 移除全部 20

	commandCardEffectAdd(eng, []InstanceID{1}, nil)    // 效果編號缺漏 → no-op
	commandCardEffectDel(eng, []InstanceID{1}, nil)    // 同上
	commandCardEffectDelAll(eng, []InstanceID{1}, nil) // 同上
	this.Equal([]int32{30, 40}, card.EffectID)

	commandCardEffectAdd(eng, []InstanceID{99}, nums(50))    // 非卡牌實例 → no-op
	commandCardEffectDel(eng, []InstanceID{99}, nums(30))    // 同上
	commandCardEffectDelAll(eng, []InstanceID{99}, nums(30)) // 同上
	this.Equal([]int32{30, 40}, card.EffectID)
}

// === 測試輔助(置尾) ===

func (this *SuiteCommandCard) engine(runtime *Runtime) *Engine {
	return NewEngine(runtime, nil, buildSheet(), fakeOperator{}, fakeRander{}, nil)
}
