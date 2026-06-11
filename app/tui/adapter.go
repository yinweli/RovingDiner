package tui

import (
	"github.com/yinweli/RovingDiner/internal/cores"
	"github.com/yinweli/RovingDiner/internal/games"
	sheeter "github.com/yinweli/RovingDiner/sheet"
)

// adapter 引擎橋接器(【營業實作規格書 | 四、解耦的關鍵：邊界介面 | TUI adapter】): 背景 goroutine 跑
// games.Run, 事件經無緩衝 channel 逐筆送出——引擎的 Emit 阻塞到消費端收走為止, 背壓天生內建,
// 快/慢/步進(M25)只是消費節奏、引擎端零改動。done 緩衝 1:事件 channel 無緩衝, Run 返回時
// 事件必已全數被收走, 終局訊號不會搶先到達(waitEvent 的 select 依賴此不變式)。
type adapter struct {
	event chan cores.EventData // 事件流(無緩衝)
	done  chan bool            // 終局成敗(緩衝 1)
}

// newAdapter 建立橋接器並立即啟動引擎 goroutine; 事件自首筆起即阻塞等待消費端(waitEvent)收取。
// 中途離開不做取消: process 結束時阻塞中的引擎 goroutine 隨之消滅, 核心不吃 context(鐵則 1)、亦無外部資源要清理。
func newAdapter(seed int64, stageID int32, sheet *sheeter.Sheeter, operator cores.Operator) *adapter {
	this := &adapter{
		event: make(chan cores.EventData),
		done:  make(chan bool, 1),
	}
	go func() {
		this.done <- games.Run(seed, stageID, sheet, operator, chanPresenter{event: this.event})
	}()
	return this
}

// chanPresenter 事件流輸出 port 的橋接半身: Emit 把事件值送進 channel(EventData 傳值即快照, 跨 goroutine 天生安全)。
type chanPresenter struct {
	event chan<- cores.EventData // 與 adapter.event 同一條
}

func (this chanPresenter) Emit(eventData cores.EventData) {
	this.event <- eventData
}
