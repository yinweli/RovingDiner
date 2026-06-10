package rules

import (
	"github.com/yinweli/RovingDiner/internal/cores"
)

// 終止階段（【營業規格書 | 十九、核心流程 | 6. 營業成功階段 / 7. 營業失敗階段】）:皆由執行結算（M16）觸發進入;
// 觸發對應時機後回 PhaseNone 作停機訊號,成功 / 失敗由停在哪一站得知（games.Run 屆時轉對外結果）。

// phaseGameSucc 營業成功階段:觸發 gameSucc、終止營業。
func phaseGameSucc(game *cores.Game) cores.PhaseKind {
	fireTrigger(game, cores.TriggerGameSucc) // 營業成功觸發
	return cores.PhaseNone
}

// phaseGameFail 營業失敗階段:觸發 gameFail、終止營業。
func phaseGameFail(game *cores.Game) cores.PhaseKind {
	fireTrigger(game, cores.TriggerGameFail) // 營業失敗觸發
	return cores.PhaseNone
}
