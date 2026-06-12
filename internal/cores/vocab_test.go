package cores

import (
	"testing"

	"github.com/stretchr/testify/suite"
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
	this.Equal("出牌點數", AttrText("energy", true))
	this.Equal("餐廳士氣值", AttrText("morale", true))
	this.Equal("士氣值", AttrText("morale", false)) // 同名鍵取顧客義
	this.Equal("出牌費用", AttrText("cost", false))
	this.Equal("效果免疫群組", AttrText(AttrEffectImmune, false))
	this.Equal("foo", AttrText("foo", true)) // 查無退回原文
	this.Equal("bar", AttrText("bar", false))
}

// TestTriggerText 驗證觸發時機中文名; 查無退回原文。
func (this *SuiteVocab) TestTriggerText() {
	this.Equal("玩家出牌", TriggerText(TriggerCardPlay))
	this.Equal("士氣受損", TriggerText(TriggerDamage))
	this.Equal("weird", TriggerText(TriggerKind("weird")))
}

// TestSourceText 驗證選取來源中文名; 查無退回原文。
func (this *SuiteVocab) TestSourceText() {
	this.Equal("手牌上限", SourceText("discardOver"))
	this.Equal("顧客指定", SourceText("guestPick"))
	this.Equal("weird", SourceText("weird"))
}

// TestStageText 驗證效果階段名; 未知顯 ?。
func (this *SuiteVocab) TestStageText() {
	this.Equal("立即", StageText(EffectStageImmed))
	this.Equal("加入", StageText(EffectStageJoin))
	this.Equal("觸發", StageText(EffectStageTrigger))
	this.Equal("啟動", StageText(EffectStageStart))
	this.Equal("結束", StageText(EffectStageEnd))
	this.Equal("條件不成立", StageText(EffectStageCondFail))
	this.Equal("?", StageText(EffectStageNone))
}

// TestAssignText 驗證賦值符字面; 未知顯 ?。
func (this *SuiteVocab) TestAssignText() {
	this.Equal("=", AssignText(AssignSet))
	this.Equal("+=", AssignText(AssignAdd))
	this.Equal("-=", AssignText(AssignSub))
	this.Equal("*=", AssignText(AssignMul))
	this.Equal("/=", AssignText(AssignDiv))
	this.Equal("%=", AssignText(AssignMod))
	this.Equal("@", AssignText(AssignLock))
	this.Equal("#", AssignText(AssignUnlock))
	this.Equal("?", AssignText(AssignKind(99)))
}

// TestContainerText 驗證容器名: 與盤面區標籤同名、None 顯 離場、未知顯 ?。
func (this *SuiteVocab) TestContainerText() {
	this.Equal("手牌", ContainerText(ContainerHand))
	this.Equal("抽牌堆", ContainerText(ContainerDeck))
	this.Equal("棄牌堆", ContainerText(ContainerDrop))
	this.Equal("流放堆", ContainerText(ContainerExile))
	this.Equal("排隊", ContainerText(ContainerWait))
	this.Equal("座位", ContainerText(ContainerSeat))
	this.Equal("遊蕩", ContainerText(ContainerRoam))
	this.Equal("卡牌化", ContainerText(ContainerCardify))
	this.Equal("離場", ContainerText(ContainerNone))
	this.Equal("?", ContainerText(ContainerKind(99)))
}

// TestTaskText 驗證行動類型全名; 未知顯 ?。
func (this *SuiteVocab) TestTaskText() {
	this.Equal("飽食", TaskText(TaskSate))
	this.Equal("耐心", TaskText(TaskCalm))
	this.Equal("?", TaskText(TaskKind(99)))
}

// TestPhaseName 驗證階段中文全名; 無階段 / 未知顯 -。
func (this *SuiteVocab) TestPhaseName() {
	this.Equal("營業開始", PhaseName(PhaseGameStart))
	this.Equal("回合開始", PhaseName(PhaseRoundStart))
	this.Equal("玩家行動", PhaseName(PhasePlayerAction))
	this.Equal("顧客行動", PhaseName(PhaseGuestAction))
	this.Equal("回合結束", PhaseName(PhaseRoundEnd))
	this.Equal("營業成功", PhaseName(PhaseGameSucc))
	this.Equal("營業失敗", PhaseName(PhaseGameFail))
	this.Equal("-", PhaseName(PhaseNone))
}

// TestNumText 驗證數值轉日誌字串: 整數去小數位、小數保留。
func (this *SuiteVocab) TestNumText() {
	this.Equal("5", NumText(5))
	this.Equal("-3", NumText(-3))
	this.Equal("2.5", NumText(2.5))
	this.Equal("0", NumText(0))
}
