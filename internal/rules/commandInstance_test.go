package rules

import (
	"testing"

	"github.com/yinweli/RovingDiner/internal/cores"

	"github.com/stretchr/testify/suite"

	"github.com/yinweli/RovingDiner/internal/exprs"
)

func TestSuiteCommandInstance(t *testing.T) {
	suite.Run(t, new(SuiteCommandInstance))
}

// SuiteCommandInstance 驗證實例化命令(commandInstance.go): *Add / *Roll / *Copy / *Clone / guestSpawn / waitAdd 與 newCard / newGuest / award 抽獎。
type SuiteCommandInstance struct {
	suite.Suite
}

func (this *SuiteCommandInstance) TestCardAdd() {
	game := newGame()

	commandHandAdd(game, nil, nums(103, 2)) // 卡 103 × 2 入手牌
	this.Require().Len(game.Hand, 2)
	card := game.Hand[0]
	this.Equal(int32(103), card.GetCardID())
	this.Equal(int32(2), card.GetCost().GetValue())                   // 卡牌資料初始費用
	this.Equal(int32(1), card.GetKeep().GetLock())                    // bool 欄 → 鎖定計數
	this.Equal(int32(1), card.GetSeal().GetLock())                    //
	this.Equal([]int32{401, 402}, card.GetEffectID().List())          // SkillID 301 → Skill.GetEffectID().List()
	this.NotEqual(game.Hand[1].GetInstanceID(), card.GetInstanceID()) // 各自實例編號

	commandDeckAdd(game, nil, nums(103, 1))
	commandDropAdd(game, nil, nums(103, 1))
	commandExileAdd(game, nil, nums(103, 1))
	this.Len(game.Deck, 1)
	this.Len(game.Drop, 1)
	this.Len(game.Exile, 1)

	commandHandAdd(game, nil, nums(103, 0)) // N <= 0 → no-op
	commandHandAdd(game, nil, nums(999, 1)) // 卡牌資料不存在 → no-op
	commandHandAdd(game, nil, nums(103))    // 缺 N → no-op
	commandHandAdd(game, nil, nil)          // 缺卡牌編號 → no-op
	this.Len(game.Hand, 2)
}

func (this *SuiteCommandInstance) TestCardRoll() {
	game := newGame()

	commandHandRoll(game, nil, nums(7, 2)) // 群組 7 抽 2(fakeRander.Weighted → index 0 → 卡 101)
	this.Require().Len(game.Hand, 2)
	this.Equal(int32(101), game.Hand[0].GetCardID())

	commandDeckRoll(game, nil, nums(7, 1))
	commandDropRoll(game, nil, nums(7, 1))
	commandExileRoll(game, nil, nums(7, 1))
	this.Len(game.Deck, 1)
	this.Len(game.Drop, 1)
	this.Len(game.Exile, 1)

	commandHandRoll(game, nil, nums(8, 1))  // 群組 8 權重 0 → no-op
	commandHandRoll(game, nil, nums(9, 1))  // 群組 9 抽中編號 999 無卡牌資料 → 跳過該張
	commandHandRoll(game, nil, nums(99, 1)) // 群組不存在 → no-op
	commandHandRoll(game, nil, nums(7, 0))  // N <= 0 → no-op
	commandHandRoll(game, nil, nums(7))     // 缺 N → no-op
	commandHandRoll(game, nil, nil)         // 缺抽獎群組編號 → no-op
	this.Len(game.Hand, 2)
}

func (this *SuiteCommandInstance) TestCardCopy() {
	game := newGame()
	source := sourceCard(game, 777)
	game.Hand = cores.CardList{source}
	id := []cores.InstanceID{source.GetInstanceID()}

	commandHandCopy(game, id, nums(1)) // 淺複製: 載卡牌資料初始值
	this.Require().Len(game.Hand, 2)
	shallow := game.Hand[0]
	this.Equal(int32(2), shallow.GetCost().GetValue())          // 初始費用 2(非 source 當前 9)
	this.Equal([]int32{401, 402}, shallow.GetEffectID().List()) // 技能效果列表

	// deck 變體: 洗牌參數於 index 1、附加效果自 index 2
	game.Deck = nil
	commandDeckCopy(game, id, []exprs.Value{exprs.NewNum(1), exprs.NewBool(true), exprs.NewNum(888)})
	this.Require().Len(game.Deck, 1)
	this.Equal([]int32{401, 402, 888}, game.Deck[0].GetEffectID().List()) // 技能效果 + 附加 888

	commandDropCopy(game, id, nums(1))
	commandExileCopy(game, id, nums(1))
	this.Len(game.Drop, 1)
	this.Len(game.Exile, 1)
}

func (this *SuiteCommandInstance) TestCardClone() {
	game := newGame()
	source := sourceCard(game, 777)
	game.Hand = cores.CardList{source}
	id := []cores.InstanceID{source.GetInstanceID()}

	commandHandClone(game, id, nums(1, 888, 999)) // 深複製 + 附加 888,999
	this.Require().Len(game.Hand, 2)
	deep := game.Hand[0]
	this.Equal(int32(9), deep.GetCost().GetValue())               // source 當前狀態
	this.Equal([]int32{777, 888, 999}, deep.GetEffectID().List()) // source 效果 + 附加
	this.NotEqual(source.GetInstanceID(), deep.GetInstanceID())   // 實例編號重生

	commandDeckClone(game, id, nums(1))
	commandDropClone(game, id, nums(1))
	commandExileClone(game, id, nums(1))
	this.Len(game.Deck, 1)
	this.Len(game.Drop, 1)
	this.Len(game.Exile, 1)
}

func (this *SuiteCommandInstance) TestCopyCloneNoop() {
	game := newGame()
	source := cores.NewCard(game, 103)
	missing := strayCard(999) // 卡 999 無 sheet
	game.Hand = cores.CardList{source, missing}

	commandHandCopy(game, []cores.InstanceID{missing.GetInstanceID()}, nums(1)) // 淺複製載入失敗 → 跳過該張
	this.Len(game.Hand, 2)

	commandHandCopy(game, []cores.InstanceID{99}, nums(1)) // 非卡牌實例 → no-op
	this.Len(game.Hand, 2)

	commandHandCopy(game, []cores.InstanceID{source.GetInstanceID()}, nums(0)) // N <= 0 → 無新卡
	this.Len(game.Hand, 2)

	commandHandCopy(game, []cores.InstanceID{source.GetInstanceID()}, nil) // 缺 N → no-op
	this.Len(game.Hand, 2)
}

func (this *SuiteCommandInstance) TestGuestSpawn() {
	game := newGame()

	commandGuestSpawn(game, nil, nums(501, 0)) // 座位 0 → 遊蕩 + 自動鎖
	this.Require().Len(game.Roam, 1)
	guest := game.Roam[0]
	this.Equal(int32(501), guest.GetGuestID())
	this.Equal(int32(5), guest.GetMorale().GetValue())   // 顧客資料初始值
	this.Equal(int32(12), guest.GetSateMax().GetValue()) // 飽食值離場線取自顧客資料
	this.Equal(int32(0), guest.GetSate().GetValue())     // 飽食值初值 0
	this.Equal(int32(1), guest.GetSate().GetLock())      // 自動鎖 +1
	this.Equal(int32(1), guest.GetCalmSeal().GetLock())  // sheet false → 0, 自動鎖 +1

	commandGuestSpawn(game, nil, nums(501, 2)) // 座位 2(buildSheet 存在且空)→ 入座
	this.Require().NotNil(game.Seat[2])
	this.Equal(int32(2), game.Seat[2].GetSeatID())

	occupied := game.Seat[2]
	commandGuestSpawn(game, nil, nums(501, 2)) // 座位已占用 → no-op
	this.Equal(occupied, game.Seat[2])

	commandGuestSpawn(game, nil, nums(501, 99)) // 座位不存在 → no-op
	this.Nil(game.Seat[99])

	commandGuestSpawn(game, nil, nums(501, -1)) // 座位 < 0 → no-op
	commandGuestSpawn(game, nil, nums(999, 0))  // 顧客資料不存在 → no-op
	commandGuestSpawn(game, nil, nums(501))     // 缺座位編號 → no-op
	commandGuestSpawn(game, nil, nil)           // 缺顧客編號 → no-op
	this.Len(game.Roam, 1)
}

func (this *SuiteCommandInstance) TestWaitAdd() {
	game := newGame()
	exist := cores.NewGuest(game, 501) // 既有 1 位
	game.Wait.Insert(exist)

	commandWaitAdd(game, nil, nums(501, 2)) // 加 2 位至前端(優先入座)
	this.Require().Len(game.Wait, 3)
	this.Equal(int32(501), game.Wait[0].GetGuestID())
	this.Same(exist, game.Wait[2]) // 原有者沉到尾

	commandWaitAdd(game, nil, nums(501, 0)) // N <= 0 → no-op
	commandWaitAdd(game, nil, nums(999, 1)) // 顧客資料不存在 → no-op
	commandWaitAdd(game, nil, nums(501))    // 缺 N → no-op
	commandWaitAdd(game, nil, nil)          // 缺顧客編號 → no-op
	this.Len(game.Wait, 3)
}

// === 測試輔助(置尾) ===

// sourceCard 建構複製來源卡牌(卡 103): 費用改 9、實例效果列表替換為指定編號(覆蓋技能載入的 401, 402)。
func sourceCard(game *cores.Game, effectID int32) *cores.Card {
	source := cores.NewCard(game, 103)
	source.GetCost().Set(9)
	source.GetEffectID().DelAll(401)
	source.GetEffectID().DelAll(402)
	source.GetEffectID().Add(effectID)
	return source
}
