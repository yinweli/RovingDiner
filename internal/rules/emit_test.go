package rules

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/yinweli/RovingDiner/internal/cores"
	"github.com/yinweli/RovingDiner/internal/tester"
)

func TestSuiteEmit(t *testing.T) {
	suite.Run(t, new(SuiteEmit))
}

// SuiteEmit 驗證投影事件發射輔助(emit.go): 各類事件的欄位組裝與選中清單建構; 發射點接線斷言見各流程測試。
type SuiteEmit struct {
	suite.Suite
}

// TestEmitEffect 驗證效果事件組裝: 對象欄載 self 本體(卡牌 / 顧客 / 空物件留零值)。
func (this *SuiteEmit) TestEmitEffect() {
	game, record := newGameRecord()
	card := cores.NewCard(game, 101)

	emitEffect(game, 401, 9, cores.NewRefCard(card), cores.EffectStageImmed)
	this.Require().Len(record.Event, 1)
	this.Equal(cores.EventData{Kind: cores.EventEffect, DataID: 101, InstanceID: card.GetInstanceID(), EffectID: 401, EffectInstanceID: 9, Stage: cores.EffectStageImmed}, record.Event[0])

	emitEffect(game, 401, 0, cores.Ref{}, cores.EffectStageEnd) // 空物件 self → 對象欄零值
	this.Require().Len(record.Event, 2)
	this.Equal(cores.EventData{Kind: cores.EventEffect, EffectID: 401, Stage: cores.EffectStageEnd}, record.Event[1])
}

// TestEmitGuestMove 驗證顧客容器搬移事件組裝: 對象編號 + From / To + 座位編號。
func (this *SuiteEmit) TestEmitGuestMove() {
	game, record := newGameRecord()
	guest := cores.NewGuest(game, 501)

	emitGuestMove(game, guest, cores.ContainerWait, cores.ContainerSeat, 2)
	this.Require().Len(record.Event, 1)
	this.Equal(cores.EventData{Kind: cores.EventContainer, DataID: 501, InstanceID: guest.GetInstanceID(), From: cores.ContainerWait, To: cores.ContainerSeat, SeatID: 2}, record.Event[0])
}

// TestEmitCardMove 驗證卡牌容器搬移事件組裝: 對象編號 + From / To; 已綁卡牌化來源者帶語境欄。
func (this *SuiteEmit) TestEmitCardMove() {
	game, record := newGameRecord()
	card := cores.NewCard(game, 101)

	emitCardMove(game, card, cores.ContainerDeck, cores.ContainerHand)
	this.Require().Len(record.Event, 1)
	this.Equal(cores.EventData{Kind: cores.EventContainer, DataID: 101, InstanceID: card.GetInstanceID(), From: cores.ContainerDeck, To: cores.ContainerHand}, record.Event[0])

	guest := cores.NewGuest(game, 501)
	card.CardifyBind(guest) // 綁卡牌化來源 → 語境欄
	emitCardMove(game, card, cores.ContainerNone, cores.ContainerHand)
	this.Require().Len(record.Event, 2)
	this.Equal(cores.EventData{Kind: cores.EventContainer, DataID: 101, InstanceID: card.GetInstanceID(), From: cores.ContainerNone, To: cores.ContainerHand, BindID: 501, BindInstanceID: guest.GetInstanceID()}, record.Event[1])
}

// TestEmitDeckOrder 驗證抽牌牌堆重整快照組裝: From == To == Deck、Pick 載全序。
func (this *SuiteEmit) TestEmitDeckOrder() {
	game, record := newGameRecord()
	game.Deck = cores.CardList{cores.NewCard(game, 101), cores.NewCard(game, 103)}

	emitDeckOrder(game)
	this.Require().Len(record.Event, 1)
	this.Equal(cores.EventData{Kind: cores.EventContainer, From: cores.ContainerDeck, To: cores.ContainerDeck, Pick: cardPickData(game.Deck)}, record.Event[0])
}

// TestEmitAction 驗證行動佇列事件組裝: 對象欄載顧客 + 技能 + 行動類型 + 在列記號。
func (this *SuiteEmit) TestEmitAction() {
	game, record := newGameRecord()
	guest := cores.NewGuest(game, 501)
	action := cores.NewAction(guest, cores.TaskCalm, 301)

	emitAction(game, action, true)
	this.Require().Len(record.Event, 1)
	this.Equal(cores.EventData{Kind: cores.EventAction, DataID: 501, InstanceID: guest.GetInstanceID(), SkillID: 301, Task: cores.TaskCalm, Alive: true}, record.Event[0])

	emitAction(game, action, false) // 出列
	this.Require().Len(record.Event, 2)
	this.False(record.Event[1].Alive)
}

// TestEmitEffectState 驗證載佇列狀態快照的效果事件組裝: Stack / Expire 絕對值 + Alive 去留記號。
func (this *SuiteEmit) TestEmitEffectState() {
	data := tester.BuildData()
	data.SetEffect(801, cores.EffectData{RunRound: 3, StackMax: 5})
	game, record := newGameDataRecord(data)
	game.GetRound().Set(2)
	card := cores.NewCard(game, 101)
	effect := cores.NewEffect(game, 801, cores.NewRefCard(card), 2)

	emitEffectState(game, effect, cores.EffectStageJoin, true)
	this.Require().Len(record.Event, 1)
	this.Equal(cores.EventData{Kind: cores.EventEffect, Round: 2, DataID: 101, InstanceID: card.GetInstanceID(), EffectID: 801, EffectInstanceID: effect.GetInstanceID(), Stage: cores.EffectStageJoin, Stack: 2, Expire: 4, Alive: true}, record.Event[0]) // 結束回合 = 2 + 3 - 1; Round 為 Emit 座標蓋章
}

// TestEmitProperty 驗證流程直寫屬性事件組裝: 詞條鍵 + 賦值符 + 右值 + 前後值。
func (this *SuiteEmit) TestEmitProperty() {
	game, record := newGameRecord()
	emitProperty(game, 501, 3, "calm", cores.AssignSub, 1, 5, 4)
	this.Require().Len(record.Event, 1)
	this.Equal(cores.EventData{Kind: cores.EventProperty, DataID: 501, InstanceID: 3, Attr: "calm", Op: cores.AssignSub, Operand: 1, Before: 5, After: 4}, record.Event[0])
}

// TestEmitSelect 驗證玩家選取紀錄組裝: 來源鍵 + 效果編號 + 選中清單。
func (this *SuiteEmit) TestEmitSelect() {
	game, record := newGameRecord()
	pick := []cores.PickData{{DataID: 101, InstanceID: 7}}

	emitSelect(game, "guestPick", 401, pick)
	this.Require().Len(record.Event, 1)
	this.Equal(cores.EventData{Kind: cores.EventSelect, Source: "guestPick", EffectID: 401, Pick: pick}, record.Event[0])
}

// TestCardPickData 驗證卡牌選中清單建構: 逐張(資料編號 + 實例編號)、每發新建切片(不共享來源底層)。
func (this *SuiteEmit) TestCardPickData() {
	game := newGame()
	c1 := cores.NewCard(game, 101)
	c2 := cores.NewCard(game, 103)

	this.Equal([]cores.PickData{
		{DataID: 101, InstanceID: c1.GetInstanceID()},
		{DataID: 103, InstanceID: c2.GetInstanceID()},
	}, cardPickData([]*cores.Card{c1, c2}))
	this.Empty(cardPickData(nil))
}

// TestGuestPickData 驗證顧客選中清單建構; 規則同 cardPickData。
func (this *SuiteEmit) TestGuestPickData() {
	game := newGame()
	guest := cores.NewGuest(game, 501)

	this.Equal([]cores.PickData{{DataID: 501, InstanceID: guest.GetInstanceID()}}, guestPickData([]*cores.Guest{guest}))
	this.Empty(guestPickData(nil))
}
