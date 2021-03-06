// 接口文件 balancer 负载均衡
//
package balancer

import (
	"github.com/pojol/braid-go/module"
	"github.com/pojol/braid-go/module/discover"
)

// IPicker 选取器
type IPicker interface {
	// Get 从当前的负载均衡算法中，选取一个匹配的节点
	Get() (nod discover.Node, err error)

	// Add 为当前的服务添加一个新的节点 service gate : [ gate1, gate2 ]
	Add(discover.Node)

	// Rmv 从当前的服务中移除一个旧的节点
	Rmv(discover.Node)

	// Update 更新一个当前服务中的节点（通常是权重信息
	Update(discover.Node)
}

// IBalancer 负载均衡器
type IBalancer interface {
	module.IModule

	// Pick 为 target 服务选取一个合适的节点
	//
	// strategy 选取所使用的策略，在构建阶段通过 opt 传入
	Pick(strategy string, target string) (discover.Node, error)
}
