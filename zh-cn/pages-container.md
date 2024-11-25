# 构建一个服务，并注册actor进去

</br>

* [服务配置](#服务配置)
* [权重](#权重)
* [actor 配置](#actor-配置)
* [actor_types生成器](#actor-生成器)
* [创建一个节点](#创建一个节点)

</br>

### 服务配置
```yaml
node:
  # 节点唯一标识符，解析自动通过环境变量 BRAID_NODE_ID 获取（也可以手动传入
  id: "{BRAID_NODE_ID}"
  weight: "{BRAID_NODE_WEIGHT}"
  
  # Actor 可选项配置列表
  actors:
    # WebSocket 接收器 Actor
    - name: "HTTP_ACCEPTOR"
      options:
        port: "8008"
```

* **ID** - 节点的id，需要全局唯一（可以通过环境变量传入
* **Weight** - 权重值，节点的权重总值

</br>

### 权重
> 权重是一个 braid 中的核心概念，它主要用于表示系统的负载情况，通常我们可以设计一个基础的算法如：
```
// 2c4g 的资源
node_total_weight = 2 * 4 * 1000 = 8000
当我们允许这个资源下运行 4000 个 user actor 时，那么一个 user actor 的 weight = 2
注：具体的配置设置应该基于压力测试的结果去调整
```

</br>

### Actor 模版配置

```yaml
actor_types:
  # WebSocket acceptor
  # for accepting WebSocket connections
  # options:
  #   - port: WebSocket server port
- name: "WEBSOCKET_ACCEPTOR"
  unique: true
  weight: 800
  limit: 1
  category: "core"
  # options:
  #  port: "8008"
```

* **unique** - 表示这个 actor 是否只能在当前节点注册一个，比如 control actor 控制器一个节点一个就已经满足需求了
* **weight** - actor 权重值， 在 braid 中我们需要设计一个权重体系，用于负载均衡； 这个值表示当前 actor 在 node 中的权重数量
[node权重](#权重)
* **limit** - 表示当前节点在集群中的可注册总数，如果为0则表示无限制； 这个字段可以控制系统的负载能力，另外也可以通过设置 1 将 actor 设置为全局唯一
* **dynamic** 标记为dynamic的actor将不会被node启动时构建
* **options** - actor 的可选项， 如 http 的对外端口，或者心跳的路径配置 /heartbeat 等

</br>

### Actor 生成器
> 使用 go generate 通过 actor_types.yml 配置文件生成 actor_types.go 文件
> 在代码中，我们不推荐使用 "WEBSOCKET_ACCEPTOR" 来表示某个 actor 类型， 应该使用 `types.ACTOR_WEBSOCKET_ACCEPTOR`
```go
const (
//  WebSocket 接收器
//  用于接受 WebSocket 连接的 Actor
//  选项:
//    - port: WebSocket 服务器端口
    ACTOR_WEBSOCKET_ACCEPTOR = "WEBSOCKET_ACCEPTOR"
)
```

</br>

### 创建一个节点

```go
// 加载配置，获得节点配置结构，和 actors 配置结构
nodeCfg, actorTypes, err := config.ParseConfig("node_conf.yml", "actor_types.yml")
if err != nil {
    panic(err)
}

// 将 actor 配置对象注册到 actor factory 中
factory := actors.BuildActorFactory(nodeCfg.Actors)
loader := actors.BuildDefaultActorLoader(factory)

// 创建一个节点，并传入 system
nod := node.BuildProcessWithOption(
  core.WithSystem(
    core.NodeWithID(nodeCfg.ID),
    core.NodeWithLoader(loader),
    core.NodeWithFactory(factory),
  ),
)

// 初始化节点
err = nod.Init()
if err != nil {
    panic(fmt.Errorf("node init err %v", err.Error()))
}

// 运行节点
nod.Update()

// 阻塞等待节点退出信号
nod.WaitClose() 
```