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
	runtime.Game.Score = cores.NewValue(10, 0)
	eng := cores.NewEngine(runtime, nil, nil, nil, nil, nil)

	command, err := Parse("score += 5") // Parse → execute → engine.ExecAssign
	this.Require().NoError(err)
	execute(eng, command)
	this.Equal(int32(15), runtime.Game.Score.GetValue())
}

func (this *SuiteExecute) TestExecuteOperate() {
	runtime := cores.NewRuntime(0)
	eng := cores.NewEngine(runtime, nil, nil, nil, nil, nil)

	command, err := Parse("phaseJump(none, '玩家行動')") // Parse → execute → engine.ExecOperate(端到端派發)
	this.Require().NoError(err)
	execute(eng, command)
	this.Equal(cores.PhasePlayerAction, runtime.Game.NextPhase)
}
