package rules

import (
	"testing"

	"github.com/yinweli/RovingDiner/internal/cores"

	"github.com/stretchr/testify/suite"

	"github.com/yinweli/RovingDiner/internal/exprs"
)

func TestSuiteCommand(t *testing.T) {
	suite.Run(t, new(SuiteCommand))
}

// SuiteCommand 驗證操作命令派發(engine.ExecOperate)、登錄查詢(HasCommand)與 bootstrap 詞條(command.go:phaseJump / deckShuffle)。
// 經 ExecOperate 真實入口測:求值失敗 / 未登錄 → 整動作 no-op;phaseJump 階段合法性;deckShuffle 委派 Rander。
type SuiteCommand struct {
	suite.Suite
}

func (this *SuiteCommand) TestCommandHas() {
	this.True(HasCommand("phaseJump")) // M9.0 bootstrap 詞條已填入
	this.True(HasCommand("deckShuffle"))
	this.True(HasCommand("handAdd"))      // M9.3 實例化
	this.True(HasCommand("cardRun"))      // M9.4 處理流程
	this.False(HasCommand("effectClear")) // 效果佇列類,留效果系統(M12/M13)
}

func (this *SuiteCommand) TestExecOperatePhaseJump() {
	game := newGame()

	game.ExecOperate("phaseJump", "none", nil, this.arg("'玩家行動'"))
	this.Equal(cores.PhasePlayerAction, game.GetNextPhase()) // 合法跳轉值 → 設定

	game.SetNextPhase(cores.PhaseNone)
	game.ExecOperate("phaseJump", "none", nil, this.arg("'營業開始'")) // 非跳轉合法值 → no-op
	this.Equal(cores.PhaseNone, game.GetNextPhase())

	game.ExecOperate("phaseJump", "none", nil, this.arg("5")) // 非字串 → no-op
	this.Equal(cores.PhaseNone, game.GetNextPhase())

	game.ExecOperate("phaseJump", "none", nil, nil) // 缺參數 → no-op
	this.Equal(cores.PhaseNone, game.GetNextPhase())
}

func (this *SuiteCommand) TestExecOperateDeckShuffle() {
	game := newGame()
	game.Deck = cores.CardList{cores.NewCard(game, 101), cores.NewCard(game, 101), cores.NewCard(game, 101)}

	game.ExecOperate("deckShuffle", "none", nil, nil)
	this.Len(game.Deck, 3) // 委派 Rander 洗牌(恆等替身:張數保留)
}

func (this *SuiteCommand) TestExecOperateNoop() {
	game := newGame()
	game.Deck = cores.CardList{cores.NewCard(game, 101)}

	game.ExecOperate("nope", "none", nil, nil)                         // 未知命令 → no-op
	game.ExecOperate("phaseJump", "nope", nil, nil)                    // 未登錄命令對象 → no-op
	game.ExecOperate("deckShuffle", "deckTop", this.arg("1 / 0"), nil) // 命令對象參數評估失敗 → 整動作 no-op
	game.ExecOperate("phaseJump", "none", nil, this.arg("1 / 0"))      // 其餘參數評估失敗 → 整動作 no-op

	this.Equal(cores.PhaseNone, game.GetNextPhase()) // 全程未變更狀態
	this.Len(game.Deck, 1)
}

// === 測試輔助(置尾) ===

// arg 把多個來源各解析為 *exprs.Expr(模擬 parser 產出的命令參數);解析失敗即測試失敗。
func (this *SuiteCommand) arg(source ...string) (result []*exprs.Expr) {
	for _, itor := range source {
		expr, err := exprs.Parse(itor)
		this.Require().NoError(err)
		result = append(result, expr)
	} // for

	return result
}
