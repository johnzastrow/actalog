import { format as formatDate } from 'date-fns'
import { formatInTimeZone, toZonedTime } from 'date-fns-tz'

// Curated list of common timezones with friendly labels
export const TIMEZONE_OPTIONS = [
  { value: 'America/New_York', label: '(UTC-05:00) Eastern Time' },
  { value: 'America/Chicago', label: '(UTC-06:00) Central Time' },
  { value: 'America/Denver', label: '(UTC-07:00) Mountain Time' },
  { value: 'America/Los_Angeles', label: '(UTC-08:00) Pacific Time' },
  { value: 'America/Anchorage', label: '(UTC-09:00) Alaska' },
  { value: 'Pacific/Honolulu', label: '(UTC-10:00) Hawaii' },
  { value: 'Europe/London', label: '(UTC+00:00) London' },
  { value: 'Europe/Paris', label: '(UTC+01:00) Paris, Berlin' },
  { value: 'Europe/Moscow', label: '(UTC+03:00) Moscow' },
  { value: 'Asia/Dubai', label: '(UTC+04:00) Dubai' },
  { value: 'Asia/Kolkata', label: '(UTC+05:30) India' },
  { value: 'Asia/Singapore', label: '(UTC+08:00) Singapore' },
  { value: 'Asia/Tokyo', label: '(UTC+09:00) Tokyo' },
  { value: 'Australia/Sydney', label: '(UTC+10:00) Sydney' },
  { value: 'Pacific/Auckland', label: '(UTC+12:00) Auckland' },
  { value: 'UTC', label: '(UTC+00:00) UTC' },
]

/**
 * Get the browser's detected timezone
 * @returns {string} IANA timezone identifier
 */
export function getBrowserTimezone() {
  try {
    return Intl.DateTimeFormat().resolvedOptions().timeZone
  } catch (e) {
    console.warn('Could not detect browser timezone:', e)
    return 'America/New_York'
  }
}

/**
 * Check if a timezone is in our curated list
 * @param {string} timezone - IANA timezone identifier
 * @returns {boolean}
 */
export function isTimezoneInList(timezone) {
  return TIMEZONE_OPTIONS.some(opt => opt.value === timezone)
}

/**
 * Get timezone options including browser timezone if not in list
 * @returns {Array} Array of timezone options
 */
export function getTimezoneOptions() {
  const browserTz = getBrowserTimezone()
  if (isTimezoneInList(browserTz)) {
    return TIMEZONE_OPTIONS
  }

  // Add browser timezone at the top if not in curated list
  return [
    { value: browserTz, label: `(Browser) ${browserTz}` },
    ...TIMEZONE_OPTIONS
  ]
}

/**
 * Format a date in a specific timezone
 * @param {Date|string} date - Date to format
 * @param {string} timezone - IANA timezone identifier
 * @param {string} formatStr - date-fns format string (default: 'MMM d, yyyy')
 * @returns {string} Formatted date string
 */
export function formatDateInTimezone(date, timezone, formatStr = 'MMM d, yyyy') {
  if (!date) return ''

  try {
    const dateObj = typeof date === 'string' ? new Date(date) : date
    return formatInTimeZone(dateObj, timezone, formatStr)
  } catch (e) {
    console.warn('Error formatting date in timezone:', e)
    // Fallback to basic formatting
    const dateObj = typeof date === 'string' ? new Date(date) : date
    return formatDate(dateObj, formatStr)
  }
}

/**
 * Format a date with time in a specific timezone
 * @param {Date|string} date - Date to format
 * @param {string} timezone - IANA timezone identifier
 * @param {string} formatStr - date-fns format string (default: 'MMM d, yyyy h:mm a')
 * @returns {string} Formatted datetime string
 */
export function formatDateTimeInTimezone(date, timezone, formatStr = 'MMM d, yyyy h:mm a') {
  return formatDateInTimezone(date, timezone, formatStr)
}

/**
 * Get today's date in a specific timezone as YYYY-MM-DD
 * @param {string} timezone - IANA timezone identifier
 * @returns {string} Date string in YYYY-MM-DD format
 */
export function getTodayInTimezone(timezone) {
  try {
    const now = new Date()
    return formatInTimeZone(now, timezone, 'yyyy-MM-dd')
  } catch (e) {
    console.warn('Error getting today in timezone:', e)
    // Fallback to local date
    return formatDate(new Date(), 'yyyy-MM-dd')
  }
}

/**
 * Convert a date to a specific timezone
 * @param {Date|string} date - Date to convert
 * @param {string} timezone - IANA timezone identifier
 * @returns {Date} Date object adjusted to the timezone
 */
export function toTimezone(date, timezone) {
  const dateObj = typeof date === 'string' ? new Date(date) : date
  return toZonedTime(dateObj, timezone)
}

/**
 * Get the timezone label for display
 * @param {string} timezone - IANA timezone identifier
 * @returns {string} Human-readable timezone label
 */
export function getTimezoneLabel(timezone) {
  const option = TIMEZONE_OPTIONS.find(opt => opt.value === timezone)
  return option ? option.label : timezone
}
