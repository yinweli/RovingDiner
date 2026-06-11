package tui

import (
	"github.com/yinweli/RovingDiner/internal/cores"
)

// passiveOperator 被動觀看替身: M19–M23 觀看模式的產品行為本體(人只看不操作), M24 鍵盤 Operator 到位後退役。
// 永不出牌、選取一律取候選前綴, 全同步立答、引擎不暫停; 行為與 tester.FakeOperator 同形但刻意不 import——
// 測試基建的章程是跟著測試需求走, 正式碼不鎖死它的修改自由。
type passiveOperator struct{}

// PlayerAction 恆回玩家結束: 被動觀看不出牌。
func (this passiveOperator) PlayerAction(game *cores.Game) *cores.Card {
	return nil
}

func (this passiveOperator) PickGuest(source []*cores.Guest, count int) []*cores.Guest {
	return source[:count]
}

func (this passiveOperator) PickCard(source []*cores.Card, count int) []*cores.Card {
	return source[:count]
}

func (this passiveOperator) PickDiscard(source []*cores.Card, over int) []*cores.Card {
	return source[:over]
}
