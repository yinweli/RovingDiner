package rodi

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/yinweli/RovingDiner/internal/cores"
	sheeter "github.com/yinweli/RovingDiner/sheet"
)

func TestSuiteModalCard(t *testing.T) {
	suite.Run(t, new(SuiteModalCard))
}

// SuiteModalCard 驗證卡牌 modal(modalCard.go): 規格 / 旗標配對對齊、靜態補欄、cardify 來源、
// 實例效果橫排重複列出。
type SuiteModalCard struct {
	suite.Suite
}

// TestModalCardBody 驗證內容行(全量釘字串): 非 cardify 卡來源留空、旗標純鎖含 [鎖0]、
// 效果多重集合重複列出。
func (this *SuiteModalCard) TestModalCardBody() {
	game := testGame()
	card := cores.NewCard(game, 101) // #1; 資料: 群組 3 / 技能 301 / 費用 2 / 額外 1~3 / Seal 鎖 1

	title, row := modalCard{card: card}.Body(game)
	this.Equal("卡牌檢視 101@上菜#1", title)
	this.Equal([]string{
		"卡牌群組 3          出牌費用 2",
		"額外發動次數下限 1  額外發動次數上限 3",
		"卡牌技能 301@開朗",
		"卡牌化來源",
		"+- 旗標",
		"不棄卡牌[鎖0]       封印卡牌[鎖1]",
		"出牌後流放[鎖0]     未出牌流放[鎖0]",
		"+- 效果",
		"401@加耐  401@加耐  402@護盾", // 技能種子的多重集合: 401 重複出現即重複列出
	}, row)

	guest := cores.NewGuest(game, 501) // #2; cardify 來源含實例段
	bound := cores.NewCard(game, 101)  // #3
	bound.CardifyBind(guest)
	_, row = modalCard{card: bound}.Body(game)
	this.Equal("卡牌化來源 501@老饕#2", row[3])

	empty := cores.NewGame(0, 0, cores.NewData(&sheeter.Sheeter{}, nil), nil, nil, nil)
	_, row = modalCard{card: card}.Body(empty) // 對空表查無編號(防禦): 靜態欄全零值照排
	this.Contains(row, "卡牌技能 0@?")
}
