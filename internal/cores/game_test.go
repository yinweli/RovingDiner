package cores

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

func TestSuiteGame(t *testing.T) {
	suite.Run(t, new(SuiteGame))
}

// SuiteGame 驗證營業實例(game.go):建構 / 取值 / 階段設定 / 事件成組寫入 / 回合計數歸零。
type SuiteGame struct {
	suite.Suite
}

// TestNewGame 驗證 NewGame 建構空白營業實例:屬性零值、無跳轉階段、累積計數零值可用。
func (this *SuiteGame) TestNewGame() {
	game := NewGame()
	this.Require().NotNil(game)
	this.Equal(int32(0), game.GetMorale().GetValue())
	this.Equal(PhaseNone, game.GetNextPhase())
	this.Equal(int32(0), game.GetDrawTotal().Sum())
	this.Nil(game.GetDamageGuest())
}

// TestGameGetMorale 驗證 GetMorale 取回餐廳士氣值組件（寫入經組件可見）。
func (this *SuiteGame) TestGameGetMorale() {
	game := &Game{morale: NewValue(25, 0)}
	this.Equal(int32(25), game.GetMorale().GetValue())

	game.GetMorale().Add(5) // 經組件寫入 → 同一實體
	this.Equal(int32(30), game.GetMorale().GetValue())
}

// TestGameGetMoraleMax 驗證 GetMoraleMax 取回餐廳士氣值上限。
func (this *SuiteGame) TestGameGetMoraleMax() {
	this.Equal(int32(50), (&Game{moraleMax: NewValue(50, 0)}).GetMoraleMax().GetValue())
}

// TestGameGetMoraleShield 驗證 GetMoraleShield 取回餐廳士氣值護盾。
func (this *SuiteGame) TestGameGetMoraleShield() {
	this.Equal(int32(9), (&Game{moraleShield: NewValue(9, 0)}).GetMoraleShield().GetValue())
}

// TestGameGetMoraleBlock 驗證 GetMoraleBlock 取回餐廳士氣值格擋。
func (this *SuiteGame) TestGameGetMoraleBlock() {
	this.Equal(int32(8), (&Game{moraleBlock: NewValue(8, 0)}).GetMoraleBlock().GetValue())
}

// TestGameGetScore 驗證 GetScore 取回餐廳滿意值。
func (this *SuiteGame) TestGameGetScore() {
	this.Equal(int32(7), (&Game{score: NewValue(7, 0)}).GetScore().GetValue())
}

// TestGameGetEnergy 驗證 GetEnergy 取回出牌點數。
func (this *SuiteGame) TestGameGetEnergy() {
	this.Equal(int32(3), (&Game{energy: NewValue(3, 0)}).GetEnergy().GetValue())
}

// TestGameGetEnergyMax 驗證 GetEnergyMax 取回出牌點數上限。
func (this *SuiteGame) TestGameGetEnergyMax() {
	this.Equal(int32(6), (&Game{energyMax: NewValue(6, 0)}).GetEnergyMax().GetValue())
}

// TestGameGetEnergyKeep 驗證 GetEnergyKeep 取回出牌點數保留（純鎖屬性,狀態於鎖定計數）。
func (this *SuiteGame) TestGameGetEnergyKeep() {
	this.Equal(int32(1), (&Game{energyKeep: NewValue(0, 1)}).GetEnergyKeep().GetLock())
}

// TestGameGetHandMax 驗證 GetHandMax 取回手牌張數上限。
func (this *SuiteGame) TestGameGetHandMax() {
	this.Equal(int32(10), (&Game{handMax: NewValue(10, 0)}).GetHandMax().GetValue())
}

// TestGameGetDrawMax 驗證 GetDrawMax 取回補牌張數上限。
func (this *SuiteGame) TestGameGetDrawMax() {
	this.Equal(int32(5), (&Game{drawMax: NewValue(5, 0)}).GetDrawMax().GetValue())
}

// TestGameGetNextPhase 驗證 GetNextPhase 取回下一階段（無跳轉目標為 PhaseNone）。
func (this *SuiteGame) TestGameGetNextPhase() {
	this.Equal(PhaseNone, (&Game{}).GetNextPhase())
	this.Equal(PhasePlayerAction, (&Game{nextPhase: PhasePlayerAction}).GetNextPhase())
}

// TestGameSetNextPhase 驗證 SetNextPhase 設下一階段（含清回 PhaseNone）。
func (this *SuiteGame) TestGameSetNextPhase() {
	game := &Game{}

	game.SetNextPhase(PhasePlayerAction)
	this.Equal(PhasePlayerAction, game.GetNextPhase())

	game.SetNextPhase(PhaseNone) // 系統轉移時清空
	this.Equal(PhaseNone, game.GetNextPhase())
}

// TestGameGetRound 驗證 GetRound 取回當前回合數。
func (this *SuiteGame) TestGameGetRound() {
	this.Equal(int32(4), (&Game{round: NewValue(4, 0)}).GetRound().GetValue())
}

// TestGameGetRoundMax 驗證 GetRoundMax 取回回合上限。
func (this *SuiteGame) TestGameGetRoundMax() {
	this.Equal(int32(12), (&Game{roundMax: NewValue(12, 0)}).GetRoundMax().GetValue())
}

// TestGameGetDamageValue 驗證 GetDamageValue 取回士氣受損值。
func (this *SuiteGame) TestGameGetDamageValue() {
	this.Equal(int32(6), (&Game{damageValue: 6}).GetDamageValue())
}

// TestGameGetDamageGuest 驗證 GetDamageGuest 取回士氣受損顧客（無來源回 nil）。
func (this *SuiteGame) TestGameGetDamageGuest() {
	guest := &Guest{instanceID: 9}
	this.Same(guest, (&Game{damageGuest: guest}).GetDamageGuest())
	this.Nil((&Game{}).GetDamageGuest()) // 無顧客來源 → nil
}

// TestGameGetSeatLast 驗證 GetSeatLast 取回最後入座顧客。
func (this *SuiteGame) TestGameGetSeatLast() {
	guest := &Guest{instanceID: 9}
	this.Same(guest, (&Game{seatLast: guest}).GetSeatLast())
}

// TestGameGetSeatCount 驗證 GetSeatCount 取回回合入座人數。
func (this *SuiteGame) TestGameGetSeatCount() {
	this.Equal(int32(2), (&Game{seatCount: 2}).GetSeatCount())
}

// TestGameGetExitLast 驗證 GetExitLast 取回最後離場顧客。
func (this *SuiteGame) TestGameGetExitLast() {
	guest := &Guest{instanceID: 9}
	this.Same(guest, (&Game{exitLast: guest}).GetExitLast())
}

// TestGameGetExitLastSeat 驗證 GetExitLastSeat 取回最後離場座位。
func (this *SuiteGame) TestGameGetExitLastSeat() {
	this.Equal(int32(3), (&Game{exitLastSeat: 3}).GetExitLastSeat())
}

// TestGameGetExitCount 驗證 GetExitCount 取回回合離場人數。
func (this *SuiteGame) TestGameGetExitCount() {
	this.Equal(int32(1), (&Game{exitCount: 1}).GetExitCount())
}

// TestGameGetTaskGuest 驗證 GetTaskGuest 取回最後行動顧客。
func (this *SuiteGame) TestGameGetTaskGuest() {
	guest := &Guest{instanceID: 9}
	this.Same(guest, (&Game{taskGuest: guest}).GetTaskGuest())
}

// TestGameGetTaskSkill 驗證 GetTaskSkill 取回最後行動技能編號。
func (this *SuiteGame) TestGameGetTaskSkill() {
	this.Equal(int32(77), (&Game{taskSkill: 77}).GetTaskSkill())
}

// TestGameGetTaskCount 驗證 GetTaskCount 取回回合行動次數。
func (this *SuiteGame) TestGameGetTaskCount() {
	this.Equal(int32(4), (&Game{taskCount: 4}).GetTaskCount())
}

// TestGameGetDrawLast 驗證 GetDrawLast 取回最後抽出卡牌。
func (this *SuiteGame) TestGameGetDrawLast() {
	card := &Card{instanceID: 1}
	this.Same(card, (&Game{drawLast: card}).GetDrawLast())
}

// TestGameGetDrawCount 驗證 GetDrawCount 取回回合抽牌張數。
func (this *SuiteGame) TestGameGetDrawCount() {
	this.Equal(int32(2), (&Game{drawCount: 2}).GetDrawCount())
}

// TestGameGetDrawTotal 驗證 GetDrawTotal 取回累積抽牌計數組件（零值可用、寫入經組件可見）。
func (this *SuiteGame) TestGameGetDrawTotal() {
	game := &Game{}
	game.GetDrawTotal().Add(1)
	this.Equal(int32(1), game.GetDrawTotal().Get(1))
}

// TestGameGetDropLast 驗證 GetDropLast 取回最後棄置卡牌。
func (this *SuiteGame) TestGameGetDropLast() {
	card := &Card{instanceID: 1}
	this.Same(card, (&Game{dropLast: card}).GetDropLast())
}

// TestGameGetDropCount 驗證 GetDropCount 取回回合棄牌張數。
func (this *SuiteGame) TestGameGetDropCount() {
	this.Equal(int32(1), (&Game{dropCount: 1}).GetDropCount())
}

// TestGameGetDropTotal 驗證 GetDropTotal 取回累積棄牌計數組件。
func (this *SuiteGame) TestGameGetDropTotal() {
	game := &Game{}
	game.GetDropTotal().Add(2)
	this.Equal(int32(1), game.GetDropTotal().Get(2))
}

// TestGameGetPlayLast 驗證 GetPlayLast 取回最後出牌卡牌。
func (this *SuiteGame) TestGameGetPlayLast() {
	card := &Card{instanceID: 1}
	this.Same(card, (&Game{playLast: card}).GetPlayLast())
}

// TestGameGetPlayCount 驗證 GetPlayCount 取回回合出牌張數。
func (this *SuiteGame) TestGameGetPlayCount() {
	this.Equal(int32(3), (&Game{playCount: 3}).GetPlayCount())
}

// TestGameGetPlayTotal 驗證 GetPlayTotal 取回累積出牌計數組件。
func (this *SuiteGame) TestGameGetPlayTotal() {
	game := &Game{}
	game.GetPlayTotal().Add(1)
	this.Equal(int32(1), game.GetPlayTotal().Get(1))
}

// TestGameGetExileLast 驗證 GetExileLast 取回最後流放卡牌。
func (this *SuiteGame) TestGameGetExileLast() {
	card := &Card{instanceID: 1}
	this.Same(card, (&Game{exileLast: card}).GetExileLast())
}

// TestGameGetExileCount 驗證 GetExileCount 取回回合流放張數。
func (this *SuiteGame) TestGameGetExileCount() {
	this.Equal(int32(1), (&Game{exileCount: 1}).GetExileCount())
}

// TestGameGetExileTotal 驗證 GetExileTotal 取回累積流放計數組件。
func (this *SuiteGame) TestGameGetExileTotal() {
	game := &Game{}
	game.GetExileTotal().Add(1)
	this.Equal(int32(1), game.GetExileTotal().Get(1))
}

// TestGameGetMorphLast 驗證 GetMorphLast 取回最後變身卡牌。
func (this *SuiteGame) TestGameGetMorphLast() {
	card := &Card{instanceID: 1}
	this.Same(card, (&Game{morphLast: card}).GetMorphLast())
}

// TestGameGetMorphCount 驗證 GetMorphCount 取回回合變身次數。
func (this *SuiteGame) TestGameGetMorphCount() {
	this.Equal(int32(2), (&Game{morphCount: 2}).GetMorphCount())
}

// TestGameGetMorphOldID 驗證 GetMorphOldID 取回變身前卡牌編號。
func (this *SuiteGame) TestGameGetMorphOldID() {
	this.Equal(int32(100), (&Game{morphOldID: 100}).GetMorphOldID())
}

// TestGameGetMorphNewID 驗證 GetMorphNewID 取回變身後卡牌編號。
func (this *SuiteGame) TestGameGetMorphNewID() {
	this.Equal(int32(200), (&Game{morphNewID: 200}).GetMorphNewID())
}

// TestGameEventDamage 驗證 EventDamage 成組設置受損值 + 來源顧客（含覆寫前次）。
func (this *SuiteGame) TestGameEventDamage() {
	guest := &Guest{instanceID: 9}
	game := &Game{}

	game.EventDamage(6, guest)
	this.Equal(int32(6), game.GetDamageValue())
	this.Same(guest, game.GetDamageGuest())

	game.EventDamage(3, nil) // 無顧客來源 → 覆寫為 nil
	this.Equal(int32(3), game.GetDamageValue())
	this.Nil(game.GetDamageGuest())
}

// TestGameEventSeat 驗證 EventSeat 成組設置最後入座 + 回合入座人數累加。
func (this *SuiteGame) TestGameEventSeat() {
	first := &Guest{instanceID: 1}
	second := &Guest{instanceID: 2}
	game := &Game{}

	game.EventSeat(first)
	game.EventSeat(second)
	this.Same(second, game.GetSeatLast())
	this.Equal(int32(2), game.GetSeatCount())
}

// TestGameEventExit 驗證 EventExit 成組設置最後離場 + 離場座位（取當下 seatID）+ 回合離場人數累加。
func (this *SuiteGame) TestGameEventExit() {
	seated := &Guest{instanceID: 1, seatID: 2}
	game := &Game{}

	game.EventExit(seated)
	this.Same(seated, game.GetExitLast())
	this.Equal(int32(2), game.GetExitLastSeat())
	this.Equal(int32(1), game.GetExitCount())

	roam := &Guest{instanceID: 3} // 自遊蕩列表離場（seatID 0）→ 離場座位 0
	game.EventExit(roam)
	this.Same(roam, game.GetExitLast())
	this.Equal(int32(0), game.GetExitLastSeat())
	this.Equal(int32(2), game.GetExitCount())
}

// TestGameEventTask 驗證 EventTask 成組設置最後行動顧客 + 行動技能 + 回合行動次數累加。
func (this *SuiteGame) TestGameEventTask() {
	guest := &Guest{instanceID: 9}
	game := &Game{}

	game.EventTask(guest, 77)
	this.Same(guest, game.GetTaskGuest())
	this.Equal(int32(77), game.GetTaskSkill())
	this.Equal(int32(1), game.GetTaskCount())
}

// TestGameEventDraw 驗證 EventDraw 成組設置最後抽出 + 回合張數累加 + 分組累積。
func (this *SuiteGame) TestGameEventDraw() {
	card := &Card{instanceID: 1}
	game := &Game{}

	game.EventDraw(card, 1)
	this.Same(card, game.GetDrawLast())
	this.Equal(int32(1), game.GetDrawCount())
	this.Equal(int32(1), game.GetDrawTotal().Get(1))
}

// TestGameEventDrop 驗證 EventDrop 成組設置最後棄置 + 回合張數累加 + 分組累積。
func (this *SuiteGame) TestGameEventDrop() {
	card := &Card{instanceID: 1}
	game := &Game{}

	game.EventDrop(card, 2)
	this.Same(card, game.GetDropLast())
	this.Equal(int32(1), game.GetDropCount())
	this.Equal(int32(1), game.GetDropTotal().Get(2))
}

// TestGameEventPlay 驗證 EventPlay 成組設置最後出牌 + 回合張數累加 + 分組累積。
func (this *SuiteGame) TestGameEventPlay() {
	card := &Card{instanceID: 1}
	game := &Game{}

	game.EventPlay(card, 1)
	this.Same(card, game.GetPlayLast())
	this.Equal(int32(1), game.GetPlayCount())
	this.Equal(int32(1), game.GetPlayTotal().Get(1))
}

// TestGameEventExile 驗證 EventExile 成組設置最後流放 + 回合張數累加 + 分組累積。
func (this *SuiteGame) TestGameEventExile() {
	card := &Card{instanceID: 1}
	game := &Game{}

	game.EventExile(card, 1)
	this.Same(card, game.GetExileLast())
	this.Equal(int32(1), game.GetExileCount())
	this.Equal(int32(1), game.GetExileTotal().Get(1))
}

// TestGameEventMorph 驗證 EventMorph 成組設置最後變身 + 前後卡牌編號 + 回合次數累加。
func (this *SuiteGame) TestGameEventMorph() {
	card := &Card{instanceID: 1}
	game := &Game{}

	game.EventMorph(card, 100, 200)
	this.Same(card, game.GetMorphLast())
	this.Equal(int32(100), game.GetMorphOldID())
	this.Equal(int32(200), game.GetMorphNewID())
	this.Equal(int32(1), game.GetMorphCount())
}

// TestGameRoundReset 驗證 RoundReset 歸零全部回合計數;Last 引用與整場累積保留。
func (this *SuiteGame) TestGameRoundReset() {
	guest := &Guest{instanceID: 9, seatID: 1}
	card := &Card{instanceID: 1}
	game := &Game{}
	game.EventSeat(guest)
	game.EventExit(guest)
	game.EventTask(guest, 77)
	game.EventDraw(card, 1)
	game.EventDrop(card, 1)
	game.EventPlay(card, 1)
	game.EventExile(card, 1)
	game.EventMorph(card, 100, 200)

	game.RoundReset()
	this.Equal(int32(0), game.GetSeatCount())
	this.Equal(int32(0), game.GetExitCount())
	this.Equal(int32(0), game.GetTaskCount())
	this.Equal(int32(0), game.GetDrawCount())
	this.Equal(int32(0), game.GetDropCount())
	this.Equal(int32(0), game.GetPlayCount())
	this.Equal(int32(0), game.GetExileCount())
	this.Equal(int32(0), game.GetMorphCount())

	this.Same(card, game.GetDrawLast())              // Last 引用保留
	this.Equal(int32(1), game.GetDrawTotal().Get(1)) // 整場累積保留
}
