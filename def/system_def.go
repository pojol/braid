package def

const (
	// 通配符表示路由到任意一个该类型的 actor
	SymbolWildcard = "?"

	// 表示路由到一组 actor
	// - 注 这个符号只能用于 send 接口（异步调用
	SymbolGroup = "#"

	// 表示路由到所有 该类型的 actor
	// - 注 这个符号只能用于 send 接口（异步调用
	SymbolAll = "*"

	// 表示随机路由到一个该类型的 actor，但优先选择本节点上的
	// 如果本节点上没有该类型的 actor，则随机选择其他节点上的
	SymbolLocalFirst = "~"
)

const (
	RedisAddressbookIDField = "braid.addressbook.id"
	RedisAddressbookTyField = "braid.addressbook.ty."
)
