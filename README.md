# go-forge

Scaffold production-ready Go REST API projects in seconds.

```bash
forge new my-api
```

Like `create-react-app` but for Go backends — opinionated, complete, and ready to deploy.

## What it generates

```
my-api/
├── cmd/api/main.go              # entry point — wires config, db, router
├── internal/
│   ├── config/config.go         # env var loading (godotenv)
│   ├── database/db.go           # sqlx PostgreSQL connection + pool
│   ├── auth/jwt.go              # JWT token generation + parsing
│   ├── auth/middleware.go       # chi JWT middleware
│   ├── handler/                 # HTTP handlers (health, auth, users)
│   ├── model/user.go            # domain structs with db/json tags
│   ├── repository/users.go      # DB queries via sqlx (no ORM)
│   ├── server/server.go         # chi router + middleware wiring
│   └── validator/request.go     # go-playground/validator helpers
├── pkg/logger/logger.go         # slog JSON logger
├── pkg/response/json.go         # JSON response helpers
├── migrations/001_create_users.sql
├── api/requests.http            # VS Code REST Client test file
├── Dockerfile                   # multi-stage build → scratch image
├── docker-compose.yml           # postgres + api services
├── Makefile                     # dev, build, test, migrate, docker targets
└── .env.example                 # all required env vars documented
```

## Install

```bash
go install github.com/akshadjaiswal/go-forge@latest
```

> This installs the `forge` command (binary is named `forge`, not `go-forge`).

Or download a binary from [Releases](https://github.com/akshadjaiswal/go-forge/releases).

## Usage

### Interactive (recommended)

```bash
forge new my-api
```

Prompts for project name, module path, and which features to include.

### Non-interactive (CI / scripting)

```bash
# Full stack: auth + docker (defaults)
forge new my-api --module github.com/username/my-api

# Skip auth
forge new my-api --module github.com/username/my-api --no-auth

# Skip docker
forge new my-api --module github.com/username/my-api --no-docker
```

## Stack

Every generated project uses:

| Layer | Technology |
|-------|-----------|
| Router | [chi](https://github.com/go-chi/chi) |
| Database | PostgreSQL + [sqlx](https://github.com/jmoiron/sqlx) (no ORM) |
| Auth | JWT ([golang-jwt](https://github.com/golang-jwt/jwt)) + bcrypt |
| Logging | Go 1.21+ `log/slog` (structured JSON) |
| Validation | [go-playground/validator](https://github.com/go-playground/validator) |
| Config | `.env` + [godotenv](https://github.com/joho/godotenv) |
| Docker | Multi-stage build → `scratch` image (~10MB) |

## After generating

```bash
cd my-api
cp .env.example .env       # fill in DATABASE_URL and JWT_SECRET
psql -U postgres -c "CREATE DATABASE my_api_db;"
make migrate               # run SQL migrations
make dev                   # start server on :8080
```

Open `api/requests.http` in VS Code with the [REST Client](https://marketplace.visualstudio.com/items?itemName=humao.rest-client) extension to test all endpoints.

## Endpoints included

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| GET | `/health` | — | Health check |
| POST | `/auth/register` | — | Register new user |
| POST | `/auth/login` | — | Login, returns JWT |
| GET | `/users/{id}` | Bearer | Get user by ID |

## Why go-forge?

Existing Go scaffolding tools either:
- Are too generic (no auth, no Docker, no logging)
- Are framework-locked (Buffalo requires its ORM)
- Are web-focused (HTML templates, not REST APIs)

go-forge generates a complete, opinionated stack that you can run immediately — the same patterns used in [go-backend-production](https://github.com/akshadjaiswal/go-backend-production).

## Author

**Akshad Jaiswal** — built while learning Go production patterns.
