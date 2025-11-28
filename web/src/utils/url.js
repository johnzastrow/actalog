/**
 * URL utility functions for handling API and asset URLs
 * Supports both development (with Vite proxy) and production environments
 */

/**
 * Get the API base URL
 * - In development: uses Vite proxy (empty string for relative URLs)
 * - In production: uses VITE_API_BASE_URL env variable or constructs from window.location
 * @returns {string} The API base URL
 */
export function getApiBaseUrl() {
  // Use environment variable if set
  if (import.meta.env.VITE_API_BASE_URL) {
    return import.meta.env.VITE_API_BASE_URL
  }

  // In development with Vite proxy, use relative URLs
  if (import.meta.env.DEV) {
    return ''
  }

  // In production, construct from window.location
  // This assumes the API is on the same host and port as the frontend
  return `${window.location.protocol}//${window.location.host}`
}

/**
 * Convert a relative path to an absolute URL for static assets
 * Handles profile images, avatars, etc. that are served from the backend
 * @param {string} path - The relative path (e.g., '/uploads/avatar.jpg')
 * @returns {string} The absolute URL
 */
export function getAssetUrl(path) {
  if (!path) return ''

  // Already an absolute URL
  if (path.startsWith('http://') || path.startsWith('https://')) {
    return path
  }

  // Use environment variable if set
  if (import.meta.env.VITE_API_BASE_URL) {
    return `${import.meta.env.VITE_API_BASE_URL}${path}`
  }

  // In development, proxy through Vite dev server
  // The Vite proxy will forward /api and /uploads requests to the backend
  if (import.meta.env.DEV) {
    // For uploads, we need to proxy through the dev server
    // The vite.config.js should handle /uploads similar to /api
    return path
  }

  // In production, construct full URL using the same host as the frontend
  // When behind a reverse proxy (nginx/Caddy), use the same protocol/host/port
  // as the current page - the proxy will route /uploads to the backend
  return `${window.location.origin}${path}`
}

/**
 * Ensure a user profile image has a full URL
 * @param {string|null} profileImage - The profile image path from the API
 * @returns {string|null} The full URL or null if no image
 */
export function getProfileImageUrl(profileImage) {
  if (!profileImage) return null
  return getAssetUrl(profileImage)
}
