package cores

import (
	"strings"

	"github.com/yinweli/RovingDiner/internal/exprs"
	sheeter "github.com/yinweli/RovingDiner/sheet"
)

// Engine 驅動引擎本體;持有一場營業的執行期狀態,並委派實作 exprs.Resolver。
// 對應【營業實作規格書 | 二、套件結構】engine.go。M6 補入屬性讀取所需的最小狀態(runtime / self / data);
// M8 補入命令對象解析所需的 operator(*Pick 暫停玩家選)/ rander(*Rand 隨機、deckTop auto-shuffle);
// 餘下的 Presenter 於後續里程碑按需補。
// 結構欄位保持私有:games 僅透過匯出方法(Attr / AttrRef / ExecAssign / Run)操作引擎,經 NewEngine 建立。
// 內建函式註冊表為套件層全域 builtin(同 attrRead 等詞彙表),非 per-instance 狀態,故不入欄位。
type Engine struct {
	runtime  *Runtime             // 一場營業的聚合狀態(全域屬性 + 全部容器)
	self     *Self                // 當前求值脈絡的 self 綁定;nil 代表 self 未固定
	data     *sheeter.Sheeter     // 靜態表格;查詢函式 / cardGroup / 座位佈局讀取用
	operator Operator             // 玩家輸入 port;命令對象 *Pick 暫停流程由玩家選取
	rander   Rander               // 亂數 port;命令對象 *Rand 隨機選取、deckTop auto-shuffle 洗牌
	award    map[int32]awardGroup // 抽獎衍生索引(群組 → 候選);NewEngine 建一次、唯讀,供 *Roll / *Morph 用
}

// NewEngine 建立驅動引擎;注入聚合狀態 / self 綁定 / 靜態表格 / 玩家輸入與亂數兩 port。
func NewEngine(runtime *Runtime, self *Self, data *sheeter.Sheeter, operator Operator, rander Rander) (engine *Engine) {
	return &Engine{
		runtime:  runtime,
		self:     self,
		data:     data,
		operator: operator,
		rander:   rander,
		award:    buildAward(data),
	}
}

// Attr 委派全域屬性詞彙表,以自身為 context 求值。
// 先查值表 attrRead;未命中且名稱以 Lock 結尾,剝去後綴改查鎖表 attrLockRead(讀鎖定計數)。
func (this *Engine) Attr(name string, arg []exprs.Value) (result exprs.Value, ok bool) {
	if read, known := attrRead[name]; known {
		return read(this, arg)
	} // if

	if base, found := strings.CutSuffix(name, "Lock"); found {
		if read, known := attrLockRead[base]; known {
			return read(this, arg)
		} // if
	} // if

	return exprs.Value{}, false
}

// AttrRef 以引用 ref 為主體委派引用屬性詞彙表。Lock 後綴路由規則同 Attr。
func (this *Engine) AttrRef(ref exprs.Ref, name string, arg []exprs.Value) (result exprs.Value, ok bool) {
	if read, known := attrRefRead[name]; known {
		return read(this, ref, arg)
	} // if

	if base, found := strings.CutSuffix(name, "Lock"); found {
		if read, known := attrRefLockRead[base]; known {
			return read(this, ref, arg)
		} // if
	} // if

	return exprs.Value{}, false
}

// ExecAssign 執行屬性修改命令(【營業規格書 | 十七、命令 | 1】);回報是否實際寫入。
// 帶值賦值先求值右值(評估失敗 / 右值非數值 → no-op);引用左值先解析引用主體(空物件 / 型別不符 / 不存在 → no-op);
// 再經寫入詞彙表(全域 attrWrite / 引用 attrRefWrite)依賦值符變更狀態。名稱可寫性由 games.Validate 先行檢查。
func (this *Engine) ExecAssign(base, refAttr string, isRef bool, op AssignKind, value *exprs.Expr) (changed bool) {
	n := float64(0)

	if op != AssignLock && op != AssignUnlock { // @ # 不帶右值
		result, ok := value.Eval(this.env())

		if ok == false || result.IsNum() == false {
			return false // 算術評估失敗 / 右值非數值 → no-op
		} // if

		n = result.Num()
	} // if

	if isRef {
		owner, ok := this.Attr(base, nil)

		if ok == false || owner.IsRef() == false {
			return false // 引用解析為空物件 / 型別不符 / 不存在 → no-op
		} // if

		write, known := attrRefWrite[refAttr]

		if known == false {
			return false
		} // if

		return write(this, owner.Ref(), op, n)
	} // if

	write, known := attrWrite[base]

	if known == false {
		return false
	} // if

	return write(this, op, n)
}

// ExecOperate 執行操作命令(【營業規格書 | 十七、命令 | 2】【二十五、操作命令清單】);走法 X:games 持 AST 型別 switch、engine 做分派。
// 流程:求值命令對象 [...] 參數 → selectObject 解析作用集合 → 求值其餘參數 → 查 command 表 → 對集合 fan-out。
// **任一參數評估失敗 → 整動作 no-op**(對齊規格);命令對象 / verb 未登錄(Validate 應先擋)亦 no-op。
// 逐元素的型別 / 位置不符 no-op、空集合整體 no-op 由各 verb 本體處理。
func (this *Engine) ExecOperate(verb, selectorName string, selectorParam, arg []*exprs.Expr) {
	selectorValue, ok := this.evalAll(selectorParam)

	if ok == false {
		return // 命令對象參數評估失敗 → 整動作 no-op
	} // if

	target, known := this.selectObject(selectorName, selectorValue)

	if known == false {
		return // 命令對象未登錄 → no-op
	} // if

	argValue, ok := this.evalAll(arg)

	if ok == false {
		return // 其餘參數評估失敗 → 整動作 no-op
	} // if

	run, found := command[verb]

	if found == false {
		return // 未知命令 → no-op
	} // if

	run(this, target, argValue)
}

// selectObject 解析命令對象為作用對象集合(【營業規格書 | 二十四、命令對象清單】);ok=false 代表命令對象名稱未登錄。
// arg 為 [...] 內參數的已求值結果(求值由呼叫端 M9 ExecOperate 負責);selector 自身不碰 exprs。
// 回身分集 []InstanceID(非具型別實例):M9 verb 自行 locate 取實例 + 查容器位置(位置不符 no-op 本就要查),
// 使本表保持同構、與 Self / effect 的 InstanceID 身分模型一致。
func (this *Engine) selectObject(name string, arg []exprs.Value) (result []InstanceID, ok bool) {
	resolve, known := selector[name]

	if known == false {
		return nil, false
	} // if

	return resolve(this, arg), true
}

// evalAll 依序求值一串算術式;任一失敗回 ok=false(供 ExecOperate 套用「任一參數評估失敗 → 整動作 no-op」)。
func (this *Engine) evalAll(expr []*exprs.Expr) (result []exprs.Value, ok bool) {
	result = make([]exprs.Value, 0, len(expr))

	for _, itor := range expr {
		value, valid := itor.Eval(this.env())

		if valid == false {
			return nil, false
		} // if

		result = append(result, value)
	} // for

	return result, true
}

// locateCard 以實例編號自四牌堆找出卡牌及其所在容器;未命中回 (nil, ContainerNone, false)。
// 供操作命令取實例 + 查容器位置(位置不符 no-op);卡牌僅存在於 手牌 / 抽牌 / 棄牌 / 流放。
func (this *Engine) locateCard(id InstanceID) (card *Card, where ContainerKind, ok bool) {
	runtime := this.runtime

	if found := findCard(runtime.Hand, id); found != nil {
		return found, ContainerHand, true
	} // if

	if found := findCard(runtime.Deck, id); found != nil {
		return found, ContainerDeck, true
	} // if

	if found := findCard(runtime.Drop, id); found != nil {
		return found, ContainerDrop, true
	} // if

	if found := findCard(runtime.Exile, id); found != nil {
		return found, ContainerExile, true
	} // if

	return nil, ContainerNone, false
}

// locateGuest 以實例編號自顧客四容器找出顧客及其所在容器;未命中回 (nil, ContainerNone, false)。
// 供操作命令取顧客實例 + 查容器位置(型別不符 / 位置不符 no-op);顧客存在於 座位 / 排隊 / 遊蕩 / 卡牌化。
func (this *Engine) locateGuest(id InstanceID) (guest *Guest, where ContainerKind, ok bool) {
	runtime := this.runtime

	for _, itor := range runtime.Seat {
		if itor != nil && itor.InstanceID == id {
			return itor, ContainerSeat, true
		} // if
	} // for

	if found := findGuest(runtime.Wait, id); found != nil {
		return found, ContainerWait, true
	} // if

	if found := findGuest(runtime.Roam, id); found != nil {
		return found, ContainerRoam, true
	} // if

	if found := findGuest(runtime.Cardify, id); found != nil {
		return found, ContainerCardify, true
	} // if

	return nil, ContainerNone, false
}

// env 組裝求值期環境:以自身為條件對象 Resolver、帶入套件層全域內建函式註冊表 builtin。
func (this *Engine) env() exprs.Env {
	return exprs.Env{Resolver: this, Builtin: builtin}
}

// 編譯期確認 Engine 滿足 exprs.Resolver(條件對象求值的接縫)。
var _ exprs.Resolver = (*Engine)(nil)
