package rodi

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/suite"

	"github.com/yinweli/RovingDiner/internal/cores"
	"github.com/yinweli/RovingDiner/internal/tester"
)

func TestSuiteModel(t *testing.T) {
	suite.Run(t, new(SuiteModel))
}

// SuiteModel 驗證 Bubble Tea 殼(model.go): Update 純函式分派、全畫面 layout、最低尺寸守門、排拍推進鏈(M25)。
// Run 為組裝入口(需要 TTY), 比照 cmd 瘦組裝不納白箱測試。
type SuiteModel struct {
	suite.Suite
}

// TestNewModel 驗證建構: 欄位就位、日誌 / 組件 / 寬高預算就緒、初始模式快速。
func (this *SuiteModel) TestNewModel() {
	target := newModel(nil)
	this.Nil(target.stepper)
	this.NotNil(target.log)
	this.Len(target.keybar.bind, 4)
	this.Len(target.comp, 6) // 六區(座位 / 場外 / 行動 / 效果 / 手牌 / 牌堆); 狀態列 / 日誌 / 鍵位列為 layout 角色專屬掛點
	this.Equal(modeFast, target.mode)
	this.Equal(0, target.gen)
	this.Equal(minWidth, target.width)
	this.Equal(minHeight, target.height)
}

// TestModelInit 驗證起跑 Cmd 存在(依初始模式排第一拍)。
func (this *SuiteModel) TestModelInit() {
	this.NotNil(newModel(nil).Init())
}

// TestModelUpdate 驗證訊息分派: timer 拍驗章後推進並排下一拍(盤面直讀引擎、無摺疊), 過期世代丟棄斷鏈,
// 終局停止排拍; 步進拍只在步進模式推一拍且不排拍; 模式循環換模式 + 世代 +1;
// 視窗尺寸更新寬高預算; 按鍵查綁定表分派(space / n / q / ctrl+c、未綁定鍵不動作)。
func (this *SuiteModel) TestModelUpdate() {
	result, cmd := tea.Model(newModel(newStepper(1, 601, tester.BuildSheet()))).Update(tickMsg{})
	this.Equal(cores.PhaseGameStart, result.(model).stepper.game.GetPhase()) // 首拍: 引擎停在首事件邊界, 直讀即見
	this.NotNil(cmd)                                                         // 已排下一拍

	for cmd != nil { // 逐拍推進到終局: 停止排拍
		result, cmd = result.(model).Update(tickMsg{})
	} // for

	this.NotEmpty(result.(model).log.line)                          // 行組已入日誌
	this.NotZero(result.(model).stepper.game.GetRound().GetValue()) // 引擎已推進整場

	target := newModel(newStepper(1, 601, tester.BuildSheet()))
	target.gen = 1
	phase := target.stepper.game.GetPhase()
	result, cmd = target.Update(tickMsg{}) // 過期世代(章 0 != 世代 1): 切模式前的在途舊拍, 不推進不排拍
	this.Equal(phase, result.(model).stepper.game.GetPhase())
	this.Nil(cmd)

	target = newModel(newStepper(1, 601, tester.BuildSheet()))
	phase = target.stepper.game.GetPhase()
	result, cmd = target.Update(stepMsg{}) // 自動模式: [N] 不插拍
	this.Equal(phase, result.(model).stepper.game.GetPhase())
	this.Nil(cmd)

	target.mode = modeStep
	result, cmd = target.Update(stepMsg{}) // 步進模式: [N] 推一拍、不排拍(等下一個 [N])
	this.Equal(cores.PhaseGameStart, result.(model).stepper.game.GetPhase())
	this.Nil(cmd)

	result, cmd = tea.Model(newModel(nil)).Update(cycleMsg{}) // 快速 → 慢速: 世代 +1、重排拍
	this.Equal(modeSlow, result.(model).mode)
	this.Equal(1, result.(model).gen)
	this.NotNil(cmd)

	result, cmd = result.(model).Update(cycleMsg{}) // 慢速 → 步進: 暫停態不排拍
	this.Equal(modeStep, result.(model).mode)
	this.Equal(2, result.(model).gen)
	this.Nil(cmd)

	result, cmd = result.(model).Update(cycleMsg{}) // 步進 → 快速: 重排拍
	this.Equal(modeFast, result.(model).mode)
	this.Equal(3, result.(model).gen)
	this.NotNil(cmd)

	result, cmd = newModel(nil).Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	this.Equal(120, result.(model).width)
	this.Equal(40, result.(model).height)
	this.Nil(cmd)

	_, cmd = newModel(nil).Update(tea.KeyMsg{Type: tea.KeySpace})
	this.NotNil(cmd)
	this.Equal(cycleMsg{}, cmd())

	_, cmd = newModel(nil).Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("n")})
	this.NotNil(cmd)
	this.Equal(stepMsg{}, cmd())

	_, cmd = newModel(nil).Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("q")})
	this.NotNil(cmd)
	this.Equal(tea.QuitMsg{}, cmd())

	_, cmd = newModel(nil).Update(tea.KeyMsg{Type: tea.KeyCtrlC})
	this.NotNil(cmd)
	this.Equal(tea.QuitMsg{}, cmd())

	_, cmd = newModel(nil).Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("x")})
	this.Nil(cmd) // 未綁定鍵不動作

	_, cmd = newModel(nil).Update(struct{}{}) // 未知訊息型別不動作
	this.Nil(cmd)
}

// TestModelView 驗證全畫面 layout: 共 height 行、左欄六區與日誌同列起頭、狀態列釘底、鍵位列收尾;
// 低於最低尺寸守門顯提示。盤面直讀暫停機(未推進 = 開局前空白盤面, 守門與排版不受內容影響)。
func (this *SuiteModel) TestModelView() {
	target := newModel(newStepper(1, 601, tester.BuildSheet()))
	row := strings.Split(target.View(), "\n")
	this.Len(row, minHeight) // 預設 100x30 → 滿版 30 行
	this.Contains(row[0], "+- 座位")
	this.Contains(row[0], "+- 事件日誌")         // 右欄與左欄同列起頭
	this.Contains(row[minHeight-3], "快速")    // 狀態列釘底(鍵位列上方)
	this.Contains(row[minHeight-1], "[Q]離開") // 鍵位列固定 2 行, 現有鍵全在第 2 行(第 1 行留白)

	target.height = 40 // 更高: 留白墊在六區與狀態列之間, 狀態列仍釘底
	row = strings.Split(target.View(), "\n")
	this.Len(row, 40)
	this.Contains(row[37], "快速")

	this.Len(strings.Split(target.leftView(70, 5), "\n"), 26) // 內容超高(防禦): 留白歸零、不裁內容

	target.width = 80 // 低於最低尺寸 → 守門提示
	this.Equal("請放大終端機 (現 80x40, 最低 100x30)", target.View())

	target.width = 100
	target.height = 20
	this.Equal("請放大終端機 (現 100x20, 最低 100x30)", target.View())
}

// TestModelTick 驗證排拍 Cmd: 快/慢回 timer Cmd 且訊息蓋上當前世代(供驗章), 步進不排拍回 nil。
func (this *SuiteModel) TestModelTick() {
	target := newModel(nil)
	target.gen = 7
	cmd := target.tick()
	this.Require().NotNil(cmd)
	this.Equal(tickMsg{gen: 7}, cmd()) // 快速 timer: 0.2 秒後投遞當前世代章

	target.mode = modeSlow
	this.NotNil(target.tick())

	target.mode = modeStep
	this.Nil(target.tick())
}

// TestStep 驗證步進訊息 Cmd: 只回推進訊息(不碰引擎)。
func (this *SuiteModel) TestStep() {
	this.Equal(stepMsg{}, step()())
}

// TestCycle 驗證模式循環訊息 Cmd: 只回循環訊息(模式切換活在 Update)。
func (this *SuiteModel) TestCycle() {
	this.Equal(cycleMsg{}, cycle()())
}
