package expr

// mockResolver 是測試用的假 Resolver:property / member 以 map 注入,未登記者回傳失敗。
// 需驗證引數傳遞或自訂行為時,設定 onProperty / onMember 覆寫對應方法。
// 預設 Member 對空物件存取屬性回傳失敗,模擬規格【二十七〇5】「引用為空物件 → 評估失敗」。
type mockResolver struct {
	property   map[string]Value
	member     map[string]Value
	onProperty func(name string, arg []Value) (value Value, ok bool)
	onMember   func(target Value, name string, arg []Value) (value Value, ok bool)
}

func (this *mockResolver) Property(name string, arg []Value) (value Value, ok bool) {
	if this.onProperty != nil {
		return this.onProperty(name, arg)
	} // if

	value, ok = this.property[name]
	return value, ok
}

func (this *mockResolver) Member(target Value, name string, arg []Value) (value Value, ok bool) {
	if this.onMember != nil {
		return this.onMember(target, name, arg)
	} // if

	if target.IsNone() == true {
		return Value{}, false // 空物件存取屬性 → 失敗
	} // if

	value, ok = this.member[name]
	return value, ok
}
