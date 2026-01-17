# Organization/Gym Feature Implementation Plan

**Feature:** Organization/Gym Infrastructure + Dashboard Active Users Statistics Card

**Version:** 0.6.0

**Date:** 2025-01-12

## Overview

This document tracks the implementation of organization/gym functionality in ActaLog and a dashboard statistics card showing active users within an organization for the current month.

## Requirements

### User Story
As a gym member, I want to see a dashboard card that shows:
- My own workout count for the current month (always displayed)
- 2 random users from my organization/gym with their workout counts
- Visual indication of who I am

### Technical Requirements
1. Add organization/gym entity to the system
2. Allow admins to create and manage organizations
3. Allow admins to assign users to organizations
4. Create API endpoint for active users statistics
5. Add dashboard UI component to display the statistics

## Implementation Progress

### ✅ Phase 1: Backend Infrastructure (COMPLETED)

#### 1. Domain Layer ✅
**Files Modified:**
- ✅ `internal/domain/organization.go` (NEW)
  - Organization entity (id, name, description, created_at, updated_at)
  - OrganizationRepository interface (CRUD + Count methods)

- ✅ `internal/domain/user.go`
  - Added field: `OrganizationID *int64` (nullable for backward compatibility)

#### 2. Database Migration ✅
**Files Modified:**
- ✅ `internal/repository/migrations.go`
  - Added migration version 0.6.0
  - Creates `organizations` table (SQLite/PostgreSQL/MySQL variants)
  - Adds `organization_id` column to `users` table (nullable, foreign key)

- ✅ `internal/repository/database.go`
  - Added `checkColumnExists()` helper function for migration safety

**Migration Details:**
```sql
-- Organizations table
CREATE TABLE organizations (
  id INTEGER PRIMARY KEY AUTOINCREMENT,  -- or BIGSERIAL/BIGINT AUTO_INCREMENT
  name TEXT UNIQUE NOT NULL,
  description TEXT,
  created_at DATETIME NOT NULL,
  updated_at DATETIME NOT NULL,
  INDEX idx_organizations_name (name)
);

-- Add to users table
ALTER TABLE users ADD COLUMN organization_id INTEGER REFERENCES organizations(id) ON DELETE SET NULL;
```

#### 3. Repository Layer ✅
**Files Modified:**
- ✅ `internal/repository/organization_repository.go` (NEW)
  - Implements OrganizationRepository interface
  - Handles database driver differences (SQLite vs PostgreSQL vs MySQL)
  - CRUD operations with proper error handling

- ✅ `internal/repository/user_repository.go`
  - Added organization_id to all SELECT queries (GetByID, GetByEmail, List)
  - Added organization_id handling in UPDATE query
  - Added sql.NullInt64 handling for nullable organization_id

- ✅ `internal/repository/user_workout_repository.go`
  - Added method: `GetActiveUsersThisMonth(userID, orgID int64)`
  - Query returns: Current user + 2 random users from same org with workout counts
  - Handles RANDOM() vs RAND() for different databases

#### 4. Service Layer ✅
**Files Modified:**
- ✅ `internal/service/organization_service.go` (NEW)
  - Create organization with validation
  - Update organization (checks name uniqueness)
  - Delete organization (foreign key handles user cleanup)
  - AssignUserToOrganization
  - RemoveUserFromOrganization
  - List with pagination

- ✅ `internal/service/user_workout_service.go`
  - Added method: `GetActiveUsersThisMonth(userID int64)`
  - Returns current user + 2 random org members if user has org
  - Returns only current user if no org assigned

### 🔄 Phase 2: API Layer (IN PROGRESS)

#### 5. Handler Layer ⏳
**Files to Modify:**
- ⏳ `internal/handler/organization_handler.go` (NEW)
  - POST /api/admin/organizations - Create organization (admin only)
  - GET /api/admin/organizations - List organizations (admin only)
  - GET /api/admin/organizations/:id - Get organization (admin only)
  - PUT /api/admin/organizations/:id - Update organization (admin only)
  - DELETE /api/admin/organizations/:id - Delete organization (admin only)
  - POST /api/admin/users/:id/organization - Assign user to org (admin only)
  - DELETE /api/admin/users/:id/organization - Remove user from org (admin only)

- ⏳ `internal/handler/user_workout_handler.go`
  - Add method: `GetActiveUsersStats(w, r)` for GET /api/stats/active-users-this-month
  - Returns JSON: `{"users": [{"id", "name", "workout_count", "is_current"}]}`

#### 6. Application Wiring ⏳
**Files to Modify:**
- ⏳ `cmd/actalog/main.go`
  - Initialize OrganizationRepository (~line 200)
  - Initialize OrganizationService (~line 244)
  - Initialize OrganizationHandler (~line 266)
  - Add admin routes for organizations (~line 437)
  - Add stats route: GET /api/stats/active-users-this-month (~line 362)

### 📋 Phase 3: Frontend (PENDING)

#### 7. Admin Frontend - Organization Management ⏳
**Files to Create:**
- ⏳ `web/src/views/AdminOrganizationsView.vue` (NEW)
  - Data table showing all organizations
  - Create/Edit/Delete organization dialogs
  - Pagination support
  - Error and success alerts

**Files to Modify:**
- ⏳ `web/src/router/index.js`
  - Add route: `/admin/organizations` → AdminOrganizationsView (admin only)

- ⏳ `web/src/views/AdminView.vue`
  - Add card linking to organization management

- ⏳ `web/src/views/AdminUsersView.vue`
  - Add organization column to user table
  - Add organization selector in user edit dialog
  - Implement assign/remove organization functionality

#### 8. Dashboard Statistics Card ⏳
**Files to Modify:**
- ⏳ `web/src/views/DashboardView.vue`
  - Add new card component after "This Week" card
  - Display current user (highlighted with cyan avatar)
  - Display 2 random users from same org (purple avatars)
  - Show workout counts for current month
  - Fetch data via GET /api/stats/active-users-this-month

**Expected UI:**
```vue
<!-- Active Users This Month Card -->
<v-row dense class="mb-1" v-if="activeUsersStats.length > 0">
  <v-col cols="12">
    <v-card elevation="0" rounded class="pa-2">
      <div class="text-caption mb-2 font-weight-bold">
        <v-icon size="small" color="#00bcd4">mdi-account-group</v-icon>
        Active This Month
      </div>
      <!-- User list with avatars and workout counts -->
    </v-card>
  </v-col>
</v-row>
```

### 🧪 Phase 4: Testing (PENDING)

#### Manual Testing Checklist
- [ ] Migration runs successfully on SQLite
- [ ] Migration runs successfully on PostgreSQL
- [ ] Migration runs successfully on MySQL
- [ ] Create organization as admin
- [ ] Edit organization name and description
- [ ] Delete empty organization
- [ ] Attempt to delete organization with users (should succeed, users set to NULL)
- [ ] Assign user to organization
- [ ] Remove user from organization
- [ ] Verify admin-only access to organization endpoints
- [ ] View dashboard as user without organization (shows only own stats)
- [ ] View dashboard as user with organization (shows self + 2 random users)
- [ ] Verify workout counts are accurate
- [ ] Verify current user is highlighted in dashboard card
- [ ] Test with organization having <2 users
- [ ] Test refresh functionality updates stats

#### Unit Tests to Add
- [ ] `internal/service/organization_service_test.go`
  - Test Create with valid/invalid data
  - Test Create with duplicate name
  - Test GetByID, Update, Delete
  - Test user assignment/removal

- [ ] `internal/repository/organization_repository_test.go`
  - Test CRUD for SQLite, PostgreSQL, MySQL
  - Test pagination

- [ ] `internal/repository/user_workout_repository_test.go`
  - Test GetActiveUsersThisMonth with various scenarios

#### Integration Tests to Add
- [ ] `test/integration/organization_api_test.go`
  - Test full CRUD workflow
  - Test admin-only access
  - Test user assignment workflow

## API Endpoints Added

### Admin Routes (Require Admin Role)
| Method | Route | Description |
|--------|-------|-------------|
| POST | /api/admin/organizations | Create organization |
| GET | /api/admin/organizations | List organizations |
| GET | /api/admin/organizations/:id | Get organization by ID |
| PUT | /api/admin/organizations/:id | Update organization |
| DELETE | /api/admin/organizations/:id | Delete organization |
| POST | /api/admin/users/:id/organization | Assign user to organization |
| DELETE | /api/admin/users/:id/organization | Remove user from organization |

### Authenticated Routes
| Method | Route | Description |
|--------|-------|-------------|
| GET | /api/stats/active-users-this-month | Get active users stats for current month |

### Request/Response Examples

**POST /api/admin/organizations**
```json
// Request
{
  "name": "CrossFit Downtown",
  "description": "Main downtown location"
}

// Response (201)
{
  "id": 1,
  "name": "CrossFit Downtown",
  "description": "Main downtown location",
  "created_at": "2025-01-12T10:00:00Z",
  "updated_at": "2025-01-12T10:00:00Z"
}
```

**GET /api/stats/active-users-this-month**
```json
// Response (200)
{
  "users": [
    {
      "id": 5,
      "name": "John Doe",
      "workout_count": 12,
      "is_current": true
    },
    {
      "id": 8,
      "name": "Jane Smith",
      "workout_count": 15
    },
    {
      "id": 3,
      "name": "Mike Johnson",
      "workout_count": 8
    }
  ]
}
```

## Key Design Decisions

### 1. Nullable organization_id
- **Decision:** Make `organization_id` nullable in users table
- **Rationale:** Backward compatibility for existing users
- **Trade-off:** Requires null checks throughout code
- **Alternative Rejected:** Require all users to be assigned (breaks existing installations)

### 2. Random User Selection
- **Decision:** Use database RANDOM() for selecting 2 users
- **Rationale:** Simplicity and database efficiency
- **Trade-off:** Different SQL syntax across databases (RANDOM() vs RAND())
- **Alternative Rejected:** Fetch all users and randomize in Go (less efficient)

### 3. Stats Card Data Fetching
- **Decision:** Separate API endpoint for stats
- **Rationale:** Keeps dashboard load time fast, allows independent caching
- **Trade-off:** Extra API call
- **Alternative Rejected:** Include in main dashboard data fetch (slower, couples data)

### 4. Admin-Only Organization Management
- **Decision:** All organization operations require admin role
- **Rationale:** Centralized control, prevents misuse
- **Trade-off:** Users can't create/manage their own organizations
- **Alternative Rejected:** Allow user-created organizations (more complex permissions)

### 5. Foreign Key ON DELETE SET NULL
- **Decision:** When organization is deleted, users' organization_id is set to NULL
- **Rationale:** Don't delete users when organization is removed, just unassign them
- **Trade-off:** Users lose organization affiliation
- **Alternative Rejected:** Prevent deletion if users exist (more restrictive)

## Database Schema Changes

### New Table: organizations
| Column | Type | Constraints |
|--------|------|-------------|
| id | INTEGER/BIGINT | PRIMARY KEY AUTOINCREMENT |
| name | TEXT/VARCHAR(255) | UNIQUE NOT NULL |
| description | TEXT | NULL |
| created_at | DATETIME/TIMESTAMP | NOT NULL |
| updated_at | DATETIME/TIMESTAMP | NOT NULL |

**Indexes:**
- `idx_organizations_name` on name

### Modified Table: users
| Column | Type | Constraints |
|--------|------|-------------|
| organization_id | INTEGER/BIGINT | NULL, FOREIGN KEY → organizations(id) ON DELETE SET NULL |

## File Structure Summary

### New Files Created
```
internal/
├── domain/
│   └── organization.go                      # Organization entity + repository interface
├── repository/
│   └── organization_repository.go           # Organization data access implementation
└── service/
    └── organization_service.go              # Organization business logic

web/src/views/
└── AdminOrganizationsView.vue               # Organization management UI (TODO)
```

### Modified Files
```
internal/
├── domain/
│   └── user.go                              # Added OrganizationID field
├── repository/
│   ├── migrations.go                        # Added v0.6.0 migration
│   ├── database.go                          # Added checkColumnExists helper
│   ├── user_repository.go                   # Updated all queries for organization_id
│   └── user_workout_repository.go           # Added GetActiveUsersThisMonth
├── service/
│   └── user_workout_service.go              # Added GetActiveUsersThisMonth
├── handler/
│   ├── organization_handler.go              # Organization HTTP endpoints (TODO)
│   └── user_workout_handler.go              # Stats endpoint (TODO)
└── cmd/actalog/
    └── main.go                               # Wiring (TODO)

web/src/
├── router/
│   └── index.js                              # Add org route (TODO)
└── views/
    ├── AdminView.vue                         # Add org link (TODO)
    ├── AdminUsersView.vue                    # User-org assignment (TODO)
    └── DashboardView.vue                     # Active users card (TODO)
```

## Potential Issues and Solutions

### Issue 1: SQLite Column Addition
**Problem:** SQLite has limited ALTER TABLE support
**Solution:** Use simple ADD COLUMN (supported), migration includes column existence check

### Issue 2: Database-Specific RANDOM()
**Problem:** Different random functions (RANDOM() vs RAND())
**Solution:** Use driver detection and conditional query building in repository

### Issue 3: Migration on Existing Installations
**Problem:** Existing users won't have organization_id
**Solution:** Column is nullable, no data transformation needed, users continue working normally

### Issue 4: Stats Card Performance
**Problem:** Multiple aggregations may be slow
**Solution:** Use indexed queries, limit to current month only, consider caching

## Next Steps

1. **Complete Handler Layer**
   - Create organization_handler.go
   - Add stats endpoint to user_workout_handler.go

2. **Wire Everything in main.go**
   - Initialize repositories, services, handlers
   - Add routes for organization management and stats

3. **Frontend Implementation**
   - Create AdminOrganizationsView component
   - Update router with organization routes
   - Add organization management link to AdminView
   - Update AdminUsersView for user-org assignment
   - Add active users stats card to DashboardView

4. **Testing**
   - Run migration on all three database types
   - Test organization CRUD operations
   - Test user assignment workflow
   - Verify dashboard card displays correctly
   - Test with various organization sizes

5. **Documentation**
   - Update CHANGELOG.md with v0.6.0 changes
   - Update README.md if needed
   - Document admin workflows for organization management

## Version History

- **v0.6.0** (2025-01-12) - Organization/Gym feature implementation started
  - Added organizations table
  - Added organization_id to users
  - Implemented organization CRUD operations
  - Added active users dashboard statistics

## References

- Plan file: `/home/jcz/.claude/plans/toasty-churning-lantern.md`
- CLAUDE.md: Project coding guidelines
- ARCHITECTURE.md: Clean Architecture patterns
- DATABASE_SCHEMA.md: Complete schema documentation
