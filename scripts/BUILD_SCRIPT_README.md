# Build Script Documentation

## Overview

The `build.sh` script is a comprehensive build and deployment tool for ActaLog. It supports three modes of operation:

1. **Fresh Installation** - Complete setup on a new Ubuntu system
2. **Update Mode** - Update all packages and rebuild
3. **Rebuild Mode** - Quick rebuild without updating system packages

## Usage

### Fresh Installation

For a brand new Ubuntu installation:

```bash
chmod +x build.sh
./build.sh
```

This will:
- Ask for confirmation before proceeding (interactive mode)
- Update system packages
- Install Go 1.25.0
- Install Node.js 24 (LTS)
- Install SQLite3
- Install development tools (make, git, curl, etc.)
- Install Go tools (air, golangci-lint, goimports)
- Install backend dependencies
- Install frontend dependencies
- Create `.env` files from examples with auto-generated JWT secret
- Build backend and frontend
- Initialize database setup
- Create `run.sh` convenience script

### Update Mode

To update an existing installation:

```bash
./build.sh --update
```

This will:
- **Non-interactive** (no prompts, fully automated)
- Update all system packages
- Update Go to the version specified in the script
- Update Node.js to the version specified in the script
- Update all Go tools to latest versions
- Update backend dependencies
- Update frontend dependencies (including security fixes)
- Rebuild both backend and frontend
- **Preserve** your `.env` files and database

**Perfect for:**
- Monthly maintenance updates
- Security updates
- Upgrading to new Go/Node.js versions
- Keeping dependencies fresh

### Rebuild Mode

For quick rebuilds after code changes:

```bash
./build.sh --rebuild
```

This will:
- **Non-interactive** (no prompts, fully automated)
- Update backend dependencies (`go mod download`)
- Update frontend dependencies (`npm install`)
- Rebuild backend
- Rebuild frontend
- **Skip** all system package and tool updates

**Perfect for:**
- Day-to-day development
- After pulling code changes
- Testing builds before deployment
- Quick iteration during development

### Help

View detailed help:

```bash
./build.sh --help
```

## Version Configuration

The script uses these version constants (edit at the top of the script):

```bash
GO_VERSION="1.25.0"  # Go version to install
NODE_VERSION="24"     # Node.js major version (LTS)
```

To change versions, edit these values before running the script.

## How Update Mode Works

### Go Updates
- Detects currently installed version
- Compares with target version in script
- Automatically upgrades if version is different
- No prompts, fully automated

### Node.js Updates
- Checks major version (e.g., v24.x.x)
- Upgrades if major version differs
- Uses NodeSource repository for latest LTS

### Go Tools Updates
- Updates `air` to latest version
- Updates `golangci-lint` to latest version
- Updates `goimports` to latest version

### Frontend Dependencies Updates
- Runs `npm update` to update packages within semver ranges
- Runs `npm audit fix` to fix security vulnerabilities
- Preserves package.json constraints

## Comparison of Modes

| Feature | Fresh Install | Update Mode | Rebuild Mode |
|---------|--------------|-------------|--------------|
| **Interactive** | Yes | No | No |
| **System Packages** | ✓ | ✓ | ✗ |
| **Go Installation** | ✓ | ✓ (upgrade) | ✗ |
| **Node.js Installation** | ✓ | ✓ (upgrade) | ✗ |
| **Go Tools** | ✓ | ✓ (update) | ✗ |
| **Backend Deps** | ✓ | ✓ | ✓ |
| **Frontend Deps** | ✓ | ✓ (update) | ✓ |
| **Build Backend** | ✓ | ✓ | ✓ |
| **Build Frontend** | ✓ | ✓ | ✓ |
| **Create .env** | ✓ | ✗ | ✗ |
| **Init Database** | ✓ | ✗ | ✗ |
| **Speed** | Slow | Medium | Fast |

## Examples

### Initial Setup

```bash
# Clone the repository
git clone https://github.com/yourusername/actionlog.git
cd actionlog

# Run fresh installation
./build.sh

# Review and edit .env file
nano .env

# Start the application
./run.sh
```

### Monthly Maintenance

```bash
# Pull latest code
cd actionlog
git pull

# Update everything and rebuild
./build.sh --update

# Restart the application
./run.sh
```

### Development Workflow

```bash
# Make code changes
vim internal/service/workout_service.go

# Quick rebuild to test
./build.sh --rebuild

# Run tests
make test

# If tests pass, commit
git add .
git commit -m "feat: improve workout service"
```

### Updating to New Go Version

```bash
# Edit build.sh
nano build.sh
# Change: GO_VERSION="1.26.0"

# Run update
./build.sh --update

# Verify
go version
```

## What Gets Preserved

The update and rebuild modes preserve:

- ✓ `.env` files (backend and frontend)
- ✓ Database files (`actalog.db`)
- ✓ User data
- ✓ Configuration settings
- ✓ Git repository state

## Troubleshooting

### Permission Denied

```bash
chmod +x build.sh
```

## Which user should run `build.sh`?

- **Short recommendation:** Run `./build.sh` as your normal (non-root) user who has `sudo` privileges. Do NOT run the whole script with `sudo` (for example, avoid `sudo ./build.sh`) or as the `root` user.

- **Why:** The script uses `sudo` internally to install system packages. Running the script as your normal user allows `sudo` to elevate only when required, while keeping files created during the build (binaries, caches, `bin/`, `.cache/`) owned by your user. Running the script as root or with `sudo` causes files in the repository to be owned by root and later leads to permission errors for non-root builds.

- **Preferred commands:**

```bash
# Fresh install (will prompt and use sudo for package installs)
./scripts/build.sh

# Rebuild only (no system package installs; safe to run without sudo)
./scripts/build.sh --rebuild
```

- **Quick pre-checks (run before build):**

```bash
whoami
sudo -v && echo "sudo OK" || echo "sudo missing or not permitted"
node --version || echo "node: missing"
npm --version || echo "npm: missing"
go version || echo "go: missing"
```

- **If you accidentally ran the script as `root` or with `sudo` and see permission errors later:**

```bash
# From the repository root — make files owned by your user again
sudo chown -R $(whoami):$(whoami) .

# Then run a rebuild as your non-root user
./scripts/build.sh --rebuild
```

- **Avoiding cache permission issues without changing ownership:**

You can set per-user Go cache directories before running the build to ensure the Go toolchain has writable cache locations:

```bash
export GOCACHE="$HOME/.cache/go-build"
export GOMODCACHE="$HOME/.cache/go-mod"
mkdir -p "$GOCACHE" "$GOMODCACHE"
./scripts/build.sh --rebuild
```

- **If you cannot use `sudo` on the machine:**
	- Ask an administrator to run the update/install parts (`./scripts/build.sh`) once, or to install required system packages, then use `./scripts/build.sh --rebuild` locally for subsequent builds.


### Go/Node Already Installed

In fresh install mode, the script will ask if you want to reinstall. In update mode, it will automatically upgrade if needed.

### NPM Audit Failures

The script uses `npm audit fix || true` to prevent failures if automatic fixes aren't available. Manual review may be needed.

### Database Locked

Stop any running ActaLog instances before running the script:

```bash
killall actalog
./build.sh --rebuild
```

## Advanced Usage

### Automated Deployment

For automated deployments (CI/CD):

```bash
# Update without prompts
./build.sh --update

# Or rebuild only
./build.sh --rebuild
```

### Custom Go/Node Versions

Edit the version constants before running:

```bash
# Edit versions
sed -i 's/GO_VERSION="1.25.0"/GO_VERSION="1.26.0"/' build.sh
sed -i 's/NODE_VERSION="24"/NODE_VERSION="26"/' build.sh

# Run update
./build.sh --update
```

## Requirements

- Ubuntu 20.04+ (tested on 20.04, 22.04, 24.04)
- sudo privileges
- Internet connection
- At least 2GB free disk space

## Exit Codes

- `0` - Success
- `1` - Invalid arguments or user cancelled
- Non-zero - Build or installation error

## Support

For issues with the build script:
1. Check `build.sh --help`
2. Review error messages (color-coded)
3. Ensure system requirements are met
4. Open an issue on GitHub
