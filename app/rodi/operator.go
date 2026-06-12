package rodi

import (
	"github.com/yinweli/RovingDiner/internal/cores"
)

// keyboardOperator 鍵盤玩家輸入(M27; M19–M26 的 passiveOperator 退役): 跑在引擎 goroutine 上,
// 把每個暫停點包成輸入請求輪次送進交棒 channel、阻塞等 UI 答覆——引擎視角仍是同步阻塞呼叫
// (邊界介面模型不變, 詳見營業實作規格書【四、解耦的關鍵：邊界介面】), 暫停 / 恢復全活在 channel 會合。
type keyboardOperator struct {
	turn chan<- turn // 交棒輪次發送端(與 stepper.turn 同一條)
}

// PlayerAction 玩家行動請求(無候選欄): 答覆空 = 玩家結束、一張 = 出該卡。
func (this keyboardOperator) PlayerAction(game *cores.Game) *cores.Card {
	ans := this.ask(request{})

	if len(ans.card) == 0 {
		return nil
	} // if

	return ans.card[0]
}

func (this keyboardOperator) PickGuest(prompt string, source []*cores.Guest, count int) []*cores.Guest {
	return this.ask(request{prompt: prompt, guest: source, count: count}).guest
}

func (this keyboardOperator) PickCard(prompt string, source []*cores.Card, count int) []*cores.Card {
	return this.ask(request{prompt: prompt, card: source, count: count}).card
}

func (this keyboardOperator) PickDiscard(prompt string, source []*cores.Card, over int) []*cores.Card {
	return this.ask(request{prompt: prompt, card: source, count: over}).card
}

// ask 送出輸入請求輪次並阻塞等答覆(答覆通道緩衝 1, UI 答覆不阻塞)。
func (this keyboardOperator) ask(req request) answer {
	req.answer = make(chan answer, 1)
	this.turn <- turn{role: turnRequest, req: &req}
	return <-req.answer
}

// request 輸入請求(turnRequest 載荷): 辨型看候選欄——guest 非 nil = 選顧客、card 非 nil = 選卡牌、
// 皆 nil = 玩家行動(出牌 / 結束)。候選身分 = 實例指標(與面板渲染同批, 指標相等即辨識、零 ID 對映; M27 拍板)。
type request struct {
	prompt string         // 選取提示前文(引擎組好, 【營業顯示規格書 | 6、畫面規格 | 6.11】; 玩家行動為空)
	guest  []*cores.Guest // 顧客候選
	card   []*cores.Card  // 卡牌候選
	count  int            // 目標上限 N(玩家行動不適用)
	answer chan answer    // 答覆通道(緩衝 1)
}

// answer 輸入答覆: 顧客選取答 guest、卡牌選取 / 棄牌答 card、玩家行動答 card(空 = 結束、一張 = 出該卡)。
type answer struct {
	guest []*cores.Guest // 選中顧客
	card  []*cores.Card  // 選中卡牌 / 出牌
}
