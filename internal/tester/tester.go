// Package tester 提供 cores / rules / games 測試共用的基建: 迷你靜態表、遊戲資料包裝與決定性替身。
// 只使用 cores 公開介面(cores 自身的白箱測試另持私有複本)。
package tester

import (
	"github.com/yinweli/RovingDiner/internal/cores"
	sheeter "github.com/yinweli/RovingDiner/sheet"
)

// BuildSheet 組裝測試用 in-memory 靜態表: 設定六鍵、3 座位(桌1:座1,2; 桌2:座3)、卡牌(101~103 單元素材 / 104~106 腳本劇情)、
// 效果群組 5 / 6 與腳本劇情效果 405~409、技能 301~304、抽獎群組 7 / 8 / 9、顧客 501 / 502、
// 關卡 601(完整開局)/ 602(壞引用)/ 603(耗到上限)/ 604(腳本劇情)。各層測試共用同一份迷你表, 測試只關注自己用到的列。
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
		104: {ID: 104, Group: 2, Cost: 1, SkillID: 302},                         // 腳本劇情: 新選顧客餵食
		105: {ID: 105, Group: 2, SkillID: 303},                                  // 腳本劇情: 新選手牌降費 + 爆抽棄牌
		106: {ID: 106, Group: 2, SkillID: 304},                                  // 腳本劇情: 掛觸發 + 常駐入列
	}
	data.Effect.Data = map[int32]*sheeter.Effect{
		201: {ID: 201, Group: 5},
		401: {ID: 401, Group: 5, RunOrder: 10, RunRound: 2},                                                      // RunRound>0 → Expire = 建立回合 + 1; RunOrder 10
		402: {ID: 402, Group: 5, RunOrder: 10, RunRound: 0},                                                      // RunRound 0 → Expire 0; 與 401 同序、EffectID 402 > 401
		403: {ID: 403, Group: 6, RunOrder: 20, RunRound: 1},                                                      // RunOrder 20 最大 → 排序最先
		405: {ID: 405, TargetKind: 1, TargetCount: 1, CommandImmed: "self.sate += 6"},                            // 立即 + 新選顧客(玩家選取)
		406: {ID: 406, TargetKind: 4, TargetCount: 1, CommandImmed: "cardCostAdd(self, -1)"},                     // 立即 + 新選手牌(玩家選取)
		407: {ID: 407, CommandImmed: "deckToHand(deckTop[7], false)"},                                            // 立即爆抽 → 手牌超上限 → 結算棄牌選取
		408: {ID: 408, Kind: 1, TriggerKind: "roundStart", CommandTrigger: "morale += 1"},                        // 觸發入列: 回合開始 + 恆成立
		409: {ID: 409, Kind: 2, RunRound: 2, CommandStart: "moraleShield += 2", CommandEnd: "moraleShield -= 2"}, // 常駐入列: 啟動與結束命令成對
	}
	data.Skill.Data = map[int32]*sheeter.Skill{
		301: {ID: 301, Group: 3, EffectID: []int32{401, 402}}, // 卡 103 的技能(群組 3)效果列表
		302: {ID: 302, Group: 3, EffectID: []int32{405}},      // 卡 104 的技能: 選顧客餵食
		303: {ID: 303, Group: 3, EffectID: []int32{406, 407}}, // 卡 105 的技能: 選手牌降費 → 爆抽
		304: {ID: 304, Group: 3, EffectID: []int32{408, 409}}, // 卡 106 的技能: 觸發 + 常駐入列
	}
	data.Award.Data = map[int32]*sheeter.Award{
		1: {ID: 1, Group: 7, CardID: 101, Weight: 3}, // 群組 7:候選 101(w3) / 102(w1)
		2: {ID: 2, Group: 7, CardID: 102, Weight: 1},
		3: {ID: 3, Group: 8, CardID: 101, Weight: 0}, // 群組 8:權重 0 → roll no-op
		4: {ID: 4, Group: 9, CardID: 999, Weight: 1}, // 群組 9:抽中編號 999 無卡牌資料 → 跳過該張
	}
	data.Guest.Data = map[int32]*sheeter.Guest{
		501: {ID: 501, Score: 0, ScoreMax: 10, Morale: 5, MoraleMax: 8, Calm: 3, SateMax: 12, SateSeal: true, SateSkillID: []string{"9^301", "6^301"}, CalmSkillID: []string{"1^301", "2^301"}},
		502: {ID: 502, Score: 2, ScoreMax: 10, Morale: 5, MoraleMax: 8, Calm: 99, SateMax: 12}, // 高耐心(供回合上限失敗局)、無門檻
	}
	data.Stage.Data = map[int32]*sheeter.Stage{
		601: {ID: 601, Name: "測試關卡", HandID: []int32{101}, DeckID: []int32{101, 102}, DropID: []int32{102}, ExileID: []int32{101}, WaitID: []int32{501, 501}, PrefixSkillID: []int32{301}},
		602: {ID: 602, Name: "壞引用", HandID: []int32{999}, DeckID: []int32{999}, WaitID: []int32{999}, PrefixSkillID: []int32{999}},                                 // 全列查無資料 → 逐筆跳過
		603: {ID: 603, Name: "耗到上限", WaitID: []int32{502}},                                                                                                         // 顧客耐心耗不完 → 回合上限失敗
		604: {ID: 604, Name: "腳本劇情", HandID: []int32{104, 105, 106}, DeckID: []int32{101, 102, 101, 102, 101, 102, 101, 102, 101, 101}, WaitID: []int32{501, 501}}, // ScriptOperator 腳本局(出牌 / 三類選取 / 觸發常駐入列)
	}
	return data
}

// BuildData 以迷你靜態表組裝遊戲資料(無命令編譯; 自訂編譯效果以 Data.SetEffect 注入)。
func BuildData() *cores.Data {
	return cores.NewData(BuildSheet(), nil)
}

// FakeOperator 決定性玩家輸入替身: Pick 取候選前綴(提示前文不消費——替身無顯示);
// EmptyPick 為真時 PickGuest 回空(驗證防禦分支)。
type FakeOperator struct {
	EmptyPick bool
}

func (this FakeOperator) PlayerAction(game *cores.Game) *cores.Card {
	return nil // 恆回玩家結束; 出牌序列見 ScriptOperator
}

func (this FakeOperator) PickGuest(prompt string, source []*cores.Guest, count int) []*cores.Guest {
	if this.EmptyPick {
		return nil
	} // if

	return source[:count]
}

func (this FakeOperator) PickCard(prompt string, source []*cores.Card, count int) []*cores.Card {
	return source[:count]
}

// PickDiscard 良性版棄牌選擇——棄掉候選的前 over 張; 行為不良分支由各測試自備替身驗證。
func (this FakeOperator) PickDiscard(prompt string, source []*cores.Card, over int) []*cores.Card {
	return source[:over]
}

// ScriptOperator 腳本化玩家輸入替身: 各決策點依佇列消費——出牌每玩家行動階段一筆(筆內依序以編號取首張同號手牌,
// 查無略過; 筆盡即該階段玩家結束), 選取每次一筆(筆內為候選索引, 越界略過、空筆 = 棄選)。
// 佇列耗盡回退 FakeOperator 行為(出牌結束 / 取候選前綴), 腳本有限故一局必然收斂; 同 seed 同腳本 = 同行序列
// (conformance golden 的輸入腳本維度,【營業實作規格書 | 七、測試策略 | 7.3】)。須以指標使用(消費推進佇列)。
type ScriptOperator struct {
	Play    [][]int32 // 出牌腳本(每玩家行動階段一筆; 筆內為卡牌編號序列)
	Guest   [][]int   // 選顧客腳本(每次選取一筆; 筆內為候選索引)
	Card    [][]int   // 選卡牌腳本(同上)
	Discard [][]int   // 棄牌腳本(同上)
}

func (this *ScriptOperator) PlayerAction(game *cores.Game) *cores.Card {
	for len(this.Play) > 0 {
		head := this.Play[0]

		if len(head) == 0 {
			this.Play = this.Play[1:] // 本階段筆耗盡 → 玩家結束, 後續筆留給下個階段
			return nil
		} // if

		this.Play[0] = head[1:]

		for _, itor := range game.Hand {
			if itor.GetCardID() == head[0] {
				return itor
			} // if
		} // for
	} // for

	return nil
}

func (this *ScriptOperator) PickGuest(prompt string, source []*cores.Guest, count int) []*cores.Guest {
	if len(this.Guest) == 0 {
		return source[:count]
	} // if

	head := this.Guest[0]
	this.Guest = this.Guest[1:]
	return pickIndex(source, head)
}

func (this *ScriptOperator) PickCard(prompt string, source []*cores.Card, count int) []*cores.Card {
	if len(this.Card) == 0 {
		return source[:count]
	} // if

	head := this.Card[0]
	this.Card = this.Card[1:]
	return pickIndex(source, head)
}

func (this *ScriptOperator) PickDiscard(prompt string, source []*cores.Card, over int) []*cores.Card {
	if len(this.Discard) == 0 {
		return source[:over]
	} // if

	head := this.Discard[0]
	this.Discard = this.Discard[1:]
	return pickIndex(source, head)
}

// pickIndex 依索引列表取候選元素(腳本選取共用): 結果順序依索引列表, 越界索引略過。
func pickIndex[T any](source []T, index []int) (result []T) {
	for _, itor := range index {
		if itor >= 0 && itor < len(source) {
			result = append(result, source[itor])
		} // if
	} // for

	return result
}

// FakeRander 決定性亂數替身: Intn 固定回 N(預設 0、取候選首位); Shuffle 恆等(randSubset 取候選前綴, 結果可預期); Weighted 恆取首位。
type FakeRander struct {
	N int
}

func (this FakeRander) Intn(n int) int {
	return this.N
}

func (this FakeRander) Shuffle(n int, swap func(i, j int)) {
	if n > 0 {
		swap(0, 0) // 恆等交換: 執行 caller 的 swap 閉包(覆蓋), 但不改變順序, 使 randSubset / deckTop 結果可預期
	} // if
}

func (this FakeRander) Weighted(weight []int32) int {
	return 0
}

// RecordPresenter 日誌流錄製器: Emit 逐拍依序追加進 Line 切片(一拍一組行), 供發射點測試與 conformance golden 比對
// (【營業實作規格書 | 七、測試策略】L3 基建); 須以指標使用(Emit 寫入切片)。
type RecordPresenter struct {
	Line [][]string // 錄得行組(發射序; 一拍一組)
}

func (this *RecordPresenter) Emit(line ...string) {
	this.Line = append(this.Line, line)
}

// Flat 攤平行組為行序列(golden 比對素材; 拍的分組是節奏、不入 golden)。
func (this *RecordPresenter) Flat() (result []string) {
	for _, itor := range this.Line {
		result = append(result, itor...)
	} // for

	return result
}
