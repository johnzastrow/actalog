# Actionlog Requirements

This document outlines the system requirements for running Actionlog, a fitness tracker app focused on logging Crossfit workouts and history for weights and reps for particular movements or named weightlifting lifts. This will be a web-based application that will mostly be used from mobile phones, though it will need to be accessible from desktop browsers as well. It will be hosted on a small servers (Windows or Linux) with a database backend.

For optimal performance and user experience, the following requirements should be met:
The application should be lightweight and responsive, ensuring quick load times and smooth navigation on mobile devices. The user interface should be intuitive, allowing users to easily log their workouts and view their progress over time.

It will be multi-user, allowing individuals to create accounts and securely log in to access their personal workout data. The application should support user authentication and data privacy.

## Features

- User Authentication: Secure login and account management.
- Workout Logging: Ability to log various types of workouts, including weights, reps, and named Crossfit workouts.
- Progress Tracking: Visual representation of progress over time through charts and graphs.
- Mobile Optimization: Responsive design for seamless use on mobile devices.
- Data Export: Option to export workout data for personal records or sharing.
- Data Import: Ability to import workout data from other fitness apps or devices. also able to import workouts from a predefined list of common Crossfit workouts.

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

- Database: A relational database should be MariaDB, PostgreSQL, or SQLite to store user data and workout logs.
- Backend: The backend should be built using a robust framework based on Go to handle server-side logic and API endpoints.
- Frontend: The frontend should be developed using modern web technologies such as HTML5, CSS3, and JavaScript (with Vue.js and Vuetify) to ensure a responsive and user-friendly interface.
- Logging and Analytics: Implement logging mechanisms to track user activity and application performance for monitoring and debugging purposes.
- Security: Ensure that the application follows best practices for security, including data encryption, secure authentication, and protection against common vulnerabilities.
- Operating System: The application should be compatible with both Windows and Linux server environments.
- Web Server: A reliable web server such as Apache or Nginx should be optionally used to host the application in a more robust manner.
- APIs: The application should expose RESTful APIs for integration with other services and applications.
- Version Control: Use a version control system like Git to manage code changes and collaborate with other developers.
- Testing: Implement automated testing to ensure the reliability and stability of the application.
- Deployment: Use containerization (e.g., Docker) for easy deployment and scalability of the application and also support traditional deployment methods on a server.
- Documentation: Provide comprehensive documentation for both users and developers to facilitate usage and maintenance of the application.
- Backup and Recovery: Implement regular backup procedures and recovery mechanisms to protect user data.
- Scalability: Design the application to handle an increasing number of users and data without significant performance degradation.
- Accessibility: Ensure the application is accessible to users with disabilities by following web accessibility standards (e.g., WCAG).
-
