package domain_test

import (
	"testing"

	"github.com/johnzastrow/actalog/internal/domain"
)

// TestProtectedUserAuditEventConstants verifies that all protected-user audit
// event constants are defined and have distinct, non-empty string values.
func TestProtectedUserAuditEventConstants(t *testing.T) {
	constants := map[string]string{
		"EventProtectedUserAttackHTTP":    domain.EventProtectedUserAttackHTTP,
		"EventProtectedUserAttackService": domain.EventProtectedUserAttackService,
		"EventProtectedUserAttackDB":      domain.EventProtectedUserAttackDB,
		"EventPasswordResetForcedByAdmin": domain.EventPasswordResetForcedByAdmin,
	}

	// All values must be non-empty
	for name, val := range constants {
		if val == "" {
			t.Errorf("constant %s must not be empty", name)
		}
	}

	// All values must be distinct
	seen := make(map[string]string)
	for name, val := range constants {
		if prev, exists := seen[val]; exists {
			t.Errorf("constant %s has duplicate value %q (same as %s)", name, val, prev)
		}
		seen[val] = name
	}

	// Verify expected string values
	if domain.EventProtectedUserAttackHTTP != "protected_user_attack_http" {
		t.Errorf("EventProtectedUserAttackHTTP = %q, want %q", domain.EventProtectedUserAttackHTTP, "protected_user_attack_http")
	}
	if domain.EventProtectedUserAttackService != "protected_user_attack_service" {
		t.Errorf("EventProtectedUserAttackService = %q, want %q", domain.EventProtectedUserAttackService, "protected_user_attack_service")
	}
	if domain.EventProtectedUserAttackDB != "protected_user_attack_db" {
		t.Errorf("EventProtectedUserAttackDB = %q, want %q", domain.EventProtectedUserAttackDB, "protected_user_attack_db")
	}
	if domain.EventPasswordResetForcedByAdmin != "password_reset_forced_by_admin" {
		t.Errorf("EventPasswordResetForcedByAdmin = %q, want %q", domain.EventPasswordResetForcedByAdmin, "password_reset_forced_by_admin")
	}
}
