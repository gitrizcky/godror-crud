# Go Demo CRUD (Oracle + godror)

Simple CRUD HTTP API for products backed by Oracle Database using the `godror` driver. The project demonstrates a clean structure (db/config, repo, api handlers), OpenAPI docs, and hot‑reloadable SQL queries via a Java‑style `application.properties` file.

## Features
- REST endpoints for products (list/get/create/update/delete)
- OpenAPI 3 spec served from the app
- Hot‑reload SQL queries from `config/application.properties` without restart
- Oracle connection via `godror` with explicit timezone to avoid timestamp warnings

## Project Layout
- `main.go` — wires config → db → repo → HTTP routes
- `internal/db/` — DB config and connection (`Config`, `Open`)
- `internal/model/` — domain structs (`Product`)
- `internal/repo/` — data access (`ProductRepo`)
- `internal/api/` — HTTP handlers + route registration
- `internal/docs/` — embedded OpenAPI (`openapi.yaml`)
- `internal/config/` — properties loader with polling hot‑reload
- `config/application.properties` — SQL statements you can edit live

## Training Guide
- Step-by-step hands-on for juniors: see `docs/TRAINING.md:1`.

## Prerequisites
- Go 1.23+ (module in `go.mod`)
- Oracle Database you can reach (host/port/service)
- Oracle client options for `godror`:
  - Easiest: Install Oracle Instant Client (Basic) and ensure the library directory is on `PATH` (Windows) / `LD_LIBRARY_PATH` (Linux) / `DYLD_LIBRARY_PATH` (macOS).
  - Alternative: Use `godror` thin mode if supported by your version. Check the `godror` docs for the `thin` connection parameter; thin mode avoids native client libs.

## Configure
- Database connection is set in `main.go` using `internal/db.Config`:
  - `Service`, `Username`, `Server`, `Port`, `Password`, `Timezone`.
  - The DSN built in `internal/db/config.go` includes `timezone="+00:00"` by default to keep timestamp semantics predictable.
- SQL statements live in `config/application.properties` and are hot‑reloaded every ~2s. Keys:
  - `listProducts`, `getProduct`, `nextProductId`, `insertProduct`, `updateProduct`, `deleteProduct`.

### Connection Pool
You can tune the standard `database/sql` pool via properties (all optional):

```
db.pool.max_open = 20              # SetMaxOpenConns
db.pool.max_idle = 10              # SetMaxIdleConns
db.pool.conn_max_lifetime = 30m    # SetConnMaxLifetime (e.g., 30m, 1h)
db.pool.conn_max_idletime = 5m     # SetConnMaxIdleTime
```

These are applied in `internal/db/Open` when present; leave blank to use defaults.

### godror Session Pooling (optional)
Enable the driver’s session pooling by adding these to `config/application.properties`:

```
# 0 enables session pool; 1 uses standalone connections
db.godror.standalone_connection = 0
db.godror.pool_min_sessions = 1
db.godror.pool_max_sessions = 20
db.godror.pool_increment = 1
```

The DSN is automatically extended with these parameters in `internal/db/Open`.

Example `config/application.properties` snippet:
```
listProducts = SELECT PRODUCT_ID, NAME, ATTRIBUTES_VARCHAR, ATTRIBUTES_CLOB, ATTRIBUTES_BLOB FROM PRODUCTS ORDER BY PRODUCT_ID
getProduct = SELECT NAME, ATTRIBUTES_VARCHAR, ATTRIBUTES_CLOB, ATTRIBUTES_BLOB FROM PRODUCTS WHERE PRODUCT_ID = :1
```

## Run
- Via Go:
  - `go run .`
- Via VS Code:
  - Use the `Launch API` configuration in `.vscode/launch.json`.
- Server listens on `http://localhost:8080`.

## Endpoints
- GET `/health`
- GET `/products`
- POST `/products`
- GET `/products/{id}`
- PUT `/products/{id}`
- DELETE `/products/{id}`

## API Docs
- OpenAPI spec: `GET /docs/openapi.yaml`
- Import to Postman:
  - Import → Link → `http://localhost:8080/docs/openapi.yaml` (or import the file at `internal/docs/openapi.yaml`).

## Hot‑Reloadable Queries
- The config watcher in `internal/config/properties.go` polls `config/application.properties` every ~2s.
- When you save changes, the repo methods immediately use the latest SQL without restarting the service.
- Invalid edits keep the last good config; errors are logged.

## Using godror to Connect to Oracle
- This project uses the key/value DSN form in `internal/db/config.go`:
  - `user="..." password="..." connectString="host:port/service" timezone="+00:00"`
- Timezone handling:
  - If you see a warning like: discrepancy between `SESSIONTIMEZONE` and `SYSTIMESTAMP`, align them.
  - We set `timezone="+00:00"` (UTC) in the DSN to match many DBs’ `SYSTIMESTAMP`.
  - Alternatives:
    - Use an IANA name: `timezone="Asia/Bangkok"` or `timezone="Local"`.
    - Or run: `ALTER SESSION SET TIME_ZONE = DBTIMEZONE` right after connect.
- Native client vs Thin mode:
  - With the native client, ensure Instant Client libraries are discoverable by your OS.
  - Thin mode (if enabled via driver params) avoids native libs; check your `godror` version docs for exact flags.

## Notes
- The sample `Product` table and credentials in the code are placeholders. Update them to match your environment.
- For binary data (`attributes_blob`), send base64 in JSON.

## License
- This repository does not include a license. Add one if you plan to distribute.

## Database DDL
Use the following SQL to create the `PRODUCTS` table used by the API.

```sql
CREATE TABLE PRODUCTS (
  PRODUCT_ID         NUMBER       NOT NULL,
  NAME               VARCHAR2(200) NOT NULL,
  ATTRIBUTES_VARCHAR VARCHAR2(4000),
  ATTRIBUTES_CLOB    CLOB,
  ATTRIBUTES_BLOB    BLOB,
  CONSTRAINT PK_PRODUCTS PRIMARY KEY (PRODUCT_ID)
);
```

The application, by default, computes the next `PRODUCT_ID` as `NVL(MAX(PRODUCT_ID), 0) + 1`. For better concurrency, consider using a sequence and update the app’s query via `config/application.properties`.

Option A — Use a sequence (recommended):
```sql
CREATE SEQUENCE PRODUCTS_SEQ START WITH 1 INCREMENT BY 1 NOCACHE NOCYCLE;
```
Then set in `config/application.properties`:
```
nextProductId = SELECT PRODUCTS_SEQ.NEXTVAL FROM DUAL
```

Option B — Identity column (Oracle 12c+):
```sql
-- If creating a new table with identity
CREATE TABLE PRODUCTS (
  PRODUCT_ID         NUMBER GENERATED BY DEFAULT AS IDENTITY PRIMARY KEY,
  NAME               VARCHAR2(200) NOT NULL,
  ATTRIBUTES_VARCHAR VARCHAR2(4000),
  ATTRIBUTES_CLOB    CLOB,
  ATTRIBUTES_BLOB    BLOB
);
```
If you choose identity, also change inserts to omit `PRODUCT_ID` and adapt the app accordingly (or use `RETURNING PRODUCT_ID INTO :id`).


Optional: Driver Session Pooling (godror)
If you want to also use godror’s session pool, we can add DSN params:

standaloneConnection=0
poolMinSessions=1
poolMaxSessions=20
poolIncrement=1
Example DSN (in internal/db/config.go) would become:

user="..." password="..." connectString="host:port/service" timezone="+00:00" standaloneConnection=0 poolMinSessions=1 poolMaxSessions=20 poolIncrement=1

Why

MaxIdleConns > 0 ensures connections are reused, preventing rapid connect/close loops that can trigger OCI issues.
Session pooling in godror further reduces churn at the Oracle session level.
