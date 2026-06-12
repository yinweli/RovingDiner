package rodi

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/yinweli/RovingDiner/internal/cores"
)

func TestSuitePick(t *testing.T) {
	suite.Run(t, new(SuitePick))
}

// SuitePick 驗證選取模式共享狀態(pick.go): 起停、候選索引(指標相等)、toggle 規則、滿選與答覆組裝、
// 選取提示、候選步進原語。
type SuitePick struct {
	suite.Suite
}

// TestPickStateStart 驗證起停: start 配已選標記、stop 清空、active 隨之; nil 接收器安全(未注入面板)。
func (this *SuitePick) TestPickStateStart() {
	target := &pickState{}
	this.False(target.active())
	target.start(&request{guest: []*cores.Guest{{}, {}}, count: 1})
	this.True(target.active())
	this.Len(target.picked, 2)
	target.stop()
	this.False(target.active())
	this.Nil(target.picked)
	this.False((*pickState)(nil).active()) // nil 接收器走非選取路徑
}

// TestPickStateGuestIndex 驗證顧客候選索引: 指標相等; 非候選 / nil / 非選取模式回 -1。
func (this *SuitePick) TestPickStateGuestIndex() {
	g1, g2 := &cores.Guest{}, &cores.Guest{}
	target := &pickState{}
	this.Equal(-1, target.guestIndex(g1)) // 非選取模式
	target.start(&request{guest: []*cores.Guest{g1, g2}, count: 1})
	this.Equal(0, target.guestIndex(g1))
	this.Equal(1, target.guestIndex(g2))
	this.Equal(-1, target.guestIndex(&cores.Guest{}))
	this.Equal(-1, target.guestIndex(nil))
}

// TestPickStateCardIndex 驗證卡牌候選索引; 規則同 guestIndex。
func (this *SuitePick) TestPickStateCardIndex() {
	c1, c2 := &cores.Card{}, &cores.Card{}
	target := &pickState{}
	this.Equal(-1, target.cardIndex(c1))
	target.start(&request{card: []*cores.Card{c1, c2}, count: 1})
	this.Equal(0, target.cardIndex(c1))
	this.Equal(1, target.cardIndex(c2))
	this.Equal(-1, target.cardIndex(&cores.Card{}))
	this.Equal(-1, target.cardIndex(nil))
}

// TestPickStateToggle 驗證加選 / 取消: 未滿加選、已選取消、滿 N 再加選 no-op(須先取消其一)、
// 越界 / -1 不動作; chosen 同步回報。
func (this *SuitePick) TestPickStateToggle() {
	target := &pickState{}
	target.start(&request{guest: []*cores.Guest{{}, {}, {}}, count: 2})
	target.toggle(0)
	this.True(target.chosen(0))
	target.toggle(1)
	this.Equal(2, target.count())
	target.toggle(2) // 滿 N: no-op
	this.False(target.chosen(2))
	target.toggle(0) // 取消
	this.False(target.chosen(0))
	this.Equal(1, target.count())
	target.toggle(2) // 取消後可再加選
	this.True(target.chosen(2))
	target.toggle(-1) // 越界不動作
	target.toggle(9)
	this.Equal(2, target.count())
	this.False(target.chosen(-1))
	this.False(target.chosen(9))
}

// TestPickStateFull 驗證滿選判定([Enter] 確認條件): 非選取模式 false、選滿 N true。
func (this *SuitePick) TestPickStateFull() {
	target := &pickState{}
	this.False(target.full())
	target.start(&request{guest: []*cores.Guest{{}, {}}, count: 1})
	this.False(target.full())
	target.toggle(1)
	this.True(target.full())
}

// TestPickStateResult 驗證答覆組裝: 依候選序收已選(與加選順序無關、保決定性); 顧客 / 卡牌兩型。
func (this *SuitePick) TestPickStateResult() {
	g1, g2, g3 := &cores.Guest{}, &cores.Guest{}, &cores.Guest{}
	target := &pickState{}
	target.start(&request{guest: []*cores.Guest{g1, g2, g3}, count: 2})
	target.toggle(2) // 倒序加選
	target.toggle(0)
	this.Equal(answer{guest: []*cores.Guest{g1, g3}}, target.result()) // 答覆仍依候選序

	c1, c2 := &cores.Card{}, &cores.Card{}
	target.start(&request{card: []*cores.Card{c1, c2}, count: 1})
	target.toggle(1)
	this.Equal(answer{card: []*cores.Card{c2}}, target.result())
}

// TestPickStateHint 驗證選取提示: 引擎前文 + 已選進度。
func (this *SuitePick) TestPickStateHint() {
	target := &pickState{}
	target.start(&request{prompt: "開朗 要求選顧客", guest: []*cores.Guest{{}, {}, {}}, count: 2})
	this.Equal("開朗 要求選顧客 (已選 0/2)", target.hint())
	target.toggle(0)
	this.Equal("開朗 要求選顧客 (已選 1/2)", target.hint())
}
