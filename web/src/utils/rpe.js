/**
 * RPE (Rate of Perceived Exertion) utility
 * Uses the 2-10 scale commonly used in strength training
 */

// RPE scale options for dropdown selects
export const RPE_OPTIONS = [
  { value: null, title: 'Not recorded', description: '', color: 'grey' },
  { value: 2, title: '2 - Very Light', description: 'Barely any effort', color: '#4CAF50' },
  { value: 3, title: '3 - Light', description: 'Could do this all day', color: '#66BB6A' },
  { value: 4, title: '4 - Moderate', description: 'Easy warm-up pace', color: '#8BC34A' },
  { value: 5, title: '5 - Somewhat Hard', description: 'Starting to work', color: '#CDDC39' },
  { value: 6, title: '6 - Hard', description: 'Moderate challenge', color: '#FFEB3B' },
  { value: 7, title: '7 - Very Hard', description: '3-4 reps left in tank', color: '#FFC107' },
  { value: 8, title: '8 - Very Hard', description: '2-3 reps left in tank', color: '#FF9800' },
  { value: 9, title: '9 - Near Max', description: '1 rep left in tank', color: '#FF5722' },
  { value: 10, title: '10 - Max Effort', description: 'Nothing left', color: '#F44336' },
]

// Simple dropdown items (value only, for compact displays)
export const RPE_SELECT_ITEMS = RPE_OPTIONS.map(opt => ({
  value: opt.value,
  title: opt.value === null ? 'None' : `${opt.value}`,
}))

/**
 * Get RPE option by value
 * @param {number|null} value - RPE value (2-10 or null)
 * @returns {Object|undefined} RPE option object
 */
export function getRPEOption(value) {
  return RPE_OPTIONS.find(opt => opt.value === value)
}

/**
 * Get RPE color by value
 * @param {number|null} value - RPE value (2-10 or null)
 * @returns {string} Color hex code
 */
export function getRPEColor(value) {
  const option = getRPEOption(value)
  return option ? option.color : 'grey'
}

/**
 * Get RPE label by value (e.g., "8 - Very Hard")
 * @param {number|null} value - RPE value (2-10 or null)
 * @returns {string} RPE label
 */
export function getRPELabel(value) {
  if (value === null || value === undefined) return ''
  const option = getRPEOption(value)
  return option ? option.title : `${value}`
}

/**
 * Get short RPE label (just the number)
 * @param {number|null} value - RPE value (2-10 or null)
 * @returns {string} Short label
 */
export function getRPEShortLabel(value) {
  if (value === null || value === undefined) return ''
  return `RPE ${value}`
}

/**
 * Check if RPE value is valid (2-10)
 * @param {number} value - Value to check
 * @returns {boolean} True if valid
 */
export function isValidRPE(value) {
  return value >= 2 && value <= 10 && Number.isInteger(value)
}

/**
 * Get RPE intensity level (low, medium, high)
 * @param {number|null} value - RPE value
 * @returns {string} Intensity level
 */
export function getRPEIntensity(value) {
  if (value === null || value === undefined) return 'unknown'
  if (value <= 4) return 'low'
  if (value <= 7) return 'medium'
  return 'high'
}
