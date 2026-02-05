# ActaLog

> A mobile-first fitness tracker for CrossFit enthusiasts to log workouts, track progress, and analyze performance.

[![Version](https://img.shields.io/badge/version-0.22.0--beta-blue)](https://github.com/johnzastrow/actalog/blob/main/docs/CHANGELOG.md)
[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)](https://golang.org)
[![Vue.js](https://img.shields.io/badge/Vue.js-3.x-4FC08D?style=flat&logo=vue.js)](https://vuejs.org)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![CI](https://github.com/johnzastrow/actalog/actions/workflows/ci.yml/badge.svg)](https://github.com/johnzastrow/actalog/actions/workflows/ci.yml)

![ActaLog Logo](docs/images/logo.png)

ActaLog is an open-source web application designed for CrossFit/Functional Fitness athletes to log their workouts, monitor progress, and analyze performance over time. Built with Go on the backend and Vue.js on the frontend, ActaLog offers a responsive and user-friendly interface optimized for mobile devices.

## Screenshots

See the [primary website](https://johnzastrow.github.io/actalog/index.html)


# Videos - be sure to view the full playlist

* [Overview](https://www.youtube.com/embed/DfLPpVaMT2g?si=rj8rfLqS_XcnhPdi)
* [Registering, Quicklog and User Menus](https://www.youtube.com/embed/aMaUYZQMWs0?si=KiUI6Gr28tpd-qRa)
* [Recording custom WODs: can be converted to WODs for all users](https://www.youtube.com/embed/Tc_LFHwmTsI?si=WanjZu_8WkGfOf8r)
* [Personal workouts to workouts for all users](https://www.youtube.com/embed/IW3manK6iHM?si=qhhh4BiHzcGeHNND)


## Roadmap — Next priorities

See the Roadmap file [ROADMAP.md](docs/ROADMAP.md) for the current project status and next priorities.

For the full backlog and lower-priority items see [TODO.md](docs/TODO.md). For release history and completed features see [CHANGELOG.md](docs/CHANGELOG.md).


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

1. **Set up environment variables**

   ```bash
   cp .env.example .env
   # Edit .env with your configuration
   ```

1. **Install Go dependencies**

   ```bash
   go mod download
   ```

1. **Install frontend dependencies**

   ```bash
   cd web
   npm install
   cd ..
   ```

1. **Run the backend**

   ```bash
   # Terminal 1
   make run
   # Or: go run cmd/actalog/main.go
   ```

1. **Run the frontend**

   ```bash
   # Terminal 2
   cd web
   npm run dev
   ```

1. **Run tests**

   ```bash
   # Backend tests
   make test

   # Frontend tests
   cd web
   npm run test:run        # Single run
   npm run test:coverage   # With coverage report
   ```

1. **Access the application**

   - Frontend: `http://localhost:3000`
   - Backend API: `http://localhost:8080`
   - API Health: `http://localhost:8080/health`
   - **API Documentation**: `http://localhost:8080/api/docs/`

## API Documentation

ActaLog provides interactive API documentation via Swagger UI, making it easy to explore and test all available endpoints.

### Accessing Swagger UI

Once the server is running, navigate to:

```
http://localhost:8080/api/docs/
```

### Features

- **Interactive Documentation**: Browse all 129 API operations across 105 endpoints
- **Try It Out**: Test endpoints directly from the browser
- **Request/Response Examples**: See expected request bodies and response formats
- **Authentication Support**: Authorize with JWT tokens to test protected endpoints

### Authentication

Most endpoints require JWT authentication. To use protected endpoints in Swagger UI:

1. First, call `POST /api/auth/login` with your credentials
2. Copy the `token` from the response
3. Click the **Authorize** button (lock icon) at the top of the page
4. Enter: `Bearer <your-token>` (include the word "Bearer" followed by a space)
5. Click **Authorize** to apply the token to all requests

### Endpoint Categories

| Tag | Description |
|-----|-------------|
| `auth` | Login, register, password reset, token management |
| `users` | User profile and settings |
| `workouts` | Workout logging and history |
| `movements` | Exercise/movement definitions |
| `wods` | Workout of the Day management |
| `templates` | Reusable workout templates |
| `performance` | Analytics and statistics |
| `prs` | Personal records tracking |
| `notifications` | User notifications |
| `sessions` | Session management |
| `subscriptions` | Subscription management |
| `organizations` | Organization management |
| `admin` | Admin-only operations |
| `import-export` | Data import/export (CSV, JSON) |
| `backups` | System backup/restore |
| `audit` | Audit logs |

### Docker Deployment (Production)

For production deployment using Docker:

1. **Pull the latest image from GitHub Container Registry**

   ```bash
   docker pull ghcr.io/johnzastrow/actalog:latest
   ```

2. **Run the container**

   ```bash
   docker run -d \
     -p 8080:8080 \
     -v actalog-data:/app/data \
     -v actalog-uploads:/app/uploads \
     --name actalog \
     ghcr.io/johnzastrow/actalog:latest
   ```

3. **Access the application**

   - Full application (frontend + API): `http://localhost:8080`
   - Health check: `http://localhost:8080/health`

**Note:** In production, the Docker container serves both the frontend and backend from a single port (8080). The frontend is pre-built into static files, and the Go backend serves both the API and the static files. This is different from local development where the frontend runs on port 3000.

**Database Options:**

The application supports multiple databases via environment variables:

- **SQLite** (default): File-based, single-container deployment
- **PostgreSQL**: For production, use with docker-compose
- **MariaDB/MySQL**: Alternative production database

See `docker/docker-compose.*.yml` for multi-container setups with PostgreSQL or MariaDB.


## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Support

For issues, questions, or feature requests, please open an issue on GitHub.

See the top-level Roadmap section for current status and next priorities (keeps a single authoritative roadmap in this README).



[imageLogoRef]: images/logo.png
