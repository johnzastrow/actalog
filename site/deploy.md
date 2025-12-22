# Deploy — Quickstart

This page contains a minimal quickstart for deploying ActaLog from the repository using Docker Compose.

## Requirements
- Docker 20.10+
- Docker Compose v2 (or `docker compose` integrated CLI)
- A Linux VPS, Docker Desktop, or a local machine

## Minimal quickstart
```bash
# clone the repo
git clone https://github.com/johnzastrow/actalog.git
cd actalog

# bring up the stack
docker compose up -d
```

The stack will start:
- `actalog` (backend service)
- `web` (frontend)
- `db` (postgres or mariadb depending on env)

## Environment
Customize database and mail settings in the repository's `.env` or `docker-compose.override.yml` as needed.

## Next steps
- Check logs with `docker compose logs -f`
- Open the frontend at `http://localhost:3000` (or your configured host)
- See the repo README for advanced deployment and backup instructions
