package games

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/yinweli/RovingDiner/internal/cores"
	"github.com/yinweli/RovingDiner/internal/rules"
)

func TestSuiteExecute(t *testing.T) {
	suite.Run(t, new(SuiteExecute))
}

// SuiteExecute 驗證命令派發(execute.go): games 解析 → Game 執行的端到端接線(走法 X); 詞彙經 rules.Register 裝備。
type SuiteExecute struct {
	suite.Suite
}

func (this *SuiteExecute) TestExecuteAssign() {
	game := cores.NewGame(0, 0, nil, nil, nil, nil)
	rules.Register(game)
	game.GetScore().Set(10)

	command, err := Parse("score += 5") // Parse → execute → Game.ExecAssign
	this.Require().NoError(err)
	execute(game, command)
	this.Equal(int32(15), game.GetScore().GetValue())
}

func (this *SuiteExecute) TestExecuteOperate() {
	game := cores.NewGame(0, 0, nil, nil, nil, nil)
	rules.Register(game)

	command, err := Parse("phaseJump(none, '玩家行動')") // Parse → execute → Game.ExecOperate(端到端派發)
	this.Require().NoError(err)
	execute(game, command)
	this.Equal(cores.PhasePlayerAction, game.GetNextPhase())
}
