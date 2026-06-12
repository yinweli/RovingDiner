package exprs

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

func TestSuiteNode(t *testing.T) {
	suite.Run(t, new(SuiteNode))
}

// SuiteNode 驗證 AST 節點(node.go): 各節點型別直接建構後經 evaluator 求值, 確認其形狀
// 正確驅動求值(與 parser_test 的「字串→AST」不同角度); 條件對象 / 函式節點以 stub 接縫求值。
type SuiteNode struct {
	suite.Suite
}

func (this *SuiteNode) TestNodeLiteral() {
	this.Equal(5.0, this.eval(nodeLiteral{value: NewNum(5)}, Env{}).Num())
}

func (this *SuiteNode) TestNodeUnary() {
	this.Equal(-5.0, this.eval(nodeUnary{op: tokenMinus, operand: nodeLiteral{value: NewNum(5)}}, Env{}).Num())
}

func (this *SuiteNode) TestNodeBinary() {
	node := nodeBinary{op: tokenPlus, lhs: nodeLiteral{value: NewNum(1)}, rhs: nodeLiteral{value: NewNum(2)}}
	this.Equal(3.0, this.eval(node, Env{}).Num())
}

func (this *SuiteNode) TestNodeTernary() {
	node := nodeTernary{cond: nodeLiteral{value: NewBool(true)}, then: nodeLiteral{value: NewNum(1)}, els: nodeLiteral{value: NewNum(2)}}
	this.Equal(1.0, this.eval(node, Env{}).Num()) // 條件真 → 取 then 分支、不評估 els
}

func (this *SuiteNode) TestNodeIdent() {
	env := Env{Resolver: &stubResolver{attr: map[string]Value{"morale": NewNum(5)}}}
	this.Equal(5.0, this.eval(nodeIdent{name: "morale"}, env).Num())
}

func (this *SuiteNode) TestNodeCall() {
	env := Env{Builtin: stubBuiltin()}
	node := nodeCall{name: "min", arg: []node{nodeLiteral{value: NewNum(3)}, nodeLiteral{value: NewNum(1)}}}
	this.Equal(1.0, this.eval(node, env).Num())
}

func (this *SuiteNode) TestNodeRef() {
	env := Env{Resolver: &stubResolver{
		attr:    map[string]Value{"self": NewRef(stubRef{id: 1})},
		refAttr: map[int]map[string]Value{1: {"calm": NewNum(7)}},
	}}
	this.Equal(7.0, this.eval(nodeRef{name: "self", attr: "calm"}, env).Num())
}

// eval 將節點包成 Expr 後求值並斷言成功, 回傳結果值。
func (this *SuiteNode) eval(n node, env Env) Value {
	expr := &Expr{root: n}
	value, ok := expr.Eval(env)
	this.Require().True(ok)
	return value
}
