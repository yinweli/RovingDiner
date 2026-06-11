package rodi

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/yinweli/RovingDiner/internal/cores"
	"github.com/yinweli/RovingDiner/internal/games"
	"github.com/yinweli/RovingDiner/internal/tester"
)

func TestSuiteStepper(t *testing.T) {
	suite.Run(t, new(SuiteStepper))
}

// SuiteStepper 驗證暫停機橋接器(stepper.go): 雙跑同序(交棒不丟事件、不改序、不死鎖)、終局成敗、
// 終局後防呆、presenter 交棒半身。
type SuiteStepper struct {
	suite.Suite
}

// TestNewStepper 驗證交棒全鏈: 同 seed + 同關卡 + 同輸入, 同步直跑(RecordPresenter)與逐拍 Next 收集的
// 事件序列完全相等, 且終局回報正確成敗(601 生氣離場清空 → 成功、603 高耐心耗到回合上限 → 失敗)。
func (this *SuiteStepper) TestNewStepper() {
	record := &tester.RecordPresenter{}
	succ := games.Run(1, 601, tester.BuildSheet(), tester.FakeOperator{}, record)
	event, bridged := drain(newStepper(1, 601, tester.BuildSheet()))
	this.Equal(succ, bridged)
	this.True(bridged)
	this.Equal(record.Event, event)

	record = &tester.RecordPresenter{}
	succ = games.Run(1, 603, tester.BuildSheet(), tester.FakeOperator{}, record)
	event, bridged = drain(newStepper(1, 603, tester.BuildSheet()))
	this.Equal(succ, bridged)
	this.False(bridged)
	this.Equal(record.Event, event)
}

// TestStepperNext 驗證終局後防呆: 排水到終局後再 Next, 直接回 more == false(不再放行, 不死鎖)。
func (this *SuiteStepper) TestStepperNext() {
	target := newStepper(1, 601, tester.BuildSheet())
	drain(target)
	eventData, more := target.Next()
	this.Equal(cores.EventData{}, eventData)
	this.False(more)
}

// TestGatePresenterEmit 驗證交棒半身: Emit 把事件值原樣送進 channel 後消費一拍放行。
func (this *SuiteStepper) TestGatePresenterEmit() {
	event := make(chan cores.EventData, 1)
	gate := make(chan struct{}, 1)
	gate <- struct{}{} // 預放行: Emit 送完事件即取走、不停棒
	gatePresenter{event: event, gate: gate}.Emit(cores.EventData{Kind: cores.EventPhase, Round: 3})
	this.Equal(cores.EventData{Kind: cores.EventPhase, Round: 3}, <-event)
	this.Empty(gate) // 放行訊號已被消費
}

// drain 排水收集: 逐拍 Next 收事件到終局為止, 回傳事件序列與成敗。
func drain(target *stepper) (event []cores.EventData, succ bool) {
	for {
		eventData, more := target.Next()

		if more == false {
			return event, target.succ
		} // if

		event = append(event, eventData)
	} // for
}
