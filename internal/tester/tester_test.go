package tester

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/yinweli/RovingDiner/internal/cores"
)

func TestSuiteTester(t *testing.T) {
	suite.Run(t, new(SuiteTester))
}

// SuiteTester 驗證測試基建(tester.go): 腳本替身與日誌流錄製器; 其餘替身與迷你表行為由各層測試間接覆蓋。
type SuiteTester struct {
	suite.Suite
}

// TestScriptOperatorPlayerAction 驗證出牌腳本: 筆內依編號序取首張同號手牌、查無略過,
// 筆盡回 nil 且後續筆留給下個階段, 腳本盡恆回玩家結束。
func (this *SuiteTester) TestScriptOperatorPlayerAction() {
	game := cores.NewGame(0, 0, cores.NewData(BuildSheet(), nil), nil, nil, nil)
	first := cores.NewCard(game, 101)
	second := cores.NewCard(game, 102)
	game.Hand = cores.CardList{first, second}
	target := &ScriptOperator{Play: [][]int32{{102, 999, 101}, {101}}}
	this.Same(second, target.PlayerAction(game)) // 編號 102 → 第二張
	this.Same(first, target.PlayerAction(game))  // 999 查無略過 → 101
	this.Nil(target.PlayerAction(game))          // 筆盡 → 玩家結束
	this.Same(first, target.PlayerAction(game))  // 下個階段消費次筆
	this.Nil(target.PlayerAction(game))          // 次筆盡
	this.Nil(target.PlayerAction(game))          // 腳本盡恆回玩家結束
}

// TestScriptOperatorPickGuest 驗證選顧客腳本: 依索引取候選(順序依筆、越界略過、空筆棄選), 佇列盡取候選前綴。
func (this *SuiteTester) TestScriptOperatorPickGuest() {
	guest := []*cores.Guest{{}, {}, {}}
	target := &ScriptOperator{Guest: [][]int{{2, 0, 9}, {}}}
	this.Equal([]*cores.Guest{guest[2], guest[0]}, target.PickGuest("", guest, 2)) // 索引 2, 0; 9 越界略過
	this.Nil(target.PickGuest("", guest, 1))                                       // 空筆 = 棄選
	this.Equal(guest[:2], target.PickGuest("", guest, 2))                          // 佇列盡 → 前綴
}

// TestScriptOperatorPickCard 驗證選卡牌腳本: 依索引取候選, 佇列盡取候選前綴。
func (this *SuiteTester) TestScriptOperatorPickCard() {
	card := []*cores.Card{{}, {}}
	target := &ScriptOperator{Card: [][]int{{1}}}
	this.Equal([]*cores.Card{card[1]}, target.PickCard("", card, 1))
	this.Equal(card[:1], target.PickCard("", card, 1)) // 佇列盡 → 前綴
}

// TestScriptOperatorPickDiscard 驗證棄牌腳本: 依索引取候選, 佇列盡棄候選前 over 張。
func (this *SuiteTester) TestScriptOperatorPickDiscard() {
	card := []*cores.Card{{}, {}, {}}
	target := &ScriptOperator{Discard: [][]int{{2}}}
	this.Equal([]*cores.Card{card[2]}, target.PickDiscard("", card, 1))
	this.Equal(card[:1], target.PickDiscard("", card, 1)) // 佇列盡 → 前綴 over 張
}

// TestRecordPresenterEmit 驗證 Emit 逐拍依序追加行組、行原樣保存。
func (this *SuiteTester) TestRecordPresenterEmit() {
	record := &RecordPresenter{}
	record.Emit("[R1 營業開始] 前置技能", "* 301@")
	record.Emit("$ x")
	this.Require().Len(record.Line, 2)
	this.Equal([]string{"[R1 營業開始] 前置技能", "* 301@"}, record.Line[0])
	this.Equal([]string{"$ x"}, record.Line[1])
}

// TestRecordPresenterFlat 驗證 Flat 攤平行組為行序列(拍的分組不入序列)。
func (this *SuiteTester) TestRecordPresenterFlat() {
	record := &RecordPresenter{}
	record.Emit("a", "b")
	record.Emit("c")
	this.Equal([]string{"a", "b", "c"}, record.Flat())
}
