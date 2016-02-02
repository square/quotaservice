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
