# Database Backup Utility (dbbackup)

A cross-DB CLI to backup and restore databases (MySQL, PostgreSQL, MongoDB, SQLite).

> Project idea from: https://roadmap.sh/projects/database-backup-utility

This project is a working utility and scaffolding that demonstrates practical, production-minded approaches:

- CLI built with Cobra (`backup`, `restore`, `serve`)
- Full and selective backups using native dump tools (mysqldump, pg_dump, mongodump) or file-copy for SQLite
- Streaming gzip compression helpers to avoid double-writing large dumps
- Local storage and AWS S3 upload (using AWS SDK v2)
- Basic logging and optional Slack webhook notifications
- Driver-based Postgres CSV restore (no external `psql`) with optional upsert keys
- Integration test (dockerized Postgres) that validates CSV restore/upsert behavior

This README documents how to install, use, and test the project.

## Quick install

Requirements:

- Go 1.20+ (module-enabled)
- Docker (for integration tests; optional for core functionality)
- Native DB clients if you rely on external tools: `mysqldump`, `pg_dump`, `mongodump`, `psql`, `mysql` (not required for driver-based restore)

To fetch dependencies and build (recommended):

```bash
git clone <repo>
cd database-utility
go mod tidy
# build a binary named `dbbackup` (this keeps CLI name and binary name consistent)
go build -o dbbackup .
```

Notes:

- The repository (module) name is `database-utility`. The CLI tool defined by the code uses the Cobra root command name `dbbackup` (so help and examples show `dbbackup`).
- You can name the output binary anything you like; the examples in this README assume you build an executable named `dbbackup`.
- Alternatively run without producing a binary:

```bash
go run . --help
```

## CLI overview

Main commands:

- `backup` — create backups
- `restore` — restore databases (full and selective)
- `serve` — run a scheduler to execute configured backup jobs

Run the generic help to see flags:

```bash
./dbbackup --help
./dbbackup backup --help
./dbbackup restore --help
```

### Common flags (examples)

- `--type` — database type: `mysql`, `postgres`, `mongodb`, `sqlite`
- `--host`, `--port`, `--user`, `--password`, `--dbname` — connection params
- `--out` — output directory for local backups
- `--gzip` — compress output (or use `.gz` extension)
- `--s3-bucket`, `--s3-key` — upload to AWS S3 after backup
- `--stream` — stream dump stdout directly to gzip and optionally S3 (recommended for large DBs)

## Examples

Local SQLite backup:

```bash
./dbbackup backup --type=sqlite --dbname=/path/to/db.sqlite --out=./backups
```

MySQL full backup to local file (requires `mysqldump`):

```bash
./dbbackup backup --type=mysql --host=127.0.0.1 --port=3306 --user=root --password=secret --dbname=mydb --out=./backups
```

Stream a live MySQL dump directly to S3 (recommended for large DBs):

```bash
./dbbackup backup --type=mysql --host=... --user=... --password=... --dbname=... --stream --s3-bucket=my-bucket --s3-key=backups/mydb-$(date +%F).sql.gz
```

Postgres selective (table-level) CSV export (incremental-ish):

```bash
./dbbackup backup --type=postgres --host=... --user=... --password=... --dbname=mydb --tables=events,users --where="updated_at > '2025-10-01'" --out=./backups
```

Restore examples:

Restore a sqlite backup (file copy):

```bash
./dbbackup restore ./backups/sample.sqlite.sql --type=sqlite --dbname=/path/to/target.sqlite
```

Restore MySQL from SQL file (requires `mysql` client):

```bash
./dbbackup restore ./backups/mydb.sql --type=mysql --host=127.0.0.1 --port=3306 --user=root --password=secret --dbname=mydb
```

Postgres: restore using driver-based CSV upsert (no external `psql` required):

```bash
./dbbackup restore ./backups/users.csv --type=postgres --mode=selective --tables=users --driver-restore --upsert-keys=id --host=... --user=... --password=... --dbname=mydb
```

If you have a directory containing per-table CSVs (e.g., `./backups/2025-10-16/`), point `restore` at the directory and specify `--tables=...`.

## Integration tests (Postgres)

This repo includes an integration test that starts a Postgres container and verifies the Postgres CSV driver restore and upsert functionality.

Prerequisites:

- Docker installed and the user can access the Docker socket
- Go toolchain

Run only the integration test (slow):

```bash
go test ./internal/db -run TestRestorePostgresCSVDriver -v
```

Notes:

- The test uses `github.com/ory/dockertest` and will skip automatically if Docker is not available or the test is being run with `-short`.
- On CI, ensure the runner has Docker and adequate permissions (or configure a job to run the integration test on a self-hosted runner).

## Scheduling (service mode)

Create a JSON `jobs.json` describing shell commands to run on cron schedules:

```json
[
  {
    "name": "daily-mysql",
    "cron": "0 0 * * *",
    "command": "./dbbackup backup --type=mysql --host=127.0.0.1 --user=root --password=secret --dbname=mydb --out=/var/backups"
  }
]
```

Start the scheduler:

```bash
./dbbackup serve --config ./jobs.json
```

This runs configured shell commands on the schedule using a simple cron scheduler.

## Logging & Notifications

- Basic logging goes to stdout. You can set a file output by calling logger setup in code or redirect stdout/stderr when running.
- Slack notifications: provide `--slack-webhook` flag to `backup` to POST a message when a job completes (success/failure). See `cmd/backup.go` flags.

## Storage adapters

- Local storage is the default (use `--out` to target a directory).
- AWS S3: provide `--s3-bucket` and `--s3-key` to upload backups to S3. AWS credentials are handled via the usual AWS SDK mechanisms (env, shared config).

Future/Planned:

- Add Google Cloud Storage and Azure Blob adapters
- Add secrets manager integration (Vault/AWS Secrets Manager)
- True incremental backups using DB WAL/PITR where applicable
- More integration tests across MySQL and MongoDB

## Troubleshooting

- Ensure external utilities are installed when using exec-based dumps: `mysqldump`, `pg_dump`, `mongodump`, `psql`, `mysql`.
- If S3 upload fails, check AWS credentials and environment variables (AWS_ACCESS_KEY_ID, AWS_SECRET_ACCESS_KEY, AWS_REGION).
- If the integration test is skipped, verify Docker availability and permissions (the test skips when Docker socket is unavailable).

## Contributing

Contributions are welcome. Please open issues or PRs for bugs, feature requests, or tests.

## License

This project is provided as-is for educational/demo purposes. Add your preferred license.

## End-to-end local demo with Docker Compose

The following small demo shows how to run a Postgres instance with Docker Compose, perform a backup of a single table using the CLI, then restore it back using the driver-based CSV restore. This is intended for local testing and demo purposes.

1. Create a `docker-compose.yml` in a temporary directory with the following content:

```yaml
version: '3.8'
services:
  postgres:
    image: postgres:14
    environment:
      POSTGRES_USER: test
      POSTGRES_PASSWORD: secret
      POSTGRES_DB: demo
    ports:
      - "5432:5432"
    volumes:
      - pgdata:/var/lib/postgresql/data

volumes:
  pgdata:
```

2. Start Postgres with Docker Compose:

```bash
docker-compose up -d
```

3. Create a sample table and insert data (use `psql` or any DB client). Example using `psql` (you can run this on the host if you have `psql` installed or `docker exec` into the container):

```bash
# using docker exec
docker exec -i $(docker ps -q -f ancestor=postgres:14) psql -U test -d demo <<'SQL'
CREATE TABLE users (id text PRIMARY KEY, name text, age integer);
INSERT INTO users (id, name, age) VALUES ('u1','Alice',30),('u2','Bob',25');
SQL
```

4. Run a selective CSV export with the CLI (exports the `users` table to `./backups`):

```bash
./dbbackup backup --type=postgres --host=localhost --port=5432 --user=test --password=secret --dbname=demo --tables=users --out=./backups
```

5. Modify the CSV to simulate changes (e.g., change Bob's age and add a new user):

```bash
# edit ./backups/<file>.csv (or create users.csv) and change/add rows
```

6. Restore using the driver-based CSV upsert:

```bash
./dbbackup restore ./backups/users.csv --type=postgres --mode=selective --tables=users --driver-restore --upsert-keys=id --host=localhost --port=5432 --user=test --password=secret --dbname=demo
```

7. Verify results (via `psql`):

```bash
docker exec -i $(docker ps -q -f ancestor=postgres:14) psql -U test -d demo -c "SELECT * FROM users ORDER BY id;"
```

8. When finished, stop and remove the containers (and optionally the volume):

```bash
docker-compose down -v
```

Notes:

- This is a lightweight demo; in production you'd manage credentials securely and use non-default network/volumes.
- If `psql` is not available locally, use `docker exec -it <container> psql ...` to run SQL commands inside the container.
