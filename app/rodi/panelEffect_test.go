package rodi

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
	"github.com/stretchr/testify/suite"

	"github.com/yinweli/RovingDiner/internal/cores"
)

func TestSuitePanelEffect(t *testing.T) {
	suite.Run(t, new(SuitePanelEffect))
}

// SuitePanelEffect 驗證效果佇列組件(panelEffect.go): 作用順序排序 / 層數與剩餘回合 / self 辨型 / 類型標記。
type SuitePanelEffect struct {
	suite.Suite
}

// TestPanelEffectView 驗證渲染: 作用順序大者優先、層數 1 省 xN、整場保留顯 永、剩餘回合 = 結束回合 - 當前回合。
func (this *SuitePanelEffect) TestPanelEffectView() {
	game := testGame()
	game.GetRound().Set(2)
	guest := cores.NewGuest(game, 501)
	stacked := cores.NewEffect(game, 401, cores.NewRefGuest(guest), 2)
	stacked.SetExpire(5)
	game.Effect.Push(stacked)
	game.Effect.Push(cores.NewEffect(game, 402, cores.Ref{}, 1))
	game.Effect.Push(cores.NewEffect(game, 403, cores.Ref{}, 1))

	this.Equal(strings.Join([]string{ // 作用順序 9 > 5 排前; 同序 402 < 403 編號小者優先
		"+- 效果佇列(3) " + strings.Repeat("-", 44) + "+",
		"| " + padTo("402@護盾 (永)  403@立即 (永)  401@加耐x2 (3)", 56) + " |",
		"| " + padTo("空 常駐"+strings.Repeat(" ", 8)+"空 ?"+strings.Repeat(" ", 11)+"501@老饕 觸發", 56) + " |",
	}, "\n"), (&panelEffect{}).View(game, 60, false))

	row := strings.Split((&panelEffect{}).View(game, 12, false), "\n") // 超寬: 內容寬 8、> 站最後內容格
	this.Equal("| 402@護 > |", row[1])

	lipgloss.SetColorProfile(termenv.ANSI) // 游標態: 索引對排序後順序(索引 1 = 403@立即), 項目區塊 2 行反白
	defer lipgloss.SetColorProfile(termenv.Ascii)
	row = strings.Split((&panelEffect{cursor: 1}).View(game, 60, true), "\n")
	this.Contains(row[1], styleCursor.Render(padTo("403@立即 (永)", 13)))

	row = strings.Split((&panelEffect{cursor: 2}).View(game, 20, true), "\n") // 窄寬: 窗格捲到游標、左緣 <
	this.Contains(row[1], "| < ")
	this.Contains(row[2], "|   ") // 第 2 行同縮排
}

// TestPanelEffectMove 驗證游標移動: 單列左右、夾界不迴繞。
func (this *SuitePanelEffect) TestPanelEffectMove() {
	game := testGame()
	game.Effect.Push(cores.NewEffect(game, 401, cores.Ref{}, 1))
	game.Effect.Push(cores.NewEffect(game, 402, cores.Ref{}, 1))
	target := &panelEffect{}
	target.Move(game, "right")
	this.Equal(1, target.cursor)
	target.Move(game, "right") // 右端夾住
	this.Equal(1, target.cursor)
	target.Move(game, "left")
	target.Move(game, "left") // 左端夾住
	this.Equal(0, target.cursor)
}

// TestPanelEffectItem 驗證游標項目: 對排序後順序(與 View 同序); 空佇列回 nil。
func (this *SuitePanelEffect) TestPanelEffectItem() {
	game := testGame()
	target := &panelEffect{}
	this.Nil(target.Item(game)) // 空佇列

	low := cores.NewEffect(game, 401, cores.Ref{}, 1)  // 作用順序 5
	high := cores.NewEffect(game, 402, cores.Ref{}, 1) // 作用順序 9 排前
	game.Effect.Push(low)
	game.Effect.Push(high)
	this.Equal(high, target.Item(game)) // 游標 0 = 排序後首位

	target.Move(game, "right")
	this.Equal(low, target.Item(game))
}

// TestEffectOrder 驗證作用順序查表; 查無回 0。
func (this *SuitePanelEffect) TestEffectOrder() {
	this.Equal(int32(9), effectOrder(testSheet(), 402))
	this.Equal(int32(0), effectOrder(testSheet(), 999))
}

// TestEffectSelf 驗證 self 主畫面摘要: Ref 自帶辨型(顧客 / 卡牌)、空物件顯 空。
func (this *SuitePanelEffect) TestEffectSelf() {
	game := testGame()
	this.Equal("501@老饕", effectSelf(testSheet(), cores.NewEffect(game, 401, cores.NewRefGuest(cores.NewGuest(game, 501)), 1)))
	this.Equal("101@上菜", effectSelf(testSheet(), cores.NewEffect(game, 401, cores.NewRefCard(cores.NewCard(game, 101)), 1)))
	this.Equal("空", effectSelf(testSheet(), cores.NewEffect(game, 401, cores.Ref{}, 1))) // 空物件
}

// TestEffectKindName 驗證類型標記: 觸發 / 常駐; 立即與查無顯 ?。
func (this *SuitePanelEffect) TestEffectKindName() {
	this.Equal("觸發", effectKindName(testSheet(), 401))
	this.Equal("常駐", effectKindName(testSheet(), 402))
	this.Equal("?", effectKindName(testSheet(), 403)) // 立即不入列
	this.Equal("?", effectKindName(testSheet(), 999))
}
