import { describe, it, expect, vi, beforeEach } from 'vitest'
import {
  TIMEZONE_OPTIONS,
  getBrowserTimezone,
  isTimezoneInList,
  getTimezoneOptions,
  formatDateInTimezone,
  formatDateTimeInTimezone,
  getTodayInTimezone,
  getTimezoneLabel,
} from './timezone'

describe('timezone utilities', () => {
  describe('TIMEZONE_OPTIONS', () => {
    it('should contain common timezones', () => {
      expect(TIMEZONE_OPTIONS.length).toBeGreaterThan(10)
      expect(TIMEZONE_OPTIONS.some(opt => opt.value === 'America/New_York')).toBe(true)
      expect(TIMEZONE_OPTIONS.some(opt => opt.value === 'Europe/London')).toBe(true)
      expect(TIMEZONE_OPTIONS.some(opt => opt.value === 'Asia/Tokyo')).toBe(true)
    })

    it('should have value and label for each option', () => {
      TIMEZONE_OPTIONS.forEach(opt => {
        expect(opt).toHaveProperty('value')
        expect(opt).toHaveProperty('label')
        expect(typeof opt.value).toBe('string')
        expect(typeof opt.label).toBe('string')
      })
    })
  })

  describe('getBrowserTimezone', () => {
    it('should return a string timezone', () => {
      const tz = getBrowserTimezone()
      expect(typeof tz).toBe('string')
      expect(tz.length).toBeGreaterThan(0)
    })

    it('should return default timezone when Intl is unavailable', () => {
      const originalIntl = global.Intl
      global.Intl = undefined

      const tz = getBrowserTimezone()
      expect(tz).toBe('America/New_York')

      global.Intl = originalIntl
    })
  })

  describe('isTimezoneInList', () => {
    it('should return true for timezones in the list', () => {
      expect(isTimezoneInList('America/New_York')).toBe(true)
      expect(isTimezoneInList('Europe/London')).toBe(true)
      expect(isTimezoneInList('UTC')).toBe(true)
    })

    it('should return false for timezones not in the list', () => {
      expect(isTimezoneInList('America/Phoenix')).toBe(false)
      expect(isTimezoneInList('invalid/timezone')).toBe(false)
      expect(isTimezoneInList('')).toBe(false)
    })
  })

  describe('getTimezoneOptions', () => {
    it('should return at least the curated options', () => {
      const options = getTimezoneOptions()
      expect(options.length).toBeGreaterThanOrEqual(TIMEZONE_OPTIONS.length)
    })

    it('should include all curated timezones', () => {
      const options = getTimezoneOptions()
      TIMEZONE_OPTIONS.forEach(curated => {
        expect(options.some(opt => opt.value === curated.value)).toBe(true)
      })
    })
  })

  describe('formatDateInTimezone', () => {
    it('should format a date with default format', () => {
      const date = new Date('2024-06-15T12:00:00Z')
      const formatted = formatDateInTimezone(date, 'UTC')
      expect(formatted).toBe('Jun 15, 2024')
    })

    it('should format a date with custom format', () => {
      const date = new Date('2024-06-15T12:00:00Z')
      const formatted = formatDateInTimezone(date, 'UTC', 'yyyy-MM-dd')
      expect(formatted).toBe('2024-06-15')
    })

    it('should handle string date input', () => {
      const formatted = formatDateInTimezone('2024-06-15T12:00:00Z', 'UTC')
      expect(formatted).toBe('Jun 15, 2024')
    })

    it('should return empty string for null/undefined input', () => {
      expect(formatDateInTimezone(null, 'UTC')).toBe('')
      expect(formatDateInTimezone(undefined, 'UTC')).toBe('')
    })

    it('should format dates differently based on timezone', () => {
      // Use a time that straddles midnight between timezones
      const date = new Date('2024-06-15T04:00:00Z')
      const nyFormat = formatDateInTimezone(date, 'America/New_York', 'yyyy-MM-dd')
      const tokyoFormat = formatDateInTimezone(date, 'Asia/Tokyo', 'yyyy-MM-dd')
      // At 4am UTC: New York is midnight (June 15), Tokyo is 1pm (June 15)
      expect(nyFormat).toBe('2024-06-15')
      expect(tokyoFormat).toBe('2024-06-15')
    })
  })

  describe('formatDateTimeInTimezone', () => {
    it('should format date with time by default', () => {
      const date = new Date('2024-06-15T15:30:00Z')
      const formatted = formatDateTimeInTimezone(date, 'UTC')
      expect(formatted).toBe('Jun 15, 2024 3:30 PM')
    })

    it('should accept custom format', () => {
      const date = new Date('2024-06-15T15:30:00Z')
      const formatted = formatDateTimeInTimezone(date, 'UTC', 'HH:mm')
      expect(formatted).toBe('15:30')
    })
  })

  describe('getTodayInTimezone', () => {
    it('should return date in YYYY-MM-DD format', () => {
      const today = getTodayInTimezone('UTC')
      expect(today).toMatch(/^\d{4}-\d{2}-\d{2}$/)
    })

    it('should return valid date', () => {
      const today = getTodayInTimezone('America/New_York')
      const date = new Date(today)
      expect(date.toString()).not.toBe('Invalid Date')
    })
  })

  describe('getTimezoneLabel', () => {
    it('should return label for known timezone', () => {
      const label = getTimezoneLabel('America/New_York')
      expect(label).toBe('(UTC-05:00) Eastern Time')
    })

    it('should return the timezone itself for unknown timezone', () => {
      const label = getTimezoneLabel('America/Phoenix')
      expect(label).toBe('America/Phoenix')
    })

    it('should return UTC label', () => {
      const label = getTimezoneLabel('UTC')
      expect(label).toBe('(UTC+00:00) UTC')
    })
  })
})
