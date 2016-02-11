# Quota Service

# Overview

The purpose of a quota service is to prevent cascading failures in micro-service environments. The service acts as a traffic cop, slowing down traffic where necessary to prevent overloading services. Further, it can be used as a global rate limiter. For this to work, remote procedure calls (RPCs) between services consult the quota service before making a call. The service isn’t strictly for RPCs between services, and can even be used to apply quotas to database calls, for example.

## Audience

The quota service is of interest to anyone building services that communicate with each other via RPC, or that communicate with shared resources such as databases over a network and are interested in limiting the impact of cascading failures due to resource starvation or poor allocation.

## Rationale

Whether the shared resource is a database or a service in itself, less critical systems could overwhelm a shared resource which could trigger outages in more critical functions that depend on this shared resource.

The quota service discussed here aims to ration how much a shared resource can be used by a given service, limiting the impact of such cascading failures. Increasing quotas would involve a capacity planning exercise to ensure the resource is capable of handling the allowed load.

## Background and Project Overview

### Service Runtime

The quota service is written as a [gRPC](http://www.grpc.io) service, implemented in [Go](http://www.golang.org) for close-to-system runtime performance. Go's abundance of relevant open source libraries such as leaky and token buckets also helped influence this decision. Using gRPC also allows for easy client implementations, since gRPC has client libraries for a wide variety of platforms.

### Clients

Naïve clients simply ask the quota service for permission to make an RPC call. Clients have sensible defaults configured, to fall back to if the quota service is unavailable or doesn't respond in a timely manner. An example of such a fallback could be a naïve rate limiter set to, say, 100 requests per second.

This design also discusses smarter clients. Smart clients maintain a client-side bucket which will

be used by the application. A client-side thread will top up the bucket at the default rate. An

asynchronous thread will periodically ask the quota service for more tokens to add to the bucket. This allows for greater resilience to latency spikes in the quota service, and takes the additional step of querying the quota service off the critical path. However this comes at the additional cost of a more complex client.

A [gRPC client interceptor](https://github.com/grpc/grpc-java/blob/master/core/src/main/java/io/grpc/ClientInterceptor.java) can be used to make sure quotas are checked before RPC calls are made, as demonstrated in [this example](https://github.com/grpc/grpc-java/blob/master/examples/src/main/java/io/grpc/examples/header/CustomHeaderClient.java). Discussions are currently underway to ensure the same level of client-side interceptor support is available across all gRPC client libraries.

## Open Source

The quota service and clients will be completely open source (NOTE:  Using the ASLv2 license) from the start. It is developed in the open, issues tracked in the open, design documentation made open. The service is designed to hook into proprietary services organizations may have, however. For example, plugging in logging systems, metrics and monitoring systems, and service discovery systems are all supported and discussed in detail below.

# Goals

In order of priority.

**Phase 1**

* Provide service that maintains in-memory data structures tracking quotas.

* Statically configure the service via a YAML configuration file.

    * This is temporary for this phase, and the ability to configure via YAML will be *removed***_ _**by the start of Phase 2.

* Explicit calls to the quota service's API on whether an RPC should be allowed to proceed or not. Services are expected to use this API *before* actually making the RPC it needs to make.

* Cluster-awareness to present a single master to update buckets.

**Phase 2**

* Admin UI to add services and quotas to the quota service to allow reconfiguration without redeployment.

* Naïve client(s) that integrate with gRPC.

* Expose metrics on the server, tracking rates of denials/throttling.

**Phase 3**

* Smart client, with client-side buckets and asynchronous, bulk token updates from the quota service.

* Context-aware buckets, allowing sub-quotas for specific requests based on contents of the request payload.

# Non-goals

* We’re not attempting to build a capacity-planning tool here.

# Eng Design: The Service

## Architecture

![image alt text](image_0.png)

Scenario 1: quotas set up for Service A and B in the Quota service.

* Service A asks the Quota Service for permission to call into Service B.

* The Quota Service responds with the permission requested.

* Service A makes the call into Service B.

Scenario 2: quotas **_not_** set up for Service A and B in the Quota service.

* Service A asks the Quota Service for permission to call into Service B.

* The Quota Service responds with a rejection message.

* Service A uses client-side defaults to call into Service B

* Service A admins alerted that quotas have not been set up and defaults are being used.

Scenario 3: quotas set up for Service A and B in the Quota service, but none available.

* Service A asks the Quota Service for permission to call into Service B.

* The Quota Service blocks, waiting for permission to become available.

* The Quota Service responds with the permission requested.

* Service A makes the call into Service B.

Scenario 4: quotas set up for Service A and B in the Quota service, but none available.

* Service A asks the Quota Service for permission to call into Service B.

* The Quota Service blocks, waiting for permission to become available.

* The Quota Service times out.

* Service A **_does not_** makes the call into Service B.

## Tokens and Token Buckets

For each service-to-service call, the quota service maintains a [token bucket](https://en.wikipedia.org/wiki/Token_bucket). Tokens are

added to the token bucket at a fixed rate. Handling an AllowRequest involves locating the appropriate token bucket and acquiring a token from it. If no tokens are available, the request will wait for a configured amount of time for new tokens to be added. If the time elapses and no tokens are available, the call responds with an error.

Token buckets have a fixed size. If a token bucket is full, no additional tokens are added.

### Naming and wildcards

A token bucket has a name and a namespace to which it belongs. Namespaces have defaults that can be applied to named buckets. Namespaces can also be configured to allow dynamically created buckets from a template. Names and namespaces are case-sensitive. Valid characters for names and namespaces are those that match this regexp: [a-zA-Z0-9_]+

#### Example 1: S2S RPCs

A token bucket namespace for requests from Service A to Service B, for all endpoints:

`service_a_service_b`

A token bucket for requests from Service A to Service B, for all endpoints on the UserService

`service_a_service_b:UserService`

A token bucket for requests from Service A to Service B, for UserService.getUser()

`service_a_service_b:UserService_getUser`

#### Example 2: Databases

A token bucket namespace for Service A making calls to its MySQL database:

`service_a_mysql`

A token bucket for Service A  making calls to the users table in its MySQL database:

`service_a_mysql:users`

A token bucket for Service A  making calls to the users table in its MySQL database, for inserts:

`service_a_mysql:users_insert`

#### Example 3: Context-specific quotas

A token bucket namespace for all public API calls in Service A:

`service_a_public`

Dynamic bucket configured for this namespace, restricting individual IP addresses to send at most 10 reqs/sec. Requests would look like:

`service_a_public:${IP_ADDRESS}`

## Data storage

Token buckets are stored in a map, allowing for constant time lookups. This map is is keyed on bucket name (as described above), pointing to an instance of a token bucket. Token buckets are created and added to the map lazily.

### Hierarchies and bucket search order

The preference is to use a named bucket as possible. For example, given an `AllowRequest` for `service_a_service_b:UserService_getUser`,  the quota service would look for token buckets in the following order:

1. `service_a_service_b:UserService_getUser`

2. Create a dynamic bucket in the `service_a_service_b` namespace, if allowed.

3. Use a default bucket in the `service_a_service_b` namespace, if allowed.

4. Use a global default bucket, if allowed.

### Dynamic token buckets

If a bucket doesn’t exist but the namespace is configured to allow dynamic buckets, a named bucket is created using defaults from a template as defined on the namespace. If configured to allow dynamic buckets, a namespace will also be configured with a limit of dynamic buckets it may create.

#### Deleting buckets

Buckets may be deleted to reclaim memory. A bucket can have a maximum idle time defined, after which it is removed. Accesses to buckets are recorded. If a bucket is removed and subsequently accessed, it is recreated. and populated to half its capacity.

### Default token buckets

If a bucket isn't found and dynamic buckets are not enabled for a namespace, behavior depends on whether a default bucket is configured on the namespace. If one is configured, it is used. If not, a global default bucket is attempted. If a global default bucket doesn’t exist, the call fails.

### Storing token buckets

Buckets are maintained solely in memory, and are not persisted. If a server fails and is restarted, buckets are recreated as per configuration and will start empty. The replenishing thread also starts immediately, providing each bucket with tokens.

In future persisting buckets to disk may be considered but for now is considered out of scope.

#### Storing configurations

Configurations for each bucket are stored in memory, alongside each bucket, after reading them from a configuration YAML file. Once YAML file support for configurations is removed, configurations will be managed via a web based admin console and persisted to a durable back-end, with adapters for storing on disk as well as other destinations such as Zookeeper for greater durability.

### Filling tokens

Tokens are added to a bucket lazily when tokens are requested and sufficient time has passed to allow additional permits to be added, taking inspiration from [Guava’s RateLimiter](https://code.google.com/p/guava-libraries/source/browse/guava/src/com/google/common/util/concurrent/RateLimiter.java?r=cb140e39acac7da75a7f28bcf406c9ff9086c7cf) library.

## API: Protobuf service

A protobuf service endpoint will be exposed by the quota service.

```protobuf
syntax = "proto2";

package quotaservice;

service QuotaService {
 rpc Allow(AllowRequest) returns (AllowResponse) {}
}

message AllowRequest {
 optional string namespace = 1;
 optional string name = 2;
 optional int32 num_tokens_requested = 3; // Defaults to 1
}

message AllowResponse {
 enum Status {
   OK = 1;
   OK_WAIT = 2;
   REJECTED = 3;
 }

 optional Status status = 1;
 optional int32 num_tokens_granted = 2;
 optional int64 wait_millis = 3; // Defaults to 0
}
```

### Alternative APIs

While we’re designing for a gRPC-based API, it is conceivable that other RPC mechanisms may also be desired, such as [Thrift](https://thrift.apache.org/) or even simple JSON-over-HTTP. To this end, the quota service is designed to plug into any request/response style RPC mechanism, by providing an interface as an extension point, that would have to be implemented to support more RPC mechanisms.

```
type RpcEndpoint interface {
  Init(cfgs *configs.Configs, qs service.QuotaService)
  Start()
  Stop()
}
```

The QuotaService interface passed in encapsulates the token buckets and provides the basic functionality of acquiring quotas.

```
type QuotaService interface {
  Allow(bucketName string, tokensRequested int, emptyBucketPolicyOverride EmptyBucketPolicyOverride) (int, error)
}
```

The built-in gRPC implementation of the RpcEndpoint interface, for example, simply adapts the protobuf service implementation to call in to QuotaService.Allow, transforming parameters accordingly.

## Clustering and High Availability

The quota service can be run as a single node, however it will have limited scalability and availability characteristics when run in this manner. As such, it is also designed to run in a cluster.

### Cluster notification

Rather than implementing cluster coordination between Quota Service nodes, we rely on an external system providing such cluster change information. Most organizations have systems that perform this function anyway, and plugging this into the Quota Service is trivial. The Quota Service declares a Clustering interface, which must be implemented and passed in when initializing the Quota Service. The interface above can be implemented in a number of ways, such as backed by [Zookeeper](https://zookeeper.apache.org/), a [RAFT](https://raft.github.io/) implementation, or a distributed consensus protocol.

```
type Clustering interface {
  // Returns the current node name
  CurrentNodeName() string
  // Returns a slice of node names that form a cluster.
  Members() []string
  // Returns a channel that is used to notify listeners of a membership change.
  MembershipChangeNotificationChannel() chan bool
}
```

### Mastering buckets

Buckets are mastered by a single node in the cluster, with remaining nodes in the cluster acting as backups for the master. When determining the master of a bucket, the namespace is used as a key in a [consistent hash](http://michaelnielsen.org/blog/consistent-hashing/) wheel.

### Proxying requests

If a node receives a request for which it is not a master, it will proxy the request to the master, and add an additional hint to the response so that the client can, for future requests, contact the master node directly.

### Caching bucket masters on clients

To prevent a lot of unnecessary proxying, clients are expected to cache bucket master hints and direct requests accordingly. If this cache gets stale, the request will still hit a non-master node and will get an updated hint in the response. If the cache contains a node that no longer exists, the client should handle the connection error and try any arbitrary, available node, again receiving an updated hint.

### Cross-datacenter concerns

The quota service is designed to be run local to a data center. In future, we could enhance this and have nodes in each datacenter communicate and share token buckets, but that is out of scope for this design.

## Alternative clustering design

Simpler architecture. All nodes answer all requests. The data structure is modeled in memcached, which all nodes connect to. **TODO**: prototype this, see if it performs well enough under load. Proxying will not be necessary in this case.

## Logging

The quota service makes use of standard Go [logging](https://golang.org/pkg/log/). However this can be overridden to allow for different logging back-ends by passing in a logger implementing Logger:

```
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


## Metrics

If enabled, metrics are gathered and exported via a MetricsHandler. The MetricsHandler can be queried for data and published to monitoring systems, and exposes the following information, per token bucket, as well as for the default bucket:

* Pass rate

* Throttle rate

* Timeout rate

* Overall response time histogram

* Pass-only response time histogram

* Volume by caller

Histograms are a high-fidelity [implementation](https://github.com/codahale/hdrhistogram), based on Gil Tene’s [HDRHistogram](https://github.com/HdrHistogram/HdrHistogram/blob/master/README.md).

## Configuration

The following configuration elements need to be provided to the quota service:

* Global:

    * Allow global default bucket (true | false)

    * Global default bucket settings

    * Filler frequency in millis (default: 1000 i.e., every second)

    * Metrics enabled (true | false)

* For each namespace:

    * Allow namespace default bucket (true | false)

    * Namespace default bucket settings

    * Allow dynamic buckets (true | false)

    * Max dynamic buckets (default: -1 i.e., unlimited)

    * Dynamic bucket settings

    * Named bucket configs

* For each bucket:

    * Size (default: 100)

    * Fill rate (default: 50/s)

    * Wait timeout millis (default: 1000)

    * Max idle time millis (default: -1)

## Service-level objectives

Since the quota service will be on the critical path for many communications, it needs to have extremely low latency. We’re targeting 2ms per invocation at 99.9th percentile, under an expected load of 50,000 QPS per node.

### Failure scenarios

Since the quota service is a best-effort system, failure scenarios are easy to deal with.

* When faced with latency spikes or communication failures with the service, clients will fall back to client-side rate-limiters preconfigured with sensible/conservative defaults.

    * After a threshold of failed calls to the quota service, clients will stop calling into the quota service and just use client-side rate-limiters, periodically checking if the quota service has resumed service.

* When faced with a loss of token buckets, a service simply initiates a new token bucket, losing no more than a few seconds of accurate traffic shaping. Based on how buckets are configured (size, fill rates, fill frequency) impact could be minimal.

* When faced with a quota service node outage, other nodes in the cluster would take over in a time specified by how quickly GNS can notify the remaining nodes of a loss of one node. During this period, token buckets are reset, and the same effect of losing a token bucket altogether is experienced.

## Rollout and test plan

TODO

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

The quota service will expose the following metrics:

* Timeout rate

* Client-side rate-limit usage frequency

## Smart clients

Smart clients allow for periodic, asynchronous communication with the quota service to request for tokens in batches, and always make use of a client-side token bucket when rate-limiting.

**_// TODO(manik) Complete this section_**

# Bibliography

[Guava RateLimiter](http://docs.guava-libraries.googlecode.com/git/javadoc/com/google/common/util/concurrent/RateLimiter.html)

[Hystrix](https://github.com/Netflix/Hystrix)

[Resilience by Design](https://www.youtube.com/watch?v=MEgyGamo79I) - a talk by Ben Christensen on Resilience by Design, and the motivation behind Hystrix

[Hierarchical Token Buckets](http://luxik.cdi.cz/~devik/qos/htb/manual/theory.htm)

