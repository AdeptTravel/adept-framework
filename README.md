# Adept Framework

Adept is a Go‑native, FreeBSD‑friendly **multi‑tenant** web and API platform.  It runs many independent sites from a single binary, lazy‑loads each tenant on the first request, and evicts idle ones to keep memory usage low.  A unified *API* layer and provider‑agnostic *AI* helpers let Components call external services (OpenAI, SendGrid, …) with one line of code.

---

## Features

* **Lazy‑loaded tenants** — cold loads on first hit, idle‑TTL + LRU eviction.
* **Per‑tenant DB pools** capped at five connections, DSN pulled from Vault or table.
* **Structured logging, metrics, tracing** — Zap JSON logs, Prometheus `/metrics`, OTEL spans.
* **Security engine** — host/IP/Geo/UA allow‑deny with *shadow mode* and Prom metrics.
* **Unified API layer** — auth, retry, rate‑limit, TTL cache; first client: **OpenAI**.
* **AI helpers** — Chat / Embed wrappers that pick the right provider at runtime.
* **Theme & Asset manager** — bundles CSS/JS, serves pre‑compressed brotli.
* **One‑directory deployment** — drop `/inet` tree onto any jail/server and start.

---

## Quick Start (development)

```bash
# Clone and build
$ git clone https://github.com/yanizio/adept.git
$ cd adept
$ go mod tidy
$ go build ./cmd/web

# Minimal config (.env or conf/global.yaml)
$ cat > .env <<EOF
GLOBAL_DB_DSN=adept:pass@tcp(127.0.0.1:3306)/adept_global?parseTime=true&loc=Local
ADEPT_ROOT=$(pwd)
EOF

# Run with an example site row already present
$ go run ./cmd/web
```

Visit [http://127.0.0.1:8080/](http://127.0.0.1:8080/) for the placeholder page and
[http://127.0.0.1:8080/metrics](http://127.0.0.1:8080/metrics) for live stats.

---

## Directory Layout (partial)

```text
cmd/
  web/                    # HTTP entry point
internal/
  api/                    # generic client helpers + service dirs (openai/…)
  ai/                     # provider‑agnostic Chat/Embed helpers
  config/                 # env + YAML loader, validation
  dbcore/                 # sqlx helpers, migrations
  tenant/                 # meta models + lazy‑load LRU cache
  security/               # IP/UA/Geo rules, shadow‑mode metrics
  observability/          # zap, prometheus, OTEL, sentry
components/               # first‑class business features
themes/                   # templates + assets
conf/                     # global.yaml, security.yaml, etc.
logs/                     # daily JSON logs
```

---

## Configuration & Environment

Most tunables live in `conf/global.yaml`; environment variables can
override any key by prefixing with `ADEPT_` and replacing dots with
double underscores (`__`).

| Variable / YAML key       | Example value                                               | Purpose                           |
| ------------------------- | ----------------------------------------------------------- | --------------------------------- |
| `database.global_dsn`     | `user:pass@tcp(127.0.0.1:3306)/adept_global?parseTime=true` | Control‑plane MySQL/Cockroach DSN |
| `http.listen_addr`        | `127.0.0.1:8080`                                            | Bind address                      |
| `http.force_https`        | `true`                                                      | 308 redirect for non‑HTTPS        |
| `adept_root` *(env only)* | `/inet`                                                     | One‑directory deployment root     |

---

## Building on FreeBSD

```bash
pkg install go git
cd /usr/local/adept
make build          # invokes go vet, go test, go build
service adept_web start
```

Systemd or rc.d script sets `WorkingDirectory=/inet` and sources
`/usr/local/etc/adept/global.env` before launch.

---

## Contributing

1. Fork the repo, create a feature branch.
2. Run `make vet test`.  Ensure `go vet` and `golangci-lint` pass.
3. Submit a pull request with a clear description.

We follow the comment style in `guidelines/comment-style.md`
---

## License

BSD 3‑Clause (see `LICENSE`).
