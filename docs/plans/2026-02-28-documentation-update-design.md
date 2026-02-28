# Documentation Update Design

**Date:** 2026-02-28
**Version:** 1.2.0-beta
**Scope:** Content-only update of all existing user-facing documentation

---

## Problem Statement

ActaLog has grown from v0.12.2-beta to v1.2.0-beta with significant new features, but existing documentation has not kept pace. The public landing page shows v0.16.0-beta, the end-user help guide is frozen at v0.12.2-beta, and the roadmap describes a development state from over a year ago. New users encountering any of these documents will find missing features, wrong version numbers, and stale role names.

---

## Approach

Full sweep of all user-facing docs. No new tooling, no new site structure. Every file gets a systematic pass: version stamps corrected, new features documented, renamed concepts fixed, stale claims removed.

Explicitly out of scope: building a static site generator documentation portal (MkDocs/VitePress/Hugo) — tracked separately in TODO.md.

---

## Files in Scope

### 1. `docs/index.html` — GitHub Pages landing page

**Changes:**
- Footer version: `0.16.0-beta` → `1.2.0-beta`; copyright `2025` → `2026`
- Stats section: version number animation and completion percentage updated
- Features grid: add cards for Class Scheduling, Coach Role, Leaderboards, Subscription System (alongside existing 6)
- Tech stack: Go `1.21+` → `1.23+`

**Key new features to surface:**
- Class scheduling with reservations, waitlist, credits
- Three-tier role system (Athlete / Coach / Admin)
- PR Leaderboards and consistency achievements
- Subscription system with read-only mode enforcement

---

### 2. `docs/help/README.md` — End-user guide (largest gap: v0.12.2 → v1.2.0)

**Version stamp:** `0.12.2-beta` → `1.2.0-beta`, date → `2026-02-28`

**Role rename throughout:** "user" role → "athlete"; add Coach role explanation where relevant.

**New sections to add:**
- Class Scheduling — browsing sessions, making reservations, joining/leaving waitlist
- My Credits — viewing credit balance, packages, document status
- My Reservations — upcoming reservations, cancellation
- Coach Dashboard — session roster, check-in, no-show, complete session (Coach/Admin only)
- Leaderboards — PR leaderboard, consistency achievements
- Notification Likes — liking notifications, seeing likers

**Existing sections to update:**
- Getting Started: add subscription status context; note three roles
- Settings: add font customization (10 font options)
- Profile: add avatar upload
- FAQ/Troubleshooting: add subscription read-only mode explanation

**Table of contents:** regenerate to match all new and updated sections.

---

### 3. `docs/ROADMAP.md` — Development roadmap

**Changes:**
- Current version: `0.17.0-beta in development` → `1.2.0-beta released`
- Overall completion: update percentage and executive summary
- Version history: add v1.1.0-beta (Coach role, scheduling Phase 1-3) and v1.2.0-beta (scheduling Phase 4, leaderboards, beta logo) as released milestones
- Collapse pre-v1.0 history into a brief "Early Development" summary section to reduce clutter
- Future enhancements section: remove items already completed

---

### 4. `docs/USER_PERMISSIONS.md` — Role and permission matrix

**Changes:**
- Version stamp: `1.1.0-beta` → `1.2.0-beta`, date → `2026-02-28`
- Content is accurate — no structural changes needed

---

### 5. `docs/admin/README.md` — Administrator guide

**Changes:**
- Version stamp updated to `1.2.0-beta`
- Add sections for:
  - Scheduling administration (gym locations, class templates, schedule slots, sessions, coach assignments)
  - Package and document management (class packages, user documents/waivers)
  - Subscription management UI (create, mark paid, cancel, view history)
  - User import/export (bulk CSV import, export, password reset)

---

## Execution Order

1. `docs/USER_PERMISSIONS.md` — version stamp only, trivial, good warm-up
2. `docs/index.html` — landing page, high visibility, bounded changes
3. `docs/ROADMAP.md` — structural restructure, no new prose needed
4. `docs/admin/README.md` — new sections, moderate effort
5. `docs/help/README.md` — largest file, most new content, save for last

---

## Success Criteria

- No document references a version earlier than `1.2.0-beta`
- All major features present in v1.1.0 and v1.2.0 are represented in the help guide
- "user" role is not used where "athlete" is correct
- The landing page feature cards reflect the current feature set
- The roadmap accurately describes what is released vs. planned
