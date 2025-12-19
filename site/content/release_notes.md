# Release Notes (summary)

This page presents concise, user-friendly summaries of recent releases pulled from the project's CHANGELOG. Use these on the marketing site or documentation landing pages.

---

## Unreleased

- Ongoing improvements and fixes. Check the full `docs/CHANGELOG.md` for details.

---

## v0.14.0-beta — Subscription Billing System (2024-12-16)

What’s new

- Dual-level subscription support (individual and organization).
- Admin APIs and tooling to create, mark-paid, cancel, and audit subscriptions.
- Read-only enforcement when subscriptions expire (write operations return HTTP 402).
- Database migrations and multi-database snapshots included for SQLite, PostgreSQL, and MariaDB.

Why it matters

- Enables gyms and teams to centrally manage access while preserving individual subscriptions.
- Admins can manage billing manually until automated billing is added.

---

## v0.12.2-beta — PWA Offline Reliability (2025-11-28)

What’s new

- Robust offline recording and sync for workouts, including PUT support for updates.
- Improved offline detection in API layer and a local sync queue.
- New, user-controlled update prompt (no sudden reloads) via `UpdatePrompt.vue`.
- "Saved Offline" notifications to keep users informed.

Why it matters

- Reliable offline recording ensures workouts aren’t lost when connectivity is poor.
- Users remain in control when the app updates, reducing accidental data loss.

---

## v0.12.1-beta — DB Compatibility Fixes (2025-11-28)

What’s new

- Database-agnostic timestamp helpers and fixes for MySQL/MariaDB compatibility.
- Minor documentation updates for Docker-hosted databases and troubleshooting.

Why it matters

- Expands production compatibility and makes migrations and snapshots more reliable across different SQL engines.

---

## v0.12.0-beta — Mobile Layout & PWA polish (2025-11-26)

What’s new

- Systematic mobile overflow fixes across many views.
- `.mobile-view-wrapper` layout pattern and safe-area handling for iOS PWAs.
- Docker image metadata added for better registry visibility.

Why it matters

- Enhances mobile experience and PWA stability, making the app feel native on phones and tablets.

---

## Full change history

For complete details, file-level notes, and migration instructions, see `docs/CHANGELOG.md` and `docs/TODO.md`.
