package cores

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/yinweli/RovingDiner/internal/exprs"
	sheeter "github.com/yinweli/RovingDiner/sheet"
)

func TestSuiteGame(t *testing.T) {
	suite.Run(t, new(SuiteGame))
}

// SuiteGame 驗證營業實例(game.go): 建構 / 取值 / 階段設定 / 事件成組寫入 / 回合計數歸零,
// 與驅動引擎面(exprs.Resolver 委派、Lock 後綴路由、命令執行分派); 詞條內容逐項驗證見 attrRead_test.go 等。
type SuiteGame struct {
	suite.Suite
}

// TestNewGame 驗證 NewGame 建構營業實例: 盤面空白(屬性零值、無跳轉階段、累積計數零值可用、座位 map 就緒、容器空、身分入欄),
// 注入靜態表與衍生索引、self 不入建構。
func (this *SuiteGame) TestNewGame() {
	sheet := buildSheet()
	game := NewGame(123, 5, NewData(sheet, nil), nil, nil, nil)
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
	this.Equal(int64(123), game.Seed) // 本場身分入欄
	this.Equal(int32(5), game.StageID)

	this.Same(sheet, game.GetSheet()) // 注入遊戲資料
	this.Nil(game.GetSelf())          // self 為 run-state, 建構不綁定

	empty := NewGame(0, 0, nil, nil, nil, nil) // data nil → 補空殼, 查詢不爆
	this.Nil(empty.GetSheet())
}

// TestGameNextID 驗證 NextID 配發遞增唯一實例編號(卡牌 / 顧客 / 效果共用同一序列)。
func (this *SuiteGame) TestGameNextID() {
	game := NewGame(0, 0, nil, nil, nil, nil)
	this.Equal(InstanceID(1), game.NextID())
	this.Equal(InstanceID(2), game.NextID())
	this.Equal(InstanceID(3), game.NextID())
}

// TestGameSetSelf 驗證 SetSelf 成對寫門: 綁定 + 還原函式逐層退棧(結算重入安全)。
func (this *SuiteGame) TestGameSetSelf() {
	game := NewGame(0, 0, nil, nil, nil, nil)
	outer := NewRefGuest(&Guest{instanceID: 1})
	inner := NewRefCard(&Card{instanceID: 2})

	restoreOuter := game.SetSelf(&outer)
	this.Same(&outer, game.GetSelf())

	restoreInner := game.SetSelf(&inner) // 巢狀綁定
	this.Same(&inner, game.GetSelf())

	restoreInner() // 逐層還原
	this.Same(&outer, game.GetSelf())

	restoreOuter()
	this.Nil(game.GetSelf())
}

// TestGameGetOperator 驗證 GetOperator 取回注入的玩家輸入 port。
func (this *SuiteGame) TestGameGetOperator() {
	this.Equal(fakeOperator{}, NewGame(0, 0, nil, fakeOperator{}, nil, nil).GetOperator())
}

// TestGameGetRander 驗證 GetRander 取回注入的亂數 port。
func (this *SuiteGame) TestGameGetRander() {
	this.Equal(fakeRander{}, NewGame(0, 0, nil, nil, fakeRander{}, nil).GetRander())
}

// TestGameEmit 驗證 Emit 發射投影事件: 統一蓋章座標(當前回合 / 階段)後轉交 presenter, 發射點不自帶;
// presenter nil 由建構正規化為無輸出替身, 發射不爆。
func (this *SuiteGame) TestGameEmit() {
	record := &fakePresenter{}
	game := NewGame(0, 0, nil, nil, nil, record)
	game.GetRound().Set(3)
	game.SetPhase(PhaseRoundStart)
	game.Emit(EventData{Kind: EventScope, Scope: ScopeSettle})
	this.Require().Len(record.event, 1)
	this.Equal(EventScope, record.event[0].Kind)
	this.Equal(ScopeSettle, record.event[0].Scope)
	this.Equal(int32(3), record.event[0].Round) // 座標由 Emit 蓋章, 發射點未填
	this.Equal(PhaseRoundStart, record.event[0].Phase)

	NewGame(0, 0, nil, nil, nil, nil).Emit(EventData{}) // presenter nil → 無輸出替身, 不爆
}

// TestGameGetMorale 驗證 GetMorale 取回餐廳士氣值組件(寫入經組件可見)。
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

// TestGameGetEnergyKeep 驗證 GetEnergyKeep 取回出牌點數保留(純鎖屬性, 狀態於鎖定計數)。
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

// TestGameGetPhase 驗證 GetPhase / SetPhase 當前階段存取(未踏站前為 PhaseNone)。
func (this *SuiteGame) TestGameGetPhase() {
	game := NewGame(0, 0, nil, nil, nil, nil)
	this.Equal(PhaseNone, game.GetPhase())

	game.SetPhase(PhaseGuestAction)
	this.Equal(PhaseGuestAction, game.GetPhase())
}

// TestGameGetNextPhase 驗證 GetNextPhase 取回下一階段(無跳轉目標為 PhaseNone)。
func (this *SuiteGame) TestGameGetNextPhase() {
	this.Equal(PhaseNone, (&Game{}).GetNextPhase())
	this.Equal(PhasePlayerAction, (&Game{phaseNext: PhasePlayerAction}).GetNextPhase())
}

// TestGameSetNextPhase 驗證 SetNextPhase 設下一階段(含清回 PhaseNone)。
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

// TestGameGetDamageGuest 驗證 GetDamageGuest 取回士氣受損顧客(無來源回 nil)。
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

// TestGameGetDrawTotal 驗證 GetDrawTotal 取回累積抽牌計數組件(零值可用、寫入經組件可見)。
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

// TestGameEventDamage 驗證 EventDamage 成組設置受損值 + 來源顧客(含覆寫前次)。
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

// TestGameEventExit 驗證 EventExit 成組設置最後離場 + 離場座位(取當下 seatID)+ 回合離場人數累加。
func (this *SuiteGame) TestGameEventExit() {
	seated := &Guest{instanceID: 1, seatID: 2}
	game := &Game{}

	game.EventExit(seated)
	this.Same(seated, game.GetExitLast())
	this.Equal(int32(2), game.GetExitLastSeat())
	this.Equal(int32(1), game.GetExitCount())

	roam := &Guest{instanceID: 3} // 自遊蕩列表離場(seatID 0)→ 離場座位 0
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

// TestGameRoundReset 驗證 RoundReset 歸零全部回合計數; Last 引用與整場累積保留。
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

// TestGameSkillEffect 驗證 SkillEffect 取技能效果列表複本(不共享靜態表底層); 技能不存在回 nil。
func (this *SuiteGame) TestGameSkillEffect() {
	game := NewGame(0, 0, NewData(buildSheet(), nil), nil, nil, nil)

	effect := game.SkillEffect(301)
	this.Equal([]int32{401, 402}, effect)

	effect[0] = 999 // 改複本不影響靜態表
	this.Equal([]int32{401, 402}, game.SkillEffect(301))

	this.Nil(game.SkillEffect(999)) // 技能不存在 → nil
}

// TestGameCardSkillGroup 驗證 CardSkillGroup 取卡牌技能群組; 卡牌 / 技能資料缺失回 0。
func (this *SuiteGame) TestGameCardSkillGroup() {
	game := NewGame(0, 0, NewData(buildSheet(), nil), nil, nil, nil)
	this.Equal(int32(3), game.CardSkillGroup(103)) // 卡 103 → 技能 301 → 群組 3
	this.Equal(int32(0), game.CardSkillGroup(101)) // 卡 101 無技能(SkillID 0)→ 0
	this.Equal(int32(0), game.CardSkillGroup(999)) // 卡牌資料不存在 → 0
}

// TestGameEffectData 驗證 EffectData 委派遊戲資料查預編譯效果; 查無回 ok=false。
func (this *SuiteGame) TestGameEffectData() {
	game := NewGame(0, 0, NewData(buildSheet(), nil), nil, nil, nil)

	meta, ok := game.EffectData(401)
	this.True(ok)
	this.Equal(int32(5), meta.Group)

	_, ok = game.EffectData(999) // 查無 → 失敗
	this.False(ok)
}

// TestGameGuestData 驗證 GuestData 委派遊戲資料查顧客門檻; 查無回 ok=false(即無門檻)。
func (this *SuiteGame) TestGameGuestData() {
	sheet := &sheeter.Sheeter{}
	sheet.Guest.Data = map[int32]*sheeter.Guest{
		501: {ID: 501, SateSkillID: []string{"6^301"}},
	}
	game := NewGame(0, 0, NewData(sheet, nil), nil, nil, nil)

	meta, ok := game.GuestData(501)
	this.True(ok)
	this.Len(meta.Sate, 1)

	_, ok = game.GuestData(999) // 查無 → 失敗
	this.False(ok)
}

// TestGameRollCard 驗證 RollCard 對抽獎群組 weighted random 抽卡牌編號; 群組不存在 / 總權重 0 回 ok=false。
func (this *SuiteGame) TestGameRollCard() {
	game := NewGame(0, 0, NewData(buildSheet(), nil), nil, fakeRander{}, nil)

	cardID, ok := game.RollCard(7) // 群組 7:Weighted 恆取首位 → 101
	this.True(ok)
	this.Equal(int32(101), cardID)

	_, ok = game.RollCard(8) // 群組 8 全 0 權重 → 失敗
	this.False(ok)

	_, ok = game.RollCard(99) // 群組不存在 → 失敗
	this.False(ok)
}

// TestGameAttr 驗證 Attr 對全域屬性詞彙表的單查分派(裝備經 RegisterAttrRead; 只驗分派機制, 真詞條驗證歸 rules)。
func (this *SuiteGame) TestGameAttr() {
	game := NewGame(0, 0, nil, nil, nil, nil)
	game.RegisterAttrRead("fake", func(game *Game, arg []exprs.Value) (result exprs.Value, ok bool) {
		return exprs.NewNum(7), true
	})

	value, ok := game.Attr("fake", nil) // 詞條命中
	this.True(ok)
	this.Equal(float64(7), value.Num())

	_, ok = game.Attr("nope", nil) // 未知名稱 → 失敗
	this.False(ok)

	_, ok = NewGame(0, 0, nil, nil, nil, nil).Attr("fake", nil) // 未裝備 → 失敗
	this.False(ok)
}

// TestGameAttrRef 驗證 AttrRef 對引用屬性詞彙表的單查分派(裝備經 RegisterAttrRefRead)。
func (this *SuiteGame) TestGameAttrRef() {
	game := NewGame(0, 0, nil, nil, nil, nil)
	game.RegisterAttrRefRead("fake", func(game *Game, ref exprs.Ref, arg []exprs.Value) (result exprs.Value, ok bool) {
		guest, valid := AsGuest(ref)

		if valid == false {
			return exprs.Value{}, false
		} // if

		return exprs.NewNum(float64(guest.GetCalm().GetValue())), true
	})

	value, ok := game.AttrRef(NewRefGuest(&Guest{calm: NewValue(5, 0)}), "fake", nil) // 詞條命中
	this.True(ok)
	this.Equal(float64(5), value.Num())

	_, ok = game.AttrRef(NewRefGuest(&Guest{}), "nope", nil) // 未知名稱 → 失敗
	this.False(ok)
}

// TestGameExecAssignGlobal 驗證 ExecAssign 全域左值路徑: 帶值賦值求值(含 builtin 注入)/ 鎖定 / 未知名稱與評估失敗 no-op;
// 屬性事件收口——進了寫入詞條就發(前後值經讀詞條、@ 載鎖定計數)、閘門前夭折無事件(M18 拍板)。
func (this *SuiteGame) TestGameExecAssignGlobal() {
	record := &fakePresenter{}
	game := NewGame(0, 0, nil, nil, nil, record)
	game.score = NewValue(10, 0)
	game.RegisterAttrWrite("score", func(game *Game, op AssignKind, n float64) bool {
		return game.GetScore().Apply(op, n)
	})
	game.RegisterAttrRead("score", func(game *Game, arg []exprs.Value) (result exprs.Value, ok bool) {
		return exprs.NewNum(float64(game.GetScore().GetValue())), true
	})
	game.RegisterAttrRead("scoreLock", func(game *Game, arg []exprs.Value) (result exprs.Value, ok bool) {
		return exprs.NewNum(float64(game.GetScore().GetLock())), true
	})
	game.RegisterBuiltin("seven", func(arg []exprs.Value) (result exprs.Value, ok bool) {
		return exprs.NewNum(7), true
	})

	// 全域帶值賦值; RHS 經 exprs 求值(內建函式 seven 驗證 builtin 裝備經 Env 帶入)
	this.True(game.ExecAssign("score", "", false, AssignAdd, this.expr("seven()")))
	this.Equal(int32(17), game.GetScore().GetValue())
	this.Require().Len(record.event, 1) // 屬性事件: 詞條鍵 + 賦值符 + 右值 + 前後值, 全域對象欄零值
	this.Equal(EventData{Kind: EventProperty, Attr: "score", Op: AssignAdd, Operand: 7, Before: 10, After: 17}, record.event[0])

	// @ 鎖定(不帶右值, value 為 nil、不求值); 事件前後值載鎖定計數(Lock 全名讀鍵)
	this.True(game.ExecAssign("score", "", false, AssignLock, nil))
	this.Equal(int32(1), game.GetScore().GetLock())
	this.Require().Len(record.event, 2)
	this.Equal(EventData{Kind: EventProperty, Attr: "score", Op: AssignLock, Before: 0, After: 1}, record.event[1])

	// 鎖定拒寫仍發事件, 以 Before == After 表達(M18 拍板)
	this.False(game.ExecAssign("score", "", false, AssignSet, this.expr("99")))
	this.Require().Len(record.event, 3)
	this.Equal(EventData{Kind: EventProperty, Attr: "score", Op: AssignSet, Operand: 99, Before: 17, After: 17}, record.event[2])

	// 無讀詞條的可寫屬性 → 事件前後值留零值(進了寫入詞條照發)
	game.RegisterAttrWrite("silent", func(game *Game, op AssignKind, n float64) bool { return true })
	this.True(game.ExecAssign("silent", "", false, AssignSet, this.expr("1")))
	this.Require().Len(record.event, 4)
	this.Equal(EventData{Kind: EventProperty, Attr: "silent", Op: AssignSet, Operand: 1}, record.event[3])

	// 未知全域屬性 → no-op、無事件
	this.False(game.ExecAssign("nope", "", false, AssignSet, this.expr("1")))

	// RHS 評估失敗(除 0)→ no-op、無事件
	this.False(game.ExecAssign("score", "", false, AssignSet, this.expr("1 / 0")))

	// RHS 非數值(字串)→ no-op、無事件
	this.False(game.ExecAssign("score", "", false, AssignSet, this.expr("'text'")))
	this.Len(record.event, 4) // 閘門前夭折三式皆無事件
}

// TestGameExecAssignRef 驗證 ExecAssign 引用左值路徑: 引用解析 / 屬性凍結 / 未知屬性 / 空物件 no-op;
// 屬性事件帶對象編號(資料編號 + 實例編號), @ 無 Lock 讀詞條時前後值留零值。
func (this *SuiteGame) TestGameExecAssignRef() {
	record := &fakePresenter{}
	card := &Card{instanceID: 1, cardID: 103}
	game := NewGame(0, 0, nil, nil, nil, record)
	game.drawLast = card
	game.RegisterAttrRead("drawLast", func(game *Game, arg []exprs.Value) (result exprs.Value, ok bool) {
		return NewRefCard(game.GetDrawLast()).Value(), true
	})
	game.RegisterAttrRefRead("cost", func(game *Game, ref exprs.Ref, arg []exprs.Value) (result exprs.Value, ok bool) {
		target, valid := AsCard(ref)

		if valid == false {
			return exprs.Value{}, false
		} // if

		return exprs.NewNum(float64(target.GetCost().GetValue())), true
	})
	game.RegisterAttrRefWrite("cost", func(game *Game, ref exprs.Ref, op AssignKind, n float64) bool {
		target, valid := AsCard(ref)

		if valid == false {
			return false
		} // if

		return target.GetCost().Apply(op, n)
	})

	// 引用左值寫入(最後抽出卡牌的出牌費用設為 3); 事件帶對象編號與前後值
	this.True(game.ExecAssign("drawLast", "cost", true, AssignSet, this.expr("3")))
	this.Equal(int32(3), card.GetCost().GetValue())
	this.Require().Len(record.event, 1)
	this.Equal(EventData{Kind: EventProperty, DataID: 103, InstanceID: 1, Attr: "cost", Op: AssignSet, Operand: 3, Before: 0, After: 3}, record.event[0])

	// @ 鎖定: 無 costLock 讀詞條 → 事件前後值留零值(進了寫入詞條照發)
	this.True(game.ExecAssign("drawLast", "cost", true, AssignLock, nil))
	this.Require().Len(record.event, 2)
	this.Equal(EventData{Kind: EventProperty, DataID: 103, InstanceID: 1, Attr: "cost", Op: AssignLock, Before: 0, After: 0}, record.event[1])

	// 未知引用屬性 → no-op、無事件
	this.False(game.ExecAssign("drawLast", "nope", true, AssignSet, this.expr("3")))
	this.Len(record.event, 2)

	// 引用解析為空物件(drawLast 為 nil → 引用 Value 為 none)→ no-op
	game.drawLast = nil
	this.False(game.ExecAssign("drawLast", "cost", true, AssignSet, this.expr("3")))

	// 屬性凍結: 寫入主體為凍結中(卡牌化列表)顧客 → no-op; 移出卡牌化列表(解凍)後寫入恢復
	guest := &Guest{instanceID: 2}
	game.Cardify = GuestList{guest}
	game.RegisterAttrRead("exitLast", func(game *Game, arg []exprs.Value) (result exprs.Value, ok bool) {
		return NewRefGuest(guest).Value(), true
	})
	game.RegisterAttrRefWrite("calm", func(game *Game, ref exprs.Ref, op AssignKind, n float64) bool {
		target, valid := AsGuest(ref)

		if valid == false {
			return false
		} // if

		return target.GetCalm().Apply(op, n)
	})

	this.False(game.ExecAssign("exitLast", "calm", true, AssignSet, this.expr("3"))) // 凍結中 → no-op
	this.Equal(int32(0), guest.GetCalm().GetValue())

	game.Cardify = GuestList{}
	this.True(game.ExecAssign("exitLast", "calm", true, AssignSet, this.expr("3"))) // 解凍 → 寫入
	this.Equal(int32(3), guest.GetCalm().GetValue())
}

// TestGameExecOperate 驗證 ExecOperate 分派: 求值命令對象參數 → 解析作用集合 → 求值參數 → 查表 fan-out;
// 任一參數評估失敗 / 名稱未登錄 → 整動作 no-op(evalAll 順帶覆蓋)。
func (this *SuiteGame) TestGameExecOperate() {
	run := []InstanceID{}
	game := NewGame(0, 0, nil, nil, nil, nil)
	game.RegisterSelector("fake", func(game *Game, arg []exprs.Value) (result []InstanceID) {
		return []InstanceID{InstanceID(arg[0].Num())}
	})
	game.RegisterCommand("record", func(game *Game, target []InstanceID, arg []exprs.Value) {
		run = append(run, target...)
	})

	game.ExecOperate("record", "fake", []*exprs.Expr{this.expr("7")}, nil) // 正常分派
	this.Equal([]InstanceID{7}, run)

	game.ExecOperate("record", "fake", []*exprs.Expr{this.expr("1 / 0")}, nil)                           // 命令對象參數評估失敗 → no-op
	game.ExecOperate("record", "nope", nil, nil)                                                         // 命令對象未登錄 → no-op
	game.ExecOperate("nope", "fake", []*exprs.Expr{this.expr("1")}, nil)                                 // 未知命令 → no-op
	game.ExecOperate("record", "fake", []*exprs.Expr{this.expr("1")}, []*exprs.Expr{this.expr("1 / 0")}) // 參數評估失敗 → no-op
	this.Equal([]InstanceID{7}, run)

	// none / 篩空正規化: none → target nil(全域掃描記號)、其他命令對象篩空 → 非 nil 空切片
	got := []InstanceID{}
	game.RegisterSelector(SelectorNone, func(game *Game, arg []exprs.Value) (result []InstanceID) {
		return nil
	})
	game.RegisterSelector("empty", func(game *Game, arg []exprs.Value) (result []InstanceID) {
		return nil
	})
	game.RegisterCommand("probe", func(game *Game, target []InstanceID, arg []exprs.Value) {
		got = target
	})

	game.ExecOperate("probe", SelectorNone, nil, nil) // none → nil
	this.Nil(got)

	game.ExecOperate("probe", "empty", nil, nil) // 篩空 → 非 nil 空切片
	this.NotNil(got)
	this.Empty(got)
}

// TestGameSelectObject 驗證 selectObject 對命令對象詞彙表的分派與未登錄回報。
func (this *SuiteGame) TestGameSelectObject() {
	game := NewGame(0, 0, nil, nil, nil, nil)
	game.RegisterSelector("fake", func(game *Game, arg []exprs.Value) (result []InstanceID) {
		return []InstanceID{7}
	})

	result, ok := game.selectObject("fake", nil) // 已登錄 → 派發至詞條
	this.True(ok)
	this.Equal([]InstanceID{7}, result)

	_, ok = game.selectObject("nope", nil) // 未登錄命令對象 → ok=false
	this.False(ok)
}

// TestGameLocateCard 驗證 LocateCard 自四牌堆定位卡牌與所在容器; 未命中回 ContainerNone。
func (this *SuiteGame) TestGameLocateCard() {
	game := NewGame(0, 0, nil, nil, nil, nil)
	game.Hand = CardList{{instanceID: 1}}
	game.Deck = CardList{{instanceID: 2}}
	game.Drop = CardList{{instanceID: 3}}
	game.Exile = CardList{{instanceID: 4}}

	for id, want := range map[InstanceID]ContainerKind{
		1: ContainerHand, 2: ContainerDeck, 3: ContainerDrop, 4: ContainerExile,
	} {
		card, where, ok := game.LocateCard(id)
		this.True(ok)
		this.Equal(want, where)
		this.Equal(id, card.GetInstanceID())
	} // for

	card, where, ok := game.LocateCard(99) // 未命中
	this.Nil(card)
	this.Equal(ContainerNone, where)
	this.False(ok)
}

// TestGameLocateGuest 驗證 LocateGuest 自顧客四容器定位顧客與所在容器; 未命中回 ContainerNone。
func (this *SuiteGame) TestGameLocateGuest() {
	game := NewGame(0, 0, nil, nil, nil, nil)
	game.Seat.Place(1, &Guest{instanceID: 1})
	game.Wait = WaitList{{instanceID: 2}}
	game.Roam = GuestList{{instanceID: 3}}
	game.Cardify = GuestList{{instanceID: 4}}

	for id, want := range map[InstanceID]ContainerKind{
		1: ContainerSeat, 2: ContainerWait, 3: ContainerRoam, 4: ContainerCardify,
	} {
		guest, where, ok := game.LocateGuest(id)
		this.True(ok)
		this.Equal(want, where)
		this.Equal(id, guest.GetInstanceID())
	} // for

	guest, where, ok := game.LocateGuest(99) // 未命中
	this.Nil(guest)
	this.Equal(ContainerNone, where)
	this.False(ok)
}

// TestGameIsFrozen 驗證 IsFrozen 以卡牌化列表成員身分判凍結(freeze 欄位不參與判定)。
func (this *SuiteGame) TestGameIsFrozen() {
	frozen := &Guest{instanceID: 1}
	free := &Guest{instanceID: 2}
	free.SetFreeze(3) // freeze 殘值不影響判定(僅作解凍補回錨點)
	game := NewGame(0, 0, nil, nil, nil, nil)
	game.Cardify = GuestList{frozen}

	this.True(game.IsFrozen(frozen))
	this.False(game.IsFrozen(free))
}

// TestGameEnv 驗證 Env 組裝求值期環境: 自身為 Resolver、帶入裝備的內建函式。
func (this *SuiteGame) TestGameEnv() {
	game := NewGame(0, 0, nil, nil, nil, nil)
	game.RegisterBuiltin("seven", func(arg []exprs.Value) (result exprs.Value, ok bool) {
		return exprs.NewNum(7), true
	})

	env := game.Env()
	this.Equal(exprs.Resolver(game), env.Resolver)
	this.Contains(env.Builtin, "seven")
}

// === 測試輔助(置尾) ===

// expr 解析算術式來源為 *exprs.Expr; 解析失敗即測試失敗。
func (this *SuiteGame) expr(source string) *exprs.Expr {
	result, err := exprs.Parse(source)
	this.Require().NoError(err)
	return result
}

// fakeOperator 玩家輸入空替身(cores 白箱測試私有; rules / games 用 tester.FakeOperator)。
type fakeOperator struct{}

func (this fakeOperator) PlayerAction(game *Game) *Card {
	return nil
}

func (this fakeOperator) PickGuest(source []*Guest, count int) []*Guest {
	return source[:count]
}

func (this fakeOperator) PickCard(source []*Card, count int) []*Card {
	return source[:count]
}

func (this fakeOperator) PickDiscard(source []*Card, over int) []*Card {
	return nil
}

// fakeRander 決定性亂數空替身(cores 白箱測試私有; rules / games 用 tester.FakeRander)。
type fakeRander struct{}

func (this fakeRander) Intn(n int) int {
	return 0
}

func (this fakeRander) Shuffle(n int, swap func(i, j int)) {
}

func (this fakeRander) Weighted(weight []int32) int {
	return 0
}

// fakePresenter 事件流錄製替身(cores 白箱測試私有; rules / games 用 tester.RecordPresenter)。
type fakePresenter struct {
	event []EventData
}

func (this *fakePresenter) Emit(eventData EventData) {
	this.event = append(this.event, eventData)
}
