# Skywalker — agent context

**Skywalker** is a batteries-included Go web framework (module `github.com/stefanlester/skywalker`),
an open re-implementation in the lineage of the "Celeritas" course framework. A companion repo,
**`skywalker_engine/`** (module `myapp`), is a demo app that consumes the framework.

## Module map (framework)
- `skywalker.go` — core `Skywalker` type + `New()` bootstrap (loads `.env`, wires every subsystem).
- `render/` — Jet v6 + `html/template` rendering. `session/` — scs sessions. `cache/` — Redis + Badger behind a `Cache` interface.
- `mailer/` — queued SMTP + API mail. `urlsigner/` — signed URLs. `data/`, `driver.go`, `migrations.go` — DB (pgx/mysql) + golang-migrate.
- `filesystems/` — `FS` interface (`Put/Get/List/Delete`) with `minio`, `s3`, `sftp`, `webdav` backends.
- `cmd/cli/` — scaffolder: `new`, `migrate`, `make {migration,auth,handler,model,session,mail}`; templates in `cmd/cli/templates/`.
- `dist/myapp/` — **generated** sample app (vendored). Treat as build output, not source.

## Conventions
- Small files, doc comments on exported symbols, `os.Getenv`-driven config, interface-first design.
- Reusable capability → framework; demo wiring → `skywalker_engine`. The framework must never import `myapp`.
- When changing a framework signature, update every consumer: `cmd/cli/templates/` stubs **and** the engine.

## Verify changes
```
go build ./...        # affected module
go vet ./...
go test ./...         # tests live in cache, render, session, mailer
```
Host is Windows/PowerShell (Bash tool also available). Keep code cross-platform.

## Do not touch / do not search
`dist/myapp/vendor/**`, any `vendor/**`, and `skywalker_engine/db-data/**` are noise — exclude from searches.
Never commit `.env`, private keys, or `db-data/`. `skywalker_engine` currently tracks secrets (`.env`, `db-data/home/id_*`) — flag, don't propagate.

## Specialized agent
Use the **`apex`** subagent (`.claude/agents/apex.md`) for non-trivial design/implementation/review across both repos.
