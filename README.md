# go-learn-grpc

gRPC in Go, one chapter per concept, built around a single product catalog service.
The contract lives in [catalog](catalog), the server implementation in [internal/catalog](internal/catalog).

Generate proto code:

    go generate ./...

Run a chapter:

    go run ./examples/c01-unary

| # | Chapter | Concepts |
|---|---------|----------|
| 01 | [unary](examples/c01-unary) | Proto contract, codegen, unary RPC |
| 02 | [streaming](examples/c02-streaming) | Server, client, and bidirectional streaming |
| 03 | [errors](examples/c03-errors) | Status codes, error handling, deadlines |