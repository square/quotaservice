# Quota Service
The purpose of a quota service is to prevent cascading failures in micro-service environments.
This service acts as a traffic cop, slowing down traffic where necessary to prevent overloading
services.

## Raw notes
### MVP
* Best effort, doesn’t have to be 100% accurate.
* Low latency - needs to be very fast. Peak latency in single digit millis. P99s sub-2 millis.
* Hot-path quota lookup requests should not need to touch disk. In-memory leaky-bucket or
  token-buckets should be sufficient.
  * Golang implementations exist: https://github.com/Clever/leakybucket and
    https://godoc.org/github.com/hotei/tokenbucket
* Clients should fall back onto sensible defaults.
  * Initial impl clients in Java/Go/Ruby

### V1
#### Asynchronous reporting
* Each client maintains a token bucket for each endpoint it intends to talk to.
* Client has a thread that fills its token bucket at a fixed rate to ensure graceful degradation.
  (rate configurable)
* Background thread pings the quota server for additional tokens periodically.
* Adds additional tokens received from the service. May also remove tokens/stop the local token
  generator thread.

##### Advantages
* No added RPC to the quota service for each remote call.
* Graceful degradation in event of quota service failure.
* Quota service still able to control rate of calls from one service to the next.
* Clients can still use the synchronous model if required.

## Messages
* Service X wants to talk to Service Y, about Z (Z = map of context values, such as merchant token,
  country code, etc) on endpoint E.
* Quotas per DC
* Sharded in some way?
* Quotas periodically sync’ed to disk, mainly in memory. Needs to be very low latency.
* RPC Backpressure is too coarse grained. Good to have as well, but orthogonal.
  Permanent rules (described in a yaml file?)
* Dynamic (and temporary) rules - will require a data store? Zookeeper? Buckets?

## Configuring quotas
* Hierarchical namespaces
* Feedback loop: success rate to be batched and fed back, and can be used to dial down quotas,
  potentially black-list and route around affected services?
* Dynamic deadlines / timeouts: allow bursts / slow requests, but if all requests get slow, failfast

## Technology stack
* Service: [Golang](https://golang.org/) (for close-to-system low-latency runtime)
* Communications: [gRPC](http://www.grpc.io/) for RPC messages
* Clients: Java, Golang, Ruby

## Background reading
* [Resilience by Design](https://www.youtube.com/watch?v=MEgyGamo79I) - a talk by Ben Christensen
  on Resilience by Design, and the motivation behind Hystrix
