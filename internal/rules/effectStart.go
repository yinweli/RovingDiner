package rules

import (
	"github.com/yinweli/RovingDiner/internal/cores"
)

// runEffectList 啟動效果列表（【營業規格書 | 二十、獨立流程 | 啟動技能】）:逐效果選目標、對每個 self 依效果類型派發。
// effectID 為效果編號列表（卡牌實例效果列表 / 技能效果列表）;skillGroup 供 skillImmune 排除免疫顧客;沿用來源空集合起步。
func runEffectList(game *cores.Game, effectID []int32, skillGroup int32) {
	var inherit []cores.Ref

	for _, itor := range effectID {
		meta, ok := game.EffectData(itor)

		if ok == false {
			continue // 查無編譯資料 → 略過該效果
		} // if

		for _, self := range selectTargets(game, meta, skillGroup, &inherit) {
			dispatchEffect(game, meta, itor, self)
		} // for
	} // for
}

// dispatchEffect 對單一 self 依效果類型派發:立即跑立即命令、觸發 / 常駐入佇列、常駐再每增一層跑一次啟動命令。
func dispatchEffect(game *cores.Game, meta cores.EffectData, effectID int32, self cores.Ref) {
	restore := game.SetSelf(&self) // self 綁定:立即命令 / 觸發條件 / 啟動命令以此 self 求值

	defer restore()

	switch meta.Kind {
	case cores.EffectImmed:
		if condPass(game, meta.Cond) {
			runEffectExec(game, meta.Immed, 1) // 立即:條件成立執行立即命令一次（不入佇列）
		} // if

	case cores.EffectTrigger:
		effectStack(game, self, effectID, 0) // 入佇列;觸發命令留 fireTrigger

	case cores.EffectPersist:
		added := effectStack(game, self, effectID, 0)
		runEffectExec(game, meta.Start, added) // 常駐:啟動命令每增加一層執行一次
	} // switch
}

// selectTargets 依目標類型選出本次 self 集合（【營業規格書 | 二十、獨立流程 | 啟動技能】目標選取段）;
// inherit 為沿用來源（跨效果列表維護）:新選 / 隨機設值、無目標清空、沿用讀取。每個 self 之後各建一個效果（【九、目標數量】）。
func selectTargets(game *cores.Game, meta cores.EffectData, skillGroup int32, inherit *[]cores.Ref) []cores.Ref {
	result := []cores.Ref{}

	switch meta.TargetKind {
	case cores.TargetNone:
		*inherit = nil
		result = []cores.Ref{{}} // 無目標 → 一個空物件 self

	case cores.TargetGuestPick:
		result = setInherit(inherit, pickGuestSelf(game, skillGroup, meta.TargetCount, false))

	case cores.TargetGuestRand:
		result = setInherit(inherit, pickGuestSelf(game, skillGroup, meta.TargetCount, true))

	case cores.TargetCardPick:
		result = setInherit(inherit, pickCardSelf(game, meta.TargetCount, false))

	case cores.TargetCardRand:
		result = setInherit(inherit, pickCardSelf(game, meta.TargetCount, true))

	case cores.TargetGuestSame:
		if hasGuest(*inherit) {
			result = *inherit // 沿用顧客:繼承前一效果 self（沿用來源不變）
		} else {
			result = setInherit(inherit, pickGuestSelf(game, skillGroup, meta.TargetCount, false)) // 退化新選顧客
		} // if

	case cores.TargetCardSame:
		if hasCard(*inherit) {
			result = *inherit
		} else {
			result = setInherit(inherit, pickCardSelf(game, meta.TargetCount, false))
		} // if

	default:
		// 不可達:目標類型已於 prepareEffect 範圍檢查
	} // switch

	return result
}

// setInherit 設沿用來源並回傳同一集合（新選 / 隨機選取後共用）。
func setInherit(inherit *[]cores.Ref, target []cores.Ref) []cores.Ref {
	*inherit = target
	return target
}

// pickGuestSelf 自座位候選（排除 skillImmune）選 count 位顧客為 self;random 為系統隨機、否則暫停由玩家選。
func pickGuestSelf(game *cores.Game, skillGroup, count int32, random bool) []cores.Ref {
	return guestSelf(selectN(game, guestCandidate(game, skillGroup), count, random, game.GetOperator().PickGuest))
}

// pickCardSelf 自手牌候選選 count 張手牌為 self;random 為系統隨機、否則暫停由玩家選。
func pickCardSelf(game *cores.Game, count int32, random bool) []cores.Ref {
	return cardSelf(selectN(game, cardCandidate(game), count, random, game.GetOperator().PickCard))
}

// guestCandidate 取座位列表顧客（依座位編號序）、排除技能群組 skillGroup 免疫者（【二十、獨立流程 | 啟動技能】顧客類 filter）。
func guestCandidate(game *cores.Game, skillGroup int32) (result []*cores.Guest) {
	for _, itor := range game.Seat.Sorted() {
		if itor.GetSkillImmune().Get(skillGroup) == 0 {
			result = append(result, itor)
		} // if
	} // for

	return result
}

// cardCandidate 取手牌全部為候選（手牌類無額外 filter）。
func cardCandidate(game *cores.Game) []*cores.Card {
	return game.Hand
}

// selectN 依【營業規格書 | 九、目標數量】退化規則選 count 個:候選 ≤ count → 全取;否則 random 走 randSubset、新選交 pick（玩家）。
func selectN[T any](game *cores.Game, candidate []T, count int32, random bool, pick func([]T, int) []T) []T {
	if int32(len(candidate)) <= count {
		return candidate // 退化:候選不足 → 全取
	} // if

	if random {
		return randSubset(game, candidate, int(count))
	} // if

	return pick(candidate, int(count))
}

// guestSelf 把顧客列表包成 self 集合（每位顧客一個 self）。
func guestSelf(guest []*cores.Guest) (result []cores.Ref) {
	for _, itor := range guest {
		result = append(result, cores.NewRefGuest(itor))
	} // for

	return result
}

// cardSelf 把手牌列表包成 self 集合（每張手牌一個 self）。
func cardSelf(card []*cores.Card) (result []cores.Ref) {
	for _, itor := range card {
		result = append(result, cores.NewRefCard(itor))
	} // for

	return result
}

// hasGuest 回報沿用來源是否為非空的顧客集合（沿用顧客有效;否則退化新選,【八、目標類型】）。
func hasGuest(inherit []cores.Ref) bool {
	return len(inherit) > 0 && inherit[0].GetGuest() != nil
}

// hasCard 回報沿用來源是否為非空的手牌集合（沿用手牌有效;否則退化新選）。
func hasCard(inherit []cores.Ref) bool {
	return len(inherit) > 0 && inherit[0].GetCard() != nil
}
