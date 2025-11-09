# Training Guide: Build an Oracle CRUD API in Go

This hands-on guide walks junior developers through recreating this project on their laptops, step by step.

## Learning Objectives
- Use `godror` to connect Go apps to Oracle DB
- Structure a small service (db, repo, api) using packages
- Serve OpenAPI docs and use Postman for testing
- Configure a connection pool and avoid Oracle OCI pitfalls
- Hot‑reload SQL queries from a properties file without restarting

## Prerequisites
- Go 1.23+ (check with `go version`)
- Git and VS Code (or your editor of choice)
- Oracle Database access (choose one):
  - Local Oracle XE (Docker or native install), or
  - Remote Oracle instance you can reach (host/port/service)
- Oracle client setup for `godror` (choose one):
  - Native: Install Oracle Instant Client (Basic). Add its `bin` to PATH (Win) / `LD_LIBRARY_PATH` (Linux) / `DYLD_LIBRARY_PATH` (macOS).
  - Thin mode (if supported): Consult `godror` docs to enable driver thin mode.

## Path A: Run This Project
1) Clone and open
- `git clone <your-repo-url> && cd godror-crud`
- Open in VS Code.

2) Configure DB
- Edit `config/application.properties:1` and set:
  - `db.service`, `db.username`, `db.password`, `db.server`, `db.port`, `db.timezone`
- Optional pool tuning (recommended):
  - `db.pool.max_open = 20`
  - `db.pool.max_idle = 10`
  - `db.pool.conn_max_lifetime = 30m`
  - `db.pool.conn_max_idletime = 5m`

3) Create database objects
- Run the DDL in `README.md: Database DDL` to create `PRODUCTS`.
- Optionally create a sequence and set `nextProductId` to `SELECT PRODUCTS_SEQ.NEXTVAL FROM DUAL`.

4) Build and run
- `go run .`
- Or VS Code → Run and Debug → `Launch API`.

5) Smoke test
- Health: `curl http://localhost:8080/health`
- List:   `curl http://localhost:8080/products`
- Create: `curl -X POST http://localhost:8080/products -H "Content-Type: application/json" -d '{"name":"Widget"}'`
- Docs:   `http://localhost:8080/docs/openapi.yaml`

6) Import to Postman
- Import → Link → `http://localhost:8080/docs/openapi.yaml` → Generate Collection.

7) Hot‑reload SQL
- Edit `config/application.properties:1` queries (e.g., add an `ORDER BY NAME`). Save.
- Re‑call the endpoint; changes apply within ~2 seconds (no restart).

## Path B: Recreate From Scratch
1) Initialize module
- `mkdir oracle-crud && cd oracle-crud`
- `go mod init go-demo-crud`
- `go get github.com/godror/godror`

2) Create folders
- `internal/db`, `internal/model`, `internal/repo`, `internal/api`, `internal/docs`, `internal/config`, `config`

3) Implement components
- DB connection: `internal/db/config.go` with `Config{ Service, Username, ... Timezone }` and `Open()` building a `godror` DSN.
- Domain model: `internal/model/product.go` with `Product` struct.
- Repo: `internal/repo/product_repo.go` with CRUD using `database/sql`.
- HTTP: `internal/api/handlers.go` for `/health`, `/products`, `/products/{id}` handlers and route registration.
- Docs: `internal/docs/openapi.yaml` (OpenAPI 3) and `internal/docs/serve.go` using `embed` to serve `/docs/openapi.yaml`.
- Properties: `internal/config/properties.go` to poll+reload SQL queries from `config/application.properties` (every ~2s).
- App config: `internal/config/app.go` to load DB settings from the same properties file.
- Main: `main.go` wires config → db → repo → api and starts server on `:8080`.

4) Add config files
- `config/application.properties` with `db.*` keys and SQL statements.

5) Run and test
- Use the same steps as Path A, sections 4–7.

## Best Practices and Tips
- Pooling: Keep `db.pool.max_idle > 0` to avoid OCI reconnect churn. Consider driver session pooling (set DSN params `standaloneConnection=0 poolMinSessions=1`).
- Timezone: Set DSN `timezone` to match DB `SYSTIMESTAMP` (often `+00:00`) to avoid warnings and timestamp confusion.
- Errors: Check logs for config reload issues; the app keeps last good SQL on parse failures.
- BLOB/JSON: For `attributes_blob`, send base64; for large text, prefer CLOB.

## Exercises
- Replace MAX+1 ID generation with a sequence; update queries via properties.
- Add a search endpoint (e.g., `/products?name=...`), including OpenAPI updates.
- Implement server‑side validation for product name length.
- Add basic middleware for request logging.

## Troubleshooting
- `discrepancy between SESSIONTIMEZONE and SYSTIMESTAMP`: set `timezone="+00:00"` (or your local zone) in the DSN.
- `OCI` library load errors: verify Instant Client install and PATH/LD_LIBRARY_PATH.
- Build status `0xc000013a` on Windows: usually indicates the process was interrupted (debugger stopped).
- Permission/auth errors: confirm Oracle credentials and network reachability.

## Reference Files in This Repo
- `main.go:1` — wiring and server startup
- `internal/db/config.go:1` — DB config, DSN, pool application
- `config/application.properties:1` — DB settings + SQL queries (hot‑reloaded)
- `internal/config/properties.go:1` — query reload manager
- `internal/config/app.go:1` — load DB config from properties
- `internal/repo/product_repo.go:1` — CRUD SQL
- `internal/api/handlers.go:1` — HTTP handlers and routes
- `internal/docs/openapi.yaml:1` — API schema (import into Postman)
