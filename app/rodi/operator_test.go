package rodi

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/yinweli/RovingDiner/internal/cores"
)

func TestSuiteOperator(t *testing.T) {
	suite.Run(t, new(SuiteOperator))
}

// SuiteOperator 驗證鍵盤玩家輸入(operator.go): 暫停點包成輸入請求輪次、阻塞等答覆、答覆原樣回引擎。
// 各測試以 goroutine 模擬引擎端呼叫、主 goroutine 扮演 UI 端答覆。
type SuiteOperator struct {
	suite.Suite
}

// TestKeyboardOperatorPlayerAction 驗證玩家行動請求: 無候選欄與提示; 答覆空 = 結束(nil)、一張 = 出該卡。
func (this *SuiteOperator) TestKeyboardOperatorPlayerAction() {
	bus := make(chan turn)
	target := keyboardOperator{turn: bus}
	result := make(chan *cores.Card, 1)

	go func() { result <- target.PlayerAction(nil) }()
	next := <-bus
	this.Equal(turnRequest, next.role)
	this.Empty(next.req.prompt)
	this.Nil(next.req.guest)
	this.Nil(next.req.card)
	next.req.answer <- answer{}
	this.Nil(<-result)

	go func() { result <- target.PlayerAction(nil) }()
	next = <-bus
	card := &cores.Card{}
	next.req.answer <- answer{card: []*cores.Card{card}}
	this.Same(card, <-result)
}

// TestKeyboardOperatorPickGuest 驗證顧客選取請求: 提示 / 候選 / 上限原樣下傳, 答覆原樣回引擎。
func (this *SuiteOperator) TestKeyboardOperatorPickGuest() {
	bus := make(chan turn)
	target := keyboardOperator{turn: bus}
	guest := []*cores.Guest{{}, {}, {}}
	result := make(chan []*cores.Guest, 1)

	go func() { result <- target.PickGuest("開朗 要求選顧客", guest, 2) }()
	next := <-bus
	this.Equal(turnRequest, next.role)
	this.Equal("開朗 要求選顧客", next.req.prompt)
	this.Equal(guest, next.req.guest)
	this.Equal(2, next.req.count)
	next.req.answer <- answer{guest: guest[1:]}
	this.Equal(guest[1:], <-result)
}

// TestKeyboardOperatorPickCard 驗證卡牌選取請求; 規則同 PickGuest。
func (this *SuiteOperator) TestKeyboardOperatorPickCard() {
	bus := make(chan turn)
	target := keyboardOperator{turn: bus}
	card := []*cores.Card{{}, {}, {}}
	result := make(chan []*cores.Card, 1)

	go func() { result <- target.PickCard("手牌指定 要求選手牌", card, 2) }()
	next := <-bus
	this.Equal("手牌指定 要求選手牌", next.req.prompt)
	this.Equal(card, next.req.card)
	this.Equal(2, next.req.count)
	next.req.answer <- answer{card: card[:2]}
	this.Equal(card[:2], <-result)
}

// TestKeyboardOperatorPickDiscard 驗證棄牌選取請求: over 載於 count; 規則同 PickCard。
func (this *SuiteOperator) TestKeyboardOperatorPickDiscard() {
	bus := make(chan turn)
	target := keyboardOperator{turn: bus}
	card := []*cores.Card{{}, {}, {}}
	result := make(chan []*cores.Card, 1)

	go func() { result <- target.PickDiscard("手牌上限 要求選手牌", card, 1) }()
	next := <-bus
	this.Equal("手牌上限 要求選手牌", next.req.prompt)
	this.Equal(card, next.req.card)
	this.Equal(1, next.req.count)
	next.req.answer <- answer{card: card[:1]}
	this.Equal(card[:1], <-result)
}
