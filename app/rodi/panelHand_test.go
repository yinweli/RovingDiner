package rodi

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/yinweli/RovingDiner/internal/cores"
)

func TestSuitePanelHand(t *testing.T) {
	suite.Run(t, new(SuitePanelHand))
}

// SuitePanelHand 驗證手牌組件(panelHand.go): 3 行卡塊 / cardify 來源段 / 旗標命中才顯 / 截斷記號。
type SuitePanelHand struct {
	suite.Suite
}

// TestPanelHandView 驗證渲染: 標題含上限、卡名行帶費用與 cardify 來源、flag A / B 命中才顯、超寬補右緣 >。
func (this *SuitePanelHand) TestPanelHandView() {
	game := testGame()
	game.GetHandMax().Set(5)
	game.Hand.Push(cores.NewCard(game, 101))
	game.Hand.Push(cores.NewCard(game, 103))
	guest := cores.NewGuest(game, 501)
	game.Cardify.Push(guest)
	bound := cores.NewCard(game, 101)
	bound.CardifyBind(guest)
	game.Hand.Push(bound)

	this.Equal(strings.Join([]string{ // 手牌序 = 新進入者在前: [綁定卡 103 101]
		"+- 手牌(3/5) " + strings.Repeat("-", 47),
		"101@上菜[501@老饕] (2)  103@結帳 (1)  101@上菜 (2)",
		"不棄 封印" + strings.Repeat(" ", 15) + "不棄" + strings.Repeat(" ", 10) + "封印",
		"",
	}, "\n"), panelHand{}.View(game, 60)) // 綁定卡: CardifyBind 入不棄鎖 + 資料封印; 103: 資料不棄; 101: 資料封印; flag B 全空留白

	row := strings.Split(panelHand{}.View(game, 12), "\n") // 超寬: 補右緣 >
	this.Equal("101@上菜[5 >", row[1])
}

// TestHandDim 驗證暗色標記判定: 出不起(費用 > 出牌點數)或封印命中、付得起且未封印不命中。
func (this *SuitePanelHand) TestHandDim() {
	game := testGame()
	game.GetEnergy().Set(2)
	sealed := cores.NewCard(game, 101) // 101: 費用 2、資料封印
	this.True(handDim(game, sealed))

	open := cores.NewCard(game, 103) // 103: 費用 1、未封印
	this.False(handDim(game, open))

	game.GetEnergy().Set(0) // 出不起
	this.True(handDim(game, open))
}

// TestHandFlagB 驗證 flag B: 出放 / 未放命中並列。
func (this *SuitePanelHand) TestHandFlagB() {
	card := cores.NewCard(testGame(), 101)
	this.Equal("", handFlagB(card))

	card.GetPlayExile().Lock()
	this.Equal("出放", handFlagB(card))

	card.GetUnplayExile().Lock()
	this.Equal("出放 未放", handFlagB(card))
}
