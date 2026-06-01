package game

import "github.com/yinweli/Project_005/internal/defines"

// NewRuntime 建立空白 Runtime；初始化 Game 與各容器（不載入任何遊戲資料）。
// 初始牌堆 / 排隊 / 前置技能等由組裝層（features / testdata）填入。
func NewRuntime(seed int64) *Runtime {
	return &Runtime{
		Game: &Game{
			DrawTotal:  map[int32]int32{},
			DropTotal:  map[int32]int32{},
			PlayTotal:  map[int32]int32{},
			ExileTotal: map[int32]int32{},
		},
		Seat: map[int32]*Guest{},
		Seed: seed,
	}
}

// Runtime 一場營業的聚合狀態：全域屬性（Game 實例）+ 全部容器 + 結算旗標。
// 對應規格書【五、實例結構】【六、容器結構】。
//
// 引擎是唯一真相、持有全部實例與容器（純可序列化資料、零顯示依賴）；
// 前端靠事件流 + InstanceID 對照投影畫面，不持有第二份規則狀態。
// 營業邊界（開始前 / 結束後）call stack 為空、Runtime 是純資料，可整包序列化；
// 營業進行中暫停狀態活在 call stack，故不可序列化（符合「營業中不存檔」）。
type Runtime struct {
	Game *Game // 全域屬性與事件型狀態

	// 卡牌容器
	Hand  []*Card // 手牌（玩家檢視序）
	Deck  []*Card // 抽牌牌堆（先進後出；新進入者置頂）
	Drop  []*Card // 棄牌牌堆（先進後出；新進入者置頂）
	Exile []*Card // 流放牌堆（先進後出；新進入者置頂）

	// 顧客容器
	Wait    []*Guest         // 排隊佇列（先進先出）
	Seat    map[int32]*Guest // 座位列表（座位編號 -> 顧客）
	Roam    []*Guest         // 遊蕩列表（順序無關）
	Cardify []*Guest         // 卡牌化列表（順序無關）

	// 效果 / 行動容器
	Effect []*Effect // 效果佇列（處理時依【十五、作用順序】排序）
	Action []*Action // 行動佇列（先進先出）

	// 流程旗標與設置
	Settling    bool               // 結算旗標；執行結算的重入防護
	Seed        int64              // 本場 PRNG 種子（執行期狀態，供顯示 / 重現）
	PrefixSkill []int32            // 前置技能列表（營業開始時逐一啟動）
	lastID      defines.InstanceID // 實例編號產生器游標
}

// NextID 配發下一個唯一實例編號（卡牌 / 顧客 / 效果共用同一序列）。
func (this *Runtime) NextID() defines.InstanceID {
	this.lastID++
	return this.lastID
}
