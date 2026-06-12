package rodi

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/yinweli/RovingDiner/internal/games"
	"github.com/yinweli/RovingDiner/internal/tester"
)

func TestSuiteStepper(t *testing.T) {
	suite.Run(t, new(SuiteStepper))
}

// SuiteStepper 驗證暫停機橋接器(stepper.go): 雙跑同序(交棒不丟行組、不改序、不死鎖)、終局成敗、
// 終局後防呆、答覆會合、presenter 交棒半身。
type SuiteStepper struct {
	suite.Suite
}

// TestNewStepper 驗證交棒全鏈: 同 seed + 同關卡 + 同輸入, 同步直跑(RecordPresenter + FakeOperator)與
// 逐拍 Next / 被動答覆收集的行組序列完全相等(輸入請求路徑走真 channel 會合), 且終局回報正確成敗
// (601 生氣離場清空 → 成功、603 高耐心耗到回合上限 → 失敗)。
func (this *SuiteStepper) TestNewStepper() {
	record := &tester.RecordPresenter{}
	succ := games.Run(1, 601, tester.BuildSheet(), tester.FakeOperator{}, record)
	line, bridged := drain(newStepper(1, 601, tester.BuildSheet()))
	this.Equal(succ, bridged)
	this.True(bridged)
	this.Equal(record.Line, line)

	record = &tester.RecordPresenter{}
	succ = games.Run(1, 603, tester.BuildSheet(), tester.FakeOperator{}, record)
	line, bridged = drain(newStepper(1, 603, tester.BuildSheet()))
	this.Equal(succ, bridged)
	this.False(bridged)
	this.Equal(record.Line, line)
}

// TestStepperNext 驗證終局後防呆: 排水到終局後再 Next, 直接回終局輪次(不再放行, 不死鎖)。
func (this *SuiteStepper) TestStepperNext() {
	target := newStepper(1, 601, tester.BuildSheet())
	drain(target)
	next := target.Next()
	this.Equal(turnOver, next.role)
	this.True(next.succ)
}

// TestStepperAnswer 驗證答覆會合: 答案送進請求的答覆通道(緩衝 1 不阻塞)後收下一輪; 終局輪次記錄 over / succ。
func (this *SuiteStepper) TestStepperAnswer() {
	target := &stepper{turn: make(chan turn, 1)}
	req := &request{answer: make(chan answer, 1)}
	target.turn <- turn{role: turnLine, line: []string{"x"}}
	next := target.Answer(req, answer{})
	this.Equal(turnLine, next.role)
	this.Equal([]string{"x"}, next.line)
	this.Equal(answer{}, <-req.answer)
	this.False(target.over)

	req = &request{answer: make(chan answer, 1)}
	target.turn <- turn{role: turnOver, succ: true}
	next = target.Answer(req, answer{})
	this.Equal(turnOver, next.role)
	this.True(target.over)
	this.True(target.succ)
}

// TestGatePresenterEmit 驗證交棒半身: Emit 把行組包成行組輪次原樣送進 channel 後消費一拍放行。
func (this *SuiteStepper) TestGatePresenterEmit() {
	bus := make(chan turn, 1)
	gate := make(chan struct{}, 1)
	gate <- struct{}{} // 預放行: Emit 送完行組即取走、不停棒
	gatePresenter{turn: bus, gate: gate}.Emit("[R3 回合開始] 執行結算", "$ x")
	next := <-bus
	this.Equal(turnLine, next.role)
	this.Equal([]string{"[R3 回合開始] 執行結算", "$ x"}, next.line)
	this.Empty(gate) // 放行訊號已被消費
}

// drain 排水收集: 逐拍 Next 收行組、輸入請求以候選前綴立答, 到終局為止, 回傳行組序列與成敗。
func drain(target *stepper) (line [][]string, succ bool) {
	next := target.Next()

	for {
		switch next.role {
		case turnLine:
			line = append(line, next.line)
			next = target.Next()

		case turnRequest:
			next = target.Answer(next.req, prefixAnswer(next.req))

		case turnOver:
			return line, next.succ
		} // switch
	} // for
}

// prefixAnswer 候選前綴答覆(排水用): 玩家行動 = 結束、選取 = 候選前綴, 與雙跑基準 tester.FakeOperator 同形。
func prefixAnswer(req *request) answer {
	if req.guest != nil {
		return answer{guest: req.guest[:req.count]}
	} // if

	if req.card != nil {
		return answer{card: req.card[:req.count]}
	} // if

	return answer{}
}
