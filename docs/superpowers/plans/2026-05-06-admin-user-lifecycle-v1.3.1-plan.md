# Admin User Lifecycle (v1.3.1) Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add admin-side user creation (`POST /api/admin/users`) and admin-set-password (`POST /api/admin/users/{id}/password`) — backend + frontend dialogs + audit events — so admins can fully manage user lifecycle from the admin UI.

**Architecture:** Two new HTTP endpoints reusing existing repository methods (`UserRepository.Create`, `UpdatePassword`, `UnlockAccount`, `RevokeAllForUser`). Two new Vuetify dialogs (`AdminUserCreateDialog`, `AdminSetPasswordDialog`) plus a `usePasswordInputs` composable extracted for reuse. Three new audit events. Existing protected-user defense (`ProtectedUserGuard` + L1/L2 audit family) covers the protected-target attack surface; protected emails are also rejected at create with their own audit event.

**Tech Stack:** Go (Chi router), SQLite/PostgreSQL/MySQL, Vue 3 + Vuetify 3 + Vitest, bcrypt cost-12 password hashing.

**Spec:** `/home/jcz/Github/actionlog/.worktrees/feature/admin-user-lifecycle-v1.3.1/docs/superpowers/specs/2026-05-06-admin-user-lifecycle-v1.3.1-design.md`

---

## File map

**New backend files:** none — all changes are in existing files.

**Modified backend files:**
- `internal/domain/audit_log.go` — three new event constants
- `internal/service/admin_user_service.go` — `CreateUser`, `SetPassword` methods
- `internal/service/admin_user_service_test.go` — tests for the above
- `internal/handler/admin_user_handler.go` — `CreateUser`, `SetPassword` handlers
- `internal/handler/admin_user_handler_test.go` — tests for the above
- `cmd/actalog/main.go` — two route registrations

**New backend test file:**
- `test/integration/admin_user_lifecycle_test.go`

**New frontend files:**
- `web/src/components/admin/AdminUserCreateDialog.vue`
- `web/src/components/admin/AdminUserCreateDialog.test.js`
- `web/src/components/admin/AdminSetPasswordDialog.vue`
- `web/src/components/admin/AdminSetPasswordDialog.test.js`
- `web/src/components/admin/composables/usePasswordInputs.js`
- `web/src/components/admin/composables/usePasswordInputs.test.js`

**Modified frontend files:**
- `web/src/views/AdminUsersView.vue` — Create button + dialog wiring
- `web/src/views/AdminUsersView.test.js` — verify Create button visibility
- `web/src/components/admin/user-edit/ProfileTab.vue` — Password Management card

**Doc updates:**
- `docs/security/THREAT_MODEL.md`
- `docs/CHANGELOG.md`
- `docs/USER_PERMISSIONS.md`
- `docs/TODO.md`
- `pkg/version/version.go`
- `web/package.json`
- `CLAUDE.md` (version line)

---

## Conventions used in this plan

- All file paths are relative to the repo root (`/home/jcz/Github/actionlog/.worktrees/feature/admin-user-lifecycle-v1.3.1`).
- Backend tests run with `go test ./<pkg>/ -count=1 -run <Name>`. Always pass `-count=1` to defeat Go's test caching.
- Frontend tests run with `cd web && npm run test:run -- <pattern>`.
- Each task ends with `git add` listing every file touched explicitly (don't `git add .` per CLAUDE.md credential-scanning rule).
- All commits are conventional-commit style: `feat:`, `fix:`, `test:`, `docs:`, `refactor:`.
- Existing pattern: `validatePassword(password)` is private to `package service` (`internal/service/user_service.go:103`). The new admin methods call it directly.

---

## Task 1: Add three new audit event constants

**Files:**
- Modify: `internal/domain/audit_log.go`

- [ ] **Step 1: Add the three constants**

Find the existing `EventProtectedUserAttack*` block (around line 161) and add the three new constants in the user-management section (around line 47-50).

In `internal/domain/audit_log.go`, find:

```go
EventUserCreated        = "user_created"
EventUserUpdated        = "user_updated"
EventUserDeleted        = "user_deleted"
EventRoleChanged        = "role_changed" // Admin promoted/demoted user
EventProfileUpdated     = "profile_updated"
EventUserSettingsUpdate = "user_settings_updated"
```

After `EventUserSettingsUpdate`, add:

```go
// Admin user lifecycle events (v1.3.1)
EventAdminUserCreated                  = "admin_user_created"                    // POST /api/admin/users 201
EventAdminPasswordSet                  = "admin_password_set"                    // POST /api/admin/users/{id}/password 204
EventAdminUserCreateRejectedProtected  = "admin_user_create_rejected_protected"  // create attempted with a protected email
```

- [ ] **Step 2: Verify the package compiles**

```bash
go build ./internal/domain/...
```

Expected: no output, exit 0.

- [ ] **Step 3: Commit**

```bash
git add internal/domain/audit_log.go
git commit -m "feat(audit): add admin user lifecycle event constants

Three new event types for v1.3.1 admin user creation + admin-set-password:
- admin_user_created
- admin_password_set
- admin_user_create_rejected_protected"
```

---

## Task 2: AdminUserService.CreateUser — service method

**Files:**
- Modify: `internal/service/admin_user_service.go`
- Modify: `internal/service/admin_user_service_test.go`

- [ ] **Step 1: Write the failing tests (table-driven happy + rejections)**

Append to `internal/service/admin_user_service_test.go`:

```go
// TestAdminUserService_CreateUser_HappyPath verifies the admin can create a
// user account that the new user can immediately authenticate with.
func TestAdminUserService_CreateUser_HappyPath(t *testing.T) {
	repo := newStubUserRepo()
	svc := newService(repo, &stubRefreshTokenRepo{}, &stubEmailSvc{}, &stubAuditLogger{})

	created, err := svc.CreateUser(1, CreateUserFields{
		Email:         "newathlete@example.com",
		Password:      "SetByAdminAtCreate1",
		Name:          "New Athlete",
		Role:          "athlete",
		EmailVerified: true,
	})
	if err != nil {
		t.Fatalf("CreateUser: %v", err)
	}
	if created.ID == 0 {
		t.Error("expected user to have an assigned ID after Create")
	}
	if created.Email != "newathlete@example.com" {
		t.Errorf("Email = %q", created.Email)
	}
	if !created.EmailVerified {
		t.Error("EmailVerified should be true (admin-vouched)")
	}
	if created.PasswordHash == "" {
		t.Error("PasswordHash must be set")
	}
	if created.PasswordHash == "SetByAdminAtCreate1" {
		t.Error("PasswordHash must be hashed, not the raw password")
	}
}

// TestAdminUserService_CreateUser_ValidationRejections covers every input
// rule with a single table-driven test. Each case must return a typed
// *domain.InvalidInputError so the handler surfaces 400.
func TestAdminUserService_CreateUser_ValidationRejections(t *testing.T) {
	cases := []struct {
		name      string
		fields    CreateUserFields
		wantField string
	}{
		{
			name:      "empty name",
			fields:    CreateUserFields{Email: "a@b.com", Password: "ValidPass123Long", Name: "", Role: "athlete"},
			wantField: "name",
		},
		{
			name:      "name over 100 chars",
			fields:    CreateUserFields{Email: "a@b.com", Password: "ValidPass123Long", Name: strings.Repeat("x", 101), Role: "athlete"},
			wantField: "name",
		},
		{
			name:      "malformed email",
			fields:    CreateUserFields{Email: "not-an-email", Password: "ValidPass123Long", Name: "X", Role: "athlete"},
			wantField: "email",
		},
		{
			name:      "password too short",
			fields:    CreateUserFields{Email: "a@b.com", Password: "Short1", Name: "X", Role: "athlete"},
			wantField: "password",
		},
		{
			name:      "password missing digit",
			fields:    CreateUserFields{Email: "a@b.com", Password: "NoDigitsHereXYZ", Name: "X", Role: "athlete"},
			wantField: "password",
		},
		{
			name:      "password missing upper",
			fields:    CreateUserFields{Email: "a@b.com", Password: "alllower123long", Name: "X", Role: "athlete"},
			wantField: "password",
		},
		{
			name:      "password missing lower",
			fields:    CreateUserFields{Email: "a@b.com", Password: "ALLUPPER123LONG", Name: "X", Role: "athlete"},
			wantField: "password",
		},
		{
			name:      "invalid role",
			fields:    CreateUserFields{Email: "a@b.com", Password: "ValidPass123Long", Name: "X", Role: "wizard"},
			wantField: "role",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			repo := newStubUserRepo()
			svc := newService(repo, &stubRefreshTokenRepo{}, &stubEmailSvc{}, &stubAuditLogger{})

			_, err := svc.CreateUser(1, tc.fields)
			if err == nil {
				t.Fatal("expected validation error, got nil")
			}
			var invErr *domain.InvalidInputError
			if !errors.As(err, &invErr) {
				t.Fatalf("expected *domain.InvalidInputError, got %T: %v", err, err)
			}
			if invErr.Field != tc.wantField {
				t.Errorf("Field = %q, want %q", invErr.Field, tc.wantField)
			}
		})
	}
}

// TestAdminUserService_CreateUser_DuplicateEmail verifies the service rejects
// a second create with an email that already exists. Returns a sentinel the
// handler can map to HTTP 409.
func TestAdminUserService_CreateUser_DuplicateEmail(t *testing.T) {
	repo := newStubUserRepo()
	repo.addUser(makeNormalUser(7, "taken@example.com", "Existing"))
	svc := newService(repo, &stubRefreshTokenRepo{}, &stubEmailSvc{}, &stubAuditLogger{})

	_, err := svc.CreateUser(1, CreateUserFields{
		Email:    "taken@example.com",
		Password: "ValidPass123Long",
		Name:     "Dup",
		Role:     "athlete",
	})
	if err == nil {
		t.Fatal("expected duplicate-email error, got nil")
	}
	if !errors.Is(err, ErrEmailAlreadyExists) {
		t.Errorf("expected ErrEmailAlreadyExists, got: %v", err)
	}
}

// TestAdminUserService_CreateUser_ProtectedEmailRejected verifies that an
// admin cannot create a new user using an email on the protected-users list.
// This must fire EventAdminUserCreateRejectedProtected, NOT the
// protected_user_attack_* family (those are for modifying existing protected
// rows; this is for attempted creation).
func TestAdminUserService_CreateUser_ProtectedEmailRejected(t *testing.T) {
	// Pick the first email from the production protected list so the test
	// stays in sync with security.IsProtectedEmail without duplicating it.
	protectedEmail := security.ProtectedEmailsList()[0]

	repo := newStubUserRepo()
	auditLog := &stubAuditLogger{}
	svc := newService(repo, &stubRefreshTokenRepo{}, &stubEmailSvc{}, auditLog)

	_, err := svc.CreateUser(1, CreateUserFields{
		Email:    protectedEmail,
		Password: "ValidPass123Long",
		Name:     "Imposter",
		Role:     "admin",
	})
	if err == nil {
		t.Fatal("expected protected-email rejection, got nil")
	}
	var invErr *domain.InvalidInputError
	if !errors.As(err, &invErr) {
		t.Fatalf("expected *domain.InvalidInputError, got %T", err)
	}
	if invErr.Field != "email" {
		t.Errorf("Field = %q, want %q", invErr.Field, "email")
	}
	if !strings.Contains(invErr.Message, "reserved") {
		t.Errorf("Message should mention 'reserved'; got %q", invErr.Message)
	}
	// Audit must capture the attempt.
	if len(auditLog.events) != 1 {
		t.Fatalf("expected exactly 1 audit event, got %d: %v", len(auditLog.events), auditLog.events)
	}
	if auditLog.events[0].eventType != domain.EventAdminUserCreateRejectedProtected {
		t.Errorf("event = %q, want %q", auditLog.events[0].eventType, domain.EventAdminUserCreateRejectedProtected)
	}
}

// TestAdminUserService_CreateUser_AuditOnSuccess verifies the success path
// fires EventAdminUserCreated with the role and email_verified in the details.
func TestAdminUserService_CreateUser_AuditOnSuccess(t *testing.T) {
	repo := newStubUserRepo()
	auditLog := &stubAuditLogger{}
	svc := newService(repo, &stubRefreshTokenRepo{}, &stubEmailSvc{}, auditLog)

	_, err := svc.CreateUser(1, CreateUserFields{
		Email:         "auditme@example.com",
		Password:      "ValidPass123Long",
		Name:          "Audit",
		Role:          "coach",
		EmailVerified: true,
	})
	if err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	if len(auditLog.events) != 1 {
		t.Fatalf("expected 1 audit event, got %d", len(auditLog.events))
	}
	ev := auditLog.events[0]
	if ev.eventType != domain.EventAdminUserCreated {
		t.Errorf("eventType = %q, want %q", ev.eventType, domain.EventAdminUserCreated)
	}
	if ev.details["role"] != "coach" {
		t.Errorf("details.role = %v, want %q", ev.details["role"], "coach")
	}
	if ev.details["email_verified"] != true {
		t.Errorf("details.email_verified = %v, want true", ev.details["email_verified"])
	}
	if ev.details["email_domain"] != "example.com" {
		t.Errorf("details.email_domain = %v, want %q", ev.details["email_domain"], "example.com")
	}
}
```

- [ ] **Step 2: Run tests to verify they fail with "method not defined"**

```bash
go test ./internal/service/ -run "TestAdminUserService_CreateUser" -count=1 -v
```

Expected: FAIL with `undefined: CreateUserFields` and `undefined: ErrEmailAlreadyExists` and `svc.CreateUser undefined`.

- [ ] **Step 3: Inspect existing helpers we'll reuse**

Read `internal/service/admin_user_service.go` lines 1-100 to confirm:
- The struct has `userRepo`, `auditLogService`
- `domain.User` shape matches what we'll build
- The package already imports `domain`, `security`, `auth`, etc.

Read `internal/service/user_service.go:103-122` to confirm `validatePassword` exists and is package-private (callable from `admin_user_service.go`).

Note: stubs `stubUserRepo`, `stubAuditLogger`, `stubEmailSvc`, `stubRefreshTokenRepo`, `newStubUserRepo`, `newService`, `makeNormalUser` are already defined in `internal/service/admin_user_service_test.go` from v1.3.0. Verify `stubAuditLogger` records `eventType` and `details`; if it doesn't, extend it (do that now while writing tests).

- [ ] **Step 4: Add the field type, sentinel, and method to admin_user_service.go**

In `internal/service/admin_user_service.go`, after the existing `ProfileUpdateFields` struct (around line 30), add:

```go
// CreateUserFields is the input to AdminUserService.CreateUser. Exactly one
// instance per admin create — every field is set or left zero by the handler
// after JSON decode. Defaults (role="athlete", EmailVerified=true) are applied
// inside CreateUser, not at decode time, so the wire format stays explicit.
type CreateUserFields struct {
	Email         string
	Password      string
	Name          string
	Role          string
	EmailVerified bool
}
```

Then in the same file, after the existing sentinels (search for `var Err`), add:

```go
// ErrEmailAlreadyExists is returned by CreateUser when the requested email is
// already in use. Mapped to HTTP 409 by the handler error pipeline.
var ErrEmailAlreadyExists = errors.New("email already exists")
```

If `errors` isn't imported, add it (`"errors"` in the import block).

Then add the method (place it near the existing `UpdateProfile` method):

```go
// CreateUser creates a new user account from admin context.
//
// Validates input strictly. On protected-email attempt, fires
// EventAdminUserCreateRejectedProtected and returns *domain.InvalidInputError.
// On success, persists the user and fires EventAdminUserCreated.
//
// actorID is the admin's user ID (used for audit attribution).
//
// Defaults applied here (not at decode time): Role="athlete" if empty;
// EmailVerified=true is the default expressed by the handler — the field is
// passed through verbatim and used as-is.
func (s *AdminUserService) CreateUser(actorID int64, fields CreateUserFields) (*domain.User, error) {
	// 1. Email syntax + protected check (before duplicate check, so the audit
	//    event is fired for protected-email attempts even if the email also
	//    happens to be a duplicate).
	addr, parseErr := mail.ParseAddress(fields.Email)
	if parseErr != nil {
		return nil, &domain.InvalidInputError{Field: "email", Message: "is not a valid address", Cause: parseErr}
	}
	emailLower := strings.ToLower(addr.Address)
	if security.IsProtectedEmail(emailLower) {
		// Audit before returning so the operator's intent is captured.
		_ = s.auditLogService.LogEvent(
			domain.EventAdminUserCreateRejectedProtected,
			&actorID,
			nil,
			nil, nil,
			map[string]interface{}{"attempted_email": emailLower},
		)
		return nil, &domain.InvalidInputError{Field: "email", Message: "is reserved"}
	}

	// 2. Name length.
	name := strings.TrimSpace(fields.Name)
	if name == "" || len(name) > 100 {
		return nil, &domain.InvalidInputError{Field: "name", Message: "must be 1–100 characters"}
	}

	// 3. Password complexity (reuses the package-private validator that
	//    Register / ChangePassword / ResetPassword all use).
	if err := validatePassword(fields.Password); err != nil {
		return nil, &domain.InvalidInputError{Field: "password", Message: err.Error(), Cause: err}
	}

	// 4. Role allowlist. Default to athlete if empty.
	role := fields.Role
	if role == "" {
		role = "athlete"
	}
	if role != "athlete" && role != "coach" && role != "admin" {
		return nil, &domain.InvalidInputError{Field: "role", Message: "must be athlete, coach, or admin"}
	}

	// 5. Duplicate email check via repo. We accept the small race between
	//    check and insert because UserRepository.Create will surface a unique
	//    constraint violation on race; treat that as the same condition.
	existing, lookupErr := s.userRepo.GetByEmail(emailLower)
	if lookupErr != nil {
		return nil, fmt.Errorf("CreateUser: lookup existing: %w", lookupErr)
	}
	if existing != nil {
		return nil, ErrEmailAlreadyExists
	}

	// 6. Hash password and persist.
	hash, hashErr := auth.HashPassword(fields.Password)
	if hashErr != nil {
		return nil, fmt.Errorf("CreateUser: hash password: %w", hashErr)
	}

	now := time.Now().UTC()
	user := &domain.User{
		Email:        emailLower,
		PasswordHash: hash,
		Name:         name,
		Role:         role,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	if createErr := s.userRepo.Create(user); createErr != nil {
		// Surface duplicate-email as the typed sentinel even if the unique
		// constraint trips on race.
		if isUniqueViolation(createErr) {
			return nil, ErrEmailAlreadyExists
		}
		return nil, fmt.Errorf("CreateUser: persist: %w", createErr)
	}

	// 7. If admin set EmailVerified=true, set it via Update (the Create INSERT
	//    only sets the six core columns).
	if fields.EmailVerified {
		user.EmailVerified = true
		verifiedAt := time.Now().UTC()
		user.EmailVerifiedAt = &verifiedAt
		if updateErr := s.userRepo.Update(user); updateErr != nil {
			// Non-fatal: user was created; admin can re-verify via the
			// existing toggle endpoint. Log + audit the partial state.
			s.log.Error("AdminUserService.CreateUser: post-create email_verified set: %v", updateErr)
		}
	}

	// 8. Audit success.
	emailDomain := emailLower
	if at := strings.LastIndex(emailLower, "@"); at >= 0 {
		emailDomain = emailLower[at+1:]
	}
	if auditErr := s.auditLogService.LogEvent(
		domain.EventAdminUserCreated,
		&actorID,
		&user.ID,
		nil, nil,
		map[string]interface{}{
			"role":           role,
			"email_verified": user.EmailVerified,
			"email_domain":   emailDomain,
		},
	); auditErr != nil {
		s.log.Error("AdminUserService.CreateUser: audit: %v", auditErr)
	}

	return user, nil
}

// isUniqueViolation returns true if err is a database UNIQUE constraint
// violation. Driver-specific message sniffing is unavoidable here — the
// database/sql interface doesn't expose a portable error code.
func isUniqueViolation(err error) bool {
	if err == nil {
		return false
	}
	s := strings.ToLower(err.Error())
	return strings.Contains(s, "unique constraint") ||
		strings.Contains(s, "unique") && strings.Contains(s, "constraint") ||
		strings.Contains(s, "duplicate entry") ||
		strings.Contains(s, "duplicate key")
}
```

Add the `auth` import if not already present:

```go
"github.com/johnzastrow/actalog/pkg/auth"
```

- [ ] **Step 5: Extend stubAuditLogger to record events for assertion**

Find `stubAuditLogger` in `internal/service/admin_user_service_test.go`. If `events` is not already a slice with `eventType` and `details` captured, extend it:

```go
type recordedAuditEvent struct {
	eventType string
	actorID   *int64
	targetID  *int64
	details   map[string]interface{}
}

type stubAuditLogger struct {
	events []recordedAuditEvent
}

func (s *stubAuditLogger) LogEvent(eventType string, actorID, targetID *int64, _, _ interface{}, details map[string]interface{}) error {
	s.events = append(s.events, recordedAuditEvent{
		eventType: eventType,
		actorID:   actorID,
		targetID:  targetID,
		details:   details,
	})
	return nil
}
```

If it already records events under different field names, leave it alone and adjust the test assertions in Step 1 to match (the test code is the *only* thing that should reference these fields).

- [ ] **Step 6: Run tests, verify they pass**

```bash
go test ./internal/service/ -run "TestAdminUserService_CreateUser" -count=1 -v
```

Expected: all 5 tests + 8 subtests PASS.

- [ ] **Step 7: Commit**

```bash
git add internal/service/admin_user_service.go internal/service/admin_user_service_test.go
git commit -m "feat(service): AdminUserService.CreateUser

Admin-context user creation with full input validation, protected-email
rejection (audited), duplicate-email detection (typed sentinel for 409
mapping), and EventAdminUserCreated on success.

Reuses validatePassword, auth.HashPassword, and security.IsProtectedEmail.
No new repository methods — UserRepository.Create + Update covers it."
```

---

## Task 3: AdminUserHandler.CreateUser — HTTP handler

**Files:**
- Modify: `internal/handler/admin_user_handler.go`
- Modify: `internal/handler/admin_user_handler_test.go`
- Modify: `internal/handler/errors.go` (extend WriteError to map ErrEmailAlreadyExists → 409)

- [ ] **Step 1: Write failing handler tests**

Append to `internal/handler/admin_user_handler_test.go`:

```go
// TestAdminUserHandler_CreateUser_201Success exercises the happy path: valid
// JSON in, 201 + user JSON out, service.CreateUser called with the decoded fields.
func TestAdminUserHandler_CreateUser_201Success(t *testing.T) {
	stub := &stubAdminUserService{
		createUserResult: &domain.User{
			ID:    42,
			Email: "newathlete@example.com",
			Name:  "New",
			Role:  "athlete",
		},
	}
	h := NewAdminUserHandler(nil, stub, nil)

	body := strings.NewReader(`{"email":"newathlete@example.com","password":"ValidPass123Long","name":"New","role":"athlete","email_verified":true}`)
	req := httptest.NewRequest(http.MethodPost, "/api/admin/users", body)
	ctx := middleware.WithUserID(req.Context(), 1)
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	h.CreateUser(rr, req)

	if rr.Code != http.StatusCreated {
		t.Errorf("Code = %d, want 201; body=%s", rr.Code, rr.Body.String())
	}
	if !stub.createUserCalled {
		t.Error("service.CreateUser not invoked")
	}
	if stub.createUserActorID != 1 {
		t.Errorf("actorID = %d, want 1", stub.createUserActorID)
	}
	if stub.createUserFields.Email != "newathlete@example.com" {
		t.Errorf("Email = %q", stub.createUserFields.Email)
	}
	if !stub.createUserFields.EmailVerified {
		t.Error("EmailVerified should be true (decoded from request)")
	}
}

// TestAdminUserHandler_CreateUser_400OnInvalidJSON returns a clear bad-request
// status without invoking the service.
func TestAdminUserHandler_CreateUser_400OnInvalidJSON(t *testing.T) {
	stub := &stubAdminUserService{}
	h := NewAdminUserHandler(nil, stub, nil)

	req := httptest.NewRequest(http.MethodPost, "/api/admin/users", strings.NewReader(`{not json`))
	ctx := middleware.WithUserID(req.Context(), 1)
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	h.CreateUser(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("Code = %d, want 400", rr.Code)
	}
	if stub.createUserCalled {
		t.Error("service should not be called on JSON decode failure")
	}
}

// TestAdminUserHandler_CreateUser_400OnValidationError surfaces InvalidInputError as 400.
func TestAdminUserHandler_CreateUser_400OnValidationError(t *testing.T) {
	stub := &stubAdminUserService{
		createUserErr: &domain.InvalidInputError{Field: "password", Message: "must be at least 12 characters"},
	}
	h := NewAdminUserHandler(nil, stub, nil)

	req := httptest.NewRequest(http.MethodPost, "/api/admin/users",
		strings.NewReader(`{"email":"a@b.com","password":"x","name":"X","role":"athlete"}`))
	ctx := middleware.WithUserID(req.Context(), 1)
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	h.CreateUser(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("Code = %d, want 400", rr.Code)
	}
	if !strings.Contains(rr.Body.String(), "password") {
		t.Errorf("body should include the field name; got %s", rr.Body.String())
	}
}

// TestAdminUserHandler_CreateUser_409OnDuplicateEmail maps the service sentinel.
func TestAdminUserHandler_CreateUser_409OnDuplicateEmail(t *testing.T) {
	stub := &stubAdminUserService{createUserErr: service.ErrEmailAlreadyExists}
	h := NewAdminUserHandler(nil, stub, nil)

	req := httptest.NewRequest(http.MethodPost, "/api/admin/users",
		strings.NewReader(`{"email":"a@b.com","password":"ValidPass123Long","name":"X","role":"athlete"}`))
	ctx := middleware.WithUserID(req.Context(), 1)
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	h.CreateUser(rr, req)

	if rr.Code != http.StatusConflict {
		t.Errorf("Code = %d, want 409; body=%s", rr.Code, rr.Body.String())
	}
}
```

- [ ] **Step 2: Extend the handler's stub service**

Find `stubAdminUserService` in `internal/handler/admin_user_handler_test.go`. Add the create-user fields:

```go
// (Inside the existing stubAdminUserService struct, add:)
createUserCalled   bool
createUserActorID  int64
createUserFields   service.CreateUserFields
createUserResult   *domain.User
createUserErr      error
```

And the method on the stub:

```go
func (s *stubAdminUserService) CreateUser(actorID int64, fields service.CreateUserFields) (*domain.User, error) {
	s.createUserCalled = true
	s.createUserActorID = actorID
	s.createUserFields = fields
	return s.createUserResult, s.createUserErr
}
```

If the handler uses an interface to talk to the service (it does — search for `type adminUserService interface` or similar), add `CreateUser` to that interface signature.

- [ ] **Step 3: Run tests to confirm failure**

```bash
go test ./internal/handler/ -run "TestAdminUserHandler_CreateUser" -count=1 -v
```

Expected: FAIL — undefined `h.CreateUser`.

- [ ] **Step 4: Add WriteError mapping for ErrEmailAlreadyExists → 409**

In `internal/handler/errors.go`, find the `errors.As(err, &invErr)` branch added in v1.3.0 hotfix #220. Add a check for the duplicate sentinel before that branch:

```go
// Service-layer duplicate-email sentinel: map to 409 with a stable error
// code the frontend can act on (e.g., highlight the email input).
if errors.Is(err, service.ErrEmailAlreadyExists) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusConflict)
	json.NewEncoder(w).Encode(ErrorResponse{
		Error:   "duplicate_email",
		Message: "A user with that email already exists.",
	})
	return
}
```

If `service` is not already imported, add `"github.com/johnzastrow/actalog/internal/service"`.

- [ ] **Step 5: Add the handler**

In `internal/handler/admin_user_handler.go`, after the existing `ForcePasswordReset` method, add:

```go
// createUserRequest is the POST /api/admin/users request body.
type createUserRequest struct {
	Email         string `json:"email"`
	Password      string `json:"password"`
	Name          string `json:"name"`
	Role          string `json:"role"`
	EmailVerified *bool  `json:"email_verified,omitempty"` // nil → default true
}

// CreateUser handles POST /api/admin/users — admin creates a new user account.
//
//   - 201 with the created user JSON on success
//   - 400 invalid_input on validation failure
//   - 409 duplicate_email if email already exists
//   - 401/403 from auth middleware
func (h *AdminUserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	actorID, ok := middleware.GetUserID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}
	var req createUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid JSON")
		return
	}
	// Default EmailVerified=true if the caller omitted the field.
	emailVerified := true
	if req.EmailVerified != nil {
		emailVerified = *req.EmailVerified
	}
	user, err := h.adminUserService.CreateUser(actorID, service.CreateUserFields{
		Email:         req.Email,
		Password:      req.Password,
		Name:          req.Name,
		Role:          req.Role,
		EmailVerified: emailVerified,
	})
	if err != nil {
		WriteError(w, err)
		return
	}
	respondJSON(w, http.StatusCreated, user)
}
```

- [ ] **Step 6: Run tests, verify they pass**

```bash
go test ./internal/handler/ -run "TestAdminUserHandler_CreateUser" -count=1 -v
```

Expected: 4 PASS.

- [ ] **Step 7: Commit**

```bash
git add internal/handler/admin_user_handler.go internal/handler/admin_user_handler_test.go internal/handler/errors.go
git commit -m "feat(handler): AdminUserHandler.CreateUser + 409 mapping for duplicate email

POST /api/admin/users handler: decode body, default email_verified=true,
delegate to service, map success to 201. WriteError gains a sentinel
branch for service.ErrEmailAlreadyExists → 409 duplicate_email."
```

---

## Task 4: Wire POST /api/admin/users route

**Files:**
- Modify: `cmd/actalog/main.go`

- [ ] **Step 1: Add the route registration**

In `cmd/actalog/main.go`, find:

```go
r.Get("/users", adminUserHandler.ListUsers)
```

(should be around line 995). Add right after:

```go
r.Post("/users", adminUserHandler.CreateUser)
```

This places it OUTSIDE the `r.Route("/users/{id}", ...)` sub-router because creation has no `{id}` and doesn't need the ProtectedUserGuard (the guard fires on existing-row writes; create attempts are filtered at the service layer).

- [ ] **Step 2: Build to verify wiring**

```bash
go build ./cmd/actalog/
```

Expected: no output, exit 0.

- [ ] **Step 3: Run a quick smoke test in-process**

```bash
go test ./internal/handler/ -count=1 -run "TestAdminUserHandler_CreateUser" -v 2>&1 | tail -8
```

Expected: 4 PASS.

- [ ] **Step 4: Commit**

```bash
git add cmd/actalog/main.go
git commit -m "feat(routes): wire POST /api/admin/users for admin user creation

Outside the {id} sub-router because create has no {id} and the
ProtectedUserGuard targets existing-row writes only. Service layer
handles protected-email rejection at create."
```

---

## Task 5: AdminUserService.SetPassword — service method

**Files:**
- Modify: `internal/service/admin_user_service.go`
- Modify: `internal/service/admin_user_service_test.go`

- [ ] **Step 1: Write failing tests**

Append to `internal/service/admin_user_service_test.go`:

```go
// TestAdminUserService_SetPassword_HappyPath verifies password is hashed +
// stored, lockout state cleared, refresh tokens revoked, and audit fired.
func TestAdminUserService_SetPassword_HappyPath(t *testing.T) {
	repo := newStubUserRepo()
	target := makeNormalUser(5, "lockedout@example.com", "Locked")
	target.FailedLoginAttempts = 3
	lockUntil := time.Now().Add(15 * time.Minute)
	target.LockedUntil = &lockUntil
	repo.addUser(target)

	tokenRepo := &stubRefreshTokenRepo{}
	auditLog := &stubAuditLogger{}
	svc := newService(repo, tokenRepo, &stubEmailSvc{}, auditLog)

	if err := svc.SetPassword(1, 5, "NewValidPass123A"); err != nil {
		t.Fatalf("SetPassword: %v", err)
	}

	// Password was hashed (not stored as plaintext).
	stored := repo.users[5]
	if stored.PasswordHash == "" {
		t.Error("PasswordHash empty after SetPassword")
	}
	if stored.PasswordHash == "NewValidPass123A" {
		t.Error("PasswordHash equals plaintext — not hashed")
	}

	// Lockout cleared.
	if stored.FailedLoginAttempts != 0 {
		t.Errorf("FailedLoginAttempts = %d, want 0", stored.FailedLoginAttempts)
	}
	if stored.LockedUntil != nil {
		t.Error("LockedUntil should be nil after SetPassword")
	}

	// Refresh tokens revoked.
	if !tokenRepo.revokeAllForUserCalled {
		t.Error("RevokeAllForUser not called")
	}
	if tokenRepo.revokeAllForUserID != 5 {
		t.Errorf("revoke for userID = %d, want 5", tokenRepo.revokeAllForUserID)
	}

	// Audit event with priors captured.
	if len(auditLog.events) != 1 {
		t.Fatalf("expected 1 audit event, got %d", len(auditLog.events))
	}
	ev := auditLog.events[0]
	if ev.eventType != domain.EventAdminPasswordSet {
		t.Errorf("eventType = %q, want %q", ev.eventType, domain.EventAdminPasswordSet)
	}
	if ev.details["cleared_failed_login_attempts"] != 3 {
		t.Errorf("details.cleared_failed_login_attempts = %v, want 3", ev.details["cleared_failed_login_attempts"])
	}
	if ev.details["cleared_lockout"] != true {
		t.Errorf("details.cleared_lockout = %v, want true", ev.details["cleared_lockout"])
	}
}

// TestAdminUserService_SetPassword_404OnUnknownUser surfaces the user-not-found path.
func TestAdminUserService_SetPassword_404OnUnknownUser(t *testing.T) {
	repo := newStubUserRepo()
	svc := newService(repo, &stubRefreshTokenRepo{}, &stubEmailSvc{}, &stubAuditLogger{})

	err := svc.SetPassword(1, 99, "NewValidPass123A")
	if err == nil {
		t.Fatal("expected ErrUserNotFound, got nil")
	}
	if !errors.Is(err, ErrUserNotFound) {
		t.Errorf("expected ErrUserNotFound, got: %v", err)
	}
}

// TestAdminUserService_SetPassword_RejectsWeakPassword exercises the policy.
func TestAdminUserService_SetPassword_RejectsWeakPassword(t *testing.T) {
	repo := newStubUserRepo()
	target := makeNormalUser(5, "u@example.com", "U")
	repo.addUser(target)
	svc := newService(repo, &stubRefreshTokenRepo{}, &stubEmailSvc{}, &stubAuditLogger{})

	err := svc.SetPassword(1, 5, "tooshort")
	if err == nil {
		t.Fatal("expected validation error, got nil")
	}
	var invErr *domain.InvalidInputError
	if !errors.As(err, &invErr) {
		t.Fatalf("expected *domain.InvalidInputError, got %T", err)
	}
	if invErr.Field != "new_password" {
		t.Errorf("Field = %q, want %q", invErr.Field, "new_password")
	}
}

// TestAdminUserService_SetPassword_NoLockoutPriorState verifies the audit
// captures cleared_lockout=false when the user wasn't locked.
func TestAdminUserService_SetPassword_NoLockoutPriorState(t *testing.T) {
	repo := newStubUserRepo()
	target := makeNormalUser(5, "calm@example.com", "Calm")
	target.FailedLoginAttempts = 0
	target.LockedUntil = nil
	repo.addUser(target)

	auditLog := &stubAuditLogger{}
	svc := newService(repo, &stubRefreshTokenRepo{}, &stubEmailSvc{}, auditLog)

	if err := svc.SetPassword(1, 5, "NewValidPass123A"); err != nil {
		t.Fatalf("SetPassword: %v", err)
	}
	if auditLog.events[0].details["cleared_lockout"] != false {
		t.Errorf("cleared_lockout should be false; got %v", auditLog.events[0].details["cleared_lockout"])
	}
	if auditLog.events[0].details["cleared_failed_login_attempts"] != 0 {
		t.Errorf("cleared_failed_login_attempts should be 0; got %v", auditLog.events[0].details["cleared_failed_login_attempts"])
	}
}
```

- [ ] **Step 2: Extend stubs to record RevokeAllForUser**

Find `stubRefreshTokenRepo` in `internal/service/admin_user_service_test.go`. Add tracking fields:

```go
type stubRefreshTokenRepo struct {
	// existing fields...
	revokeAllForUserCalled bool
	revokeAllForUserID     int64
}

func (s *stubRefreshTokenRepo) RevokeAllForUser(userID int64) error {
	s.revokeAllForUserCalled = true
	s.revokeAllForUserID = userID
	return nil
}
```

If the method is already there, just add the tracking fields and update the existing method.

- [ ] **Step 3: Run tests to verify they fail**

```bash
go test ./internal/service/ -run "TestAdminUserService_SetPassword" -count=1 -v
```

Expected: FAIL — `svc.SetPassword undefined`.

- [ ] **Step 4: Add the method to admin_user_service.go**

Append to `internal/service/admin_user_service.go` (near `ForcePasswordReset`):

```go
// SetPassword sets a specific password on a target user account from admin
// context. Bundles the recovery flow that admins typically need together:
//
//   - Hash + store the new password
//   - Clear failed_login_attempts and any active lockout
//   - Revoke all refresh tokens (force re-auth on existing devices)
//   - Audit the operation with prior counter/lockout state for forensics
//
// Does NOT touch account_disabled, email_verified, or any role/identity
// field — those are separate admin concerns and have their own endpoints.
//
// Returns *domain.InvalidInputError if the password fails the complexity
// policy, ErrUserNotFound if the target doesn't exist.
func (s *AdminUserService) SetPassword(actorID, targetID int64, newPassword string) error {
	// 1. Validate password complexity FIRST so we never load the user when
	//    we know the input is bad. Return the typed error so the handler
	//    surfaces 400 with a useful message.
	if err := validatePassword(newPassword); err != nil {
		return &domain.InvalidInputError{Field: "new_password", Message: err.Error(), Cause: err}
	}

	// 2. Load target to capture prior state for audit.
	target, err := s.userRepo.GetByID(targetID)
	if err != nil {
		return fmt.Errorf("SetPassword: GetByID(%d): %w", targetID, err)
	}
	if target == nil {
		return fmt.Errorf("SetPassword: user %d: %w", targetID, ErrUserNotFound)
	}
	priorAttempts := target.FailedLoginAttempts
	priorLocked := target.LockedUntil != nil

	// 3. Hash and persist the new password.
	hash, hashErr := auth.HashPassword(newPassword)
	if hashErr != nil {
		return fmt.Errorf("SetPassword: hash: %w", hashErr)
	}
	if updErr := s.userRepo.UpdatePassword(targetID, hash); updErr != nil {
		return fmt.Errorf("SetPassword: UpdatePassword: %w", updErr)
	}

	// 4. Clear lockout state. UnlockAccount sets failed_login_attempts=0,
	//    locked_at=NULL, locked_until=NULL atomically.
	if unlockErr := s.userRepo.UnlockAccount(targetID); unlockErr != nil {
		return fmt.Errorf("SetPassword: UnlockAccount: %w", unlockErr)
	}

	// 5. Revoke refresh tokens. A failure here is a partial-state warning —
	//    the password was set, but existing sessions weren't invalidated.
	//    Log + record in audit detail; the operation is still considered
	//    successful from the admin's perspective.
	revokeOK := true
	if revokeErr := s.refreshTokenRepo.RevokeAllForUser(targetID); revokeErr != nil {
		s.log.Warn("SetPassword: RevokeAllForUser(%d): %v", targetID, revokeErr)
		revokeOK = false
	}

	// 6. Audit.
	if auditErr := s.auditLogService.LogEvent(
		domain.EventAdminPasswordSet,
		&actorID,
		&targetID,
		nil, nil,
		map[string]interface{}{
			"cleared_failed_login_attempts": priorAttempts,
			"cleared_lockout":               priorLocked,
			"revoked_refresh_tokens":        revokeOK,
		},
	); auditErr != nil {
		s.log.Error("AdminUserService.SetPassword: audit: %v", auditErr)
	}

	return nil
}
```

If the logger doesn't have a `Warn` method, use `Error` (search the existing codebase for the pattern). Verify `s.log.Warn` or `.Error` is callable.

- [ ] **Step 5: Run tests, verify they pass**

```bash
go test ./internal/service/ -run "TestAdminUserService_SetPassword" -count=1 -v
```

Expected: 4 PASS.

- [ ] **Step 6: Commit**

```bash
git add internal/service/admin_user_service.go internal/service/admin_user_service_test.go
git commit -m "feat(service): AdminUserService.SetPassword

Admin sets a specific password directly. Bundles the credential-recovery
flow: hash + store password, clear failed_login_attempts and lockout
(via UnlockAccount), revoke refresh tokens. Audit captures prior counter
+ lockout state for forensics. Validates complexity via the existing
validatePassword helper. account_disabled / email_verified / role are
NOT touched."
```

---

## Task 6: AdminUserHandler.SetPassword — HTTP handler

**Files:**
- Modify: `internal/handler/admin_user_handler.go`
- Modify: `internal/handler/admin_user_handler_test.go`

- [ ] **Step 1: Write failing tests**

Append to `internal/handler/admin_user_handler_test.go`:

```go
// TestAdminUserHandler_SetPassword_204Success returns no body on success.
func TestAdminUserHandler_SetPassword_204Success(t *testing.T) {
	stub := &stubAdminUserService{}
	h := NewAdminUserHandler(nil, stub, nil)

	req := httptest.NewRequest(http.MethodPost, "/api/admin/users/42/password",
		strings.NewReader(`{"new_password":"ValidPass123Long"}`))
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "42")
	ctx := middleware.WithUserID(req.Context(), 1)
	ctx = context.WithValue(ctx, chi.RouteCtxKey, rctx)
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	h.SetPassword(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Errorf("Code = %d, want 204; body=%s", rr.Code, rr.Body.String())
	}
	if rr.Body.Len() != 0 {
		t.Errorf("body should be empty for 204; got %s", rr.Body.String())
	}
	if !stub.setPasswordCalled {
		t.Error("service.SetPassword not invoked")
	}
	if stub.setPasswordTargetID != 42 {
		t.Errorf("targetID = %d, want 42", stub.setPasswordTargetID)
	}
}

// TestAdminUserHandler_SetPassword_400OnInvalidJSON
func TestAdminUserHandler_SetPassword_400OnInvalidJSON(t *testing.T) {
	stub := &stubAdminUserService{}
	h := NewAdminUserHandler(nil, stub, nil)

	req := httptest.NewRequest(http.MethodPost, "/api/admin/users/42/password",
		strings.NewReader(`{not json`))
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "42")
	ctx := middleware.WithUserID(req.Context(), 1)
	ctx = context.WithValue(ctx, chi.RouteCtxKey, rctx)
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	h.SetPassword(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("Code = %d, want 400", rr.Code)
	}
}

// TestAdminUserHandler_SetPassword_404OnInvalidID
func TestAdminUserHandler_SetPassword_404OnInvalidID(t *testing.T) {
	stub := &stubAdminUserService{}
	h := NewAdminUserHandler(nil, stub, nil)

	req := httptest.NewRequest(http.MethodPost, "/api/admin/users/notanint/password",
		strings.NewReader(`{"new_password":"ValidPass123Long"}`))
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "notanint")
	ctx := middleware.WithUserID(req.Context(), 1)
	ctx = context.WithValue(ctx, chi.RouteCtxKey, rctx)
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	h.SetPassword(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("Code = %d, want 404", rr.Code)
	}
}

// TestAdminUserHandler_SetPassword_400OnWeakPassword surfaces InvalidInputError.
func TestAdminUserHandler_SetPassword_400OnWeakPassword(t *testing.T) {
	stub := &stubAdminUserService{
		setPasswordErr: &domain.InvalidInputError{Field: "new_password", Message: "too short"},
	}
	h := NewAdminUserHandler(nil, stub, nil)

	req := httptest.NewRequest(http.MethodPost, "/api/admin/users/42/password",
		strings.NewReader(`{"new_password":"x"}`))
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "42")
	ctx := middleware.WithUserID(req.Context(), 1)
	ctx = context.WithValue(ctx, chi.RouteCtxKey, rctx)
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	h.SetPassword(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("Code = %d, want 400", rr.Code)
	}
	if !strings.Contains(rr.Body.String(), "new_password") {
		t.Errorf("body should include field name; got %s", rr.Body.String())
	}
}
```

- [ ] **Step 2: Extend stub service**

In the `stubAdminUserService` struct, add:

```go
setPasswordCalled    bool
setPasswordActorID   int64
setPasswordTargetID  int64
setPasswordPassword  string
setPasswordErr       error
```

And the method:

```go
func (s *stubAdminUserService) SetPassword(actorID, targetID int64, newPassword string) error {
	s.setPasswordCalled = true
	s.setPasswordActorID = actorID
	s.setPasswordTargetID = targetID
	s.setPasswordPassword = newPassword
	return s.setPasswordErr
}
```

If the handler talks to the service via an interface, add `SetPassword(actorID, targetID int64, newPassword string) error` to that interface.

- [ ] **Step 3: Run tests to confirm failure**

```bash
go test ./internal/handler/ -run "TestAdminUserHandler_SetPassword" -count=1 -v
```

Expected: FAIL — `h.SetPassword undefined`.

- [ ] **Step 4: Add the handler**

Append to `internal/handler/admin_user_handler.go`:

```go
// setPasswordRequest is the POST /api/admin/users/{id}/password request body.
type setPasswordRequest struct {
	NewPassword string `json:"new_password"`
}

// SetPassword handles POST /api/admin/users/{id}/password — admin sets a
// specific password on a user account.
//
//   - 204 on success (empty body)
//   - 400 on invalid JSON or password-policy failure
//   - 403 from L1 ProtectedUserGuard if target is protected
//   - 404 if {id} is not a valid integer or user doesn't exist
func (h *AdminUserHandler) SetPassword(w http.ResponseWriter, r *http.Request) {
	actorID, ok := middleware.GetUserID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}
	targetID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		respondError(w, http.StatusNotFound, "Not found")
		return
	}
	var req setPasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid JSON")
		return
	}
	if err := h.adminUserService.SetPassword(actorID, targetID, req.NewPassword); err != nil {
		WriteError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
```

- [ ] **Step 5: Run tests, verify they pass**

```bash
go test ./internal/handler/ -run "TestAdminUserHandler_SetPassword" -count=1 -v
```

Expected: 4 PASS.

- [ ] **Step 6: Commit**

```bash
git add internal/handler/admin_user_handler.go internal/handler/admin_user_handler_test.go
git commit -m "feat(handler): AdminUserHandler.SetPassword

POST /api/admin/users/{id}/password — empty 204 body on success, 400 on
weak password, 404 on unknown user. Delegates to service which performs
the lockout-clear + token-revoke bundle."
```

---

## Task 7: Wire POST /api/admin/users/{id}/password route

**Files:**
- Modify: `cmd/actalog/main.go`

- [ ] **Step 1: Add the route inside the protected-user sub-router**

In `cmd/actalog/main.go`, find the `r.Route("/users/{id}", func(r chi.Router) {` block (around line 1000). After `r.Post("/force-password-reset", adminUserHandler.ForcePasswordReset)` (line 1025), add:

```go
r.Post("/password", adminUserHandler.SetPassword) // v1.3.1: admin sets password directly
```

Inside-the-subrouter placement gives us `ProtectedUserGuard` (L1) for free — protected users get 403 before the handler runs.

- [ ] **Step 2: Build to verify**

```bash
go build ./cmd/actalog/
```

Expected: no output, exit 0.

- [ ] **Step 3: Commit**

```bash
git add cmd/actalog/main.go
git commit -m "feat(routes): wire POST /api/admin/users/{id}/password

Inside the protected-user sub-router so L1 ProtectedUserGuard auto-blocks
attempts against protected accounts."
```

---

## Task 8: Backend integration tests (multi-DB)

**Files:**
- Create: `test/integration/admin_user_lifecycle_test.go`

- [ ] **Step 1: Write the integration test file**

Create `test/integration/admin_user_lifecycle_test.go`:

```go
package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/johnzastrow/actalog/internal/domain"
	"github.com/johnzastrow/actalog/pkg/security"
)

// TestAdminCreateUser_ThenLogin verifies the round trip: admin creates a user
// via POST /api/admin/users, the new user can then POST /api/auth/login with
// the admin-set credentials.
func TestAdminCreateUser_ThenLogin(t *testing.T) {
	srv := mustStartTestServer(t)
	defer srv.Close()

	adminToken := mustRegisterAndGetToken(t, srv, "adm@example.com", "AdminPass123Long", "Adm")

	// Create a user via the admin endpoint.
	body, _ := json.Marshal(map[string]any{
		"email":          "newathlete@example.com",
		"password":       "NewAthletePass123",
		"name":           "New Athlete",
		"role":           "athlete",
		"email_verified": true,
	})
	req, _ := http.NewRequest(http.MethodPost, srv.URL+"/api/admin/users", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+adminToken)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("POST /api/admin/users: %v", err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("create status = %d, want 201", resp.StatusCode)
	}
	resp.Body.Close()

	// New user logs in.
	loginBody, _ := json.Marshal(map[string]string{
		"email":    "newathlete@example.com",
		"password": "NewAthletePass123",
	})
	loginResp, err := http.Post(srv.URL+"/api/auth/login", "application/json", bytes.NewReader(loginBody))
	if err != nil {
		t.Fatalf("login: %v", err)
	}
	if loginResp.StatusCode != http.StatusOK {
		t.Fatalf("login status = %d, want 200", loginResp.StatusCode)
	}
	loginResp.Body.Close()
}

// TestAdminCreateUser_ProtectedEmailRejected covers the rejection path with
// real HTTP machinery (not just the unit test stub).
func TestAdminCreateUser_ProtectedEmailRejected(t *testing.T) {
	srv := mustStartTestServer(t)
	defer srv.Close()

	adminToken := mustRegisterAndGetToken(t, srv, "adm2@example.com", "AdminPass123Long", "Adm")

	protectedEmail := security.ProtectedEmailsList()[0]
	body, _ := json.Marshal(map[string]any{
		"email":    protectedEmail,
		"password": "ValidPass123Long",
		"name":     "Imp",
		"role":     "athlete",
	})
	req, _ := http.NewRequest(http.MethodPost, srv.URL+"/api/admin/users", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+adminToken)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("POST: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", resp.StatusCode)
	}
	resp.Body.Close()
}

// TestAdminSetPassword_ClearsLockoutAndAllowsLogin: admin sets password on a
// locked-out user; verify the user can then log in with the new password and
// failed_login_attempts/locked_until were cleared.
func TestAdminSetPassword_ClearsLockoutAndAllowsLogin(t *testing.T) {
	srv := mustStartTestServer(t)
	defer srv.Close()
	db, driver := mustOpenTestDB(t)

	adminToken := mustRegisterAndGetToken(t, srv, "admin3@example.com", "AdminPass123Long", "A3")

	// Create the target via admin endpoint, then artificially lock them at
	// the DB level (cheaper than triggering the lockout via failed logins).
	createBody, _ := json.Marshal(map[string]any{
		"email":          "victim@example.com",
		"password":       "OriginalPass123A",
		"name":           "Victim",
		"role":           "athlete",
		"email_verified": true,
	})
	req, _ := http.NewRequest(http.MethodPost, srv.URL+"/api/admin/users", bytes.NewReader(createBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+adminToken)
	cresp, err := http.DefaultClient.Do(req)
	if err != nil || cresp.StatusCode != http.StatusCreated {
		t.Fatalf("create: %v / %d", err, cresp.StatusCode)
	}
	var createdUser domain.User
	json.NewDecoder(cresp.Body).Decode(&createdUser)
	cresp.Body.Close()

	// Lock the target directly in DB.
	lockSQL := "UPDATE users SET failed_login_attempts = 5, locked_at = ?, locked_until = ? WHERE id = ?"
	if driver == "postgres" {
		lockSQL = "UPDATE users SET failed_login_attempts = 5, locked_at = $1, locked_until = $2 WHERE id = $3"
	}
	if _, err := db.Exec(lockSQL, "2026-01-01T00:00:00Z", "2099-01-01T00:00:00Z", createdUser.ID); err != nil {
		t.Fatalf("lock target: %v", err)
	}

	// Admin sets a new password.
	pwBody, _ := json.Marshal(map[string]string{"new_password": "AdminFixedPass45A"})
	pwReq, _ := http.NewRequest(http.MethodPost, srv.URL+"/api/admin/users/"+strconv.FormatInt(createdUser.ID, 10)+"/password", bytes.NewReader(pwBody))
	pwReq.Header.Set("Content-Type", "application/json")
	pwReq.Header.Set("Authorization", "Bearer "+adminToken)
	pwResp, err := http.DefaultClient.Do(pwReq)
	if err != nil {
		t.Fatalf("set-password: %v", err)
	}
	if pwResp.StatusCode != http.StatusNoContent {
		t.Fatalf("set-password status = %d, want 204", pwResp.StatusCode)
	}
	pwResp.Body.Close()

	// Verify lockout cleared in DB.
	var attempts int
	var lockedUntil *string
	q := "SELECT failed_login_attempts, locked_until FROM users WHERE id = ?"
	if driver == "postgres" {
		q = "SELECT failed_login_attempts, locked_until FROM users WHERE id = $1"
	}
	if err := db.QueryRow(q, createdUser.ID).Scan(&attempts, &lockedUntil); err != nil {
		t.Fatalf("read state: %v", err)
	}
	if attempts != 0 {
		t.Errorf("failed_login_attempts = %d, want 0", attempts)
	}
	if lockedUntil != nil {
		t.Errorf("locked_until should be NULL; got %v", lockedUntil)
	}

	// Log in with the new password.
	loginBody, _ := json.Marshal(map[string]string{
		"email":    "victim@example.com",
		"password": "AdminFixedPass45A",
	})
	loginResp, err := http.Post(srv.URL+"/api/auth/login", "application/json", bytes.NewReader(loginBody))
	if err != nil || loginResp.StatusCode != http.StatusOK {
		t.Fatalf("login with new password: err=%v status=%d", err, loginResp.StatusCode)
	}
	loginResp.Body.Close()
}

```

> This test depends on helpers `mustStartTestServer` and `mustRegisterAndGetToken`. If they don't exist in `test/integration/`, look for similar bootstrap helpers in `test/integration/api_test.go` and reuse the existing pattern. If new helpers are needed, place them in `test/integration/` next to existing ones — don't put them in this new test file. The existing `mustOpenTestDB` helper (used by the v1.3.0 protected-user tests) is reusable.

- [ ] **Step 2: Run on sqlite3 (default for local)**

```bash
go test ./test/integration/ -run "TestAdminCreateUser|TestAdminSetPassword" -count=1 -v
```

Expected: 3 PASS.

- [ ] **Step 3: Commit**

```bash
git add test/integration/admin_user_lifecycle_test.go
git commit -m "test(integration): admin user lifecycle end-to-end

Three scenarios — create-then-login, protected-email rejection at
create, set-password-clears-lockout-and-allows-login. CI matrix runs
these against sqlite3, postgres, and mysql."
```

---

## Task 9: usePasswordInputs composable

**Files:**
- Create: `web/src/components/admin/composables/usePasswordInputs.js`
- Create: `web/src/components/admin/composables/usePasswordInputs.test.js`

- [ ] **Step 1: Write failing tests**

Create `web/src/components/admin/composables/usePasswordInputs.test.js`:

```js
import { describe, it, expect } from 'vitest'
import { usePasswordInputs } from './usePasswordInputs.js'

describe('usePasswordInputs', () => {
  it('starts with empty values and not visible', () => {
    const pw = usePasswordInputs()
    expect(pw.password.value).toBe('')
    expect(pw.confirm.value).toBe('')
    expect(pw.visible.value).toBe(false)
  })

  it('toggleVisible flips the boolean', () => {
    const pw = usePasswordInputs()
    pw.toggleVisible()
    expect(pw.visible.value).toBe(true)
    pw.toggleVisible()
    expect(pw.visible.value).toBe(false)
  })

  it('isValid is false when empty', () => {
    const pw = usePasswordInputs()
    expect(pw.isValid.value).toBe(false)
  })

  it('isValid is false when confirm does not match', () => {
    const pw = usePasswordInputs()
    pw.password.value = 'ValidPass123Long'
    pw.confirm.value = 'DifferentPass1Long'
    expect(pw.isValid.value).toBe(false)
    expect(pw.errorMessage.value).toMatch(/match/i)
  })

  it('isValid is false when password fails complexity', () => {
    const pw = usePasswordInputs()
    pw.password.value = 'short'
    pw.confirm.value = 'short'
    expect(pw.isValid.value).toBe(false)
    expect(pw.errorMessage.value).toMatch(/12|character|upper|lower|digit/i)
  })

  it('isValid is true with matching valid password', () => {
    const pw = usePasswordInputs()
    pw.password.value = 'ValidPass123Long'
    pw.confirm.value = 'ValidPass123Long'
    expect(pw.isValid.value).toBe(true)
    expect(pw.errorMessage.value).toBe('')
  })

  it('reset clears all state', () => {
    const pw = usePasswordInputs()
    pw.password.value = 'ValidPass123Long'
    pw.confirm.value = 'ValidPass123Long'
    pw.visible.value = true
    pw.reset()
    expect(pw.password.value).toBe('')
    expect(pw.confirm.value).toBe('')
    expect(pw.visible.value).toBe(false)
  })
})
```

- [ ] **Step 2: Run tests to verify failure**

```bash
cd web && npm run test:run -- usePasswordInputs
```

Expected: FAIL — `Cannot find module 'usePasswordInputs.js'`.

- [ ] **Step 3: Implement the composable**

Create `web/src/components/admin/composables/usePasswordInputs.js`:

```js
import { ref, computed } from 'vue'

/**
 * usePasswordInputs — shared password+confirm state for admin dialogs.
 *
 * Returns reactive refs and computeds that wire directly into Vuetify
 * v-text-field bindings:
 *
 *   const pw = usePasswordInputs()
 *   <v-text-field v-model="pw.password.value" :type="pw.visible.value ? 'text' : 'password'" />
 *   <v-text-field v-model="pw.confirm.value" :type="pw.visible.value ? 'text' : 'password'" />
 *   <v-btn @click="pw.toggleVisible">{{ pw.visible.value ? 'Hide' : 'Show' }}</v-btn>
 *
 * Validation policy mirrors the backend (pkg/auth + internal/service/user_service.go):
 *   - ≥12 characters
 *   - at least one uppercase
 *   - at least one lowercase
 *   - at least one digit
 *
 * The backend remains the source of truth — this is just a UX hint to
 * disable the submit button before a doomed request goes out.
 */
export function usePasswordInputs() {
  const password = ref('')
  const confirm = ref('')
  const visible = ref(false)

  const errorMessage = computed(() => {
    if (!password.value && !confirm.value) return ''
    if (password.value.length < 12) {
      return 'Password must be at least 12 characters.'
    }
    const hasUpper = /[A-Z]/.test(password.value)
    const hasLower = /[a-z]/.test(password.value)
    const hasDigit = /[0-9]/.test(password.value)
    if (!hasUpper || !hasLower || !hasDigit) {
      return 'Password must include uppercase, lowercase, and a digit.'
    }
    if (password.value !== confirm.value) {
      return 'Passwords do not match.'
    }
    return ''
  })

  const isValid = computed(() => {
    return password.value.length > 0 && errorMessage.value === ''
  })

  function toggleVisible() {
    visible.value = !visible.value
  }

  function reset() {
    password.value = ''
    confirm.value = ''
    visible.value = false
  }

  return { password, confirm, visible, errorMessage, isValid, toggleVisible, reset }
}
```

- [ ] **Step 4: Run tests, verify they pass**

```bash
cd web && npm run test:run -- usePasswordInputs
```

Expected: 7 PASS.

- [ ] **Step 5: Commit**

```bash
cd /home/jcz/Github/actionlog/.worktrees/feature/admin-user-lifecycle-v1.3.1
git add web/src/components/admin/composables/usePasswordInputs.js web/src/components/admin/composables/usePasswordInputs.test.js
git commit -m "feat(web): usePasswordInputs composable for admin dialogs

Shared password+confirm state with show/hide toggle, complexity hint,
matching-validation, and reset. Mirrors backend policy as a UX hint;
backend remains source of truth."
```

---

## Task 10: AdminUserCreateDialog component

**Files:**
- Create: `web/src/components/admin/AdminUserCreateDialog.vue`
- Create: `web/src/components/admin/AdminUserCreateDialog.test.js`

- [ ] **Step 1: Write failing tests**

Create `web/src/components/admin/AdminUserCreateDialog.test.js`:

```js
import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount } from '@vue/test-utils'
import AdminUserCreateDialog from './AdminUserCreateDialog.vue'
import { createVuetify } from 'vuetify'
import * as components from 'vuetify/components'
import * as directives from 'vuetify/directives'

const vuetify = createVuetify({ components, directives })

vi.mock('@/utils/axios', () => ({
  default: {
    post: vi.fn()
  }
}))
import axios from '@/utils/axios'

function mountDialog(props = {}) {
  return mount(AdminUserCreateDialog, {
    global: { plugins: [vuetify] },
    props: { modelValue: true, ...props },
    attachTo: document.body,
  })
}

describe('AdminUserCreateDialog', () => {
  beforeEach(() => {
    axios.post.mockReset()
  })

  it('renders all required fields', () => {
    const wrapper = mountDialog()
    const html = document.body.innerHTML
    expect(html).toContain('Create User')
    // Email, password, confirm, name, role, email_verified
    expect(document.querySelectorAll('input[type="email"], input[type="password"], input[type="text"]').length).toBeGreaterThanOrEqual(4)
    wrapper.unmount()
  })

  it('disables submit until all fields valid', async () => {
    const wrapper = mountDialog()
    // Without input, the Create button should be disabled.
    const createBtn = Array.from(document.querySelectorAll('button')).find(b => b.textContent.includes('Create'))
    expect(createBtn.disabled).toBe(true)
    wrapper.unmount()
  })

  it('POSTs to /api/admin/users on submit and emits created on 201', async () => {
    axios.post.mockResolvedValueOnce({
      status: 201,
      data: { id: 5, email: 'new@example.com', name: 'New', role: 'athlete' },
    })
    const wrapper = mountDialog()

    // Fill the form via the component's exposed setters (or directly via DOM input).
    // Simplest: drive via wrapper.vm internals.
    wrapper.vm.email = 'new@example.com'
    wrapper.vm.name = 'New'
    wrapper.vm.role = 'athlete'
    wrapper.vm.emailVerified = true
    wrapper.vm.pw.password.value = 'ValidPass123Long'
    wrapper.vm.pw.confirm.value = 'ValidPass123Long'
    await wrapper.vm.$nextTick()
    await wrapper.vm.submit()

    expect(axios.post).toHaveBeenCalledWith('/api/admin/users', expect.objectContaining({
      email: 'new@example.com',
      name: 'New',
      role: 'athlete',
      email_verified: true,
      password: 'ValidPass123Long',
    }))
    expect(wrapper.emitted('created')).toBeTruthy()
    expect(wrapper.emitted('created')[0][0]).toMatchObject({ id: 5 })
    wrapper.unmount()
  })

  it('shows server error message on 400 invalid_input', async () => {
    axios.post.mockRejectedValueOnce({
      response: { status: 400, data: { error: 'invalid_input', message: 'email: is reserved' } },
    })
    const wrapper = mountDialog()
    wrapper.vm.email = 'br8kwall@gmail.com'
    wrapper.vm.name = 'X'
    wrapper.vm.role = 'athlete'
    wrapper.vm.pw.password.value = 'ValidPass123Long'
    wrapper.vm.pw.confirm.value = 'ValidPass123Long'
    await wrapper.vm.submit()

    expect(wrapper.vm.errorMessage).toContain('reserved')
    wrapper.unmount()
  })

  it('shows duplicate-email error on 409', async () => {
    axios.post.mockRejectedValueOnce({
      response: { status: 409, data: { error: 'duplicate_email', message: 'A user with that email already exists.' } },
    })
    const wrapper = mountDialog()
    wrapper.vm.email = 'taken@example.com'
    wrapper.vm.name = 'X'
    wrapper.vm.role = 'athlete'
    wrapper.vm.pw.password.value = 'ValidPass123Long'
    wrapper.vm.pw.confirm.value = 'ValidPass123Long'
    await wrapper.vm.submit()

    expect(wrapper.vm.errorMessage).toMatch(/exists/i)
    wrapper.unmount()
  })
})
```

- [ ] **Step 2: Run tests to verify failure**

```bash
cd web && npm run test:run -- AdminUserCreateDialog
```

Expected: FAIL — module not found.

- [ ] **Step 3: Implement the component**

Create `web/src/components/admin/AdminUserCreateDialog.vue`:

```vue
<template>
  <v-dialog
    :model-value="modelValue"
    @update:model-value="(v) => $emit('update:modelValue', v)"
    max-width="560"
    persistent
  >
    <v-card>
      <v-card-title class="text-h6">Create User</v-card-title>
      <v-card-text>
        <v-form @submit.prevent="submit">
          <v-text-field
            v-model="email"
            label="Email"
            type="email"
            autofocus
            required
          />
          <v-text-field
            v-model="pw.password.value"
            label="Password"
            :type="pw.visible.value ? 'text' : 'password'"
            :append-inner-icon="pw.visible.value ? 'mdi-eye-off' : 'mdi-eye'"
            @click:append-inner="pw.toggleVisible"
            hint="Min 12 chars; uppercase, lowercase, digit."
            persistent-hint
            required
          />
          <v-text-field
            v-model="pw.confirm.value"
            label="Confirm password"
            :type="pw.visible.value ? 'text' : 'password'"
            :error-messages="pw.errorMessage.value"
            required
          />
          <v-text-field
            v-model="name"
            label="Name"
            required
          />
          <v-select
            v-model="role"
            label="Role"
            :items="['athlete', 'coach', 'admin']"
            required
          />
          <v-checkbox
            v-model="emailVerified"
            label="Mark email as verified (admin-vouched)"
          />

          <v-alert v-if="errorMessage" type="error" density="compact" class="mb-2">
            {{ errorMessage }}
          </v-alert>
        </v-form>
      </v-card-text>
      <v-card-actions>
        <v-spacer />
        <v-btn variant="text" @click="cancel">Cancel</v-btn>
        <v-btn
          color="primary"
          :disabled="!canSubmit || submitting"
          :loading="submitting"
          @click="submit"
        >
          Create
        </v-btn>
      </v-card-actions>
    </v-card>
  </v-dialog>
</template>

<script setup>
import { ref, computed, watch } from 'vue'
import axios from '@/utils/axios'
import { usePasswordInputs } from '@/components/admin/composables/usePasswordInputs.js'

const props = defineProps({
  modelValue: { type: Boolean, required: true },
})
const emit = defineEmits(['update:modelValue', 'created'])

const email = ref('')
const name = ref('')
const role = ref('athlete')
const emailVerified = ref(true)
const pw = usePasswordInputs()
const submitting = ref(false)
const errorMessage = ref('')

const canSubmit = computed(() => {
  return email.value.length > 0 &&
    name.value.trim().length > 0 &&
    pw.isValid.value
})

async function submit() {
  errorMessage.value = ''
  submitting.value = true
  try {
    const res = await axios.post('/api/admin/users', {
      email: email.value.trim(),
      password: pw.password.value,
      name: name.value.trim(),
      role: role.value,
      email_verified: emailVerified.value,
    })
    emit('created', res.data)
    closeAndReset()
  } catch (e) {
    if (e?.response?.data?.message) {
      errorMessage.value = e.response.data.message
    } else {
      errorMessage.value = 'Could not create user. Check server logs.'
    }
  } finally {
    submitting.value = false
  }
}

function cancel() {
  closeAndReset()
}

function closeAndReset() {
  email.value = ''
  name.value = ''
  role.value = 'athlete'
  emailVerified.value = true
  pw.reset()
  errorMessage.value = ''
  emit('update:modelValue', false)
}

// Reset state any time the dialog reopens (in case it was closed mid-edit).
watch(() => props.modelValue, (v) => {
  if (v) {
    errorMessage.value = ''
  }
})

defineExpose({ email, name, role, emailVerified, pw, submit, errorMessage })
</script>
```

- [ ] **Step 4: Run tests, verify they pass**

```bash
cd web && npm run test:run -- AdminUserCreateDialog
```

Expected: 5 PASS.

- [ ] **Step 5: Commit**

```bash
cd /home/jcz/Github/actionlog/.worktrees/feature/admin-user-lifecycle-v1.3.1
git add web/src/components/admin/AdminUserCreateDialog.vue web/src/components/admin/AdminUserCreateDialog.test.js
git commit -m "feat(web): AdminUserCreateDialog component

Modal dialog for admin user creation. Reuses usePasswordInputs for the
password+confirm pair. Surfaces server validation errors inline. Resets
state on close so reopening starts clean."
```

---

## Task 11: Wire Create button into AdminUsersView

**Files:**
- Modify: `web/src/views/AdminUsersView.vue`
- Modify: `web/src/views/AdminUsersView.test.js`

- [ ] **Step 1: Add the failing test**

Append to `web/src/views/AdminUsersView.test.js`:

```js
describe('AdminUsersView Create User button', () => {
  it('renders a Create User button visible to admin', async () => {
    // mount AdminUsersView with admin auth state
    const wrapper = mountAdminUsersView({ isAdmin: true })
    await wrapper.vm.$nextTick()
    const createBtn = wrapper.find('[data-test="create-user-button"]')
    expect(createBtn.exists()).toBe(true)
    expect(createBtn.text()).toContain('Create User')
    wrapper.unmount()
  })

  it('opens the create dialog when clicked', async () => {
    const wrapper = mountAdminUsersView({ isAdmin: true })
    await wrapper.find('[data-test="create-user-button"]').trigger('click')
    await wrapper.vm.$nextTick()
    expect(wrapper.vm.showCreateDialog).toBe(true)
    wrapper.unmount()
  })
})
```

> If a `mountAdminUsersView` helper doesn't exist in the test file, look for the existing mount pattern at the top — the v1.3.0 work added tests for this view. Reuse the pattern.

- [ ] **Step 2: Run tests to confirm failure**

```bash
cd web && npm run test:run -- AdminUsersView
```

Expected: FAIL — `data-test="create-user-button"` not found.

- [ ] **Step 3: Add the button + dialog wiring**

In `web/src/views/AdminUsersView.vue`, find the toolbar (search for `v-toolbar` or the existing search/filter row). Add the Create User button. Then in the script section, import the dialog and add show/hide state.

Template addition (in toolbar):

```vue
<v-btn
  color="primary"
  prepend-icon="mdi-account-plus"
  data-test="create-user-button"
  @click="showCreateDialog = true"
>
  Create User
</v-btn>
```

Add the dialog at the end of the template (before closing `</template>`):

```vue
<AdminUserCreateDialog
  v-model="showCreateDialog"
  @created="handleUserCreated"
/>
```

Script additions:

```js
import { ref } from 'vue'
import AdminUserCreateDialog from '@/components/admin/AdminUserCreateDialog.vue'

const showCreateDialog = ref(false)

function handleUserCreated(/* user */) {
  // refresh the user list after a successful create
  fetchUsers()
}
```

(Replace `fetchUsers` with whatever the existing list-loader is named in this view.)

- [ ] **Step 4: Run tests, verify pass**

```bash
cd web && npm run test:run -- AdminUsersView
```

Expected: PASS for the new tests + existing tests still pass.

- [ ] **Step 5: Commit**

```bash
cd /home/jcz/Github/actionlog/.worktrees/feature/admin-user-lifecycle-v1.3.1
git add web/src/views/AdminUsersView.vue web/src/views/AdminUsersView.test.js
git commit -m "feat(web): Create User button on AdminUsersView toolbar

Opens AdminUserCreateDialog; refreshes the user list on successful create."
```

---

## Task 12: AdminSetPasswordDialog component

**Files:**
- Create: `web/src/components/admin/AdminSetPasswordDialog.vue`
- Create: `web/src/components/admin/AdminSetPasswordDialog.test.js`

- [ ] **Step 1: Write failing tests**

Create `web/src/components/admin/AdminSetPasswordDialog.test.js`:

```js
import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount } from '@vue/test-utils'
import AdminSetPasswordDialog from './AdminSetPasswordDialog.vue'
import { createVuetify } from 'vuetify'
import * as components from 'vuetify/components'
import * as directives from 'vuetify/directives'

const vuetify = createVuetify({ components, directives })

vi.mock('@/utils/axios', () => ({
  default: { post: vi.fn() }
}))
import axios from '@/utils/axios'

function mountDialog(targetUser = { id: 5, email: 'target@example.com' }, props = {}) {
  return mount(AdminSetPasswordDialog, {
    global: { plugins: [vuetify] },
    props: { modelValue: true, targetUser, ...props },
    attachTo: document.body,
  })
}

describe('AdminSetPasswordDialog', () => {
  beforeEach(() => {
    axios.post.mockReset()
  })

  it('shows the target email in the title', () => {
    const wrapper = mountDialog({ id: 5, email: 'target@example.com' })
    expect(document.body.innerHTML).toContain('target@example.com')
    wrapper.unmount()
  })

  it('disables submit until passwords valid and matching', () => {
    const wrapper = mountDialog()
    const submitBtn = Array.from(document.querySelectorAll('button')).find(b => b.textContent.includes('Set Password'))
    expect(submitBtn.disabled).toBe(true)
    wrapper.unmount()
  })

  it('POSTs to /api/admin/users/{id}/password and emits success on 204', async () => {
    axios.post.mockResolvedValueOnce({ status: 204, data: '' })
    const wrapper = mountDialog({ id: 42, email: 't@example.com' })
    wrapper.vm.pw.password.value = 'ValidPass123Long'
    wrapper.vm.pw.confirm.value = 'ValidPass123Long'
    await wrapper.vm.submit()
    expect(axios.post).toHaveBeenCalledWith('/api/admin/users/42/password', { new_password: 'ValidPass123Long' })
    expect(wrapper.emitted('password-set')).toBeTruthy()
    wrapper.unmount()
  })

  it('shows complexity error from server on 400', async () => {
    axios.post.mockRejectedValueOnce({
      response: { status: 400, data: { error: 'invalid_input', message: 'new_password: must be at least 12 characters' } },
    })
    const wrapper = mountDialog({ id: 42, email: 't@example.com' })
    wrapper.vm.pw.password.value = 'short'
    wrapper.vm.pw.confirm.value = 'short'
    await wrapper.vm.submit()
    expect(wrapper.vm.errorMessage).toContain('12 characters')
    wrapper.unmount()
  })
})
```

- [ ] **Step 2: Run tests, confirm failure**

```bash
cd web && npm run test:run -- AdminSetPasswordDialog
```

Expected: FAIL — module not found.

- [ ] **Step 3: Implement the dialog**

Create `web/src/components/admin/AdminSetPasswordDialog.vue`:

```vue
<template>
  <v-dialog
    :model-value="modelValue"
    @update:model-value="(v) => $emit('update:modelValue', v)"
    max-width="500"
    persistent
  >
    <v-card>
      <v-card-title class="text-h6">
        Set password for <span class="font-weight-regular">{{ targetUser?.email }}</span>
      </v-card-title>
      <v-card-text>
        <v-alert type="info" density="compact" class="mb-3">
          Sets a new password directly. The user will be signed out of all current
          sessions and any account lockout will be cleared.
        </v-alert>
        <v-form @submit.prevent="submit">
          <v-text-field
            v-model="pw.password.value"
            label="New password"
            :type="pw.visible.value ? 'text' : 'password'"
            :append-inner-icon="pw.visible.value ? 'mdi-eye-off' : 'mdi-eye'"
            @click:append-inner="pw.toggleVisible"
            hint="Min 12 chars; uppercase, lowercase, digit."
            persistent-hint
            autofocus
            required
          />
          <v-text-field
            v-model="pw.confirm.value"
            label="Confirm new password"
            :type="pw.visible.value ? 'text' : 'password'"
            :error-messages="pw.errorMessage.value"
            required
          />
          <v-alert v-if="errorMessage" type="error" density="compact" class="mt-2">
            {{ errorMessage }}
          </v-alert>
        </v-form>
      </v-card-text>
      <v-card-actions>
        <v-spacer />
        <v-btn variant="text" @click="cancel">Cancel</v-btn>
        <v-btn
          color="primary"
          :disabled="!pw.isValid.value || submitting"
          :loading="submitting"
          @click="submit"
        >
          Set Password
        </v-btn>
      </v-card-actions>
    </v-card>
  </v-dialog>
</template>

<script setup>
import { ref, watch } from 'vue'
import axios from '@/utils/axios'
import { usePasswordInputs } from '@/components/admin/composables/usePasswordInputs.js'

const props = defineProps({
  modelValue: { type: Boolean, required: true },
  targetUser: { type: Object, required: true },
})
const emit = defineEmits(['update:modelValue', 'password-set'])

const pw = usePasswordInputs()
const submitting = ref(false)
const errorMessage = ref('')

async function submit() {
  errorMessage.value = ''
  submitting.value = true
  try {
    await axios.post(`/api/admin/users/${props.targetUser.id}/password`, {
      new_password: pw.password.value,
    })
    emit('password-set', { id: props.targetUser.id })
    closeAndReset()
  } catch (e) {
    if (e?.response?.data?.message) {
      errorMessage.value = e.response.data.message
    } else {
      errorMessage.value = 'Could not set password. Check server logs.'
    }
  } finally {
    submitting.value = false
  }
}

function cancel() { closeAndReset() }

function closeAndReset() {
  pw.reset()
  errorMessage.value = ''
  emit('update:modelValue', false)
}

watch(() => props.modelValue, (v) => {
  if (v) errorMessage.value = ''
})

defineExpose({ pw, submit, errorMessage })
</script>
```

- [ ] **Step 4: Run tests, verify pass**

```bash
cd web && npm run test:run -- AdminSetPasswordDialog
```

Expected: 4 PASS.

- [ ] **Step 5: Commit**

```bash
cd /home/jcz/Github/actionlog/.worktrees/feature/admin-user-lifecycle-v1.3.1
git add web/src/components/admin/AdminSetPasswordDialog.vue web/src/components/admin/AdminSetPasswordDialog.test.js
git commit -m "feat(web): AdminSetPasswordDialog component

Targeted dialog with the target user's email in the title (anti-mistarget),
warning about session sign-out + lockout clear, and the shared
usePasswordInputs composable."
```

---

## Task 13: Add Password Management card to ProfileTab

**Files:**
- Modify: `web/src/components/admin/user-edit/ProfileTab.vue`

- [ ] **Step 1: Add the card to the template**

In `web/src/components/admin/user-edit/ProfileTab.vue`, find the existing field grid. After it (before any closing tags), add:

```vue
<v-card class="mt-4" variant="outlined">
  <v-card-title class="text-subtitle-1">Password Management</v-card-title>
  <v-card-text>
    <p class="text-caption mb-3">
      Force Reset sends an email link; the user picks their own password.
      Set Directly lets you type the password yourself — both will sign the
      user out of all current sessions and clear any account lockout.
    </p>
    <div class="d-flex ga-2">
      <v-btn
        variant="outlined"
        prepend-icon="mdi-email-fast"
        @click="confirmForceReset = true"
      >
        Force Password Reset
      </v-btn>
      <v-btn
        color="primary"
        prepend-icon="mdi-key-change"
        @click="showSetPasswordDialog = true"
      >
        Set Password Directly
      </v-btn>
    </div>
  </v-card-text>
</v-card>

<!-- Existing force-reset confirmation dialog stays as-is. -->

<AdminSetPasswordDialog
  v-model="showSetPasswordDialog"
  :target-user="user"
  @password-set="onPasswordSet"
/>
```

In the script:

```js
import { ref } from 'vue'
import AdminSetPasswordDialog from '@/components/admin/AdminSetPasswordDialog.vue'

const showSetPasswordDialog = ref(false)

function onPasswordSet() {
  // optional: toast handled inside the dialog; just close + maybe refresh
}
```

If `confirmForceReset` is already a defined ref (used by the existing button), don't redeclare it.

- [ ] **Step 2: Run frontend tests to verify nothing regressed**

```bash
cd web && npm run test:run -- ProfileTab
```

Expected: PASS — existing ProfileTab tests still pass; nothing new tested here yet (the dialog is independently tested in Task 12).

- [ ] **Step 3: Commit**

```bash
cd /home/jcz/Github/actionlog/.worktrees/feature/admin-user-lifecycle-v1.3.1
git add web/src/components/admin/user-edit/ProfileTab.vue
git commit -m "feat(web): Password Management card on ProfileTab

Two buttons (Force Reset, Set Directly) with explanatory copy that
distinguishes them. Set Directly opens AdminSetPasswordDialog."
```

---

## Task 14: Documentation updates

**Files:**
- Modify: `docs/security/THREAT_MODEL.md`
- Modify: `docs/CHANGELOG.md`
- Modify: `docs/USER_PERMISSIONS.md`
- Modify: `docs/TODO.md`

- [ ] **Step 1: Threat model addendum**

In `docs/security/THREAT_MODEL.md`, find the residual-risks section. Add a row:

```markdown
| Admin compromise → mass user creation / password reset | Audit log review (`admin_user_created` and `admin_password_set` events per actor over a window). Not blocked at the system level — accepted residual risk for a small-team deployment with admin trust. |
```

(Place under whichever existing residual-risks heading matches; the document evolved through v1.3.0 — match the existing structure.)

- [ ] **Step 2: CHANGELOG entry**

At the top of `docs/CHANGELOG.md` (above the v1.3.0 entry), add:

```markdown
## [1.3.1] - YYYY-MM-DD — Admin user lifecycle

### Added
- **`POST /api/admin/users`** — admins can create user accounts with email + password + role; new user can sign in immediately
- **`POST /api/admin/users/{id}/password`** — admin sets a specific password directly; bundles lockout-clear + refresh-token revocation in one operation; protected accounts blocked at L1
- **AdminUserCreateDialog** on the User Management screen — modal with email/password/name/role/email-verified inputs
- **AdminSetPasswordDialog** on the Profile tab Password Management card — sits alongside the existing Force Password Reset button with explanatory copy distinguishing the two
- **`usePasswordInputs`** composable — shared password+confirm state with show/hide toggle, complexity hint, matching validation; used by both new dialogs
- **Three new audit events**: `admin_user_created`, `admin_password_set` (with prior failed-attempts/lockout state captured for forensics), `admin_user_create_rejected_protected`

### Security
- Protected emails are rejected at create with their own audit event — distinct from the `protected_user_attack_*` family which targets modifications of existing protected rows
- L1 `ProtectedUserGuard` covers the new set-password endpoint automatically (inside the `/users/{id}` sub-router)
- Password complexity policy unchanged (12+ chars, upper, lower, digit) — single source of truth in `validatePassword`

### Documentation
- `docs/security/THREAT_MODEL.md` — admin-compromise residual-risk row added
- `docs/USER_PERMISSIONS.md` — new endpoints + UI surfaces

(Note: replace `YYYY-MM-DD` with the merge date when finalizing.)
```

- [ ] **Step 3: USER_PERMISSIONS update**

In `docs/USER_PERMISSIONS.md`, find the admin endpoints section. Add:

```markdown
- `POST /api/admin/users` — admin creates a user (admin-only; rejects protected emails)
- `POST /api/admin/users/{id}/password` — admin sets a specific password (admin-only; protected accounts return 403 via L1)
```

And in the UI section, add:

```markdown
- **AdminUsersView** "Create User" button — admin-only; opens AdminUserCreateDialog
- **ProfileTab Password Management** card — admin-only; "Force Password Reset" + "Set Password Directly" buttons
```

- [ ] **Step 4: TODO mark items complete**

In `docs/TODO.md`, find these lines:

```markdown
- [ ] `[HIGH]` **Admin: Create User flow on User Management screen (v1.3.1)** — ...
- [ ] `[HIGH]` **Admin: edit full user details on Profile tab (v1.3.1)** — ...
```

Change the `[ ]` to `[x]` and append `*(Completed v1.3.1)*` to each.

Leave the `Admin: break-glass CLI for protected users (v1.3.1)` line as `[ ]` — that's a separate spec.

- [ ] **Step 5: Commit**

```bash
git add docs/security/THREAT_MODEL.md docs/CHANGELOG.md docs/USER_PERMISSIONS.md docs/TODO.md
git commit -m "docs: v1.3.1 admin user lifecycle

Threat-model residual-risks row, CHANGELOG entry, USER_PERMISSIONS
updates for new endpoints + UI surfaces, TODO checkboxes for the two
items shipped in this release."
```

---

## Task 15: Version bump

**Files:**
- Modify: `pkg/version/version.go`
- Modify: `web/package.json`
- Modify: `CLAUDE.md`

- [ ] **Step 1: Update Go version constant**

In `pkg/version/version.go`, find:

```go
Major = 1
Minor = 3
Patch = 0
```

Change `Patch = 0` to `Patch = 1`. Leave `Build` alone — `make build` increments it automatically.

- [ ] **Step 2: Update web package version**

In `web/package.json`, change `"version": "1.3.0"` → `"version": "1.3.1"`.

- [ ] **Step 3: Update CLAUDE.md version line**

In `CLAUDE.md`, change `**Version:** 1.3.0` → `**Version:** 1.3.1`.

Also update any version references in the docker test snippets (search for `1.3.0` and update to `1.3.1`).

- [ ] **Step 4: Build to confirm version bump**

```bash
go build -o /tmp/actalog-version-check ./cmd/actalog/
/tmp/actalog-version-check --version 2>&1 | head -2
```

Expected: prints version starting with `1.3.1`.

- [ ] **Step 5: Commit**

```bash
git add pkg/version/version.go web/package.json CLAUDE.md
git commit -m "chore(version): 1.3.0 → 1.3.1"
```

---

## Task 16: Local Docker smoke test + final review

**Files:** none changed (verification step).

- [ ] **Step 1: Pre-flight cleanup per CLAUDE.md**

```bash
pkill -9 -f actalog 2>/dev/null || true
docker ps -q --filter "publish=8080" | xargs -r docker stop
sleep 1
lsof -i :8080 -sTCP:LISTEN 2>&1 || echo "port free"
```

Expected: `port free`.

- [ ] **Step 2: Build the docker image**

```bash
./docker/scripts/build.sh dev 2>&1 | tail -5
```

Expected: `Build Complete!` line.

- [ ] **Step 3: Start fresh SQLite container**

```bash
mkdir -p /tmp/actalog-v1.3.1-smoke && rm -f /tmp/actalog-v1.3.1-smoke/*.db
docker run -d --name actalog-v1.3.1-smoke -p 8080:8080 \
  -v /tmp/actalog-v1.3.1-smoke:/app/data \
  -e DB_DRIVER=sqlite3 -e DB_NAME=/app/data/actalog.db \
  -e JWT_SECRET=smoke-test-only-not-for-prod \
  ghcr.io/johnzastrow/actalog:dev
sleep 3
curl -s -o /dev/null -w "health: %{http_code}\n" http://localhost:8080/health
```

Expected: `health: 200`.

- [ ] **Step 4: Manually exercise the full lifecycle via curl**

```bash
# Bootstrap admin
curl -s -X POST http://localhost:8080/api/auth/register \
  -H 'Content-Type: application/json' \
  -d '{"email":"admin@smoke.test","password":"AdminSmokeTest1","name":"Adm"}'

# Login
TOKEN=$(curl -s -X POST http://localhost:8080/api/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"email":"admin@smoke.test","password":"AdminSmokeTest1"}' \
  | python3 -c 'import sys,json;print(json.load(sys.stdin)["token"])')

# Create user
echo "=== create ==="
curl -s -X POST http://localhost:8080/api/admin/users \
  -H "Authorization: Bearer $TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{"email":"newathlete@smoke.test","password":"NewAthletePass1A","name":"New","role":"athlete","email_verified":true}' \
  -w "\nHTTP: %{http_code}\n"

# Login as the new user
echo "=== new user login ==="
curl -s -X POST http://localhost:8080/api/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"email":"newathlete@smoke.test","password":"NewAthletePass1A"}' \
  -w "\nHTTP: %{http_code}\n" | head -1

# Admin sets password directly
NEW_USER_ID=$(curl -s -H "Authorization: Bearer $TOKEN" "http://localhost:8080/api/admin/users" | python3 -c 'import sys,json; users=json.load(sys.stdin)["data"]; print([u for u in users if u["email"]=="newathlete@smoke.test"][0]["id"])')
echo "=== admin set-password (user id: $NEW_USER_ID) ==="
curl -s -X POST "http://localhost:8080/api/admin/users/$NEW_USER_ID/password" \
  -H "Authorization: Bearer $TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{"new_password":"AdminFixedPass99B"}' \
  -w "\nHTTP: %{http_code}\n"

# Login with the admin-set password
echo "=== login with admin-set password ==="
curl -s -X POST http://localhost:8080/api/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"email":"newathlete@smoke.test","password":"AdminFixedPass99B"}' \
  -w "\nHTTP: %{http_code}\n" | head -1

# Protected email rejection
echo "=== protected email at create (should 400) ==="
PROTECTED=$(grep -E '^\s*"' pkg/security/protected_users.go | head -1 | tr -d '", \t')
curl -s -X POST http://localhost:8080/api/admin/users \
  -H "Authorization: Bearer $TOKEN" \
  -H 'Content-Type: application/json' \
  -d "{\"email\":\"$PROTECTED\",\"password\":\"ValidPass123Long\",\"name\":\"X\",\"role\":\"athlete\"}" \
  -w "\nHTTP: %{http_code}\n"
```

Expected:
- create: HTTP 201
- new user login: HTTP 200
- admin set-password: HTTP 204
- login with admin-set password: HTTP 200
- protected email at create: HTTP 400

- [ ] **Step 5: Open the app in a browser and exercise the UI**

Open http://localhost:8080 in a browser. Log in as `admin@smoke.test`. Click `Admin → Users`, then `Create User`. Create a new user via the dialog. Verify success toast and that the user appears in the list. Click the new user's edit icon, scroll to `Password Management`, click `Set Password Directly`, type a new password, submit. Verify success.

- [ ] **Step 6: Tear down container**

```bash
docker stop actalog-v1.3.1-smoke && docker rm actalog-v1.3.1-smoke
```

- [ ] **Step 7: Final test sweep**

```bash
go test ./... -count=1 2>&1 | grep -E "^(FAIL|ok)" | tail -20
cd web && npm run test:run 2>&1 | tail -5
```

Expected: all PASS.

- [ ] **Step 8: Push and open PR**

```bash
cd /home/jcz/Github/actionlog/.worktrees/feature/admin-user-lifecycle-v1.3.1
git push -u origin feature/admin-user-lifecycle-v1.3.1
gh pr create --base main --head feature/admin-user-lifecycle-v1.3.1 \
  --title "v1.3.1: admin user lifecycle — create + admin-set-password" \
  --body "$(cat <<'EOF'
## Summary

- POST /api/admin/users — admin creates a user; new user can sign in with provided creds (primary onboarding flow)
- POST /api/admin/users/{id}/password — admin sets a specific password directly; bundles lockout-clear + token-revoke
- AdminUserCreateDialog on the User Management screen
- AdminSetPasswordDialog on the Profile tab "Password Management" card (alongside the existing Force Password Reset)
- usePasswordInputs composable — shared password+confirm state with complexity hint, matching validation, show/hide
- Three new audit events: admin_user_created, admin_password_set, admin_user_create_rejected_protected
- Protected emails rejected at create; L1 ProtectedUserGuard covers the new set-password endpoint

## Test plan

- [ ] CI matrix green
- [ ] Local Docker smoke test passes (curl flows + browser flows)
- [ ] Audit log shows the three new event types after exercising the flows
- [ ] Protected user account cannot be set-password'd via admin UI

🤖 Generated with [Claude Code](https://claude.com/claude-code)
EOF
)"
```

---

## End of plan

After PR merges, follow the v1.3.0 cleanup pattern: tag `v1.3.1`, push docker images (`1.3.1`, `dev`, `latest`), delete branch + worktree. Per project preference (memory `feedback_release_cadence`), no GitHub release page until at least v1.3.2.
