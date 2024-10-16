# hello braid

### 使用 braid-cli 工具安装脚手架项目

1. 安装 CLI 工具
```bash
$ go install github.com/pojol/braid-cli@latest
```

2. 使用 CLI 生成一个新的空项目
```bash
$ braid-cli new "you-project-name"
```

3. 从 Actor 模板配置创建 .go 文件
```bash
$ cd you-project-name/template
$ go generate
```

4. 进入到 services 目录，然后尝试构建并运行 demo
```bash
$ cd you-project-name/services/demo-1
$ go run main.go
```

```
├── actors      # 用户设计的 actor 存放在这个目录
├── template    # 模版配置目录
├── constant    # 常量目录
├── server      # 服务 main 文件目录，主要用于配置服务各种参数和启动项
├── errcode     # errcode 目录
├── chains      # 消息处理函数目录
├── middleware  # 通用的中间件目录
└── states          # state 目录 （建议 state 拥有统一的序列化，如 protobuf, msgpack, json 等
    ├── commproto   # 通用结构，如 item, mail 等跨服务，甚至跨语言，跨工具（后台管理）
    ├── gameproto   # 游戏通讯协议
    ├── chat        # 聊天 state 模块， 定义数据结构，并实现数据结构提供的一些计算函数
    └── user        # 用户模块（entity）
```