package games

import (
	"github.com/yinweli/RovingDiner/internal/cores"
	"github.com/yinweli/RovingDiner/internal/infra"
	"github.com/yinweli/RovingDiner/internal/rules"
	sheeter "github.com/yinweli/RovingDiner/sheet"
)

// Run 營業執行對外入口（【營業實作規格書 | 四、解耦的關鍵：邊界介面 | 核心怎麼跑】）:
// 以本場身分（seed / 關卡編號）與靜態表建營業實例——infra.NewRander(seed) 為唯一亂數來源、
// compiler 閉包注入效果預編譯（Parse → 捕捉 AST → execute 派發 + Settle 結算尾,即【營業規格書 | 二十、獨立流程 | 執行命令】）、
// rules.Register 裝備詞彙——自營業開始階段驅動 RunPhase 迴圈至停機,回報營業成功與否（踏入終止站時記錄）。
// Presenter 事件流於 M17 接入。
func Run(seed int64, stageID int32, sheet *sheeter.Sheeter, operator cores.Operator) bool {
	compiler := func(source string) (command cores.EffectExec, err error) {
		parsed, err := Parse(source)

		if err != nil {
			return nil, err
		} // if

		return func(game *cores.Game) {
			execute(game, parsed)
			rules.Settle(game) // 執行命令結算尾:結算旗標 == false → 執行結算
		}, nil
	}

	game := cores.NewGame(seed, stageID, cores.NewData(sheet, compiler), operator, infra.NewRander(seed))
	rules.Register(game)
	succ := false

	for phase := cores.PhaseGameStart; phase != cores.PhaseNone; {
		if phase == cores.PhaseGameSucc {
			succ = true // 兩終止站皆回 PhaseNone,踏站時記錄成敗
		} // if

		phase = rules.RunPhase(game, phase)
	} // for

	return succ
}
