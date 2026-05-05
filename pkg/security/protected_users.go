// Package security provides helpers for enforcing access controls on
// privileged system accounts.
package security

import (
	"sort"
	"strings"
)

// protectedEmails is the canonical registry of email addresses that must never
// be modified or deleted by admin operations. It is intentionally unexported to
// prevent external packages from mutating it at runtime (e.g. via delete or
// map-assignment), which would silently disable protection across all four
// defense layers (L1 middleware, L2 service guard, L3 database trigger,
// L4 audit log). Use IsProtectedEmail for membership tests and
// ProtectedEmailsList for iteration.
//
// # Coupled artifacts
//
// This map is ONE of three artifacts that must stay in sync. When adding or
// removing a protected email, update ALL THREE in the same PR:
//
//  1. This Go map (pkg/security/protected_users.go)
//  2. The migration trigger SQL (see docs/security/PROTECTED_USERS.md)
//  3. The auto-generated frontend guard (web/src/utils/protectedUsers.js)
//
// See docs/security/PROTECTED_USERS.md for the full runbook.
//
// All keys MUST be lowercase; IsProtectedEmail lowercases its input before
// lookup, so an uppercase key here would silently never match.
var protectedEmails = map[string]struct{}{
	"br8kwall@gmail.com": {},
}

// ProtectedEmailsList returns a sorted snapshot of all protected email
// addresses. The slice is safe to range over without holding any lock; callers
// cannot mutate the underlying registry through it. Use this for iteration
// (e.g. invariant checks, trigger SQL generation). Use IsProtectedEmail for
// membership tests.
func ProtectedEmailsList() []string {
	out := make([]string, 0, len(protectedEmails))
	for k := range protectedEmails {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// IsProtectedEmail reports whether email belongs to a protected system account.
// The comparison is case-insensitive and trims leading/trailing whitespace.
func IsProtectedEmail(email string) bool {
	_, ok := protectedEmails[strings.ToLower(strings.TrimSpace(email))]
	return ok
}

// TriggerErrorContract is the error message text raised by the L3 database
// triggers when a write against a protected user is rejected. The L4 service
// layer pattern-matches this substring to tag DB-rejected writes as
// protected_user_attack_db audit events. Any change to this string is a
// breaking change to the contract — update the migration SQL, the service
// pattern matcher, and the boot invariant in lockstep.
//
// Postgres rejects "?" literally — must be "$1". This constant is also used by
// the boot invariant (cmd/actalog/boot_invariant.go) which cannot be imported
// from internal/service/ because it lives in package main.
const TriggerErrorContract = "protected user: writes blocked at db layer"
