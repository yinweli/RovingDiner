package games

import (
	"flag"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/yinweli/RovingDiner/internal/cores"
	"github.com/yinweli/RovingDiner/internal/tester"
)

// updateGolden 重錄 golden 快照(go test ./internal/games -run TestSuiteGolden -update)。
var updateGolden = flag.Bool("update", false, "重錄 golden 行歷史快照")

func TestSuiteGolden(t *testing.T) {
	suite.Run(t, new(SuiteGolden))
}

// SuiteGolden 驗證日誌行歷史 golden: 固定 seed×關卡跑完整營業, 引擎直出的行序列與快照逐 byte 相同
// (CRLF 正規化除外); -update 重錄。快照檔常駐 internal/tester(與錄製來源 BuildSheet / 替身同包)。
// 引擎決定性使快照釘住整個規則引擎的敘事行為——M24 R0 以舊管線(事件→journal)錄製、
// R2 換軌後由引擎直出逐 byte 重現, 其後常駐為核心回歸網(M24 拍板④)。
type SuiteGolden struct {
	suite.Suite
}

// TestGolden 對矩陣逐案比對(或重錄): 601 完整開局→成功 / 603 耗到上限→失敗 各三 seed、
// 602 壞引用防禦路徑一 seed(輸出與 seed 無關)、604 腳本劇情三 seed(M29 輸入腳本維度:
// 出牌 / 三類選取 / 觸發常駐入列皆入快照)。
func (this *SuiteGolden) TestGolden() {
	dir := filepath.Join("..", "tester")

	for _, itor := range []struct {
		seed    int64
		stageID int32
	}{{1, 601}, {2, 601}, {3, 601}, {1, 603}, {2, 603}, {3, 603}, {1, 602}, {1, 604}, {2, 604}, {3, 604}} {
		name := "golden_" + strconv.FormatInt(int64(itor.stageID), 10) + "_" + strconv.FormatInt(itor.seed, 10) + ".txt"
		path := filepath.Join(dir, name)
		text := strings.Join(goldenLine(itor.seed, itor.stageID), "\n") + "\n"

		if *updateGolden {
			this.NoError(os.MkdirAll(dir, 0o750))
			this.NoError(os.WriteFile(path, []byte(text), 0o644))
			continue
		} // if

		data, err := os.ReadFile(path)
		this.NoError(err, name)
		this.Equal(text, strings.ReplaceAll(string(data), "\r\n", "\n"), name)
	} // for
}

// goldenLine 跑完整營業並收集攤平行序列(引擎直出最終行; Operator 用決定性替身)。
func goldenLine(seed int64, stageID int32) []string {
	record := &tester.RecordPresenter{}
	Run(seed, stageID, tester.BuildSheet(), goldenOperator(stageID), record)
	return record.Flat()
}

// goldenOperator 依關卡配發輸入替身: 604 用出牌腳本(每局取新替身, 佇列消費不跨局)——
// 首階段出 104(選顧客餵食)/ 105(選手牌降費 → 爆抽 → 結算棄牌)/ 106(掛觸發 + 常駐), 次階段出素卡 101;
// 其餘關卡玩家不動(FakeOperator)。
func goldenOperator(stageID int32) cores.Operator {
	if stageID == 604 {
		return &tester.ScriptOperator{
			Play:    [][]int32{{104, 105, 106}, {101}},
			Guest:   [][]int{{1}},
			Card:    [][]int{{3}},
			Discard: [][]int{{0}},
		}
	} // if

	return tester.FakeOperator{}
}
