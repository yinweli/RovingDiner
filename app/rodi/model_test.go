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

// TestNewModel 驗證建構: 欄位就位、日誌 / 組件 / 寬高預算就緒、初始模式快速、起始聚焦座位(0)。
func (this *SuiteModel) TestNewModel() {
	target := newModel(nil)
	this.Nil(target.stepper)
	this.NotNil(target.log)
	this.Len(target.keybar.bind[keyModeNormal], 15)
	this.Empty(target.modal)
	this.Nil(target.wait)
	this.Len(target.comp, 6) // 六區(座位 / 場外 / 行動 / 效果 / 手牌 / 牌堆); 狀態列 / 日誌 / 鍵位列為 layout 角色專屬掛點
	this.Equal(modeFast, target.mode)
	this.Equal(0, target.gen)
	this.Equal(0, target.focus)
	this.Equal(minWidth, target.width)
	this.Equal(minHeight, target.height)
}

// TestModelInit 驗證起跑 Cmd 存在(依初始模式排第一拍)。
func (this *SuiteModel) TestModelInit() {
	this.NotNil(newModel(nil).Init())
}

// TestModelUpdate 驗證訊息分派: timer 拍驗章後推進並排下一拍(盤面直讀引擎、無摺疊), 過期世代丟棄斷鏈,
// 終局停止排拍; 步進拍只在步進模式推一拍且不排拍; 模式循環換模式 + 世代 +1; 切區循環移動(雙向迴繞);
// 視窗尺寸更新寬高預算; 按鍵查當前鍵盤模式綁定表分派(tab / space / n / q / ctrl+c、未綁定鍵不動作)。
func (this *SuiteModel) TestModelUpdate() {
	result, cmd := tea.Model(newModel(newStepper(1, 601, tester.BuildSheet()))).Update(tickMsg{})
	this.Equal(cores.PhaseGameStart, result.(model).stepper.game.GetPhase()) // 首拍: 引擎停在首事件邊界, 直讀即見
	this.NotNil(cmd)                                                         // 已排下一拍

	for { // 逐拍推進到終局: 玩家行動等待以 [E] 結束答覆並恢復排拍, 終局停止排拍
		if cmd != nil {
			result, cmd = result.(model).Update(tickMsg{})
			continue
		} // if

		if result.(model).wait == nil {
			break // 停止排拍且無等待 = 終局
		} // if

		result, cmd = result.(model).Update(endMsg{})
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

	target = newModel(newStepper(1, 601, tester.BuildSheet()))
	result, cmd = tea.Model(target).Update(tickMsg{})

	for result.(model).wait == nil { // 推進到第一個玩家行動等待
		result, cmd = result.(model).Update(tickMsg{})
	} // for

	this.Nil(cmd)                                                               // 等待輸入: 不排拍
	this.Equal(cores.PhasePlayerAction, result.(model).stepper.game.GetPhase()) // 引擎停在 PlayerAction 暫停點

	hold := result.(model)
	hold.mode = modeStep
	result, cmd = hold.Update(stepMsg{}) // 等待中 [N] 防呆: 不放行不推進(引擎停在 Operator)
	this.NotNil(result.(model).wait)
	this.Nil(cmd)

	hold = result.(model)
	hold.mode = modeFast
	hold.stepper.game.Hand[0].GetSeal().Lock() // 游標卡封印 → [P] no-op(M27 拍板; 暗色已提示)
	result, cmd = hold.Update(playMsg{})
	this.NotNil(result.(model).wait)
	this.Nil(cmd)

	hold = result.(model)
	hold.stepper.game.Hand[0].GetSeal().Unlock() // 解封 → [P] 出游標卡: 答覆會合、行組續收、恢復排拍
	result, cmd = hold.Update(playMsg{})
	this.Nil(result.(model).wait)
	this.NotNil(cmd)

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

	result, cmd = tea.Model(newModel(nil)).Update(tabMsg{delta: 1}) // 正向切區: 座位 → 場外
	this.Equal(1, result.(model).focus)
	this.Nil(cmd)

	result, _ = result.(model).Update(tabMsg{delta: -1}) // 反向切回
	this.Equal(0, result.(model).focus)

	result, _ = result.(model).Update(tabMsg{delta: -1}) // 反向自 0 迴繞到事件日誌(末區)
	this.Equal(focusLog, result.(model).focus)

	result, _ = result.(model).Update(tabMsg{delta: 1}) // 正向自末區迴繞回座位
	this.Equal(0, result.(model).focus)

	target = newModel(newStepper(1, 601, tester.BuildSheet()))
	target.focus = 1 // 聚焦場外: 方向鍵分派給組件自持游標
	result, cmd = target.Update(moveMsg{key: "down"})
	this.Equal(1, result.(model).comp[1].(*panelPool).curRow)
	this.Nil(cmd)

	target.focus = focusStatus // 狀態列無游標: 落空不動作
	_, cmd = target.Update(moveMsg{key: "down"})
	this.Nil(cmd)

	target.focus = focusLog // 聚焦日誌: 分派 viewport 捲動
	target.log.Append("a", "b")
	_, _ = target.Update(moveMsg{key: "up"})
	this.Equal(1, target.log.offset)

	result, cmd = tea.Model(target).Update(countMsg{}) // F2 開計數 modal: 入棧 + 世代 +1(在途拍作廢、暫停消費)
	this.Len(result.(model).modal, 1)
	this.Equal(1, result.(model).gen)
	this.Equal(keyModeModal, result.(model).keymode())
	this.Nil(cmd)

	result, cmd = result.(model).Update(moveMsg{key: "down"}) // modal 態: 方向鍵轉頂層捲動(計數內容 22 行 > 可視 20, 上限 2)
	this.Equal(1, result.(model).modalOff)
	this.Nil(cmd)

	result, _ = result.(model).Update(moveMsg{key: "down"})
	result, _ = result.(model).Update(moveMsg{key: "down"}) // 過衝夾住上限
	this.Equal(2, result.(model).modalOff)

	result, _ = result.(model).Update(moveMsg{key: "up"})
	result, _ = result.(model).Update(moveMsg{key: "up"})
	result, _ = result.(model).Update(moveMsg{key: "up"}) // 過衝夾回 0
	this.Equal(0, result.(model).modalOff)

	result, cmd = result.(model).Update(popMsg{}) // Esc 關閉: 出棧 + 依當前模式(快速)恢復排拍
	this.Empty(result.(model).modal)
	this.Equal(keyModeNormal, result.(model).keymode())
	this.NotNil(cmd)

	result, cmd = result.(model).Update(popMsg{}) // 空棧防呆, 不動作也不重排拍以免雙鏈
	this.Empty(result.(model).modal)
	this.Nil(cmd)

	result, cmd = result.(model).Update(helpMsg{}) // F1 開格式說明 modal: 同入棧路徑(開著時暫停消費)
	this.Equal(modalHelp{}, result.(model).modal[0])
	this.Equal(keyModeModal, result.(model).keymode())
	this.Nil(cmd)

	target = newModel(newStepper(1, 601, tester.BuildSheet()))
	target.modal = []modal{fakeModal{row: []string{"短內容"}}} // 內容未溢出: 捲動上限 0、down 不動
	result, _ = target.Update(moveMsg{key: "down"})
	this.Equal(0, result.(model).modalOff)

	target = newModel(newStepper(1, 601, tester.BuildSheet()))
	target.focus = focusStatus
	result, cmd = target.Update(enterMsg{}) // 狀態列聚焦 Enter = 開計數 modal
	this.Len(result.(model).modal, 1)
	this.Nil(cmd)

	target.focus = 0
	result, _ = target.Update(enterMsg{}) // 聚焦座位但全空位: 游標下無項目, 不動作
	this.Empty(result.(model).modal)

	game := target.stepper.game // 擺盤後對游標項目辨型開檢視 modal(R4)
	game.Wait.Insert(cores.NewGuest(game, 1))
	game.Action.Push(cores.NewAction(game.Wait[0], cores.TaskSate, 1))
	game.Effect.Push(cores.NewEffect(game, 1, cores.Ref{}, 1))
	game.Hand.Push(cores.NewCard(game, 1))

	for focus, want := range map[int]modal{
		1: modalGuest{guest: game.Wait[0]},
		2: modalAction{action: game.Action[0]},
		3: modalEffect{effect: game.Effect[0]},
		4: modalCard{card: game.Hand[0]},
	} {
		target.focus = focus
		result, cmd = target.Update(enterMsg{})
		this.Equal(want, result.(model).modal[0])
		this.Nil(cmd)
		target.modal = nil // 收回(value model: target 未持堆疊, 防呆歸位)
	} // for

	target.focus = focusLog // 事件日誌無 modal
	result, _ = target.Update(enterMsg{})
	this.Empty(result.(model).modal)

	pick := &request{guest: []*cores.Guest{{}, {}}, count: 1, answer: make(chan answer, 1)}
	fab := newModel(&stepper{turn: make(chan turn, 2)}) // 偽 stepper 預填輪次: 驗答覆後續收的 Pick 被動立答
	fab.wait = &request{answer: make(chan answer, 1)}
	fab.stepper.turn <- turn{role: turnRequest, req: pick}
	fab.stepper.turn <- turn{role: turnLine, line: []string{"x"}}
	result, cmd = fab.Update(endMsg{})
	this.Nil(result.(model).wait)
	this.NotNil(cmd)                                         // 行組後依快速恢復排拍
	this.Equal(answer{guest: pick.guest[:1]}, <-pick.answer) // Pick 請求已被動立答(R4 換真選取前的過渡)
	this.NotEmpty(result.(model).log.line)

	fab = newModel(&stepper{turn: make(chan turn, 1)})
	fab.wait = &request{answer: make(chan answer, 1)}
	fab.stepper.turn <- turn{role: turnOver, succ: true}
	result, cmd = fab.Update(endMsg{}) // 答覆後直接終局: 停止排拍
	this.Nil(result.(model).wait)
	this.Nil(cmd)

	fab = newModel(newStepper(1, 601, tester.BuildSheet())) // 未推進: 空白盤面空手牌
	fab.wait = &request{answer: make(chan answer, 1)}
	_, cmd = fab.Update(playMsg{}) // 游標下無卡: no-op
	this.Nil(cmd)

	_, cmd = newModel(nil).Update(playMsg{}) // 非等待輸入: [P]/[E] 不動作
	this.Nil(cmd)

	_, cmd = newModel(nil).Update(endMsg{})
	this.Nil(cmd)

	result, cmd = newModel(nil).Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	this.Equal(120, result.(model).width)
	this.Equal(40, result.(model).height)
	this.Nil(cmd)

	_, cmd = newModel(nil).Update(tea.KeyMsg{Type: tea.KeyTab})
	this.NotNil(cmd)
	this.Equal(tabMsg{delta: 1}, cmd())

	_, cmd = newModel(nil).Update(tea.KeyMsg{Type: tea.KeyShiftTab})
	this.NotNil(cmd)
	this.Equal(tabMsg{delta: -1}, cmd())

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

// TestModelView 驗證全畫面 layout: 共 height 行、左欄六區與日誌同列起頭、狀態列釘底、底框列、鍵位列收尾、
// 日誌寬 flex(餘寬先給日誌到上限); 低於最低尺寸守門顯提示。盤面直讀暫停機(未推進 = 開局前空白盤面,
// 守門與排版不受內容影響)。
func (this *SuiteModel) TestModelView() {
	target := newModel(newStepper(1, 601, tester.BuildSheet()))
	row := strings.Split(target.View(), "\n")
	this.Len(row, minHeight) // 預設 100x30 → 滿版 30 行
	this.Contains(row[0], "+- 座位")
	this.Contains(row[0], "- 事件日誌")                                                           // 右欄與左欄同列起頭(接縫版標題無左端 +)
	this.Contains(row[minHeight-4], "快速")                                                     // 狀態列釘底(底框列上方)
	this.Equal("+"+strings.Repeat("-", 68)+"+"+strings.Repeat("-", 29)+"+", row[minHeight-3]) // 父層全寬底框列(左欄 70 + 日誌 30)
	this.Contains(row[minHeight-1], "[Q]離開")                                                  // 鍵位列固定 2 行

	target.height = 40 // 更高: 帶框空行墊在六區與狀態列之間(格線不開洞), 狀態列仍釘底
	row = strings.Split(target.View(), "\n")
	this.Len(row, 40)
	this.Contains(row[36], "快速")
	this.Equal(boxRow("", 70)+boxRowSeam("", 30), row[30]) // gap 帶框空行 + 同列的日誌接縫空行

	target.width = 130 // 更寬: 餘寬先給日誌(130-70 = 60 → 上限 50), 之後才給左欄(80)
	row = strings.Split(target.View(), "\n")
	this.Equal("+"+strings.Repeat("-", 78)+"+"+strings.Repeat("-", 49)+"+", row[37])
	target.width = minWidth

	target.focus = focusStatus // 聚焦狀態列 / 日誌: 高亮純上色(無 TTY 渲染原文), 版面與內容不變
	row = strings.Split(target.View(), "\n")
	this.Contains(row[34], "+- 狀態列")

	target.focus = focusLog
	row = strings.Split(target.View(), "\n")
	this.Contains(row[0], "- 事件日誌")
	target.focus = 0

	target.modal = []modal{modalCount{seed: 42}} // modal 開著: 置中疊在全畫面上、總行數不變
	row = strings.Split(target.View(), "\n")
	this.Len(row, 40)
	this.Contains(target.View(), "計數檢視 (seed 42)")
	target.modal = nil

	this.Len(strings.Split(target.leftView(70, 5), "\n"), 27) // 內容超高(防禦): 留白歸零、不裁內容

	target.width = 80 // 低於最低尺寸 → 守門提示
	this.Equal("請放大終端機 (現 80x40, 最低 100x30)", target.View())

	target.width = 100
	target.height = 20
	this.Equal("請放大終端機 (現 100x20, 最低 100x30)", target.View())
}

// TestModelTick 驗證排拍 Cmd: 快/慢回 timer Cmd 且訊息蓋上當前世代(供驗章), 步進不排拍回 nil,
// modal 開著不排拍回 nil(暫停消費)。
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

	target.mode = modeFast
	target.modal = []modal{modalCount{}}
	this.Nil(target.tick())

	target.modal = nil
	target.wait = &request{}
	this.Nil(target.tick()) // 等待輸入不排拍(引擎停在 Operator)
}

// TestModelKeymode 驗證鍵盤模式導出: modal 堆疊非空 = modal 態、否則常態(選取等待 M27 補來源)。
func (this *SuiteModel) TestModelKeymode() {
	target := newModel(nil)
	this.Equal(keyModeNormal, target.keymode())

	target.modal = []modal{modalCount{}}
	this.Equal(keyModeModal, target.keymode())
}

// TestStep 驗證步進訊息 Cmd: 只回推進訊息(不碰引擎)。
func (this *SuiteModel) TestStep() {
	this.Equal(stepMsg{}, step()())
}

// TestCycle 驗證模式循環訊息 Cmd: 只回循環訊息(模式切換活在 Update)。
func (this *SuiteModel) TestCycle() {
	this.Equal(cycleMsg{}, cycle()())
}

// TestTab 驗證切區訊息 Cmd: 載移動方向(聚焦移動活在 Update)。
func (this *SuiteModel) TestTab() {
	this.Equal(tabMsg{delta: 1}, tab(1)())
	this.Equal(tabMsg{delta: -1}, tab(-1)())
}

// TestMove 驗證游標移動訊息 Cmd: 載方向鍵名(分派活在 Update)。
func (this *SuiteModel) TestMove() {
	this.Equal(moveMsg{key: "up"}, move("up")())
	this.Equal(moveMsg{key: "left"}, move("left")())
}

// TestHelp 驗證開格式說明 modal 訊息 Cmd。
func (this *SuiteModel) TestHelp() {
	this.Equal(helpMsg{}, help()())
}

// TestCount 驗證開計數 modal 訊息 Cmd。
func (this *SuiteModel) TestCount() {
	this.Equal(countMsg{}, count()())
}

// TestEnter 驗證檢視訊息 Cmd。
func (this *SuiteModel) TestEnter() {
	this.Equal(enterMsg{}, enter()())
}

// TestPlay 驗證出牌訊息 Cmd。
func (this *SuiteModel) TestPlay() {
	this.Equal(playMsg{}, play()())
}

// TestEnd 驗證玩家結束訊息 Cmd。
func (this *SuiteModel) TestEnd() {
	this.Equal(endMsg{}, end()())
}

// TestPop 驗證關閉 modal 訊息 Cmd。
func (this *SuiteModel) TestPop() {
	this.Equal(popMsg{}, pop()())
}
