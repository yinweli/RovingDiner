package rodi

import (
	"fmt"
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/stretchr/testify/suite"

	"github.com/yinweli/RovingDiner/internal/cores"
)

func TestSuiteModal(t *testing.T) {
	suite.Run(t, new(SuiteModal))
}

// SuiteModal 驗證 modal 框架(modal.go): 盒渲染(content-fit 夾上界 / 分隔線代換 / 捲動與位置指示)、
// 底框列、置中疊層。
type SuiteModal struct {
	suite.Suite
}

// TestModalView 驗證盒渲染: 寬 = 內容 content-fit、分隔線前綴行代換成具名分隔線、未溢出素底框;
// 溢出依 offset 開窗 + 底框位置指示(在頂 0% / 中途百分比 / 到底 END)。
func (this *SuiteModal) TestModalView() {
	game := testGame()
	this.Equal([]string{
		"+- 標題 -+",
		"| 內容一 |",
		"+- 群組 -+",
		"| 內容二 |",
		"+--------+",
	}, modalView(game, fakeModal{title: "標題", row: []string{"內容一", modalGroup + "群組", "內容二"}}, 0))

	long := fakeModal{title: "t"}

	for i := 0; i < 25; i++ { // 25 行 > 可視 20: 捲動上限 5
		long.row = append(long.row, fmt.Sprintf("row%02d", i))
	} // for

	box := modalView(game, long, 0)
	this.Require().Len(box, modalHeightMax)
	this.Equal("| row00 |", box[1])
	this.Equal("| row19 |", box[20])
	this.Contains(box[21], " 0% ") // 在頂

	box = modalView(game, long, 2)
	this.Equal("| row02 |", box[1])
	this.Contains(box[21], " 40% ") // 中途: 2/5

	box = modalView(game, long, 5)
	this.Equal("| row24 |", box[20])
	this.Contains(box[21], " END ") // 到底

	wide := modalView(game, fakeModal{title: "t", row: []string{strings.Repeat("x", 90)}}, 0) // 內容超寬: 夾上界 80、行內截斷
	this.Equal(modalWidthMax, lipgloss.Width(wide[0]))
	this.Equal(modalWidthMax, lipgloss.Width(wide[1]))
}

// TestWrapToken 驗證 token 換行: 未超寬單行、超寬貪婪打包成多列、空列表回單一空行。
func (this *SuiteModal) TestWrapToken() {
	this.Equal([]string{"aa  bb"}, wrapToken([]string{"aa", "bb"}))
	this.Equal([]string{""}, wrapToken(nil)) // 空列表: 區塊恆顯、列留空

	token := []string{}

	for i := 0; i < 10; i++ { // 10 個 10 格 token: 預算 76 一列容 6 個(6x10+5x2 = 70)
		token = append(token, strings.Repeat("x", 10))
	} // for

	row := wrapToken(token)
	this.Require().Len(row, 2)
	this.Equal(70, lipgloss.Width(row[0]))
	this.Equal(46, lipgloss.Width(row[1]))
}

// TestWrapText 驗證長字串換行: 未超寬原樣單行、超寬依顯示寬切段(全形不切半字)。
func (this *SuiteModal) TestWrapText() {
	this.Equal([]string{"短文"}, wrapText("短文"))

	row := wrapText(strings.Repeat("文", 50)) // 100 格 > 76: 切 38 字 + 12 字
	this.Require().Len(row, 2)
	this.Equal(76, lipgloss.Width(row[0]))
	this.Equal(24, lipgloss.Width(row[1]))
}

// TestModalBottom 驗證底框列: 無標示素線收尾; 有標示嵌右下(仿 less)。
func (this *SuiteModal) TestModalBottom() {
	this.Equal("+--------+", modalBottom(10, ""))
	this.Equal("+- 57% -+", modalBottom(9, "57%"))
	this.Equal("+-- END -+", modalBottom(10, "END"))
}

// TestOverlay 驗證置中疊層: 盒列疊在底圖正中、被遮列左右兩側保留; 盒高超出底圖(防禦)捨去超出列。
func (this *SuiteModal) TestOverlay() {
	base := strings.Join([]string{"aaaaaaaaaa", "bbbbbbbbbb", "cccccccccc", "dddddddddd", "eeeeeeeeee"}, "\n")
	this.Equal(strings.Join([]string{
		"aaaaaaaaaa",
		"bbb[XX]bbb",
		"ccc[YY]ccc",
		"dddddddddd",
		"eeeeeeeeee",
	}, "\n"), overlay(base, []string{"[XX]", "[YY]"}, 10, 5))

	this.Equal("[YY]", overlay("zzzz", []string{"[XX]", "[YY]", "[ZZ]"}, 4, 1)) // 防禦: 超出底圖的盒列捨去
}

// === 測試輔助(置尾) ===

// fakeModal 測試替身: 回固定標題與內容行, 供框架渲染驗證。
type fakeModal struct {
	title string
	row   []string
}

func (this fakeModal) Body(game *cores.Game) (title string, row []string) {
	return this.title, this.row
}
