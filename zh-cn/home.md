# Braid 编程指南

> Braid 是一个轻量的分布式游戏框架，采用了 actor 设计模型，支持微服务，适用于游戏后端、物联网等场景。

### Braid 的亮点
* Actor 模型: Braid 采用 Actor 模型，使得并发编程变得简单直观。每个 Actor 都是独立的处理单元，可以接收消息、更新状态、发送消息给其他 Actor。
* 动态负载均衡: 通过内置的负载均衡机制，Braid 可以动态地在集群中分配和管理 Actor，确保系统资源的最优利用。
* 灵活的消息处理: Braid 支持自定义消息处理逻辑，包括中间件、定时器和事件订阅，让你的微服务更加灵活多变。
* 高性能: 基于 Go 语言的并发特性，Braid 提供了出色的性能表现，适合构建高吞吐量的分布式系统。
* 易于使用: Braid 的 API 设计简洁明了，大大降低了学习曲线，让开发者可以快速上手。

## Website Pages
- [1.通过脚手架构建hello braid](zh-cn/pages-hello-world.md)
- [2.构建一个服务，并注册actor进去](zh-cn/pages-container.md)
- [3.搭建一个websocket服务器](zh-cn/pages-websocket.md)
- [4.为actor引入状态](zh-cn/pages-entity-state.md)
- [5.处理消息](zh-cn/pages-actor-message.md)
- [6.发送消息](zh-cn/pages-actor-send.md)
- [7.使用actor状态（增删改查](zh-cn/pages-actor-state.md)
- [8.设计一个聊天服务器](zh-cn/pages-chat.md)
- [9.搭建一个集群](zh-cn/pages-cluster.md)
---
