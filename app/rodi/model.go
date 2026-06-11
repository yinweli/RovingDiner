package rodi

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/yinweli/RovingDiner/internal/cores"
	sheeter "github.com/yinweli/RovingDiner/sheet"
)

// Run 顯示/操作層對外入口: 組裝橋接器(被動觀看 operator)與 Bubble Tea 殼後跑到離開。
// seed 由呼叫端先行定案(cmd 對 0 取時間亂數), 本層只負責驅動與消費。
func Run(seed int64, stageID int32, dataDir string, sheet *sheeter.Sheeter) error {
	_, err := tea.NewProgram(newModel(newAdapter(seed, stageID, sheet, passiveOperator{}), seed, stageID, dataDir, sheet)).Run()

	if err != nil {
		return fmt.Errorf("rodi: %w", err)
	} // if

	return nil
}

// model Bubble Tea 殼(M20 渲染地基): 維持 inline 模式——事件逐行印進終端機 scrollback(純文字 dump,
// 一事件一行「序號 + %+v 原始欄位」, 事件日誌格式由 M22 整套重做), View 渲染組件 footer(狀態列 + 鍵位列);
// 父層只組合與發寬度預算, alt-screen / 高度分配 / 最低尺寸守門綁 M22(M20 拍板)。
// 消費採波浪式 Cmd: waitEvent 收一筆 → Update 摺疊鏡像、印行並再發 waitEvent; 收到終局即停止消費、等 q 離開——
// 消費節奏全活在 Update 迴圈, 即 M25 速率的掛點(【營業顯示規格書 | 3、事件流的消費：速率與步進】)。
type model struct {
	adapter *adapter    // 引擎橋接器
	seed    int64       // 本場 seed(顯示用)
	stageID int32       // 關卡編號(顯示用)
	dataDir string      // 靜態表目錄(顯示用)
	world   *mirror     // 世界鏡像(事件摺疊一次、組件唯讀共用)
	keybar  keyBar      // 鍵位列(按鍵分派入口, 同時掛載於組件列表)
	comp    []component // 掛載組件(順序 = 堆疊順序)
	width   int         // 寬度預算(WindowSizeMsg 前用設計寬 100; 【營業顯示規格書 | 8、終端機尺寸與 flex】)
	serial  int         // 已收事件序號(dump 行前綴)
}

func newModel(adapter *adapter, seed int64, stageID int32, dataDir string, sheet *sheeter.Sheeter) model {
	keybar := newKeyBar()
	return model{
		adapter: adapter,
		seed:    seed,
		stageID: stageID,
		dataDir: dataDir,
		world:   newMirror(sheet),
		keybar:  keybar,
		comp:    []component{seatPanel{}, poolPanel{}, actionPanel{}, effectPanel{}, handPanel{}, pilePanel{}, statusBar{}, keybar},
		width:   100,
	}
}

// Init 起跑: 印 header(seed/關卡/資料目錄, 供重現與對帳)並開始等第一筆事件。
// 印行與續等用 tea.Sequence 循序執行(先印完、再等下一筆): tea.Batch 並發會讓相鄰兩行 Printf 搶序, 事件 dump 行序必須保序。
func (this model) Init() tea.Cmd {
	return tea.Sequence(tea.Printf("營業開始 (seed %v, stage %v, data %v)", this.seed, this.stageID, this.dataDir), waitEvent(this.adapter))
}

// Update 訊息分派: 事件 → 摺疊鏡像 + 印行 + 續等; 終局 → 印終局行、停止消費(成敗常駐顯示活在狀態列階段欄,
// 終局事件座標 = 營業成功 / 失敗站); 視窗尺寸 → 更新寬度預算; 按鍵 → 查鍵綁定表分派(未綁定不動作)。
func (this model) Update(msg tea.Msg) (result tea.Model, cmd tea.Cmd) {
	switch msg := msg.(type) {
	case eventMsg:
		this.serial++
		this.world.Apply(cores.EventData(msg))
		return this, tea.Sequence(tea.Printf("%5d %+v", this.serial, cores.EventData(msg)), waitEvent(this.adapter)) // 循序保行序, 依 Init 同註

	case doneMsg:
		state := "失敗"

		if bool(msg) {
			state = "成功"
		} // if

		return this, tea.Printf("營業結束: %v (seed %v, stage %v)", state, this.seed, this.stageID)

	case tea.WindowSizeMsg:
		this.width = msg.Width
		return this, nil

	case tea.KeyMsg:
		return this, this.keybar.Find(msg.String())
	} // switch

	return this, nil
}

// View 組件 footer: 父層只組合(狀態列 + 鍵位列)與發寬度預算; 事件本文已印進 scrollback、不歸 View 管。
// 終局不自動退出(留人看尾巴), 等 q。
func (this model) View() string {
	return composeView(this.world, this.width, this.comp) + "\n"
}

// eventMsg 一筆引擎事件抵達(Bubble Tea 訊息殼)。
type eventMsg cores.EventData

// doneMsg 營業跑完(載成敗)。
type doneMsg bool

// waitEvent 等待下一筆事件或終局的 Cmd: 兩者同一個等待點——done 緩衝 1 且引擎送完所有事件才送終局
// (adapter 不變式), select 不會在尚有事件未收時撿到終局。
func waitEvent(adapter *adapter) tea.Cmd {
	return func() tea.Msg {
		select {
		case eventData := <-adapter.event:
			return eventMsg(eventData)

		case succ := <-adapter.done:
			return doneMsg(succ)
		}
	}
}
