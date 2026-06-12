package rules

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/yinweli/RovingDiner/internal/cores"
	"github.com/yinweli/RovingDiner/internal/tester"
)

func TestSuiteRegister(t *testing.T) {
	suite.Run(t, new(SuiteRegister))
}

// SuiteRegister 驗證詞彙裝備入口(register.go): Register 後 Game 可分派全部詞彙類別。
type SuiteRegister struct {
	suite.Suite
}

// TestRegister 驗證 Register 對 Game 裝備七類詞彙(屬性讀寫 / 引用讀寫 / 命令 / 命令對象 / 內建函式)。
func (this *SuiteRegister) TestRegister() {
	game := cores.NewGame(0, 0, tester.BuildData(), tester.FakeOperator{}, tester.FakeRander{}, nil)
	Register(game)

	_, ok := game.Attr("morale", nil) // 屬性讀
	this.True(ok)
	_, ok = game.Attr("moraleLock", nil) // Lock 全名詞條併入主表
	this.True(ok)

	ref := cores.NewRefGuest(cores.NewGuest(game, 501))
	_, ok = game.AttrRef(ref, "calm", nil) // 引用屬性讀
	this.True(ok)
	_, ok = game.AttrRef(ref, "calmLock", nil)
	this.True(ok)

	this.True(game.ExecAssign("score", "", false, cores.AssignLock, nil)) // 屬性寫詞條(@ 不帶右值)
	this.Equal(int32(1), game.GetScore().GetLock())

	game.ExecOperate("phaseJump", "none", nil, nil) // 命令 + 命令對象詞條已裝備(缺參數 → 詞條內 no-op)
	this.Contains(game.Env().Builtin, "min")        // builtin 詞條
}

// === 測試輔助(置尾) ===

// newGame 組裝全裝備測試營業: 迷你遊戲資料 + 決定性替身 + 全詞彙裝備。各 suite 共用。
func newGame() *cores.Game {
	return newGameData(tester.BuildData())
}

// newGameData 以指定遊戲資料組裝全裝備測試營業(供 Data.SetEffect 注入自訂編譯效果的測試)。
func newGameData(data *cores.Data) *cores.Game {
	game := cores.NewGame(0, 0, data, tester.FakeOperator{}, tester.FakeRander{}, nil)
	Register(game)
	return game
}

// newGameRecord 組裝全裝備測試營業 + 日誌流錄製器(發射接線斷言用)。
func newGameRecord() (game *cores.Game, record *tester.RecordPresenter) {
	return newGameDataRecord(tester.BuildData())
}

// newGameDataRecord 以指定遊戲資料組裝全裝備測試營業 + 日誌流錄製器。
func newGameDataRecord(data *cores.Data) (game *cores.Game, record *tester.RecordPresenter) {
	record = &tester.RecordPresenter{}
	game = cores.NewGame(0, 0, data, tester.FakeOperator{}, tester.FakeRander{}, record)
	Register(game)
	return game, record
}

// filterLine 自錄製行組攤平後篩出指定前綴的行(發射接線測試的行類過濾)。
func filterLine(record *tester.RecordPresenter, prefix string) (result []string) {
	for _, itor := range record.Flat() {
		if strings.HasPrefix(itor, prefix) {
			result = append(result, itor)
		} // if
	} // for

	return result
}
