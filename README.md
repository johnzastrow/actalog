# ActaLog

> A mobile-first fitness tracker for CrossFit enthusiasts to log workouts, track progress, and analyze performance.

[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)](https://golang.org)
[![Vue.js](https://img.shields.io/badge/Vue.js-3.x-4FC08D?style=flat&logo=vue.js)](https://vuejs.org)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![CI](https://github.com/johnzastrow/actalog/actions/workflows/ci.yml/badge.svg)](https://github.com/johnzastrow/actalog/actions/workflows/ci.yml)

![ActaLog Logo](docs/images/logo_sm.png)

ActaLog is an open-source web application designed for CrossFit/Functional Fitness athletes to log their workouts, monitor progress, and analyze performance over time. Built with Go on the backend and Vue.js on the frontend, ActaLog offers a responsive and user-friendly interface optimized for mobile devices.

## Screenshots

<!--
Source - https://stackoverflow.com/a
Posted by alciregi, modified by community. See post 'Timeline' for change history
Retrieved 2025-11-23, License - CC BY-SA 4.0
-->

<img src="https://github.com/johnzastrow/actalog/blob/dc64830e14e2624e7ba35dfd4b6386b620b12230/docs/images/dash_with_annie.png" width="250">
1. Dashboard with Workout Summary
   



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

1. **Access the application**

   - Frontend: `http://localhost:3000`
   - Backend API: `http://localhost:8080`
   - API Health: `http://localhost:8080/health`


## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Support

For issues, questions, or feature requests, please open an issue on GitHub.

See the top-level Roadmap section for current status and next priorities (keeps a single authoritative roadmap in this README).



[imageLogoRef]: images/logo.png
