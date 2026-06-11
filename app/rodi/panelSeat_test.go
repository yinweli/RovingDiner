package rodi

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
	"github.com/stretchr/testify/suite"

	"github.com/yinweli/RovingDiner/internal/cores"
	sheeter "github.com/yinweli/RovingDiner/sheet"
)

func TestSuitePanelSeat(t *testing.T) {
	suite.Run(t, new(SuitePanelSeat))
}

// SuitePanelSeat 驗證座位組件(panelSeat.go): 桌 strip / 2 行摘要 / 旗標列 / 固定桌欄寬 / 截斷記號。
type SuitePanelSeat struct {
	suite.Suite
}

// TestPanelSeatView 驗證渲染: 桌號列 + 桌內 2 座各 2 行、桌欄固定寬 25、空位顯 空、超寬桌號列補右緣 >。
func (this *SuitePanelSeat) TestPanelSeatView() {
	game := testGame()
	guest := cores.NewGuest(game, 501)
	game.Seat.Place(1, guest)
	game.Effect.Push(cores.NewEffect(game, 401, cores.NewRefGuest(guest), 1))

	this.Equal(strings.Join([]string{ // 顧客 501: 封(資料 SateSeal)+ 效(佇列效果)命中、免 留白
		"+- 座位 " + strings.Repeat("-", 51) + "+",
		"| " + padTo("桌1"+strings.Repeat(" ", 24)+"桌2", 56) + " |",
		"| " + padTo("501@老饕 飽0耐3"+strings.Repeat(" ", 12)+"空", 56) + " |",
		"| " + padTo(strings.Repeat(" ", 9)+"封  效", 56) + " |",
		"| " + padTo("空", 56) + " |",
		"| " + padTo("", 56) + " |",
	}, "\n"), (&panelSeat{}).View(game, 60, false))

	guest.GetEffectImmune().Add(5) // 免疫計數 > 0 → 免 點亮(M22 拍板)
	this.Contains((&panelSeat{}).View(game, 60, false), strings.Repeat(" ", 9)+"封免效")

	row := strings.Split((&panelSeat{}).View(game, 10, false), "\n") // 超寬: 內容寬 6、桌號列 > 站最後內容格
	this.Equal("| 桌1  > |", row[1])

	game.Seat[2] = nil // 防禦: 座位表 nil 項 → 顯 空
	this.Contains((&panelSeat{}).View(game, 60, false), "空")

	game.Effect = nil // 無效果 → 效 槽熄滅(hasEffect 掃完未命中)
	this.NotContains((&panelSeat{}).View(game, 60, false), "效")
}

// TestPanelSeatMove 驗證游標移動: 左右換桌、上下切座、夾界不迴繞(桌2 僅 1 座切不下去);
// 聚焦時游標顧客格 2 行整格反白(未聚焦不顯游標)。
func (this *SuitePanelSeat) TestPanelSeatMove() {
	game := testGame() // 桌1(座1,2) / 桌2(座3)
	target := &panelSeat{}
	target.Move(game, "right")
	this.Equal(1, target.curTable)
	target.Move(game, "right") // 右端夾住
	this.Equal(1, target.curTable)
	target.Move(game, "down") // 桌2 僅 1 座 → 夾回 0
	this.Equal(0, target.curSeat)
	target.Move(game, "left")
	target.Move(game, "down")
	this.Equal(0, target.curTable)
	this.Equal(1, target.curSeat)
	target.Move(game, "up")
	this.Equal(0, target.curSeat)

	lipgloss.SetColorProfile(termenv.ANSI) // 臨時升 profile 使樣式可見(同 TestFocusView)
	defer lipgloss.SetColorProfile(termenv.Ascii)
	row := strings.Split((&panelSeat{}).View(game, 60, true), "\n")
	this.Contains(row[2], styleCursor.Render(padTo("空", seatColumn))) // 游標格(空位)整格 2 行反白
	this.Contains(row[3], styleCursor.Render(padTo("", seatColumn)))
	this.NotContains((&panelSeat{}).View(game, 60, false), styleCursor.Render(padTo("空", seatColumn))) // 未聚焦不顯游標

	row = strings.Split((&panelSeat{curTable: 1}).View(game, 33, true), "\n") // 窄寬: 桌窗格捲到游標桌
	this.Contains(row[1], "| < 桌2")                                           // 左緣 < 於桌號列
	this.Contains(row[2], styleCursor.Render(padTo("空", seatColumn)))         // 其餘行同縮排, 游標格仍反白
}

// TestPanelSeatItem 驗證游標項目: 游標座位上的顧客; 空位與無桌(防禦)回 nil。
func (this *SuitePanelSeat) TestPanelSeatItem() {
	game := testGame()
	guest := cores.NewGuest(game, 501)
	game.Seat.Place(1, guest)
	target := &panelSeat{}
	this.Equal(guest, target.Item(game)) // 桌1 座位1

	target.Move(game, "right") // 桌2 空位
	this.Nil(target.Item(game))

	empty := cores.NewGame(0, 0, cores.NewData(&sheeter.Sheeter{}, nil), nil, nil, nil) // 防禦: 座位表無桌
	this.Nil(target.Item(empty))
}
