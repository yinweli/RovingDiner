package rules

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/yinweli/RovingDiner/internal/cores"
	"github.com/yinweli/RovingDiner/internal/tester"
)

func TestSuiteSettle(t *testing.T) {
	suite.Run(t, new(SuiteSettle))
}

// SuiteSettle 驗證執行結算流程（settle.go）:門檻命中與標記 / 離場與再定位防禦 / sate 鎖定暫停 / 終止判定哨兵 /
// 手牌上限棄牌 / 結算旗標不重入。
type SuiteSettle struct {
	suite.Suite
}

// TestSettle 驗證結算全程:飽食門檻入行動佇列（已觸發不重複）、飽食 / 生氣離場、sate 鎖定暫停飽食線、旗標不重入。
func (this *SuiteSettle) TestSettle() {
	data := tester.BuildData()
	game := newGameData(data)
	game.GetRoundMax().Set(99)
	game.GetMorale().Set(30)
	game.GetHandMax().Set(10)

	hungry := cores.NewGuest(game, 501) // 飽食 7:過門檻 6、未過 9;耐心 3 不動
	hungry.GetSate().Set(7)
	game.Seat.Place(1, hungry)

	full := cores.NewGuest(game, 501) // 飽食 12 = 離場線 → 飽食離場(給滿意值)
	full.GetSate().Set(12)
	full.GetScore().Set(4)
	game.Seat.Place(2, full)

	roam := cores.NewGuest(game, 501) // 遊蕩:sate 鎖定 → 飽食線暫停(門檻 / 離場皆跳過)
	roam.GetSate().Set(12)
	roam.RoamLock()
	game.Roam.Push(roam)

	survivor := cores.NewGuest(game, 502) // 高耐心無門檻 → 全程在座(擋住「全場清空」終止判定)
	game.Seat.Place(3, survivor)

	Settle(game)
	this.Require().Len(game.Action, 3) // hungry 命中 6;full 離場前命中 6 / 9;roam 暫停
	this.Equal(hungry, game.Action[0].GetGuest())
	this.Equal(cores.TaskSate, game.Action[0].GetKind())
	this.Equal(int32(301), game.Action[0].GetSkillID())
	this.True(hungry.GetSateHit().IsHit(6))
	this.False(hungry.GetSateHit().IsHit(9)) // 7 < 9 未達

	this.Nil(game.Seat[2])                           // 飽食離場
	this.Equal(int32(4), game.GetScore().GetValue()) // 給滿意值
	this.Equal(full, game.GetExitLast())
	this.Len(game.Roam, 1)    // sate 鎖定 → 不離場
	this.False(game.Settling) // 旗標除

	Settle(game)
	this.Len(game.Action, 3) // 已觸發門檻不重複入佇列

	// 生氣離場:耐心 0 → 扣士氣;遊蕩照常運作(解鎖後其門檻不再暫停,但本案驗證耐心線)
	hungry.GetCalm().Set(0)
	roam.GetCalm().Set(0)
	Settle(game)
	this.Nil(game.Seat[1])                             // 座位生氣離場
	this.Empty(game.Roam)                              // 遊蕩生氣離場照常
	this.Equal(int32(20), game.GetMorale().GetValue()) // 30 - 5 - 5
	this.Equal(survivor, game.Seat[3])                 // 倖存者在座

	// 結算旗標已立 → no-op
	game.Seat.Place(1, cores.NewGuest(game, 501))
	game.Settling = true
	Settle(game)
	this.True(game.Settling)
}

// TestSettleEmit 驗證執行結算的事件接線:無事結算（settleBusy 全不中）靜默不發題;有事結算先發範圍標題;
// 手牌上限棄牌發玩家輸入紀錄（流程名 discardOver）。
func (this *SuiteSettle) TestSettleEmit() {
	game, record := newGameRecord()
	game.GetRoundMax().Set(99)
	game.GetMorale().Set(30)
	game.GetHandMax().Set(2)
	game.Seat.Place(1, cores.NewGuest(game, 502)) // 高耐心、無門檻
	game.Seat.Place(2, cores.NewGuest(game, 502)) // 第二位:離場後仍有人在座,擋全場清空

	Settle(game) // 無事結算 → 靜默,不發題不立旗標
	this.Empty(record.Event)
	this.False(game.Settling)

	angry := game.Seat[1]
	angry.GetCalm().Set(0) // 生氣離場線 → 有事
	Settle(game)
	this.Require().NotEmpty(record.Event)
	this.Equal(cores.EventData{Kind: cores.EventScope, Scope: cores.ScopeSettle}, record.Event[0]) // 範圍標題先行
	this.Nil(game.Seat[1])

	record.Event = nil
	game.Hand = cores.CardList{cores.NewCard(game, 101), cores.NewCard(game, 101), cores.NewCard(game, 103)} // 3 > 上限 2 → 棄 1
	Settle(game)
	pick := []cores.EventData{}

	for itor := range record.Event {
		if record.Event[itor].Kind == cores.EventSelect {
			pick = append(pick, record.Event[itor])
		} // if
	} // for

	this.Require().Len(pick, 1) // 玩家輸入紀錄:流程名 + 選中清單
	this.Equal("discardOver", pick[0].Source)
	this.Equal([]cores.PickData{{DataID: 101, InstanceID: game.Drop[0].GetInstanceID()}}, pick[0].Pick)
}

// TestJudgeEnd 驗證終止判定:失敗(回合上限 / 士氣歸零,失敗優先)與成功(全場清空)以哨兵跳出、命中前清旗標;未命中無事。
func (this *SuiteSettle) TestJudgeEnd() {
	game := newGame()
	game.GetRoundMax().Set(5)
	game.GetMorale().Set(10)
	game.Seat.Place(1, cores.NewGuest(game, 501))

	judgeEnd(game) // 未命中 → 無事
	this.False(game.Settling)

	game.Settling = true
	game.GetRound().Set(5) // 回合 >= 上限 → 失敗
	this.PanicsWithValue(gameEnd{phase: cores.PhaseGameFail}, func() { judgeEnd(game) })
	this.False(game.Settling) // 命中前清旗標

	game.GetRound().Set(1)
	game.GetMorale().Set(0) // 士氣歸零 → 失敗
	this.PanicsWithValue(gameEnd{phase: cores.PhaseGameFail}, func() { judgeEnd(game) })

	game.GetMorale().Set(10)
	game.Seat.Remove(game.Seat[1]) // 全場清空 → 成功
	this.PanicsWithValue(gameEnd{phase: cores.PhaseGameSucc}, func() { judgeEnd(game) })

	game.GetRound().Set(5) // 失敗優先於成功:雙條件同時成立 → 失敗
	this.PanicsWithValue(gameEnd{phase: cores.PhaseGameFail}, func() { judgeEnd(game) })
}

// TestExitSate 驗證飽食離場的再定位防禦:離場觸發把後續候選搬離 → 略過不重複離場。
func (this *SuiteSettle) TestExitSate() {
	data := tester.BuildData()
	game := newGameData(data)
	first := cores.NewGuest(game, 501)
	first.GetSate().Set(12)
	game.Seat.Place(1, first)
	second := cores.NewGuest(game, 501)
	second.GetSate().Set(12)
	game.Seat.Place(2, second)

	data.SetEffect(901, cores.EffectData{Kind: cores.EffectTrigger, TriggerKind: cores.TriggerExitAny, Trigger: func(game *cores.Game) {
		game.Seat.Remove(second) // 首位離場觸發把第二位搬離座位 → 再定位略過
	}})
	game.Effect.Push(cores.NewEffect(game, 901, cores.Ref{}, 1))

	exitSate(game)
	this.Equal(first, game.GetExitLast()) // 僅首位走完離場流程
	this.Equal(int32(1), game.GetExitCount())
	this.Nil(game.Seat[1])
	this.Nil(game.Seat[2])
}

// TestDiscardOver 驗證手牌上限棄牌:逐張驗證棄置至達標;Operator 無進展 → 防禦跳出。
func (this *SuiteSettle) TestDiscardOver() {
	game := newGame()
	game.GetHandMax().Set(1)
	game.Hand = cores.CardList{cores.NewCard(game, 101), cores.NewCard(game, 101), cores.NewCard(game, 101)}

	discardOver(game) // FakeOperator 良性版:取前綴 2 張棄置
	this.Len(game.Hand, 1)
	this.Len(game.Drop, 2)
	this.Equal(int32(2), game.GetDropCount()) // 走 placeCard → 照設棄牌事件

	bad := cores.NewGame(0, 0, tester.BuildData(), badOperator{}, tester.FakeRander{}, nil)
	Register(bad)
	bad.GetHandMax().Set(1)
	bad.Hand = cores.CardList{cores.NewCard(bad, 101), cores.NewCard(bad, 101)}

	discardOver(bad) // 回傳場外卡 → 無進展 → 防禦跳出不掛死
	this.Len(bad.Hand, 2)
}

// === 測試輔助（置尾） ===

// badOperator 行為不良替身:PickDiscard 回傳不在手牌的卡（驗證 discardOver 無進展防禦）。
type badOperator struct {
	tester.FakeOperator
}

func (this badOperator) PickDiscard(source []*cores.Card, over int) []*cores.Card {
	return []*cores.Card{{}}
}
