# hello braid


### Install the scaffold project using the braid-cli tool

1. Install CLI Tool
```bash
$ go install github.com/pojol/braid-cli@latest
```

2. Using the CLI to Generate a New Empty Project
```bash
$ braid-cli new "you-project-name"
```

3. Creating .go Files from Actor Template Configurations
```bash
$ cd you-project-name/template
$ go generate
```

4. Navigate to the services directory, then try to build and run the demo
```bash
$ cd you-project-name/services/demo-1
$ go run main.go
```

</br>

```
├── actors      # Directory for user-designed actors
├── template    # Configuration directory
├── constant    # Constants directory
├── server      # Main file directory for services, mainly used to configure various service parameters and startup items
├── errcode     # Error code directory
├── chains      # Message handling function directory
├── middleware  # Common middleware directory
└── states          # State directory (it's recommended that states have unified serialization, such as protobuf, msgpack, json, etc.)
    ├── commproto   # Common structures, such as items, mail, etc., across services, even across languages and tools (backend management)
    ├── gameproto   # Game communication protocol
    ├── chat        # Chat state module, defines data structures and implements some calculation functions provided by the data structures
    └── user        # User module (entity)
```
