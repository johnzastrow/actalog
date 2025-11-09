

# Actalog Requirements

Design just a mobile-first app called ActaLog with just the core mobile app structure and features needed to log workouts per the requirements below. Focus on the essential components and user flows without going into extensive detail. Do not develop a marketing website or additional features beyond the core functionality needed for logging workouts.

This document outlines the system requirements for running Actalog, a fitness tracker app focused on logging Crossfit workouts and history for weights and reps for particular movements or named weightlifting lifts. This is a **Progressive Web App (PWA)** that will mostly be used from mobile phones, though it will be accessible from desktop browsers as well. It will be hosted on small servers (Windows or Linux) with a database backend.

## PWA Capabilities

ActaLog is built as a Progressive Web App, providing:

- **Installable**: Users can install the app to their home screen without app stores
- **Offline-First**: Full functionality even without internet connection
- **Fast Loading**: Instant loading with service worker caching
- **Automatic Updates**: Silent updates without manual app store processes
- **Cross-Platform**: Single codebase works on iOS, Android, and desktop
- **Push Notifications**: Workout reminders and achievement alerts (future)
- **Background Sync**: Automatic data synchronization when connection is restored

For optimal performance and user experience, the following requirements should be met:
The application should be lightweight and responsive, ensuring quick load times and smooth navigation on mobile devices. The user interface should be intuitive, allowing users to easily log their workouts and view their progress over time.

It will be multi-user, allowing individuals to create accounts and securely log in to access their personal workout data. The application should support user authentication and data privacy.

## Features

- **User Authentication**: Secure login and account management with JWT tokens.
- **Workout Logging**: Ability to log various types of workouts, including weights, reps, and named Crossfit workouts.
- **Progress Tracking**: Visual representation of progress over time through charts and graphs.
- **Mobile Optimization**: Responsive PWA design for seamless use on mobile devices.
- **Offline Support**: Full offline functionality with local data storage and background sync.
- **Installable App**: Add to home screen for native-like app experience.
- **Data Export**: Option to export workout data for personal records or sharing.
- **Data Import**: Ability to import workout data from other fitness apps or devices, plus predefined list of common Crossfit workouts.
- **Auto-Updates**: Automatic application updates via service worker.
- **Cached Performance**: Fast loading times through intelligent caching strategies.

## User Stories

1. As a user, I want to create an account so that I can securely log my workouts. My profile should store my personal information and picture.
2. As a user, I want to log my workouts easily so that I can track my fitness progress.
3. As a user, I want to view my workout history so that I can see my improvements over time. Views should include filters by date, workout type, and specific movements be shown a list, timeline and calendar views
4. As a user, I want to export my workout data so that I can keep a personal record. Exports should be in common formats like CSV and JSON.
5. as a user, I want to import workouts from other apps so that I can consolidate my fitness data in one place. Imports should be from CSV and JSON files.
6. As a user, I want to access the app from both my mobile phone and desktop so that I can log workouts conveniently.
7. As a user, I want to see visual representations of my progress so that I can stay motivated. Visuals should include charts and graphs for weights lifted, reps completed, and workout frequency.
8. As a user, I want to have a list of common Crossfit workouts and standard weightlifting movements available for use so that I can quickly log standard workouts without manual entry. The list should be customizable to add or remove workouts.
9. As a user, I want to be able to manually enter custom workouts so that I can log unique or personalized workout routines.
10. as an admin, I want to manage user accounts so that I can ensure the security and integrity of the application.
11. as an admin, I want to monitor application performance so that I can ensure a smooth user experience.
12. as an admin, I want to access user activity logs so that I can identify and address any potential issues or concerns.
13. as an admin, I want to back up user data regularly so that I can prevent data loss in case of system failures.
14. as an admin, I want to update the list of common Crossfit workouts and weightlifting movements so that users have access to the latest standards and trends in fitness.
15. as an admin, I want to manage application settings so that I can customize the user experience and ensure optimal performance.
16. as an admin, I want to generate reports on user activity and application usage so that I can make informed decisions for future improvements.
17. as an admin, I want to implement security measures so that user data is protected from unauthorized access.
18. as an admin, I want to provide support and assistance to users so that they have a positive experience with the application.
19. as an admin, I want to ensure compliance with data protection regulations so that user privacy is maintained.
20. as admin , I want to manage user roles and permissions so that access to sensitive features and data is controlled appropriately.
21. as an admin, I want to be able to export user data for analysis and reporting purposes.
22. as an admin, I want to be able to import user data from other systems to facilitate user onboarding and data migration.
23. as an admin, I want to be able to edit and delete user accounts and associated data so that I can manage the user base effectively.

## Technical Requirements

### Backend
- **Database**: MariaDB, PostgreSQL, or SQLite to store user data and workout logs.
- **Framework**: Go-based backend to handle server-side logic and RESTful API endpoints.
- **Logging and Analytics**: Structured logging with OpenTelemetry for monitoring and debugging.
- **Security**: Best practices including data encryption, JWT authentication, bcrypt password hashing, and protection against OWASP Top 10 vulnerabilities.

### Frontend (PWA)
- **Framework**: Vue.js 3 with Vuetify 3 for responsive UI components.
- **Service Worker**: Workbox-powered service worker for offline functionality and caching.
- **Web App Manifest**: Full PWA manifest with app metadata and icons.
- **Offline Storage**: IndexedDB for local data persistence and offline workouts.
- **Background Sync**: Automatic synchronization when connection is restored.
- **Progressive Enhancement**: Works on all devices, enhanced on modern browsers.
- **HTTPS**: Required in production for PWA features (service workers, install prompt).

### Infrastructure
- **Operating System**: Compatible with Windows and Linux server environments.
- **Web Server**: Nginx or Apache (optional reverse proxy for production).
- **SSL/TLS**: HTTPS required for PWA functionality (Let's Encrypt recommended).
- **APIs**: RESTful APIs with JWT authentication.
- **Version Control**: Git for source code management.
- **Testing**: Automated unit and integration tests for both frontend and backend.
- **Deployment**: Docker containerization plus traditional deployment support.
- **Documentation**: Comprehensive docs for users and developers.
- **Backup and Recovery**: Regular automated backups with point-in-time recovery.
- **Scalability**: Horizontal scaling support for multiple instances.
- **Accessibility**: WCAG 2.1 AA compliance for inclusive design.

### PWA-Specific Requirements
- **Manifest File**: Complete web app manifest with all required fields.
- **App Icons**: Multiple icon sizes (72px to 512px) for all platforms.
- **Offline Page**: Graceful offline experience with cached content.
- **Update Strategy**: Automatic updates with user notification for breaking changes.
- **Install Prompt**: Custom install promotion for better UX.
- **Lighthouse Score**: Target score of 90+ for PWA audit.

## UI Components: 
- Navigation Bar: Global navigation for product sections; includes links to Dashboard, Performance, Workouts, Profile, and Settings.
- Dashboard: Overview page with timeline of recent activity, and quick access to logging today's workout. 
- Large text fields should render markdown for formatting.
- Performance Page: Visual charts and graphs showing progress over time for the selected named workout or weight movement. The details for the selected movement should include a list with dates and details such as times and reps. Also a line chart with date along the X axis and weight or time for the Y axis. The list should show a star for the PRs for that movement in history.
- Workout Logging Page: Form to log a new workout, including fields for date (default to today), workout type (named WOD or custom), movements (select from common list or enter custom), weights, reps, and notes. Include a submit button to save the workout, reset button to clear the form, and a cancel button to return to the previous page without saving. also include a next button to add another movement to the same workout log.
- Profile Page: User profile management with fields for name, email, profile picture upload, and password change.
- Settings Page: Application settings including notification preferences, data export/import options, and account deletion.
- Help Page: FAQs, troubleshooting tips, and contact information for support.
- About Page: Information about the application, its purpose, and the development team.
- Terms of Service Page: Legal agreements and terms governing the use of the application.
- Privacy Policy Page: Information on data collection, usage, and user rights.
- Cookie Policy Page: Information on the use of cookies and tracking technologies.
- Accessibility Statement Page: Commitment to ensuring digital accessibility for users with disabilities.
- User Guide Page: Comprehensive documentation for users on how to use the application effectively.
- API Documentation Page: Technical documentation for developers on how to integrate with the application's API.
- Changelog Page: Record of changes, updates, and improvements made to the application over time.
- 

Visual Style: 
- Theme: Light theme with optional dark mode
- Primary color: #2c3657ff
- Secondary color: #597a6aff
- Accent color: #5a4e68ff
- Error/Alert: Red #DF3F40
- Spacing: Consistent 20px outer padding, 16px gutter spacing between items
- Borders: 1px solid light gray #E3E6EA on cards and input fields; slightly rounded corners (6px radius)
- Typography: Sans-serif, medium font weight (500) for headings, regular (400) for body, base size 16px
- Icons/images: Simple, filled vector icons for navigation and actions; illustrative flat images used occasionally for empty states
-


## Logical Data Model

* Each workout is composed of a warmup and zero or more strength movements, whose details include: weight, reps, and sets, plus zero or more WODs. 
* WODs are predefined combinations of activities that users can select from when logging their workouts. Users can also create custom WODs by specifying their own movements and details. Details of a WOD are Name, Source (Crossfit named workout, Other Coach, Self-recorded  with username), Type (Benchmark, Hero, Girl, Notables, Games, Endurance, Self-created), Regime (EMOM,AMRAP, Fastest Time, Slowest Round, Get Stronger, Skills), Score Type (Time [HH:MM:SS], Rounds and Reps [Rounds:Reps], Max Weight [Decimal] ), Description [Text], URL for Video or other online research [Text], Notes [Text].
  
A workout is created before logging

Said another way, the main entities in the data model are:
Workout = Warmup + WOD(s) + Strength Movement(s)

Each user can log multiple workouts each day, and each workout can include multiple strength movements and WODs. The system should also track user profiles, settings, and historical data for progress tracking. A workout is independent of other workouts and is not linked to users in the workout definition.

Each WOD can be linked to zero or more workouts, and each strength movement can also be linked to zero or more workouts. 


### Entities and Attributes

- **User**: id, name, email, password_hash, profile_picture_url, created_at, updated_at, updated_by
- **WOD**: id, name, source, type, regime, score, description, url, notes, created_at, updated_at, updated_by
- **Strength**: id, name, movement_type (weightlifting/cardio/gymnastics), created_at, updated_at, updated_by
- **Workout**: id, user_id (FK), date, notes, created_at, updated_at
- **WorkoutWOD**: id, workout_id (FK), wod_id (FK), created_at, updated_at
- **WorkoutStrength**: id, workout_id (FK), strength_id (FK), weight, reps, sets, created_at, updated_at 
- **UserWorkout**: id, user_id (FK), workout_id (FK), created_at, updated_at
- **UserSetting**: id, user_id (FK), notification_preferences, data_export_format, created_at, updated_at  
- **AuditLog**: id, user_id (FK), action, timestamp, details
- **Backup**: id, backup_date, status, file_location


### Relationships
- A **User** can log multiple **Workouts** and a workout may be used by many users on multiple days (many-to-Many) via UserWorkouts.
- A **Workout** can include multiple **Strength Movements** (Many-to-Many via WorkoutStrength).
- A **Strength Movement** can be included in multiple **Workouts** (Many-to-Many via WorkoutStrength).
- A **User** can have multiple **UserSettings**.
- A **User** can have multiple **AuditLogs**.
- A **User** can have multiple **Backups**.
- A UserWorkout links a User to a Workout they have logged.
- A **User** can have multiple **UserWorkouts**.
- A **Workout** can have multiple **UserWorkouts**.
- A **UserWorkout** links a User to a Workout they have logged.
- A **Workout** can include multiple **WODs** (Many-to-Many via WorkoutWOD).
- A **WOD** can be included in multiple **Workouts** (Many-to-Many via WorkoutWOD).
- A **User** can create multiple **WODs**.
- A **WOD** can be created by a **User**.
- A **User** can create multiple **Strength Movements**.
- A **Strength Movement** can be created by a **User**.
- Each entity should have created_at, updated_at, and updated_by fields for auditing purposes.
- Each entity should have a unique identifier (id) as the primary key.
- each WorkoutWod and WorkoutStrength should have a unique identifier (id) as the primary key. 
- Each WOD in a UserWorkout should allow a score to be recorded based on the score type defined in the WOD entity.
- Each WorkoutStrength in a UserWorkout should allow weight, reps, and sets to be recorded.
- 
- Each UserSetting should be linked to a specific User via user_id foreign key.
- 
- ### Additional Considerations
- Foreign keys (FK) should be used to establish relationships between entities.
- Indexes should be created on frequently queried fields for performance optimization.
- Data integrity constraints should be enforced to maintain consistency across related entities.
- Consideration for scalability and future expansion of the data model should be taken into account during design.
- Normalization should be applied to reduce data redundancy and improve data integrity.
- Backup and recovery mechanisms should be implemented to protect against data loss.
- Security measures should be in place to protect sensitive user data, including encryption and access controls.
- Audit logging should capture significant actions performed by users for accountability and troubleshooting.



