# Setting up a Cluster

</br>

* [Building a Multi-Node Cluster Locally](#building-a-multi-node-cluster-locally)
* [Setting up a Multi-Node Cluster in Containers](#setting-up-a-multi-node-cluster-in-swarm)
* [Setting up a Multi-Node Cluster in Kubernetes](#setting-up-a-multi-node-cluster-in-kubernetes)

</br>

> Setting up a cluster in Braid is a very easy task. All you need to do is assign a unique ID to each node, then start any number of nodes. The distribution of actors within the cluster can be set through the 'weight' in the configuration. Braid will automatically distribute actors across different nodes through a load balancer.

</br>

### Building a Multi-Node Cluster Locally
```go
id := flag.String("id", "", "Node ID (required)")
flag.Parse()

// Modify the node ID in the configuration file
nodeCfg.ID = *id
```

</br>

### Setting up a Multi-Node Cluster in Swarm
> Use the task ID as BRAID_NODE_ID
```yaml
version: '3.8'
services:
  your_service:
    image: your_image
    deploy:
      mode: replicated
      replicas: 3
    environment:
      - BRAID_NODE_ID={{.Task.ID}}
```

</br>

### Setting up a Multi-Node Cluster in Kubernetes
> Use the pod name as BRAID_NODE_ID
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: your-deployment
spec:
  template:
    spec:
      containers:
      - name: your-container
        image: your-image
        env:
        - name: BRAID_NODE_ID
          valueFrom:
            fieldRef:
              fieldPath: metadata.name
```