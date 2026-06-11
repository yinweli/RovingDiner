package rules

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/yinweli/RovingDiner/internal/cores"
	"github.com/yinweli/RovingDiner/internal/tester"
)

func TestSuitePhasePlayerAction(t *testing.T) {
	suite.Run(t, new(SuitePhasePlayerAction))
}

// SuitePhasePlayerAction 驗證玩家行動階段(phasePlayerAction.go): 補牌迴圈 / 玩家事件主迴圈 / 出牌(playCard)/
// 額外次數(extraCount)/ 玩家結束(playerEnd)。
type SuitePhasePlayerAction struct {
	suite.Suite
}

// TestPhasePlayerAction 驗證補牌(含棄牌堆洗回)與「封印不出 → 出牌 → 用盡結束」的完整事件序列。
func (this *SuitePhasePlayerAction) TestPhasePlayerAction() {
	count := 0
	op := &playOperator{}
	data := tester.BuildData()
	game := cores.NewGame(0, 0, data, op, tester.FakeRander{}, nil)
	Register(game)
	game.GetDrawMax().Set(6)
	game.GetEnergy().Set(5)
	data.SetEffect(401, cores.EffectData{Kind: cores.EffectImmed, Immed: func(game *cores.Game) { count++ }})
	data.SetEffect(402, cores.EffectData{Kind: cores.EffectImmed, Immed: func(game *cores.Game) { count++ }})

	game.Deck = cores.CardList{cores.NewCard(game, 101), cores.NewCard(game, 101), cores.NewCard(game, 101)}
	game.Drop = cores.CardList{cores.NewCard(game, 102)}

	play := cores.NewCard(game, 103) // 卡 103:Cost 2、Keep / Seal 鎖、技能 301 → 效果 401 / 402
	play.GetSeal().Unlock()          // 解封使其可出
	sealed := cores.NewCard(game, 103)
	game.Hand = cores.CardList{play, sealed}
	op.play = []*cores.Card{sealed, play} // 封印卡不執行出牌(迴圈續問)→ 出 play → 用盡 nil 結束

	this.Equal(cores.PhaseGuestAction, phasePlayerAction(game)) // 無跳轉 → 顧客行動
	this.Equal(int32(4), game.GetDrawCount())                   // 補 4 張至上限 6(抽牌堆 3 + 棄牌堆洗回 1)
	this.Equal(int32(3), game.GetEnergy().GetValue())           // 5 - 出牌費用 2
	this.Equal(2, count)                                        // 實例效果列表 401 / 402 各一次(額外次數 0)
	this.Equal(int32(1), game.GetPlayCount())
	this.Equal(play, game.GetPlayLast())
	this.Len(game.Hand, 1) // 玩家結束: 不棄(sealed)留手、其餘棄置
	this.Same(sealed, game.Hand[0])
	this.Empty(game.Deck)
	this.Len(game.Drop, 5) // 出牌 1 + 棄置 4
}

// TestPhasePlayerActionJump 驗證主迴圈頂的階段跳轉視為玩家結束(userStart / userEnd 各觸發一次, 不問 Operator)。
func (this *SuitePhasePlayerAction) TestPhasePlayerActionJump() {
	start := 0
	end := 0
	data := tester.BuildData()
	game := newGameData(data)
	data.SetEffect(901, cores.EffectData{Kind: cores.EffectTrigger, TriggerKind: cores.TriggerUserStart, Trigger: func(game *cores.Game) { start++ }})
	data.SetEffect(902, cores.EffectData{Kind: cores.EffectTrigger, TriggerKind: cores.TriggerUserEnd, Trigger: func(game *cores.Game) { end++ }})
	game.Effect.Push(cores.NewEffect(game, 901, cores.Ref{}, 1))
	game.Effect.Push(cores.NewEffect(game, 902, cores.Ref{}, 1))

	game.SetNextPhase(cores.PhaseRoundEnd)
	this.Equal(cores.PhaseRoundEnd, phasePlayerAction(game)) // 跳轉回合結束 → 玩家結束分派
	this.Equal(cores.PhaseNone, game.GetNextPhase())
	this.Equal(1, start)
	this.Equal(1, end)

	game.SetNextPhase(cores.PhaseGuestAction)
	this.Equal(cores.PhaseGuestAction, phasePlayerAction(game)) // 跳轉顧客行動 → 玩家結束預設分派
	this.Equal(cores.PhaseNone, game.GetNextPhase())
	this.Equal(2, end)
}

// TestPlayCard 驗證出牌閘門(非手牌 / 封印 / 卡牌化無空位 / 點數不足)、成功路徑與流放路由、卡牌化自動還原、效果搬卡防禦。
func (this *SuitePhasePlayerAction) TestPlayCard() {
	count := 0
	fired := 0
	data := tester.BuildData()
	game := newGameData(data)
	data.SetEffect(401, cores.EffectData{Kind: cores.EffectImmed, Immed: func(game *cores.Game) { count++ }})
	data.SetEffect(402, cores.EffectData{Kind: cores.EffectImmed, Immed: func(game *cores.Game) { count++ }})
	data.SetEffect(901, cores.EffectData{Kind: cores.EffectTrigger, TriggerKind: cores.TriggerCardPlay, Trigger: func(game *cores.Game) { fired++ }})
	game.Effect.Push(cores.NewEffect(game, 901, cores.Ref{}, 1))

	stray := cores.NewCard(game, 101)
	playCard(game, stray) // 防禦: 不在任何牌堆 → 不執行出牌
	this.Equal(int32(0), game.GetPlayCount())

	deckCard := cores.NewCard(game, 101)
	game.Deck = cores.CardList{deckCard}
	playCard(game, deckCard) // 位於抽牌牌堆(非手牌)→ 不執行出牌
	this.Equal(int32(0), game.GetPlayCount())

	sealed := cores.NewCard(game, 103)
	game.Hand = cores.CardList{sealed}
	playCard(game, sealed) // 封印 → 不執行出牌
	this.Equal(int32(0), game.GetPlayCount())

	play := cores.NewCard(game, 103)
	play.GetSeal().Unlock()
	game.Hand.Push(play)
	game.GetEnergy().Set(1)
	playCard(game, play) // 點數 1 < 費用 2 → 不執行出牌
	this.Equal(int32(0), game.GetPlayCount())

	game.GetEnergy().Set(5)
	playCard(game, play) // 成功: 扣點、效果、進棄牌堆、cardPlay 觸發
	this.Equal(int32(3), game.GetEnergy().GetValue())
	this.Equal(2, count)
	this.Equal(int32(1), game.GetPlayCount())
	this.Equal(1, fired)
	this.True(game.Drop.Has(play.GetInstanceID()))

	exilePlay := cores.NewCard(game, 101)
	exilePlay.GetPlayExile().Lock()
	game.Hand.Push(exilePlay)
	playCard(game, exilePlay) // 出牌後流放 → 進流放牌堆
	this.True(game.Exile.Has(exilePlay.GetInstanceID()))

	vanish := cores.NewCard(game, 101)
	vanish.GetEffectID().Add(902)
	data.SetEffect(902, cores.EffectData{Kind: cores.EffectImmed, Immed: func(game *cores.Game) { game.Hand.Remove(vanish.GetInstanceID()) }})
	game.Hand.Push(vanish)
	playCard(game, vanish) // 效果把本卡移出牌堆 → 路由略過(重新定位防禦)
	this.Equal(vanish, game.GetPlayLast())
	this.False(game.Drop.Has(vanish.GetInstanceID()))

	guest := cores.NewGuest(game, 501)
	bind := cores.NewCard(game, 101)
	bind.CardifyBind(guest)
	game.Cardify.Push(guest)
	game.Hand.Push(bind)
	game.Seat.Place(1, cores.NewGuest(game, 501)) // 座位全占
	game.Seat.Place(2, cores.NewGuest(game, 501))
	game.Seat.Place(3, cores.NewGuest(game, 501))
	playCard(game, bind) // 卡牌化卡無空位 → 不執行出牌
	this.NotEqual(bind, game.GetPlayLast())

	game.Seat.Remove(game.Seat[3]) // 空出座位 3
	playCard(game, bind)           // 卡牌化卡出牌 → 自動還原綁定顧客
	this.Equal(bind, game.GetPlayLast())
	this.Same(guest, game.Seat[3])
	this.Empty(game.Cardify)
	this.Nil(bind.GetCardify())
	this.True(game.Drop.Has(bind.GetInstanceID()))
}

// TestPlayCardEmit 驗證玩家出牌的發射接線: 範圍標題(操作元 = 卡牌 + 技能)先行、出牌耗能屬性行接續(流程寫入白名單)。
func (this *SuitePhasePlayerAction) TestPlayCardEmit() {
	game, record := newGameRecord()
	card := cores.NewCard(game, 103) // 費用 2、技能 301; 實例編號 1
	card.GetSeal().Unlock()
	game.Hand = cores.CardList{card}
	game.GetEnergy().Set(5)

	playCard(game, card)
	this.Require().True(len(record.Line) >= 2)
	this.Equal([]string{"[R0 -] 玩家出牌", "* 103@#1", "* 301@"}, record.Line[0])
	this.Equal([]string{"$ 出牌點數 -= 2 >> 3"}, record.Line[1])

	record.Line = nil
	sealed := cores.NewCard(game, 103) // 封印閘門擋下 → 不執行出牌、無行
	game.Hand.Push(sealed)
	playCard(game, sealed)
	this.Empty(record.Line)
}

// TestPlayerEndEmit 驗證玩家結束的發射接線: 手動結束範圍標題先行(流程, 無操作元; 階段跳轉路徑亦發)。
func (this *SuitePhasePlayerAction) TestPlayerEndEmit() {
	game, record := newGameRecord()
	playerEnd(game)
	this.Require().NotEmpty(record.Line)
	this.Equal([]string{"[R0 -] 手動結束"}, record.Line[0])
}

// TestExtraCount 驗證額外發動次數: Intn 取下限 / 帶位移、下限 > 上限與上限 < 0 為 0、負結果視為 0。
func (this *SuitePhasePlayerAction) TestExtraCount() {
	game := newGame() // FakeRander Intn 恆 0
	card := cores.NewCard(game, 101)
	this.Equal(int32(0), extraCount(game, card)) // [0, 0] → 0

	card.GetExtraRunMin().Set(2)
	card.GetExtraRunMax().Set(4)
	this.Equal(int32(2), extraCount(game, card)) // Intn 恆 0 → 下限

	gameN := cores.NewGame(0, 0, tester.BuildData(), tester.FakeOperator{}, tester.FakeRander{N: 2}, nil)
	this.Equal(int32(4), extraCount(gameN, card)) // 下限 2 + 位移 2

	card.GetExtraRunMin().Set(5)
	this.Equal(int32(0), extraCount(game, card)) // 下限 > 上限 → 0

	card.GetExtraRunMin().Set(-3)
	card.GetExtraRunMax().Set(-1)
	this.Equal(int32(0), extraCount(game, card)) // 上限 < 0 → 0

	card.GetExtraRunMax().Set(0)
	this.Equal(int32(0), extraCount(game, card)) // [-3, 0] 取得 -3 → 視為 0(防禦)
}

// TestPlayerEnd 驗證玩家結束: userEnd 觸發、剩餘手牌三路處理(流放 / 留手 / 棄置)、觸發搬卡防禦、下一階段分派。
func (this *SuitePhasePlayerAction) TestPlayerEnd() {
	end := 0
	data := tester.BuildData()
	game := newGameData(data)
	exileCard := cores.NewCard(game, 101)
	exileCard.GetUnplayExile().Lock()
	keepCard := cores.NewCard(game, 103) // Keep 鎖 → 留手
	dropCard := cores.NewCard(game, 101)
	stolen := cores.NewCard(game, 101)
	game.Hand = cores.CardList{exileCard, keepCard, dropCard, stolen}

	data.SetEffect(901, cores.EffectData{Kind: cores.EffectTrigger, TriggerKind: cores.TriggerUserEnd, Trigger: func(game *cores.Game) { end++ }})
	data.SetEffect(902, cores.EffectData{Kind: cores.EffectTrigger, TriggerKind: cores.TriggerCardExile, Trigger: func(game *cores.Game) {
		game.Hand.Remove(stolen.GetInstanceID()) // 快照後、處理中段觸發效果搬走手牌 → 後續迭代須略過
	}})
	game.Effect.Push(cores.NewEffect(game, 901, cores.Ref{}, 1))
	game.Effect.Push(cores.NewEffect(game, 902, cores.Ref{}, 1))

	this.Equal(cores.PhaseGuestAction, playerEnd(game)) // 預設 → 顧客行動
	this.Equal(1, end)
	this.True(game.Exile.Has(exileCard.GetInstanceID())) // 未出牌流放
	this.Len(game.Hand, 1)                               // 不棄留手
	this.Same(keepCard, game.Hand[0])
	this.True(game.Drop.Has(dropCard.GetInstanceID())) // 棄置
	this.False(game.Drop.Has(stolen.GetInstanceID()))  // 已被觸發效果搬離 → 略過

	game.SetNextPhase(cores.PhaseRoundEnd)
	this.Equal(cores.PhaseRoundEnd, playerEnd(game)) // 跳轉回合結束
	this.Equal(cores.PhaseNone, game.GetNextPhase())
}

// === 測試輔助(置尾) ===

// playOperator 出牌腳本替身: PlayerAction 依序回傳 play 佇列的卡、用盡回 nil(玩家結束); 其餘輸入沿用 FakeOperator。
type playOperator struct {
	tester.FakeOperator
	play []*cores.Card
	at   int
}

func (this *playOperator) PlayerAction(game *cores.Game) *cores.Card {
	if this.at >= len(this.play) {
		return nil
	} // if

	result := this.play[this.at]
	this.at++
	return result
}
