package rules

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/yinweli/RovingDiner/internal/cores"
)

func TestSuiteEmit(t *testing.T) {
	suite.Run(t, new(SuiteEmit))
}

// SuiteEmit 驗證日誌行發射輔助(emit.go): 搬移 / 行動 / 變身 / 標題 / 選取的本文組裝;
// 組行原語(標題 / 效果頭 / 屬性行)的模板與歸因語意歸 cores 發射台測試。
type SuiteEmit struct {
	suite.Suite
}

// TestEmitGuestMove 驗證顧客搬移行: 識別碼 >> 去向(入座帶座位編號、ContainerNone 顯 離場)。
func (this *SuiteEmit) TestEmitGuestMove() {
	game, record := newGameRecord()
	guest := cores.NewGuest(game, 501) // 實例編號 1(新局首發; 迷你表無名稱欄)

	emitGuestMove(game, guest, cores.ContainerSeat, 2)
	emitGuestMove(game, guest, cores.ContainerNone, 0)
	this.Require().Len(record.Line, 2)
	this.Equal([]string{"$ 501@#1 >> 座位2"}, record.Line[0])
	this.Equal([]string{"$ 501@#1 >> 離場"}, record.Line[1])
}

// TestEmitCardMove 驗證卡牌搬移行: 識別碼 >> 去向。
func (this *SuiteEmit) TestEmitCardMove() {
	game, record := newGameRecord()
	card := cores.NewCard(game, 101)

	emitCardMove(game, card, cores.ContainerHand)
	this.Require().Len(record.Line, 1)
	this.Equal([]string{"$ 101@#1 >> 手牌"}, record.Line[0])
}

// TestEmitDeckOrder 驗證抽牌牌堆重整行: 全序不印(順序直讀盤面)、只立洗牌行。
func (this *SuiteEmit) TestEmitDeckOrder() {
	game, record := newGameRecord()

	emitDeckOrder(game)
	this.Require().Len(record.Line, 1)
	this.Equal([]string{"$ 抽牌堆 洗牌"}, record.Line[0])
}

// TestEmitAction 驗證行動佇列入列行: 顧客識別碼 >> 行動佇列(行動類型)。
func (this *SuiteEmit) TestEmitAction() {
	game, record := newGameRecord()
	action := cores.NewAction(cores.NewGuest(game, 501), cores.TaskCalm, 301)

	emitAction(game, action)
	this.Require().Len(record.Line, 1)
	this.Equal([]string{"$ 501@#1 >> 行動佇列(耐心)"}, record.Line[0])
}

// TestEmitMorph 驗證變身行: 舊識別碼 >> 新識別碼(一行收口)。
func (this *SuiteEmit) TestEmitMorph() {
	game, record := newGameRecord()
	card := cores.NewCard(game, 101)

	emitMorph(game, 102, 9, card)
	this.Require().Len(record.Line, 1)
	this.Equal([]string{"$ 102@#9 >> 101@#1"}, record.Line[0])
}

// TestEmitPlayTitle 驗證玩家出牌範圍標題: 操作元 = 卡牌 + 技能(卡 103 技能 301; 無技能卡僅卡牌行)。
func (this *SuiteEmit) TestEmitPlayTitle() {
	game, record := newGameRecord()

	emitPlayTitle(game, cores.NewCard(game, 103))
	this.Require().Len(record.Line, 1)
	this.Equal([]string{"[R0 -] 玩家出牌", "* 103@#1", "* 301@"}, record.Line[0])

	emitPlayTitle(game, cores.NewCard(game, 101)) // 卡 101 無技能 → 無技能操作元行
	this.Equal([]string{"[R0 -] 玩家出牌", "* 101@#2"}, record.Line[1])
}

// TestEmitGuestTitle 驗證顧客行動範圍標題: 操作元 = 顧客 + 技能。
func (this *SuiteEmit) TestEmitGuestTitle() {
	game, record := newGameRecord()
	action := cores.NewAction(cores.NewGuest(game, 501), cores.TaskSate, 301)

	emitGuestTitle(game, action)
	this.Require().Len(record.Line, 1)
	this.Equal([]string{"[R0 -] 顧客行動", "* 501@#1", "* 301@"}, record.Line[0])
}

// TestActionOperator 驗證操作元清單組裝: 主體 + 技能(技能編號零值防禦略過)。
func (this *SuiteEmit) TestActionOperator() {
	game := newGame()
	this.Equal([]string{"主體", "301@"}, actionOperator(game, "主體", 301))
	this.Equal([]string{"主體"}, actionOperator(game, "主體", 0))
}

// TestEmitSelect 驗證玩家選取紀錄行: 效果目標選取顯效果識別碼、其餘查詞彙對照; 空清單防禦顯 空。
func (this *SuiteEmit) TestEmitSelect() {
	game, record := newGameRecord()

	emitSelect(game, "guestPick", 401, []string{"501@#1", "501@#2"})
	emitSelect(game, "discardOver", 0, nil)
	this.Require().Len(record.Line, 2)
	this.Equal([]string{"$ 選取 401@ -> 501@#1 501@#2"}, record.Line[0])
	this.Equal([]string{"$ 選取 手牌上限 -> 空"}, record.Line[1])
}

// TestCardIdentList 驗證卡牌識別碼清單組裝(選取紀錄行的選中段)。
func (this *SuiteEmit) TestCardIdentList() {
	game := newGame()
	c1 := cores.NewCard(game, 101)
	c2 := cores.NewCard(game, 103)
	this.Equal([]string{"101@#1", "103@#2"}, cardIdentList(game, []*cores.Card{c1, c2}))
	this.Empty(cardIdentList(game, nil))
}

// TestGuestIdentList 驗證顧客識別碼清單組裝; 規則同 cardIdentList。
func (this *SuiteEmit) TestGuestIdentList() {
	game := newGame()
	this.Equal([]string{"501@#1"}, guestIdentList(game, []*cores.Guest{cores.NewGuest(game, 501)}))
	this.Empty(guestIdentList(game, nil))
}
