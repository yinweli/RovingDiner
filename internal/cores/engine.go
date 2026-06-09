package cores

import (
	"strings"

	"github.com/yinweli/RovingDiner/internal/exprs"
	sheeter "github.com/yinweli/RovingDiner/sheet"
)

// NewEngine 建立驅動引擎;注入聚合狀態 / self 綁定 / 靜態表格。
func NewEngine(runtime *Runtime, self *Self, data *sheeter.Sheeter) (engine *Engine) {
	return &Engine{
		runtime: runtime,
		self:    self,
		data:    data,
	}
}

// Engine 驅動引擎本體;持有一場營業的執行期狀態,並委派實作 exprs.Resolver。
// 對應【營業實作規格書 | 二、套件結構】engine.go。M6 補入屬性讀取所需的最小狀態(runtime / self / data);
// 三 port(Operator / Presenter / Rander)等其餘欄位於後續里程碑按需補。
// 結構欄位保持私有:games 僅透過匯出方法(Attr / AttrRef / ExecAssign)操作引擎,經 NewEngine 建立。
// 內建函式註冊表為套件層全域 builtin(同 attrRead 等詞彙表),非 per-instance 狀態,故不入欄位。
type Engine struct {
	runtime *Runtime         // 一場營業的聚合狀態(全域屬性 + 全部容器)
	self    *Self            // 當前求值脈絡的 self 綁定;nil 代表 self 未固定
	data    *sheeter.Sheeter // 靜態表格;查詢函式 / cardGroup / 座位佈局讀取用
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

// env 組裝求值期環境:以自身為條件對象 Resolver、帶入套件層全域內建函式註冊表 builtin。
func (this *Engine) env() exprs.Env {
	return exprs.Env{Resolver: this, Builtin: builtin}
}

// 編譯期確認 Engine 滿足 exprs.Resolver(條件對象求值的接縫)。
var _ exprs.Resolver = (*Engine)(nil)
