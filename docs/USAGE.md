# Using Skywalker

A practical guide to building applications with the Skywalker framework. For installation and a
feature overview, see the [README](../README.md).

Everything below is driven by two things: the `Skywalker` application type (created by
`New()`) and your `.env` file.

## 1. Create and run an application

```bash
./skywalker new myapp
cd myapp
go run .
```

`skywalker new` clones the skeleton, writes a fresh `.env` (with a random 32-char `KEY`), rewrites
`go.mod` to your app name, and runs `go mod tidy`. On first boot, `New()` creates any missing
standard folders: `handlers`, `migrations`, `views`, `data`, `public`, `tmp`, `logs`, `middleware`.

The app boots with **no external services** when `.env` has `SESSION_TYPE=cookie` and empty
`DATABASE_TYPE`/`CACHE`.

## 2. Application structure

Your app wraps the framework in an `application` struct (see the skeleton's `main.go`):

```go
type application struct {
    App        *skywalker.Skywalker
    Handlers   *handlers.Handlers
    Models     data.Models
    Middleware *middleware.Middleware
}
```

`init-skywalker.go` calls `skywalker.New(rootPath)`, which reads `.env` and wires every subsystem:
router, renderer, sessions, DB, cache, mailer, scheduler, and filesystems.

## 3. Routing

Routes live in `routes.go`, using thin helpers over [chi](https://github.com/go-chi/chi):

```go
func (a *application) routes() *chi.Mux {
    a.get("/", a.Handlers.Home)
    a.get("/files/upload", a.Handlers.UploadToFS)
    a.post("/files/upload", a.Handlers.PostUploadToFS)

    // static assets
    fileServer := http.FileServer(http.Dir("./public"))
    a.App.Routes.Handle("/public/*", http.StripPrefix("/public", fileServer))

    return a.App.Routes
}
```

`a.use(...)` adds middleware. The framework pre-installs chi's `RequestID`, `RealIP`, `Recoverer`
(plus `Logger` when `DEBUG=true`), session loading, and **CSRF protection (`nosurf`)**.

> Because of nosurf, every `POST` must carry the CSRF token — include
> `<input type="hidden" name="csrf_token" value="{{.CSRFToken}}">` in forms (the skeleton's
> layout exposes it as a meta tag too). A token-less POST gets `400 Bad Request`.

## 4. Handlers and rendering

Handlers hang off a `Handlers` struct that carries the app:

```go
func (h *Handlers) Home(w http.ResponseWriter, r *http.Request) {
    err := h.render(w, r, "home", nil, nil)   // wraps h.App.Render.Page
    if err != nil {
        h.App.ErrorLog.Println("error rendering:", err)
    }
}
```

`RENDERER` selects the engine: `jet` (default) renders `views/<name>.jet`; `go` renders
`views/<name>.page.tmpl`. Pass Jet variables via `jet.VarMap`:

```go
vars := make(jet.VarMap)
vars.Set("list", files)
h.render(w, r, "list-fs", vars, nil)
```

Templates always receive `TemplateData` with `.CSRFToken`, `.Flash`, `.Error`,
`.IsAuthenticated`, `.ServerName`, `.Secure`, and `.Port`. Jet templates support layouts:
`{{extends "./layouts/base.jet"}}` with `{{block ...}}` sections.

**Performance note:** with `DEBUG=true`, templates re-parse on every request (edit-and-refresh).
With `DEBUG=false`, both Jet and Go templates are parsed once and cached — use this in production.

## 5. Sessions and flash messages

`App.Session` is an [scs](https://github.com/alexedwards/scs) manager; the store follows
`SESSION_TYPE` (`cookie`, `redis`, `mysql`, `postgres`):

```go
h.App.Session.Put(r.Context(), "userID", id)
id := h.App.Session.GetInt(r.Context(), "userID")
h.App.Session.Remove(r.Context(), "userID")
_ = h.App.Session.RenewToken(r.Context())   // do this on login
_ = h.App.Session.Destroy(r.Context())      // and this on logout
```

Flash messages are just session keys the renderer pops for you: put `"flash"` or `"error"` and the
next rendered page exposes them as `.Flash` / `.Error`:

```go
h.App.Session.Put(r.Context(), "flash", "File uploaded!")
http.Redirect(w, r, "/files/upload", http.StatusSeeOther)
```

Setting `IsAuthenticated` is driven by the presence of the `userID` session key.

## 6. Database, models, and migrations

Set `DATABASE_TYPE` (`postgres` or `mysql`) plus `DATABASE_HOST/PORT/USER/PASS/NAME/SSL_MODE`.
`New()` opens the pool (`App.DB.Pool`); models use [upper/db](https://upper.io/) via `data.New`.

Scaffold and run migrations with the CLI (run from the app root, reads `.env`):

```bash
./skywalker make migration create_widgets   # writes migrations/*.up.sql / *.down.sql
./skywalker migrate                          # apply all pending
./skywalker migrate down                     # roll back the most recent
./skywalker migrate reset                    # all down, then all up
./skywalker make model widget                # data/widget.go with CRUD via upper/db
./skywalker make auth                        # auth tables migration + user/token models + middleware
./skywalker make session                     # sessions table migration (for DB session stores)
```

## 7. Caching

`App.Cache` implements one interface regardless of backend (`CACHE=redis` or `CACHE=badger`;
Badger is embedded — no server needed):

```go
err := app.Cache.Set("answer", 42)          // forever
err  = app.Cache.Set("answer", 42, 300)     // TTL in seconds
v, err := app.Cache.Get("answer")
ok, err := app.Cache.Has("answer")
err  = app.Cache.Forget("answer")
err  = app.Cache.EmptyByMatch("user-")      // delete by prefix
err  = app.Cache.Empty()
```

Values are gob-encoded; register custom types with `gob.Register` if you cache structs.

## 8. Sending mail

`App.Mail` is a queued mailer — push a `Message` onto the job channel and a background listener
sends it (API service if `MAILER_API`/`MAILER_KEY`/`MAILER_URL` are set — mailgun, sendgrid, or
sparkpost — otherwise SMTP):

```go
msg := mailer.Message{
    To:       "user@example.com",
    Subject:  "Welcome!",
    Template: "welcome",              // mail/welcome.html.tmpl + welcome.plain.tmpl
    Data:     someData,
}
app.Mail.Jobs <- msg
res := <-app.Mail.Results
if !res.Success {
    app.ErrorLog.Println(res.Error)
}
```

Create template pairs with `./skywalker make mail welcome`. HTML mail is run through premailer
(inlines CSS) before sending.

## 9. Remote filesystems

Four backends behind one interface — construction is gated on env (`MINIO_SECRET`, `S3_SECRET`,
`SFTP_HOST`, `WEBDAV_HOST`); configured backends appear in `App.FileSystems` keyed `"MINIO"`,
`"S3"`, `"SFTP"`, `"WEBDAV"`:

```go
fs, ok := app.FileSystems[fsType].(filesystems.FS)
if !ok { /* backend not configured */ }

err := fs.Put("./tmp/report.pdf", "reports")   // upload local file into folder
items, err := fs.List("reports/")              // Listing{Key, Size(MB), LastModified, IsDir}
err  = fs.Get("./downloads", "reports/report.pdf")
ok   = fs.Delete([]string{"reports/report.pdf"})
```

`Listing.Size` is in megabytes (`filesystems.SizeToMB`). See the engine's upload/list/delete
handlers for a complete worked example.

The `MINIO` backend speaks the standard S3 protocol via the actively maintained minio-go client,
so it works with **any S3-compatible server**: SeaweedFS (the skeleton's bundled dev default),
Garage, Cloudflare R2, AWS S3, or an existing MinIO deployment. (The MinIO server community
edition itself is no longer maintained upstream — hence SeaweedFS as the bundled default.)

## 10. Encryption, signed URLs, JSON APIs

```go
// authenticated AES-GCM encryption using the 32-char KEY from .env
// (tamper-proof: Decrypt errors on any modified or wrong-key ciphertext)
enc := skywalker.Encryption{Key: []byte(app.EncryptionKey)}
c, _ := enc.Encrypt("secret")
p, _ := enc.Decrypt(c)

// tamper-proof, expiring links (password resets etc.)
signer := urlsigner.Signer{Secret: []byte(app.EncryptionKey)}
signed := signer.GenerateTokenFromString("https://example.com/reset?email=a@b.c")
valid  := signer.VerifyToken(signed)
expired := signer.Expired(signed, 60)   // minutes

// JSON/XML APIs (compact output)
var input payload
err := app.ReadJSON(w, r, &input)               // 1 MB limit, rejects trailing JSON
err  = app.WriteJSON(w, http.StatusOK, data)
err  = app.WriteXML(w, http.StatusOK, data)
err  = app.DownloadFile(w, r, "./tmp", "report.pdf")
app.Error404(w, r); app.Error500(w, r)          // status helpers
```

## 11. Background jobs

`App.Scheduler` is a [robfig/cron](https://github.com/robfig/cron) instance:

```go
id, _ := app.Scheduler.AddFunc("@every 1h", func() { /* work */ })
app.Scheduler.Start()
```

## 12. Production checklist

- `DEBUG=false` — enables template caching, disables the request logger, and switches all logs
  (`App.Log`, plus the `InfoLog`/`ErrorLog` bridges) to structured JSON on stdout.
- `SECURE=true` + `COOKIE_SECURE=true` behind TLS.
- `KEY` must be exactly 32 characters; treat `.env` as a secret (never commit it).
- Use a persistent session store (`redis`/`postgres`/`mysql`) if you run more than one instance.
- SFTP backend skips host-key verification (demo-grade) — front it with known hosts before
  production use.
