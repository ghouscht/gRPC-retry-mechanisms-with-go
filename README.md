# gRPC retry mechanisms in go
My first ever talk I held at the [Bärner Go Talks 2023 no. 3](https://www.meetup.com/berner-go-meetup/events/293782118/) 
about gRPC and retry mechanisms in go.

## gRPC in general

## gRPC retry mechanisms
Read more on this [here](https://pkg.go.dev/google.golang.org/grpc/examples/features/retry#section-readme) and [here](https://github.com/grpc/proposal/blob/master/A6-client-retries.md).

gRPC supports client side retries but they are disabled by default and must be configured with a so called [service config](https://github.com/grpc/grpc/blob/master/doc/service_config.md).

### unary RPCs
```mermaid
sequenceDiagram
    participant Client Application
    participant gRPC Client Lib
    participant gRPC Server Lib
    participant Server Application

    Client Application->>Server Application: sends RPC
    note left of Client Application: Client makes gRPC call
    note right of Server Application: failure occurs

    Server Application->>gRPC Client Lib: returns UNAVAILABLE
    note left of gRPC Client Lib: retry occurs after backoff

    gRPC Client Lib->>Server Application: retried RPC
    note right of Server Application: successful processing
    Server Application->>Client Application: returns OK
```

### streaming RPCs
```mermaid
sequenceDiagram
    client->>gRPC stub: get users stream
	gRPC stub->>gRPC server: get users stream
	gRPC server->>database: get user 1
	database->>gRPC server: user 1
    gRPC server-->>client: user 1
	gRPC server->>database: get user 2
	database->>gRPC server: user 2
	gRPC server-->>client: user 2
	gRPC server->>database: get user 3
	database->>gRPC server: user 3
	gRPC server-->>client: user 3

	Kubernetes->>gRPC server: SIGTERM (e.g. on deployment)
	gRPC server->>gRPC server: initiate shutdown
	gRPC server->>client: GOAWAY "UNAVAILABLE" (closing connection)
	gRPC server->>gRPC server: exit 0
	Kubernetes->>gRPC server: Start new server

	gRPC stub-xgRPC server: stub can't automatically retry, doesn't know at what user we stopped
```
TL;DR: streaming RPC’s can’t be retried automatically.
