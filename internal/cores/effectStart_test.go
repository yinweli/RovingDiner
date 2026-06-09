package cores

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/yinweli/RovingDiner/internal/exprs"
)

func TestSuiteEffectStart(t *testing.T) {
	suite.Run(t, new(SuiteEffectStart))
}

// SuiteEffectStart 驗證啟動效果列表的目標選取（effectStart.go）:無目標 / 新選 / 隨機 / 沿用 × 顧客·手牌、退化、skillImmune、沿用來源。
type SuiteEffectStart struct {
	suite.Suite
}

func (this *SuiteEffectStart) TestRunEffectList() {
	immedRan := 0
	startRan := 0
	immed := func(*Engine) { immedRan++ }
	start := func(*Engine) { startRan++ }

	runtime := NewRuntime(0)
	eng := this.engine(runtime)
	eng.effect = map[int32]effectData{
		901: {Kind: EffectImmed, TargetKind: TargetNone, Immed: immed},
		902: {Kind: EffectImmed, TargetKind: TargetNone, Cond: this.expr("morale > 5"), Immed: immed},
		903: {Kind: EffectTrigger, TargetKind: TargetNone, Stack: 1},
		904: {Kind: EffectPersist, TargetKind: TargetNone, Stack: 2, Start: start},
	}

	runEffectList(eng, []int32{901, 902, 903, 904, 999}, 0) // 999 查無 → 略過

	this.Equal(1, immedRan)               // 901 跑;902 條件假（morale 0）不跑
	this.Equal(2, startRan)               // 904 常駐 2 層 → 啟動命令 ×2
	this.Require().Len(runtime.Effect, 2) // 903 觸發 + 904 常駐 入列;立即不入列
}

func (this *SuiteEffectStart) TestSelectTargets() {
	g1 := &Guest{InstanceID: 1, SkillImmune: map[int32]int32{}}
	g2 := &Guest{InstanceID: 2, SkillImmune: map[int32]int32{}}
	g3 := &Guest{InstanceID: 3, SkillImmune: map[int32]int32{7: 1}} // 技能群組 7 免疫
	runtime := NewRuntime(0)
	runtime.Seat[1], runtime.Seat[2], runtime.Seat[3] = g1, g2, g3
	runtime.Hand = []*Card{{InstanceID: 10}, {InstanceID: 11}}
	eng := this.engine(runtime)

	// 無目標 → 一個空物件、清沿用來源
	inherit := []Self{{Guest: g1}}
	this.Equal([]Self{{}}, selectTargets(eng, effectData{TargetKind: TargetNone}, 0, &inherit))
	this.Nil(inherit)

	// 新選顧客 count 2、群組 7:候選 g1/g2（g3 免疫排除）、候選 2 ≤ 2 → 退化全取、設沿用來源
	got := selectTargets(eng, effectData{TargetKind: TargetGuestPick, TargetCount: 2}, 7, &inherit)
	this.Equal([]Self{{Guest: g1}, {Guest: g2}}, got)
	this.Equal(got, inherit)

	// 新選顧客 count 1、群組 0:候選 3 > 1 → Operator.PickGuest 回前綴 [g1]
	this.Equal([]Self{{Guest: g1}}, selectTargets(eng, effectData{TargetKind: TargetGuestPick, TargetCount: 1}, 0, &inherit))

	// 隨機顧客 count 1、群組 0:候選 3 > 1 → randSubset 決定性取 1
	this.Len(selectTargets(eng, effectData{TargetKind: TargetGuestRand, TargetCount: 1}, 0, &inherit), 1)

	// 新選手牌 count 1:候選 2 > 1 → PickCard 回前綴 [card 10]
	this.Equal([]Self{{Card: runtime.Hand[0]}}, selectTargets(eng, effectData{TargetKind: TargetCardPick, TargetCount: 1}, 0, &inherit))

	// 隨機手牌 count 3:候選 2 ≤ 3 → 退化全取
	this.Len(selectTargets(eng, effectData{TargetKind: TargetCardRand, TargetCount: 3}, 0, &inherit), 2)
}

func (this *SuiteEffectStart) TestSelectTargetsSame() {
	g1 := &Guest{InstanceID: 1, SkillImmune: map[int32]int32{}}
	c1 := &Card{InstanceID: 10}
	runtime := NewRuntime(0)
	runtime.Seat[1] = g1
	runtime.Hand = []*Card{c1}
	eng := this.engine(runtime)

	// 沿用顧客:沿用來源有顧客 → 回沿用、沿用來源不變
	inherit := []Self{{Guest: g1}}
	this.Equal([]Self{{Guest: g1}}, selectTargets(eng, effectData{TargetKind: TargetGuestSame, TargetCount: 1}, 0, &inherit))
	this.Equal([]Self{{Guest: g1}}, inherit)

	// 沿用顧客退化:沿用來源空 → 退化新選顧客（候選 g1）
	inherit = nil
	this.Equal([]Self{{Guest: g1}}, selectTargets(eng, effectData{TargetKind: TargetGuestSame, TargetCount: 1}, 0, &inherit))

	// 沿用顧客退化:沿用來源為手牌（型別不符）→ 退化新選顧客
	inherit = []Self{{Card: c1}}
	this.Equal([]Self{{Guest: g1}}, selectTargets(eng, effectData{TargetKind: TargetGuestSame, TargetCount: 1}, 0, &inherit))

	// 沿用手牌:沿用來源有手牌 → 回沿用
	inherit = []Self{{Card: c1}}
	this.Equal([]Self{{Card: c1}}, selectTargets(eng, effectData{TargetKind: TargetCardSame, TargetCount: 1}, 0, &inherit))

	// 沿用手牌退化:沿用來源空 → 退化新選手牌
	inherit = nil
	this.Equal([]Self{{Card: c1}}, selectTargets(eng, effectData{TargetKind: TargetCardSame, TargetCount: 1}, 0, &inherit))
}

// === 測試輔助（置尾） ===

func (this *SuiteEffectStart) engine(runtime *Runtime) *Engine {
	return NewEngine(runtime, nil, buildSheet(), fakeOperator{}, fakeRander{}, nil)
}

// expr 解析運算式字串;語法錯即測試失敗。供觸發條件構造。
func (this *SuiteEffectStart) expr(source string) *exprs.Expr {
	result, err := exprs.Parse(source)
	this.Require().NoError(err)

	return result
}
