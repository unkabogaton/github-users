# GitHub Users Service

A GoLang service that exposes both REST and gRPC interfaces. The service also supports caching with Redis and storage with MySQL.

---

## Prerequisites

- Go >= 1.24.4
- Docker & Docker Compose
- `protoc` (for generating gRPC code)
- GitHub API token (set in `.env`)

---

## Running the Sync Worker (Fetch GitHub Users)

1. Make sure your `.env` is loaded:

```bash
source .env
```

2. Run the sync worker:

```bash
go run ./cmd/sync-users
```

- This will fetch users from the GitHub API and store them in MySQL.
- Uses concurrency and worker pool for faster fetches.
- Supports rate limiting and retries.

---

## Running Locally with Docker

1. Copy `.env.example` to `.env` and fill in your credentials:

```bash
cp .env.example .env
```

2. Start all services (MySQL, Redis, REST server, gRPC server):

```bash
docker compose up --build
```

- REST server will be available at `localhost:8080`
- gRPC server will be available at `localhost:9090`

---

## Generating gRPC Code from `.proto` Files

To regenerate Go code from `.proto` definitions:

```bash
protoc \
  -I api/proto \
  --go_out=internal/infrastructure/grpc/gen --go_opt=paths=source_relative \
  --go-grpc_out=internal/infrastructure/grpc/gen --go-grpc_opt=paths=source_relative \
  api/proto/users.proto
```

- This will generate code in `internal/infrastructure/grpc/gen`.

---

## Notes

- Make sure Docker Compose is running MySQL and Redis before starting the services.
- The REST and gRPC services share the same business logic via a service layer (internal/application/services).
- Migrations are handled with [Goose](https://github.com/pressly/goose).
