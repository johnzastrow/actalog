# Test Plan for v0.3.0-beta

This document outlines the testing strategy for version 0.3.0-beta features.

## New Features to Test

1. **PR (Personal Records) Tracking System**
2. **Password Reset Flow**
3. **Multi-Database Support (MariaDB)**
4. **Environment Configuration (.env loading)**

---

## 1. PR Tracking System Testing

### Backend Endpoints

#### Test 1.1: Get Personal Records
**Endpoint:** `GET /api/workouts/prs`
**Authentication:** Required (JWT token)

**Test Steps:**
```bash
# 1. Login and get token
TOKEN=$(curl -s -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"Test12345"}' | \
  python3 -c "import sys, json; print(json.load(sys.stdin)['token'])")

# 2. Get personal records
curl -s "http://localhost:8080/api/workouts/prs" \
  -H "Authorization: Bearer $TOKEN" | python3 -m json.tool
```

**Expected Response:**
```json
{
  "records": [
    {
      "movement_id": 1,
      "movement_name": "Back Squat",
      "max_weight": 225.0,
      "max_reps": 10,
      "best_time": null,
      "workout_id": 5,
      "workout_date": "2025-11-09T00:00:00Z"
    }
  ]
}
```

**Test Cases:**
- ✅ Empty response when user has no PRs
- ✅ Returns only the user's PRs (not other users')
- ✅ Returns max weight, max reps, best time per movement
- ✅ Returns workout_id and workout_date for when PR was achieved
- ❌ 401 Unauthorized if no token provided
- ❌ 401 Unauthorized if invalid/expired token

---

#### Test 1.2: Get Recent PR Movements
**Endpoint:** `GET /api/workouts/pr-movements?limit=5`
**Authentication:** Required

**Test Steps:**
```bash
curl -s "http://localhost:8080/api/workouts/pr-movements?limit=5" \
  -H "Authorization: Bearer $TOKEN" | python3 -m json.tool
```

**Expected Response:**
```json
{
  "movements": [
    {
      "id": 123,
      "workout_id": 5,
      "movement_id": 1,
      "movement_name": "Back Squat",
      "weight": 225.0,
      "reps": 5,
      "is_pr": true,
      "workout_date": "2025-11-09T00:00:00Z"
    }
  ]
}
```

**Test Cases:**
- ✅ Limits results to requested number (default 10, max as configured)
- ✅ Returns movements ordered by date (most recent first)
- ✅ Only returns movements with `is_pr = true`
- ✅ Returns movement names, not just IDs
- ❌ 401 Unauthorized if no token

---

#### Test 1.3: Toggle PR Flag
**Endpoint:** `POST /api/workouts/movements/{id}/toggle-pr`
**Authentication:** Required

**Test Steps:**
```bash
# Toggle PR flag on workout_movement ID 123
curl -s -X POST "http://localhost:8080/api/workouts/movements/123/toggle-pr" \
  -H "Authorization: Bearer $TOKEN" | python3 -m json.tool
```

**Expected Response:**
```json
{
  "message": "PR flag toggled successfully"
}
```

**Test Cases:**
- ✅ Toggles is_pr from false → true
- ✅ Toggles is_pr from true → false
- ✅ Only the workout owner can toggle PR flag
- ❌ 403 Forbidden if user doesn't own the workout
- ❌ 400 Bad Request if invalid movement ID
- ❌ 404 Not Found if movement doesn't exist

---

### Frontend Testing

#### Test 1.4: PR History View
**Route:** `/prs`
**File:** `web/src/views/PRHistoryView.vue`

**Test Steps:**
1. Navigate to `/prs` route
2. Verify page loads without errors
3. Check "Recent PRs" section displays
4. Check "All Personal Records" section displays

**Visual Checks:**
- ✅ Gold trophy icons (mdi-trophy) visible
- ✅ PR badges with gold/amber color (#ffc107)
- ✅ Date formatting (Today, Yesterday, X days ago)
- ✅ Empty state shows "Start logging workouts to track PRs"
- ✅ Loading spinner appears while fetching data
- ✅ Error alert shows if API fails

---

#### Test 1.5: PR Badges on Workout Cards
**Component:** `web/src/components/RecentWorkoutsCards.vue`

**Test Steps:**
1. Navigate to Dashboard (`/`)
2. Check recent workout cards
3. Verify gold PR chip appears on workouts with PRs

**Visual Checks:**
- ✅ Gold PR chip with trophy icon on card header
- ✅ Individual trophy badges next to PR movements
- ✅ Movement details (weight, reps) display correctly

---

#### Test 1.6: PR Indicators in Workouts List
**Component:** `web/src/views/WorkoutsView.vue`

**Test Steps:**
1. Navigate to Workouts view (`/workouts`)
2. Check workout list items
3. Verify PR indicators next to movements

**Visual Checks:**
- ✅ Gold trophy icon with "PR" text next to PR movements
- ✅ Consistent styling with Rx badges

---

## 2. Password Reset Flow Testing

### Backend Endpoints

#### Test 2.1: Forgot Password
**Endpoint:** `POST /api/auth/forgot-password`
**Authentication:** Not required

**Test Steps:**
```bash
curl -s -X POST http://localhost:8080/api/auth/forgot-password \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com"}' | python3 -m json.tool
```

**Expected Response:**
```json
{
  "message": "If an account with that email exists, a password reset link has been sent"
}
```

**Test Cases:**
- ✅ Returns success message even if email doesn't exist (security)
- ✅ Generates secure random token (32 bytes hex)
- ✅ Token expires after configured duration (default 1 hour)
- ✅ Sends email with reset link (if SMTP configured)
- ✅ Logs warning if email service not configured
- ❌ 400 Bad Request if email missing

---

#### Test 2.2: Reset Password
**Endpoint:** `POST /api/auth/reset-password`
**Authentication:** Not required (uses token)

**Test Steps:**
```bash
# Use token from email or database
curl -s -X POST http://localhost:8080/api/auth/reset-password \
  -H "Content-Type: application/json" \
  -d '{"token":"RESET_TOKEN_HERE","new_password":"NewPass12345"}' | \
  python3 -m json.tool
```

**Expected Response:**
```json
{
  "message": "Password has been reset successfully. You can now login with your new password"
}
```

**Test Cases:**
- ✅ Validates token exists and is not expired
- ✅ Validates token hasn't been used already
- ✅ Hashes new password with bcrypt
- ✅ Updates user's password_hash
- ✅ Marks token as used (sets used_at timestamp)
- ❌ 400 Bad Request if token invalid
- ❌ 400 Bad Request if token expired
- ❌ 400 Bad Request if password too short (< 8 chars)

---

### Frontend Testing

#### Test 2.3: Forgot Password View
**Route:** `/forgot-password`
**File:** `web/src/views/ForgotPasswordView.vue`

**Test Steps:**
1. Click "Forgot password?" link on login page
2. Navigate to `/forgot-password`
3. Enter email address
4. Submit form

**Visual Checks:**
- ✅ Form displays with email input
- ✅ Loading state shows on submit
- ✅ Success message displays after submission
- ✅ Error alert shows if API fails
- ✅ Link to return to login page works

---

#### Test 2.4: Reset Password View
**Route:** `/reset-password/:token`
**File:** `web/src/views/ResetPasswordView.vue`

**Test Steps:**
1. Navigate to `/reset-password/TOKEN` (from email link)
2. Enter new password
3. Confirm new password
4. Submit form

**Visual Checks:**
- ✅ Form displays with password inputs
- ✅ Password visibility toggle works
- ✅ Password confirmation validation works
- ✅ Success message redirects to login
- ✅ Error alert shows if token invalid/expired

---

## 3. Multi-Database Support Testing

### Test 3.1: SQLite (Development)
**Configuration:**
```env
DB_DRIVER=sqlite3
DB_NAME=actalog.db
```

**Test Steps:**
1. Update `.env` with SQLite config
2. Restart application
3. Check logs confirm "Database Driver: sqlite3"
4. Test basic operations (register, login, create workout)

---

### Test 3.2: PostgreSQL (Production)
**Configuration:**
```env
DB_DRIVER=postgres
DB_HOST=localhost
DB_PORT=5432
DB_USER=actalog
DB_PASSWORD=secret
DB_NAME=actalog
DB_SSLMODE=disable
```

**Test Steps:**
1. Set up PostgreSQL database
2. Update `.env` with PostgreSQL config
3. Restart application
4. Check logs confirm "Database Driver: postgres"
5. Verify migrations run successfully
6. Test all CRUD operations

---

### Test 3.3: MariaDB/MySQL (Production)
**Configuration:**
```env
DB_DRIVER=mysql
DB_HOST=192.168.1.234
DB_PORT=3306
DB_USER=jcz
DB_PASSWORD=secret
DB_NAME=actalog
```

**Test Steps:**
1. Set up MariaDB database
2. Update `.env` with MariaDB config
3. Restart application
4. Check logs confirm "Database Driver: mysql"
5. Verify schema created correctly
6. Test movement seeding (31 standard movements)
7. Verify timestamp functions work (NOW() vs datetime('now'))

---

## 4. Environment Configuration Testing

### Test 4.1: .env File Loading
**File:** `.env`

**Test Steps:**
1. Create/modify `.env` file
2. Set `LOG_FILE_ENABLED=true`
3. Set `LOG_FILE_PATH=/path/to/logs/actalog.log`
4. Set `DB_DRIVER=mysql`
5. Restart application
6. Verify settings applied

**Verification:**
```bash
# Check log file created
ls -lh /path/to/logs/actalog.log

# Check database driver logged
tail -20 /path/to/logs/actalog.log | grep "Database Driver"
```

**Test Cases:**
- ✅ .env file loaded on startup
- ✅ LOG_FILE_PATH respected
- ✅ LOG_FILE_ENABLED=true creates log file
- ✅ DB_DRIVER setting applied
- ✅ All environment variables accessible

---

## 5. Security Testing

### Test 5.1: Authentication Required
**Test all protected endpoints without token:**

```bash
# Should all return 401 Unauthorized
curl -s http://localhost:8080/api/workouts
curl -s http://localhost:8080/api/workouts/prs
curl -s http://localhost:8080/api/workouts/pr-movements
```

---

### Test 5.2: User Isolation
**Test users can only access their own data:**

```bash
# User 1 creates workout
TOKEN1=$(curl -s -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"user1@example.com","password":"Pass12345"}' | \
  python3 -c "import sys, json; print(json.load(sys.stdin)['token'])")

# User 2 tries to access User 1's data
TOKEN2=$(curl -s -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"user2@example.com","password":"Pass12345"}' | \
  python3 -c "import sys, json; print(json.load(sys.stdin)['token'])")

# Should only see their own workouts
curl -s http://localhost:8080/api/workouts -H "Authorization: Bearer $TOKEN2"
```

---

### Test 5.3: PR Toggle Authorization
**Test users can't toggle PRs on other users' workouts:**

```bash
# User 2 tries to toggle PR on User 1's workout_movement
curl -s -X POST http://localhost:8080/api/workouts/movements/USER1_MOVEMENT_ID/toggle-pr \
  -H "Authorization: Bearer $TOKEN2"

# Should return 403 Forbidden
```

---

## 6. Integration Testing

### Test 6.1: End-to-End Workout + PR Flow

**Full workflow:**
1. Register new user
2. Login
3. Create workout with Back Squat @ 185 lbs x 5 reps
4. Create second workout with Back Squat @ 205 lbs x 5 reps (PR!)
5. Check `/api/workouts/prs` shows max weight 205 lbs
6. Check `/api/workouts/pr-movements` shows recent PR
7. Navigate to `/prs` in frontend
8. Verify PR displays with gold badge

---

### Test 6.2: Password Reset Flow

**Full workflow:**
1. Go to login page
2. Click "Forgot password?"
3. Enter email address
4. Submit request
5. Check email for reset link (or check logs if SMTP disabled)
6. Click reset link (opens `/reset-password/:token`)
7. Enter new password
8. Submit reset
9. Verify redirect to login
10. Login with new password

---

## Test Checklist Summary

### Backend
- [ ] PR tracking endpoints authenticated
- [ ] User isolation enforced
- [ ] PR auto-detection on workout creation
- [ ] Manual PR toggle works
- [ ] Password reset token generation
- [ ] Password reset token validation
- [ ] Password reset token expiration
- [ ] Multi-database support (SQLite, PostgreSQL, MariaDB)
- [ ] .env file loading

### Frontend
- [ ] PR History View displays correctly
- [ ] PR badges on workout cards
- [ ] PR indicators in workouts list
- [ ] Forgot password form works
- [ ] Reset password form works
- [ ] Loading states display
- [ ] Error handling works
- [ ] Navigation flows correctly

### Security
- [ ] Authentication required on protected routes
- [ ] User can only access own data
- [ ] Password reset tokens secure and time-limited
- [ ] Bcrypt password hashing (cost ≥ 12)
- [ ] SQL injection prevention (parameterized queries)

### Database
- [ ] Migrations run successfully
- [ ] Schema matches documentation
- [ ] Seeding works (31 movements)
- [ ] Multi-database compatibility
- [ ] Timestamp functions work across databases

---

## Known Issues

1. ~~**Hardcoded user IDs in handlers** - FIXED~~
   - All workout handlers now extract userID from JWT context
   - PR tracking handlers now use proper authentication

2. **Email service optional**
   - Password reset works but email not sent if SMTP not configured
   - Check logs for reset tokens in development

3. **Profile image field nullable**
   - Changed from `string` to `*string` in domain model
   - Prevents SQL scan errors for NULL values

---

## Next Steps

After testing v0.3.0-beta:
1. Create test user accounts with different scenarios
2. Populate database with sample workouts
3. Verify PR detection algorithm accuracy
4. Test across different browsers (mobile-first)
5. Load testing for multi-user scenarios
6. Review logs for any warnings/errors
