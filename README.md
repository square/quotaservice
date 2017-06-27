# Quota Service
[![license](https://img.shields.io/badge/license-apache_2.0-red.svg?style=flat)](https://raw.githubusercontent.com/square/quotaservice/master/LICENSE)
[![Travis status](https://travis-ci.org/square/quotaservice.svg?branch=master "Travis status")](https://travis-ci.org/square/quotaservice)
[![GoDoc](https://godoc.org/github.com/square/quotaservice?status.png)](https://godoc.org/github.com/square/quotaservice)
![Project Status](https://img.shields.io/badge/status-beta-orange.svg)
[![Gitter](https://badges.gitter.im/Join%20Chat.svg)](https://gitter.im/square/quotaservice?utm_source=badge&utm_medium=badge&utm_campaign=pr-badge)

# Overview

The purpose of a quota service is to prevent cascading failures in micro-service environments. The service acts as a traffic cop, slowing down traffic where necessary to prevent overloading services. Further, it can be used as a global rate limiter. For this to work, remote procedure calls (RPCs) between services consult the quota service before making a call. The service isn’t strictly for RPCs between services, and can even be used to apply quotas to database calls, for example.

## Audience

The quota service is of interest to anyone building services that communicate with each other via RPC, or that communicate with shared resources such as databases over a network and are interested in limiting the impact of cascading failures due to resource starvation or poor allocation.

## Rationale

Whether the shared resource is a database or a service in itself, less critical systems could overwhelm a shared resource which could trigger outages in more critical functions that depend on this shared resource.

We have seen this as a practical, real-life in the past, and have attached custom rate limiter code to various parts of various services, and/or just increased capacity without visibility or measurement.

The quota service discussed here aims to ration how much a shared resource can be used by a given service, limiting the impact of such cascading failures. Increasing quotas would involve a capacity planning exercise to ensure the resource is capable of handling the allowed load.

## Background and Project Overview

### Service Runtime

The quota service is written as a [gRPC](http://www.grpc.io) service, implemented in [Go](http://www.golang.org) for close-to-system runtime performance. Go's abundance of relevant open source libraries such as leaky and token buckets also helped influence this decision. Using gRPC also allows for easy client implementations, since gRPC has client libraries for a wide variety of platforms.

### Clients

Naïve clients simply ask the quota service for permission to make an RPC call. Clients have sensible defaults configured, to fall back to if the quota service is unavailable or doesn't respond in a timely manner. An example of such a fallback could be a naïve rate limiter set to, say, 100 requests per second.

This design also discusses smarter clients. Smart clients maintain a client-side bucket which will be used by the application. A client-side thread will top up the bucket at the default rate. An asynchronous thread will periodically ask the quota service for more tokens to add to the bucket. This allows for greater resilience to latency spikes in the quota service, and takes the additional step of querying the quota service off the critical path. However this comes at the additional cost of a more complex client.

A [gRPC client interceptor](https://github.com/grpc/grpc-java/blob/master/core/src/main/java/io/grpc/ClientInterceptor.java) can be used to make sure quotas are checked before RPC calls are made, as demonstrated in [this example](https://github.com/grpc/grpc-java/blob/master/examples/src/main/java/io/grpc/examples/header/CustomHeaderClient.java). Discussions are currently underway to ensure the same level of client-side interceptor support is available across all gRPC client libraries.

## Open Source

The quota service and clients are all completely open source, under the Apache Software Foundation License v2.0 (ASLv2). See [LICENSE](https://raw.githubusercontent.com/square/quotaservice/master/LICENSE) for details.

# Goals

In order of priority.

![Status](https://img.shields.io/badge/status-complete-green.svg)
* Provide service that maintains in-memory data structures tracking quotas.
* Statically configure the service via a YAML configuration file.
* Explicit calls to the quota service's API on whether an RPC should be allowed to proceed or not. Services are expected to use this API *before* actually making the RPC it needs to make.
* Weighted quotas.
* Expose a listener on the server, so events can be tracked and statistics gathered.
* REST/HTTP endpoint.
* Admin CLI to add services and quotas to the quota service to allow reconfiguration without redeployment.
* Admin UI to add services and quotas to the quota service to allow reconfiguration without redeployment.

![Status](https://img.shields.io/badge/status-WIP-blue.svg)
* Naïve client(s) that integrate with gRPC.

![Status](https://img.shields.io/badge/status-unscheduled-red.svg)
* Smart client, with client-side buckets and asynchronous, bulk token updates from the quota service.
* Allow for bursting (hard limits vs soft limits)
* Sharded back-end

# Use cases

A general purpose rate limiter should restrict the rate at which an application makes use of a shared resource such as a database or service. Different usage patterns exist, including:

* Simple rate limiting where all usage is treated equal. The calling application acquires a single token for a named service and if the token is granted, the application proceeds using the service.
    * Example: when trying to restrict the rate of SQL queries to a database from a specific service.

* Weighted rate limiting, where all usage is not considered to be equal. The calling application acquires a number of tokens for a named service, based on how expensive the call is considered to be.
    * Example: when some SQL queries are proportionately more expensive than other, similar queries, perhaps based on the `LIMIT` clause restricting the number of records retrieved.

* Dynamically named shared resources. Same as above, except that quotas for resources are created and destroyed on the fly, based on a template.
    * Example: when trying to restrict logins from a given IP address.

* After-the-fact quota consumption updates. Same as simple rate limiting, except that a single token is modeled around a unit of cost. For example, a millisecond of processing time. A calling application would acquire a single token to perform a task, and after performing the task, would update the quota service with the actual number of "tokens" used. After-the-fact updates could be batched to reduce network calls.
    * Example: when trying to limit requests to the amount of parallelism available on a shared resource.

# Eng Design: The Service

## Architecture

### Scenario 1
Quotas set up for `Pinky` to call into `TheBrain` in the Quota Service.

![Sequence Diagram](/resources/sequences_1.png?raw=true)

* `Pinky` asks the Quota Service for a token to call into `TheBrain`.
* The Quota Service responds with the token requested, and status of `OK`.
* `Pinky` makes the call into `TheBrain`.

### Scenario 2
Quotas set up for `Pinky` to call into `TheBrain` in the Quota service, but no tokens immediately available.

![Sequence Diagram](/resources/sequences_2.png?raw=true)

* `Pinky` asks the Quota Service for a token to call into `TheBrain`.
* The Quota Service returns with status `OK_WAIT` and the number of millis to wait, indicating that the client has to wait before proceeding.
* `Pinky` waits, then makes the call into `TheBrain`.

### Scenario 3
Quotas **_not_** set up for `Pinky` to call into `TheBrain` in the Quota Service.

![Sequence Diagram](/resources/sequences_3.png?raw=true)

* `Pinky` asks the Quota Service for a token to call into `TheBrain` (namespace `Pinky_TheBrain`).
* Namespace `Pinky_TheBrain` does not have a default bucket or allows for dynamic buckets.
* The Quota Service responds with a rejection message (status `REJECT`)
* `Pinky` uses client-side defaults to call into `TheBrain` and logs/alerts accordingly.

### Scenario 4
Quotas set up for `Pinky` to call into `TheBrain` in the Quota service, but no tokens available for a long while.

![Sequence Diagram](/resources/sequences_4.png?raw=true)

* `Pinky` asks the Quota Service for a token to call into `TheBrain`.
* No quota immediately available.
* Time before more quota is available exceeds `maxWaitTime`
    * `maxWaitTime` is configured per bucket, and can be overridden per request.
* The Quota Service does not claim tokens, responds with status `REJECTED`.
* `Pinky` **_does not_** makes the call into `TheBrain`.

### Scenario 5
Quotas set up for `Pinky` to call into `TheBrain` in the Quota service, pre-fetching tokens.

![Sequence Diagram](/resources/sequences_5.png?raw=true)

* `Pinky` asks the Quota Service for 10 tokens to call into `TheBrain`.
* The Quota Service responds with the tokens requested, and status of `OK`.
* `Pinky` uses the tokens to make calls into `TheBrain`, using 1 token each time.
* When `Pinky` runs out of tokens, it asks the Quota Service for more tokens.

### Scenario 6
Quotas set up for `Pinky` to call into `TheBrain` in the Quota service, post-accounting.

![Sequence Diagram](/resources/sequences_6.png?raw=true)

* Buckets set up so that 1 token == number of millis taken to perform a task.
* `Pinky` asks the Quota Service for a token to call into `TheBrain`.
* The Quota Service responds with the token requested, and status of `OK`.
* `Pinky` uses the tokens to make calls into `TheBrain`, measuring how long the call takes.
* After the call, `Pinky` asks the Quota Service for `time_taken_in_millis` tokens, as post-accounting for the call just completed.

## Tokens and Token Buckets

For each service-to-service call, the quota service maintains a [token bucket](https://en.wikipedia.org/wiki/Token_bucket). Tokens are added to the token bucket at a fixed rate (see detailed algorithm below). Handling an `AllowRequest` involves:

1. Locating the appropriate token bucket, or creating one if dynamic buckets are allowed.
1. Acquiring a token from it.
1. If no tokens are available, the request will respond with a wait time, after which a token will be available, and claim a future for this token - provided the wait time is below a configured maximum.
1. Maximum wait times are configured on each bucket, but can be overridden in the `AllowRequest`, however overridden maximum wait times cannot exceed a configured maximum for the bucket.
1. If the wait time exceeds the maximum, no future tokens are claimed, and a rejected response is returned.

Token buckets have a fixed size. If a token bucket is full, no additional tokens are added.

### Naming and wildcards

A token bucket has a name and a namespace to which it belongs. Namespaces have defaults that can be applied to named buckets. Namespaces can also be configured to allow dynamically created buckets from a template. Names and namespaces are case-sensitive. Valid characters for names and namespaces are those that match this regexp: `[a-zA-Z0-9_]+`.

#### Example 1: S2S RPCs

A token bucket namespace for requests from `Pinky` to `TheBrain`, for all services:

```
Pinky_TheBrain
```

A token bucket for requests from `Pinky` to `TheBrain`, for all endpoints on the `UserService`

```
Pinky_TheBrain:UserService
```

A token bucket for requests from `Pinky` to `TheBrain`, for `UserService.getUser()`

```
Pinky_TheBrain:UserService_getUser
```

#### Example 2: Databases

A token bucket namespace for `Pinky` making calls to its MySQL database:

```
Pinky_PinkyMySQL
```

A token bucket for `Pinky` making calls to the `users` table in its MySQL database:

```
Pinky_PinkyMySQL:users
```

A token bucket for `Pinky` making calls to the `users` table in its MySQL database, for inserts:

```
Pinky_PinkyMySQL:users_insert
```

#### Example 3: Context-specific quotas

A token bucket namespace for all user login requests in `TheBrain`:

```
TheBrain_userLogins
```

Dynamic bucket configured for this namespace, restricting user logins to at most 1/sec. Requests would look like:

```
TheBrain_userLogins:${userId}
```

## Data storage

Token buckets are stored in a map, allowing for constant time lookups. This map is is keyed on bucket name (as described above), pointing to an instance of a token bucket. Token buckets are created and added to the map lazily.

### Hierarchies and bucket search order

The preference is to use a named bucket as possible. For example, given an `AllowRequest` for  `Pinky_TheBrain:UserService_getUser`,  the quota service would look for token buckets in the following order:

1. `Pinky_TheBrain:UserService_getUser`

2. Create a dynamic bucket in the `Pinky_TheBrain` namespace, if allowed.

3. Use a default bucket in the `Pinky_TheBrain` namespace, if allowed.

4. Use a global default bucket, if allowed.

### Dynamic token buckets

If a bucket doesn’t exist but the namespace is configured to allow dynamic buckets, a named bucket is created using defaults from a template as defined on the namespace. If configured to allow dynamic buckets, a namespace will also be configured with a limit of dynamic buckets it may create.

#### Deleting buckets

Buckets may be deleted to reclaim memory. A bucket can have a maximum idle time defined, after which it is removed. Accesses to buckets are recorded. If a bucket is removed and subsequently accessed, it is recreated. and filled.

### Default token buckets

If a bucket isn't found and dynamic buckets are not enabled for a namespace, behavior depends on whether a default bucket is configured on the namespace. If one is configured, it is used. If not, a global default bucket is attempted. If a global default bucket doesn’t exist, the call fails.

### Storing token buckets

Buckets are maintained solely in-memory, and are not persisted. If a server fails and is restarted, buckets are recreated as per configuration and will start empty. The replenishing thread also starts immediately, providing each bucket with tokens.

In future persisting buckets to disk may be considered but for now is considered out of scope.

#### Storing configurations

Configurations for each bucket are stored in memory, alongside each bucket, after reading them from a configuration YAML file. Once YAML file support for configurations is removed, configurations will be managed via a web based admin console and persisted to a durable back-end, with adapters for storing on disk as well as other destinations such as MySQL, Zookeeper or etcd as examples, for greater durability.

### Filling tokens

Tokens are added to a bucket lazily when tokens are requested and sufficient time has passed to allow additional permits to be added, taking inspiration from [Guava’s RateLimiter](https://code.google.com/p/guava-libraries/source/browse/guava/src/com/google/common/util/concurrent/RateLimiter.java?r=cb140e39acac7da75a7f28bcf406c9ff9086c7cf) library.

#### Algorithm

The algorithm below lazily accumulates tokens, up to the size of the bucket, based on the time difference between when tokens will next become available. If a caller requests for more tokens than is available, the tokens are granted, but the next caller will have to wait for more tokens to become available. This is achieved by incrementing `tokensNextAvailable` for a specified period of time to pay back the "token debt" of the previous caller.

Note that a `maxDebtNanos` configuration parameter is maintained per bucket, to ensure requests don’t attempt to acquire a large number of tokens, thereby locking up the quota service for a long period of time.

```
var tokensNextAvailableNanos // Stored per bucket
var accumulatedTokens // Stored per bucket

var bucketSize // from config
var fillRate // from config, as tokens/sec

var nanosBetweenTokens = 1E+9 / fillRate
var freshTokens = 0

if currentTimeNanos > tokensNextAvailableNanos {
  freshTokens = (currentTimeNanos() - tokensNextAvailableNanos) / nanosBetweenTokens
  accumulatedTokens = min(bucketSize, accumulatedTokens + freshTokens)
  tokensNextAvailableNanos = currentTimeNanos
}

waitTime = tokensNextAvailableNanos - currentTimeNanos
accumulatedTokensUsed = min(accumulatedTokens, requested)
tokensToWaitFor = requested - accumulatedTokensUsed
futureWaitNanos = tokensToWaitFor * nanosBetweenTokens

if tokensNextAvailableNanos + futureWaitNanos - currentTimeNanos > maxDebtNanos {
  // Return rejection response
} else {
  tokensNextAvailableNanos = tokensNextAvailableNanos + futureWaitNanos
  accumulatedTokens = accumulatedTokens - accumulatedTokensUsed

  // waitTime contains how long the caller has to wait for, before using token requested.
  // Return success response with waitTime
}
```


## API: Protobuf service

A protobuf service endpoint will be exposed by the quota service, as defined [here](https://github.com/square/quotaservice/blob/master/protos/quota_service.proto).

### Alternative APIs

While we’re designing for a gRPC-based API, it is conceivable that other RPC mechanisms may also be desired, such as [Thrift](https://thrift.apache.org/) or even simple JSON-over-HTTP. To this end, the quota service is designed to plug into any request/response style RPC mechanism, by providing an interface as an extension point, that would have to be implemented to support more RPC mechanisms.

```go
type RpcEndpoint interface {
  Init(cfgs *configs.Configs, qs service.QuotaService)
  Start()
  Stop()
}
```

The `QuotaService` interface passed in encapsulates the token buckets and provides the basic functionality of acquiring quotas.

```go
type QuotaService interface {
  Allow(bucketName string, tokensRequested int, emptyBucketPolicyOverride EmptyBucketPolicyOverride) (int, error)
}
```

The built-in gRPC implementation of the RpcEndpoint interface, for example, simply adapts the protobuf service implementation to call in to QuotaService.Allow, transforming parameters accordingly.

## Clustering and High Availability

The quota service can be run as a single node, however it will have limited scalability and availability characteristics when run in this manner. As such, it is also designed to run in a cluster, backed by a shared data structure that holds the token buckets. Any node may update the data structure so requests can be load balanced to all quota service nodes.

![Shared Storage Diagram](/resources/shared_storage.png?raw=true)

### Shared data structure

The shared data structure is treated as ephemeral, and as such, complexities of persistence, replication, disaster recovery are all averted. If the shared data structure’s contents are lost, buckets are lazily rebuilt as needed, as per best-effort guarantees.

### Redis implementation

The only available shared data structure at the moment is backed by Redis. Redis performs to within expectations (see below on SLOs). Redis is treated as ephemeral, so persisting or adding durability to Redis' state is unnecessary.

Other implementations - including ones based on distributed consensus algorithms - can easily be plugged in.

### Sharding

The shared data structure could be sharded, hashed on namespace, to provide greater concurrency and capacity if needed, though out of scope for this design. This is trivial to add at a later date, and libraries that perform sharded connection pool management exist.

## Logging

The quota service makes use of standard Go [logging](https://golang.org/pkg/log/). However this can be overridden to allow for different logging back-ends by passing in a logger implementing Logger:

```go
// Logger mimics golang's standard Logger as an interface.
type Logger interface {
  Fatal(args ...interface{})
  Fatalf(format string, args ...interface{})
  Fatalln(args ...interface{})
  Print(args ...interface{})
  Printf(format string, args ...interface{})
  Println(args ...interface{})
}
```


## Listeners

Listeners can be attached, to be notified of events that take place, such as:

* Tokens served (including whether a wait time was imposed)
* Tokens not served due to:
  * Timeout (max wait exceeded wait time imposed)
  * Too many tokens requested
  * Bucket miss (non-existent, or too many dynamic buckets)
  * Dynamic bucket created
  * Bucket removed (garbage-collected)

Each event callback passes the caller the following details:

```go
type Event interface {
	EventType() EventType
	Namespace() string
	BucketName() string
	Dynamic() bool
	NumTokens() int64
	WaitTime() time.Duration
}
```

where `EventType` is defined as:

```go
type EventType int

const (
	EVENT_TOKENS_SERVED EventType = iota
	EVENT_TIMEOUT_SERVING_TOKENS
	EVENT_TOO_MANY_TOKENS_REQUESTED
	EVENT_BUCKET_MISS
	EVENT_BUCKET_CREATED
	EVENT_BUCKET_REMOVED
)

```

### Metrics
Metrics can be implemented by attaching an event listener and collecting data from the event.

## Configuration

The following configuration elements need to be provided to the quota service:

* Global:
    * Global default bucket settings (*disabled if unset*)

* For each namespace:
    * Namespace default bucket settings (*disabled if unset*)
    * Max dynamic buckets (default: `0` i.e., unlimited)
    * Dynamic bucket template (*disabled if unset*)

* For each bucket:
    * Size (default: `100`)
    * Fill rate per second (default: `50`)
    * Wait timeout millis (default: `1000`)
    * Max idle time millis (default: `-1`)
    * Max debt millis - the maximum amount of time in the future a request can pre-reserve tokens (default: `10000`)
    * Max tokens per request (default: `fill_rate`)

See the GoDocs on [`configs.ServiceConfig`](https://godoc.org/github.com/square/quotaservice/configs#ServiceConfig) for more details.

## Service-level objectives

### Load testing the prototype

A prototype of the quota service, currently deployed in staging, backed by a single Redis node, exhibits the following performance characteristics when served by 2 QuotaService nodes. Each node restricted to 4 CPU cores.

```
Completed 1,519,404 iterations, mean 947.051µs each, at 135,040 requests/sec.

Latency at percentiles:
  50 th percentile: 884.735µs
  75 th percentile: 983.039µs
  90 th percentile: 1.114111ms
  95 th percentile: 1.179647ms
  99 th percentile: 2.031615ms
  99.9 th percentile: 6.553599ms
  99.99 th percentile: 209.715199ms
Test time 10m0.114231921s
```

While latencies 99th percentile and above aren’t ideal, this is just a prototype and will go through profiling and performance tuning before being deployed to production. During this test, host resources both on the quota servers and the shared Redis back-end were under-utilized.

### Greater throughput

The same prototype was pushed harder, with greater load, yielding the following results:

```
Completed 5,895,093 iterations, mean 2.235415ms each, at 228,864 requests/sec.

Latency at percentiles:
  50 th percentile: 1.114111ms
  75 th percentile: 1.310719ms
  90 th percentile: 1.966079ms
  95 th percentile: 3.801087ms
  99 th percentile: 25.165823ms
  99.9 th percentile: 142.606335ms
  99.99 th percentile: 318.767103ms
Test time 10m1.060592248s
```

While capable of dealing with far higher throughput, we do see latency spike after the 95th percentile. More investigation will be needed, but this level of throughput is well above our targets in any case.

### Failure scenarios

Since the quota service is a best-effort system, failure scenarios are easy to deal with.

* When faced with latency spikes or communication failures with the service, clients will fall back to client-side rate-limiters preconfigured with sensible/conservative defaults.

    * After a threshold of failed calls to the quota service, clients will stop calling into the quota service and just use client-side rate-limiters, periodically checking if the quota service has resumed service.

* When faced with a loss of token buckets (i.e., loss of Redis state), a service simply initiates a new token bucket, losing no more than a few seconds of accurate traffic shaping. Based on how buckets are configured (size, fill rates, fill frequency) impact could be minimal.

# Eng Design: Clients

## Manual clients

Since clients are generated gRPC service stubs, there is no design beyond the protobuf service definition for this phase. Services actively call into the quota service to obtain permissions before making RPC calls.

## Naïve clients

These clients are implemented as gRPC client-side interceptors, and are designed to rate-limit  gRPC-based service-to-service calls. The calling service would just need to make an RPC call to a target service, and the interceptor will, transparent to the caller, first call into the quota service, blocking if necessary. Such clients are configured with fallback policies, so in the event of being unable to contact the quota service, will fall back to a simple [rate limiter](http://docs.guava-libraries.googlecode.com/git/javadoc/com/google/common/util/concurrent/RateLimiter.html) with configured limits.

### Configuration

Naive clients are configured with:

* Default rate limit per service name -> service name pair
* Timeout when communicating with the quota service

### Metrics

The quota service clients will expose the following metrics:

* Timeout rate
* Client-side rate-limit usage frequency

## Smart clients

Smart clients allow for periodic, asynchronous communication with the quota service to request for tokens in batches, and always make use of a client-side token bucket when rate-limiting.

_**// TODO(manik)** Complete this section_

# Bibliography

* [Guava RateLimiter](http://docs.guava-libraries.googlecode.com/git/javadoc/com/google/common/util/concurrent/RateLimiter.html)
* [Hystrix](https://github.com/Netflix/Hystrix)
* [Resilience by Design](https://www.youtube.com/watch?v=MEgyGamo79I) - a talk by Ben Christensen on Resilience by Design, and the motivation behind Hystrix
* [Hierarchical Token Buckets](http://luxik.cdi.cz/~devik/qos/htb/manual/theory.htm)
