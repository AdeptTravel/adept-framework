# Adept · Roadmap

> **Generated:** 2025‑06‑06 — reflects post‑MVP‑0.4 progress and new API / AI work.

---

## Vision

Adept is a FreeBSD‑friendly, Go‑native multi‑tenant platform that delivers
websites, APIs, and AI‑powered features with minimal ops friction.  Each
site lives in a single directory tree for drop‑in deploys, while
Components (feature packages) and Widgets (view fragments) allow rapid,
AI‑assisted scaffolding.

---

## Current Milestone – *MVP‑0.5* (Target September 2025)

| Pri    | Task                                                                      | Owner | Notes                                       |
| ------ | ------------------------------------------------------------------------- | ----- | ------------------------------------------- |
| **P0** | **API core helpers** (`client.go`, retry, rate, cache) with OpenAI client | GPT   | Enables AI layer and future providers       |
| **P0** | **AI layer** Chat + Embed helpers (OpenAI provider)                       | GPT   | Provider router, tenant credentials         |
| **P0** | **Tenant credential loader** (Vault or AES‑enc table fallback)            | BJY   | Inject creds into `Tenant.API` map          |
| **P1** | Per‑tenant rate‑limit middleware (global + scoped buckets)                | GPT   | Token bucket using `golang.org/x/time/rate` |
| **P1** | Security engine hard‑enforcement (block modes)                            | GPT   | Shadow mode already shipping                |
| **P2** | CLI `adept ai test-prompt --site` for ops sanity                          | GPT   | Uses AI layer                               |
| **P2** | RouterPath helper and Content Component refactor                          | GPT   | String‑safe URL builder                     |
| **P3** | Makefile targets (vet, lint, test, run) + GitHub Actions CI               | GPT   | Lint via `golangci-lint`                    |

*Definition of done:* Server boots with API & AI layers; `curl -H "Host: example.com" /api/chat` returns OpenAI response using tenant key; Prometheus shows rate‑limit metrics.

---

## Next Milestones

### *MVP‑0.6* (Q4‑2025)

* Forms subsystem (YAML loader, CSRF, hCaptcha/reCAPTCHA v3, email/store/pdf/webhook actions).
* Image variant generator (Libvips, watermark, smart‑contrast).
* Second API provider (SendGrid or Anthropic) to exercise abstraction.
* Observability polish: ClickHouse consumer for metrics, Tempo dashboards.

### *MVP‑0.7* (Q1‑2026)

* Feature‑flag service (sync via API layer).
* Admin UI seed (HTMX or Vue) for tenant CRUD and form builder.
* Background job runner (CRON expressions) for pre‑render & email.

---

## Delivered (MVP‑0.4)

* Module → Component rename and import rewrite.
* Merged `internal/site` into `internal/tenant`; added host preload list.
* Config loader (YAML + env); `koanf` tags + struct validation.
* Zap logger with daily rotation; structured spans throughout app.
* Request‑info pipeline (UA, Geo, IP) with DEBUG logging.
* Tenant LRU cache instrumented, scan error on zero‑dates fixed.

---

## Blocked / Research

| Topic                              | Status                               |
| ---------------------------------- | ------------------------------------ |
| Vault secret renewal loop          | PoC pending in staging.              |
| Queue engine (JetStream vs Kafka)  | Spike running — measure FreeBSD pkg. |
| Tracing exporter (Tempo vs Jaeger) | Evaluate OTLP gRPC latency.          |

---

## Reference Docs

* Comment style – `guidelines/comment-style.md`
* Architecture – `ARCHITECTURE.md`
* Component & Widget scaffold prompts – `prompts/`

---

*Roadmap is reviewed bi‑weekly after stand‑up; update priorities as needed.*
