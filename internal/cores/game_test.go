package cores

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/yinweli/RovingDiner/internal/exprs"
)

func TestSuiteGame(t *testing.T) {
	suite.Run(t, new(SuiteGame))
}

// SuiteGame 驗證營業實例(game.go):建構 / 取值 / 階段設定 / 事件成組寫入 / 回合計數歸零,
// 與驅動引擎面(exprs.Resolver 委派、Lock 後綴路由、命令執行分派);詞條內容逐項驗證見 attrRead_test.go 等。
type SuiteGame struct {
	suite.Suite
}

// TestNewGame 驗證 NewGame 建構營業實例:盤面空白(屬性零值、無跳轉階段、累積計數零值可用、座位 map 就緒、容器空、種子入欄),
// 注入靜態表與衍生索引、self 不入建構。
func (this *SuiteGame) TestNewGame() {
	game := NewGame(123, buildSheet(), fakeOperator{}, fakeRander{}, nil)
	this.Require().NotNil(game)
	this.Equal(int32(0), game.GetMorale().GetValue())
	this.Equal(PhaseNone, game.GetNextPhase())
	this.Equal(int32(0), game.GetDrawTotal().Sum())
	this.Nil(game.GetDamageGuest())

	this.NotNil(game.Seat) // 座位 map 必須可用
	this.Empty(game.Hand)
	this.Empty(game.Deck)
	this.Empty(game.Effect)
	this.Empty(game.Action)
	this.Equal(int64(123), game.Seed)

	this.NotNil(game.data)       // 注入靜態表
	this.NotNil(game.awardData)  // 衍生索引於建構時整理
	this.NotNil(game.effectData) //
	this.Nil(game.self)          // self 為 run-state,建構不綁定
}

// TestGameNextID 驗證 NextID 配發遞增唯一實例編號（卡牌 / 顧客 / 效果共用同一序列）。
func (this *SuiteGame) TestGameNextID() {
	game := NewGame(0, nil, nil, nil, nil)
	this.Equal(InstanceID(1), game.NextID())
	this.Equal(InstanceID(2), game.NextID())
	this.Equal(InstanceID(3), game.NextID())
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

// TestGameSkillEffect 驗證 SkillEffect 取技能效果列表複本(不共享靜態表底層);技能不存在回 nil。
func (this *SuiteGame) TestGameSkillEffect() {
	game := NewGame(0, buildSheet(), nil, nil, nil)

	effect := game.SkillEffect(301)
	this.Equal([]int32{401, 402}, effect)

	effect[0] = 999 // 改複本不影響靜態表
	this.Equal([]int32{401, 402}, game.SkillEffect(301))

	this.Nil(game.SkillEffect(999)) // 技能不存在 → nil
}

// TestGameCardSkillGroup 驗證 CardSkillGroup 取卡牌技能群組;卡牌 / 技能資料缺失回 0。
func (this *SuiteGame) TestGameCardSkillGroup() {
	game := NewGame(0, buildSheet(), nil, nil, nil)
	this.Equal(int32(3), game.CardSkillGroup(103)) // 卡 103 → 技能 301 → 群組 3
	this.Equal(int32(0), game.CardSkillGroup(101)) // 卡 101 無技能（SkillID 0）→ 0
	this.Equal(int32(0), game.CardSkillGroup(999)) // 卡牌資料不存在 → 0
}

// TestGameAttr 驗證 Attr 對全域屬性詞彙表的委派與 Lock 後綴路由(只驗分派機制)。
func (this *SuiteGame) TestGameAttr() {
	game := NewGame(0, nil, nil, nil, nil)
	game.morale = NewValue(30, 2)

	// 值表命中
	value, ok := game.Attr("morale", nil)
	this.True(ok)
	this.Equal(float64(30), value.Num())

	// Lock 後綴 → 剝後綴查鎖表
	value, ok = game.Attr("moraleLock", nil)
	this.True(ok)
	this.Equal(float64(2), value.Num())

	// 未知名稱 → 失敗
	_, ok = game.Attr("nope", nil)
	this.False(ok)

	// 非鎖屬性的 Lock 後綴 → 失敗(round 為「寫」無鎖,鎖表無此鍵)
	_, ok = game.Attr("roundLock", nil)
	this.False(ok)
}

// TestGameAttrRef 驗證 AttrRef 對引用屬性詞彙表的委派與 Lock 後綴路由。
func (this *SuiteGame) TestGameAttrRef() {
	game := NewGame(0, nil, nil, nil, nil)
	ref := NewRefGuest(&Guest{calm: NewValue(5, 1)})

	// 值表命中
	value, ok := game.AttrRef(ref, "calm", nil)
	this.True(ok)
	this.Equal(float64(5), value.Num())

	// Lock 後綴 → 剝後綴查鎖表
	value, ok = game.AttrRef(ref, "calmLock", nil)
	this.True(ok)
	this.Equal(float64(1), value.Num())

	// 未知名稱 → 失敗
	_, ok = game.AttrRef(ref, "nope", nil)
	this.False(ok)
}

// TestGameExecAssignGlobal 驗證 ExecAssign 全域左值路徑:帶值賦值求值 / 鎖定 / 未知名稱與評估失敗 no-op。
func (this *SuiteGame) TestGameExecAssignGlobal() {
	game := NewGame(0, nil, nil, nil, nil)
	game.score = NewValue(10, 0)

	// 全域帶值賦值;RHS 經 exprs 求值(含內建函式 min,驗證 builtin 注入)
	this.True(game.ExecAssign("score", "", false, AssignAdd, this.expr("min(5, 8)")))
	this.Equal(int32(15), game.GetScore().GetValue())

	// @ 鎖定(不帶右值,value 為 nil、不求值)
	this.True(game.ExecAssign("score", "", false, AssignLock, nil))
	this.Equal(int32(1), game.GetScore().GetLock())

	// 未知全域屬性 → no-op
	this.False(game.ExecAssign("nope", "", false, AssignSet, this.expr("1")))

	// RHS 評估失敗(除 0)→ no-op
	this.False(game.ExecAssign("energy", "", false, AssignSet, this.expr("1 / 0")))

	// RHS 非數值(字串)→ no-op
	this.False(game.ExecAssign("energy", "", false, AssignSet, this.expr("'text'")))
}

// TestGameExecAssignRef 驗證 ExecAssign 引用左值路徑:引用解析 / 型別不符 / 未知屬性 / 空物件 no-op。
func (this *SuiteGame) TestGameExecAssignRef() {
	card := &Card{instanceID: 1}
	game := NewGame(0, nil, nil, nil, nil)
	game.drawLast = card

	// 引用左值寫入(最後抽出卡牌的出牌費用設為 3)
	this.True(game.ExecAssign("drawLast", "cost", true, AssignSet, this.expr("3")))
	this.Equal(int32(3), card.GetCost().GetValue())

	// 引用屬性型別不符(卡牌引用寫顧客屬性 calm)→ no-op
	this.False(game.ExecAssign("drawLast", "calm", true, AssignSet, this.expr("3")))

	// 未知引用屬性 → no-op
	this.False(game.ExecAssign("drawLast", "nope", true, AssignSet, this.expr("3")))

	// 引用解析為空物件(drawLast 為 nil)→ no-op
	game.drawLast = nil
	this.False(game.ExecAssign("drawLast", "cost", true, AssignSet, this.expr("3")))
}

// TestGameSelectObject 驗證 selectObject 對命令對象詞彙表的分派與未登錄回報。
func (this *SuiteGame) TestGameSelectObject() {
	game := NewGame(0, buildSheet(), fakeOperator{}, fakeRander{}, nil)
	game.drawLast = &Card{instanceID: 7}

	result, ok := game.selectObject("drawLast", nil) // 已登錄 → 派發至詞條
	this.True(ok)
	this.Equal([]InstanceID{7}, result)

	_, ok = game.selectObject("nope", nil) // 未登錄命令對象 → ok=false
	this.False(ok)
}

// === 測試輔助（置尾） ===

// expr 解析算術式來源為 *exprs.Expr;解析失敗即測試失敗。
func (this *SuiteGame) expr(source string) *exprs.Expr {
	result, err := exprs.Parse(source)
	this.Require().NoError(err)
	return result
}

// injectPort 注入測試替身:共用 buildSheet 靜態表、決定性 fake Operator / Rander,並重建衍生索引(等價 NewGame 的注入半部)。
// 供各 suite 對已佈置狀態的 Game 補注入;全空 Game 直接以 NewGame 帶參建構即可。
func injectPort(game *Game) {
	game.data = buildSheet()
	game.operator = fakeOperator{}
	game.rander = fakeRander{}
	game.awardData = prepareAward(game.data)
	game.effectData = prepareEffect(game.data, nil)
}
