// 營業 TUI 瘦進入點(【營業實作規格書 | 二、套件結構】): flag 解析 → 載表 → 組裝後呼叫 app; CLI 框架統一用 cobra。
package main

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/yinweli/RovingDiner/app/rodi"
	"github.com/yinweli/RovingDiner/internal/infra"
)

// main 組裝營業 TUI: --seed 預設 0 = 以當下時間取亂(實際 seed 於進 alt-screen 前印進 scrollback,
// 退出後仍可見、供重現), 明給非零值即決定性同一局; --stage 查無關卡空盤面照走(M15 寬鬆策略);
// --data 指向 Sheeter 生成的資料目錄。
func main() {
	var seed int64
	var stage int32
	var data string

	command := &cobra.Command{
		Use:          "rodi",
		Short:        "流浪食堂營業 TUI(debug viewer)",
		SilenceUsage: true, // 執行期錯誤(載表失敗等)不重印用法, 只在旗標解析錯時顯示
		RunE: func(_ *cobra.Command, _ []string) error {
			if seed == 0 {
				seed = time.Now().UnixNano()
			} // if

			sheet, err := infra.Load(data)

			if err != nil {
				return fmt.Errorf("載表失敗: %w", err)
			} // if

			fmt.Printf("營業開始 (seed %v, stage %v, data %v)\n", seed, stage, data) // 進 alt-screen 前印, 留 scrollback
			return rodi.Run(seed, stage, sheet)
		},
	}
	command.Flags().Int64Var(&seed, "seed", 0, "亂數種子; 0 = 以時間取亂")
	command.Flags().Int32Var(&stage, "stage", 0, "關卡編號")
	command.Flags().StringVar(&data, "data", "sheetdata", "靜態表資料目錄")
	command.SetFlagErrorFunc(func(c *cobra.Command, err error) error {
		return fmt.Errorf("%v\n%v", err, c.UsageString()) // SilenceUsage 連旗標解析錯也吞用法, 這裡補回: 解析錯誤帶上用法再回報
	})

	if command.Execute() != nil {
		os.Exit(1) // 錯誤訊息已由 cobra 印出
	} // if
}
