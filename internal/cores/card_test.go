package cores

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

func TestSuiteCard(t *testing.T) {
	suite.Run(t, new(SuiteCard))
}

// SuiteCard 驗證卡牌實例與卡牌容器(card.go): 建構 / 複製 / 取值 / 卡牌化綁定 / 變身, 牌堆置頂 / 移除 / 查找。
type SuiteCard struct {
	suite.Suite
}

// TestNewCard 驗證 NewCard 載入卡牌資料初始值、bool 欄轉鎖、SkillID 取技能效果列表; 資料不存在回 nil。
func (this *SuiteCard) TestNewCard() {
	game := NewGame(0, 0, NewData(buildSheet(), nil), nil, nil, nil)

	card := NewCard(game, 103) // 卡 103:Cost 2、Keep / Seal bool → 鎖、SkillID 301 → 效果列表
	this.Require().NotNil(card)
	this.Equal(int32(103), card.GetCardID())
	this.Equal(int32(2), card.GetCost().GetValue())
	this.Equal(int32(1), card.GetKeep().GetLock())           // bool 欄 → 鎖定計數
	this.Equal(int32(1), card.GetSeal().GetLock())           //
	this.Equal(int32(0), card.GetPlayExile().GetLock())      // sheet 未設 → 0
	this.Equal([]int32{401, 402}, card.GetEffectID().List()) // SkillID 301 → Skill.EffectID
	this.NotEqual(InstanceID(0), card.GetInstanceID())       // 配發實例編號

	this.Nil(NewCard(game, 999)) // 卡牌資料不存在 → nil
}

// TestCopyCard 驗證 CopyCard 淺複製載初始值、深複製複製 source 當前狀態與效果列表; 實例編號皆重生。
func (this *SuiteCard) TestCopyCard() {
	game := NewGame(0, 0, NewData(buildSheet(), nil), nil, nil, nil)
	source := &Card{instanceID: game.NextID(), cardID: 103, cost: NewValue(9, 0), effectID: NewIDList(777)}

	shallow := CopyCard(game, source, false) // 淺複製: 載卡牌資料初始值(非 source 當前狀態)
	this.Require().NotNil(shallow)
	this.Equal(int32(2), shallow.GetCost().GetValue())          // 初始費用 2(非 source 的 9)
	this.Equal([]int32{401, 402}, shallow.GetEffectID().List()) // 技能效果列表
	this.NotEqual(source.GetInstanceID(), shallow.GetInstanceID())

	deep := CopyCard(game, source, true) // 深複製: 複製 source 當前狀態 + 效果列表
	this.Require().NotNil(deep)
	this.Equal(int32(9), deep.GetCost().GetValue())             // source 當前費用
	this.Equal([]int32{777}, deep.GetEffectID().List())         // source 效果列表深複製
	this.NotEqual(source.GetInstanceID(), deep.GetInstanceID()) // 實例編號重生

	source.GetEffectID().Add(888) // 深複製不共享底層
	this.Equal([]int32{777}, deep.GetEffectID().List())
}

// TestCardGetInstanceID 驗證 GetInstanceID 取回實例編號。
func (this *SuiteCard) TestCardGetInstanceID() {
	this.Equal(InstanceID(7), (&Card{instanceID: 7}).GetInstanceID())
}

// TestCardGetCardID 驗證 GetCardID 取回卡牌編號。
func (this *SuiteCard) TestCardGetCardID() {
	this.Equal(int32(101), (&Card{cardID: 101}).GetCardID())
}

// TestCardGetCost 驗證 GetCost 取回出牌費用組件(寫入經組件可見)。
func (this *SuiteCard) TestCardGetCost() {
	card := &Card{cost: NewValue(2, 0)}
	this.Equal(int32(2), card.GetCost().GetValue())

	card.GetCost().Add(1) // 經組件寫入 → 同一實體
	this.Equal(int32(3), card.GetCost().GetValue())
}

// TestCardGetExtraRunMin 驗證 GetExtraRunMin 取回額外發動次數下限。
func (this *SuiteCard) TestCardGetExtraRunMin() {
	this.Equal(int32(1), (&Card{extraRunMin: NewValue(1, 0)}).GetExtraRunMin().GetValue())
}

// TestCardGetExtraRunMax 驗證 GetExtraRunMax 取回額外發動次數上限。
func (this *SuiteCard) TestCardGetExtraRunMax() {
	this.Equal(int32(5), (&Card{extraRunMax: NewValue(5, 0)}).GetExtraRunMax().GetValue())
}

// TestCardGetKeep 驗證 GetKeep 取回不棄卡牌。
func (this *SuiteCard) TestCardGetKeep() {
	this.Equal(int32(1), (&Card{keep: NewValueLock(true)}).GetKeep().GetLock())
}

// TestCardGetSeal 驗證 GetSeal 取回封印卡牌。
func (this *SuiteCard) TestCardGetSeal() {
	this.Equal(int32(1), (&Card{seal: NewValueLock(true)}).GetSeal().GetLock())
}

// TestCardGetPlayExile 驗證 GetPlayExile 取回出牌後流放。
func (this *SuiteCard) TestCardGetPlayExile() {
	this.Equal(int32(1), (&Card{playExile: NewValueLock(true)}).GetPlayExile().GetLock())
}

// TestCardGetUnplayExile 驗證 GetUnplayExile 取回未出牌流放。
func (this *SuiteCard) TestCardGetUnplayExile() {
	this.Equal(int32(1), (&Card{unplayExile: NewValueLock(true)}).GetUnplayExile().GetLock())
}

// TestCardGetEffectID 驗證 GetEffectID 取回實例效果列表(零值可用)。
func (this *SuiteCard) TestCardGetEffectID() {
	card := &Card{}
	card.GetEffectID().Add(801)
	this.Equal([]int32{801}, card.GetEffectID().List())
}

// TestCardGetCardify 驗證 GetCardify 取回卡牌化來源(未綁定回 nil)。
func (this *SuiteCard) TestCardGetCardify() {
	guest := &Guest{instanceID: 9}
	this.Same(guest, (&Card{cardify: guest}).GetCardify())
	this.Nil((&Card{}).GetCardify()) // 未綁定 → nil
}

// TestCardCardifyBind 驗證 CardifyBind 成對紀律: 綁來源 + 不棄鎖 + 1。
func (this *SuiteCard) TestCardCardifyBind() {
	guest := &Guest{instanceID: 9}
	card := &Card{}

	card.CardifyBind(guest)
	this.Same(guest, card.GetCardify())
	this.Equal(int32(1), card.GetKeep().GetLock())
}

// TestCardCardifyFree 驗證 CardifyFree 成對紀律: 解綁 + 不棄鎖 - 1(夾 ≥ 0)。
func (this *SuiteCard) TestCardCardifyFree() {
	card := &Card{}
	card.CardifyBind(&Guest{instanceID: 9})

	card.CardifyFree()
	this.Nil(card.GetCardify())
	this.Equal(int32(0), card.GetKeep().GetLock())

	card.CardifyFree() // 已歸零再解 → 夾 ≥ 0
	this.Equal(int32(0), card.GetKeep().GetLock())
}

// TestCardMorph 驗證 Morph 變身重設: 重配實例編號、換編號、僅載四欄; 查無資料回 false 不動。
func (this *SuiteCard) TestCardMorph() {
	game := NewGame(0, 0, NewData(buildSheet(), nil), nil, nil, nil)
	card := &Card{instanceID: game.NextID(), cardID: 102, cost: NewValue(9, 0), extraRunMin: NewValue(7, 0)}
	old := card.GetInstanceID()

	this.True(card.Morph(game, 103))                         // 卡 103:Cost 2、Keep / Seal 鎖、效果 401, 402
	this.NotEqual(old, card.GetInstanceID())                 // 重分配實例編號
	this.Equal(int32(103), card.GetCardID())                 // 換卡牌編號
	this.Equal(int32(2), card.GetCost().GetValue())          // 載新卡費用
	this.Equal(int32(1), card.GetKeep().GetLock())           // 載新卡不棄鎖
	this.Equal(int32(1), card.GetSeal().GetLock())           // 載新卡封印鎖
	this.Equal([]int32{401, 402}, card.GetEffectID().List()) // 載新卡效果列表
	this.Equal(int32(7), card.GetExtraRunMin().GetValue())   // 額外發動次數不在重載四欄 → 保留

	id := card.GetInstanceID()
	this.False(card.Morph(game, 999)) // 查無卡牌資料 → false 不動
	this.Equal(id, card.GetInstanceID())
	this.Equal(int32(103), card.GetCardID())
}

// TestCardListPush 驗證 Push 加入牌堆頂端(後入者居頂)。
func (this *SuiteCard) TestCardListPush() {
	first := &Card{instanceID: 1}
	second := &Card{instanceID: 2}
	list := CardList{}

	list.Push(first)
	list.Push(second)
	this.Require().Len(list, 2)
	this.Same(second, list[0]) // 新進入者置頂
	this.Same(first, list[1])
}

// TestCardListRemove 驗證 Remove 依實例編號移除、未命中不變。
func (this *SuiteCard) TestCardListRemove() {
	keep := &Card{instanceID: 1}
	list := CardList{keep, {instanceID: 2}}

	list.Remove(InstanceID(2)) // 命中 → 移除
	this.Require().Len(list, 1)
	this.Same(keep, list[0])

	list.Remove(InstanceID(9)) // 未命中 → 不變
	this.Len(list, 1)
}

// TestCardListFind 驗證 Find 依實例編號查找、未命中回 nil。
func (this *SuiteCard) TestCardListFind() {
	card := &Card{instanceID: 5}
	list := CardList{{instanceID: 1}, card}

	this.Same(card, list.Find(InstanceID(5)))
	this.Nil(list.Find(InstanceID(9))) // 未命中 → nil
}

// TestCardListHas 驗證 Has 依實例編號回報牌堆歸屬。
func (this *SuiteCard) TestCardListHas() {
	list := CardList{{instanceID: 1}, {instanceID: 5}}

	this.True(list.Has(InstanceID(5)))
	this.False(list.Has(InstanceID(9))) // 不在牌堆

	empty := CardList{}
	this.False(empty.Has(InstanceID(5))) // 空牌堆
}

// TestCardListCountGroup 驗證 CountGroup 分組計數: group == 0 全量、資料缺失卡牌不計入任何群組。
func (this *SuiteCard) TestCardListCountGroup() {
	data := buildSheet()
	list := CardList{{cardID: 101}, {cardID: 101}, {cardID: 102}, {cardID: 999}} // 999 無靜態資料

	this.Equal(int32(4), list.CountGroup(0, data)) // group==0 全量(不過濾)
	this.Equal(int32(2), list.CountGroup(1, data)) // 群組 1 = 卡 101 兩張
	this.Equal(int32(1), list.CountGroup(2, data)) // 群組 2 = 卡 102 一張
	this.Equal(int32(0), list.CountGroup(9, data)) // 無此群組(含資料缺失卡牌)
}

// === 測試輔助(置尾) ===
