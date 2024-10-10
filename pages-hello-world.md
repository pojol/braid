# hello braid


```
├── actors      # Directory for user-designed actors
├── config      # Configuration directory
├── constant    # Constants directory
├── server      # Main file directory for services, mainly used to configure various service parameters and startup items
├── errcode     # Error code directory
├── events      # Message handling function directory
├── middleware  # Common middleware directory
└── models          # State directory (it's recommended that states have unified serialization, such as protobuf, msgpack, json, etc.)
    ├── commproto   # Common structures, such as items, mail, etc., across services, even across languages and tools (backend management)
    ├── gameproto   # Game communication protocol
    ├── chat        # Chat state module, defines data structures and implements some calculation functions provided by the data structures
    └── user        # User module (entity)
```