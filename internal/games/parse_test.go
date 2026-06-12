package games

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/yinweli/RovingDiner/internal/cores"
	"github.com/yinweli/RovingDiner/internal/exprs"
)

func TestSuiteParse(t *testing.T) {
	suite.Run(t, new(SuiteParse))
}

// SuiteParse 驗證命令文法解析(parse.go): 屬性修改命令(全域 / 引用左值、賦值符、鎖定 / 解鎖、右值算術式)、
// 操作命令(命令對象、[...] 參數、varargs、巢狀函式參數)、空白容錯, 以及語法錯誤(含出錯位置)。
// 名稱以原始字串擷取、不驗成員(成員合法性屬 Validate); AST 以 S-運算式字串比對。
type SuiteParse struct {
	suite.Suite
}

func (this *SuiteParse) TestParseAssign() {
	this.Equal("(set morale 10)", this.parse("morale = 10"))
	this.Equal("(add score 5)", this.parse("score += 5"))
	this.Equal("(sub morale 3)", this.parse("morale -= 3"))
	this.Equal("(mul energy 2)", this.parse("energy *= 2"))
	this.Equal("(div drawMax 2)", this.parse("drawMax /= 2"))
	this.Equal("(mod round 2)", this.parse("round %= 2"))
}

func (this *SuiteParse) TestParseAssignLock() {
	this.Equal("(lock energy)", this.parse("energy @"))   // @ 不帶算術式
	this.Equal("(unlock energy)", this.parse("energy #")) // # 不帶算術式
	this.Equal("(lock drawLast.cardSeal)", this.parse("drawLast.cardSeal @"))
}

func (this *SuiteParse) TestParseAssignRef() {
	this.Equal("(set self.calm 7)", this.parse("self.calm = 7"))
	this.Equal("(add self.sate 1)", this.parse("self.sate += 1"))
	this.Equal("(add damageGuest.sate <expr>)", this.parse("damageGuest.sate += damageValue")) // 右值含全域屬性

	command := this.assign("self.calm = 7")
	this.Equal("self", command.base)
	this.Equal("calm", command.refAttr)
	this.True(command.isRef)
}

func (this *SuiteParse) TestParseAssignExpr() {
	// 右值算術式委由 exprs 解析; 求值確認運算子與優先序被正確擷取
	command := this.assign("morale = 1 + 2 * 3")
	this.Equal("morale", command.base)
	this.Equal(cores.AssignSet, command.op)
	this.False(command.isRef)

	value, ok := command.value.Eval(exprs.Env{})
	this.Require().True(ok)
	this.Equal(7.0, value.Num())
}

func (this *SuiteParse) TestParseOperate() {
	this.Equal("(op handAdd none 10031 1)", this.parse("handAdd(none, 10031, 1)"))
	this.Equal("(op phaseJump none '玩家行動')", this.parse("phaseJump(none, '玩家行動')")) // 字串字面值參數
	this.Equal("(op guestSeat none)", this.parse("guestSeat(none)"))                // 僅命令對象、無其餘參數
	this.Equal("(op cardRun self true false)", this.parse("cardRun(self, true, false)"))

	command := this.operate("deckToHand(deckTop[2], false)")
	this.Equal("deckToHand", command.verb)
	this.Equal("deckTop", command.selector.selector)
	this.Len(command.selector.param, 1)
	this.Len(command.arg, 1)
}

func (this *SuiteParse) TestParseVararg() {
	this.Equal("(op handCopy handPick[1 0] 3 20042 20055)", this.parse("handCopy(handPick[1, 0], 3, 20042, 20055)")) // 尾端 varargs

	// 巢狀函式參數內部的逗號不可被誤判為外層分隔: min(10, 3) 須為單一參數
	command := this.operate("deckAdd(none, min(10, 3), 1)")
	this.Len(command.arg, 2)
	this.Equal("(op deckAdd none <expr> 1)", commandString(command))
}

func (this *SuiteParse) TestParseSelectorParam() {
	this.Equal("(op deckToHand deckTop[2] false)", this.parse("deckToHand(deckTop[2], false)"))
	this.Equal("(op cardEffectAdd handPick[1 0] 20042)", this.parse("cardEffectAdd(handPick[1, 0], 20042)"))
	this.Equal("(op deckToHand deckTop[<expr>] false)", this.parse("deckToHand(deckTop[drawMax - 2], false)")) // 命令對象參數含全域屬性
	this.Equal("(op deckShuffle none)", this.parse("deckShuffle(none)"))                                       // 無 [...] 參數
	this.Equal("(op deckToHand deckTop[] false)", this.parse("deckToHand(deckTop[], false)"))                  // 空中括號(數量驗證留待 M8)
}

func (this *SuiteParse) TestParseWhitespace() {
	this.Equal("(set morale 10)", this.parse("   morale   =   10   "))
	this.Equal("(op cardRun self true false)", this.parse("cardRun ( self , true , false )"))
	this.Equal("(set self.calm 7)", this.parse(" self . calm = 7 "))
}

func (this *SuiteParse) TestParseError() {
	source := []string{
		"",                             // 空輸入
		"   ",                          // 全空白
		"123",                          // 數字開頭非識別子
		"morale",                       // 缺賦值符
		"morale +",                     // '+' 非賦值符(缺 =)
		"morale = ",                    // 缺右值算術式
		"morale @ 5",                   // 鎖定符後不可再帶內容
		"self. = 1",                    // '.' 後缺引用屬性名
		"morale = 1 +",                 // 右值算術式不完整
		"morale = 1 2",                 // 右值結尾多餘內容
		"deckToHand(self",              // 主括號未閉合
		"deckToHand()",                 // 命令對象缺漏
		"deckToHand(self,)",            // 尾隨逗號缺參數
		"deckToHand(deckTop[2] false)", // 參數缺逗號分隔
		"deckToHand(deckTop[2)",        // 命令對象中括號未閉合
		"min(1, 2)",                    // 命令對象需為識別子, 數字非法
		"handAdd(none, 1 +, 2)",        // 參數算術式不完整(內嵌 exprs 錯誤上拋)
		"deckToHand(deckTop[1 +], 0)",  // 命令對象參數算術式不完整
		"phaseJump(none, 'oops)",       // 參數字串未結束(掃描至結尾仍委由 exprs 報錯)
	}

	for _, itor := range source {
		_, err := Parse(itor)
		this.Require().Error(err, itor)
	} // for
}

func (this *SuiteParse) TestParseErrorPosition() {
	this.Equal(0, this.errorPos(""))                       // 開頭即缺識別子
	this.Equal(11, this.errorPos("morale = 1 2"))          // 右值多餘內容, 定位回算至第二個 token
	this.Equal(15, this.errorPos("deckToHand(self"))       // 缺 ) , 定位於來源結尾
	this.Equal(15, this.errorPos("deckToHand(self.)"))     // 命令對象不可為引用, 停在 '.'
	this.Equal(17, this.errorPos("handAdd(none, 1 +, 2)")) // 參數內嵌 exprs 錯誤, 偏移回算至命令來源
}

// parse 解析 source 並回傳其 AST 的 S-運算式字串(解析失敗即 fail 測試)。
func (this *SuiteParse) parse(source string) string {
	command, err := Parse(source)
	this.Require().NoError(err, source)
	return commandString(command)
}

// assign 解析 source 並斷言為屬性修改命令, 回傳其結構(供右值 / 欄位斷言)。
func (this *SuiteParse) assign(source string) commandAssign {
	command, err := Parse(source)
	this.Require().NoError(err, source)
	result, ok := command.(commandAssign)
	this.Require().True(ok, source)
	return result
}

// operate 解析 source 並斷言為操作命令, 回傳其結構(供命令對象 / 參數斷言)。
func (this *SuiteParse) operate(source string) commandOperate {
	command, err := Parse(source)
	this.Require().NoError(err, source)
	result, ok := command.(commandOperate)
	this.Require().True(ok, source)
	return result
}

// errorPos 解析 source(預期失敗)並回傳 SyntaxError 的位置。
func (this *SuiteParse) errorPos(source string) int {
	_, err := Parse(source)
	this.Require().Error(err, source)
	syntaxError, ok := err.(*exprs.SyntaxError)
	this.Require().True(ok, source)
	return syntaxError.Pos
}

// commandString 把命令 AST 轉成 S-運算式字串, 便於結構比對。
func commandString(command Command) string {
	switch c := command.(type) {
	case commandAssign:
		lvalue := c.base

		if c.isRef {
			lvalue += "." + c.refAttr
		} // if

		if c.value == nil {
			return "(" + assignTag(c.op) + " " + lvalue + ")"
		} // if

		return "(" + assignTag(c.op) + " " + lvalue + " " + exprText(c.value) + ")"

	case commandOperate:
		out := "(op " + c.verb + " " + selectorString(c.selector)

		for _, itor := range c.arg {
			out += " " + exprText(itor)
		} // for

		return out + ")"

	default:
		return "?"
	} // switch
}

// selectorString 把命令對象轉成 名稱[參數...] 字串(無 [...] 時僅名稱)。
func selectorString(selector selectorArg) string {
	out := selector.selector

	if selector.param != nil {
		out += "["

		for i, itor := range selector.param {
			if i > 0 {
				out += " "
			} // if

			out += exprText(itor)
		} // for

		out += "]"
	} // if

	return out
}

// assignTag 把賦值符轉成顯示用標記。
func assignTag(op cores.AssignKind) string {
	switch op {
	case cores.AssignSet:
		return "set"

	case cores.AssignAdd:
		return "add"

	case cores.AssignSub:
		return "sub"

	case cores.AssignMul:
		return "mul"

	case cores.AssignDiv:
		return "div"

	case cores.AssignMod:
		return "mod"

	case cores.AssignLock:
		return "lock"

	case cores.AssignUnlock:
		return "unlock"

	default:
		return "?"
	} // switch
}

// exprText 把內嵌算術式以求值結果呈現: 常數以其值顯示、需 Resolver 者(求值失敗)以 <expr> 表示。
func exprText(expr *exprs.Expr) string {
	if expr == nil {
		return "-"
	} // if

	value, ok := expr.Eval(exprs.Env{})

	if ok == false {
		return "<expr>"
	} // if

	switch {
	case value.IsNum():
		return strconv.FormatFloat(value.Num(), 'g', -1, 64)

	case value.IsBool():
		return strconv.FormatBool(value.Bool())

	case value.IsText():
		return "'" + value.Text() + "'"

	case value.IsNone():
		return "none"

	default:
		return "<expr>"
	} // switch
}
