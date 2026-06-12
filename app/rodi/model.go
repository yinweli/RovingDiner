package rodi

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/yinweli/RovingDiner/internal/cores"
	sheeter "github.com/yinweli/RovingDiner/sheet"
)

// 畫面尺寸常數(【營業顯示規格書 | 8、終端機尺寸與 flex】): 以 ~100 欄為主要設計寬度——
// 最低尺寸時左欄 70(座位 2 桌硬需求)/ 日誌 30; 餘寬先給日誌長到上限、之後才全給左欄
// (M26 拍板: 日誌不斷尾優先於多視一桌); 低於最低支援尺寸顯提示、不硬塞。
const (
	logWidthMin = 30  // 右欄事件日誌寬度地板(最低尺寸時)
	logWidthMax = 50  // 右欄事件日誌寬度上限(M26 拍板)
	minWidth    = 100 // 最低支援寬
	minHeight   = 30  // 最低支援高
)

// 聚焦區索引(【營業顯示規格書 | 7、互動規格 | 7.1】8 區循環; M26 拍板循環序 = 規格列序):
// 0..5 對應 comp 堆疊順序(座位 / 場外 / 行動 / 效果 / 手牌 / 牌堆), 狀態列與事件日誌為專屬掛點接末;
// 起始聚焦 = 座位(0)且恆有聚焦區(不設無聚焦態)。鍵位列非聚焦元件, 不入循環。
const (
	focusStatus = 6 // 狀態列
	focusLog    = 7 // 事件日誌
	focusCount  = 8 // 循環模數
)

// Run 顯示/操作層對外入口: 組裝暫停機(被動觀看 operator)與 Bubble Tea 殼(alt-screen; M22)後跑到離開。
// seed 由呼叫端先行定案(cmd 對 0 取時間亂數), 並由 cmd 在進 alt-screen 前把 seed / 關卡印進
// scrollback(退出後仍可見, 供重現與對帳; M22 拍板)。
func Run(seed int64, stageID int32, sheet *sheeter.Sheeter) error {
	_, err := tea.NewProgram(newModel(newStepper(seed, stageID, sheet)), tea.WithAltScreen()).Run()

	if err != nil {
		return fmt.Errorf("rodi: %w", err)
	} // if

	return nil
}

// model Bubble Tea 殼(M22 alt-screen 換裝, M19 dump 退役): 全畫面 layout——左欄六區堆疊 + 狀態列釘底、
// 右欄事件日誌與左欄同高、鍵位列橫跨底部全寬; 父層只組合與分配空間(M20 拍板)。
// 消費依模式排拍(M25, 【營業顯示規格書 | 3、日誌流的消費：速率與步進】): 快/慢 = timer Cmd 投遞 tickMsg,
// Update 同步 Next 推一拍(行組入日誌; 盤面組件直讀引擎、不持拷貝)再排下一拍; 步進 = 不排拍,
// 等 [N] 的 stepMsg 逐拍前進; [Space] 經 cycleMsg 循環三模式, 世代 +1 作廢在途舊拍(驗章斷鏈)。
// 終局即停止推進、等 q 離開(成敗常駐顯示活在狀態列階段欄)。
// Next 只在 Update 內呼叫(stepper 直讀安全窗的前提), 推進節奏全活在 Update 迴圈。
type model struct {
	stepper  *stepper    // 暫停機橋接器(盤面唯一真相 = stepper.game, 組件直讀、不持拷貝)
	log      *panelLog   // 事件日誌組件(右欄; 行歷史自持、簽章自立, 不入 comp; M22 拍板)
	keybar   barKey      // 鍵位列(按鍵分派入口 + 底部全寬列)
	status   barStatus   // 狀態列(左欄釘底)
	comp     []component // 左欄堆疊組件(六區; 順序 = 堆疊順序)
	mode     mode        // 執行模式(model 持有的 UI 狀態; [Space] 經 cycleMsg 切換, 排拍節奏據此)
	gen      int         // 排拍世代(切模式 / 開 modal +1; tickMsg 載章比對, 在途舊拍作廢——模式值當章不夠, 快→步→快 回同名模式會雙鏈)
	focus    int         // 聚焦區索引(model 持有的跨組件 UI 狀態; Tab 經 tabMsg 循環 8 區, 聚焦高亮據此)
	modal    []modal     // modal 堆疊(M26 R3; 開著時暫停消費、鍵盤入 modal 態, 頂層互動)
	modalOff int         // 頂層 modal 捲動偏移(0 = 在頂; 開 / 關歸零, Update 捲動時夾界)
	width    int         // 寬度預算(WindowSizeMsg 前用設計寬)
	height   int         // 高度預算(WindowSizeMsg 前用最低高)
}

func newModel(stepper *stepper) model {
	return model{
		stepper: stepper,
		log:     newPanelLog(),
		keybar:  newBarKey(),
		status:  barStatus{},
		comp:    []component{&panelSeat{}, &panelPool{}, &panelAction{}, &panelEffect{}, &panelHand{}, &panelPile{}},
		mode:    modeFast,
		width:   minWidth,
		height:  minHeight,
	}
}

// Init 起跑: 依初始模式(快速)排第一拍(header 已由 cmd 印於 scrollback, 殼不再印行)。
func (this model) Init() tea.Cmd {
	return this.tick()
}

// Update 訊息分派: timer 拍 → 驗章後推一拍再排下一拍(過期世代 = 切模式前的在途舊拍, 丟棄斷鏈);
// 步進拍 → 步進模式才推一拍、不排拍([N] 為步進專用, 自動模式下不插拍); 模式循環 → 換模式 + 世代 +1,
// 新模式為自動即重排拍; 終局停止推進(成敗活在狀態列階段欄直讀, 不另動作; 終局後切模式排的拍經 Next
// 防呆自然 no-op); 切區 → 聚焦索引循環移動(迴繞); 游標 → modal 態捲動頂層 modal、常態分派聚焦區 Move
// (語意隨區, 狀態列落空不動作); 說明 / 計數 / 檢視 → 推 modal 入棧(開著時暫停消費); 關閉 → 出棧並恢復排拍;
// 視窗尺寸 → 更新寬高預算; 按鍵 → 查當前鍵盤模式的綁定表分派(未綁定不動作)。
func (this model) Update(msg tea.Msg) (result tea.Model, cmd tea.Cmd) {
	switch msg := msg.(type) {
	case tickMsg:
		if msg.gen != this.gen {
			return this, nil
		} // if

		if this.advance() == false {
			return this, nil
		} // if

		return this, this.tick()

	case stepMsg:
		if this.mode != modeStep {
			return this, nil
		} // if

		this.advance()
		return this, nil

	case cycleMsg:
		this.mode = this.mode.next()
		this.gen++
		return this, this.tick()

	case tabMsg:
		this.focus = (this.focus + msg.delta + focusCount) % focusCount
		return this, nil

	case moveMsg:
		if len(this.modal) > 0 { // modal 態: 上下捲動頂層 modal(夾 [0, 內容行數 - 可視行數])
			if msg.key == keyUp {
				this.modalOff--
			} // if

			if msg.key == keyDown {
				this.modalOff++
			} // if

			_, row := this.modal[len(this.modal)-1].Body(this.stepper.game)
			limit := len(row) - (modalHeightMax - 2)

			if limit < 0 {
				limit = 0
			} // if

			if this.modalOff > limit {
				this.modalOff = limit
			} // if

			if this.modalOff < 0 {
				this.modalOff = 0
			} // if

			return this, nil
		} // if

		switch {
		case this.focus < len(this.comp):
			this.comp[this.focus].Move(this.stepper.game, msg.key)

		case this.focus == focusLog:
			this.log.Move(msg.key)
		} // switch

		return this, nil // 狀態列無游標(聚焦即整列), 落空不動作

	case helpMsg:
		return this.push(modalHelp{}), nil

	case countMsg:
		return this.push(modalCount{seed: this.stepper.game.Seed}), nil

	case enterMsg:
		if this.focus == focusStatus { // 狀態列聚焦 Enter = 開計數 modal(【營業顯示規格書 | 7、互動規格 | 7.3】)
			return this.push(modalCount{seed: this.stepper.game.Seed}), nil
		} // if

		if this.focus < len(this.comp) { // 對游標項目辨型開檢視 modal(M26 R4; 候選身分 = 實例指標)
			switch item := this.comp[this.focus].Item(this.stepper.game).(type) {
			case *cores.Guest:
				return this.push(modalGuest{guest: item}), nil

			case *cores.Action:
				return this.push(modalAction{action: item}), nil

			case *cores.Effect:
				return this.push(modalEffect{effect: item}), nil

			case *cores.Card:
				return this.push(modalCard{card: item}), nil
			} // switch
		} // if

		return this, nil // 游標下無項目 / 事件日誌無 modal: 不動作

	case popMsg:
		if size := len(this.modal); size > 0 {
			this.modal = this.modal[:size-1]
			this.modalOff = 0
			return this, this.tick() // 全關後依當前模式恢復排拍(仍有 modal 或步進時 tick 自回 nil)
		} // if

		return this, nil // 空棧防呆: 不重排拍, 免與既有拍鏈雙鏈

	case tea.WindowSizeMsg:
		this.width = msg.Width
		this.height = msg.Height
		return this, nil

	case tea.KeyMsg:
		return this, this.keybar.Find(this.keymode(), msg.String())
	} // switch

	return this, nil
}

// View 全畫面組合: 低於最低尺寸守門(顯提示不硬塞; 【營業顯示規格書 | 8、終端機尺寸與 flex】);
// 日誌寬 = 總寬扣左欄基準後夾在地板與上限間(餘寬先給日誌、到頂才給左欄; M26 拍板),
// 左右兩欄水平拼接(日誌與左欄同高)、父層補全寬底框列(M26 R1.5)、鍵位列收尾, 共 height 行;
// 聚焦日誌時其標題列高亮(左欄聚焦歸 leftView)。
func (this model) View() string {
	if this.width < minWidth || this.height < minHeight {
		return fmt.Sprintf("請放大終端機 (現 %vx%v, 最低 %vx%v)", this.width, this.height, minWidth, minHeight)
	} // if

	logw := this.width - (minWidth - logWidthMin) // 守門後必 >= 地板(寬 100 → 30), 不需下限 clamp

	if logw > logWidthMax {
		logw = logWidthMax
	} // if

	body := this.height - 3 // 鍵位列固定 2 行 + 底框列 1 行
	left := this.leftView(this.width-logw, body)
	logview := this.log.View(logw, body)

	if this.focus == focusLog {
		logview = focusView(logview)
	} // if

	bottom := "+" + strings.Repeat("-", this.width-logw-2) + "+" + strings.Repeat("-", logw-1) + "+"
	view := lipgloss.JoinHorizontal(lipgloss.Top, left, logview) + "\n" + bottom + "\n" + this.keybar.View(this.keymode(), this.width)

	if size := len(this.modal); size > 0 { // 頂層 modal 置中疊在全畫面上(M26 R3)
		view = overlay(view, modalView(this.stepper.game, this.modal[size-1], this.modalOff), this.width, this.height)
	} // if

	return view
}

// leftView 左欄主畫面(共 height 行): 六區堆疊 + 帶框空行墊高(格線不開洞; M26 R1.5)+ 狀態列釘底
// (【營業顯示規格書 | 6、畫面規格 | 6.2】置底); 各區固定高、唯一會變長的日誌在右欄, 內容超高(防禦)不裁。
// 盤面直讀暫停機的營業實例(停點間引擎必停)。聚焦六區之一時該組件標題列高亮(經 composeView)、
// 聚焦狀態列時其標題列高亮。
func (this model) leftView(width, height int) string {
	stack := composeView(this.stepper.game, width, this.comp, this.focus)
	status := this.status.View(this.stepper.game, this.mode, width)

	if this.focus == focusStatus {
		status = focusView(status)
	} // if

	gap := height - strings.Count(stack, "\n") - strings.Count(status, "\n") - 2

	if gap < 0 {
		gap = 0
	} // if

	row := []string{stack}

	for i := 0; i < gap; i++ {
		row = append(row, boxRow("", width))
	} // for

	row = append(row, status)
	return strings.Join(row, "\n")
}

// push 開 modal: 疊層入棧、捲動歸零、世代 +1 作廢在途拍(開著時 tick 門控停排——
// M26 拍板「modal 開著時暫停消費」; 速率 mode 不動, 關閉依當前模式恢復)。
func (this model) push(top modal) model {
	this.modal = append(this.modal, top)
	this.modalOff = 0
	this.gen++
	return this
}

// advance 推一拍: 同步 Next 收輪次——行組入日誌(盤面組件直讀引擎、不持拷貝); 輸入請求暫以被動答覆立答
// 再續收(R2 過渡, R3 / R4 換成真等待態); 終局回 false(呼叫端據此停止排拍)。
func (this model) advance() bool {
	next := this.stepper.Next()

	for next.role == turnRequest {
		next = this.stepper.Answer(next.req, passiveAnswer(next.req))
	} // for

	if next.role == turnOver {
		return false
	} // if

	this.log.Append(next.line...)
	return true
}

// tick 依當前模式排下一拍的 Cmd: 快/慢回 timer Cmd(訊息蓋上當前世代供驗章)、步進不排拍回 nil(等 [N])、
// modal 開著不排拍(暫停消費, 凍結盤面供檢視; M26 R3)。timer Cmd 跑在別條 goroutine,
// 只回訊息不碰引擎(Next 必須留在 Update 內)。
func (this model) tick() tea.Cmd {
	if this.mode == modeStep || len(this.modal) > 0 {
		return nil
	} // if

	return tea.Tick(this.mode.interval(), func(time.Time) tea.Msg {
		return tickMsg{gen: this.gen}
	})
}

// keymode 當前鍵盤模式(三模式框架; M26 R1 立框架): modal 堆疊非空 = modal 態、
// 選取等待 = 選取模式(M27 補來源), 其餘常態——模式由 model 狀態導出、不另存欄位。
func (this model) keymode() keyMode {
	if len(this.modal) > 0 {
		return keyModeModal
	} // if

	return keyModeNormal
}

// tickMsg 自動排拍訊息(快/慢模式 timer 投遞; 不載行組資料——行組由 Update 內同步 Next 取得)。
type tickMsg struct {
	gen int // 排出當下的世代(與 model.gen 不符 = 切模式前的在途舊拍, 丟棄)
}

// stepMsg 步進一拍訊息([N] 鍵投遞; 步進模式專用, 自動模式下被 Update 忽略)。
type stepMsg struct{}

// step 排步進訊息的 Cmd([N] 鍵綁定): 只回推進訊息、不碰引擎。
func step() tea.Cmd {
	return func() tea.Msg {
		return stepMsg{}
	}
}

// cycleMsg 模式循環訊息([Space] 鍵投遞): 快速 → 慢速 → 步進 循環。
type cycleMsg struct{}

// cycle 排模式循環訊息的 Cmd([Space] 鍵綁定)。
func cycle() tea.Cmd {
	return func() tea.Msg {
		return cycleMsg{}
	}
}

// tabMsg 切區訊息(Tab/Shift+Tab 鍵投遞): 聚焦索引循環移動。
type tabMsg struct {
	delta int // 移動方向(+1 正向 / -1 反向; 模數迴繞)
}

// tab 排切區訊息的 Cmd(Tab/Shift+Tab 鍵綁定)。
func tab(delta int) tea.Cmd {
	return func() tea.Msg {
		return tabMsg{delta: delta}
	}
}

// moveMsg 游標移動訊息(方向鍵投遞): 常態分派給聚焦區自持游標(語意隨區; 狀態列無游標不動作)、
// modal 態轉頂層 modal 捲動。
type moveMsg struct {
	key string // 方向鍵名(up / down / left / right)
}

// move 排游標移動訊息的 Cmd(方向鍵綁定)。
func move(key string) tea.Cmd {
	return func() tea.Msg {
		return moveMsg{key: key}
	}
}

// helpMsg 開格式說明 modal 訊息([F1] 鍵投遞)。
type helpMsg struct{}

// help 排開格式說明 modal 訊息的 Cmd([F1] 鍵綁定)。
func help() tea.Cmd {
	return func() tea.Msg {
		return helpMsg{}
	}
}

// countMsg 開計數 modal 訊息([F2] 鍵投遞)。
type countMsg struct{}

// count 排開計數 modal 訊息的 Cmd([F2] 鍵綁定)。
func count() tea.Cmd {
	return func() tea.Msg {
		return countMsg{}
	}
}

// enterMsg 檢視訊息([Enter] 鍵投遞): 狀態列聚焦開計數 modal; 其餘區的檢視 modal 歸 R4。
type enterMsg struct{}

// enter 排檢視訊息的 Cmd([Enter] 鍵綁定)。
func enter() tea.Cmd {
	return func() tea.Msg {
		return enterMsg{}
	}
}

// popMsg 關閉頂層 modal 訊息([Esc] 鍵投遞)。
type popMsg struct{}

// pop 排關閉 modal 訊息的 Cmd([Esc] 鍵綁定)。
func pop() tea.Cmd {
	return func() tea.Msg {
		return popMsg{}
	}
}
