## GUH — Go Universal Helper

Go Universal Helper (GUH) is a CLI and a collection of Go packages that help you bootstrap and operate Go services without repeating the same setup every time. It creates project structure, initializes Go modules, wires Docker Compose, scaffolds config and logging, and includes reusable libs for logging, env handling, DB access, retries, HTTP helpers, and more.

### Highlights
- **CLI tools**: `init`, `structure`, `mod`, `compose`, `config`, `api`, `help`, `-v`
- **CLI tools**: `init`, `structure`, `mod`, `compose`, `config`, `api`, `db`, `help`, `-v`
- **Batteries included**: Dockerfile, Docker Compose for Postgres, `.env` conventions, Gin-friendly HTTP scaffolding
- **Libraries**: logging, env loader, Postgres DB wrapper, retry/timer utilities, HTTP helpers, project config


## Installation

### Prerequisites
- Go (matching your environment; Dockerfile targets Go 1.24)
- Git (required by `mod --github`)
- Docker and Docker Compose (required by `compose` and `api --serve`)
- curl (for `api --get` convenience)

### Install the CLI

Option 1: Install from source
```bash
git clone https://github.com/Arthur-Conti/guh.git
cd guh
go build -o guh
./guh -v
```

Option 2: Install via `go install`
```bash
go install github.com/Arthur-Conti/guh@latest
# ensure $GOPATH/bin or $GOBIN is on your PATH
guh -v
```


## Quick start
```bash
# 1) Create base structure (required to persist service info)
guh structure --create --serviceName=my-service

# 2) Initialize go.mod and optionally add Gin
guh mod --github=github.com/you/my-service --gin

# 3) Create Docker Compose with Postgres, add your service, and run it
guh compose --dbName=Postgres --addService --run

# 4) Generate config files (logger + init)
guh config --all

# 5) Serve and test
guh api --serve &            # or use --bg
guh api --get=/health
guh api --kill
```

### Faster start with `init`
If you prefer a one-command setup, use `guh init` to orchestrate structure, module init, Docker Compose, and configs in a single step:
```bash
guh init --serviceName=my-service \
  --dbName=postgres \
  --github=github.com/you/my-service \
  --gin \
  --all
```


## CLI reference

Global usage:
```bash
guh <command> [flags]
```

Global commands:
- `guh help`: prints help and command list
- `guh -v`: prints the GUH version (e.g., `GUH version: dev`)

### compose — generate and run Docker Compose
Creates a `docker-compose.yml` with the services you need (currently Postgres) and can attach your service container.

Flags:
- `--dbName` (string, default: `Postgres`): database to include. Currently supported: `Postgres`.
- `--addService` (bool): add your app service to the compose file (uses service name from project config).
- `--run` (bool): run `docker compose up --build -d` after creating/updating the file.
- `--help`: show help for this command.

Examples:
```bash
guh compose --dbName=Postgres
guh compose --addService
guh compose --run
```

Notes:
- The Postgres service is created with envs loaded from `.env` (see Env section below).
- `--addService` requires a saved service name. Run `guh structure --create --serviceName=<name>` first.

### config — scaffold config files
Creates configuration boilerplate under a target path (defaults to `./internal/config/`).

Flags:
- `--filePath` (string, default: `./internal/config/`): target directory for generated files.
- `--all` (bool): create all config files (currently `init.go` and `logger.go`).
- `--logger` (bool): create only `logger.go`.
- `--help`: show help for this command.

Generated files (when `--all`):
- `init.go` with `config.Init()` that sets up logging
- `logger.go` with a default logger wired to a plain stdout output

Examples:
```bash
guh config --all
guh config --filePath=./config/ --logger
```

### mod — initialize go.mod and optionally sync with GitHub
Helps initialize your module name and optionally download useful packages.

Flags:
- `--github` (string): full GitHub module path; sets git remotes and runs `go mod init <github-url>`.
- `--gin` (bool): download Gin (`github.com/gin-gonic/gin`).
- `--help`: show help for this command.

Behavior:
- If `--github` is provided, GUH runs: `git init`, `git remote add origin <url>`, `git branch -M main`, and `go mod init <url>`.
- If `--github` is not provided, GUH uses the saved `serviceName` (from project config) as the module name.
- GUH then cleans/tidies the module and fetches GUH packages: `go get -u github.com/Arthur-Conti/guh@v0.1.0`.

Examples:
```bash
guh mod --github=github.com/you/my-service --gin
guh mod --gin
```

Requirements:
- Git must be installed and in PATH when using `--github`.

### structure — bootstrap project layout
Creates the initial directory tree, a `Dockerfile`, a starter `.env`, and a minimal Gin server setup.

Flags:
- `--serviceName` (string, required): name of your service; persisted as project metadata.
- `--create` (bool): create the directories and files.
- `--showFirst` (bool): preview the structure and confirm before creating.
- `--help`: show help for this command.

What gets created (high-level):
```
├── cmd/
│   └── main.go
├── internal/
│   ├── domain/
│   ├── config/
│   ├── application/
│   │   └── services/
│   └── infra/
│       ├── http/
│       │   ├── controllers/
│       │   └── routes/
│       │      └── routes.go
│       └── repositories/
├── Dockerfile
└── .env
```

Examples:
```bash
guh structure --create --serviceName=my-service
guh structure --showFirst --serviceName=my-service
```

### api — run, stop, and extend your application API
Convenience command to serve the app using Docker Compose and perform quick actions.

Flags:
- `--serve` (bool): run `docker compose up --build` and stream logs; with `--bg` runs detached.
- `--bg` (bool): if used with `--serve`, runs in background (`-d`).
- `--kill` (bool): run `docker compose down` to stop the stack.
- `--newRoute` (string): add a new route group (creates `internal/infra/http/routes/<route>.go` and registers it).
- `--get` (string): send a GET `curl` request to the app base URL (`http://localhost:8080`) with the given path.
- `--help`: show help for this command.

Examples:
```bash
guh api --serve
guh api --serve --bg
guh api --kill
guh api --newRoute=/orders
guh api --get=/health
```

Notes:
- `--newRoute` expects a Gin setup and adds a stub like `func OrdersRoutes(group *gin.RouterGroup) {}`. It also registers the group in `internal/infra/http/routes/routes.go`.
- Base URL for `--get` is `http://localhost:8080`.

### help and version
```bash
guh help
guh <command> --help
guh -v
```

### init — bootstrap everything quickly
### db — database migrations and seeds (Postgres)
Manages SQL file-based migrations and seeds. Defaults:
- Migrations: `./internal/infra/db/migrations`
- Seeds: `./internal/infra/db/seeds`

Flags:
- `--dir` (string, default: `./internal/infra/db/migrations`): directory for migration files
- `--init` (bool): create the migrations directory and `schema_migrations` table
- `--new` (string): create a new migration pair `<timestamp>_<name>.up.sql` and `.down.sql`
- `--up` (bool): apply all pending migrations in order
- `--down` (bool): revert the last N migrations (use `--steps`)
- `--steps` (int, default: 1): number of steps for `--down`
- `--status` (bool): show migration status (applied vs pending)
- `--seedDir` (string, default: `./internal/infra/db/seeds`): directory for seeds
- `--initSeeds` (bool): create seeds directory and `schema_seeds` table
- `--newSeed` (string): create a new seed file `<name>.sql`
- `--seed` (bool): apply all pending seeds
- `--seedStatus` (bool): show seed status

Examples:
```bash
guh db --init
guh db --new=create_users_table
guh db --up
guh db --down --steps=1
guh db --status
guh db --initSeeds
guh db --newSeed=seed_users
guh db --seed
guh db --seedStatus
```

Notes:
- Uses `.env` (`DB_USER`, `DB_PASS`, `DB_IP`, `DB_PORT`, `DB_DATABASE`) via `libs/db` conventions.
- Creates and reads `schema_migrations` to track applied versions.
- File naming format: `<YYYYMMDDHHMMSS>_<snake_case_name>.up.sql|.down.sql`.
- Seeds use `schema_seeds` (`name`, `applied_at`) and run in lexical order once.

Run a one-shot project bootstrap that creates structure, initializes `go.mod`, prepares Docker Compose (Postgres + app service), and generates configs.

Flags:
- `--serviceName` (string, required): the name of your service.
- `--dbName` (string, default: `postgres`): database to provision in Docker Compose.
- `--github` (string): GitHub module path; if provided, GUH configures git remotes and runs `go mod init <github-url>`. If omitted, GUH uses `serviceName` as module name.
- `--gin` (bool): download Gin.
- `--all` (bool): generate all config files (logger + init).
- `--showFirst` (bool): preview the structure and confirm before creating.
- `--help`: show help for this command.

Example:
```bash
guh init --serviceName=my-service \
  --dbName=postgres \
  --github=github.com/you/my-service \
  --gin \
  --all
```


## Configuration and conventions

### Environment variables (`.env`)
GUH’s Postgres helpers expect these variables to be present. The structure command seeds a starter file; adjust as needed.
```dotenv
DB_USER=user_test
DB_PASS=pass_test
DB_IP=localhost
DB_PORT=5432
DB_DATABASE=default
```

### Docker Compose
`guh compose --dbName=Postgres` generates a `docker-compose.yml` with:
- `postgres` service (image `postgres:15`, volume, port mapping)
- Credentials pulled from your `.env`
- `--addService` attaches your app service with `build: .`, port `8080:8080`, and `depends_on: [postgres]`.

### Generated config
- `config.Init()` initializes `config.Config.Logger` using the default plain stdout output and `debug` level.
- Use it in your `main()` before invoking commands (already done in GUH’s own `main.go`).


## Libraries overview

GUH ships with reusable packages under `libs/` that you can import in your services.

- `libs/log/*`: structured logger with outputs and application package tagging
  - Initialize via the generated `config.Init()` or construct manually using `outputs.NewPlainOutput` and `logger.NewLogger`.

- `libs/env_handler` and `libs/env_handler/env_locations`: simple env loader using `joho/godotenv`
  - Example: `env := env_handler.NewEnvs(env_locations.NewLocalEnvs("./.env")); env.EnvLocation.LoadDotEnv()`

- `libs/db`: Postgres helper with sane defaults
  - Quick connect using `.env` defaults:
    ```go
    p, err := db.DefaultPostgres()
    if err != nil { panic(err) }
    defer p.Close()
    ```
  - Query into a struct (matching fields by `db:"column"` tag or lowercased field name):
    ```go
    type User struct { ID int `db:"id"`; Name string `db:"name"` }
    var u User
    _ = p.QueryRow(&u, "select id, name from users where id = $1", 1)
    ```
  - Query into a slice of structs:
    ```go
    var users []User
    _ = p.Query(&users, "select id, name from users")
    ```

- `libs/http_handler`: HTTP helpers (see package for details)
- `libs/retry_handler`: small retry utilities
- `libs/timer`: simple timing utilities
- `libs/project_config`: reads/writes project metadata (e.g., service name, module)


## Troubleshooting
- Docker command not found: install Docker and ensure `docker` is in PATH.
- `compose --addService` complains about service name: run `guh structure --create --serviceName=<name>` first.
- `mod --github` fails on git commands: ensure Git is installed and that the URL is correct.
- Port `8080` in use: change your app or compose mappings accordingly.


## Contributing
Issues and PRs are welcome. If you add new commands or flags, please update this README and the per-command help text.
