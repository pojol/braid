package workerthread

import (
	"testing"
)

// 同步调用
//
//	阻塞等待调用，无论这个call在服务中存在多少跳，都会阻塞等待消息的返回（或超时返回）
func TestSystemCall(t *testing.T) {

	// 调用某个 actor 的 event
	//Call(ctx, router.Target{ID: "mock_actor_1", Ty: "mock_actor", Ev: "mock_test"}, nil)

	// 调用某一种 actor 类型中的一个
	//Call(ctx, router.Target{ID: def.SymbolWildcard, Ty: "mock_actor", Ev: "mock_test"}, nil)
}

// 异步调用
//
//	异步调用通常用于，调用一个需要较长时间处理的函数（比如大于10s），不用立刻返回的消息发布类型
func TestSystemSend(t *testing.T) {

	// 调用某个 actor 的 event
	// 调用某一种 actor 类型中的一个
	// 同上，只是接口换成 Send

	// 调用某组 actor 的 event
	/*
		Send(ctx, router.Target{
			ID:    def.SymbolGroup,
			Ty:    "mock_actor",
			Ev:    "mock_test",
			Group: []string{"mock_actor_1", "mock_actor_2", "mock_actor_3"}}, nil)
	*/

	// 调用所有该 actor 类型的 actor
	/*
		Send(ctx, router.Target{
			ID: def.SymbolAll,
			Ty: "mock_actor",
			Ev: "mock_test",
		}, nil)
	*/
}

// pubsub 发布到队列
//
//	队列型的消息会缓存在磁盘或redis的队列中，等待接收端消费，是一种处理大规模或者安全要求高（必定完成）的消息发布类型
func TestSystemPub(t *testing.T) {

	// 发布给某一个指定的 actor
	// - 注 : pubsub 的消息 和 call | send 的消息id 不能重复（消费模式不一致
	//Pub(ctx, router.Target{ID: "mock_actor_1", Ty: "mock_actor", Ev: "ps_mock_test"}, nil)

	// 发布一个消息给某个 actor 类型随机消费
	//Pub(ctx, router.Target{ID: def.SymbolWildcard, Ty: "mock_actor", Ev: "ps_mock_test"}, nil)

	// 发布一个消息给所有指定类型的 actor
	//Pub(ctx, router.Target{ID: def.SymbolAll, Ty: "mock_actor", Ev: "ps_mock_test"}, nil)
}
