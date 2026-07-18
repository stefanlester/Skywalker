# Skywalker

[![CI](https://github.com/stefanlester/Skywalker/actions/workflows/ci.yml/badge.svg)](https://github.com/stefanlester/Skywalker/actions/workflows/ci.yml)

**Skywalker** is a batteries-included web framework for Go — an open re-implementation in the
lineage of the "Celeritas" course framework. It wires together routing, rendering, sessions,
caching, mail, a database + migration layer, pluggable remote filesystems, and a code-generating
CLI, so you can start building instead of assembling boilerplate.

One `App.New()` call reads your `.env`, creates the project layout, and boots every subsystem.

```go
func main() {
    sky := &skywalker.Skywalker{}
    if err := sky.New(rootPath); err != nil {
        log.Fatal(err)
    }
    sky.ListenAndServe()
}
```

## Features

- **Routing & middleware** — [chi](https://github.com/go-chi/chi) router, with CSRF (`nosurf`) and session middleware wired in.
- **Rendering** — [Jet](https://github.com/CloudyKit/jet) templates (default) or Go `html/template`, selected by `RENDERER`.
- **Sessions** — [scs](https://github.com/alexedwards/scs) with cookie, Redis, MySQL, or Postgres stores.
- **Caching** — Redis or [Badger](https://github.com/dgraph-io/badger) behind a single `Cache` interface.
- **Mail** — a queued mailer over channels; SMTP or an API service (Mailgun / SendGrid / SparkPost).
- **Database & migrations** — Postgres (pgx) or MySQL, with [golang-migrate](https://github.com/golang-migrate/migrate).
- **Remote filesystems** — one `FS` interface with four real backends: **MinIO, S3, SFTP, WebDAV**.
- **Extras** — authenticated AES-GCM encryption, HMAC-signed URLs, structured logging (`slog`: text in debug, JSON in production), a cron scheduler, and JSON/XML response helpers.
- **CLI** — scaffolds new apps, models, handlers, migrations, auth, sessions, and mail templates.

## Requirements

- **Go 1.25+** (both this module and generated apps build in module mode — no vendoring).
- Optional, depending on which subsystems you enable: Postgres/MySQL, Redis, an SMTP server, and a
  MinIO/S3/SFTP/WebDAV endpoint.

## Install the CLI

```bash
git clone https://github.com/stefanlester/Skywalker.git
cd Skywalker
go build -o skywalker ./cmd/cli      # Windows: -o skywalker.exe
```

Run `./skywalker help` to see every command.

## Create and run an app

```bash
./skywalker new myapp
```

This clones the [skywalker_engine](https://github.com/stefanlester/skywalker_engine) skeleton,
generates a fresh `.env` (with a random 32-char encryption key), rewrites the module path, and runs
`go mod tidy`. Then:

```bash
cd myapp
go run .                 # starts on the PORT set in .env (skeleton default: 4000)
```

Open <http://localhost:4000>. On first boot `New()` creates the standard folders: `handlers`,
`migrations`, `views`, `data`, `public`, `tmp`, `logs`, `middleware`.

**For the full guide — handlers, rendering, sessions, DB, cache, mail, filesystems, APIs — see
[docs/USAGE.md](docs/USAGE.md).**

## Configuration

Everything is driven by environment variables (loaded from `.env`). The common ones:

| Variable | Purpose |
|---|---|
| `PORT`, `APP_URL`, `SERVER_NAME`, `SECURE` | HTTP server |
| `DEBUG` | `true` puts Jet in development (re-parse) mode |
| `RENDERER` | `jet` or `go` |
| `DATABASE_TYPE` + `DATABASE_*` | `postgres` / `mysql`; empty = no DB |
| `CACHE` | `redis` or `badger`; empty = no cache |
| `SESSION_TYPE` | `cookie`, `redis`, `mysql`, or `postgres` |
| `COOKIE_*` | session cookie name/lifetime/domain/secure |
| `KEY` | AES key — **must be exactly 32 characters** |
| `SMTP_*` / `MAILER_API`,`MAILER_KEY`,`MAILER_URL` | SMTP or API mail |
| `MINIO_*`, `S3_*`, `SFTP_*`, `WEBDAV_*` | filesystem backends (see below) |

An app boots with **no external services** using `SESSION_TYPE=cookie`, empty `DATABASE_TYPE`, and
empty `CACHE` — handy for a first run.

## Remote filesystems

All four backends implement one interface:

```go
type FS interface {
    Put(fileName, folder string) error
    Get(destination string, items ...string) error
    List(prefix string) ([]Listing, error)
    Delete(itemsToDelete []string) bool
}
```

`New()` constructs each backend **only if its gate variable is set** — `MINIO_SECRET`, `S3_SECRET`,
`SFTP_HOST`, `WEBDAV_HOST` — and stores it in `App.FileSystems` keyed by name (`"MINIO"`, `"S3"`,
`"SFTP"`, `"WEBDAV"`). Pick one at request time through the interface:

```go
fs, ok := app.FileSystems[fsType].(filesystems.FS)   // fsType: "MINIO" | "S3" | "SFTP" | "WEBDAV"
if ok {
    files, _ := fs.List("/")
}
```

| Backend | Library | Key env vars |
|---|---|---|
| MinIO (any S3-compatible store) | minio-go | `MINIO_ENDPOINT`, `MINIO_KEY`, `MINIO_SECRET`, `MINIO_BUCKET`, `MINIO_USESSL`, `MINIO_REGION` |
| S3 | minio-go (S3-compatible) | `S3_ENDPOINT`, `S3_KEY`, `S3_SECRET`, `S3_REGION`, `S3_BUCKET` |
| SFTP | pkg/sftp | `SFTP_HOST`, `SFTP_USER`, `SFTP_PASS`, `SFTP_PORT` |
| WebDAV | studio-b12/gowebdav | `WEBDAV_HOST`, `WEBDAV_USER`, `WEBDAV_PASS` |

> SFTP currently uses `InsecureIgnoreHostKey`; verify host keys via `known_hosts` before production use.

> **Note on MinIO:** the MinIO *server* community edition is [no longer maintained upstream](https://github.com/minio/minio)
> (source-only, steered toward commercial AIStor). Skywalker is unaffected in code: both S3-family backends
> use the still-maintained [minio-go](https://github.com/minio/minio-go) client and speak the standard S3
> protocol, so they work with any S3-compatible server — [SeaweedFS](https://github.com/seaweedfs/seaweedfs)
> (the dev default bundled with the skeleton, verified end-to-end), [Garage](https://garagehq.deuxfleurs.fr/),
> Cloudflare R2, AWS S3, or an existing MinIO deployment.

## CLI reference

```
skywalker new <name>          clone + scaffold a new application
skywalker make migration <n>  create up/down migration files
skywalker make auth           create auth migrations, models, and middleware
skywalker make handler <n>    create a stub handler
skywalker make model <n>      create a model in the data directory
skywalker make session        create a sessions table migration
skywalker make mail <n>       create starter mail templates
skywalker migrate             run all up migrations
skywalker migrate down        reverse the most recent migration
skywalker migrate reset       run all down, then all up migrations
skywalker version | help
```

## Development

```bash
go build ./...
go vet ./...
go test ./...     # tests in cache, render, session, mailer, filesystems
```

Tests run with no external services; the mailer's SMTP tests spin up a MailHog container via
`dockertest` when Docker is available and skip cleanly otherwise. CI (build + vet + test) runs on
every push and pull request via GitHub Actions.

## Repositories

- **[Skywalker](https://github.com/stefanlester/Skywalker)** — this framework (`github.com/stefanlester/skywalker`).
- **[skywalker_engine](https://github.com/stefanlester/skywalker_engine)** — the demo/skeleton app that `skywalker new` clones.

## Status

Actively developed. The filesystem layer (all four backends), module-mode build, and CI are in place.
Live end-to-end coverage for the S3/SFTP/WebDAV backends and a project `LICENSE` are still to come.
