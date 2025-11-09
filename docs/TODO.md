# TODO

## Database Schema Migration (v0.3.0) - HIGH PRIORITY

### Schema Changes Required
- [ ] Create database migration from v0.2.0 to v0.3.0
- [ ] Create `wods` table with all attributes (name, source, type, regime, score_type, etc.)
- [ ] Rename `movements` table to `strength_movements`
- [ ] Add `movement_type` column to `strength_movements`
- [ ] Modify `workouts` table (remove user_id, workout_date, workout_type, workout_name, total_time)
- [ ] Create `user_workouts` junction table
- [ ] Rename `workout_movements` to `workout_strength`
- [ ] Create `workout_wods` junction table
- [ ] Create `user_settings` table
- [ ] Create `audit_logs` table
- [ ] Add `updated_by` columns to all relevant tables
- [ ] Migrate existing data to new schema structure
- [ ] Test migration on development database
- [ ] Create rollback migration script

### Backend Updates for New Schema
- [ ] Update domain models for new entities (WOD, Strength, UserWorkout, etc.)
- [ ] Create repository interfaces and implementations for new entities
- [ ] Update service layer to work with new schema
- [ ] Update API handlers for new data structure
- [ ] Add validation for WOD attributes (source, type, regime, score_type)
- [ ] Implement audit logging functionality
- [ ] Create user settings management endpoints

### Seed Data
- [ ] Create seed data for standard CrossFit WODs (Fran, Grace, Helen, Diane, Karen, Murph, DT, etc.)
- [ ] Create seed data for standard strength movements
- [ ] Categorize movements by type (weightlifting, cardio, gymnastics)
- [ ] Add descriptions and URLs for standard WODs

## PWA Features (v0.2.0)

### Completed ✅
- [x] Configure vite-plugin-pwa
- [x] Create web app manifest
- [x] Set up service worker with Workbox
- [x] Implement IndexedDB offline storage
- [x] Add background sync queue
- [x] Service worker registration
- [x] Auto-update notification system
- [x] PWA documentation (DEPLOYMENT.md)
- [x] PWA development setup in SETUP.md

### Remaining PWA Tasks
- [ ] Generate all PWA icon sizes (72px - 512px)
- [ ] Create apple-touch-icon.png for iOS
- [ ] Test offline workout creation
- [ ] Test background sync functionality
- [ ] Implement offline indicator in UI
- [ ] Add sync status indicator
- [ ] Test install prompt on all platforms
- [ ] Run Lighthouse PWA audit
- [ ] Optimize service worker cache size

## High Priority

### Authentication & User Management
- [ ] Implement password reset functionality
- [ ] Add email verification for new users
- [ ] Implement "Remember Me" functionality
- [ ] Add profile picture upload
- [ ] Add user profile editing

### Workout Logging (Updated for v0.3.0 Schema)
- [ ] Implement workout template creation API endpoints
- [ ] Implement user_workout logging endpoints (link user to workout on specific date)
- [ ] Add WOD creation/editing for custom WODs
- [ ] Add strength movement creation/editing for custom movements
- [ ] Implement workout_wod association endpoints
- [ ] Implement workout_strength association endpoints with weight/sets/reps
- [ ] Implement workout history retrieval (via user_workouts)
- [ ] Add workout template editing and deletion
- [ ] Add workout search and filtering
- [ ] Implement PR (Personal Record) tracking across user_workouts
- [ ] Add scoring for WODs (time, rounds+reps, max weight)

### Movement Database
- [ ] Seed database with standard CrossFit movements
- [ ] Add movement categories (Weightlifting, Gymnastics, etc.)
- [ ] Implement movement search functionality
- [ ] Add movement details and instructions
- [ ] Support for custom movements per user

### Progress Tracking
- [ ] Implement data aggregation for charts
- [ ] Add progress by movement endpoint
- [ ] Add progress by date range endpoint
- [ ] Calculate and display PRs
- [ ] Add workout frequency analytics

## Medium Priority

### Data Import/Export
- [ ] Implement CSV export for workouts
- [ ] Implement JSON export for workouts
- [ ] Add CSV import functionality
- [ ] Add JSON import functionality
- [ ] Validate imported data

### Admin Features
- [ ] Admin dashboard
- [ ] User management interface
- [ ] System settings management
- [ ] Database backup functionality
- [ ] User activity monitoring

### Frontend Enhancements
- [ ] Connect all views to backend APIs
- [ ] Add loading states and error handling
- [ ] Implement data caching with Pinia and IndexedDB
- [x] Add offline support (PWA) - v0.2.0
- [ ] Add pull-to-refresh on mobile (can use PWA techniques)
- [ ] Integrate offline storage with workout forms
- [ ] Show network status indicator
- [ ] Display sync status for pending workouts

### Testing
- [ ] Write unit tests for services
- [ ] Write unit tests for repositories
- [ ] Write integration tests for API endpoints
- [ ] Add frontend component tests
- [ ] Set up CI/CD pipeline

## Low Priority

### Performance
- [ ] Add database query optimization
- [x] Implement PWA caching (service worker) - v0.2.0
- [ ] Add Redis for session storage
- [ ] Optimize frontend bundle size
- [ ] Add lazy loading for images
- [x] Precache static assets - v0.2.0
- [x] Implement code splitting preparation - v0.2.0

### Social Features
- [ ] Add workout sharing
- [ ] Add leaderboards
- [ ] Add workout comments
- [ ] Add friend system
- [ ] Add workout templates

### Notifications
- [ ] Implement email notifications
- [ ] Add in-app notifications
- [ ] Add workout reminders via push notifications (PWA)
- [ ] Add achievement notifications
- [ ] Implement Web Push API for PWA notifications
- [ ] Add notification preferences in settings

### Documentation
- [ ] Complete API documentation
- [ ] Add user guide
- [ ] Create developer setup guide
- [ ] Add deployment guide
- [ ] Create video tutorials

## Future Considerations

- [x] Progressive Web App (completed v0.2.0)
- [ ] Advanced PWA features:
  - [ ] Periodic background sync for data refresh
  - [ ] Web Share API for workout sharing
  - [ ] File System Access API for bulk operations
  - [ ] Badging API for unsynced notifications
- [ ] Mobile native apps (iOS/Android) - may not be needed with PWA
- [ ] Apple Watch integration
- [ ] Wearable device sync
- [ ] Nutrition tracking
- [ ] Workout planning/programming
- [ ] Coach/athlete relationship features
- [ ] Gym/box management features
- [ ] Payment/subscription system
- [ ] Multi-language support

## Technical Debt

- [ ] Add comprehensive error handling
- [ ] Improve logging with structured logging
- [ ] Add request rate limiting
- [ ] Implement API versioning
- [ ] Add database migrations system
- [ ] Set up monitoring and alerting
- [ ] Add security headers
- [ ] Implement CSRF protection
- [ ] Clean up old service worker caches
- [ ] Implement PWA update strategy testing

## Deployment Tasks

- [ ] Set up production HTTPS (Let's Encrypt)
- [ ] Configure Nginx for PWA (see DEPLOYMENT.md)
- [ ] Generate production PWA icons
- [ ] Test PWA install on all platforms
- [ ] Set up automated backups
- [ ] Configure monitoring and alerting
- [ ] Set up SSL auto-renewal
- [ ] Performance testing and optimization
- [ ] Security audit

---

**Last Updated:** 2025-11-08
**Version:** 0.2.0
