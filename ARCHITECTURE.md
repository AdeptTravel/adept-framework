# Adept · System Architecture (MVP)

> **Status:** Revised 2025‑06‑13 — adds Messaging subsystem and queue stage; directory layout flags `(planned)`.

---

## 0 · Glossary

| Term          | Definition                                                                                              |
| ------------- | ------------------------------------------------------------------------------------------------------- |
| **Tenant**    | A single website or API instance, loaded on‑demand, isolated by DB schema and in‑memory cache.          |
| **Component** | Primary feature package that owns routes, business logic, and optional Widgets.                         |
| **Widget**    | View fragment nested *inside* its parent Component.                                                     |
| **Theme**     | Site‑wide skin: base templates plus static assets.                                                      |
| **Variant**   | Resized / cached image or pre‑compressed static asset.                                                  |
| **API layer** | Re‑usable clients for third‑party services (OpenAI, SendGrid, …) providing auth, retry, and rate‑limit. |
| **AI layer**  | Provider‑agnostic helpers (Chat, Embed, Classify) built atop the API layer.                             |
| **Messaging** | Queue‑backed email, SMS, and push notification subsystem with pluggable providers.                      |

---

## 1 · Directory Layout

```
adept/
├── cmd/
│   └── web/                     # main(), flags, HTTP listener
├── internal/
│   ├── api/                     # generic helpers + service clients
│   ├── ai/                      # Chat / Embed helpers, provider router
│   ├── message/ (planned)       # sender/, template/, queue/, worker/, provider/
│   ├── config/                  # env, YAML, Vault integration
│   ├── vault/                   # singleton client + KV helpers
│   ├── database/                  # sqlx helpers, migrations
│   ├── tenant/                  # meta models + lazy‑load LRU cache
│   ├── security/                # IP / UA / Geo rules, shadow‑mode
│   ├── observability/           # zap, prometheus, OTEL, sentry
│   ├── routerpath/              # URL builder + reverse lookup
│   └── cache/                   # TinyLRU + singleflight thin wrapper
├── components/ …                # Components + Widgets
├── themes/ …                    # Theme templates + assets
├── sites/                       # per‑tenant overrides at <host>/
│   └── <host>/                  # e.g., my.adept.travel/
│       ├── assets/
│       │   └── images/
│       └── templates/           # override hierarchy
│           ├── theme/
│           ├── components/
│           └── widgets/
├── images/                      # originals/, variants/, tmp/
├── assets/                      # global static files + cache/
├── sql/                         # global & tenant schema
└── README.md
```

adept/
├── cmd/
│   └── web/                      # main(), flags, HTTP listener
├── internal/
│   ├── api/                      # generic helpers + service clients
│   ├── ai/                       # Chat / Embed helpers, provider router
│   ├── message/ (planned)        # sender/, template/, queue/, worker/, provider/
│   ├── config/                   # env, YAML, Vault integration
│   ├── vault/                    # singleton client + KV helpers
│   ├── database/                   # sqlx helpers, migrations
│   ├── tenant/                   # meta models + lazy‑load LRU cache
│   ├── security/                 # IP / UA / Geo rules, shadow‑mode
│   ├── observability/            # zap, prometheus, OTEL, sentry
│   ├── routerpath/               # URL builder + reverse lookup
│   └── cache/                    # TinyLRU + singleflight thin wrapper
├── components/ …                 # Components + Widgets
├── themes/ …                     # Theme templates + assets
├── images/                       # originals/, variants/, tmp/
├── assets/                       # global static files + cache/
├── sql/                          # global & tenant schema
└── README.md

````
*`(planned)` denotes a directory scaffolded in future milestones.*

---

## 2 · Boot Sequence

| #   | Stage                                             | Package            | Fatal? |
| --- | ------------------------------------------------- | ------------------ | ------ |
| 1   | Resolve **ADEPT_ROOT**                            | `config`           | yes    |
| 2   | Logger core (Zap) + Prometheus `/metrics`         | `observability`    | yes    |
| 3   | Load `.env`, YAML, Vault secrets, env overrides   | `config`           | yes    |
| 4   | DB connect via lazy DSN provider (Vault‑aware)    | `database`         | yes    |
| 5   | Load global + tenant API credentials              | `api`              | yes    |
| 6   | Security engine ruleset                           | `security`         | yes    |
| 7   | Tenant registry & cold‑load cache                 | `tenant/cache`     | no     |
| 8   | AI provider router init                           | `ai`               | no     |
| 9   | Analytics batch queue                             | `observability`    | yes    |
| 9a  | Messaging worker pool + queue listener (planned)  | `message/worker`   | no     |
| 10  | Theme template pre‑parse                          | `themes`           | no     |
| 11  | Signal traps, readiness toggle                    | `cmd/web`          | —      |
| 12  | HTTP listener (`/healthz`, `/ready`, `/metrics`)  | `chi`              | —      |

*`/ready` returns 200 after stage 9; `/healthz` after stage 2.*

---

## 3 · Request Pipeline

```text
Security.Secure
    ↓
Alias.Resolve
    ↓
RateLimit (global + tenant buckets)
    ↓
Tenant.Router
    ├─ Component (HTML / JSON / PDF)
    │     └─ optional Widget(s)
    ├─ API layer  →  external service
    │      └─ AI layer (Chat / Embed / Classify)
    └─ Message sender (Email / SMS / Push)  →  queue → provider
````

---

## 4 · Subsystems

### 4.1 Security Engine

* Host, IP/CIDR, Geo, UA checks with *shadow‑mode* metrics.
* Config file `conf/security.yaml`.
* Metrics: `adept_security_hits_total`, `adept_blocklist_hits_total`.

### 4.2 API Layer

* Package `internal/api`.
* Shared helpers: `client.go`, `retry.go`, `rate.go`, `cache.go`.
* Per‑service directories (first: **openai**).
* Per‑tenant credentials injected by the Tenant loader.

### 4.3 AI Layer

* Package `internal/ai`.
* Interfaces `LLM`, `Vision`, etc.
* Router picks provider based on tenant/global config.
* First tasks: `Chat`, `Embed`.  Default provider: **OpenAI**.

### 4.4 Messaging (planned)

* Package `internal/message`.
* Facade helpers: `Email`, `SMS`, `Push` attach to queue.
* Persistent `message_queue` table per tenant; retry with back‑off.
* Provider adapters: SendGrid, SES, Twilio, FCM (later).
* Metrics: `adept_msg_sent_total`, `adept_msg_failed_total`, latency histogram.

### 4.5 Forms

* YAML definition, CSRF token, pluggable CAPTCHA (hCaptcha, reCAPTCHA v3).
* Actions: `email`, `store`, `pdf`, `webhook`.

### 4.6 Image Variants

* Libvips resize pipeline, optional watermark with smart‑contrast.
* Variants cached under `images/variants/`.

### 4.7 Assets & Bundles

* Per‑page CSS/JS registration → minify + hashed bundle key.
* Pre‑compressed `.br` / `.gz`, `Cache‑Control: immutable`.

### 4.8 Observability

* **Logging:** Zap JSON, daily rotation via Lumberjack.
* **Tracing:** OTLP exporter (Tempo / Jaeger), 1 % sample rate.
* **Metrics:** Prometheus, ClickHouse batch via JetStream.

### 4.9 Vault Secrets

* Vault client in `internal/vault` (singleton, renewal‑aware).
* Config loader resolves `vault:<path>#<key>` for any field.
* Global DB credentials are pulled securely from Vault at boot.
* Supports future dynamic credentials and per‑tenant paths.

---

### 4.10 Site Overrides

* Directory structure `sites/<host>/` mirrors the theme hierarchy, letting a tenant override **theme**, **component**, or **widget** templates.
* Asset lookup order: **site → theme → default**.
* Overrides are discovered during tenant cold‑load and injected into the template engine’s lookup path.

## 5 · Deployment Model

* **One‑directory** runtime at `/inet` (`/inet/bin`, `/inet/conf`, `/inet/logs`, `/inet/themes`, `/inet/assets`, `/inet/images`, `/inet/sites`).
* `ADEPT_ROOT` overrides auto‑detection; other `ADEPT_*_DIR` vars allow relocation.
* systemd `WorkingDirectory=/inet`, readiness probe `/ready`.

---

## 6 · Outstanding TODOs

* ✅ Global DB credential pulled from Vault via AppRole.
* ⏳ Tenant credential loader (Vault path or AES‑enc fallback).
* ⏳ Messaging subsystem scaffold (queue, sender, SendGrid + Twilio) (**Q3‑2025**).
* ⏳ API core helpers + OpenAI client (**Q3‑2025**).
* ⏳ AI layer Chat / Embed with provider router (**Q3‑2025**).
* ⏳ Per‑tenant rate‑limit middleware.
* ⏳ Forms subsystem, Image variant generator, CLI utilities.
* ⏳ Observability polish: ClickHouse consumer, Tempo dashboards.
