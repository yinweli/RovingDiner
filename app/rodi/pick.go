package rodi

import (
	"fmt"

	"github.com/yinweli/RovingDiner/internal/cores"
)

// 標題提醒文字(【營業顯示規格書 | 7、互動規格 | 7.1 | 等待落點標題提醒】): 輸入等待時加註於候選
// 所在面板的標題(出牌等待 = 手牌), 不論聚焦何區恆顯、答覆後消失——Tab 切走仍看得到等待落點。
const (
	notePlay  = " (請出牌)"
	noteGuest = " (請選顧客)"
	noteCard  = " (請選卡牌)"
)

// pickState 輸入等待共享狀態(M27 R4; 【營業顯示規格書 | 7、互動規格 | 7.4】): model 與候選面板
// (座位 / 手牌 / 牌堆)經 newModel 注入同一份指標——model 寫(輸入請求到達 start、答覆 stop)、
// 面板讀(選取三視覺態著色與游標限縮吸附、標題提醒)。出牌等待(玩家行動請求, 無候選)也入此狀態,
// 僅供手牌標題提醒、不啟動選取模式。候選身分 = 實例指標(與面板渲染同批, 指標相等即辨識、
// 零 ID 對映; M27 拍板); 確認答覆依候選序組裝, 與加選順序無關、保決定性。
type pickState struct {
	req    *request // 進行中的選取請求(nil = 非選取模式)
	picked []bool   // 已選標記(與候選同序)
}

// start 進入輸入等待(已選標記歸零重配; 出牌等待無候選即零長)。
func (this *pickState) start(req *request) {
	this.req = req
	this.picked = make([]bool, len(req.guest)+len(req.card))
}

// stop 離開輸入等待(答覆後清空)。
func (this *pickState) stop() {
	this.req = nil
	this.picked = nil
}

// active 回報選取模式是否進行中(請求帶候選; 出牌等待不算)。nil 接收器安全:
// 未注入共享狀態的面板恆走非選取路徑。
func (this *pickState) active() bool {
	return this != nil && this.req != nil && (this.req.guest != nil || this.req.card != nil)
}

// playing 回報出牌等待是否進行中(玩家行動請求, 無候選; 辨型同 operator 的 request): 手牌標題提醒用。
func (this *pickState) playing() bool {
	return this != nil && this.req != nil && this.req.guest == nil && this.req.card == nil
}

// guestIndex 顧客候選索引(指標相等; 非候選 / 非選取模式回 -1)。
func (this *pickState) guestIndex(guest *cores.Guest) int {
	if this.req == nil || guest == nil {
		return -1
	} // if

	for index, itor := range this.req.guest {
		if itor == guest {
			return index
		} // if
	} // for

	return -1
}

// cardIndex 卡牌候選索引(指標相等; 非候選 / 非選取模式回 -1)。
func (this *pickState) cardIndex(card *cores.Card) int {
	if this.req == nil || card == nil {
		return -1
	} // if

	for index, itor := range this.req.card {
		if itor == card {
			return index
		} // if
	} // for

	return -1
}

// chosen 回報候選是否已選(越界 / -1 視為未選)。
func (this *pickState) chosen(index int) bool {
	return index >= 0 && index < len(this.picked) && this.picked[index]
}

// toggle 加選 / 取消選候選(【營業顯示規格書 | 7、互動規格 | 7.4】): 已選取消;
// 未選且未滿 N 加選(已選達 N 再加選須先取消其一, no-op); 越界 / -1 不動作。
func (this *pickState) toggle(index int) {
	if index < 0 || index >= len(this.picked) {
		return
	} // if

	if this.picked[index] {
		this.picked[index] = false
		return
	} // if

	if this.count() < this.req.count {
		this.picked[index] = true
	} // if
}

// count 已選數。
func (this *pickState) count() (result int) {
	for _, itor := range this.picked {
		if itor {
			result++
		} // if
	} // for

	return result
}

// full 回報已選滿 N([Enter] 確認條件——選滿才可確認; 候選 > N 才進互動, 必選得滿;
// 出牌等待非選取模式, 恆不可確認)。
func (this *pickState) full() bool {
	return this.active() && this.count() == this.req.count
}

// result 組確認答覆: 依候選序收已選(與加選順序無關、保決定性)。
func (this *pickState) result() (ans answer) {
	for index, itor := range this.req.guest {
		if this.picked[index] {
			ans.guest = append(ans.guest, itor)
		} // if
	} // for

	for index, itor := range this.req.card {
		if this.picked[index] {
			ans.card = append(ans.card, itor)
		} // if
	} // for

	return ans
}

// hint 鍵位列行 2 選取提示(【營業顯示規格書 | 6、畫面規格 | 6.11】): 提示前文引擎組好, 此處附加已選進度。
func (this *pickState) hint() string {
	return fmt.Sprintf("%v (已選 %v/%v)", this.req.prompt, this.count(), this.req.count)
}
