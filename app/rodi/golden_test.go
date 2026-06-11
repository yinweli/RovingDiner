package rodi

import (
	"flag"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/yinweli/RovingDiner/internal/cores"
	"github.com/yinweli/RovingDiner/internal/games"
	"github.com/yinweli/RovingDiner/internal/tester"
)

// updateGolden 重錄 golden 快照(go test ./app/rodi -run TestSuiteGolden -update)。
var updateGolden = flag.Bool("update", false, "重錄 golden 行歷史快照")

func TestSuiteGolden(t *testing.T) {
	suite.Run(t, new(SuiteGolden))
}

// SuiteGolden 驗證日誌行歷史 golden(M24 R0 錄製): 固定 seed×關卡跑完整營業, 行歷史與 testdata 快照
// 逐 byte 相同(CRLF 正規化除外); -update 重錄。引擎決定性使快照釘住整個規則引擎的敘事行為——
// 它是 M24 換軌的驗收基準(引擎直出行必須逐 byte 重現), 其後常駐為核心回歸網(家隨 R2 遷往 games)。
type SuiteGolden struct {
	suite.Suite
}

// TestGolden 對矩陣逐案比對(或重錄): 601 完整開局→成功 / 603 耗到上限→失敗 各三 seed、
// 602 壞引用防禦路徑一 seed(輸出與 seed 無關)。
func (this *SuiteGolden) TestGolden() {
	for _, itor := range []struct {
		seed    int64
		stageID int32
	}{{1, 601}, {2, 601}, {3, 601}, {1, 603}, {2, 603}, {3, 603}, {1, 602}} {
		name := "golden_" + strconv.FormatInt(int64(itor.stageID), 10) + "_" + strconv.FormatInt(itor.seed, 10) + ".txt"
		path := filepath.Join("testdata", name)
		text := strings.Join(goldenLine(itor.seed, itor.stageID), "\n") + "\n"

		if *updateGolden {
			this.NoError(os.MkdirAll("testdata", 0o750))
			this.NoError(os.WriteFile(path, []byte(text), 0o644))
			continue
		} // if

		data, err := os.ReadFile(path)
		this.NoError(err, name)
		this.Equal(text, strings.ReplaceAll(string(data), "\r\n", "\n"), name)
	} // for
}

// goldenLine 跑完整營業並收集行歷史(現管線: 事件 → journal 轉寫; Operator 與 games 側測試同用 FakeOperator)。
func goldenLine(seed int64, stageID int32) []string {
	target := newJournal(tester.BuildSheet())
	games.Run(seed, stageID, tester.BuildSheet(), tester.FakeOperator{}, journalPresenter{journal: target})
	return target.line
}

// journalPresenter 把事件流直灌 journal 的錄製 presenter(golden 管線的接頭)。
type journalPresenter struct {
	journal *journal // 行合成器(行歷史收集處)
}

func (this journalPresenter) Emit(eventData cores.EventData) {
	this.journal.Append(eventData)
}
