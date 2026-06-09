package games

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/yinweli/RovingDiner/internal/cores"
)

func TestSuiteExecute(t *testing.T) {
	suite.Run(t, new(SuiteExecute))
}

// SuiteExecute 驗證命令派發(execute.go):games 解析 → engine 執行的端到端接線(走法 X)。
type SuiteExecute struct {
	suite.Suite
}

func (this *SuiteExecute) TestExecuteAssign() {
	runtime := cores.NewRuntime(0)
	runtime.Game.Score.Value = 10
	eng := cores.NewEngine(runtime, nil, nil)

	command, err := Parse("score += 5") // Parse → execute → engine.ExecAssign
	this.Require().NoError(err)
	execute(eng, command)
	this.Equal(int32(15), runtime.Game.Score.Value)
}

func (this *SuiteExecute) TestExecuteOperate() {
	eng := cores.NewEngine(cores.NewRuntime(0), nil, nil)

	command, err := Parse("deckToHand(none)") // 操作命令於 M9 前為 no-op,此處只驗派發不 panic
	this.Require().NoError(err)
	execute(eng, command)
}
