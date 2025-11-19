# Frontend Starter Scripts

This guide explains the `scripts/start-frontend.sh` (Linux / Git Bash / WSL) helper and the Windows wrapper `scripts/start-frontend.bat`. The scripts guide you through running the frontend in development or preview (production-like) modes and can optionally create local HTTPS certs using `mkcert`.

## Usage (Linux / Git Bash / WSL)

From the repository root run:

```bash
./scripts/start-frontend.sh
```

The script will prompt for:

- Mode: `dev` (vite dev server, HMR) or `preview` (build + vite preview)
- Host type: `localhost` or a domain name (for testing a mapped host)
- Whether to use HTTPS (the script can optionally generate certs using `mkcert`)

When HTTPS is selected and `mkcert` is available the script can:
- Run `mkcert -install` to install a local CA (may require elevation)
- Generate cert and key files into `web/certs/<hostname>.pem` and `web/certs/<hostname>-key.pem`

It prints the exact `vite.config.js` snippet to add (or you can run preview with `--https --host`).

### Usage (Windows)

Run the wrapper if you have `bash` available (Git Bash, MSYS, or WSL):

```powershell
scripts\start-frontend.bat
```

The wrapper checks for `bash` and invokes the same guided script. If `mkcert` is installed on Windows but not visible to `bash`, the wrapper/script attempts to locate `mkcert.exe` using `where mkcert` and convert the returned path so the bash environment can execute it.

If you have Git Bash but `mkcert` is installed via Chocolatey/Scoop, add the directory containing `mkcert.exe` to your Git Bash PATH (e.g. add an `export PATH=$PATH:/c/Users/you/scoop/shims` line to `~/.bashrc`).

### Vite configuration

If you generated certs, add the following (example) to `web/vite.config.js` so Vite uses the created certs for dev/preview HTTPS. The `start-frontend.sh` script prints the recommended snippet; here is the canonical example:

```js
import fs from 'fs'

export default defineConfig({
  // ... other config ...
  server: {
    host: '<your-hostname>',
    https: {
      key: fs.readFileSync('certs/<your-hostname>-key.pem'),
      cert: fs.readFileSync('certs/<your-hostname>.pem')
    }
  }
})
```

Alternatively you can avoid editing `vite.config.js` and run the preview server with the generated certs using:

```bash
cd web
npm run build
npm run preview -- --https --host <your-hostname>
```

### Windows-specific notes

- To find where Windows installed `mkcert`, run in PowerShell:

```powershell
where mkcert
# or
Get-Command mkcert | Select-Object -ExpandProperty Source
```

- Convert the path for Git Bash if needed (example):

```bash
export PATH=$PATH:/c/Users/you/scoop/shims
# or call the exe directly:
/c/Users/you/scoop/shims/mkcert.exe -install
```

### Troubleshooting

- mkcert not found in Bash: Either add the Windows binary's parent folder to your Bash `PATH` or install mkcert inside WSL.
- Permission prompts: `mkcert -install` may ask for admin/administrator permissions to add the local CA to the OS store.
- HMR and origin mismatch: If you map a hostname in `/etc/hosts` (or Windows hosts file), start Vite with that host so HMR and service worker origins match: `npm run dev -- --host <your-hostname>`.
- If preview doesn't bind: confirm the `--host` and `--https` flags are passed and that `vite.config.js` does not override `server` in a conflicting way.

### Security reminder

Generated mkcert certs are only for local development and trust within your machine. Use CA-signed certs (Let's Encrypt, etc.) for public-facing production deployments. Do not commit `web/certs` to source control.

**Want automation?**

If you'd like, I can edit `web/vite.config.js` automatically to set `server.https` when cert files exist (the repo already contains optional logic to read from `web/certs`). I can also add an explicit `scripts` entry to `web/package.json` to start HTTPS dev mode if you prefer a single npm command.

---

File locations:

- `scripts/start-frontend.sh` — guided script (bash)
- `scripts/start-frontend.bat` — Windows wrapper
- `web/certs/` — location where generated certs are stored (ignored by default; do not commit)
