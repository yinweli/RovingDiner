package rodi

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/yinweli/RovingDiner/internal/cores"
)

func TestSuiteModalAction(t *testing.T) {
	suite.Run(t, new(SuiteModalAction))
}

// SuiteModalAction 驗證行動 modal(modalAction.go): 三欄單列 / 技能效果橫排重複列出 / 離場標記。
type SuiteModalAction struct {
	suite.Suite
}

// TestModalActionBody 驗證內容行: 顧客 / 類型 / 技能單欄逐列、效果多重集合重複列出;
// 顧客不在任一容器附 (已離場)。
func (this *SuiteModalAction) TestModalActionBody() {
	game := testGame()
	guest := cores.NewGuest(game, 501) // #1
	game.Wait.Insert(guest)

	title, row := modalAction{action: cores.NewAction(guest, cores.TaskSate, 301)}.Body(game)
	this.Equal("行動檢視", title)
	this.Equal([]string{
		"顧客 501@老饕#1",
		"行動類型 飽食",
		"技能 301@開朗",
		"+- 效果",
		"401@加耐  401@加耐  402@護盾", // 多重集合: 401 重複引用即重複列出
	}, row)

	gone := cores.NewGuest(game, 501) // 不在任一容器 → 已離場標記
	_, row = modalAction{action: cores.NewAction(gone, cores.TaskCalm, 999)}.Body(game)
	this.Equal("顧客 501@老饕#2 (已離場)", row[0])
	this.Equal("行動類型 耐心", row[1])
	this.Equal("", row[4]) // 查無技能(999): 效果區塊恆顯、列留空
}

// TestTaskText 驗證行動類型全名; 越界顯 ?。
func (this *SuiteModalAction) TestTaskText() {
	this.Equal("飽食", taskText(cores.TaskSate))
	this.Equal("耐心", taskText(cores.TaskCalm))
	this.Equal("?", taskText(cores.TaskKind(9)))
}

// TestGuestGone 驗證離場判定: 在座 / 排隊 / 遊蕩 / 卡牌化任一即在場; nil 與都不在為已離場。
func (this *SuiteModalAction) TestGuestGone() {
	game := testGame()
	this.True(guestGone(game, nil))

	seated := cores.NewGuest(game, 501)
	game.Seat.Place(1, seated)
	this.False(guestGone(game, seated))

	wait := cores.NewGuest(game, 501)
	game.Wait.Insert(wait)
	this.False(guestGone(game, wait))

	this.True(guestGone(game, cores.NewGuest(game, 501)))
}
