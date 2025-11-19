# ActaLog Development Scripts

Quick reference for development scripts to manage backend and frontend services.

## Scripts Overview

| Script | Purpose | Platform |
|--------|---------|----------|
| `restart.sh` | Kill and restart both services | Linux/Mac |
| `restart.bat` | Kill and restart both services | Windows |
| `stop.sh` | Stop all services | Linux/Mac |
| `stop.bat` | Stop all services | Windows |

## Usage

### Linux/Mac

**Restart everything (with latest code):**
```bash
./restart.sh
```

**Stop everything:**
```bash
./stop.sh
```

### Windows

**Restart everything (with latest code):**
```cmd
restart.bat
```

**Stop everything:**
```cmd
stop.bat
```

## What the scripts do

### Restart Scripts (`restart.sh` / `restart.bat`)
1. Kill all running backend processes (ANY port, by process name and working directory)
2. Kill all running frontend processes (ANY port, by process name and working directory)
3. Build the backend (`make build`)
4. Start backend in background (logs to `backend.log`)
5. Start frontend in background (logs to `frontend.log`)
6. Display service URLs and PIDs

### Stop Scripts (`stop.sh` / `stop.bat`)
1. Kill all running backend processes (ANY port, by process name and working directory)
2. Kill all running frontend processes (ANY port, by process name and working directory)

**Process Detection:**
- **Linux/Mac:** Finds processes by name (`actalog`, `go`, `npm`, `vite`, `node`) and verifies they're running from the project directory
- **Windows:** Finds processes by command line patterns and process names, then validates against listening ports
- **No hardcoded ports:** Scripts will find and kill backend/frontend instances on ANY port

## Service URLs

After running `restart.sh` or `restart.bat`:

- **Backend API:** [http://localhost:8080](http://localhost:8080)
- **Frontend:** [http://localhost:3000](http://localhost:3000) or [http://localhost:5173](http://localhost:5173)

## Log Files

Both scripts create log files in the project root:

- `backend.log` - Backend server logs
- `frontend.log` - Frontend dev server logs

These files are git-ignored.

## Manual Control

**View backend logs:**
```bash
tail -f backend.log
```

**View frontend logs:**
```bash
tail -f frontend.log
```

**Check running services:**
```bash
lsof -ti:8080,3000,5173
```

**Manually kill services:**
```bash
lsof -ti:8080,3000,5173 | xargs kill -9
```

## Troubleshooting

**Scripts won't execute (Linux/Mac):**
```bash
chmod +x restart.sh stop.sh
```

**Ports already in use:**
- Run `./stop.sh` (or `stop.bat`) first
- Scripts automatically detect and kill processes on any port
- Manual kill (specific port): `lsof -ti:8080 | xargs kill -9`

**Multiple instances running:**
- Scripts intelligently find ALL instances related to the project
- Works regardless of which port the services are using
- Both direct process name and command line pattern matching

**Frontend not starting:**
- Check if `web/node_modules` exists
- Run `cd web && npm install` if needed

**Backend build fails:**
- Check `backend.log` for errors
- Ensure Go dependencies are up to date: `go mod tidy`

**Process detection notes:**
- **Windows:** Checks process command lines for `actionlog`, `actalog`, `vite`, `npm`, and validates listening ports in range 3000-6000 for frontend
- **Linux/Mac:** Uses working directory checking via `/proc/$pid/cwd` to ensure only project-related processes are killed
