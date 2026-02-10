# User Permissions Matrix

**Last Updated:** 2026-02-09
**Version:** 1.1.0-beta

This document details all user actions and their permission requirements in ActaLog.

---

## Role Overview

| Role | Description | Assignment |
|------|-------------|------------|
| **User** | Regular user with access to personal workout tracking features | Default role for all new registrations |
| **Admin** | Full system access including user management and system configuration | First registered user; manually assigned thereafter |

**Notes:**
- A user has exactly one role at any time
- Role changes require admin action via Admin > Users
- Subscription status affects feature access for regular users (admins bypass subscription checks)

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

| Action | User | Admin | Notes |
|--------|------|-------|-------|
| Register new account | Y | Y | Public action, no auth required |
| Login | Y | Y | Public action |
| Logout | Y | Y | |
| Request password reset | Y | Y | Public action |
| Reset password with token | Y | Y | Public action |
| Verify email address | Y | Y | Via email link |
| Resend verification email | Y | Y | |
| Refresh JWT token | Y | Y | |
| Revoke refresh token | Y | Y | |
| View own profile | Y | Y | |
| Edit own profile (name, email, birthday) | Y | Y | |
| Upload own avatar | Y | Y | |
| Delete own avatar | Y | Y | |
| Change own password | Y | Y | |
| View own settings | Y | Y | |
| Update own settings | Y | Y | |
| View own active sessions | Y | Y | |
| Revoke specific session | Y | Y | Own sessions only |
| Revoke all own sessions | Y | Y | |
| View own audit logs | Y | Y | |
| Check own subscription status | Y | Y | |

---

## Movements Library

| Action | User | Admin | Notes |
|--------|------|-------|-------|
| View all movements | Y | Y | Includes standard and custom |
| Search movements | Y | Y | |
| View movement details | Y | Y | |
| Create custom movement | S | Y | User-owned, subscription required |
| Edit own custom movement | S | Y | Own movements only |
| Delete own custom movement | S | Y | Own movements only |
| Edit standard movement | N | Y | |
| Delete standard movement | N | Y | |
| Toggle PR tracking on movement | S | Y | |
| View movement performance history | S | Y | Own history only |

---

## WODs (Workouts of the Day)

| Action | User | Admin | Notes |
|--------|------|-------|-------|
| View all WODs | Y | Y | Includes standard and custom |
| View standard WODs | Y | Y | |
| Search WODs | Y | Y | |
| View WOD details | Y | Y | |
| View own custom WODs | S | Y | |
| Create custom WOD | S | Y | User-owned |
| Edit own custom WOD | S | Y | Own WODs only |
| Delete own custom WOD | S | Y | Own WODs only |
| Edit standard WOD | N | Y | |
| Delete standard WOD | N | Y | |
| View WOD performance history | S | Y | Own history only |

---

## Workout Templates

| Action | User | Admin | Notes |
|--------|------|-------|-------|
| View standard templates | Y | Y | Public templates |
| View template details | Y | Y | |
| View own templates | S | Y | |
| Create workout template | S | Y | User-owned |
| Edit own template | S | Y | Own templates only |
| Delete own template | S | Y | Own templates only |
| Add WOD to own template | S | Y | |
| Remove WOD from own template | S | Y | |
| Update WOD in own template | S | Y | |
| Toggle PR flag on template WOD | S | Y | |
| Edit standard template | N | Y | |
| Delete standard template | N | Y | |

---

## Workout Logging

| Action | User | Admin | Notes |
|--------|------|-------|-------|
| Log a workout | S | Y | Creates user workout record |
| View own workout history | S | Y | |
| View own workout details | S | Y | |
| Edit own logged workout | S | Y | |
| Delete own logged workout | S | Y | |
| View monthly workout statistics | S | Y | Own stats |
| View workout calendar | S | Y | Own calendar |
| View other users' workouts | N | Y | Admin can view all |

---

## Personal Records (PRs)

| Action | User | Admin | Notes |
|--------|------|-------|-------|
| View own personal records | S | Y | |
| View movements with PRs | S | Y | Own PR movements |
| Flag retroactive PRs | S | Y | Recalculate historical PRs |
| View PR history | S | Y | Own history |
| View other users' PRs | N | Y | Admin can view all |

---

## Performance Analytics

| Action | User | Admin | Notes |
|--------|------|-------|-------|
| Search own performance data | S | Y | Unified search |
| View movement performance trends | S | Y | Own data |
| View WOD performance trends | S | Y | Own data |
| View active users statistics | S | Y | Anonymized stats |

---

## Notifications

| Action | User | Admin | Notes |
|--------|------|-------|-------|
| View own notifications | Y | Y | No subscription required |
| View unread notifications | Y | Y | |
| Get unread notification count | Y | Y | |
| Mark notification as read | Y | Y | |
| Mark all notifications as read | Y | Y | |
| Delete own notification | Y | Y | |
| Like a notification | S | Y | |
| Unlike a notification | S | Y | |
| View notification likes | Y | Y | |
| Create system announcement | N | Y | Sends to all users |

---

## Data Import/Export

| Action | User | Admin | Notes |
|--------|------|-------|-------|
| Export own WODs | S | Y | CSV/JSON format |
| Export own movements | S | Y | CSV/JSON format |
| Export own workouts | S | Y | CSV/JSON format |
| Preview WOD import | S | Y | |
| Confirm WOD import | S | Y | Into own account |
| Preview movement import | S | Y | |
| Confirm movement import | S | Y | Into own account |
| Preview workout import | S | Y | |
| Confirm workout import | S | Y | Into own account |
| Preview Wodify import | S | Y | Third-party format |
| Confirm Wodify import | S | Y | Into own account |
| Export all users' data | N | Y | Admin bulk export |
| Import users in bulk | N | Y | Admin bulk import |

---

## User Management (Admin Only)

| Action | User | Admin | Notes |
|--------|------|-------|-------|
| List all users | N | Y | |
| View any user's details | N | Y | |
| Unlock locked user account | N | Y | After failed login attempts |
| Disable user account | N | Y | Prevents login |
| Enable user account | N | Y | Restores access |
| Change user role | N | Y | Promote/demote |
| Toggle email verification status | N | Y | Manual verification |
| Delete user account | N | Y | Permanent deletion |
| Preview user import | N | Y | Bulk user creation |
| Confirm user import | N | Y | |
| Export users list | N | Y | |
| Filter users by criteria | N | Y | |
| Send batch password reset emails | N | Y | |

---

## Organization Management (Admin Only)

| Action | User | Admin | Notes |
|--------|------|-------|-------|
| Create organization | N | Y | |
| List organizations | N | Y | |
| View organization details | N | Y | |
| Update organization | N | Y | |
| Delete organization | N | Y | |
| Assign user to organization | N | Y | |
| Remove user from organization | N | Y | |
| View user's organizations | N | Y | |
| View organization's users | N | Y | |

---

## Subscription Management (Admin Only)

| Action | User | Admin | Notes |
|--------|------|-------|-------|
| List all user subscriptions | N | Y | |
| List all organization subscriptions | N | Y | |
| List expiring user subscriptions | N | Y | |
| List expired user subscriptions | N | Y | |
| List expiring organization subscriptions | N | Y | |
| List expired organization subscriptions | N | Y | |
| Create user subscription | N | Y | |
| View user's subscriptions | N | Y | |
| Mark subscription as paid | N | Y | |
| Cancel user subscription | N | Y | |
| Set subscription as permanent | N | Y | Never expires |
| Create organization subscription | N | Y | |
| View organization's subscriptions | N | Y | |
| Mark organization subscription paid | N | Y | |
| Cancel organization subscription | N | Y | |
| Set organization subscription permanent | N | Y | |

---

## User-Generated Content Management (Admin Only)

| Action | User | Admin | Notes |
|--------|------|-------|-------|
| List all user-created WODs | N | Y | Across all users |
| Copy user WOD to standard library | N | Y | Promotes to standard |
| List all user-created movements | N | Y | Across all users |
| Copy user movement to standard library | N | Y | Promotes to standard |
| List all user-created workouts | N | Y | Across all users |
| Copy user workout to standard library | N | Y | Promotes to standard |

---

## Backup Management (Admin Only)

| Action | User | Admin | Notes |
|--------|------|-------|-------|
| Create database backup | N | Y | Full DB backup |
| List available backups | N | Y | |
| Upload backup file | N | Y | |
| Download backup | N | Y | |
| View backup metadata | N | Y | |
| Delete backup | N | Y | |
| Restore from backup | N | Y | Replaces current data |

---

## Email Management (Admin Only)

| Action | User | Admin | Notes |
|--------|------|-------|-------|
| View email configuration | N | Y | SMTP settings |
| Send test email | N | Y | Verify configuration |
| List email logs | N | Y | Delivery history |
| View email statistics | N | Y | Success/failure rates |
| View recent email failures | N | Y | |
| Cleanup old email logs | N | Y | |
| View email log details | N | Y | |

---

## Data Quality & Cleanup (Admin Only)

| Action | User | Admin | Notes |
|--------|------|-------|-------|
| Detect WOD score type mismatches | N | Y | Data integrity check |
| Fix WOD score type mismatches | N | Y | Bulk correction |
| Update WOD record directly | N | Y | Manual fix |
| Run full data quality scan | N | Y | Comprehensive check |
| Scan for duplicate records | N | Y | All entity types |
| View duplicate summary | N | Y | |
| Scan duplicates by entity type | N | Y | Movements, WODs, etc. |
| Preview duplicate merge | N | Y | Dry run |
| Confirm duplicate merge | N | Y | Execute merge |
| View data quality issues | N | Y | |

---

## Audit & Logging (Admin Only)

| Action | User | Admin | Notes |
|--------|------|-------|-------|
| List all audit logs | N | Y | System-wide |
| View audit log details | N | Y | |
| Cleanup old audit logs | N | Y | Retention policy |
| List data change logs | N | Y | Entity modifications |
| View data change log details | N | Y | |
| View entity modification history | N | Y | By entity type/ID |
| Cleanup old data change logs | N | Y | |

---

## System Administration (Admin Only)

| Action | User | Admin | Notes |
|--------|------|-------|-------|
| View admin dashboard metrics | N | Y | System overview |
| Run benchmark test | S | Y | Performance test |
| View benchmark status | S | Y | |
| Cleanup benchmark data | N | Y | |

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

### Screens Available to Admins Only

| Screen | Route | Purpose |
|--------|-------|---------|
| Admin Metrics | `/admin/metrics` | System metrics dashboard |
| Admin Users | `/admin/users` | User management |
| Admin User Import/Export | `/admin/users/import-export` | Bulk user operations |
| Admin User Content | `/admin/user-content` | User-generated content |
| Admin Organizations | `/admin/organizations` | Organization management |
| Admin Subscriptions | `/admin/subscriptions` | Subscription management |
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

| Category | Total Endpoints | Public | User | Admin Only |
|----------|-----------------|--------|------|------------|
| Authentication | 12 | 10 | 2 | 0 |
| Account/Settings | 12 | 0 | 12 | 0 |
| Sessions | 3 | 0 | 3 | 0 |
| Subscriptions (User) | 1 | 0 | 1 | 0 |
| Notifications | 9 | 0 | 9 | 0 |
| Movements | 6 | 4 | 2 | 0 |
| WODs | 8 | 4 | 4 | 0 |
| Workout Templates | 9 | 2 | 7 | 0 |
| Workout Logging | 8 | 0 | 8 | 0 |
| Personal Records | 4 | 0 | 4 | 0 |
| Performance | 3 | 0 | 3 | 0 |
| Data Export | 3 | 0 | 3 | 0 |
| Data Import | 8 | 0 | 8 | 0 |
| Admin User Management | 14 | 0 | 0 | 14 |
| Admin Organizations | 9 | 0 | 0 | 9 |
| Admin Subscriptions | 16 | 0 | 0 | 16 |
| Admin User Content | 6 | 0 | 0 | 6 |
| Admin Backups | 7 | 0 | 0 | 7 |
| Admin Email | 7 | 0 | 0 | 7 |
| Admin Data Quality | 9 | 0 | 0 | 9 |
| Admin Audit/Logs | 7 | 0 | 0 | 7 |
| Admin Metrics | 1 | 0 | 0 | 1 |
| **TOTAL** | **163** | **20** | **66** | **77** |

---

## Implementation Notes

### Middleware Stack

**Public endpoints:** No middleware

**Authenticated endpoints:**
```go
middleware.Auth(jwtSecret) // Extracts user ID, email, role from JWT
```

**Subscription-gated endpoints:**
```go
middleware.Auth(jwtSecret)
middleware.RequireActiveSubscription(subscriptionService) // Checks subscription status
```

**Admin-only endpoints:**
```go
middleware.Auth(jwtSecret)
middleware.AdminOnly // Checks role == "admin"
```

### Key Source Files

| Purpose | Location |
|---------|----------|
| Route definitions | `cmd/actalog/main.go:485-776` |
| Auth middleware | `pkg/middleware/auth.go` |
| Subscription middleware | `pkg/middleware/subscription.go` |
| Frontend router | `web/src/router/index.js` |

---

## Future Considerations

The following roles are planned but not yet implemented:

- **Coach** - Can check athletes in/out of classes, view assigned sessions
- **Gym Owner** - Multi-gym management capabilities

See `screenshots/Scheduling/scheduling_v2.md` for the planned Coach role specification.
