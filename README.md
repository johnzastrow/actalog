# ActaLog

> A mobile-first fitness tracker for CrossFit enthusiasts to log workouts, track progress, and analyze performance.

[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)](https://golang.org)
[![Vue.js](https://img.shields.io/badge/Vue.js-3.x-4FC08D?style=flat&logo=vue.js)](https://vuejs.org)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

## Overview

ActaLog is a fitness tracker app focused on logging CrossFit workouts and tracking history for weights, reps, and named weightlifting lifts. Built with a Go backend and Vue.js/Vuetify frontend, it provides a clean, mobile-first interface for tracking your fitness journey.

**Version:** 0.1.0-alpha

## Features

- 🏋️ **Workout Logging**: Track named WODs and custom workouts
- 📊 **Progress Tracking**: Visual charts and graphs showing improvement over time
- 📱 **Mobile-First**: Optimized for mobile with responsive desktop support
- 👤 **Multi-User**: Secure authentication and personal workout data
- 📤 **Data Export/Import**: CSV and JSON support
- 🎨 **Clean Design**: Material Design UI with Vuetify
- 🔒 **Secure**: JWT authentication, bcrypt password hashing

## Technology Stack

### Backend
- **Language**: Go 1.21+
- **Router**: Chi
- **Database**: SQLite (dev), PostgreSQL (prod), MariaDB (supported)
- **Authentication**: JWT with golang-jwt/jwt
- **ORM**: sqlx
- **Testing**: testify

### Frontend
- **Framework**: Vue.js 3
- **UI Library**: Vuetify 3
- **State Management**: Pinia
- **Build Tool**: Vite
- **Charts**: Chart.js with vue-chartjs

### Infrastructure
- **Containerization**: Docker + Docker Compose
- **Database Migrations**: golang-migrate
- **Reverse Proxy**: Nginx (optional)

## Quick Start

### Prerequisites

- Go 1.21 or higher
- Node.js 18+ and npm
- Docker and Docker Compose (optional)

### Local Development

1. **Clone the repository**
   ```bash
   git clone https://github.com/johnzastrow/actalog.git
   cd actalog
   ```

2. **Set up environment variables**
   ```bash
   cp .env.example .env
   # Edit .env with your configuration
   ```

3. **Install Go dependencies**
   ```bash
   go mod download
   ```

4. **Install frontend dependencies**
   ```bash
   cd web
   npm install
   cd ..
   ```

5. **Run the backend**
   ```bash
   # Terminal 1
   make run
   # Or: go run cmd/actalog/main.go
   ```

6. **Run the frontend**
   ```bash
   # Terminal 2
   cd web
   npm run dev
   ```

7. **Access the application**
   - Frontend: http://localhost:3000
   - Backend API: http://localhost:8080
   - API Health: http://localhost:8080/health

### Using Docker

```bash
# Start all services
make docker-up

# Stop all services
make docker-down

# View logs
make docker-logs
```

## Project Structure

```
actalog/
├── cmd/actalog/          # Application entry point
├── internal/             # Private application code
│   ├── domain/          # Business entities and interfaces
│   ├── repository/      # Data access layer
│   ├── service/         # Business logic layer
│   └── handler/         # HTTP handlers
├── pkg/                 # Public packages
│   ├── auth/           # Authentication utilities
│   ├── middleware/     # HTTP middleware
│   ├── utils/          # Helper functions
│   └── version/        # Version management
├── api/                 # API definitions
├── configs/            # Configuration
├── test/               # Tests
├── web/                # Frontend Vue.js app
├── docs/               # Documentation
├── design/             # Design assets
└── migrations/         # Database migrations
```

## Available Commands

### Backend (Makefile)

```bash
make help              # Show all available commands
make build             # Build the application
make run               # Run the application
make test              # Run all tests with coverage
make test-unit         # Run unit tests only
make lint              # Run linters
make fmt               # Format code
make clean             # Clean build artifacts
make install-tools     # Install development tools
```

### Frontend

```bash
npm run dev            # Start development server
npm run build          # Build for production
npm run preview        # Preview production build
npm run lint           # Run ESLint
npm run format         # Format code with Prettier
```

## Documentation

Comprehensive documentation is available in the `docs/` directory:

- [Architecture](docs/ARCHITECTURE.md) - System architecture and design patterns
- [Database Schema](docs/DATABASE_SCHEMA.md) - Database structure and ERD
- [Requirements](docs/REQUIIREMENTS.md) - Project requirements and user stories
- [AI Instructions](docs/AI_INSTRUCTIONS.md) - Development guidelines

## Configuration

Configuration is managed through environment variables. See [.env.example](.env.example) for all available options.

Key configuration:
- `APP_ENV`: Environment (development, staging, production)
- `DB_DRIVER`: Database driver (sqlite, postgres, mysql)
- `JWT_SECRET`: Secret key for JWT tokens (MUST change in production!)
- `SERVER_PORT`: Server port (default: 8080)

## Testing

```bash
# Run all tests
make test

# Run unit tests only
make test-unit

# Run integration tests
make test-integration

# View coverage report
make coverage
```

## Security

- **Passwords**: Hashed with bcrypt (cost factor 12+)
- **Authentication**: JWT with secure secret keys
- **SQL Injection**: Parameterized queries only
- **CORS**: Configurable allowed origins
- **TLS/SSL**: Required in production

⚠️ **Important**: Change `JWT_SECRET` before deploying to production!

## Contributing

See [CONTRIBUTING.md](docs/CONTRIBUTING.md) for development guidelines.

1. Follow Clean Architecture principles
2. Write tests for new features
3. Run linters before committing
4. Follow Go and Vue.js best practices

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Support

For issues, questions, or feature requests, please open an issue on GitHub.

## Roadmap

- [ ] User authentication and registration
- [ ] Workout logging functionality
- [ ] Movement database with standard CrossFit movements
- [ ] Progress tracking with charts
- [ ] Data import/export
- [ ] PWA support for offline access
- [ ] Mobile apps (iOS/Android)
- [ ] Social features and leaderboards

---

**Built with ❤️ for the CrossFit community** 
