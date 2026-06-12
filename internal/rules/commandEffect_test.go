package rules

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/yinweli/RovingDiner/internal/cores"
	"github.com/yinweli/RovingDiner/internal/exprs"
	"github.com/yinweli/RovingDiner/internal/tester"
)

func TestSuiteCommandEffect(t *testing.T) {
	suite.Run(t, new(SuiteCommandEffect))
}

// SuiteCommandEffect 驗證效果佇列命令(commandEffect.go: effectClear / effectDel / effectRun);
// 經 ExecOperate 真實入口測: none 全域掃描 vs 篩空 no-op、群組 / 同份比對、退層與出佇列、依類型派發與 N override。
type SuiteCommandEffect struct {
	suite.Suite
}

func (this *SuiteCommandEffect) TestCommandEffectClear() {
	ended := []int32{}
	record := func(id int32) cores.EffectExec {
		return func(game *cores.Game) { ended = append(ended, id) }
	}

	data := tester.BuildData()
	game := newGameData(data)
	seated := cores.NewGuest(game, 501)
	game.Seat.Place(1, seated)
	frozen := cores.NewGuest(game, 501)
	game.Cardify.Push(frozen)

	data.SetEffect(901, cores.EffectData{Kind: cores.EffectPersist, Group: 5, RunOrder: 10, End: record(901)})
	data.SetEffect(902, cores.EffectData{Kind: cores.EffectPersist, Group: 6, RunOrder: 20, End: record(902)})
	data.SetEffect(903, cores.EffectData{Kind: cores.EffectPersist, Group: 7, End: record(903)})
	data.SetEffect(904, cores.EffectData{Kind: cores.EffectPersist, Group: 0, End: record(904)})
	data.SetEffect(905, cores.EffectData{Kind: cores.EffectPersist, Group: 5, End: record(905)})

	game.Effect.Push(cores.NewEffect(game, 901, cores.NewRefGuest(seated), 2)) // self ∈ 對象、群組 5
	game.Effect.Push(cores.NewEffect(game, 902, cores.Ref{}, 1))               // self 空物件、群組 6 → 僅全域掃描可清
	game.Effect.Push(cores.NewEffect(game, 903, cores.NewRefGuest(seated), 1)) // 群組 7 → 未指定不清
	game.Effect.Push(cores.NewEffect(game, 904, cores.Ref{}, 1))               // 群組 0 → 永不命中
	game.Effect.Push(cores.NewEffect(game, 905, cores.NewRefGuest(frozen), 1)) // 凍結 → 不清

	game.ExecOperate("effectClear", "guestAll", nil, this.arg("5", "6")) // 對象集合 {seated}
	this.Equal([]int32{901, 901}, ended)                                 // 901 層數 2 → 結束命令 × 2;902 self 不屬於集合
	this.Len(game.Effect, 4)

	ended = nil
	game.ExecOperate("effectClear", "none", nil, this.arg("6", "0")) // none → 全域掃描; 指定 0 不命中群組 0
	this.Equal([]int32{902}, ended)
	this.Len(game.Effect, 3)

	ended = nil
	game.ExecOperate("effectClear", "none", nil, this.arg("5")) // 全域掃描遇凍結中顧客的效果 → 跳過
	this.Empty(ended)
	this.Len(game.Effect, 3)

	// 卡牌 self 的身分比對
	card := cores.NewCard(game, 101)
	game.Hand = cores.CardList{card}
	data.SetEffect(906, cores.EffectData{Kind: cores.EffectPersist, Group: 5, End: record(906)})
	game.Effect.Push(cores.NewEffect(game, 906, cores.NewRefCard(card), 1))

	ended = nil
	game.ExecOperate("effectClear", "handAll", this.arg("0"), this.arg("5"))
	this.Equal([]int32{906}, ended)
	this.Len(game.Effect, 3)

	ended = nil
	game.Seat.Remove(seated) // 對象篩空(非 none)→ 整體 no-op, 非全域掃描
	game.ExecOperate("effectClear", "guestAll", nil, this.arg("7"))
	this.Empty(ended)
	this.Len(game.Effect, 3)
}

func (this *SuiteCommandEffect) TestCommandEffectDel() {
	ended := []int32{}
	data := tester.BuildData()
	game := newGameData(data)
	seated := cores.NewGuest(game, 501)
	game.Seat.Place(1, seated)

	data.SetEffect(901, cores.EffectData{Kind: cores.EffectTrigger, TriggerKind: cores.TriggerCardPlay, End: func(game *cores.Game) {
		this.Equal(seated, game.GetSelf().GetGuest()) // 退層結束命令以該效果 self 綁定執行
		ended = append(ended, 901)
	}})
	target := cores.NewEffect(game, 901, cores.NewRefGuest(seated), 5)
	game.Effect.Push(target)

	game.ExecOperate("effectDel", "guestAll", nil, this.arg("901", "2")) // 退 2 層 → 結束命令 × 2、留佇列
	this.Equal([]int32{901, 901}, ended)
	this.Len(game.Effect, 1)
	this.Equal(int32(3), target.GetStack())

	ended = nil
	game.ExecOperate("effectDel", "guestAll", nil, this.arg("901", "9")) // 超量 → 扣到 0、出佇列
	this.Equal([]int32{901, 901, 901}, ended)
	this.Empty(game.Effect)

	ended = nil
	game.ExecOperate("effectDel", "guestAll", nil, this.arg("901", "1")) // 同份不存在 → no-op
	game.ExecOperate("effectDel", "guestAll", nil, this.arg("901", "0")) // N <= 0 → no-op
	game.ExecOperate("effectDel", "guestAll", nil, this.arg("901"))      // 缺 N → no-op
	game.ExecOperate("effectDel", "guestAll", nil, this.arg("999", "1")) // 查無編譯資料 → no-op
	commandEffectDel(game, []cores.InstanceID{99999}, nums(901, 1))      // 實例不存在 → 該項 no-op(白箱)
	this.Empty(ended)
}

// TestEffectDelEmit 驗證 effectDel 的效果頭: 退層與歸零出佇列各發一組 結束 三行
// (層數與去留不入行——佇列項直讀盤面即見; M24 拍板)。
func (this *SuiteCommandEffect) TestEffectDelEmit() {
	data := tester.BuildData()
	data.SetEffect(901, cores.EffectData{Stack: 3, StackMax: 5})
	game, record := newGameDataRecord(data)
	guest := cores.NewGuest(game, 501) // 實例編號 1; 效果實例編號 2
	game.Seat.Place(1, guest)
	effect := cores.NewEffect(game, 901, cores.NewRefGuest(guest), 3)
	game.Effect.Push(effect)

	commandEffectDel(game, []cores.InstanceID{guest.GetInstanceID()}, nums(901, 1)) // 退 1 層 → 留佇列
	this.Require().Len(record.Line, 1)
	this.Equal([]string{"- 901@?#2", "  501@#1", "  結束"}, record.Line[0])

	commandEffectDel(game, []cores.InstanceID{guest.GetInstanceID()}, nums(901, 9)) // 歸零 → 出佇列
	this.Require().Len(record.Line, 2)
	this.Equal(record.Line[0], record.Line[1]) // 退層與退場行文相同
	this.Empty(game.Effect)
}

func (this *SuiteCommandEffect) TestCommandEffectRun() {
	started := []int32{}
	immed := 0
	data := tester.BuildData()
	game := newGameData(data)
	card := cores.NewCard(game, 101)
	game.Hand = cores.CardList{card}

	data.SetEffect(901, cores.EffectData{Kind: cores.EffectImmed, Immed: func(game *cores.Game) {
		this.Equal(card, game.GetSelf().GetCard()) // 以命令對象元素為 self
		immed++
	}})
	data.SetEffect(902, cores.EffectData{Kind: cores.EffectTrigger, TriggerKind: cores.TriggerCardPlay, Stack: 1})
	data.SetEffect(903, cores.EffectData{Kind: cores.EffectPersist, Stack: 2, Start: func(game *cores.Game) { started = append(started, 903) }})

	game.ExecOperate("effectRun", "handAll", this.arg("0"), this.arg("901", "5")) // 立即: 忽略 N、條件過即跑一次, 不入佇列
	this.Equal(1, immed)
	this.Empty(game.Effect)

	game.ExecOperate("effectRun", "handAll", this.arg("0"), this.arg("902", "3")) // 觸發: N > 0 覆寫本次增加層數
	this.Require().Len(game.Effect, 1)
	this.Equal(int32(3), game.Effect[0].GetStack())
	this.Equal(card, game.Effect[0].GetSelf().GetCard())

	game.ExecOperate("effectRun", "handAll", this.arg("0"), this.arg("903", "0")) // 常駐: N = 0 → 用效果堆疊層數 2; 啟動命令 × 2
	this.Equal([]int32{903, 903}, started)
	this.Len(game.Effect, 2)

	game.ExecOperate("effectRun", "handAll", this.arg("0"), this.arg("999", "1")) // 查無編譯資料 → no-op
	game.ExecOperate("effectRun", "handAll", this.arg("0"), this.arg("901"))      // 缺 N → no-op
	commandEffectRun(game, []cores.InstanceID{99999}, nums(901, 1))               // 實例不存在 → 該項 no-op(白箱)
	this.Equal(1, immed)
	this.Len(game.Effect, 2)
}

// === 測試輔助(置尾) ===

// arg 把多個來源各解析為 *exprs.Expr(模擬 parser 產出的命令參數); 解析失敗即測試失敗。
func (this *SuiteCommandEffect) arg(source ...string) (result []*exprs.Expr) {
	for _, itor := range source {
		expr, err := exprs.Parse(itor)
		this.Require().NoError(err)
		result = append(result, expr)
	} // for

	return result
}
