# Braid 编程指南
> Braid 是一个轻量级的分布式 Actor 框架，专为构建高性能、可扩展的微服务应用而设计。它提供了一种简单而强大的方式来处理分布式系统中的并发和通信问题。

</br>

### 🌟 特性
* 轻量级 Actor 模型: 基于 Go 协程实现的高效 Actor 系统，每个 Actor 都是独立的计算单元
* 灵活的消息路由: 支持点对点通信、广播和通配符路由
* 分布式寻址: 内置分布式地址簿，支持动态服务发现和负载均衡
* 高性能通信: 基于 gRPC 的高效节点间通信
* 可观测性: 内置追踪和监控支持
* 容错机制: 内置故障恢复和错误处理机制
* 发布订阅: 支持基于主题的消息发布和订阅

</br>

## 🔧 使用指引 (通过脚手架搭建一个分布式聊天服务器
- [1. 通过脚手架初始化项目](zh-cn/pages-chat-init.md)
- [2. 设计 chat actors](zh-cn/pages-chat-actors.md)
- [3. 实现各个 actor 的 handlers](zh-cn/pages-chat-handlers.md)
- [4. 部署分布式聊天服务器](zh-cn/pages-chat-deploy.md)

</br>

## 💡 核心概念
* [Actor 计算单元](zh-cn/pages-actor.md)
* [Handler 计算处理函数](zh-cn/pages-handler.md)
* [State 计算状态](zh-cn/pages-state.md)

</br>

## 🧩 核心接口
* Call, Send, Pub
* OnEvent, OnTimer, Sub

---

### 🌐 加入 Braid 社区

Braid 正在快速发展中，现在是加入这个充满活力的社区的最佳时机！

- GitHub: [https://github.com/pojol/braid](https://github.com/pojol/braid)
- Discord: [加入我们的 Discord](https://discord.gg/yXJgTrkWxT)