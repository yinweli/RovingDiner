package cores

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/yinweli/RovingDiner/internal/exprs"
)

func TestSuiteCommandInstance(t *testing.T) {
	suite.Run(t, new(SuiteCommandInstance))
}

// SuiteCommandInstance 驗證實例化命令(commandInstance.go):*Add / *Roll / *Copy / *Clone / guestSpawn / waitAdd 與 newCard / newGuest / award 抽獎。
type SuiteCommandInstance struct {
	suite.Suite
}

func (this *SuiteCommandInstance) TestCardAdd() {
	runtime := NewRuntime(0)
	eng := this.engine(runtime)

	commandHandAdd(eng, nil, nums(103, 2)) // 卡 103 × 2 入手牌
	this.Require().Len(runtime.Hand, 2)
	card := runtime.Hand[0]
	this.Equal(int32(103), card.CardID)
	this.Equal(int32(2), card.Cost.Value)                      // 卡牌資料初始費用
	this.Equal(int32(1), card.Keep.Lock)                       // bool 欄 → 鎖定計數
	this.Equal(int32(1), card.Seal.Lock)                       //
	this.Equal([]int32{401, 402}, card.EffectID)               // SkillID 301 → Skill.EffectID
	this.NotEqual(runtime.Hand[1].InstanceID, card.InstanceID) // 各自實例編號

	commandDeckAdd(eng, nil, nums(103, 1))
	commandDropAdd(eng, nil, nums(103, 1))
	commandExileAdd(eng, nil, nums(103, 1))
	this.Len(runtime.Deck, 1)
	this.Len(runtime.Drop, 1)
	this.Len(runtime.Exile, 1)

	commandHandAdd(eng, nil, nums(103, 0)) // N <= 0 → no-op
	commandHandAdd(eng, nil, nums(999, 1)) // 卡牌資料不存在 → no-op
	commandHandAdd(eng, nil, nums(103))    // 缺 N → no-op
	commandHandAdd(eng, nil, nil)          // 缺卡牌編號 → no-op
	this.Len(runtime.Hand, 2)
}

func (this *SuiteCommandInstance) TestCardRoll() {
	runtime := NewRuntime(0)
	eng := this.engine(runtime)

	commandHandRoll(eng, nil, nums(7, 2)) // 群組 7 抽 2(fakeRander.Weighted → index 0 → 卡 101)
	this.Require().Len(runtime.Hand, 2)
	this.Equal(int32(101), runtime.Hand[0].CardID)

	commandDeckRoll(eng, nil, nums(7, 1))
	commandDropRoll(eng, nil, nums(7, 1))
	commandExileRoll(eng, nil, nums(7, 1))
	this.Len(runtime.Deck, 1)
	this.Len(runtime.Drop, 1)
	this.Len(runtime.Exile, 1)

	commandHandRoll(eng, nil, nums(8, 1))  // 群組 8 權重 0 → no-op
	commandHandRoll(eng, nil, nums(9, 1))  // 群組 9 抽中編號 999 無卡牌資料 → 跳過該張
	commandHandRoll(eng, nil, nums(99, 1)) // 群組不存在 → no-op
	commandHandRoll(eng, nil, nums(7, 0))  // N <= 0 → no-op
	commandHandRoll(eng, nil, nums(7))     // 缺 N → no-op
	commandHandRoll(eng, nil, nil)         // 缺抽獎群組編號 → no-op
	this.Len(runtime.Hand, 2)
}

func (this *SuiteCommandInstance) TestCardCopy() {
	runtime := NewRuntime(0)
	source := &Card{InstanceID: runtime.NextID(), CardID: 103, Cost: Value{Value: 9}, EffectID: []int32{777}}
	runtime.Hand = []*Card{source}
	eng := this.engine(runtime)

	commandHandCopy(eng, []InstanceID{1}, nums(1)) // 淺複製:載卡牌資料初始值
	this.Require().Len(runtime.Hand, 2)
	shallow := runtime.Hand[0]
	this.Equal(int32(2), shallow.Cost.Value)        // 初始費用 2(非 source 當前 9)
	this.Equal([]int32{401, 402}, shallow.EffectID) // 技能效果列表

	// deck 變體:洗牌參數於 index 1、附加效果自 index 2
	runtime.Deck = nil
	commandDeckCopy(eng, []InstanceID{1}, []exprs.Value{exprs.NewNum(1), exprs.NewBool(true), exprs.NewNum(888)})
	this.Require().Len(runtime.Deck, 1)
	this.Equal([]int32{401, 402, 888}, runtime.Deck[0].EffectID) // 技能效果 + 附加 888

	commandDropCopy(eng, []InstanceID{1}, nums(1))
	commandExileCopy(eng, []InstanceID{1}, nums(1))
	this.Len(runtime.Drop, 1)
	this.Len(runtime.Exile, 1)
}

func (this *SuiteCommandInstance) TestCardClone() {
	runtime := NewRuntime(0)
	source := &Card{InstanceID: runtime.NextID(), CardID: 103, Cost: Value{Value: 9}, EffectID: []int32{777}}
	runtime.Hand = []*Card{source}
	eng := this.engine(runtime)

	commandHandClone(eng, []InstanceID{1}, nums(1, 888, 999)) // 深複製 + 附加 888,999
	this.Require().Len(runtime.Hand, 2)
	deep := runtime.Hand[0]
	this.Equal(int32(9), deep.Cost.Value)             // source 當前狀態
	this.Equal([]int32{777, 888, 999}, deep.EffectID) // source 效果 + 附加
	this.NotEqual(source.InstanceID, deep.InstanceID) // 實例編號重生

	commandDeckClone(eng, []InstanceID{1}, nums(1))
	commandDropClone(eng, []InstanceID{1}, nums(1))
	commandExileClone(eng, []InstanceID{1}, nums(1))
	this.Len(runtime.Deck, 1)
	this.Len(runtime.Drop, 1)
	this.Len(runtime.Exile, 1)
}

func (this *SuiteCommandInstance) TestCopyCloneNoop() {
	runtime := NewRuntime(0)
	source := &Card{InstanceID: 1, CardID: 103}
	missing := &Card{InstanceID: 2, CardID: 999} // 卡 999 無 sheet
	runtime.Hand = []*Card{source, missing}
	eng := this.engine(runtime)

	commandHandCopy(eng, []InstanceID{2}, nums(1)) // 淺複製載入失敗 → 跳過該張
	this.Len(runtime.Hand, 2)

	commandHandCopy(eng, []InstanceID{99}, nums(1)) // 非卡牌實例 → no-op
	this.Len(runtime.Hand, 2)

	commandHandCopy(eng, []InstanceID{1}, nums(0)) // N <= 0 → 無新卡
	this.Len(runtime.Hand, 2)

	commandHandCopy(eng, []InstanceID{1}, nil) // 缺 N → no-op
	this.Len(runtime.Hand, 2)
}

func (this *SuiteCommandInstance) TestGuestSpawn() {
	runtime := NewRuntime(0)
	eng := this.engine(runtime)

	commandGuestSpawn(eng, nil, nums(501, 0)) // 座位 0 → 遊蕩 + 自動鎖
	this.Require().Len(runtime.Roam, 1)
	guest := runtime.Roam[0]
	this.Equal(int32(501), guest.GuestID)
	this.Equal(int32(5), guest.Morale.Value)   // 顧客資料初始值
	this.Equal(int32(12), guest.SateMax.Value) // 飽食值離場線取自顧客資料
	this.Equal(int32(0), guest.Sate.Value)     // 飽食值初值 0
	this.Equal(int32(1), guest.Sate.Lock)      // 自動鎖 +1
	this.Equal(int32(1), guest.CalmSeal.Lock)  // sheet false → 0,自動鎖 +1

	commandGuestSpawn(eng, nil, nums(501, 2)) // 座位 2(buildSheet 存在且空)→ 入座
	this.Require().NotNil(runtime.Seat[2])
	this.Equal(int32(2), runtime.Seat[2].SeatID)

	occupied := runtime.Seat[2]
	commandGuestSpawn(eng, nil, nums(501, 2)) // 座位已占用 → no-op
	this.Equal(occupied, runtime.Seat[2])

	commandGuestSpawn(eng, nil, nums(501, 99)) // 座位不存在 → no-op
	this.Nil(runtime.Seat[99])

	commandGuestSpawn(eng, nil, nums(501, -1)) // 座位 < 0 → no-op
	commandGuestSpawn(eng, nil, nums(999, 0))  // 顧客資料不存在 → no-op
	commandGuestSpawn(eng, nil, nums(501))     // 缺座位編號 → no-op
	commandGuestSpawn(eng, nil, nil)           // 缺顧客編號 → no-op
	this.Len(runtime.Roam, 1)
}

func (this *SuiteCommandInstance) TestWaitAdd() {
	runtime := NewRuntime(0)
	runtime.Wait = []*Guest{{InstanceID: 1}} // 既有 1 位
	eng := this.engine(runtime)

	commandWaitAdd(eng, nil, nums(501, 2)) // 加 2 位至前端(優先入座)
	this.Require().Len(runtime.Wait, 3)
	this.Equal(int32(501), runtime.Wait[0].GuestID)
	this.Equal(InstanceID(1), runtime.Wait[2].InstanceID) // 原有者沉到尾

	commandWaitAdd(eng, nil, nums(501, 0)) // N <= 0 → no-op
	commandWaitAdd(eng, nil, nums(999, 1)) // 顧客資料不存在 → no-op
	commandWaitAdd(eng, nil, nums(501))    // 缺 N → no-op
	commandWaitAdd(eng, nil, nil)          // 缺顧客編號 → no-op
	this.Len(runtime.Wait, 3)
}

// === 測試輔助(置尾) ===

func (this *SuiteCommandInstance) engine(runtime *Runtime) *Engine {
	return NewEngine(runtime, nil, buildSheet(), fakeOperator{}, fakeRander{})
}
