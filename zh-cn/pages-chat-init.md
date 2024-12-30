# hello braid

### 使用 braid-cli 工具安装脚手架项目

1. 安装 CLI 工具
```bash
$ go install gitee.com/pojol/braidcn-cli@latest
```

2. 使用 CLI 生成一个新的空项目
```bash
$ braidcn-cli new "chat-server" v0.1.9
```

3. 项目所使用的目录结构
```
├── actors      # 用户设计的 actor 存放在这个目录
├── template    # 模版配置目录
├── constant    # 常量目录
    ├── fields  # 通常用于标记系统中各种唯一键值映射，如（actorid, sessionid, roomid 等
├── node        # 节点目录, 服务入口, 主要用于配置服务各种参数和启动项
├── errcode     # errcode 目录
├── handlers    # 消息处理函数目录
├── middleware  # 通用的中间件目录
└── states          # state 目录 （建议 state 拥有统一的序列化，如 protobuf, msgpack, json 等
    ├── commproto   # 通用结构，如 item, mail 等跨服务，甚至跨语言，跨工具（后台管理）
    ├── gameproto   # 游戏通讯协议
    ├── chat        # 聊天 state 模块， 定义数据结构，并实现数据结构提供的一些计算函数
    └── user        # 用户模块（entity）
```