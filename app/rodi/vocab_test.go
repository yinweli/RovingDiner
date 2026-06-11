package rodi

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/yinweli/RovingDiner/internal/cores"
)

func TestSuiteVocab(t *testing.T) {
	suite.Run(t, new(SuiteVocab))
}

// SuiteVocab 驗證詞彙對照(vocab.go): 各轉換函式的中文名與查無退回原文。
type SuiteVocab struct {
	suite.Suite
}

// TestAttrText 驗證屬性詞條鍵轉中文名: 全域 / 引用分表、同名鍵分流、查無退回原文。
func (this *SuiteVocab) TestAttrText() {
	this.Equal("出牌點數", attrText("energy", true))
	this.Equal("餐廳士氣值", attrText("morale", true))
	this.Equal("士氣值", attrText("morale", false)) // 同名鍵取顧客義
	this.Equal("出牌費用", attrText("cost", false))
	this.Equal("效果免疫群組", attrText("effectImmune", false))
	this.Equal("foo", attrText("foo", true)) // 查無退回原文
	this.Equal("bar", attrText("bar", false))
}

// TestTriggerText 驗證觸發時機中文名; 查無退回原文。
func (this *SuiteVocab) TestTriggerText() {
	this.Equal("玩家出牌", triggerText(cores.TriggerCardPlay))
	this.Equal("士氣受損", triggerText(cores.TriggerDamage))
	this.Equal("weird", triggerText(cores.TriggerKind("weird")))
}

// TestSourceText 驗證選取來源中文名; 查無退回原文。
func (this *SuiteVocab) TestSourceText() {
	this.Equal("手牌上限", sourceText("discardOver"))
	this.Equal("顧客指定", sourceText("guestPick"))
	this.Equal("weird", sourceText("weird"))
}

// TestScopeText 驗證範圍事件名: 六種範圍 + 時機帶時機名 + 未知顯 ?。
func (this *SuiteVocab) TestScopeText() {
	this.Equal("玩家出牌", scopeText(cores.ScopePlay, ""))
	this.Equal("顧客行動", scopeText(cores.ScopeGuest, ""))
	this.Equal("前置技能", scopeText(cores.ScopePrefix, ""))
	this.Equal("時機:玩家出牌", scopeText(cores.ScopeTrigger, cores.TriggerCardPlay))
	this.Equal("執行結算", scopeText(cores.ScopeSettle, ""))
	this.Equal("手動結束", scopeText(cores.ScopeManual, ""))
	this.Equal("?", scopeText(cores.ScopeNone, ""))
}

// TestStageText 驗證效果階段名; 未知顯 ?。
func (this *SuiteVocab) TestStageText() {
	this.Equal("立即", stageText(cores.EffectStageImmed))
	this.Equal("加入", stageText(cores.EffectStageJoin))
	this.Equal("觸發", stageText(cores.EffectStageTrigger))
	this.Equal("啟動", stageText(cores.EffectStageStart))
	this.Equal("結束", stageText(cores.EffectStageEnd))
	this.Equal("條件不成立", stageText(cores.EffectStageCondFail))
	this.Equal("?", stageText(cores.EffectStageNone))
}

// TestAssignText 驗證賦值符字面; 未知顯 ?。
func (this *SuiteVocab) TestAssignText() {
	this.Equal("=", assignText(cores.AssignSet))
	this.Equal("+=", assignText(cores.AssignAdd))
	this.Equal("-=", assignText(cores.AssignSub))
	this.Equal("*=", assignText(cores.AssignMul))
	this.Equal("/=", assignText(cores.AssignDiv))
	this.Equal("%=", assignText(cores.AssignMod))
	this.Equal("@", assignText(cores.AssignLock))
	this.Equal("#", assignText(cores.AssignUnlock))
	this.Equal("?", assignText(cores.AssignKind(99)))
}

// TestContainerText 驗證容器名: 與盤面區標籤同名、None 顯 離場、未知顯 ?。
func (this *SuiteVocab) TestContainerText() {
	this.Equal("手牌", containerText(cores.ContainerHand))
	this.Equal("抽牌堆", containerText(cores.ContainerDeck))
	this.Equal("棄牌堆", containerText(cores.ContainerDrop))
	this.Equal("流放堆", containerText(cores.ContainerExile))
	this.Equal("排隊", containerText(cores.ContainerWait))
	this.Equal("座位", containerText(cores.ContainerSeat))
	this.Equal("遊蕩", containerText(cores.ContainerRoam))
	this.Equal("卡牌化", containerText(cores.ContainerCardify))
	this.Equal("離場", containerText(cores.ContainerNone))
	this.Equal("?", containerText(cores.ContainerKind(99)))
}

// TestTaskText 驗證行動類型全名; 未知顯 ?。
func (this *SuiteVocab) TestTaskText() {
	this.Equal("飽食", taskText(cores.TaskSate))
	this.Equal("耐心", taskText(cores.TaskCalm))
	this.Equal("?", taskText(cores.TaskKind(99)))
}

// TestPhaseName 驗證階段中文全名; 無階段 / 未知顯 -。
func (this *SuiteVocab) TestPhaseName() {
	this.Equal("營業開始", phaseName(cores.PhaseGameStart))
	this.Equal("回合開始", phaseName(cores.PhaseRoundStart))
	this.Equal("玩家行動", phaseName(cores.PhasePlayerAction))
	this.Equal("顧客行動", phaseName(cores.PhaseGuestAction))
	this.Equal("回合結束", phaseName(cores.PhaseRoundEnd))
	this.Equal("營業成功", phaseName(cores.PhaseGameSucc))
	this.Equal("營業失敗", phaseName(cores.PhaseGameFail))
	this.Equal("-", phaseName(cores.PhaseNone))
}
