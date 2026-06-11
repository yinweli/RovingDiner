package rules

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/yinweli/RovingDiner/internal/cores"
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
