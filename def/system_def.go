package def

const (
	// Wildcard symbol represents routing to any actor of this type
	SymbolWildcard = "?"

	// Represents routing to a group of actors
	// - Note: This symbol can only be used with the send interface (asynchronous call)
	SymbolGroup = "#"

	// Represents routing to all actors of this type
	// - Note: This symbol can only be used with the send interface (asynchronous call)
	SymbolAll = "*"

	// Represents random routing to an actor of this type, but prioritizes actors on the current node
	// If there are no actors of this type on the current node, it randomly selects from other nodes
	SymbolLocalFirst = "~"
)

// 内置的 actors
const (
	ActorDynamicPicker   = "braid.actor_dynamic_picker"
	ActorDynamicRegister = "braid.actor_dynamic_register"
	ActorControl         = "braid.actor_control"
)

// 丑陋的传参方式，等优化 [todo]
const (
	// EvDynamicPick is used to pick an actor
	// customOptions:
	// - actor_id: string
	// - actor_ty: string
	EvDynamicPick = "braid.event_dynamic_pick"

	// EvDynamicRegister is used to register an actor
	// customOptions:
	// - actor_ty: string
	EvDynamicRegister = "braid.event_dynamic_register"

	// EvUnregister is used to unregister an actor
	// customOptions:
	// - actor_id: string
	EvUnregister = "braid.event_unregister"
)

const (
	RedisAddressbookIDField = "braid.addressbook.id"
	// set
	RedisAddressbookTyField = "braid.addressbook.ty."
	// hash
	RedisAddressbookNodesField = "braid.addressbook.nodes"
)
