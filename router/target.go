package router

type Target struct {
	ID    string   // 唯一标识 （也可以使用 def.Symbol 来表示一些特殊的发送方式
	Ty    string   // actor type
	Ev    string   // event
	Group []string // 需要对一组actor进行消息发送时使用，只能在 send 接口使用
}
