# Quota Service

[![license](https://img.shields.io/badge/license-apache_2.0-red.svg?style=flat)](https://raw.githubusercontent.com/maniksurtani/quotaservice/master/LICENSE)
![Travis status](https://travis-ci.org/maniksurtani/quotaservice.svg?branch=master "Travis status")

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

# Detailed Design
**COMING SOON**

# Bibliography

[Guava RateLimiter](http://docs.guava-libraries.googlecode.com/git/javadoc/com/google/common/util/concurrent/RateLimiter.html)

[Hystrix](https://github.com/Netflix/Hystrix)

[Resilience by Design](https://www.youtube.com/watch?v=MEgyGamo79I) - a talk by Ben Christensen on Resilience by Design, and the motivation behind Hystrix

[Hierarchical Token Buckets](http://luxik.cdi.cz/~devik/qos/htb/manual/theory.htm)

