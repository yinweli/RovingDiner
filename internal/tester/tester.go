// Package tester 提供 cores / rules / games 測試共用的基建：迷你靜態表、遊戲資料包裝與決定性替身。
// 只使用 cores 公開介面（cores 自身的白箱測試另持私有複本）。
package tester

import (
	"github.com/yinweli/RovingDiner/internal/cores"
	sheeter "github.com/yinweli/RovingDiner/sheet"
)

// BuildSheet 組裝測試用 in-memory 靜態表:設定六鍵、3 座位(桌1:座1,2;桌2:座3)、3 卡牌(群組 1 / 2)、效果群組 5 / 6、
// 技能 301、抽獎群組 7 / 8 / 9、顧客 501、關卡 601(完整開局)/ 602(壞引用)。各層測試共用同一份迷你表,測試只關注自己用到的列。
func BuildSheet() *sheeter.Sheeter {
	data := &sheeter.Sheeter{}
	data.Setting.Data = map[string]*sheeter.Setting{
		"RoundMax":  {ID: "RoundMax", Value: []string{"12"}},
		"Morale":    {ID: "Morale", Value: []string{"30"}},
		"MoraleMax": {ID: "MoraleMax", Value: []string{"50"}},
		"EnergyMax": {ID: "EnergyMax", Value: []string{"3"}},
		"HandMax":   {ID: "HandMax", Value: []string{"10"}},
		"DrawMax":   {ID: "DrawMax", Value: []string{"5"}},
	}
	data.Seat.Data = map[int32]*sheeter.Seat{
		1: {ID: 1, TableID: 1, SameSeatID: []int32{1, 2}, NearSeatID: []int32{3}},
		2: {ID: 2, TableID: 1, SameSeatID: []int32{1, 2}, NearSeatID: []int32{3}},
		3: {ID: 3, TableID: 2, SameSeatID: []int32{3}, NearSeatID: []int32{1, 2}},
	}
	data.Card.Data = map[int32]*sheeter.Card{
		101: {ID: 101, Group: 1},
		102: {ID: 102, Group: 2},
		103: {ID: 103, Group: 1, Cost: 2, Keep: true, Seal: true, SkillID: 301}, // bool 欄 → 鎖、SkillID → 效果列表
	}
	data.Effect.Data = map[int32]*sheeter.Effect{
		201: {ID: 201, Group: 5},
		401: {ID: 401, Group: 5, RunOrder: 10, RunRound: 2}, // RunRound>0 → Expire = 建立回合 + 1；RunOrder 10
		402: {ID: 402, Group: 5, RunOrder: 10, RunRound: 0}, // RunRound 0 → Expire 0；與 401 同序、EffectID 402 > 401
		403: {ID: 403, Group: 6, RunOrder: 20, RunRound: 1}, // RunOrder 20 最大 → 排序最先
	}
	data.Skill.Data = map[int32]*sheeter.Skill{
		301: {ID: 301, Group: 3, EffectID: []int32{401, 402}}, // 卡 103 的技能（群組 3）效果列表
	}
	data.Award.Data = map[int32]*sheeter.Award{
		1: {ID: 1, Group: 7, CardID: 101, Weight: 3}, // 群組 7：候選 101(w3) / 102(w1)
		2: {ID: 2, Group: 7, CardID: 102, Weight: 1},
		3: {ID: 3, Group: 8, CardID: 101, Weight: 0}, // 群組 8：權重 0 → roll no-op
		4: {ID: 4, Group: 9, CardID: 999, Weight: 1}, // 群組 9：抽中編號 999 無卡牌資料 → 跳過該張
	}
	data.Guest.Data = map[int32]*sheeter.Guest{
		501: {ID: 501, Score: 0, ScoreMax: 10, Morale: 5, MoraleMax: 8, Calm: 3, SateMax: 12, SateSeal: true, SateSkillID: []string{"9^301", "6^301"}, CalmSkillID: []string{"1^301", "2^301"}},
		502: {ID: 502, Score: 2, ScoreMax: 10, Morale: 5, MoraleMax: 8, Calm: 99, SateMax: 12}, // 高耐心(供回合上限失敗局)、無門檻
	}
	data.Stage.Data = map[int32]*sheeter.Stage{
		601: {ID: 601, Name: "測試關卡", HandID: []int32{101}, DeckID: []int32{101, 102}, DropID: []int32{102}, ExileID: []int32{101}, WaitID: []int32{501, 501}, PrefixSkillID: []int32{301}},
		602: {ID: 602, Name: "壞引用", HandID: []int32{999}, DeckID: []int32{999}, WaitID: []int32{999}, PrefixSkillID: []int32{999}}, // 全列查無資料 → 逐筆跳過
		603: {ID: 603, Name: "耗到上限", WaitID: []int32{502}},                                                                         // 顧客耐心耗不完 → 回合上限失敗
	}
	return data
}

// BuildData 以迷你靜態表組裝遊戲資料（無命令編譯;自訂編譯效果以 Data.SetEffect 注入）。
func BuildData() *cores.Data {
	return cores.NewData(BuildSheet(), nil)
}

// FakeOperator 決定性玩家輸入替身:Pick 取候選前綴;EmptyPick 為真時 PickGuest 回空(驗證防禦分支)。
type FakeOperator struct {
	EmptyPick bool
}

func (this FakeOperator) PlayerAction(game *cores.Game) *cores.Card {
	return nil // 恆回玩家結束;出牌序列由各測試自備腳本 Operator
}

func (this FakeOperator) PickGuest(source []*cores.Guest, count int) []*cores.Guest {
	if this.EmptyPick {
		return nil
	} // if

	return source[:count]
}

func (this FakeOperator) PickCard(source []*cores.Card, count int) []*cores.Card {
	return source[:count]
}

// PickDiscard 良性版棄牌選擇——棄掉候選的前 over 張;行為不良分支由各測試自備替身驗證。
func (this FakeOperator) PickDiscard(source []*cores.Card, over int) []*cores.Card {
	return source[:over]
}

// FakeRander 決定性亂數替身:Intn 固定回 N(預設 0、取候選首位);Shuffle 恆等(randSubset 取候選前綴,結果可預期);Weighted 恆取首位。
type FakeRander struct {
	N int
}

func (this FakeRander) Intn(n int) int {
	return this.N
}

func (this FakeRander) Shuffle(n int, swap func(i, j int)) {
	if n > 0 {
		swap(0, 0) // 恆等交換:執行 caller 的 swap 閉包(覆蓋),但不改變順序,使 randSubset / deckTop 結果可預期
	} // if
}

func (this FakeRander) Weighted(weight []int32) int {
	return 0
}
