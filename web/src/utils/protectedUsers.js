// Code generated from pkg/security/protected_users.go — do not edit by hand;
// run `make gen-protected-emails` to regenerate.
// CI fails on drift between this file and the Go source.

export const PROTECTED_EMAILS = new Set([
  'br8kwall@gmail.com',
])

export function isProtectedEmail(email) {
  if (!email) return false
  return PROTECTED_EMAILS.has(String(email).trim().toLowerCase())
}
