# Welcome project for newcomers!

# Introduction
In this project you have to implement a message broker, based on `broker.Broker`
interface. There are unit tests to specify requirements and also validate your implementation.

# Roadmap
- [ ] Implement `broker.Broker` interface and pass all tests (Use memory for database)
- [ ] Create *dockerfile* and *docker-compose* files for your deployment
- [ ] Add basic logs and prometheus metrics
  - Metrics for each RPCs:
    - `method_count` to show count of failed/successful RPC calls
    - `method_duration` for latency of each call, in 99, 95, 50 quantiles
    - `active_subscribers` to display total active subscriptions
  - Env metrics:
    - Metrics for your application memory, cpu load, cpu utilization, GCs
- [ ] Implement gRPC API for the broker and main functionalities
- [ ] Persist messages in postgres
- [ ] Persist messages in cassandra
- [ ] Deploy your app on k8s
