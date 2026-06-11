package tester

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/yinweli/RovingDiner/internal/cores"
)

func TestSuiteTester(t *testing.T) {
	suite.Run(t, new(SuiteTester))
}

// SuiteTester 驗證測試基建(tester.go); 現僅事件流錄製器, 其餘替身與迷你表行為由各層測試間接覆蓋。
type SuiteTester struct {
	suite.Suite
}

// TestRecordPresenterEmit 驗證 Emit 逐筆依序追加事件、欄位原樣保存。
func (this *SuiteTester) TestRecordPresenterEmit() {
	record := &RecordPresenter{}
	record.Emit(cores.EventData{Kind: cores.EventPhase, Phase: cores.PhaseGameStart})
	record.Emit(cores.EventData{Kind: cores.EventScope, Scope: cores.ScopeSettle, Round: 3})
	this.Require().Len(record.Event, 2)
	this.Equal(cores.EventPhase, record.Event[0].Kind)
	this.Equal(cores.PhaseGameStart, record.Event[0].Phase)
	this.Equal(cores.EventScope, record.Event[1].Kind)
	this.Equal(cores.ScopeSettle, record.Event[1].Scope)
	this.Equal(int32(3), record.Event[1].Round)
}
