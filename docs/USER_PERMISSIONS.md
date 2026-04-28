# User Permissions Matrix

**Last Updated:** 2026-04-28
**Version:** 1.2.4

This document details all user actions and their permission requirements in ActaLog.

---

## Role Summary

ActaLog has three user roles: **Athlete**, **Coach**, and **Admin**. Every person who registers receives the Athlete role by default, with one exception: the very first account created on a fresh installation is automatically promoted to Admin. After that, only an existing Admin can change another person's role.

### Athletes

An athlete can track personal workouts, browse the movement and WOD libraries, log results, view personal records, analyze performance trends, and manage their own profile and settings. They can also browse class schedules, make reservations, join waitlists, and view their credits and documents.

All of these capabilities are gated behind an **active subscription**. When an athlete's subscription is active, they have full read and write access to every feature route. When their subscription expires, they enter a **read-only mode**: they can still view their data (GET requests succeed), but any attempt to create, update, or delete records on feature routes returns HTTP 402 (Payment Required). This ensures athletes never lose visibility into their history while encouraging renewal.

A handful of routes are exempt from subscription enforcement entirely. Account essentials like viewing and editing your profile, managing settings, changing your password, viewing notifications, checking subscription status, and managing login sessions always work regardless of subscription state. These are considered security-critical or baseline account operations that should never be locked behind a paywall.

Athletes are strictly scoped to their own data. They cannot view other users' workouts, personal records, or account details. All user-facing queries are filtered by the authenticated user's ID at the service layer.

### Coaches

Coaches inherit every capability an athlete has, plus access to the Coach Dashboard where they can manage class sessions for their assigned gyms. Coach capabilities include: viewing session rosters, checking in athletes, marking no-shows, and completing sessions.

Coaches access these features through dedicated `/api/coaches/` routes protected by `CoachOrAdmin` middleware. The service layer additionally verifies that a coach is assigned to the session's organization before allowing roster/check-in actions.

Critically, **coaches bypass all subscription checks** (like admins). A coach's access is never degraded by subscription status.

Coach actions do NOT include: template/location/schedule management, user management, or any other admin-only operations. Those remain restricted to the Admin role.

### Administrators

Admins inherit every capability a coach has, plus full access to all administrative functions: user management (list, disable, enable, unlock, delete, role changes), organization and subscription management, class scheduling configuration (locations, templates, schedule slots, sessions, coaches), document and package management, data quality tools, backup and restore, email configuration, audit logs, and system metrics.

Critically, **admins bypass all subscription checks**. An admin's access is never degraded by subscription status. This ensures the system can always be managed even if the admin's own subscription has lapsed.

Admins also have exclusive access to bulk operations: importing and exporting users, sending batch password reset emails, promoting user-created content to the standard library, and running data quality scans and duplicate merges.

### Constraints and Exceptions

- **One role at a time.** A user is Athlete, Coach, or Admin, never multiple. Role changes take effect immediately.
- **Rate limiting applies to everyone.** Public authentication endpoints (login, register, password reset) are rate-limited by IP regardless of role. Login and registration allow 5 attempts per 15 minutes; password reset allows 3 attempts per hour.
- **Coach role + coach assignments.** The Coach role grants access to coach routes. Coach assignments (via `coach_assignments` table) determine which specific gyms a coach can manage. Both are required: a user must have the coach role AND be assigned to an organization to manage its sessions.
- **Protected accounts.** Certain system accounts are protected and must never have their data modified, regardless of who is logged in. See CLAUDE.md for the protected accounts list.

---

## Role Overview

| Role | Description | Assignment |
|------|-------------|------------|
| **Athlete** | Regular user with access to personal workout tracking features | Default role for all new registrations |
| **Coach** | Elevated user with roster/check-in access for assigned gyms | Assigned by admin via Admin > Users |
| **Admin** | Full system access including user management and system configuration | First registered user; manually assigned thereafter |

**Notes:**
- A user has exactly one role at any time
- Role changes require admin action via Admin > Users
- Subscription status affects feature access for athletes only (coaches and admins bypass subscription checks)
- Expired subscription athletes get read-only access: GET requests succeed, POST/PUT/DELETE return HTTP 402

---

## Permission Legend

| Symbol | Meaning |
|--------|---------|
| Y | Yes - Action permitted |
| N | No - Action not permitted |
| S | Subscription required (User needs active subscription; Admin always permitted) |
| O | Own data only (User can only access their own records) |

---

## Account & Authentication

| Action | Athlete | Admin | API Endpoint | Notes |
|--------|---------|-------|-------------|-------|
| Register new account | Y | Y | `POST /api/auth/register` | Public, rate limited (5/15min per IP) |
| Login | Y | Y | `POST /api/auth/login` | Public, rate limited (5/15min per IP) |
| Logout | Y | Y | Client-side token removal | |
| Request password reset | Y | Y | `POST /api/auth/forgot-password` | Public, rate limited (3/hr per IP) |
| Reset password with token | Y | Y | `POST /api/auth/reset-password` | Public, rate limited (3/hr per IP) |
| Verify email address | Y | Y | `GET /api/auth/verify-email` | Public, via email link |
| Resend verification email | Y | Y | `POST /api/auth/resend-verification` | Public, rate limited (5/15min per IP) |
| Refresh JWT token | Y | Y | `POST /api/auth/refresh` | Public |
| Revoke refresh token | Y | Y | `POST /api/auth/revoke` | Public |
| View own profile | Y | Y | `GET /api/users/profile` | |
| Edit own profile (name, email, birthday) | Y | Y | `PUT /api/users/profile` | |
| Upload own avatar | Y | Y | `POST /api/users/avatar` | Image-only (magic-byte sniffed); allowed extensions: `.jpg`/`.jpeg`/`.png`/`.gif`/`.webp`; max 5MB |
| Delete own avatar | Y | Y | `DELETE /api/users/avatar` | |
| Change own password | Y | Y | `PUT /api/users/password` | |
| View own settings | Y | Y | `GET /api/users/settings` | |
| Update own settings | Y | Y | `PUT /api/users/settings` | |
| View own active sessions | Y | Y | `GET /api/sessions` | |
| Revoke specific session | Y | Y | `DELETE /api/sessions/{id}` | Own sessions only |
| Revoke all own sessions | Y | Y | `POST /api/sessions/revoke-all` | |
| View own audit logs | Y | Y | `GET /api/users/me/audit-logs` | |
| Check own subscription status | Y | Y | `GET /api/subscriptions/status` | |

---

## Movements Library

| Action | Athlete | Admin | API Endpoint | Notes |
|--------|---------|-------|-------------|-------|
| View all movements | Y | Y | `GET /api/movements` | Public, includes standard and custom |
| Search movements | Y | Y | `GET /api/movements/search` | Public |
| View movement details | Y | Y | `GET /api/movements/{id}` | Public |
| Create custom movement | S | Y | `POST /api/movements` | User-owned, subscription required |
| Edit own custom movement | S | Y | `PUT /api/movements/{id}` | Own movements only |
| Delete own custom movement | S | Y | `DELETE /api/movements/{id}` | Own movements only |
| Edit standard movement | N | Y | `PUT /api/movements/{id}` | |
| Delete standard movement | N | Y | `DELETE /api/movements/{id}` | |
| Toggle PR tracking on movement | S | Y | `POST /api/movements/toggle-pr` | |
| View movement performance history | S | Y | `GET /api/performance/movements/{id}` | Own history only |

---

## WODs (Workouts of the Day)

| Action | Athlete | Admin | API Endpoint | Notes |
|--------|---------|-------|-------------|-------|
| View all WODs | Y | Y | `GET /api/wods` | Public, includes standard and custom |
| View standard WODs | Y | Y | `GET /api/wods/standard` | Public |
| Search WODs | Y | Y | `GET /api/wods/search` | Public |
| View WOD details | Y | Y | `GET /api/wods/{id}` | Public |
| View own custom WODs | S | Y | `GET /api/wods/my-wods` | |
| Create custom WOD | S | Y | `POST /api/wods` | User-owned |
| Edit own custom WOD | S | Y | `PUT /api/wods/{id}` | Own WODs only |
| Delete own custom WOD | S | Y | `DELETE /api/wods/{id}` | Own WODs only |
| Edit standard WOD | N | Y | `PUT /api/wods/{id}` | |
| Delete standard WOD | N | Y | `DELETE /api/wods/{id}` | |
| View WOD performance history | S | Y | `GET /api/performance/wods/{id}` | Own history only |

---

## Workout Templates

| Action | Athlete | Admin | API Endpoint | Notes |
|--------|---------|-------|-------------|-------|
| View standard templates | Y | Y | `GET /api/templates` | Public |
| View template details | Y | Y | `GET /api/templates/{id}` | Public |
| View own templates | S | Y | `GET /api/workouts/my-templates` | |
| Create workout template | S | Y | `POST /api/templates` | User-owned |
| Edit own template | S | Y | `PUT /api/templates/{id}` | Own templates only |
| Delete own template | S | Y | `DELETE /api/templates/{id}` | Own templates only |
| Add WOD to own template | S | Y | `POST /api/templates/{workout_id}/wods` | |
| List WODs in template | S | Y | `GET /api/templates/{workout_id}/wods` | |
| Update WOD in own template | S | Y | `PUT /api/templates/wods/{workout_wod_id}` | |
| Remove WOD from own template | S | Y | `DELETE /api/templates/wods/{workout_wod_id}` | |
| Toggle PR flag on template WOD | S | Y | `POST /api/templates/wods/{workout_wod_id}/toggle-pr` | |
| Edit standard template | N | Y | `PUT /api/templates/{id}` | |
| Delete standard template | N | Y | `DELETE /api/templates/{id}` | |

---

## Workout Logging

| Action | Athlete | Admin | API Endpoint | Notes |
|--------|---------|-------|-------------|-------|
| Log a workout | S | Y | `POST /api/workouts` | Creates user workout record |
| View own workout history | S | Y | `GET /api/workouts` | |
| View standard workouts | S | Y | `GET /api/workouts/standard` | |
| View own workout details | S | Y | `GET /api/workouts/{id}` | |
| Edit own logged workout | S | Y | `PUT /api/workouts/{id}` | |
| Delete own logged workout | S | Y | `DELETE /api/workouts/{id}` | |
| View monthly workout statistics | S | Y | `GET /api/workouts/stats/monthly` | Own stats |
| View workout personal records | S | Y | `GET /api/workouts/personal-records` | |
| Flag retroactive PRs | S | Y | `POST /api/workouts/retroactive-flag-prs` | Recalculate historical PRs |
| View other users' workouts | N | Y | | Admin can view all |

---

## Personal Records (PRs)

| Action | Athlete | Admin | API Endpoint | Notes |
|--------|---------|-------|-------------|-------|
| View own personal records | S | Y | `GET /api/prs` | |
| View movements with PRs | S | Y | `GET /api/pr-movements` | Own PR movements |
| Toggle movement PR tracking | S | Y | `POST /api/movements/toggle-pr` | |
| View other users' PRs | N | Y | | Admin can view all |

---

## Performance Analytics

| Action | Athlete | Admin | API Endpoint | Notes |
|--------|---------|-------|-------------|-------|
| Search own performance data | S | Y | `GET /api/performance/search` | Unified search |
| View movement performance trends | S | Y | `GET /api/performance/movements/{id}` | Own data |
| View WOD performance trends | S | Y | `GET /api/performance/wods/{id}` | Own data |
| View active users statistics | S | Y | `GET /api/stats/active-users-this-month` | Anonymized stats |

---

## Notifications

| Action | Athlete | Admin | API Endpoint | Notes |
|--------|---------|-------|-------------|-------|
| View own notifications | Y | Y | `GET /api/notifications` | No subscription required |
| View unread notifications | Y | Y | `GET /api/notifications/unread` | |
| Get unread notification count | Y | Y | `GET /api/notifications/count` | |
| Mark notification as read | Y | Y | `PUT /api/notifications/{id}/read` | |
| Mark all notifications as read | Y | Y | `PUT /api/notifications/read-all` | |
| Delete own notification | Y | Y | `DELETE /api/notifications/{id}` | |
| Like a notification | S | Y | `POST /api/notifications/{id}/like` | |
| Unlike a notification | S | Y | `DELETE /api/notifications/{id}/like` | |
| View notification likes | Y | Y | `GET /api/notifications/{id}/likes` | |
| Create system announcement | N | Y | `POST /api/admin/notifications/announce` | Sends to all users |

---

## Class Scheduling (User-Facing)

### Browsing Classes

| Action | Athlete | Admin | API Endpoint | Notes |
|--------|---------|-------|-------------|-------|
| View gym sessions list | Y | Y | `GET /api/gyms/{gym_id}/sessions` | Auth required, no subscription |
| View session details | Y | Y | `GET /api/gyms/{gym_id}/sessions/{id}` | Auth required, no subscription |
| View gym locations | Y | Y | `GET /api/gyms/{gym_id}/locations` | Auth required, no subscription |
| View location details | Y | Y | `GET /api/gyms/{gym_id}/locations/{id}` | Auth required, no subscription |
| View gym class templates | Y | Y | `GET /api/gyms/{gym_id}/templates` | Auth required, no subscription |

### Reservations

| Action | Athlete | Admin | API Endpoint | Notes |
|--------|---------|-------|-------------|-------|
| Reserve a session spot | S | Y | `POST /api/sessions/{session_id}/reserve` | Subscription required |
| Cancel reservation | S | Y | `DELETE /api/sessions/{session_id}/reserve` | Subscription required |
| View own upcoming reservations | Y | Y | `GET /api/users/me/reservations/upcoming` | No subscription required |

### Waitlist

| Action | Athlete | Admin | API Endpoint | Notes |
|--------|---------|-------|-------------|-------|
| View session waitlist | Y | Y | `GET /api/sessions/{session_id}/waitlist` | Auth required, no subscription |
| View own waitlist position | Y | Y | `GET /api/sessions/{session_id}/waitlist/position` | Auth required, no subscription |
| Join session waitlist | S | Y | `POST /api/sessions/{session_id}/waitlist` | Subscription required |
| Leave session waitlist | S | Y | `DELETE /api/sessions/{session_id}/waitlist` | Subscription required |
| View own waitlist entries | Y | Y | `GET /api/users/me/waitlist` | No subscription required |

---

## Documents & Packages (User-Facing)

| Action | Athlete | Admin | API Endpoint | Notes |
|--------|---------|-------|-------------|-------|
| View gym documents | Y | Y | `GET /api/gyms/{gym_id}/documents` | Auth required, no subscription |
| View document details | Y | Y | `GET /api/gyms/{gym_id}/documents/{id}` | Auth required, no subscription |
| View own documents (all gyms) | Y | Y | `GET /api/users/me/documents` | No subscription required |
| View own documents (by gym) | Y | Y | `GET /api/gyms/{gym_id}/users/me/documents` | Auth required, no subscription |
| View pending documents | Y | Y | `GET /api/gyms/{gym_id}/users/me/documents/pending` | Auth required, no subscription |
| View gym class packages | Y | Y | `GET /api/gyms/{gym_id}/packages` | Auth required, no subscription |
| View package details | Y | Y | `GET /api/gyms/{gym_id}/packages/{id}` | Auth required, no subscription |

---

## Credits (User-Facing)

| Action | Athlete | Admin | API Endpoint | Notes |
|--------|---------|-------|-------------|-------|
| View own credits (by gym) | Y | Y | `GET /api/gyms/{gym_id}/users/me/credits` | Auth required, no subscription |
| View available credits (by gym) | Y | Y | `GET /api/gyms/{gym_id}/users/me/credits/available` | Auth required, no subscription |
| View own notifications (phase4) | Y | Y | `GET /api/users/me/notifications` | No subscription required |

---

## Coach Portal

| Action | Athlete | Coach | Admin | API Endpoint | Notes |
|--------|---------|-------|-------|-------------|-------|
| View assigned coaching sessions | N | Y | Y | `GET /api/coaches/me/sessions` | Requires coach or admin role |
| View session roster | N | Y | Y | `GET /api/coaches/sessions/{id}/roster` | Coach must be assigned to session's org |
| Check in athlete | N | Y | Y | `POST /api/coaches/sessions/{id}/check-in/{res_id}` | Coach must be assigned to session's org |
| Mark no-show | N | Y | Y | `POST /api/coaches/sessions/{id}/no-show/{res_id}` | Coach must be assigned to session's org |
| Complete session | N | Y | Y | `POST /api/coaches/sessions/{id}/complete` | Coach must be assigned to session's org |

**Note:** Coach routes require the `coach` or `admin` role (enforced by `CoachOrAdmin` middleware). Additionally, the service layer verifies that a coach is assigned to the session's organization via `coach_assignments` before allowing roster/check-in actions. Admins bypass the org-assignment check.

---

## Data Import/Export

| Action | Athlete | Admin | API Endpoint | Notes |
|--------|---------|-------|-------------|-------|
| Export own WODs | S | Y | `GET /api/export/wods` | CSV/JSON format |
| Export own movements | S | Y | `GET /api/export/movements` | CSV/JSON format |
| Export own workouts | S | Y | `GET /api/export/user-workouts` | CSV/JSON format |
| Preview WOD import | S | Y | `POST /api/import/wods/preview` | |
| Confirm WOD import | S | Y | `POST /api/import/wods/confirm` | Into own account |
| Preview movement import | S | Y | `POST /api/import/movements/preview` | |
| Confirm movement import | S | Y | `POST /api/import/movements/confirm` | Into own account |
| Preview workout import | S | Y | `POST /api/import/user-workouts/preview` | |
| Confirm workout import | S | Y | `POST /api/import/user-workouts/confirm` | Into own account |
| Preview Wodify import | S | Y | `POST /api/import/wodify/preview` | Third-party format |
| Confirm Wodify import | S | Y | `POST /api/import/wodify/confirm` | Into own account |
| Export all users' data | N | Y | `GET /api/admin/user-management/export` | Admin bulk export |
| Import users in bulk | N | Y | `POST /api/admin/user-management/import/*` | Admin bulk import |

---

## User Management (Admin Only)

| Action | Athlete | Admin | API Endpoint | Notes |
|--------|---------|-------|-------------|-------|
| List all users | N | Y | `GET /api/admin/users` | |
| View any user's details | N | Y | `GET /api/admin/users/{id}` | |
| Unlock locked user account | N | Y | `POST /api/admin/users/{id}/unlock` | After failed login attempts |
| Disable user account | N | Y | `POST /api/admin/users/{id}/disable` | Prevents login |
| Enable user account | N | Y | `POST /api/admin/users/{id}/enable` | Restores access |
| Change user role | N | Y | `PUT /api/admin/users/{id}/role` | Promote/demote |
| Toggle email verification status | N | Y | `POST /api/admin/users/{id}/toggle-email-verification` | Manual verification |
| Delete user account | N | Y | `DELETE /api/admin/users/{id}` | Permanent deletion |
| Preview user import | N | Y | `POST /api/admin/user-management/import/preview` | Bulk user creation |
| Confirm user import | N | Y | `POST /api/admin/user-management/import/confirm` | |
| Export users list | N | Y | `GET /api/admin/user-management/export` | |
| Filter users by criteria | N | Y | `GET /api/admin/user-management/filter` | |
| Send batch password reset emails | N | Y | `POST /api/admin/user-management/batch-password-reset` | |

---

## Organization Management (Admin Only)

| Action | Athlete | Admin | API Endpoint | Notes |
|--------|---------|-------|-------------|-------|
| Create organization | N | Y | `POST /api/admin/organizations` | |
| List organizations | N | Y | `GET /api/admin/organizations` | |
| View organization details | N | Y | `GET /api/admin/organizations/{id}` | |
| Update organization | N | Y | `PUT /api/admin/organizations/{id}` | |
| Delete organization | N | Y | `DELETE /api/admin/organizations/{id}` | |
| Assign user to organization | N | Y | `POST /api/admin/users/{id}/organization` | |
| Remove user from organization | N | Y | `DELETE /api/admin/users/{id}/organization/{org_id}` | |
| View user's organizations | N | Y | `GET /api/admin/users/{id}/organizations` | |
| View organization's users | N | Y | `GET /api/admin/organizations/{id}/users` | |

---

## Subscription Management (Admin Only)

| Action | Athlete | Admin | API Endpoint | Notes |
|--------|---------|-------|-------------|-------|
| List all user subscriptions | N | Y | `GET /api/admin/subscriptions/users` | |
| List all organization subscriptions | N | Y | `GET /api/admin/subscriptions/organizations` | |
| List expiring user subscriptions | N | Y | `GET /api/admin/subscriptions/users/expiring` | |
| List expired user subscriptions | N | Y | `GET /api/admin/subscriptions/users/expired` | |
| List expiring organization subscriptions | N | Y | `GET /api/admin/subscriptions/organizations/expiring` | |
| List expired organization subscriptions | N | Y | `GET /api/admin/subscriptions/organizations/expired` | |
| Create user subscription | N | Y | `POST /api/admin/subscriptions/user` | |
| View user's subscriptions | N | Y | `GET /api/admin/subscriptions/user/{user_id}` | |
| Mark subscription as paid | N | Y | `POST /api/admin/subscriptions/user/{id}/mark-paid` | |
| Cancel user subscription | N | Y | `POST /api/admin/subscriptions/user/{id}/cancel` | |
| Set subscription as permanent | N | Y | `POST /api/admin/subscriptions/user/{id}/set-permanent` | Never expires |
| Create organization subscription | N | Y | `POST /api/admin/subscriptions/organization` | |
| View organization's subscriptions | N | Y | `GET /api/admin/subscriptions/organization/{org_id}` | |
| Mark organization subscription paid | N | Y | `POST /api/admin/subscriptions/organization/{id}/mark-paid` | |
| Cancel organization subscription | N | Y | `POST /api/admin/subscriptions/organization/{id}/cancel` | |
| Set organization subscription permanent | N | Y | `POST /api/admin/subscriptions/organization/{id}/set-permanent` | |

---

## Class Scheduling Management (Admin Only)

### Location Management

| Action | Athlete | Admin | API Endpoint | Notes |
|--------|---------|-------|-------------|-------|
| Create gym location | N | Y | `POST /api/admin/gyms/{gym_id}/locations` | |
| Update gym location | N | Y | `PUT /api/admin/gyms/{gym_id}/locations/{id}` | |
| Delete gym location | N | Y | `DELETE /api/admin/gyms/{gym_id}/locations/{id}` | |

### Class Template Management

| Action | Athlete | Admin | API Endpoint | Notes |
|--------|---------|-------|-------------|-------|
| Create class template | N | Y | `POST /api/admin/gyms/{gym_id}/templates` | |
| View class template | N | Y | `GET /api/admin/scheduling/templates/{id}` | |
| Update class template | N | Y | `PUT /api/admin/scheduling/templates/{id}` | |
| Delete class template | N | Y | `DELETE /api/admin/scheduling/templates/{id}` | Supports `?mode=with_all_sessions` or `?mode=with_future_sessions` |
| Preview template schedule | N | Y | `GET /api/admin/scheduling/templates/{id}/preview-schedule` | |

### Schedule Slot Management

| Action | Athlete | Admin | API Endpoint | Notes |
|--------|---------|-------|-------------|-------|
| Create schedule slot | N | Y | `POST /api/admin/scheduling/templates/{id}/slots/` | Recurring time pattern |
| List schedule slots | N | Y | `GET /api/admin/scheduling/templates/{id}/slots/` | |
| Update schedule slot | N | Y | `PUT /api/admin/scheduling/templates/{id}/slots/{slot_id}` | |
| Delete schedule slot | N | Y | `DELETE /api/admin/scheduling/templates/{id}/slots/{slot_id}` | |

### Template Coach Management

| Action | Athlete | Admin | API Endpoint | Notes |
|--------|---------|-------|-------------|-------|
| View template coaches | N | Y | `GET /api/admin/scheduling/templates/{id}/coaches` | Default coaches |
| Add template coach | N | Y | `POST /api/admin/scheduling/templates/{id}/coaches` | |
| Remove template coach | N | Y | `DELETE /api/admin/scheduling/templates/{id}/coaches/{user_id}` | |

### Session Management

| Action | Athlete | Admin | API Endpoint | Notes |
|--------|---------|-------|-------------|-------|
| Create session manually | N | Y | `POST /api/admin/gyms/{gym_id}/sessions` | |
| Update session | N | Y | `PUT /api/admin/gyms/{gym_id}/sessions/{id}` | |
| Cancel session | N | Y | `POST /api/admin/gyms/{gym_id}/sessions/{id}/cancel` | |
| Batch update session workout | N | Y | `PUT /api/admin/gyms/{gym_id}/sessions/batch-workout` | |

### Session Roster & Check-In

| Action | Athlete | Admin | API Endpoint | Notes |
|--------|---------|-------|-------------|-------|
| View session roster | N | Y | `GET /api/admin/sessions/{session_id}/roster` | |
| Check in reservation | N | Y | `POST /api/admin/sessions/{session_id}/check-in/{reservation_id}` | |
| Mark no-show | N | Y | `POST /api/admin/sessions/{session_id}/no-show/{reservation_id}` | |
| Complete session | N | Y | `POST /api/admin/sessions/{session_id}/complete` | |

### Session Coach Management

| Action | Athlete | Admin | API Endpoint | Notes |
|--------|---------|-------|-------------|-------|
| View session coaches | N | Y | `GET /api/admin/sessions/{session_id}/coaches` | |
| Add session coach | N | Y | `POST /api/admin/sessions/{session_id}/coaches` | |
| Remove session coach | N | Y | `DELETE /api/admin/sessions/{session_id}/coaches/{user_id}` | |

### Coach Assignment

| Action | Athlete | Admin | API Endpoint | Notes |
|--------|---------|-------|-------------|-------|
| Assign coach to gym | N | Y | `POST /api/admin/gyms/{gym_id}/coaches` | |
| List gym coaches | N | Y | `GET /api/admin/gyms/{gym_id}/coaches` | |
| Unassign coach from gym | N | Y | `DELETE /api/admin/gyms/{gym_id}/coaches/{id}` | |

---

## Document & Package Management (Admin Only)

| Action | Athlete | Admin | API Endpoint | Notes |
|--------|---------|-------|-------------|-------|
| Create document | N | Y | `POST /api/admin/gyms/{gym_id}/documents` | |
| Update document | N | Y | `PUT /api/admin/gyms/{gym_id}/documents/{id}` | |
| Delete document | N | Y | `DELETE /api/admin/gyms/{gym_id}/documents/{id}` | |
| Create class package | N | Y | `POST /api/admin/gyms/{gym_id}/packages` | |
| Update class package | N | Y | `PUT /api/admin/gyms/{gym_id}/packages/{id}` | |
| Delete class package | N | Y | `DELETE /api/admin/gyms/{gym_id}/packages/{id}` | |
| Mark user document completed | N | Y | `POST /api/admin/gyms/{gym_id}/user-documents/{id}/complete` | |
| Initialize user documents | N | Y | `POST /api/admin/gyms/{gym_id}/users/{user_id}/documents/init` | |
| Purchase credits for user | N | Y | `POST /api/admin/gyms/{gym_id}/users/{user_id}/credits` | |

---

## User-Generated Content Management (Admin Only)

| Action | Athlete | Admin | API Endpoint | Notes |
|--------|---------|-------|-------------|-------|
| List all user-created WODs | N | Y | `GET /api/admin/user-created/wods` | Across all users |
| Copy user WOD to standard library | N | Y | `POST /api/admin/user-created/wods/{id}/copy-to-standard` | Promotes to standard |
| List all user-created movements | N | Y | `GET /api/admin/user-created/movements` | Across all users |
| Copy user movement to standard library | N | Y | `POST /api/admin/user-created/movements/{id}/copy-to-standard` | Promotes to standard |
| List all user-created workouts | N | Y | `GET /api/admin/user-created/workouts` | Across all users |
| Copy user workout to standard library | N | Y | `POST /api/admin/user-created/workouts/{id}/copy-to-standard` | Promotes to standard |

---

## Backup Management (Admin Only)

| Action | Athlete | Admin | API Endpoint | Notes |
|--------|---------|-------|-------------|-------|
| Create database backup | N | Y | `POST /api/admin/backups` | Full DB backup |
| List available backups | N | Y | `GET /api/admin/backups` | |
| Upload backup file | N | Y | `POST /api/admin/backups/upload` | |
| Download backup | N | Y | `GET /api/admin/backups/{filename}` | |
| View backup metadata | N | Y | `GET /api/admin/backups/{filename}/metadata` | |
| Delete backup | N | Y | `DELETE /api/admin/backups/{filename}` | |
| Restore from backup | N | Y | `POST /api/admin/backups/{filename}/restore` | Replaces current data |

---

## Email Management (Admin Only)

| Action | Athlete | Admin | API Endpoint | Notes |
|--------|---------|-------|-------------|-------|
| View email configuration | N | Y | `GET /api/admin/email/config` | SMTP settings |
| Send test email | N | Y | `POST /api/admin/email/test` | Verify configuration |
| List email logs | N | Y | `GET /api/admin/email-logs/` | Delivery history |
| View email statistics | N | Y | `GET /api/admin/email-logs/stats` | Success/failure rates |
| View recent email failures | N | Y | `GET /api/admin/email-logs/failures` | |
| Cleanup old email logs | N | Y | `POST /api/admin/email-logs/cleanup` | |
| View email log details | N | Y | `GET /api/admin/email-logs/{id}` | |

---

## Data Quality & Cleanup (Admin Only)

| Action | Athlete | Admin | API Endpoint | Notes |
|--------|---------|-------|-------------|-------|
| Detect WOD score type mismatches | N | Y | `GET /api/admin/data-cleanup/wod-mismatches` | Data integrity check |
| Fix WOD score type mismatches | N | Y | `DELETE /api/admin/data-cleanup/wod-mismatches` | Bulk correction |
| Update WOD record directly | N | Y | `PUT /api/admin/data-cleanup/wod-record/{id}` | Manual fix |
| Run full data quality scan | N | Y | `GET /api/admin/data-quality/full-scan` | Comprehensive check |
| Scan for duplicate records | N | Y | `GET /api/admin/data-quality/duplicates` | All entity types |
| View duplicate summary | N | Y | `GET /api/admin/data-quality/duplicates/summary` | |
| Scan duplicates by entity type | N | Y | `GET /api/admin/data-quality/duplicates/{entity}` | Movements, WODs, etc. |
| Preview duplicate merge | N | Y | `POST /api/admin/data-quality/duplicates/merge/preview` | Dry run |
| Confirm duplicate merge | N | Y | `POST /api/admin/data-quality/duplicates/merge/confirm` | Execute merge |
| View data quality issues | N | Y | `GET /api/admin/data-quality/issues` | |

---

## Audit & Logging (Admin Only)

| Action | Athlete | Admin | API Endpoint | Notes |
|--------|---------|-------|-------------|-------|
| List all audit logs | N | Y | `GET /api/admin/audit-logs` | System-wide |
| View audit log details | N | Y | `GET /api/admin/audit-logs/{id}` | |
| Cleanup old audit logs | N | Y | `POST /api/admin/audit-logs/cleanup` | Retention policy |
| List data change logs | N | Y | `GET /api/admin/data-change-logs` | Entity modifications |
| View data change log details | N | Y | `GET /api/admin/data-change-logs/{id}` | |
| View entity modification history | N | Y | `GET /api/admin/data-change-logs/entity/{entity_type}/{entity_id}` | By entity type/ID |
| Cleanup old data change logs | N | Y | `POST /api/admin/data-change-logs/cleanup` | |

---

## System Administration (Admin Only)

| Action | Athlete | Admin | API Endpoint | Notes |
|--------|---------|-------|-------------|-------|
| View admin dashboard metrics | N | Y | `GET /api/admin/metrics` | System overview |
| Run benchmark test | S | Y | `POST /api/benchmark` | Performance test |
| View benchmark status | S | Y | `GET /api/benchmark/status` | |
| Cleanup benchmark data | N | Y | `DELETE /api/admin/benchmark/data` | |

---

## Public System Endpoints

| Action | API Endpoint | Notes |
|--------|-------------|-------|
| Health check | `GET /health` | Returns `{"status":"healthy","version":"..."}` |
| Version info | `GET /api/version` | Returns version, build number, scheduling status |
| API documentation | `GET /api/docs/*` | Swagger UI |

---

## UI Screen Access

### Screens Available to All Authenticated Users

| Screen | Route | Purpose |
|--------|-------|---------|
| Dashboard | `/dashboard` | Main home screen |
| Workouts List | `/workouts` | View workout history |
| Log Workout | `/workouts/log` | Record new workout |
| Create Template | `/workouts/templates/create` | Create workout template |
| Edit Template | `/workouts/templates/:id/edit` | Modify template |
| Workout Detail | `/workouts/:id` | View workout details |
| Workout Calendar | `/workouts/calendar` | Calendar view |
| Movements Library | `/movements` | Browse movements |
| Create Movement | `/movements/create` | Add custom movement |
| Edit Movement | `/movements/:id/edit` | Modify movement |
| Movement Detail | `/movements/:id` | View movement history |
| WOD Library | `/wods` | Browse WODs |
| Create WOD | `/wods/create` | Add custom WOD |
| Edit WOD | `/wods/:id/edit` | Modify WOD |
| WOD Detail | `/wods/:id` | View WOD details |
| Performance | `/performance` | Analytics dashboard |
| PR History | `/prs` | Personal records |
| Profile | `/profile` | User profile |
| Settings | `/settings` | User preferences |
| Notifications | `/notifications` | View notifications |
| Export Data | `/settings/export` | Export user data |
| Import Data | `/settings/import` | Import user data |
| Class Schedule | `/schedule/:gym_id?` | Browse and reserve classes |
| My Reservations | `/reservations` | View upcoming reservations |
| My Credits | `/my-credits` | View class credits |
| Coach Dashboard | `/coach` | View assigned coaching sessions |

### Screens Available to Admins Only

| Screen | Route | Purpose |
|--------|-------|---------|
| Admin Metrics | `/admin/metrics` | System metrics dashboard |
| Admin Users | `/admin/users` | User management |
| Admin User Import/Export | `/admin/users/import-export` | Bulk user operations |
| Admin User Content | `/admin/user-content` | User-generated content |
| Admin Organizations | `/admin/organizations` | Organization management |
| Admin Organization Detail | `/admin/organizations/:id` | Single organization view |
| Admin Subscriptions | `/admin/subscriptions` | Subscription management |
| Admin Scheduling | `/admin/scheduling` | Class templates, sessions, coaches |
| Admin Packages | `/admin/packages` | Class package management |
| Admin Backups | `/admin/backups` | Backup management |
| Admin Audit Logs | `/admin/audit-logs` | System audit logs |
| Admin Data Change Logs | `/admin/data-change-logs` | Entity change history |
| Admin Data Quality | `/admin/data-quality` | Data quality monitoring |
| Admin Data Cleanup | `/admin/data-cleanup` | Fix data issues |
| Admin Announcements | `/admin/announcements` | System announcements |
| Admin Email Settings | `/admin/email-settings` | Email configuration |
| Admin Email Logs | `/admin/email-logs` | Email delivery logs |

---

## API Endpoint Summary

| Category | Total | Public | Auth (No Sub) | Subscription | Coach/Admin | Admin Only |
|----------|-------|--------|---------------|--------------|-------------|------------|
| System (health/version/docs) | 3 | 3 | 0 | 0 | 0 | 0 |
| Authentication | 8 | 8 | 0 | 0 | 0 | 0 |
| Browse (movements/WODs/templates) | 9 | 9 | 0 | 0 | 0 | 0 |
| Account/Profile/Settings | 8 | 0 | 8 | 0 | 0 | 0 |
| Sessions (auth) | 3 | 0 | 3 | 0 | 0 | 0 |
| Subscription Status | 1 | 0 | 1 | 0 | 0 | 0 |
| Notifications | 9 | 0 | 7 | 2 | 0 | 0 |
| User Scheduling Data | 4 | 0 | 4 | 0 | 0 | 0 |
| Coach Portal | 5 | 0 | 0 | 0 | 5 | 0 |
| Gym Browsing | 13 | 0 | 13 | 0 | 0 | 0 |
| Waitlist (read) | 2 | 0 | 2 | 0 | 0 | 0 |
| Movement Management | 3 | 0 | 0 | 3 | 0 | 0 |
| WOD Management | 4 | 0 | 0 | 4 | 0 | 0 |
| Template Management | 4 | 0 | 0 | 4 | 0 | 0 |
| Template WOD Linking | 5 | 0 | 0 | 5 | 0 | 0 |
| Workout Logging | 9 | 0 | 0 | 9 | 0 | 0 |
| PR Tracking | 3 | 0 | 0 | 3 | 0 | 0 |
| Performance | 3 | 0 | 0 | 3 | 0 | 0 |
| Statistics | 1 | 0 | 0 | 1 | 0 | 0 |
| Data Export | 3 | 0 | 0 | 3 | 0 | 0 |
| Data Import | 8 | 0 | 0 | 8 | 0 | 0 |
| Benchmark | 2 | 0 | 0 | 2 | 0 | 0 |
| Reservations | 2 | 0 | 0 | 2 | 0 | 0 |
| Waitlist (actions) | 2 | 0 | 0 | 2 | 0 | 0 |
| Admin User Management | 8 | 0 | 0 | 0 | 0 | 8 |
| Admin User Import/Export | 5 | 0 | 0 | 0 | 0 | 5 |
| Admin Organizations | 9 | 0 | 0 | 0 | 0 | 9 |
| Admin Subscriptions | 16 | 0 | 0 | 0 | 0 | 16 |
| Admin Scheduling (locations) | 3 | 0 | 0 | 0 | 0 | 3 |
| Admin Scheduling (templates) | 5 | 0 | 0 | 0 | 0 | 5 |
| Admin Scheduling (slots) | 4 | 0 | 0 | 0 | 0 | 4 |
| Admin Scheduling (template coaches) | 3 | 0 | 0 | 0 | 0 | 3 |
| Admin Scheduling (sessions) | 4 | 0 | 0 | 0 | 0 | 4 |
| Admin Roster & Check-In | 4 | 0 | 0 | 0 | 0 | 4 |
| Admin Session Coaches | 3 | 0 | 0 | 0 | 0 | 3 |
| Admin Coach Assignment | 3 | 0 | 0 | 0 | 0 | 3 |
| Admin Documents & Packages | 9 | 0 | 0 | 0 | 0 | 9 |
| Admin User Content | 6 | 0 | 0 | 0 | 0 | 6 |
| Admin Backups | 7 | 0 | 0 | 0 | 0 | 7 |
| Admin Email | 7 | 0 | 0 | 0 | 0 | 7 |
| Admin Data Quality | 10 | 0 | 0 | 0 | 0 | 10 |
| Admin Audit/Logs | 7 | 0 | 0 | 0 | 0 | 7 |
| Admin Metrics | 1 | 0 | 0 | 0 | 0 | 1 |
| Admin Benchmark Cleanup | 1 | 0 | 0 | 0 | 0 | 1 |
| Admin Announcements | 1 | 0 | 0 | 0 | 0 | 1 |
| **TOTAL** | **230** | **20** | **38** | **51** | **5** | **116** |

**Note:** "Coach/Admin" endpoints require `CoachOrAdmin` middleware — accessible to users with `coach` or `admin` role, with no subscription check. Coaches are additionally verified against org-level assignments at the service layer.

---

## Rate Limiting

| Endpoint Group | Limit | Window |
|---------------|-------|--------|
| Login / Register / Resend Verification | 5 requests | Per 15 minutes per IP |
| Forgot Password / Reset Password | 3 requests | Per hour per IP |

---

## Subscription Enforcement Behavior

| HTTP Method | Active Subscription | Expired Subscription | Coach | Admin |
|-------------|-------------------|---------------------|-------|-------|
| GET | Allowed | Allowed (read-only) | Always allowed | Always allowed |
| POST | Allowed | HTTP 402 Payment Required | Always allowed | Always allowed |
| PUT | Allowed | HTTP 402 Payment Required | Always allowed | Always allowed |
| DELETE | Allowed | HTTP 402 Payment Required | Always allowed | Always allowed |

**Note:** Account routes (profile, settings, sessions, notifications) are exempt from subscription checks. Only feature routes (workouts, movements, WODs, etc.) enforce subscription requirements. Coaches and admins bypass all subscription enforcement.

---

## Implementation Notes

### Middleware Stack

**Public endpoints:** No middleware (or rate limiting only)

**Authenticated endpoints:**
```
middleware.Auth(jwtSecret) // Extracts user ID, email, role from JWT
```

**Subscription-gated endpoints:**
```
middleware.Auth(jwtSecret)
middleware.RequireActiveSubscription(subscriptionService)
// GET requests pass through; POST/PUT/DELETE blocked for expired users
```

**Coach/admin endpoints:**
```
middleware.Auth(jwtSecret)
middleware.CoachOrAdmin // Checks role == "coach" or "admin"
// Service layer additionally verifies coach org assignment
```

**Admin-only endpoints:**
```
middleware.Auth(jwtSecret)
middleware.AdminOnly // Checks role == "admin"
```

### Key Source Files

| Purpose | Location |
|---------|----------|
| Route definitions | `cmd/actalog/main.go:594-995` |
| Auth middleware | `pkg/middleware/auth.go` |
| Subscription middleware | `pkg/middleware/subscription.go` |
| Rate limiter | `pkg/middleware/rate_limit.go` |
| Frontend router | `web/src/router/index.js` |

---

## Future Considerations

- **Gym Owner** - Multi-gym management capabilities with delegated admin permissions
