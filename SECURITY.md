# Security Policy

We are committed to protecting the data and infrastructure of every tenant that relies on the framework.  We welcome—and greatly appreciate—coordinated vulnerability disclosures following industry‑standard best practices.

---

## Supported Versions

| Version branch | Status      | Supported until        |
| -------------- | ----------- | ---------------------- |
| **main**       | Development | Always supported       |
| **1.0.x**      | Stable      | 12 months after 1.0 GA |
| **0.9.x**      | Maintenance | 6 months after 1.0 GA  |

All earlier branches receive security fixes **only** if the issue is both
critical (CVSS ≥ 9.0) and the patch is low‑risk.  Otherwise, upgrading to a
supported branch is required.

---

## Reporting a Vulnerability

1. **Email `security@yaniz.io`.**  Include:

   * A clear description of the issue and its impact.
   * Steps to reproduce (proof‑of‑concept, demo site, or code snippet).
   * Affected Adept version(s) and any relevant environment details.
2. Do **not** open a public GitHub issue or discuss the vulnerability in
   public forums until we publish an advisory.
3. We will acknowledge your report within **72 hours** and keep you
   informed of the remediation timeline.

> **Encryption:** We do not yet publish a public PGP key.  If you require
> encrypted communication, note this in your initial email and we will
> arrange a secure channel.

---

## Disclosure Timeline

| Phase                  | Target window             |
| ---------------------- | ------------------------- |
| Acknowledge receipt    | 72 hours                  |
| Initial assessment     | 7 days                    |
| Fix in private repo    | ≤ 30 days\*               |
| Coordinated disclosure | Same day as patch release |

\* Complex issues may take longer.  We will provide weekly status updates
and coordinate an extended timeline when necessary.

---

## Public Advisories

* Security releases are tagged on GitHub and published on the project
  website.
* CVEs are requested for all critical or high‑severity issues once a fix
  is available.
* Release notes include upgrade instructions and mitigation steps.

---

Thank you for helping keeping this project and its tenants safe.
