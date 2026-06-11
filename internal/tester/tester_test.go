package tester

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

func TestSuiteTester(t *testing.T) {
	suite.Run(t, new(SuiteTester))
}

// SuiteTester 驗證測試基建(tester.go); 現僅日誌流錄製器, 其餘替身與迷你表行為由各層測試間接覆蓋。
type SuiteTester struct {
	suite.Suite
}

// TestRecordPresenterEmit 驗證 Emit 逐拍依序追加行組、行原樣保存。
func (this *SuiteTester) TestRecordPresenterEmit() {
	record := &RecordPresenter{}
	record.Emit("[R1 營業開始] 前置技能", "* 301@")
	record.Emit("$ x")
	this.Require().Len(record.Line, 2)
	this.Equal([]string{"[R1 營業開始] 前置技能", "* 301@"}, record.Line[0])
	this.Equal([]string{"$ x"}, record.Line[1])
}

// TestRecordPresenterFlat 驗證 Flat 攤平行組為行序列(拍的分組不入序列)。
func (this *SuiteTester) TestRecordPresenterFlat() {
	record := &RecordPresenter{}
	record.Emit("a", "b")
	record.Emit("c")
	this.Equal([]string{"a", "b", "c"}, record.Flat())
}
