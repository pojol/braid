# Braid Programming Guide

> Braid is a lightweight distributed game framework that adopts the actor design model, supports microservices, and is suitable for game backends, IoT, and other scenarios.

### Feature
* Actor Model: Braid adopts the Actor model, making concurrent programming simple and intuitive. Each Actor is an independent processing unit that can receive messages, update states, and send messages to other Actors.
* Dynamic Load Balancing: Through built-in load balancing mechanisms, Braid can dynamically allocate and manage Actors in the cluster, ensuring optimal utilization of system resources.
* Flexible Message Handling: Braid supports custom message handling logic, including middleware, timers, and event subscriptions, making your microservices more flexible and versatile.
* High Performance: Based on Go language's concurrency features, Braid offers excellent performance, suitable for building high-throughput distributed systems.
* Easy to Use: Braid's API design is concise and clear, greatly reducing the learning curve and allowing developers to quickly get started.

## Website Pages
- [1. Building Hello Braid with Scaffolding](pages-hello-world.md)
- [2. Building a Service and Registering Actors](pages-container.md)
- [3. Setting Up a WebSocket Server](pages-websocket.md)
- [4. Introducing State to Actors](pages-entity-state.md)
- [5. Handling Messages](pages-actor-message.md)
- [6. Sending Messages](pages-actor-send.md)
- [7. Using Actor State (CRUD)](pages-actor-state.md)
- [8. Designing a Chat Server](pages-chat.md)
- [9. Setting Up a Cluster](pages-cluster.md)
---