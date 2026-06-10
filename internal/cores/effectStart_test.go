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
	immed := func(*Game) { immedRan++ }
	start := func(*Game) { startRan++ }

	game := NewGame(0, nil, nil, nil, nil)
	injectPort(game)
	game.effectData = map[int32]effectData{
		901: {Kind: EffectImmed, TargetKind: TargetNone, Immed: immed},
		902: {Kind: EffectImmed, TargetKind: TargetNone, Cond: this.expr("morale > 5"), Immed: immed},
		903: {Kind: EffectTrigger, TargetKind: TargetNone, Stack: 1},
		904: {Kind: EffectPersist, TargetKind: TargetNone, Stack: 2, Start: start},
	}

	runEffectList(game, []int32{901, 902, 903, 904, 999}, 0) // 999 查無 → 略過

	this.Equal(1, immedRan)            // 901 跑;902 條件假（morale 0）不跑
	this.Equal(2, startRan)            // 904 常駐 2 層 → 啟動命令 ×2
	this.Require().Len(game.Effect, 2) // 903 觸發 + 904 常駐 入列;立即不入列
}

func (this *SuiteEffectStart) TestSelectTargets() {
	g1 := &Guest{instanceID: 1, skillImmune: Immune{count: map[int32]int32{}}}
	g2 := &Guest{instanceID: 2, skillImmune: Immune{count: map[int32]int32{}}}
	g3 := &Guest{instanceID: 3, skillImmune: Immune{count: map[int32]int32{7: 1}}} // 技能群組 7 免疫
	game := NewGame(0, nil, nil, nil, nil)
	game.Seat[1], game.Seat[2], game.Seat[3] = g1, g2, g3
	game.Hand = []*Card{{instanceID: 10}, {instanceID: 11}}
	injectPort(game)

	// 無目標 → 一個空物件、清沿用來源
	inherit := []Ref{NewRefGuest(g1)}
	this.Equal([]Ref{{}}, selectTargets(game, effectData{TargetKind: TargetNone}, 0, &inherit))
	this.Nil(inherit)

	// 新選顧客 count 2、群組 7:候選 g1/g2（g3 免疫排除）、候選 2 ≤ 2 → 退化全取、設沿用來源
	got := selectTargets(game, effectData{TargetKind: TargetGuestPick, TargetCount: 2}, 7, &inherit)
	this.Equal([]Ref{NewRefGuest(g1), NewRefGuest(g2)}, got)
	this.Equal(got, inherit)

	// 新選顧客 count 1、群組 0:候選 3 > 1 → Operator.PickGuest 回前綴 [g1]
	this.Equal([]Ref{NewRefGuest(g1)}, selectTargets(game, effectData{TargetKind: TargetGuestPick, TargetCount: 1}, 0, &inherit))

	// 隨機顧客 count 1、群組 0:候選 3 > 1 → randSubset 決定性取 1
	this.Len(selectTargets(game, effectData{TargetKind: TargetGuestRand, TargetCount: 1}, 0, &inherit), 1)

	// 新選手牌 count 1:候選 2 > 1 → PickCard 回前綴 [card 10]
	this.Equal([]Ref{NewRefCard(game.Hand[0])}, selectTargets(game, effectData{TargetKind: TargetCardPick, TargetCount: 1}, 0, &inherit))

	// 隨機手牌 count 3:候選 2 ≤ 3 → 退化全取
	this.Len(selectTargets(game, effectData{TargetKind: TargetCardRand, TargetCount: 3}, 0, &inherit), 2)
}

func (this *SuiteEffectStart) TestSelectTargetsSame() {
	g1 := &Guest{instanceID: 1, skillImmune: Immune{count: map[int32]int32{}}}
	c1 := &Card{instanceID: 10}
	game := NewGame(0, nil, nil, nil, nil)
	game.Seat[1] = g1
	game.Hand = []*Card{c1}
	injectPort(game)

	// 沿用顧客:沿用來源有顧客 → 回沿用、沿用來源不變
	inherit := []Ref{NewRefGuest(g1)}
	this.Equal([]Ref{NewRefGuest(g1)}, selectTargets(game, effectData{TargetKind: TargetGuestSame, TargetCount: 1}, 0, &inherit))
	this.Equal([]Ref{NewRefGuest(g1)}, inherit)

	// 沿用顧客退化:沿用來源空 → 退化新選顧客（候選 g1）
	inherit = nil
	this.Equal([]Ref{NewRefGuest(g1)}, selectTargets(game, effectData{TargetKind: TargetGuestSame, TargetCount: 1}, 0, &inherit))

	// 沿用顧客退化:沿用來源為手牌（型別不符）→ 退化新選顧客
	inherit = []Ref{NewRefCard(c1)}
	this.Equal([]Ref{NewRefGuest(g1)}, selectTargets(game, effectData{TargetKind: TargetGuestSame, TargetCount: 1}, 0, &inherit))

	// 沿用手牌:沿用來源有手牌 → 回沿用
	inherit = []Ref{NewRefCard(c1)}
	this.Equal([]Ref{NewRefCard(c1)}, selectTargets(game, effectData{TargetKind: TargetCardSame, TargetCount: 1}, 0, &inherit))

	// 沿用手牌退化:沿用來源空 → 退化新選手牌
	inherit = nil
	this.Equal([]Ref{NewRefCard(c1)}, selectTargets(game, effectData{TargetKind: TargetCardSame, TargetCount: 1}, 0, &inherit))
}

// === 測試輔助（置尾） ===

// expr 解析運算式字串;語法錯即測試失敗。供觸發條件構造。
func (this *SuiteEffectStart) expr(source string) *exprs.Expr {
	result, err := exprs.Parse(source)
	this.Require().NoError(err)

	return result
}
