package features

import (
	"sort"

	"github.com/yinweli/Project_005/internal/game"
	sheeter "github.com/yinweli/Project_005/sheet"
)

// NewDater 包裝表格資料並預建衍生索引。
func NewDater(s *sheeter.Sheeter) *Dater {
	this := &Dater{
		sheeter: s,
		award:   map[int32][]*sheeter.Award{},
	}

	for _, itor := range s.Award.Values() {
		this.award[itor.Group] = append(this.award[itor.Group], itor)
	} // for

	for _, itor := range this.award {
		sort.Slice(itor, func(i, j int) bool {
			return itor[i].ID < itor[j].ID
		})
	} // for

	return this
}

// Dater 以 *sheeter.Sheeter 為後端的唯讀靜態表格，實作 game.Dater。
// 載入後預建衍生索引（抽獎依群組聚合），核心拿到 ready-to-use 的資料。
type Dater struct {
	sheeter *sheeter.Sheeter
	award   map[int32][]*sheeter.Award // 抽獎群組編號 -> 該群組紀錄（依 ID 排序，確保決定性）
}

// Card 取得卡牌資料；不存在時回傳 nil。
func (this *Dater) Card(id int32) *sheeter.Card {
	return this.sheeter.Card.Get(id)
}

// Guest 取得顧客資料；不存在時回傳 nil。
func (this *Dater) Guest(id int32) *sheeter.Guest {
	return this.sheeter.Guest.Get(id)
}

// Skill 取得技能資料；不存在時回傳 nil。
func (this *Dater) Skill(id int32) *sheeter.Skill {
	return this.sheeter.Skill.Get(id)
}

// Effect 取得效果資料；不存在時回傳 nil。
func (this *Dater) Effect(id int32) *sheeter.Effect {
	return this.sheeter.Effect.Get(id)
}

// Seat 取得座位資料；不存在時回傳 nil。
func (this *Dater) Seat(id int32) *sheeter.Seat {
	return this.sheeter.Seat.Get(id)
}

// Setting 取得設定資料；不存在時回傳 nil。
func (this *Dater) Setting(id string) *sheeter.Setting {
	return this.sheeter.Setting.Get(id)
}

// Award 取得指定抽獎群組的全部紀錄（依 ID 排序）；無此群組時回傳 nil。
func (this *Dater) Award(group int32) []*sheeter.Award {
	return this.award[group]
}

// 編譯期確認 Dater 滿足 game.Dater 介面。
var _ game.Dater = (*Dater)(nil)
