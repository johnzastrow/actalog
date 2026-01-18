# New Capability: Class Schedules

**Summary:**

Add a first-class scheduling capability for gyms that lets administrators create recurring class templates and individual class sessions, let coaches manage attendance, and let athletes reserve and attend classes. The feature must be toggleable (disable/enable scheduling) by admins, respect gym boundaries, and integrate with existing workout templates and user workout records.

**Goals:**

- Provide a clear data model that separates a recurring class template from scheduled class sessions.
- Support reservations, waitlists, and capacity enforcement.
- Allow coaches to check-in/check-out attendees and record attendance and results (creating user workout records).
- Enable reporting: per-class participation, per-user attendance history, missed-class analysis and subscription/credit accounting.
- Make the feature administratively toggleable and safe to disable.

**Non-Goals:**

- Replace full calendar systems or enterprise resource scheduling — keep scope focused on classes and gym sessions.
- Manage complex multi-location resource bookings (beyond gym/location attribute and simple room/resource name).

**High-level Requirements:**

1. Admins can enable/disable scheduling globally (or per-gym). When disabled, only admins can access scheduling UIs.
2. A `ClassTemplate` holds recurring attributes: title, gym, location, default_workout_template_id, duration_minutes, day/time or recurrence_rule, max_capacity, required_documents, subscription/package constraints.
3. A `ClassSession` is a concrete occurrence with scheduled start/end, optional overrides (workout, coach assignment), and maintains reservations and attendance lists.
4. Roles: a user has a single primary role (athlete, coach, admin) for UI actions. Coaches may be assigned to sessions (not templates) and can check in/out attendees.
5. Reservations support states: `reserved` (holds seat), `confirmed` (paid or credit applied), `canceled`, `attended`, `no_show`. Optional waitlist behavior when session is full.
6. When a `ClassSession` is marked completed, create a `UserWorkout` for attendees optionally populated with results; users may edit their own results later without affecting others.

**User Flows / UX:**

- Athlete: browse classes (filter by gym, class type, day/time), view required documents and subscription availability, reserve or cancel a reservation, check-in at class (coach or athlete), view attendance history.
- Coach: view assigned upcoming sessions, view roster and reservation statuses, check users in/out, mark no-shows, add notes.
- Admin: manage templates and sessions, override capacity, manage subscriptions/credits, export attendance reports, enable/disable feature per gym/global.

Ignore the navigation bars in the example screenshots — focus on attributes and flows.

### Screenshots for attributes and ideas

![Documents](/home/jcz/Github/actionlog/screenshots/Scheduling/image_001.png)

Figure 1 — Document checklist: users must sign required documents before confirming reservations.

![Subscription packages](/home/jcz/Github/actionlog/screenshots/Scheduling/image_002.png)

Figure 2 — Subscription/package UI: subscriptions may be time-bound or credit-based (punch card). Credits decrement on confirmed attendance.

![Attendance tracking](/home/jcz/Github/actionlog/screenshots/Scheduling/image_003.png)

Figure 3 — Attendance tracking with filters.

![User classes](/home/jcz/Github/actionlog/screenshots/Scheduling/image_004.png)

Figure 4 — Classes taken by user.

![Class details](/home/jcz/Github/actionlog/screenshots/Scheduling/image_006.png)

Figure 5 — Class details and attributes.

![LOI](/home/jcz/Github/actionlog/screenshots/Scheduling/image_007.png)

Figure 6 — (example layout/origin image)

**Data model (recommended entities):**

- `Gym` (existing)
- `User` (existing, with role)
- `WorkoutTemplate` (existing)
- `ClassTemplate` (id, gym_id, title, description, default_workout_template_id, location, recurrence_rule, duration_minutes, max_capacity, required_documents[], visibility)
- `ClassSession` (id, template_id nullable, gym_id, start_time, end_time, location, coach_user_id nullable, capacity_override nullable, status [scheduled|completed|cancelled])
- `Reservation` (id, session_id, user_id, state [reserved|confirmed|canceled|attended|no_show], reserved_at, confirmed_at, canceled_at, note)
- `Attendance` (id, session_id, user_id, check_in_time, check_out_time, recorded_by_user_id)
- `CoachAssignment` (session_id, coach_user_id)
- `Document` / `UserDocument` (document requirements and signed state)
- `Subscription` / `UserSubscription` (credit/punch tracking and constraints)

ClassTemplate vs ClassSession: templates represent recurrence and default attributes; sessions are concrete instances created by a scheduler job or admin action and can override template data (notably coach assignments and capacity overrides).

**Recurrence & Scheduling Rules:**

- Support basic recurrence: weekly on day/time, daily, or custom RRULE (RFC 5545 subset). Store recurrence_rule on `ClassTemplate` and materialize `ClassSession` instances for an upcoming window (e.g., 90 days).
- Materialization strategy: run a background job to ensure sessions exist for the next N days. Admin edits to a template should support "apply to future sessions only" or "apply to all existing instances" choices.

**Reservation lifecycle & capacity handling:**

- Reservation request attempts to create a `Reservation` row and decrement available capacity.
- Enforce capacity with transactional checks and a unique constraint on (session_id, user_id) to avoid duplicates.
- For high concurrency, use DB transactions and an integer `reserved_count` on `ClassSession` updated with optimistic locking (version) or use SQL counter with WHERE reserved_count < capacity.
- Waitlist: when full, place reservation into `waitlist` state; when a spot opens (cancellation), notify next waitlist user and allow a TTL for confirmation.

**Check-in / Attendance:**

- Coaches or admins can mark `Attendance` rows; check-in sets `attended` state for reservation and may consume credits if not already consumed.
- Late/no-show rules: if a user fails to check-in and session passes, mark `no_show` and optionally record missed-class in user history.

**Subscription & Credit accounting:**

- Support two models: time-bound subscription (unlimited classes within time window) and credit-based punch cards.
- On confirmed reservation or on check-in (configurable), decrement credits and record `UserSubscription` usage.

**APIs (examples):**

- GET /gyms/{gym_id}/classes?from=2026-01-01&to=2026-03-31 — list sessions
- POST /gyms/{gym_id}/templates — create template
- POST /sessions — create or materialize session
- POST /sessions/{id}/reserve — { user_id } → creates `Reservation`
- POST /sessions/{id}/cancel_reservation — { user_id, reason }
- POST /sessions/{id}/checkin — { user_id, recorded_by }
- GET /users/{id}/attendance — attendance history and credits used

Return patterns: include reservation state and whether user has required documents and sufficient subscription/credits.

**Notifications & Webhooks:**

- Events to emit: `reservation_created`, `reservation_canceled`, `waitlist_promoted`, `session_canceled`, `session_rescheduled`, `attendance_recorded`.
- Support email/SMS push notifications for reservation confirmations and waitlist promotions.

**Reporting / Admin dashboards:**

- Class participation by template/session (count, attendance rate, average capacity used)
- User attendance history and missed-class counts
- Subscription revenue / credit usage tied to sessions

**Concurrency & Data Integrity:**

- Use DB constraints to prevent duplicate reservations (unique(session_id, user_id)).
- Use transactional counters or SELECT ... FOR UPDATE patterns when reserving seats.
- Consider background reconciliation job to resolve transient inconsistencies (e.g., overbooked sessions).

**Testing:**

- Unit tests: reservation logic, capacity enforcement, waitlist promotion, credit accounting.
- Integration tests: materialization job creates sessions for templates, API flows for reservations and check-ins.
- Manual QA: simulate heavy concurrent reservation load to validate locking strategy.

**Migration Plan:**

1. Add new tables for `class_templates`, `class_sessions`, `reservations`, `attendances` and `user_documents` if not present.
2. Backfill minimal sessions for existing important templates if needed.
3. Deploy materialization job and set conservative horizon (e.g., 30 days) initially.

**Security & Privacy:**

- Respect gym-level boundaries for all queries; verify `gym_id` in writes.
- Protect personal data (attendance records and documents) per privacy rules; only allow exports to authorized roles.

**Monitoring & Metrics:**

- Track metrics: reservations_created, reservations_confirmed, sessions_completed, average_capacity_utilization, waitlist_promotions, checkins_per_session.

**Open Questions / Decisions Needed:**

- When should credits be charged: on reservation confirmed or on check-in? (Configurable per gym). Answer: on check-in.
- Waitlist promotion TTL: how long does a promoted user have to confirm? (e.g., 10 minutes). Answer: configurable, default 15 minutes.
- Recurrence complexity: support full RRULE or only simple weekly schedules? Answer: I need to see examples.

**Next Steps (implementation sketch):**

1. Implement DB schema and model objects.
2. Implement API endpoints and permissions checks.
3. Implement materialization background job and an admin UI to preview and edit future sessions.
4. Implement reservation UI with clear required document and subscription/credit checks.
5. Add tests for concurrency and reconciliation.

----

If you'd like, I can now:

- convert the above into a more formal PR-ready design doc, or
- implement the DB migration SQL and skeleton Go models and APIs in this repo.

Which would you prefer next?
# New Capability: Class Schedules

**Summary:** 

Add ability for administrators to create and manage class schedules for users, and that a new user type of "Coach" can check the users (athletes) into specified types of classes. 



1. The entire scheduling and class function can be hidden from athletes and coaches when disabled through the UI. Enabling it shows the screens to athletes and coaches. Admins will see a warning when it is disabled.
2. Classes will contain the same attributes as Workout Templates, but will also have a scheduled date and time, duration, maximum number of participants, and a list of users who reserved their spot in each session of the class. 
3. Users can have one role at one time (athlete, admin, coach). Only the user can change workout 
4. The coaches will be able to see a list of upcoming classes they are assigned to, and will be able to check users in and out of those classes. the system should track attendance for each class session, and generate reports for administrators on class participation rates, as well as individual user attendance history, and missed classes per user and class type and time (e.g., Yoga on Mondays at 6pm).
5. Class gym, location, title, day of week, time window, and workout should be recurring, but not the coach assignments. Editing each instance of the class then allows all those items to be adjusted.
6. 



This new feature should be able to be disabled by administrators through the UI, which would make it only visible to administrators. If disabled, they would be able to perform scheduling tasks 



Ignore the navigation bars in these screens. They are from another app to use as examples for attributes mostly.

### Example Screenshots

1. 

<img src="/home/jcz/Github/actionlog/screenshots/Scheduling/image_001.png" style="zoom:25%;" />

Figure 1. Documents

Checklist of documents list that users must sign before being able to reserve and be confirmed in classes. Documents move form unsigned to signed. Create needed attributes for document management.



<img src="/home/jcz/Github/actionlog/screenshots/Scheduling/image_002.png" style="zoom: 25%;" />

Figure 2. User Subscription packages. 

Subscription Attributes: Name, Description, Beginning and ending date (generated by system based on trigger date and Subscription Term [weekly, monthly, yearly, or indefinite]). Indefinite is also accompanied by a  "punch or credit count" starting balance that declines with every confirmed class taken. System records history or subscriptions paid for, and credits used. 

Subscription statuses: Active, Inactive, Canceled. That have active status as well. Subscriptions are per gym, so users may have subscriptions to multiple gyms with different statuses. Each subscription may expire independently. 

The user:subscription relationship should support optional notes. For example to explain that a user was refunded and subscription set to Canceled.

All classes, scheduling, attendance are per-gym. But the system may choose to aggregate them when reporting. For example, dashboards may sum classes attendance per user across multiple gyms.



A class is a combination of class attributes plus a workout (template). When a user completes a class, it creates a user workout record whose results (overall elapsed time for the workout, as well as regime scores for any associated WODS and Movements) can be edited by the user without affecting others. All user entered attributes are optional.



<img src="/home/jcz/Github/actionlog/screenshots/Scheduling/image_003.png" style="zoom: 25%;" />

Figure 3. attendance tracking for each user with filters





<img src="/home/jcz/Github/actionlog/screenshots/Scheduling/image_004.png" style="zoom: 25%;" />

figure 4. list of classes take by the user



<img src="/home/jcz/Github/actionlog/screenshots/Scheduling/image_006.png" style="zoom: 50%;" />

Figure 5. Class details. Document all the attributes you see. 



<img src="/home/jcz/Github/actionlog/screenshots/Scheduling/image_007.png" style="zoom: 50%;" />

Figure 6. Example of an athlete class list



