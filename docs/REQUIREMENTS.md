

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

## Design Decisions

The following design decisions define the core functionality and user experience of ActaLog:

### 1. Workout Template System
**Decision**: Hybrid Approach
- Users can use pre-defined WODs and admin-created workout templates
- Users can also create and save their own custom workout templates
- This provides flexibility while maintaining a curated library of standard CrossFit workouts

### 2. Personal Records (PR) Tracking
**Decision**: Auto-Detection with Manual Override
- System automatically detects PRs by comparing workout history (weight lifted, time, rounds+reps)
- Users can manually flag or unflag workouts as PRs for accuracy
- PR badges appear on workout cards in history and dashboard
- Movement detail screens show PR history with star indicators (⭐)

### 3. Social & Sharing Features
**Decision**: Public Leaderboards with Scaled Divisions
- **Leaderboard Structure**:
  - Separate leaderboards for Rx (as prescribed), Scaled (modified), and Beginner divisions
  - Users self-select their division when logging WOD scores
  - Global leaderboards for each standard WOD (benchmark, hero, girl workouts)
- **Privacy**: Email verification required to appear on leaderboards
- **No friend system initially**: Focus on global community competition
- **Future consideration**: Web Share API for social media sharing

### 4. Data Import/Export Formats
**Decision**: CSV, JSON, and Markdown Support
- **CSV**: For spreadsheet compatibility and data analysis in Excel/Google Sheets
- **JSON**: Complete structured data backup for full restore capability
- **Markdown**: For formatted workout reports and notes
- Date range selection for partial exports
- Import with data preview before confirmation

### 5. User Roles & Permissions
**Decision**: Simple Two-Tier System
- **Regular Users**: Can create workouts, log exercises, manage their own data
- **Admins**: Can manage all user accounts, create/edit standard WODs and movements, access audit logs
- **First user becomes admin** automatically
- No coach or gym owner roles initially (future consideration for multi-gym support)

### 6. Workout Scheduling
**Decision**: Future Planning Support
- Users can schedule workouts on future dates using the calendar
- No push notifications initially (PWA notification infrastructure in place for future)
- No dedicated rest day tracking (users simply don't log workouts on rest days)
- Calendar view shows scheduled vs. completed workouts with different visual indicators

### 7. Strength Movement Library
**Decision**: Hybrid Predefined + Custom Library
- App includes predefined library of standard CrossFit movements (back squat, deadlift, clean & jerk, etc.)
- Users can create custom movements for specialized exercises
- Both standard and custom movements searchable and filterable by type (weightlifting, cardio, gymnastics)
- Admin can add to standard library; users cannot edit standard movements

### 8. Offline Data Sync Conflict Resolution
**Decision**: Last Write Wins by Timestamp
- When offline data conflicts with server during sync, most recent timestamp takes precedence
- Suitable for single-user workout logging (rarely conflicts)
- Sync status indicator shows pending operations
- Future enhancement: conflict detection UI for edge cases

### 9. Performance Analytics & Metrics
**Decision**: Three Primary Visualizations
- **Weight Progression Charts**: Line graphs showing strength gains over time for specific movements (e.g., back squat 1RM)
- **Workout Frequency Heatmap**: Calendar heatmap showing workout consistency and streaks
- **WOD Leaderboards**: Global rankings for benchmark WODs with division filters
- Not initially: Volume tracking (total weight × reps), though data model supports future addition

### 10. Email Verification
**Decision**: Optional Verification with Feature Unlock
- Users can immediately use core features (logging workouts, tracking progress) without email verification
- Email verification unlocks:
  - Public leaderboard participation
  - Data export functionality
  - Future: Workout sharing and social features
- Verification email sent on registration, can be resent from profile
- Balances security with user experience

### 11. Workout Notes & Documentation
**Decision**: Per-Workout Notes with Markdown
- Each logged workout can have notes field (how it felt, modifications made, injuries, etc.)
- Notes support markdown formatting for rich text (bold, lists, links)
- Rendered as formatted text in workout detail view
- Not initially: Movement-specific notes, photo/video attachments, voice notes (future consideration)

### 12. Leaderboard Divisions & Scoring
**Decision**: Three-Division System
- **Rx (As Prescribed)**: Workout performed exactly as specified (weight, reps, movements)
- **Scaled**: Workout modified (lighter weight, fewer reps, substitute movements)
- **Beginner**: Simplified version for newer athletes
- Users self-report division when logging WOD
- Separate leaderboards prevent unfair comparisons
- Future: Admin verification for top leaderboard entries

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

## Screens & Navigation Flow

### Authentication Screens (Public)

#### 1. Login Screen
- **Route**: `/login`
- **Purpose**: User authentication
- **Components**:
  - Email input field
  - Password input field
  - "Remember Me" checkbox
  - Login button
  - "Forgot Password?" link
  - "Create Account" link to Register screen
- **Navigation**:
  - Success → Dashboard
  - Create Account → Register Screen
  - Forgot Password → Password Reset Screen (future)

#### 2. Register Screen
- **Route**: `/register`
- **Purpose**: New user account creation
- **Components**:
  - Name input field
  - Email input field
  - Password input field
  - Confirm Password input field
  - Terms & Conditions checkbox
  - Register button
  - "Already have an account?" link to Login
- **Navigation**:
  - Success → Dashboard (auto-login)
  - Login link → Login Screen

### Main Application Screens (Authenticated)

#### 3. Dashboard
- **Route**: `/dashboard`
- **Purpose**: Overview of recent activity and quick actions
- **Components** (as per design/dashboard.png):
  - Top app bar with logo and current date
  - Monthly calendar view showing workout days (colored circles)
  - "Workouts in last 7 Days" cards section
  - Bottom navigation bar
- **Navigation**:
  - Calendar date click → Filter workouts by date or show day detail
  - Workout card click → Workout Detail Screen
  - Bottom nav → Performance, Log Workout, Workouts, Profile

#### 4. Log Workout Screen
- **Route**: `/workouts/log`
- **Purpose**: Create new workout or log existing workout template
- **Components**:
  - Date picker (default: today)
  - Workout template selector or "Create New" option
  - **If creating template:**
    - Workout name input
    - Add WOD section (search/select from library or create custom)
    - Add Strength Movement section (search/select from library)
    - For each movement: weight, sets, reps inputs
    - Order/sequence controls
  - **If logging existing template:**
    - Show template structure (read-only)
    - For each WOD: score input field (time, rounds+reps, or weight based on score_type)
    - For each strength movement: weight, sets, reps inputs
  - Notes field (markdown supported)
  - Save button
  - Cancel button
- **Navigation**:
  - Save → Dashboard or Workouts Screen
  - Cancel → Previous screen
  - Add WOD → WOD Library Screen (modal/overlay)
  - Add Movement → Movement Library Screen (modal/overlay)

#### 5. Workouts Screen (History)
- **Route**: `/workouts`
- **Purpose**: View all logged workouts with filtering and search
- **Components**:
  - Search bar
  - Filter controls (date range, workout type, WOD name)
  - Sort options (date, name, type)
  - Workout list/cards grouped by date
  - PR badges on personal record workouts
  - "Log Workout" FAB button
- **Navigation**:
  - Workout card click → Workout Detail Screen
  - FAB button → Log Workout Screen
  - Bottom nav → Dashboard, Performance, Profile

#### 6. Workout Detail Screen
- **Route**: `/workouts/:id`
- **Purpose**: View details of a specific logged workout
- **Components**:
  - Workout date and name
  - WODs performed with scores
  - Strength movements with weight/sets/reps
  - Notes (rendered as markdown)
  - Edit button
  - Delete button
  - Share button (future - Web Share API)
- **Navigation**:
  - Back → Workouts Screen
  - Edit → Edit Workout Screen (similar to Log Workout)
  - Delete → Confirmation dialog → Workouts Screen

#### 7. Performance/Progress Screen
- **Route**: `/performance`
- **Purpose**: Visualize progress over time for specific movements or WODs
- **Components**:
  - Movement/WOD selector dropdown
  - Date range selector
  - Line chart (X: date, Y: weight or time)
  - Data table with dates, values, and PR indicators (⭐)
  - Filter by movement type (weightlifting, cardio, gymnastics)
  - Export data button
- **Navigation**:
  - Movement click → Filter chart for that movement
  - Bottom nav → Dashboard, Workouts, Profile

#### 8. WOD Library Screen
- **Route**: `/wods`
- **Purpose**: Browse and search standard and custom WODs
- **Components**:
  - Search bar
  - Filter by type (Benchmark, Hero, Girl, Games, Custom)
  - Filter by source (CrossFit, Other Coach, Self-recorded)
  - WOD cards showing name, type, and description preview
  - "Create Custom WOD" button
- **Navigation**:
  - WOD card click → WOD Detail Screen
  - Create button → Create WOD Screen
  - Select WOD (when opened as modal from Log Workout) → Returns to Log Workout with WOD added

#### 9. WOD Detail Screen
- **Route**: `/wods/:id`
- **Purpose**: View full WOD information
- **Components**:
  - WOD name and type badges
  - Source and regime
  - Score type
  - Full description (markdown)
  - Video URL (if available)
  - Notes
  - "Use This WOD" button
  - Edit button (if custom WOD created by user)
  - Delete button (if custom WOD)
- **Navigation**:
  - Back → WOD Library
  - Use This → Log Workout Screen with WOD pre-selected
  - Edit → Edit WOD Screen (if custom)

#### 10. Create/Edit WOD Screen
- **Route**: `/wods/create` or `/wods/:id/edit`
- **Purpose**: Create custom WOD or edit user-created WOD
- **Components**:
  - Name input
  - Source dropdown (CrossFit, Other Coach, Self-recorded)
  - Type dropdown (Benchmark, Hero, Girl, etc.)
  - Regime dropdown (EMOM, AMRAP, Fastest Time, etc.)
  - Score Type dropdown (Time, Rounds+Reps, Max Weight)
  - Description textarea (markdown supported)
  - URL input (optional video/reference)
  - Notes textarea
  - Save button
  - Cancel button
- **Navigation**:
  - Save → WOD Detail Screen or WOD Library
  - Cancel → Previous screen

#### 11. Movement Library Screen
- **Route**: `/movements`
- **Purpose**: Browse standard and custom strength movements
- **Components**:
  - Search bar
  - Filter by type (weightlifting, cardio, gymnastics)
  - Movement list/cards
  - "Create Custom Movement" button
- **Navigation**:
  - Movement click → Movement Detail Screen
  - Create button → Create Movement Screen
  - Select movement (when opened as modal) → Returns to Log Workout with movement added

#### 12. Movement Detail Screen
- **Route**: `/movements/:id`
- **Purpose**: View movement information and personal history
- **Components**:
  - Movement name and type
  - Description
  - Personal best (PR)
  - History of this movement in workouts
  - Progress chart
  - "Use This Movement" button
  - Edit button (if custom)
- **Navigation**:
  - Back → Movement Library
  - Use This → Log Workout Screen
  - Edit → Edit Movement Screen (if custom)

#### 13. Create/Edit Movement Screen
- **Route**: `/movements/create` or `/movements/:id/edit`
- **Purpose**: Create custom movement
- **Components**:
  - Name input
  - Type dropdown (weightlifting, cardio, gymnastics)
  - Description textarea
  - Save button
  - Cancel button
- **Navigation**:
  - Save → Movement Detail or Movement Library
  - Cancel → Previous screen

#### 14. Settings Menu Screen (Flyout/Main Menu)
- **Route**: `/settings`
- **Purpose**: Main settings menu with links to all management screens
- **Access**: User Avatar icon in bottom navigation
- **Components** (List/Menu):
  - **Profile** → Profile Screen
  - **My WODs** → WOD Management Screen
  - **My Strength Movements** → Strength Management Screen
  - **My Workout Templates** → Workout Templates Screen
  - **Import Data** → Import Screen
  - **Export Data** → Export Screen
  - **App Preferences** → App Preferences Screen
  - **Help** → Help Screen
  - **About** → About Screen
  - **Logout** button
- **Navigation**:
  - Each menu item → Respective screen
  - Logout → Login Screen
  - Close/Back → Previous screen

#### 15. Profile Screen
- **Route**: `/settings/profile`
- **Purpose**: View and edit user profile
- **Components**:
  - Profile picture with upload button
  - Name input (editable)
  - Email (display only)
  - Birthday input (date picker)
  - Member since date (read-only)
  - Total workouts count (read-only)
  - Personal records count (read-only)
  - "Change Password" button
  - "Save Changes" button
  - "Cancel" button
- **Navigation**:
  - Change Password → Change Password Screen
  - Save → Settings Menu
  - Cancel → Settings Menu
  - Back → Settings Menu

#### 16. Change Password Screen
- **Route**: `/settings/profile/change-password`
- **Purpose**: Update user password
- **Components**:
  - Current password input
  - New password input
  - Confirm new password input
  - "Update Password" button
  - "Cancel" button
- **Navigation**:
  - Update → Profile Screen (with success message)
  - Cancel → Profile Screen

#### 17. WOD Management Screen
- **Route**: `/settings/wods`
- **Purpose**: Manage user's custom WODs
- **Components**:
  - Search bar
  - Filter by type (All, My Custom, Standard)
  - WOD list/cards showing:
    - Name, type, source
    - Edit icon (only for user's custom WODs)
    - Delete icon (only for user's custom WODs)
  - "Create New WOD" FAB button
- **Navigation**:
  - WOD card click → WOD Detail Screen (view only for standard, edit for custom)
  - Edit icon → Edit WOD Screen
  - Delete icon → Confirmation dialog → Delete → Refresh list
  - Create button → Create WOD Screen
  - Back → Settings Menu

#### 18. Create WOD Screen
- **Route**: `/settings/wods/create`
- **Purpose**: Create custom WOD
- **Components**:
  - Name input (required)
  - Source dropdown (CrossFit, Other Coach, Self-recorded)
  - Type dropdown (Benchmark, Hero, Girl, Notables, Games, Endurance, Self-created)
  - Regime dropdown (EMOM, AMRAP, Fastest Time, Slowest Round, Get Stronger, Skills)
  - Score Type dropdown (Time, Rounds+Reps, Max Weight)
  - Description textarea (markdown supported)
  - URL input (optional video/reference)
  - Notes textarea
  - "Save WOD" button
  - "Cancel" button
- **Navigation**:
  - Save → WOD Management Screen (with success message)
  - Cancel → WOD Management Screen

#### 19. Edit WOD Screen
- **Route**: `/settings/wods/:id/edit`
- **Purpose**: Edit user's custom WOD
- **Components**: Same as Create WOD Screen with pre-filled values
- **Navigation**:
  - Save → WOD Management Screen (with success message)
  - Cancel → WOD Management Screen
  - Delete → Confirmation dialog → WOD Management Screen

#### 20. Strength Management Screen
- **Route**: `/settings/movements`
- **Purpose**: Manage user's custom strength movements
- **Components**:
  - Search bar
  - Filter by type (All, Weightlifting, Cardio, Gymnastics, My Custom, Standard)
  - Movement list/cards showing:
    - Name, type
    - Edit icon (only for user's custom movements)
    - Delete icon (only for user's custom movements)
  - "Create New Movement" FAB button
- **Navigation**:
  - Movement card click → Movement Detail Screen
  - Edit icon → Edit Movement Screen
  - Delete icon → Confirmation dialog → Delete → Refresh list
  - Create button → Create Movement Screen
  - Back → Settings Menu

#### 21. Create Movement Screen
- **Route**: `/settings/movements/create`
- **Purpose**: Create custom strength movement
- **Components**:
  - Name input (required)
  - Type dropdown (Weightlifting, Cardio, Gymnastics)
  - Description textarea
  - "Save Movement" button
  - "Cancel" button
- **Navigation**:
  - Save → Strength Management Screen (with success message)
  - Cancel → Strength Management Screen

#### 22. Edit Movement Screen
- **Route**: `/settings/movements/:id/edit`
- **Purpose**: Edit user's custom movement
- **Components**: Same as Create Movement Screen with pre-filled values
- **Navigation**:
  - Save → Strength Management Screen (with success message)
  - Cancel → Strength Management Screen
  - Delete → Confirmation dialog → Strength Management Screen

#### 23. Workout Templates Screen
- **Route**: `/settings/workouts`
- **Purpose**: Manage saved workout templates
- **Components**:
  - Search bar
  - Filter controls
  - Template list/cards showing:
    - Template name
    - WODs and movements included
    - Times used count
    - Edit icon
    - Delete icon
  - "Create New Template" FAB button
- **Navigation**:
  - Template card click → Template Detail Screen
  - Edit icon → Edit Template Screen
  - Delete icon → Confirmation dialog → Delete → Refresh list
  - Create button → Create Template Screen
  - Back → Settings Menu

#### 24. Create Workout Template Screen
- **Route**: `/settings/workouts/create`
- **Purpose**: Create reusable workout template
- **Components**:
  - Template name input
  - "Add WOD" button → WOD Library (modal)
  - "Add Strength Movement" button → Movement Library (modal)
  - Order/sequence drag controls
  - List of added WODs and movements
  - Remove buttons for each item
  - Notes textarea
  - "Save Template" button
  - "Cancel" button
- **Navigation**:
  - Save → Workout Templates Screen
  - Cancel → Workout Templates Screen
  - Add WOD → WOD Library Modal → Select → Return
  - Add Movement → Movement Library Modal → Select → Return

#### 25. Edit Workout Template Screen
- **Route**: `/settings/workouts/:id/edit`
- **Purpose**: Edit existing workout template
- **Components**: Same as Create Workout Template Screen with pre-filled values
- **Navigation**:
  - Save → Workout Templates Screen
  - Cancel → Workout Templates Screen
  - Delete → Confirmation dialog → Workout Templates Screen

#### 26. Import Data Screen
- **Route**: `/settings/import`
- **Purpose**: Import workout data from files
- **Components**:
  - File upload dropzone
  - Supported formats info (CSV, JSON)
  - Preview imported data table
  - "Import" button
  - "Cancel" button
  - Import history list
- **Navigation**:
  - Import → Processing → Success message → Settings Menu
  - Cancel → Settings Menu
  - Back → Settings Menu

#### 27. Export Data Screen
- **Route**: `/settings/export`
- **Purpose**: Export user data
- **Components**:
  - Export format selector (CSV, JSON)
  - Date range selector
  - Data type checkboxes (Workouts, WODs, Movements, Profile)
  - "Export" button
  - "Cancel" button
  - Export history list with download links
- **Navigation**:
  - Export → Generate file → Download
  - Cancel → Settings Menu
  - Back → Settings Menu

#### 28. App Preferences Screen
- **Route**: `/settings/preferences`
- **Purpose**: Application settings and preferences
- **Components**:
  - **Appearance Section:**
    - Theme toggle (light/dark)
  - **Units Section:**
    - Weight unit (lbs/kg)
    - Distance unit (miles/km)
  - **Notifications Section:**
    - Push notification toggle
    - Workout reminder toggle and time
  - **Account Section:**
    - Delete account button (with confirmation)
  - **About Section:**
    - Version number
    - Privacy Policy link
    - Terms of Service link
- **Navigation**:
  - Back → Settings Menu
  - Privacy Policy → Privacy Policy Screen
  - Terms → Terms of Service Screen

### Admin Screens (Admin Role Only)

#### 29. Admin Dashboard
- **Route**: `/admin`
- **Purpose**: Admin overview and management
- **Components**:
  - User statistics
  - System health
  - Recent activity
  - Quick links to management screens
- **Navigation**:
  - User Management → Admin Users Screen
  - System Logs → Audit Logs Screen

#### 30. Admin Users Screen
- **Route**: `/admin/users`
- **Purpose**: User management
- **Components**:
  - User search and filter
  - User list with roles
  - Edit/delete actions
  - Export user data
- **Navigation**:
  - Edit user → Edit User Screen
  - Back → Admin Dashboard

### Informational Screens

#### 31. Help/FAQ Screen
- **Route**: `/help`
- **Purpose**: User assistance
- **Components**:
  - FAQ accordion
  - Search bar
  - Contact support button
  - User guide link

#### 32. Privacy Policy Screen
- **Route**: `/privacy`
- **Purpose**: Privacy policy
- **Components**:
  - Privacy policy text
  - Last updated date

#### 33. Terms of Service Screen
- **Route**: `/terms`
- **Purpose**: Terms of service
- **Components**:
  - Terms text
  - Last updated date

### Navigation Flow Diagram

```
Login/Register
    ↓
Dashboard (Hub)
    ├→ Log Workout → WOD Library (modal)
    │               → Movement Library (modal)
    │               → Save → Dashboard
    │
    ├→ Workouts → Workout Detail → Edit/Delete
    │
    ├→ Performance → Movement Detail → Use in Workout
    │
    └→ Settings Menu (Flyout)
        ├→ Profile → Change Password
        ├→ My WODs → Create/Edit/Delete WOD
        ├→ My Strength Movements → Create/Edit/Delete Movement
        ├→ My Workout Templates → Create/Edit/Delete Template
        ├→ Import Data
        ├→ Export Data
        ├→ App Preferences
        ├→ Help
        ├→ About
        └→ Logout → Login

Admin: Settings Menu → Admin Dashboard → User Management
                                       → Audit Logs
```

### Bottom Navigation (Always Visible When Authenticated)

1. **Dashboard** - Home icon
2. **Performance** - Chart icon
3. **Log Workout** - Large center FAB with plus icon
4. **Workouts** - Dumbbell icon
5. **Settings** - Avatar or account icon (opens Settings Menu flyout)

### Screen Count Summary

- **Authentication**: 2 screens (Login, Register)
- **Main App**: 26 screens (Dashboard through App Preferences)
  - Core: 7 screens (Dashboard, Log Workout, Workouts, Workout Detail, Performance, Movement Library, Movement Detail)
  - Settings Hub: 1 screen (Settings Menu)
  - Profile Management: 2 screens (Profile, Change Password)
  - WOD Management: 3 screens (WOD List, Create, Edit)
  - Strength Management: 3 screens (Movement List, Create, Edit)
  - Workout Templates: 3 screens (Template List, Create, Edit)
  - Data Management: 2 screens (Import, Export)
  - App Settings: 1 screen (Preferences)
  - Legacy Screens: 4 screens (WOD Library, WOD Detail, Create/Edit WOD from library, Create/Edit Movement from library)
- **Admin**: 2 screens
- **Informational**: 3 screens (Help, Privacy, Terms)
- **Total**: 33 core screens

### PWA-Specific Screens

#### Install Prompt (Progressive Enhancement)
- Custom install prompt UI when PWA installability criteria met
- Shows on first visit or after user engagement
- "Install App" and "Not Now" buttons

#### Offline Indicator
- Toast/banner notification when app goes offline
- Shows cached data availability
- Sync status indicator for pending workouts


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



#### Import and Export
Import and export functions should allow round-tripping key data entities data through CSV files, in addition to other more comprehensive functions through JSON and copying data through SQLite copies. Implement import and export tools via screens from the Profile menu. The administrator should be able to export global data, but normal users should only be able to import and export data visible to them.
1. Import CSV of WODs, and export to the same format. Transform the CSV for WODs in the seeds/ directory to match what you would need to export to round-trip the data. The add seeding the database of a new database with these data. Do the same process for the movement CSV
2. Export and Import of User Workouts with enough information to rebuild the associated workouts/WODs and Movements. Might need to be in JSON format
3. Export of Workouts in CSV format for use in spreadsheets. Flatten out the workout details form other tables into a flattened CSV file
4. Export of Performance information for each WOD and Movement into CSV format, with flags for PRs


#### Database Backup and Restore System

A comprehensive backup and restore system allows administrators to create full database backups and restore them when needed. This feature is critical for disaster recovery, data migration between installations, and changing database technologies.

**User Story:**
- As an admin, I want to create full database backups so that I can restore the entire application in case of data loss or system failure.
- As an admin, I want to download backups to my local device so that I have offline copies for safekeeping.
- As an admin, I want to restore from a backup so that I can recover from data loss or migrate to a new server.
- As an admin, I want to list all available backups with metadata so that I can identify which backup to restore.
- As an admin, I want to delete old backups so that I can manage storage space on the server.

**Phase 1 Requirements (v0.6.0):**

1. **Backup Creation (Admin Only)**
   - Manual backup trigger via web UI ("Create Backup" button)
   - Backup includes ALL application data:
     - All database tables (users, workouts, movements, WODs, performance data, audit logs, settings, etc.)
     - Uploaded files (avatars, attachments stored in `uploads/` directory)
     - Application configuration files (`.env` settings, relevant config)
   - Backup formats:
     - **Primary format**: JSON (portable, database-agnostic, allows migration between database technologies)
     - **Optional format**: SQLite file (if SQLite is the active database, offer direct database file download)
   - Compression: All backups packaged as `.zip` files for reduced storage
   - No encryption yet (passwords already hashed in database)
   - Storage location: Server filesystem in `backups/` directory
   - Backup filename format: `actalog_backup_YYYY-MM-DD_HH-MM-SS.zip`

2. **Backup Metadata**
   - Each backup includes a `metadata.json` file with:
     - Backup creation date/time
     - Application version
     - Database driver used (sqlite3, postgres, mysql)
     - Admin user who created backup
     - Data date range (earliest and latest workout dates)
     - Total record counts (users, workouts, movements, WODs)
     - Backup file size
     - Database schema version

3. **Backup Management UI (Admin Only)**
   - **Location**: `/admin/backups` route
   - **List View**: Table showing all available backups with:
     - Creation date/time
     - File size (human-readable: KB, MB, GB)
     - Data date range
     - Created by (admin username)
     - Record counts summary
     - Actions: Download, Delete
   - **Actions**:
     - **Create Backup**: Button triggers backup creation with progress indicator
     - **Download Backup**: Download `.zip` file to admin's device
     - **Delete Backup**: Delete backup file from server (with confirmation dialog)
   - **No preview**: Backups cannot be previewed before restore (too complex for Phase 1)

4. **Restore Process (Admin Only) - Full Restore**
   - **Phase 1**: Full restore only (wipes existing database and replaces with backup data)
   - **Restore workflow**:
     1. Admin selects backup from list
     2. Confirmation dialog warns: "This will DELETE all current data and replace with backup data. Are you sure?"
     3. Admin uploads backup `.zip` file (if not already on server)
     4. System validates backup file (checks metadata, file integrity)
     5. System stops application (maintenance mode)
     6. System wipes current database tables
     7. System restores all data from JSON backup (parses and inserts into current database driver)
     8. System restores uploaded files to `uploads/` directory
     9. System restarts application
     10. Admin sees success message with summary (records restored, files restored)
   - **Duplicate handling (Phase 1)**: Overwrite all existing data (full wipe and replace)
   - **Database migration support**: JSON format allows restoring into different database technology than original backup

5. **Multi-Database Support**
   - Backup/restore must work with all supported database drivers:
     - SQLite (`sqlite3`)
     - PostgreSQL (`postgres`)
     - MySQL/MariaDB (`mysql`)
   - JSON format ensures portability between database types
   - SQLite file download only available when SQLite is active database

6. **Access Control**
   - **Admin only**: Only users with `role = admin` can access backup/restore features
   - Route protected by `middleware.AdminOnly`
   - Non-admin users see 403 Forbidden if they attempt access

7. **Error Handling**
   - Backup creation errors: Disk space, file permissions, database read errors
   - Restore errors: Invalid backup file, corrupted data, database write errors, schema mismatch
   - User-friendly error messages with action hints
   - Rollback mechanism if restore fails mid-process (restore database from pre-restore snapshot)

**Future Enhancements (v0.7.0+):**

1. **Selective Restore**
   - Restore specific tables only (e.g., just workouts, just users)
   - Restore data from specific date ranges
   - Choose which data to include/exclude during restore

2. **Merge Mode Restore**
   - Import backup data into existing database without wiping
   - Smart duplicate detection and resolution
   - User chooses: Skip, Overwrite, or Rename duplicates

3. **Automated Backups**
   - Scheduled automatic backups (daily, weekly, monthly)
   - Configurable retention policy (keep last N backups, auto-delete old)
   - Email notifications when backup completes or fails

4. **Encryption**
   - Password-protected backups
   - Encryption at rest for backup files
   - Secure key management

5. **Remote Storage**
   - Upload backups to cloud storage (S3, Google Drive, Dropbox)
   - Remote backup retention policy
   - Download from cloud for restore

6. **Point-in-Time Recovery (Advanced)**
   - Transaction log backups for databases that support it
   - Restore database to exact point in time
   - Incremental backups between full backups

7. **CLI Tool**
   - Command-line interface for backup/restore operations
   - Useful for server administrators and automation
   - Example: `actalog backup create`, `actalog backup restore <file>`

**Technical Implementation Notes:**

- **JSON Structure** for backup:
  ```json
  {
    "metadata": {
      "backup_version": "1.0",
      "app_version": "0.6.0",
      "created_at": "2025-01-21T10:30:00Z",
      "database_driver": "sqlite3",
      "created_by": "admin@example.com",
      "data_date_range": {
        "earliest_workout": "2024-01-01",
        "latest_workout": "2025-01-21"
      },
      "record_counts": {
        "users": 150,
        "workouts": 3420,
        "movements": 45,
        "wods": 120,
        "audit_logs": 8950
      },
      "schema_version": "0.5.0"
    },
    "users": [...],
    "movements": [...],
    "wods": [...],
    "workouts": [...],
    "user_workouts": [...],
    "user_workout_movements": [...],
    "user_workout_wods": [...],
    "audit_logs": [...],
    "user_settings": [...]
  }
  ```

- **File Structure** in backup `.zip`:
  ```
  actalog_backup_2025-01-21_10-30-00.zip
  ├── metadata.json
  ├── database.json (all tables)
  ├── database.db (optional, if SQLite)
  ├── config/
  │   └── .env.backup
  └── uploads/
      ├── avatars/
      └── attachments/
  ```

- **Service Layer** (`internal/service/backup_service.go`):
  - `CreateBackup(adminUserID)` - Creates full backup
  - `ListBackups()` - Returns list of available backups with metadata
  - `GetBackupMetadata(filename)` - Reads metadata from backup file
  - `RestoreBackup(filename, adminUserID)` - Full restore from backup
  - `DeleteBackup(filename, adminUserID)` - Deletes backup file

- **Repository Layer**: No new repositories needed (uses existing repos to read/write data)

- **Handler Layer** (`internal/handler/backup_handler.go`):
  - `POST /api/admin/backups` - Create backup
  - `GET /api/admin/backups` - List backups
  - `GET /api/admin/backups/{filename}` - Download backup
  - `DELETE /api/admin/backups/{filename}` - Delete backup
  - `POST /api/admin/backups/{filename}/restore` - Restore from backup

- **Frontend** (`web/src/views/AdminBackupsView.vue`):
  - Backup management interface
  - Create, list, download, delete backups
  - Restore with confirmation dialogs

**Dependencies:**
- `archive/zip` (Go standard library) - ZIP compression
- Existing repositories for data access
- File system utilities for file operations

