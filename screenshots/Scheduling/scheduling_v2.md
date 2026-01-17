# Class Scheduling Feature Specification

**Version:** 2.0 (Consolidated)
**Last Updated:** 2026-01-15
**Status:** Draft - Pending Requirements Clarification

---

## Summary

Add a first-class scheduling capability that lets administrators create recurring class templates and individual class sessions, coaches manage attendance, and athletes reserve and attend classes. The feature must be toggleable by admins, respect gym boundaries, and integrate with existing workout templates and user workout records.

---

## Goals

1. Provide a clear data model separating recurring class templates from scheduled class sessions
2. Support reservations, waitlists, and capacity enforcement
3. Allow coaches to check-in/check-out attendees and record attendance results (creating user workout records)
4. Enable reporting: per-class participation, per-user attendance history, missed-class analysis, subscription/credit accounting
5. Make the feature administratively toggleable and safe to disable

## Non-Goals

- Replace full calendar systems or enterprise resource scheduling
- Manage complex multi-location resource bookings (beyond gym/location attribute and simple room/resource name)
- Payment processing (subscriptions are managed externally; we track credits only)

---

## User Roles

| Role | Capabilities |
|------|--------------|
| **Athlete** | Browse/filter classes, view required documents, reserve/cancel spots, self-check-in (if enabled), view attendance history |
| **Coach** | View assigned sessions, view roster, check users in/out, mark no-shows, add notes |
| **Admin** | All coach capabilities + manage templates/sessions, override capacity, manage subscriptions/credits, export reports, enable/disable feature |

**Constraint:** A user has exactly one role at any time.

---

## Feature Toggle

| State | Athlete/Coach View | Admin View |
|-------|-------------------|------------|
| **Enabled** | Full scheduling UI visible | Full scheduling UI |
| **Disabled** | Scheduling UI hidden | Scheduling UI with warning banner; can still manage data |

Toggle can be set globally or per-gym (TBD - see Question 1).

---

## Data Model

### Core Entities

```
┌─────────────────┐       ┌─────────────────┐
│  ClassTemplate  │──────▶│   ClassSession  │
│  (recurring)    │ 1:N   │   (instance)    │
└─────────────────┘       └─────────────────┘
                                   │
                    ┌──────────────┼──────────────┐
                    ▼              ▼              ▼
             ┌────────────┐ ┌────────────┐ ┌────────────┐
             │Reservation │ │ Attendance │ │CoachAssign │
             └────────────┘ └────────────┘ └────────────┘
```

### Entity Definitions

#### ClassTemplate
Defines a recurring class pattern.

| Field | Type | Description |
|-------|------|-------------|
| id | int64 | Primary key |
| gym_id | int64 | FK to gyms |
| title | string | Class name (e.g., "CrossFit WOD") |
| description | text | Optional description |
| location | string | Room/area name |
| default_workout_template_id | int64? | FK to workout_templates (nullable) |
| recurrence_rule | string | Recurrence pattern (see Recurrence section) |
| duration_minutes | int | Class duration |
| max_capacity | int | Maximum attendees |
| required_document_ids | []int64 | Documents user must sign |
| allowed_subscription_ids | []int64 | Which subscriptions can book (empty = any) |
| visibility | enum | public, members_only, private |
| is_active | bool | Soft delete / pause |
| created_at, updated_at | timestamp | Audit fields |

#### ClassSession
A concrete scheduled occurrence.

| Field | Type | Description |
|-------|------|-------------|
| id | int64 | Primary key |
| template_id | int64? | FK to class_templates (nullable for ad-hoc) |
| gym_id | int64 | FK to gyms |
| title | string | Inherited or overridden |
| location | string | Inherited or overridden |
| workout_template_id | int64? | Inherited or overridden |
| start_time | timestamp | Scheduled start |
| end_time | timestamp | Scheduled end |
| capacity_override | int? | Override template capacity |
| status | enum | scheduled, in_progress, completed, cancelled |
| cancellation_reason | string? | If cancelled |
| created_at, updated_at | timestamp | Audit fields |

#### Reservation
User's spot in a session.

| Field | Type | Description |
|-------|------|-------------|
| id | int64 | Primary key |
| session_id | int64 | FK to class_sessions |
| user_id | int64 | FK to users |
| state | enum | reserved, confirmed, waitlist, canceled, attended, no_show |
| waitlist_position | int? | Position if waitlisted |
| reserved_at | timestamp | When reserved |
| confirmed_at | timestamp? | When confirmed/paid |
| canceled_at | timestamp? | When canceled |
| canceled_by | int64? | User who canceled (self or admin) |
| cancellation_reason | string? | Optional reason |
| note | text? | Admin notes |
| created_at, updated_at | timestamp | Audit fields |

**Unique constraint:** (session_id, user_id) - one reservation per user per session

#### Attendance
Check-in/check-out record.

| Field | Type | Description |
|-------|------|-------------|
| id | int64 | Primary key |
| session_id | int64 | FK to class_sessions |
| user_id | int64 | FK to users |
| reservation_id | int64? | FK to reservations (nullable for walk-ins) |
| check_in_time | timestamp | When checked in |
| check_out_time | timestamp? | When checked out (optional) |
| recorded_by_user_id | int64 | Coach/admin who recorded |
| user_workout_id | int64? | FK to user_workouts (created on completion) |
| notes | text? | Coach notes |
| created_at | timestamp | Audit field |

#### CoachAssignment
Links coaches to sessions.

| Field | Type | Description |
|-------|------|-------------|
| session_id | int64 | FK to class_sessions |
| coach_user_id | int64 | FK to users (role=coach) |
| is_primary | bool | Primary coach for the session |
| created_at | timestamp | Audit field |

**Primary key:** (session_id, coach_user_id)

#### Document
Required documents (waivers, agreements).

| Field | Type | Description |
|-------|------|-------------|
| id | int64 | Primary key |
| gym_id | int64 | FK to gyms |
| title | string | Document name |
| content | text | Document content/HTML |
| version | int | Version number |
| is_required | bool | Required for class booking |
| is_active | bool | Currently in use |
| created_at, updated_at | timestamp | Audit fields |

#### UserDocument
User's signature status.

| Field | Type | Description |
|-------|------|-------------|
| id | int64 | Primary key |
| user_id | int64 | FK to users |
| document_id | int64 | FK to documents |
| document_version | int | Version signed |
| signed_at | timestamp | When signed |
| ip_address | string? | IP at signing |

**Unique constraint:** (user_id, document_id)

#### Subscription
Subscription/package definition.

| Field | Type | Description |
|-------|------|-------------|
| id | int64 | Primary key |
| gym_id | int64 | FK to gyms |
| name | string | Package name |
| description | text? | Description |
| type | enum | time_bound, credit_based |
| term | enum | weekly, monthly, yearly, indefinite |
| credit_amount | int? | Starting credits (for credit_based) |
| price | decimal? | Display price (not used for billing) |
| is_active | bool | Available for purchase |
| created_at, updated_at | timestamp | Audit fields |

#### UserSubscription
User's active subscription.

| Field | Type | Description |
|-------|------|-------------|
| id | int64 | Primary key |
| user_id | int64 | FK to users |
| subscription_id | int64 | FK to subscriptions |
| gym_id | int64 | FK to gyms |
| status | enum | active, inactive, canceled, expired |
| start_date | date | Subscription start |
| end_date | date? | Subscription end (null for indefinite) |
| credits_remaining | int? | Current balance (for credit_based) |
| credits_used | int | Total credits used |
| notes | text? | Admin notes (e.g., refund reason) |
| created_at, updated_at | timestamp | Audit fields |

#### SubscriptionUsage
Credit usage history.

| Field | Type | Description |
|-------|------|-------------|
| id | int64 | Primary key |
| user_subscription_id | int64 | FK to user_subscriptions |
| session_id | int64 | FK to class_sessions |
| credits_used | int | Credits consumed |
| used_at | timestamp | When consumed |
| reason | enum | reservation, check_in, admin_adjustment |

---

## Recurrence Rules

### Supported Patterns

| Pattern | Example | Storage |
|---------|---------|---------|
| Weekly on specific day(s) | Every Mon/Wed/Fri at 6pm | `WEEKLY;BYDAY=MO,WE,FR;TIME=18:00` |
| Daily | Every day at 9am | `DAILY;TIME=09:00` |
| Specific days of month | 1st and 15th at noon | `MONTHLY;BYMONTHDAY=1,15;TIME=12:00` |

**Note:** Full RFC 5545 RRULE support deferred (see Question 2).

### Session Materialization

- Background job runs daily to materialize sessions for next N days (configurable, default 30)
- On template edit, admin chooses: "Apply to future unmaterialized sessions" or "Apply to all existing sessions"
- Sessions can be individually edited after materialization
- Deleting a template does not delete existing sessions (orphaned sessions remain)

---

## Reservation Lifecycle

```
                    ┌─────────────┐
                    │   reserve   │
                    └──────┬──────┘
                           │
              ┌────────────┴────────────┐
              ▼                         ▼
       ┌────────────┐           ┌────────────┐
       │  reserved  │           │  waitlist  │
       └──────┬─────┘           └──────┬─────┘
              │                        │
              │ confirm                │ spot opens
              ▼                        ▼
       ┌────────────┐           ┌────────────┐
       │ confirmed  │◀──────────│  promoted  │
       └──────┬─────┘           └────────────┘
              │
    ┌─────────┴─────────┐
    ▼                   ▼
┌─────────┐      ┌───────────┐      ┌──────────┐
│attended │      │  no_show  │      │ canceled │
└─────────┘      └───────────┘      └──────────┘
```

### State Transitions

| From | To | Trigger |
|------|----|---------|
| (none) | reserved | User requests reservation, capacity available |
| (none) | waitlist | User requests reservation, capacity full |
| reserved | confirmed | Payment/credit applied (TBD timing) |
| reserved | canceled | User or admin cancels |
| waitlist | reserved | Spot opens, user promoted |
| confirmed | attended | Check-in recorded |
| confirmed | no_show | Session ends without check-in |
| confirmed | canceled | User or admin cancels |

---

## Capacity Enforcement

```sql
-- Atomic reservation with capacity check
BEGIN;
SELECT reserved_count, max_capacity FROM class_sessions WHERE id = ? FOR UPDATE;
-- If reserved_count < max_capacity:
INSERT INTO reservations (session_id, user_id, state) VALUES (?, ?, 'reserved');
UPDATE class_sessions SET reserved_count = reserved_count + 1 WHERE id = ?;
COMMIT;
```

- `reserved_count` maintained on `class_sessions` for fast capacity checks
- Unique constraint `(session_id, user_id)` prevents duplicate reservations
- Background reconciliation job periodically verifies `reserved_count` matches actual reservation count

---

## Check-in / Attendance

### Check-in Process

1. Coach/admin selects user from roster
2. System creates `Attendance` record with `check_in_time`
3. System updates `Reservation.state` to `attended`
4. If credits charged on check-in (configurable), decrement `UserSubscription.credits_remaining`

### Session Completion

1. Admin/coach marks session as `completed`
2. For each `attended` reservation:
   - Create `UserWorkout` record linked to session's workout template
   - User can later edit their own results (time, scores)
3. For each `confirmed` reservation without attendance:
   - Mark as `no_show`
   - Optionally restore credits (configurable)

---

## Subscription & Credit Accounting

### Time-Bound Subscriptions
- Unlimited classes within date range
- No credit tracking
- Expired subscriptions block new reservations

### Credit-Based (Punch Card)
- Fixed number of credits
- Each confirmed class consumes 1 credit (or configurable amount per class type)
- Zero credits blocks new reservations
- Credits do not expire separately from subscription (TBD - see Question 3)

### Multi-Gym Support
- Subscriptions are per-gym
- User may have subscriptions at multiple gyms
- Reporting can aggregate across gyms for admin dashboards

---

## API Endpoints

### Templates
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/gyms/{gym_id}/templates` | List templates |
| POST | `/api/gyms/{gym_id}/templates` | Create template |
| GET | `/api/gyms/{gym_id}/templates/{id}` | Get template |
| PUT | `/api/gyms/{gym_id}/templates/{id}` | Update template |
| DELETE | `/api/gyms/{gym_id}/templates/{id}` | Deactivate template |

### Sessions
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/gyms/{gym_id}/sessions` | List sessions (filter by date range) |
| POST | `/api/gyms/{gym_id}/sessions` | Create ad-hoc session |
| GET | `/api/sessions/{id}` | Get session with roster |
| PUT | `/api/sessions/{id}` | Update session |
| POST | `/api/sessions/{id}/cancel` | Cancel session |
| POST | `/api/sessions/{id}/complete` | Mark completed |

### Reservations
| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/sessions/{id}/reserve` | Reserve spot |
| POST | `/api/sessions/{id}/cancel-reservation` | Cancel reservation |
| GET | `/api/sessions/{id}/roster` | Get reservation list |
| POST | `/api/sessions/{id}/checkin` | Check in user |
| POST | `/api/sessions/{id}/checkout` | Check out user |

### User
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/users/{id}/classes` | User's upcoming reservations |
| GET | `/api/users/{id}/attendance` | Attendance history |
| GET | `/api/users/{id}/subscriptions` | User's subscriptions |

### Admin
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/admin/reports/attendance` | Attendance reports |
| GET | `/api/admin/reports/subscriptions` | Subscription usage |
| POST | `/api/admin/subscriptions/{id}/adjust-credits` | Manual credit adjustment |

---

## Notifications

### Events
| Event | Recipients | Channel |
|-------|-----------|---------|
| `reservation_confirmed` | User | Email |
| `reservation_canceled` | User | Email |
| `waitlist_promoted` | User | Email + Push |
| `session_canceled` | All reserved users | Email + Push |
| `session_reminder` | All confirmed users | Push (1hr before) |
| `no_show_recorded` | User | Email |

---

## Reporting

### Admin Dashboards
1. **Class Participation**: attendance rate, capacity utilization by template/day/time
2. **User Attendance**: per-user history, missed classes, streak tracking
3. **Subscription Usage**: credits consumed, revenue by subscription type
4. **Coach Performance**: sessions led, average attendance

### Filters
- Date range
- Gym
- Class template
- User
- Subscription type

---

## Database Indexes

```sql
-- Session queries
CREATE INDEX idx_class_sessions_gym_start ON class_sessions(gym_id, start_time);
CREATE INDEX idx_class_sessions_template ON class_sessions(template_id);
CREATE INDEX idx_class_sessions_status ON class_sessions(status);

-- Reservation queries
CREATE INDEX idx_reservations_session ON reservations(session_id, state);
CREATE INDEX idx_reservations_user ON reservations(user_id, state);

-- Attendance queries
CREATE INDEX idx_attendance_session ON attendance(session_id);
CREATE INDEX idx_attendance_user ON attendance(user_id, check_in_time);

-- Subscription queries
CREATE INDEX idx_user_subscriptions_user ON user_subscriptions(user_id, status);
CREATE INDEX idx_user_subscriptions_gym ON user_subscriptions(gym_id, status);
```

---

## Security & Privacy

- All queries filter by `gym_id` to enforce gym boundaries
- Attendance records visible only to: the user, assigned coaches, admins
- Document signatures include IP address for audit
- Export functions restricted to admin role
- Personal data handling per privacy policy

---

## Testing Strategy

| Type | Coverage |
|------|----------|
| Unit | Reservation state machine, capacity logic, credit accounting |
| Integration | Session materialization, API flows, notification triggers |
| Load | Concurrent reservation stress test (target: 100 concurrent) |
| E2E | Full booking flow from browse to check-in |

---

## Migration Plan

1. Add new tables (class_templates, class_sessions, reservations, attendance, etc.)
2. Add user role field if not present (default: athlete)
3. Add gym settings for feature toggle
4. Deploy with feature disabled
5. Admin enables and creates initial templates
6. Materialization job creates first sessions

---

## Open Questions for Clarification

See below for scenario-based questions requiring decisions.

---

## Reference Screenshots

| Figure | Description |
|--------|-------------|
| 1 | Document checklist - required signatures |
| 2 | Subscription packages - time-bound and credit-based |
| 3 | Attendance tracking with filters |
| 4 | User's class history |
| 5 | Class details and attributes |
| 6 | Athlete class list view |

