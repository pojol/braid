# build a service and register actor into it

</br>

* [Service Configuration](#service-configuration)
* [Weight](#weight)
* [Actor Configuration](#actor-configuration)
* [Actor Types Generator](#actor-types-generator)
* [Creating a Node](#creating-a-node)

</br>

### Service Configuration
```yaml
node:
  # Unique node identifier, can be passed through the NODE_ID environment variable
  id: "{NODE_ID}"
  weight: "{NODE_WEIGHT}"
  
  # List of optional Actor configurations
  actors:
    # WebSocket Acceptor Actor
    - name: "HTTP_ACCEPTOR"
      options:
        port: "8008"
```

* **ID** - The node's ID, must be globally unique (can be passed through environment variables)
* **Weight** - Weight value, total weight value of the node

</br>

### Weight
> Weight is a core concept in a braid, mainly used to represent the system's load situation. Usually, we can design a basic algorithm like:
```
// 2c4g resources
node_total_weight = 2 * 4 * 1000 = 8000
When we allow 4000 user actors to run under this resource, then the weight of one user actor = 2
Note: The specific configuration settings should be adjusted based on stress test results
```

</br>

### Actor Configuration

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

* **Unique** - Indicates whether only one of this actor can be registered on the current node. For example, one control actor per node is sufficient.
* **Weight** - Actor weight value. In braid, we need to design a weight system for load balancing; this value represents the weight quantity of the current actor in the node.
[Node Weight](#weight)
* **Limit** - Indicates the total number that can be registered in the cluster for the current node. If 0, it means unlimited. This field can control the system's load capacity, and by setting it to 1, the actor can be set as globally unique.
* **Category** - "core" (basic capability), "static" (static actor), or "dynamic" (dynamic actor)
* **Options** - Optional items for the actor, such as the external port for HTTP, or heartbeat path configuration /heartbeat, etc.

</br>

### Actor Types Generator
> Use go generate to generate the actor_types.go file from the actor_types.yml configuration file
> In the code, we don't recommend using "WEBSOCKET_ACCEPTOR" to represent an actor type, instead use `types.ACTOR_WEBSOCKET_ACCEPTOR`
```go
const (
//  WebSocket Acceptor
//  Actor for accepting WebSocket connections
//  Options:
//    - port: WebSocket server port
    ACTOR_WEBSOCKET_ACCEPTOR = "WEBSOCKET_ACCEPTOR"
)
```

</br>

### Creating a Node

```go
// Load configuration, obtain node configuration structure and actors configuration structure
nodeCfg, actorTypes, err := config.ParseConfig("node_conf.yml", "actor_types.yml")
if err != nil {
    panic(err)
}

// Register actor configuration objects to the actor factory
factory := actors.BuildActorFactory(actorTypes)

// Create a node and pass in the system
nod := node.BuildProcessWithOption(
		core.WithSystem(
			node.BuildSystemWithOption(nodeCfg.ID, factory),
		),
	)

// Register actors to this node
// Note: 
// - Failure to register basic capability actors will cause a panic
// - Static actors will be skipped if registration fails due to limit or unique fields
// - Dynamic actors will not be registered during the startup phase
for _, regActor := range nodeCfg.ActorOpts {
    builder := nod.System().Loader(regActor.Name).WithID(nodeCfg.ID + "_" + regActor.Name)
    for key, val := range regActor.Options {
        builder.WithOpt(key, val)
    }
    _, err = builder.Build()
    if err != nil {
        panic(err.Error())
    }
}

// Initialize the node
err = nod.Init()
if err != nil {
    panic(fmt.Errorf("node init err %v", err.Error()))
}

// Run the node
nod.Update()

// Block and wait for the node exit signal
nod.WaitClose() 
```