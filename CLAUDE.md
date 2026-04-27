# CLAUDE.md — go-forge

Context for Claude to work effectively on this project across sessions.

---

## What this project is

`go-forge` is a CLI tool that scaffolds production-ready Go REST API projects — like `create-react-app` but for Go backends. Run `forge new my-api` and get a complete, compiling project with chi router, sqlx + PostgreSQL, JWT auth, slog logging, Docker, and more.

**GitHub:** `git@github-personal:akshadjaiswal/go-forge.git`
**Module path:** `github.com/akshadjaiswal/go-forge`
**Owner:** Akshad Jaiswal (Go learner, JS/Node background)
**Latest release:** v0.1.0 — published on GitHub Releases with pre-built binaries
**Binary name:** `go-forge` — `go install` names the binary after the last module path segment, not `rootCmd.Use`

---

## Repository structure

```
go-forge/
├── main.go                              # entry point — calls cmd.Execute(version)
├── cmd/
│   ├── root.go                          # cobra root command + Execute()
│   └── new.go                           # `forge new` subcommand, prompt flow, flags
├── internal/
│   └── generator/
│       ├── options.go                   # ProjectOptions + TemplateData structs
│       ├── generator.go                 # core: embed FS, render loop, file writer
│       └── post.go                      # post-gen: go mod tidy, git init, initial commit
├── internal/generator/templates/        # ALL embedded templates (//go:embed all:templates)
│   ├── base/                            # always generated (no flags needed)
│   │   ├── cmd/api/main.go.tmpl
│   │   ├── internal/config/config.go.tmpl
│   │   ├── internal/server/server.go.tmpl
│   │   ├── internal/handler/health.go.tmpl
│   │   ├── pkg/logger/logger.go.tmpl
│   │   ├── pkg/response/json.go.tmpl
│   │   ├── go.mod.tmpl
│   │   ├── Makefile.tmpl
│   │   ├── .env.example.tmpl
│   │   ├── .gitignore.tmpl
│   │   └── README.md.tmpl
│   ├── database/                        # generated when WithAuth=true
│   │   └── internal/database/db.go.tmpl
│   ├── auth/                            # generated when WithAuth=true
│   │   ├── internal/auth/jwt.go.tmpl
│   │   ├── internal/auth/middleware.go.tmpl
│   │   ├── internal/handler/auth.go.tmpl
│   │   ├── internal/handler/users.go.tmpl
│   │   ├── internal/model/user.go.tmpl
│   │   ├── internal/repository/users.go.tmpl
│   │   ├── internal/validator/request.go.tmpl
│   │   └── migrations/001_create_users.sql.tmpl
│   ├── docker/                          # generated when WithDocker=true
│   │   ├── Dockerfile.tmpl
│   │   ├── .dockerignore.tmpl
│   │   └── docker-compose.yml.tmpl
│   └── requests/
│       └── api/requests.http.tmpl       # always generated
├── Makefile                             # build, install, test targets for forge itself
├── .gitignore                           # ignores bin/
└── README.md
```

---

## Key architecture decisions — read before touching generator code

### Template delimiters: `[[ ]]` not `{{ }}`

All `.tmpl` files use `[[` and `]]` as delimiters:
```go
template.New("").Delims("[[", "]]").Parse(...)
```
**Why:** Go source files use `{{ }}` in fmt strings, map literals, struct literals. Using `{{ }}` as template delimiters causes parse errors in generated code. `[[ ]]` avoids all collisions.

**Never change this.** If you add a new template, always use `[[ .FieldName ]]`, `[[ if .WithAuth ]]`, etc.

### Embed directive: `all:templates`

```go
//go:embed all:templates
var templateFS embed.FS
```
The `all:` prefix is required to include dot-files like `.env.example.tmpl` and `.gitignore.tmpl`. Without it, Go's embed skips files starting with `.`.

### Template group → feature flag mapping

| Template dir | Condition | What it generates |
|---|---|---|
| `base/` | always | main.go, config, server, health, logger, response, go.mod, Makefile, README, .env.example, .gitignore |
| `database/` | `WithAuth=true` | `internal/database/db.go` (sqlx connection pool) |
| `auth/` | `WithAuth=true` | JWT, middleware, auth/users handlers, model, repository, validator, migration |
| `docker/` | `WithDocker=true` | Dockerfile, .dockerignore, docker-compose.yml |
| `requests/` | always | `api/requests.http` |

**Critical:** `database/` and `auth/` must stay coupled — `auth/` imports `internal/database` and `internal/model`. Do not generate one without the other.

### Output path derivation

Template path → output path formula:
```
"templates/base/cmd/api/main.go.tmpl"
  strip prefix "templates/base/"  → "cmd/api/main.go.tmpl"
  strip suffix ".tmpl"            → "cmd/api/main.go"
  prepend project name            → "my-api/cmd/api/main.go"
```
Implemented in `generator.go:deriveOutputPath()`.

---

## ProjectOptions and TemplateData

```go
// What the user configures (from CLI flags or prompts)
type ProjectOptions struct {
    Name       string // "my-api"
    ModulePath string // "github.com/user/my-api"
    WithAuth   bool   // includes JWT auth + DB (default true)
    WithDocker bool   // includes Dockerfile + docker-compose (default true)
    GoVersion  string // auto-detected from runtime.Version()
}

// Passed into every .tmpl file during rendering
type TemplateData struct {
    ProjectName string // "my-api"
    ModulePath  string // "github.com/user/my-api"
    PackageName string // "myapi" (hyphens stripped, valid Go identifier)
    GoVersion   string // "1.22"
    DBName      string // "my_api_db"
    WithAuth    bool
    WithDocker  bool
}
```

---

## CLI flags

```
forge new [project-name]
  --module string    Go module path (skips module prompt, enables non-interactive mode)
  --no-auth          Skip JWT auth files (WithAuth=false)
  --no-docker        Skip Docker files (WithDocker=false)
```

**Interactive mode:** runs when `--module` is NOT provided — prompts for name (if not arg), module path, auth Y/n, docker Y/n using promptui.

**Non-interactive mode:** triggered when `--module` is provided — skips all confirm prompts, uses flag defaults for auth/docker.

---

## Generated project structure (with auth + docker)

```
my-api/                              ← 25 files
├── cmd/api/main.go
├── internal/
│   ├── config/config.go
│   ├── database/db.go              ← sqlx.DB wrapper type
│   ├── auth/jwt.go                 ← GenerateToken, ParseToken
│   ├── auth/middleware.go          ← chi JWT middleware + UserIDFromContext
│   ├── handler/health.go           ← GET /health
│   ├── handler/auth.go             ← POST /auth/register, POST /auth/login
│   ├── handler/users.go            ← GET /users/{id} (protected)
│   ├── model/user.go               ← User struct (db/json tags, password json:"-")
│   ├── repository/users.go         ← Create, GetByEmail, GetByID
│   ├── server/server.go            ← chi router, all routes wired
│   └── validator/request.go        ← DecodeAndValidate helper
├── pkg/logger/logger.go            ← slog JSON handler
├── pkg/response/json.go            ← JSON(), Error() helpers
├── migrations/001_create_users.sql
├── api/requests.http
├── Dockerfile
├── .dockerignore
├── docker-compose.yml
├── Makefile
├── .env.example
├── .gitignore
├── go.mod
└── README.md
```

---

## forge's own dependencies

```
github.com/spf13/cobra v1.10.x        # CLI framework
github.com/manifoldco/promptui v0.9.x  # interactive prompts
```

Nothing else. Keep forge's own dependency footprint minimal.

**go.mod minimum:** `go 1.21` (covers `log/slog` + `embed`). Do NOT raise this above the current Go stable release — CI uses `go-version: stable` and will fail if go.mod requires a pre-release version.

---

## CI / Release pipeline

```
.github/workflows/ci.yml       # build + vet + test on every push/PR to main
.github/workflows/release.yml  # GoReleaser on v*.*.* tags → GitHub Release
.goreleaser.yml                # multi-platform build config
```

**Release process:**
1. Work on a feature branch, open PR to main
2. CI must pass before merging
3. After merge, tag: `git tag -a vX.Y.Z -m "..." && git push origin vX.Y.Z`
4. GoReleaser auto-builds binaries for linux/darwin/windows × amd64/arm64

**No GITHUB_TOKEN setup needed** — Actions provides it automatically.

**Binary naming note:** `go install github.com/akshadjaiswal/go-forge@latest` installs a binary named `go-forge` (last path segment of module). GoReleaser pre-built archives contain a binary named `forge` (from `.goreleaser.yml binary: forge`). The two install paths produce different binary names — this is a known tradeoff, documented in README.

---

## Generated project dependencies

```
github.com/go-chi/chi/v5              # router (always)
github.com/go-playground/validator/v10 # validation (always)
github.com/joho/godotenv              # .env loading (always)
github.com/jmoiron/sqlx               # DB (WithAuth only)
github.com/lib/pq                     # postgres driver (WithAuth only)
github.com/golang-jwt/jwt/v5          # JWT (WithAuth only)
golang.org/x/crypto                   # bcrypt (WithAuth only)
```

---

## How to build and test forge

```bash
# Build binary (local dev)
make build          # → bin/forge

# Install to GOPATH/bin (installs as go-forge via go install)
make install

# Test generated output compiles
make integration-test

# Manual test — generate with all features
./bin/forge new test-api --module github.com/akshadjaiswal/test-api
cd test-api && go build ./... && go vet ./...

# Manual test — no auth, no docker
./bin/forge new bare-api --module github.com/akshadjaiswal/bare-api --no-auth --no-docker
cd bare-api && go build ./... && go vet ./...

# Install from GitHub and test
go install github.com/akshadjaiswal/go-forge@latest
go-forge new test-api --module github.com/akshadjaiswal/test-api
```

---

## What to do when adding a new template

1. Create `.tmpl` file in the correct group directory (`base/`, `auth/`, `docker/`, or new group)
2. Use `[[ .FieldName ]]` delimiters — never `{{ }}`
3. If new group: add a new dir under `templates/`, add it to `collectTemplatePaths()` in `generator.go`, add the flag to `ProjectOptions` + `TemplateData` + `cmd/new.go`
4. If new TemplateData field needed: add to both structs in `options.go` and populate in `BuildTemplateData()`
5. Re-run `go build ./...` on forge, then generate a test project and verify `go build ./...` on the output

---

## What to do when adding a new CLI flag

1. Add field to `ProjectOptions` in `options.go`
2. Add field to `TemplateData` in `options.go`, populate in `BuildTemplateData()`
3. Add `var flag...` + `newCmd.Flags().BoolVar(...)` in `cmd/new.go`
4. Wire into `runNew()` — check if `flagModule != ""` to decide whether to prompt or use flag
5. Add to `collectTemplatePaths()` in `generator.go` if it controls a template group

---

## Conventions — do not change

- Handler struct pattern: `type XHandler struct { ... }` with `NewXHandler()` constructor
- `writeJSON` / `response.JSON` — all handlers use pkg/response helpers, never raw `json.Marshal`
- `mustGetEnv` panics on missing required env — fail fast at startup, not runtime
- `sql.ErrNoRows` checked specifically before generic error → 404
- `RETURNING *` on all INSERT/UPDATE SQL
- Parameterized queries only — `$1, $2` placeholders, never string concat
- No ORM — raw sqlx + SQL throughout

---

## Do not

- Do NOT change template delimiters from `[[ ]]` to `{{ }}`
- Do NOT change the embed directive from `all:templates` to `templates/*` (breaks dot-files)
- Do NOT generate `database/` without `auth/` — `auth` imports model which only exists with auth
- Do NOT add heavy dependencies to forge itself (cobra + promptui is the full list)
- Do NOT commit without explicit user approval — user verifies manually first
- Do NOT add new generated features without updating this CLAUDE.md
