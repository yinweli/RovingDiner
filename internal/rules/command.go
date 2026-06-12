package rules

import (
	"github.com/yinweli/RovingDiner/internal/cores"
	"github.com/yinweli/RovingDiner/internal/exprs"
)

// HasCommand 回報操作命令詞彙表是否登錄 name; 供 games.Validate 檢查命令 verb。
func HasCommand(name string) bool {
	_, ok := command[name]
	return ok
}

// command 操作命令詞彙表(名稱 → 執行行為); 對應【營業規格書 | 二十五、操作命令清單】。
// 每一詞條對應一個獨立的 command* 具名函式(比照讀寫 / 命令對象詞彙表); 詞條依類別分檔
// (commandMove / commandCard / commandEffect / commandGuest / commandInstance / commandFlow.go), 本檔僅持型別、註冊表與 bootstrap 詞條。
var command = map[string]cores.CommandFunc{
	// bootstrap(command.go)
	"phaseJump":   commandPhaseJump,
	"deckShuffle": commandDeckShuffle,

	// 容器搬移(commandMove.go)
	"handToDeck":  commandHandToDeck,
	"handToDrop":  commandHandToDrop,
	"handToExile": commandHandToExile,
	"deckToDrop":  commandDeckToDrop,
	"deckToExile": commandDeckToExile,
	"deckToHand":  commandDeckToHand,
	"dropToDeck":  commandDropToDeck,
	"dropToExile": commandDropToExile,
	"dropToHand":  commandDropToHand,
	"exileToDeck": commandExileToDeck,
	"exileToDrop": commandExileToDrop,
	"exileToHand": commandExileToHand,

	// 卡牌屬性(commandCard.go)
	"cardCostAdd":      commandCardCostAdd,
	"cardCostMul":      commandCardCostMul,
	"cardCostSet":      commandCardCostSet,
	"cardEffectAdd":    commandCardEffectAdd,
	"cardEffectDel":    commandCardEffectDel,
	"cardEffectDelAll": commandCardEffectDelAll,

	// 效果佇列(commandEffect.go)
	"effectClear": commandEffectClear,
	"effectDel":   commandEffectDel,
	"effectRun":   commandEffectRun,

	// 顧客免疫 / 行動(commandGuest.go)
	"effectImmuneAdd": commandEffectImmuneAdd,
	"effectImmuneDel": commandEffectImmuneDel,
	"skillImmuneAdd":  commandSkillImmuneAdd,
	"skillImmuneDel":  commandSkillImmuneDel,
	"taskAdd":         commandTaskAdd,

	// 實例化(commandInstance.go)
	"handAdd":    commandHandAdd,
	"deckAdd":    commandDeckAdd,
	"dropAdd":    commandDropAdd,
	"exileAdd":   commandExileAdd,
	"handRoll":   commandHandRoll,
	"deckRoll":   commandDeckRoll,
	"dropRoll":   commandDropRoll,
	"exileRoll":  commandExileRoll,
	"handCopy":   commandHandCopy,
	"deckCopy":   commandDeckCopy,
	"dropCopy":   commandDropCopy,
	"exileCopy":  commandExileCopy,
	"handClone":  commandHandClone,
	"deckClone":  commandDeckClone,
	"dropClone":  commandDropClone,
	"exileClone": commandExileClone,
	"guestSpawn": commandGuestSpawn,
	"waitAdd":    commandWaitAdd,

	// 處理流程(commandFlow.go)
	"cardRun":     commandCardRun,
	"handMorph":   commandHandMorph,
	"deckMorph":   commandDeckMorph,
	"dropMorph":   commandDropMorph,
	"exileMorph":  commandExileMorph,
	"cardify":     commandCardify,
	"restore":     commandRestore,
	"guestExit":   commandGuestExit,
	"guestReturn": commandGuestReturn,
	"guestRoam":   commandGuestRoam,
	"guestSeat":   commandGuestSeat,
}

// commandPhaseJump 設下一階段(【營業規格書 | 二十五、操作命令清單 | phaseJump】);
// 階段名稱 ∈ {玩家行動, 顧客行動, 回合結束}(PhaseJumpLegal), 不在此集合 / 非字串 → no-op。命令對象固定 none、不使用 target。
func commandPhaseJump(game *cores.Game, target []cores.InstanceID, arg []exprs.Value) {
	if len(arg) < 1 || arg[0].IsText() == false {
		return
	} // if

	phase := cores.PhaseKind(arg[0].Text())

	if cores.PhaseJumpLegal[phase] == false {
		return
	} // if

	game.SetNextPhase(phase)
}

// commandDeckShuffle 將抽牌牌堆隨機洗牌(【營業規格書 | 二十五、操作命令清單 | deckShuffle】); 命令對象固定 none、無參數。
// 成員不變、僅順序重排, 故只發重整快照(<= 1 張無事不發; M21 拍板, 原 M18 靜默作廢)。
func commandDeckShuffle(game *cores.Game, target []cores.InstanceID, arg []exprs.Value) {
	shuffleCard(game, game.Deck)

	if len(game.Deck) > 1 {
		emitDeckOrder(game)
	} // if
}
