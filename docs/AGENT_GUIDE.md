# Agent Guide: ActaLog

This guide makes it easy for AI agents and humans to collaborate on this codebase. It explains how to run, test, and safely change things with minimal context switching.

## Quickstart

- Backend (Go): use Makefile
  - Build: make build
  - Run: make run
  - Dev auto-reload: make dev (requires air)
  - Tests: make test (coverage at coverage.html)
  - Lint/format: make lint and make fmt

- Frontend (Vue + Vite) in `web/`
  - Dev server: npm run dev
  - Build: npm run build
  - Lint: npm run lint

VS Code: Press Ctrl+Shift+B to see tasks; use the Launch configs:
- Go: Launch Backend
- Web: Vite Dev

## Project Map

- cmd/actalog/main.go: entrypoint and HTTP server
- internal/
  - domain/: core entities
  - repository/: DB access + migrations
  - service/: business logic + tests
  - handler/: HTTP handlers
- pkg/: auth, email, logger, middleware, version utilities
- configs/: configuration loader
- web/: Vue app
- docs/: documentation set

## Conventions for Agents

- Keep public behavior stable; add tests when changing handler/service logic.
- Prefer small scoped PRs; update docs if behavior or endpoints change.
- Follow Clean Architecture boundaries: handler -> service -> repository -> domain.
- Log and validate inputs; avoid panics; return typed errors from services.

## Common Tasks

1) Run full checks (lint + test)
- make fmt && make lint && make test

2) Add a new REST endpoint
- Add method to the relevant service with tests in internal/service/*_test.go
- Add handler in internal/handler/*
- Wire route in cmd/actalog/main.go
- Update docs/API.md if schema changes

3) Update DB schema
- Create migration via make migrate-create name=your_change
- Implement up/down SQL under internal/repository/migrations/ or top-level migrations/ (follow existing pattern)
- Update repositories + tests

## Prompts That Work Well

- “Add GET /api/ping endpoint returning version and time; include unit tests for handler and service. Keep APIs backwards compatible.”
- “Refactor UserWorkoutService.UpdateLoggedWorkout to reduce duplication; keep tests green.”
- “Create migration and repository changes to add ‘intensity’ optional field to user_workouts; surface in handlers and Vue UI minimally.”

## Security and Safety

- Validate all inputs at handlers; sanitize IDs and strings.
- Never log secrets or tokens.
- Use context timeouts for DB calls in services/repos.
- JWT secrets must come from config; keep defaults safe.

## Versioning

The application version is defined in pkg/version/version.go and surfaced at /version and logs. Increment on behavior changes.

## CI/CD

Lightweight CI can run: go fmt, golangci-lint, go test, and web lint/build. See .github/workflows/ci.yml (added by agents when requested).

## Troubleshooting

- Build cache is local to project (.cache); run make clean for a reset.
- If `air` not installed, use go run via make run.
