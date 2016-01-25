# Quota Service

## Overview
The purpose of a quota service is to prevent cascading failures in micro-service environments.
The service acts as a traffic cop, slowing down traffic where necessary to prevent overloading
services. For this to work, remote procedure calls (RPCs) between services consult the quota
service before making a call.

### Service Runtime
The Quota Service is written as a [gRPC][1] service, implemented in [Go][2] for close-to-system
runtime performance. Go's abundance of relevant open source libraries such as [leaky][4] and
[token buckets][5] also helped influence this decision.

#### gRPC
Using gRPC allows for easy client implementations, however the service is designed to allow for
alternate RPC frameworks to be used. gRPC clients exist for C, C++, Objective-C, Java, Go, Python,
Ruby, PHP, C# and Javascript.

### Clients
Naive clients simply make an RPC call to the quota service asking for permission to make an RPC
call. Clients have sensible defaults configured, to fall back to if the quota service is unavailable
or doesn't respond in time. An example of such a fallback could be a naive rate limiter set to, say,
100 requests/sec.

Smarter clients will also be provided. Such clients will maintain a client-side bucket which will
be used by the application. A client-side thread will top up the bucket at the default rate. An
asynchronous thread will periodically ask the quota service for more tokens to add to the bucket.
This allows for greater resilience to latency spikes in the quota service, and takes this additional
request off the critical path. However this comes at the additional cost of a more complex client.

#### Integration with gRPC clients
A [gRPC client interceptor][6] can be used to make sure quotas are checked before RPC calls are
made. [Example][7].

**TODO(manik)**: Make sure we can do this in Go and Ruby.

## Goals
In order of priority.
#### Phase 1
* Provide service that maintains in-memory data structures tracking quotas.
* Statically configured via YAML.
* Explicit calls to the quota service's API on whether an RPC should be allowed to proceed or not.
  Services are expected to use this API *before* actually making the RPC it needs to make.

#### Phase 2
* Naive client that integrates with gRPC, transparently makes a call to the quota service before
  making an RPC.
* Configured using coarse-grained fallbacks (e.g., 100 requests/sec when talking to Service X; 200
  requests/sec when talking to Service Y)
* Expose metrics on the server, tracking rates of denials/throttling.

#### Phase 3
* Admin UI to add services and quotas to the quota service to allow reconfiguration without
  redeployment.

#### Phase 4
* Smart client, with client-side buckets and asynchronous, bulk token updates from the quota service.
* Context-aware buckets, allowing sub-quotas for specific requests based on contents of the request
  payload.

## Design: Phase 1

### Dependencies
This is designed to be open sourced from the beginning. As such, it will not have any dependencies
on any proprietary libraries. Its primary dependencies are gRPC, protocol buffers and standard Go
libraries.

### Protobuf service
A proto3 as well as "legacy" proto2 service endpoint will be exposed. They will be functionally
identical but will allow for a wider range of clients to make use of the service, since proto3 is
not as widely adopted. The protobuf compiler used will still have to be protoc3 though, using the
proto2 definition format.

```proto
syntax = "proto3";

package quotaservice;

service QuotaService {
  rpc Allow(AllowRequest) returns (AllowResponse) {}
}

message AllowRequest {
  string source_system = 1;
  string destination_system = 2;
  string service_name = 3;
  string endpoint_name = 4;
}

message AllowResponse {
  enum Status {
    OK = 1;
    TIMED_OUT = 2;
    BUCKET_REJECTED = 3;
  }

  Status status = 1;
  bool granted = 2;
}

```

A functionally identical endpoint in proto2 syntax will be available in package `quotaservice.proto2`.

### Tokens and Buckets
For each service-to-service call, the quota service maintains a [token bucket][8]. Tokens are
added to the bucket at a fixed rate. When a call has to be made, it will take a token from the
bucket. If no tokens are available, it will wait for a configured amount of time for new tokens to
be added. If the time elapses and no tokens are available, the call returns with an error. Token
buckets have a fixed size. If a token bucket is full, no additional tokens are added.

#### Bucket naming and wildcards
Categories are case-insensitive, and are named as:

`<SourceSystem>/<DestinationSystem>/<ServiceName>/<EndpointName>`

##### Examples
Quota for Spot to Esperanto, for all services:

`spot/esperanto/*/*`

Quota for Spot to Esperanto, for `PaymentService`:

`spot/esperanto/PaymentService/*`

Quota for Spot to Esperanto, for `PaymentService.lookupPayment()`:

`spot/esperanto/PaymentService/lookupPayment`

#### Hierarchies

#### Default Buckets
If a bucket isn't found (assume the 3 buckets defined above), behavior depends on a configured
bucket miss policy. If this policy is `REJECT`, the token request fails with an appropriate error.
If the policy is `USE_DEFAULTS`, a default bucket is used.

#### Storing buckets
Buckets are maintained solely in memory, and are not persisted. If a server fails and is restarted,
buckets are recreated as per configuration and will start empty. The replenishing thread also starts
immediately, providing each bucket with tokens.

##### In-memory structure
**TODO(manik)**

#### Filling tokens
A thread runs periodically, topping up all token buckets with their configured fill rates. All fill
rates are defined per second, and fill amounts calculated based on when the filler last updated a
given bucket.

### Configuration
The following configuration elements need to be provided to the quota service:

* For each named bucket, as well as a default:
  * Token bucket size
  * Token bucket fill rate
  * Token bucket wait timeout
* Bucket miss policy (`REJECT` | `USE_DEFAULTS`)
* Filler frequency in millis (default: `1000` i.e., every second)

### Client design
Since clients are generated gRPC service stubs, there is no design beyond the protobuf service
definition.

### Cross-datacenter, high availability and disaster recovery concerns
**TODO(manik)**

### Availability concerns
**TODO(manik)**

### Performance concerns
**TODO(manik)**

# Raw notes
### MVP
* Best effort, doesn’t have to be 100% accurate.
* Low latency - needs to be very fast. Peak latency in single digit millis. P99s sub-2 millis.
* Hot-path quota lookup requests should not need to touch disk. In-memory leaky-bucket or
  token-buckets should be sufficient.
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
* Service: [Golang][2] (for close-to-system low-latency runtime)
* Communications: [gRPC][1] for RPC messages
* Clients: Java, Golang, Ruby

## Background reading
* [Resilience by Design][3] - a talk by Ben Christensen
  on Resilience by Design, and the motivation behind Hystrix
* [Hierarchical Token Buckets in Linux][9]


[1]: http://www.grpc.io
[2]: https://golang.org
[3]: https://www.youtube.com/watch?v=MEgyGamo79I "Resilience by Design by Ben Christensen"
[4]: https://github.com/Clever/leakybucket
[5]: https://godoc.org/github.com/hotei/tokenbucket
[6]: https://github.com/grpc/grpc-java/blob/master/core/src/main/java/io/grpc/ClientInterceptor.java
[7]: https://github.com/grpc/grpc-java/blob/master/examples/src/main/java/io/grpc/examples/header/CustomHeaderClient.java
[8]: https://en.wikipedia.org/wiki/Token_bucket
[9]: http://luxik.cdi.cz/~devik/qos/htb/manual/theory.htm


