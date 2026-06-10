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
	game := cores.NewGame(0, nil, nil, nil, nil)
	game.GetScore().Set(10)

	command, err := Parse("score += 5") // Parse → execute → engine.ExecAssign
	this.Require().NoError(err)
	execute(game, command)
	this.Equal(int32(15), game.GetScore().GetValue())
}

func (this *SuiteExecute) TestExecuteOperate() {
	game := cores.NewGame(0, nil, nil, nil, nil)

	command, err := Parse("phaseJump(none, '玩家行動')") // Parse → execute → engine.ExecOperate(端到端派發)
	this.Require().NoError(err)
	execute(game, command)
	this.Equal(cores.PhasePlayerAction, game.GetNextPhase())
}
