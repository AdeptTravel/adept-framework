# Adept Framework – System Architecture

## Overview

The Adept Framework is a multi-tenant Go web platform designed to serve many independent **sites** from a single binary.  Each site has its own database schema, static assets, and configuration, while the runtime maintains a lightweight **tenant** object that is created on demand and evicted after a period of inactivity.  The framework runs inside FreeBSD jails, relies on a Cockroach (or MariaDB/MySQL) control-plane database, and exposes metrics through Prometheus.

---

## Key Terminology

* **Site** – A persistent definition stored in the `site` table (host name, DSN, theme, status).  Site files live under `sites/{host}/`.
* **Tenant** – An in-memory wrapper that combines a site’s metadata with an open `*sqlx.DB` pool and a `LastSeen` timestamp.  Tenants are cached, lazily loaded, and evicted.

---

## High-Level Flow

1. **Startup**

   1. Load environment variables from `/usr/local/etc/adept/global.env`, else `.env`.
   2. Connect to the global control-plane database using `GLOBAL_DB_DSN`.
   3. Construct a `tenant.Cache` with default Idle TTL (30 minutes) and Max Entries (100).
2. **Request Handling**

   1. Strip `:port` from the `Host` header to obtain `hostKey`.
   2. Call `cache.Get(hostKey)`.

      * **Cache hit** → return the existing tenant, update `LastSeen`.
      * **Cache miss** → singleflight loader queries `site` table, opens DSN, stores tenant, updates Prometheus counters.
   3. Pass the tenant to the router, templates, modules, and widgets.
3. **Eviction**

   * A goroutine scans every five minutes.
   * Idle tenants (`now - LastSeen > IdleTTL`) are closed and removed.
   * If total entries > MaxEntries, the oldest tenants are evicted (simple LRU).

---

## Component Diagram

```text
┌─────────────┐          ┌──────────┐         ┌──────────────┐
│  HTTP Mux   │──host──▶│ tenantPkg │──miss──▶│  sitePkg SQL │
└─────────────┘          │  cache   │         │   (global)   │
       ▲                 └──────────┘         └──────────────┘
       │hit                                    ▲
       │                                       │
       │                         opens DSN     │
       │                                       │
       │          ┌────────────────────────────┘
       ▼          ▼
 Static files  Tenant.DB
```

---

## Data Stores

* **Global DB** – Single cluster of Cockroach (or MariaDB) holding the `site` table and future global tables (modules, OAuth clients, etc.).
* **Tenant DB** – Each site’s schema or separate database, connected via DSN stored in the `site` row.
* **Static Assets** – Synced to each server via rsync or object storage.

---

## Environment Variables

| Variable           | Purpose                             | Default | Notes                       |
| ------------------ | ----------------------------------- | ------- | --------------------------- |
| `GLOBAL_DB_DSN`    | DSN for control-plane database      | –       | Required.                   |
| `TENANT_IDLE_TTL`  | Idle eviction timeout (e.g., `45m`) | `30m`   | Optional.                   |
| `TENANT_CACHE_MAX` | LRU capacity                        | `100`   | `0` disables size eviction. |

---

## Metrics

* `active_tenants` – Gauge of currently cached tenants.
* `tenant_load_total` – Counter of successful tenant loads.
* `tenant_evict_total` – Counter of evicted tenants.
* `tenant_load_errors_total` – Counter of load failures.

---

## Future Extensions

* Secret manager abstraction (Vault OSS) behind `internal/secret`.
* Asset pipeline with fingerprinting and SRI hashes.
* Background job runners for itinerary updates, AI tasks, and webhooks.
* Admin CLI (`adeptctl`) for warm-loading or flushing tenants.
