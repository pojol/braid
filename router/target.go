package router

type Target struct {
	// 可以使用模式匹配
	//  * 表示任意一个actor
	//  + 表示所有的actor
	ID    string
	Ty    string   // actor 类型
	Ev    string   // 事件
	Group []string // 需要对一组actor进行消息发送时使用，只能在 send 接口使用
}
