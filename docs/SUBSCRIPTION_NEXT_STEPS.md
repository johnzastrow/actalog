# Subscription System - Next Steps

**Date:** 2024-12-16
**Version:** 0.14.0-beta
**Status:** Backend Complete, Frontend Pending

---

## Summary

The subscription billing system backend is **100% complete** and production-ready. All backend layers (domain, repository, service, middleware, handler) are implemented, tested, and documented. The system supports:

- ✅ Dual-level billing (user-level AND organization-level)
- ✅ Three subscription types (Free, Monthly, Annual)
- ✅ Permanent Free subscriptions (never expire)
- ✅ Manual admin payment control
- ✅ Immediate read-only mode enforcement
- ✅ Complete audit trail
- ✅ Multi-database support (SQLite, PostgreSQL, MariaDB)
- ✅ Zero downtime backward compatibility

**All existing users** have been automatically seeded with permanent free subscriptions and maintain full access.

---

## What's Complete

### Backend Implementation (v0.14.0)

**Domain Layer:**
- `internal/domain/subscription.go` - Entities and repository interfaces

**Repository Layer:**
- `internal/repository/user_subscription_repository.go` - User subscription CRUD
- `internal/repository/organization_subscription_repository.go` - Organization subscription CRUD
- `internal/repository/subscription_access_repository.go` - Performance-optimized access checking

**Service Layer:**
- `internal/service/subscription_service.go` - Business logic for subscription management
- Admin operations: Create, MarkAsPaid, Cancel
- User operations: CheckAccess, GetStatus

**Middleware Layer:**
- `pkg/middleware/subscription.go` - Read-only mode enforcement
- Allows GET requests when expired (view/export/dashboard)
- Blocks POST/PUT/PATCH/DELETE when expired (returns HTTP 402)

**Handler Layer:**
- `internal/handler/subscription_handler.go` - 10 API endpoints

**Route Configuration:**
- `cmd/actalog/main.go` - All routes wired and ready

**Database:**
- Migration 0.14.0 creates subscription tables
- All existing users seeded with permanent free subscriptions
- Version snapshots created for SQLite, PostgreSQL, MariaDB

**Documentation:**
- Updated CHANGELOG.md with v0.14.0 release notes
- Updated TODO.md with completed work and next steps
- Updated ROADMAP.md with subscription system status
- Updated DATABASE_SCHEMA.md with subscription table definitions
- Created VERSION_DATABASES.md for database version management
- Created MIGRATION_TEST_0.14.0.md with test results

---

## Next Steps: Frontend Integration

### Priority 1: User Subscription Status Display (High Priority)

**Goal:** Users can view their subscription status and understand their access level.

**Tasks:**
1. **Create SubscriptionStatusBadge component** (`web/src/components/SubscriptionStatusBadge.vue`)
   - Display subscription type (Free, Monthly, Annual, Permanent Free)
   - Show expiration date for paid subscriptions
   - Show "Permanent Free" badge for permanent users
   - Show "Active" or "Expired" status with color coding
   - List organization subscriptions user benefits from

2. **Add subscription status to Settings/Profile view** (`web/src/views/SettingsView.vue`)
   - Fetch subscription status from `GET /api/subscriptions/status`
   - Display SubscriptionStatusBadge component
   - Show next billing date (if applicable)
   - Show "Renew Subscription" CTA if expired

3. **Create subscription store** (`web/src/stores/subscription.js`)
   - Store user subscription status
   - Provide computed properties: `hasAccess`, `isExpired`, `expiresAt`, `source`
   - Auto-refresh subscription status on app load
   - Expose methods: `fetchStatus()`, `isActive()`

**API Endpoint:**
- `GET /api/subscriptions/status` - Already implemented

**Estimated Effort:** 2-3 hours

---

### Priority 2: Read-Only Mode UI Feedback (High Priority)

**Goal:** Users with expired subscriptions understand why they cannot create/edit content.

**Tasks:**
1. **Handle HTTP 402 responses globally** (`web/src/utils/axios.js`)
   - Intercept 402 Payment Required responses
   - Store expired state in subscription store
   - Show subscription expired notification

2. **Create SubscriptionExpiredBanner component** (`web/src/components/SubscriptionExpiredBanner.vue`)
   - Persistent banner at top of app when subscription expired
   - Message: "Your subscription has expired. You can view and export data, but cannot create or edit content."
   - "Renew Subscription" button (links to admin/contact)
   - Dismissible but reappears on page reload

3. **Disable create/edit buttons when expired**
   - Check `subscriptionStore.hasAccess` before enabling buttons
   - Add disabled state with tooltip: "Subscription required"
   - Apply to: Create Workout, Log Workout, Edit buttons, Delete buttons

4. **Toast notifications for blocked operations**
   - When user tries to perform write operation while expired
   - Message: "Subscription expired. Please renew to create or edit content."

**Files to Modify:**
- `web/src/utils/axios.js` - Add 402 interceptor
- `web/src/App.vue` - Add SubscriptionExpiredBanner
- `web/src/views/WorkoutsView.vue` - Disable "Log Workout" button
- `web/src/views/MovementsView.vue` - Disable "Create Movement" button
- `web/src/views/WODsView.vue` - Disable "Create WOD" button
- `web/src/views/TemplatesView.vue` - Disable "Create Template" button

**Estimated Effort:** 3-4 hours

---

### Priority 3: Admin Subscription Management UI (High Priority)

**Goal:** Admins can manage user and organization subscriptions through the admin panel.

**Tasks:**
1. **Create AdminSubscriptionsView** (`web/src/views/AdminSubscriptionsView.vue`)
   - Two tabs: "User Subscriptions" and "Organization Subscriptions"
   - Data table with columns: ID, Name, Type, Status, Start Date, End Date, Actions
   - Filters: Status (All, Active, Expired, Cancelled), Type (All, Free, Monthly, Annual)
   - Search by user email or organization name
   - Actions: View Details, Mark as Paid, Cancel

2. **Create SubscriptionDetailDialog component** (`web/src/components/admin/SubscriptionDetailDialog.vue`)
   - Show full subscription details
   - Display subscription history
   - "Mark as Paid" button → opens payment confirmation dialog
   - "Cancel Subscription" button → opens cancellation dialog with reason input
   - Show audit log entries for this subscription

3. **Create CreateSubscriptionDialog component** (`web/src/components/admin/CreateSubscriptionDialog.vue`)
   - Select user OR organization (autocomplete)
   - Select subscription type (Free, Monthly, Annual)
   - Checkbox: "Permanent Free"
   - Notes field (admin notes)
   - Create button → calls `POST /api/admin/subscriptions/user` or `/organization`

4. **Create MarkAsPaidDialog component** (`web/src/components/admin/MarkAsPaidDialog.vue`)
   - Confirm payment date (defaults to today)
   - Shows calculated new end_date
   - Confirm button → calls `POST /api/admin/subscriptions/user/{id}/mark-paid`

5. **Create CancelSubscriptionDialog component** (`web/src/components/admin/CancelSubscriptionDialog.vue`)
   - Reason input (required)
   - Confirm button → calls `POST /api/admin/subscriptions/user/{id}/cancel`

6. **Add "Expiring Soon" and "Overdue" views**
   - List subscriptions expiring in next 30 days
   - List expired subscriptions that need renewal
   - Quick "Mark as Paid" action

7. **Add subscription management link to Admin menu**
   - Update `web/src/views/AdminProfileView.vue`
   - Add "Manage Subscriptions" card/link

**API Endpoints:**
- `POST /api/admin/subscriptions/user` - Create user subscription
- `GET /api/admin/subscriptions/user/{user_id}` - List user subscriptions
- `POST /api/admin/subscriptions/user/{id}/mark-paid` - Mark as paid
- `POST /api/admin/subscriptions/user/{id}/cancel` - Cancel subscription
- Organization endpoints (same pattern)

**Estimated Effort:** 8-10 hours

---

## Implementation Sequence

### Phase 1: User-Facing Features (Week 1)
1. Create subscription store (`subscription.js`)
2. Create SubscriptionStatusBadge component
3. Add subscription status to Settings view
4. Handle HTTP 402 responses in axios
5. Create SubscriptionExpiredBanner
6. Disable buttons when subscription expired

**Deliverable:** Users can view their subscription status and understand read-only mode.

### Phase 2: Admin Features (Week 2)
1. Create AdminSubscriptionsView
2. Create CreateSubscriptionDialog
3. Create SubscriptionDetailDialog
4. Create MarkAsPaidDialog
5. Create CancelSubscriptionDialog
6. Add expiring/overdue views
7. Add subscription management to admin menu

**Deliverable:** Admins can fully manage subscriptions through UI.

---

## Testing Checklist

### User Features
- [ ] User can view subscription status in Settings
- [ ] Subscription badge shows correct type (Free, Monthly, Annual, Permanent)
- [ ] Expiration date displays correctly for paid subscriptions
- [ ] "Permanent Free" badge appears for permanent users
- [ ] Organization subscriptions are listed
- [ ] Subscription expired banner appears when subscription expires
- [ ] Create/edit buttons are disabled when expired
- [ ] Tooltip explains why buttons are disabled
- [ ] Toast notification appears when trying write operation while expired
- [ ] User can view/export data while expired
- [ ] User cannot create/edit while expired

### Admin Features
- [ ] Admin can view all user subscriptions
- [ ] Admin can view all organization subscriptions
- [ ] Admin can create user subscription
- [ ] Admin can create organization subscription
- [ ] Admin can mark subscription as paid (end_date extends)
- [ ] Admin can cancel subscription (with reason)
- [ ] Admin can view subscription history
- [ ] Expiring subscriptions list shows correct items (next 30 days)
- [ ] Overdue subscriptions list shows expired items
- [ ] All subscription operations create audit log entries
- [ ] Admin cannot modify their own subscription

### Access Control
- [ ] User with active personal subscription has write access
- [ ] User with expired personal subscription but active org subscription has write access
- [ ] User with both expired has read-only access
- [ ] Read operations (GET) work when expired
- [ ] Write operations (POST/PUT/PATCH/DELETE) return HTTP 402 when expired
- [ ] Middleware correctly enforces read-only mode

---

## API Endpoints Reference

### User Endpoints (Authenticated)
```
GET /api/subscriptions/status
  Response: {
    has_access: boolean,
    source: "user" | "organization" | "both" | "none",
    user_subscription: {...} | null,
    org_subscriptions: [...],
    expires_at: "2025-01-15T00:00:00Z" | null
  }
```

### Admin Endpoints (Admin Only)

**User Subscriptions:**
```
POST /api/admin/subscriptions/user
  Body: {
    user_id: number,
    subscription_type: "free" | "monthly" | "annual",
    is_permanent_free: boolean,
    notes: string
  }

GET /api/admin/subscriptions/user/{user_id}
  Response: [UserSubscription, ...]

POST /api/admin/subscriptions/user/{id}/mark-paid
  Body: {} (uses current date)
  Effect: Updates last_payment_date, extends end_date

POST /api/admin/subscriptions/user/{id}/cancel
  Body: {reason: string}
  Effect: Sets status to 'cancelled', records reason
```

**Organization Subscriptions:** (Same structure as user subscriptions)
```
POST /api/admin/subscriptions/organization
GET /api/admin/subscriptions/organization/{org_id}
POST /api/admin/subscriptions/organization/{id}/mark-paid
POST /api/admin/subscriptions/organization/{id}/cancel
```

---

## Database Reference

### user_subscriptions Table

| Column | Type | Description |
|--------|------|-------------|
| id | BIGINT | Primary key |
| user_id | BIGINT | FK to users.id |
| subscription_type | VARCHAR(20) | 'free', 'monthly', 'annual' |
| status | VARCHAR(20) | 'active', 'expired', 'cancelled' |
| is_permanent_free | BOOLEAN | Never expires if TRUE |
| start_date | TIMESTAMP | When subscription started |
| end_date | TIMESTAMP | When expires (NULL for permanent) |
| last_payment_date | TIMESTAMP | Last payment received |
| next_billing_date | TIMESTAMP | Next billing due (NULL for free) |
| cancelled_at | TIMESTAMP | When cancelled |
| cancelled_reason | TEXT | Cancellation reason |
| notes | TEXT | Admin notes |
| created_at | TIMESTAMP | Record creation |
| updated_at | TIMESTAMP | Last update |
| created_by_user_id | BIGINT | FK to users.id (admin) |

**Access Logic:**
```
User has write access if:
  1. User has active personal subscription (is_permanent_free OR end_date > NOW())
  OR
  2. User belongs to ≥1 organization with active subscription
```

---

## Future Enhancements (Post-MVP)

These features can be added after the core frontend is complete:

1. **Stripe Integration** (v0.15.0+)
   - Replace manual admin control with automated billing webhooks
   - Credit card payment processing
   - Automatic subscription renewal

2. **Email Notifications** (v0.15.0+)
   - Notify users 7/3/1 days before expiration
   - Notify users when subscription expires
   - Notify admins of failed payments

3. **Self-Service Portal** (v0.16.0+)
   - Users can upgrade/downgrade themselves
   - View payment history
   - Update payment method

4. **Usage Limits** (v0.16.0+)
   - Track API usage for free tier
   - Enforce limits on free subscriptions
   - Display usage metrics

5. **Bulk Operations** (v0.17.0+)
   - Admin can bulk-update subscriptions
   - Bulk extend expiration dates
   - Bulk cancel subscriptions

6. **Grace Period Configuration** (v0.17.0+)
   - Make grace period configurable per organization
   - Different grace periods for different subscription tiers

---

## Resources

### Documentation
- **CHANGELOG.md** - v0.14.0 release notes
- **DATABASE_SCHEMA.md** - Subscription table definitions
- **VERSION_DATABASES.md** - Database version management guide
- **MIGRATION_TEST_0.14.0.md** - Migration test report

### Code Reference
- **Domain:** `internal/domain/subscription.go`
- **Repository:** `internal/repository/*_subscription_repository.go`
- **Service:** `internal/service/subscription_service.go`
- **Middleware:** `pkg/middleware/subscription.go`
- **Handler:** `internal/handler/subscription_handler.go`
- **Routes:** `cmd/actalog/main.go:414, 477-489`

### Database Snapshots
- **SQLite:** `db_versions/actalog_0.14.0.db` (564 KB)
- **PostgreSQL:** Schema `actalog_0_14_0` on 192.168.1.143
- **MariaDB:** Database `actalog_0_14_0` on 192.168.1.234

### Scripts
- `scripts/create-db-snapshot.sh` - Create version snapshot
- `scripts/verify-version-databases.sh` - Verify all version databases

---

## Questions?

If you need clarification on any part of the subscription system:

1. Review the comprehensive documentation in `db_versions/VERSION_DATABASES.md`
2. Check the migration test report in `db_versions/MIGRATION_TEST_0.14.0.md`
3. Review the implementation plan (available in Claude Code session history)
4. Examine the API handler code in `internal/handler/subscription_handler.go`

The backend is **production-ready** and fully tested across all three database engines.
