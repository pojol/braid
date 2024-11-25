# hello braid


### Install the scaffold project using the braid-cli tool

1. Install CLI Tool
```bash
$ go install github.com/pojol/braid-cli@latest
```

2. Using the CLI to Generate a New Empty Project
```bash
$ braid-cli new "you-project-name" v0.0.1
```

3. Creating .go Files from Actor Template Configurations
```bash
$ cd you-project-name/template
$ go generate
```

4. Navigate to the services directory, then try to build and run the demo
```bash
$ cd you-project-name/node
$ go run main.go
```

</br>

```
├── actors      # Directory for user-designed actors
├── template    # Configuration directory
├── constant    # Constants directory
    ├── fields  # Used to mark various unique key-value mappings 
                # (actorid, sessionid, roomid, etc.)
├── node        # Main file directory for services
                # mainly used to configure service parameters and startup items
├── errcode     # Error code directory
├── handlers    # Message handling function directory
├── middleware  # Common middleware directory
└── states      # State directory 
                # (recommended to use unified serialization: protobuf, msgpack, json, etc.)
    ├── commproto   # Common structures across services
                    # (items, mail, etc., shared across languages and tools)
    ├── gameproto   # Game communication protocol
    ├── chat        # Chat state module
                    # defines data structures and calculation functions
    └── user        # User module (entity)
```