import { describe, it, expect } from 'vitest'
import { isProtectedEmail, PROTECTED_EMAILS } from './protectedUsers'

describe('protectedUsers module', () => {
  // =========================================================================
  // PROTECTED_EMAILS export
  // =========================================================================

  describe('PROTECTED_EMAILS', () => {
    it('exports a Set', () => {
      expect(PROTECTED_EMAILS).toBeInstanceOf(Set)
    })

    it('is non-empty', () => {
      expect(PROTECTED_EMAILS.size).toBeGreaterThan(0)
    })

    it('contains the project owner email', () => {
      expect(PROTECTED_EMAILS.has('br8kwall@gmail.com')).toBe(true)
    })
  })

  // =========================================================================
  // isProtectedEmail — exact match
  // =========================================================================

  describe('isProtectedEmail — exact match', () => {
    it('returns true for the project owner email', () => {
      expect(isProtectedEmail('br8kwall@gmail.com')).toBe(true)
    })
  })

  // =========================================================================
  // isProtectedEmail — case-insensitive
  // =========================================================================

  describe('isProtectedEmail — case-insensitive', () => {
    it('returns true for all-uppercase input', () => {
      expect(isProtectedEmail('BR8KWALL@GMAIL.COM')).toBe(true)
    })

    it('returns true for mixed-case local part', () => {
      expect(isProtectedEmail('BR8KWALL@gmail.com')).toBe(true)
    })

    it('returns true for mixed-case throughout', () => {
      expect(isProtectedEmail('Br8kWaLl@Gmail.Com')).toBe(true)
    })
  })

  // =========================================================================
  // isProtectedEmail — whitespace trimming
  // =========================================================================

  describe('isProtectedEmail — whitespace trimming', () => {
    it('returns true when email has leading and trailing spaces', () => {
      expect(isProtectedEmail('  br8kwall@gmail.com  ')).toBe(true)
    })

    it('returns true when email has only leading spaces', () => {
      expect(isProtectedEmail('  br8kwall@gmail.com')).toBe(true)
    })

    it('returns true when email has only trailing spaces', () => {
      expect(isProtectedEmail('br8kwall@gmail.com  ')).toBe(true)
    })
  })

  // =========================================================================
  // isProtectedEmail — plus-addressing does NOT match
  // =========================================================================

  describe('isProtectedEmail — plus-addressing', () => {
    it('returns false for plus-addressed variant of a protected email', () => {
      expect(isProtectedEmail('br8kwall+test@gmail.com')).toBe(false)
    })

    it('returns false for plus-addressed variant with uppercase', () => {
      expect(isProtectedEmail('BR8KWALL+FILTER@gmail.com')).toBe(false)
    })
  })

  // =========================================================================
  // isProtectedEmail — non-protected emails
  // =========================================================================

  describe('isProtectedEmail — non-protected emails', () => {
    it('returns false for an arbitrary email', () => {
      expect(isProtectedEmail('alice@example.com')).toBe(false)
    })

    it('returns false for a similar but different email', () => {
      expect(isProtectedEmail('br8kwall@yahoo.com')).toBe(false)
    })
  })

  // =========================================================================
  // isProtectedEmail — falsy / empty inputs
  // =========================================================================

  describe('isProtectedEmail — falsy inputs', () => {
    it('returns false for empty string', () => {
      expect(isProtectedEmail('')).toBe(false)
    })

    it('returns false for null', () => {
      expect(isProtectedEmail(null)).toBe(false)
    })

    it('returns false for undefined', () => {
      expect(isProtectedEmail(undefined)).toBe(false)
    })
  })
})
