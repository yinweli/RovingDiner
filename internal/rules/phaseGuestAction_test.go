package rules

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/yinweli/RovingDiner/internal/cores"
	"github.com/yinweli/RovingDiner/internal/tester"
)

func TestSuitePhaseGuestAction(t *testing.T) {
	suite.Run(t, new(SuitePhaseGuestAction))
}

// SuitePhaseGuestAction 驗證顧客行動階段(phaseGuestAction.go): 佇列消費 / 行動事件 / 封印閘門(sealTask)/ 技能啟動 / 觸發收尾。
type SuitePhaseGuestAction struct {
	suite.Suite
}

func (this *SuitePhaseGuestAction) TestPhaseGuestAction() {
	count := 0
	start := 0
	task := 0
	end := 0
	data := tester.BuildData()
	game := newGameData(data)
	data.SetEffect(401, cores.EffectData{Kind: cores.EffectImmed, Immed: func(game *cores.Game) { count++ }})
	data.SetEffect(402, cores.EffectData{Kind: cores.EffectImmed, Immed: func(game *cores.Game) { count++ }})
	data.SetEffect(901, cores.EffectData{Kind: cores.EffectTrigger, TriggerKind: cores.TriggerGuestStart, Trigger: func(game *cores.Game) { start++ }})
	data.SetEffect(902, cores.EffectData{Kind: cores.EffectTrigger, TriggerKind: cores.TriggerGuestTask, Trigger: func(game *cores.Game) { task++ }})
	data.SetEffect(903, cores.EffectData{Kind: cores.EffectTrigger, TriggerKind: cores.TriggerGuestEnd, Trigger: func(game *cores.Game) { end++ }})

	for _, itor := range []int32{901, 902, 903} {
		game.Effect.Push(cores.NewEffect(game, itor, cores.Ref{}, 1))
	} // for

	sateGuest := cores.NewGuest(game, 501)
	sateGuest.GetSateSeal().Unlock()       // 解除顧客資料的預設封印 → 飽食行動可啟動
	calmGuest := cores.NewGuest(game, 501) // calmSeal 未鎖 → 耐心行動可啟動
	sealGuest := cores.NewGuest(game, 501) // sateSeal 鎖(資料預設)→ 跳過啟動
	game.Action.Push(cores.NewAction(sateGuest, cores.TaskSate, 301))
	game.Action.Push(cores.NewAction(calmGuest, cores.TaskCalm, 301))
	game.Action.Push(cores.NewAction(sealGuest, cores.TaskSate, 301))
	game.Action.Push(cores.NewAction(sealGuest, cores.TaskKind(9), 301)) // 越界行動類型 → 跳過啟動(防禦)

	this.Equal(cores.PhaseRoundEnd, phaseGuestAction(game))
	this.Equal(1, start)
	this.Equal(1, end)
	this.Equal(4, task)  // 每筆行動觸發一次, 含跳過啟動者
	this.Equal(4, count) // 僅前兩筆啟動技能 301:2 筆 × 效果 401 / 402
	this.Empty(game.Action)
	this.Equal(int32(4), game.GetTaskCount()) // 行動事件每筆都設
	this.Same(sealGuest, game.GetTaskGuest())
	this.Equal(int32(301), game.GetTaskSkill())

	// 跳轉回合結束 → 立即收尾, 行動不消費
	game.Action.Push(cores.NewAction(calmGuest, cores.TaskCalm, 301))
	game.SetNextPhase(cores.PhaseRoundEnd)
	this.Equal(cores.PhaseRoundEnd, phaseGuestAction(game))
	this.Equal(cores.PhaseNone, game.GetNextPhase())
	this.Len(game.Action, 1)
	this.Equal(2, end)
}

// TestPhaseGuestActionEmit 驗證顧客行動的發射接線: 逐筆行動發範圍標題(操作元 = 顧客 + 技能;
// 出列不發行——標題已承載; M22 拍板)。
func (this *SuitePhaseGuestAction) TestPhaseGuestActionEmit() {
	game, record := newGameRecord()
	guest := cores.NewGuest(game, 501) // 實例編號 1
	game.Seat.Place(1, guest)
	game.Action.Push(cores.NewAction(guest, cores.TaskCalm, 301))

	phaseGuestAction(game)
	this.Require().NotEmpty(record.Line)
	this.Equal([]string{"[R0 -] 顧客行動", "* 501@#1", "* 301@"}, record.Line[0])
}

// TestSealTask 驗證行動封印閘門: 飽食 / 耐心各查對應封印鎖、越界行動類型視為封印。
func (this *SuitePhaseGuestAction) TestSealTask() {
	game := newGame()
	guest := cores.NewGuest(game, 501) // sateSeal 鎖(資料預設)、calmSeal 未鎖

	this.True(sealTask(cores.NewAction(guest, cores.TaskSate, 301))) // 封印飽食 → 跳過
	guest.GetSateSeal().Unlock()
	this.False(sealTask(cores.NewAction(guest, cores.TaskSate, 301)))

	this.False(sealTask(cores.NewAction(guest, cores.TaskCalm, 301)))
	guest.GetCalmSeal().Lock()
	this.True(sealTask(cores.NewAction(guest, cores.TaskCalm, 301))) // 封印耐心 → 跳過

	this.True(sealTask(cores.NewAction(guest, cores.TaskKind(9), 301))) // 越界 → 視為封印(防禦)
}
