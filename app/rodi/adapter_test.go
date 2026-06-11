package rodi

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/yinweli/RovingDiner/internal/cores"
	"github.com/yinweli/RovingDiner/internal/games"
	"github.com/yinweli/RovingDiner/internal/tester"
)

func TestSuiteAdapter(t *testing.T) {
	suite.Run(t, new(SuiteAdapter))
}

// SuiteAdapter 驗證引擎橋接器(adapter.go): 雙跑同序(橋不丟事件、不改序、不死鎖)、終局成敗、presenter 半身。
type SuiteAdapter struct {
	suite.Suite
}

// TestNewAdapter 驗證橋接全鏈: 同 seed + 同關卡 + 同輸入, 同步直跑(RecordPresenter)與走橋排水收集的
// 事件序列完全相等, 且 done 回報正確成敗(601 生氣離場清空 → 成功、603 高耐心耗到回合上限 → 失敗)。
func (this *SuiteAdapter) TestNewAdapter() {
	record := &tester.RecordPresenter{}
	succ := games.Run(1, 601, tester.BuildSheet(), tester.FakeOperator{}, record)
	event, bridged := drain(newAdapter(1, 601, tester.BuildSheet(), passiveOperator{}))
	this.Equal(succ, bridged)
	this.True(bridged)
	this.Equal(record.Event, event)

	record = &tester.RecordPresenter{}
	succ = games.Run(1, 603, tester.BuildSheet(), tester.FakeOperator{}, record)
	event, bridged = drain(newAdapter(1, 603, tester.BuildSheet(), passiveOperator{}))
	this.Equal(succ, bridged)
	this.False(bridged)
	this.Equal(record.Event, event)
}

// TestChanPresenterEmit 驗證 presenter 半身: Emit 把事件值原樣送進 channel。
func (this *SuiteAdapter) TestChanPresenterEmit() {
	event := make(chan cores.EventData, 1)
	chanPresenter{event: event}.Emit(cores.EventData{Kind: cores.EventPhase, Round: 3})
	this.Equal(cores.EventData{Kind: cores.EventPhase, Round: 3}, <-event)
}

// drain 排水收集: 收事件到終局為止, 回傳事件序列與成敗。
func drain(target *adapter) (event []cores.EventData, succ bool) {
	for {
		select {
		case eventData := <-target.event:
			event = append(event, eventData)

		case succ = <-target.done:
			return event, succ
		}
	} // for
}
