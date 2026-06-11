package rodi

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	sheeter "github.com/yinweli/RovingDiner/sheet"
)

// 畫面尺寸常數(【營業顯示規格書 | 8、終端機尺寸與 flex】): 以 ~100 欄為主要設計寬度——
// 右欄事件日誌固定 30 欄、左欄主畫面吃剩餘(終端機更寬時 flex 全給左欄, 座位可視更多桌);
// 低於最低支援尺寸顯提示、不硬塞。
const (
	logWidth  = 30  // 右欄事件日誌固定寬
	minWidth  = 100 // 最低支援寬
	minHeight = 30  // 最低支援高
)

// Run 顯示/操作層對外入口: 組裝暫停機(被動觀看 operator)與 Bubble Tea 殼(alt-screen; M22)後跑到離開。
// seed 由呼叫端先行定案(cmd 對 0 取時間亂數), 並由 cmd 在進 alt-screen 前把 seed / 關卡印進
// scrollback(退出後仍可見, 供重現與對帳; M22 拍板)。
func Run(seed int64, stageID int32, sheet *sheeter.Sheeter) error {
	_, err := tea.NewProgram(newModel(newStepper(seed, stageID, sheet), sheet), tea.WithAltScreen()).Run()

	if err != nil {
		return fmt.Errorf("rodi: %w", err)
	} // if

	return nil
}

// model Bubble Tea 殼(M22 alt-screen 換裝, M19 dump 退役): 全畫面 layout——左欄六區堆疊 + 狀態列釘底、
// 右欄事件日誌與左欄同高、鍵位列橫跨底部全寬; 父層只組合與分配空間(M20 拍板)。
// 消費採自我訊息鏈: stepMsg 抵達 → Update 同步 Next 推一拍(摺疊鏡像 + 轉寫日誌)、再排下一拍;
// 終局即停止推進、等 q 離開(成敗常駐顯示活在狀態列階段欄與日誌標題前綴)。
// Next 只在 Update 內呼叫(stepper 直讀安全窗的前提), 推進節奏全活在 Update 迴圈,
// 即 M26 速率的掛點(【營業顯示規格書 | 3、事件流的消費：速率與步進】)。
type model struct {
	stepper *stepper    // 暫停機橋接器
	world   *mirror     // 世界鏡像(事件摺疊一次、組件唯讀共用)
	log     *logPanel   // 事件日誌組件(右欄; 行歷史自持、簽章自立, 不入 comp; M22 拍板)
	keybar  keyBar      // 鍵位列(按鍵分派入口 + 底部全寬列)
	status  statusBar   // 狀態列(左欄釘底)
	comp    []component // 左欄堆疊組件(六區; 順序 = 堆疊順序)
	width   int         // 寬度預算(WindowSizeMsg 前用設計寬)
	height  int         // 高度預算(WindowSizeMsg 前用最低高)
}

func newModel(stepper *stepper, sheet *sheeter.Sheeter) model {
	return model{
		stepper: stepper,
		world:   newMirror(sheet),
		log:     newLogPanel(sheet),
		keybar:  newKeyBar(),
		status:  statusBar{},
		comp:    []component{seatPanel{}, poolPanel{}, actionPanel{}, effectPanel{}, handPanel{}, pilePanel{}},
		width:   minWidth,
		height:  minHeight,
	}
}

// Init 起跑: 排第一拍(header 已由 cmd 印於 scrollback, 殼不再印行)。
func (this model) Init() tea.Cmd {
	return step()
}

// Update 訊息分派: 推進拍 → 同步 Next 收一筆事件摺疊鏡像 + 轉寫日誌、再排下一拍, 終局停止推進
// (成敗已由終局 phase 事件投影, 不另動作); 視窗尺寸 → 更新寬高預算; 按鍵 → 查鍵綁定表分派(未綁定不動作)。
func (this model) Update(msg tea.Msg) (result tea.Model, cmd tea.Cmd) {
	switch msg := msg.(type) {
	case stepMsg:
		eventData, more := this.stepper.Next()

		if more == false {
			return this, nil
		} // if

		this.world.Apply(eventData)
		this.log.Append(eventData)
		return this, step()

	case tea.WindowSizeMsg:
		this.width = msg.Width
		this.height = msg.Height
		return this, nil

	case tea.KeyMsg:
		return this, this.keybar.Find(msg.String())
	} // switch

	return this, nil
}

// View 全畫面組合: 低於最低尺寸守門(顯提示不硬塞; 【營業顯示規格書 | 8、終端機尺寸與 flex】);
// 左右兩欄水平拼接(左欄寬 = 總寬 - 日誌欄寬、日誌與左欄同高)、鍵位列收尾, 共 height 行。
func (this model) View() string {
	if this.width < minWidth || this.height < minHeight {
		return fmt.Sprintf("請放大終端機 (現 %vx%v, 最低 %vx%v)", this.width, this.height, minWidth, minHeight)
	} // if

	body := this.height - 2 // 鍵位列固定 2 行
	left := this.leftView(this.width-logWidth, body)
	return lipgloss.JoinHorizontal(lipgloss.Top, left, this.log.View(logWidth, body)) + "\n" + this.keybar.View(this.world, this.width)
}

// leftView 左欄主畫面(共 height 行): 六區堆疊 + 留白墊高 + 狀態列釘底(【營業顯示規格書 | 6、畫面規格 | 6.2】
// 置底); 各區固定高、唯一會變長的日誌在右欄, 內容超高(防禦)不裁。
func (this model) leftView(width, height int) string {
	stack := composeView(this.world, width, this.comp)
	status := this.status.View(this.world, width)
	gap := height - strings.Count(stack, "\n") - strings.Count(status, "\n") - 2

	if gap < 0 {
		gap = 0
	} // if

	return stack + strings.Repeat("\n", gap+1) + status
}

// stepMsg 推進一拍(Bubble Tea 訊息殼; 不載資料——事件由 Update 內同步 Next 取得)。
type stepMsg struct{}

// step 排下一拍的 Cmd: 只回推進訊息、不碰引擎(Next 必須留在 Update 內, Cmd 跑在別條 goroutine)。
func step() tea.Cmd {
	return func() tea.Msg {
		return stepMsg{}
	}
}
