// 營業 TUI 瘦進入點（【營業實作規格書 | 二、套件結構】）：flag 解析 → 載表 → 組裝後呼叫 app。
package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/yinweli/RovingDiner/app/tui"
	"github.com/yinweli/RovingDiner/internal/infra"
)

// main 組裝營業 TUI：-seed 預設 0 = 以當下時間取亂（實際 seed 由 dump 首行印出供重現），
// 明給非零值即決定性同一局；-stage 查無關卡空盤面照走（M15 寬鬆策略）；-data 指向 Sheeter 生成的資料目錄。
func main() {
	seed := flag.Int64("seed", 0, "亂數種子;0 = 以時間取亂")
	stage := flag.Int("stage", 0, "關卡編號")
	data := flag.String("data", "sheetdata", "靜態表資料目錄")
	flag.Parse()

	if *seed == 0 {
		*seed = time.Now().UnixNano()
	} // if

	sheet, err := infra.Load(*data)

	if err != nil {
		fmt.Fprintln(os.Stderr, "載表失敗:", err)
		os.Exit(1)
	} // if

	if err := tui.Run(*seed, int32(*stage), *data, sheet); err != nil {
		fmt.Fprintln(os.Stderr, "TUI 執行失敗:", err)
		os.Exit(1)
	} // if
}
