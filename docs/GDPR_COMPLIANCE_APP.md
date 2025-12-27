# GDPR Compliance Assessment - ActaLog Application

**Assessment Date:** December 2025
**Version Assessed:** 0.17.0-beta
**Assessor:** Automated Code Analysis

---

## Executive Summary

ActaLog demonstrates **good technical security** but has **significant GDPR compliance gaps** primarily in legal documentation, comprehensive data export capabilities, and frontend token storage. The application has a strong foundation for compliance with excellent audit logging and no third-party data sharing, but requires immediate attention on several fronts.

### Overall Status: PARTIALLY COMPLIANT

| Area | Status |
|------|--------|
| Data Security | GOOD |
| Audit Logging | EXCELLENT |
| Third-Party Sharing | EXCELLENT |
| Legal Documentation | INCOMPLETE |
| Data Subject Rights | PARTIAL |
| Token Security | NEEDS IMPROVEMENT |

---

## 1. Data Collection Assessment

### 1.1 Personal Data Collected

**During Registration** (`internal/domain/user.go:8-36`, `internal/handler/auth_handler.go:25-30`):

| Field | Required | Sensitivity | Purpose |
|-------|----------|-------------|---------|
| Email | Yes | Medium | Account identification, communications |
| Password | Yes | High | Authentication (stored as bcrypt hash) |
| Name | Yes | Low | Display name |
| Role | Auto | Low | Access control |
| Birthday | Optional | High | Age-related features (CONCERN) |
| Profile Image | Optional | Low | Personalization |

**Auto-Generated Security Fields:**
- Email verification status and tokens
- Password reset tokens
- Failed login attempt counters
- Account lock timestamps
- Last login timestamp

**CONCERN:** Birthday collection lacks documented legal basis. This is potentially sensitive data under GDPR Article 9 (health-related inference).

### 1.2 Data Minimization

The application collects only data necessary for its fitness tracking purpose, with the exception of the optional birthday field which may not be strictly necessary.

**Recommendation:** Remove birthday field or document explicit consent and purpose.

---

## 2. Data Storage Security

### 2.1 Password Security

**File:** `pkg/auth/password.go`

- **Algorithm:** bcrypt
- **Cost Factor:** 12 (industry standard, strong)
- **Status:** COMPLIANT

### 2.2 Sensitive Field Protection

**File:** `internal/domain/user.go`

Sensitive fields are properly protected from JSON serialization:
```go
PasswordHash         `json:"-"`
VerificationToken    `json:"-"`
ResetToken           `json:"-"`
FailedLoginAttempts  `json:"-"`
LockedAt             `json:"-"`
LockedUntil          `json:"-"`
```

**Status:** COMPLIANT

### 2.3 Database Encryption

- SQLite: Unencrypted (development only)
- PostgreSQL/MySQL: Depends on deployment configuration

**Recommendation:** Document at-rest encryption requirements for production deployments.

---

## 3. Session Management and Cookies

### 3.1 JWT Token Implementation

**File:** `pkg/auth/jwt.go:16-22`

Claims structure:
- User ID
- Email address
- Role
- Standard claims (iat, exp, nbf)

Expiration: Configurable via `JWT_EXPIRATION` (default: 24 hours)

### 3.2 Token Storage (CRITICAL ISSUE)

**File:** `web/src/stores/auth.js`

```javascript
localStorage.setItem('token', token.value)
localStorage.setItem('user', JSON.stringify(user.value))
localStorage.setItem('refreshToken', refreshToken.value)
```

**Issues:**
- Tokens stored in localStorage are accessible to JavaScript
- Vulnerable to XSS attacks
- No secure, httpOnly, or sameSite flags possible

**Status:** NON-COMPLIANT (Article 32 - Security of Processing)

**Recommendation:** Migrate to HTTP-only cookies with:
- `Secure` flag (HTTPS only)
- `SameSite=Strict`
- `HttpOnly` flag

### 3.3 Refresh Token Sessions

**File:** `internal/domain/user.go:38-47`

- Supports device tracking
- Supports session revocation
- Can revoke all sessions for a user

**Status:** COMPLIANT

---

## 4. Email Communications

### 4.1 Emails Sent

**File:** `pkg/email/email.go`

| Email Type | Trigger | Token Expiry |
|------------|---------|--------------|
| Email Verification | Registration | 24 hours |
| Password Reset | User request | 1 hour |

### 4.2 Email Content

- Clear "If you didn't request this" disclaimers
- No marketing content without consent
- Configurable SMTP (no third-party dependencies)

**Status:** COMPLIANT

---

## 5. Data Subject Rights

### 5.1 Right of Access (Article 15)

**Current Exports Available** (`internal/handler/export_handler.go`):
- WODs (Workouts of the Day)
- Movements
- User Workouts

**Missing from Exports:**
- Audit logs (user's own actions)
- User settings
- Subscription data
- Organization relationships
- Notification history
- Login history

**Status:** PARTIALLY COMPLIANT

**Recommendation:** Create comprehensive `/api/user/export-all-data` endpoint returning ALL personal data.

### 5.2 Right to Erasure (Article 17)

**File:** `internal/service/user_service.go:978-1022`

- Admin-only deletion
- Hard delete from users table
- Audit logged

**CONCERN:** Cascade deletion not verified for related tables:
- refresh_tokens
- audit_logs (user_id, target_user_id)
- data_change_logs
- user_workouts
- user_settings
- user_subscriptions
- user_organizations
- notifications

**Status:** PARTIALLY COMPLIANT

**Recommendation:** Implement transactional cascade delete with verification, or add database-level CASCADE constraints.

### 5.3 Right to Data Portability (Article 20)

Export functionality exists but is incomplete (see 5.1).

**Status:** PARTIALLY COMPLIANT

### 5.4 Right to Rectification (Article 16)

Users can update their profile information through the application.

**Status:** COMPLIANT

---

## 6. Data Retention

### 6.1 Current State

**Files:** `internal/domain/audit_log.go:97-119`, `internal/domain/data_change_log.go:58-61`

Cleanup methods exist:
```go
DeleteOlderThan(before time.Time) (int, error)
```

**Issue:** No automatic retention policy is configured or enforced.

**Status:** NON-COMPLIANT (Article 5.1(e) - Storage Limitation)

**Recommendation:** Implement scheduled cleanup jobs with configurable retention periods.

### 6.2 Recommended Retention Periods

| Data Type | Suggested Retention |
|-----------|---------------------|
| Audit logs | 12 months |
| Data change logs | 12 months |
| Expired refresh tokens | 30 days after expiry |
| Password reset tokens | Immediate on use or expiry |
| Email verification tokens | 7 days after expiry |

---

## 7. Third-Party Data Sharing

### 7.1 External Services

| Service | Data Shared | Purpose |
|---------|-------------|---------|
| Google Fonts | IP address (CDN request) | Typography |
| None others detected | - | - |

### 7.2 Assessment

- No analytics services
- No crash reporting services
- No payment processors integrated
- No marketing pixels
- No social media SDKs

**Status:** COMPLIANT

---

## 8. Audit Logging

### 8.1 Events Tracked

**File:** `internal/domain/audit_log.go`

**Authentication Events:**
- login_success, login_failed, logout, token_refresh

**Account Security Events:**
- account_locked_auto, account_unlocked_admin
- account_disabled, account_enabled

**Data Modification Events:**
- password_changed, password_reset
- email_changed, email_verified
- profile_updated, user_settings_updated
- user_created, user_updated, user_deleted
- role_changed

### 8.2 Audit Log Fields

- User ID (actor)
- Target User ID (subject)
- Event Type
- IP Address
- User-Agent
- Details (JSON)
- Timestamp

**Status:** EXCELLENT (Articles 5.2, 32 - Accountability and Security)

---

## 9. Missing Legal Documentation

### 9.1 Required Documents Not Found

1. **Privacy Policy**
   - Data processing purposes
   - Legal basis for processing
   - Data retention periods
   - Data subject rights
   - Contact information

2. **Terms of Service**
   - Data processing terms
   - User consent provisions

3. **Cookie Policy**
   - Cookie usage (currently none)
   - Third-party cookies

4. **Data Processing Agreement (DPA)**
   - For organizational customers
   - Standard Contractual Clauses if applicable

### 9.2 Consent Mechanisms

- No consent checkboxes during registration
- No granular consent preferences
- No consent withdrawal mechanism

**Status:** NON-COMPLIANT (Articles 6, 7)

---

## 10. CORS and Access Control

### 10.1 CORS Configuration

**File:** `pkg/middleware/cors.go:10-50`

- Configurable allowed origins
- Credentials allowed
- Proper headers exposed

**Status:** COMPLIANT

### 10.2 Role-Based Access Control

- User/Admin roles
- Middleware enforcement
- User-scoped data access

**Status:** COMPLIANT

---

## 11. Incident Response

### 11.1 Breach Detection

No evidence of:
- Breach detection mechanisms
- Automated alerting
- Intrusion detection

### 11.2 Breach Notification

Missing:
- Notification procedures
- User notification templates
- DPA notification process
- Breach log

**Status:** NON-COMPLIANT (Article 33, 34)

---

## Compliance Gap Summary

### Critical (Immediate Action Required)

| Gap | Article | Recommendation |
|-----|---------|----------------|
| Missing Privacy Policy | 13, 14 | Create and publish privacy policy |
| Token Storage Insecurity | 32 | Migrate to HTTP-only cookies |
| Incomplete Data Export | 15, 20 | Add comprehensive export endpoint |
| No Consent Mechanism | 6, 7 | Add registration consent checkboxes |

### High Priority

| Gap | Article | Recommendation |
|-----|---------|----------------|
| Cascade Delete Unverified | 17 | Verify/implement cascade deletion |
| No Data Retention Policy | 5.1(e) | Configure automatic data cleanup |
| No Breach Procedures | 33, 34 | Document incident response |

### Medium Priority

| Gap | Article | Recommendation |
|-----|---------|----------------|
| Birthday Field | 9 | Remove or add explicit consent |
| Terms of Service | 6 | Create and require acceptance |
| DPA Template | 28 | Create for organizational use |

---

## Implementation Recommendations

### Phase 1: Legal Foundation (Critical)

1. **Create Privacy Policy** (`docs/PRIVACY_POLICY.md` or hosted page)
   - All data categories collected
   - Legal basis for each
   - Retention periods
   - User rights
   - Contact information

2. **Create Terms of Service**
   - Data processing terms
   - User responsibilities
   - Service limitations

3. **Add Consent Mechanism**
   - Registration checkbox for privacy policy
   - Link to privacy policy
   - Store consent timestamp

### Phase 2: Technical Compliance

4. **Fix Token Storage**
   ```javascript
   // Replace localStorage with secure cookie handling
   // Backend: Set-Cookie with HttpOnly, Secure, SameSite=Strict
   ```

5. **Comprehensive Data Export**
   ```go
   // Add endpoint: GET /api/user/gdpr-export
   // Returns: all personal data in JSON format
   ```

6. **Cascade Delete Verification**
   ```go
   // Transaction with explicit deletion from all tables
   // Or add ON DELETE CASCADE constraints
   ```

### Phase 3: Operational Compliance

7. **Data Retention Jobs**
   - Scheduled cleanup for audit logs
   - Cleanup for expired tokens

8. **Breach Response Procedures**
   - Detection mechanisms
   - Response playbook
   - Notification templates

---

## Appendix: File References

| Concern | File | Lines |
|---------|------|-------|
| User data model | `internal/domain/user.go` | 8-36 |
| Auth handler | `internal/handler/auth_handler.go` | 25-30, 52-103 |
| JWT implementation | `pkg/auth/jwt.go` | 16-39 |
| Password hashing | `pkg/auth/password.go` | All |
| Audit logging | `internal/domain/audit_log.go` | All |
| User deletion | `internal/service/user_service.go` | 978-1022 |
| Data exports | `internal/handler/export_handler.go` | All |
| Email service | `pkg/email/email.go` | All |
| Frontend auth | `web/src/stores/auth.js` | All |
| CORS config | `pkg/middleware/cors.go` | 10-50 |

---

## Document History

| Date | Version | Changes |
|------|---------|---------|
| December 2025 | 1.0 | Initial assessment |
