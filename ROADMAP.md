# Adept · Roadmap

> **Generated:** 2025‑06‑13 — merged messaging subsystem milestones, directory rename (database/), and recent planning decisions.  Review bi‑weekly, and adjust priorities as needed.

---

## Vision

Adept is a Go‑native, multi‑tenant platform that delivers websites, APIs, AI‑powered features, **and multi‑channel messaging** with minimal ops friction.  Each site lives in a single directory tree for drop‑in deploys, while Components (feature packages) and Widgets (view fragments) allow rapid, AI‑assisted scaffolding.

---

## Version Overview

This table tracks delivery toward **MVP 1.0**.  Each task carries a status badge.

✅ done ⏳ in progress 🔜 next 🟦 planned 📄 design only

| Version | Target date | Theme                                | Key tasks                                                                                                                                                                          |
| ------- | ----------- | ------------------------------------ | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| **0.5** | 2025‑09     | Foundations                          | ✅ Vault client + renew, ✅ structured logging, ✅ Prometheus metrics, ⏳ AI provider abstraction, ⏳ OpenAI Chat + Embed, ⏳ rate‑limit middleware                                      |
| **0.6** | 2025‑11     | Secure Input **+ Messaging Phase 1** | 🔜 Forms subsystem (YAML, validation, CSRF), 🔜 UA normalization, 🔜 `/ready` endpoint, 🔜 **Messaging queue + SendGrid/Twilio adapters**, 🔜 per‑tenant opt‑out table             |
| **0.7** | 2026‑01     | Identity **+ Messaging Phase 2**     | 🟦 Auth GA (local, OAuth 2, OIDC, passkeys), 🟦 profile selection UI, 🟦 magic link, 🟦 **Template editor & in‑app/push notifications (FCM/APNS)**, 🟦 provider webhook processing |
| **0.8** | 2026‑03     | Ops and UX                           | 🟦 Template cache shared across tenants, 🟦 graceful shutdown with drain, 🟦 OpenTelemetry traces end‑to‑end, 🟦 tenant CLI backup + restore                                       |
| **0.9** | 2026‑06     | Ecosystem                            | 🟦 `adept component scaffold`, 🟦 image variant service, 🟦 second AI provider fallback, 🟦 plug‑in registry site                                                                  |
| **1.0** | 2026‑08     | Production GA                        | 🟦 Admin UI, 🟦 job runner with retries, 🟦 load test harness, 🟦 security audit pass, 🟦 docs freeze                                                                              |

---

## Current Milestone – *MVP‑0.5* (Target September 2025)

| Pri    | Task                                                                      | Owner | Notes                                               |
| ------ | ------------------------------------------------------------------------- | ----- | --------------------------------------------------- |
| **P0** | **API core helpers** (`client.go`, retry, rate, cache) with OpenAI client | GPT   | Enables AI layer and future providers.              |
| **P0** | **AI layer** Chat + Embed helpers (OpenAI provider)                       | GPT   | Provider router, tenant credentials.                |
| **P0** | **Tenant credential loader** (Vault path or AES‑enc fallback)             | BJY   | Inject creds into `Tenant.API` map — global done ✅. |
| **P1** | Per‑tenant rate‑limit middleware (global + scoped buckets)                | GPT   | Token bucket via `x/time/rate`.                     |
| **P1** | Security engine hard‑enforcement (block modes)                            | GPT   | Shadow mode already shipping.                       |
| **P2** | Messaging design doc & queue schema draft                                 | GPT   | Blocks 0.6 work.                                    |
| **P2** | CLI `adept ai test‑prompt --site` for ops sanity                          | GPT   | Uses AI layer.                                      |
| **P2** | RouterPath helper + Content Component refactor                            | GPT   | String‑safe URL builder.                            |
| **P3** | Makefile targets (vet, lint, test, run) + GitHub Actions CI               | GPT   | Lint via `golangci‑lint`.                           |

*Definition of done:* Server boots with API & AI layers.  `curl -H "Host: example.com" /api/chat` returns OpenAI response using tenant key.  Prometheus shows rate‑limit metrics.  Messaging design doc merged.

---

## Next Milestones

### *MVP‑0.6* (Q4‑2025)

* **Messaging Phase 1** — `internal/message` package: queue table, worker pool, SendGrid (email) + Twilio (SMS) adapters, opt‑out table, Prometheus metrics.
* Forms subsystem (YAML loader, CSRF, hCaptcha / reCAPTCHA v3, email / store / pdf / webhook actions) integrated with Messaging.
* UA normalization middleware; `/ready` health endpoint.
* Image variant generator (Libvips, watermark, smart‑contrast).
* Observability polish: ClickHouse consumer + Tempo dashboards.

### *MVP‑0.7* (Q1‑2026)

* **Messaging Phase 2** — template editor, i18n, push notifications (FCM/APNS), provider webhooks for delivery logs.
* Auth Component GA: local creds, OAuth 2, OIDC, WebAuthn passkeys.
* Account‑profile selection UI + optional passwordless magic link flow.
* Feature‑flag service (sync via API layer).
* Admin UI seed (HTMX or Vue) for tenant CRUD + form builder.
* Background job runner (CRON expressions) leveraging Messaging for digest emails.

### *MVP‑0.8* (Q1‑2026)

* Template cache shared across tenants for memory savings.
* Graceful shutdown with drain and hot reload.
* End‑to‑end OpenTelemetry tracing.
* Tenant CLI backup and restore.

### *MVP‑0.9* (Q2‑2026)

* `adept component scaffold` generator.
* Image variant service GA.
* Second AI provider fallback routing.
* Plug‑in registry site.

### *MVP‑1.0* (Q3‑2026)

* Admin UI GA.
* Job runner with retries.
* Load test harness.
* Security audit pass.
* Documentation freeze.

---

## Delivered (MVP‑0.4 → 0.5)

* Vault client (`internal/vault`) with background token renewal.
* Config loader resolves `vault:` URIs; global DB password now pulled from Vault.
* Lazy DSN provider; per‑tenant pools capped at five connections.
* Module → Component rename and import rewrite.
* Merged `internal/site` into `internal/tenant`; added host preload list.
* Config loader (YAML + env); `koanf` tags + struct validation.
* Zap logger with daily rotation; structured spans throughout app.
* Request‑info pipeline (UA, Geo, IP) with DEBUG logging.
* Tenant LRU cache instrumented; scan error on zero‑dates fixed.

---

## Blocked / Research

| Topic                              | Status                                  |
| ---------------------------------- | --------------------------------------- |
| Queue engine (JetStream vs Kafka)  | Spike running — measure package impact. |
| Tracing exporter (Tempo vs Jaeger) | Evaluate OTLP gRPC latency.             |

*(Vault renewal loop delivered; removed from Blocked.)*

---

## Reference Docs

* Comment style — `guidelines/comment-style.md`
* Architecture — `ARCHITECTURE.md`
* Component & Widget scaffold prompts — `prompts/`

---

Roadmap is reviewed bi‑weekly after stand‑up.  Update priorities as needed.
