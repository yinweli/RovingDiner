package games

import (
	"github.com/yinweli/RovingDiner/internal/cores"
	"github.com/yinweli/RovingDiner/internal/infra"
	"github.com/yinweli/RovingDiner/internal/rules"
	sheeter "github.com/yinweli/RovingDiner/sheet"
)

// Build 組裝營業實例(【營業實作規格書 | 四、解耦的關鍵：邊界介面 | 核心怎麼跑】):
// 以本場身分(seed / 關卡編號)與靜態表建 Game——infra.NewRander(seed) 為唯一亂數來源、
// compiler 閉包注入效果預編譯(Parse → 捕捉 AST → execute 派發 + Settle 結算尾, 即【營業規格書 | 二十、獨立流程 | 執行命令】)、
// rules.Register 裝備詞彙。presenter 為事件流輸出 port(nil 由 NewGame 正規化為無輸出替身)。
// 建後盤面空白且引擎靜止, 開局建置由 Loop 的營業開始階段執行; 建造與驅動分離,
// 供消費端先取得 *cores.Game 再決定驅動節奏(同步直跑 Run / 顯示層停點直讀)。
func Build(seed int64, stageID int32, sheet *sheeter.Sheeter, operator cores.Operator, presenter cores.Presenter) *cores.Game {
	compiler := func(source string) (command cores.EffectExec, err error) {
		parsed, err := Parse(source)

		if err != nil {
			return nil, err
		} // if

		return func(game *cores.Game) {
			execute(game, parsed)
			rules.Settle(game) // 執行命令結算尾: 結算旗標 == false → 執行結算
		}, nil
	}

	game := cores.NewGame(seed, stageID, cores.NewData(sheet, compiler), operator, infra.NewRander(seed), presenter)
	rules.Register(game)
	return game
}

// Loop 驅動營業到停機: 自營業開始階段驅動 RunPhase 迴圈, 回報營業成功與否(踏入終止站時記錄)。
// 引擎逐單位同步 Emit(發射點對照詳見【營業實作規格書 | 四、解耦的關鍵：邊界介面 | 四之二】)、
// 需要玩家輸入時阻塞呼叫 operator。
func Loop(game *cores.Game) bool {
	succ := false

	for phase := cores.PhaseGameStart; phase != cores.PhaseNone; {
		if phase == cores.PhaseGameSucc {
			succ = true // 兩終止站皆回 PhaseNone, 踏站時記錄成敗
		} // if

		phase = rules.RunPhase(game, phase)
	} // for

	return succ
}

// Run 營業執行對外入口: Build + Loop 一站到底(同步直跑; 測試與 golden 比對用)。
func Run(seed int64, stageID int32, sheet *sheeter.Sheeter, operator cores.Operator, presenter cores.Presenter) bool {
	return Loop(Build(seed, stageID, sheet, operator, presenter))
}
