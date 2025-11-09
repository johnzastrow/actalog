# TODO

This document tracks planned features, improvements, and known issues for ActaLog.

## Current Status (v0.2.0-beta)

### Completed Features
- [x] User authentication (registration, login, JWT)
- [x] Workout CRUD operations
- [x] Movement management with 31 seeded CrossFit movements
- [x] Workout logging functionality
- [x] Workouts list with movement details
- [x] Dashboard with live statistics
- [x] Searchable movement selection (autocomplete)
- [x] Modern mobile-first UI design
- [x] Bottom navigation
- [x] Profile view with logout
- [x] Clean Architecture implementation
- [x] API authentication middleware
- [x] Database seeding for movements
- [x] Responsive scrolling on all views

## High Priority

### Features
- [ ] Workout detail view (individual workout page)
- [ ] Edit workout functionality
- [ ] Delete workout with confirmation
- [ ] Performance charts for movement progress over time
- [ ] Add custom movements from the UI
- [ ] Workout history filtering by date range
- [ ] Search workouts by movement or date

### UI/UX Improvements
- [ ] Loading states for all API calls
- [ ] Success notifications after save operations
- [ ] Error boundary for better error handling
- [ ] Pull-to-refresh on mobile
- [ ] Skeleton loaders for better perceived performance

### Backend
- [ ] Pagination for workout lists
- [ ] Workout search and filtering endpoints
- [ ] Movement statistics endpoint (PR tracking)
- [ ] User preferences/settings storage
- [ ] Data export functionality (CSV, JSON)
- [ ] Enable support for Mariadb

## Medium Priority

### Features
- [ ] Multiple movements per workout (currently supports one)
- [ ] Timed workouts (AMRAP, EMOM, For Time)
- [ ] Workout templates for common WODs
- [ ] Personal records (PR) tracking and display
- [ ] Workout sharing between users
- [ ] Comments/notes on specific movements
- [ ] Photo upload for form checks
- [ ] Rest timer integration

### Data & Analytics
- [ ] Weekly/monthly workout summaries
- [ ] Volume tracking (total weight lifted)
- [ ] Movement frequency analysis
- [ ] Progress streak tracking
- [ ] Goal setting and tracking

### UI/UX
- [ ] Dark mode support
- [ ] Workout calendar view
- [ ] Movement video demonstrations
- [ ] Onboarding tutorial for new users
- [ ] Settings page (units, preferences)

## Low Priority

### Features
- [ ] Social features (friends, leaderboards)
- [ ] Gym/box integration
- [ ] Coach accounts with athlete management
- [ ] Workout planning and scheduling
- [ ] Integration with wearable devices
- [ ] Mobile app (React Native or Flutter)
- [ ] Offline support with sync

### Backend
- [ ] Redis caching layer
- [ ] Background jobs for data aggregation
- [ ] Email notifications
- [ ] API rate limiting
- [ ] GraphQL API option
- [ ] Multi-tenant support for gyms

### DevOps
- [ ] Docker Compose for development
- [ ] CI/CD pipeline (GitHub Actions)
- [ ] Automated testing (unit, integration, e2e)
- [ ] Database migration system
- [ ] Monitoring and logging (Prometheus, Grafana)
- [ ] Production deployment guide

## Known Issues

### Bugs
- [ ] Date picker may show timezone offset issues
- [ ] Console warning about iterable (third-party library)
- [ ] Movement type filter not yet implemented

### Technical Debt
- [ ] Add comprehensive test coverage (currently minimal)
- [ ] Implement proper error logging system
- [ ] Add API request/response validation with schemas
- [ ] Implement database migrations instead of manual schema
- [ ] Add API versioning strategy
- [ ] Refactor duplicate bottom navigation code into component

## Future Considerations

### Architecture
- [ ] Microservices architecture for scaling
- [ ] Event-driven architecture for analytics
- [ ] CQRS pattern for complex queries
- [ ] WebSocket support for real-time features

### Business
- [ ] Subscription/pricing model
- [ ] White-label solution for gyms
- [ ] Mobile app monetization strategy
- [ ] API for third-party integrations

---

**Last Updated**: 2025-11-06
**Version**: 0.2.0-beta
