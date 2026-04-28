# Protected Users — Runbook

> **Status:** Stub. Full runbook arrives in Task 24 of the v1.3.0 plan.
>
> Until then, see `docs/superpowers/specs/2026-04-28-admin-user-edit-design.md`
> sections §3 and §6 for the design and `docs/superpowers/plans/2026-04-28-admin-user-edit-v1.3.0-plan.md`
> for the implementation order.

## Quick reference

The protected-user system has four defense layers (L1 middleware, L2 service
guard, L3 database trigger, L4 audit log) plus a boot-time invariant. The
single source of truth is `pkg/security/protected_users.go`. The frontend
mirror at `web/src/utils/protectedUsers.js` is auto-generated from the Go
source via `make gen-protected-emails`.

To add or remove a protected user, see the operator runbook in this file
(populated by Task 24 of the implementation plan).
