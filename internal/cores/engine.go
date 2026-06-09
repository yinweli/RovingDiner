package cores

import (
	"strings"

	"github.com/yinweli/RovingDiner/internal/exprs"
	sheeter "github.com/yinweli/RovingDiner/sheet"
)

// engine 驅動引擎本體;持有一場營業的執行期狀態,並委派實作 exprs.Resolver。
// 對應【營業實作規格書 | 二、套件結構】engine.go。M6 補入屬性讀取所需的最小狀態(runtime / self / data);
// 三 port(Operator / Presenter / Rander)等其餘欄位於後續里程碑按需補。
type engine struct {
	runtime *Runtime         // 一場營業的聚合狀態(全域屬性 + 全部容器)
	self    *Self            // 當前求值脈絡的 self 綁定;nil 代表 self 未固定
	data    *sheeter.Sheeter // 靜態表格;查詢函式 / cardGroup / 座位佈局讀取用
}

// Attr 委派全域屬性詞彙表,以自身為 context 求值。
// 先查值表 attrRead;未命中且名稱以 Lock 結尾,剝去後綴改查鎖表 attrLockRead(讀鎖定計數)。
func (this *engine) Attr(name string, arg []exprs.Value) (result exprs.Value, ok bool) {
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
func (this *engine) AttrRef(ref exprs.Ref, name string, arg []exprs.Value) (result exprs.Value, ok bool) {
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

// 編譯期確認 engine 滿足 exprs.Resolver(條件對象求值的接縫)。
var _ exprs.Resolver = (*engine)(nil)
