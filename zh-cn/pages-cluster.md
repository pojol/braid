# 搭建一个集群

</br>

* [在本地构建一个多节点的集群](#在本地构建一个多节点的集群)
* [在容器中搭建一个多节点集群](#在-swarm-中搭建一个多节点集群)
* [在 k8s 中搭建一个多节点集群](#在-k8s-中搭建一个多节点集群)

</br>

> 在 braid 中搭建一个集群是一件非常容易的事情，您要做的事情就是赋予 node 一个唯一的 ID；然后启动任意的节点数即可，集群内的 actor 分布，可以在配置中通过 weight 进行设置，braid 会通过负载均衡器自动将 actor 分布在不同的节点中

</br>

### 在本地构建一个多节点的集群
```go
id := flag.String("id", "", "Node ID (required)")
flag.Parse()

// 改写配置文件中的 node id
nodeCfg.ID = *id
```

</br>

### 在 swarm 中搭建一个多节点集群
> 将 taskid 作为 BRAID_NODE_ID
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

### 在 k8s 中搭建一个多节点集群
> 将 pod name 作为 BRAID_NODE_ID
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