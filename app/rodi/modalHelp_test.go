package rodi

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/stretchr/testify/suite"
)

func TestSuiteModalHelp(t *testing.T) {
	suite.Run(t, new(SuiteModalHelp))
}

// SuiteModalHelp 驗證格式說明 modal(modalHelp.go): 全靜態內容——標題、章序、行寬預算。
type SuiteModalHelp struct {
	suite.Suite
}

// TestModalHelpBody 驗證內容行: 標題、頂部通則段貼標題下、章序 = Tab 循環序(座位 → 牌堆 → 事件日誌)、
// 全行寬在 modal 內容寬上界內(靜態字面不靠 wrap)、各章範例與圖例行抽查。
func (this *SuiteModalHelp) TestModalHelpBody() {
	title, row := modalHelp{}.Body(nil)
	this.Equal("格式說明", title)
	this.Equal("列首「<」: 左端尚有未顯示項目", row[0]) // 頂部通則段直接貼標題下, 不立章
	group := []string{}

	for _, itor := range row {
		if strings.HasPrefix(itor, modalGroup) {
			group = append(group, strings.TrimPrefix(itor, modalGroup))
		} // if

		this.LessOrEqual(lipgloss.Width(itor), modalWidthMax-4, itor) // 行寬預算: 含框後不超 80 欄上界
	} // for

	this.Equal([]string{"座位", "行動佇列", "效果佇列", "手牌", "牌堆", "事件日誌"}, group)
	this.Contains(row, "10001@學生 飽3耐5 <- 資料編號@顧客名稱")
	this.Contains(row, "牌堆名稱(12): 1001@招呼 1002@上菜 <- (12)=張數")
	this.Contains(row, "行首「[」: 事件標題, 其下各行都屬於這個事件")
	this.Contains(row, "$ 1001@招呼#9 >> 棄牌堆 <- 搬移: 對象 >> 去向")
}
