package tui

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/yinweli/RovingDiner/internal/cores"
)

func TestSuiteOperator(t *testing.T) {
	suite.Run(t, new(SuiteOperator))
}

// SuiteOperator 驗證被動觀看替身（operator.go）：永不出牌、選取一律取候選前綴。
type SuiteOperator struct {
	suite.Suite
}

func (this *SuiteOperator) TestPassiveOperatorPlayerAction() {
	this.Nil(passiveOperator{}.PlayerAction(nil))
}

func (this *SuiteOperator) TestPassiveOperatorPickGuest() {
	guest := []*cores.Guest{{}, {}, {}}
	this.Equal(guest[:2], passiveOperator{}.PickGuest(guest, 2))
}

func (this *SuiteOperator) TestPassiveOperatorPickCard() {
	card := []*cores.Card{{}, {}, {}}
	this.Equal(card[:2], passiveOperator{}.PickCard(card, 2))
}

func (this *SuiteOperator) TestPassiveOperatorPickDiscard() {
	card := []*cores.Card{{}, {}, {}}
	this.Equal(card[:1], passiveOperator{}.PickDiscard(card, 1))
}
