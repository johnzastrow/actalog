# Documentation Update Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Update all existing user-facing documentation and marketing site to accurately reflect ActaLog v1.2.0-beta, adding all features introduced since v0.12.2-beta.

**Architecture:** Content-only update — no new tooling, no site generator. Five files edited in order of increasing effort: USER_PERMISSIONS.md (stamp only), index.html (landing page), ROADMAP.md (restructure), admin/README.md (new sections), help/README.md (largest, most new content).

**Tech Stack:** Markdown, HTML/CSS. No build step required.

---

## Background: What Changed Between v0.12.2 and v1.2.0

Major features absent from current docs:
- **Three-tier roles:** "user" renamed to "athlete"; "coach" added as middle tier
- **Subscription system:** active/expired states, read-only mode at 402
- **Class Scheduling (Phase 1-4):** gym locations, class templates, schedule slots, sessions, coach assignments, reservations, waitlist, credits/packages, documents/waivers
- **Delete class with sessions:** three cascade modes
- **PR Leaderboards & consistency achievements**
- **Notification likes** (social — thumbs up on any notification)
- **User-customizable fonts** (10 options in Settings)
- **Avatar upload** in Profile
- **Configurable beta logo** via `LOGO_VARIANT` env var
- **Backup/restore modes:** replace, merge, skip
- **Import duplicate protection** across all import flows
- **OpenAPI/Swagger** at `/docs/swagger.json`

---

## Task 1: Update `docs/USER_PERMISSIONS.md` version stamp

**Files:**
- Modify: `docs/USER_PERMISSIONS.md` (lines 2-3)

**Step 1: Make the edit**

Change:
```
**Last Updated:** 2026-02-12
**Version:** 1.1.0-beta
```
To:
```
**Last Updated:** 2026-02-28
**Version:** 1.2.0-beta
```

**Step 2: Verify**

Run: `grep -n "Version\|Last Updated" docs/USER_PERMISSIONS.md`
Expected output:
```
2:**Last Updated:** 2026-02-28
3:**Version:** 1.2.0-beta
```

**Step 3: Commit**

```bash
git add docs/USER_PERMISSIONS.md
git commit -m "docs: bump USER_PERMISSIONS.md version to 1.2.0-beta"
```

---

## Task 2: Update `docs/index.html` — public landing page

**Files:**
- Modify: `docs/index.html`

The landing page is served via GitHub Pages at `https://johnzastrow.github.io/actalog/`. It currently shows v0.16.0-beta, Go 1.21+, and six feature cards that are missing all major features added since v0.16.

### Step 1: Fix the footer version and copyright

Find:
```html
                <div class="cta-meta-item">
                    <span class="cta-meta-label">Version:</span>
                    <span class="cta-meta-value">0.16.0-beta</span>
                </div>
```
Replace with:
```html
                <div class="cta-meta-item">
                    <span class="cta-meta-label">Version:</span>
                    <span class="cta-meta-value">1.2.0-beta</span>
                </div>
```

Find:
```html
            <p>&copy; 2025 ActaLog. Open source under MIT License.</p>
```
Replace with:
```html
            <p>&copy; 2026 ActaLog. Open source under MIT License.</p>
```

### Step 2: Fix the stats section version number

Find:
```html
                <div class="stat-item">
                    <div class="stat-number" data-target="0.16">0.0</div>
                    <div class="stat-label">CURRENT VERSION</div>
                    <div class="stat-sublabel">Beta Release</div>
                </div>
```
Replace with:
```html
                <div class="stat-item">
                    <div class="stat-number">1.2.0</div>
                    <div class="stat-label">CURRENT VERSION</div>
                    <div class="stat-sublabel">Beta Release</div>
                </div>
```

Find:
```html
                <div class="stat-item">
                    <div class="stat-number" data-target="93">0</div>
                    <div class="stat-label">% CORE COMPLETE</div>
                    <div class="stat-sublabel">Production Ready</div>
                </div>
```
Replace with:
```html
                <div class="stat-item">
                    <div class="stat-number">99</div>
                    <div class="stat-label">% CORE COMPLETE</div>
                    <div class="stat-sublabel">Production Ready</div>
                </div>
```

### Step 3: Fix Go version in tech stack

Find:
```html
                    <div class="tech-item">
                        <div class="tech-label">BACKEND</div>
                        <div class="tech-value">Go 1.21+ / Chi Router</div>
                    </div>
```
Replace with:
```html
                    <div class="tech-item">
                        <div class="tech-label">BACKEND</div>
                        <div class="tech-value">Go 1.23+ / Chi Router</div>
                    </div>
```

### Step 4: Add new feature cards

The existing `features-grid` div ends after the PWA card (`</div>` closing `features-grid`). Add three new cards **before** that closing `</div>`:

```html
                <div class="feature-card">
                    <div class="feature-icon">
                        <svg width="40" height="40" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                            <path d="M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2"/>
                            <circle cx="9" cy="7" r="4"/>
                            <path d="M23 21v-2a4 4 0 0 0-3-3.87"/>
                            <path d="M16 3.13a4 4 0 0 1 0 7.75"/>
                        </svg>
                    </div>
                    <h3 class="feature-title">CLASS SCHEDULING</h3>
                    <p class="feature-description">Full gym scheduling: locations, class templates, sessions, coach assignments, reservations, waitlist, and credit packages.</p>
                </div>

                <div class="feature-card">
                    <div class="feature-icon">
                        <svg width="40" height="40" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                            <polyline points="23 6 13.5 15.5 8.5 10.5 1 18"/>
                            <polyline points="17 6 23 6 23 12"/>
                        </svg>
                    </div>
                    <h3 class="feature-title">LEADERBOARDS</h3>
                    <p class="feature-description">PR leaderboards across the gym. Consistency achievements for showing up. Compare your lifts against the community.</p>
                </div>

                <div class="feature-card">
                    <div class="feature-icon">
                        <svg width="40" height="40" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                            <path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z"/>
                        </svg>
                    </div>
                    <h3 class="feature-title">MULTI-ROLE ACCESS</h3>
                    <p class="feature-description">Three-tier roles: Athlete, Coach, and Admin. Coaches manage rosters and check-ins. Admins control the full system.</p>
                </div>
```

### Step 5: Verify the HTML is valid

Run: `grep -c "feature-card" docs/index.html`
Expected: `9` (6 original + 3 new)

Also confirm no broken tags: `python3 -c "from html.parser import HTMLParser; p = HTMLParser(); p.feed(open('docs/index.html').read()); print('OK')" 2>&1`
Expected: `OK`

### Step 6: Commit

```bash
git add docs/index.html
git commit -m "docs: update landing page to v1.2.0-beta with new features"
```

---

## Task 3: Update `docs/ROADMAP.md` — development roadmap

**Files:**
- Modify: `docs/ROADMAP.md`

The roadmap currently describes v0.17.0-beta as "in development" and its executive summary is a year out of date. Rather than updating every historical version entry (unnecessary), restructure as follows.

### Step 1: Replace the header block

Find and replace the entire header section (everything from line 1 through the first `---`):

```markdown
# ActaLog Development Roadmap

**Current Version:** 1.2.0-beta (Released)
**Last Updated:** 2026-02-28
**Overall Completion:** ~99% of core requirements

---

## Executive Summary

ActaLog is a mobile-first Progressive Web App (PWA) for CrossFit workout tracking. The application is **production-ready** with all core features implemented: user authentication with email verification, three-tier role system (Athlete/Coach/Admin), workout logging, performance tracking and PR detection, import/export (Wodify CSV, JSON backup/restore), subscription billing, social features (notification likes), **full class scheduling** (gym locations, templates, sessions, coach assignments, reservations, waitlist, credit packages, and document management), PR leaderboards, and consistency achievements. The application deploys as a single Docker container with automatic database migrations and seed data across SQLite, PostgreSQL, and MariaDB.

---
```

### Step 2: Add v1.2.0-beta and v1.1.0-beta released entries

After the new header block, add the following two version entries **before** the existing version history content:

```markdown
## Released Versions

### v1.2.0-beta (Released 2026-02-19)
**Status:** Class scheduling Phase 4, PR leaderboards, configurable beta logo

**Highlights:**
- ✅ Class packages (credit system), waitlist management, class notifications
- ✅ User documents / waivers per gym
- ✅ Delete class templates with cascade modes (template only / future sessions / all sessions)
- ✅ PR Leaderboards — gym-wide personal record comparison
- ✅ Consistency achievements
- ✅ Configurable beta logo via `LOGO_VARIANT` environment variable
- ✅ Leaderboard search and header nav icon

---

### v1.1.0-beta (Released 2026-01-xx)
**Status:** Three-tier roles, class scheduling Phase 1-3, coach dashboard

**Highlights:**
- ✅ Renamed "user" role to "athlete"; added "coach" as middle tier
- ✅ `CoachOrAdmin` middleware; coaches bypass subscription checks
- ✅ Dedicated `/api/coaches/` routes (sessions, roster, check-in, no-show, complete)
- ✅ Gym locations, class templates, schedule slots, class sessions with capacity
- ✅ Coach assignments per gym; reservations with check-in flow
- ✅ Coach Dashboard frontend; Admin all-sessions visibility

---

## Early Development History (v0.12 – v0.16)

See [CHANGELOG.md](CHANGELOG.md) for the full release history prior to v1.0.

---
```

### Step 3: Update the Future Enhancements section

Find and replace items in "Future Enhancements (Post-MVP)" that are already completed. Remove or mark completed:
- Stripe Integration — still future
- Email Notifications — still future (partially done via SMTP)
- Documentation Website — still future (MkDocs/VitePress)

Leave those three as-is; they are genuinely future work.

### Step 4: Verify version header

Run: `head -5 docs/ROADMAP.md`
Expected first line: `# ActaLog Development Roadmap`
Expected line 3: `**Current Version:** 1.2.0-beta (Released)`

### Step 5: Commit

```bash
git add docs/ROADMAP.md
git commit -m "docs: update ROADMAP.md to reflect v1.2.0-beta released state"
```

---

## Task 4: Update `docs/admin/README.md` — administrator guide

**Files:**
- Modify: `docs/admin/README.md`

The admin guide is at v0.22.0-beta (January 2026) and is missing three major areas: subscription management UI, class scheduling administration, and the three-tier role system.

### Step 1: Update version stamp and table of contents header

Find:
```
**Version:** 0.22.0-beta
**Last Updated:** 2026-01-13
```
Replace with:
```
**Version:** 1.2.0-beta
**Last Updated:** 2026-02-28
```

### Step 2: Update the role description in Administrator Overview

Find in "Administrator Overview" → "How to Become an Administrator":
```
- Admin status is controlled by the `role` field in the users table (`admin` vs `user`)
```
Replace with:
```
- Admin status is controlled by the `role` field in the users table. ActaLog has three roles: `athlete` (default), `coach`, and `admin`
```

Find in the overview list:
```
- Full user account management (view, unlock, disable, delete, role changes)
```
Replace with:
```
- Full user account management (view, unlock, disable, delete, role changes to athlete/coach/admin)
```

### Step 3: Update TOC to add new sections

Replace the existing Table of Contents block with:

```markdown
## Table of Contents

1. [Administrator Overview](#administrator-overview)
2. [System Metrics Dashboard](#system-metrics-dashboard)
3. [User Account Management](#user-account-management)
4. [User Import/Export](#user-importexport)
5. [Subscription Management](#subscription-management)
6. [Class Scheduling Administration](#class-scheduling-administration)
7. [Database Backup and Restore](#database-backup-and-restore)
8. [Audit Log Monitoring](#audit-log-monitoring)
9. [Data Change Logs](#data-change-logs)
10. [System Configuration](#system-configuration)
11. [Security Best Practices](#security-best-practices)
12. [Database Management](#database-management)
13. [Troubleshooting](#troubleshooting)
14. [API Reference](#api-reference)
15. [Admin FAQ](#admin-faq)
```

### Step 4: Add Subscription Management section

Insert the following section after the existing "User Import/Export" section (before "Database Backup and Restore"):

```markdown
---

## Subscription Management

ActaLog includes a built-in subscription billing system. Admins control all subscription operations manually; there is no automated payment processor by default.

### Access

Navigate to **Profile** → **Admin** → **Subscriptions**.

### Subscription Types

| Type | Description |
|------|-------------|
| **User Subscription** | Applies to a single athlete account |
| **Organization Subscription** | Applies to all members of a gym organization |
| **Permanent Free** | Never expires; used for special accounts |

### Subscription States

- **Active** — Athlete has full read/write access to all features
- **Expired / None** — Athlete enters read-only mode (HTTP 402 on write operations). They can still view all their historical data.

Coaches and Admins are **never** subject to subscription enforcement regardless of subscription state.

### Common Admin Tasks

**Create a subscription:**
1. Go to **Subscriptions** tab
2. Click **Create Subscription**
3. Select user or organization, set expiration date, choose type
4. Click **Save**

**Mark a subscription as paid:**
1. Find the subscription in the list
2. Click the **Mark as Paid** button
3. Confirm the action

**Cancel a subscription:**
1. Find the subscription
2. Click **Cancel**
3. Enter a cancellation reason (for audit trail)

**View expiring subscriptions:**
- The **Expiring Soon** tab shows subscriptions expiring within 30 days

**View expired subscriptions:**
- The **Expired** tab lists all lapsed subscriptions

### Subscription Enforcement

When an athlete's subscription expires:
- All `POST`, `PUT`, `PATCH`, `DELETE` requests to feature routes return **HTTP 402 Payment Required**
- `GET` requests continue to succeed — athletes never lose access to their data
- A banner appears in the UI prompting renewal
- Routes exempt from enforcement: profile, settings, password change, notifications, subscription status, login/logout

---
```

### Step 5: Add Class Scheduling Administration section

Insert the following section after "Subscription Management" (before "Database Backup and Restore"):

```markdown
---

## Class Scheduling Administration

The class scheduling system allows gyms to manage locations, class templates, session schedules, coach assignments, reservations, waitlists, credit packages, and required documents.

Access scheduling administration via **Profile** → **Admin** → **Scheduling**.

### Gym Locations

Locations represent physical gym facilities. Each organization can have multiple locations.

- **Create:** Click **Add Location**, enter name and address
- **Edit:** Click location row → edit form
- **Delete:** Available if no active sessions are attached

### Class Templates

Templates define the blueprint for a recurring class type (e.g., "6am CrossFit", "Open Gym").

Fields:
- **Name** — Display name for the class
- **Location** — Which gym location
- **Capacity** — Maximum athletes per session
- **Duration** — Minutes
- **Description** — Optional notes shown to athletes

**Delete modes:** When deleting a template, choose:
- `Template only` — Sessions become orphaned (still visible in history)
- `With future sessions` — Deletes template and all future sessions; past sessions preserved
- `With all sessions` — Deletes template and every associated session (use with caution)

Credit refunds are issued automatically for unconfirmed reservations on deleted sessions.

### Schedule Slots

Slots attach a template to a recurring time (e.g., every Monday/Wednesday/Friday at 6:00am). The system generates individual class sessions from slots.

### Class Sessions

Individual occurrences of a class. Each session has:
- Date/time, capacity, coach assignment(s)
- Reservation list, waitlist, attendance records

**Session actions:**
- View roster and check-in status
- Add/remove coaches
- Cancel individual sessions

### Coach Assignments

Coaches must be assigned to an organization before they can manage its sessions. Assign coaches via **Admin** → **Users** → select user → change role to **Coach**, then assign to organization(s) via the scheduling panel.

A coach must have:
1. The `coach` role (set in Admin > Users)
2. An assignment to the specific organization (set in Scheduling admin)

### Credit Packages

Packages define how athletes purchase class access (e.g., "10-Class Pack at $100", "Monthly Unlimited").

- **Create:** Click **Add Package**, set name, credit count, price, expiration days
- **Issue credits to user:** Admin > Users > select user > issue credits from package

### User Documents

Documents are gym-defined forms that athletes must complete before attending classes (waivers, liability forms, health questionnaires).

- **Create document type:** Set name, whether required, expiration period
- **Mark as completed:** Admin can mark a user's document as completed on their behalf
- Athletes view their document status in **My Credits**

### Waitlist Management

When a class is at capacity, athletes can join the waitlist. If a reservation is cancelled, the first waitlisted athlete is automatically promoted and notified.

---
```

### Step 6: Verify new sections are present

Run: `grep -n "## Subscription Management\|## Class Scheduling" docs/admin/README.md`
Expected: two matching lines with correct section headers.

### Step 7: Commit

```bash
git add docs/admin/README.md
git commit -m "docs: add subscription and scheduling sections to admin guide, bump to v1.2.0-beta"
```

---

## Task 5: Update `docs/help/README.md` — end-user guide

**Files:**
- Modify: `docs/help/README.md`

This is the largest update. The file is at v0.12.2-beta (November 2025) and is missing all major features from the last 14 months.

### Step 1: Update version stamp

Find:
```
**Version:** 0.12.2-beta
**Last Updated:** 2025-11-28
```
Replace with:
```
**Version:** 1.2.0-beta
**Last Updated:** 2026-02-28
```

### Step 2: Replace the Table of Contents

Replace the existing TOC with:

```markdown
## Table of Contents

1. [Getting Started](#getting-started)
2. [Understanding Your Role](#understanding-your-role)
3. [Subscription Status](#subscription-status)
4. [Logging Your First Workout](#logging-your-first-workout)
5. [Tracking Personal Records (PRs)](#tracking-personal-records-prs)
6. [Using Quick Log](#using-quick-log)
7. [Creating and Using Workout Templates](#creating-and-using-workout-templates)
8. [Viewing Performance Trends](#viewing-performance-trends)
9. [Leaderboards](#leaderboards)
10. [Class Schedule](#class-schedule)
11. [My Reservations](#my-reservations)
12. [My Credits and Documents](#my-credits-and-documents)
13. [Coach Dashboard](#coach-dashboard)
14. [Notifications and Likes](#notifications-and-likes)
15. [Importing Data from Wodify](#importing-data-from-wodify)
16. [Exporting and Backing Up Your Data](#exporting-and-backing-up-your-data)
17. [Profile and Settings](#profile-and-settings)
18. [Installing the Progressive Web App (PWA)](#installing-the-progressive-web-app-pwa)
19. [FAQ](#frequently-asked-questions)
20. [Troubleshooting](#troubleshooting)
21. [Glossary](#glossary)
```

### Step 3: Add "Understanding Your Role" section

Insert after the "Getting Started" section:

```markdown
---

## Understanding Your Role

ActaLog has three user roles. Your role is shown in your profile.

| Role | Who It's For | What You Can Do |
|------|-------------|-----------------|
| **Athlete** | All regular members | Log workouts, track PRs, browse classes, make reservations, view leaderboards |
| **Coach** | Gym coaches | Everything an athlete can do, plus manage class rosters, check in athletes, mark no-shows |
| **Admin** | Gym administrators | Everything a coach can do, plus full system management (users, scheduling, subscriptions, backups) |

The first account created on a new ActaLog installation is automatically assigned the Admin role. All subsequent registrations are Athletes by default. An Admin can change any user's role via Admin → Users.

---
```

### Step 4: Add "Subscription Status" section

Insert after "Understanding Your Role":

```markdown
---

## Subscription Status

Access to ActaLog features requires an active subscription (configured by your gym admin).

### Active Subscription

You have full read and write access to all features.

### Expired or No Subscription

You enter **read-only mode**:
- You can still view all your historical workout data, PRs, and performance charts
- You cannot log new workouts, create templates, or make reservations
- A banner at the top of the screen will prompt you to contact your admin to renew

### Always Available (regardless of subscription)

- Viewing and editing your profile
- Changing your password and settings
- Viewing notifications
- Logging in and out

**Note:** Coaches and Admins are never subject to subscription restrictions.

---
```

### Step 5: Add "Leaderboards" section

Insert after the existing "Viewing Performance Trends" section:

```markdown
---

## Leaderboards

The Leaderboards page lets you compare your personal records against other athletes in your gym.

### Accessing Leaderboards

Navigate to the **Leaderboards** icon in the bottom navigation bar (trophy icon).

### PR Leaderboard

- Shows the top performances for a selected movement across all athletes
- Filter by movement using the search bar
- Your own entry is highlighted

### Consistency Achievements

Consistency achievements reward showing up regularly:
- Achievements are awarded based on workout frequency milestones
- View your achievements and compare with the community on the leaderboard

---
```

### Step 6: Add "Class Schedule" section

Insert after "Leaderboards":

```markdown
---

## Class Schedule

If your gym has class scheduling enabled, you can browse upcoming sessions and reserve your spot.

### Accessing the Schedule

Navigate to **Schedule** in the bottom navigation (calendar icon).

### Browsing Sessions

- Sessions are shown by date with class name, time, location, coach, and available spots
- Tap a session to see full details

### Making a Reservation

1. Tap the session you want to attend
2. Click **Reserve**
3. Your spot is confirmed — you'll receive a notification

Requirements: You must have an active subscription and available credits (if your gym uses the credit system).

### Joining the Waitlist

If a class is full:
1. Tap the session
2. Click **Join Waitlist**
3. You'll be automatically promoted and notified if a spot opens

### Cancelling a Reservation

1. Go to **My Reservations** (or tap the session)
2. Click **Cancel Reservation**

Cancelling may refund your credit depending on your gym's policy.

---
```

### Step 7: Add "My Reservations" section

```markdown
---

## My Reservations

View all your upcoming class reservations in one place.

### Accessing My Reservations

Navigate to **Profile** → **My Reservations**, or tap the reservations shortcut if shown on your dashboard.

### What You'll See

- Upcoming session date, time, class name, location
- Reservation status (confirmed, waitlisted, checked in)
- Cancel button for reservations you no longer need

---
```

### Step 8: Add "My Credits and Documents" section

```markdown
---

## My Credits and Documents

If your gym uses the credit system, this section shows your class credit balance and required document status.

### Accessing My Credits

Navigate to **Profile** → **My Credits**.

### Credits

- **Balance:** How many class credits you have remaining
- **Expiration:** Credits may have an expiration date set by your gym
- Credits are deducted when you make a reservation
- Credits may be refunded when you cancel (depending on gym policy)

### Documents

Your gym may require you to complete forms before attending classes (liability waivers, health forms, etc.):
- **Pending** — Document needs to be completed; contact your gym admin
- **Completed** — Document is on file
- **Expired** — Document needs to be renewed (e.g., annual waiver)

### Waitlist

Any classes you're currently waitlisted for are also shown here.

---
```

### Step 9: Add "Coach Dashboard" section

```markdown
---

## Coach Dashboard

*This section applies to users with the Coach or Admin role.*

The Coach Dashboard gives coaches visibility into their assigned sessions and tools to manage class check-ins.

### Accessing the Coach Dashboard

Navigate to the **Coach** icon in the bottom navigation (visible only to coaches and admins).

### Viewing Your Sessions

- Coaches see their upcoming assigned sessions
- Admins see ALL upcoming sessions across all gyms

### Managing a Session

Tap a session to open the roster view:

| Action | How |
|--------|-----|
| **Check in athlete** | Tap the athlete's name → Check In |
| **Mark no-show** | Tap athlete → No Show |
| **Complete session** | Tap **Complete Session** after class ends |

Completing a session locks the roster and finalizes attendance records.

---
```

### Step 10: Add "Notifications and Likes" section

Insert after "Coach Dashboard":

```markdown
---

## Notifications and Likes

### Notifications

ActaLog sends notifications for:
- New personal records (PR achievements)
- Class reservation confirmations and reminders
- Waitlist promotions (when a spot opens for you)
- Gym-wide announcements from your admin

Access notifications via the **bell icon** in the navigation bar. Unread notifications are highlighted. Tap any notification to mark it as read.

### Liking Notifications

You can react to any notification (yours or visible community notifications) with a thumbs-up like:

1. Open the **Notifications** page
2. Tap the 👍 icon on any notification
3. The like count updates immediately

Liking a notification marks it as unread for the recipient so they see your reaction.

---
```

### Step 11: Add Profile and Settings section (replacing any outdated existing coverage)

Find and replace (or add if missing) a "Profile and Settings" section with:

```markdown
---

## Profile and Settings

### Profile

Access your profile via the **Profile** icon in the bottom navigation.

**Avatar:** Tap your avatar image to upload a new photo. Supported formats: JPG, PNG, WebP. Tap anywhere on the avatar area to open the file picker.

**Account details:** View your email, role, and subscription status.

### Settings

Navigate to **Profile** → **Settings** to customize your experience:

**Font:** Choose from 10 font options including accessibility-optimized fonts:
- System Default, Inter, Roboto, Lato, Fira Sans, Lexend
- OpenDyslexic *(accessibility)*, Atkinson Hyperlegible *(accessibility)*
- Source Serif Pro, JetBrains Mono

Your font preference is synced to your account and applies across all devices.

**Theme:** Choose a color theme (including Sunrise and other options).

**Password:** Change your password from the Settings page.

---
```

### Step 12: Update the Glossary

Find the `## Glossary` section and add missing terms:

```markdown
| **Athlete** | The default role for all ActaLog users. Previously called "User" in older versions. |
| **Coach** | A role with access to class roster management and check-in capabilities for assigned gyms. |
| **Credit** | A unit of class access purchased via a credit package. One credit is typically consumed per class reservation. |
| **Waitlist** | A queue for a full class session. Athletes are automatically promoted when a spot opens. |
| **Leaderboard** | A gym-wide comparison of personal records across all athletes. |
| **Session** | A specific scheduled occurrence of a class (e.g., Monday 6am CrossFit on March 3rd). |
| **Template** | A class blueprint (name, capacity, location, duration) used to generate sessions. |
| **Subscription** | A time-limited access grant. Expired subscriptions put athletes in read-only mode. |
```

### Step 13: Verify section count

Run: `grep -c "^## " docs/help/README.md`
Expected: ≥ 18 (new sections added)

Run: `grep "^## " docs/help/README.md`
Expected output includes: Getting Started, Understanding Your Role, Subscription Status, Leaderboards, Class Schedule, My Reservations, My Credits and Documents, Coach Dashboard, Notifications and Likes, Profile and Settings

### Step 14: Commit

```bash
git add docs/help/README.md
git commit -m "docs: overhaul help guide to v1.2.0-beta — add scheduling, coach, leaderboards, subscription sections"
```

---

## Task 6: Final verification pass

### Step 1: Check for stale version strings across all updated files

```bash
grep -rn "0\.12\|0\.16\|0\.17\|0\.22\|1\.1\.0-beta\|2025" \
  docs/index.html docs/help/README.md docs/ROADMAP.md \
  docs/USER_PERMISSIONS.md docs/admin/README.md
```

Expected: no matches (all stale versions replaced)

### Step 2: Check for old "user" role references where "athlete" is now correct

```bash
grep -n '"user" role\|user role\|role.*user\b' \
  docs/help/README.md docs/admin/README.md docs/index.html
```

Expected: no matches (or only matches that are genuinely about the word "user" as a generic noun, not the role name)

### Step 3: Confirm new sections exist

```bash
grep -l "Class Scheduling\|Coach Dashboard\|Leaderboard\|Subscription" \
  docs/help/README.md docs/admin/README.md docs/index.html
```

Expected: all three files listed

### Step 4: Final commit if any cleanup needed

```bash
git add -p  # review any remaining changes
git commit -m "docs: final cleanup pass on v1.2.0-beta documentation update"
```

---

## Task 7: Update `site/` — static marketing website

**Files:**
- Modify: `site/index.html`
- Modify: `site/features.html`
- Modify: `site/faq.html`
- Modify: `site/content/harvested_features.md`

The `site/` directory is a separate marketing website (distinct from `docs/index.html`). It has its own multi-page structure with an index, features page, FAQ, deploy guide, and tech page. The structured data in `site/index.html` shows `softwareVersion: "0.24.0-beta"` and the features page describes only two roles ("Admin" and "member"), missing Coach entirely, and makes no mention of class scheduling, leaderboards, or the credit system.

### Step 1: Fix structured data version in `site/index.html`

Find:
```json
"softwareVersion": "0.24.0-beta",
```
Replace with:
```json
"softwareVersion": "1.2.0-beta",
```

### Step 2: Update role description in `site/features.html`

Find:
```html
<h3 class="feature-title">ROLE-BASED ACCESS</h3>
<p class="feature-description">Admin, and member roles. To manage different parts of the system.</p>
```
Replace with:
```html
<h3 class="feature-title">ROLE-BASED ACCESS</h3>
<p class="feature-description">Three roles: Athlete (default), Coach (roster & check-in access), and Admin (full system control). Coaches bypass subscription checks and manage sessions at their assigned gyms.</p>
```

### Step 3: Add class scheduling feature card to `site/features.html`

Locate the member management features grid section. After the last existing feature card in that section (before the closing `</div>` of the grid), add:

```html
<div class="feature-card">
    <h3 class="feature-title">CLASS SCHEDULING</h3>
    <p class="feature-description">Full gym scheduling: locations, class templates, recurring schedule slots, and individual sessions with capacity management. Athletes reserve spots, join waitlists, and get automatic notifications when a spot opens.</p>
</div>

<div class="feature-card">
    <h3 class="feature-title">CREDIT PACKAGES</h3>
    <p class="feature-description">Sell class packages (e.g., "10-Class Pack", "Monthly Unlimited"). Credits deduct automatically on reservation. Refunds issued when sessions are cancelled. Expiration tracking built in.</p>
</div>

<div class="feature-card">
    <h3 class="feature-title">LEADERBOARDS</h3>
    <p class="feature-description">Gym-wide PR leaderboards for any movement. Consistency achievement tracking. Athletes see how their lifts stack up against the community — motivation built in.</p>
</div>
```

### Step 4: Fix stale role references in `site/features.html`

Find:
```html
With role-based access control, you decide who sees what. Admins have full control. Members see only their own data.
```
Replace with:
```html
With three-tier role-based access control, you decide who sees what. Admins have full control. Coaches manage class rosters and check-ins. Athletes see only their own data.
```

### Step 5: Add scheduling FAQ entry to `site/faq.html`

Find the first `<div class="faq-item">` in the file and insert a new FAQ item before it:

```html
<div class="faq-item">
    <div class="faq-question">Does ActaLog support class scheduling and reservations?</div>
    <div class="faq-answer">
        <p>Yes. ActaLog includes a full class scheduling system: gym locations, class templates, recurring schedule slots, and individual sessions with capacity management. Athletes can reserve spots, join waitlists, and receive automatic notifications when a spot opens. Admins sell credit packages (e.g., "10-Class Pack") and track document requirements (waivers, liability forms). Coaches can check in athletes and mark attendance from the Coach Dashboard.</p>
    </div>
</div>

<div class="faq-item">
    <div class="faq-question">What are the user roles in ActaLog?</div>
    <div class="faq-answer">
        <p>ActaLog has three roles:</p>
        <ul>
            <li><strong>Athlete</strong> — Default role. Full workout tracking, class reservations, and leaderboard access.</li>
            <li><strong>Coach</strong> — Everything an athlete can do, plus class roster management and check-in capabilities for assigned gyms. Coaches are never subject to subscription restrictions.</li>
            <li><strong>Admin</strong> — Full system access including user management, class scheduling configuration, subscription billing, backups, and audit logs.</li>
        </ul>
    </div>
</div>
```

### Step 6: Update `site/content/harvested_features.md`

Replace the "Recent Notable Releases" section with:

```markdown
## Recent Notable Releases

### v1.2.0-beta (Released 2026-02-19)
- Class packages (credit system) and waitlist management
- User documents / waivers per gym organization
- Delete class templates with cascade modes
- PR Leaderboards — gym-wide personal record comparison
- Consistency achievements
- Configurable beta logo via `LOGO_VARIANT` environment variable

### v1.1.0-beta
- Renamed "user" role to "athlete"; added "coach" as middle tier
- Three-tier role system: Athlete < Coach < Admin
- Dedicated coach API routes and Coach Dashboard
- Class Scheduling Phase 1-3: gym locations, templates, sessions, coach assignments, reservations

### v0.16.0-beta — Notification Likes
- Users can like any notification (PR achievements, announcements)
- Social engagement: thumbs-up icon with liker count

### v0.14.0-beta — Subscription Billing System
- Dual-level subscriptions: user and organization
- Read-only mode enforcement when subscriptions expire (HTTP 402)
- Admin subscription management UI
```

Also update the "High-priority Backlog" section — mark subscription frontend as completed:

```markdown
## Current Status

All core features are implemented and production-ready as of v1.2.0-beta:
- ✅ Subscription management UI (frontend complete)
- ✅ Class scheduling (Phase 1-4 complete)
- ✅ Three-tier role system (Athlete/Coach/Admin)
- ✅ PR Leaderboards and consistency achievements
- ⏳ Stripe integration (future)
- ⏳ Documentation website / static site generator (future)
```

### Step 7: Verify version and role fixes

```bash
grep -n "softwareVersion\|0\.24\|member role\|Admin.*member" \
  site/index.html site/features.html site/faq.html
```
Expected: no matches (all stale references replaced)

```bash
grep -n "1\.2\.0-beta\|Athlete\|Coach.*Admin\|class scheduling" \
  site/index.html site/features.html site/faq.html -i | head -20
```
Expected: matches in all three files confirming updates landed.

### Step 8: Commit

```bash
git add site/index.html site/features.html site/faq.html site/content/harvested_features.md
git commit -m "docs: update site/ marketing pages to v1.2.0-beta — add scheduling, coach role, leaderboards"
```
