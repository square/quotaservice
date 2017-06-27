# QuotaService client libraries

## Protocol
QuotaService listens over gRPC, using the protobuf service and messages defined in
[`../protos/quota_service.proto`](https://github.com/square/quotaservice/blob/master/protos/quota_service.proto). Additional endpoints (such as JSON over HTTP) may be defined as
well, and some of the clients provided here may support such interaction too. Please see
platform-specific client READMEs.

## Supported platforms
Clients are available for **Golang**, **Java** and **Ruby**, and are organized in subdirectories
under this directory, with the exception of the **Golang** client, which is in this directory to
deal with Go's package importing (allowing for `import "github.com/square/quotaservice/client"`).

## Contributing more clients
Issue a pull request, contributions are welcome! Please make sure your contributed client can at
least communicate over gRPC, has necessary tests and a README. Please see other client libs in this
directory for inspiration.

# Golang client
[![license](https://img.shields.io/badge/license-apache_2.0-red.svg?style=flat)](https://raw.githubusercontent.com/square/quotaservice/master/LICENSE)
[![GoDoc](https://godoc.org/github.com/square/quotaservice/client?status.png)](https://godoc.org/github.com/square/quotaservice/client)
![Project Status](https://img.shields.io/badge/status-beta-orange.svg)

![Status](https://img.shields.io/badge/status-complete-green.svg)
* N/a

![Status](https://img.shields.io/badge/status-WIP-blue.svg)
* Naïve client(s) that integrate with gRPC.

![Status](https://img.shields.io/badge/status-unscheduled-red.svg)
* Smart client, with client-side buckets and asynchronous, bulk token updates from the quota service.

Usage:



