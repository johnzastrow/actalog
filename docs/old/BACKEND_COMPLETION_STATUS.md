# Backend & Frontend Completion Status

**Date:** November 12, 2025
**Current Version:** 0.3.2-beta

## Summary

This document tracks the implementation status of missing backend endpoints and frontend integrations identified during the completion audit.

---

## Phase 1: Movement Management ✅ **COMPLETE**

### Backend Implementation
- **File:** `internal/handler/movement_handler.go`
- **Lines Added:** 149-229
- **Endpoints:**
  - `PUT /api/movements/{id}` - Update custom movement ✅ TESTED
  - `DELETE /api/movements/{id}` - Delete custom movement ✅ TESTED

### Routes Wired
- **File:** `cmd/actalog/main.go`
- **Lines:** 246-247

### Testing Results
```bash
# Test results from curl commands:
✅ PUT /api/movements/32 - Successfully updated movement
✅ DELETE /api/movements/32 - Successfully deleted movement
```

### Frontend Integration
- **MovementEditView.vue** - Ready to use endpoints
- **MovementsLibraryView.vue** - Edit button functional

---

## Phase 2: User Settings & Security 🔄 **IN PROGRESS**

### Completed Files ✅

1. **Domain Layer**
   - `internal/domain/user_settings.go` ✅ CREATED
   - Defines UserSettings entity and UserSettingsRepository interface

2. **Repository Layer**
   - `internal/repository/user_settings_repository.go` ✅ CREATED
   - Implements CRUD operations for user_settings table
   - Methods: GetByUserID, Create, Update, Delete

3. **Service Layer**
   - `internal/service/user_settings_service.go` ✅ CREATED
   - Business logic for settings management
   - Auto-creates default settings if none exist

4. **Handler Layer**
   - `internal/handler/settings_handler.go` ✅ CREATED
   - Endpoints: GetSettings, UpdateSettings

### Remaining Tasks ⏳

1. **Password Change Functionality**
   - Add method to `internal/service/user_service.go`:
     ```go
     func (s *UserService) ChangePassword(userID int64, oldPassword, newPassword string) error
     ```
   - Add handler to `internal/handler/user_handler.go`:
     ```go
     func (h *UserHandler) ChangePassword(w http.ResponseWriter, r *http.Request)
     ```
   - Validate old password, hash new password with bcrypt

2. **Wire Routes in main.go**
   - Add to authenticated routes section:
     ```go
     // Settings routes
     r.Get("/users/settings", settingsHandler.GetSettings)
     r.Put("/users/settings", settingsHandler.UpdateSettings)
     r.Put("/users/password", userHandler.ChangePassword)
     ```
   - Initialize settings components in main():
     ```go
     settingsRepo := repository.NewSQLiteUserSettingsRepository(db)
     settingsService := service.NewUserSettingsService(settingsRepo)
     settingsHandler := handler.NewSettingsHandler(settingsService, appLogger)
     ```

3. **Frontend Integration**
   - **File:** `web/src/views/SettingsView.vue`
   - **Current State:** Uses localStorage for settings
   - **Changes Needed:**
     - Replace localStorage calls with API calls
     - GET /api/users/settings on component mount
     - PUT /api/users/settings on save
     - PUT /api/users/password for password change

### Database Schema
**Table:** `user_settings` (ALREADY EXISTS in v0.3.0 migration)
- No schema changes required ✅

---

## Phase 3: PR Tracking ⏳ **PENDING**

### Requirements Analysis

**Existing PR Infrastructure:**
- `workout_wods.is_pr` column exists ✅
- `workout_strength.is_pr` column exists ✅
- Toggle PR functionality exists for workout WODs ✅

### Missing Components

1. **PR Aggregation Repository Methods**
   - **File:** Create `internal/repository/pr_repository.go` OR extend existing repositories
   - **Methods Needed:**
     ```go
     // Get all personal records for a user
     GetPRsByUserID(userID int64) ([]*PRRecord, error)

     // Get recent PRs (last N days)
     GetRecentPRs(userID int64, days int) ([]*PRRecord, error)

     // Get PR movements (movements with PRs)
     GetPRMovements(userID int64, limit int) ([]*Movement, error)
     ```
   - **PR Record Structure:**
     ```go
     type PRRecord struct {
         MovementID   int64
         MovementName string
         Weight       float64
         Reps         int
         WorkoutDate  time.Time
     }
     ```

2. **PR Handler**
   - **File:** Create `internal/handler/pr_handler.go`
   - **Endpoints:**
     ```go
     GET /api/workouts/prs - Get all PRs for user
     GET /api/workouts/pr-movements?limit=5 - Get movements with PRs
     ```

3. **Routes** (commented out in main.go:279-282)
   ```go
   r.Get("/prs", prHandler.GetPersonalRecords)
   r.Get("/pr-movements", prHandler.GetPRMovements)
   ```

### Frontend Integration
- **File:** `web/src/views/PRHistoryView.vue` (lines 180-242)
- **Current Calls:**
   - Line 186: `GET /api/workouts/prs`
   - Line 190: `GET /api/workouts/pr-movements?limit=5`
- **Status:** Frontend ready, waiting for backend ⏳

---

## Additional Features (Future Work)

### Account Management

1. **Account Deletion**
   - Endpoint: `DELETE /api/users/account`
   - Service method: Delete user and all related data
   - Frontend: SettingsView.vue line 561

2. **Data Export**
   - Endpoint: `GET /api/export?format=json|csv`
   - Export user data (workouts, PRs, settings)
   - Frontend: SettingsView.vue line 532

### Admin Features

1. **Admin Middleware**
   - **File:** Create `pkg/middleware/admin.go`
   - **Function:** `RequireAdmin()` middleware
   - Uses existing `GetUserRole` from auth.go

2. **User Management Endpoints**
   - `GET /api/admin/users` - List all users
   - `GET /api/admin/users/{id}` - Get user details
   - `PUT /api/admin/users/{id}/role` - Change user role
   - `PUT /api/admin/users/{id}/status` - Enable/disable user
   - `DELETE /api/admin/users/{id}` - Delete user

---

## Implementation Guide

### Quick Start for Completing Phase 2

1. **Add password change to user_service.go:**
   ```go
   func (s *UserService) ChangePassword(userID int64, oldPassword, newPassword string) error {
       user, err := s.userRepo.GetByID(userID)
       if err != nil || user == nil {
           return ErrUserNotFound
       }

       if !auth.CheckPassword(oldPassword, user.Password) {
           return ErrInvalidCredentials
       }

       hashedPassword, err := auth.HashPassword(newPassword)
       if err != nil {
           return err
       }

       return s.userRepo.UpdatePassword(userID, hashedPassword)
   }
   ```

2. **Add UpdatePassword to user repository interface (domain/user.go):**
   ```go
   UpdatePassword(userID int64, hashedPassword string) error
   ```

3. **Implement in repository/user_repository.go:**
   ```go
   func (r *SQLiteUserRepository) UpdatePassword(userID int64, hashedPassword string) error {
       query := `UPDATE users SET password = ?, updated_at = ? WHERE id = ?`
       _, err := r.db.Exec(query, hashedPassword, time.Now(), userID)
       return err
   }
   ```

4. **Add handler method to user_handler.go:**
   ```go
   func (h *UserHandler) ChangePassword(w http.ResponseWriter, r *http.Request) {
       userID, ok := middleware.GetUserID(r.Context())
       if !ok {
           respondError(w, http.StatusUnauthorized, "Unauthorized")
           return
       }

       var req struct {
           OldPassword string `json:"old_password"`
           NewPassword string `json:"new_password"`
       }

       if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
           respondError(w, http.StatusBadRequest, "Invalid request body")
           return
       }

       if err := h.userService.ChangePassword(userID, req.OldPassword, req.NewPassword); err != nil {
           if err == service.ErrInvalidCredentials {
               respondError(w, http.StatusUnauthorized, "Current password is incorrect")
               return
           }
           respondError(w, http.StatusInternalServerError, "Failed to change password")
           return
       }

       respondJSON(w, http.StatusOK, map[string]string{"message": "Password changed successfully"})
   }
   ```

5. **Wire everything in main.go** (see section above)

6. **Build and test:**
   ```bash
   make build
   curl -X PUT http://localhost:8080/api/users/settings \
     -H "Authorization: Bearer $TOKEN" \
     -H "Content-Type: application/json" \
     -d '{"theme":"dark","weight_unit":"kg"}'
   ```

---

## Testing Checklist

### Phase 1: Movement Management
- [x] PUT /api/movements/{id} - Update movement
- [x] DELETE /api/movements/{id} - Delete movement
- [x] Frontend MovementEditView integration

### Phase 2: User Settings
- [ ] GET /api/users/settings - Retrieve settings
- [ ] PUT /api/users/settings - Update settings
- [ ] PUT /api/users/password - Change password
- [ ] Frontend SettingsView integration

### Phase 3: PR Tracking
- [ ] GET /api/workouts/prs - Get all PRs
- [ ] GET /api/workouts/pr-movements - Get PR movements
- [ ] Frontend PRHistoryView integration

---

## Estimated Completion Times

- **Phase 2 Remaining:** 1-2 hours
- **Phase 3 Complete:** 2-3 hours
- **Admin Features:** 3-4 hours
- **Total Remaining:** 6-9 hours

---

## Notes

- All database tables exist from v0.3.0 migration
- No schema changes required for current scope
- Repository patterns established and consistent
- Frontend views are already implemented and waiting for backend
