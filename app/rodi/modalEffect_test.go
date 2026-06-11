package rodi

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/yinweli/RovingDiner/internal/cores"
	sheeter "github.com/yinweli/RovingDiner/sheet"
)

func TestSuiteModalEffect(t *testing.T) {
	suite.Run(t, new(SuiteModalEffect))
}

// SuiteModalEffect 驗證效果 modal(modalEffect.go): 實例 + 靜態 def 配對對齊、恆顯留空依適用類型矩陣、
// self 辨型與失效標記、列舉值中文。
type SuiteModalEffect struct {
	suite.Suite
}

// TestModalEffectBody 驗證內容行(觸發類型全量釘字串): 命令欄空值僅留標籤、矩陣不適用欄留空。
func (this *SuiteModalEffect) TestModalEffectBody() {
	game := testGame()
	guest := cores.NewGuest(game, 501) // #1
	game.Seat.Place(1, guest)
	effect := cores.NewEffect(game, 401, cores.NewRefGuest(guest), 2) // #2; 401 = 觸發類型帶滿安全靜態欄
	effect.SetExpire(4)

	title, row := modalEffect{effect: effect}.Body(game)
	this.Equal("效果檢視 401@加耐#2", title)
	this.Equal([]string{
		"結束回合 4         當前層數 2",
		"self 501@老饕#1",
		"+- 規格",
		"效果類型 觸發      效果群組 0",
		"目標類型 沿用顧客  目標數量 1",
		"作用回合 3         作用順序 5",
		"堆疊層數 1         堆疊上限 3",
		"堆疊時間 刷新",
		"+- 觸發",
		"觸發時機 玩家出牌  觸發後行為 保留",
		"觸發條件 self.sate > 0",
		"觸發次數 self.calm",
		"+- 命令",
		"立即",
		"觸發",
		"啟動",
		"結束",
	}, row)
}

// TestModalEffectMatrix 驗證恆顯留空: 立即類型(403)不適用佇列 / 觸發欄全空、條件族欄照顯;
// self 空物件顯 空。
func (this *SuiteModalEffect) TestModalEffectMatrix() {
	game := testGame()
	_, row := modalEffect{effect: cores.NewEffect(game, 403, cores.Ref{}, 1)}.Body(game)
	this.Equal("self 空", row[1])
	this.Contains(row, "效果類型 立即    效果群組")
	this.Contains(row, "作用回合         作用順序 9")
	this.Contains(row, "堆疊層數         堆疊上限")
	this.Contains(row, "堆疊時間")
	this.Contains(row, "觸發時機         觸發後行為")

	empty := cores.NewGame(0, 0, cores.NewData(&sheeter.Sheeter{}, nil), nil, nil, nil)
	_, row = modalEffect{effect: cores.NewEffect(game, 403, cores.Ref{}, 1)}.Body(empty) // 對空表查無編號(防禦): 靜態欄全零值照排
	this.Contains(row, "效果類型 立即    效果群組")
}

// TestRefTarget 驗證 self 顯示: 顧客 / 卡牌辨型含實例段、已失效附標記、空物件顯 空。
func (this *SuiteModalEffect) TestRefTarget() {
	game := testGame()
	guest := cores.NewGuest(game, 501) // #1; 不在任一容器 → 已失效
	this.Equal("501@老饕#1 (已失效)", refTarget(game, cores.NewEffect(game, 401, cores.NewRefGuest(guest), 1)))

	card := cores.NewCard(game, 101) // #3
	game.Hand.Push(card)
	this.Equal("101@上菜#3", refTarget(game, cores.NewEffect(game, 401, cores.NewRefCard(card), 1)))

	gone := cores.NewCard(game, 103) // #5; 不在任一牌堆 → 已失效
	this.Equal("103@結帳#5 (已失效)", refTarget(game, cores.NewEffect(game, 401, cores.NewRefCard(gone), 1)))

	this.Equal("空", refTarget(game, cores.NewEffect(game, 401, cores.Ref{}, 1)))
}

// TestCardGone 驗證卡牌失效判定: 手牌 / 抽 / 棄 / 流放任一即在場。
func (this *SuiteModalEffect) TestCardGone() {
	game := testGame()
	card := cores.NewCard(game, 101)
	this.True(cardGone(game, card))

	game.Deck.Push(card)
	this.False(cardGone(game, card))
}

// TestKindText 驗證效果類型中文名; 越界顯 ?。
func (this *SuiteModalEffect) TestKindText() {
	this.Equal("立即", kindText(cores.EffectImmed))
	this.Equal("觸發", kindText(cores.EffectTrigger))
	this.Equal("常駐", kindText(cores.EffectPersist))
	this.Equal("?", kindText(cores.EffectKind(9)))
}

// TestTargetText 驗證目標類型中文名; 越界顯 ?。
func (this *SuiteModalEffect) TestTargetText() {
	this.Equal("無目標", targetText(cores.TargetNone))
	this.Equal("新選顧客", targetText(cores.TargetGuestPick))
	this.Equal("沿用顧客", targetText(cores.TargetGuestSame))
	this.Equal("隨機顧客", targetText(cores.TargetGuestRand))
	this.Equal("新選手牌", targetText(cores.TargetCardPick))
	this.Equal("沿用手牌", targetText(cores.TargetCardSame))
	this.Equal("隨機手牌", targetText(cores.TargetCardRand))
	this.Equal("?", targetText(cores.TargetKind(9)))
}

// TestAfterText 驗證觸發後行為中文名; 越界顯 ?。
func (this *SuiteModalEffect) TestAfterText() {
	this.Equal("保留", afterText(cores.TriggerAfterKeep))
	this.Equal("移除", afterText(cores.TriggerAfterRemove))
	this.Equal("?", afterText(cores.TriggerAfter(9)))
}

// TestStackText 驗證堆疊時間中文名; 越界顯 ?。
func (this *SuiteModalEffect) TestStackText() {
	this.Equal("不變", stackText(cores.StackTimeStay))
	this.Equal("刷新", stackText(cores.StackTimeRefresh))
	this.Equal("?", stackText(cores.StackTime(9)))
}
