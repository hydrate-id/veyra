# Project Overview

Veyra is a single-binary local ingress controller / reverse proxy control plane based on Caddy and BoltDB.

## Setup Commands

- Get dependencies: `go mod download`
- Run the CLI tool locally: `go run main.go`

## Build & Test

- Run all unit and integration tests: `go test ./...`
- Run tests with coverage: `go test -coverprofile=coverage.out ./...`
- Build production binary: `go build -o veyra main.go`

## Architecture

- `main.go`: Entrypoint that invokes the Cobra root command execution.
- `cmd/`: Cobra CLI commands (e.g., `start`, `add`, `remove`, `reload`, `list`, `get`, `validate`, `export`, `import`).
- `core/`: Core domain logic including rule validation, routing table registry, and Caddyfile generator.
- `caddy/`: Caddy integration wrapper handling adaptation of Caddyfiles, running embedded Caddy instances, and dynamic configuration reloads via HTTP POST.
- `store/`: BoltDB wrapper for persistence of reverse proxy rules and key-value configuration settings.
- `model/`: Data structures representing rules and backup objects.

## Code Style

- Follow standard Go formatting rules via `gofmt`.

## Boundaries

- Absolutely no comments (`//` or `#`) are allowed in newly added code blocks.
- All new CLI features must be supported by tests to maintain statement coverage above 90%.
- HTTPS is disabled by default for the local proxy. It can be enabled using the `--https` flag on `veyra start`.
- Silent Caddy reloads: During `add` and `remove`, failures in reloading Caddy should remain silenced unless `--verbose` (or `-v`) is explicitly set. During `reload`, failures should print a simplified error message unless `--verbose` (or `-v`) is set, which prints the raw trace.
