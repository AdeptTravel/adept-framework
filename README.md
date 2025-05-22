# Adept Framework

The Adept Framework is a Go‑based, multi‑tenant web platform engineered for travel‑industry workloads.  It serves many independent sites from a single binary, lazy‑loads each site’s resources on the first request, and evicts idle tenants to keep resource usage low.  The project is developed on FreeBSD jails, but it runs anywhere Go 1.24+ is available.

---

## Features

* **Lazy‑loaded tenants** – Sites are loaded on first hit, with idle‑TTL and LRU eviction.
* **Per‑tenant database pools** capped at five open connections.
* **Prometheus metrics** on `/metrics`.
* **Pluggable AI and third‑party API adapters** (scaffold stage).
* **Future‑ready for Vault secrets** and Cockroach multi‑region clusters.

---

## Quick Start

```bash
# Clone and build
git clone https://github.com/AdeptTravel/adept-framework.git
cd adept-framework
go mod tidy
go build ./cmd/web

# Set environment (development)
cat > .env <<EOF
GLOBAL_DB_DSN=adept:pass@tcp(127.0.0.1:3306)/adept_global
TENANT_IDLE_TTL=30m
TENANT_CACHE_MAX=100
EOF

# Run with an example site row already present
go run ./cmd/web
```

Visit `http://localhost:8080/` for the placeholder page, and `http://localhost:8080/metrics` for live stats.

---

## Directory Layout

```
cmd/                  # entry points (web server, future CLI)
  web/                # main server binary
internal/
  tenant/             # lazy‑loading cache
  site/               # SQL helpers for site table
  database/           # sqlx helpers
  metrics/            # Prometheus collectors
config/               # project‑wide constants (default TTLs)
sites/                # static files per host (synced via rsync)
sql/                  # schema and migrations
```

---

## Environment Variables

| Variable           | Description                        | Example                                |
| ------------------ | ---------------------------------- | -------------------------------------- |
| `GLOBAL_DB_DSN`    | DSN for the control‑plane database | `user:pass@tcp(10.0.0.5:26257)/global` |
| `TENANT_IDLE_TTL`  | Idle eviction timeout              | `45m`                                  |
| `TENANT_CACHE_MAX` | Max cached tenants (LRU)           | `200`                                  |

---

## Building on FreeBSD

```bash
sudo pkg install go git
cd /opt/adept-framework
make build
service adept_web start
```

Environment variables are read first from `/usr/local/etc/adept‑framework/global.env`, falling back to `.env` in the working directory.

---

## Contributing

1. Fork the repo, create a feature branch, and run `make vet test`.
2. Submit a pull request with a clear description.

---

## License

BSD 3‑Clause (see `LICENSE`).
