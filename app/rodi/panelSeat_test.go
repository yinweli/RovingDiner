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

// TestPanelSeatPick 驗證選取模式(M27 R4): 游標吸附第一個候選; 方向語意保留——左右沿同座位列掃描換桌
// (略過非候選 / 無此座位的桌)、上下限桌內(掃過非候選), 掃不到不動; 已選 / 非候選著色路徑、
// 非本區候選(卡牌選取)照常渲染、stop 後回原語意。
func (this *SuitePanelSeat) TestPanelSeatPick() {
	game := testGame()
	g1 := cores.NewGuest(game, 501)
	g2 := cores.NewGuest(game, 501)
	g3 := cores.NewGuest(game, 501)
	game.Seat.Place(1, g1) // 桌1 上座
	game.Seat.Place(2, g2) // 桌1 下座(先作非候選)
	game.Seat.Place(3, g3) // 桌2(單座)
	pick := &pickState{}
	pick.start(&request{guest: []*cores.Guest{g1, g3}, count: 1})
	target := &panelSeat{curTable: 1, curSeat: 1, pick: pick}
	target.View(game, 60, true) // 游標在桌2 無此座位 → 吸附桌1 上座
	this.Equal(0, target.curTable)
	this.Equal(0, target.curSeat)

	target.Move(game, "down") // 桌1 下座非候選且無更下: 不動(上下限桌內)
	this.Equal(0, target.curSeat)

	target.Move(game, "up") // 無更上: 不動
	this.Equal(0, target.curSeat)

	target.Move(game, "left") // 無更左: 不動
	this.Equal(0, target.curTable)

	target.Move(game, "right") // 同座位列右掃: 桌2 上座 g3
	this.Equal(1, target.curTable)
	this.Equal(0, target.curSeat)

	target.Move(game, "right") // 無更右: 不動
	this.Equal(1, target.curTable)

	target.Move(game, "down") // 桌2 無下座: 不動
	this.Equal(0, target.curSeat)

	target.Move(game, "left") // 同座位列左掃: 回桌1 上座
	this.Equal(0, target.curTable)

	pick.start(&request{guest: []*cores.Guest{g1, g2, g3}, count: 1}) // 三人皆候選: 驗下座視角
	target.curTable, target.curSeat = 0, 1                            // 游標在桌1 下座 g2

	target.Move(game, "right") // 桌2 無下座: 不動(左右不跨座位列)
	this.Equal(0, target.curTable)
	this.Equal(1, target.curSeat)

	target.Move(game, "up") // 桌內上掃: 桌1 上座 g1
	this.Equal(0, target.curSeat)

	pick.toggle(0) // g1 已選 → 已選色底路徑(無 TTY 樣式渲原文, 內容不變)
	this.Contains(target.View(game, 60, false), "501@老饕")
	this.Contains(target.View(game, 60, false), "座位 (請選顧客)") // 標題加註等待落點提醒

	pick.start(&request{card: []*cores.Card{{}}, count: 1}) // 非本區候選: 照常渲染
	this.Equal((&panelSeat{}).View(game, 60, false), (&panelSeat{pick: pick}).View(game, 60, false))

	pick.stop()
	target.curTable, target.curSeat = 0, 0
	target.Move(game, "down") // 非選取模式: 回切座原語意
	this.Equal(1, target.curSeat)
}
