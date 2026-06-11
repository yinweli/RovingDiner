package rodi

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/yinweli/RovingDiner/internal/cores"
)

func TestSuiteJournal(t *testing.T) {
	suite.Run(t, new(SuiteJournal))
}

// SuiteJournal 驗證事件日誌轉寫器(journal.go): 逐類事件的行形狀、歸因狀態機、不立行事件。
type SuiteJournal struct {
	suite.Suite
}

// TestJournalScope 驗證範圍標題 + 操作元行: 六種範圍事件名、動作類操作元依序卡牌(或顧客)、技能,
// 技能編號零值略過, 並重置歸因(標題後的命令行回到流程直屬)。
func (this *SuiteJournal) TestJournalScope() {
	target := newJournal(testSheet())
	target.Append(cores.EventData{Kind: cores.EventScope, Round: 3, Phase: cores.PhasePlayerAction, Scope: cores.ScopePlay, DataID: 101, InstanceID: 5, SkillID: 301})
	this.Equal([]string{"[R3 玩家行動] 玩家出牌", "* 101@上菜#5", "* 301@開朗"}, target.line)

	target = newJournal(testSheet())
	target.Append(cores.EventData{Kind: cores.EventScope, Round: 2, Phase: cores.PhaseGuestAction, Scope: cores.ScopeGuest, DataID: 501, InstanceID: 3, SkillID: 301})
	this.Equal([]string{"[R2 顧客行動] 顧客行動", "* 501@老饕#3", "* 301@開朗"}, target.line)

	target = newJournal(testSheet())
	target.Append(cores.EventData{Kind: cores.EventScope, Phase: cores.PhaseGameStart, Scope: cores.ScopePrefix, SkillID: 301})
	this.Equal([]string{"[R0 營業開始] 前置技能", "* 301@開朗"}, target.line)

	target = newJournal(testSheet())
	target.Append(cores.EventData{Kind: cores.EventScope, Round: 3, Phase: cores.PhasePlayerAction, Scope: cores.ScopeTrigger, Trigger: cores.TriggerCardPlay})
	this.Equal([]string{"[R3 玩家行動] 時機:玩家出牌"}, target.line)

	target = newJournal(testSheet())
	target.Append(cores.EventData{Kind: cores.EventScope, Round: 3, Phase: cores.PhaseRoundEnd, Scope: cores.ScopeSettle})
	target.Append(cores.EventData{Kind: cores.EventScope, Round: 3, Phase: cores.PhasePlayerAction, Scope: cores.ScopeManual})
	this.Equal([]string{"[R3 回合結束] 執行結算", "[R3 玩家行動] 手動結束"}, target.line)

	target = newJournal(testSheet()) // 出牌無技能 → 略過技能操作元行
	target.Append(cores.EventData{Kind: cores.EventScope, Round: 1, Phase: cores.PhasePlayerAction, Scope: cores.ScopePlay, DataID: 101, InstanceID: 5})
	this.Equal([]string{"[R1 玩家行動] 玩家出牌", "* 101@上菜#5"}, target.line)

	target = newJournal(testSheet()) // 標題重置歸因: 效果後的標題使命令行回到 $
	target.Append(cores.EventData{Kind: cores.EventEffect, EffectID: 401, EffectInstanceID: 7, DataID: 501, InstanceID: 3, Stage: cores.EffectStageTrigger})
	target.Append(cores.EventData{Kind: cores.EventScope, Round: 3, Phase: cores.PhaseRoundEnd, Scope: cores.ScopeSettle})
	target.Append(cores.EventData{Kind: cores.EventProperty, Attr: "score", Op: cores.AssignAdd, Operand: 14, Before: 1250, After: 1264})
	this.Equal("$ 餐廳滿意值 += 14 >> 1264", target.line[len(target.line)-1])
}

// TestJournalEffect 驗證效果行 + 效果欄位行: - 識別碼、欄 2 對象(無目標顯 空)、欄 2 階段。
func (this *SuiteJournal) TestJournalEffect() {
	target := newJournal(testSheet())
	target.Append(cores.EventData{Kind: cores.EventEffect, EffectID: 401, EffectInstanceID: 7, DataID: 501, InstanceID: 3, Stage: cores.EffectStageTrigger})
	this.Equal([]string{"- 401@加耐#7", "  501@老饕#3", "  觸發"}, target.line)

	target = newJournal(testSheet()) // 立即類無實例段、無目標顯 空
	target.Append(cores.EventData{Kind: cores.EventEffect, EffectID: 403, Stage: cores.EffectStageImmed})
	this.Equal([]string{"- 403@立即", "  空", "  立即"}, target.line)
}

// TestJournalProperty 驗證屬性命令行: 全域全名 / 效果 self 屬性開頭 / 異對象識別碼開頭、
// 鎖定無算術式、免疫詞條查詢函式形、歸因分流欄 2 與 $。
func (this *SuiteJournal) TestJournalProperty() {
	target := newJournal(testSheet())
	target.Append(cores.EventData{Kind: cores.EventProperty, Attr: "energy", Op: cores.AssignSub, Operand: 2, Before: 10, After: 8})
	this.Equal([]string{"$ 出牌點數 -= 2 >> 8"}, target.line) // 全域 + 無歸因 → $

	target = newJournal(testSheet())
	target.Append(cores.EventData{Kind: cores.EventEffect, EffectID: 401, EffectInstanceID: 7, DataID: 501, InstanceID: 3, Stage: cores.EffectStageTrigger})
	target.Append(cores.EventData{Kind: cores.EventProperty, Attr: "sate", DataID: 501, InstanceID: 3, Op: cores.AssignSub, Operand: 2, Before: 5, After: 3})
	target.Append(cores.EventData{Kind: cores.EventProperty, Attr: "cost", DataID: 101, InstanceID: 5, Op: cores.AssignAdd, Operand: 1, Before: 2, After: 3})
	target.Append(cores.EventData{Kind: cores.EventProperty, Attr: "morale", Op: cores.AssignSub, Operand: 2, Before: 25, After: 23})
	this.Equal([]string{
		"- 401@加耐#7", "  501@老饕#3", "  觸發",
		"  飽食值 -= 2 >> 3",           // self 同對象 → 屬性開頭
		"  101@上菜#5 出牌費用 += 1 >> 3", // 異對象 fan-out → 識別碼開頭
		"  餐廳士氣值 -= 2 >> 23",        // 全域於效果命令中 → 欄 2 全名
	}, target.line)

	target = newJournal(testSheet())
	target.Append(cores.EventData{Kind: cores.EventProperty, Attr: "sateSeal", DataID: 501, InstanceID: 3, Op: cores.AssignLock, After: 1})
	this.Equal([]string{"$ 501@老饕#3 封印飽食技能 @ >> 1"}, target.line) // 鎖定無算術式

	target = newJournal(testSheet())
	target.Append(cores.EventData{Kind: cores.EventProperty, Attr: "effectImmune", DataID: 501, InstanceID: 3, Op: cores.AssignAdd, Operand: 5, Before: 1, After: 2})
	this.Equal([]string{"$ 501@老饕#3 效果免疫群組(5) += 1 >> 2"}, target.line) // 免疫: 群組維度鍵轉查詢函式形
}

// TestJournalContainer 驗證容器命令行: 搬移 >> 去向、入座帶座位編號、離場、重整洗牌行。
func (this *SuiteJournal) TestJournalContainer() {
	target := newJournal(testSheet())
	target.Append(cores.EventData{Kind: cores.EventContainer, DataID: 101, InstanceID: 9, From: cores.ContainerHand, To: cores.ContainerDrop})
	target.Append(cores.EventData{Kind: cores.EventContainer, DataID: 501, InstanceID: 3, From: cores.ContainerWait, To: cores.ContainerSeat, SeatID: 2})
	target.Append(cores.EventData{Kind: cores.EventContainer, DataID: 501, InstanceID: 3, From: cores.ContainerSeat, To: cores.ContainerNone})
	target.Append(cores.EventData{Kind: cores.EventContainer, From: cores.ContainerDeck, To: cores.ContainerDeck, Pick: []cores.PickData{{DataID: 101, InstanceID: 9}}})
	this.Equal([]string{
		"$ 101@上菜#9 >> 棄牌堆",
		"$ 501@老饕#3 >> 座位2",
		"$ 501@老饕#3 >> 離場",
		"$ 抽牌堆 洗牌", // 重整快照: 全序不印(M22 拍板)
	}, target.line)
}

// TestJournalInstance 驗證變身配對: 銷毀暫存不立行、建立合併一行 舊 >> 新; 未見銷毀的建立以 ? 起頭。
func (this *SuiteJournal) TestJournalInstance() {
	target := newJournal(testSheet())
	target.Append(cores.EventData{Kind: cores.EventInstance, DataID: 101, InstanceID: 5, Alive: false})
	this.Empty(target.line) // 銷毀只暫存

	target.Append(cores.EventData{Kind: cores.EventInstance, DataID: 103, InstanceID: 6, Alive: true})
	this.Equal([]string{"$ 101@上菜#5 >> 103@結帳#6"}, target.line)

	target.Append(cores.EventData{Kind: cores.EventInstance, DataID: 103, InstanceID: 8, Alive: true}) // 防禦: 無配對
	this.Equal("$ ? >> 103@結帳#8", target.line[len(target.line)-1])
}

// TestJournalSelect 驗證選取行: 來源詞彙對照 / 效果目標選取顯效果識別碼、選中清單並列、空清單防禦顯 空。
func (this *SuiteJournal) TestJournalSelect() {
	target := newJournal(testSheet())
	target.Append(cores.EventData{Kind: cores.EventSelect, Source: "discardOver", Pick: []cores.PickData{{DataID: 101, InstanceID: 9}, {DataID: 103, InstanceID: 4}}})
	target.Append(cores.EventData{Kind: cores.EventSelect, Source: "guestPick", EffectID: 401, Pick: []cores.PickData{{DataID: 501, InstanceID: 3}}})
	target.Append(cores.EventData{Kind: cores.EventSelect, Source: "weird"})
	this.Equal([]string{
		"$ 選取 手牌上限 -> 101@上菜#9 103@結帳#4",
		"$ 選取 401@加耐 -> 501@老饕#3", // 效果目標選取: 來源顯效果識別碼
		"$ 選取 weird -> 空",
	}, target.line)
}

// TestJournalAction 驗證行動佇列行: 入列立行(歸因分流)、出列不立行(顧客行動標題已承載; M22 拍板)。
func (this *SuiteJournal) TestJournalAction() {
	target := newJournal(testSheet())
	target.Append(cores.EventData{Kind: cores.EventAction, DataID: 501, InstanceID: 3, SkillID: 301, Task: cores.TaskSate, Alive: true})
	this.Equal([]string{"$ 501@老饕#3 >> 行動佇列(飽食)"}, target.line)

	target.Append(cores.EventData{Kind: cores.EventAction, DataID: 501, InstanceID: 3, SkillID: 301, Task: cores.TaskSate, Alive: false})
	this.Len(target.line, 1) // 出列不立行
}

// TestJournalPhase 驗證 phase 切換不立行(階段值併入下一標題前綴)。
func (this *SuiteJournal) TestJournalPhase() {
	target := newJournal(testSheet())
	target.Append(cores.EventData{Kind: cores.EventPhase, Phase: cores.PhaseRoundStart})
	this.Empty(target.line)
}

// TestNumText 驗證數值轉顯示字串: 整數去小數位、小數保留、負值。
func (this *SuiteJournal) TestNumText() {
	this.Equal("2", cores.NumText(2))
	this.Equal("1.5", cores.NumText(1.5))
	this.Equal("-3", cores.NumText(-3))
	this.Equal("0", cores.NumText(0))
}
