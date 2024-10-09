# build a service and register actor into it


### service config

### actor config
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
  # options:
  #  port: "8008"
```

* Unique
* Weight
* Limit
* Options

