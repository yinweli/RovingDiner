package rules

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/yinweli/RovingDiner/internal/cores"
	"github.com/yinweli/RovingDiner/internal/exprs"
	"github.com/yinweli/RovingDiner/internal/tester"
)

func TestSuiteEffectStart(t *testing.T) {
	suite.Run(t, new(SuiteEffectStart))
}

// SuiteEffectStart 驗證啟動效果列表的目標選取(effectStart.go): 無目標 / 新選 / 隨機 / 沿用 × 顧客·手牌、退化、skillImmune、沿用來源。
type SuiteEffectStart struct {
	suite.Suite
}

func (this *SuiteEffectStart) TestRunEffectList() {
	immedRan := 0
	startRan := 0
	immed := func(*cores.Game) { immedRan++ }
	start := func(*cores.Game) { startRan++ }

	data := tester.BuildData()
	data.SetEffect(901, cores.EffectData{Kind: cores.EffectImmed, TargetKind: cores.TargetNone, Immed: immed})
	data.SetEffect(902, cores.EffectData{Kind: cores.EffectImmed, TargetKind: cores.TargetNone, Cond: this.expr("morale > 5"), Immed: immed})
	data.SetEffect(903, cores.EffectData{Kind: cores.EffectTrigger, TargetKind: cores.TargetNone, Stack: 1})
	data.SetEffect(904, cores.EffectData{Kind: cores.EffectPersist, TargetKind: cores.TargetNone, Stack: 2, Start: start})
	game := newGameData(data)

	runEffectList(game, []int32{901, 902, 903, 904, 999}, 0) // 999 查無 → 略過

	this.Equal(1, immedRan)            // 901 跑; 902 條件假(morale 0)不跑
	this.Equal(2, startRan)            // 904 常駐 2 層 → 啟動命令 ×2
	this.Require().Len(game.Effect, 2) // 903 觸發 + 904 常駐 入列; 立即不入列
}

// TestDispatchEffectEmit 驗證派發的效果頭: 立即 / 條件不成立(效果實例編號零值、識別碼省實例段)、
// 觸發入列僅 加入、常駐 加入 + 啟動。
func (this *SuiteEffectStart) TestDispatchEffectEmit() {
	data := tester.BuildData()
	game, record := newGameDataRecord(data)

	dispatchEffect(game, cores.EffectData{Kind: cores.EffectImmed}, 901, cores.Ref{}, 0) // 條件空 → 成立 → 立即
	this.Require().Len(record.Line, 1)
	this.Equal([]string{"- 901@?", "  空", "  立即"}, record.Line[0])

	dispatchEffect(game, cores.EffectData{Kind: cores.EffectImmed, Cond: this.expr("0")}, 901, cores.Ref{}, 0) // 條件不成立
	this.Require().Len(record.Line, 2)
	this.Equal([]string{"- 901@?", "  空", "  條件不成立"}, record.Line[1])

	record.Line = nil
	data.SetEffect(902, cores.EffectData{Kind: cores.EffectTrigger, Stack: 1})
	dispatchEffect(game, cores.EffectData{Kind: cores.EffectTrigger, Stack: 1}, 902, cores.Ref{}, 0) // 觸發入列 → 僅 加入(實例編號 1)
	this.Require().Len(record.Line, 1)
	this.Equal([]string{"- 902@?#1", "  空", "  加入"}, record.Line[0])

	record.Line = nil
	data.SetEffect(903, cores.EffectData{Kind: cores.EffectPersist, Stack: 2})
	dispatchEffect(game, cores.EffectData{Kind: cores.EffectPersist, Stack: 2}, 903, cores.Ref{}, 0) // 常駐 → 加入 + 啟動(實例編號 2)
	this.Require().Len(record.Line, 2)
	this.Equal([]string{"- 903@?#2", "  空", "  加入"}, record.Line[0])
	this.Equal([]string{"- 903@?#2", "  空", "  啟動"}, record.Line[1])
}

func (this *SuiteEffectStart) TestSelectTargets() {
	game := newGame()
	g1 := &cores.Guest{}
	g2 := &cores.Guest{}
	g3 := &cores.Guest{}
	g3.GetSkillImmune().Add(7) // 技能群組 7 免疫
	game.Seat[1], game.Seat[2], game.Seat[3] = g1, g2, g3
	game.Hand = cores.CardList{cores.NewCard(game, 101), cores.NewCard(game, 101)}

	// 無目標 → 一個空物件、清沿用來源
	inherit := []cores.Ref{cores.NewRefGuest(g1)}
	this.Equal([]cores.Ref{{}}, selectTargets(game, cores.EffectData{TargetKind: cores.TargetNone}, 0, 0, &inherit))
	this.Nil(inherit)

	// 新選顧客 count 2、群組 7:候選 g1/g2(g3 免疫排除)、候選 2 ≤ 2 → 退化全取、設沿用來源
	got := selectTargets(game, cores.EffectData{TargetKind: cores.TargetGuestPick, TargetCount: 2}, 0, 7, &inherit)
	this.Equal([]cores.Ref{cores.NewRefGuest(g1), cores.NewRefGuest(g2)}, got)
	this.Equal(got, inherit)

	// 新選顧客 count 1、群組 0:候選 3 > 1 → Operator.PickGuest 回前綴 [g1]
	this.Equal([]cores.Ref{cores.NewRefGuest(g1)}, selectTargets(game, cores.EffectData{TargetKind: cores.TargetGuestPick, TargetCount: 1}, 0, 0, &inherit))

	// 隨機顧客 count 1、群組 0:候選 3 > 1 → randSubset 決定性取 1
	this.Len(selectTargets(game, cores.EffectData{TargetKind: cores.TargetGuestRand, TargetCount: 1}, 0, 0, &inherit), 1)

	// 新選手牌 count 1:候選 2 > 1 → PickCard 回前綴 [card 10]
	this.Equal([]cores.Ref{cores.NewRefCard(game.Hand[0])}, selectTargets(game, cores.EffectData{TargetKind: cores.TargetCardPick, TargetCount: 1}, 0, 0, &inherit))

	// 隨機手牌 count 3:候選 2 ≤ 3 → 退化全取
	this.Len(selectTargets(game, cores.EffectData{TargetKind: cores.TargetCardRand, TargetCount: 3}, 0, 0, &inherit), 2)
}

func (this *SuiteEffectStart) TestSelectTargetsSame() {
	game := newGame()
	g1 := &cores.Guest{}
	c1 := cores.NewCard(game, 101)
	game.Seat[1] = g1
	game.Hand = cores.CardList{c1}

	// 沿用顧客: 沿用來源有顧客 → 回沿用、沿用來源不變
	inherit := []cores.Ref{cores.NewRefGuest(g1)}
	this.Equal([]cores.Ref{cores.NewRefGuest(g1)}, selectTargets(game, cores.EffectData{TargetKind: cores.TargetGuestSame, TargetCount: 1}, 0, 0, &inherit))
	this.Equal([]cores.Ref{cores.NewRefGuest(g1)}, inherit)

	// 沿用顧客退化: 沿用來源空 → 退化新選顧客(候選 g1)
	inherit = nil
	this.Equal([]cores.Ref{cores.NewRefGuest(g1)}, selectTargets(game, cores.EffectData{TargetKind: cores.TargetGuestSame, TargetCount: 1}, 0, 0, &inherit))

	// 沿用顧客退化: 沿用來源為手牌(型別不符)→ 退化新選顧客
	inherit = []cores.Ref{cores.NewRefCard(c1)}
	this.Equal([]cores.Ref{cores.NewRefGuest(g1)}, selectTargets(game, cores.EffectData{TargetKind: cores.TargetGuestSame, TargetCount: 1}, 0, 0, &inherit))

	// 沿用手牌: 沿用來源有手牌 → 回沿用
	inherit = []cores.Ref{cores.NewRefCard(c1)}
	this.Equal([]cores.Ref{cores.NewRefCard(c1)}, selectTargets(game, cores.EffectData{TargetKind: cores.TargetCardSame, TargetCount: 1}, 0, 0, &inherit))

	// 沿用手牌退化: 沿用來源空 → 退化新選手牌
	inherit = nil
	this.Equal([]cores.Ref{cores.NewRefCard(c1)}, selectTargets(game, cores.EffectData{TargetKind: cores.TargetCardSame, TargetCount: 1}, 0, 0, &inherit))
}

// TestSelectTargetsEmit 驗證目標選取的玩家輸入紀錄行: 真選取(交 Operator)發 $ 選取 行
// (效果目標選取顯效果識別碼 + 選中識別碼); 退化全取與系統隨機由 seed 涵蓋, 不記(M18 拍板)。
func (this *SuiteEffectStart) TestSelectTargetsEmit() {
	game, record := newGameRecord()
	g1 := cores.NewGuest(game, 501) // 實例編號 1
	g2 := cores.NewGuest(game, 501)
	game.Seat.Place(1, g1)
	game.Seat.Place(2, g2)
	inherit := []cores.Ref{}

	selectTargets(game, cores.EffectData{TargetKind: cores.TargetGuestPick, TargetCount: 1}, 401, 0, &inherit) // 候選 2 > 1 → 真選取
	this.Require().Len(record.Line, 1)
	this.Equal([]string{"$ 選取 401@ -> 501@#1"}, record.Line[0])

	record.Line = nil
	selectTargets(game, cores.EffectData{TargetKind: cores.TargetGuestPick, TargetCount: 5}, 401, 0, &inherit) // 退化全取 → 不記
	selectTargets(game, cores.EffectData{TargetKind: cores.TargetGuestRand, TargetCount: 1}, 401, 0, &inherit) // 系統隨機 → 不記
	this.Empty(record.Line)

	game.Hand = cores.CardList{cores.NewCard(game, 101), cores.NewCard(game, 103)}                            // 卡實例編號 3、4
	selectTargets(game, cores.EffectData{TargetKind: cores.TargetCardPick, TargetCount: 1}, 402, 0, &inherit) // 手牌真選取
	this.Require().Len(record.Line, 1)
	this.Equal([]string{"$ 選取 402@ -> 101@#3"}, record.Line[0])
}

// === 測試輔助(置尾) ===

// expr 解析運算式字串; 語法錯即測試失敗。供觸發條件構造。
func (this *SuiteEffectStart) expr(source string) *exprs.Expr {
	result, err := exprs.Parse(source)
	this.Require().NoError(err)
	return result
}
