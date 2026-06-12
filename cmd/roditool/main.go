// 企劃驗證器瘦進入點(【營業實作規格書 | 附錄：企劃驗證器】): 單筆檢查(expr / command / threshold 子命令)
// 與表單檢查(sheet 子命令)兩種輸入模式, 共用 app/roditool 檢查引擎; CLI 框架統一用 cobra。
package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/yinweli/RovingDiner/app/roditool"
	"github.com/yinweli/RovingDiner/internal/infra"
)

// main 組裝企劃驗證器 CLI: 子命令分派單筆 / 表單檢查; 通過印 通過、發現問題回非零結束碼(供 task / CI 把關)。
func main() {
	command := &cobra.Command{
		Use:          "roditool",
		Short:        "流浪食堂企劃驗證器(單筆 / 表單檢查)",
		SilenceUsage: true, // 檢查不通過屬資料問題, 不重印用法; 旗標 / 參數解析錯由 FlagErrorFunc 補回
	}
	command.AddCommand(fieldCommand("expr", "檢查運算式欄位(觸發條件 / 觸發次數)", roditool.CheckExpr))
	command.AddCommand(fieldCommand("command", "檢查命令欄位(立即 / 觸發 / 啟動 / 結束命令)", roditool.CheckCommand))
	command.AddCommand(fieldCommand("threshold", "檢查門檻配對欄位(門檻值^技能編號)", roditool.CheckThreshold))
	command.AddCommand(sheetCommand())
	command.SetFlagErrorFunc(func(c *cobra.Command, err error) error {
		return fmt.Errorf("%v\n%v", err, c.UsageString()) // 比照 cmd/rodi: 解析錯誤帶上用法再回報
	})

	if command.Execute() != nil {
		os.Exit(1) // 錯誤訊息已由 cobra 印出
	} // if
}

// fieldCommand 組裝單筆檢查子命令: 一個位置參數為欄位內容, 經 check 分派對應文法驗證; 通過印 通過。
func fieldCommand(use, short string, check func(source string) error) *cobra.Command {
	return &cobra.Command{
		Use:   use + " <內容>",
		Short: short,
		Args:  cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, arg []string) error {
			if err := check(arg[0]); err != nil {
				return err
			} // if

			fmt.Println("通過")
			return nil
		},
	}
}

// sheetCommand 組裝表單檢查子命令: 載入 --data 目錄、CheckSheet 逐表逐欄檢查, 逐筆列印 表 / 列 / 欄 / 錯誤;
// 發現問題回錯誤(非零結束碼)。
func sheetCommand() *cobra.Command {
	var data string

	command := &cobra.Command{
		Use:   "sheet",
		Short: "表單檢查(掃描靜態表資料全表)",
		RunE: func(_ *cobra.Command, _ []string) error {
			sheet, err := infra.Load(data)

			if err != nil {
				return fmt.Errorf("載表失敗: %w", err)
			} // if

			issue := roditool.CheckSheet(sheet)

			for _, itor := range issue {
				fmt.Println(itor.String())
			} // for

			if len(issue) > 0 {
				return fmt.Errorf("表單檢查發現 %v 筆問題", len(issue))
			} // if

			fmt.Println("表單檢查通過")
			return nil
		},
	}
	command.Flags().StringVar(&data, "data", "sheetdata", "靜態表資料目錄")
	return command
}
